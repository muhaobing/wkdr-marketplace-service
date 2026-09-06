<template>
  <div class="user-benefits">
    <div class="page-header">
      <h1>用户权益</h1>
      <p class="page-desc">查询商城用户，调整积分或赠送商品（零元订单自动履约）</p>
    </div>

    <div class="card lookup-section">
      <h2 class="section-title">用户查询</h2>
      <div class="lookup-grid">
        <div class="field">
          <label>商城用户 ID</label>
          <input v-model="lookup.user_id" type="text" placeholder="例如 10001" @keyup.enter="handleLookup">
        </div>
        <div class="field">
          <label>手机号</label>
          <input v-model="lookup.tel_no" type="text" maxlength="11" placeholder="11 位手机号" @keyup.enter="handleLookup">
        </div>
        <div class="field">
          <label>企业 ID（企业用户选填，个人填 0）</label>
          <input v-model="lookup.company_id" type="text" placeholder="0 表示个人" @keyup.enter="handleLookup">
        </div>
        <div class="field">
          <label>业务平台 biz_code</label>
          <input v-model="lookup.biz_code" type="text" placeholder="如 LawMind_ToC" @keyup.enter="handleLookup">
        </div>
        <div class="field">
          <label>业务用户 biz_user_id</label>
          <input v-model="lookup.biz_user_id" type="text" placeholder="LawSharp 用户 ID" @keyup.enter="handleLookup">
        </div>
        <div class="field actions">
          <button class="btn btn-primary" :disabled="lookupLoading" @click="handleLookup">
            {{ lookupLoading ? '查询中…' : '查询用户' }}
          </button>
        </div>
      </div>

      <div v-if="userProfile" class="user-summary">
        <div class="summary-item"><span class="label">商城用户 ID</span><span>{{ userProfile.user.id }}</span></div>
        <div class="summary-item"><span class="label">手机号</span><span>{{ userProfile.user.tel_no || '--' }}</span></div>
        <div class="summary-item"><span class="label">企业 ID</span><span>{{ userProfile.user.company_id || 0 }}</span></div>
        <div class="summary-item"><span class="label">积分账户</span><span>{{ ecoinAccountLabel }}</span></div>
        <div class="summary-item highlight">
          <span class="label">当前积分</span>
          <span class="balance">{{ Number(userProfile.ecoin?.available_stock || 0).toFixed(2) }}</span>
        </div>
        <div v-if="userProfile.bindings?.length" class="bindings">
          <span class="label">业务绑定</span>
          <div class="binding-tags">
            <span v-for="b in userProfile.bindings" :key="`${b.biz_code}-${b.biz_user_id}`" class="tag">
              {{ b.biz_code }} · {{ b.biz_user_id }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="userProfile" class="ops-tabs card">
      <div class="tab-bar">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="tab-btn"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >{{ tab.label }}</button>
      </div>

      <div v-show="activeTab === 'adjust'" class="tab-panel">
        <div class="mode-row">
          <label v-for="m in adjustModes" :key="m.value" class="mode-option">
            <input v-model="adjustForm.mode" type="radio" :value="m.value">
            <span>{{ m.label }}</span>
          </label>
        </div>
        <div v-if="adjustForm.mode === 'set'" class="form-row">
          <label>目标积分余额</label>
          <input v-model.number="adjustForm.target_balance" type="number" min="0" step="0.01">
        </div>
        <div v-else class="form-row">
          <label>{{ adjustForm.mode === 'add' ? '增加' : '扣减' }}积分数量</label>
          <input v-model.number="adjustForm.amount" type="number" min="0.01" step="0.01">
        </div>
        <div class="form-row">
          <label>备注（选填）</label>
          <input v-model="adjustForm.description" type="text" placeholder="写入流水说明">
        </div>
        <button class="btn btn-primary" :disabled="adjustSubmitting" @click="submitAdjust">
          {{ adjustSubmitting ? '提交中…' : '确认调整' }}
        </button>
      </div>

      <div v-show="activeTab === 'gift_ecoin'" class="tab-panel">
        <div class="form-row">
          <label>赠送积分数量</label>
          <input v-model.number="giftEcoinForm.amount" type="number" min="0.01" step="0.01">
        </div>
        <div class="form-row">
          <label>备注（选填）</label>
          <input v-model="giftEcoinForm.description" type="text" placeholder="如：活动补偿">
        </div>
        <button class="btn btn-primary" :disabled="giftEcoinSubmitting" @click="submitGiftEcoin">
          {{ giftEcoinSubmitting ? '提交中…' : '确认赠送积分' }}
        </button>
      </div>

      <div v-show="activeTab === 'gift_sku'" class="tab-panel">
        <div class="form-row">
          <label>商品</label>
          <OpsSelect
            v-model="giftSkuForm.sku_id"
            :options="skuOptions"
            variant="filter"
          />
        </div>
        <div class="form-row">
          <label>数量</label>
          <input v-model.number="giftSkuForm.quantity" type="number" min="1" step="1">
        </div>
        <div class="form-row">
          <label>备注（选填）</label>
          <input v-model="giftSkuForm.remark" type="text" placeholder="赠送原因">
        </div>
        <button class="btn btn-primary" :disabled="giftSkuSubmitting" @click="submitGiftSku">
          {{ giftSkuSubmitting ? '提交中…' : '确认赠送商品' }}
        </button>
      </div>

      <div v-show="activeTab === 'gift_membership'" class="tab-panel">
        <div class="form-row">
          <label>会员档位</label>
          <OpsSelect
            v-model="giftMembershipForm.vip_role"
            :options="membershipRoleOptions"
            variant="filter"
          />
        </div>
        <div class="form-row">
          <label>时长</label>
          <div class="mode-row">
            <label class="mode-option">
              <input v-model="giftMembershipForm.period" type="radio" value="monthly">
              <span>包月</span>
            </label>
            <label class="mode-option">
              <input v-model="giftMembershipForm.period" type="radio" value="yearly">
              <span>包年</span>
            </label>
          </div>
        </div>
        <div class="form-row">
          <label>备注（选填）</label>
          <input v-model="giftMembershipForm.remark" type="text" placeholder="赠送原因">
        </div>
        <button class="btn btn-primary" :disabled="giftMembershipSubmitting" @click="submitGiftMembership">
          {{ giftMembershipSubmitting ? '提交中…' : '确认赠送会员' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import OpsSelect from '../../components/OpsSelect.vue'
import { opsSkuApi, opsUserBenefitsApi } from '../../api/index.js'
import { toast } from '../../utils/toast.js'

const lookup = reactive({
  user_id: '',
  tel_no: '',
  company_id: '',
  biz_code: '',
  biz_user_id: '',
})

const tabs = [
  { key: 'adjust', label: '调整积分' },
  { key: 'gift_ecoin', label: '赠送积分' },
  { key: 'gift_sku', label: '赠送商品' },
  { key: 'gift_membership', label: '赠送会员' },
]
const activeTab = ref('adjust')

const adjustModes = [
  { value: 'set', label: '设为指定余额' },
  { value: 'add', label: '增加积分' },
  { value: 'deduct', label: '扣减积分' },
]

const lookupLoading = ref(false)
const userProfile = ref(null)

const adjustForm = reactive({
  mode: 'set',
  target_balance: 0,
  amount: 0,
  description: '',
})
const adjustSubmitting = ref(false)

const giftEcoinForm = reactive({
  amount: 100,
  description: '',
})
const giftEcoinSubmitting = ref(false)

const giftSkuForm = reactive({
  sku_id: '',
  quantity: 1,
  remark: '',
})
const giftSkuSubmitting = ref(false)
const skuOptions = ref([{ value: '', label: '请选择商品' }])

const giftMembershipForm = reactive({
  vip_role: 'VIP_PRO',
  period: 'monthly',
  remark: '',
})
const giftMembershipSubmitting = ref(false)
const membershipRoleOptions = [
  { value: 'VIP_PRO', label: '个人专业版 VIP_PRO' },
  { value: 'VIP_MAX', label: '个人旗舰版 VIP_MAX' },
  { value: 'ENTERPRISE_BASIC', label: '企业基础版' },
  { value: 'ENTERPRISE_STANDARD', label: '企业标准版' },
  { value: 'ENTERPRISE_PRO', label: '企业专业版' },
  { value: 'ENTERPRISE_FLAGSHIP', label: '企业旗舰版' },
]

const ecoinAccountLabel = computed(() => {
  if (!userProfile.value?.ecoin) return '--'
  return userProfile.value.ecoin.account_type === 'company' ? '企业共享积分' : '个人积分'
})

function buildLookupParams() {
  const params = {}
  if (lookup.user_id.trim()) params.user_id = lookup.user_id.trim()
  if (lookup.tel_no.trim()) {
    params.tel_no = lookup.tel_no.trim()
    params.company_id = lookup.company_id.trim() !== '' ? lookup.company_id.trim() : '0'
  }
  if (lookup.biz_code.trim()) params.biz_code = lookup.biz_code.trim()
  if (lookup.biz_user_id.trim()) params.biz_user_id = lookup.biz_user_id.trim()
  return params
}

async function handleLookup() {
  const params = buildLookupParams()
  const hasUserId = !!params.user_id
  const hasPhone = !!params.tel_no
  const hasBiz = !!(params.biz_code && params.biz_user_id)
  if (!hasUserId && !hasPhone && !hasBiz) {
    toast.warning('请填写用户 ID、手机号+企业 ID，或 biz_code+biz_user_id')
    return
  }
  lookupLoading.value = true
  try {
    userProfile.value = await opsUserBenefitsApi.lookup(params)
    adjustForm.target_balance = Number(userProfile.value?.ecoin?.available_stock || 0)
    toast.success('用户查询成功')
  } catch (error) {
    userProfile.value = null
    toast.error(error.message || '查询失败')
  } finally {
    lookupLoading.value = false
  }
}

async function refreshProfile() {
  if (!userProfile.value?.user?.id) return
  lookup.user_id = String(userProfile.value.user.id)
  lookup.tel_no = userProfile.value.user.tel_no || ''
  lookup.company_id = String(userProfile.value.user.company_id || 0)
  await handleLookup()
}

async function submitAdjust() {
  if (!userProfile.value?.user?.id) return
  adjustSubmitting.value = true
  try {
    const payload = {
      user_id: userProfile.value.user.id,
      mode: adjustForm.mode,
      description: adjustForm.description.trim(),
    }
    if (adjustForm.mode === 'set') {
      payload.target_balance = Number(adjustForm.target_balance)
    } else {
      payload.amount = Number(adjustForm.amount)
    }
    const result = await opsUserBenefitsApi.adjustEcoin(payload)
    toast.success(`调整成功：${Number(result.before_balance).toFixed(2)} → ${Number(result.after_balance).toFixed(2)}`)
    await refreshProfile()
  } catch (error) {
    toast.error(error.message || '调整失败')
  } finally {
    adjustSubmitting.value = false
  }
}

async function submitGiftEcoin() {
  if (!userProfile.value?.user?.id) return
  if (!giftEcoinForm.amount || giftEcoinForm.amount <= 0) {
    toast.warning('请输入大于 0 的积分数量')
    return
  }
  giftEcoinSubmitting.value = true
  try {
    const result = await opsUserBenefitsApi.giftEcoin({
      user_id: userProfile.value.user.id,
      amount: Number(giftEcoinForm.amount),
      description: giftEcoinForm.description.trim(),
    })
    toast.success(`赠送成功：+${Number(result.delta).toFixed(2)}，当前 ${Number(result.after_balance).toFixed(2)}`)
    await refreshProfile()
  } catch (error) {
    toast.error(error.message || '赠送失败')
  } finally {
    giftEcoinSubmitting.value = false
  }
}

async function submitGiftSku() {
  if (!userProfile.value?.user?.id) return
  if (!giftSkuForm.sku_id) {
    toast.warning('请选择商品')
    return
  }
  giftSkuSubmitting.value = true
  try {
    const order = await opsUserBenefitsApi.giftSku({
      user_id: userProfile.value.user.id,
      sku_id: Number(giftSkuForm.sku_id),
      quantity: Number(giftSkuForm.quantity) || 1,
      remark: giftSkuForm.remark.trim(),
    })
    toast.success(`赠送成功，订单号 ${order.order_no}`)
  } catch (error) {
    toast.error(error.message || '赠送商品失败')
  } finally {
    giftSkuSubmitting.value = false
  }
}

async function submitGiftMembership() {
  if (!userProfile.value?.user?.id) return
  if (!giftMembershipForm.vip_role) {
    toast.warning('请选择会员档位')
    return
  }
  giftMembershipSubmitting.value = true
  try {
    const result = await opsUserBenefitsApi.giftMembership({
      user_id: String(userProfile.value.user.id),
      vip_role: giftMembershipForm.vip_role,
      period: giftMembershipForm.period,
      remark: giftMembershipForm.remark.trim(),
    })
    toast.success(`赠送成功：${result.sku_name || ''}，订单号 ${result.order_no || '--'}`)
  } catch (error) {
    toast.error(error.message || '赠送会员失败')
  } finally {
    giftMembershipSubmitting.value = false
  }
}

async function loadSkuOptions() {
  try {
    const resp = await opsSkuApi.list({ limit: 200, offset: 0, status: 1 })
    const list = resp?.list || []
    skuOptions.value = [
      { value: '', label: '请选择商品' },
      ...list.map(s => ({
        value: String(s.id),
        label: `[${s.biz_code}] ${s.sku_name}（¥${Number(s.cost || 0).toFixed(2)}）`,
      })),
    ]
  } catch {
    skuOptions.value = [{ value: '', label: '商品加载失败' }]
  }
}

onMounted(loadSkuOptions)
</script>

<style scoped>
.user-benefits {
  padding: 24px 28px 40px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0 0 8px;
  font-size: 24px;
  color: #0f172a;
}

.page-desc {
  margin: 0;
  color: #64748b;
  font-size: 14px;
}

.card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  padding: 20px 24px;
  margin-bottom: 20px;
}

.section-title {
  margin: 0 0 16px;
  font-size: 16px;
  color: #1e293b;
}

.lookup-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 14px 16px;
  align-items: end;
}

.field label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  color: #64748b;
}

