<script setup>
import { ref, computed, onMounted, watch, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useEnhancedI18n } from '../composables/useI18n'
import { useApi } from '../composables/useApi'
import LogoutButton from './LogoutButton.vue'
import ChatModal from './ChatModal.vue'
import ChatMessages from './chat/ChatMessages.vue'

const router = useRouter()
const { t } = useEnhancedI18n()
const { callApi, getAssets } = useApi()
const user = ref(null)
const logoUrl = ref(null)
const showChatModal = ref(false)
const unreadMessageCount = ref(0)
const chatMessagesRef = ref(null)
const pendingTasksCount = ref(0)

// User activity tracking
const isUserActive = ref(true)
const lastActivityTime = ref(Date.now())
const inactivityTimeout = 5 * 60 * 1000 // 5 minutes of inactivity
let inactivityTimer = null

// Adaptive polling intervals
const ACTIVE_INTERVALS = {
  tasks: 2 * 60 * 1000, // 2 minutes when active
  chat: 10 * 1000, // 10 seconds when active
}

const INACTIVE_INTERVALS = {
  tasks: 5 * 60 * 1000, // 5 minutes when inactive
  chat: 30 * 1000, // 30 seconds when inactive
}

// Get user from localStorage
const getUser = async () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    // If user is loaded, fetch data immediately
    if (user.value) {
      await forceRefreshData()
    }
  } else {
    user.value = null
  }
}

// Fetch pending tasks count
const fetchPendingTasksCount = async () => {
  if (!user.value) {
    return
  }

  try {
    const result = await callApi({
      query: `
        SELECT COUNT(*) as count
        FROM tasks t
        WHERE t.date_declare_done IS NULL
        AND (t.id_user_receive = ? OR t.id_user_create = ?)
      `,
      params: [user.value.id, user.value.id],
      requiresAuth: true,
    })

    if (result.success && result.data.length > 0) {
      const count = result.data[0].count
      pendingTasksCount.value = count
    } else {
      pendingTasksCount.value = 0
    }
  } catch (error) {
    pendingTasksCount.value = 0
  }
}

// User activity tracking functions
const updateUserActivity = () => {
  lastActivityTime.value = Date.now()
  if (!isUserActive.value) {
    isUserActive.value = true
    restartPolling()
  }
  resetInactivityTimer()
}

const resetInactivityTimer = () => {
  if (inactivityTimer) {
    clearTimeout(inactivityTimer)
  }
  inactivityTimer = setTimeout(() => {
    isUserActive.value = false
    restartPolling()
  }, inactivityTimeout)
}

const restartPolling = () => {
  // Clear existing intervals
  if (unreadCountInterval) {
    clearInterval(unreadCountInterval)
  }
  if (tasksCountInterval) {
    clearInterval(tasksCountInterval)
  }

  // Get current intervals based on activity
  const currentIntervals = isUserActive.value ? ACTIVE_INTERVALS : INACTIVE_INTERVALS

  // Restart with new intervals
  unreadCountInterval = setInterval(updateUnreadCount, currentIntervals.chat)
  tasksCountInterval = setInterval(fetchPendingTasksCount, currentIntervals.tasks)
}

// Set up activity listeners
const setupActivityListeners = () => {
  const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click']
  events.forEach((event) => {
    document.addEventListener(event, updateUserActivity, { passive: true })
  })

  // Also track visibility changes
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
      updateUserActivity()
    }
  })
}

// Clean up activity listeners
const cleanupActivityListeners = () => {
  const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click']
  events.forEach((event) => {
    document.removeEventListener(event, updateUserActivity)
  })
  document.removeEventListener('visibilitychange', updateUserInterval)
}

// Watch for localStorage changes (when user logs in/out)
const watchUserChanges = () => {
  window.addEventListener('storage', () => getUser())
  // Also watch for custom events
  window.addEventListener('userLogin', () => getUser())
  window.addEventListener('userLogout', () => getUser())
}

// Check if we're on the login page
const isLoginPage = computed(() => {
  return router.currentRoute.value.name === 'login'
})

// Get current page name for display
const currentPageName = computed(() => {
  const route = router.currentRoute.value
  const routeNames = {
    dashboard: t('navigation.dashboard'),
    users: t('navigation.users'),
    roles: t('navigation.roles'),
    transfers: t('navigation.transfers'),
    'send-transfer': t('navigation.sendTransfer'),
    'receive-transfer': t('navigation.receiveTransfer'),
    'sell-bills': t('navigation.sellBills'),
    'buy-payments': t('navigation.buyPayments'),
    params: t('navigation.params'),
    'advanced-sql': t('navigation.advancedSql'),
    'transfers-list': t('navigation.transfersList'),
    cars: t('navigation.cars'),
    'cars-stock': t('navigation.carsStock'),
    warehouses: t('navigation.warehouses'),
    containers: t('navigation.containers'),
    clients: t('navigation.clients'),
    cashier: t('navigation.cashier'),
    rates: t('navigation.rates'),
    tasks: t('navigation.tasks'),
    statistics: t('navigation.statistics'),
    chat: t('navigation.chat'),
  }
  return routeNames[route.name] || t('navigation.dashboard')
})

