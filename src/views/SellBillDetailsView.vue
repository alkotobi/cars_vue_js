<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const route = useRoute()
const router = useRouter()
const { callApi } = useApi()

const billId = route.params.id
const bill = ref(null)
const cars = ref([])
const loading = ref(true)
const error = ref(null)

const fetchData = async () => {
  loading.value = true
  error.value = null
  try {
    console.log('[DEBUG] billId:', billId)
    // Fetch sell bill with client name using id_client
    const billResult = await callApi({
      query: `SELECT sb.*, c.name as client_name FROM sell_bill sb LEFT JOIN clients c ON sb.id_broker = c.id WHERE sb.id = ?`,
      params: [billId],
      requiresAuth: false,
    })
    console.log('[DEBUG] billResult:', billResult)
    if (billResult.success && billResult.data.length) {
      bill.value = billResult.data[0]
      console.log('[DEBUG] bill row:', bill.value)
    } else {
      error.value = 'Sell bill not found.'
      loading.value = false
      return
    }
    // Fetch cars in this sell bill with car name, freight, price CFR DA, discharge port
    const carsResult = await callApi({
      query: `SELECT cs.*, cn.car_name, dp.discharge_port
              FROM cars_stock cs
              LEFT JOIN cars_names cn ON cs.id_buy_details = cn.id
              LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
              WHERE cs.id_sell = ?`,
      params: [billId],
      requiresAuth: false,
    })
    console.log('[DEBUG] carsResult:', carsResult)
    if (carsResult.success) {
      cars.value = carsResult.data
    } else {
      error.value = carsResult.error || 'Failed to fetch cars.'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred.'
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="sell-bill-details-view">
    <button class="back-btn" @click="router.back()">&larr; Back</button>
    <h2>Sell Bill Details</h2>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="bill">
      <div class="bill-info">
        <div><strong>ID:</strong> {{ bill.id }}</div>
        <div><strong>Reference:</strong> {{ bill.bill_ref || 'N/A' }}</div>
        <div>
          <strong>Date:</strong>
          {{ bill.date_sell ? new Date(bill.date_sell).toLocaleDateString() : 'N/A' }}
        </div>
        <div>
          <strong>Amount:</strong>
          {{ bill.amount !== null && bill.amount !== undefined ? bill.amount : 'N/A' }}
        </div>
        <div>
          <strong>Received:</strong>
          {{ bill.received !== null && bill.received !== undefined ? bill.received : 'N/A' }}
        </div>
        <div><strong>Broker:</strong> {{ bill.client_name || 'N/A' }}</div>
      </div>
      <h3>Cars in this Sell Bill</h3>
      <div v-if="cars.length === 0" class="no-cars">No cars found for this sell bill.</div>
      <table v-else class="cars-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Car Name</th>
            <th>Freight</th>
            <th>Price</th>
            <th>Rate</th>
            <th>Price CFR DA</th>
            <th>Discharge Port</th>
            <th>Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="car in cars" :key="car.id">
            <td>{{ car.id }}</td>
            <td>{{ car.car_name || 'N/A' }}</td>
            <td>{{ car.freight !== null && car.freight !== undefined ? car.freight : 'N/A' }}</td>
            <td>
              {{ car.price_cell !== null && car.price_cell !== undefined ? car.price_cell : 'N/A' }}
            </td>
            <td>{{ car.rate !== null && car.rate !== undefined ? car.rate : 'N/A' }}</td>
            <td>
              {{
                car.price_cell !== null &&
                car.freight !== null &&
                car.rate !== null &&
                car.price_cell !== undefined &&
                car.freight !== undefined &&
                car.rate !== undefined
                  ? (Number(car.price_cell) + Number(car.freight)) * Number(car.rate)
                  : 'N/A'
              }}
            </td>
            <td>{{ car.discharge_port || 'N/A' }}</td>
            <td>{{ car.notes || 'N/A' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.sell-bill-details-view {
  max-width: 700px;
  margin: 0 auto;
  padding: 24px 12px;
}
.back-btn {
  margin-bottom: 16px;
  background: #f3f4f6;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 1em;
}
.loading,
.error,
.no-cars {
  margin: 24px 0;
  text-align: center;
  color: #888;
}
.bill-info {
  background: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 24px;
}
.cars-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 12px;
}
.cars-table th,
.cars-table td {
  border: 1px solid #e5e7eb;
  padding: 8px;
  text-align: left;
}
.cars-table th {
  background: #f3f4f6;
}
@media (max-width: 600px) {
  .bill-info {
    grid-template-columns: 1fr;
  }
  .cars-table th,
  .cars-table td {
    font-size: 0.95em;
    padding: 6px;
  }
}
</style>
