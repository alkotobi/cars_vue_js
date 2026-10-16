<script setup>
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppHeader from './components/AppHeader.vue'
import AlertsView from './views/AlertsView.vue'
import { onMounted, onUnmounted, computed, ref } from 'vue'
import { useApi } from './composables/useApi'
import DatabaseVersionCheck from '@/components/DatabaseVersionCheck.vue'
import I18nDebugPanel from '@/components/I18nDebugPanel.vue'
import SessionExpiredModal from '@/components/SessionExpiredModal.vue'

const { t } = useI18n()
const { handleCookieVerification } = useApi()
const route = useRoute()
const router = useRouter()

const user = ref(null)
const hideAlertsRoutes = ['/login', '/print', '/print-car', '/alert-cars', '/clients']
const isAdmin = computed(() => user.value?.role_id === 1)
const canViewAlerts = computed(() => {
  if (!user.value) return false
  if (isAdmin.value) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_c_alerts')
})
const showAlerts = computed(() => {
  return canViewAlerts.value && !hideAlertsRoutes.some((hideRoute) => route.path.startsWith(hideRoute))
})

// The chat page goes immersive on anything up to a tablet, not just phones: at
// 769-1024px (iPad portrait, phones in landscape) the chat page now renders the
// mobile list/conversation, so the header must step aside there too.
const isMobile = ref(false)
const mobileQuery = typeof window !== 'undefined' ? window.matchMedia('(max-width: 1024px)') : null
const setMobile = () => {
  isMobile.value = mobileQuery ? mobileQuery.matches : false
}
const hideChrome = computed(() => isMobile.value && route.name === 'chat')

onMounted(async () => {
  setMobile()
  if (mobileQuery) mobileQuery.addEventListener('change', setMobile)
  try {
    await handleCookieVerification()
  } catch (error) {
    console.error('Cookie verification failed:', error)
  }
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
})

onUnmounted(() => {
  if (mobileQuery) mobileQuery.removeEventListener('change', setMobile)
})
</script>

<template>
  <div id="app">
    <DatabaseVersionCheck />
    <AlertsView v-if="showAlerts" />
    <AppHeader v-if="!hideChrome" />
    <main class="app-main" :class="{ 'with-alerts': showAlerts, 'app-main--immersive': hideChrome }">
      <RouterView />
    </main>
    <I18nDebugPanel />
    <SessionExpiredModal />
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
  padding-top: 70px;
}

.app-main.with-alerts {
  padding-top: calc(70px + var(--alerts-banner-height, 100px));
}

.app-main--immersive {
  padding-top: 0;
  height: 100dvh;
  height: 100vh;
  min-height: 0;
  max-height: 100dvh;
  max-height: 100vh;
  overflow: hidden;
}

@media (max-width: 768px) {
  .app-main {
    padding-top: 60px;
  }

  .app-main.with-alerts {
    padding-top: 120px;
  }
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
