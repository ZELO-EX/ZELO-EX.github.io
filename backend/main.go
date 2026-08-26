package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const defaultPort = "54345"

func main() {
	blogRoot := resolveDir(getenv("BLOG_DIR", "blog"))
	frontendRoot := resolveDir(getenv("FRONTEND_DIR", "frontend"))
	port := getenv("PORT", defaultPort)

	s := &server{blogRoot: blogRoot, frontendRoot: frontendRoot}

	cmd := "generate" // default: build static site into frontend/
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		cmd = "serve"
	}
	switch cmd {
	case "serve":
		s.serve(port)
	default:
		base := getenv("SITE_BASE", "https://zelo-ex.github.io")
		if len(os.Args) > 1 && os.Args[1] == "generate" && len(os.Args) > 2 {
			base = os.Args[2]
		}
		if err := s.generate(base); err != nil {
			log.Fatal(err)
		}
		log.Printf("generated static site into %s (base=%s)", frontendRoot, base)
	}
}

func (s *server) serve(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/posts", s.handlePosts)
	mux.HandleFunc("GET /api/post", s.handlePost)
	mux.HandleFunc("GET /api/tags", s.handleTags)
	mux.HandleFunc("GET /api/about", s.handleAbout)
	mux.HandleFunc("GET /posts.json", s.handlePosts)
	mux.HandleFunc("GET /tags.json", s.handleTags)
	mux.HandleFunc("GET /about.json", s.handleAbout)
	mux.HandleFunc("GET /rss.xml", s.handleRSS)
	mux.HandleFunc("/", s.handleStatic)

	handler := noStore(logRequests(mux))
	log.Printf("zal-blog listening on http://localhost:%s (blog=%s frontend=%s)", port, s.blogRoot, s.frontendRoot)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	blogRoot     string
	frontendRoot string
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// resolveDir allows running the binary both from the repo root
// (go run ./backend) and from inside backend/ (cd backend && go run .).
// Symlinks are resolved to their real path: filepath.WalkDir does not
// traverse a symlinked root, which would break post scanning.
func resolveDir(name string) string {
	if resolved, err := filepath.EvalSymlinks(name); err == nil {
		return resolved
	}
	if alt := filepath.Join("..", name); alt != "" {
		if resolved, err := filepath.EvalSymlinks(alt); err == nil {
			return resolved
		}
	}
	return name
}

func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no caching at all: edits to org/posts/frontend show up immediately
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf("%s %s", r.Method, r.URL.Path)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---------------------------------------------------------------------------
// post scanning
// ---------------------------------------------------------------------------

var postPathRe = regexp.MustCompile(`^\d{4}/\d{2}/[A-Za-z0-9._-]+\.org$`)

// scanPosts walks blogRoot for published posts (YYYY/MM/name.org).
// Drafts (*.sec.org) are excluded.
func (s *server) scanPosts() ([]PostMeta, error) {
	var posts []PostMeta
	err := filepath.WalkDir(s.blogRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != s.blogRoot {
				name := d.Name()
				if name == "assets" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
					return filepath.SkipDir
				}
			}
			return nil
		}
		rel, err := filepath.Rel(s.blogRoot, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !postPathRe.MatchString(rel) {
			return nil
		}
		base := filepath.Base(rel)
		if strings.HasPrefix(base, ".") || strings.HasPrefix(base, "_") {
			return nil
		}
		if strings.HasSuffix(base, ".sec.org") {
			return nil
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		meta := parseMeta(string(src))
		meta.Path = strings.TrimSuffix(rel, ".org")
		meta.fillFallbacks()
		posts = append(posts, meta)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sortPosts(posts)
	return posts, nil
}

// fillFallbacks derives Title/Date from content/path when keywords are absent.
func (m *PostMeta) fillFallbacks() {
	if m.Date == "" {
		parts := strings.Split(m.Path, "/")
		if len(parts) >= 2 {
			m.Date = parts[0] + "-" + parts[1]
		}
	}
}

func fallbackTitle(src string, pathSlug string) string {
	if m := headingRe.FindStringSubmatch(src); m != nil {
		return strings.TrimSpace(m[2])
	}
	return filepath.Base(pathSlug)
}

var dateLayouts = []string{
	"2006-01-02", "2006-01-02 15:04", "2006-01-02 15:04:05",
	"2006-01-02T15:04:05", "2006/01/02", "2006.01.02",
}

// parsePostTime tries the DATE field, then the YYYY/MM path prefix.
func parsePostTime(meta PostMeta) (time.Time, bool) {
	d := strings.TrimSpace(meta.Date)
	d = strings.Trim(d, "<>[]")
	for _, l := range dateLayouts {
		if t, err := time.Parse(l, d); err == nil {
			return t, true
		}
	}
	parts := strings.Split(meta.Path, "/")
	if len(parts) >= 2 {
		if t, err := time.Parse("2006-01", parts[0]+"-"+parts[1]); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func sortPosts(posts []PostMeta) {
	sort.Slice(posts, func(a, b int) bool {
		ta, oka := parsePostTime(posts[a])
		tb, okb := parsePostTime(posts[b])
		if oka && okb && !ta.Equal(tb) {
			return ta.After(tb)
		}
		if oka != okb {
			return oka
		}
		return posts[a].Path > posts[b].Path
	})
}

func displayDate(d string) string {
	return strings.TrimSpace(strings.Trim(d, "<>[]"))
}

// ---------------------------------------------------------------------------
// API handlers
// ---------------------------------------------------------------------------

type postSummary struct {
	Path  string   `json:"path"`
	Title string   `json:"title"`
	Date  string   `json:"date"`
	Tags  []string `json:"tags"`
}

func summarize(p PostMeta) postSummary {
	return postSummary{Path: p.Path, Title: p.Title, Date: displayDate(p.Date), Tags: p.Tags}
}

func (s *server) handlePosts(w http.ResponseWriter, r *http.Request) {
	posts, err := s.scanPosts()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	tag := r.URL.Query().Get("tag")
	list := make([]postSummary, 0, len(posts))
	for _, p := range posts {
		if tag != "" && !hasTag(p.Tags, tag) {
			continue
		}
		list = append(list, summarize(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"posts": list})
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}

var postSlugRe = regexp.MustCompile(`^\d{4}/\d{2}/[A-Za-z0-9._-]+$`)

func (s *server) handlePost(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Query().Get("p"), "/")
	if !postSlugRe.MatchString(p) {
		writeErr(w, http.StatusNotFound, "post not found")
		return
	}
	meta, htmlOut, err := s.loadPost(p)
	if err != nil {
		writeErr(w, http.StatusNotFound, "post not found")
		return
	}
	posts, err := s.scanPosts()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	prev, next := findNav(posts, p)
	detail := struct {
		postSummary
		HTML string       `json:"html"`
		Prev *postSummary `json:"prev"`
		Next *postSummary `json:"next"`
	}{postSummary: summarize(meta), HTML: htmlOut, Prev: prev, Next: next}
	writeJSON(w, http.StatusOK, detail)
}

// findNav returns the newer (prev) and older (next) neighbours of path
// within the date-descending published list.
func findNav(posts []PostMeta, path string) (prev, next *postSummary) {
	idx := -1
	for i, q := range posts {
		if q.Path == path {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, nil
	}
	if idx > 0 {
		ps := summarize(posts[idx-1])
		prev = &ps // newer
	}
	if idx < len(posts)-1 {
		ps := summarize(posts[idx+1])
		next = &ps // older
	}
	return prev, next
}

// loadPost reads and renders a single post (works for drafts too, for preview).
func (s *server) loadPost(slug string) (PostMeta, string, error) {
	full := filepath.Join(s.blogRoot, filepath.FromSlash(slug+".org"))
	src, err := os.ReadFile(full)
	if err != nil {
		return PostMeta{}, "", err
	}
	meta := parseMeta(string(src))
	meta.Path = slug
	meta.fillFallbacks()
	if meta.Title == "" {
		meta.Title = fallbackTitle(string(src), slug)
	}
	dir := strings.TrimSuffix(slug, "/"+filepath.Base(slug))
	dir = strings.TrimSuffix(dir, "/")
	htmlOut := renderOrg(string(src), dir)
	return meta, htmlOut, nil
}

type tagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

func buildTags(posts []PostMeta) []tagCount {
	counts := map[string]int{}
	for _, p := range posts {
		for _, t := range p.Tags {
			counts[t]++
		}
	}
	list := make([]tagCount, 0, len(counts))
	for t, c := range counts {
		list = append(list, tagCount{Tag: t, Count: c})
	}
	sort.Slice(list, func(a, b int) bool {
		if list[a].Count != list[b].Count {
			return list[a].Count > list[b].Count
		}
		return list[a].Tag < list[b].Tag
	})
	return list
}

func (s *server) handleTags(w http.ResponseWriter, r *http.Request) {
	posts, err := s.scanPosts()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tags": buildTags(posts)})
}

// buildAbout renders blog/about.org into the {title, html} shape used by
// both the API and the static generator.
func (s *server) buildAbout() (map[string]string, error) {
	src, err := os.ReadFile(filepath.Join(s.blogRoot, "about.org"))
	if err != nil {
		return nil, err
	}
	meta := parseMeta(string(src))
	title := meta.Title
	if title == "" {
		title = fallbackTitle(string(src), "about")
	}
	if title == "" || title == "about" {
		title = "About"
	}
	return map[string]string{
		"title": title,
		"html":  renderOrg(string(src), ""),
	}, nil
}

func (s *server) handleAbout(w http.ResponseWriter, r *http.Request) {
	about, err := s.buildAbout()
	if err != nil {
		writeErr(w, http.StatusNotFound, "about.org not found")
		return
	}
	writeJSON(w, http.StatusOK, about)
}

// ---------------------------------------------------------------------------
// static files
// ---------------------------------------------------------------------------

func (s *server) handleStatic(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	switch {
	case p == "/":
		s.serveFile(w, r, filepath.Join(s.frontendRoot, "index.html"))
	case p == "/blog" || p == "/tags" || p == "/about":
		s.serveFile(w, r, filepath.Join(s.frontendRoot, strings.TrimPrefix(p, "/")+".html"))
	case strings.HasPrefix(p, "/t/") && len(p) > 3:
		s.serveFile(w, r, filepath.Join(s.frontendRoot, "tags.html"))
	case strings.HasPrefix(p, "/posts/") && len(p) > 8:
		s.handlePostPage(w, r)
	case strings.HasPrefix(p, "/assets/"):
		rel := filepath.FromSlash(strings.TrimPrefix(p, "/assets/"))
		full, ok := safeJoin(filepath.Join(s.blogRoot, "assets"), rel)
		if !ok {
			s.notFound(w, r)
			return
		}
		s.serveFile(w, r, full)
	default:
		rel := filepath.FromSlash(strings.TrimPrefix(p, "/"))
		full, ok := safeJoin(s.frontendRoot, rel)
		if !ok {
			s.notFound(w, r)
			return
		}
		if _, err := os.Stat(full); err != nil {
			s.notFound(w, r)
			return
		}
		s.serveFile(w, r, full)
	}
}

func (s *server) serveFile(w http.ResponseWriter, r *http.Request, file string) {
	f, err := os.Open(file)
	if err != nil {
		s.notFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		s.notFound(w, r)
		return
	}
	http.ServeContent(w, r, filepath.Base(file), st.ModTime(), f)
}

func (s *server) notFound(w http.ResponseWriter, r *http.Request) {
	_ = r
	body, err := os.ReadFile(filepath.Join(s.frontendRoot, "404.html"))
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write(body); err != nil {
		log.Printf("notFound: write: %v", err)
	}
}

// safeJoin joins root and rel ensuring the result stays inside root.
func safeJoin(root, rel string) (string, bool) {
	clean := filepath.Clean(filepath.Join(root, rel))
	if clean == root {
		return clean, true
	}
	if strings.HasPrefix(clean, root+string(os.PathSeparator)) {
		return clean, true
	}
	return "", false
}
