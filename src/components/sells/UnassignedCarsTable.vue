<script setup>
import { ref, onMounted, defineEmits, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import CarAssignmentForm from './CarAssignmentForm.vue'

const emit = defineEmits(['refresh', 'assign-car'])

const { callApi } = useApi()
const unassignedCars = ref([])
const allUnassignedCars = ref([]) // Store all unassigned cars before filtering
const loading = ref(true)
const error = ref(null)
const selectedSellBillId = ref(null)
const showAssignmentForm = ref(false)
const selectedCarId = ref(null)
const isProcessing = ref(false)
const isBatchSell = ref(false)

// Filter states
const filters = ref({
  carName: '',
  color: '',
  vin: '',
  buyBillRef: '',
})
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)

// Add sort config after the filters
const sortConfig = ref({
  field: 'id',
  direction: 'desc',
})

// Add computed property for sorted cars
const sortedCars = computed(() => {
  if (!unassignedCars.value.length) return []

  return [...unassignedCars.value].sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle date comparison
    if (sortConfig.value.field === 'date_sell') {
      aValue = aValue ? new Date(aValue).getTime() : 0
      bValue = bValue ? new Date(bValue).getTime() : 0
    }

    // Handle numeric fields
    if (['id', 'price_cell', 'buy_price'].includes(sortConfig.value.field)) {
      aValue = Number(aValue) || 0
      bValue = Number(bValue) || 0
    }

    // Handle null values
    if (aValue === null || aValue === undefined) aValue = ''
    if (bValue === null || bValue === undefined) bValue = ''

    // Compare values based on direction
    if (sortConfig.value.direction === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })
})

const handleSort = (field) => {
  if (sortConfig.value.field === field) {
    // Toggle direction if clicking the same field
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    // Set new field and default to descending
    sortConfig.value.field = field
    sortConfig.value.direction = 'desc'
  }
}

// Function to fetch all unassigned cars
const fetchUnassignedCars = async () => {
  loading.value = true
  error.value = null

  try {
    let myQuery = `SELECT 
          cs.id,
          cs.vin,
          cs.notes,
          cs.price_cell,
          cs.freight,
          cs.path_documents,
          bd.amount as buy_price,
          cn.car_name,
          clr.color,
          clr.hexa,
          bb.bill_ref as buy_bill_ref
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
        WHERE cs.id_sell IS NULL
        `
    if (!isAdmin.value) {
      myQuery += ' AND cs.hidden = 0'
    }
    myQuery += ' ORDER BY cs.id DESC'

    const result = await callApi({
      query: myQuery,
      params: [],
    })

    if (result.success) {
      allUnassignedCars.value = result.data
      applyFilters() // Apply filters to the fetched data
    } else {
      error.value = result.error || 'Failed to fetch unassigned cars'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Apply filters to the cars list
const applyFilters = () => {
  unassignedCars.value = allUnassignedCars.value.filter((car) => {
    const matchCarName =
      !filters.value.carName ||
      (car.car_name && car.car_name.toLowerCase().includes(filters.value.carName.toLowerCase()))

    const matchColor =
      !filters.value.color ||
      (car.color && car.color.toLowerCase().includes(filters.value.color.toLowerCase()))

    const matchVin =
      !filters.value.vin ||
      (car.vin && car.vin.toLowerCase().includes(filters.value.vin.toLowerCase()))

    const matchBuyBillRef =
      !filters.value.buyBillRef ||
      (car.buy_bill_ref &&
        car.buy_bill_ref.toLowerCase().includes(filters.value.buyBillRef.toLowerCase()))

    return matchCarName && matchColor && matchVin && matchBuyBillRef
  })
}

// Reset all filters
const resetFilters = () => {
  filters.value = {
    carName: '',
    color: '',
    vin: '',
    buyBillRef: '',
  }
  applyFilters()
}

// Watch for filter changes
const handleFilterChange = () => {
  applyFilters()
}

// Function to check if the selected bill is a batch sell
const checkBatchSell = async (billId) => {
  try {
    const result = await callApi({
      query: `
        SELECT is_batch_sell
        FROM sell_bill
        WHERE id = ?
      `,
      params: [billId],
    })

    if (result.success && result.data.length > 0) {
      isBatchSell.value = result.data[0].is_batch_sell === 1
    }
  } catch (err) {
    console.error('Error checking batch sell:', err)
  }
}

// Add persistent pricing for batch sell
const batchPrice = ref(null)
const batchFreight = ref(null)
const batchRate = ref(null)
const batchDischargePort = ref(null)
const showBatchPricing = ref(false)

// Add discharge ports for batch sell
const dischargePorts = ref([])

// Function to calculate text color based on background color
const getTextColor = (backgroundColor) => {
  if (!backgroundColor) return '#374151'

  // Convert hex to RGB
  const hex = backgroundColor.replace('#', '')
  const r = parseInt(hex.substr(0, 2), 16)
  const g = parseInt(hex.substr(2, 2), 16)
  const b = parseInt(hex.substr(4, 2), 16)

  // Calculate luminance
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255

  // Return white text for dark backgrounds, black text for light backgrounds
  return luminance > 0.5 ? '#000000' : '#ffffff'
}

// Function to fetch discharge ports
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
    }
  } catch (err) {
    console.error('Error fetching discharge ports:', err)
  }
}

