<template>
  <div class="cart-page">
    <div class="container">
      <div class="page-header">
        <h1>购物车</h1>
      </div>

      <div v-if="cartItems.length === 0" class="empty-state">
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
            <div v-for="item in cartItems" :key="item.id" class="cart-item">
              <label class="item-checkbox">
                <input type="checkbox" :checked="item.selected" @change="toggleSelect(item.id)">
              </label>
              
              <div class="item-product" @click="goToDetail(item.id)">
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
                </div>
              </div>

              <div class="item-price">
                {{ item.cost.toFixed(2) }} 积分
              </div>

              <div class="item-quantity">
                <div class="quantity-control">
                  <button @click="decreaseQty(item)" :disabled="item.quantity <= 1">-</button>
                  <input type="number" :value="item.quantity" @change="updateQty(item, $event)" min="1" max="99">
                  <button @click="increaseQty(item)" :disabled="item.quantity >= 99">+</button>
                </div>
              </div>

              <div class="item-total">
                {{ (item.cost * item.quantity).toFixed(2) }} 积分
              </div>

              <div class="item-action">
                <button class="delete-btn" @click="removeItem(item.id)">
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
                合计: <strong>{{ selectedTotalPrice.toFixed(2) }}</strong> 积分
              </span>
            </div>
            <button 
              class="btn btn-primary checkout-btn" 
              :disabled="selectedItems.length === 0"
              @click="checkout"
            >
              结算
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 支付方式弹窗 -->
    <div v-if="showPaymentModal" class="modal-overlay" @click.self="closePaymentModal">
      <div class="modal-content card">
        <h3>选择支付方式</h3>
        <div class="payment-methods">
          <label 
            v-for="method in paymentMethods" 
            :key="method.channel + method.pay_method"
            class="payment-option"
            :class="{ active: selectedPayment === method }"
          >
            <input type="radio" :value="method" v-model="selectedPayment">
            <span class="option-name">{{ method.name }}</span>
          </label>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closePaymentModal">取消</button>
          <button class="btn btn-primary" @click="confirmOrder" :disabled="!selectedPayment || ordering">
            {{ ordering ? '处理中...' : '确认下单' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { useUserStore } from '../stores/user'
import { orderApi, paymentApi } from '../api'

const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()

const cartItems = computed(() => cartStore.items)
const selectedItems = computed(() => cartStore.selectedItems)
const selectedTotalPrice = computed(() => cartStore.selectedTotalPrice)
const isAllSelected = computed(() => cartStore.isAllSelected)

const showPaymentModal = ref(false)
const paymentMethods = ref([])
const selectedPayment = ref(null)
const ordering = ref(false)

function toggleSelect(id) {
  cartStore.toggleSelect(id)
}

function toggleSelectAll() {
  cartStore.toggleSelectAll()
}

function decreaseQty(item) {
  if (item.quantity > 1) {
    cartStore.updateQuantity(item.id, item.quantity - 1)
  }
}

function increaseQty(item) {
  if (item.quantity < 99) {
    cartStore.updateQuantity(item.id, item.quantity + 1)
  }
}

function updateQty(item, event) {
  const value = parseInt(event.target.value)
  if (value >= 1 && value <= 99) {
    cartStore.updateQuantity(item.id, value)
  }
}

function removeItem(id) {
  if (confirm('确定要删除这个商品吗？')) {
    cartStore.removeItem(id)
  }
}

function clearSelected() {
  if (confirm('确定要删除选中的商品吗？')) {
    cartStore.clearSelected()
  }
}

function goToDetail(id) {
  router.push(`/product/${id}`)
}

async function fetchPaymentMethods() {
  try {
    paymentMethods.value = await paymentApi.methods()
  } catch (error) {
    paymentMethods.value = [
      { channel: 'ecoin', name: '积分支付', pay_method: 'ecoin' },
      { channel: 'wechat', name: '微信扫码支付', pay_method: 'native' }
    ]
  }
}

function checkout() {
  if (selectedItems.value.length === 0) return
  showPaymentModal.value = true
  fetchPaymentMethods()
}

function closePaymentModal() {
  showPaymentModal.value = false
  selectedPayment.value = null
}

async function confirmOrder() {
  if (!selectedPayment.value || ordering.value) return
  
  ordering.value = true
  try {
    // 构建订单商品列表
    const skuItems = selectedItems.value.map(item => ({
      sku_id: item.id,
      quantity: item.quantity
    }))

    // 创建订单
    const orderRes = await orderApi.create({
      user_id: userStore.userId,
      sku_items: skuItems,
      pay_type: selectedPayment.value.channel === 'ecoin' ? 'ecoin' : 'money'
    })

    // 清空选中的商品
    cartStore.clearSelected()

    if (selectedPayment.value.channel === 'ecoin') {
      alert('下单成功！')
      userStore.refreshEcoin()
      router.push(`/orders/${orderRes.order_no}`)
    } else {
      const payRes = await orderApi.pay(orderRes.order_no, {
        channel: selectedPayment.value.channel,
        pay_method: selectedPayment.value.pay_method
      })
      
      if (payRes.code_url) {
        alert(`请使用微信扫描二维码完成支付\n${payRes.code_url}`)
      }
      router.push(`/orders/${orderRes.order_no}`)
    }
    closePaymentModal()
  } catch (error) {
    alert('下单失败: ' + error.message)
  } finally {
    ordering.value = false
  }
}
</script>

<style scoped>
.cart-page {
  padding-top: 20px;
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
  padding: 16px 20px;
  background-color: var(--gray-50);
  font-size: 14px;
  color: var(--gray-500);
  font-weight: 500;
}

.select-all {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.select-all input {
  width: 18px;
  height: 18px;
  accent-color: var(--primary-color);
}

.cart-list {
  padding: 0 20px;
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

.item-checkbox input {
  width: 18px;
  height: 18px;
  accent-color: var(--primary-color);
}

.item-product {
  display: flex;
  gap: 16px;
  cursor: pointer;
}

.product-image {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  overflow: hidden;
  background-color: var(--gray-100);
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
  width: 32px;
  height: 32px;
}

.product-info h4 {
  font-size: 14px;
  font-weight: 500;
  color: var(--gray-700);
  margin-bottom: 4px;
}

.product-info p {
  font-size: 12px;
  color: var(--gray-400);
}

.item-price {
  font-size: 14px;
  color: var(--gray-600);
}

.quantity-control {
  display: flex;
  align-items: center;
  border: 1px solid var(--gray-200);
  border-radius: 6px;
  overflow: hidden;
}

.quantity-control button {
  width: 28px;
  height: 28px;
  background-color: var(--gray-50);
  color: var(--gray-600);
  font-size: 14px;
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
  height: 28px;
  text-align: center;
  border: none;
  border-left: 1px solid var(--gray-200);
  border-right: 1px solid var(--gray-200);
  font-size: 14px;
}

.quantity-control input::-webkit-outer-spin-button,
.quantity-control input::-webkit-inner-spin-button {
  -webkit-appearance: none;
}

.item-total {
  font-size: 14px;
  font-weight: 600;
  color: var(--primary-color);
}

.delete-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--gray-400);
  background: none;
  transition: color 0.2s;
}

.delete-btn:hover {
  color: var(--danger);
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
  padding: 16px 24px;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 24px;
}

.clear-btn {
  color: var(--gray-500);
  background: none;
  font-size: 14px;
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
  color: var(--gray-700);
}

.total-price strong {
  font-size: 20px;
  color: var(--primary-color);
}

.checkout-btn {
  padding: 12px 48px;
  font-size: 16px;
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal-content {
  width: 400px;
  padding: 24px;
}

.modal-content h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--gray-700);
  margin-bottom: 20px;
}

.payment-methods {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
}

.payment-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.payment-option:hover {
  border-color: var(--primary-color);
}

.payment-option.active {
  border-color: var(--primary-color);
  background-color: rgba(26, 54, 93, 0.05);
}

.payment-option input {
  accent-color: var(--primary-color);
}

.option-name {
  font-size: 14px;
  color: var(--gray-700);
}

.modal-footer {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
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
