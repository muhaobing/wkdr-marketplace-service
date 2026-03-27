<template>
  <div class="cart-page">
    <div class="container">
      <div class="page-header">
        <h1>购物车</h1>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <p>加载中...</p>
      </div>

      <div v-else-if="cartItems.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="9" cy="21" r="1"/>
          <circle cx="20" cy="21" r="1"/>
          <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
        </svg>
        <p>购物车是空的</p>
        <router-link to="/" class="btn btn-primary" style="margin-top: 16px;">去逛逛</router-link>
      </div>

      <div v-else class="cart-content">
        <div class="cart-main card">
          <!-- 全选 -->
          <div class="cart-header">
            <label class="select-all">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll">
              <span>全选</span>
            </label>
            <span class="header-product">商品信息</span>
            <span class="header-price">单价</span>
            <span class="header-quantity">数量</span>
            <span class="header-total">小计</span>
            <span class="header-action">操作</span>
          </div>

          <!-- 商品列表 -->
          <div class="cart-list">
            <div v-for="item in cartItems" :key="item.sku_id" class="cart-item" :class="{ 'offline': item.sku_status !== 1 }">
              <label class="item-checkbox">
                <input type="checkbox" :checked="item.selected" @change="toggleSelect(item.sku_id)" :disabled="item.sku_status !== 1">
              </label>
              
              <div class="item-product" @click="goToDetail(item.sku_id)">
                <div class="product-image">
                  <img v-if="item.sku_avatar" :src="item.sku_avatar" :alt="item.sku_name">
                  <div v-else class="image-placeholder">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                      <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                    </svg>
                  </div>
                </div>
                <div class="product-info">
                  <h4>{{ item.sku_name }}</h4>
                  <p>{{ item.sku_code }}</p>
                  <span v-if="item.sku_status !== 1" class="offline-badge">已下架</span>
                </div>
              </div>

              <div class="item-price">
                ¥{{ item.cost.toFixed(2) }}
                <div class="ecoin-price">({{ toEcoin(item.cost) }} 积分)</div>
              </div>

              <div class="item-quantity">
                <div class="quantity-control">
                  <button @click="decreaseQty(item)" :disabled="item.quantity <= 1">-</button>
                  <input type="number" :value="item.quantity" @change="updateQty(item, $event)" min="1" max="99">
                  <button @click="increaseQty(item)" :disabled="item.quantity >= 99">+</button>
                </div>
              </div>

              <div class="item-total">
                ¥{{ (item.cost * item.quantity).toFixed(2) }}
                <div class="ecoin-price">({{ toEcoin(item.cost * item.quantity) }} 积分)</div>
              </div>

              <div class="item-action">
                <button class="delete-btn" @click="removeItem(item.sku_id)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"/>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 结算栏 -->
        <div class="cart-footer card">
          <div class="footer-left">
            <label class="select-all">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll">
              <span>全选</span>
            </label>
            <button class="clear-btn" @click="clearSelected" :disabled="selectedItems.length === 0">
              删除选中
            </button>
          </div>
          <div class="footer-right">
            <div class="summary">
              <span>已选 <strong>{{ selectedItems.length }}</strong> 件商品</span>
              <span class="total-price">
                合计: <strong>¥{{ selectedTotalPrice.toFixed(2) }}</strong>
                <span class="ecoin-price">({{ toEcoin(selectedTotalPrice) }} 积分)</span>
              </span>
            </div>
            <button 
              class="btn btn-primary checkout-btn" 
              :disabled="selectedItems.length === 0 || ordering"
              @click="handleCheckout"
            >
              {{ ordering ? '下单中...' : '结算' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { useUserStore } from '../stores/user'
import { ecoinApi } from '../api'
import { toast, confirm } from '../utils/toast'

const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()
const ecoinUnitPrice = ref(0)
const ordering = ref(false)

function toEcoin(cost) {
  if (!ecoinUnitPrice.value || ecoinUnitPrice.value <= 0) return '--'
  return (cost / ecoinUnitPrice.value).toFixed(2)
}

const cartItems = computed(() => cartStore.itemsWithSelected)
const selectedItems = computed(() => cartStore.selectedItems)
const selectedTotalPrice = computed(() => cartStore.selectedTotalPrice)
const isAllSelected = computed(() => cartStore.isAllSelected)
const loading = computed(() => cartStore.loading)

onMounted(async () => {
  cartStore.init()
  try {
    const cfg = await ecoinApi.getRechargeConfig()
    ecoinUnitPrice.value = cfg.unit_price || 0
  } catch (e) { /* ignore */ }
})

function toggleSelect(skuId) {
  cartStore.toggleSelect(skuId)
}

function toggleSelectAll() {
  cartStore.toggleSelectAll()
}

async function decreaseQty(item) {
  if (item.quantity > 1) {
    await cartStore.updateQuantity(item.sku_id, item.quantity - 1)
  }
}

async function increaseQty(item) {
  if (item.quantity < 99) {
    await cartStore.updateQuantity(item.sku_id, item.quantity + 1)
  }
}

async function updateQty(item, event) {
  const value = parseInt(event.target.value)
  if (value >= 1 && value <= 99) {
    await cartStore.updateQuantity(item.sku_id, value)
  }
}

async function removeItem(skuId) {
  if (await confirm('确定要删除这个商品吗？')) {
    await cartStore.removeItem(skuId)
  }
}

async function clearSelected() {
  if (await confirm('确定要删除选中的商品吗？')) {
    await cartStore.clearSelected()
  }
}

function goToDetail(skuId) {
  router.push(`/product/${skuId}`)
}

async function handleCheckout() {
  if (selectedItems.value.length === 0 || ordering.value) return
  ordering.value = true
  try {
    const orderRes = await cartStore.checkout()
    const orderNo = orderRes.order?.order_no || orderRes.order_no
    router.push(`/orders/${orderNo}`)
  } catch (error) {
    toast.error('下单失败: ' + error.message)
  } finally {
    ordering.value = false
  }
}
</script>

<style scoped>
.cart-page {
  padding-top: 8px;
}

.loading-state {
  text-align: center;
  padding: 60px 0;
  color: var(--gray-500);
}

.cart-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.cart-main {
  overflow: hidden;
}

.cart-header {
  display: grid;
  grid-template-columns: 50px 2fr 1fr 120px 1fr 80px;
  gap: 16px;
  align-items: center;
  padding: 14px 24px;
  background-color: var(--gray-50);
  font-size: 13px;
  color: var(--gray-400);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-bottom: 1px solid var(--gray-100);
}

.select-all {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  text-transform: none;
  letter-spacing: normal;
  font-weight: 500;
  color: var(--gray-600);
}

.select-all input {
  width: 18px;
  height: 18px;
  accent-color: var(--accent);
}

.cart-list {
  padding: 0 24px;
}

.cart-item {
  display: grid;
  grid-template-columns: 50px 2fr 1fr 120px 1fr 80px;
  gap: 16px;
  align-items: center;
  padding: 20px 0;
  border-bottom: 1px solid var(--gray-100);
}

.cart-item:last-child {
  border-bottom: none;
}

.cart-item.offline {
  opacity: 0.5;
}

.item-checkbox input {
  width: 18px;
  height: 18px;
  accent-color: var(--accent);
}

.item-product {
  display: flex;
  gap: 16px;
  cursor: pointer;
}

.product-image {
  width: 72px;
  height: 72px;
  border-radius: 10px;
  overflow: hidden;
  background: linear-gradient(135deg, #f8fafc, #e2e8f0);
  flex-shrink: 0;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--gray-300);
}

.image-placeholder svg {
  width: 28px;
  height: 28px;
}

.product-info h4 {
  font-size: 14px;
  font-weight: 600;
  color: var(--gray-800);
  margin-bottom: 4px;
}

.product-info p {
  font-size: 12px;
  color: var(--gray-400);
}

.offline-badge {
  display: inline-block;
  margin-top: 4px;
  padding: 2px 8px;
  font-size: 11px;
  color: var(--danger);
  background-color: #fef2f2;
  border-radius: 4px;
  font-weight: 500;
}

.item-price {
  font-size: 14px;
  color: #b91c1c;
  font-weight: 700;
}

.ecoin-price {
  font-size: 12px;
  color: #f59e0b;
  font-weight: 600;
  margin-top: 2px;
}

.quantity-control {
  display: flex;
  align-items: center;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  overflow: hidden;
}

.quantity-control button {
  width: 30px;
  height: 30px;
  background-color: var(--gray-50);
  color: var(--gray-600);
  font-size: 14px;
  transition: background-color 0.15s;
}

.quantity-control button:hover:not(:disabled) {
  background-color: var(--gray-100);
}

.quantity-control button:disabled {
  color: var(--gray-300);
  cursor: not-allowed;
}

.quantity-control input {
  width: 40px;
  height: 30px;
  text-align: center;
  border: none;
  border-left: 1px solid var(--gray-200);
  border-right: 1px solid var(--gray-200);
  font-size: 13px;
  font-weight: 600;
}

.quantity-control input::-webkit-outer-spin-button,
.quantity-control input::-webkit-inner-spin-button {
  -webkit-appearance: none;
}

.item-total {
  font-size: 15px;
  font-weight: 700;
  color: #b91c1c;
}

.delete-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--gray-400);
  background: none;
  border-radius: 8px;
  transition: all 0.15s;
}

