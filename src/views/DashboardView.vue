<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import LogoutButton from '../components/LogoutButton.vue'
import { useApi } from '../composables/useApi'

const router = useRouter()
const user = ref(null)
const latestRate = ref(null)
const { callApi } = useApi()

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

const canAccessRates = computed(() => {
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for exchange sender permission
  return user.value.permissions?.some(p => p.permission_name === 'is_exchange_sender')
})

const navigateToTransfers = () => {
  router.push('/transfers')
}

const navigateToRates = () => {
  router.push('/rates')
}

const fetchLatestRate = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT rate, created_on 
        FROM rates 
        ORDER BY created_on DESC 
        LIMIT 1
      `
    })
    if (result.success && result.data.length > 0) {
      latestRate.value = result.data[0]
    }
  } catch (error) {
    console.error('Error fetching latest rate:', error)
  }
}

onMounted(async () => {
  const userStr = localStorage.getItem('user')
  if (!userStr) {
    router.push('/login')
    return
  }
  user.value = JSON.parse(userStr)
  await fetchLatestRate()
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

const navigateToCashier = () => {
  router.push('/cashier')
}

const canAccessCashier = computed(() => {
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for specific permission
  return user.value.permissions?.some(p => p.permission_name === 'can_access_cashier')
})
</script>

<template>
  <div class="dashboard">
    <div class="header">
      <div class="user-info">
        <h1>Welcome {{ user?.username }}</h1>
        <p class="role">Role: {{ user?.role_name }}</p>
        <p v-if="latestRate" class="rate-info">
          The rate is {{ latestRate.rate }} on {{ new Date(latestRate.created_on).toLocaleString() }}
        </p>
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
      <button 
        v-if="canAccessCashier"
        @click="navigateToCashier" 
        class="action-btn cashier-btn"
      >
        Cashier
      </button>
      <button 
        v-if="canAccessRates"
        @click="navigateToRates" 
        class="action-btn rates-btn"
      >
        Rates
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
  padding: 15px 30px;
  margin-right: 10px;
  border: none;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 500;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.action-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.users-btn {
  background-color: #3b82f6;
  color: white;
}

.transfers-btn {
  background-color: #10b981;
  color: white;
}

.cars-btn {
  background-color: #f59e0b;
  color: white;
}

.cashier-btn {
  background-color: #8b5cf6;
  color: white;
}

.cashier-btn:hover {
  background-color: #7c3aed;
}

.rates-btn {
  background-color: #9c27b0;
  color: white;
}

.rates-btn:hover {
  background-color: #7b1fa2;
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

.rate-info {
  margin-top: 10px;
  padding: 8px 12px;
  background-color: #e3f2fd;
  border-radius: 4px;
  color: #1976d2;
  font-weight: 500;
  display: inline-block;
}
</style>