// Add broker info for batch sell
const brokerInfo = ref(null)

// Function to fetch broker info
const fetchBrokerInfo = async (billId) => {
  try {
    const result = await callApi({
      query: `
        SELECT c.name as broker_name, c.mobiles as broker_phone
        FROM sell_bill sb
        LEFT JOIN clients c ON sb.id_broker = c.id
        WHERE sb.id = ?
      `,
      params: [billId],
    })

    if (result.success && result.data.length > 0) {
      brokerInfo.value = result.data[0]
    }
  } catch (err) {
    console.error('Error fetching broker info:', err)
  }
}

// Function to directly assign car for batch sell with pricing
const assignCarToBatchSell = async (carId) => {
  if (isProcessing.value) return
  isProcessing.value = true

  try {
    // Get the bill_ref and broker_id from the sell_bill
    const billResult = await callApi({
      query: `
        SELECT bill_ref, id_broker
        FROM sell_bill
        WHERE id = ?
      `,
      params: [selectedSellBillId.value],
    })

    if (!billResult.success || !billResult.data.length) {
      error.value = 'Failed to get bill reference'
      return
    }

    const billRef = billResult.data[0].bill_ref
    const brokerId = billResult.data[0].id_broker
    const currentDate = new Date().toISOString().split('T')[0]

    const result = await callApi({
      query: `
        UPDATE cars_stock
        SET id_sell = ?,
            id_client = ?,
            id_port_discharge = ?,
            price_cell = ?,
            freight = ?,
            rate = ?,
            date_sell = ?,
            id_sell_pi = ?,
            date_assigned = NOW()
        WHERE id = ?
      `,
      params: [
        selectedSellBillId.value,
        brokerId, // Set broker as client
        batchDischargePort.value || null,
        batchPrice.value || null,
        batchFreight.value || null,
        batchRate.value || null,
        currentDate,
        billRef,
        carId,
      ],
    })

    if (result.success) {
      await fetchUnassignedCars()
      emit('refresh')
    } else {
      error.value = result.error || 'Failed to assign car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isProcessing.value = false
  }
}

// Function to toggle batch pricing visibility
const toggleBatchPricing = () => {
  showBatchPricing.value = !showBatchPricing.value
}

// Function to clear batch pricing
const clearBatchPricing = () => {
  batchPrice.value = null
  batchFreight.value = null
  batchRate.value = null
  batchDischargePort.value = null
}

// Function to set batch pricing from defaults
const setDefaultBatchPricing = async () => {
  try {
    const result = await callApi({
      query: 'SELECT freight_small, freight_big FROM defaults LIMIT 1',
      params: [],
    })

    if (result.success && result.data.length > 0) {
      const defaults = result.data[0]
      // Set default freight (you can choose small or big based on your needs)
      batchFreight.value = defaults.freight_small || null
    }
  } catch (err) {
    console.error('Error fetching default pricing:', err)
  }
}

// Open assignment form for a car
const openAssignmentForm = async (carId) => {
  if (!selectedSellBillId.value) {
    alert('Please select a sell bill first')
    return
  }

  if (isProcessing.value) return

  // Check if it's a batch sell
  await checkBatchSell(selectedSellBillId.value)

  if (isBatchSell.value) {
    // If it's a batch sell, assign directly without showing the form
    await assignCarToBatchSell(carId)
  } else {
    // If it's not a batch sell, show the assignment form as usual
    selectedCarId.value = carId
    showAssignmentForm.value = true
  }
}

// Handle successful assignment
const handleAssignSuccess = () => {
  showAssignmentForm.value = false
  fetchUnassignedCars()
  emit('refresh')
}

// Set the selected sell bill ID
const setSellBillId = async (billId) => {
  selectedSellBillId.value = billId
  await checkBatchSell(billId)
  if (billId) {
    await fetchBrokerInfo(billId)
  }
}

// Calculate profit
const calculateProfit = (sellPrice, buyPrice, freight) => {
  if (!sellPrice || !buyPrice) return 'N/A'
  const sell = parseFloat(sellPrice)
  const buy = parseFloat(buyPrice)
  const freightCost = freight ? parseFloat(freight) : 0
  return (sell - buy - freightCost).toFixed(2)
}

// Expose methods to parent component
defineExpose({
  fetchUnassignedCars,
  setSellBillId,
})

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchUnassignedCars()
  fetchDischargePorts()
})
</script>

