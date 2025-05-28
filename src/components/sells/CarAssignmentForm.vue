<script setup>
import { ref, onMounted, watch, defineProps, defineEmits, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { ElSelect, ElOption } from 'element-plus'
import 'element-plus/dist/index.css'

const props = defineProps({
  carId: {
    type: Number,
    default: null,
  },
  sellBillId: {
    type: Number,
    default: null,
  },
  visible: {
    type: Boolean,
    default: false,
  },
})
const user = ref(JSON.parse(localStorage.getItem('user')))
const isAdmin = computed(() => user.value?.role_id === 1)

const emit = defineEmits(['close', 'assign-success'])

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const clients = ref([])
const filteredClients = ref([])
const dischargePorts = ref([])
const carDetails = ref(null)

// Add loading state for form submission
const isSubmitting = ref(false)
const isProcessing = ref(false)

const formData = ref({
  id_client: null,
  id_port_discharge: null,
  price_cell: null,
  freight: null,
  rate: null,
})

// Add ref for CFR DA input
const cfrDaInput = ref(null)

// Add function to calculate price from CFR DA
const calculatePriceFromCFRDA = (cfrDa, rate, freight) => {
  if (!cfrDa || !rate) return null
  const parsedCfrDa = parseFloat(cfrDa)
  const parsedRate = parseFloat(rate)
  const parsedFreight = parseFloat(freight) || 0
  return (parsedCfrDa / parsedRate - parsedFreight).toFixed(2)
}

// Add function to calculate CFR DA from price
const calculateCFRDAFromPrice = (price, rate, freight) => {
  if (!price || !rate) return null
  const parsedPrice = parseFloat(price)
  const parsedRate = parseFloat(rate)
  const parsedFreight = parseFloat(freight) || 0
  return ((parsedPrice + parsedFreight) * parsedRate).toFixed(2)
}

// Add handlers for input changes
const handlePriceChange = (event) => {
  const newPrice = event.target.value
  if (newPrice && formData.value.rate) {
    cfrDaInput.value = calculateCFRDAFromPrice(
      newPrice,
      formData.value.rate,
      formData.value.freight,
    )
  }
}

const handleCFRDAChange = (event) => {
  const newCFRDA = event.target.value
  if (newCFRDA && formData.value.rate) {
    formData.value.price_cell = calculatePriceFromCFRDA(
      newCFRDA,
      formData.value.rate,
      formData.value.freight,
    )
  }
}

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
      params: [],
    })

    if (result.success) {
      clients.value = result.data
      filteredClients.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch clients'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const remoteMethod = (query) => {
  if (query) {
    // Filter the existing clients data
    filteredClients.value = clients.value.filter((client) =>
      client.name.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    // If no query, show all clients
    filteredClients.value = clients.value
  }
}

const handleClientChange = () => {
  // Handle client change if needed
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
      params: [],
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

// Add new function to fetch default freight values
const fetchDefaultFreight = async (isSmallCar) => {
  try {
    const result = await callApi({
      query: 'SELECT freight_small, freight_big FROM defaults LIMIT 1',
      params: [],
    })

    if (result.success && result.data.length > 0) {
      const defaultFreight = isSmallCar ? result.data[0].freight_small : result.data[0].freight_big
      formData.value.freight = defaultFreight
    }
  } catch (err) {
    console.error('Error fetching default freight:', err)
  }
}

// Update the fetchCarDetails function to include car size info and set freight
const fetchCarDetails = async () => {
  if (!props.carId) return

  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.price_cell,
          cs.is_big_car,
          cn.car_name,
          clr.color,
          bd.price_sell as buy_price
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        WHERE cs.id = ?
      `,
      params: [props.carId],
    })

    if (result.success && result.data.length > 0) {
      carDetails.value = result.data[0]
      // Set the default sell price if available
      if (carDetails.value.price_cell) {
        formData.value.price_cell = carDetails.value.price_cell
      }
      // Fetch and set default freight based on car size
      await fetchDefaultFreight(!carDetails.value.is_big_car)
      // After fetching default rate, calculate initial CFR DA
      await fetchDefaultRate()
      if (formData.value.price_cell && formData.value.rate) {
        cfrDaInput.value = calculateCFRDAFromPrice(
          formData.value.price_cell,
          formData.value.rate,
          formData.value.freight,
        )
      }
    } else {
      error.value = result.error || 'Failed to fetch car details'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Update fetchDefaultRate to calculate initial CFR DA
const fetchDefaultRate = async () => {
  try {
    const result = await callApi({
      query: 'SELECT rate FROM defaults LIMIT 1',
      params: [],
    })

    if (result.success && result.data.length > 0) {
      formData.value.rate = result.data[0].rate
      // Calculate initial CFR DA if price is available
      if (formData.value.price_cell) {
        cfrDaInput.value = calculateCFRDAFromPrice(
          formData.value.price_cell,
          formData.value.rate,
          formData.value.freight,
        )
      }
    }
  } catch (err) {
    console.error('Error fetching default rate:', err)
  }
}

// Add computed property for CFR DA calculation
const calculateCFRDA = computed(() => {
  if (!formData.value.price_cell || !formData.value.rate) return 'N/A'
  const sellPrice = parseFloat(formData.value.price_cell) || 0
  const freight = parseFloat(formData.value.freight) || 0
  const rate = parseFloat(formData.value.rate) || 0
  return ((sellPrice + freight) * rate).toFixed(2)
})

// Assign car with collected data
const assignCar = async () => {
  // Prevent multiple submissions
  if (isSubmitting.value) {
    return
  }

  if (!validateForm()) return

  try {
    isSubmitting.value = true
    loading.value = true
    error.value = null

    // First get the bill_ref from the sell_bill
    const billResult = await callApi({
      query: `
        SELECT bill_ref
        FROM sell_bill
        WHERE id = ?
      `,
      params: [props.sellBillId],
    })

    if (!billResult.success || !billResult.data.length) {
      error.value = 'Failed to get bill reference'
      return
    }

    const billRef = billResult.data[0].bill_ref
    const currentDate = new Date().toISOString().split('T')[0]

    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = ?,
            id_client = ?,
            id_port_discharge = ?,
            price_cell = ?,
            freight = ?,
            date_sell = ?,
            id_sell_pi = ?,
            rate = ?
        WHERE id = ?
      `,
      params: [
        props.sellBillId,
        formData.value.id_client,
        formData.value.id_port_discharge,
        formData.value.price_cell,
        formData.value.freight || null,
        currentDate,
        billRef,
        formData.value.rate || null,
        props.carId,
      ],
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
    isSubmitting.value = false
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

  if (!formData.value.rate) {
    error.value = 'Please enter a rate'
    return false
  }

  return true
}

// Reset form after submission
const resetForm = () => {
  const currentRate = formData.value.rate // Preserve current rate
  formData.value = {
    id_client: null,
    id_port_discharge: null,
    price_cell: null,
    freight: null,
    rate: currentRate, // Keep the default rate
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
  fetchDefaultRate()
  if (props.carId) {
    fetchCarDetails()
  }
})
</script>

<template>
  <div v-if="visible" class="car-assignment-form-overlay">
    <div class="car-assignment-form">
      <!-- Loading Overlay -->
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>{{ isProcessing ? 'Assigning car...' : 'Loading...' }}</span>
      </div>

      <div class="form-header">
        <h3>
          <i class="fas fa-link"></i>
          Assign Car to Bill
        </h3>
        <button @click="$emit('close')" class="close-btn">&times;</button>
      </div>

      <div v-if="error" class="error">{{ error }}</div>

      <div v-if="carDetails" class="car-details">
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-car"></i>
            Car:
          </span>
          <span class="value">{{ carDetails.car_name }}</span>
        </div>
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-palette"></i>
            Color:
          </span>
          <span class="value">{{ carDetails.color }}</span>
        </div>
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-fingerprint"></i>
            VIN:
          </span>
          <span class="value">{{ carDetails.vin || 'N/A' }}</span>
        </div>
        <div v-if="isAdmin" class="detail-item">
          <span class="label">
            <i class="fas fa-dollar-sign"></i>
            Buy Price:
          </span>
          <span class="value">${{ carDetails.buy_price }}</span>
        </div>
      </div>

      <form @submit.prevent="assignCar">
        <div class="form-group">
          <label for="client">
            <i class="fas fa-user"></i>
            Client: <span class="required">*</span>
          </label>
          <el-select
            v-model="formData.id_client"
            filterable
            remote
            :remote-method="remoteMethod"
            :loading="loading"
            placeholder="Select or search client"
            style="width: 100%"
            @change="handleClientChange"
          >
            <el-option
              v-for="client in filteredClients"
              :key="client.id"
              :label="client.name"
              :value="client.id"
            >
              <i class="fas fa-user"></i>
              {{ client.name }}
              <small v-if="client.mobiles">
                <i class="fas fa-phone"></i>
                {{ client.mobiles }}
              </small>
            </el-option>
          </el-select>
        </div>

        <div class="form-group">
          <label for="discharge-port">
            <i class="fas fa-anchor"></i>
            Discharge Port: <span class="required">*</span>
          </label>
          <select id="discharge-port" v-model="formData.id_port_discharge" required>
            <option value="">Select Discharge Port</option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
              {{ port.discharge_port }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label for="cfr-da">
            <i class="fas fa-calculator"></i>
            CFR DA:
          </label>
          <div class="input-with-info">
            <input
              type="number"
              id="cfr-da"
              v-model="cfrDaInput"
              step="0.01"
              min="0"
              @input="handleCFRDAChange"
              :class="{ 'calculated-value': true }"
            />
          </div>
        </div>
        <div class="form-group">
          <label for="sell-price">
            <i class="fas fa-dollar-sign"></i>
            Sell Price: <span class="required">*</span>
          </label>
          <div class="input-with-info">
            <input
              type="number"
              id="sell-price"
              v-model="formData.price_cell"
              step="0.01"
              min="0"
              required
              @input="handlePriceChange"
              :class="{
                'default-value':
                  carDetails?.price_cell && formData.price_cell === carDetails.price_cell,
              }"
            />
            <span
              v-if="carDetails?.price_cell && formData.price_cell === carDetails.price_cell"
              class="default-badge"
            >
              Default
            </span>
          </div>
        </div>

        <div class="form-group">
          <label for="freight">
            <i class="fas fa-ship"></i>
            Freight:
          </label>
          <div class="input-with-info">
            <input
              type="number"
              id="freight"
              v-model="formData.freight"
              step="0.01"
              min="0"
              :class="{ 'default-value': formData.freight !== null }"
            />
            <span v-if="formData.freight !== null" class="default-badge">
              Default ({{ carDetails?.is_big_car ? 'Big' : 'Small' }} Car)
            </span>
          </div>
        </div>

        <div class="form-group">
          <label for="rate">
            <i class="fas fa-exchange-alt"></i>
            Rate: <span class="required">*</span>
          </label>
          <div class="input-with-info">
            <input
              type="number"
              id="rate"
              v-model="formData.rate"
              step="0.01"
              min="0"
              required
              :class="{ 'default-value': formData.rate !== null }"
            />
            <span v-if="formData.rate !== null" class="default-badge">Default</span>
          </div>
        </div>

        <div
          v-if="formData.price_cell && carDetails?.buy_price && isAdmin"
          class="profit-calculation"
        >
          <span class="label">
            <i class="fas fa-dollar-sign"></i>
            Estimated Profit:
          </span>
          <span class="value profit">${{ calculateProfit() }}</span>
        </div>

        <div class="form-actions">
          <button type="button" @click="$emit('close')" class="cancel-btn">
            <i class="fas fa-times"></i>
            Cancel
          </button>
          <button type="submit" class="assign-btn" :disabled="isProcessing">
            <i class="fas fa-save"></i>
            {{ isProcessing ? 'Assigning...' : 'Assign Car' }}
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

select,
input {
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
  opacity: 0.6;
}

.spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid #ffffff;
  border-radius: 50%;
  border-top-color: transparent;
  animation: spin 1s ease-in-out infinite;
  margin-right: 8px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error {
  color: #ef4444;
  background-color: #fee2e2;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
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
  gap: 1rem;
  z-index: 10;
}

.form-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1f2937;
  margin-bottom: 1.5rem;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  font-weight: 500;
}

.form-group label i {
  color: #6b7280;
}

.car-details {
  background-color: #f9fafb;
  padding: 1rem;
  border-radius: 0.5rem;
  border: 1px solid #e5e7eb;
}

.car-details p {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.5rem 0;
  color: #374151;
}

.car-details i {
  color: #6b7280;
}

.client-select {
  width: 100%;
}

input[type='number'],
select,
textarea {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

input[type='number']:focus,
select:focus,
textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

textarea {
  min-height: 100px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}

.cancel-btn,
.save-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #059669;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Element Plus Select Customization */
:deep(.el-select-dropdown__item) {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

:deep(.el-select-dropdown__item i) {
  color: #6b7280;
}

:deep(.el-select-dropdown__item small) {
  margin-left: auto;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

/* Add smooth transitions */
.form-group input,
.form-group select,
.form-group textarea {
  transition: all 0.2s;
}

button i {
  transition: transform 0.2s;
}

button:hover:not(:disabled) i {
  transform: scale(1.1);
}

.calculated-value {
  background-color: #f3f4f6;
  color: #1f2937;
  font-weight: 500;
  font-family: monospace;
}

.input-with-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-with-info input {
  flex: 1;
}
</style>
