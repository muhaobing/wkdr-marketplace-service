<template>
  <div class="product-detail">
    <div class="container">
      <div v-if="loading" class="loading"></div>

      <div v-else-if="!product" class="empty-state">
        <p>商品不存在</p>
        <router-link to="/" class="btn btn-primary" style="margin-top: 16px;">返回首页</router-link>
      </div>

      <div v-else class="detail-content">
        <!-- 返回按钮 -->
        <button class="back-btn" @click="goBack">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="19" y1="12" x2="5" y2="12"/>
            <polyline points="12 19 5 12 12 5"/>
          </svg>
          返回
        </button>

        <div class="detail-main card">
          <div class="detail-grid">
            <!-- 商品图片 -->
            <div class="product-image">
              <img v-if="product.sku_avatar" :src="product.sku_avatar" :alt="product.sku_name">
              <div v-else class="image-placeholder">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <polyline points="21 15 16 10 5 21"/>
                </svg>
              </div>
            </div>

            <!-- 商品信息 -->
            <div class="product-info">
              <h1 class="product-name">{{ product.sku_name }}</h1>
              <p class="product-code">商品编码: {{ product.sku_code }}</p>
              
              <div class="product-price">
                <span class="price-value">{{ product.cost.toFixed(2) }}</span>
                <span class="price-unit">积分</span>
              </div>

              <div class="product-desc">
                <h3>商品描述</h3>
                <p>{{ product.sku_desc || '暂无描述' }}</p>
              </div>

              <!-- 数量选择 -->
              <div class="quantity-section">
                <span class="quantity-label">数量</span>
                <div class="quantity-control">
                  <button class="qty-btn" @click="decreaseQty" :disabled="quantity <= 1">-</button>
                  <input type="number" v-model.number="quantity" min="1" max="99">
                  <button class="qty-btn" @click="increaseQty" :disabled="quantity >= 99">+</button>
                </div>
              </div>

              <!-- 操作按钮 -->
              <div class="action-buttons">
                <button class="btn btn-secondary" @click="addToCart">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="9" cy="21" r="1"/>
                    <circle cx="20" cy="21" r="1"/>
                    <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
                  </svg>
                  加入购物车
                </button>
                <button class="btn btn-primary" @click="buyNow">
                  立即购买
                </button>
              </div>
            </div>
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
            <input 
              type="radio" 
              :value="method" 
              v-model="selectedPayment"
            >
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
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { skuApi, orderApi, paymentApi } from '../api'
import { useCartStore } from '../stores/cart'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()

const product = ref(null)
const loading = ref(false)
const quantity = ref(1)

const showPaymentModal = ref(false)
const paymentMethods = ref([])
const selectedPayment = ref(null)
const ordering = ref(false)

// 获取商品详情
async function fetchProduct() {
  loading.value = true
  try {
    product.value = await skuApi.detail(route.params.id)
  } catch (error) {
    console.error('获取商品详情失败:', error)
    // Mock 数据
    product.value = {
      id: parseInt(route.params.id),
      sku_code: 'SKU001',
      sku_name: '虚拟商品A',
      sku_desc: '这是一个详细的商品描述，包含商品的各种信息和特点。',
      cost: 100,
      sku_avatar: ''
    }
  } finally {
    loading.value = false
  }
}

// 获取支付方式
async function fetchPaymentMethods() {
  try {
    paymentMethods.value = await paymentApi.methods()
  } catch (error) {
    console.error('获取支付方式失败:', error)
    // Mock 数据
    paymentMethods.value = [
      { channel: 'ecoin', name: '积分支付', pay_method: 'ecoin' },
      { channel: 'wechat', name: '微信扫码支付', pay_method: 'native' }
    ]
  }
}

function decreaseQty() {
  if (quantity.value > 1) quantity.value--
}

function increaseQty() {
  if (quantity.value < 99) quantity.value++
}

function goBack() {
  router.back()
}

function addToCart() {
  cartStore.addItem(product.value, quantity.value)
  alert('已加入购物车')
}

function buyNow() {
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
    // 创建订单
    const orderRes = await orderApi.create({
      user_id: userStore.userId,
      sku_items: [{ sku_id: product.value.id, quantity: quantity.value }],
      pay_type: selectedPayment.value.channel === 'ecoin' ? 'ecoin' : 'money'
    })

    // 如果是积分支付，直接成功
    if (selectedPayment.value.channel === 'ecoin') {
      alert('下单成功！')
      userStore.refreshEcoin()
      router.push(`/orders/${orderRes.order_no}`)
    } else {
      // 发起支付
      const payRes = await orderApi.pay(orderRes.order_no, {
        channel: selectedPayment.value.channel,
        pay_method: selectedPayment.value.pay_method
      })
      
      // 显示支付信息（如二维码等）
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

onMounted(() => {
  fetchProduct()
})
</script>

<style scoped>
.product-detail {
  padding-top: 20px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  margin-bottom: 20px;
  background: none;
  color: var(--gray-600);
  font-size: 14px;
  transition: color 0.2s;
}

.back-btn:hover {
  color: var(--primary-color);
}

.back-btn svg {
  width: 18px;
  height: 18px;
}

.detail-main {
  padding: 32px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 400px 1fr;
  gap: 40px;
}

.product-image {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 12px;
  overflow: hidden;
  background-color: var(--gray-100);
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
  width: 80px;
  height: 80px;
}

.product-name {
  font-size: 28px;
  font-weight: 600;
  color: var(--gray-700);
  margin-bottom: 8px;
}

.product-code {
  font-size: 14px;
  color: var(--gray-400);
  margin-bottom: 20px;
}

.product-price {
  margin-bottom: 24px;
}

.price-value {
  font-size: 36px;
  font-weight: 700;
  color: var(--primary-color);
}

.price-unit {
  font-size: 16px;
  color: var(--gray-500);
  margin-left: 4px;
}

.product-desc {
  margin-bottom: 24px;
  padding: 20px;
  background-color: var(--gray-50);
  border-radius: 8px;
}

.product-desc h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--gray-600);
  margin-bottom: 8px;
}

.product-desc p {
  font-size: 14px;
  color: var(--gray-500);
  line-height: 1.6;
}

.quantity-section {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
}

.quantity-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--gray-600);
}

.quantity-control {
  display: flex;
  align-items: center;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  overflow: hidden;
}

.qty-btn {
  width: 36px;
  height: 36px;
  background-color: var(--gray-50);
  color: var(--gray-600);
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.qty-btn:hover:not(:disabled) {
  background-color: var(--gray-100);
}

.qty-btn:disabled {
  color: var(--gray-300);
  cursor: not-allowed;
}

.quantity-control input {
  width: 60px;
  height: 36px;
  text-align: center;
  border: none;
  border-left: 1px solid var(--gray-200);
  border-right: 1px solid var(--gray-200);
  font-size: 16px;
}

.quantity-control input::-webkit-outer-spin-button,
.quantity-control input::-webkit-inner-spin-button {
  -webkit-appearance: none;
}

.action-buttons {
  display: flex;
  gap: 16px;
}

.action-buttons .btn {
  flex: 1;
  padding: 14px 24px;
  font-size: 16px;
}

.action-buttons .btn svg {
  width: 20px;
  height: 20px;
  margin-right: 8px;
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
  .detail-grid {
    grid-template-columns: 1fr;
    gap: 24px;
  }

  .detail-main {
    padding: 20px;
  }

  .product-name {
    font-size: 22px;
  }

  .price-value {
    font-size: 28px;
  }

  .action-buttons {
    flex-direction: column;
  }
}
</style>
