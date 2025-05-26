<script setup>
import { ref, defineProps, defineEmits, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  car: {
    type: Object,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])
const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const success = ref(null)
const isProcessing = ref(false)
const isFetchingWarehouses = ref(false)

// Data for warehouse dropdown
const warehouses = ref([])
const selectedWarehouse = ref(props.car.id_warehouse || '')

// Add user data
const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

// Add permission check
const can_receive_car = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true // Admin
  return user.value.permissions?.some((p) => p.permission_name === 'can_receive_car')
})

// Fetch warehouses data
const fetchWarehouses = async () => {
  isFetchingWarehouses.value = true
  try {
    console.log('Fetching warehouses...')
    const result = await callApi({
      query: 'SELECT id, warhouse_name as name FROM warehouses ORDER BY warhouse_name',
      params: [],
    })
    console.log('API Response:', result)

    if (result.success) {
      warehouses.value = result.data
      console.log('Warehouses loaded:', warehouses.value)
    } else {
      console.error('API Error:', result.error)
      throw new Error(result.error || 'Failed to fetch warehouses')
    }
  } catch (err) {
    console.error('Full error details:', {
      message: err.message,
      stack: err.stack,
      response: err.response,
    })
    error.value = `Failed to load warehouses data: ${err.message}`
  } finally {
    isFetchingWarehouses.value = false
  }
}

const handleGetCar = async () => {
  if (isProcessing.value || loading.value) return
  if (!can_receive_car.value) {
    error.value = 'You do not have permission to receive cars'
    return
  }

  if (props.car.in_wharhouse_date) {
    return // Button should be disabled anyway
  }

  loading.value = true
  isProcessing.value = true
  error.value = null
  success.value = null

  try {
    const currentDate = new Date().toISOString().split('T')[0]
    const result = await callApi({
      query: 'UPDATE cars_stock SET in_wharhouse_date = ? WHERE id = ?',
      params: [currentDate, props.car.id],
    })

    if (result.success) {
      success.value = 'Car received in warehouse'
      const updatedCar = {
        ...props.car,
        in_wharhouse_date: currentDate,
      }
      // Update the local car data
      Object.assign(props.car, updatedCar)
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || 'Failed to update warehouse date')
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

const handleWarehouseChange = async () => {
  if (isProcessing.value || loading.value) return
  if (!selectedWarehouse.value) {
    error.value = 'Please select a warehouse'
    return
  }

  loading.value = true
  isProcessing.value = true
  error.value = null
  success.value = null

  try {
    const result = await callApi({
      query: 'UPDATE cars_stock SET id_warehouse = ? WHERE id = ?',
      params: [selectedWarehouse.value, props.car.id],
    })

    if (result.success) {
      const warehouse = warehouses.value.find((w) => w.id === parseInt(selectedWarehouse.value))
      const updatedCar = {
        ...props.car,
        id_warehouse: parseInt(selectedWarehouse.value),
        warehouse_name: warehouse?.name,
      }
      success.value = 'Warehouse updated successfully'
      emit('save', updatedCar)
    } else {
      throw new Error(result.error || 'Failed to update warehouse')
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = null
  selectedWarehouse.value = props.car.id_warehouse || ''
  emit('close')
}

// Fetch warehouses data when component is mounted
onMounted(fetchWarehouses)
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content" :class="{ 'is-processing': isProcessing }">
      <div class="modal-header">
        <h3>
          <i class="fas fa-warehouse"></i>
          Warehouse Management
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <!-- Warehouse Status Section -->
        <div class="status-section">
          <h4>
            <i class="fas fa-info-circle"></i>
            Warehouse Status
          </h4>
          <div class="status-info">
            <span class="status-label">
              <i class="fas fa-circle-check"></i>
              Current Status:
            </span>
            <span
              :class="[
                'status-value',
                props.car.in_wharhouse_date ? 'in-warehouse' : 'not-in-warehouse',
              ]"
            >
              <i :class="props.car.in_wharhouse_date ? 'fas fa-check-circle' : 'fas fa-clock'"></i>
              {{ props.car.in_wharhouse_date ? 'In Warehouse' : 'Not In Warehouse' }}
            </span>
          </div>
          <div v-if="props.car.in_wharhouse_date" class="date-info">
            <span class="date-label">
              <i class="fas fa-calendar-alt"></i>
              Received Date:
            </span>
            <span class="date-value">
              {{ new Date(props.car.in_wharhouse_date).toLocaleDateString() }}
            </span>
          </div>
        </div>

        <!-- Warehouse Selection Section -->
        <div v-if="isFetchingWarehouses" class="loading-state">
          <i class="fas fa-spinner fa-spin"></i>
          Loading warehouses...
        </div>

        <template v-else>
          <div class="form-group">
            <label for="warehouse">
              <i class="fas fa-building"></i>
              Warehouse:
            </label>
            <div class="select-wrapper">
              <select
                id="warehouse"
                v-model="selectedWarehouse"
                class="select-field"
                :disabled="isProcessing"
              >
                <option value="">Select Warehouse</option>
                <option v-for="warehouse in warehouses" :key="warehouse.id" :value="warehouse.id">
                  {{ warehouse.name }}
                </option>
              </select>
              <i class="fas fa-chevron-down select-arrow"></i>
            </div>
            <button
              class="update-warehouse-btn"
              @click="handleWarehouseChange"
              :disabled="isProcessing || !selectedWarehouse"
              :class="{ 'is-processing': isProcessing }"
            >
              <i class="fas fa-sync-alt"></i>
              <span>{{ isProcessing ? 'Updating...' : 'Update Warehouse' }}</span>
              <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>

          <!-- Get Car Button -->
          <div class="get-car-section">
            <button
              class="get-car-btn"
              @click="handleGetCar"
              :disabled="isProcessing || !!props.car.in_wharhouse_date || !can_receive_car"
              :class="{
                disabled: !!props.car.in_wharhouse_date || !can_receive_car,
                'is-processing': isProcessing,
              }"
            >
              <i class="fas fa-truck-loading"></i>
              <span>{{ isProcessing ? 'Processing...' : 'We Get Car' }}</span>
              <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
            <div v-if="!can_receive_car && !props.car.in_wharhouse_date" class="permission-notice">
              <i class="fas fa-lock"></i>
              You need permission to receive cars
            </div>
          </div>

          <div v-if="error" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ error }}
          </div>
          <div v-if="success" class="success-message">
            <i class="fas fa-check-circle"></i>
            {{ success }}
          </div>
        </template>
      </div>

      <div class="modal-footer">
        <button class="close-btn-secondary" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
}

