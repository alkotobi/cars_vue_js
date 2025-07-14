<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import BanksTable from '../components/params/BanksTable.vue'
import DefaultsForm from '../components/params/DefaultsForm.vue'
import AlertSettingsForm from '../components/params/AlertSettingsForm.vue'

const router = useRouter()
const loading = ref(false)
const error = ref(null)
const activeSection = ref('general')
const isProcessing = ref({
  dashboard: false,
  general: false,
  company: false,
  billing: false,
  alerts: false,
  notifications: false,
  security: false,
  advanced_sql: false,
})

// Check if user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Redirect if not admin
onMounted(() => {
  if (!isAdmin.value) {
    router.push('/dashboard')
    return
  }
})

const returnToDashboard = async () => {
  if (isProcessing.value.dashboard) return
  isProcessing.value.dashboard = true
  try {
    await router.push('/dashboard')
  } finally {
    isProcessing.value.dashboard = false
  }
}

const setActiveSection = (section) => {
  if (isProcessing.value[section]) return
  isProcessing.value[section] = true
  try {
    activeSection.value = section
  } finally {
    isProcessing.value[section] = false
  }
}

const navigateToAdvancedSql = () => {
  if (isProcessing.value.advanced_sql) return
  isProcessing.value.advanced_sql = true
  try {
    router.push('/advanced-sql')
  } finally {
    isProcessing.value.advanced_sql = false
  }
}
</script>

<template>
  <div class="params-view" :class="{ 'is-loading': loading }">
    <div class="sidebar">
      <div class="sidebar-header">
        <h2>
          <i class="fas fa-cogs"></i>
          Parameters
        </h2>
        <button
          @click="returnToDashboard"
          class="sidebar-btn return-btn"
          :disabled="isProcessing.dashboard"
          :class="{ processing: isProcessing.dashboard }"
        >
          <i class="fas fa-arrow-left"></i>
          <span>Return to Dashboard</span>
          <i v-if="isProcessing.dashboard" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>

      <div class="sidebar-content">
        <button
          @click="setActiveSection('general')"
          class="sidebar-btn"
          :class="{ active: activeSection === 'general', processing: isProcessing.general }"
          :disabled="isProcessing.general"
        >
          <i class="fas fa-sliders-h"></i>
          <span>General Settings</span>
          <i v-if="isProcessing.general" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>

        <button
          @click="setActiveSection('company')"
          class="sidebar-btn"
          :class="{ active: activeSection === 'company', processing: isProcessing.company }"
          :disabled="isProcessing.company"
        >
          <i class="fas fa-building"></i>
          <span>Company Info</span>
          <i v-if="isProcessing.company" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>

        <button
          @click="setActiveSection('billing')"
          class="sidebar-btn"
          :class="{ active: activeSection === 'billing', processing: isProcessing.billing }"
          :disabled="isProcessing.billing"
        >
          <i class="fas fa-file-invoice-dollar"></i>
          <span>Billing Settings</span>
          <i v-if="isProcessing.billing" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>

        <button
          @click="setActiveSection('alerts')"
          class="sidebar-btn"
          :class="{ active: activeSection === 'alerts', processing: isProcessing.alerts }"
          :disabled="isProcessing.alerts"
        >
          <i class="fas fa-bell"></i>
          <span>Alert Settings</span>
          <i v-if="isProcessing.alerts" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>

        <button
          @click="setActiveSection('notifications')"
          class="sidebar-btn"
          :class="{
            active: activeSection === 'notifications',
            processing: isProcessing.notifications,
          }"
          :disabled="isProcessing.notifications"
        >
          <i class="fas fa-bell"></i>
          <span>Notifications</span>
          <i v-if="isProcessing.notifications" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>

        <button
          @click="setActiveSection('security')"
          class="sidebar-btn"
          :class="{ active: activeSection === 'security', processing: isProcessing.security }"
          :disabled="isProcessing.security"
        >
          <i class="fas fa-shield-alt"></i>
          <span>Security</span>
          <i v-if="isProcessing.security" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>

        <button
          @click="navigateToAdvancedSql"
          class="sidebar-btn"
          :class="{
            processing: isProcessing.advanced_sql,
          }"
          :disabled="isProcessing.advanced_sql"
        >
          <i class="fas fa-database"></i>
          <span>Advanced SQL</span>
          <i v-if="isProcessing.advanced_sql" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <div class="main-content">
      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <!-- Content area -->
      <div class="content">
        <div v-if="activeSection === 'company'">
          <BanksTable />
        </div>
        <div v-else-if="activeSection === 'billing'">
          <DefaultsForm />
        </div>
        <div v-else-if="activeSection === 'alerts'">
          <AlertSettingsForm />
        </div>
        <div v-else class="empty-state">
          <i class="fas fa-tools fa-3x"></i>
          <p>{{ activeSection }} settings configuration coming soon...</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.params-view {
  display: flex;
  min-height: 100vh;
  background-color: #f9fafb;
}

.params-view.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.sidebar {
  width: 280px;
  background-color: #1f2937;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  box-shadow: 4px 0 8px rgba(0, 0, 0, 0.1);
}

.sidebar-header {
  border-bottom: 2px solid #374151;
  padding-bottom: 20px;
}

.sidebar-header h2 {
  color: white;
  font-size: 1.5rem;
  margin: 0 0 20px 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.sidebar-header h2 i {
  color: #6366f1;
}

.sidebar-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  border: none;
  border-radius: 8px;
  background-color: transparent;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  text-align: left;
}

.sidebar-btn i:not(.loading-indicator) {
  width: 20px;
  text-align: center;
}

.sidebar-btn span {
  flex: 1;
}

.sidebar-btn:not(:disabled):hover {
  background-color: #374151;
  transform: translateX(4px);
}

.sidebar-btn.active {
  background-color: #6366f1;
}

.sidebar-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.return-btn {
  background-color: #4b5563;
}

.return-btn:not(:disabled):hover {
  background-color: #374151;
}

.main-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.content {
  background-color: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  min-height: calc(100vh - 48px);
}

.error-message {
  padding: 16px;
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  border-radius: 8px;
  color: #b91c1c;
  margin-bottom: 24px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  color: #6b7280;
  text-align: center;
}

.empty-state i {
  margin-bottom: 16px;
  color: #9ca3af;
}

.empty-state p {
  font-size: 1.1rem;
  margin: 0;
  text-transform: capitalize;
}

.loading-indicator {
  margin-left: auto;
  font-size: 0.9em;
}

.sidebar-btn.processing {
  position: relative;
  pointer-events: none;
}

.sidebar-btn.processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fa-spin {
  animation: spin 1s linear infinite;
}

.advanced-sql-section {
  padding: 24px;
}

.advanced-sql-section h2 {
  color: #1f2937;
  font-size: 1.75rem;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.advanced-sql-section h2 i {
  color: #6366f1;
}

.advanced-sql-section p {
  color: #6b7280;
  font-size: 1.1rem;
  margin-bottom: 24px;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .params-view {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    padding: 16px;
  }

  .main-content {
    padding: 16px;
  }

  .content {
    min-height: auto;
  }

  .sidebar-btn {
    padding: 12px;
  }
}
</style>
