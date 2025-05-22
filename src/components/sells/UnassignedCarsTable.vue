<script setup>
import { ref, onMounted, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'
import CarAssignmentForm from './CarAssignmentForm.vue'

const emit = defineEmits(['refresh', 'assign-car'])

const { callApi } = useApi()
const unassignedCars = ref([])
const loading = ref(true)
const error = ref(null)
const selectedSellBillId = ref(null)
const showAssignmentForm = ref(false)
const selectedCarId = ref(null)

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
      unassignedCars.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch unassigned cars'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
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
    
    <div v-if="loading" class="loading">Loading cars...</div>
    
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else-if="unassignedCars.length === 0" class="no-data">
      No unassigned cars found
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