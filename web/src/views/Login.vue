<template>
  <div class="login-page">
    <div class="login-bg"></div>
    <div
      class="login-container"
      :class="{ 'login-container--bind': bindCheckPending || pageMode === 'bind' }"
    >
      <div class="login-card">
        <div class="login-header">
          <div class="logo-mark">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="9" cy="21" r="1"/>
              <circle cx="20" cy="21" r="1"/>
              <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
            </svg>
          </div>
          <h1 class="title">商城中心</h1>
          <p class="subtitle">
            {{
              bindCheckPending
                ? '正在校验绑定状态…'
                : pageMode === 'bind'
                  ? '绑定您的账户以继续'
                  : '登录您的账户以继续'
            }}
          </p>
        </div>

        <div v-if="bindCheckPending" class="login-bind-check">
          <span class="btn-loading login-bind-check__spinner" aria-hidden="true"></span>
          <p class="login-bind-check__text">校验平台账号是否已绑定商城…</p>
        </div>

        <!-- 绑定：双列网格 + 紧凑间距 -->
        <form
          v-else-if="pageMode === 'bind'"
          @submit.prevent="handleBind"
          class="login-form login-form--bind"
        >
          <div class="bind-grid">
            <div class="form-group form-group--compact">
              <label class="form-label" :id="bindPlatformLabelId">绑定平台 <span class="req">*</span></label>
              <div
                ref="bindPlatformSelectRef"
                class="custom-select"
                :class="{
                  'custom-select--open': bindPlatformOpen,
                  'custom-select--disabled':
                    bindPlatformFromQuery ||
                    bindPlatformOptionsLoading ||
                    !bindPlatformOptions.length
                }"
              >
                <button
                  type="button"
                  class="custom-select__trigger form-input form-input--compact"
                  :class="{
                    'custom-select__trigger--placeholder':
                      !bindForm.biz_code || bindPlatformOptionsLoading
                  }"
                  :disabled="
                    bindPlatformFromQuery ||
                    bindPlatformOptionsLoading ||
                    !bindPlatformOptions.length
                  "
                  :aria-expanded="bindPlatformOpen"
                  :aria-haspopup="!bindPlatformFromQuery"
                  :aria-labelledby="bindPlatformLabelId"
                  @click.stop="toggleBindPlatform"
                >
                  <span class="custom-select__value">{{ bindPlatformTriggerLabel }}</span>
                  <svg class="custom-select__chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                    <polyline points="6 9 12 15 18 9" />
                  </svg>
                </button>
                <ul v-show="bindPlatformOpen && !bindPlatformFromQuery" class="custom-select__menu" role="listbox">
                  <li
                    v-for="opt in bindPlatformOptions"
                    :key="opt.value"
                    role="option"
                    class="custom-select__item"
                    :class="{ 'custom-select__item--active': bindForm.biz_code === opt.value }"
                    @click.stop="selectBindPlatform(opt.value)"
                  >
                    {{ opt.label }}
                  </li>
                </ul>
              </div>
            </div>

            <div class="form-group form-group--compact">
              <label class="form-label">绑定用户 ID <span class="req">*</span></label>
              <input
                v-model="bindForm.biz_user_id"
                type="text"
                inputmode="numeric"
                class="form-input form-input--compact"
                :class="{ 'form-input--prefill-locked': bindUserIdFromQuery }"
                placeholder="绑定用户ID"
                :readonly="bindUserIdFromQuery"
                required
              />
            </div>

            <div class="form-group form-group--compact">
              <label class="form-label form-label--with-hint">
                电话号码
                <button
                  type="button"
                  class="hint-bubble hint-bubble--align-start"
                  aria-label="填写说明"
                >
                  <span class="hint-bubble__icon" aria-hidden="true">
                    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none"/>
                      <circle cx="12" cy="8" r="1.35" fill="currentColor"/>
                      <path d="M12 10.75V16" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                    </svg>
                  </span>
                  <span class="hint-bubble__pop" role="tooltip">电话号码与邮箱地址至少填写一项</span>
                </button>
              </label>
              <input
                v-model="bindForm.tel_no"
                type="tel"
                class="form-input form-input--compact"
                placeholder="电话号码"
              />
            </div>

            <div class="form-group form-group--compact">
              <label class="form-label form-label--with-hint">
                邮箱地址
                <button
                  type="button"
                  class="hint-bubble hint-bubble--align-end"
                  aria-label="填写说明"
                >
                  <span class="hint-bubble__icon" aria-hidden="true">
                    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none"/>
                      <circle cx="12" cy="8" r="1.35" fill="currentColor"/>
                      <path d="M12 10.75V16" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                    </svg>
                  </span>
                  <span class="hint-bubble__pop" role="tooltip">电话号码与邮箱地址至少填写一项</span>
                </button>
              </label>
              <input
                v-model="bindForm.email"
                type="email"
                class="form-input form-input--compact"
                placeholder="邮箱地址"
              />
            </div>

            <div class="form-group form-group--compact">
              <label class="form-label">输入密码 <span class="req">*</span></label>
              <input
                v-model="bindForm.password"
                type="password"
                class="form-input form-input--compact"
                placeholder="设置登录密码"
                required
                autocomplete="new-password"
              />
            </div>

            <div class="form-group form-group--compact">
              <label class="form-label">确认密码 <span class="req">*</span></label>
              <input
                v-model="bindForm.password2"
                type="password"
                class="form-input form-input--compact"
                placeholder="再次输入"
                required
                autocomplete="new-password"
              />
            </div>
          </div>

          <div v-if="errorMsg" class="error-message">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="15" y1="9" x2="9" y2="15"/>
              <line x1="9" y1="9" x2="15" y2="15"/>
            </svg>
            {{ errorMsg }}
          </div>

          <button type="submit" class="login-btn" :disabled="loading">
            <span v-if="loading" class="btn-loading"></span>
            <span v-else>提交绑定</span>
          </button>
        </form>

        <!-- 登录（原表单） -->
        <form v-else @submit.prevent="handleLogin" class="login-form">
          <div class="login-tabs">
            <button
              type="button"
              :class="['tab', { active: loginType === 'email' }]"
              @click="loginType = 'email'"
            >
              邮箱登录
            </button>
            <button
              type="button"
              :class="['tab', { active: loginType === 'phone' }]"
              @click="loginType = 'phone'"
            >
              手机号登录
            </button>
          </div>

          <div v-if="loginType === 'email'" class="form-group">
            <label class="form-label">邮箱地址</label>
            <input
              v-model="formData.email"
              type="email"
              class="form-input"
              placeholder="请输入邮箱地址"
              required
            />
          </div>

          <div v-if="loginType === 'phone'" class="form-group">
            <label class="form-label">手机号码</label>
            <input
              v-model="formData.tel_no"
              type="tel"
              class="form-input"
              placeholder="请输入手机号码"
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label">用户密钥</label>
            <input
              v-model="formData.secret"
              type="password"
              class="form-input"
              placeholder="请输入用户密钥"
              required
            />
          </div>

          <div v-if="errorMsg" class="error-message">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="15" y1="9" x2="9" y2="15"/>
              <line x1="9" y1="9" x2="15" y2="15"/>
            </svg>
            {{ errorMsg }}
          </div>

          <button type="submit" class="login-btn" :disabled="loading">
            <span v-if="loading" class="btn-loading"></span>
            <span v-else>登 录</span>
          </button>
        </form>

        <div v-if="!bindCheckPending" class="login-mode-switch">
          <button
            v-if="pageMode === 'bind'"
            type="button"
            class="text-link"
            @click="switchToLogin"
          >
            已有账号，去登录
          </button>
          <button
            v-else
            type="button"
            class="text-link text-link-muted"
            @click="switchToBind"
          >
            没有账号？去绑定
          </button>
        </div>

        <div v-if="!bindCheckPending" class="login-footer">
          <p class="tip">
            {{ pageMode === 'bind' ? '绑定成功后自动登录商城' : '请使用您的账户密钥进行登录' }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../stores/user'
import { authApi, metaApi } from '../api'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

function hasBothValidBindQuery() {
  const qCode = route.query.biz_code
  const qUid = route.query.biz_user_id
  if (qCode === undefined || qCode === null || String(qCode).trim() === '') return false
  if (qUid === undefined || qUid === null || String(qUid).trim() === '') return false
  const n = parseInt(String(qUid), 10)
  return Number.isFinite(n) && n >= 1
}

const bindPlatformLabelId = 'bind-platform-label'
const bindPlatformOpen = ref(false)
const bindPlatformSelectRef = ref(null)
/** { value: code, label: name }，来自 GET /marketplace/biz_codes */
const bindPlatformOptions = ref([])
const bindPlatformOptionsLoading = ref(true)

const bindForm = reactive({
  biz_code: '',
  biz_user_id: '',
  tel_no: '',
  email: '',
  password: '',
  password2: ''
})

const bindPlatformDisplay = computed(() => {
  const v = bindForm.biz_code
  if (!v) return '请选择绑定平台'
  const o = bindPlatformOptions.value.find((x) => x.value === v)
  return o ? o.label : v
})

const bindPlatformTriggerLabel = computed(() => {
  if (bindPlatformOptionsLoading.value) return '加载平台列表…'
  if (!bindPlatformOptions.value.length && !bindForm.biz_code) return '暂无可用平台'
  return bindPlatformDisplay.value
})

async function loadBizCodes() {
  bindPlatformOptionsLoading.value = true
  try {
    const list = await metaApi.listBizCodes()
    bindPlatformOptions.value = (list || []).map(row => ({
      value: row.code,
      label: row.name
    }))
  } catch {
    bindPlatformOptions.value = []
  } finally {
    bindPlatformOptionsLoading.value = false
  }
}

function toggleBindPlatform() {
  if (bindPlatformFromQuery.value) return
  if (bindPlatformOptionsLoading.value) return
  if (!bindPlatformOptions.value.length) return
  bindPlatformOpen.value = !bindPlatformOpen.value
}

function selectBindPlatform(value) {
  if (bindPlatformFromQuery.value) return
  bindForm.biz_code = value
  bindPlatformOpen.value = false
}

function onBindDocClick(e) {
  const el = bindPlatformSelectRef.value
  if (!el || !bindPlatformOpen.value) return
  if (!el.contains(e.target)) bindPlatformOpen.value = false
}

function onBindDocKeydown(e) {
  if (e.key === 'Escape') bindPlatformOpen.value = false
}

const pageMode = ref('bind')
const bindCheckPending = ref(hasBothValidBindQuery())
/** 绑定平台 / 用户 ID 是否来自 URL 预填（预填时禁止修改） */
const bindPlatformFromQuery = ref(false)
const bindUserIdFromQuery = ref(false)
const loginType = ref('email')
const loading = ref(false)
const errorMsg = ref('')

const formData = reactive({
  email: '',
  tel_no: '',
  secret: ''
})

function syncBindPrefillLocksFromRoute() {
  const qCode = route.query.biz_code
  const qUid = route.query.biz_user_id
  bindPlatformFromQuery.value =
    qCode !== undefined && qCode !== null && String(qCode).trim() !== ''
  bindUserIdFromQuery.value =
    qUid !== undefined && qUid !== null && String(qUid).trim() !== ''
}

async function applyQueryAndBindCheck() {
  const qCode = route.query.biz_code
  const qUid = route.query.biz_user_id
  if (qCode !== undefined && qCode !== null && String(qCode).trim() !== '') {
    bindForm.biz_code = String(qCode).trim()
  }
  if (qUid !== undefined && qUid !== null && String(qUid) !== '') {
    bindForm.biz_user_id = String(qUid)
  }
  syncBindPrefillLocksFromRoute()
  if (bindPlatformFromQuery.value) {
    bindPlatformOpen.value = false
  }

  if (!hasBothValidBindQuery()) {
    bindCheckPending.value = false
    pageMode.value = 'bind'
    return
  }

  bindCheckPending.value = true
  const code = String(bindForm.biz_code || '').trim()
  const idStr = String(bindForm.biz_user_id || '').trim()
  const bizUserId = parseInt(idStr, 10)
  try {
    const data = await authApi.checkBinding({ biz_code: code, biz_user_id: bizUserId })
    pageMode.value = data.bound ? 'login' : 'bind'
  } catch {
    pageMode.value = 'bind'
  } finally {
    bindCheckPending.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onBindDocClick)
  document.addEventListener('keydown', onBindDocKeydown)
  loadBizCodes()
  applyQueryAndBindCheck()
})