// Chat functions
const openChat = () => {
  showChatModal.value = true
  // Reset unread count when opening chat
  unreadMessageCount.value = 0

  // Reset unread counts for all groups
  if (chatMessagesRef.value?.fetchAllGroupsMessages) {
    setTimeout(async () => {
      await chatMessagesRef.value.fetchAllGroupsMessages()
      updateUnreadCount()
    }, 100)
  }
}

const closeChat = () => {
  showChatModal.value = false
}

// Function to update unread message count
const updateUnreadCount = async () => {
  try {
    // Fetch all groups messages to update unread counts
    if (chatMessagesRef.value?.fetchAllGroupsMessages) {
      await chatMessagesRef.value.fetchAllGroupsMessages()
    }

    // Then get the updated counts
    if (chatMessagesRef.value?.getAllNewMessagesCounts) {
      const counts = chatMessagesRef.value.getAllNewMessagesCounts()
      // Sum up all unread counts from all groups
      const totalUnread = Object.values(counts).reduce((sum, count) => sum + (count || 0), 0)
      unreadMessageCount.value = totalUnread
    }
  } catch (error) {}
}

// Set up periodic updates for unread count
let unreadCountInterval = null
let tasksCountInterval = null

// Get base path helper (same logic as useApi.js)
const getBasePath = () => {
  let baseUrl = import.meta.env.BASE_URL || './'
  if (baseUrl === './' || baseUrl.startsWith('./')) {
    const pathname = window.location.pathname
    const match = pathname.match(/^(\/[^/]+\/)/)
    return match ? match[1] : '/'
  }
  return baseUrl
}

// Load logo from assets
const loadLogo = async () => {
  console.log('[AppHeader] loadLogo() called')
  const basePath = getBasePath()
  
  // Get assets version from localStorage
  const STORAGE_KEY = 'assets_version'
  const assetsVersion = localStorage.getItem(STORAGE_KEY) || Date.now().toString()
  
  // Add cache-busting with both version and timestamp for maximum cache invalidation
  const timestamp = Date.now()
  const fallbackLogo = `${basePath}logo.png?v=${assetsVersion}&t=${timestamp}`

  try {
    console.log('[AppHeader] Calling getAssets()...')
    const assets = await getAssets()
    console.log('[AppHeader] getAssets() returned:', assets)

    if (assets && assets.logo) {
      // Add additional timestamp to force reload even if version is same
      const logoWithTimestamp = `${assets.logo}&t=${timestamp}`
      console.log('[AppHeader] Setting logoUrl to:', logoWithTimestamp)
      // Force reload by setting to empty first, then to new URL
      logoUrl.value = ''
      await nextTick()
      logoUrl.value = logoWithTimestamp
      console.log('[AppHeader] logoUrl.value is now:', logoUrl.value)
    } else {
      // Fallback to default logo path if getAssets doesn't return logo
      console.log('[AppHeader] No logo in assets, using fallback:', fallbackLogo)
      logoUrl.value = ''
      await nextTick()
      logoUrl.value = fallbackLogo
      console.log('[AppHeader] logoUrl.value is now:', logoUrl.value)
    }
  } catch (err) {
    console.warn('[AppHeader] Failed to load logo from assets, using default:', err)
    // Fallback to default logo path
    logoUrl.value = ''
    await nextTick()
    logoUrl.value = fallbackLogo
    console.log('[AppHeader] logoUrl.value set to fallback:', logoUrl.value)
  }

  console.log('[AppHeader] Final logoUrl.value:', logoUrl.value)
}

// Initialize user on component mount
onMounted(async () => {
  await getUser()
  loadLogo() // Load logo asynchronously
  watchUserChanges()
  
  // Listen for assets update event (when new logo/letterhead/stamp is uploaded)
  window.addEventListener('assetsUpdated', loadLogo)

  // Set up activity listeners
  setupActivityListeners()

  // Initialize adaptive polling with active intervals (user just loaded the page)
  const currentIntervals = ACTIVE_INTERVALS
  unreadCountInterval = setInterval(updateUnreadCount, currentIntervals.chat)
  tasksCountInterval = setInterval(fetchPendingTasksCount, currentIntervals.tasks)

  // Listen for global events to force update badge
  window.addEventListener('forceUpdateBadge', updateUnreadCount)
  window.addEventListener('forceUpdateTasks', fetchPendingTasksCount)
})

