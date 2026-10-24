// Клиент REST API. Все запросы идут на /api (его проксирует nginx на бэкенд),
// поэтому адрес сервера нигде не зашит и CORS не нужен.
const BASE = '/api';
const KEY = 'petmatch.session';

export const session = {
  read() {
    try { return JSON.parse(localStorage.getItem(KEY)) || null; } catch { return null; }
  },
  get token() { return this.read()?.token || null; },
  get user() { return this.read()?.user || null; },
  save(token, user) {
    try { localStorage.setItem(KEY, JSON.stringify({ token, user })); } catch { /* приватный режим */ }
    memory = { token, user };
  },
  clear() {
    try { localStorage.removeItem(KEY); } catch { /* ignore */ }
    memory = null;
  },
};
// Запасной вариант, если localStorage недоступен: сессия живёт до перезагрузки.
let memory = null;
const origRead = session.read.bind(session);
session.read = () => origRead() || memory;

export class ApiError extends Error {
  constructor(message, status = 0, code = 'unknown') {
    super(message);
    this.status = status;
    this.code = code;
  }
}

export async function request(method, path, body) {
  const headers = {};
  let payload;
  if (body instanceof FormData) {
    payload = body;
  } else if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
    payload = JSON.stringify(body);
  }
  if (session.token) headers.Authorization = `Bearer ${session.token}`;

  let res;
  try {
    res = await fetch(BASE + path, { method, headers, body: payload });
  } catch {
    throw new ApiError('Нет связи с сервером. Проверьте подключение и попробуйте ещё раз.', 0, 'network');
  }

  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { /* не JSON */ }

  if (!res.ok) {
    if (res.status === 401 && session.token && !path.startsWith('/auth/')) {
      session.clear();
      window.dispatchEvent(new Event('pm:logout'));
    }
    const fallback = res.status >= 500
      ? 'Сервер временно недоступен. Попробуйте через минуту.'
      : 'Не удалось выполнить запрос.';
    throw new ApiError(data?.error || fallback, res.status, data?.code);
  }
  return data;
}

// Фото из локального хранилища лежат на бэкенде (/uploads), демо-фото — в статике фронта.
export function photoUrl(url) {
  if (!url) return '';
  return url.startsWith('/uploads/') ? BASE + url : url;
}

export const api = {
  health: () => request('GET', '/health'),
  login: (email, password) => request('POST', '/auth/login', { email, password }),
  register: (data) => request('POST', '/auth/register', data),
  me: () => request('GET', '/me'),
  updateMe: (data) => request('PUT', '/me', data),
  feed: () => request('GET', '/pets/feed'),
  myPets: () => request('GET', '/my-pets'),
  createPet: (data) => request('POST', '/pets', data),
  uploadPhoto: (petId, file) => {
    const fd = new FormData();
    fd.append('photo', file);
    return request('POST', `/pets/${petId}/photos`, fd);
  },
  swipe: (petId, isLike) => request('POST', '/swipes', { pet_id: petId, is_like: isLike }),
  matches: () => request('GET', '/matches'),
  messages: (matchId) => request('GET', `/matches/${matchId}/messages`),
  send: (matchId, text) => request('POST', '/messages', { match_id: matchId, text }),
};