<template>
  <div class="unassigned-cars-table">
    <div class="table-header">
      <h3>
        <i class="fas fa-car"></i>
        Unassigned Cars
      </h3>
      <div class="total-info" v-if="unassignedCars.length > 0">
        <span>
          <i class="fas fa-hashtag"></i>
          Total Cars: {{ unassignedCars.length }}
        </span>
      </div>
    </div>
    <!-- Filters Section -->
    <div class="filters-section">
      <div class="filters-header">
        <h3>
          <i class="fas fa-filter"></i>
          Filters
        </h3>
        <button @click="resetFilters" class="reset-btn">
          <i class="fas fa-undo"></i>
          Reset
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>
            <i class="fas fa-car"></i>
            Car Name:
          </label>
          <input
            type="text"
            v-model="filters.carName"
            @input="handleFilterChange"
            placeholder="Search car name..."
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-palette"></i>
            Color:
          </label>
          <input
            type="text"
            v-model="filters.color"
            @input="handleFilterChange"
            placeholder="Search color..."
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-fingerprint"></i>
            VIN:
          </label>
          <input
            type="text"
            v-model="filters.vin"
            @input="handleFilterChange"
            placeholder="Search VIN..."
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-file-alt"></i>
            Buy Bill Ref:
          </label>
          <input
            type="text"
            v-model="filters.buyBillRef"
            @input="handleFilterChange"
            placeholder="Search buy bill ref..."
          />
        </div>
      </div>
    </div>

    <!-- Batch Pricing Section (only show for batch sell) -->
    <div v-if="isBatchSell && selectedSellBillId" class="batch-pricing-section">
      <div class="batch-pricing-header">
        <h4>
          <i class="fas fa-dollar-sign"></i>
          Batch Sell Pricing
          <span
            v-if="batchPrice || batchFreight || batchRate || batchDischargePort"
            class="pricing-active-indicator"
          >
            <i class="fas fa-check-circle"></i>
            Active
          </span>
        </h4>
        <div class="batch-pricing-actions">
          <button
            @click="toggleBatchPricing"
            class="toggle-pricing-btn"
            :class="{ active: showBatchPricing }"
          >
            <i class="fas fa-cog"></i>
            {{ showBatchPricing ? 'Hide' : 'Show' }} Pricing
          </button>
          <button @click="setDefaultBatchPricing" class="default-btn" title="Set Default Freight">
            <i class="fas fa-magic"></i>
            Defaults
          </button>
          <button @click="clearBatchPricing" class="clear-btn" title="Clear All Pricing">
            <i class="fas fa-eraser"></i>
            Clear
          </button>
        </div>
      </div>

      <!-- Broker Info -->
      <div v-if="brokerInfo" class="broker-info">
        <i class="fas fa-user-tie"></i>
        <span><strong>Broker/Client:</strong> {{ brokerInfo.broker_name }}</span>
        <span v-if="brokerInfo.broker_phone" class="broker-phone">
          <i class="fas fa-phone"></i>
          {{ brokerInfo.broker_phone }}
        </span>
      </div>

      <div v-if="showBatchPricing" class="batch-pricing-form">
        <div class="pricing-info">
          <i class="fas fa-info-circle"></i>
          <span
            >Set pricing and discharge port once and they will be applied to all cars you assign.
            The broker will automatically be set as the client for all cars. You can change the
            settings at any time.</span
          >
        </div>

        <div class="pricing-grid">
          <div class="pricing-group">
            <label>
              <i class="fas fa-dollar-sign"></i>
              Price (USD):
            </label>
            <input
              type="number"
              v-model="batchPrice"
              placeholder="Enter price..."
              step="0.01"
              min="0"
            />
          </div>

          <div class="pricing-group">
            <label>
              <i class="fas fa-ship"></i>
              Freight (USD):
            </label>
            <input
              type="number"
              v-model="batchFreight"
              placeholder="Enter freight..."
              step="0.01"
              min="0"
            />
          </div>

          <div class="pricing-group">
            <label>
              <i class="fas fa-exchange-alt"></i>
              Rate (DA):
            </label>
            <input
              type="number"
              v-model="batchRate"
              placeholder="Enter rate..."
              step="0.01"
              min="0"
            />
          </div>

          <div class="pricing-group">
            <label>
              <i class="fas fa-anchor"></i>
              Discharge Port:
            </label>
            <select v-model="batchDischargePort" class="discharge-port-select">
              <option value="">Select discharge port...</option>
              <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
                {{ port.discharge_port }}
              </option>
            </select>
          </div>
        </div>

        <div
          class="pricing-summary"
          v-if="batchPrice || batchFreight || batchRate || batchDischargePort"
        >
          <div class="summary-item">
            <span class="label">Price:</span>
            <span class="value">{{
              batchPrice ? '$' + parseFloat(batchPrice).toLocaleString() : 'Not set'
            }}</span>
          </div>
          <div class="summary-item">
            <span class="label">Freight:</span>
            <span class="value">{{
              batchFreight ? '$' + parseFloat(batchFreight).toLocaleString() : 'Not set'
            }}</span>
          </div>
          <div class="summary-item">
            <span class="label">Rate:</span>
            <span class="value">{{
              batchRate ? parseFloat(batchRate).toLocaleString() + ' DA' : 'Not set'
            }}</span>
          </div>
          <div class="summary-item">
            <span class="label">Discharge Port:</span>
            <span class="value">{{
              batchDischargePort
                ? dischargePorts.find((p) => p.id === batchDischargePort)?.discharge_port ||
                  'Selected'
                : 'Not set'
            }}</span>
          </div>
          <div class="summary-item total" v-if="batchPrice && batchRate">
            <span class="label">Total (DA):</span>
            <span class="value"
              >{{
                (
                  (parseFloat(batchPrice) + (parseFloat(batchFreight) || 0)) *
                  parseFloat(batchRate)
                ).toLocaleString()
              }}
              DA</span
            >
          </div>
        </div>
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Loading unassigned cars...</span>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-else-if="unassignedCars.length === 0" class="no-data">
      <i class="fas fa-car fa-2x"></i>
      <p>No unassigned cars available</p>
    </div>

    <table v-else class="cars-table">
      <thead>
        <tr>
          <th @click="handleSort('id')" class="sortable">
            <i class="fas fa-hashtag"></i> ID
            <i
              v-if="sortConfig.field === 'id'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('car_name')" class="sortable">
            <i class="fas fa-car"></i> Car
            <i
              v-if="sortConfig.field === 'car_name'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('color')" class="sortable">
            <i class="fas fa-palette"></i> Color
            <i
              v-if="sortConfig.field === 'color'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('vin')" class="sortable">
            <i class="fas fa-fingerprint"></i> VIN
            <i
              v-if="sortConfig.field === 'vin'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('price_cell')" class="sortable">
            <i class="fas fa-dollar-sign"></i> Sell Price
            <i
              v-if="sortConfig.field === 'price_cell'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th v-if="isAdmin" @click="handleSort('buy_price')" class="sortable">
            <i class="fas fa-tags"></i> Buy Price
            <i
              v-if="sortConfig.field === 'buy_price'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('buy_bill_ref')" class="sortable">
            <i class="fas fa-file-alt"></i> Buy Bill Ref
            <i
              v-if="sortConfig.field === 'buy_bill_ref'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('notes')" class="sortable">
            <i class="fas fa-sticky-note"></i> Notes
            <i
              v-if="sortConfig.field === 'notes'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th><i class="fas fa-cog"></i> Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in sortedCars" :key="car.id">
          <td>{{ car.id }}</td>
          <td>{{ car.car_name }}</td>
          <td>
            <span
              v-if="car.color"
              class="badge color-badge"
              :style="{
                backgroundColor: car.hexa || '#000000',
                color: getTextColor(car.hexa || '#000000'),
              }"
            >
              {{ car.color }}
            </span>
            <span v-else>N/A</span>
          </td>
          <td>{{ car.vin || 'N/A' }}</td>
          <td>{{ car.price_cell ? '$' + car.price_cell.toLocaleString() : 'N/A' }}</td>
          <td v-if="isAdmin">{{ car.buy_price ? '$' + car.buy_price.toLocaleString() : 'N/A' }}</td>
          <td>{{ car.buy_bill_ref || 'N/A' }}</td>
          <td>{{ car.notes || 'N/A' }}</td>
          <td class="actions">
            <button
              @click="openAssignmentForm(car.id)"
              class="assign-btn"
              :disabled="isProcessing"
              :class="{ processing: isProcessing }"
            >
              <i class="fas fa-link"></i>
              <span>
                {{
                  isBatchSell
                    ? batchPrice
                      ? `Quick Assign ($${parseFloat(batchPrice).toLocaleString()})`
                      : 'Quick Assign'
                    : 'Assign'
                }}
                {{ isBatchSell && brokerInfo ? ` → ${brokerInfo.broker_name}` : '' }}
              </span>
              <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Assignment Form Dialog -->
    <div v-if="showAssignmentForm && !isBatchSell" class="dialog-overlay">
      <CarAssignmentForm
        :visible="showAssignmentForm"
        :carId="selectedCarId"
        :sellBillId="selectedSellBillId"
        @close="showAssignmentForm = false"
        @assign-success="handleAssignSuccess"
      />
    </div>
  </div>
