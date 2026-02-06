import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { cartApi } from '../api/index.js'
import { useUserStore } from './user.js'

export const useCartStore = defineStore('cart', () => {
  // 购物车商品列表
  const items = ref([])
  
  // 加载状态
  const loading = ref(false)

  // 选中状态（保存在前端）
  const selectedMap = ref({})

  // 购物车商品数量
  const count = computed(() => {
    return items.value.reduce((sum, item) => sum + item.quantity, 0)
  })

  // 购物车总价
  const totalPrice = computed(() => {
    return items.value.reduce((sum, item) => sum + item.cost * item.quantity, 0)
  })

  // 为 items 添加 selected 属性
  const itemsWithSelected = computed(() => {
    return items.value.map(item => ({
      ...item,
      selected: selectedMap.value[item.sku_id] !== false // 默认选中
    }))
  })

  // 选中的商品
  const selectedItems = computed(() => {
    return itemsWithSelected.value.filter(item => item.selected)
  })

  // 选中商品的总价
  const selectedTotalPrice = computed(() => {
    return selectedItems.value.reduce((sum, item) => sum + item.cost * item.quantity, 0)
  })

  // 是否全选
  const isAllSelected = computed(() => {
    return items.value.length > 0 && items.value.every(item => selectedMap.value[item.sku_id] !== false)
  })

  // 获取当前用户 ID
  function getUserId() {
    const userStore = useUserStore()
    return userStore.user?.id
  }

  // 从后端获取购物车列表
  async function fetchCart() {
    const userId = getUserId()
    if (!userId) return

    loading.value = true
    try {
      const response = await cartApi.list({ user_id: userId })
      items.value = response?.items || []
      // 同步选中状态：对于新加入的商品默认选中
      items.value.forEach(item => {
        if (selectedMap.value[item.sku_id] === undefined) {
          selectedMap.value[item.sku_id] = true
        }
      })
      saveSelectedToStorage()
    } catch (error) {
      console.error('Failed to fetch cart:', error)
    } finally {
      loading.value = false
    }
  }

  // 添加商品到购物车
  async function addItem(sku, quantity = 1) {
    const userId = getUserId()
    if (!userId) return

    try {
      await cartApi.add({
        user_id: userId,
        sku_id: sku.id,
        quantity
      })
      // 默认选中新加入的商品
      selectedMap.value[sku.id] = true
      saveSelectedToStorage()
      await fetchCart()
    } catch (error) {
      console.error('Failed to add to cart:', error)
      throw error
    }
  }

  // 更新商品数量
  async function updateQuantity(skuId, quantity) {
    const userId = getUserId()
    if (!userId) return

    try {
      if (quantity <= 0) {
        await cartApi.remove({
          user_id: userId,
          sku_id: skuId
        })
        delete selectedMap.value[skuId]
      } else {
        await cartApi.update({
          user_id: userId,
          sku_id: skuId,
          quantity
        })
      }
      saveSelectedToStorage()
      await fetchCart()
    } catch (error) {
      console.error('Failed to update cart:', error)
      throw error
    }
  }

  // 移除商品
  async function removeItem(skuId) {
    const userId = getUserId()
    if (!userId) return

    try {
      await cartApi.remove({
        user_id: userId,
        sku_id: skuId
      })
      delete selectedMap.value[skuId]
      saveSelectedToStorage()
      await fetchCart()
    } catch (error) {
      console.error('Failed to remove from cart:', error)
      throw error
    }
  }

  // 切换选中状态（前端操作）
  function toggleSelect(skuId) {
    selectedMap.value[skuId] = !selectedMap.value[skuId]
    saveSelectedToStorage()
  }

  // 全选/取消全选（前端操作）
  function toggleSelectAll() {
    const newValue = !isAllSelected.value
    items.value.forEach(item => {
      selectedMap.value[item.sku_id] = newValue
    })
    saveSelectedToStorage()
  }

  // 清空选中的商品
  async function clearSelected() {
    const userId = getUserId()
    if (!userId) return

    const skuIdsToRemove = selectedItems.value.map(item => item.sku_id)
    
    // 依次删除选中的商品
    for (const skuId of skuIdsToRemove) {
      try {
        await cartApi.remove({
          user_id: userId,
          sku_id: skuId
        })
        delete selectedMap.value[skuId]
      } catch (error) {
        console.error('Failed to remove item:', error)
      }
    }
    saveSelectedToStorage()
    await fetchCart()
  }

  // 清空购物车
  async function clearCart() {
    const userId = getUserId()
    if (!userId) return

    try {
      await cartApi.clear({ user_id: userId })
      items.value = []
      selectedMap.value = {}
      saveSelectedToStorage()
    } catch (error) {
      console.error('Failed to clear cart:', error)
      throw error
    }
  }

  // 购物车下单
  async function checkout(payType, remark = '') {
    const userId = getUserId()
    if (!userId) return

    const skuIds = selectedItems.value.map(item => item.sku_id)
    if (skuIds.length === 0) {
      throw new Error('请选择要下单的商品')
    }

    try {
      const response = await cartApi.checkout({
        user_id: userId,
        sku_ids: skuIds,
        pay_type: payType,
        remark
      })
      // 下单成功后清除已下单商品的选中状态
      skuIds.forEach(skuId => {
        delete selectedMap.value[skuId]
      })
      saveSelectedToStorage()
      await fetchCart()
      return response
    } catch (error) {
      console.error('Failed to checkout:', error)
      throw error
    }
  }

  // 保存选中状态到本地存储
  function saveSelectedToStorage() {
    localStorage.setItem('cart_selected', JSON.stringify(selectedMap.value))
  }

  // 从本地存储加载选中状态
  function loadSelectedFromStorage() {
    const stored = localStorage.getItem('cart_selected')
    if (stored) {
      try {
        selectedMap.value = JSON.parse(stored)
      } catch (e) {
        selectedMap.value = {}
      }
    }
  }

  // 初始化
  function init() {
    loadSelectedFromStorage()
    fetchCart()
  }

  return {
    items,
    itemsWithSelected,
    count,
    totalPrice,
    selectedItems,
    selectedTotalPrice,
    isAllSelected,
    loading,
    addItem,
    updateQuantity,
    removeItem,
    toggleSelect,
    toggleSelectAll,
    clearSelected,
    clearCart,
    checkout,
    fetchCart,
    init
  }
})
