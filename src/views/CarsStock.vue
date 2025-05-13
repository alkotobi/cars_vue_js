<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi } = useApi()
const cars = ref([])
const loading = ref(true)
const error = ref(null)

const fetchCarsStock = async () => {
  loading.value = true
  error.value = null
  
  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.price_cell,
          cs.date_loding,
          cs.date_sell,
          cs.deposit,
          cs.balance,
          cs.notes,
          c.name as client_name,
          cn.car_name,
          clr.color,
          lp.loading_port,
          dp.discharge_port,
          bd.price_sell as buy_price
        FROM cars_stock cs
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        WHERE cs.hidden = 0
        ORDER BY cs.id DESC
      `,
      params: []
    })
    
    if (result.success) {
      cars.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch cars stock'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

onMounted(fetchCarsStock)
</script>

<template>
  <div class="cars-stock">
    <div class="header">
      <h2>Cars Stock Management</h2>
    </div>
    
    <div class="content">
      <div v-if="loading" class="loading">Loading...</div>
      <div v-else-if="error" class="error">{{ error }}</div>
      <div v-else-if="cars.length === 0" class="empty-state">No cars in stock</div>
      
      <table v-else class="cars-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Car</th>
            <th>Color</th>
            <th>VIN</th>
            <th>Loading Port</th>
            <th>Discharge Port</th>
            <th>Buy Price</th>
            <th>Sell Price</th>
            <th>Loading Date</th>
            <th>Status</th>
            <th>Client</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="car in cars" :key="car.id">
            <td>{{ car.id }}</td>
            <td>{{ car.car_name || 'N/A' }}</td>
            <td>{{ car.color || 'N/A' }}</td>
            <td>{{ car.vin || 'N/A' }}</td>
            <td>{{ car.loading_port || 'N/A' }}</td>
            <td>{{ car.discharge_port || 'N/A' }}</td>
            <td>{{ car.buy_price ? '$' + car.buy_price : 'N/A' }}</td>
            <td>{{ car.price_cell ? '$' + car.price_cell : 'N/A' }}</td>
            <td>{{ car.date_loding ? new Date(car.date_loding).toLocaleDateString() : 'N/A' }}</td>
            <td :class="car.date_sell ? 'status-sold' : 'status-available'">
              {{ car.date_sell ? 'Sold' : 'Available' }}
            </td>
            <td>{{ car.client_name || 'N/A' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.cars-stock {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.loading, .error, .empty-state {
  padding: 20px;
  text-align: center;
}

.error {
  color: #ef4444;
}

.cars-table {
  width: 100%;
  border-collapse: collapse;
}

.cars-table th,
.cars-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.cars-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.cars-table tbody tr:hover {
  background-color: #f5f5f5;
}

.status-available {
  color: #10b981;
  font-weight: 500;
}

.status-sold {
  color: #ef4444;
  font-weight: 500;
}
</style>