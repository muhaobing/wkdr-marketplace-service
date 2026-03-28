<template>
  <div class="profile-page">
    <div class="container">
      <div class="page-header">
        <h1>个人中心</h1>
      </div>

      <!-- 账户信息 -->
      <div class="card info-card">
        <div class="section-header">
          <h2>账户信息</h2>
        </div>
        <dl class="info-grid">
          <div class="info-row">
            <dt>用户 ID</dt>
            <dd>{{ user?.id ?? '—' }}</dd>
          </div>
          <div class="info-row">
            <dt>手机号</dt>
            <dd>{{ user?.tel_no || '未设置' }}</dd>
          </div>
          <div class="info-row">
            <dt>邮箱</dt>
            <dd>{{ user?.email || '未设置' }}</dd>
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
              </tr>
            </thead>
            <tbody>
              <tr v-for="b in bindings" :key="b.id ? String(b.id) : `${b.biz_code}-${b.biz_user_id}`">
                <td>{{ bizNameMap[b.biz_code] || b.biz_code }}</td>
                <td><code>{{ b.biz_code }}</code></td>
                <td>{{ b.biz_user_id }}</td>
                <td>{{ formatTime(b.ctime) }}</td>
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
import { toast } from '../utils/toast'
import PasswordInput from '../components/PasswordInput.vue'

const userStore = useUserStore()

const user = computed(() => userStore.user)

const bindings = ref([])
const bindingsLoading = ref(false)
const bizNameMap = ref({})

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

async function loadBindings() {
  bindingsLoading.value = true
  try {
    const list = await userStore.fetchBindingsRemote()
    bindings.value = Array.isArray(list) ? list : []
  } catch (e) {
    const status = e?.response?.status
    // 401：全局拦截器已清 token 并跳转登录，勿再 toast「Unauthorized」造成误解
    if (status !== 401) {
      toast.error(e?.message || '加载绑定列表失败')
    }
    bindings.value = userStore.getCachedBindings()
  } finally {
    bindingsLoading.value = false
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

.btn-sm {
  padding: 8px 16px;
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
