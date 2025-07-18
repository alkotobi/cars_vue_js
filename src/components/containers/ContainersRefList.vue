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
      <!-- Filters -->
      <div class="filters-section">
        <div class="filters-header">
          <h3>{{ t('containersRef.filters') }}</h3>
          <button @click="resetFilters" class="btn btn-secondary">
            <i class="fas fa-times"></i>
            {{ t('containersRef.clearFilters') }}
          </button>
        </div>
        <div class="filters-grid">
          <div class="filter-item">
            <label>{{ t('containersRef.containerRef') }}:</label>
            <input
              v-model="filters.containerRef"
              type="text"
              :placeholder="t('containersRef.searchContainerRef')"
              class="filter-input"
            />
          </div>
          <div class="filter-item">
            <label>{{ t('containersRef.location') }}:</label>
            <input
              v-model="filters.location"
              type="text"
              :placeholder="t('containersRef.searchLocation')"
              class="filter-input"
            />
          </div>
          <div class="filter-item">
            <label>{{ t('containersRef.user') }}:</label>
            <input
              v-model="filters.user"
              type="text"
              :placeholder="t('containersRef.searchUser')"
              class="filter-input"
            />
          </div>
          <div class="filter-item">
            <label>{{ t('containersRef.hasLocation') }}:</label>
            <select v-model="filters.hasLocation" class="filter-select">
              <option value="">{{ t('containersRef.all') }}</option>
              <option value="yes">{{ t('containersRef.yes') }}</option>
              <option value="no">{{ t('containersRef.no') }}</option>
            </select>
          </div>
        </div>
      </div>
      <div class="stats">
        <div class="stat-item">
          <span class="stat-label">{{ t('containersRef.totalContainers') }}:</span>
          <span class="stat-value">{{ containersRef.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">{{ t('containersRef.filteredContainers') }}:</span>
          <span class="stat-value">{{ filteredAndSortedContainers.length }}</span>
        </div>
      </div>

      <div class="table-wrapper">
        <table class="containers-table">
          <thead>
            <tr>
              <th>#</th>
              <th @click="handleSort('container_ref')" class="sortable">
                {{ t('containersRef.containerRef') }}
                <i
                  v-if="sortConfig.field === 'container_ref'"
                  :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
                ></i>
              </th>
              <th @click="handleSort('tracking')" class="sortable">
                {{ t('containersRef.location') }}
                <i
                  v-if="sortConfig.field === 'tracking'"
                  :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
                ></i>
              </th>
              <th @click="handleSort('time')" class="sortable">
                {{ t('containersRef.timestamp') }}
                <i
                  v-if="sortConfig.field === 'time'"
                  :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
                ></i>
              </th>
              <th @click="handleSort('username')" class="sortable">
                {{ t('containersRef.user') }}
                <i
                  v-if="sortConfig.field === 'username'"
                  :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
                ></i>
              </th>
              <th>{{ t('containersRef.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(container, index) in filteredAndSortedContainers"
              :key="container.container_ref"
              class="table-row"
            >
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
import { ref, onMounted, computed } from 'vue'
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

// Filter and sort state
const filters = ref({
  containerRef: '',
  location: '',
  user: '',
  hasLocation: '', // 'all', 'yes', 'no'
})

const sortConfig = ref({
  field: 'container_ref',
  direction: 'asc',
})

// Filtered and sorted containers
const filteredAndSortedContainers = computed(() => {
  let result = [...containersRef.value]

  // Apply filters
  if (filters.value.containerRef) {
    result = result.filter((container) =>
      container.container_ref?.toLowerCase().includes(filters.value.containerRef.toLowerCase()),
    )
  }

  if (filters.value.location) {
    result = result.filter((container) =>
      container.tracking?.toLowerCase().includes(filters.value.location.toLowerCase()),
    )
  }

  if (filters.value.user) {
    result = result.filter(
      (container) =>
        container.username?.toLowerCase().includes(filters.value.user.toLowerCase()) ||
        container.id_user?.toString().includes(filters.value.user),
    )
  }

  if (filters.value.hasLocation) {
    if (filters.value.hasLocation === 'yes') {
      result = result.filter((container) => container.tracking)
    } else if (filters.value.hasLocation === 'no') {
      result = result.filter((container) => !container.tracking)
    }
  }

  // Apply sorting
  result.sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle timestamp comparison
    if (sortConfig.value.field === 'time') {
      aValue = aValue ? new Date(aValue).getTime() : 0
      bValue = bValue ? new Date(bValue).getTime() : 0
    }

    // Handle string comparison
    if (typeof aValue === 'string') {
      aValue = aValue.toLowerCase()
      bValue = bValue.toLowerCase()
    }

    if (sortConfig.value.direction === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })

  return result
})

const handleSort = (field) => {
  if (sortConfig.value.field === field) {
    // Toggle direction if clicking the same field
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    // Set new field and default to ascending
    sortConfig.value.field = field
    sortConfig.value.direction = 'asc'
  }
}

const resetFilters = () => {
  filters.value = {
    containerRef: '',
    location: '',
    user: '',
    hasLocation: '',
  }
}

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

.filters-section {
  background: #f8f9fa;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.filters-header h3 {
  margin: 0;
  font-size: 16px;
  color: #495057;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
  font-size: 12px;
  padding: 6px 12px;
}

.btn-secondary:hover {
  background-color: #5a6268;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.filter-item label {
  font-size: 12px;
  font-weight: 600;
  color: #495057;
}

.filter-input,
.filter-select {
  padding: 8px 12px;
  border: 1px solid #ced4da;
  border-radius: 4px;
  font-size: 14px;
  background-color: white;
  transition: border-color 0.2s ease;
}

.filter-input:focus,
.filter-select:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
}

.sortable:hover {
  background-color: #e9ecef;
}

.sortable i {
  margin-left: 5px;
  font-size: 12px;
  color: #6c757d;
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
