<template>
  <div class="order-detail">
    <div class="container">
      <div v-if="loading" class="loading"></div>

      <div v-else-if="!order" class="empty-state">
        <p>订单不存在</p>
        <router-link to="/orders" class="btn btn-primary" style="margin-top: 16px;">返回订单列表</router-link>
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

        <!-- 订单状态卡片 -->
        <div class="status-card card">
          <div class="status-icon" :class="getStatusClass(order.status)">
            <svg v-if="order.status === 0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            <svg v-else-if="order.status === 1 || order.status === 2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
              <polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="15" y1="9" x2="9" y2="15"/>
              <line x1="9" y1="9" x2="15" y2="15"/>
            </svg>
          </div>
          <div class="status-info">
            <h2>{{ getStatusText(order.status) }}</h2>
            <p v-if="order.status === 0" class="countdown-text">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="countdown-icon">
                <circle cx="12" cy="12" r="10"/>
                <polyline points="12 6 12 12 16 14"/>
              </svg>
              剩余支付时间：<strong>{{ countdownText }}</strong>
            </p>
            <p v-else-if="order.status === 1">订单正在处理中</p>
            <p v-else-if="order.status === 2">订单已完成</p>
            <p v-else-if="order.status === 3">订单已取消</p>
            <p v-else-if="order.status === 4">订单已退款</p>
          </div>
          <div v-if="canCancel(order.status) || canPay(order.status)" class="status-actions">
            <button 
              v-if="canCancel(order.status)"
              class="btn btn-secondary"
              @click="cancelOrder"
              :disabled="canceling"
            >
              {{ canceling ? '处理中...' : '取消订单' }}
            </button>
            <button 
              v-if="canPay(order.status)"
              class="btn btn-primary"
              @click="showPaymentModal = true; fetchPaymentMethods()"
            >
              去支付
            </button>
          </div>
        </div>

        <!-- 订单信息 -->
        <div class="info-card card">
          <h3>订单信息</h3>
          <div class="info-grid">
            <div class="info-item">
              <span class="label">订单编号</span>
              <span class="value">{{ order.order_no }}</span>
            </div>
            <div class="info-item">
              <span class="label">创建时间</span>
              <span class="value">{{ formatTime(order.ctime) }}</span>
            </div>
            <div class="info-item">
              <span class="label">支付类型</span>
              <span class="value">{{ order.pay_type === 'ecoin' ? '积分支付' : '在线支付' }}</span>
            </div>
            <div class="info-item" v-if="order.pay_time">
              <span class="label">支付时间</span>
              <span class="value">{{ formatTime(order.pay_time) }}</span>
            </div>
          </div>
        </div>

        <!-- 商品列表 -->
        <div class="items-card card">
          <h3>商品清单</h3>
          <div class="items-list">
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
                <p>{{ item.sku_code }}</p>
              </div>
              <div class="item-quantity">x{{ item.quantity }}</div>
              <div class="item-price">{{ (item.unit_price || 0).toFixed(2) }} 积分</div>
            </div>
          </div>

          <div class="items-summary">
            <div class="summary-row">
              <span>商品总价</span>
              <span>{{ (order.original_amount || 0).toFixed(2) }} 积分</span>
            </div>
            <div class="summary-row total">
              <span>实付金额</span>
              <span>{{ (order.pay_amount || 0).toFixed(2) }} 积分</span>
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
            <input type="radio" :value="method" v-model="selectedPayment">
            <span class="option-name">{{ method.name }}</span>
          </label>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closePaymentModal">取消</button>
          <button class="btn btn-primary" @click="payOrder" :disabled="!selectedPayment || paying">
            {{ paying ? '处理中...' : '确认支付' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 支付二维码弹窗 -->
    <div v-if="showQrcodeModal" class="modal-overlay" @click.self="closeQrcodeModal">
      <div class="modal-content card qrcode-modal">
        <h3>扫码支付</h3>
        <div class="qrcode-body">
          <div class="qrcode-container">
            <img v-if="qrcodeUrl" :src="qrcodeUrl" alt="支付二维码">
            <div v-else class="qrcode-placeholder">二维码加载中...</div>
          </div>
          <p class="qrcode-tip">请使用微信扫描二维码完成支付</p>
          <p class="qrcode-amount">支付金额：<strong>{{ (order?.pay_amount || 0).toFixed(2) }} 积分</strong></p>
          <p class="qrcode-polling">正在等待支付结果...</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeQrcodeModal">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { orderApi, paymentApi } from '../api'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const PAYMENT_TIMEOUT = 15 * 60

const order = ref(null)
const loading = ref(false)
const canceling = ref(false)
const paying = ref(false)

const showPaymentModal = ref(false)
const paymentMethods = ref([])
const selectedPayment = ref(null)

const showQrcodeModal = ref(false)
const qrcodeUrl = ref('')
let pollTimer = null

const now = ref(Math.floor(Date.now() / 1000))
let countdownTimer = null

const countdownText = computed(() => {
  if (!order.value) return ''
  const remaining = (order.value.ctime + PAYMENT_TIMEOUT) - now.value
  if (remaining <= 0) return '即将关闭'
  const min = Math.floor(remaining / 60)
  const sec = remaining % 60
  return `${min}:${sec.toString().padStart(2, '0')}`
})

const statusMap = {
  0: { text: '待支付', class: 'status-warning' },
  1: { text: '已支付', class: 'status-info' },
  2: { text: '已履约', class: 'status-success' },
  3: { text: '已取消', class: 'status-danger' },
  4: { text: '已退款', class: 'status-danger' }
}

function getStatusText(status) {
  return statusMap[status]?.text || '未知'
}

function getStatusClass(status) {
  return statusMap[status]?.class || ''
}

function canCancel(status) {
  return status === 0 || status === 1
}

function canPay(status) {
  return status === 0
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

async function fetchOrder(silent = false) {
  if (!silent) loading.value = true
  try {
    order.value = await orderApi.detail(route.params.orderNo)
  } catch (error) {
    console.error('获取订单详情失败:', error)
    if (!silent) {
      order.value = {
        order_no: route.params.orderNo,
        status: 0,
        pay_type: 'ecoin',
        original_amount: 300,
        pay_amount: 300,
        ctime: Date.now() / 1000 - 3600,
        items: [
          { id: 1, sku_code: 'SKU001', sku_name: '虚拟商品A', sku_avatar: '', quantity: 2, unit_price: 100 },
          { id: 2, sku_code: 'SKU002', sku_name: '虚拟商品B', sku_avatar: '', quantity: 1, unit_price: 100 }
        ]
      }
    }
  } finally {
    if (!silent) loading.value = false
  }
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

function goBack() {
  router.back()
}

async function cancelOrder() {
  if (!confirm('确定要取消这个订单吗？')) return
  
  canceling.value = true
  try {
    await orderApi.cancel(order.value.order_no, '用户主动取消')
    alert('订单已取消')
    userStore.refreshEcoin()
    fetchOrder(true)
  } catch (error) {
    alert('取消订单失败: ' + error.message)
  } finally {
    canceling.value = false
  }
}

function closePaymentModal() {
  showPaymentModal.value = false
  selectedPayment.value = null
}

async function payOrder() {
  if (!selectedPayment.value || paying.value) return
  
  paying.value = true
  try {
    const payRes = await orderApi.pay(order.value.order_no, {
      channel: selectedPayment.value.channel,
      pay_method: selectedPayment.value.pay_method
    })
    
    closePaymentModal()

    if (selectedPayment.value.channel === 'ecoin') {
      alert('支付成功！')
      userStore.refreshEcoin()
      fetchOrder(true)
    } else if (payRes.code_url) {
      qrcodeUrl.value = `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(payRes.code_url)}`
      showQrcodeModal.value = true
      startPolling()
    } else {
      fetchOrder(true)
    }
  } catch (error) {
    alert('支付失败: ' + error.message)
  } finally {
    paying.value = false
  }
}

function closeQrcodeModal() {
  showQrcodeModal.value = false
  qrcodeUrl.value = ''
  stopPolling()
}

function isActiveStatus() {
  return order.value && (order.value.status === 0 || order.value.status === 1)
}

async function pollOrderStatus() {
  if (!order.value || !isActiveStatus()) {
    stopPolling()
    return
  }
  try {
    const updated = await orderApi.sync(order.value.order_no)
    if (updated && updated.status !== order.value.status) {
      const prevStatus = order.value.status
      order.value = updated
      userStore.refreshEcoin()
      if (showQrcodeModal.value && prevStatus === 0 && (updated.status === 1 || updated.status === 2)) {
        closeQrcodeModal()
        alert('支付成功！')
      }
      if (!isActiveStatus()) {
        stopPolling()
        stopCountdown()
      }
    }
  } catch (e) {
    // 忽略
  }
}

function startPolling() {
  stopPolling()
  if (!isActiveStatus()) return
  pollTimer = setInterval(pollOrderStatus, 3000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function stopCountdown() {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

onMounted(async () => {
  await fetchOrder()
  if (isActiveStatus()) {
    startPolling()
    countdownTimer = setInterval(() => {
      now.value = Math.floor(Date.now() / 1000)
      if (order.value && order.value.status === 0) {
        const remaining = (order.value.ctime + PAYMENT_TIMEOUT) - now.value
        if (remaining <= 0) {
          fetchOrder(true).then(() => {
            if (!isActiveStatus()) {
              stopPolling()
              stopCountdown()
            }
          })
        }
      }
    }, 1000)
  }
})

onBeforeUnmount(() => {
  stopPolling()
  stopCountdown()
})
</script>

<style scoped>
.order-detail {
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

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 状态卡片 */
.status-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px;
}

.status-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.status-icon svg {
  width: 32px;
  height: 32px;
}

.status-warning {
  background-color: #fef3c7;
  color: #d97706;
}

.status-info {
  background-color: #dbeafe;
  color: #2563eb;
}

.status-success {
  background-color: #d1fae5;
  color: #059669;
}

.status-danger {
  background-color: #fee2e2;
  color: #dc2626;
}

.status-info h2 {
  font-size: 20px;
  font-weight: 600;
  color: var(--gray-700);
  margin-bottom: 4px;
}

.status-info p {
  font-size: 14px;
  color: var(--gray-500);
}

.countdown-text {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #d97706 !important;
  font-variant-numeric: tabular-nums;
}

.countdown-text strong {
  font-size: 18px;
  font-weight: 600;
}

.countdown-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.status-actions {
  margin-left: auto;
  display: flex;
  gap: 12px;
}

/* 信息卡片 */
.info-card,
.items-card {
  padding: 24px;
}

.info-card h3,
.items-card h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--gray-700);
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--gray-100);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item .label {
  font-size: 13px;
  color: var(--gray-400);
}

.info-item .value {
  font-size: 14px;
  color: var(--gray-700);
}

/* 商品列表 */
.items-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 20px;
}

.order-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.item-image {
  width: 64px;
  height: 64px;
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

.item-quantity {
  font-size: 14px;
  color: var(--gray-500);
}

.item-price {
  font-size: 14px;
  font-weight: 500;
  color: var(--gray-700);
  min-width: 100px;
  text-align: right;
}

.items-summary {
  padding-top: 20px;
  border-top: 1px solid var(--gray-100);
}

.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 14px;
  color: var(--gray-500);
}

.summary-row.total {
  font-size: 16px;
  font-weight: 600;
  color: var(--gray-700);
}

.summary-row.total span:last-child {
  color: var(--primary-color);
  font-size: 20px;
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

@media (max-width: 640px) {
  .status-card {
    flex-direction: column;
    text-align: center;
  }

  .status-actions {
    margin-left: 0;
    width: 100%;
    flex-direction: column;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .order-item {
    flex-wrap: wrap;
  }

  .item-price {
    width: 100%;
    text-align: left;
    margin-top: 8px;
  }
}

/* 二维码弹窗 */
.qrcode-modal {
  max-width: 360px;
}

.qrcode-body {
  padding: 20px;
  text-align: center;
}

.qrcode-container {
  width: 200px;
  height: 200px;
  margin: 0 auto 16px;
  background-color: var(--gray-100);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qrcode-container img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.qrcode-placeholder {
  color: var(--gray-400);
  font-size: 14px;
}

.qrcode-tip {
  font-size: 14px;
  color: var(--gray-500);
  margin-bottom: 8px;
}

.qrcode-amount {
  font-size: 14px;
  color: var(--gray-600);
  margin-bottom: 8px;
}

.qrcode-amount strong {
  color: var(--primary-color);
  font-size: 18px;
}

.qrcode-polling {
  font-size: 12px;
  color: var(--gray-400);
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
