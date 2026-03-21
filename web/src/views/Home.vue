<template>
  <div class="home">
    <div class="container">
      <!-- 搜索栏 -->
      <div class="search-section">
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input 
            v-model="searchQuery" 
            type="text" 
            placeholder="搜索商品..."
            @input="handleSearch"
          >
        </div>
      </div>

      <!-- 商品列表 -->
      <div v-if="loading" class="loading"></div>

      <div v-else-if="filteredProducts.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
        </svg>
        <p>暂无商品</p>
      </div>

      <div v-else class="product-grid">
        <div 
          v-for="product in filteredProducts" 
          :key="product.id" 
          class="product-card card"
          @click="goToDetail(product.id)"
        >
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
          <div class="product-info">
            <h3 class="product-name">{{ product.sku_name }}</h3>
            <p class="product-desc">{{ product.sku_desc || '暂无描述' }}</p>
            <div class="product-footer">
              <span class="product-price">¥{{ product.cost.toFixed(2) }} <span class="ecoin-price">({{ toEcoin(product.cost) }} 积分)</span></span>
              <button class="btn btn-primary btn-sm" @click.stop="addToCart(product)">
                加入购物车
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { skuApi, ecoinApi } from '../api'
import { useCartStore } from '../stores/cart'

const router = useRouter()
const cartStore = useCartStore()

const products = ref([])
const loading = ref(false)
const searchQuery = ref('')
const ecoinUnitPrice = ref(0)

function toEcoin(cost) {
  if (!ecoinUnitPrice.value || ecoinUnitPrice.value <= 0) return '--'
  return (cost / ecoinUnitPrice.value).toFixed(2)
}

// 过滤后的商品列表
const filteredProducts = computed(() => {
  if (!searchQuery.value.trim()) {
    return products.value
  }
  const query = searchQuery.value.toLowerCase()
  return products.value.filter(p => 
    p.sku_name.toLowerCase().includes(query) ||
    (p.sku_desc && p.sku_desc.toLowerCase().includes(query))
  )
})

// 获取商品列表
async function fetchProducts() {
  loading.value = true
  try {
    const res = await skuApi.list({ biz_code: 'marketplace', limit: 100 })
    products.value = res?.list || []
  } catch (error) {
    console.error('获取商品列表失败:', error)
    // Mock 数据
    products.value = [
      { id: 1, sku_code: 'SKU001', sku_name: '虚拟商品A', sku_desc: '这是一个测试商品', cost: 100, sku_avatar: '' },
      { id: 2, sku_code: 'SKU002', sku_name: '虚拟商品B', sku_desc: '另一个测试商品', cost: 200, sku_avatar: '' },
      { id: 3, sku_code: 'SKU003', sku_name: '高级会员', sku_desc: '一个月高级会员权益', cost: 500, sku_avatar: '' },
      { id: 4, sku_code: 'SKU004', sku_name: '专属道具', sku_desc: '限量版专属道具', cost: 300, sku_avatar: '' },
    ]
  } finally {
    loading.value = false
  }
}

// 搜索处理
function handleSearch() {
  // 本地过滤，无需额外操作
}

// 跳转到商品详情
function goToDetail(id) {
  router.push(`/product/${id}`)
}

// 加入购物车
function addToCart(product) {
  cartStore.addItem(product)
  // 简单提示
  alert('已加入购物车')
}

onMounted(async () => {
  fetchProducts()
  try {
    const cfg = await ecoinApi.getRechargeConfig()
    ecoinUnitPrice.value = cfg.unit_price || 0
  } catch (e) { /* ignore */ }
})
</script>

<style scoped>
.home {
  padding-top: 20px;
}

.search-section {
  margin-bottom: 30px;
}

.search-box {
  position: relative;
  max-width: 500px;
  margin: 0 auto;
}

.search-icon {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  color: var(--gray-400);
}

.search-box input {
  width: 100%;
  padding: 14px 16px 14px 48px;
  border: 1px solid var(--gray-200);
  border-radius: 12px;
  font-size: 16px;
  background-color: var(--white);
  transition: all 0.2s;
}

.search-box input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(26, 54, 93, 0.1);
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 24px;
}

.product-card {
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.1);
}

.product-image {
  width: 100%;
  height: 180px;
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
  width: 48px;
  height: 48px;
}

.product-info {
  padding: 16px;
}

.product-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--gray-700);
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-desc {
  font-size: 13px;
  color: var(--gray-500);
  margin-bottom: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.product-price {
  font-size: 18px;
  font-weight: 600;
  color: #e53e3e;
}

.product-price .ecoin-price {
  font-size: 13px;
  font-weight: 400;
  color: #888;
}

.btn-sm {
  padding: 8px 14px;
  font-size: 13px;
}

@media (max-width: 640px) {
  .product-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .product-image {
    height: 140px;
  }

  .product-info {
    padding: 12px;
  }

  .product-price {
    font-size: 14px;
  }

  .btn-sm {
    padding: 6px 10px;
    font-size: 12px;
  }
}
</style>
