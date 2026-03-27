import axios from 'axios'

const api = axios.create({
  baseURL: '/marketplace',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 运营 API 实例
const opsApi = axios.create({
  baseURL: '/ops',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加 token
const requestInterceptor = config => {
  const token = localStorage.getItem('token')
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

const responseErrorHandler = error => {
  // 401 未授权，跳转到登录页
  if (error.response && error.response.status === 401) {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    window.location.href = '/login'
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
  // 登录（电话号码/邮箱）
  login(data) {
    return api.post('/login', data)
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
