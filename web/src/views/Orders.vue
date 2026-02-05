<template>
  <div class="orders-page">
    <div class="container">
      <div class="page-header">
        <h1>订单中心</h1>
      </div>

      <div v-if="loading" class="loading"></div>

      <div v-else-if="orders.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
        </svg>
        <p>暂无订单</p>
        <router-link to="/" class="btn btn-primary" style="margin-top: 16px;">去逛逛</router-link>
      </div>

      <div v-else class="orders-list">
        <div 
          v-for="order in orders" 
          :key="order.order_no" 
          class="order-card card"
          @click="goToDetail(order.order_no)"
        >
          <div class="order-header">
            <div class="order-info">
              <span class="order-no">订单号: {{ order.order_no }}</span>
              <span class="order-time">{{ formatTime(order.created_at) }}</span>
            </div>
            <span class="order-status" :class="getStatusClass(order.status)">
              {{ getStatusText(order.status) }}
            </span>
          </div>

          <div class="order-items">
            <div v-for="item in order.items" :key="item.id" class="order-item">
              <div class="item-image">
                <img v-if="item.sku_avatar" :src="item.sku_avatar" :alt="item.sku_name">
                <div v-else class="image-placeholder">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                  </svg>
                </div>
              </div>
              <div class="item-info">
                <h4>{{ item.sku_name }}</h4>
                <p>x{{ item.quantity }}</p>
              </div>
              <div class="item-price">
                {{ item.price.toFixed(2) }} 积分
              </div>
            </div>
          </div>

          <div class="order-footer">
            <div class="order-total">
              共 {{ getTotalQuantity(order) }} 件商品，合计: 
              <strong>{{ order.total_amount.toFixed(2) }}</strong> 积分
            </div>
            <div class="order-actions" @click.stop>
              <button 
                v-if="canCancel(order.status)"
                class="btn btn-secondary btn-sm"
                @click="cancelOrder(order)"
              >
                取消订单
              </button>
              <button 
                v-if="canPay(order.status)"
                class="btn btn-primary btn-sm"
                @click="payOrder(order)"
              >
                去支付
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { orderApi } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const orders = ref([])
const loading = ref(false)

// 订单状态映射
const statusMap = {
  0: { text: '待支付', class: 'tag-warning' },
  1: { text: '已支付', class: 'tag-info' },
  2: { text: '已履约', class: 'tag-success' },
  3: { text: '已取消', class: 'tag-danger' },
  4: { text: '已退款', class: 'tag-danger' }
}

function getStatusText(status) {
  return statusMap[status]?.text || '未知'
}

function getStatusClass(status) {
  return statusMap[status]?.class || ''
}

function canCancel(status) {
  return status === 0 || status === 1 // 待支付或已支付可取消
}

function canPay(status) {
  return status === 0 // 待支付可支付
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function getTotalQuantity(order) {
  return order.items?.reduce((sum, item) => sum + item.quantity, 0) || 0
}

async function fetchOrders() {
  loading.value = true
  try {
    const res = await orderApi.list({ user_id: userStore.userId, limit: 100 })
    orders.value = res?.list || []
  } catch (error) {
    console.error('获取订单列表失败:', error)
    // Mock 数据
    orders.value = [
      {
        order_no: 'ORD20240101001',
        status: 0,
        total_amount: 300,
        created_at: Date.now() / 1000 - 3600,
        items: [
          { id: 1, sku_name: '虚拟商品A', sku_avatar: '', quantity: 2, price: 100 },
          { id: 2, sku_name: '虚拟商品B', sku_avatar: '', quantity: 1, price: 100 }
        ]
      },
      {
        order_no: 'ORD20240101002',
        status: 2,
        total_amount: 500,
        created_at: Date.now() / 1000 - 86400,
        items: [
          { id: 3, sku_name: '高级会员', sku_avatar: '', quantity: 1, price: 500 }
        ]
      }
    ]
  } finally {
    loading.value = false
  }
}

function goToDetail(orderNo) {
  router.push(`/orders/${orderNo}`)
}

async function cancelOrder(order) {
  if (!confirm('确定要取消这个订单吗？')) return
  
  try {
    await orderApi.cancel(order.order_no, '用户主动取消')
    alert('订单已取消')
    fetchOrders()
    userStore.refreshEcoin()
  } catch (error) {
    alert('取消订单失败: ' + error.message)
  }
}

function payOrder(order) {
  router.push(`/orders/${order.order_no}`)
}

onMounted(() => {
  fetchOrders()
})
</script>

<style scoped>
.orders-page {
  padding-top: 20px;
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.order-card {
  padding: 20px;
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.order-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--gray-100);
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.order-no {
  font-size: 14px;
  font-weight: 500;
  color: var(--gray-700);
}

.order-time {
  font-size: 12px;
  color: var(--gray-400);
}

.order-status {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}

.order-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.order-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.item-image {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  overflow: hidden;
  background-color: var(--gray-100);
  flex-shrink: 0;
}

.item-image img {
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
  width: 24px;
  height: 24px;
}

.item-info {
  flex: 1;
}

.item-info h4 {
  font-size: 14px;
  font-weight: 500;
  color: var(--gray-700);
  margin-bottom: 2px;
}

.item-info p {
  font-size: 12px;
  color: var(--gray-400);
}

.item-price {
  font-size: 14px;
  color: var(--gray-600);
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 16px;
  border-top: 1px solid var(--gray-100);
}

.order-total {
  font-size: 14px;
  color: var(--gray-500);
}

.order-total strong {
  font-size: 18px;
  color: var(--primary-color);
}

.order-actions {
  display: flex;
  gap: 12px;
}

.btn-sm {
  padding: 8px 16px;
  font-size: 13px;
}

@media (max-width: 640px) {
  .order-card {
    padding: 16px;
  }

  .order-footer {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .order-actions {
    justify-content: flex-end;
  }
}
</style>
