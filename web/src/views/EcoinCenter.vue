<template>
  <div class="ecoin-center-page">
    <div class="container">
      <div class="page-header">
        <h1>积分中心</h1>
      </div>

      <!-- 积分余额卡片 -->
      <div class="balance-card card">
        <div class="balance-info">
          <div class="balance-label">当前积分余额{{ ecoinEnterpriseSuffix }}</div>
          <div class="balance-value">
            <span class="balance-number">{{ balance.toFixed(2) }}</span>
            <span class="balance-unit">积分</span>
          </div>
        </div>
        <button class="btn btn-primary recharge-btn" @click="openRechargeModal">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          充值
        </button>
      </div>

      <!-- 积分分组明细 -->
      <div class="stock-groups-section card">
        <div class="section-header">
          <h2>积分明细{{ ecoinEnterpriseSuffix }}</h2>
        </div>

        <div v-if="groupLoading" class="loading"></div>

        <div v-else-if="stockGroups.length === 0" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M20 7L9 18l-5-5"/>
          </svg>
          <p>暂无可用积分批次</p>
        </div>

        <div v-else class="groups-list">
          <div v-for="group in stockGroups" :key="group.id" class="group-item">
            <div class="group-main">
              <div class="group-balance">
                <span class="group-remaining">{{ Number(group.remaining_stock || 0).toFixed(2) }}</span>
                <span class="group-unit">积分</span>
              </div>
              <div class="group-meta">
                <span class="group-source">{{ getSourceTypeText(group.source_type) }}</span>
                <span class="group-expire" :class="{ 'expire-soon': isExpiringSoon(group.expire_time) }">
                  {{ formatExpireTime(group.expire_time) }}
                </span>
              </div>
            </div>
            <div class="group-progress">
              <div class="group-progress-bar" :style="{ width: `${groupProgress(group)}%` }"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- 积分流水 -->
      <div class="transactions-section card">
        <div class="section-header">
          <h2>积分流水{{ ecoinEnterpriseSuffix }}</h2>
        </div>

        <div v-if="loading" class="loading"></div>

        <div v-else-if="transactions.length === 0" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
          </svg>
          <p>暂无积分流水</p>
        </div>

        <div v-else class="transactions-list">
          <div 
            v-for="tx in transactions" 
            :key="tx.id" 
            class="transaction-item"
          >
            <div class="tx-info">
              <div class="tx-desc">{{ tx.description || getSourceTypeText(tx.source_type) }}</div>
              <div v-if="userStore.isEnterpriseAccount" class="tx-operator">
                操作人：{{ formatTxOperator(tx) }}
              </div>
              <div class="tx-time">{{ formatTime(tx.ctime) }}</div>
            </div>
            <div class="tx-amount" :class="tx.amount >= 0 ? 'positive' : 'negative'">
              {{ tx.amount >= 0 ? '+' : '' }}{{ tx.amount.toFixed(2) }}
            </div>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="total > pageSize" class="pagination">
          <button 
            class="page-btn" 
            :disabled="currentPage === 1"
            @click="changePage(currentPage - 1)"
          >
            上一页
          </button>
          <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
          <button 
            class="page-btn" 
            :disabled="currentPage === totalPages"
            @click="changePage(currentPage + 1)"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- 充值弹窗 -->
    <div v-if="showRechargeModal" class="modal-overlay" @click.self="closeRechargeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>积分充值</h3>
          <button class="close-btn" @click="closeRechargeModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>充值数量</label>
            <input 
              type="number" 
              v-model.number="rechargeAmount" 
              :placeholder="`最低充值 ${minRechargeAmount} 积分`"
              :min="minRechargeAmount"
              class="form-input"
            >
            <div class="form-hint">
              最低充值 {{ minRechargeAmount }} 积分<template v-if="ecoinUnitPrice">，单价 {{ ecoinUnitPrice }} 元/积分</template>
            </div>
          </div>
          <div class="form-group">
            <label>支付方式</label>
            <div class="payment-methods">
              <div 
                v-for="method in paymentMethods" 
                :key="method.channel + method.pay_method"
                class="payment-method"
                :class="{ active: selectedPayment?.channel === method.channel && selectedPayment?.pay_method === method.pay_method }"
                @click="selectPayment(method)"
              >
                <div class="method-icon">
                  <svg v-if="method.channel === 'wechat'" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1-.023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-6.656-6.088V8.89c-.135-.003-.27-.012-.407-.033zm-2.53 3.274c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982z"/>
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
                    <line x1="1" y1="10" x2="23" y2="10"/>
                  </svg>
                </div>
                <div class="method-info">
                  <div class="method-name">{{ method.name }}</div>
                </div>
                <div class="method-check" v-if="selectedPayment?.channel === method.channel && selectedPayment?.pay_method === method.pay_method">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"/>
                  </svg>
                </div>
              </div>
            </div>
          </div>
          <div class="recharge-summary">
            <div>充值数量：<strong>{{ rechargeAmount || 0 }} 积分</strong></div>
            <div v-if="ecoinUnitPrice" class="pay-amount">
              需支付：<strong>{{ ((rechargeAmount || 0) * ecoinUnitPrice).toFixed(2) }} 元</strong>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeRechargeModal">取消</button>
          <button 
            class="btn btn-primary" 
            :disabled="!rechargeAmount || rechargeAmount < minRechargeAmount || !selectedPayment || recharging"
            @click="handleRecharge"
          >
            {{ recharging ? '处理中...' : '确认充值' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 支付二维码弹窗 -->
    <div v-if="showPaymentModal" class="modal-overlay" @click.self="closePaymentModal">
      <div class="modal payment-modal">
        <div class="modal-header">
          <h3>扫码支付</h3>
          <button class="close-btn" @click="closePaymentModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body payment-body">
          <div class="qrcode-container">
            <img :src="qrcodeUrl" alt="支付二维码" v-if="qrcodeUrl">
            <div v-else class="qrcode-placeholder">二维码加载中...</div>
          </div>
          <p class="payment-tip">请使用微信扫描二维码完成支付</p>
          <p class="payment-amount">支付金额：<strong>{{ pendingRechargeAmount }} 积分</strong></p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="checkPaymentStatus">我已支付</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { ecoinApi, paymentApi, orderApi } from '../api'
import { useUserStore } from '../stores/user'
import { toast } from '../utils/toast'

const userStore = useUserStore()

const loading = ref(false)
const groupLoading = ref(false)
const transactions = ref([])
const stockGroups = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 10

const showRechargeModal = ref(false)
const rechargeAmount = ref(100)
const selectedPayment = ref(null)
const paymentMethods = ref([])
const recharging = ref(false)
const minRechargeAmount = ref(1)
const ecoinUnitPrice = ref(0)

const showPaymentModal = ref(false)
const qrcodeUrl = ref('')
const pendingOrderNo = ref('')
const pendingRechargeAmount = ref(0)
let pollTimer = null

const balance = computed(() => userStore.balance)
const totalPages = computed(() => Math.ceil(total.value / pageSize))
const ecoinEnterpriseSuffix = computed(() => (userStore.isEnterpriseAccount ? '（企业）' : ''))

function formatTxOperator(tx) {
  if (tx.operator_label) return tx.operator_label
  const oid = tx.operator_user_id
  if (oid != null && Number(oid) > 0) return `用户 #${oid}`
  return '系统'
}

const sourceTypeMap = {
  'order': '订单消费',
  'refund': '订单退款',
  'recharge': '积分充值',
  'admin': '系统调整',
  'system': '系统发放',
  'expire': '过期失效'
}

function getSourceTypeText(sourceType) {
  return sourceTypeMap[sourceType] || sourceType || '积分变动'
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

function formatExpireTime(timestamp) {
  if (!timestamp || timestamp <= 0) return '永久有效'
  const date = new Date(timestamp * 1000)
  return `到期时间：${date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })}`
}

function isExpiringSoon(timestamp) {
  if (!timestamp || timestamp <= 0) return false
  const now = Date.now()
  return timestamp * 1000 - now <= 3 * 24 * 3600 * 1000
}

function groupProgress(group) {
  const total = Number(group.total_stock || 0)
  const left = Number(group.remaining_stock || 0)
  if (total <= 0) return 0
  return Math.max(0, Math.min(100, (left / total) * 100))
}

async function fetchStockGroups() {
  groupLoading.value = true
  try {
    const res = await ecoinApi.getStockGroups(userStore.userId)
    stockGroups.value = res?.list || []
  } catch (error) {
    console.error('获取积分明细失败:', error)
    stockGroups.value = []
  } finally {
    groupLoading.value = false
  }
}

async function fetchTransactions() {
  loading.value = true
  try {
    const offset = (currentPage.value - 1) * pageSize
    const res = await ecoinApi.getTransactions({
      user_id: userStore.userId,
      offset,
      limit: pageSize
    })
    transactions.value = res?.list || []
    total.value = res?.total || 0
  } catch (error) {
    console.error('获取积分流水失败:', error)
    transactions.value = []
  } finally {
    loading.value = false
  }
}

async function fetchPaymentMethods() {
  try {
    const methods = await paymentApi.methods()
    // 过滤掉积分支付
    paymentMethods.value = (methods || []).filter(m => m.channel !== 'ecoin')
    if (paymentMethods.value.length > 0) {
      selectedPayment.value = paymentMethods.value[0]
    }
  } catch (error) {
    console.error('获取支付方式失败:', error)
    // Mock 数据
    paymentMethods.value = [
      { channel: 'wechat', name: '微信支付', pay_method: 'native' }
    ]
    selectedPayment.value = paymentMethods.value[0]
  }
}

function changePage(page) {
  currentPage.value = page
  fetchTransactions()
}

async function openRechargeModal() {
  showRechargeModal.value = true
  fetchPaymentMethods()
  try {
    const cfg = await ecoinApi.getRechargeConfig()
    minRechargeAmount.value = cfg?.min_amount || 1
    ecoinUnitPrice.value = cfg?.unit_price || 0
    rechargeAmount.value = Math.max(100, minRechargeAmount.value)
  } catch (e) {
    minRechargeAmount.value = 1
    rechargeAmount.value = 100
  }
}

function closeRechargeModal() {
  showRechargeModal.value = false
}

function selectPayment(method) {
  selectedPayment.value = method
}

async function handleRecharge() {
  if (!rechargeAmount.value || rechargeAmount.value <= 0) {
    toast.warning('请输入正确的充值数量')
    return
  }
  if (rechargeAmount.value < minRechargeAmount.value) {
    toast.warning(`最低充值数量为 ${minRechargeAmount.value} 积分`)
    return
  }
  if (!selectedPayment.value) {
    toast.warning('请选择支付方式')
    return
  }

  recharging.value = true
  try {
    const res = await ecoinApi.recharge({
      user_id: userStore.userId,
      amount: rechargeAmount.value,
      pay_type: 'money',
      channel: selectedPayment.value.channel,
      pay_method: selectedPayment.value.pay_method
    })

    closeRechargeModal()

    if (res?.payment_info?.code_url) {
      pendingOrderNo.value = res.payment_info.order_no || res.order?.order_no
      pendingRechargeAmount.value = rechargeAmount.value
      qrcodeUrl.value = `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(res.payment_info.code_url)}`
      showPaymentModal.value = true
      startPolling()
    } else if (res?.order) {
      toast.info('充值订单已创建，请完成支付')
    }
  } catch (error) {
    toast.error('充值失败: ' + error.message)
  } finally {
    recharging.value = false
  }
}

function closePaymentModal() {
  showPaymentModal.value = false
  qrcodeUrl.value = ''
  pendingOrderNo.value = ''
  stopPolling()
}

async function pollRechargeStatus() {
  if (!pendingOrderNo.value) return
  try {
    const updated = await orderApi.sync(pendingOrderNo.value)
    if (updated && updated.status !== 0) {
      await userStore.fetchEcoin()
      await fetchStockGroups()
      await fetchTransactions()
      closePaymentModal()
      if (updated.status === 1 || updated.status === 2) {
        toast.success('充值成功！积分已到账')
      }
    }
  } catch (e) {
    // 忽略
  }
}

function startPolling() {
  stopPolling()
  pollTimer = setInterval(pollRechargeStatus, 3000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function checkPaymentStatus() {
  await userStore.fetchEcoin()
  await fetchStockGroups()
  await fetchTransactions()
  closePaymentModal()
  toast.info('积分余额已刷新，如未到账请稍后再试')
}

onMounted(() => {
  userStore.fetchEcoin()
  fetchStockGroups()
  fetchTransactions()
})

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<style scoped>
.ecoin-center-page {
  padding-top: 8px;
}

.balance-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 36px 40px;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 60%, #334155 100%);
  color: white;
  margin-bottom: 28px;
  border: none;
  position: relative;
  overflow: hidden;
}

.balance-card::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -20%;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.15), transparent 70%);
  pointer-events: none;
}

.balance-info {
  position: relative;
}

.balance-label {
  font-size: 13px;
  opacity: 0.7;
  margin-bottom: 8px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.balance-value {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.balance-number {
  font-size: 44px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.balance-unit {
  font-size: 16px;
  opacity: 0.7;
  font-weight: 500;
}

.recharge-btn {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 12px 28px;
  border-radius: 10px;
  backdrop-filter: blur(8px);
  box-shadow: none;
  font-weight: 600;
}

.recharge-btn:hover {
  background: rgba(255, 255, 255, 0.25);
  transform: translateY(-1px);
}

.recharge-btn svg {
  width: 18px;
  height: 18px;
}

.stock-groups-section {
  padding: 24px 28px;
  margin-bottom: 20px;
}

.groups-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.group-item {
  border: 1px solid var(--gray-100);
  border-radius: 12px;
  padding: 14px 16px;
  background: #fff;
}

.group-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 10px;
}

.group-balance {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.group-remaining {
  font-size: 24px;
  font-weight: 700;
  color: var(--warning);
}

.group-unit {
  font-size: 13px;
  color: var(--gray-500);
}

.group-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.group-source {
  font-size: 12px;
  color: var(--gray-500);
}

.group-expire {
  font-size: 12px;
  color: var(--gray-400);
}

.group-expire.expire-soon {
  color: #d97706;
  font-weight: 600;
}

.group-progress {
  height: 6px;
  border-radius: 999px;
  background: var(--gray-100);
  overflow: hidden;
}

.group-progress-bar {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #f59e0b, #fbbf24);
}

.transactions-section {
  padding: 28px 32px;
}

.section-header {
  margin-bottom: 20px;
}

.section-header h2 {
  font-size: 13px;
  font-weight: 700;
  color: var(--gray-400);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.transactions-list {
  display: flex;
  flex-direction: column;
}

.transaction-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid var(--gray-100);
}

.transaction-item:last-child {
  border-bottom: none;
}

.tx-info {
  flex: 1;
}

.tx-desc {
  font-size: 14px;
  font-weight: 500;
  color: var(--gray-800);
  margin-bottom: 4px;
}

.tx-operator {
  font-size: 12px;
  color: var(--gray-500);
  margin-top: 4px;
  font-weight: 500;
}

.tx-time {
  font-size: 12px;
  color: var(--gray-400);
}

.tx-amount {
  font-size: 16px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.tx-amount.positive {
  color: var(--success);
}

.tx-amount.negative {
  color: var(--danger);
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--gray-100);
}

.page-btn {
  padding: 8px 18px;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  background-color: white;
  color: var(--gray-600);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-info {
  font-size: 13px;
  color: var(--gray-400);
  font-weight: 500;
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal {
  background-color: white;
  border-radius: 20px;
  width: 90%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: var(--shadow-xl);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 28px;
  border-bottom: 1px solid var(--gray-100);
}

.modal-header h3 {
  font-size: 18px;
  font-weight: 700;
  color: var(--gray-800);
  letter-spacing: -0.01em;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: none;
  cursor: pointer;
  color: var(--gray-400);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.15s;
}

.close-btn:hover {
  background-color: var(--gray-100);
  color: var(--gray-600);
}

.close-btn svg {
  width: 20px;
  height: 20px;
}

.modal-body {
  padding: 28px;
}

.form-group {
  margin-bottom: 22px;
}

.form-group label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--gray-600);
  margin-bottom: 8px;
  letter-spacing: 0.02em;
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--gray-200);
  border-radius: 10px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.form-hint {
  margin-top: 6px;
  font-size: 12px;
  color: var(--gray-400);
}

.payment-methods {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.payment-method {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px;
  border: 2px solid var(--gray-200);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.payment-method:hover {
  border-color: var(--gray-300);
}

.payment-method.active {
  border-color: var(--accent);
  background-color: rgba(99, 102, 241, 0.04);
}

.method-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background-color: var(--gray-50);
  display: flex;
  align-items: center;
  justify-content: center;
}

.method-icon svg {
  width: 22px;
  height: 22px;
  color: #07c160;
}

.method-info {
  flex: 1;
}

.method-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--gray-800);
}

.method-check {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), var(--accent-light));
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
}

.method-check svg {
  width: 14px;
  height: 14px;
}

.recharge-summary {
  padding: 18px;
  background: linear-gradient(135deg, #f8fafc, #f1f5f9);
  border-radius: 12px;
  text-align: center;
  font-size: 14px;
  color: var(--gray-600);
  border: 1px solid var(--gray-100);
}

.recharge-summary strong {
  font-size: 22px;
  color: var(--primary-color);
  font-weight: 700;
  letter-spacing: -0.02em;
}

.pay-amount {
  margin-top: 8px;
  font-size: 13px;
  color: var(--gray-500);
}

.pay-amount strong {
  font-size: 16px;
}

.modal-footer {
  display: flex;
  gap: 10px;
  padding: 20px 28px;
  border-top: 1px solid var(--gray-100);
  justify-content: flex-end;
}

/* 支付二维码弹窗 */
.payment-modal {
  max-width: 380px;
}

.payment-body {
  text-align: center;
}

.qrcode-container {
  width: 200px;
  height: 200px;
  margin: 0 auto 20px;
  background-color: var(--gray-50);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--gray-100);
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

.payment-tip {
  font-size: 14px;
  color: var(--gray-500);
  margin-bottom: 8px;
}

.payment-amount {
  font-size: 14px;
  color: var(--gray-600);
}

.payment-amount strong {
  color: var(--primary-color);
  font-size: 20px;
  font-weight: 700;
}

@media (max-width: 640px) {
  .balance-card {
    flex-direction: column;
    gap: 24px;
    text-align: center;
    padding: 28px;
  }

  .balance-number {
    font-size: 36px;
  }

  .recharge-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
