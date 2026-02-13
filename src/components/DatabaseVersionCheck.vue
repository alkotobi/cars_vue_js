<template>
  <div v-if="hasVersionMismatch" class="version-modal-overlay">
    <div class="version-modal">
      <div class="version-modal-header">
        <i class="fas fa-exclamation-triangle"></i>
        <h2>{{ t('app.updateRequired') }}</h2>
      </div>

      <div class="version-modal-content">
        <p>{{ t('app.updateMessage') }}</p>

        <div class="version-info">
          <div class="version-item">
            <span class="version-label">{{ t('app.databaseName') || 'Database Name' }}:</span>
            <span class="version-value">{{ dbName || 'N/A' }}</span>
          </div>
          <div class="version-item">
            <span class="version-label">{{ t('app.databaseVersion') }}:</span>
            <span class="version-value">{{ dbVersion }}</span>
          </div>
          <div class="version-item">
            <span class="version-label">{{ t('app.applicationVersion') }}:</span>
            <span class="version-value">{{ currentAppVersion }}</span>
          </div>
        </div>

        <!-- Admin-only information -->
        <div v-if="isAdmin" class="admin-info">
          <div class="admin-info-header">
            <i class="fas fa-shield-alt"></i>
            <span>{{ t('app.adminInfo') || 'Admin Information' }}</span>
          </div>
          <div class="admin-info-item">
            <span class="admin-label">{{ t('app.directoryPath') || 'Directory Path' }}:</span>
            <span class="admin-value">{{ directoryPath || 'N/A' }}</span>
          </div>
        </div>

        <div class="version-warning">
          <i class="fas fa-info-circle"></i>
          <span>{{ t('app.refreshMessage') }}</span>
        </div>

        <div class="version-actions">
          <button @click="forceRefresh" class="btn-refresh">
            <i class="fas fa-sync-alt"></i>
            {{ t('app.refreshNow') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useVersionCheck } from '@/composables/useVersionCheck'
import { useApi } from '@/composables/useApi'

const { t } = useI18n()
const { currentAppVersion, dbVersion, isLoading, hasVersionMismatch, checkVersion, forceRefresh } =
  useVersionCheck()
const { loadDbName } = useApi()

const user = ref(null)
const dbName = ref('')
const directoryPath = ref('')

const isAdmin = computed(() => {
  return user.value?.role_id === 1
})

// Get base path where index.html exists (same logic as useApi.js)
const getBasePath = () => {
  let baseUrl = import.meta.env.BASE_URL || './'
  
  // If base is relative, convert to absolute path
  if (baseUrl === './' || baseUrl.startsWith('./')) {
    const pathname = window.location.pathname
    
    // Known route patterns that should NOT be treated as base paths
    const knownRoutes = [
      '/login', '/dashboard', '/users', '/roles', '/transfers', '/send', '/receive',
      '/sell-bills', '/buy-payments', '/params', '/advanced-sql', '/transfers-list',
      '/cars', '/warehouses', '/containers', '/print', '/clients', '/cashier',
      '/rates', '/tasks', '/statistics', '/chat', '/invitations', '/containers-ref',
      '/db-manager', '/alert-cars'
    ]
    
    // Check if pathname starts with a known route at root level (e.g., '/login' or '/db-manager')
    // But NOT if it's in a subdirectory (e.g., '/mig_26/db-manager' should extract '/mig_26/')
    const startsWithKnownRouteAtRoot = knownRoutes.some(route => {
      // Check if pathname exactly matches the route or starts with route followed by / or end of string
      return pathname === route || pathname.startsWith(route + '/') || pathname.startsWith(route + '?')
    })
    
    // If it starts with a known route at root level (not in subdirectory), base path is '/'
    if (startsWithKnownRouteAtRoot && !pathname.match(/^\/[^/]+\//)) {
      return '/'
    }
    
    // If pathname is like '/mig_26/login', extract '/mig_26/'
    // If pathname is like '/login', use '/'
    const match = pathname.match(/^(\/[^/]+\/)/)
    return match ? match[1] : '/'
  }
  
  // If base is already absolute, use it as is
  return baseUrl
}

onMounted(async () => {
  // Get user from localStorage
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }

  await checkVersion()

  // Always load database name for version dialog
  try {
    const name = await loadDbName()
    if (name) {
      dbName.value = name || 'N/A'
    }
  } catch (error) {
    console.error('Failed to load database config:', error)
    dbName.value = 'Error loading'
  }

  // Load directory path for admin users
  if (isAdmin.value) {
    directoryPath.value = getBasePath() || '/'
  }
})
</script>

<style scoped>
.version-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.version-modal {
  background: white;
  border-radius: 12px;
  padding: 32px;
  max-width: 500px;
  width: 90%;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  text-align: center;
}

.version-modal-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 24px;
  color: #dc2626;
}

.version-modal-header i {
  font-size: 28px;
}

.version-modal-header h2 {
  margin: 0;
  color: #dc2626;
  font-size: 24px;
}

.version-modal-content p {
  margin: 16px 0;
  color: #374151;
  line-height: 1.6;
  font-size: 16px;
}

.version-info {
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 20px;
  margin: 20px 0;
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 8px 0;
  padding: 8px 0;
}

.version-label {
  font-weight: 600;
  color: #374151;
}

.version-value {
  font-weight: 700;
  color: #1f2937;
  background: #e5e7eb;
  padding: 4px 12px;
  border-radius: 6px;
  font-family: monospace;
}

.version-warning {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin: 20px 0;
  padding: 16px;
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 8px;
  color: #92400e;
}

.version-warning i {
  color: #f59e0b;
  font-size: 18px;
}

.version-actions {
  margin-top: 24px;
}

.btn-refresh {
  padding: 16px 32px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 16px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 12px;
  transition: all 0.2s;
  background: #dc2626;
  color: white;
  box-shadow: 0 4px 6px rgba(220, 38, 38, 0.2);
}

.btn-refresh:hover {
  background: #b91c1c;
  transform: translateY(-1px);
  box-shadow: 0 6px 8px rgba(220, 38, 38, 0.3);
}

.btn-refresh:active {
  transform: translateY(0);
}

.admin-info {
  background: #eff6ff;
  border: 1px solid #3b82f6;
  border-radius: 8px;
  padding: 16px;
  margin: 20px 0;
  text-align: left;
}

.admin-info-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-weight: 600;
  color: #1e40af;
  font-size: 14px;
}

.admin-info-header i {
  color: #3b82f6;
  font-size: 16px;
}

.admin-info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin: 8px 0;
  padding: 8px 0;
}

.admin-label {
  font-weight: 600;
  color: #374151;
  font-size: 13px;
}

.admin-value {
  font-weight: 500;
  color: #1f2937;
  background: #dbeafe;
  padding: 6px 12px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
}
</style>
