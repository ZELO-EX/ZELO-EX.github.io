import { getJSON } from '/js/common.js';

const titleEl = document.getElementById('about-title');
const contentEl = document.getElementById('about');

try {
  const d = await getJSON('/api/about');
  document.title = d.title + ' - zal blog';
  titleEl.textContent = d.title;
  contentEl.innerHTML = d.html;
} catch (e) {
  contentEl.textContent = '[Error] ' + e.message;
}
