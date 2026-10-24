import { h, icon, toast, sleep, petPhoto, typeLabel, ageLabel, emptyState } from '../ui.js';
import { api, photoUrl } from '../api.js';

const SWIPE_DISTANCE = 110;

export function FeedPage(container, { go }) {
  let pets = [];
  let busy = false;
  let alive = true;
  let overlay = null;

  const stage = h('div', { class: 'stage' });
  const likeBtn = h('button', { class: 'round round-like', 'aria-label': 'Хочу познакомиться', onclick: () => decide(true) }, icon('heartFill', 30));
  const passBtn = h('button', { class: 'round round-pass', 'aria-label': 'Пропустить', onclick: () => decide(false) }, icon('x', 28));
  const actions = h('div', { class: 'swipe-actions' }, passBtn, likeBtn);

  container.append(
    h('header', { class: 'topbar' },
      h('div', { class: 'brand' }, icon('paw', 26), h('span', null, 'PetMatch')),
    ),
    h('section', { class: 'feed' }, stage, actions),
  );

  const onKey = (e) => {
    if (overlay || e.target.closest('input, textarea')) return;
    if (e.key === 'ArrowRight') decide(true);
    if (e.key === 'ArrowLeft') decide(false);
  };
  document.addEventListener('keydown', onKey);

  load();

  async function load() {
    stage.replaceChildren(h('div', { class: 'skeleton-card', 'aria-label': 'Загрузка' }));
    actions.hidden = true;
    try {
      pets = await api.feed();
    } catch (err) {
      if (!alive) return;
      stage.replaceChildren(emptyState({
        title: 'Лента не загрузилась',
        text: err.message,
        actions: [h('button', { class: 'btn btn-primary', onclick: load }, icon('refresh', 18), 'Повторить')],
      }));
      return;
    }
    if (alive) draw();
  }

  function draw() {
    stage.replaceChildren();
    actions.hidden = pets.length === 0;
    if (!pets.length) {
      stage.append(emptyState({
        title: 'Пока все питомцы просмотрены',
        text: 'Новые анкеты появляются каждый день. Загляните позже или опубликуйте своего питомца.',
        actions: [
          h('button', { class: 'btn btn-primary', onclick: () => go('#/add') }, icon('plus', 18), 'Добавить питомца'),
          h('button', { class: 'btn btn-soft', onclick: load }, icon('refresh', 18), 'Обновить'),
        ],
      }));
      return;
    }
    if (pets[1]) stage.append(card(pets[1], false));
    stage.append(card(pets[0], true));
  }

  function card(pet, top) {
    const stampLike = h('span', { class: 'stamp stamp-like' }, 'Хочу!');
    const stampPass = h('span', { class: 'stamp stamp-pass' }, 'Пока нет');
    const desc = h('p', { class: 'card-desc' }, pet.description || 'Подробное описание пока не добавлено.');

    const el = h('article', { class: 'card' + (top ? ' top' : ' behind'), 'aria-label': `${pet.name}, ${typeLabel(pet.type)}` },
      petPhoto(photoUrl(pet.photos?.[0]), 'card-photo'),
      h('div', { class: 'card-shade' }),
      h('div', { class: 'owner-chip' }, icon('paw', 16), h('span', null, pet.owner_name || 'Владелец')),
      stampLike, stampPass,
      h('div', { class: 'card-info' },
        h('div', { class: 'card-title' },
          h('h2', null, pet.name),
          h('span', { class: 'chip' }, ageLabel(pet.age)),
        ),
        h('p', { class: 'card-sub' }, pet.breed || typeLabel(pet.type)),
        desc,
      ),
    );
    if (top) enableDrag(el, stampLike, stampPass);
    return el;
  }

  function enableDrag(el, stampLike, stampPass) {
    let startX = 0, startY = 0, dx = 0, dragging = false, moved = false;

    el.addEventListener('pointerdown', (e) => {
      if (busy) return;
      dragging = true; moved = false; startX = e.clientX; startY = e.clientY; dx = 0;
      el.setPointerCapture(e.pointerId);
      el.style.transition = 'none';
    });
    el.addEventListener('pointermove', (e) => {
      if (!dragging) return;
      dx = e.clientX - startX;
      if (Math.abs(dx) > 6 || Math.abs(e.clientY - startY) > 6) moved = true;
      el.style.transform = `translateX(${dx}px) rotate(${dx * 0.05}deg)`;
      stampLike.style.opacity = Math.min(1, Math.max(0, dx / SWIPE_DISTANCE));
      stampPass.style.opacity = Math.min(1, Math.max(0, -dx / SWIPE_DISTANCE));
    });
    const end = () => {
      if (!dragging) return;
      dragging = false;
      el.style.transition = '';
      if (Math.abs(dx) > SWIPE_DISTANCE) {
        decide(dx > 0);
      } else {
        el.style.transform = '';
        stampLike.style.opacity = stampPass.style.opacity = 0;
        if (!moved) el.classList.toggle('open'); // тап — раскрыть описание
      }
    };
    el.addEventListener('pointerup', end);
    el.addEventListener('pointercancel', end);
  }

  async function decide(like) {
    if (busy || !pets.length) return;
    busy = true;
    const pet = pets[0];
    const el = stage.querySelector('.card.top');

    if (el) {
      el.classList.add('flying');
      el.style.transform = `translateX(${like ? 140 : -140}%) rotate(${like ? 18 : -18}deg)`;
      el.style.opacity = '0';
    }

    try {
      const [res] = await Promise.all([api.swipe(pet.id, like), sleep(260)]);
      pets.shift();
      if (res.match) showMatch(pet);
    } catch (err) {
      if (err.status === 409) {
        pets.shift(); // уже оценён — просто убираем из ленты
      } else {
        toast(err.message, 'error');
        if (el) { el.classList.remove('flying'); el.style.transform = ''; el.style.opacity = ''; }
        busy = false;
        return;
      }
    }
    if (!alive) return;
    if (pets.length) draw(); else await load();
    busy = false;
  }

  function showMatch(pet) {
    const close = () => { overlay?.remove(); overlay = null; };
    overlay = h('div', { class: 'match', role: 'dialog', 'aria-modal': 'true', 'aria-label': 'Новый мэтч' },
      h('div', { class: 'match-body' },
        h('div', { class: 'match-photos' },
          petPhoto(photoUrl(pet.photos?.[0]), 'match-photo'),
          h('span', { class: 'match-heart' }, icon('heartFill', 30)),
        ),
        h('h2', null, 'Это мэтч!'),
        h('p', null, `Вы понравились друг другу: ${pet.name} и ${pet.owner_name || 'хозяин'}. Чат с хозяином скоро появится — мы сохранили ваш мэтч.`),
        h('div', { class: 'match-actions' },
          h('button', { class: 'btn btn-cream btn-block btn-lg', onclick: close }, 'Смотреть дальше'),
        ),
      ),
    );
    document.getElementById('app').append(overlay);
    overlay.querySelector('button').focus();
  }

  return () => {
    alive = false;
    document.removeEventListener('keydown', onKey);
    overlay?.remove();
  };
}
