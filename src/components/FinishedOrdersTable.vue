<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi } = useApi()
const finishedOrders = ref([])
const loading = ref(false)

const filters = ref({
  vin: '',
  client: '',
  date: '',
})

const sortConfig = ref({
  key: 'date_send_documents',
  direction: 'desc',
})

const fetchFinishedOrders = async () => {
  loading.value = true
  try {
    const result = await callApi({
      query: `SELECT cs.*, c.name as client_name FROM cars_stock cs LEFT JOIN clients c ON cs.id_client = c.id WHERE cs.date_send_documents IS NOT NULL ORDER BY cs.date_send_documents DESC`,
      params: [],
    })
    if (result.success) {
      finishedOrders.value = result.data
    }
  } finally {
    loading.value = false
  }
}

onMounted(fetchFinishedOrders)

const filteredAndSorted = computed(() => {
  let data = finishedOrders.value
  if (filters.value.vin) {
    data = data.filter((car) =>
      (car.vin || '').toLowerCase().includes(filters.value.vin.toLowerCase()),
    )
  }
  if (filters.value.client) {
    data = data.filter((car) =>
      (car.client_name || '').toLowerCase().includes(filters.value.client.toLowerCase()),
    )
  }
  if (filters.value.date) {
    data = data.filter((car) => (car.date_send_documents || '').startsWith(filters.value.date))
  }
  if (sortConfig.value.key) {
    data = [...data].sort((a, b) => {
      let aVal = a[sortConfig.value.key]
      let bVal = b[sortConfig.value.key]
      if (sortConfig.value.key === 'price_cell') {
        aVal = parseFloat(aVal) || 0
        bVal = parseFloat(bVal) || 0
      }
      if (aVal == null) return 1
      if (bVal == null) return -1
      if (sortConfig.value.direction === 'asc') {
        return aVal > bVal ? 1 : aVal < bVal ? -1 : 0
      } else {
        return aVal < bVal ? 1 : aVal > bVal ? -1 : 0
      }
    })
  }
  return data
})

function toggleSort(key) {
  if (sortConfig.value.key === key) {
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    sortConfig.value.key = key
    sortConfig.value.direction = 'asc'
  }
}
</script>

<template>
  <div>
    <button @click="$emit('close')" class="btn return-btn">
      <i class="fas fa-arrow-left"></i> Return
    </button>
    <h2>Finished Orders</h2>
    <div class="filters-row">
      <input v-model="filters.vin" placeholder="Filter by VIN" />
      <input v-model="filters.client" placeholder="Filter by Client Name" />
      <input v-model="filters.date" type="date" placeholder="Filter by Date Sent" />
    </div>
    <div v-if="loading" class="loading">Loading...</div>
    <table v-else class="finished-orders-table">
      <thead>
        <tr>
          <th @click="toggleSort('id')">
            ID
            <span v-if="sortConfig.key === 'id'">{{
              sortConfig.direction === 'asc' ? '▲' : '▼'
            }}</span>
          </th>
          <th @click="toggleSort('vin')">
            VIN
            <span v-if="sortConfig.key === 'vin'">{{
              sortConfig.direction === 'asc' ? '▲' : '▼'
            }}</span>
          </th>
          <th @click="toggleSort('client_name')">
            Client
            <span v-if="sortConfig.key === 'client_name'">{{
              sortConfig.direction === 'asc' ? '▲' : '▼'
            }}</span>
          </th>
          <th @click="toggleSort('date_send_documents')">
            Date Sent Documents
            <span v-if="sortConfig.key === 'date_send_documents'">{{
              sortConfig.direction === 'asc' ? '▲' : '▼'
            }}</span>
          </th>
          <th @click="toggleSort('price_cell')">
            Price
            <span v-if="sortConfig.key === 'price_cell'">{{
              sortConfig.direction === 'asc' ? '▲' : '▼'
            }}</span>
          </th>
          <th>Notes</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in filteredAndSorted" :key="car.id">
          <td>{{ car.id }}</td>
          <td>{{ car.vin }}</td>
          <td>{{ car.client_name }}</td>
          <td>{{ car.date_send_documents }}</td>
          <td>{{ car.price_cell }}</td>
          <td>{{ car.notes }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.finished-orders-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 1.5rem;
}
.finished-orders-table th,
.finished-orders-table td {
  border: 1px solid #e0e0e0;
  padding: 0.75rem 1rem;
  text-align: left;
  cursor: pointer;
}
.finished-orders-table th {
  background: #f8f9fa;
  user-select: none;
}
.filters-row {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}
.filters-row input {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  min-width: 160px;
}
.return-btn {
  margin-bottom: 1rem;
  background: #eee;
  color: #333;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
}
.return-btn:hover {
  background: #ddd;
}
</style>
