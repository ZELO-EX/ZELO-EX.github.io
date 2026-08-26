package main

import (
	"strings"
	"testing"
)

func TestParseMeta(t *testing.T) {
	src := `#+TITLE: My Post
#+DATE: 2025-08-25
#+TAGS: :go:blog:
#+DRAFT: true

content`
	m := parseMeta(src)
	if m.Title != "My Post" {
		t.Errorf("Title = %q", m.Title)
	}
	if m.Date != "2025-08-25" {
		t.Errorf("Date = %q", m.Date)
	}
	if len(m.Tags) != 2 || m.Tags[0] != "go" || m.Tags[1] != "blog" {
		t.Errorf("Tags = %v", m.Tags)
	}
	if !m.Draft {
		t.Error("Draft should be true")
	}
}

func TestParseMetaDraftBare(t *testing.T) {
	if !parseMeta("#+DRAFT\n").Draft {
		t.Error("bare #+DRAFT should set Draft")
	}
	if parseMeta("#+DRAFT: false\n").Draft {
		t.Error("#+DRAFT: false should not set Draft")
	}
}

func TestHeadingLevels(t *testing.T) {
	html := renderOrg("* one\n** two\n***** five\n", "")
	want := "<h2>one</h2><h3>two</h3><h6>five</h6>"
	if html != want {
		t.Errorf("got %q want %q", html, want)
	}
}

func TestInlineMarkup(t *testing.T) {
	cases := []struct{ in, want string }{
		{"*bold*", "<p><strong>bold</strong></p>"},
		{"/italic/", "<p><em>italic</em></p>"},
		{"_under_", "<p><u>under</u></p>"},
		{"+strike+", "<p><del>strike</del></p>"},
		{"~code~", "<p><code>code</code></p>"},
		{"=code2=", "<p><code>code2</code></p>"},
		{"[[https://example.com][text]]", `<p><a href="https://example.com">text</a></p>`},
		{"see https://example.com now", `<p>see <a href="https://example.com">https://example.com</a> now</p>`},
		{"<script>", "<p>&lt;script&gt;</p>"},
	}
	for _, c := range cases {
		got := renderOrg(c.in, "")
		if got != c.want {
			t.Errorf("renderOrg(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNoEmphasisInsideWord(t *testing.T) {
	got := renderOrg("foo_bar and a/b and 2*3", "")
	want := "<p>foo_bar and a/b and 2*3</p>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestChineseEmphasisWithSpace(t *testing.T) {
	got := renderOrg("你好 *世界* 结束", "")
	want := "<p>你好 <strong>世界</strong> 结束</p>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSrcBlock(t *testing.T) {
	in := "#+begin_src go\nfunc f() {}\n#+end_src\n"
	got := renderOrg(in, "")
	want := `<pre class="src"><code class="language-go">func f() {}</code></pre>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestQuoteBlock(t *testing.T) {
	in := "#+begin_quote\nhello *world*\n#+end_quote\n"
	got := renderOrg(in, "")
	want := "<blockquote><p>hello <strong>world</strong></p></blockquote>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestTable(t *testing.T) {
	in := "| name | value |\n|------+-------|\n| foo  | 1     |\n"
	got := renderOrg(in, "")
	want := "<table><tr><th>name</th><th>value</th></tr><tr><td>foo</td><td>1</td></tr></table>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestNestedLists(t *testing.T) {
	in := "- a\n- b\n  1. x\n  2. y\n- c\n"
	got := renderOrg(in, "")
	want := "<ul><li>a</li><li>b<ol><li>x</li><li>y</li></ol></li><li>c</li></ul>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestHR(t *testing.T) {
	got := renderOrg("para\n-----\nafter\n", "")
	want := "<p>para</p><hr><p>after</p>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestImagePath(t *testing.T) {
	got := renderOrg("[[file:pic.png]]", "2025/08")
	want := `<p><img src="/assets/2025/08/pic.png" alt="pic.png"></p>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	// with alt text
	got = renderOrg("[[file:pic.png][a pic]]", "2025/08")
	want = `<p><img src="/assets/2025/08/pic.png" alt="a pic"></p>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestParagraphJoin(t *testing.T) {
	got := renderOrg("line one\nline two\n", "")
	want := "<p>line one\nline two</p>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestListContinuationKeepsBreak(t *testing.T) {
	got := renderOrg("- item\n  continued line\n", "")
	want := "<ul><li>item\ncontinued line</li></ul>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestURLTrailingPunct(t *testing.T) {
	got := renderOrg("(https://example.com).", "")
	want := `<p>(<a href="https://example.com">https://example.com</a>).</p>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestInlineCodeInBoldTextNotDoubled(t *testing.T) {
	got := renderOrg("*bold with ~code~ inside*", "")
	if strings.Count(got, "<code>") != 1 {
		t.Errorf("got %q", got)
	}
}
