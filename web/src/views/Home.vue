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
          class="product-card"
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
            <div class="image-overlay">
              <button class="quick-cart-btn" @click.stop="addToCart(product)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="9" cy="21" r="1"/>
                  <circle cx="20" cy="21" r="1"/>
                  <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
                </svg>
                加入购物车
              </button>
            </div>
          </div>
          <div class="product-info">
            <h3 class="product-name">{{ product.sku_name }}</h3>
            <p class="product-desc">{{ product.sku_desc || '暂无描述' }}</p>
            <div class="product-footer">
              <div class="product-price">
                <span class="price-rmb">¥{{ product.cost.toFixed(2) }}</span>
                <span class="price-ecoin">{{ toEcoin(product.cost) }} 积分</span>
              </div>
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
import { toast } from '../utils/toast'

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

async function fetchProducts() {
  loading.value = true
  try {
    const res = await skuApi.list({ biz_code: 'marketplace', limit: 100 })
    products.value = res?.list || []
  } catch (error) {
    console.error('获取商品列表失败:', error)
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

function handleSearch() {}

function goToDetail(id) {
  router.push(`/product/${id}`)
}

function addToCart(product) {
  cartStore.addItem(product)
  toast.success('已加入购物车')
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
  padding-top: 8px;
}

.search-section {
  margin-bottom: 36px;
}

.search-box {
  position: relative;
  max-width: 520px;
  margin: 0 auto;
}

.search-icon {
  position: absolute;
  left: 18px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  color: var(--gray-400);
  pointer-events: none;
}

.search-box input {
  width: 100%;
  padding: 14px 20px 14px 50px;
  border: 1px solid var(--gray-200);
  border-radius: 14px;
  font-size: 15px;
  background-color: var(--white);
  transition: all 0.25s;
  box-shadow: var(--shadow-sm);
}

.search-box input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.08), var(--shadow-md);
}

.search-box input::placeholder {
  color: var(--gray-400);
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 24px;
}

.product-card {
  background: white;
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(0, 0, 0, 0.04);
  box-shadow: var(--shadow-sm);
}

.product-card:hover {
  transform: translateY(-6px);
  box-shadow: var(--shadow-xl);
  border-color: transparent;
}

.product-image {
  width: 100%;
  height: 200px;
  overflow: hidden;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  position: relative;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.4s ease;
}

.product-card:hover .product-image img {
  transform: scale(1.05);
}

.image-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.5) 0%, transparent 50%);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: 16px;
  opacity: 0;
  transition: opacity 0.3s;
}

.product-card:hover .image-overlay {
  opacity: 1;
}

.quick-cart-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: white;
  color: var(--gray-800);
  font-size: 13px;
  font-weight: 600;
  border-radius: 8px;
  transition: all 0.2s;
}

.quick-cart-btn:hover {
  background: var(--gray-100);
}

.quick-cart-btn svg {
  width: 16px;
  height: 16px;
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
  padding: 18px 20px 20px;
}

.product-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--gray-800);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  letter-spacing: -0.01em;
}

.product-desc {
  font-size: 13px;
  color: var(--gray-400);
  margin-bottom: 14px;
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
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.price-rmb {
  font-size: 20px;
  font-weight: 700;
  color: #b91c1c;
  letter-spacing: -0.02em;
}

.price-ecoin {
  font-size: 12px;
  font-weight: 600;
  color: #f59e0b;
}

@media (max-width: 640px) {
  .product-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .product-image {
    height: 150px;
  }

  .product-info {
    padding: 12px 14px 14px;
  }

  .price-rmb {
    font-size: 16px;
  }

  .image-overlay {
    display: none;
  }
}
</style>
