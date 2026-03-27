import { createRouter, createWebHistory } from 'vue-router'
import { toast } from '../utils/toast'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/product/:id',
    name: 'ProductDetail',
    component: () => import('../views/ProductDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/cart',
    name: 'Cart',
    component: () => import('../views/Cart.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/orders',
    name: 'Orders',
    component: () => import('../views/Orders.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/orders/:orderNo',
    name: 'OrderDetail',
    component: () => import('../views/OrderDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/ecoin',
    name: 'EcoinCenter',
    component: () => import('../views/EcoinCenter.vue'),
    meta: { requiresAuth: true }
  },
  // 运营中心路由
  {
    path: '/ops',
    component: () => import('../components/OpsLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        redirect: { name: 'OpsSkuManagement' }
      },
      {
        path: 'skus',
        name: 'OpsSkuManagement',
        component: () => import('../views/ops/SkuManagement.vue'),
        meta: { requiresAuth: true, requiresAdmin: true }
      },
      {
        path: 'data',
        name: 'OpsDataCenter',
        component: () => import('../views/ops/DataCenter.vue'),
        meta: { requiresAuth: true, requiresAdmin: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 路由守卫 - 检查登录状态和权限
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  const isLoggedIn = !!token

  // 如果页面需要登录但用户未登录
  if (to.meta.requiresAuth && !isLoggedIn) {
    next({
      name: 'Login',
      query: { redirect: to.fullPath }
    })
    return
  }

  // 如果用户已登录但访问登录页
  if (to.name === 'Login' && isLoggedIn) {
    next({ name: 'Home' })
    return
  }

  // 检查管理员权限
  if (to.meta.requiresAdmin && isLoggedIn) {
    try {
      const user = JSON.parse(userStr)
      if (user.role !== 1) {
        toast.error('没有访问权限')
        next({ name: 'Home' })
        return
      }
    } catch (e) {
      next({ name: 'Home' })
      return
    }
  }

  next()
})

export default router
