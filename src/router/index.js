import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import PrintPage from '../views/PrintPage.vue'
import PrintCarPage from '../views/PrintCarPage.vue'
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
      children: [
        {
          path: 'client-details',
          name: 'client-transfers-detail',
          component: () => import('../views/ClientTransfersDetail.vue'),
          meta: { requiresTransferAccess: true },
        },
      ],
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
      path: '/sell-bills/:id',
      name: 'sell-bill-details',
      component: () => import('../views/SellBillDetailsView.vue'),
    },
    {
      path: '/buy-payments/:id',
      name: 'buy-payments',
      component: BuyBillPaymentsView,
      meta: { requiresAuth: true },
    },
    {
      path: '/params',
      name: 'params',
      component: () => import('../views/ParamsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/advanced-sql',
      name: 'advanced-sql',
      component: () => import('../views/AdvancedSqlView.vue'),
      meta: { requiresAdmin: true },
    },
    {
      path: '/transfers-list',
      name: 'transfers-list',
      component: () => import('../views/TransfersListView.vue'),
      meta: { requiresTransferAccess: true },
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
      path: '/containers',
      name: 'containers',
      component: () => import('../views/ContainersView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/print/:billId',
      name: 'print',
      component: PrintPage,
    },
    {
      path: '/print-car/:carId',
      name: 'print-car',
      component: () => import('../views/PrintCarPage.vue'),
    },
    {
      path: '/clients',
      name: 'clients',
      component: () => import('../views/ClientsView.vue'),
    },
    {
      path: '/clients/:id',
      name: 'client-details',
      component: () => import('../views/ClientDetailsView.vue'),
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
      path: '/tasks',
      name: 'tasks',
      component: () => import('../views/TasksView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/statistics',
      name: 'statistics',
      component: () => import('../views/StatisticsView.vue'),
      meta: { requiresAdmin: true },
    },
    {
      path: '/chat',
      name: 'chat',
      component: () => import('../views/ChatView.vue'),
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

// Navigation guard for authentication and admin routes
router.beforeEach((to, from, next) => {
  const publicPages = ['/login']
  // Check if the current route is a client details page
  if (to.name === 'client-details') {
    return next()
  }
  const authRequired = !publicPages.includes(to.path)
  const user = localStorage.getItem('user')
  const userData = user ? JSON.parse(user) : null

  if (authRequired && !user) {
    return next('/login')
  }

  if (to.path === '/login' && user) {
    return next('/dashboard')
  }

  // Check for admin-only routes
  if (to.meta.requiresAdmin && (!userData || userData.role_id !== 1)) {
    return next('/dashboard')
  }

  // Check for transfer access routes
  if (to.meta.requiresTransferAccess && userData) {
    const hasAccess =
      userData.role_id === 1 || // Admin
      userData.permissions?.some(
        (p) =>
          p.permission_name === 'is_exchange_sender' ||
          p.permission_name === 'is_exchange_receiver',
      )

    if (!hasAccess) {
      return next('/dashboard')
    }
  }

  next()
})

export default router
