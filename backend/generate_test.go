package main

import (
	"strings"
	"testing"
)

func TestPostNavHTML(t *testing.T) {
	meta := PostMeta{Tags: []string{"go", "blog"}}
	prev := &postSummary{Path: "2025/08/newer", Title: "Newer"}
	next := &postSummary{Path: "2025/08/older", Title: "Older"}

	got := postNavHTML(meta, prev, next)
	for _, want := range []string{
		`<a href="/tags.html?t=go">#go</a>`,
		`<a href="/posts/2025/08/newer.html"><- newer: Newer</a>`,
		`<a href="/posts/2025/08/older.html">older ->: Older</a>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("postNavHTML missing %q in %q", want, got)
		}
	}
}

func TestBuildPostPage(t *testing.T) {
	meta := PostMeta{Path: "2025/08/x", Title: "X <Post>", Date: "2025-08-01"}
	page := buildPostPage(meta, "<p>raw &amp; html</p>", nil, nil)

	if !strings.Contains(page, "<title>X &lt;Post&gt; - zal-blog</title>") {
		t.Error("title not escaped")
	}
	if !strings.Contains(page, "<p>raw &amp; html</p>") {
		t.Error("content should be inserted raw")
	}
	if !strings.Contains(page, `<hr class="splitter" />`) {
		t.Error("splitter missing")
	}
}
