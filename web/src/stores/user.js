import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, ecoinApi } from '../api'
import { STORAGE_TOKEN_KEY, STORAGE_BINDINGS_KEY } from '../constants/storage.js'

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
  /** 企业账号：company_id > 0（与后端 user_tab.company_id 一致） */
  const isEnterpriseAccount = computed(() => {
    const id = user.value?.company_id
    return id != null && Number(id) > 0
  })
  const companyName = computed(() => user.value?.company_name || '')

  function setCachedBindingsList(list) {
    try {
      localStorage.setItem(STORAGE_BINDINGS_KEY, JSON.stringify(Array.isArray(list) ? list : []))
    } catch (_) {}
  }

  /** 已有登录态但本地无绑定缓存时补拉一次（如旧版本仅写了 token、或用户清过 storage） */
  async function hydrateBindingsIfNeeded() {
    const savedToken = localStorage.getItem(STORAGE_TOKEN_KEY)
    if (!savedToken) return
    try {
      const raw = localStorage.getItem(STORAGE_BINDINGS_KEY)
      if (raw != null && raw !== '') return
    } catch (_) {}
    await fetchBindingsRemote().catch(() => {})
  }

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
      if (Array.isArray(response.bindings)) {
        setCachedBindingsList(response.bindings)
      } else {
        await fetchBindingsRemote().catch(() => {})
      }

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
      if (Array.isArray(response.bindings)) {
        setCachedBindingsList(response.bindings)
      } else {
        await fetchBindingsRemote().catch(() => {})
      }

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
    localStorage.removeItem(STORAGE_BINDINGS_KEY)
  }

  /** 拉取并缓存当前用户的业务平台绑定（登录/绑定接口已带 bindings 时可不调用） */
  async function fetchBindingsRemote() {
    const list = await authApi.listUserBindings()
    setCachedBindingsList(list ?? [])
    return Array.isArray(list) ? list : []
  }

  /** 合并更新当前用户信息（个人中心修改手机号/邮箱后同步 Pinia 与 localStorage） */
  function patchUser(partial) {
    if (!user.value || !partial || typeof partial !== 'object') return
    user.value = { ...user.value, ...partial }
    try {
      localStorage.setItem('user', JSON.stringify(user.value))
    } catch (_) {}
  }

  /** 仅读 localStorage，路由守卫与 LawMind 跳转校验用此数据 */
  function getCachedBindings() {
    try {
      const raw = localStorage.getItem(STORAGE_BINDINGS_KEY)
      if (!raw) return []
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
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
    isEnterpriseAccount,
    companyName,
    init,
    login,
    bindAccount,
    logout,
    fetchEcoin,
    refreshEcoin,
    fetchBindingsRemote,
    getCachedBindings,
    hydrateBindingsIfNeeded,
    patchUser
  }
})
