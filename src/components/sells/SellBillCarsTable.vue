<script setup>
import { ref, watch, defineProps, defineEmits, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { ElSelect, ElOption } from 'element-plus'
import 'element-plus/dist/index.css'

const props = defineProps({
  sellBillId: {
    type: Number,
    default: null,
  },
})

const emit = defineEmits(['refresh'])
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)

const { callApi } = useApi()
const cars = ref([])
const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)

// Add edit dialog state and form data
const showEditDialog = ref(false)
const editingCar = ref(null)
const clients = ref([])
const dischargePorts = ref([])
const editFormData = ref({
  id_client: null,
  id_port_discharge: null,
  price_cell: null,
  freight: null,
  rate: null,
  notes: null,
})

const filteredClients = ref([])
const filteredPorts = ref([])

// Add CFR DA calculation functions
const cfrDaInput = ref(null)

const calculatePriceFromCFRDA = (cfrDa, rate, freight) => {
  if (!cfrDa || !rate) return null
  const parsedCfrDa = parseFloat(cfrDa)
  const parsedRate = parseFloat(rate)
  const parsedFreight = parseFloat(freight) || 0
  return (parsedCfrDa / parsedRate - parsedFreight).toFixed(2)
}

const calculateCFRDAFromPrice = (price, rate, freight) => {
  if (!price || !rate) return null
  const parsedPrice = parseFloat(price)
  const parsedRate = parseFloat(rate)
  const parsedFreight = parseFloat(freight) || 0
  return ((parsedPrice + parsedFreight) * parsedRate).toFixed(2)
}

const handlePriceChange = (event) => {
  const newPrice = event.target.value
  if (newPrice && editFormData.value.rate) {
    cfrDaInput.value = calculateCFRDAFromPrice(
      newPrice,
      editFormData.value.rate,
      editFormData.value.freight,
    )
  }
}

const handleCFRDAChange = (event) => {
  const newCFRDA = event.target.value
  if (newCFRDA && editFormData.value.rate) {
    editFormData.value.price_cell = calculatePriceFromCFRDA(
      newCFRDA,
      editFormData.value.rate,
      editFormData.value.freight,
    )
  }
}

// Function to unassign a car from the sell bill
const unassignCar = async (carId) => {
  if (!confirm('Are you sure you want to unassign this car from the sell bill?')) {
    return
  }

  isProcessing.value = true
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = NULL, 
            id_client = NULL, 
            id_port_discharge = NULL, 
            freight = NULL,
            id_sell_pi = NULL
        WHERE id = ?
      `,
      params: [carId],
    })

    if (result.success) {
      // Refresh the cars list for this component
      await fetchCarsByBillId(props.sellBillId)

      // Emit refresh event to parent to update other components
      emit('refresh')
    } else {
      error.value = result.error || 'Failed to unassign car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

// Update client search method
const remoteClientSearch = (query) => {
  if (query) {
    filteredClients.value = clients.value.filter((client) =>
      client.name.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    filteredClients.value = clients.value
  }
}

// Update port search method
const remotePortSearch = (query) => {
  if (query) {
    filteredPorts.value = dischargePorts.value.filter((port) =>
      port.discharge_port.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    filteredPorts.value = dischargePorts.value
  }
}

// Update fetchClients to set filtered clients initially
const fetchClients = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name, mobiles
        FROM clients
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

// Update fetchDischargePorts to set filtered ports initially
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
      filteredPorts.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch discharge ports'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Handle edit button click
const handleEdit = async (car) => {
  isProcessing.value = true
  try {
    // Fetch full car details including client and port
    const result = await callApi({
      query: `
        SELECT 
          cs.*,
          c.name as client_name,
          dp.discharge_port
        FROM cars_stock cs
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        WHERE cs.id = ?
      `,
      params: [car.id],
    })

    if (result.success && result.data.length > 0) {
      const carData = result.data[0]
      editingCar.value = carData
      editFormData.value = {
        id_client: carData.id_client,
        id_port_discharge: carData.id_port_discharge,
        price_cell: carData.price_cell,
        freight: carData.freight,
        rate: carData.rate,
        notes: carData.notes,
      }

      // Calculate initial CFR DA
      if (carData.price_cell && carData.rate) {
        cfrDaInput.value = calculateCFRDAFromPrice(
          carData.price_cell,
          carData.rate,
          carData.freight,
        )
      }

      // Fetch necessary data for dropdowns
      await Promise.all([fetchClients(), fetchDischargePorts()])

      showEditDialog.value = true
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isProcessing.value = false
  }
}

// Handle save edit
const handleSaveEdit = async () => {
  if (!editingCar.value) return

  isProcessing.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_client = ?,
            id_port_discharge = ?,
            price_cell = ?,
            freight = ?,
            rate = ?,
            notes = ?
        WHERE id = ?
      `,
      params: [
        editFormData.value.id_client,
        editFormData.value.id_port_discharge,
        editFormData.value.price_cell,
        editFormData.value.freight,
        editFormData.value.rate,
        editFormData.value.notes || null,
        editingCar.value.id,
      ],
    })

    if (result.success) {
      showEditDialog.value = false
      editingCar.value = null
      await fetchCarsByBillId(props.sellBillId)
      emit('refresh')
    } else {
      error.value = result.error || 'Failed to update car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isProcessing.value = false
  }
}