.status-section {
  margin-bottom: 24px;
  padding: 16px;
  background-color: #f9fafb;
  border-radius: 6px;
}

.status-section h4 {
  margin: 0 0 12px 0;
  color: #374151;
  font-size: 1.1rem;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.status-label,
.date-label {
  font-weight: 500;
  color: #6b7280;
}

.status-value {
  font-weight: 600;
}

.in-warehouse {
  color: #059669;
}

.not-in-warehouse {
  color: #dc2626;
}

.date-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.date-value {
  color: #374151;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.select-wrapper {
  position: relative;
  margin-bottom: 12px;
}

.select-arrow {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
  pointer-events: none;
}

.select-field {
  width: 100%;
  padding: 8px 12px;
  padding-right: 32px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  appearance: none;
  background-color: white;
  transition: all 0.2s ease;
}

.select-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.select-field:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.update-warehouse-btn,
.get-car-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.update-warehouse-btn {
  background-color: #3b82f6;
  color: white;
  width: 100%;
}

.get-car-btn {
  background-color: #10b981;
  color: white;
  width: 100%;
  margin-top: 24px;
}

.update-warehouse-btn:hover:not(:disabled),
.get-car-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.update-warehouse-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.get-car-btn:hover:not(:disabled) {
  background-color: #059669;
}

.update-warehouse-btn:disabled,
.get-car-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.is-processing {
  position: relative;
}

.is-processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

.permission-notice {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  color: #ef4444;
  font-size: 14px;
}

.error-message,
.success-message {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 16px 0;
  padding: 12px;
  border-radius: 4px;
  font-size: 14px;
}

.error-message {
  background-color: #fee2e2;
  color: #ef4444;
}

.success-message {
  background-color: #d1fae5;
  color: #10b981;
}

.close-btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn-secondary:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.loading-indicator {
  margin-left: 4px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fa-spin {
  animation: spin 1s linear infinite;
}

.modal-content.is-processing {
  opacity: 0.7;
  pointer-events: none;
}

.modal-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-header h3 i {
  color: #3b82f6;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px;
  color: #6b7280;
  font-size: 16px;
}

.loading-state i {
  color: #3b82f6;
  font-size: 24px;
}

.status-section h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 16px;
  color: #1f2937;
}

.status-section h4 i {
  color: #3b82f6;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.status-label,
.date-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  font-weight: 500;
}

.status-value {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.status-value.in-warehouse {
  background-color: #d1fae5;
  color: #10b981;
}

.status-value.not-in-warehouse {
  background-color: #fee2e2;
  color: #ef4444;
}

.date-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}

.date-value {
  color: #1f2937;
  font-weight: 500;
}

.form-group {
  margin: 24px 0;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}
</style>
