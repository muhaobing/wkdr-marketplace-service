<template>
  <div class="sku-management">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1>商品管理</h1>
      <button class="btn btn-primary" @click="openCreateModal">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"/>
          <line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
        新增商品
      </button>
    </div>

    <!-- 搜索筛选 -->
    <div class="filter-section card">
      <div class="filter-row">
        <div class="filter-item">
          <label>商品名称</label>
          <input 
            type="text" 
            v-model="filters.sku_name" 
            placeholder="输入商品名称搜索"
            @keyup.enter="handleSearch"
          >
        </div>
        <div class="filter-item">
          <label>业务域</label>
          <input 
            type="text" 
            v-model="filters.biz_code" 
            placeholder="输入业务域代码"
            @keyup.enter="handleSearch"
          >
        </div>
        <div class="filter-item">
          <label>上架状态</label>
          <select v-model="filters.status">
            <option value="">全部</option>
            <option value="0">待上架</option>
            <option value="1">已上架</option>
            <option value="2">已下架</option>
          </select>
        </div>
        <div class="filter-actions">
          <button class="btn btn-primary" @click="handleSearch">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"/>
              <line x1="21" y1="21" x2="16.65" y2="16.65"/>
            </svg>
            搜索
          </button>
          <button class="btn btn-secondary" @click="handleReset">重置</button>
        </div>
      </div>
    </div>

    <!-- 商品列表 -->
    <div class="table-section card">
      <div v-if="loading" class="loading-state">
        <p>加载中...</p>
      </div>
      
      <table v-else class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>商品编码</th>
            <th>商品名称</th>
            <th>业务域</th>
            <th>价格</th>
            <th>多选</th>
            <th>状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="skuList.length === 0">
            <td colspan="9" class="empty-row">暂无数据</td>
          </tr>
          <tr v-for="sku in skuList" :key="sku.id">
            <td>{{ sku.id }}</td>
            <td>{{ sku.sku_code }}</td>
            <td>
              <div class="sku-info">
                <img v-if="sku.sku_avatar" :src="sku.sku_avatar" class="sku-avatar">
                <div v-else class="sku-avatar-placeholder"></div>
                <span>{{ sku.sku_name }}</span>
              </div>
            </td>
            <td>{{ sku.biz_code }}</td>
            <td><span class="sku-price-rmb">¥{{ sku.cost?.toFixed(2) }}</span> <span class="ecoin-price">({{ toEcoin(sku.cost) }} 积分)</span></td>
            <td>
              <span :class="sku.multi_select === 1 ? 'badge-success' : 'badge-default'">
                {{ sku.multi_select === 1 ? '支持' : '不支持' }}
              </span>
            </td>
            <td>
              <label class="toggle-switch" :class="{ disabled: sku.toggling }">
                <input 
                  type="checkbox" 
                  :checked="sku.sku_status === 1"
                  @change="handleToggleStatus(sku)"
                  :disabled="sku.toggling"
                >
                <span class="toggle-slider"></span>
                <span class="toggle-label">{{ sku.sku_status === 1 ? '已上架' : '未上架' }}</span>
              </label>
            </td>
            <td>{{ formatTime(sku.ctime) }}</td>
            <td>
              <div class="action-buttons">
                <button class="action-btn" @click="openEditModal(sku)" title="编辑">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                  </svg>
                </button>
                <button 
                  v-if="sku.sku_status !== 1" 
                  class="action-btn danger" 
                  @click="handleDelete(sku)" 
                  title="删除"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"/>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页 -->
      <div class="pagination" v-if="total > pageSize">
        <button 
          class="page-btn" 
          :disabled="currentPage <= 1" 
          @click="changePage(currentPage - 1)"
        >
          上一页
        </button>
        <span class="page-info">
          第 {{ currentPage }} 页 / 共 {{ totalPages }} 页 ({{ total }} 条)
        </span>
        <button 
          class="page-btn" 
          :disabled="currentPage >= totalPages" 
          @click="changePage(currentPage + 1)"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content card">
        <div class="modal-header">
          <h3>{{ isEditing ? '编辑商品' : '新增商品' }}</h3>
          <button class="close-btn" @click="closeModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        
        <div class="modal-body">
          <div class="form-group">
            <label>商品编码 <span class="required">*</span></label>
            <input 
              type="text" 
              v-model="formData.sku_code" 
              placeholder="请输入商品编码"
              :disabled="isEditing"
            >
          </div>
          
          <div class="form-group">
            <label>商品名称 <span class="required">*</span></label>
            <input 
              type="text" 
              v-model="formData.sku_name" 
              placeholder="请输入商品名称"
            >
          </div>
          
          <div class="form-group">
            <label>业务域 <span class="required">*</span></label>
            <input 
              type="text" 
              v-model="formData.biz_code" 
              placeholder="请输入业务域代码"
              :disabled="isEditing"
            >
          </div>
          
          <div class="form-group">
            <label>商品价格(元) <span class="required">*</span></label>
            <input 
              type="number" 
              v-model.number="formData.cost" 
              placeholder="请输入商品价格（人民币）"
              min="0"
              step="0.01"
            >
            <div v-if="formData.cost > 0 && ecoinUnitPrice > 0" class="form-hint">
              ≈ {{ (formData.cost / ecoinUnitPrice).toFixed(2) }} 积分
            </div>
          </div>
          
          <div class="form-group">
            <label>商品图片</label>
            <div class="image-upload">
              <div v-if="formData.sku_avatar" class="image-preview">
                <img :src="formData.sku_avatar" alt="商品图片">
                <button type="button" class="image-remove-btn" @click="formData.sku_avatar = ''">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </div>
              <label v-else class="image-upload-trigger" for="sku-image-input">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <polyline points="21 15 16 10 5 21"/>
                </svg>
                <span>点击上传图片 <span class="upload-hint">(最大 800×800px)</span></span>
              </label>
              <input 
                id="sku-image-input"
                type="file" 
                accept="image/*"
                class="image-file-input"
                @change="handleImageUpload"
              >
            </div>
          </div>
          
          <div class="form-group">
            <label>商品描述</label>
            <textarea 
              v-model="formData.sku_desc" 
              placeholder="请输入商品描述"
              rows="4"
            ></textarea>
          </div>

          <div class="form-group">
            <label>是否支持多选</label>
            <div class="toggle-row">
              <label class="form-toggle">
                <input type="checkbox" v-model="multiSelectChecked">
                <span class="form-toggle-track">
                  <span class="form-toggle-thumb"></span>
                </span>
              </label>
              <span class="form-toggle-text" :class="{ active: multiSelectChecked }">
                {{ multiSelectChecked ? '支持' : '不支持' }}
              </span>
            </div>
          </div>
          
          <div class="form-group">
            <label>履约回调接口</label>
            <input 
              type="text" 
              v-model="formData.delivery_method" 
              placeholder="请输入履约回调接口URL"
            >
          </div>
        </div>
        
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="handleSubmit" :disabled="submitting">
            {{ submitting ? '提交中...' : '确定' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { opsSkuApi, ecoinApi } from '../../api'
import { toast, confirm } from '../../utils/toast'

// 列表数据
const skuList = ref([])
const loading = ref(false)
const total = ref(0)
const pageSize = 10
const currentPage = ref(1)

// 筛选条件
const filters = reactive({
  sku_name: '',
  biz_code: '',
  status: ''
})

// 弹窗相关
const showModal = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref(null)

const formData = reactive({
  sku_code: '',
  sku_name: '',
  biz_code: '',
  cost: '',
  sku_avatar: '',
  sku_desc: '',
  delivery_method: '',
  multi_select: 0
})

const MAX_IMAGE_WIDTH = 800
const MAX_IMAGE_HEIGHT = 800
const MAX_IMAGE_SIZE = 500 * 1024

function handleImageUpload(event) {
  const file = event.target.files[0]
  if (!file) return
  if (!file.type.startsWith('image/')) {
    toast.warning('请选择图片文件')
    event.target.value = ''
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    toast.warning('原始图片不能超过 5MB')
    event.target.value = ''
    return
  }
  const reader = new FileReader()
  reader.onload = (e) => {
    const img = new Image()
    img.onload = () => {
      let { width, height } = img
      if (width > MAX_IMAGE_WIDTH || height > MAX_IMAGE_HEIGHT) {
        const ratio = Math.min(MAX_IMAGE_WIDTH / width, MAX_IMAGE_HEIGHT / height)
        width = Math.round(width * ratio)
        height = Math.round(height * ratio)
      }
      const canvas = document.createElement('canvas')
      canvas.width = width
      canvas.height = height
      const ctx = canvas.getContext('2d')
      ctx.drawImage(img, 0, 0, width, height)
      let quality = 0.85
      let result = canvas.toDataURL('image/jpeg', quality)
      while (result.length * 0.75 > MAX_IMAGE_SIZE && quality > 0.1) {
        quality -= 0.1
        result = canvas.toDataURL('image/jpeg', quality)
      }
      if (result.length * 0.75 > MAX_IMAGE_SIZE) {
        toast.warning('图片压缩后仍超过 500KB，请选择更小的图片')
        return
      }
      formData.sku_avatar = result
    }
    img.src = e.target.result
  }
  reader.readAsDataURL(file)
  event.target.value = ''
}

const multiSelectChecked = computed({
  get: () => formData.multi_select === 1,
  set: (val) => { formData.multi_select = val ? 1 : 0 }
})

const ecoinUnitPrice = ref(0)

function toEcoin(cost) {
  if (!cost || !ecoinUnitPrice.value || ecoinUnitPrice.value <= 0) return '--'
  return (cost / ecoinUnitPrice.value).toFixed(2)
}

// 计算总页数
const totalPages = computed(() => Math.ceil(total.value / pageSize))

// 获取商品列表
async function fetchSkuList() {
  loading.value = true
  try {
    const params = {
      offset: (currentPage.value - 1) * pageSize,
      limit: pageSize
    }
    if (filters.sku_name) params.sku_name = filters.sku_name
    if (filters.biz_code) params.biz_code = filters.biz_code
    if (filters.status !== '') params.status = parseInt(filters.status)

    const res = await opsSkuApi.list(params)
    skuList.value = res?.list || []
    total.value = res?.total || 0
  } catch (error) {
    console.error('获取商品列表失败:', error)
    // Mock 数据
    skuList.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 搜索
function handleSearch() {
  currentPage.value = 1
  fetchSkuList()
}

// 重置
function handleReset() {
  filters.sku_name = ''
  filters.biz_code = ''
  filters.status = ''
  currentPage.value = 1
  fetchSkuList()
}

// 翻页
function changePage(page) {
  currentPage.value = page
  fetchSkuList()
}

// 打开新增弹窗
function openCreateModal() {
  isEditing.value = false
  editingId.value = null
  resetFormData()
  showModal.value = true
}

// 打开编辑弹窗
function openEditModal(sku) {
  isEditing.value = true
  editingId.value = sku.id
  formData.sku_code = sku.sku_code
  formData.sku_name = sku.sku_name
  formData.biz_code = sku.biz_code
  formData.cost = sku.cost
  formData.sku_avatar = sku.sku_avatar || ''
  formData.sku_desc = sku.sku_desc || ''
  formData.delivery_method = sku.delivery_method || ''
  formData.multi_select = sku.multi_select || 0
  showModal.value = true
}

// 关闭弹窗
function closeModal() {
  showModal.value = false
  resetFormData()
}

// 重置表单
function resetFormData() {
  formData.sku_code = ''
  formData.sku_name = ''
  formData.biz_code = ''
  formData.cost = ''
  formData.sku_avatar = ''
  formData.sku_desc = ''
  formData.delivery_method = ''
  formData.multi_select = 0
}

// 提交表单
async function handleSubmit() {
  // 验证
  if (!formData.sku_code || !formData.sku_name || !formData.biz_code || formData.cost === '') {
    toast.warning('请填写必填字段')
    return
  }

  submitting.value = true
  try {
    const data = {
      sku_code: formData.sku_code,
      sku_name: formData.sku_name,
      biz_code: formData.biz_code,
      cost: parseFloat(formData.cost),
      sku_avatar: formData.sku_avatar,
      sku_desc: formData.sku_desc,
      delivery_method: formData.delivery_method,
      multi_select: formData.multi_select
    }

    if (isEditing.value) {
      data.id = editingId.value
      await opsSkuApi.edit(data)
      toast.success('编辑成功')
    } else {
      await opsSkuApi.create(data)
      toast.success('创建成功')
    }

    closeModal()
    fetchSkuList()
  } catch (error) {
    toast.error('操作失败: ' + error.message)
  } finally {
    submitting.value = false
  }
}

// 切换上架状态
async function handleToggleStatus(sku) {
  sku.toggling = true
  try {
    if (sku.sku_status === 1) {
      // 下架
      await opsSkuApi.delisting(sku.id)
      sku.sku_status = 0
    } else {
      // 上架
      await opsSkuApi.listing(sku.id)
      sku.sku_status = 1
    }
  } catch (error) {
    toast.error('操作失败: ' + error.message)
    fetchSkuList()
  } finally {
    sku.toggling = false
  }
}

// 上架（保留备用）
async function handleListing(sku) {
  try {
    await opsSkuApi.listing(sku.id)
    fetchSkuList()
  } catch (error) {
    toast.error('上架失败: ' + error.message)
  }
}

// 下架（保留备用）
async function handleDelisting(sku) {
  try {
    await opsSkuApi.delisting(sku.id)
    toast.success('下架成功')
    fetchSkuList()
  } catch (error) {
    toast.error('下架失败: ' + error.message)
  }
}

// 删除
async function handleDelete(sku) {
  if (!await confirm(`确定要删除商品 "${sku.sku_name}" 吗？此操作不可恢复！`)) return
  
  try {
    await opsSkuApi.delete(sku.id)
    toast.success('删除成功')
    fetchSkuList()
  } catch (error) {
    toast.error('删除失败: ' + error.message)
  }
}

// 获取状态样式
function getStatusClass(status) {
  switch (status) {
    case 0: return 'status-pending'
    case 1: return 'status-online'
    case 2: return 'status-offline'
    default: return ''
  }
}

// 获取状态文字
function getStatusText(status) {
  switch (status) {
    case 0: return '待上架'
    case 1: return '已上架'
    case 2: return '已下架'
    default: return '未知'
  }
}

// 格式化时间
function formatTime(timestamp) {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(async () => {
  fetchSkuList()
  try {
    const cfg = await ecoinApi.getRechargeConfig()
    ecoinUnitPrice.value = cfg.unit_price || 0
  } catch (e) { /* ignore */ }
})
</script>

<style scoped>
.sku-management {
  padding: 32px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28px;
}

.page-header h1 {
  font-size: 26px;
  font-weight: 700;
  color: var(--gray-800);
  letter-spacing: -0.02em;
}

.page-header .btn svg {
  width: 18px;
  height: 18px;
  margin-right: 6px;
}

/* 筛选区域 */
.filter-section {
  padding: 20px 24px;
  margin-bottom: 20px;
}

.filter-row {
  display: flex;
  gap: 16px;
  align-items: flex-end;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-item label {
  font-size: 12px;
  color: var(--gray-400);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.filter-item input,
.filter-item select {
  height: 38px;
  padding: 0 14px;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  font-size: 14px;
  min-width: 180px;
  transition: all 0.2s;
}

.filter-item input:focus,
.filter-item select:focus {
  border-color: var(--accent);
  outline: none;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.08);
}

.filter-actions {
  display: flex;
  gap: 8px;
}

.filter-actions .btn {
  height: 38px;
  padding: 0 18px;
}

.filter-actions .btn svg {
  width: 16px;
  height: 16px;
  margin-right: 4px;
}

/* 表格区域 */
.table-section {
  padding: 0;
  overflow: hidden;
}

.loading-state {
  padding: 60px;
  text-align: center;
  color: var(--gray-400);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  background-color: var(--gray-50);
  padding: 12px 16px;
  text-align: left;
  font-size: 12px;
  font-weight: 700;
  color: var(--gray-400);
  border-bottom: 1px solid var(--gray-200);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.data-table td .sku-price-rmb {
  color: #b91c1c;
  font-weight: 700;
}

.data-table td .ecoin-price {
  font-size: 12px;
  color: #f59e0b;
  font-weight: 600;
}

.data-table td {
  padding: 14px 16px;
  font-size: 14px;
  color: var(--gray-700);
  border-bottom: 1px solid var(--gray-100);
}

.data-table tr:hover td {
  background-color: rgba(99, 102, 241, 0.02);
}

.empty-row {
  text-align: center;
  color: var(--gray-400);
  padding: 60px !important;
}

.sku-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sku-avatar {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: cover;
}

.sku-avatar-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: linear-gradient(135deg, #f1f5f9, #e2e8f0);
}

.status-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.status-pending {
  background-color: #fffbeb;
  color: #b45309;
}

.status-online {
  background-color: #ecfdf5;
  color: #047857;
}

.status-offline {
  background-color: #fef2f2;
  color: #b91c1c;
}

/* Toggle 开关 */
.toggle-switch {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.toggle-switch.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.toggle-switch input {
  display: none;
}

.toggle-slider {
  position: relative;
  width: 44px;
  height: 24px;
  background-color: var(--gray-300);
  border-radius: 12px;
  transition: all 0.25s;
}

.toggle-slider::before {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background-color: white;
  border-radius: 50%;
  transition: all 0.25s;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}

.toggle-switch input:checked + .toggle-slider {
  background-color: var(--success);
}

.toggle-switch input:checked + .toggle-slider::before {
  transform: translateX(20px);
}

.toggle-label {
  font-size: 13px;
  color: var(--gray-500);
  font-weight: 500;
}

.toggle-switch input:checked ~ .toggle-label {
  color: var(--success);
  font-weight: 600;
}

.badge-success {
  display: inline-block;
  padding: 3px 10px;
  font-size: 12px;
  border-radius: 6px;
  background-color: #ecfdf5;
  color: #047857;
  font-weight: 500;
}

.badge-default {
  display: inline-block;
  padding: 3px 10px;
  font-size: 12px;
  border-radius: 6px;
  background-color: var(--gray-100);
  color: var(--gray-500);
  font-weight: 500;
}

.action-buttons {
  display: flex;
  gap: 6px;
}

.action-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background-color: var(--gray-50);
  color: var(--gray-500);
  transition: all 0.15s;
  border: 1px solid var(--gray-200);
}

.action-btn:hover {
  background-color: var(--accent);
  border-color: var(--accent);
  color: white;
}

.action-btn.danger:hover {
  background-color: var(--danger);
  border-color: var(--danger);
}

.action-btn svg {
  width: 15px;
  height: 15px;
}

/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 20px;
  border-top: 1px solid var(--gray-100);
}

.page-btn {
  padding: 8px 18px;
  background-color: white;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--gray-600);
  transition: all 0.15s;
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

/* 弹窗 */
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

.modal-content {
  width: 560px;
  max-height: 90vh;
  overflow-y: auto;
  border-radius: 20px;
  box-shadow: var(--shadow-xl);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: none;
  color: var(--gray-400);
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
  margin-bottom: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--gray-500);
  letter-spacing: 0.02em;
}

.form-group .required {
  color: var(--danger);
}

.form-group input,
.form-group textarea,
.form-group select {
  width: 100%;
  padding: 11px 14px;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus {
  border-color: var(--accent);
  outline: none;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.08);
}

.form-group input:disabled {
  background-color: var(--gray-50);
  cursor: not-allowed;
}

.form-group textarea {
  resize: vertical;
}

.form-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--gray-400);
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
}

.form-toggle {
  position: relative;
  display: inline-block;
  width: 48px;
  height: 26px;
  cursor: pointer;
  flex-shrink: 0;
}

.form-toggle input {
  display: none;
}

.form-toggle-track {
  position: absolute;
  inset: 0;
  background-color: var(--gray-300);
  border-radius: 13px;
  transition: background-color 0.25s;
}

.form-toggle input:checked + .form-toggle-track {
  background-color: var(--success);
}

.form-toggle-thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 20px;
  height: 20px;
  background-color: white;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  transition: transform 0.25s;
}

.form-toggle input:checked + .form-toggle-track .form-toggle-thumb {
  transform: translateX(22px);
}

.form-toggle-text {
  font-size: 14px;
  color: var(--gray-500);
  transition: color 0.2s;
  user-select: none;
  font-weight: 500;
}

.form-toggle-text.active {
  color: var(--success);
  font-weight: 600;
}

.image-upload {
  position: relative;
}

.image-file-input {
  display: none;
}

.image-upload-trigger {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 120px;
  height: 120px;
  border: 2px dashed var(--gray-300);
  border-radius: 12px;
  cursor: pointer;
  color: var(--gray-400);
  transition: all 0.2s;
}

.image-upload-trigger:hover {
  border-color: var(--accent);
  color: var(--accent);
  background-color: rgba(99, 102, 241, 0.02);
}

.image-upload-trigger svg {
  width: 28px;
  height: 28px;
}

.image-upload-trigger span {
  font-size: 12px;
  font-weight: 500;
}

.upload-hint {
  color: var(--gray-400);
  font-size: 11px !important;
}

.image-preview {
  position: relative;
  display: inline-block;
  width: 120px;
  height: 120px;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--gray-200);
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-remove-btn {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.5);
  color: white;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  transition: background-color 0.15s;
}

.image-remove-btn:hover {
  background-color: rgba(239, 68, 68, 0.8);
}

.image-remove-btn svg {
  width: 14px;
  height: 14px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 20px 28px;
  border-top: 1px solid var(--gray-100);
}
</style>
