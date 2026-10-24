// Мини-хелперы для построения DOM без innerHTML: пользовательский текст
// всегда попадает на страницу как текст, поэтому XSS через имена и описания невозможен.
export function h(tag, props, ...children) {
  const el = document.createElement(tag);
  for (const [key, value] of Object.entries(props || {})) {
    if (value == null || value === false) continue;
    if (key === 'class') el.className = value;
    else if (key.startsWith('on')) el.addEventListener(key.slice(2).toLowerCase(), value);
    else if (key === 'dataset') Object.assign(el.dataset, value);
    else if (key in el && key !== 'list') el[key] = value;
    else el.setAttribute(key, value === true ? '' : value);
  }
  const add = (child) => {
    if (child == null || child === false) return;
    if (Array.isArray(child)) child.forEach(add);
    else el.append(child instanceof Node ? child : document.createTextNode(String(child)));
  };
  children.forEach(add);
  return el;
}

const STROKE = 'fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"';
const PATHS = {
  paw: `<g fill="currentColor"><ellipse cx="6.500" cy="10.500" rx="2" ry="2.800"/><ellipse cx="10.200" cy="6" rx="2" ry="2.800"/><ellipse cx="15" cy="6" rx="2" ry="2.800"/><ellipse cx="18.600" cy="10.500" rx="2" ry="2.800"/><path d="M12.600 11.500c-3 0-5.800 3.300-5.800 6 0 2 1.700 3 3.300 2.600 1-.2 1.600-.5 2.500-.5s1.500.3 2.500.5c1.600.4 3.300-.6 3.300-2.600 0-2.700-2.800-6-5.800-6z"/></g>`,
  heart: `<path ${STROKE} d="M19 14c1.490-1.460 3-3.210 3-5.500A5.500 5.500 0 0 0 16.500 3c-1.760 0-3 .5-4.500 2-1.500-1.500-2.740-2-4.500-2A5.500 5.500 0 0 0 2 8.500c0 2.300 1.500 4.050 3 5.500l7 7Z"/>`,
  heartFill: `<path fill="currentColor" d="M19 14c1.490-1.460 3-3.210 3-5.500A5.500 5.500 0 0 0 16.500 3c-1.760 0-3 .5-4.500 2-1.500-1.500-2.740-2-4.500-2A5.500 5.500 0 0 0 2 8.500c0 2.300 1.500 4.050 3 5.500l7 7Z"/>`,
  x: `<path ${STROKE} d="M18 6 6 18M6 6l12 12"/>`,
  chat: `<path ${STROKE} d="M7.900 20A9 9 0 1 0 4 16.100L2 22Z"/>`,
  plus: `<path ${STROKE} d="M5 12h14M12 5v14"/>`,
  user: `<g ${STROKE}><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></g>`,
  send: `<path ${STROKE} d="m22 2-7 20-4-9-9-4ZM22 2 11 13"/>`,
  back: `<path ${STROKE} d="m12 19-7-7 7-7M19 12H5"/>`,
  logout: `<path ${STROKE} d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>`,
  pin: `<g ${STROKE}><path d="M20 10c0 6-8 12-8 12S4 16 4 10a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></g>`,
  camera: `<g ${STROKE}><path d="M14.500 4h-5L7 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-3l-2.500-3z"/><circle cx="12" cy="13" r="3"/></g>`,
  refresh: `<path ${STROKE} d="M21 12a9 9 0 1 1-3-6.700L21 8M21 3v5h-5"/>`,
};

export function icon(name, size = 22) {
  const span = document.createElement('span');
  span.className = 'icon';
  span.setAttribute('aria-hidden', 'true');
  span.style.width = span.style.height = size + 'px';
  span.innerHTML = `<svg viewBox="0 0 24 24" width="${size}" height="${size}">${PATHS[name] || ''}</svg>`;
  return span;
}

export function toast(message, kind = 'info') {
  const host = document.getElementById('toasts');
  const el = h('div', { class: `toast toast-${kind}` }, message);
  host.append(el);
  setTimeout(() => {
    el.classList.add('leaving');
    setTimeout(() => el.remove(), 250);
  }, 3600);
}

export const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const TYPES = { dog: 'Собака', cat: 'Кошка', other: 'Питомец' };
export const typeLabel = (t) => TYPES[t] || 'Питомец';

export function ageLabel(age) {
  if (!age) return 'меньше года';
  const m10 = age % 10, m100 = age % 100;
  if (m10 === 1 && m100 !== 11) return `${age} год`;
  if (m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14)) return `${age} года`;
  return `${age} лет`;
}


export function timeLabel(iso) {
  const d = new Date(iso);
  return isNaN(d) ? '' : d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

export function dayLabel(iso) {
  const d = new Date(iso);
  if (isNaN(d)) return '';
  const today = new Date();
  const yesterday = new Date(Date.now() - 864e5);
  if (d.toDateString() === today.toDateString()) return 'Сегодня';
  if (d.toDateString() === yesterday.toDateString()) return 'Вчера';
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
}

// Картинка питомца или заглушка с лапкой, если фото нет или оно не загрузилось.
export function petPhoto(url, className = '') {
  const wrap = h('div', { class: `photo ${className}` }, h('div', { class: 'photo-fallback' }, icon('paw', 56)));
  if (url) {
    const img = h('img', { src: url, alt: '', draggable: false, loading: 'lazy' });
    img.addEventListener('error', () => img.remove());
    wrap.append(img);
  }
  return wrap;
}

export function emptyState({ title, text, actions = [] }) {
  return h('div', { class: 'empty' },
    h('div', { class: 'empty-art' }, icon('paw', 44)),
    h('h2', null, title),
    h('p', null, text),
    actions.length ? h('div', { class: 'empty-actions' }, actions) : null,
  );
}
