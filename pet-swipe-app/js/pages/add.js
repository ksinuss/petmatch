import { h, icon, toast } from '../ui.js';
import { api } from '../api.js';

const TYPES = [
  ['dog', 'Собака'],
  ['cat', 'Кошка'],
  ['other', 'Другой'],
];

export function AddPetPage(container, { go }) {
  let type = 'dog';
  let file = null;
  let previewUrl = '';

  const fileInput = h('input', { type: 'file', accept: 'image/jpeg,image/png,image/webp', hidden: true });
  const drop = h('button', { type: 'button', class: 'drop', onclick: () => fileInput.click() });
  const errorBox = h('p', { class: 'form-error', role: 'alert' });
  const typeSeg = h('div', { class: 'seg', role: 'radiogroup', 'aria-label': 'Вид животного' });
  const submitBtn = h('button', { class: 'btn btn-primary btn-block btn-lg', type: 'submit' }, 'Опубликовать анкету');

  function drawDrop() {
    drop.replaceChildren(previewUrl
      ? h('img', { src: previewUrl, alt: 'Выбранное фото' })
      : h('span', { class: 'drop-empty' }, icon('camera', 34), h('strong', null, 'Добавить фото'), h('small', null, 'JPEG, PNG или WebP до 5 МБ')));
    drop.classList.toggle('has', !!previewUrl);
  }

  function drawType() {
    typeSeg.replaceChildren(...TYPES.map(([value, label]) =>
      h('button', {
        type: 'button', role: 'radio', 'aria-checked': String(type === value),
        class: 'seg-btn' + (type === value ? ' on' : ''),
        onclick: () => { type = value; drawType(); },
      }, label)));
  }

  fileInput.addEventListener('change', () => {
    const f = fileInput.files?.[0];
    if (!f) return;
    if (!/^image\/(jpeg|png|webp)$/.test(f.type)) { toast('Подойдут только JPEG, PNG или WebP', 'error'); fileInput.value = ''; return; }
    if (f.size > 5 * 1024 * 1024) { toast('Фото больше 5 МБ — выберите поменьше', 'error'); fileInput.value = ''; return; }
    file = f;
    if (previewUrl) URL.revokeObjectURL(previewUrl);
    previewUrl = URL.createObjectURL(f);
    drawDrop();
  });

  const form = h('form', { class: 'form', novalidate: true, onsubmit: submit },
    drop, fileInput,
    h('div', { class: 'field' }, h('span', null, 'Кто это'), typeSeg),
    field('name', 'Кличка', 'text', 'Например, Рич', 50),
    field('breed', 'Порода', 'text', 'Можно оставить пустым', 60),
    h('label', { class: 'field field-age' },
      h('span', null, 'Возраст, лет'),
      h('input', { name: 'age', type: 'number', min: 0, max: 40, value: '1', inputMode: 'numeric' }),
      h('small', null, '0 — если меньше года'),
    ),
    h('label', { class: 'field' },
      h('span', null, 'Расскажите о характере'),
      h('textarea', { name: 'description', rows: 4, maxLength: 1000, placeholder: 'Привит, ладит с детьми, любит гулять…' }),
    ),
    errorBox,
    submitBtn,
  );

  container.append(
    h('header', { class: 'topbar' }, h('h1', { class: 'topbar-title' }, 'Новая анкета')),
    h('section', { class: 'page' }, form),
  );
  drawDrop();
  drawType();

  async function submit(e) {
    e.preventDefault();
    const f = new FormData(form);
    const name = String(f.get('name') || '').trim();
    const age = parseInt(String(f.get('age') || '0'), 10);

    errorBox.textContent = '';
    if (!name) { errorBox.textContent = 'Укажите кличку питомца'; return; }
    if (Number.isNaN(age) || age < 0 || age > 40) { errorBox.textContent = 'Возраст — число от 0 до 40'; return; }

    submitBtn.disabled = true;
    submitBtn.textContent = 'Публикуем…';
    try {
      const pet = await api.createPet({
        name, type, age,
        breed: String(f.get('breed') || '').trim(),
        description: String(f.get('description') || '').trim(),
        adoption_mode: 'open',
      });
      let photoFailed = false;
      if (file) {
        try { await api.uploadPhoto(pet.id, file); } catch { photoFailed = true; }
      }
      toast(photoFailed ? 'Анкета опубликована, но фото не загрузилось. Добавьте его позже.' : `Анкета «${pet.name}» опубликована`, photoFailed ? 'error' : 'success');
      go('#/profile');
    } catch (err) {
      errorBox.textContent = err.message;
      submitBtn.disabled = false;
      submitBtn.textContent = 'Опубликовать анкету';
    }
  }

  return () => { if (previewUrl) URL.revokeObjectURL(previewUrl); };
}

function field(name, label, type, placeholder, maxLength) {
  return h('label', { class: 'field' },
    h('span', null, label),
    h('input', { name, type, placeholder, maxLength, autocomplete: 'off' }),
  );
}
