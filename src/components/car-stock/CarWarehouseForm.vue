<script setup>
import { ref, defineProps, defineEmits, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  car: {
    type: Object,
    required: true
  },
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'save'])
const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const success = ref(null)

// Data for warehouse dropdown
const warehouses = ref([])
const selectedWarehouse = ref(props.car.id_warehouse || '')

// Add user data
const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

// Add permission check
const can_receive_car = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true // Admin
  return user.value.permissions?.some(p => p.permission_name === 'can_receive_car')
})

// Fetch warehouses data
const fetchWarehouses = async () => {
  try {
    console.log('Fetching warehouses...')
    const result = await callApi({
      query: 'SELECT id, warhouse_name as name FROM warehouses ORDER BY warhouse_name',
      params: []
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
      response: err.response
    })
    error.value = `Failed to load warehouses data: ${err.message}`
  }
}

const handleGetCar = async () => {
  if (!can_receive_car.value) {
    error.value = 'You do not have permission to receive cars'
    return
  }

  if (props.car.in_wharhouse_date) {
    return // Button should be disabled anyway
  }

  loading.value = true
  error.value = null
  success.value = null

  try {
    const currentDate = new Date().toISOString().split('T')[0]
    const result = await callApi({
      query: 'UPDATE cars_stock SET in_wharhouse_date = ? WHERE id = ?',
      params: [currentDate, props.car.id]
    })

    if (result.success) {
      success.value = 'Car received in warehouse'
      const updatedCar = { 
        ...props.car,
        in_wharhouse_date: currentDate
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
  }
}

const handleWarehouseChange = async () => {
  if (!selectedWarehouse.value) {
    error.value = 'Please select a warehouse'
    return
  }

  loading.value = true
  error.value = null
  success.value = null

  try {
    const result = await callApi({
      query: 'UPDATE cars_stock SET id_warehouse = ? WHERE id = ?',
      params: [selectedWarehouse.value, props.car.id]
    })

    if (result.success) {
      const warehouse = warehouses.value.find(w => w.id === parseInt(selectedWarehouse.value))
      const updatedCar = { 
        ...props.car,
        id_warehouse: parseInt(selectedWarehouse.value),
        warehouse_name: warehouse?.name
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
    <div class="modal-content">
      <div class="modal-header">
        <h3>Warehouse Management</h3>
        <button class="close-btn" @click="closeModal">&times;</button>
      </div>

      <div class="modal-body">
        <!-- Warehouse Status Section -->
        <div class="status-section">
          <h4>Warehouse Status</h4>
          <div class="status-info">
            <span class="status-label">Current Status:</span>
            <span :class="['status-value', props.car.in_wharhouse_date ? 'in-warehouse' : 'not-in-warehouse']">
              {{ props.car.in_wharhouse_date ? 'In Warehouse' : 'Not In Warehouse' }}
            </span>
          </div>
          <div v-if="props.car.in_wharhouse_date" class="date-info">
            <span class="date-label">Received Date:</span>
            <span class="date-value">{{ new Date(props.car.in_wharhouse_date).toLocaleDateString() }}</span>
          </div>
        </div>

        <!-- Warehouse Selection Section -->
        <div class="form-group">
          <label for="warehouse">Warehouse:</label>
          <select 
            id="warehouse" 
            v-model="selectedWarehouse"
            class="select-field"
            :disabled="loading"
          >
            <option value="">Select Warehouse</option>
            <option 
              v-for="warehouse in warehouses" 
              :key="warehouse.id" 
              :value="warehouse.id"
            >
              {{ warehouse.name }}
            </option>
          </select>
          <button 
            class="update-warehouse-btn" 
            @click="handleWarehouseChange"
            :disabled="loading || !selectedWarehouse"
          >
            Update Warehouse
          </button>
        </div>

        <!-- Get Car Button -->
        <div class="get-car-section">
          <button 
            class="get-car-btn" 
            @click="handleGetCar"
            :disabled="loading || !!props.car.in_wharhouse_date || !can_receive_car"
            :class="{ 
              'disabled': !!props.car.in_wharhouse_date || !can_receive_car 
            }"
          >
            {{ loading ? 'Processing...' : 'We Get Car' }}
          </button>
          <div v-if="!can_receive_car && !props.car.in_wharhouse_date" class="permission-notice">
            You need permission to receive cars
          </div>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
        <div v-if="success" class="success-message">
          {{ success }}
        </div>
      </div>

      <div class="modal-footer">
        <button 
          class="close-btn-secondary" 
          @click="closeModal"
          :disabled="loading"
        >
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

.status-label, .date-label {
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

.select-field {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  margin-bottom: 12px;
}

.update-warehouse-btn {
  width: 100%;
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.update-warehouse-btn:hover {
  background-color: #2563eb;
}

.get-car-section {
  margin-top: 24px;
}

.get-car-btn {
  width: 100%;
  padding: 12px 24px;
  background-color: #059669;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  transition: background-color 0.2s;
}

.get-car-btn:hover:not(:disabled) {
  background-color: #047857;
}

.get-car-btn.disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.error-message {
  color: #ef4444;
  margin: 16px 0;
  font-size: 14px;
}

.success-message {
  color: #10b981;
  margin: 16px 0;
  font-size: 14px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.close-btn-secondary {
  padding: 8px 16px;
  background-color: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.close-btn-secondary:hover {
  background-color: #e5e7eb;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.permission-notice {
  color: #dc2626;
  font-size: 0.875rem;
  margin-top: 0.5rem;
  text-align: center;
}
</style> 