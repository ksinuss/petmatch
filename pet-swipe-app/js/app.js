import { session } from './api.js';
import { h, icon } from './ui.js';
import { WelcomePage } from './pages/welcome.js';
import { AuthPage } from './pages/auth.js';
import { FeedPage } from './pages/feed.js';
import { AddPetPage } from './pages/add.js';
import { ChatsPage } from './pages/chats.js';
import { ProfilePage } from './pages/profile.js';

// Маршруты: hash-роутер, без перезагрузок страницы.
const ROUTES = {
  welcome: { page: WelcomePage, public: true },
  auth: { page: AuthPage, public: true },
  feed: { page: FeedPage, tab: 'feed' },
  add: { page: AddPetPage, tab: 'add' },
  chats: { page: ChatsPage, tab: 'chats' },
  profile: { page: ProfilePage, tab: 'profile' },
};

const TABS = [
  { id: 'feed', label: 'Лента', icon: 'paw', href: '#/feed' },
  { id: 'add', label: 'Анкета', icon: 'plus', href: '#/add' },
  { id: 'chats', label: 'Чаты', icon: 'chat', href: '#/chats' },
  { id: 'profile', label: 'Профиль', icon: 'user', href: '#/profile' },
];

const view = document.getElementById('view');
const tabs = document.getElementById('tabs');
let cleanup = null;

function go(hash) {
  if (location.hash === hash) render();
  else location.hash = hash;
}

function logout() {
  session.clear();
  go('#/welcome');
}

function drawTabs(active) {
  tabs.hidden = !active;
  if (!active) return;
  tabs.replaceChildren(...TABS.map((t) =>
    h('a', { href: t.href, class: 'tab' + (t.id === active ? ' on' : ''), 'aria-current': t.id === active ? 'page' : null },
      icon(t.icon, 24), h('span', null, t.label))));
}

function render() {
  if (typeof cleanup === 'function') cleanup();
  cleanup = null;

  const [name = '', ...params] = location.hash.replace(/^#\/?/, '').split('/');
  const authed = !!session.token;
  let route = ROUTES[name];

  // Гостя ведём на приветствие, а вошедшего — сразу в ленту.
  if (!route || (!route.public && !authed)) {
    return go(authed ? '#/feed' : '#/welcome');
  }
  if (route.public && authed) return go('#/feed');

  view.replaceChildren();
  view.scrollTop = 0;
  view.dataset.page = name;
  drawTabs(route.tab);
  cleanup = route.page(view, { params, go, logout });
}

window.addEventListener('hashchange', render);
window.addEventListener('pm:logout', () => go('#/welcome'));
render();
