import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import PrintPage from '../views/PrintPage.vue'
import SellBillPaymentsView from '../views/SellBillPaymentsView.vue'
import MoneyMovements from '../components/cashier/MoneyMovements.vue'
import TransfersInterTable from '../components/cashier/TransfersInterTable.vue'
import ChinaCash from '../components/cashier/ChinaCash.vue'
import BuyBillPaymentsView from '../views/BuyBillPaymentsView.vue'
import RatesView from '../views/RatesView.vue'

// Use root path since we're using a subdomain
const router = createRouter({
  history: createWebHistory('/mig/'),
  routes: [
    {
      path: '/',
      redirect: '/login',
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView,
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('../views/UsersView.vue'),
    },
    {
      path: '/roles',
      name: 'roles',
      component: () => import('../views/RolesView.vue'),
    },
    {
      path: '/transfers',
      name: 'transfers',
      component: () => import('../views/TransfersView.vue'),
    },
    {
      path: '/send',
      name: 'send-transfer',
      component: () => import('../views/SenderView.vue'),
    },
    {
      path: '/receive',
      name: 'receive-transfer',
      component: () => import('../views/ReceiverView.vue'),
    },
    {
      path: '/sell-bills',
      name: 'sell-bills',
      component: () => import('../views/SellBillsView.vue'),
    },
    {
      path: '/sell-bills/:id/payments',
      name: 'sell-bill-payments',
      component: SellBillPaymentsView,
      meta: { requiresAuth: true },
    },
    {
      path: '/buy-payments/:id',
      name: 'buy-payments',
      component: BuyBillPaymentsView,
      meta: { requiresAuth: true },
    },
    {
      path: '/transfers-list',
      name: 'transfers-list',
      component: () => import('../views/TransfersListView.vue'),
      meta: { requiresAdmin: true },
    },
    {
      path: '/cars',
      name: 'cars',
      component: () => import('../views/CarsView.vue'),
    },
    {
      path: '/cars/stock',
      name: 'cars-stock',
      component: () => import('../views/CarsStock.vue'),
    },
    {
      path: '/warehouses',
      name: 'warehouses',
      component: () => import('../views/WarehousesView.vue'),
    },
    {
      path: '/print/:billId',
      name: 'print',
      component: PrintPage,
    },
    {
      path: '/cashier',
      name: 'cashier',
      component: () => import('../views/CashierView.vue'),
      children: [
        {
          path: 'transfers',
          name: 'cashier-transfers',
          component: TransfersInterTable,
        },
        {
          path: 'money-movements',
          name: 'money-movements',
          component: MoneyMovements,
        },
        {
          path: 'china-cash',
          name: 'china-cash',
          component: ChinaCash,
          meta: { requiresAuth: true },
        },
      ],
    },
    {
      path: '/rates',
      name: 'rates',
      component: RatesView,
      meta: { requiresAuth: true },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

// Remove this navigation guard since we have the catch-all route
// router.beforeEach((to, from, next) => {
//   if (to.matched.length === 0) {
//     next('/')
//   } else {
//     next()
//   }
// })

// Navigation guard for authentication
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
