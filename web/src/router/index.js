import { createRouter, createWebHistory } from 'vue-router'
import { toast } from '../utils/toast'
import { bindingListContains, parseBizQueryFromRoute } from '../utils/bizBindings.js'
import { STORAGE_TOKEN_KEY } from '../constants/storage.js'

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
router.beforeEach(async (to, from, next) => {
  const token = localStorage.getItem(STORAGE_TOKEN_KEY)
  const userStr = localStorage.getItem('user')
  const isLoggedIn = !!token

  const bizParams = parseBizQueryFromRoute(to.query)

  // LawMind 跳转带 biz_code、biz_user_id：已登录则校验当前账号是否绑定该业务身份
  if (bizParams && isLoggedIn) {
    const { useUserStore } = await import('../stores/user.js')
    const userStore = useUserStore()
    try {
      const bindings = await userStore.fetchBindingsRemote()
      if (!bindingListContains(bindings, bizParams.biz_code, bizParams.biz_user_id)) {
        userStore.logout()
        toast.error('当前账号与 LawMind 跳转参数不一致，请重新登录')
        next({
          name: 'Login',
          query: {
            ...to.query,
            redirect: to.path || '/'
          }
        })
        return
      }
    } catch (e) {
      console.error(e)
      userStore.logout()
      toast.error('无法校验业务绑定，请重新登录')
      next({
        name: 'Login',
        query: {
          ...to.query,
          redirect: to.path || '/'
        }
      })
      return
    }
  }

  // 如果页面需要登录但用户未登录
  if (to.meta.requiresAuth && !isLoggedIn) {
    // 业务参数（biz_code、biz_user_id 等）只放在顶层 query，由 ...to.query 带上
    // redirect 只用「路径」：若用 to.fullPath，会把 ?& 再塞进 redirect，易与顶层参数冲突或导致 URL 解析截断
    next({
      name: 'Login',
      query: {
        ...to.query,
        redirect: to.path || '/'
      }
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
