import { h, emptyState, icon } from '../ui.js';

// Заглушка: переписка пока не реализована.
export function ChatsPage(container, { go }) {
  container.append(
    h('header', { class: 'topbar' }, h('h1', { class: 'topbar-title' }, 'Сообщения')),
    h('section', { class: 'page page-flush stub' },
      emptyState({
        title: 'Чаты скоро появятся',
        text: 'Здесь будет переписка с хозяевами питомцев после мэтча. Пока можно листать ленту и добавлять своих питомцев.',
        actions: [h('button', { class: 'btn btn-primary', onclick: () => go('#/feed') }, icon('paw', 18), 'Открыть ленту')],
      }),
    ),
  );
}
