<template>
  <div class="password-input-wrap">
    <input
      :type="visible ? 'text' : 'password'"
      :value="modelValue"
      :class="['pi-input', { 'pi-input--compact': compact }]"
      v-bind="$attrs"
      @input="$emit('update:modelValue', $event.target.value)"
    />
    <button
      type="button"
      class="password-toggle"
      tabindex="-1"
      :aria-label="visible ? '隐藏密码' : '显示密码'"
      :aria-pressed="visible"
      @click.prevent="visible = !visible"
    >
      <!-- eye open -->
      <svg v-if="!visible" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
        <circle cx="12" cy="12" r="3"/>
      </svg>
      <!-- eye off -->
      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
        <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
        <line x1="1" y1="1" x2="23" y2="23"/>
      </svg>
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue'

defineOptions({ inheritAttrs: false })

defineProps({
  modelValue: { type: String, default: '' },
  compact: { type: Boolean, default: false }
})

defineEmits(['update:modelValue'])

const visible = ref(false)
</script>

<style scoped>
.password-input-wrap {
  position: relative;
  width: 100%;
}

.pi-input {
  width: 100%;
  box-sizing: border-box;
  padding: 12px 44px 12px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  transition: all 0.2s;
  background: white;
  color: #0f172a;
}

.pi-input--compact {
  padding: 10px 40px 10px 12px;
  border-radius: 8px;
  font-size: 13px;
}

.pi-input:focus {
  outline: none;
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.pi-input::placeholder {
  color: #cbd5e1;
}

.password-toggle {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  border: none;
  background: transparent;
  color: #94a3b8;
  border-radius: 8px;
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
}

.password-toggle:hover {
  color: #6366f1;
  background: rgba(99, 102, 241, 0.08);
}

.password-toggle svg {
  width: 20px;
  height: 20px;
}
</style>
