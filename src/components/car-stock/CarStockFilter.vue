<script setup>
import { ref, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'

const emit = defineEmits(['filter'])

const { callApi } = useApi()

// Reference data for dropdowns
const carNames = ref([])
const colors = ref([])
const loadingPorts = ref([])
const dischargePorts = ref([])
const clients = ref([])
const warehouses = ref([])

// Basic filter
const basicFilter = ref('')

// Advanced filter state
const showAdvancedFilter = ref(false)
const advancedFilters = ref({
  id: '',
  car_name: '',
  color: '',
  vin: '',
  loading_port: '',
  discharge_port: '',
  freight_min: '',
  freight_max: '',
  price_min: '',
  price_max: '',
  loading_date_from: '',
  loading_date_to: '',
  status: '',
  client: '',
  warehouse: ''
})

// Fetch reference data for filter dropdowns
const fetchReferenceData = async () => {
  try {
    // Fetch car names
    const carNamesResult = await callApi({
      query: 'SELECT id, car_name FROM cars_names ORDER BY car_name',
      params: []
    })
    if (carNamesResult.success) {
      carNames.value = carNamesResult.data
    }

    // Fetch colors
    const colorsResult = await callApi({
      query: 'SELECT id, color FROM colors ORDER BY color',
      params: []
    })
    if (colorsResult.success) {
      colors.value = colorsResult.data
    }

    // Fetch loading ports
    const loadingPortsResult = await callApi({
      query: 'SELECT id, loading_port FROM loading_ports ORDER BY loading_port',
      params: []
    })
    if (loadingPortsResult.success) {
      loadingPorts.value = loadingPortsResult.data
    }

    // Fetch discharge ports
    const dischargePortsResult = await callApi({
      query: 'SELECT id, discharge_port FROM discharge_ports ORDER BY discharge_port',
      params: []
    })
    if (dischargePortsResult.success) {
      dischargePorts.value = dischargePortsResult.data
    }

    // Fetch clients
    const clientsResult = await callApi({
      query: 'SELECT id, name FROM clients ORDER BY name',
      params: []
    })
    if (clientsResult.success) {
      clients.value = clientsResult.data
    }

    // Fetch warehouses
    const warehousesResult = await callApi({
      query: 'SELECT id, name FROM warehouses ORDER BY name',
      params: []
    })
    if (warehousesResult.success) {
      warehouses.value = warehousesResult.data
    }
  } catch (error) {
    console.error('Error fetching reference data:', error)
  }
}

// Apply basic filter
const applyBasicFilter = () => {
  emit('filter', { basic: basicFilter.value, advanced: null })
}

// Apply advanced filters
const applyAdvancedFilters = () => {
  emit('filter', { basic: null, advanced: advancedFilters.value })
}

// Reset filters
const resetFilters = () => {
  basicFilter.value = ''
  Object.keys(advancedFilters.value).forEach(key => {
    advancedFilters.value[key] = ''
  })
  emit('filter', { basic: '', advanced: null })
}

// Toggle advanced filter visibility
const toggleAdvancedFilter = () => {
  showAdvancedFilter.value = !showAdvancedFilter.value
}

// Fetch reference data on component mount
fetchReferenceData()
</script>

<template>
  <div class="car-stock-filter">
    <!-- Basic Filter -->
    <div class="basic-filter">
      <div class="filter-row">
        <div class="filter-input">
          <input 
            type="text" 
            v-model="basicFilter" 
            placeholder="Search by ID, Car, Color, VIN, Ports, Client..." 
            @input="applyBasicFilter"
          />
        </div>
        <div class="filter-actions">
          <button @click="toggleAdvancedFilter" class="toggle-btn">
            {{ showAdvancedFilter ? 'Hide Advanced' : 'Show Advanced' }}
          </button>
          <button @click="resetFilters" class="reset-btn">Reset</button>
        </div>
      </div>
    </div>

    <!-- Advanced Filters -->
    <div v-if="showAdvancedFilter" class="advanced-filters">
      <div class="filter-grid">
        <!-- ID Filter -->
        <div class="filter-field">
          <label for="id-filter">ID</label>
          <input 
            id="id-filter" 
            type="text" 
            v-model="advancedFilters.id" 
            placeholder="Car ID"
          />
        </div>

        <!-- Car Name Filter -->
        <div class="filter-field">
          <label for="car-name-filter">Car</label>
          <select 
            id="car-name-filter" 
            v-model="advancedFilters.car_name"
          >
            <option value="">All Cars</option>
            <option 
              v-for="car in carNames" 
              :key="car.id" 
              :value="car.car_name"
            >
              {{ car.car_name }}
            </option>
          </select>
        </div>

        <!-- Color Filter -->
        <div class="filter-field">
          <label for="color-filter">Color</label>
          <select 
            id="color-filter" 
            v-model="advancedFilters.color"
          >
            <option value="">All Colors</option>
            <option 
              v-for="color in colors" 
              :key="color.id" 
              :value="color.color"
            >
              {{ color.color }}
            </option>
          </select>
        </div>

        <!-- VIN Filter -->
        <div class="filter-field">
          <label for="vin-filter">VIN</label>
          <input 
            id="vin-filter" 
            type="text" 
            v-model="advancedFilters.vin" 
            placeholder="VIN Number"
          />
        </div>

        <!-- Loading Port Filter -->
        <div class="filter-field">
          <label for="loading-port-filter">Loading Port</label>
          <select 
            id="loading-port-filter" 
            v-model="advancedFilters.loading_port"
          >
            <option value="">All Loading Ports</option>
            <option 
              v-for="port in loadingPorts" 
              :key="port.id" 
              :value="port.loading_port"
            >
              {{ port.loading_port }}
            </option>
          </select>
        </div>

        <!-- Discharge Port Filter -->
        <div class="filter-field">
          <label for="discharge-port-filter">Discharge Port</label>
          <select 
            id="discharge-port-filter" 
            v-model="advancedFilters.discharge_port"
          >
            <option value="">All Discharge Ports</option>
            <option 
              v-for="port in dischargePorts" 
              :key="port.id" 
              :value="port.discharge_port"
            >
              {{ port.discharge_port }}
            </option>
          </select>
        </div>

        <!-- Freight Range Filter -->
        <div class="filter-field">
          <label>Freight Range</label>
          <div class="range-inputs">
            <input 
              type="number" 
              v-model="advancedFilters.freight_min" 
              placeholder="Min"
            />
            <span>to</span>
            <input 
              type="number" 
              v-model="advancedFilters.freight_max" 
              placeholder="Max"
            />
          </div>
        </div>

        <!-- Price Range Filter -->
        <div class="filter-field">
          <label>Price Range</label>
          <div class="range-inputs">
            <input 
              type="number" 
              v-model="advancedFilters.price_min" 
              placeholder="Min"
            />
            <span>to</span>
            <input 
              type="number" 
              v-model="advancedFilters.price_max" 
              placeholder="Max"
            />
          </div>
        </div>

        <!-- Loading Date Range Filter -->
        <div class="filter-field">
          <label>Loading Date Range</label>
          <div class="date-inputs">
            <input 
              type="date" 
              v-model="advancedFilters.loading_date_from" 
            />
            <span>to</span>
            <input 
              type="date" 
              v-model="advancedFilters.loading_date_to" 
            />
          </div>
        </div>

        <!-- Status Filter -->
        <div class="filter-field">
          <label for="status-filter">Status</label>
          <select 
            id="status-filter" 
            v-model="advancedFilters.status"
          >
            <option value="">All Status</option>
            <option value="available">Available</option>
            <option value="sold">Sold</option>
          </select>
        </div>

        <!-- Client Filter -->
        <div class="filter-field">
          <label for="client-filter">Client</label>
          <select 
            id="client-filter" 
            v-model="advancedFilters.client"
          >
            <option value="">All Clients</option>
            <option 
              v-for="client in clients" 
              :key="client.id" 
              :value="client.name"
            >
              {{ client.name }}
            </option>
          </select>
        </div>

        <!-- Warehouse Filter -->
        <div class="filter-field">
          <label for="warehouse-filter">Warehouse</label>
          <select 
            id="warehouse-filter" 
            v-model="advancedFilters.warehouse"
          >
            <option value="">All Warehouses</option>
            <option 
              v-for="warehouse in warehouses" 
              :key="warehouse.id" 
              :value="warehouse.name"
            >
              {{ warehouse.name }}
            </option>
          </select>
        </div>
      </div>

      <div class="advanced-filter-actions">
        <button @click="applyAdvancedFilters" class="apply-btn">Apply Filters</button>
        <button @click="resetFilters" class="reset-btn">Reset All</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.car-stock-filter {
  margin-bottom: 20px;
  background-color: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
}

.filter-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.filter-input {
  flex: 1;
}

.filter-input input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.filter-actions {
  display: flex;
  gap: 8px;
}

.toggle-btn, .reset-btn, .apply-btn {
  padding: 10px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.toggle-btn {
  background-color: #6366f1;
  color: white;
}

.reset-btn {
  background-color: #9ca3af;
  color: white;
}

.apply-btn {
  background-color: #10b981;
  color: white;
}

.toggle-btn:hover {
  background-color: #4f46e5;
}

.reset-btn:hover {
  background-color: #6b7280;
}

.apply-btn:hover {
  background-color: #059669;
}

.advanced-filters {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
}

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-field label {
  font-size: 14px;
  font-weight: 500;
  color: #4b5563;
}

.filter-field input,
.filter-field select {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.range-inputs,
.date-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.range-inputs input,
.date-inputs input {
  flex: 1;
}

.range-inputs span,
.date-inputs span {
  color: #6b7280;
  font-size: 12px;
}

.advanced-filter-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}
</style>