<template>
  <div class="alerts-view" v-if="hasAlerts" :class="{ refreshing: isRefreshing }">
    <div class="alerts-container">
      <div class="alerts-badges">
        <!-- Not Loaded Cars Alert -->
        <div v-if="unloadedCount > 0" class="alert-badge unloaded" @click="handleUnloadedClick">
          <i class="fas fa-truck"></i>
          <span class="badge-text"
            >{{ unloadedCount }} {{ t('alerts.carsNotLoaded') }}
            {{ defaults?.alert_unloaded_after_days || 0 }} {{ t('alerts.daysFromSellDate') }}</span
          >
        </div>

        <!-- Not Arrived Cars Alert -->
        <div
          v-if="notArrivedCount > 0"
          class="alert-badge not-arrived"
          @click="handleNotArrivedClick"
        >
          <i class="fas fa-ship"></i>
          <span class="badge-text"
            >{{ notArrivedCount }} {{ t('alerts.carsNotArrived') }}
            {{ defaults?.alert_not_arrived_after_days || 0 }}
            {{ t('alerts.daysFromBuyDate') }}</span
          >
        </div>

        <!-- No License Cars Alert -->
        <div v-if="noLicenseCount > 0" class="alert-badge no-license" @click="handleNoLicenseClick">
          <i class="fas fa-passport"></i>
          <span class="badge-text"
            >{{ noLicenseCount }} {{ t('alerts.carsNoLicense') }}
            {{ defaults?.alert_no_licence_after_days || 0 }} {{ t('alerts.daysFromBuyDate') }}</span
          >
        </div>

        <!-- No Docs Sent Cars Alert -->
        <div v-if="noDocsSentCount > 0" class="alert-badge no-docs" @click="handleNoDocsClick">
          <i class="fas fa-file-alt"></i>
          <span class="badge-text"
            >{{ noDocsSentCount }} {{ t('alerts.carsNoDocsSent') }}
            {{ defaults?.alert_no_docs_sent_after_days || 0 }}
            {{ t('alerts.daysFromSellDate') }}</span
          >
        </div>
      </div>

      <!-- Manual refresh button -->
      <button
        @click="handleManualRefresh"
        class="refresh-button"
        :disabled="isRefreshing"
        :title="isRefreshing ? t('alerts.refreshing') : t('alerts.refreshAlerts')"
      >
        <i class="fas fa-sync-alt" :class="{ spinning: isRefreshing }"></i>
        <span>{{ isRefreshing ? t('alerts.refreshing') : t('alerts.refresh') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useApi, BASE_PATH } from '../composables/useApi'

const { t } = useI18n()
const { callApi } = useApi()

const defaults = ref(null)
const unloadedCount = ref(0)
const notArrivedCount = ref(0)
const noLicenseCount = ref(0)
const noDocsSentCount = ref(0)
const loading = ref(false)

// Auto-refresh functionality
const refreshInterval = ref(null)
const isRefreshing = ref(false)
const lastRefreshTime = ref(null)

// Computed property to check if there are any alerts
const hasAlerts = computed(() => {
  return (
    unloadedCount.value > 0 ||
    notArrivedCount.value > 0 ||
    noLicenseCount.value > 0 ||
    noDocsSentCount.value > 0
  )
})

// Format last refresh time
const formatLastRefreshTime = () => {
  if (!lastRefreshTime.value) return ''
  const now = new Date()
  const diff = now - lastRefreshTime.value
  const minutes = Math.floor(diff / 60000)
  const seconds = Math.floor((diff % 60000) / 1000)

  if (minutes > 0) {
    return `${minutes}m ${seconds}s ago`
  }
  return `${seconds}s ago`
}

// Fetch defaults settings
const fetchDefaults = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM defaults LIMIT 1',
      params: [],
    })
    if (result.success && result.data.length > 0) {
      defaults.value = result.data[0]
    }
  } catch (error) {
    console.error('Error fetching defaults:', error)
  }
}

