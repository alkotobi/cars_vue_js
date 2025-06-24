<template>
  <div class="assigned-cars-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-car"></i>
        Assigned Cars
      </h3>
      <div class="header-actions" v-if="selectedLoadedContainerId">
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

    <div v-if="!selectedLoadedContainerId" class="empty-state">
      <i class="fas fa-mouse-pointer"></i>
      <p>Click on a container line above to view assigned cars</p>
    </div>

    <div v-else class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin"></i>
        <span>Loading assigned cars...</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-triangle"></i>
        <span>{{ error }}</span>
      </div>

      <div v-else-if="assignedCars.length === 0" class="empty-state">
        <i class="fas fa-car-side"></i>
        <p>No cars assigned to containers in this loading record</p>
      </div>

      <div v-else class="table-content">
        <table class="assigned-cars-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Car Name</th>
              <th>Color</th>
              <th>VIN</th>
              <th>Container</th>
              <th>Container Ref</th>
              <th>Price</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="car in assignedCars"
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
              <td class="price-cell">{{ car.price_cell ? `$${car.price_cell}` : '-' }}</td>
              <td class="status-cell">
                <span :class="getStatusClass(car.id_sell)">
                  {{ car.id_sell ? 'Sold' : 'Available' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="table-footer">
      <span class="record-count"> Showing {{ assignedCars.length }} assigned cars </span>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useApi } from '@/composables/useApi'

const props = defineProps({
  selectedLoadedContainerId: {
    type: Number,
    default: null,
  },
})

const { callApi } = useApi()

const assignedCars = ref([])
const loading = ref(false)
const error = ref(null)
const selectedCarId = ref(null)

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
          c.color,
          cs.vin,
          cont.name as container_name,
          lc.ref_container as container_ref,
          cs.price_cell,
          cs.id_sell,
          cs.id_loaded_container
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors c ON bd.id_color = c.id
        LEFT JOIN loaded_containers lc ON cs.id_loaded_container = lc.id
        LEFT JOIN containers cont ON lc.id_container = cont.id
        WHERE cs.id_loaded_container = ?
        ORDER BY cs.id DESC
      `,
      params: [props.selectedLoadedContainerId],
    })

    if (result.success) {
      assignedCars.value = result.data || []
    } else {
      error.value = result.error || 'Failed to load assigned cars'
    }
  } catch (err) {
    error.value = 'Error loading assigned cars: ' + err.message
  } finally {
    loading.value = false
  }
}

const handleCarClick = (car) => {
  console.log('Car clicked:', car.id)
  selectedCarId.value = car.id
}

const getStatusClass = (idSell) => {
  return idSell ? 'status-sold' : 'status-available'
}

const refreshData = () => {
  fetchAssignedCars()
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
  display: inline-block;
  padding: 2px 8px;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
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
</style>