</template>

<style scoped>
.unassigned-cars-table {
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

.badge {
  padding: 4px 10px;
  font-size: 0.95em;
  font-weight: 500;
  display: inline-block;
  border-radius: 4px;
  border: 1px solid #d1d5db;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  width: 100%;
  text-align: center;
}

.color-badge {
  background-color: #fef2f2;
  color: #dc2626;
}

.assign-btn {
  background-color: #10b981;
  color: white;
}

.assign-btn:hover:not(:disabled) {
  background-color: #059669;
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

/* Add filter styles */
.filters-section {
  background-color: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filters-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #1f2937;
  font-size: 1.1rem;
}

.reset-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #e5e7eb;
  border: none;
  border-radius: 4px;
  color: #4b5563;
  cursor: pointer;
  transition: all 0.2s;
}

.reset-btn:hover {
  background-color: #d1d5db;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 0.9rem;
}

.filter-group input {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.filter-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

@media (max-width: 768px) {
  .filters-grid {
    grid-template-columns: 1fr;
  }
}

/* Add sorting styles */
.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 24px; /* Space for sort icon */
}

.sortable i:last-child {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 0.75rem;
  opacity: 0.5;
}

.sortable:hover {
  background-color: #f1f5f9;
}

.sortable:hover i:last-child {
  opacity: 1;
}

/* Add batch pricing styles */
.batch-pricing-section {
  background-color: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.batch-pricing-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.batch-pricing-header h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #1f2937;
  font-size: 1.1rem;
}

.pricing-active-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  background-color: #10b981;
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  margin-left: 8px;
}

