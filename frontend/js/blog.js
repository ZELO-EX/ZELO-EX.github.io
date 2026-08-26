import { getJSON, esc, postLink, tagsHTML } from '/js/common.js';

const list = document.getElementById('blog-list');
try {
  const data = await getJSON('/posts.json');
  list.innerHTML = data.posts.length
    ? data.posts.map(p =>
        `<a href="${postLink(p.path)}">${esc(p.title)}</a>` +
        `<span class="dim">${esc(p.date)}</span>${tagsHTML(p.tags)}`
      ).join('\n')
    : 'no posts yet';
} catch (e) {
  list.textContent = '[Error] ' + e.message;
}
