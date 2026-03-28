<template>
  <div class="profile-page">
    <div class="container">
      <div class="page-header">
        <h1>个人中心</h1>
      </div>

      <!-- 账户信息：手机号、邮箱支持行内编辑 -->
      <div class="card info-card">
        <div class="section-header section-header--row">
          <h2>账户信息</h2>
          <div v-if="!accountEditing" class="header-actions">
            <button type="button" class="btn btn-secondary btn-sm" @click="startAccountEdit">编辑</button>
          </div>
          <div v-else class="header-actions">
            <button type="button" class="btn btn-secondary btn-sm" :disabled="contactSaving" @click="cancelAccountEdit">
              取消
            </button>
            <button type="button" class="btn btn-primary btn-sm" :disabled="contactSaving" @click="submitAccountEdit">
              {{ contactSaving ? '提交中…' : '提交' }}
            </button>
          </div>
        </div>
        <div v-if="contactError" class="contact-error" role="alert">{{ contactError }}</div>
        <dl class="info-grid">
          <div class="info-row">
            <dt>用户 ID</dt>
            <dd>{{ user?.id ?? '—' }}</dd>
          </div>
          <div class="info-row">
            <dt>手机号</dt>
            <dd v-if="!accountEditing" class="field-text">{{ user?.tel_no || '未设置' }}</dd>
            <dd v-else class="field-edit">
              <input v-model="contactDraft.tel_no" type="text" class="input" maxlength="11" placeholder="11 位手机号" />
            </dd>
          </div>
          <div class="info-row">
            <dt>邮箱</dt>
            <dd v-if="!accountEditing" class="field-text">{{ user?.email || '未设置' }}</dd>
            <dd v-else class="field-edit">
              <input v-model="contactDraft.email" type="email" class="input" maxlength="128" placeholder="邮箱" />
            </dd>
          </div>
          <div class="info-row">
            <dt>角色</dt>
            <dd>
              <span :class="['role-tag', user?.role === 1 ? 'role-admin' : 'role-user']">
                {{ user?.role === 1 ? '管理员' : '普通用户' }}
              </span>
            </dd>
          </div>
          <div class="info-row">
            <dt>注册时间</dt>
            <dd>{{ formatTime(user?.ctime) }}</dd>
          </div>
        </dl>
      </div>

      <!-- 业务平台绑定 -->
      <div class="card bindings-card">
        <div class="section-header section-header--row">
          <h2>业务平台绑定</h2>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="bindingsLoading" @click="loadBindings">
            {{ bindingsLoading ? '刷新中…' : '刷新列表' }}
          </button>
        </div>

        <div v-if="bindingsLoading && !bindings.length" class="loading-line">加载中…</div>

        <div v-else-if="!bindings.length" class="empty-state">
          <p>暂无业务平台绑定</p>
        </div>

        <div v-else class="table-wrap">
          <table class="bindings-table">
            <thead>
              <tr>
                <th>平台名称</th>
                <th>平台代码</th>
                <th>业务用户 ID</th>
                <th>绑定时间</th>
                <th class="col-action">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="b in bindings" :key="b.id ? String(b.id) : `${b.biz_code}-${b.biz_user_id}`">
                <td>{{ bizNameMap[b.biz_code] || b.biz_code }}</td>
                <td><code>{{ b.biz_code }}</code></td>
                <td>{{ b.biz_user_id }}</td>
                <td>{{ formatTime(b.ctime) }}</td>
                <td class="col-action">
                  <button
                    type="button"
                    class="unbind-btn"
                    :disabled="unbindingCode === b.biz_code"
                    @click="confirmUnbind(b)"
                  >
                    <span v-if="unbindingCode === b.biz_code" class="unbind-btn__spinner" aria-hidden="true" />
                    <svg
                      v-else
                      class="unbind-btn__icon"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      aria-hidden="true"
                    >
                      <path d="M18.84 12.25l1.72-1.71a4 4 0 0 0-5.66-5.66l-1.72 1.71" />
                      <path d="M5.16 12.75l-1.72 1.71a4 4 0 0 0 5.66 5.66l1.72-1.71" />
                      <line x1="2" y1="2" x2="22" y2="22" />
                    </svg>
                    <span>{{ unbindingCode === b.biz_code ? '处理中…' : '解除绑定' }}</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 修改密码 -->
      <div class="card password-card">
        <div class="section-header">
          <h2>修改密码</h2>
        </div>
        <p class="password-hint">修改后请使用新密码登录；当前会话仍然有效。</p>
        <form class="password-form" @submit.prevent="handleChangePassword">
          <div class="form-group">
            <label>当前密码</label>
            <PasswordInput
              v-model="pwdForm.oldSecret"
              placeholder="请输入当前登录密码"
              required
              autocomplete="current-password"
            />
          </div>
          <div class="form-group">
            <label>新密码</label>
            <PasswordInput
              v-model="pwdForm.newSecret"
              placeholder="请输入新密码"
              required
              autocomplete="new-password"
            />
          </div>
          <div class="form-group">
            <label>确认新密码</label>
            <PasswordInput
              v-model="pwdForm.newSecret2"
              placeholder="请再次输入新密码"
              required
              autocomplete="new-password"
            />
          </div>
          <div v-if="pwdError" class="form-error">{{ pwdError }}</div>
          <button type="submit" class="btn btn-primary" :disabled="pwdSubmitting">
            {{ pwdSubmitting ? '提交中…' : '确认修改' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import { authApi, metaApi } from '../api'
import { toast, confirm } from '../utils/toast'
import PasswordInput from '../components/PasswordInput.vue'

const userStore = useUserStore()

const user = computed(() => userStore.user)

const bindings = ref([])
const bindingsLoading = ref(false)
const bizNameMap = ref({})
const unbindingCode = ref('')

const accountEditing = ref(false)
const contactSaving = ref(false)
const contactDraft = reactive({ tel_no: '', email: '' })
const contactError = ref('')

const pwdForm = reactive({
  oldSecret: '',
  newSecret: '',
  newSecret2: ''
})
const pwdError = ref('')
const pwdSubmitting = ref(false)

function formatTime(ts) {
  if (ts === undefined || ts === null || ts === 0) return '—'
  const sec = Number(ts)
  if (!Number.isFinite(sec)) return '—'
  const d = new Date(sec * 1000)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

async function loadBizNames() {
  try {
    const list = await metaApi.listBizCodes()
    const map = {}
    for (const row of list || []) {
      if (row.code) map[row.code] = row.name || row.code
    }
    bizNameMap.value = map
  } catch {
    bizNameMap.value = {}
  }
}

function startAccountEdit() {
  contactError.value = ''
  contactDraft.tel_no = user.value?.tel_no || ''
  contactDraft.email = user.value?.email || ''
  accountEditing.value = true
}

function cancelAccountEdit() {
  accountEditing.value = false
  contactError.value = ''
}

function validateContact() {
  const tel = String(contactDraft.tel_no || '').trim()
  const email = String(contactDraft.email || '').trim()
  if (!tel && !email) {
    return '手机号与邮箱至少填写一项'
  }
  if (tel && !/^1[3-9]\d{9}$/.test(tel)) {
    return '请输入正确的手机号'
  }
  if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    return '请输入正确的邮箱地址'
  }
  return ''
}

async function submitAccountEdit() {
  contactError.value = ''
  const err = validateContact()
  if (err) {
    contactError.value = err
    return
  }
  contactSaving.value = true
  try {
    const data = await authApi.updateProfile({
      tel_no: String(contactDraft.tel_no || '').trim(),
      email: String(contactDraft.email || '').trim()
    })
    userStore.patchUser(data)
    accountEditing.value = false
    toast.success('已保存')
  } catch (e) {
    contactError.value = e?.message || '保存失败'
    toast.error(contactError.value)
  } finally {
    contactSaving.value = false
  }
}

async function loadBindings() {
  bindingsLoading.value = true
  try {
    const list = await userStore.fetchBindingsRemote()
    bindings.value = Array.isArray(list) ? list : []
  } catch (e) {
    const status = e?.response?.status
    if (status !== 401) {
      toast.error(e?.message || '加载绑定列表失败')
    }
    bindings.value = userStore.getCachedBindings()
  } finally {
    bindingsLoading.value = false
  }
}

async function confirmUnbind(row) {
  const ok = await confirm(`确定解除与「${bizNameMap.value[row.biz_code] || row.biz_code}」的绑定吗？`)
  if (!ok) return
  unbindingCode.value = row.biz_code
  try {
    await authApi.unbindBiz(row.biz_code)
    toast.success('已解除绑定')
    await loadBindings()
  } catch (e) {
    toast.error(e?.message || '解绑失败')
  } finally {
    unbindingCode.value = ''
  }
}

async function handleChangePassword() {
  pwdError.value = ''
  if (!pwdForm.oldSecret || !pwdForm.newSecret || !pwdForm.newSecret2) {
    pwdError.value = '请填写完整'
    return
  }
  if (pwdForm.newSecret !== pwdForm.newSecret2) {
    pwdError.value = '两次输入的新密码不一致'
    return
  }
  if (pwdForm.oldSecret === pwdForm.newSecret) {
    pwdError.value = '新密码不能与当前密码相同'
    return
  }

  pwdSubmitting.value = true
  try {
    await authApi.changePassword({
      old_secret: pwdForm.oldSecret,
      new_secret: pwdForm.newSecret
    })
    toast.success('密码已修改')
    pwdForm.oldSecret = ''
    pwdForm.newSecret = ''
    pwdForm.newSecret2 = ''
  } catch (e) {
    pwdError.value = e?.message || '修改失败'
  } finally {
    pwdSubmitting.value = false
  }
}

onMounted(() => {
  loadBizNames()
  loadBindings()
})
</script>

<style scoped>
.profile-page {
  padding-top: 8px;
  padding-bottom: 48px;
}

.info-card,
.bindings-card,
.password-card {
  padding: 24px 28px;
  margin-bottom: 20px;
}

.section-header {
  margin-bottom: 20px;
}

.section-header h2 {
  font-size: 18px;
  font-weight: 700;
  color: var(--gray-800);
  margin: 0;
}

.section-header--row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.btn-sm {
  padding: 8px 16px;
  font-size: 13px;
}

.contact-error {
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fef2f2;
  color: #b91c1c;
  font-size: 13px;
}

.info-grid {
  display: grid;
  gap: 0;
  margin: 0;
}

.info-row {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid var(--gray-100);
  align-items: center;
}

.info-row:last-child {
  border-bottom: none;
}

.info-row dt {
  font-size: 13px;
  font-weight: 600;
  color: var(--gray-500);
  margin: 0;
}

.info-row dd {
  margin: 0;
  font-size: 14px;
  color: var(--gray-800);
}

.field-text {
  font-weight: 500;
}

.field-edit .input {
  width: 100%;
  max-width: 360px;
  padding: 10px 12px;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  font-size: 14px;
}

.field-edit .input:focus {
  outline: none;
  border-color: var(--primary, #2563eb);
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.15);
}

.role-tag {
  display: inline-flex;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

.role-user {
  background: var(--gray-100);
  color: var(--gray-600);
}

.role-admin {
  background: #eef2ff;
  color: #4338ca;
}

.loading-line {
  padding: 24px;
  text-align: center;
  color: var(--gray-400);
  font-size: 14px;
}

.empty-state {
  padding: 32px;
  text-align: center;
  color: var(--gray-400);
  font-size: 14px;
}

.table-wrap {
  overflow-x: auto;
}

.bindings-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.bindings-table th,
.bindings-table td {
  padding: 12px 14px;
  text-align: left;
  border-bottom: 1px solid var(--gray-100);
}

.bindings-table th {
  font-weight: 600;
  color: var(--gray-500);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.bindings-table code {
  font-size: 13px;
  background: var(--gray-50);
  padding: 2px 6px;
  border-radius: 4px;
}

.col-action {
  width: 132px;
  white-space: nowrap;
  vertical-align: middle;
}

.unbind-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 7px 14px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: #b91c1c;
  background: transparent;
  border: 1px solid var(--gray-200, #e2e8f0);
  border-radius: 9px;
  box-shadow: none;
  cursor: pointer;
  transition:
    background 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    color 0.18s ease,
    transform 0.12s ease;
}

.unbind-btn:hover:not(:disabled) {
  color: #991b1b;
  background: transparent;
  border-color: var(--gray-300, #cbd5e1);
}

.unbind-btn:active:not(:disabled) {
  transform: translateY(1px);
}

.unbind-btn:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.45);
}

.unbind-btn:disabled {
  opacity: 0.72;
  cursor: not-allowed;
  transform: none;
}

.unbind-btn__icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  color: inherit;
}

.unbind-btn__spinner {
  width: 13px;
  height: 13px;
  flex-shrink: 0;
  border: 2px solid rgba(185, 28, 28, 0.2);
  border-top-color: #b91c1c;
  border-radius: 50%;
  animation: unbind-spin 0.55s linear infinite;
}

@keyframes unbind-spin {
  to {
    transform: rotate(360deg);
  }
}

.password-hint {
  font-size: 13px;
  color: var(--gray-500);
  margin: -8px 0 20px;
}

.password-form {
  max-width: 420px;
}

.form-group {
  margin-bottom: 18px;
}

.form-group label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--gray-600);
  margin-bottom: 8px;
}

.form-error {
  font-size: 13px;
  color: var(--danger);
  margin-bottom: 12px;
}

@media (max-width: 640px) {
  .info-row {
    grid-template-columns: 1fr;
    gap: 4px;
  }
}
</style>
