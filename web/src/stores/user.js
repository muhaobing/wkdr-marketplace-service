import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ecoinApi } from '../api'

export const useUserStore = defineStore('user', () => {
  // Mock 用户数据
  const user = ref({
    id: 10001,
    name: '测试用户'
  })

  const ecoin = ref(null)
  const loading = ref(false)

  const userId = computed(() => user.value.id)
  const userName = computed(() => user.value.name)
  const balance = computed(() => ecoin.value?.balance || 0)

  // 获取用户积分
  async function fetchEcoin() {
    if (loading.value) return
    loading.value = true
    try {
      ecoin.value = await ecoinApi.get(user.value.id)
    } catch (error) {
      console.error('获取积分失败:', error)
      // Mock 数据
      ecoin.value = { balance: 1000 }
    } finally {
      loading.value = false
    }
  }

  // 刷新积分
  function refreshEcoin() {
    fetchEcoin()
  }

  return {
    user,
    ecoin,
    loading,
    userId,
    userName,
    balance,
    fetchEcoin,
    refreshEcoin
  }
})
