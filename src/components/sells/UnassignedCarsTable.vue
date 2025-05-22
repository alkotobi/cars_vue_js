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

// Filter states
const filters = ref({
  carName: '',
  color: '',
  vin: '',
  loadingPort: ''
})

// Function to fetch all unassigned cars
const fetchUnassignedCars = async () => {
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
          cs.date_loding,
          cs.path_documents,
          lp.loading_port,
          bd.price_sell as buy_price,
          cn.car_name,
          clr.color
        FROM cars_stock cs
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        WHERE cs.id_sell IS NULL
        AND cs.hidden = 0
        ORDER BY cs.id DESC
      `,
      params: []
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
  unassignedCars.value = allUnassignedCars.value.filter(car => {
    const matchCarName = !filters.value.carName || 
      (car.car_name && car.car_name.toLowerCase().includes(filters.value.carName.toLowerCase()))
    
    const matchColor = !filters.value.color || 
      (car.color && car.color.toLowerCase().includes(filters.value.color.toLowerCase()))
    
    const matchVin = !filters.value.vin || 
      (car.vin && car.vin.toLowerCase().includes(filters.value.vin.toLowerCase()))
    
    const matchLoadingPort = !filters.value.loadingPort || 
      (car.loading_port && car.loading_port.toLowerCase().includes(filters.value.loadingPort.toLowerCase()))
    
    return matchCarName && matchColor && matchVin && matchLoadingPort
  })
}

// Reset all filters
const resetFilters = () => {
  filters.value = {
    carName: '',
    color: '',
    vin: '',
    loadingPort: ''
  }
  applyFilters()
}

// Watch for filter changes
const handleFilterChange = () => {
  applyFilters()
}

// Open assignment form for a car
const openAssignmentForm = (carId) => {
  if (!selectedSellBillId.value) {
    alert('Please select a sell bill first')
    return
  }
  
  selectedCarId.value = carId
  showAssignmentForm.value = true
}

// Handle successful assignment
const handleAssignSuccess = () => {
  showAssignmentForm.value = false
  fetchUnassignedCars()
  emit('refresh')
}

// Set the selected sell bill ID
const setSellBillId = (billId) => {
  selectedSellBillId.value = billId
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
  setSellBillId
})

onMounted(() => {
  fetchUnassignedCars()
})
</script>

<template>
  <div class="unassigned-cars-table">
    <h3>Unassigned Cars</h3>
    
    <!-- Filters -->
    <div class="filters">
      <div class="filter-row">
        <div class="filter-group">
          <label for="car-name-filter">Car Name:</label>
          <input 
            id="car-name-filter" 
            v-model="filters.carName" 
            @input="handleFilterChange" 
            placeholder="Filter by car name"
          />
        </div>
        
        <div class="filter-group">
          <label for="color-filter">Color:</label>
          <input 
            id="color-filter" 
            v-model="filters.color" 
            @input="handleFilterChange" 
            placeholder="Filter by color"
          />
        </div>
        
        <div class="filter-group">
          <label for="vin-filter">VIN:</label>
          <input 
            id="vin-filter" 
            v-model="filters.vin" 
            @input="handleFilterChange" 
            placeholder="Filter by VIN"
          />
        </div>
        
        <div class="filter-group">
          <label for="loading-port-filter">Loading Port:</label>
          <input 
            id="loading-port-filter" 
            v-model="filters.loadingPort" 
            @input="handleFilterChange" 
            placeholder="Filter by loading port"
          />
        </div>
        
        <button @click="resetFilters" class="reset-btn">Reset Filters</button>
      </div>
    </div>
    
    <div v-if="loading" class="loading">Loading cars...</div>
    
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else-if="unassignedCars.length === 0" class="no-data">
      No unassigned cars found matching your filters
    </div>
    
    <table v-else class="cars-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Car</th>
          <th>Color</th>
          <th>VIN</th>
          <th>Loading Port</th>
          <th>Buy Price</th>
          <th>Notes</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in unassignedCars" :key="car.id">
          <td>{{ car.id }}</td>
          <td>{{ car.car_name || 'N/A' }}</td>
          <td>{{ car.color || 'N/A' }}</td>
          <td>{{ car.vin || 'N/A' }}</td>
          <td>{{ car.loading_port || 'N/A' }}</td>
          <td>${{ car.buy_price || 'N/A' }}</td>
          <td>{{ car.notes || 'N/A' }}</td>
          <td>
            <button 
              @click="openAssignmentForm(car.id)" 
              class="assign-btn"
              :disabled="!selectedSellBillId"
            >
              Assign to Bill
            </button>
          </td>
        </tr>
      </tbody>
    </table>
    
    <!-- Car Assignment Form -->
    <CarAssignmentForm
      :visible="showAssignmentForm"
      :carId="selectedCarId"
      :sellBillId="selectedSellBillId"
      @close="showAssignmentForm = false"
      @assign-success="handleAssignSuccess"
    />
  </div>
</template>

<style scoped>
.unassigned-cars-table {
  margin-top: 30px;
}

h3 {
  margin-bottom: 15px;
  font-size: 1.2rem;
  color: #333;
}

.filters {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f9fafb;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  align-items: flex-end;
}

.filter-group {
  flex: 1;
  min-width: 200px;
}

.filter-group label {
  display: block;
  margin-bottom: 5px;
  font-size: 0.9rem;
  font-weight: 500;
  color: #4b5563;
}

.filter-group input {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.9rem;
}

.reset-btn {
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 0.9rem;
  height: 38px;
}

.reset-btn:hover {
  background-color: #4b5563;
}

.loading, .error, .no-data {
  padding: 20px;
  text-align: center;
  background-color: #f9f9f9;
  border-radius: 4px;
  color: #666;
}

.error {
  color: #ef4444;
  background-color: #fee2e2;
}

.cars-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.cars-table th,
.cars-table td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.cars-table th {
  background-color: #f3f4f6;
  font-weight: 600;
}

.cars-table tbody tr:hover {
  background-color: #f9fafb;
}

.assign-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 5px 10px;
  cursor: pointer;
  font-size: 0.8rem;
}

.assign-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.assign-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}
</style>