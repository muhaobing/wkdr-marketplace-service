<template>
  <div ref="rootRef" class="ops-select" :class="[`ops-select--${variant}`]">
    <button
      type="button"
      class="ops-select__trigger"
      :class="{ 'is-open': open }"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="toggle"
    >
      <span class="ops-select__label">{{ displayLabel }}</span>
      <svg class="ops-select__chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
        <polyline points="6 9 12 15 18 9" />
      </svg>
    </button>

    <Teleport to="body">
      <Transition name="ops-select-fade">
        <div
          v-if="open"
          ref="menuRef"
          class="ops-select__menu"
          :class="[`ops-select__menu--${variant}`]"
          :style="menuStyle"
          role="listbox"
        >
          <ul class="ops-select__list">
            <li
              v-for="(opt, idx) in options"
              :key="optionKey(opt, idx)"
              role="option"
              class="ops-select__item"
              :class="{ 'is-selected': isSelected(opt) }"
              @mousedown.prevent="choose(opt)"
            >
              <span class="ops-select__item-text">{{ opt.label }}</span>
              <svg
                v-if="isSelected(opt)"
                class="ops-select__check"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
                aria-hidden="true"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </li>
          </ul>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, required: true },
  /** filter: 与筛选行输入框同高；form: 与弹窗表单项一致 */
  variant: { type: String, default: 'form' },
  placeholder: { type: String, default: '请选择' }
})

const emit = defineEmits(['update:modelValue'])

const rootRef = ref(null)
const menuRef = ref(null)
const open = ref(false)
const menuStyle = ref({})

const displayLabel = computed(() => {
  const opt = props.options.find(o => isSelected(o))
  return opt ? opt.label : props.placeholder
})

function optionKey(opt, idx) {
  const v = opt.value
  if (v === '' || v === null || v === undefined) return `empty-${idx}`
  return String(v)
}

function isSelected(opt) {
  const a = props.modelValue
  const b = opt.value
  if (Object.is(a, b)) return true
  if (b === '' && (a === '' || a === undefined || a === null)) return true
  return false
}

function choose(opt) {
  emit('update:modelValue', opt.value)
  open.value = false
}

function toggle() {
  open.value = !open.value
}

function updatePosition() {
  const el = rootRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const gap = 4
  const w = Math.max(r.width, props.variant === 'filter' ? 180 : 120)
  menuStyle.value = {
    position: 'fixed',
    top: `${r.bottom + gap}px`,
    left: `${r.left}px`,
    width: `${w}px`,
    zIndex: 4000
  }
}

function onDocPointerDown(e) {
  if (!open.value) return
  const root = rootRef.value
  const menu = menuRef.value
  const t = e.target
  if (root?.contains(t) || menu?.contains(t)) return
  open.value = false
}

function onScrollResize() {
  if (open.value) updatePosition()
}

function onKeydown(e) {
  if (e.key === 'Escape') open.value = false
}

let scrollParent = null

function bindScrollParent() {
  unbindScrollParent()
  const el = rootRef.value?.closest('.modal-content, .modal-overlay')
  if (el) {
    scrollParent = el
    scrollParent.addEventListener('scroll', onScrollResize, { passive: true })
  }
}

function unbindScrollParent() {
  if (scrollParent) {
    scrollParent.removeEventListener('scroll', onScrollResize)
    scrollParent = null
  }
}

watch(open, async v => {
  if (v) {
    await nextTick()
    updatePosition()
    bindScrollParent()
  } else {
    unbindScrollParent()
  }
})

onMounted(() => {
  document.addEventListener('pointerdown', onDocPointerDown, true)
  document.addEventListener('keydown', onKeydown)
  window.addEventListener('scroll', onScrollResize, true)
  window.addEventListener('resize', onScrollResize)
})

onUnmounted(() => {
  unbindScrollParent()
  document.removeEventListener('pointerdown', onDocPointerDown, true)
  document.removeEventListener('keydown', onKeydown)
  window.removeEventListener('scroll', onScrollResize, true)
  window.removeEventListener('resize', onScrollResize)
})
</script>

<style scoped>
.ops-select {
  position: relative;
  display: block;
  width: 100%;
}

.ops-select--filter {
  min-width: 180px;
}

.ops-select__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
  margin: 0;
  padding: 11px 14px;
  padding-right: 12px;
  border: 1px solid var(--gray-200);
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  font-family: inherit;
  color: var(--gray-800);
  line-height: 1.45;
  background-color: var(--gray-50);
  cursor: pointer;
  text-align: left;
  transition: border-color 0.2s, box-shadow 0.2s, background-color 0.2s;
}

.ops-select--filter .ops-select__trigger {
  height: 38px;
  padding: 0 12px 0 14px;
  box-sizing: border-box;
}

.ops-select__trigger:hover {
  background-color: var(--white);
  border-color: var(--gray-300);
}

.ops-select__trigger:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
  background-color: var(--white);
}

.ops-select__trigger.is-open {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
  background-color: var(--white);
}

.ops-select__label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ops-select__chevron {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  color: var(--gray-500);
  transition: transform 0.2s ease;
}

.ops-select__trigger.is-open .ops-select__chevron {
  transform: rotate(180deg);
  color: var(--accent);
}

.ops-select__menu {
  padding: 6px;
  border-radius: 12px;
  background: var(--white);
  border: 1px solid var(--gray-200);
  box-shadow: var(--shadow-lg), 0 0 0 1px rgba(15, 23, 42, 0.04);
  max-height: min(280px, 70vh);
  overflow-y: auto;
}

.ops-select__list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.ops-select__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 12px;
  margin: 2px 0;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--gray-700);
  cursor: pointer;
  transition: background-color 0.12s, color 0.12s;
}

.ops-select__item:first-child {
  margin-top: 0;
}

.ops-select__item:last-child {
  margin-bottom: 0;
}

.ops-select__item:hover {
  background-color: var(--gray-50);
  color: var(--gray-900);
}

.ops-select__item.is-selected {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1) 0%, rgba(99, 102, 241, 0.06) 100%);
  color: var(--accent);
}

.ops-select__item-text {
  flex: 1;
  min-width: 0;
}

.ops-select__check {
  flex-shrink: 0;
  width: 16px;
  height: 16px;
  color: var(--accent);
}

.ops-select-fade-enter-active,
.ops-select-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.ops-select-fade-enter-from,
.ops-select-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
