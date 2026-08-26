export async function getJSON(url) {
  const res = await fetch(url);
  if (!res.ok) {
    let msg = url + ' -> ' + res.status;
    try {
      const body = await res.json();
      if (body.error) msg = body.error;
    } catch (_) { /* ignore */ }
    throw new Error(msg);
  }
  return res.json();
}

export function esc(s) {
  return String(s).replace(/[&<>"']/g, c => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
  }[c]));
}

export function postLink(path) {
  return '/post.html?p=' + encodeURIComponent(path);
}

export function tagsHTML(tags) {
  if (!tags || !tags.length) return '';
  return ' ' + tags.map(t =>
    `<a href="/tags.html?t=${encodeURIComponent(t)}">#${esc(t)}</a>`).join(' ');
}
