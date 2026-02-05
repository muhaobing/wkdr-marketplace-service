import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useCartStore = defineStore('cart', () => {
  // 购物车商品列表
  const items = ref([])

  // 购物车商品数量
  const count = computed(() => {
    return items.value.reduce((sum, item) => sum + item.quantity, 0)
  })

  // 购物车总价
  const totalPrice = computed(() => {
    return items.value.reduce((sum, item) => sum + item.cost * item.quantity, 0)
  })

  // 选中的商品
  const selectedItems = computed(() => {
    return items.value.filter(item => item.selected)
  })

  // 选中商品的总价
  const selectedTotalPrice = computed(() => {
    return selectedItems.value.reduce((sum, item) => sum + item.cost * item.quantity, 0)
  })

  // 是否全选
  const isAllSelected = computed(() => {
    return items.value.length > 0 && items.value.every(item => item.selected)
  })

  // 添加商品到购物车
  function addItem(sku, quantity = 1) {
    const existingItem = items.value.find(item => item.id === sku.id)
    if (existingItem) {
      existingItem.quantity += quantity
    } else {
      items.value.push({
        id: sku.id,
        sku_code: sku.sku_code,
        sku_name: sku.sku_name,
        sku_avatar: sku.sku_avatar,
        cost: sku.cost,
        quantity,
        selected: true
      })
    }
    saveToStorage()
  }

  // 更新商品数量
  function updateQuantity(skuId, quantity) {
    const item = items.value.find(item => item.id === skuId)
    if (item) {
      item.quantity = Math.max(1, quantity)
      saveToStorage()
    }
  }

  // 移除商品
  function removeItem(skuId) {
    const index = items.value.findIndex(item => item.id === skuId)
    if (index > -1) {
      items.value.splice(index, 1)
      saveToStorage()
    }
  }

  // 切换选中状态
  function toggleSelect(skuId) {
    const item = items.value.find(item => item.id === skuId)
    if (item) {
      item.selected = !item.selected
      saveToStorage()
    }
  }

  // 全选/取消全选
  function toggleSelectAll() {
    const newValue = !isAllSelected.value
    items.value.forEach(item => {
      item.selected = newValue
    })
    saveToStorage()
  }

  // 清空选中的商品
  function clearSelected() {
    items.value = items.value.filter(item => !item.selected)
    saveToStorage()
  }

  // 清空购物车
  function clearCart() {
    items.value = []
    saveToStorage()
  }

  // 保存到本地存储
  function saveToStorage() {
    localStorage.setItem('cart', JSON.stringify(items.value))
  }

  // 从本地存储加载
  function loadFromStorage() {
    const stored = localStorage.getItem('cart')
    if (stored) {
      try {
        items.value = JSON.parse(stored)
      } catch (e) {
        items.value = []
      }
    }
  }

  // 初始化时加载
  loadFromStorage()

  return {
    items,
    count,
    totalPrice,
    selectedItems,
    selectedTotalPrice,
    isAllSelected,
    addItem,
    updateQuantity,
    removeItem,
    toggleSelect,
    toggleSelectAll,
    clearSelected,
    clearCart,
    loadFromStorage
  }
})
