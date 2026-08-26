import { getJSON, esc, postLink, tagsHTML } from '/js/common.js';

const titleEl = document.getElementById('post-title');
const contentEl = document.getElementById('post-content');
const navEl = document.getElementById('post-nav');

let p = new URLSearchParams(location.search).get('p');
if (!p && location.pathname.startsWith('/p/')) {
  p = decodeURIComponent(location.pathname.slice(3));
}

if (!p) {
  titleEl.textContent = encodeURIComponent(p);
  contentEl.textContent = 'no post specified, please <a href="/blog.html">visit blog</a>';
} else {
  try {
    const d = await getJSON('/api/post?p=' + encodeURIComponent(p));
    document.title = d.title + ' - zal blog';
    titleEl.textContent = d.title;
    contentEl.innerHTML = d.html;
    const nav = [];
    if (d.tags && d.tags.length) nav.push('tags:' + tagsHTML(d.tags));
    if (d.prev) nav.push(`<a href="${postLink(d.prev.path)}"><- newer: ${esc(d.prev.title)}</a>`);
    if (d.next) nav.push(`<a href="${postLink(d.next.path)}">older ->: ${esc(d.next.title)}</a>`);
    navEl.innerHTML = nav.join('\n');
  } catch (e) {
    titleEl.textContent = 'post';
    contentEl.textContent = '[Error] ' + e.message;
  }
}
