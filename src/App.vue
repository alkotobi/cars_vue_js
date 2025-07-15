<script setup>
import { RouterView, useRoute } from 'vue-router'
import AppHeader from './components/AppHeader.vue'
import AlertsView from './views/AlertsView.vue'
import { onMounted, computed } from 'vue'
import { useApi } from './composables/useApi'

const { handleCookieVerification } = useApi()
const route = useRoute()

// Define routes where alerts should be hidden
const hideAlertsRoutes = [
  '/login',
  '/print',
  '/print-car',
  '/advanced-sql',
  '/alert-cars',
  '/clients',
]

// Computed property to determine if alerts should be shown
const showAlerts = computed(() => {
  return !hideAlertsRoutes.some((hideRoute) => route.path.startsWith(hideRoute))
})

onMounted(async () => {
  // Handle cookie verification on app start
  try {
    await handleCookieVerification()
  } catch (error) {
    console.error('Cookie verification failed:', error)
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
