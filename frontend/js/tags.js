import { getJSON, esc, postLink } from '/js/common.js';

const box = document.getElementById('tags-box');
const postsBox = document.getElementById('tag-posts');
const q = new URLSearchParams(location.search);
const active = q.get('t');

try {
  const data = await getJSON('/api/tags');
  box.innerHTML = data.tags.length
    ? data.tags.map(t =>
        `<a href="/tags.html?t=${encodeURIComponent(t.tag)}">#${esc(t.tag)}</a>` +
        `<span class="dim">(${t.count})</span>`
      ).join('\n')
    : 'no tags yet';

  if (active) {
    const posts = await getJSON('/api/posts?tag=' + encodeURIComponent(active));
    postsBox.innerHTML = '\n# ' + esc(active) + '\n' + (posts.posts.length
      ? posts.posts.map(p =>
          `<a href="${postLink(p.path)}">${esc(p.title)}</a>` +
          `<span class="dim">${esc(p.date)}</span>`
        ).join('\n')
      : 'no posts');
  }
} catch (e) {
  box.textContent = '[Error] ' + e.message;
}
