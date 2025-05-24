<script setup>
import { ref, defineProps, defineEmits, onMounted } from 'vue'
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

// Data for dropdowns
const loadingPorts = ref([])
const dischargePorts = ref([])

// Selected values
const selectedLoadingPort = ref(props.car.id_port_loading || '')
const selectedDischargePort = ref(props.car.id_port_discharge || '')

// Fetch ports data
const fetchPorts = async () => {
  try {
    // Fetch loading ports
    const loadingPortsResult = await callApi({
      query: 'SELECT id, loading_port FROM loading_ports ORDER BY loading_port',
      params: []
    })
    if (loadingPortsResult.success) {
      loadingPorts.value = loadingPortsResult.data
    }

    // Fetch discharge ports
    const dischargePortsResult = await callApi({
      query: 'SELECT id, discharge_port FROM discharge_ports ORDER BY discharge_port',
      params: []
    })
    if (dischargePortsResult.success) {
      dischargePorts.value = dischargePortsResult.data
    }
  } catch (err) {
    error.value = 'Failed to load ports data'
    console.error('Error fetching ports:', err)
  }
}

const handleSubmit = async () => {
  if (!selectedLoadingPort.value && !selectedDischargePort.value) {
    error.value = 'Please select at least one port to update'
    return
  }

  loading.value = true
  error.value = null

  try {
    const updates = []
    const params = []
    
    if (selectedLoadingPort.value) {
      updates.push('id_port_loading = ?')
      params.push(selectedLoadingPort.value)
    }
    
    if (selectedDischargePort.value) {
      updates.push('id_port_discharge = ?')
      params.push(selectedDischargePort.value)
    }
    
    params.push(props.car.id)

    const result = await callApi({
      query: `UPDATE cars_stock SET ${updates.join(', ')} WHERE id = ?`,
      params
    })

    if (result.success) {
      // Find the port names for the response
      const updatedCar = { ...props.car }
      if (selectedLoadingPort.value) {
        const loadingPort = loadingPorts.value.find(p => p.id === parseInt(selectedLoadingPort.value))
        updatedCar.loading_port = loadingPort?.loading_port
        updatedCar.id_port_loading = parseInt(selectedLoadingPort.value)
      }
      if (selectedDischargePort.value) {
        const dischargePort = dischargePorts.value.find(p => p.id === parseInt(selectedDischargePort.value))
        updatedCar.discharge_port = dischargePort?.discharge_port
        updatedCar.id_port_discharge = parseInt(selectedDischargePort.value)
      }
      
      emit('save', updatedCar)
      emit('close')
    } else {
      throw new Error(result.error || 'Failed to update ports')
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  selectedLoadingPort.value = props.car.id_port_loading || ''
  selectedDischargePort.value = props.car.id_port_discharge || ''
  emit('close')
}

// Fetch ports data when component is mounted
onMounted(fetchPorts)
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Edit Ports</h3>
        <button class="close-btn" @click="closeModal">&times;</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="loading-port">Loading Port:</label>
          <select 
            id="loading-port" 
            v-model="selectedLoadingPort"
            class="select-field"
          >
            <option value="">Select Loading Port</option>
            <option 
              v-for="port in loadingPorts" 
              :key="port.id" 
              :value="port.id"
            >
              {{ port.loading_port }}
            </option>
          </select>
        </div>

        <div class="form-group">
          <label for="discharge-port">Discharge Port:</label>
          <select 
            id="discharge-port" 
            v-model="selectedDischargePort"
            class="select-field"
          >
            <option value="">Select Discharge Port</option>
            <option 
              v-for="port in dischargePorts" 
              :key="port.id" 
              :value="port.id"
            >
              {{ port.discharge_port }}
            </option>
          </select>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
      </div>

      <div class="modal-footer">
        <button 
          class="cancel-btn" 
          @click="closeModal"
          :disabled="loading"
        >
          Cancel
        </button>
        <button 
          class="save-btn" 
          @click="handleSubmit"
          :disabled="loading"
        >
          {{ loading ? 'Saving...' : 'Save Changes' }}
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

.form-group {
  margin-bottom: 16px;
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
  background-color: white;
}

.select-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.error-message {
  color: #ef4444;
  margin-bottom: 16px;
  font-size: 14px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.cancel-btn, .save-btn {
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  border: none;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
}

.cancel-btn:hover {
  background-color: #e5e7eb;
}

.save-btn:hover {
  background-color: #2563eb;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style> 