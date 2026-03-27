import { reactive } from 'vue'

const state = reactive({
  toasts: [],
  confirmData: null
})

let toastId = 0

function addToast(message, type = 'info', duration = 3000) {
  const id = ++toastId
  state.toasts.push({ id, message, type, leaving: false })
  if (duration > 0) {
    setTimeout(() => removeToast(id), duration)
  }
}

function removeToast(id) {
  const idx = state.toasts.findIndex(t => t.id === id)
  if (idx !== -1) {
    state.toasts[idx].leaving = true
    setTimeout(() => {
      const i = state.toasts.findIndex(t => t.id === id)
      if (i !== -1) state.toasts.splice(i, 1)
    }, 300)
  }
}

export const toast = {
  success: (msg) => addToast(msg, 'success'),
  error: (msg) => addToast(msg, 'error', 5000),
  warning: (msg) => addToast(msg, 'warning', 4000),
  info: (msg) => addToast(msg, 'info'),
}

export function confirm(message) {
  return new Promise((resolve) => {
    state.confirmData = { message, resolve }
  })
}

export function useToastState() {
  return state
}

export function resolveConfirm(value) {
  if (state.confirmData) {
    state.confirmData.resolve(value)
    state.confirmData = null
  }
}

export { removeToast }
