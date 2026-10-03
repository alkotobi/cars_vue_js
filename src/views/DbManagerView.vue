<template>
  <div class="db-manager">
    <!-- Show login if not authenticated -->
    <div v-if="!isLoggedIn" class="login-container">
      <p v-if="sessionLostMessage" class="session-lost" role="alert">
        {{ sessionLostMessage }}
      </p>
      <LoginSignup @login-success="handleLoginSuccess" />
    </div>

    <!-- Show sidebar and content if authenticated -->
    <div v-else class="manager-layout">
      <DbManagerSidebar
        :active-item="activeItem"
        @select-item="handleSelectItem"
        @logout="handleLogout"
        @toggle="handleSidebarToggle"
      />
      <div class="main-content" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
        <div class="content-area">
          <!-- SQL Component -->
          <div v-if="activeItem === 'sql'" class="sql-section">
            <h2>SQL Query</h2>
            <p>SQL query interface will be implemented here</p>
          </div>

          <!-- Databases Component -->
          <Databases v-else-if="activeItem === 'databases'" />

          <!-- Update DB Structure Component -->
          <UpdateDbStructure v-else-if="activeItem === 'update-structure'" />

          <!-- Default view when no item selected -->
          <div v-else class="default-view">
            <h2>Database Manager</h2>
            <p>Select an option from the sidebar to get started</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import LoginSignup from '../components/db-manager/LoginSignup.vue'
import DbManagerSidebar from '../components/db-manager/DbManagerSidebar.vue'
import Databases from '../components/db-manager/Databases.vue'
import UpdateDbStructure from '../components/db-manager/UpdateDbStructure.vue'
import {
  getDbManagerToken,
  onDbManagerSessionLost,
} from '../composables/useDbManagerApi.js'

const isLoggedIn = ref(false)
const activeItem = ref('')
const sidebarCollapsed = ref(true)

// Why the login form replaced the panels, if it did. Null otherwise, so a normal
// sign-in does not greet the user with a stale warning.
const sessionLostMessage = ref('')

// A token in localStorage is not proof of a live session.
//
// This used to be `!!localStorage.getItem('db_manager_user')`, which only asked
// whether a key existed. A token revoked server-side - signed out in another tab, or
// api_token cleared by hand - left the key in place, so the panels kept rendering
// and every one of their requests came back "Not authenticated", producing a shell
// that looked logged in and did nothing. Reading the parsed token fixes the
// corrupt-value half of that; the revoked-token half is handled by the listener
// below, because only the server knows.
const checkLoginStatus = () => {
  isLoggedIn.value = !!getDbManagerToken()
}

const handleLoginSuccess = (userData) => {
  isLoggedIn.value = true
  sessionLostMessage.value = ''
  console.log('Login successful:', userData)
}

const handleLogout = () => {
  isLoggedIn.value = false
  activeItem.value = ''
}

const handleSelectItem = (item) => {
  activeItem.value = item
}

const handleSidebarToggle = (collapsed) => {
  sidebarCollapsed.value = collapsed
}

// The token was rejected mid-session. Drop back to the login form.
//
// useDbManagerApi.js has already cleared the stale credential by the time this runs,
// so isLoggedIn cannot be re-derived from storage here - it has to be set directly.
// Unmount cleanup matters because this component is behind a route guard: without it
// a listener would outlive the view and keep writing to a ref nobody renders.
let stopListening = null

onMounted(() => {
  checkLoginStatus()
  stopListening = onDbManagerSessionLost(() => {
    isLoggedIn.value = false
    activeItem.value = ''
    sessionLostMessage.value =
      'Your database manager session has ended. Please sign in again.'
  })
})

onUnmounted(() => {
  if (stopListening) {
    stopListening()
    stopListening = null
  }
})
</script>

<style scoped>
.db-manager {
  min-height: 100vh;
  background-color: #f5f7fa;
}

.login-container {
  display: flex;
  /* Column so the session-lost banner stacks above the form. align-items stays at its
     default `stretch` rather than being set to `center`: LoginSignup.vue's .auth-form
     is width:100% inside a stretched parent, so centering it in a column would
     shrink the card to its content width. */
  flex-direction: column;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.session-lost {
  max-width: 32rem;
  margin: 0 auto;
  padding: 0.75rem 1rem;
  border: 1px solid #d9a441;
  border-radius: 6px;
  background: #fdf6e3;
  color: #7a5a12;
  text-align: center;
  font-size: 0.9rem;
}

.manager-layout {
  display: flex;
  min-height: calc(100vh - 70px); /* Account for fixed header */
}

.main-content {
  margin-left: 250px;
  flex: 1;
  padding: 2rem;
  transition: margin-left 0.3s ease;
}

.main-content.sidebar-collapsed {
  margin-left: 70px;
}

.content-area {
  background: white;
  border-radius: 8px;
  padding: 2rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  min-height: calc(100vh - 70px - 4rem); /* Account for header and padding */
}

h2 {
  color: #2c3e50;
  margin-bottom: 1rem;
}

.default-view {
  text-align: center;
  padding: 4rem 2rem;
  color: #666;
}

.sql-section h2 {
  margin-bottom: 1.5rem;
}

@media (max-width: 768px) {
  .manager-layout {
    min-height: calc(100vh - 60px); /* Smaller header on mobile */
  }

  .main-content {
    margin-left: 0;
    padding: 1rem;
  }

  .content-area {
    padding: 1rem;
    min-height: calc(100vh - 60px - 2rem); /* Account for header and padding on mobile */
  }
}
</style>
