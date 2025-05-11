import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'

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
      component: () => import('../views/TransfersView.vue'),
      children: [
        {
          path: 'send',
          name: 'send-transfer',
          component: () => import('../views/SenderView.vue')
        },
        {
          path: 'receive',
          name: 'receive-transfer',
          component: () => import('../views/ReceiverView.vue')
        }
      ]
    }
  ]
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
