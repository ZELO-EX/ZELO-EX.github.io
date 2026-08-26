package main

import (
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"time"
)

// buildRSS renders an RSS 2.0 feed manually: this Go toolchain's
// encoding/xml does not support the ",cdata" field option.
func (s *server) buildRSS(base string) []byte {
	posts, err := s.scanPosts()
	if err != nil {
		log.Printf("rss: scan: %v", err)
		return nil
	}
	var b strings.Builder
	b.WriteString(xml.Header)
	b.WriteString(`<rss version="2.0"><channel>`)
	b.WriteString("<title>" + xmlEscape("zal blog") + "</title>")
	b.WriteString("<link>" + xmlEscape(base+"/") + "</link>")
	b.WriteString("<description>" + xmlEscape("zal blog - personal notes") + "</description>")
	b.WriteString("<language>zh-CN</language>")

	for _, p := range posts {
		_, htmlOut, err := s.loadPost(p.Path)
		if err != nil {
			log.Printf("rss: skip %s: %v", p.Path, err)
			continue
		}
		link := base + "/posts/" + p.Path + ".html"
		b.WriteString("<item>")
		b.WriteString("<title>" + xmlEscape(p.Title) + "</title>")
		b.WriteString("<link>" + xmlEscape(link) + "</link>")
		b.WriteString("<guid>" + xmlEscape(link) + "</guid>")
		if t, ok := parsePostTime(p); ok {
			b.WriteString("<pubDate>" + xmlEscape(t.UTC().Format(time.RFC1123Z)) + "</pubDate>")
		}
		b.WriteString("<description><![CDATA[" + cdataSafe(htmlOut) + "]]></description>")
		b.WriteString("</item>")
	}
	b.WriteString("</channel></rss>")
	return []byte(b.String())
}

func (s *server) handleRSS(w http.ResponseWriter, r *http.Request) {
	base := "http://" + r.Host
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		base = "https://" + r.Host
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write(s.buildRSS(base))
}

func xmlEscape(s string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return s
	}
	return b.String()
}

// cdataSafe makes content safe inside <![CDATA[...]]> by splitting
// any "]]>" sequence.
func cdataSafe(s string) string {
	return strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>")
}