// Fetch unloaded cars count
const fetchUnloadedCount = async () => {
  if (!defaults.value?.alert_unloaded_after_days) return

  try {
    const result = await callApi({
      query: `
        SELECT COUNT(*) as count
        FROM cars_stock cs
        INNER JOIN sell_bill sb ON cs.id_sell = sb.id
        WHERE cs.date_loding IS NULL
        AND cs.date_send_documents IS NULL
        AND cs.hidden = 0
        AND cs.is_batch = 0
        AND sb.is_batch_sell = 0
        AND (cs.container_ref IS NULL OR cs.container_ref = '')
        AND sb.date_sell < DATE_SUB(NOW(), INTERVAL ? DAY)
      `,
      params: [defaults.value.alert_unloaded_after_days],
    })
    if (result.success && result.data.length > 0) {
      unloadedCount.value = result.data[0].count
    }
  } catch (error) {
    console.error('Error fetching unloaded count:', error)
  }
}

// Fetch not arrived cars count
const fetchNotArrivedCount = async () => {
  if (!defaults.value?.alert_not_arrived_after_days) return

  try {
    const result = await callApi({
      query: `
        SELECT COUNT(*) as count
        FROM cars_stock cs
        INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
        INNER JOIN buy_bill bb ON bd.id_buy_bill = bb.id
        WHERE cs.date_loding IS NULL
        AND cs.in_wharhouse_date IS NULL
        AND cs.date_send_documents IS NULL
        AND cs.hidden = 0
        AND cs.is_batch = 0
        AND (cs.container_ref IS NULL OR cs.container_ref = '')
        AND bb.date_buy < DATE_SUB(NOW(), INTERVAL ? DAY)
      `,
      params: [defaults.value.alert_not_arrived_after_days],
    })
    if (result.success && result.data.length > 0) {
      notArrivedCount.value = result.data[0].count
    }
  } catch (error) {
    console.error('Error fetching not arrived count:', error)
  }
}

// Fetch no license cars count
const fetchNoLicenseCount = async () => {
  if (!defaults.value?.alert_no_licence_after_days) return

  try {
    const result = await callApi({
      query: `
        SELECT COUNT(*) as count
        FROM cars_stock cs
        INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
        INNER JOIN buy_bill bb ON bd.id_buy_bill = bb.id
        WHERE cs.date_loding IS NULL
        AND (cs.export_lisence_ref IS NULL OR cs.export_lisence_ref = '')
        AND cs.date_send_documents IS NULL
        AND cs.hidden = 0
        AND cs.is_batch = 0
        AND (cs.container_ref IS NULL OR cs.container_ref = '')
        AND bb.date_buy < DATE_SUB(NOW(), INTERVAL ? DAY)
      `,
      params: [defaults.value.alert_no_licence_after_days],
    })
    if (result.success && result.data.length > 0) {
      noLicenseCount.value = result.data[0].count
    }
  } catch (error) {
    console.error('Error fetching no license count:', error)
  }
}

// Fetch no docs sent cars count
const fetchNoDocsSentCount = async () => {
  if (!defaults.value?.alert_no_docs_sent_after_days) return

  try {
    const result = await callApi({
      query: `
        SELECT COUNT(*) as count
        FROM cars_stock cs
        INNER JOIN sell_bill sb ON cs.id_sell = sb.id
        WHERE cs.date_send_documents IS NULL
        AND cs.hidden = 0
        AND cs.is_batch = 0
        AND sb.is_batch_sell = 0
        AND sb.date_sell < DATE_SUB(NOW(), INTERVAL ? DAY)
      `,
      params: [defaults.value.alert_no_docs_sent_after_days],
    })
    if (result.success && result.data.length > 0) {
      noDocsSentCount.value = result.data[0].count
    }
  } catch (error) {
    console.error('Error fetching no docs sent count:', error)
  }
}

// Fetch all alert data with refresh indicator
const fetchAlertData = async (isAutoRefresh = false) => {
  if (isAutoRefresh) {
    isRefreshing.value = true
  } else {
    loading.value = true
  }

  try {
    await fetchDefaults()
    if (defaults.value) {
      await Promise.all([
        fetchUnloadedCount(),
        fetchNotArrivedCount(),
        fetchNoLicenseCount(),
        fetchNoDocsSentCount(),
      ])
    }
    lastRefreshTime.value = new Date()
  } catch (error) {
    console.error('Error fetching alert data:', error)
  } finally {
    loading.value = false
    isRefreshing.value = false
  }
}

