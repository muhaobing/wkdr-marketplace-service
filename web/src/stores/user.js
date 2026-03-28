import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, ecoinApi } from '../api'
import { STORAGE_TOKEN_KEY } from '../constants/storage.js'

export const useUserStore = defineStore('user', () => {
  // 用户数据
  const user = ref(null)
  const token = ref(null)
  const ecoin = ref(null)
  const loading = ref(false)

  // 计算属性
  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const userId = computed(() => user.value?.id || 0)
  const userName = computed(() => user.value?.name || user.value?.email || user.value?.tel_no || '')
  const balance = computed(() => ecoin.value?.available_stock || 0)
  const isAdmin = computed(() => user.value?.role === 1)

  // 初始化 - 从 localStorage 恢复登录状态
  function init() {
    const savedToken = localStorage.getItem(STORAGE_TOKEN_KEY)
    const savedUser = localStorage.getItem('user')
    if (savedToken && savedUser) {
      token.value = savedToken
      try {
        user.value = JSON.parse(savedUser)
      } catch (e) {
        console.error('解析用户信息失败:', e)
        logout()
      }
    }
  }

  // 登录
  async function login(credentials) {
    loading.value = true
    try {
      const response = await authApi.login(credentials)
      token.value = response.token
      user.value = response.user

      localStorage.setItem(STORAGE_TOKEN_KEY, response.token)
      localStorage.setItem('user', JSON.stringify(response.user))

      await fetchEcoin()

      return response
    } catch (error) {
      console.error('登录失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 绑定（返回 token，与登录后状态一致）
  async function bindAccount(payload) {
    loading.value = true
    try {
      const response = await authApi.bind(payload)
      token.value = response.token
      user.value = response.user

      localStorage.setItem(STORAGE_TOKEN_KEY, response.token)
      localStorage.setItem('user', JSON.stringify(response.user))

      await fetchEcoin()

      return response
    } catch (error) {
      console.error('绑定失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 登出
  function logout() {
    token.value = null
    user.value = null
    ecoin.value = null
    localStorage.removeItem(STORAGE_TOKEN_KEY)
    localStorage.removeItem('user')
  }

  // 获取用户积分
  async function fetchEcoin() {
    if (!user.value?.id) return
    if (loading.value) return
    
    loading.value = true
    try {
      ecoin.value = await ecoinApi.getBalance(user.value.id)
    } catch (error) {
      console.error('获取积分失败:', error)
      ecoin.value = { available_stock: 0 }
    } finally {
      loading.value = false
    }
  }

  // 刷新积分
  function refreshEcoin() {
    fetchEcoin()
  }

  // 初始化
  init()

  return {
    user,
    token,
    ecoin,
    loading,
    isLoggedIn,
    isAdmin,
    userId,
    userName,
    balance,
    init,
    login,
    bindAccount,
    logout,
    fetchEcoin,
    refreshEcoin
  }
})
