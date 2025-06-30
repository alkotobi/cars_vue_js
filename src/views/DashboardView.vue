<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
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
  tasks: false,
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
const handleTasksClick = async () => {
  if (isProcessing.value.tasks) return
  isProcessing.value.tasks = true
  try {
    await router.push('/tasks')
  } finally {
    isProcessing.value.tasks = false
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
  await Promise.all([fetchLatestRate(), fetchPendingTasks()])
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

// Add pending tasks functionality
const pendingTasks = ref([])
const pendingTasksLoading = ref(false)

const fetchPendingTasks = async () => {
  if (!user.value) return

  pendingTasksLoading.value = true
  try {
    const result = await callApi({
      query: `
        SELECT 
          t.id,
          t.title,
          t.desciption,
          t.id_priority as priority,
          t.date_create,
          t.notes,
          p.priority as priority_name,
          p.power as priority_power,
          u.username as creator_name
        FROM tasks t
        LEFT JOIN priorities p ON t.id_priority = p.id
        LEFT JOIN users u ON t.id_user_create = u.id
        WHERE t.date_declare_done IS NULL
        AND (t.id_user_receive = ? OR t.id_user_create = ?)
        ORDER BY t.date_create DESC
        LIMIT 5
      `,
      params: [user.value.id, user.value.id],
      requiresAuth: true,
    })

    if (result.success) {
      pendingTasks.value = result.data
    }
  } catch (error) {
    console.error('Error fetching pending tasks:', error)
  } finally {
    pendingTasksLoading.value = false
  }
}

const getPriorityBadge = (task) => {
  const power = task.priority_power || 1
  if (power >= 4) return { text: task.priority_name, class: 'priority-high' }
  if (power >= 2) return { text: task.priority_name, class: 'priority-medium' }
  return { text: task.priority_name, class: 'priority-low' }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString()
}
</script>
<template>
  <div class="dashboard" :class="{ 'is-loading': loading }">
    <div class="welcome-section">
      <div class="welcome-info">
        <h1><i class="fas fa-home"></i> Welcome to Dashboard</h1>
        <p v-if="latestRate" class="rate-info">
          <i class="fas fa-chart-line"></i>
          Current Rate: <strong>{{ latestRate.rate }}</strong> (Updated:
          {{ new Date(latestRate.created_on).toLocaleString() }})
        </p>
      </div>
    </div>

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
      <button
        @click="handleTasksClick"
        class="action-btn tasks-btn"
        :disabled="isProcessing.tasks"
        :class="{ processing: isProcessing.tasks }"
      >
        <i class="fas fa-tasks"></i>
        <span>Tasks</span>
        <i v-if="isProcessing.tasks" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>
    </div>

    <!-- Pending Tasks Section -->
    <div class="pending-tasks-section">
      <div class="section-header">
        <h2><i class="fas fa-clock"></i> Your Pending Tasks ({{ pendingTasks.length }})</h2>
        <div class="header-actions">
          <button @click="fetchPendingTasks" class="refresh-btn" title="Refresh tasks">
            <i class="fas fa-sync-alt"></i>
          </button>
          <button @click="handleTasksClick" class="view-all-btn">
            <i class="fas fa-external-link-alt"></i>
            View All Tasks
          </button>
        </div>
      </div>

      <div v-if="pendingTasksLoading" class="loading-tasks">
        <i class="fas fa-spinner fa-spin"></i>
        Loading your tasks...
      </div>

      <div v-else-if="pendingTasks.length === 0" class="no-tasks">
        <i class="fas fa-check-circle"></i>
        <p>No pending tasks! You're all caught up.</p>
      </div>

      <div v-else class="tasks-list">
        <div v-for="task in pendingTasks" :key="task.id" class="task-item">
          <div class="task-header">
            <h3 class="task-title">{{ task.title }}</h3>
            <span :class="getPriorityBadge(task).class" class="priority-badge">
              {{ getPriorityBadge(task).text }}
            </span>
          </div>
          <p v-if="task.desciption" class="task-description">{{ task.desciption }}</p>
          <div class="task-meta">
            <span class="task-date">
              <i class="fas fa-calendar"></i>
              {{ formatDate(task.date_create) }}
            </span>
            <span v-if="task.creator_name" class="task-creator">
              <i class="fas fa-user"></i>
              Created by {{ task.creator_name }}
            </span>
          </div>
          <p v-if="task.notes" class="task-notes">
            <i class="fas fa-sticky-note"></i>
            {{ task.notes }}
          </p>
        </div>
      </div>
    </div>

    <div class="copyright">© Merhab Noureddine 2025</div>
  </div>
</template>
<style scoped>
.dashboard {
  padding: 30px;
  position: relative;
  max-width: 1400px;
  margin: 0 auto;
}
.dashboard.is-loading {
  opacity: 0.7;
  pointer-events: none;
}
.welcome-section {
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 30px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border: 1px solid #e2e8f0;
}
.welcome-info h1 {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 10px 0;
  color: #1e293b;
  font-size: 1.8rem;
  font-weight: 700;
}
.welcome-info h1 i {
  color: #3b82f6;
}
.rate-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #64748b;
  font-size: 1rem;
}
.rate-info i {
  color: #10b981;
}
.rate-info strong {
  color: #059669;
  font-weight: 600;
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
.tasks-btn {
  background-color: #3b82f6;
  color: white;
}
button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
.copyright {
  text-align: center;
  color: #6b7280;
  font-size: 0.875rem;
  margin-top: 2rem;
  padding-bottom: 1rem;
}

/* Pending Tasks Section Styles */
.pending-tasks-section {
  background: white;
  border-radius: 12px;
  padding: 24px;
  margin: 20px 0;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  border: 1px solid #e5e7eb;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header h2 {
  margin: 0;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1.5rem;
}

.section-header h2 i {
  color: #f59e0b;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.refresh-btn {
  padding: 8px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.9rem;
  transition: background-color 0.2s ease;
}

.refresh-btn:hover {
  background-color: #2563eb;
}

.view-all-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.9rem;
  transition: background-color 0.2s ease;
}

.view-all-btn:hover {
  background-color: #2563eb;
}

.loading-tasks,
.no-tasks {
  text-align: center;
  padding: 40px 20px;
  color: #6b7280;
}

.loading-tasks i,
.no-tasks i {
  font-size: 2rem;
  margin-bottom: 12px;
  color: #10b981;
}

.no-tasks p {
  margin: 0;
  font-size: 1.1rem;
}

.tasks-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-item {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;
}

.task-item:hover {
  border-color: #d1d5db;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.task-title {
  margin: 0;
  font-size: 1.1rem;
  color: #1f2937;
  font-weight: 600;
  flex: 1;
}

.priority-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  white-space: nowrap;
  margin-left: 12px;
}

.priority-high {
  background-color: #fee2e2;
  color: #dc2626;
}

.priority-medium {
  background-color: #fef3c7;
  color: #d97706;
}

.priority-low {
  background-color: #d1fae5;
  color: #059669;
}

.task-description {
  margin: 8px 0;
  color: #4b5563;
  line-height: 1.5;
}

.task-meta {
  display: flex;
  gap: 16px;
  margin: 8px 0;
  font-size: 0.9rem;
  color: #6b7280;
}

.task-date,
.task-creator {
  display: flex;
  align-items: center;
  gap: 4px;
}

.task-notes {
  margin: 8px 0 0 0;
  padding: 8px 12px;
  background-color: #f3f4f6;
  border-radius: 4px;
  color: #4b5563;
  font-size: 0.9rem;
  display: flex;
  align-items: flex-start;
  gap: 6px;
}

.task-notes i {
  margin-top: 2px;
  color: #6b7280;
}
</style>