// Separate function to initialize chat data
const initializeChatData = async () => {
  // Wait for the hidden ChatMessages component to be ready (reduced wait time)
  await new Promise((resolve) => setTimeout(resolve, 500))

  // Initialize unread counts
  if (chatMessagesRef.value?.fetchAllGroupsMessages) {
    await chatMessagesRef.value.fetchAllGroupsMessages()
    updateUnreadCount()
  }
}

// Force refresh all data (called on login)
const forceRefreshData = async () => {
  if (user.value) {
    // Small delay to ensure API is ready
    await new Promise((resolve) => setTimeout(resolve, 100))
    await fetchPendingTasksCount()
    await initializeChatData()
  }
}

onUnmounted(() => {
  if (unreadCountInterval) {
    clearInterval(unreadCountInterval)
  }
  if (tasksCountInterval) {
    clearInterval(tasksCountInterval)
  }
  if (inactivityTimer) {
    clearTimeout(inactivityTimer)
  }
  if (chatMessagesRef.value?.cleanup) {
    chatMessagesRef.value.cleanup()
  }
  // Remove global event listener
  window.removeEventListener('forceUpdateBadge', updateUnreadCount)
  window.removeEventListener('forceUpdateTasks', fetchPendingTasksCount)

  // Clean up activity listeners
  cleanupActivityListeners()
})
</script>

<template>
  <header v-if="!isLoginPage" class="app-header">
    <div class="header-content">
      <div class="header-left">
        <div class="logo-section">
          <img
            :key="logoUrl"
            :src="logoUrl || `${getBasePath()}logo.png`"
            alt="Company Logo"
            class="company-logo"
            @load="
              () =>
                console.log(
                  '[AppHeader] Logo image loaded successfully from:',
                  logoUrl || `${getBasePath()}logo.png`,
                )
            "
            @error="
              (e) =>
                console.error(
                  '[AppHeader] Logo image failed to load from:',
                  logoUrl || `${getBasePath()}logo.png`,
                  e,
                )
            "
          />
          <div class="company-info">
            <h1 class="company-name">{{ t('app.companyName') }}</h1>
            <p class="company-tagline">{{ t('app.subtitle') }}</p>
          </div>
        </div>
      </div>

      <div class="header-center">
        <div class="current-page">
          <i class="fas fa-map-marker-alt"></i>
          <span>{{ currentPageName }}</span>
        </div>
      </div>

      <div class="header-right">
        <div class="user-info" v-if="user">
          <div class="user-details">
            <i class="fas fa-user-circle"></i>
            <span class="username">{{ user.username }}</span>
            <span class="user-role">{{ user.role_name || t('app.defaultRole') }}</span>
          </div>
        </div>

        <button @click="openChat" class="chat-btn" :title="t('app.openChat')">
          <i class="fas fa-comments"></i>
          <span class="chat-label">{{ t('navigation.chat') }}</span>
          <span v-if="unreadMessageCount > 0" class="notification-badge">
            {{ unreadMessageCount > 99 ? '99+' : unreadMessageCount }}
          </span>
        </button>

        <button @click="router.push('/tasks')" class="tasks-btn" :title="t('app.viewTasks')">
          <i class="fas fa-tasks"></i>
          <span class="tasks-label">{{ t('navigation.tasks') }}</span>
          <span v-if="pendingTasksCount > 0" class="notification-badge">
            {{ pendingTasksCount > 99 ? '99+' : pendingTasksCount }}
          </span>
        </button>

        <LogoutButton />
      </div>
    </div>
  </header>

  <!-- Chat Modal -->
  <div v-if="showChatModal" class="chat-modal-overlay" @click="closeChat">
    <div class="chat-modal" @click.stop>
      <div class="chat-modal-header">
        <h3>
          <i class="fas fa-comments"></i>
          {{ t('app.teamChat') }}
        </h3>
        <button @click="closeChat" class="close-chat-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="chat-modal-content">
        <ChatModal />
      </div>
    </div>
  </div>

  <!-- Hidden ChatMessages component for tracking unread counts -->
  <div style="display: none">
    <ChatMessages ref="chatMessagesRef" :group-id="0" :group-name="'header-init'" />
  </div>
</template>

<style scoped>
.app-header {
  background: linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%);
  color: white;
  padding: 0;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  width: 100%;
  z-index: 1001; /* Increased to be above dialog overlays */
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 70px;
  max-width: 1400px;
  margin: 0 auto;
}

.header-left {
  display: flex;
  align-items: center;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 15px;
  position: relative;
  z-index: 1;
}

