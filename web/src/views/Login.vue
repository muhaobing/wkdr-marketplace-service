<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-header">
        <h1 class="title">商城中心</h1>
        <p class="subtitle">欢迎登录</p>
      </div>
      
      <form @submit.prevent="handleLogin" class="login-form">
        <!-- 登录方式切换 -->
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
        
        <!-- 邮箱输入 -->
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
        
        <!-- 手机号输入 -->
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
        
        <!-- 密钥输入 -->
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
        
        <!-- 错误提示 -->
        <div v-if="errorMsg" class="error-message">
          {{ errorMsg }}
        </div>
        
        <!-- 登录按钮 -->
        <button
          type="submit"
          class="login-btn"
          :disabled="loading"
        >
          <span v-if="loading">登录中...</span>
          <span v-else>登 录</span>
        </button>
      </form>
      
      <div class="login-footer">
        <p class="tip">提示：请使用您的账户密钥进行登录</p>
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
  
  // 验证表单
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
    
    // 登录成功，跳转到目标页面或首页
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
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  padding: 20px;
}

.login-container {
  width: 100%;
  max-width: 420px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  padding: 40px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.title {
  font-size: 28px;
  font-weight: 700;
  color: #1a365d;
  margin: 0 0 8px 0;
}

.subtitle {
  font-size: 14px;
  color: #718096;
  margin: 0;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.login-tabs {
  display: flex;
  background: #f7fafc;
  border-radius: 8px;
  padding: 4px;
}

.tab {
  flex: 1;
  padding: 10px 16px;
  border: none;
  background: transparent;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #718096;
  cursor: pointer;
  transition: all 0.2s;
}

.tab.active {
  background: white;
  color: #1a365d;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.tab:hover:not(.active) {
  color: #4a5568;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 14px;
  font-weight: 500;
  color: #4a5568;
}

.form-input {
  padding: 12px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #1a365d;
  box-shadow: 0 0 0 3px rgba(26, 54, 93, 0.1);
}

.form-input::placeholder {
  color: #a0aec0;
}

.error-message {
  padding: 12px 16px;
  background: #fff5f5;
  border: 1px solid #fed7d7;
  border-radius: 8px;
  color: #c53030;
  font-size: 14px;
}

.login-btn {
  padding: 14px 24px;
  background: #1a365d;
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.login-btn:hover:not(:disabled) {
  background: #2d3748;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(26, 54, 93, 0.3);
}

.login-btn:disabled {
  background: #a0aec0;
  cursor: not-allowed;
}

.login-footer {
  margin-top: 24px;
  text-align: center;
}

.tip {
  font-size: 12px;
  color: #a0aec0;
  margin: 0;
}
</style>
