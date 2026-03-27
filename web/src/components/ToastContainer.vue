<template>
  <!-- Toast 通知 -->
  <div class="toast-wrapper">
    <div
      v-for="t in toasts"
      :key="t.id"
      :class="['toast-item', `toast-${t.type}`, { 'toast-leaving': t.leaving }]"
    >
      <div class="toast-icon">
        <svg v-if="t.type === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
          <polyline points="22 4 12 14.01 9 11.01"/>
        </svg>
        <svg v-else-if="t.type === 'error'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="15" y1="9" x2="9" y2="15"/>
          <line x1="9" y1="9" x2="15" y2="15"/>
        </svg>
        <svg v-else-if="t.type === 'warning'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
          <line x1="12" y1="9" x2="12" y2="13"/>
          <line x1="12" y1="17" x2="12.01" y2="17"/>
        </svg>
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="16" x2="12" y2="12"/>
          <line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
      </div>
      <span class="toast-message">{{ t.message }}</span>
      <button class="toast-close" @click="dismiss(t.id)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"/>
          <line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
    </div>
  </div>

  <!-- Confirm 弹窗 -->
  <Teleport to="body">
    <div v-if="confirmData" class="confirm-overlay" @click.self="handleConfirm(false)">
      <div class="confirm-dialog">
        <div class="confirm-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <p class="confirm-message">{{ confirmData.message }}</p>
        <div class="confirm-actions">
          <button class="confirm-btn confirm-cancel" @click="handleConfirm(false)">取消</button>
          <button class="confirm-btn confirm-ok" @click="handleConfirm(true)">确定</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { useToastState, removeToast, resolveConfirm } from '../utils/toast'

const state = useToastState()
const toasts = computed(() => state.toasts)
const confirmData = computed(() => state.confirmData)

function dismiss(id) {
  removeToast(id)
}

function handleConfirm(value) {
  resolveConfirm(value)
}
</script>

<style scoped>
.toast-wrapper {
  position: fixed;
  top: 80px;
  right: 24px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 10px;
  pointer-events: none;
}

.toast-item {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 300px;
  max-width: 440px;
  padding: 14px 16px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(0, 0, 0, 0.04);
  pointer-events: auto;
  animation: toastIn 0.35s cubic-bezier(0.16, 1, 0.3, 1);
  will-change: transform, opacity;
}

.toast-leaving {
  animation: toastOut 0.3s cubic-bezier(0.4, 0, 1, 1) forwards;
}

@keyframes toastIn {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes toastOut {
  from {
    opacity: 1;
    transform: translateX(0);
  }
  to {
    opacity: 0;
    transform: translateX(100%);
  }
}

.toast-icon {
  width: 22px;
  height: 22px;
  flex-shrink: 0;
}

.toast-icon svg {
  width: 22px;
  height: 22px;
}

.toast-success .toast-icon { color: #10b981; }
.toast-error .toast-icon { color: #ef4444; }
.toast-warning .toast-icon { color: #f59e0b; }
.toast-info .toast-icon { color: #6366f1; }

.toast-success { border-left: 3px solid #10b981; }
.toast-error { border-left: 3px solid #ef4444; }
.toast-warning { border-left: 3px solid #f59e0b; }
.toast-info { border-left: 3px solid #6366f1; }

.toast-message {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
  line-height: 1.4;
}

.toast-close {
  width: 20px;
  height: 20px;
  background: none;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 0;
  flex-shrink: 0;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.toast-close:hover {
  color: #475569;
  background: #f1f5f9;
}

.toast-close svg {
  width: 16px;
  height: 16px;
}

/* Confirm dialog */
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  animation: fadeIn 0.15s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.confirm-dialog {
  background: white;
  border-radius: 16px;
  padding: 32px;
  width: 380px;
  max-width: 90vw;
  box-shadow: 0 25px 60px rgba(0, 0, 0, 0.15);
  text-align: center;
  animation: dialogIn 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes dialogIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.confirm-icon {
  width: 48px;
  height: 48px;
  margin: 0 auto 16px;
  background: #fffbeb;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.confirm-icon svg {
  width: 24px;
  height: 24px;
  color: #f59e0b;
}

.confirm-message {
  font-size: 15px;
  font-weight: 500;
  color: #334155;
  line-height: 1.5;
  margin-bottom: 24px;
}

.confirm-actions {
  display: flex;
  gap: 10px;
}

.confirm-btn {
  flex: 1;
  padding: 11px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  border: none;
}

.confirm-cancel {
  background: #f1f5f9;
  color: #64748b;
}

.confirm-cancel:hover {
  background: #e2e8f0;
}

.confirm-ok {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.25);
}

.confirm-ok:hover {
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.35);
  transform: translateY(-1px);
}
</style>
