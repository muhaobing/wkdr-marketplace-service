<template>
  <div class="login-page">
    <div class="login-bg"></div>
    <div class="login-container">
      <div class="login-card">
        <div class="login-header">
          <div class="logo-mark">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
              <polyline points="9 22 9 12 15 12 15 22"/>
            </svg>
          </div>
          <h1 class="title">商城中心</h1>
          <p class="subtitle">登录您的账户以继续</p>
        </div>
        
        <form @submit.prevent="handleLogin" class="login-form">
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
          
          <button
            type="submit"
            class="login-btn"
            :disabled="loading"
          >
            <span v-if="loading" class="btn-loading"></span>
            <span v-else>登 录</span>
          </button>
        </form>
        
        <div class="login-footer">
          <p class="tip">请使用您的账户密钥进行登录</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loginType = ref('email')
const loading = ref(false)
const errorMsg = ref('')

const formData = reactive({
  email: '',
  tel_no: '',
  secret: ''
})

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

.login-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.3);
  padding: 44px 40px;
  border: 1px solid rgba(255, 255, 255, 0.2);
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

.form-input {
  padding: 12px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  transition: all 0.2s;
  background: white;
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

.login-footer {
  margin-top: 24px;
  text-align: center;
}

.tip {
  font-size: 12px;
  color: #94a3b8;
  margin: 0;
}
</style>
