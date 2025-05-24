import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import PrintPage from '../views/PrintPage.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/login'
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('../views/UsersView.vue')
    },
    {
      path: '/roles',
      name: 'roles',
      component: () => import('../views/RolesView.vue')
    },
    {
      path: '/transfers',
      name: 'transfers',
      component: () => import('../views/TransfersView.vue')
    },
    {
      path: '/send',
      name: 'send-transfer',
      component: () => import('../views/SenderView.vue')
    },
    {
      path: '/receive',
      name: 'receive-transfer',
      component: () => import('../views/ReceiverView.vue')
    },
    {
      path: '/sell-bills',
      name: 'sell-bills',
      component: () => import('../views/SellBillsView.vue')
    },
    {
      path: '/transfers-list',
      name: 'transfers-list',
      component: () => import('../views/TransfersListView.vue'),
      meta: { requiresAdmin: true }
    },
    {
      path: '/cars',
      name: 'cars',
      component: () => import('../views/CarsView.vue')
    },
    {
      path: '/cars/stock',
      name: 'cars-stock',
      component: () => import('../views/CarsStock.vue')
    },
    {
      path: '/warehouses',
      name: 'warehouses',
      component: () => import('../views/WarehousesView.vue')
    },
    {
      path: '/print/:billId',
      name: 'print',
      component: PrintPage
    }
  ],
  scrollBehavior() {
    return { top: 0 }
  }
})

// Add navigation guard to handle page reloads
router.beforeEach((to, from, next) => {
  if (to.matched.length === 0) {
    next('/')
  } else {
    next()
  }
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const publicPages = ['/login']
  const authRequired = !publicPages.includes(to.path)
  const user = localStorage.getItem('user')

  if (authRequired && !user) {
    return next('/login')
  }

  if (to.path === '/login' && user) {
    return next('/dashboard')
  }

  next()
})

export default router
