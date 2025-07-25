<template>
  <div class="assigned-cars-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-car"></i>
        {{ t('loading.assigned_cars') }}
      </h3>
      <div class="header-actions" v-if="selectedLoadedContainerId">
        <button
          @click="refreshData"
          class="refresh-btn"
          :disabled="loading"
          :class="{ processing: loading }"
        >
          <i class="fas fa-sync-alt"></i>
          <span>{{ t('loading.refresh') }}</span>
          <i v-if="loading" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <div v-if="!selectedLoadedContainerId" class="empty-state">
      <i class="fas fa-mouse-pointer fa-2x"></i>
      <p>{{ t('loading.click_container_to_view_assigned_cars') }}</p>
    </div>

    <div v-else class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin"></i>
        <span>{{ t('loading.loading_assigned_cars') }}</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="assignedCars.length === 0" class="empty-state">
        <i class="fas fa-car-side"></i>
        <p>{{ t('loading.no_cars_assigned_to_containers') }}</p>
      </div>

      <div v-else class="table-content">
        <table class="assigned-cars-table">
          <thead>
            <tr>
              <th @click="sortByColumn('id')" class="sortable-header">
                {{ t('loading.id') }}
                <i
                  v-if="sortBy === 'id'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('car_name')" class="sortable-header">
                {{ t('loading.car_name') }}
                <i
                  v-if="sortBy === 'car_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('color')" class="sortable-header">
                {{ t('loading.color') }}
                <i
                  v-if="sortBy === 'color'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('vin')" class="sortable-header">
                {{ t('loading.vin') }}
                <i
                  v-if="sortBy === 'vin'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('container_name')" class="sortable-header">
                {{ t('loading.container') }}
                <i
                  v-if="sortBy === 'container_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('container_ref')" class="sortable-header">
                {{ t('loading.container_ref') }}
                <i
                  v-if="sortBy === 'container_ref'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('client_name')" class="sortable-header">
                {{ t('loading.client') }}
                <i
                  v-if="sortBy === 'client_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('sell_bill_id')" class="sortable-header">
                {{ t('loading.sell_bill_id') }}
                <i
                  v-if="sortBy === 'sell_bill_id'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('sell_bill_date')" class="sortable-header">
                {{ t('loading.sell_bill_date') }}
                <i
                  v-if="sortBy === 'sell_bill_date'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('sell_bill_ref')" class="sortable-header">
                {{ t('loading.sell_bill_ref') }}
                <i
                  v-if="sortBy === 'sell_bill_ref'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th>{{ t('loading.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="car in sortedAssignedCars"
              :key="car.id"
              class="table-row"
              @click="handleCarClick(car)"
            >
              <td class="id-cell">#{{ car.id }}</td>
              <td class="car-name-cell">{{ car.car_name || '-' }}</td>
              <td class="color-cell">{{ car.color_name || '-' }}</td>
              <td class="vin-cell">{{ car.vin || '-' }}</td>
              <td class="container-cell">{{ car.container_name || '-' }}</td>
              <td class="ref-cell">{{ car.container_ref || '-' }}</td>
              <td class="client-cell">
                <div v-if="car.id_client && car.id_copy_path" class="client-info">
                  <img
                    :src="getFileUrl(car.id_copy_path)"
                    :alt="
                      t('loading.id_for_client', { client: car.client_name || t('loading.client') })
                    "
                    class="client-id-image"
                    @click.stop="openClientId(car.id_copy_path)"
                  />
                  <span class="client-name">{{ car.client_name }}</span>
                </div>
                <span v-else class="no-client">-</span>
              </td>
              <td class="sell-bill-id-cell">{{ car.sell_bill_id || '-' }}</td>
              <td class="sell-bill-date-cell">
                <div class="date-badges">
                  <span class="badge sell-bill-date-badge">
                    <i class="fas fa-calendar"></i>
                    {{ t('loading.sell') }}: {{ formatDate(car.sell_bill_date) }}
                  </span>
                  <span v-if="isAdmin" class="badge assigned-date-badge">
                    <i class="fas fa-clock"></i>
                    {{ t('loading.assigned') }}:
                    {{ car.date_assigned ? formatDate(car.date_assigned) : t('loading.not_set') }}
                  </span>
                </div>
              </td>
              <td class="sell-bill-ref-cell">{{ car.sell_bill_ref || '-' }}</td>
              <td class="actions-cell">
                <div class="action-buttons">
                  <button
                    @click.stop="unassignCar(car)"
                    class="action-btn unassign-btn"
                    :class="{ disabled: car.date_on_board }"
                    :title="
                      car.date_on_board
                        ? t('loading.container_on_board_since', { date: car.date_on_board })
                        : t('loading.unassign_from_container')
                    "
                    :disabled="isUnassigning || car.date_on_board"
                  >
                    <i class="fas fa-unlink"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="table-footer">
      <span class="record-count">
        {{ t('loading.showing_assigned_cars', { count: sortedAssignedCars.length }) }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, defineExpose, computed } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useApi } from '@/composables/useApi'

