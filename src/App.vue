<script setup>
import { RouterView, useRoute } from 'vue-router'
import AppHeader from './components/AppHeader.vue'
import AlertsView from './views/AlertsView.vue'
import { onMounted, computed, ref } from 'vue'
import { useApi } from './composables/useApi'

const { handleCookieVerification } = useApi()
const route = useRoute()

// Add user state
const user = ref(null)

// Define routes where alerts should be hidden
const hideAlertsRoutes = [
  '/login',
  '/print',
  '/print-car',
  '/advanced-sql',
  '/alert-cars',
  '/clients',
]

// Computed property to check if user is admin
const isAdmin = computed(() => {
  return user.value?.role_id === 1
})

// Computed property to determine if alerts should be shown
const showAlerts = computed(() => {
  // Only show alerts if user is admin and not on hidden routes
  return isAdmin.value && !hideAlertsRoutes.some((hideRoute) => route.path.startsWith(hideRoute))
})

onMounted(async () => {
  // Handle cookie verification on app start
  try {
    await handleCookieVerification()
  } catch (error) {
    console.error('Cookie verification failed:', error)
  }

  // Get user from localStorage
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
})
</script>

<template>
  <div id="app">
    <AlertsView v-if="showAlerts" />
    <AppHeader />
    <main class="app-main">
      <RouterView />
    </main>
  </div>
</template>

<style>
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-main {
  flex: 1;
  background-color: #f8fafc;
}

@media print {
  .app-header,
  header,
  .app-header *,
  .alerts-view,
  .alerts-view * {
    display: none !important;
    visibility: hidden !important;
    height: 0 !important;
    overflow: hidden !important;
    margin: 0 !important;
    padding: 0 !important;
  }
}
</style>