.pricing-active-indicator i {
  font-size: 0.7rem;
}

.batch-pricing-actions {
  display: flex;
  gap: 8px;
}

.toggle-pricing-btn {
  background-color: #10b981;
  color: white;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-pricing-btn:hover {
  background-color: #059669;
}

.toggle-pricing-btn.active {
  background-color: #059669;
}

.default-btn {
  background-color: #e5e7eb;
  color: #4b5563;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.default-btn:hover {
  background-color: #d1d5db;
}

.clear-btn {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-btn:hover {
  background-color: #f3e8e8;
}

.batch-pricing-form {
  margin-top: 16px;
}

.pricing-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #dbeafe;
  color: #1e40af;
  border-radius: 4px;
  margin-bottom: 16px;
  font-size: 0.9rem;
}

.pricing-info i {
  color: #3b82f6;
}

.pricing-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.pricing-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.pricing-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 0.9rem;
}

.pricing-group input {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.pricing-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.discharge-port-select {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.95rem;
  transition: all 0.2s;
  background-color: white;
}

.discharge-port-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.pricing-summary {
  margin-top: 16px;
  padding: 12px;
  background-color: #f3f4f6;
  border-radius: 4px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.summary-item .label {
  font-weight: 600;
}

.summary-item .value {
  font-weight: 500;
}

.broker-info {
  margin-top: 16px;
  padding: 12px;
  background-color: #f3f4f6;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.broker-info i {
  color: #3b82f6;
}

.broker-info span {
  display: flex;
  align-items: center;
  gap: 6px;
}

.broker-phone {
  font-size: 0.9rem;
  color: #6b7280;
  margin-left: auto;
}
</style>
