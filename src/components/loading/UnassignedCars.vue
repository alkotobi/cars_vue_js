<template>
  <div class="unassigned-cars-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-car"></i>
        Unassigned Cars
      </h3>
      <div class="header-actions" v-if="unassignedCars.length > 0">
        <button
          @click="refreshData"
          class="refresh-btn"
          :disabled="loading"
          :class="{ processing: loading }"
        >
          <i class="fas fa-sync-alt"></i>
          <span>Refresh</span>
          <i v-if="loading" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <div class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading...</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="!props.selectedLoadedContainerId" class="empty-state">
        <i class="fas fa-mouse-pointer fa-2x"></i>
        <p>Please select a loading record above to view unassigned cars</p>
      </div>

      <div v-else-if="unassignedCars.length === 0" class="empty-state">
        <i class="fas fa-check-circle fa-2x"></i>
        <p>All cars are assigned to containers</p>
      </div>

      <div v-else class="table-content">
        <table class="unassigned-cars-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Car Name</th>
              <th>Color</th>
              <th>VIN</th>
              <th>Client</th>
              <th>Price</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="car in unassignedCars"
              :key="car.id"
              class="table-row"
              @click="handleCarClick(car)"
            >
              <td class="id-cell">#{{ car.id }}</td>
              <td class="car-name-cell">{{ car.car_name || '-' }}</td>
              <td class="color-cell">{{ car.color || '-' }}</td>
              <td class="vin-cell">{{ car.vin || '-' }}</td>
              <td class="client-cell">
                <div v-if="car.id_client && car.id_copy_path" class="client-info">
                  <img
                    :src="getFileUrl(car.id_copy_path)"
                    :alt="'ID for ' + (car.client_name || 'Client')"
                    class="client-id-image"
                    @click.stop="openClientId(car.id_copy_path)"
                  />
                  <span class="client-name">{{ car.client_name }}</span>
                </div>
                <span v-else class="no-client">-</span>
              </td>
              <td class="price-cell">{{ car.price_cell ? `$${car.price_cell}` : '-' }}</td>
              <td class="status-cell">
                <span :class="getStatusClass(car.id_sell)">
                  {{ car.id_sell ? 'Sold' : 'Available' }}
                </span>
              </td>
              <td class="actions-cell">
                <div class="action-buttons">
                  <button
                    @click.stop="assignCar(car)"
                    class="action-btn assign-btn"
                    :title="getAssignButtonTitle()"
                    :disabled="isAssigning || props.selectedContainerOnBoard"
                  >
                    <i class="fas fa-link"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="table-footer">
      <span class="record-count"> Showing {{ unassignedCars.length }} unassigned cars </span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '@/composables/useApi'