watch(
  () => [route.query.biz_code, route.query.biz_user_id],
  () => {
    applyQueryAndBindCheck()
  }
)

onBeforeUnmount(() => {
  document.removeEventListener('click', onBindDocClick)
  document.removeEventListener('keydown', onBindDocKeydown)
})

function switchToLogin() {
  errorMsg.value = ''
  pageMode.value = 'login'
}

function switchToBind() {
  errorMsg.value = ''
  pageMode.value = 'bind'
}

async function handleBind() {
  errorMsg.value = ''

  const bizCode = String(bindForm.biz_code || '').trim()
  if (!bizCode) {
    errorMsg.value = '请选择绑定平台'
    return
  }

  const tel = String(bindForm.tel_no || '').trim()
  const email = String(bindForm.email || '').trim()
  if (!tel && !email) {
    errorMsg.value = '手机号码与邮箱至少填写一项'
    return
  }

  if (bindForm.password !== bindForm.password2) {
    errorMsg.value = '两次输入的密码不一致'
    return
  }

  const idStr = String(bindForm.biz_user_id || '').trim()
  if (!idStr) {
    errorMsg.value = '请填写绑定用户 ID'
    return
  }
  const bizUserId = parseInt(idStr, 10)
  if (!Number.isFinite(bizUserId) || bizUserId < 1) {
    errorMsg.value = '绑定用户 ID 须为正整数'
    return
  }

  loading.value = true
  try {
    const payload = {
      biz_code: bizCode,
      biz_user_id: bizUserId,
      password: bindForm.password
    }
    if (tel) payload.tel_no = tel
    if (email) payload.email = email

    await userStore.bindAccount(payload)

    const redirect = route.query.redirect || '/'
    router.push(redirect)
  } catch (error) {
    errorMsg.value = error.message || '绑定失败，请检查填写信息'
  } finally {
    loading.value = false
  }
}

