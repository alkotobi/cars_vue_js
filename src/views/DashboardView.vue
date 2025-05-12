<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import LogoutButton from '../components/LogoutButton.vue'

const router = useRouter()
const user = ref(null)

const canManageUsers = computed(() => {
  console.log(user.value);
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for specific permission
  return user.value.permissions?.some(p => p.permission_name === 'can_manage_users')
})

const canAccessTransfers = computed(() => {
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for exchange permissions
  return user.value.permissions?.some(p => 
    p.permission_name === 'is_exchange_sender' || 
    p.permission_name === 'is_exchange_receiver'
  )
})

const navigateToTransfers = () => {
  router.push('/transfers')
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (!userStr) {
    router.push('/login')
    return
  }
  user.value = JSON.parse(userStr)
})

const navigateToUsers = () => {
  router.push('/users')
}

const canManageCars = computed(() => {
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for specific permission
  return user.value.permissions?.some(p => p.permission_name === 'can_manage_cars')
})

const navigateToCars = () => {
  router.push('/cars')
}
</script>

<template>
  <div class="dashboard">
    <div class="header">
      <div class="user-info">
        <h1>Welcome {{ user?.username }}</h1>
        <p class="role">Role: {{ user?.role_name }}</p>
      </div>
    </div>
    <LogoutButton />
    <div class="actions-section">
      <button 
        v-if="canManageUsers" 
        @click="navigateToUsers" 
        class="action-btn users-btn"
      >
        Users
      </button>
      <button 
        v-if="canAccessTransfers" 
        @click="navigateToTransfers" 
        class="action-btn transfers-btn"
      >
        Transfers 
      </button>
      <button 
        v-if="canManageCars"
        @click="navigateToCars" 
        class="action-btn cars-btn"
      >
        Cars
      </button>

    </div>
    <div class="permissions-section" v-if="user?.permissions?.length">
      <h2>Your Permissions:</h2>
      <ul class="permissions-list">
        <li v-for="permission in user.permissions" :key="permission.permission_name">
          <strong>{{ permission.permission_name }}</strong>
          <p v-if="permission.description">{{ permission.description }}</p>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.logout-btn {
  padding: 8px 16px;
  background-color: #f44336;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.logout-btn:hover {
  background-color: #d32f2f;
}

.user-info {
  flex: 1;
}

.role {
  color: #666;
  margin-top: 5px;
}

.permissions-section {
  margin-top: 30px;
  padding: 20px;
  background-color: #f5f5f5;
  border-radius: 8px;
}

.permissions-list {
  list-style: none;
  padding: 0;
}

.permissions-list li {
  margin-bottom: 15px;
  padding: 10px;
  background-color: white;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.permissions-list strong {
  color: #2c3e50;
  display: block;
  margin-bottom: 5px;
}

.permissions-list p {
  color: #666;
  margin: 0;
  font-size: 0.9em;
}

.actions-section {
  margin: 20px 0;
}

.action-btn {
  padding: 10px 20px;
  margin: 5px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1em;
  transition: background-color 0.3s;
}

.users-btn {
  background-color: #2196F3;
  color: white;
}

.users-btn:hover {
  background-color: #1976D2;
}

.transfers-btn {
  background-color: #4CAF50;
  color: white;
}

.transfers-btn:hover {
  background-color: #388E3C;
}

.cars-btn {
  background-color: #9C27B0;
  color: white;
}

.cars-btn:hover {
  background-color: #7B1FA2;
}
</style>