const props = defineProps({
  selectedLoadedContainerId: {
    type: Number,
    default: null,
  },
  selectedContainerOnBoard: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['car-assigned'])

const { callApi, getFileUrl } = useApi()

const unassignedCars = ref([])
const loading = ref(false)
const error = ref(null)
const selectedCarId = ref(null)
const isAssigning = ref(false)

const fetchUnassignedCars = async () => {
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cn.car_name,
          c.color,
          cs.vin,
          cs.price_cell,
          cs.id_sell,
          cs.id_client,
          cl.name as client_name,
          cl.id_copy_path
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors c ON bd.id_color = c.id
        LEFT JOIN clients cl ON cs.id_client = cl.id
        WHERE cs.id_loaded_container IS NULL 
        AND cs.date_loding IS NULL
        ORDER BY cs.id DESC
      `,
      params: [],
    })

    if (result.success) {
      unassignedCars.value = result.data || []
    } else {
      error.value = result.error || 'Failed to load unassigned cars'
    }
  } catch (err) {
    error.value = 'Error loading unassigned cars: ' + err.message
  } finally {
    loading.value = false
  }
}

const handleCarClick = (car) => {
  console.log('Unassigned car clicked:', car.id)
  selectedCarId.value = car.id
}

const openClientId = (path) => {
  window.open(getFileUrl(path), '_blank')
}

const getStatusClass = (idSell) => {
  return idSell ? 'status-sold' : 'status-available'
}

const refreshData = () => {
  fetchUnassignedCars()
}

const assignCar = async (car) => {
  if (isAssigning.value) return

  if (!props.selectedLoadedContainerId) {
    alert('Please select a container first')
    return
  }

  // Check if the selected container is already on board
  try {
    const containerResult = await callApi({
      query: 'SELECT date_on_board FROM loaded_containers WHERE id = ?',
      params: [props.selectedLoadedContainerId],
    })

    if (containerResult.success && containerResult.data && containerResult.data.length > 0) {
      const container = containerResult.data[0]
      if (container.date_on_board) {
        alert(
          `Cannot assign car to this container. Container is already on board since ${container.date_on_board}`,
        )
        return
      }
    }
  } catch (err) {
    error.value = 'Error checking container status: ' + err.message
    return
  }

  if (!confirm(`Are you sure you want to assign car #${car.id} to this container?`)) {
    return
  }

  isAssigning.value = true

  try {
    const result = await callApi({
      query: 'UPDATE cars_stock SET id_loaded_container = ? WHERE id = ?',
      params: [props.selectedLoadedContainerId, car.id],
    })

    if (result.success) {
      // Refresh the data to update both tables
      await fetchUnassignedCars()
      // Emit an event to refresh assigned cars table
      emit('car-assigned')
    } else {
      error.value = result.error || 'Failed to assign car'
    }
  } catch (err) {
    error.value = 'Error assigning car: ' + err.message
  } finally {
    isAssigning.value = false
  }
}

const getAssignButtonTitle = () => {
  if (props.selectedContainerOnBoard) {
    return 'Container is already on board'
  }
  return 'Assign to Container'
}

onMounted(() => {
  fetchUnassignedCars()
})

// Expose methods to parent component
defineExpose({
  refreshData,
  fetchUnassignedCars,
})
</script>

<style scoped>
.unassigned-cars-container {
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
  color: #f59e0b;
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
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  z-index: 10;
}

.loading-overlay i {
  color: #3b82f6;
}

.loading-overlay span {
  color: #6b7280;
  font-weight: 500;
}

.error-message {
  padding: 20px;
  text-align: center;
  color: #dc2626;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  margin: 16px;
}

.error-message i {
  margin-right: 8px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: #6b7280;
}

.empty-state i {
  margin-bottom: 16px;
  color: #10b981;
}

.empty-state p {
  margin: 0;
  font-size: 1rem;
}

.table-content {
  overflow-x: auto;
}

.unassigned-cars-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.unassigned-cars-table th {
  background-color: #f8fafc;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
  position: sticky;
  top: 0;
  z-index: 5;
}

.unassigned-cars-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f3f4f6;
  vertical-align: middle;
}

.table-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.table-row:hover {
  background-color: #f9fafb;
}

.table-row:active {
  background-color: #f3f4f6;
}

.id-cell {
  font-weight: 600;
  color: #6b7280;
  font-family: monospace;
}

.car-name-cell {
  font-weight: 500;
  color: #1f2937;
}

.color-cell {
  color: #374151;
}

.vin-cell {
  font-family: monospace;
  color: #6b7280;
  font-size: 0.85rem;
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

.price-cell {
  font-weight: 600;
  color: #059669;
}

.status-cell {
  text-align: center;
}

.status-available {
  background-color: #d1fae5;
  color: #065f46;
  padding: 4px 12px;
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
  justify-content: center;
  gap: 8px;
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

.action-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.assign-btn {
  background-color: #dbeafe;
  color: #1d4ed8;
  border: 1px solid #93c5fd;
}

.assign-btn:hover:not(:disabled) {
  background-color: #bfdbfe;
  color: #1e40af;
}

.assign-btn:disabled {
  background-color: #e5e7eb;
  color: #9ca3af;
  border-color: #d1d5db;
  cursor: not-allowed;
}

.assign-btn:disabled:hover {
  background-color: #e5e7eb;
  color: #9ca3af;
}

.table-footer {
  padding: 12px 20px;
  border-top: 1px solid #e5e7eb;
  background-color: #f8fafc;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.record-count {
  color: #6b7280;
  font-size: 0.9rem;
}
</style>