async function handleLogin() {
  errorMsg.value = ''

  if (loginType.value === 'email' && !formData.email) {
    errorMsg.value = '请输入邮箱地址'
    return
  }
  if (loginType.value === 'phone' && !formData.tel_no) {
    errorMsg.value = '请输入手机号码'
    return
  }
  if (!formData.secret) {
    errorMsg.value = '请输入用户密钥'
    return
  }

  loading.value = true

  try {
    const credentials = {
      secret: formData.secret
    }

    if (loginType.value === 'email') {
      credentials.email = formData.email
    } else {
      credentials.tel_no = formData.tel_no
    }

    await userStore.login(credentials)

    const redirect = route.query.redirect || '/'
    router.push(redirect)
  } catch (error) {
    errorMsg.value = error.message || '登录失败，请检查账号密钥是否正确'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: #0f172a;
}

.login-bg {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 50% at 50% -20%, rgba(99, 102, 241, 0.3), transparent),
    radial-gradient(ellipse 60% 40% at 80% 50%, rgba(59, 130, 246, 0.15), transparent),
    radial-gradient(ellipse 50% 30% at 20% 80%, rgba(99, 102, 241, 0.1), transparent);
}

.login-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 440px;
  padding: 20px;
}

.login-container--bind {
  max-width: 520px;
}

