import { h, icon } from '../ui.js';
import { api, session } from '../api.js';

export function AuthPage(container, { params, go }) {
  let mode = params[0] === 'login' ? 'login' : 'register';
  const root = h('section', { class: 'auth' });
  container.append(root);

  async function submit(form, btn, errorBox) {
    const f = new FormData(form);
    const email = String(f.get('email') || '').trim();
    const password = String(f.get('password') || '');
    const name = String(f.get('name') || '').trim();

    const problem = validate({ email, password, name });
    if (problem) { errorBox.textContent = problem; return; }

    errorBox.textContent = '';
    btn.disabled = true;
    const label = btn.textContent;
    btn.textContent = 'Подождите…';
    try {
      const res = mode === 'login'
        ? await api.login(email, password)
        : await api.register({ email, password, display_name: name, role: 'user' });
      session.save(res.token, res.user);
      go('#/feed');
    } catch (err) {
      errorBox.textContent = err.message;
      btn.disabled = false;
      btn.textContent = label;
    }
  }

  function validate({ email, password, name }) {
    if (!email.includes('@') || !email.includes('.')) return 'Введите email в формате name@example.com';
    if (mode === 'register' && !name) return 'Укажите, как вас зовут';
    if (mode === 'register' && password.length < 6) return 'Пароль должен быть не короче 6 символов';
    if (!password) return 'Введите пароль';
    return '';
  }

  function draw() {
    const errorBox = h('p', { class: 'form-error', role: 'alert' });
    const submitBtn = h('button', { class: 'btn btn-primary btn-block', type: 'submit' },
      mode === 'login' ? 'Войти' : 'Создать аккаунт');

    const form = h('form', { class: 'form', novalidate: true, onsubmit: (e) => { e.preventDefault(); submit(form, submitBtn, errorBox); } },
      mode === 'register' && field('name', 'Как вас зовут', 'text', 'name', 'Анна'),
      field('email', 'Email', 'email', 'email', 'name@example.com'),
      field('password', 'Пароль', 'password', mode === 'login' ? 'current-password' : 'new-password', mode === 'login' ? '' : 'Не короче 6 символов'),
      errorBox,
      submitBtn,
    );

    root.replaceChildren(
      h('button', { class: 'icon-btn auth-back', onclick: () => go('#/welcome'), 'aria-label': 'Назад' }, icon('back')),
      h('div', { class: 'auth-head' },
        h('div', { class: 'brand' }, icon('paw', 26), h('span', null, 'PetMatch')),
        h('h1', null, mode === 'login' ? 'С возвращением' : 'Создайте аккаунт'),
        h('p', null, mode === 'login' ? 'Войдите, чтобы продолжить листать анкеты и переписываться.' : 'Это займёт минуту — и можно искать друга.'),
      ),
      h('div', { class: 'tabs-lite', role: 'tablist' },
        tabButton('register', 'Регистрация'),
        tabButton('login', 'Вход'),
      ),
      form,
    );
  }

  function tabButton(value, text) {
    return h('button', {
      type: 'button', role: 'tab', 'aria-selected': String(mode === value),
      class: 'tab-lite' + (mode === value ? ' on' : ''),
      onclick: () => { mode = value; draw(); },
    }, text);
  }

  function field(name, label, type, autocomplete, placeholder) {
    return h('label', { class: 'field' },
      h('span', null, label),
      h('input', { name, type, autocomplete, placeholder, required: true, maxLength: name === 'name' ? 60 : 120 }),
    );
  }

  draw();
}