// Add handleUnassign function
const handleUnassign = async (carId) => {
  await unassignCar(carId)
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
})

const fetchCarsByBillId = async (billId) => {
  if (!billId) {
    cars.value = []
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.notes,
          cs.price_cell,
          cs.freight,
          cs.rate,
          cs.date_loding,
          cs.path_documents,
          cs.is_big_car,
          cs.is_used_car,
          lp.loading_port,
          c.name as client_name,
          dp.discharge_port,
          bd.amount as buy_price,
          cn.car_name,
          clr.color
        FROM cars_stock cs
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        WHERE cs.id_sell = ?
      `,
      params: [billId],
    })

    if (result.success) {
      cars.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch cars'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Watch for changes in the sellBillId prop
watch(
  () => props.sellBillId,
  (newId) => {
    if (newId) {
      fetchCarsByBillId(newId)
    } else {
      cars.value = []
    }
  },
  { immediate: true },
)

// Calculate profit
const calculateProfit = (sellPrice, buyPrice, freight) => {
  if (!sellPrice || !buyPrice) return 'N/A'
  const sell = parseFloat(sellPrice)
  const buy = parseFloat(buyPrice)
  const freightCost = freight ? parseFloat(freight) : 0
  return (sell - buy - freightCost).toFixed(2)
}

// Add totalValue computed property
const totalValue = computed(() => {
  if (!cars.value || cars.value.length === 0) return 0
  return cars.value.reduce((total, car) => {
    const price = parseFloat(car.price_cell) || 0
    const freight = parseFloat(car.freight) || 0
    return total + price + freight
  }, 0)
})

// Add computed function for CFR DA calculation
const calculateCFRDA = (car) => {
  if (!car.price_cell || !car.rate) return 'N/A'
  const sellPrice = parseFloat(car.price_cell) || 0
  const freight = parseFloat(car.freight) || 0
  const rate = parseFloat(car.rate) || 0
  return ((sellPrice + freight) * rate).toFixed(2)
}

// Add filter refs
const filters = ref({
  carName: '',
  clientName: '',
  vin: '',
  minPrice: '',
  maxPrice: '',
})

// Add computed property for filtered cars
const filteredCars = computed(() => {
  return cars.value.filter((car) => {
    // Car name filter
    if (
      filters.value.carName &&
      !car.car_name?.toLowerCase().includes(filters.value.carName.toLowerCase())
    ) {
      return false
    }

    // Client name filter
    if (
      filters.value.clientName &&
      !car.client_name?.toLowerCase().includes(filters.value.clientName.toLowerCase())
    ) {
      return false
    }

    // VIN filter
    if (filters.value.vin && !car.vin?.toLowerCase().includes(filters.value.vin.toLowerCase())) {
      return false
    }

    // Price range filter
    const price = parseFloat(car.price_cell)
    if (filters.value.minPrice && price < parseFloat(filters.value.minPrice)) {
      return false
    }
    if (filters.value.maxPrice && price > parseFloat(filters.value.maxPrice)) {
      return false
    }

    return true
  })
})

// Add function to clear filters
const clearFilters = () => {
  filters.value = {
    carName: '',
    clientName: '',
    vin: '',
    minPrice: '',
    maxPrice: '',
  }
}

// Expose methods to parent component
defineExpose({
  fetchCarsByBillId,
})
</script>

<template>
  <div class="cars-table-component">
    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Loading cars...</span>
    </div>

    <div class="table-header">
      <h3>
        <i class="fas fa-car"></i>
        Cars in Bill
      </h3>
      <div class="total-info" v-if="cars.length > 0">
        <span>
          <i class="fas fa-hashtag"></i>
          Total Cars: {{ filteredCars.length }} / {{ cars.length }}
        </span>
        <span v-if="totalValue > 0">
          <i class="fas fa-dollar-sign"></i>
          Total Value: ${{ totalValue.toLocaleString() }}
        </span>
      </div>
    </div>

    <!-- Add Filters Section -->
    <div class="filters-section">
      <div class="filters-header">
        <h4>
          <i class="fas fa-filter"></i>
          Filters
        </h4>
        <button @click="clearFilters" class="clear-filters-btn">
          <i class="fas fa-times"></i>
          Clear Filters
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>
            <i class="fas fa-car"></i>
            Car Name
          </label>
          <input type="text" v-model="filters.carName" placeholder="Filter by car name..." />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-user"></i>
            Client Name
          </label>
          <input type="text" v-model="filters.clientName" placeholder="Filter by client name..." />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-fingerprint"></i>
            VIN
          </label>
          <input type="text" v-model="filters.vin" placeholder="Filter by VIN..." />
        </div>

        <div class="filter-group price-range">
          <label>
            <i class="fas fa-dollar-sign"></i>
            Price Range
          </label>
          <div class="price-inputs">
            <input type="number" v-model="filters.minPrice" placeholder="Min" step="0.01" min="0" />
            <span class="separator">-</span>
            <input type="number" v-model="filters.maxPrice" placeholder="Max" step="0.01" min="0" />
          </div>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-else-if="!sellBillId" class="no-selection">
      <i class="fas fa-hand-pointer fa-2x"></i>
      <p>Please select a sell bill to view its cars</p>
    </div>

    <div v-else-if="cars.length === 0" class="no-data">
      <i class="fas fa-car fa-2x"></i>
      <p>No cars found in this bill</p>
    </div>

    <div v-else-if="filteredCars.length === 0" class="no-data">
      <i class="fas fa-filter fa-2x"></i>
      <p>No cars match the current filters</p>
    </div>

    <table v-else class="cars-table">
      <thead>
        <tr>
          <th><i class="fas fa-hashtag"></i> ID</th>
          <th><i class="fas fa-car"></i> Car</th>
          <th><i class="fas fa-palette"></i> Color</th>
          <th><i class="fas fa-fingerprint"></i> VIN</th>
          <th><i class="fas fa-user"></i> Client</th>
          <th><i class="fas fa-dollar-sign"></i> Price</th>
          <th><i class="fas fa-ship"></i> Freight</th>
          <th><i class="fas fa-exchange-alt"></i> Rate</th>
          <th><i class="fas fa-calculator"></i> CFR DA</th>
          <th><i class="fas fa-sticky-note"></i> Notes</th>
          <th><i class="fas fa-cog"></i> Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in filteredCars" :key="car.id">
          <td>{{ car.id }}</td>
          <td>{{ car.car_name }}</td>
          <td>{{ car.color }}</td>
          <td>{{ car.vin || 'N/A' }}</td>
          <td>{{ car.client_name || 'N/A' }}</td>
          <td>{{ car.price_cell ? '$' + car.price_cell.toLocaleString() : 'N/A' }}</td>
          <td>{{ car.freight ? '$' + car.freight.toLocaleString() : 'N/A' }}</td>
          <td>{{ car.rate || 'N/A' }}</td>
          <td>{{ calculateCFRDA(car) }}</td>
          <td>{{ car.notes || 'N/A' }}</td>
          <td class="actions">
            <button
              @click="handleEdit(car)"
              :disabled="isProcessing"
              class="btn edit-btn"
              title="Edit Car"
            >
              <i class="fas fa-edit"></i>
            </button>
            <button
              @click="handleUnassign(car.id)"
              :disabled="isProcessing"
              class="btn unassign-btn"
              title="Unassign Car"
            >
              <i class="fas fa-unlink"></i>
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Edit Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-edit"></i>
            Edit Car Details
          </h3>
          <button @click="showEditDialog = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div v-if="editingCar" class="car-details">
          <p><strong>Car:</strong> {{ editingCar.car_name }}</p>
          <p><strong>VIN:</strong> {{ editingCar.vin }}</p>
        </div>

        <form @submit.prevent="handleSaveEdit" class="edit-form">
          <div class="form-group">
            <label>Client: <span class="required">*</span></label>
            <el-select
              v-model="editFormData.id_client"
              filterable
              remote
              :remote-method="remoteClientSearch"
              :loading="isProcessing"
              placeholder="Search client"
              class="custom-select"
              required
            >
              <el-option
                v-for="client in filteredClients"
                :key="client.id"
                :label="client.name"
                :value="client.id"
              >
                <div class="custom-option">
                  <i class="fas fa-user"></i>
                  <span>{{ client.name }}</span>
                  <small v-if="client.mobiles">
                    <i class="fas fa-phone"></i>
                    {{ client.mobiles }}
                  </small>
                </div>
              </el-option>
            </el-select>
          </div>

          <div class="form-group">
            <label>Discharge Port: <span class="required">*</span></label>
            <el-select
              v-model="editFormData.id_port_discharge"
              filterable
              remote
              :remote-method="remotePortSearch"
              :loading="isProcessing"
              placeholder="Search port"
              class="custom-select"
              required
            >
              <el-option
                v-for="port in filteredPorts"
                :key="port.id"
                :label="port.discharge_port"
                :value="port.id"
              >
                <div class="custom-option">
                  <i class="fas fa-anchor"></i>
                  <span>{{ port.discharge_port }}</span>
                </div>
              </el-option>
            </el-select>
          </div>

          <div class="form-group">
            <label>
              <i class="fas fa-dollar-sign"></i>
              Price: <span class="required">*</span>
            </label>
            <div class="input-with-info">
              <input
                type="number"
                v-model="editFormData.price_cell"
                step="0.01"
                min="0"
                required
                @input="handlePriceChange"
                :class="{ 'default-value': editFormData.price_cell === editingCar?.price_cell }"
              />
            </div>
          </div>

          <div class="form-group">
            <label>
              <i class="fas fa-ship"></i>
              Freight:
            </label>
            <div class="input-with-info">
              <input
                type="number"
                v-model="editFormData.freight"
                step="0.01"
                min="0"
                :class="{ 'default-value': editFormData.freight === editingCar?.freight }"
              />
            </div>
          </div>

          <div class="form-group">
            <label for="edit-rate">
              <i class="fas fa-exchange-alt"></i>
              Rate:
            </label>
            <input
              type="number"
              id="edit-rate"
              v-model="editFormData.rate"
              step="0.01"
              min="0"
              :disabled="isProcessing"
            />
          </div>

          <div class="form-group">
            <label for="edit-notes">
              <i class="fas fa-sticky-note"></i>
              Notes:
            </label>
            <textarea
              id="edit-notes"
              v-model="editFormData.notes"
              placeholder="Add any notes about this car..."
              rows="3"
              class="textarea-field"
              :disabled="isProcessing"
            ></textarea>
          </div>

          <div class="form-group">
            <label>
              <i class="fas fa-calculator"></i>
              CFR DA:
            </label>
            <div class="input-with-info">
              <input
                type="number"
                v-model="cfrDaInput"
                step="0.01"
                min="0"
                @input="handleCFRDAChange"
                :class="{ 'calculated-value': true }"
              />
            </div>
          </div>

          <div class="form-actions">
            <button
              type="button"
              @click="showEditDialog = false"
              class="cancel-btn"
              :disabled="isProcessing"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button type="submit" class="save-btn" :disabled="isProcessing">
              <i class="fas fa-save"></i>
              {{ isProcessing ? 'Saving...' : 'Save Changes' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cars-table-component {
  position: relative;
  margin: 1.5rem 0;
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

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.table-header h3 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1f2937;
  margin: 0;
}

.total-info {
  display: flex;
  gap: 1.5rem;
}

.total-info span {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  font-weight: 500;
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

.no-selection,
.no-data {
  text-align: center;
  padding: 2rem;
  background-color: #f9fafb;
  border-radius: 0.5rem;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.cars-table {
  width: 100%;
  border-collapse: collapse;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.cars-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #f3f4f6;
}

.cars-table th,
.cars-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.cars-table th {
  background-color: #f3f4f6;
  font-weight: 600;
  color: #374151;
}

.cars-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.cars-table tr:hover {
  background-color: #f9fafb;
}

.cars-table td i {
  margin-right: 0.5rem;
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.unassign-btn {
  background-color: #ef4444;
  color: white;
}

.unassign-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

/* Add smooth transitions */
.cars-table tr {
  transition: background-color 0.2s;
}

.btn i {
  transition: transform 0.2s;
}

.btn:hover:not(:disabled) i {
  transform: scale(1.1);
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.dialog-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  color: #666;
}

.car-details {
  background: #f9fafb;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 20px;
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.form-group label {
  font-weight: 500;
}

.form-group select,
.form-group input {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.cancel-btn,
.save-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
}

.cancel-btn {
  background: #f3f4f6;
  color: #374151;
}

.save-btn {
  background: #3b82f6;
  color: white;
}

.cancel-btn:hover:not(:disabled) {
  background: #e5e7eb;
}

.save-btn:hover:not(:disabled) {
  background: #2563eb;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.custom-select {
  width: 100%;
}

.custom-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.custom-option i {
  color: #6b7280;
  width: 16px;
}

.custom-option small {
  margin-left: auto;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 4px;
}

.required {
  color: #ef4444;
  margin-left: 2px;
}

:deep(.el-select) {
  width: 100%;
}

:deep(.el-select .el-input__wrapper) {
  background-color: white;
}

:deep(.el-select-dropdown__item) {
  padding: 0 12px;
}

:deep(.el-select-dropdown__item.selected) {
  background-color: #e5edff;
  color: #3b82f6;
}

:deep(.el-select-dropdown__item:hover) {
  background-color: #f3f4f6;
}

.input-with-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-with-info input {
  flex: 1;
}

.calculated-value {
  background-color: #f3f4f6;
  color: #1f2937;
  font-weight: 500;
  font-family: monospace;
}

.default-value {
  border-color: #10b981;
}

.filters-section {
  background-color: #f9fafb;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
  border: 1px solid #e5e7eb;
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filters-header h4 {
  margin: 0;
  font-size: 1rem;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.clear-filters-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  background-color: #ef4444;
  color: white;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.clear-filters-btn:hover {
  background-color: #dc2626;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #374151;
  font-weight: 500;
  font-size: 0.875rem;
}

.filter-group input {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
  transition: border-color 0.2s;
}

.filter-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.price-range .price-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.price-range .separator {
  color: #6b7280;
}

.price-range input {
  width: 100%;
}

/* Add transition for filtered rows */
.cars-table tbody tr {
  transition: opacity 0.2s;
}
</style>
