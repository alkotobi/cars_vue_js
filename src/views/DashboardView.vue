<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import LogoutButton from '../components/LogoutButton.vue'
import { useApi } from '../composables/useApi'
const router = useRouter()
const user = ref(null)
const latestRate = ref(null)
const { callApi } = useApi()
// Add loading and processing states
const loading = ref(false)
const isProcessing = ref({
  users: false,
  transfers: false,
  cars: false,
  cashier: false,
  rates: false,
  params: false,
})
const canManageUsers = computed(() => {
  console.log(user.value)
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for specific permission
  return user.value.permissions?.some((p) => p.permission_name === 'can_manage_users')
})
const canAccessTransfers = computed(() => {
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for exchange permissions
  return user.value.permissions?.some(
    (p) =>
      p.permission_name === 'is_exchange_sender' || p.permission_name === 'is_exchange_receiver',
  )
})
const canAccessRates = computed(() => {
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for exchange sender permission
  return user.value.permissions?.some((p) => p.permission_name === 'is_exchange_sender')
})
const handleUsersClick = async () => {
  if (isProcessing.value.users) return
  isProcessing.value.users = true
  try {
    await router.push('/users')
  } finally {
    isProcessing.value.users = false
  }
}
const handleTransfersClick = async () => {
  if (isProcessing.value.transfers) return
  isProcessing.value.transfers = true
  try {
    await router.push('/transfers')
  } finally {
    isProcessing.value.transfers = false
  }
}
const handleCarsClick = async () => {
  if (isProcessing.value.cars) return
  isProcessing.value.cars = true
  try {
    await router.push('/cars')
  } finally {
    isProcessing.value.cars = false
  }
}
const handleCashierClick = async () => {
  if (isProcessing.value.cashier) return
  isProcessing.value.cashier = true
  try {
    await router.push('/cashier')
  } finally {
    isProcessing.value.cashier = false
  }
}
const handleRatesClick = async () => {
  if (isProcessing.value.rates) return
  isProcessing.value.rates = true
  try {
    await router.push('/rates')
  } finally {
    isProcessing.value.rates = false
  }
}
const handleParamsClick = async () => {
  if (isProcessing.value.params) return
  isProcessing.value.params = true
  try {
    await router.push('/params')
  } finally {
    isProcessing.value.params = false
  }
}
const fetchLatestRate = async () => {
  loading.value = true
  try {
    const result = await callApi({
      query: `
        SELECT rate, created_on 
        FROM rates 
        ORDER BY created_on DESC 
        LIMIT 1
      `,
    })
    if (result.success && result.data.length > 0) {
      latestRate.value = result.data[0]
    }
  } catch (error) {
    console.error('Error fetching latest rate:', error)
  } finally {
    loading.value = false
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
const canManageCars = computed(() => {
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for specific permission
  return user.value.permissions?.some((p) => p.permission_name === 'can_manage_cars')
})
const canAccessCashier = computed(() => {
  if (!user.value) return false
  // Admin role (role_id = 1) can do everything
  if (user.value.role_id === 1) return true
  // Check for specific permission
  return user.value.permissions?.some((p) => p.permission_name === 'can_access_cashier')
})
const isAdmin = computed(() => {
  return user.value?.role_id === 1
})
</script>
<template>
  <div class="dashboard" :class="{ 'is-loading': loading }">
    <div class="header">
      <div class="user-info">
        <h1><i class="fas fa-user-circle"></i> Welcome {{ user?.username }}</h1>
        <p class="role"><i class="fas fa-id-badge"></i> Role: {{ user?.role_name }}</p>
        <p v-if="latestRate" class="rate-info">
          <i class="fas fa-chart-line"></i>
          The rate is {{ latestRate.rate }} on
          {{ new Date(latestRate.created_on).toLocaleString() }}
        </p>
      </div>
    </div>
    <LogoutButton />
    <div class="actions-section">
      <button
        v-if="canManageUsers"
        @click="handleUsersClick"
        class="action-btn users-btn"
        :disabled="isProcessing.users"
        :class="{ processing: isProcessing.users }"
      >
        <i class="fas fa-users"></i>
        <span>Users</span>
        <i v-if="isProcessing.users" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>
      <button
        v-if="canAccessTransfers"
        @click="handleTransfersClick"
        class="action-btn transfers-btn"
        :disabled="isProcessing.transfers"
        :class="{ processing: isProcessing.transfers }"
      >
        <i class="fas fa-exchange-alt"></i>
        <span>Transfers</span>
        <i v-if="isProcessing.transfers" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>
      <button
        v-if="canManageCars"
        @click="handleCarsClick"
        class="action-btn cars-btn"
        :disabled="isProcessing.cars"
        :class="{ processing: isProcessing.cars }"
      >
        <i class="fas fa-car"></i>
        <span>Cars</span>
        <i v-if="isProcessing.cars" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>
      <button
        v-if="canAccessCashier"
        @click="handleCashierClick"
        class="action-btn cashier-btn"
        :disabled="isProcessing.cashier"
        :class="{ processing: isProcessing.cashier }"
      >
        <i class="fas fa-cash-register"></i>
        <span>Cashier</span>
        <i v-if="isProcessing.cashier" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>
      <button
        v-if="canAccessRates"
        @click="handleRatesClick"
        class="action-btn rates-btn"
        :disabled="isProcessing.rates"
        :class="{ processing: isProcessing.rates }"
      >
        <i class="fas fa-percentage"></i>
        <span>Rates</span>
        <i v-if="isProcessing.rates" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>
      <button
        v-if="isAdmin"
        class="action-btn params-btn"
        :disabled="isProcessing.params"
        :class="{ processing: isProcessing.params }"
      >
        <i class="fas fa-cogs"></i>
        <span>Params</span>
        <i v-if="isProcessing.params" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>
    </div>
    <div class="copyright">© Merhab Noureddine 2025</div>
  </div>
</template>
<style scoped>
.dashboard {
  padding: 20px;
  position: relative;
}
.dashboard.is-loading {
  opacity: 0.7;
  pointer-events: none;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.user-info h1 {
  display: flex;
  align-items: center;
  gap: 10px;
}
.role {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
  margin-top: 5px;
}
.actions-section {
  margin: 20px 0;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.action-btn {
  padding: 15px 30px;
  border: none;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 160px;
  justify-content: center;
}
.action-btn:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}
.action-btn.processing {
  position: relative;
  pointer-events: none;
}
.action-btn.processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.2);
  border-radius: inherit;
}
.loading-indicator {
  margin-left: 4px;
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
.rates-btn {
  background-color: #9c27b0;
  color: white;
}
.params-btn {
  background-color: #6366f1;
  color: white;
}
button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
.rate-info {
  margin-top: 10px;
  padding: 12px 16px;
  background-color: #e3f2fd;
  border-radius: 8px;
  color: #1976d2;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.fa-spin {
  animation: spin 1s linear infinite;
}
/* Responsive adjustments */
@media (max-width: 768px) {
  .actions-section {
    flex-direction: column;
  }
  .action-btn {
    width: 100%;
  }
}
.copyright {
  text-align: center;
  color: #6b7280;
  font-size: 0.875rem;
  margin-top: 2rem;
  padding-bottom: 1rem;
}
</style>
