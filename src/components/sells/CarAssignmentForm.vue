<script setup>
import { ref, onMounted, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  carId: {
    type: Number,
    default: null
  },
  sellBillId: {
    type: Number,
    default: null
  },
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'assign-success'])

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const clients = ref([])
const dischargePorts = ref([])
const carDetails = ref(null)

const formData = ref({
  id_client: null,
  id_port_discharge: null,
  price_cell: null,
  freight: null
})

// Fetch clients for dropdown
const fetchClients = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name
        FROM clients
        WHERE is_broker = 0
        ORDER BY name ASC
      `,
      params: []
    })
    
    if (result.success) {
      clients.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch clients'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Fetch discharge ports for dropdown
const fetchDischargePorts = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, discharge_port
        FROM discharge_ports
        ORDER BY discharge_port ASC
      `,
      params: []
    })
    
    if (result.success) {
      dischargePorts.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch discharge ports'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Fetch car details to show in the form
const fetchCarDetails = async () => {
  if (!props.carId) return
  
  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cn.car_name,
          clr.color,
          bd.price_sell as buy_price
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        WHERE cs.id = ?
      `,
      params: [props.carId]
    })
    
    if (result.success && result.data.length > 0) {
      carDetails.value = result.data[0]
    } else {
      error.value = result.error || 'Failed to fetch car details'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Assign car with collected data
const assignCar = async () => {
  if (!validateForm()) return
  
  loading.value = true
  error.value = null
  
  try {
    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = ?,
            id_client = ?,
            id_port_discharge = ?,
            price_cell = ?,
            freight = ?
        WHERE id = ?
      `,
      params: [
        props.sellBillId,
        formData.value.id_client,
        formData.value.id_port_discharge,
        formData.value.price_cell,
        formData.value.freight || null,
        props.carId
      ]
    })
    
    if (result.success) {
      emit('assign-success')
      resetForm()
    } else {
      error.value = result.error || 'Failed to assign car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Validate form before submission
const validateForm = () => {
  if (!formData.value.id_client) {
    error.value = 'Please select a client'
    return false
  }
  
  if (!formData.value.id_port_discharge) {
    error.value = 'Please select a discharge port'
    return false
  }
  
  if (!formData.value.price_cell) {
    error.value = 'Please enter a sell price'
    return false
  }
  
  return true
}

// Reset form after submission
const resetForm = () => {
  formData.value = {
    id_client: null,
    id_port_discharge: null,
    price_cell: null,
    freight: null
  }
  error.value = null
}

// Calculate potential profit
const calculateProfit = () => {
  if (!carDetails.value?.buy_price || !formData.value.price_cell) return 'N/A'
  
  const buyPrice = parseFloat(carDetails.value.buy_price)
  const sellPrice = parseFloat(formData.value.price_cell)
  const freightCost = formData.value.freight ? parseFloat(formData.value.freight) : 0
  
  return (sellPrice - buyPrice - freightCost).toFixed(2)
}

onMounted(() => {
  fetchClients()
  fetchDischargePorts()
  if (props.carId) {
    fetchCarDetails()
  }
})
</script>

<template>
  <div v-if="visible" class="car-assignment-form-overlay">
    <div class="car-assignment-form">
      <div class="form-header">
        <h3>Assign Car to Sell Bill</h3>
        <button @click="$emit('close')" class="close-btn">&times;</button>
      </div>
      
      <div v-if="error" class="error">{{ error }}</div>
      
      <div v-if="carDetails" class="car-details">
        <div class="detail-item">
          <span class="label">Car:</span>
          <span class="value">{{ carDetails.car_name }}</span>
        </div>
        <div class="detail-item">
          <span class="label">Color:</span>
          <span class="value">{{ carDetails.color }}</span>
        </div>
        <div class="detail-item">
          <span class="label">VIN:</span>
          <span class="value">{{ carDetails.vin }}</span>
        </div>
        <div class="detail-item">
          <span class="label">Buy Price:</span>
          <span class="value">${{ carDetails.buy_price }}</span>
        </div>
      </div>
      
      <form @submit.prevent="assignCar">
        <div class="form-group">
          <label for="client">Client: <span class="required">*</span></label>
          <select id="client" v-model="formData.id_client" required>
            <option value="">Select Client</option>
            <option v-for="client in clients" :key="client.id" :value="client.id">
              {{ client.name }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="discharge-port">Discharge Port: <span class="required">*</span></label>
          <select id="discharge-port" v-model="formData.id_port_discharge" required>
            <option value="">Select Discharge Port</option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
              {{ port.discharge_port }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="sell-price">Sell Price: <span class="required">*</span></label>
          <input 
            type="number" 
            id="sell-price" 
            v-model="formData.price_cell" 
            step="0.01" 
            min="0" 
            required
          >
        </div>
        
        <div class="form-group">
          <label for="freight">Freight:</label>
          <input 
            type="number" 
            id="freight" 
            v-model="formData.freight" 
            step="0.01" 
            min="0"
          >
        </div>
        
        <div v-if="formData.price_cell && carDetails?.buy_price" class="profit-calculation">
          <span class="label">Estimated Profit:</span>
          <span class="value profit">${{ calculateProfit() }}</span>
        </div>
        
        <div class="form-actions">
          <button type="button" @click="$emit('close')" class="cancel-btn">Cancel</button>
          <button type="submit" class="assign-btn" :disabled="loading">
            {{ loading ? 'Assigning...' : 'Assign Car' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.car-assignment-form-overlay {
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

.car-assignment-form {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.form-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #666;
}

.car-details {
  background-color: #f9fafb;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  margin-bottom: 8px;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.label {
  font-weight: 600;
  width: 100px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
}

.required {
  color: #ef4444;
}

select, input {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.9rem;
}

.profit-calculation {
  background-color: #ecfdf5;
  padding: 10px;
  border-radius: 4px;
  margin: 15px 0;
  display: flex;
  align-items: center;
}

.profit {
  color: #10b981;
  font-weight: 600;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}

.assign-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}

.assign-btn:disabled {
  background-color: #93c5fd;
  cursor: not-allowed;
}

.error {
  color: #ef4444;
  background-color: #fee2e2;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
}
</style>