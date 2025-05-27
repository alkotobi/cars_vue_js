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
  loadingPort: '',
})
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)
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
          cs.date_loding,
          cs.path_documents,
          lp.loading_port,
          bd.amount as buy_price,
          cn.car_name,
          clr.color
        FROM cars_stock cs
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
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

    const matchLoadingPort =
      !filters.value.loadingPort ||
      (car.loading_port &&
        car.loading_port.toLowerCase().includes(filters.value.loadingPort.toLowerCase()))

    return matchCarName && matchColor && matchVin && matchLoadingPort
  })
}

// Reset all filters
const resetFilters = () => {
  filters.value = {
    carName: '',
    color: '',
    vin: '',
    loadingPort: '',
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

// Function to directly assign car for batch sell
const assignCarToBatchSell = async (carId) => {
  if (isProcessing.value) return
  isProcessing.value = true

  try {
    const result = await callApi({
      query: `
        UPDATE cars_stock
        SET id_sell = ?
        WHERE id = ?
      `,
      params: [selectedSellBillId.value, carId],
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
})
</script>

<template>
  <div class="unassigned-cars-table">
    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Loading unassigned cars...</span>
    </div>

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
          <th><i class="fas fa-hashtag"></i> ID</th>
          <th><i class="fas fa-car"></i> Car</th>
          <th><i class="fas fa-palette"></i> Color</th>
          <th><i class="fas fa-fingerprint"></i> VIN</th>
          <th><i class="fas fa-dollar-sign"></i> Sell Price</th>
          <th v-if="isAdmin"><i class="fas fa-tags"></i> Buy Price</th>
          <th><i class="fas fa-ship"></i> Loading Port</th>
          <th><i class="fas fa-calendar"></i> Loading Date</th>
          <th><i class="fas fa-cog"></i> Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in unassignedCars" :key="car.id">
          <td>{{ car.id }}</td>
          <td>{{ car.car_name }}</td>
          <td>{{ car.color }}</td>
          <td>{{ car.vin || 'N/A' }}</td>
          <td>{{ car.price_cell ? '$' + car.price_cell.toLocaleString() : 'N/A' }}</td>
          <td v-if="isAdmin">{{ car.buy_price ? '$' + car.buy_price.toLocaleString() : 'N/A' }}</td>
          <td>{{ car.loading_port || 'N/A' }}</td>
          <td>{{ car.date_loding || 'N/A' }}</td>
          <td class="actions">
            <button
              @click="openAssignmentForm(car.id)"
              class="assign-btn"
              :disabled="isProcessing"
              :class="{ processing: isProcessing }"
            >
              <i class="fas fa-link"></i>
              <span>{{ isBatchSell ? 'Quick Assign' : 'Assign' }}</span>
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
</style>
