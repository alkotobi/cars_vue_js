<script setup>
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import LogoutButton from './LogoutButton.vue'
import ChatModal from './ChatModal.vue'
import ChatMessages from './chat/ChatMessages.vue'

const router = useRouter()
const user = ref(null)
const showChatModal = ref(false)
const unreadMessageCount = ref(0)
const chatMessagesRef = ref(null)

// Get user from localStorage
const getUser = () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  } else {
    user.value = null
  }
}

// Watch for localStorage changes (when user logs in/out)
const watchUserChanges = () => {
  window.addEventListener('storage', getUser)
  // Also watch for custom events
  window.addEventListener('userLogin', getUser)
  window.addEventListener('userLogout', getUser)
}

// Check if we're on the login page
const isLoginPage = computed(() => {
  return router.currentRoute.value.name === 'login'
})

// Get current page name for display
const currentPageName = computed(() => {
  const route = router.currentRoute.value
  const routeNames = {
    dashboard: 'Dashboard',
    users: 'Users Management',
    roles: 'Roles Management',
    transfers: 'Transfers',
    'send-transfer': 'Send Transfer',
    'receive-transfer': 'Receive Transfer',
    'sell-bills': 'Sell Bills',
    'buy-payments': 'Buy Payments',
    params: 'Parameters',
    'advanced-sql': 'Advanced SQL',
    'transfers-list': 'Transfers List',
    cars: 'Cars',
    'cars-stock': 'Cars Stock',
    warehouses: 'Warehouses',
    containers: 'Containers',
    clients: 'Clients',
    cashier: 'Cashier',
    rates: 'Exchange Rates',
    tasks: 'Tasks',
    statistics: 'Statistics',
    chat: 'Chat',
  }
  return routeNames[route.name] || 'Dashboard'
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
      console.log('Header: Updated unread count to', totalUnread, 'from counts:', counts)
    }
  } catch (error) {
    console.error('Error updating unread count:', error)
  }
}

// Set up periodic updates for unread count
let unreadCountInterval = null

// Initialize user on component mount
onMounted(async () => {
  getUser()
  watchUserChanges()

  // Wait for the hidden ChatMessages component to be ready
  await new Promise((resolve) => setTimeout(resolve, 1000))

  // Initialize unread counts
  if (chatMessagesRef.value?.fetchAllGroupsMessages) {
    await chatMessagesRef.value.fetchAllGroupsMessages()
    updateUnreadCount()
  }

  // Set up periodic updates for unread count
  unreadCountInterval = setInterval(updateUnreadCount, 3000) // Update every 3 seconds

  // Listen for global events to force update badge
  window.addEventListener('forceUpdateBadge', updateUnreadCount)
})

onUnmounted(() => {
  if (unreadCountInterval) {
    clearInterval(unreadCountInterval)
  }
  if (chatMessagesRef.value?.cleanup) {
    chatMessagesRef.value.cleanup()
  }
  // Remove global event listener
  window.removeEventListener('forceUpdateBadge', updateUnreadCount)
})
</script>

<template>
  <header v-if="!isLoginPage" class="app-header">
    <div class="header-content">
      <div class="header-left">
        <div class="logo-section">
          <img src="../assets/logo.png" alt="Company Logo" class="company-logo" />
          <div class="company-info">
            <h1 class="company-name">GML Trading</h1>
            <p class="company-tagline">Car Trading Management System</p>
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
            <span class="user-role">{{ user.role_name || 'User' }}</span>
          </div>
        </div>

        <button @click="openChat" class="chat-btn" title="Open Chat">
          <i class="fas fa-comments"></i>
          <span class="chat-label">Chat</span>
          <span v-if="unreadMessageCount > 0" class="notification-badge">
            {{ unreadMessageCount > 99 ? '99+' : unreadMessageCount }}
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
          Team Chat
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
  position: sticky;
  top: 0;
  z-index: 1000;
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
}

.company-logo {
  width: 50px;
  height: 50px;
  border-radius: 8px;
  object-fit: cover;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
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
  padding: 8px 12px;
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

  .chat-btn {
    padding: 8px;
    min-width: 40px;
    justify-content: center;
  }
}
</style>