.delete-btn:hover {
  color: var(--danger);
  background-color: #fef2f2;
}

.delete-btn svg {
  width: 18px;
  height: 18px;
}

.cart-footer {
  position: sticky;
  bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px 28px;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 24px;
}

.clear-btn {
  color: var(--gray-400);
  background: none;
  font-size: 13px;
  font-weight: 500;
  transition: color 0.15s;
}

.clear-btn:hover:not(:disabled) {
  color: var(--danger);
}

.clear-btn:disabled {
  color: var(--gray-300);
  cursor: not-allowed;
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 32px;
}

.summary {
  display: flex;
  align-items: center;
  gap: 24px;
  font-size: 14px;
  color: var(--gray-500);
}

.summary strong {
  color: var(--gray-800);
}

.total-price strong {
  font-size: 22px;
  color: #b91c1c;
  letter-spacing: -0.02em;
}

.checkout-btn {
  padding: 12px 48px;
  font-size: 15px;
  font-weight: 600;
}

@media (max-width: 768px) {
  .cart-header {
    display: none;
  }

  .cart-item {
    grid-template-columns: 40px 1fr;
    gap: 12px;
  }

  .item-product {
    grid-column: 2;
  }

  .item-price,
  .item-quantity,
  .item-total {
    display: none;
  }

  .item-action {
    position: absolute;
    right: 20px;
  }

  .cart-footer {
    flex-direction: column;
    gap: 16px;
  }

  .footer-right {
    width: 100%;
    justify-content: space-between;
  }

  .checkout-btn {
    padding: 12px 24px;
  }
}
</style>
