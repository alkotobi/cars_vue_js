<script setup>
import { ref, defineProps, defineEmits, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  selectedCars: {
    type: Array,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
  isAdmin: {
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
const selectedWarehouse = ref('')

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

const handleWarehouseChange = async () => {
  if (isProcessing.value || loading.value) return
  if (!selectedWarehouse.value) {
    error.value = 'Please select a warehouse'
    return
  }

  if (!confirm(`Are you sure you want to assign ${props.selectedCars.length} cars to the selected warehouse?`)) {
    return
  }

  loading.value = true
  isProcessing.value = true
  error.value = null
  success.value = null

  try {
    const carIds = props.selectedCars.map(car => car.id)
    const result = await callApi({
      query: 'UPDATE cars_stock SET id_warehouse = ? WHERE id IN (' + carIds.map(() => '?').join(',') + ')',
      params: [selectedWarehouse.value, ...carIds],
    })

    if (result.success) {
      const warehouse = warehouses.value.find((w) => w.id === parseInt(selectedWarehouse.value))
      success.value = `Successfully assigned ${props.selectedCars.length} cars to ${warehouse?.name || 'warehouse'}`
      
      // Update the cars data
      const updatedCars = props.selectedCars.map(car => ({
        ...car,
        id_warehouse: parseInt(selectedWarehouse.value),
        warehouse_name: warehouse?.name,
      }))
      
      emit('save', updatedCars)
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

const handleRevertWarehouse = async () => {
  if (!props.isAdmin) return

  if (!confirm(`Are you sure you want to revert warehouse assignment to null for ${props.selectedCars.length} cars? This action cannot be undone.`)) {
    return
  }

  loading.value = true
  isProcessing.value = true
  error.value = null
  success.value = null

  try {
    const carIds = props.selectedCars.map(car => car.id)
    const result = await callApi({
      query: 'UPDATE cars_stock SET id_warehouse = NULL WHERE id IN (' + carIds.map(() => '?').join(',') + ')',
      params: carIds,
    })

    if (result.success) {
      success.value = `Successfully reverted warehouse assignment for ${props.selectedCars.length} cars`
      
      // Update the cars data
      const updatedCars = props.selectedCars.map(car => ({
        ...car,
        id_warehouse: null,
        warehouse_name: null,
      }))
      
      emit('save', updatedCars)
    } else {
      throw new Error(result.error || 'Failed to revert warehouse')
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
  selectedWarehouse.value = ''
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
          Bulk Warehouse Assignment
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="info-section">
          <h4>
            <i class="fas fa-info-circle"></i>
            Selected Cars
          </h4>
          <p>You have selected {{ selectedCars.length }} cars to assign to a warehouse.</p>
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
              <span>{{ isProcessing ? 'Updating...' : 'Assign to Warehouse' }}</span>
              <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>

          <!-- Revert Button for Admin -->
          <div v-if="isAdmin" class="form-group">
            <button
              class="revert-btn"
              @click="handleRevertWarehouse"
              :disabled="isProcessing"
            >
              <i class="fas fa-undo"></i>
              Revert Warehouse to Null
            </button>
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
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
}

.info-section {
  margin-bottom: 24px;
  padding: 16px;
  background-color: #f9fafb;
  border-radius: 6px;
}

.info-section h4 {
  margin: 0 0 12px 0;
  color: #374151;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-section h4 i {
  color: #3b82f6;
}

.info-section p {
  margin: 0;
  color: #6b7280;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
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
.revert-btn {
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
  width: 100%;
}

.update-warehouse-btn {
  background-color: #3b82f6;
  color: white;
}

.revert-btn {
  background-color: #dc2626;
  color: white;
}

.update-warehouse-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.revert-btn:hover:not(:disabled) {
  background-color: #b91c1c;
  transform: translateY(-1px);
}

.update-warehouse-btn:disabled,
.revert-btn:disabled {
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
</style> 