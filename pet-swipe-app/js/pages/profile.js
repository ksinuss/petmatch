import { h, icon, toast, petPhoto, typeLabel, ageLabel, emptyState } from '../ui.js';
import { api, session, photoUrl } from '../api.js';

export function ProfilePage(container, { go, logout }) {
  const head = h('div', { class: 'me' });
  const petsBox = h('div', { class: 'pet-grid' });
  container.append(
    h('header', { class: 'topbar' },
      h('h1', { class: 'topbar-title' }, 'Профиль'),
      h('button', { class: 'icon-btn', onclick: logout, 'aria-label': 'Выйти из аккаунта', title: 'Выйти' }, icon('logout')),
    ),
    h('section', { class: 'page' },
      head,
      h('div', { class: 'section-head' },
        h('h2', null, 'Мои питомцы'),
        h('button', { class: 'btn btn-soft btn-sm', onclick: () => go('#/add') }, icon('plus', 16), 'Добавить'),
      ),
      petsBox,
    ),
  );

  let alive = true;

  async function load() {
    try {
      const [me, pets] = await Promise.all([api.me(), api.myPets()]);
      if (!alive) return;
      session.save(session.token, me);
      drawMe(me);
      drawPets(pets);
    } catch (err) {
      if (!alive) return;
      head.replaceChildren(emptyState({
        title: 'Профиль не загрузился', text: err.message,
        actions: [h('button', { class: 'btn btn-primary', onclick: load }, 'Повторить')],
      }));
    }
  }

  function drawMe(me) {
    const errorBox = h('p', { class: 'form-error', role: 'alert' });
    const btn = h('button', { class: 'btn btn-primary', type: 'submit' }, 'Сохранить');
    const form = h('form', { class: 'form', novalidate: true,
      onsubmit: async (e) => {
        e.preventDefault();
        const f = new FormData(form);
        const name = String(f.get('display_name') || '').trim();
        errorBox.textContent = '';
        if (!name) { errorBox.textContent = 'Имя не может быть пустым'; return; }
        btn.disabled = true;
        try {
          const updated = await api.updateMe({
            display_name: name,
            location: String(f.get('location') || '').trim(),
            bio: String(f.get('bio') || '').trim(),
          });
          session.save(session.token, updated);
          drawMe(updated);
          toast('Профиль сохранён', 'success');
        } catch (err) {
          errorBox.textContent = err.message;
          btn.disabled = false;
        }
      } },
      h('label', { class: 'field' }, h('span', null, 'Имя'),
        h('input', { name: 'display_name', value: me.display_name, maxLength: 60 })),
      h('label', { class: 'field' }, h('span', null, 'Город'),
        h('input', { name: 'location', value: me.location || '', maxLength: 80, placeholder: 'Москва' })),
      h('label', { class: 'field' }, h('span', null, 'О себе'),
        h('textarea', { name: 'bio', rows: 3, maxLength: 500, placeholder: 'Пара слов о вас' }, me.bio || '')),
      errorBox, btn,
    );

    head.replaceChildren(
      h('div', { class: 'me-card' },
        h('div', { class: 'me-avatar', 'aria-hidden': 'true' }, (me.display_name || '?').trim().charAt(0).toUpperCase()),
        h('div', null,
          h('h2', null, me.display_name),
          h('p', null, me.email),
          me.location && h('p', { class: 'me-loc' }, icon('pin', 14), me.location),
        ),
      ),
      form,
    );
  }

  function drawPets(pets) {
    if (!pets.length) {
      petsBox.replaceChildren(emptyState({
        title: 'Анкет пока нет',
        text: 'Опубликуйте анкету питомца, которому ищете дом, — её увидят в ленте.',
        actions: [h('button', { class: 'btn btn-primary', onclick: () => go('#/add') }, 'Создать анкету')],
      }));
      return;
    }
    petsBox.replaceChildren(...pets.map((p) =>
      h('article', { class: 'pet-tile' },
        petPhoto(photoUrl(p.photos?.[0]), 'tile-photo'),
        h('div', { class: 'tile-info' },
          h('strong', null, p.name),
          h('span', null, `${typeLabel(p.type)}, ${ageLabel(p.age)}`),
        ),
      )));
  }

  load();
  return () => { alive = false; };
}
