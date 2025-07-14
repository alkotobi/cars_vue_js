<template>
  <div class="alerts-view" v-if="hasAlerts">
    <div class="alerts-container">
      <!-- Unloaded Cars Alert -->
      <div v-if="unloadedCount > 0" class="alert-badge unloaded" @click="handleUnloadedClick">
        <i class="fas fa-truck"></i>
        <span class="badge-text"
          >{{ unloadedCount }} cars unloaded after
          {{ defaults?.alert_unloaded_after_days || 0 }} days from sell date</span
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
          >{{ notArrivedCount }} cars not arrived after
          {{ defaults?.alert_not_arrived_after_days || 0 }} days from buy date</span
        >
      </div>

      <!-- No License Cars Alert -->
      <div v-if="noLicenseCount > 0" class="alert-badge no-license" @click="handleNoLicenseClick">
        <i class="fas fa-passport"></i>
        <span class="badge-text"
          >{{ noLicenseCount }} cars no license after
          {{ defaults?.alert_no_licence_after_days || 0 }} days from buy date</span
        >
      </div>

      <!-- No Docs Sent Cars Alert -->
      <div v-if="noDocsSentCount > 0" class="alert-badge no-docs" @click="handleNoDocsClick">
        <i class="fas fa-file-alt"></i>
        <span class="badge-text"
          >{{ noDocsSentCount }} cars no docs sent after
          {{ defaults?.alert_no_docs_sent_after_days || 0 }} days from sell date</span
        >
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi } = useApi()

const defaults = ref(null)
const unloadedCount = ref(0)
const notArrivedCount = ref(0)
const noLicenseCount = ref(0)
const noDocsSentCount = ref(0)
const loading = ref(false)

// Computed property to check if there are any alerts
const hasAlerts = computed(() => {
  return (
    unloadedCount.value > 0 ||
    notArrivedCount.value > 0 ||
    noLicenseCount.value > 0 ||
    noDocsSentCount.value > 0
  )
})

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
        WHERE cs.in_wharhouse_date IS NULL
        AND cs.date_send_documents IS NULL
        AND cs.hidden = 0
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
        WHERE (cs.export_lisence_ref IS NULL OR cs.export_lisence_ref = '')
        AND cs.date_send_documents IS NULL
        AND cs.hidden = 0
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

// Fetch all alert data
const fetchAlertData = async () => {
  loading.value = true
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
  } catch (error) {
    console.error('Error fetching alert data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchAlertData()
})

// Click handlers for alert badges
const handleUnloadedClick = () => {
  const url = `/alert-cars/unloaded/${defaults.value?.alert_unloaded_after_days || 0}`
  window.open(url, '_blank')
}

const handleNotArrivedClick = () => {
  const url = `/alert-cars/not_arrived/${defaults.value?.alert_not_arrived_after_days || 0}`
  window.open(url, '_blank')
}

const handleNoLicenseClick = () => {
  const url = `/alert-cars/no_licence/${defaults.value?.alert_no_licence_after_days || 0}`
  window.open(url, '_blank')
}

const handleNoDocsClick = () => {
  const url = `/alert-cars/no_docs_sent/${defaults.value?.alert_no_docs_sent_after_days || 0}`
  window.open(url, '_blank')
}
</script>

<style scoped>
.alerts-view {
  width: 100%;
  padding: 10px 20px;
  background-color: #fef2f2;
  border-bottom: 1px solid #fecaca;
}

.alerts-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
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

@media (max-width: 768px) {
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
