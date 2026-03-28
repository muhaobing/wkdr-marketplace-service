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
                <span class="price-value">¥{{ product.cost.toFixed(2) }}</span>
                <span v-if="!isEcoinGrantSku(product)" class="price-ecoin">({{ toEcoin(product.cost) }} 积分)</span>
              </div>

              <div class="product-desc">
                <h3>商品描述</h3>
                <p>{{ product.sku_desc || '暂无描述' }}</p>
              </div>

              <!-- 数量选择 -->
              <div v-if="product.multi_select === 1" class="quantity-section">
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
                <button class="btn btn-primary" @click="buyNow" :disabled="ordering">
                  {{ ordering ? '下单中...' : '立即购买' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { skuApi, orderApi, ecoinApi } from '../api'
import { useCartStore } from '../stores/cart'
import { useUserStore } from '../stores/user'
import { toast } from '../utils/toast'
import { isEcoinGrantSku } from '../utils/sku'

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()

const product = ref(null)
const loading = ref(false)
const quantity = ref(1)
const ecoinUnitPrice = ref(0)
const ordering = ref(false)

function toEcoin(cost) {
  if (!ecoinUnitPrice.value || ecoinUnitPrice.value <= 0) return '--'
  return (cost / ecoinUnitPrice.value).toFixed(2)
}

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

function decreaseQty() {
  if (quantity.value > 1) quantity.value--
}

function increaseQty() {
  if (quantity.value < 99) quantity.value++
}

function goBack() {
  router.back()
}

async function addToCart() {
  try {
    await cartStore.addItem(product.value, quantity.value)
    toast.success('已加入购物车')
  } catch (error) {
    toast.error('添加失败: ' + error.message)
  }
}

async function buyNow() {
  if (ordering.value) return
  ordering.value = true
  try {
    const orderRes = await orderApi.create({
      user_id: userStore.userId,
      sku_items: [{ sku_id: product.value.id, quantity: quantity.value }]
    })
    const orderNo = orderRes.order?.order_no || orderRes.order_no
    router.push(`/orders/${orderNo}`)
  } catch (error) {
    toast.error('下单失败: ' + error.message)
  } finally {
    ordering.value = false
  }
}

onMounted(async () => {
  fetchProduct()
  try {
    const cfg = await ecoinApi.getRechargeConfig()
    ecoinUnitPrice.value = cfg.unit_price || 0
  } catch (e) { /* ignore */ }
})
</script>

<style scoped>
.product-detail {
  padding-top: 8px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  margin-bottom: 24px;
  background: none;
  color: var(--gray-500);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
  border-radius: 8px;
}

.back-btn:hover {
  color: var(--gray-800);
  background-color: var(--gray-100);
}

.back-btn svg {
  width: 18px;
  height: 18px;
}

.detail-main {
  padding: 40px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 420px 1fr;
  gap: 48px;
}

.product-image {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 16px;
  overflow: hidden;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
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
  font-weight: 700;
  color: var(--gray-800);
  margin-bottom: 8px;
  letter-spacing: -0.02em;
  line-height: 1.3;
}

.product-code {
  font-size: 13px;
  color: var(--gray-400);
  margin-bottom: 24px;
  font-weight: 500;
}

.product-price {
  margin-bottom: 28px;
  padding: 20px 24px;
  background: linear-gradient(135deg, #fafafa 0%, #f5f5f5 100%);
  border-radius: 12px;
  border: 1px solid rgba(0,0,0,0.04);
}

.price-value {
  font-size: 36px;
  font-weight: 700;
  color: #b91c1c;
  letter-spacing: -0.02em;
}

.price-ecoin {
  font-size: 15px;
  color: #f59e0b;
  margin-left: 10px;
  font-weight: 600;
}

.product-desc {
  margin-bottom: 28px;
}

.product-desc h3 {
  font-size: 13px;
  font-weight: 600;
  color: var(--gray-500);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.product-desc p {
  font-size: 14px;
  color: var(--gray-500);
  line-height: 1.7;
}

.quantity-section {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
}

.quantity-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--gray-500);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.quantity-control {
  display: flex;
  align-items: center;
  border: 1px solid var(--gray-200);
  border-radius: 10px;
  overflow: hidden;
}

.qty-btn {
  width: 40px;
  height: 40px;
  background-color: var(--gray-50);
  color: var(--gray-600);
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.qty-btn:hover:not(:disabled) {
  background-color: var(--gray-100);
}

.qty-btn:disabled {
  color: var(--gray-300);
  cursor: not-allowed;
}

.quantity-control input {
  width: 64px;
  height: 40px;
  text-align: center;
  border: none;
  border-left: 1px solid var(--gray-200);
  border-right: 1px solid var(--gray-200);
  font-size: 15px;
  font-weight: 600;
}

.quantity-control input::-webkit-outer-spin-button,
.quantity-control input::-webkit-inner-spin-button {
  -webkit-appearance: none;
}

.action-buttons {
  display: flex;
  gap: 14px;
}

.action-buttons .btn {
  flex: 1;
  padding: 15px 24px;
  font-size: 15px;
  font-weight: 600;
}

.action-buttons .btn svg {
  width: 20px;
  height: 20px;
  margin-right: 8px;
}

@media (max-width: 768px) {
  .detail-grid {
    grid-template-columns: 1fr;
    gap: 24px;
  }

  .detail-main {
    padding: 24px;
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
