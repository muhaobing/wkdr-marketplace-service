import axios from 'axios'

const api = axios.create({
  baseURL: '/marketplace',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加 token
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    const { data } = response
    if (data.retcode === 0) {
      return data.data
    }
    return Promise.reject(new Error(data.message || '请求失败'))
  },
  error => {
    // 401 未授权，跳转到登录页
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

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
  // 创建订单
  create(data) {
    return api.post('/orders', data)
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

// 积分相关 API
export const ecoinApi = {
  // 获取用户积分
  get(userId) {
    return api.get('/ecoin', { params: { user_id: userId } })
  }
}

// 支付方式 API
export const paymentApi = {
  // 获取支付方式列表
  methods() {
    return api.get('/payment-methods')
  }
}

export default api
