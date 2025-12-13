<template>
  <div class="db-manager">
    <!-- Show login if not authenticated -->
    <div v-if="!isLoggedIn" class="login-container">
      <LoginSignup @login-success="handleLoginSuccess" />
    </div>

    <!-- Show sidebar and content if authenticated -->
    <div v-else class="manager-layout">
      <DbManagerSidebar
        :active-item="activeItem"
        @select-item="handleSelectItem"
        @logout="handleLogout"
      />
      <div class="main-content">
        <div class="content-area">
          <!-- SQL Component -->
          <div v-if="activeItem === 'sql'" class="sql-section">
            <h2>SQL Query</h2>
            <p>SQL query interface will be implemented here</p>
          </div>

          <!-- Databases Component -->
          <Databases v-else-if="activeItem === 'databases'" />

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
import { ref, onMounted } from 'vue'
import LoginSignup from '../components/db-manager/LoginSignup.vue'
import DbManagerSidebar from '../components/db-manager/DbManagerSidebar.vue'
import Databases from '../components/db-manager/Databases.vue'

const isLoggedIn = ref(false)
const activeItem = ref('')

const checkLoginStatus = () => {
  const user = localStorage.getItem('db_manager_user')
  isLoggedIn.value = !!user
}

const handleLoginSuccess = (userData) => {
  isLoggedIn.value = true
  console.log('Login successful:', userData)
}

const handleLogout = () => {
  isLoggedIn.value = false
  activeItem.value = ''
}

const handleSelectItem = (item) => {
  activeItem.value = item
}

onMounted(() => {
  checkLoginStatus()
})
</script>

<style scoped>
.db-manager {
  min-height: 100vh;
  background-color: #f5f7fa;
}

.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.manager-layout {
  display: flex;
  min-height: calc(100vh - 70px); /* Account for fixed header */
}

.main-content {
  margin-left: 250px;
  flex: 1;
  padding: 2rem;
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