.field input {
  width: 100%;
  height: 40px;
  padding: 0 12px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 14px;
  box-sizing: border-box;
}

.field.actions {
  display: flex;
  align-items: flex-end;
}

.user-summary {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px dashed #e2e8f0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px 20px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 14px;
  color: #0f172a;
}

.summary-item .label,
.bindings .label {
  font-size: 12px;
  color: #94a3b8;
}

.summary-item.highlight .balance {
  font-size: 22px;
  font-weight: 700;
  color: #4f46e5;
}

.bindings {
  grid-column: 1 / -1;
}

.binding-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 6px;
}

.tag {
  padding: 4px 10px;
  border-radius: 999px;
  background: #eef2ff;
  color: #4338ca;
  font-size: 12px;
}

.tab-bar {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 12px;
}

.tab-btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 14px;
  cursor: pointer;
}

.tab-btn.active {
  background: #eef2ff;
  color: #4338ca;
  font-weight: 600;
}

.tab-panel {
  max-width: 480px;
}

.mode-row {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.mode-option {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #334155;
}

.form-row {
  margin-bottom: 14px;
}

.form-row label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  color: #64748b;
}

.form-row input {
  width: 100%;
  height: 40px;
  padding: 0 12px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 14px;
  box-sizing: border-box;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  padding: 0 18px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.btn-primary {
  background: #4f46e5;
  color: #fff;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