const { t } = useEnhancedI18n()

const props = defineProps({
  selectedLoadedContainerId: {
    type: Number,
    default: null,
  },
})

const emit = defineEmits(['car-unassigned'])

const { callApi, getFileUrl } = useApi()

const assignedCars = ref([])
const loading = ref(false)
const error = ref(null)
const selectedCarId = ref(null)
const isUnassigning = ref(false)
const sortBy = ref('id')
const sortOrder = ref('desc')
const user = ref(null)

// Check if user is admin
const isAdmin = computed(() => {
  const adminStatus = user.value?.role_id === 1
  console.log('LoadingAssignedCars - Admin check:', {
    user: user.value,
    role_id: user.value?.role_id,
    isAdmin: adminStatus,
  })
  return adminStatus
})

// Computed property for sorted cars
const sortedAssignedCars = computed(() => {
  if (!assignedCars.value) return []

  return [...assignedCars.value].sort((a, b) => {
    let aVal = a[sortBy.value]
    let bVal = b[sortBy.value]

    // Handle numeric values
    if (['id', 'sell_bill_id'].includes(sortBy.value)) {
      aVal = parseFloat(aVal) || 0
      bVal = parseFloat(bVal) || 0
    }

    // Handle null values
    if (aVal == null) return 1
    if (bVal == null) return -1

    // Compare values
    if (sortOrder.value === 'asc') {
      return aVal > bVal ? 1 : aVal < bVal ? -1 : 0
    } else {
      return aVal < bVal ? 1 : aVal > bVal ? -1 : 0
    }
  })
})

