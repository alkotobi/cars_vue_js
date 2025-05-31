<script setup>
import { ref, onMounted, defineProps, computed, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  show: {
    type: Boolean,
    default: true,
  },
})

const { callApi, getFileUrl } = useApi()
const cars = ref([])
const loading = ref(true)
const error = ref(null)

// Filter states
const filters = ref({
  carName: '',
  client: '',
  dateFrom: '',
  dateTo: '',
  vin: '',
})

// Sort states
const sortColumn = ref('date_send_documents')
const sortDirection = ref('DESC')

// Watch filters and sort changes
watch(
  [filters, sortColumn, sortDirection],
  () => {
    fetchFinishedCars()
  },
  { deep: true },
)

const handleSort = (column) => {
  if (sortColumn.value === column) {
    // Toggle direction if same column
    sortDirection.value = sortDirection.value === 'ASC' ? 'DESC' : 'ASC'
  } else {
    // New column, set default direction
    sortColumn.value = column
    sortDirection.value = 'ASC'
  }
}

const getSortIcon = (column) => {
  if (sortColumn.value !== column) return 'fas fa-sort'
  return sortDirection.value === 'ASC' ? 'fas fa-sort-up' : 'fas fa-sort-down'
}

const fetchFinishedCars = async () => {
  loading.value = true
  error.value = null

  try {
    let query = `
      SELECT 
        cs.id,
        cs.vin,
        cs.price_cell,
        cs.date_loding,
        cs.date_sell,
        cs.notes,
        cs.freight,
        cs.path_documents,
        cs.date_send_documents,
        c.name as client_name,
        cn.car_name,
        clr.color,
        lp.loading_port,
        dp.discharge_port
      FROM cars_stock cs
      LEFT JOIN clients c ON cs.id_client = c.id
      LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
      LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
      LEFT JOIN colors clr ON bd.id_color = clr.id
      LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
      LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
      WHERE cs.hidden = 0   
      AND cs.date_send_documents IS NOT NULL
    `

    const params = []

    // Apply filters
    if (filters.value.carName) {
      query += ` AND cn.car_name LIKE ?`
      params.push(`%${filters.value.carName}%`)
    }

    if (filters.value.client) {
      query += ` AND c.name LIKE ?`
      params.push(`%${filters.value.client}%`)
    }

    if (filters.value.vin) {
      query += ` AND cs.vin LIKE ?`
      params.push(`%${filters.value.vin}%`)
    }

    if (filters.value.dateFrom) {
      query += ` AND cs.date_send_documents >= ?`
      params.push(filters.value.dateFrom)
    }

    if (filters.value.dateTo) {
      query += ` AND cs.date_send_documents <= ?`
      params.push(filters.value.dateTo)
    }

    // Apply column sorting
    const sortMap = {
      id: 'cs.id',
      car_name: 'cn.car_name',
      client_name: 'c.name',
      date_send_documents: 'cs.date_send_documents',
      vin: 'cs.vin',
      color: 'clr.color',
      loading_port: 'lp.loading_port',
      discharge_port: 'dp.discharge_port',
    }

    const sortColumnSQL = sortMap[sortColumn.value] || 'cs.date_send_documents'
    query += ` ORDER BY ${sortColumnSQL} ${sortDirection.value}`

    const result = await callApi({ query, params })

    if (result.success) {
      cars.value = result.data
    } else {
      throw new Error(result.error || 'Failed to fetch finished cars')
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const clearFilters = () => {
  filters.value = {
    carName: '',
    client: '',
    dateFrom: '',
    dateTo: '',
    vin: '',
  }
}

onMounted(() => {
  if (props.show) {
    fetchFinishedCars()
  }
})

// Expose the fetch method
defineExpose({
  fetchFinishedCars,
})
</script>

<template>
  <div class="finished-cars-table">
    <!-- Filters Section -->
    <div class="filters-section">
      <div class="filters-grid">
        <div class="filter-group">
          <label for="car-name">Car Name</label>
          <input
            id="car-name"
            v-model="filters.carName"
            type="text"
            placeholder="Search by car name"
          />
        </div>
        <div class="filter-group">
          <label for="client">Client</label>
          <input
            id="client"
            v-model="filters.client"
            type="text"
            placeholder="Search by client name"
          />
        </div>
        <div class="filter-group">
          <label for="vin">VIN</label>
          <input id="vin" v-model="filters.vin" type="text" placeholder="Search by VIN" />
        </div>
        <div class="filter-group">
          <label for="date-from">From Date</label>
          <input id="date-from" v-model="filters.dateFrom" type="date" />
        </div>
        <div class="filter-group">
          <label for="date-to">To Date</label>
          <input id="date-to" v-model="filters.dateTo" type="date" />
        </div>
      </div>
      <button class="clear-filters-btn" @click="clearFilters">
        <i class="fas fa-times"></i>
        Clear Filters
      </button>
    </div>

    <!-- Results Section -->
    <div v-if="loading" class="loading">
      <i class="fas fa-spinner fa-spin"></i>
      Loading...
    </div>
    <div v-else-if="error" class="error">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>
    <div v-else-if="cars.length === 0" class="empty-state">
      <i class="fas fa-check-circle fa-2x"></i>
      <p>No finished orders found</p>
    </div>
    <div v-else class="table-container">
      <table class="cars-table">
        <thead>
          <tr>
            <th @click="handleSort('id')" class="sortable">
              ID
              <i :class="getSortIcon('id')"></i>
            </th>
            <th @click="handleSort('car_name')" class="sortable">
              Car
              <i :class="getSortIcon('car_name')"></i>
            </th>
            <th @click="handleSort('color')" class="sortable">
              Color
              <i :class="getSortIcon('color')"></i>
            </th>
            <th @click="handleSort('vin')" class="sortable">
              VIN
              <i :class="getSortIcon('vin')"></i>
            </th>
            <th @click="handleSort('loading_port')" class="sortable">
              Loading Port
              <i :class="getSortIcon('loading_port')"></i>
            </th>
            <th @click="handleSort('discharge_port')" class="sortable">
              Discharge Port
              <i :class="getSortIcon('discharge_port')"></i>
            </th>
            <th @click="handleSort('client_name')" class="sortable">
              Client
              <i :class="getSortIcon('client_name')"></i>
            </th>
            <th @click="handleSort('date_send_documents')" class="sortable">
              Documents Sent Date
              <i :class="getSortIcon('date_send_documents')"></i>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="car in cars" :key="car.id">
            <td>#{{ car.id }}</td>
            <td>{{ car.car_name }}</td>
            <td>{{ car.color }}</td>
            <td>{{ car.vin || '-' }}</td>
            <td>{{ car.loading_port || '-' }}</td>
            <td>{{ car.discharge_port || '-' }}</td>
            <td>{{ car.client_name || '-' }}</td>
            <td>{{ new Date(car.date_send_documents).toLocaleDateString() }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.finished-cars-table {
  width: 100%;
  overflow-x: auto;
  position: relative;
}

/* Filters Section Styles */
.filters-section {
  background-color: #f8fafc;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid #e5e7eb;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.filter-group input,
.filter-group select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
  color: #1f2937;
  background-color: white;
}

.filter-group input:focus,
.filter-group select:focus {
  outline: none;
  border-color: #3b82f6;
  ring: 2px solid rgba(59, 130, 246, 0.5);
}

.clear-filters-btn {
  padding: 8px 16px;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background-color 0.2s;
}

.clear-filters-btn:hover {
  background-color: #dc2626;
}

/* Existing styles... */
.loading,
.error,
.empty-state {
  padding: 40px;
  text-align: center;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  background-color: #f9fafb;
  border-radius: 8px;
  margin: 20px 0;
}

.loading i {
  font-size: 2em;
  color: #3b82f6;
}

.error {
  color: #ef4444;
}

.error i {
  font-size: 2em;
}

.empty-state i {
  color: #10b981;
}

.empty-state p {
  margin: 0;
  font-size: 1.1em;
}

.table-container {
  position: relative;
  max-height: calc(100vh - 250px);
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.cars-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.cars-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #f8f9fa;
}

.cars-table th {
  padding: 12px;
  text-align: left;
  border-bottom: 2px solid #e5e7eb;
  font-weight: 600;
  white-space: nowrap;
  color: #374151;
}

.cars-table td {
  padding: 12px;
  border-bottom: 1px solid #e5e7eb;
  background-color: white;
}

.cars-table tbody tr:hover {
  background-color: #f9fafb;
}

@media (max-width: 1024px) {
  .table-container {
    max-height: none;
    overflow-x: auto;
  }

  .cars-table {
    min-width: 1000px;
  }

  .filters-grid {
    grid-template-columns: 1fr;
  }
}

/* Add new styles for sortable columns */
.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 24px !important;
}

.sortable i {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
  font-size: 0.875rem;
}

.sortable:hover {
  background-color: #f3f4f6;
}

.sortable:hover i {
  color: #6b7280;
}

/* When column is actively sorted */
.sortable i.fa-sort-up,
.sortable i.fa-sort-down {
  color: #3b82f6;
}
</style>