.login-container--bind .login-card {
  padding: 32px 28px;
}

.login-container--bind .login-header {
  margin-bottom: 22px;
}

.login-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.3);
  padding: 44px 40px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  overflow: visible;
}

.login-header {
  text-align: center;
  margin-bottom: 36px;
}

.logo-mark {
  width: 56px;
  height: 56px;
  margin: 0 auto 16px;
  background: linear-gradient(135deg, #0f172a, #334155);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.3);
}

.logo-mark svg {
  width: 28px;
  height: 28px;
  color: white;
}

.title {
  font-size: 26px;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 6px 0;
  letter-spacing: -0.02em;
}

.subtitle {
  font-size: 14px;
  color: #94a3b8;
  margin: 0;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.login-form--bind {
  gap: 14px;
}

.login-bind-check {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  min-height: 200px;
  padding: 8px 0 24px;
}

.login-bind-check__spinner {
  border-color: rgba(15, 23, 42, 0.2);
  border-top-color: #0f172a;
}

.login-bind-check__text {
  margin: 0;
  font-size: 13px;
  color: #64748b;
  text-align: center;
}

.bind-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 14px;
  align-items: start;
  overflow: visible;
}

@media (max-width: 420px) {
  .bind-grid {
    grid-template-columns: 1fr;
  }
}

.form-label--with-hint {
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}

.hint-bubble {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin: 0;
  padding: 0;
  border: none;
  background: transparent;
  vertical-align: middle;
  cursor: help;
  outline: none;
  flex-shrink: 0;
}

.hint-bubble__icon {
  display: flex;
  width: 16px;
  height: 16px;
  color: #94a3b8;
  transition: color 0.15s;
}

.hint-bubble__icon svg {
  width: 100%;
  height: 100%;
}

.hint-bubble:hover .hint-bubble__icon,
.hint-bubble:focus-visible .hint-bubble__icon {
  color: #6366f1;
}

/* 气泡在图标下方展开，避免被页面 overflow 裁切；双列时左右对齐防溢出 */
.hint-bubble__pop {
  position: absolute;
  top: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
  padding: 9px 12px;
  min-width: 200px;
  max-width: min(260px, calc(100vw - 32px));
  font-size: 12px;
  font-weight: 500;
  line-height: 1.45;
  color: #f1f5f9;
  text-align: left;
  background: #0f172a;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 10px;
  box-shadow:
    0 4px 6px -1px rgba(15, 23, 42, 0.15),
    0 12px 28px rgba(15, 23, 42, 0.28);
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  transition: opacity 0.16s ease, visibility 0.16s ease, transform 0.16s ease;
  z-index: 100;
}

.hint-bubble__pop::before {
  content: '';
  position: absolute;
  bottom: 100%;
  left: 50%;
  margin-left: -6px;
  border: 6px solid transparent;
  border-bottom-color: #0f172a;
}

.hint-bubble--align-start .hint-bubble__pop {
  left: 0;
  transform: none;
}

.hint-bubble--align-start .hint-bubble__pop::before {
  left: 11px;
  margin-left: 0;
}

.hint-bubble--align-end .hint-bubble__pop {
  left: auto;
  right: 0;
  transform: none;
}

.hint-bubble--align-end .hint-bubble__pop::before {
  left: auto;
  right: 11px;
  margin-left: 0;
}

.hint-bubble:hover .hint-bubble__pop,
.hint-bubble:focus-visible .hint-bubble__pop {
  opacity: 1;
  visibility: visible;
}

.custom-select {
  position: relative;
  width: 100%;
}

.custom-select__trigger {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  text-align: left;
  font-family: inherit;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
}

.custom-select__trigger--placeholder .custom-select__value {
  color: #cbd5e1;
  font-weight: 400;
}

.custom-select__value {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #0f172a;
  font-weight: 600;
}

.custom-select__trigger:disabled {
  cursor: not-allowed;
  background: #f8fafc;
  border-color: #e2e8f0;
  color: #334155;
}

.custom-select__trigger:disabled .custom-select__value {
  font-weight: 600;
  color: #0f172a;
}

.custom-select__trigger:disabled .custom-select__chevron {
  opacity: 0.35;
}

.custom-select__chevron {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  color: #94a3b8;
  transition: transform 0.2s ease;
}

.custom-select--open .custom-select__chevron {
  transform: rotate(180deg);
  color: #6366f1;
}

.custom-select--open .custom-select__trigger {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

.custom-select__menu {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  z-index: 40;
  margin: 0;
  padding: 4px;
  list-style: none;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  box-shadow: 0 12px 40px rgba(15, 23, 42, 0.12), 0 0 0 1px rgba(15, 23, 42, 0.04);
  max-height: 220px;
  overflow-y: auto;
}

.custom-select__item {
  padding: 10px 12px;
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #334155;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.12s ease, color 0.12s ease;
}

.custom-select__item:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.custom-select__item--active {
  background: #eef2ff;
  color: #4338ca;
  font-weight: 700;
}

.form-group--compact {
  gap: 4px;
  overflow: visible;
}

.form-group--compact .form-label {
  font-size: 12px;
  font-weight: 600;
}

.form-input--compact {
  padding: 10px 12px;
  border-radius: 8px;
  font-size: 13px;
}

.form-input--prefill-locked {
  cursor: default;
  font-weight: 600;
  color: #0f172a;
  background: #f8fafc;
  border-color: #e2e8f0;
}

.login-tabs {
  display: flex;
  background: #f1f5f9;
  border-radius: 10px;
  padding: 4px;
}

.tab {
  flex: 1;
  padding: 10px 16px;
  border: none;
  background: transparent;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s;
}

.tab.active {
  background: white;
  color: #0f172a;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.tab:hover:not(.active) {
  color: #64748b;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
  letter-spacing: 0.02em;
}

.req {
  color: #dc2626;
}

.form-input {
  padding: 12px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  transition: all 0.2s;
  background: white;
}

select.form-input {
  cursor: pointer;
}

.form-input:focus {
  outline: none;
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.form-input::placeholder {
  color: #cbd5e1;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 10px;
  color: #dc2626;
  font-size: 14px;
}

.error-message svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.login-btn {
  padding: 13px 24px;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  border: none;
  border-radius: 10px;
  color: white;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
}

.login-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.4);
}

.login-btn:active:not(:disabled) {
  transform: translateY(0);
}

.login-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-loading {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.login-mode-switch {
  margin-top: 16px;
  text-align: right;
}

.text-link {
  background: none;
  border: none;
  padding: 0;
  font-size: 14px;
  font-weight: 500;
  color: #2563eb;
  cursor: pointer;
  text-decoration: none;
}

.text-link:hover {
  text-decoration: underline;
}

.text-link-muted {
  color: #64748b;
}

.login-footer {
  margin-top: 20px;
  text-align: center;
}

.tip {
  font-size: 12px;
  color: #94a3b8;
  margin: 0;
}
</style>
