import { h, icon } from '../ui.js';

export function WelcomePage(container, { go }) {
  container.append(
    h('section', { class: 'welcome' },
      h('div', { class: 'welcome-art', 'aria-hidden': 'true' },
        h('span', { class: 'paw-big p1' }, icon('paw', 150)),
        h('span', { class: 'paw-big p2' }, icon('paw', 90)),
        h('span', { class: 'paw-big p3' }, icon('paw', 64)),
      ),
      h('div', { class: 'welcome-body' },
        h('div', { class: 'brand brand-light' }, icon('paw', 28), h('span', null, 'PetMatch')),
        h('h1', null, 'Найди друга, который уже ждёт дом'),
        h('p', null, 'Листайте анкеты питомцев, которым нужен дом, ставьте лайки и публикуйте своих подопечных.'),
        h('div', { class: 'welcome-actions' },
          h('button', { class: 'btn btn-cream btn-lg', onclick: () => go('#/auth/register') }, 'Начать'),
          h('button', { class: 'btn btn-ghost-light', onclick: () => go('#/auth/login') }, 'У меня уже есть аккаунт'),
        ),
      ),
    ),
  );
}