.company-logo {
  width: 50px;
  height: 50px;
  border-radius: 8px;
  object-fit: cover;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  display: block;
  visibility: visible;
  opacity: 1;
  position: relative;
  z-index: 1;
}

.company-info {
  display: flex;
  flex-direction: column;
}

.company-name {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

.company-tagline {
  margin: 0;
  font-size: 0.85rem;
  color: #bfdbfe;
  font-weight: 400;
}

.header-center {
  display: flex;
  align-items: center;
}

.current-page {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: rgba(255, 255, 255, 0.1);
  padding: 8px 16px;
  border-radius: 20px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.current-page i {
  color: #fbbf24;
  font-size: 0.9rem;
}

.current-page span {
  font-weight: 500;
  font-size: 0.95rem;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.chat-btn {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: rgba(255, 255, 255, 0.1);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 8px;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.2s;
  backdrop-filter: blur(10px);
  font-size: 0.9rem;
  font-weight: 500;
}

.chat-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
  transform: translateY(-1px);
}

.chat-btn i {
  color: #fbbf24;
  font-size: 1rem;
}

.chat-label {
  font-weight: 500;
}

.tasks-btn {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: rgba(255, 255, 255, 0.1);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 8px 12px;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.2s;
  backdrop-filter: blur(10px);
  font-size: 0.9rem;
  font-weight: 500;
}

.tasks-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
  transform: translateY(-1px);
}

.tasks-btn i {
  color: #10b981;
  font-size: 1rem;
}

.tasks-label {
  font-weight: 500;
}

.notification-badge {
  position: absolute;
  top: -8px;
  right: -8px;
  background-color: #ef4444;
  color: white;
  border-radius: 50%;
  min-width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 600;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.user-info {
  display: flex;
  align-items: center;
}

.user-details {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: rgba(255, 255, 255, 0.1);
  padding: 8px 12px;
  border-radius: 20px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.user-details i {
  font-size: 1.2rem;
  color: #fbbf24;
}

.username {
  font-weight: 600;
  font-size: 0.9rem;
}

.user-role {
  font-size: 0.75rem;
  color: #bfdbfe;
  background-color: rgba(255, 255, 255, 0.1);
  padding: 2px 6px;
  border-radius: 10px;
}

/* Responsive Design */
@media (max-width: 768px) {
  .header-content {
    padding: 0 15px;
    height: 60px;
  }

  .company-name {
    font-size: 1.2rem;
  }

  .company-tagline {
    font-size: 0.75rem;
  }

  .company-logo {
    width: 40px;
    height: 40px;
  }

  .current-page {
    padding: 6px 12px;
  }

  .current-page span {
    font-size: 0.85rem;
  }

  .user-details {
    padding: 6px 10px;
  }

  .username {
    font-size: 0.8rem;
  }

  .user-role {
    font-size: 0.7rem;
  }

  .chat-btn {
    padding: 8px;
    min-width: 40px;
    justify-content: center;
  }

  .tasks-btn {
    padding: 8px;
    min-width: 40px;
    justify-content: center;
  }

  .chat-label {
    display: none;
  }

  .tasks-label {
    display: none;
  }
}

@media (max-width: 480px) {
  .header-content {
    flex-direction: column;
    height: auto;
    padding: 10px 15px;
    gap: 10px;
  }

  .header-center {
    order: 3;
  }

  .current-page {
    font-size: 0.8rem;
  }
}

/* Chat Modal Styles */
.chat-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
  backdrop-filter: blur(5px);
}

.chat-modal {
  background-color: white;
  border-radius: 12px;
  width: 95%;
  max-width: 1200px;
  height: 90%;
  max-height: 800px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%);
  color: white;
  border-radius: 12px 12px 0 0;
}

.chat-modal-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1.3rem;
  font-weight: 600;
}

.chat-modal-header i {
  color: #fbbf24;
}

.close-chat-btn {
  background: none;
  border: none;
  color: white;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 5px;
  border-radius: 50%;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
}

.close-chat-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
  transform: scale(1.1);
}

.chat-modal-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

/* Responsive adjustments for chat modal */
@media (max-width: 768px) {
  .chat-modal {
    width: 100%;
    height: 100%;
    max-height: none;
    border-radius: 0;
  }

  .chat-modal-header {
    border-radius: 0;
    padding: 15px;
  }

  .chat-modal-header h3 {
    font-size: 1.1rem;
  }

  .chat-label {
    display: none;
  }

  .tasks-label {
    display: none;
  }

  .chat-btn {
    padding: 8px;
    min-width: 40px;
    justify-content: center;
  }

  .tasks-btn {
    padding: 8px;
    min-width: 40px;
    justify-content: center;
  }
}
</style>
