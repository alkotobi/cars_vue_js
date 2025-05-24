<script setup>
import { ref, watch, defineProps, defineEmits,onMounted ,computed } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  sellBillId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits(['refresh'])
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)

const { callApi } = useApi()
const cars = ref([])
const loading = ref(false)
const error = ref(null)

// Function to unassign a car from the sell bill
const unassignCar = async (carId) => {
  if (!confirm('Are you sure you want to unassign this car from the sell bill?')) {
    return
  }
  
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
      params: [carId]
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
  }
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
          cs.date_sell,
          cs.price_cell,
          cs.freight,
          cs.date_loding,
          cs.date_send_documents,
          c.name as client_name,
          lp.loading_port,
          dp.discharge_port,
          bd.price_sell as buy_price,
          cn.car_name,
          clr.color
        FROM cars_stock cs
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        WHERE cs.id_sell = ?
        ORDER BY cs.id DESC
      `,
      params: [billId]
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
watch(() => props.sellBillId, (newId) => {
  if (newId) {
    fetchCarsByBillId(newId)
  } else {
    cars.value = []
  }
}, { immediate: true })

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
  fetchCarsByBillId
})
</script>

<template>
  <div class="sell-bill-cars-table">
    <h3>Cars in Selected Bill</h3>
    
    <div v-if="!props.sellBillId" class="no-selection">
      Please select a sell bill to view associated cars
    </div>
    
    <div v-else-if="loading" class="loading">Loading cars...</div>
    
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else-if="cars.length === 0" class="no-data">
      No cars found for this sell bill
    </div>
    
    <table v-else class="cars-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Car</th>
          <th>Color</th>
          <th>VIN</th>
          <th>Client</th>
          <th>Loading Port</th>
          <th>Discharge Port</th>
          <th>Buy Price</th>
          <th>Sell Price</th>
          <th>Freight</th>
          <th>Profit</th>
          <th>Notes</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in cars" :key="car.id">
          <td>{{ car.id }}</td>
          <td>{{ car.car_name || 'N/A' }}</td>
          <td>{{ car.color || 'N/A' }}</td>
          <td>{{ car.vin || 'N/A' }}</td>
          <td>{{ car.client_name || 'N/A' }}</td>
          <td>{{ car.loading_port || 'N/A' }}</td>
          <td>{{ car.discharge_port || 'N/A' }}</td>
          <td>${{ car.buy_price || 'N/A' }}</td>
          <td>${{ car.price_cell || 'N/A' }}</td>
          <td>${{ car.freight || '0.00' }}</td>
          <td>${{ calculateProfit(car.price_cell, car.buy_price, car.freight) }}</td>
          <td>{{ car.notes || 'N/A' }}</td>
          <td>
            <button @click="unassignCar(car.id)" class="unassign-btn">Unassign</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.sell-bill-cars-table {
  margin-top: 30px;
}

h3 {
  margin-bottom: 15px;
  font-size: 1.2rem;
  color: #333;
}

.no-selection, .loading, .error, .no-data {
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

.unassign-btn {
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 5px 10px;
  cursor: pointer;
  font-size: 0.8rem;
}

.unassign-btn:hover {
  background-color: #dc2626;
}
</style>