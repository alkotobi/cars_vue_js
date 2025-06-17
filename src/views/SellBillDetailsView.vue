<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const route = useRoute()
const router = useRouter()
const { callApi, getFileUrl } = useApi()

const billId = route.params.id
const bill = ref(null)
const cars = ref([])
const loading = ref(true)
const error = ref(null)

// Helper to check if a file is an image
function isImageFile(path) {
  if (!path) return false
  const extension = path.split('.').pop().toLowerCase()
  return ['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(extension)
}

function handleImageClick(path) {
  if (path) {
    window.open(getFileUrl(path), '_blank')
  }
}

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
      query: `SELECT cs.*, cn.car_name, dp.discharge_port, cl.name as client_name, cl.id_copy_path as client_id_copy
              FROM cars_stock cs
              LEFT JOIN cars_names cn ON cs.id_buy_details = cn.id
              LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
              LEFT JOIN clients cl ON cs.id_client = cl.id
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

const amountUsd = computed(() => {
  if (!cars.value.length) return 'N/A'
  let sum = 0
  for (const car of cars.value) {
    if (car.price_cell != null && car.freight != null) {
      sum += Number(car.price_cell) + Number(car.freight)
    }
  }
  return sum ? sum : 'N/A'
})

const amountDa = computed(() => {
  if (!cars.value.length) return 'N/A'
  let sum = 0
  for (const car of cars.value) {
    if (car.price_cell != null && car.freight != null && car.rate != null) {
      sum += (Number(car.price_cell) + Number(car.freight)) * Number(car.rate)
    }
  }
  return sum ? sum : 'N/A'
})

onMounted(fetchData)
</script>

<template>
  <div class="sell-bill-details-view">
    <button v-if="false" class="back-btn" @click="router.back()">&larr; Back</button>
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
        <div><strong>Amount USD:</strong> {{ amountUsd }}</div>
        <div><strong>Amount DA:</strong> {{ amountDa }}</div>
        <div v-if="false">
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
            <th>Client</th>
            <th>Client ID</th>
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
            <td>{{ car.client_name || 'N/A' }}</td>
            <td class="id-document-cell">
              <div
                v-if="car.client_id_copy && isImageFile(car.client_id_copy)"
                class="image-preview"
                @click="handleImageClick(car.client_id_copy)"
              >
                <img :src="getFileUrl(car.client_id_copy)" alt="Client ID" />
              </div>
              <a
                v-else-if="car.client_id_copy"
                :href="getFileUrl(car.client_id_copy)"
                target="_blank"
                class="document-link"
              >
                <i class="fas fa-file-download"></i>
                View Document
              </a>
              <span v-else class="no-document">
                <i class="fas fa-times-circle"></i>
                No ID
              </span>
            </td>
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
.id-document-cell {
  width: 120px;
  text-align: center;
}
.image-preview {
  width: 60px;
  height: 40px;
  margin: 0 auto;
  cursor: pointer;
  overflow: hidden;
  border-radius: 4px;
  border: 1px solid #d1d5db;
  transition: transform 0.2s;
}
.image-preview:hover {
  transform: scale(1.05);
}
.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.document-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.9em;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  transition: color 0.2s;
}
.document-link:hover {
  color: #2563eb;
  text-decoration: underline;
}
.no-document {
  color: #6b7280;
  font-size: 0.9em;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.no-document i {
  color: #ef4444;
}
</style>
