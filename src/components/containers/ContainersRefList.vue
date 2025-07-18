<template>
  <div class="containers-ref-list">
    <div class="header">
      <h2>{{ t('containersRef.title') }}</h2>
      <div class="header-actions">
        <button @click="refreshData" :disabled="loading" class="btn btn-primary">
          <i class="fas fa-sync-alt"></i>
          {{ t('containersRef.refresh') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <i class="fas fa-spinner fa-spin"></i>
      {{ t('containersRef.loading') }}
    </div>

    <div v-else-if="error" class="error">
      <i class="fas fa-exclamation-triangle"></i>
      {{ error }}
    </div>

    <div v-else-if="containersRef.length === 0" class="no-data">
      <i class="fas fa-box-open"></i>
      {{ t('containersRef.noContainers') }}
    </div>

    <div v-else class="table-container">
      <div class="stats">
        <div class="stat-item">
          <span class="stat-label">{{ t('containersRef.totalContainers') }}:</span>
          <span class="stat-value">{{ containersRef.length }}</span>
        </div>
      </div>

      <div class="table-wrapper">
        <table class="containers-table">
          <thead>
            <tr>
              <th>#</th>
              <th>{{ t('containersRef.containerRef') }}</th>
              <th>{{ t('containersRef.location') }}</th>
              <th>{{ t('containersRef.timestamp') }}</th>
              <th>{{ t('containersRef.user') }}</th>
              <th>{{ t('containersRef.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(container, index) in containersRef" :key="index" class="table-row">
              <td class="id-cell">{{ index + 1 }}</td>
              <td class="ref-cell">
                <div class="ref-content">
                  <i class="fas fa-box"></i>
                  {{ container.container_ref }}
                </div>
              </td>
              <td class="location-cell">
                <div v-if="container.tracking" class="location-info">
                  <i class="fas fa-map-marker-alt"></i>
                  <span class="coordinates">{{ container.tracking }}</span>
                </div>
                <div v-else class="no-location">
                  <i class="fas fa-question-circle"></i>
                  <span>{{ t('containersRef.noLocation') }}</span>
                </div>
              </td>
              <td class="timestamp-cell">
                <div v-if="container.time" class="timestamp-info">
                  <i class="fas fa-clock"></i>
                  <span>{{ formatTimestamp(container.time) }}</span>
                </div>
                <div v-else class="no-timestamp">
                  <span>-</span>
                </div>
              </td>
              <td class="user-cell">
                <div v-if="container.username" class="user-info">
                  <i class="fas fa-user"></i>
                  <span>{{ container.username }}</span>
                </div>
                <div v-else-if="container.id_user" class="user-info">
                  <i class="fas fa-user"></i>
                  <span>User #{{ container.id_user }}</span>
                </div>
                <div v-else class="no-user">
                  <span>-</span>
                </div>
              </td>
              <td class="actions-cell">
                <button @click="trackContainer(container.container_ref)" class="track-btn">
                  <i class="fas fa-map-marker-alt"></i>
                  {{ t('containersRef.track') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Google Maps Popup -->
    <GoogleMapPopup
      v-if="showMapPopup"
      :is-visible="showMapPopup"
      :container-ref="selectedContainerRef"
      @close="showMapPopup = false"
      @location-selected="handleLocationSelected"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useApi } from '@/composables/useApi'
import GoogleMapPopup from './GoogleMapPopup.vue'

const { t } = useI18n()
const { callApi } = useApi()

const containersRef = ref([])
const loading = ref(false)
const error = ref(null)
const showMapPopup = ref(false)
const selectedContainerRef = ref('')

const fetchContainersRef = async () => {
  loading.value = true
  error.value = null

  try {
    // First get unique container references
    const containersResult = await callApi({
      action: 'execute_sql',
      query:
        'SELECT DISTINCT container_ref FROM cars_stock WHERE container_ref IS NOT NULL ORDER BY container_ref',
    })

    if (containersResult.success) {
      const containerRefs = containersResult.results || []

      // Get tracking data for each container reference
      const trackingPromises = containerRefs.map(async (container) => {
        const trackingResult = await callApi({
          action: 'execute_sql',
          query: `
            SELECT t.*, u.username 
            FROM tracking t 
            LEFT JOIN users u ON t.id_user = u.id 
            WHERE t.container_ref = ? 
            ORDER BY t.time DESC 
            LIMIT 1
          `,
          params: [container.container_ref],
        })

        if (trackingResult.success && trackingResult.results.length > 0) {
          return {
            ...container,
            ...trackingResult.results[0],
          }
        } else {
          return {
            ...container,
            tracking: null,
            time: null,
            id_user: null,
            username: null,
          }
        }
      })

      containersRef.value = await Promise.all(trackingPromises)
    } else {
      error.value = containersResult.message || t('containersRef.failedToLoad')
    }
  } catch (err) {
    error.value = t('containersRef.errorLoading', { message: err.message })
  } finally {
    loading.value = false
  }
}

const refreshData = () => {
  fetchContainersRef()
}

const trackContainer = (containerRef) => {
  selectedContainerRef.value = containerRef
  showMapPopup.value = true
}

const handleLocationSelected = (location) => {
  // Refresh the data to show the updated tracking information
  fetchContainersRef()
}

const formatContainerRef = (containerRef) => {
  if (!containerRef) return '-'

  // Show only first 4 characters, replace the rest with asterisks
  if (containerRef.length <= 4) {
    return containerRef
  }

  return containerRef.substring(0, 4) + '*'.repeat(containerRef.length - 4)
}

const formatTimestamp = (timestamp) => {
  if (!timestamp) return '-'

  try {
    const date = new Date(timestamp)
    return date.toLocaleString()
  } catch (error) {
    return timestamp
  }
}

onMounted(() => {
  fetchContainersRef()
})
</script>

<style scoped>
.containers-ref-list {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 15px;
  border-bottom: 2px solid #e0e0e0;
}

.header h2 {
  margin: 0;
  color: #333;
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-primary {
  background-color: #007bff;
  color: white;
}

.btn-primary:hover {
  background-color: #0056b3;
}

.btn-primary:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}

.loading,
.error,
.no-data {
  text-align: center;
  padding: 40px;
  font-size: 16px;
  color: #666;
}

.loading i,
.error i,
.no-data i {
  font-size: 24px;
  margin-bottom: 10px;
  display: block;
}

.error {
  color: #dc3545;
}

.no-data {
  color: #6c757d;
}

.table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.stats {
  background: #f8f9fa;
  padding: 15px 20px;
  border-bottom: 1px solid #e0e0e0;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.stat-label {
  font-weight: 600;
  color: #495057;
}

.stat-value {
  font-weight: 700;
  color: #007bff;
  font-size: 18px;
}

.table-wrapper {
  overflow-x: auto;
}

.containers-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.containers-table th {
  background-color: #f8fafc;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
}

.containers-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
}

.table-row:hover {
  background-color: #f9fafb;
  transition: background-color 0.2s ease;
}

.id-cell {
  font-weight: 600;
  color: #6b7280;
  width: 60px;
}

.ref-cell {
  font-family: monospace;
}

.ref-content {
  display: flex;
  align-items: center;
  gap: 6px;
}

.location-cell {
  min-width: 200px;
}

.location-info {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #059669;
}

.location-info i {
  color: #059669;
}

.coordinates {
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
}

.no-location {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  font-style: italic;
}

.no-location i {
  color: #9ca3af;
}

.timestamp-cell {
  min-width: 150px;
}

.timestamp-info {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #374151;
}

.timestamp-info i {
  color: #6b7280;
}

.no-timestamp {
  color: #9ca3af;
}

.user-cell {
  min-width: 80px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #374151;
}

.user-info i {
  color: #6b7280;
}

.no-user {
  color: #9ca3af;
}

.actions-cell {
  width: 120px;
}

.track-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.track-btn:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.track-btn:active {
  transform: translateY(0);
}

@media (max-width: 768px) {
  .header {
    flex-direction: column;
    gap: 15px;
    align-items: flex-start;
  }

  .containers-ref-list {
    padding: 15px;
  }

  .table-wrapper {
    font-size: 0.8rem;
  }

  .containers-table th,
  .containers-table td {
    padding: 8px 4px;
  }
}
</style>