// Start auto-refresh interval
const startAutoRefresh = () => {
  // Clear any existing interval
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }

  // Set up new interval (15 minutes = 900000 milliseconds)
  refreshInterval.value = setInterval(() => {
    fetchAlertData(true)
  }, 900000)
}

// Manual refresh handler
const handleManualRefresh = () => {
  if (!isRefreshing.value) {
    fetchAlertData(true)
    // Add visual effect to badges after manual refresh
    setTimeout(() => {
      const badges = document.querySelectorAll('.alert-badge')
      badges.forEach((badge) => {
        badge.classList.add('refresh-effect')
        setTimeout(() => {
          badge.classList.remove('refresh-effect')
        }, 1000)
      })
    }, 100) // Small delay to ensure data is updated
  }
}

// Stop auto-refresh interval
const stopAutoRefresh = () => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
    refreshInterval.value = null
  }
}

onMounted(() => {
  fetchAlertData()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})

// Click handlers for alert badges
const handleUnloadedClick = () => {
  const url = `${BASE_PATH}alert-cars/unloaded/${defaults.value?.alert_unloaded_after_days || 0}`
  window.open(url, '_blank')
}

const handleNotArrivedClick = () => {
  const url = `${BASE_PATH}alert-cars/not_arrived/${defaults.value?.alert_not_arrived_after_days || 0}`
  window.open(url, '_blank')
}

const handleNoLicenseClick = () => {
  const url = `${BASE_PATH}alert-cars/no_licence/${defaults.value?.alert_no_licence_after_days || 0}`
  window.open(url, '_blank')
}

const handleNoDocsClick = () => {
  const url = `${BASE_PATH}alert-cars/no_docs_sent/${defaults.value?.alert_no_docs_sent_after_days || 0}`
  window.open(url, '_blank')
}
</script>

<style scoped>
.alerts-view {
  width: 100%;
  padding: 10px 20px;
  background-color: #fef2f2;
  border-bottom: 1px solid #fecaca;
  transition: all 0.3s ease;
  position: fixed;
  top: 70px;
  left: 0;
  right: 0;
  z-index: 999;
}

.alerts-view.refreshing {
  background-color: #fef7f0;
  border-bottom-color: #fbbf24;
  box-shadow: 0 0 10px rgba(251, 191, 36, 0.3);
}

.alerts-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.alerts-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  flex: 1;
}

.refresh-button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background-color: transparent;
  color: #6b7280;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 400;
  cursor: pointer;
  transition: all 0.2s ease;
  height: 32px;
  white-space: nowrap;
  opacity: 0.7;
}

.refresh-button:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #374151;
  border-color: #9ca3af;
  opacity: 1;
  transform: none;
  box-shadow: none;
}

.refresh-button:disabled {
  background-color: transparent;
  color: #9ca3af;
  border-color: #e5e7eb;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
  opacity: 0.5;
}

.refresh-button i {
  font-size: 12px;
  transition: transform 0.3s ease;
}

.refresh-button i.spinning {
  animation: spin 1s linear infinite;
}

.alert-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.alert-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.alert-badge.refresh-effect {
  animation: refreshPulse 1s ease-in-out;
}

@keyframes refreshPulse {
  0% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.05);
    opacity: 0.8;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

.alert-badge.unloaded {
  background-color: #fef3c7;
  color: #d97706;
  border: 1px solid #fbbf24;
}

.alert-badge.not-arrived {
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fca5a5;
}

.alert-badge.no-license {
  background-color: #f0f9ff;
  color: #0369a1;
  border: 1px solid #7dd3fc;
}

.alert-badge.no-docs {
  background-color: #fdf4ff;
  color: #a855f7;
  border: 1px solid #d8b4fe;
}

.alert-badge i {
  font-size: 14px;
}

.badge-text {
  white-space: nowrap;
}

.refresh-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 8px;
  font-size: 11px;
  color: #6b7280;
  opacity: 0.8;
}

.refresh-indicator i {
  font-size: 12px;
  transition: transform 0.3s ease;
}

.refresh-indicator i.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .alerts-view {
    top: 60px; /* Smaller header on mobile */
  }

  .alerts-container {
    flex-direction: column;
    align-items: center;
  }

  .alert-badge {
    width: 100%;
    max-width: 300px;
    justify-content: center;
  }
}
</style>