const fetchAssignedCars = async () => {
  if (!props.selectedLoadedContainerId) {
    assignedCars.value = []
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cn.car_name,
          c.color as color_name,
          cs.vin,
          cont.name as container_name,
          lc.ref_container as container_ref,
          cs.price_cell,
          cs.id_sell,
          cs.date_sell,
          cs.id_loaded_container,
          cs.id_client,
          cl.name as client_name,
          cl.id_copy_path,
          lc.date_on_board,
          sb.id as sell_bill_id,
          sb.date_sell as sell_bill_date,
          sb.bill_ref as sell_bill_ref,
          cs.date_assigned
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors c ON cs.id_color = c.id
        LEFT JOIN loaded_containers lc ON cs.id_loaded_container = lc.id
        LEFT JOIN containers cont ON lc.id_container = cont.id
        LEFT JOIN clients cl ON cs.id_client = cl.id
        LEFT JOIN sell_bill sb ON cs.id_sell = sb.id
        WHERE cs.id_loaded_container = ?
        ORDER BY cs.id DESC
      `,
      params: [props.selectedLoadedContainerId],
    })

    if (result.success) {
      assignedCars.value = result.data || []
    } else {
      error.value = result.error || t('loading.failed_to_load_assigned_cars')
    }
  } catch (err) {
    error.value = t('loading.error_loading_assigned_cars', { message: err.message })
  } finally {
    loading.value = false
  }
}

const handleCarClick = (car) => {
  console.log('Car clicked:', car.id)
  selectedCarId.value = car.id
}

const openClientId = (path) => {
  window.open(getFileUrl(path), '_blank')
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

const unassignCar = async (car) => {
  if (isUnassigning.value) return

  // Check if container has date_on_board (container is on board)
  if (car.date_on_board) {
    alert(
      t('loading.cannot_unassign_car_container_on_board', {
        carId: car.id,
        date: car.date_on_board,
      }),
    )
    return
  }

  if (!confirm(t('loading.confirm_unassign_car', { carId: car.id }))) {
    return
  }

  isUnassigning.value = true

  try {
    const result = await callApi({
      query: 'UPDATE cars_stock SET id_loaded_container = NULL, container_ref = NULL WHERE id = ?',
      params: [car.id],
    })

    if (result.success) {
      // Refresh the data to update both tables
      await fetchAssignedCars()
      // Emit an event to refresh unassigned cars table
      emit('car-unassigned')
    } else {
      error.value = result.error || t('loading.failed_to_unassign_car')
    }
  } catch (err) {
    error.value = t('loading.error_unassigning_car', { message: err.message })
  } finally {
    isUnassigning.value = false
  }
}

const getStatusClass = (idSell) => {
  return idSell ? 'status-sold' : 'status-available'
}

const refreshData = () => {
  fetchAssignedCars()
}

const sortByColumn = (column) => {
  if (sortBy.value === column) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = column
    sortOrder.value = 'asc'
  }
}

// Watch for changes in selectedLoadedContainerId
watch(
  () => props.selectedLoadedContainerId,
  (newId) => {
    if (newId) {
      fetchAssignedCars()
    } else {
      assignedCars.value = []
    }
  },
  { immediate: true },
)

// Load user data on component mount
const loadUserData = () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    console.log('LoadingAssignedCars - User loaded:', user.value)
    console.log('LoadingAssignedCars - Is admin:', isAdmin.value)
  } else {
    console.log('LoadingAssignedCars - No user found in localStorage')
  }
}

// Expose methods to parent component
defineExpose({
  refreshData,
  fetchAssignedCars,
  assignedCars,
})

// Load user data when component is mounted
loadUserData()
</script>

<style scoped>
.assigned-cars-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  margin-top: 16px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
}

.table-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.1rem;
}

.table-header h3 i {
  color: #10b981;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  background-color: #f8fafc;
  color: #374151;
  border: 1px solid #d1d5db;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.refresh-btn.processing {
  opacity: 0.7;
}

.loading-indicator {
  margin-left: 4px;
}

.table-wrapper {
  position: relative;
  max-height: 400px;
  overflow-y: auto;
}

.loading-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 20px;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 4px;
  margin: 16px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.empty-state i {
  color: #d1d5db;
  font-size: 2rem;
}

.table-content {
  overflow-x: auto;
}

.assigned-cars-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.assigned-cars-table th {
  background-color: #f8fafc;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  position: sticky;
  top: 0;
  z-index: 10;
}

.assigned-cars-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
}

.table-row:hover {
  background-color: #f9fafb;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.table-row:active {
  background-color: #e5e7eb;
}

.selected-row {
  background-color: #dbeafe !important;
  border-left: 4px solid #3b82f6;
}

.selected-row:hover {
  background-color: #bfdbfe !important;
}

.id-cell {
  font-weight: 600;
  color: #6b7280;
  width: 60px;
}

.car-name-cell {
  font-weight: 500;
}

.color-cell {
  color: #6b7280;
}

.vin-cell {
  font-family: monospace;
  background-color: #f3f4f6;
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 0.8rem;
}

.container-cell {
  font-weight: 500;
}

.ref-cell {
  font-family: monospace;
  background-color: #f3f4f6;
  border-radius: 4px;
  padding: 2px 6px;
}

.selected-row .ref-cell {
  background-color: #e5e7eb;
}

.price-cell {
  font-weight: 500;
  color: #059669;
}

.status-cell {
  text-align: center;
}

.status-available {
  display: inline-block;
  padding: 2px 8px;
  background-color: #dcfce7;
  color: #166534;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-sold {
  background-color: #fee2e2;
  color: #991b1b;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.actions-cell {
  text-align: center;
}

.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.9rem;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.unassign-btn {
  background-color: #fef3c7;
  color: #d97706;
  border: 1px solid #fbbf24;
}

.unassign-btn:hover:not(:disabled) {
  background-color: #fde68a;
  color: #b45309;
}

.unassign-btn.disabled {
  background-color: #f3f4f6;
  color: #9ca3af;
  border-color: #d1d5db;
  cursor: not-allowed;
}

.unassign-btn.disabled:hover {
  background-color: #f3f4f6;
  color: #9ca3af;
}

.table-footer {
  padding: 12px 20px;
  border-top: 1px solid #e5e7eb;
  background-color: #f8fafc;
}

.record-count {
  color: #6b7280;
  font-size: 0.9rem;
}

.client-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.client-id-image {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid #e5e7eb;
  cursor: pointer;
  transition: all 0.2s ease;
}

.client-id-image:hover {
  transform: scale(1.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.client-name {
  font-weight: 500;
  color: #374151;
  font-size: 0.9rem;
}

.no-client {
  color: #6b7280;
  font-style: italic;
}

.sell-bill-id-cell {
  text-align: center;
  font-size: 0.9rem;
  color: #6b7280;
}

.sell-bill-date-cell {
  text-align: center;
  font-size: 0.9rem;
  color: #6b7280;
}

.sell-bill-ref-cell {
  text-align: center;
  font-size: 0.9rem;
  color: #6b7280;
  font-family: monospace;
}

.date-assigned-cell {
  text-align: center;
  font-size: 0.9rem;
  color: #6b7280;
  font-family: monospace;
}

.sortable-header {
  cursor: pointer;
}

.sort-inactive {
  opacity: 0.5;
}

.date-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.sell-bill-date-badge,
.assigned-date-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  font-family: inherit;
  text-align: center;
  width: 100%;
  box-sizing: border-box;
  justify-content: center;
}

.sell-bill-date-badge {
  background-color: #eff6ff;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}

.assigned-date-badge {
  background-color: #f3e8ff;
  color: #6b21a8;
  border: 1px solid #d9b8ff;
}
</style>
