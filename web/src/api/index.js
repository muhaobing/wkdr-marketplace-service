import axios from 'axios'
import router from '../router/index.js'
import { STORAGE_TOKEN_KEY } from '../constants/storage.js'

// VITE_ENV=local：请求 /marketplace、/ops，由 Vite 代理到 localhost:10302，无 /market/api 前缀
// 否则：生产构建默认 https://lawmind.top/market/api；开发可用相对路径走 Vite 的 /market/api 代理（或设 VITE_API_ORIGIN）
const isLocal = import.meta.env.VITE_ENV === 'local'

function joinApiBase(suffix) {
  if (isLocal) {
    return suffix
  }
  const explicit = import.meta.env.VITE_API_ORIGIN
  if (explicit !== undefined && explicit !== '') {
    return `${String(explicit).replace(/\/$/, '')}/market/api${suffix}`
  }
  if (import.meta.env.PROD) {
    return `https://lawmind.top/market/api${suffix}`
  }
  return `/market/api${suffix}`
}

const api = axios.create({
  baseURL: joinApiBase('/marketplace'),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 运营 API 实例
const opsApi = axios.create({
  baseURL: joinApiBase('/ops'),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器：从 localStorage（marketplace_token）附加 Authorization；若为空则回退 Pinia（例如其它标签页清过 storage）
const requestInterceptor = async config => {
  let token = localStorage.getItem(STORAGE_TOKEN_KEY)
  if (!token) {
    try {
      const { useUserStore } = await import('../stores/user.js')
      const u = useUserStore()
      token = u.token
      if (token) {
        try {
          localStorage.setItem(STORAGE_TOKEN_KEY, token)
        } catch (_) {}
      }
    } catch (_) {}
  }
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
}

const requestErrorHandler = error => {
  return Promise.reject(error)
}

// 响应拦截器
const responseInterceptor = response => {
  const { data } = response
  if (data.retcode === 0) {
    return data.data
  }
  return Promise.reject(new Error(data.message || '请求失败'))
}

const responseErrorHandler = async error => {
  // 401 未授权：同步 Pinia 登出 + 跳转登录（与 localStorage 一致）
  if (error.response && error.response.status === 401) {
    try {
      const { useUserStore } = await import('../stores/user.js')
      useUserStore().logout()
    } catch (_) {
      localStorage.removeItem(STORAGE_TOKEN_KEY)
      localStorage.removeItem('user')
    }
    const q = { ...router.currentRoute.value.query }
    router.replace({ name: 'Login', query: q }).catch(() => {})
  }
  // 403 禁止访问
  if (error.response && error.response.status === 403) {
    return Promise.reject(new Error('没有权限访问'))
  }
  return Promise.reject(error)
}

// 应用拦截器到 marketplace api
api.interceptors.request.use(requestInterceptor, requestErrorHandler)
api.interceptors.response.use(responseInterceptor, responseErrorHandler)

// 应用拦截器到 ops api
opsApi.interceptors.request.use(requestInterceptor, requestErrorHandler)
opsApi.interceptors.response.use(responseInterceptor, responseErrorHandler)

// 用户认证相关 API
export const authApi = {
  // 登录（电话号码/邮箱 + 密钥）
  login(data) {
    return api.post('/login', data)
  },
  // 绑定业务平台账号（免登录，成功后返回 token，与登录一致）
  bind(data) {
    return api.post('/user/bind', data)
  },
  // 是否已有 biz_code + biz_user_id 绑定（免登录）
  checkBinding(params) {
    return api.get('/user/bind/check', { params })
  },
  /** 当前登录用户的业务平台绑定列表（需 Authorization） */
  listUserBindings() {
    return api.get('/user/bindings')
  },
  /** 修改登录密钥（需 Authorization） */
  changePassword(data) {
    return api.post('/user/password', data)
  },
  /** 更新手机号、邮箱（需 Authorization，服务端校验唯一性） */
  updateProfile(data) {
    return api.post('/user/profile', data)
  },
  /** 解除与某业务平台的绑定（需 Authorization） */
  unbindBiz(bizCode) {
    return api.post('/user/unbind', { biz_code: bizCode })
  }
}

// 商城元数据（免登录）
export const metaApi = {
  /** 业务平台 biz_code 枚举，供绑定页下拉 */
  listBizCodes() {
    return api.get('/biz_codes')
  }
}

// 商品相关 API
export const skuApi = {
  // 获取商品列表
  list(params) {
    return api.get('/skus', { params })
  },
  // 获取商品详情
  detail(id) {
    return api.get(`/skus/${id}`)
  }
}

// 订单相关 API
export const orderApi = {
  // 创建订单（直接下单）
  create(data) {
    return api.post('/checkout', data)
  },
  // 获取订单列表
  list(params) {
    return api.get('/orders', { params })
  },
  // 获取订单详情
  detail(orderNo) {
    return api.get(`/orders/${orderNo}`)
  },
  // 取消订单
  cancel(orderNo, reason = '') {
    return api.post(`/orders/${orderNo}/cancel`, { reason })
  },
  // 支付订单
  pay(orderNo, data) {
    return api.post(`/orders/${orderNo}/pay`, data)
  },
  // 同步订单状态
  sync(orderNo) {
    return api.post(`/orders/${orderNo}/sync`)
  }
}

// 购物车相关 API
export const cartApi = {
  // 添加商品到购物车
  add(data) {
    return api.post('/shopping_cart/add', data)
  },
  // 从购物车移除商品
  remove(data) {
    return api.post('/shopping_cart/remove', data)
  },
  // 更新购物车商品数量
  update(data) {
    return api.post('/shopping_cart/update', data)
  },
  // 清空购物车
  clear(data) {
    return api.post('/shopping_cart/clear', data)
  },
  // 获取购物车列表
  list(params) {
    return api.get('/shopping_cart/list', { params })
  },
  // 购物车下单
  checkout(data) {
    return api.post('/shopping_cart/checkout', data)
  }
}

// 积分相关 API
export const ecoinApi = {
  // 获取用户积分余额
  getBalance(userId) {
    return api.get('/ecoin/balance', { params: { user_id: userId } })
  },
  // 获取积分库存分组
  getStockGroups(userId) {
    return api.get('/ecoin/stock_groups', { params: { user_id: userId } })
  },
  // 获取积分流水列表
  getTransactions(params) {
    return api.get('/ecoin/transactions', { params })
  },
  // 获取充值配置
  getRechargeConfig() {
    return api.get('/ecoin/recharge_config')
  },
  // 积分充值
  recharge(data) {
    return api.post('/ecoin/recharge', data)
  }
}

// 支付方式 API
export const paymentApi = {
  // 获取支付方式列表
  methods() {
    return api.get('/payment-methods')
  }
}

// ==================== 运营相关 API ====================

// 运营端商品管理 API
export const opsSkuApi = {
  // 获取商品列表（支持多条件搜索）
  list(params) {
    return opsApi.get('/skus', { params })
  },
  // 获取商品详情
  detail(id) {
    return opsApi.get(`/skus/${id}`)
  },
  // 创建商品
  create(data) {
    return opsApi.post('/skus/create', data)
  },
  // 编辑商品
  edit(data) {
    return opsApi.post('/skus/edit', data)
  },
  // 删除商品
  delete(id) {
    return opsApi.post('/skus/delete', { id })
  },
  // 商品上架
  listing(id) {
    return opsApi.post('/skus/listing', { id })
  },
  // 商品下架
  delisting(id) {
    return opsApi.post('/skus/delisting', { id })
  }
}

// 运营端订单管理 API
export const opsOrderApi = {
  // 获取订单列表
  list(params) {
    return opsApi.get('/orders', { params })
  },
  // 获取订单详情
  detail(orderNo) {
    return opsApi.get(`/orders/${orderNo}`)
  },
  // 退款
  refund(orderNo, reason = '') {
    return opsApi.post('/orders/refund', { order_no: orderNo, reason })
  },
  // 履约
  fulfill(orderNo, bizUserId = '') {
    return opsApi.post('/orders/fulfill', { order_no: orderNo, biz_user_id: bizUserId })
  }
}

export default api
