<script setup>
import { ref, defineEmits, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const emit = defineEmits(['filter'])

const { callApi } = useApi()

// Add loading states
const isLoading = ref(false)
const isProcessing = ref({
  basic: false,
  advanced: false,
  reset: false,
})

// Reference data for dropdowns
const carNames = ref([])
const colors = ref([])
const loadingPorts = ref([])
const dischargePorts = ref([])
const clients = ref([])
const warehouses = ref([])

// Basic filter
const basicFilter = ref('')

// Add operator choice for multi-word search
const useAndOperator = ref(true) // Default to AND operator

// Advanced filter state
const showAdvancedFilter = ref(false)

// Watch for changes in showAdvancedFilter and handle express filter accordingly
watch(showAdvancedFilter, (newValue) => {
  if (newValue) {
    // When showing advanced filter: clear and disable express filter
    basicFilter.value = ''
    // The input will be disabled via the template
  } else {
    // When hiding advanced filter: enable express filter (input will be enabled via template)
    // Don't clear the filter here as user might want to keep their search
  }
})

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
  client_id_no: '',
  warehouse: '',
  container_ref: '',
  loading_status: '',
  documents_status: '',
  bl_status: '',
  warehouse_status: '',
  bill_ref: '',
  sell_bill_ref: '',
  tmp_client_status: '',
  exclude_whole_sale: false,
})

// Fetch reference data for filter dropdowns
const fetchReferenceData = async () => {
  isLoading.value = true
  try {
    // Fetch car names
    const carNamesResult = await callApi({
      query: 'SELECT id, car_name FROM cars_names ORDER BY car_name',
      params: [],
    })
    if (carNamesResult.success) {
      carNames.value = carNamesResult.data
    }

    // Fetch colors
    const colorsResult = await callApi({
      query: 'SELECT id, color FROM colors ORDER BY color',
      params: [],
    })
    if (colorsResult.success) {
      colors.value = colorsResult.data
    }

    // Fetch loading ports
    const loadingPortsResult = await callApi({
      query: 'SELECT id, loading_port FROM loading_ports ORDER BY loading_port',
      params: [],
    })
    if (loadingPortsResult.success) {
      loadingPorts.value = loadingPortsResult.data
    }

    // Fetch discharge ports
    const dischargePortsResult = await callApi({
      query: 'SELECT id, discharge_port FROM discharge_ports ORDER BY discharge_port',
      params: [],
    })
    if (dischargePortsResult.success) {
      dischargePorts.value = dischargePortsResult.data
    }

    // Fetch clients
    const clientsResult = await callApi({
      query: 'SELECT id, name FROM clients ORDER BY name',
      params: [],
    })
    if (clientsResult.success) {
      clients.value = clientsResult.data
    }

    // Fetch warehouses
    const warehousesResult = await callApi({
      query: 'SELECT id, warhouse_name FROM warehouses ORDER BY warhouse_name',
      params: [],
    })
    if (warehousesResult.success) {
      warehouses.value = warehousesResult.data
    }
  } catch (error) {
    console.error('Error fetching reference data:', error)
  } finally {
    isLoading.value = false
  }
}

// Apply basic filter
const applyBasicFilter = async () => {
  if (isProcessing.value.basic) return
  isProcessing.value.basic = true
  try {
    // If the basic filter is empty, reset all filters
    if (!basicFilter.value.trim()) {
      await resetFilters()
      return
    }

    // Split the search term into multiple words and filter out empty strings
    const searchWords = basicFilter.value
      .trim()
      .split(/\s+/)
      .filter((word) => word.length > 0)

    emit('filter', {
      basic: searchWords,
      basicOperator: useAndOperator.value ? 'AND' : 'OR',
      advanced: null,
    })
  } finally {
    isProcessing.value.basic = false
  }
}

// Handle Enter key press for basic filter
const handleBasicFilterKeydown = async (event) => {
  if (event.key === 'Enter') {
    await applyBasicFilter()
  }
}

// Apply advanced filters
const applyAdvancedFilters = async () => {
  if (isProcessing.value.advanced) return
  isProcessing.value.advanced = true
  try {
    emit('filter', { basic: null, advanced: advancedFilters.value })
  } finally {
    isProcessing.value.advanced = false
  }
}

// Reset filters
const resetFilters = async () => {
  if (isProcessing.value.reset) return
  isProcessing.value.reset = true
  try {
    basicFilter.value = ''
    Object.keys(advancedFilters.value).forEach((key) => {
      advancedFilters.value[key] = ''
    })
    emit('filter', { basic: '', advanced: null })
  } finally {
    isProcessing.value.reset = false
  }
}

// Toggle advanced filter visibility
const toggleAdvancedFilter = () => {
  showAdvancedFilter.value = !showAdvancedFilter.value
}

// Fetch reference data on component mount
fetchReferenceData()
</script>

<template>
  <div class="car-stock-filter" :class="{ 'is-loading': isLoading }">
    <!-- Basic Filter -->
    <div class="basic-filter">
      <div class="filter-row">
        <div class="filter-input">
          <i class="fas fa-search search-icon"></i>
          <input
            type="text"
            v-model="basicFilter"
            placeholder="Search: ID, Car, Color, VIN, Ports, Client... (Multiple words = AND/OR search)"
            @keydown="handleBasicFilterKeydown"
            :disabled="isProcessing.basic || showAdvancedFilter"
          />
          <button
            @click="applyBasicFilter"
            class="search-btn"
            :disabled="isProcessing.basic || showAdvancedFilter"
            :class="{ processing: isProcessing.basic }"
            :title="`Press Enter or click to search. Multiple words will be combined with ${useAndOperator ? 'AND' : 'OR'} operator.`"
          >
            <i class="fas fa-search"></i>
            <i v-if="isProcessing.basic" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </div>
        <div class="operator-choice">
          <label class="operator-checkbox">
            <input
              type="checkbox"
              v-model="useAndOperator"
              :disabled="isProcessing.basic || showAdvancedFilter"
            />
            <span class="checkbox-label">
              <i class="fas fa-link"></i>
              If checked all words must mached
            </span>
          </label>
          <span class="operator-hint"> {{ useAndOperator ? 'AND' : 'OR' }} mode </span>
        </div>
        <div class="filter-actions">
          <button
            @click="toggleAdvancedFilter"
            class="toggle-btn"
            :disabled="isProcessing.advanced"
          >
            <i class="fas fa-sliders-h"></i>
            <span>{{ showAdvancedFilter ? 'Hide Advanced' : 'Show Advanced' }}</span>
          </button>
          <button
            @click="resetFilters"
            class="reset-btn"
            :disabled="isProcessing.reset"
            :class="{ processing: isProcessing.reset }"
          >
            <i class="fas fa-undo-alt"></i>
            <span>Reset</span>
            <i v-if="isProcessing.reset" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </div>
      </div>
    </div>

    <!-- Advanced Filters -->
    <div v-if="showAdvancedFilter" class="advanced-filters">
      <div class="filter-grid">
        <!-- ID Filter -->
        <div class="filter-field">
          <label for="id-filter">
            <i class="fas fa-hashtag"></i>
            ID
          </label>
          <input
            id="id-filter"
            type="text"
            v-model="advancedFilters.id"
            placeholder="Car ID"
            :disabled="isProcessing.advanced"
          />
        </div>

        <!-- Car Name Filter -->
        <div class="filter-field">
          <label for="car-name-filter">
            <i class="fas fa-car"></i>
            Car
          </label>
          <select
            id="car-name-filter"
            v-model="advancedFilters.car_name"
            :disabled="isProcessing.advanced"
          >
            <option value="">All Cars</option>
            <option v-for="car in carNames" :key="car.id" :value="car.car_name">
              {{ car.car_name }}
            </option>
          </select>
        </div>

        <!-- Color Filter -->
        <div class="filter-field">
          <label for="color-filter">
            <i class="fas fa-palette"></i>
            Color
          </label>
          <select
            id="color-filter"
            v-model="advancedFilters.color"
            :disabled="isProcessing.advanced"
          >
            <option value="">All Colors</option>
            <option v-for="color in colors" :key="color.id" :value="color.color">
              {{ color.color }}
            </option>
          </select>
        </div>

        <!-- VIN Filter -->
        <div class="filter-field">
          <label for="vin-filter">
            <i class="fas fa-barcode"></i>
            VIN
          </label>
          <input
            id="vin-filter"
            type="text"
            v-model="advancedFilters.vin"
            placeholder="VIN Number"
            :disabled="isProcessing.advanced"
          />
        </div>

        <!-- Loading Port Filter -->
        <div class="filter-field">
          <label for="loading-port-filter">
            <i class="fas fa-ship"></i>
            Loading Port
          </label>
          <select
            id="loading-port-filter"
            v-model="advancedFilters.loading_port"
            :disabled="isProcessing.advanced"
          >
            <option value="">All Loading Ports</option>
            <option v-for="port in loadingPorts" :key="port.id" :value="port.loading_port">
              {{ port.loading_port }}
            </option>
          </select>
        </div>

        <!-- Discharge Port Filter -->
        <div class="filter-field">
          <label for="discharge-port-filter">
            <i class="fas fa-anchor"></i>
            Discharge Port
          </label>
          <select
            id="discharge-port-filter"
            v-model="advancedFilters.discharge_port"
            :disabled="isProcessing.advanced"
          >
            <option value="">All Discharge Ports</option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.discharge_port">
              {{ port.discharge_port }}
            </option>
          </select>
        </div>

        <!-- Freight Range Filter -->
        <div class="filter-field range-field">
          <label>
            <i class="fas fa-dollar-sign"></i>
            Freight Range
          </label>
          <div class="range-inputs">
            <input
              type="number"
              v-model="advancedFilters.freight_min"
              placeholder="Min"
              :disabled="isProcessing.advanced"
            />
            <input
              type="number"
              v-model="advancedFilters.freight_max"
              placeholder="Max"
              :disabled="isProcessing.advanced"
            />
          </div>
        </div>

        <!-- Price Range Filter -->
        <div class="filter-field range-field">
          <label>
            <i class="fas fa-tags"></i>
            Price Range
          </label>
          <div class="range-inputs">
            <input
              type="number"
              v-model="advancedFilters.price_min"
              placeholder="Min"
              :disabled="isProcessing.advanced"
            />
            <input
              type="number"
              v-model="advancedFilters.price_max"
              placeholder="Max"
              :disabled="isProcessing.advanced"
            />
          </div>
        </div>

        <!-- Loading Date Range Filter -->
        <div class="filter-field range-field">
          <label>
            <i class="fas fa-calendar-alt"></i>
            Loading Date Range
          </label>
          <div class="range-inputs">
            <input
              type="date"
              v-model="advancedFilters.loading_date_from"
              :disabled="isProcessing.advanced"
            />
            <input
              type="date"
              v-model="advancedFilters.loading_date_to"
              :disabled="isProcessing.advanced"
            />
          </div>
        </div>

        <!-- Status Filter -->
        <div class="filter-field">
          <label for="status-filter">
            <i class="fas fa-info-circle"></i>
            Status
          </label>
          <select
            id="status-filter"
            v-model="advancedFilters.status"
            :disabled="isProcessing.advanced"
          >
            <option value="">All Status</option>
            <option value="available">Available</option>
            <option value="sold">Sold</option>
          </select>
        </div>

        <!-- Client Filter -->
        <div class="filter-field">
          <label for="client-filter">
            <i class="fas fa-user"></i>
            Client
          </label>
          <input
            id="client-filter"
            type="text"
            v-model="advancedFilters.client"
            placeholder="Client Name"
            :disabled="isProcessing.advanced"
          />
        </div>

        <!-- Client ID Number Filter -->
        <div class="filter-field">
          <label for="client-id-no-filter">
            <i class="fas fa-id-card"></i>
            Client ID Number
          </label>
          <input
            id="client-id-no-filter"
            type="text"
            v-model="advancedFilters.client_id_no"
            placeholder="Client ID Number"
            :disabled="isProcessing.advanced"
          />
        </div>

        <!-- Warehouse Filter -->
        <div class="filter-field">
          <label for="warehouse-filter">
            <i class="fas fa-warehouse"></i>
            Warehouse
          </label>
          <select
            id="warehouse-filter"
            v-model="advancedFilters.warehouse"
            :disabled="isProcessing.advanced"
          >
            <option value="">All Warehouses</option>
            <option
              v-for="warehouse in warehouses"
              :key="warehouse.id"
              :value="warehouse.warhouse_name"
            >
              {{ warehouse.warhouse_name }}
            </option>
          </select>
        </div>

        <!-- Container Reference Filter -->
        <div class="filter-field">
          <label for="container-ref-filter">
            <i class="fas fa-box"></i>
            Container Ref
          </label>
          <input
            id="container-ref-filter"
            type="text"
            v-model="advancedFilters.container_ref"
            placeholder="Container Reference"
            :disabled="isProcessing.advanced"
          />
        </div>

        <!-- Loading Status Filter -->
        <div class="filter-field">
          <label for="loading-status-filter">
            <i class="fas fa-truck-loading"></i>
            Loading Status
          </label>
          <select
            id="loading-status-filter"
            v-model="advancedFilters.loading_status"
            :disabled="isProcessing.advanced"
          >
            <option value="">All</option>
            <option value="loaded">Loaded</option>
            <option value="not_loaded">Not Loaded</option>
          </select>
        </div>

        <!-- Documents Status Filter -->
        <div class="filter-field">
          <label for="documents-status-filter">
            <i class="fas fa-file-alt"></i>
            Documents Status
          </label>
          <select
            id="documents-status-filter"
            v-model="advancedFilters.documents_status"
            :disabled="isProcessing.advanced"
          >
            <option value="">All</option>
            <option value="received">Documents Received</option>
            <option value="not_received">Documents Not Received</option>
          </select>
        </div>

        <!-- BL Status Filter -->
        <div class="filter-field">
          <label for="bl-status-filter">
            <i class="fas fa-ship"></i>
            BL Status
          </label>
          <select
            id="bl-status-filter"
            v-model="advancedFilters.bl_status"
            :disabled="isProcessing.advanced"
          >
            <option value="">All</option>
            <option value="received">BL Received</option>
            <option value="not_received">BL Not Received</option>
          </select>
        </div>

        <!-- Warehouse Status Filter -->
        <div class="filter-field">
          <label for="warehouse-status-filter">
            <i class="fas fa-warehouse"></i>
            Warehouse Status
          </label>
          <select
            id="warehouse-status-filter"
            v-model="advancedFilters.warehouse_status"
            :disabled="isProcessing.advanced"
          >
            <option value="">All</option>
            <option value="in_warehouse">In Warehouse</option>
            <option value="not_in_warehouse">Not In Warehouse</option>
          </select>
        </div>

        <!-- Buy Bill Filter -->
        <div class="filter-field">
          <label for="bill-ref-filter">
            <i class="fas fa-file-invoice"></i>
            Buy Bill
          </label>
          <input
            id="bill-ref-filter"
            type="text"
            v-model="advancedFilters.bill_ref"
            placeholder="Buy Bill Reference"
            :disabled="isProcessing.advanced"
          />
        </div>

        <!-- Sell Bill Filter -->
        <div class="filter-field">
          <label for="sell-bill-ref-filter">
            <i class="fas fa-file-invoice-dollar"></i>
            Sell Bill
          </label>
          <input
            id="sell-bill-ref-filter"
            type="text"
            v-model="advancedFilters.sell_bill_ref"
            placeholder="Sell Bill Reference"
            :disabled="isProcessing.advanced"
          />
        </div>

        <!-- Temporary Client Status Filter -->
        <div class="filter-field">
          <label for="tmp-client-status-filter">
            <i class="fas fa-user-clock"></i>
            Temporary Client
          </label>
          <select
            id="tmp-client-status-filter"
            v-model="advancedFilters.tmp_client_status"
            :disabled="isProcessing.advanced"
          >
            <option value="">All</option>
            <option value="tmp">Temporary Client</option>
            <option value="permanent">Permanent Client</option>
          </select>
        </div>

        <!-- Exclude Whole Sale Filter -->
        <div class="filter-field checkbox-field">
          <label for="exclude-whole-sale-filter" class="checkbox-label">
            <input
              id="exclude-whole-sale-filter"
              type="checkbox"
              v-model="advancedFilters.exclude_whole_sale"
              :disabled="isProcessing.advanced"
            />
            <span class="checkbox-text">
              <i class="fas fa-ban"></i>
              Exclude Whole Sale Cars
            </span>
          </label>
        </div>

        <!-- Add the Apply and Reset buttons at the bottom -->
        <div class="advanced-filter-actions">
          <button
            @click="applyAdvancedFilters"
            class="apply-btn"
            :disabled="isProcessing.advanced"
            :class="{ processing: isProcessing.advanced }"
          >
            <i class="fas fa-check"></i>
            <span>Apply Filters</span>
            <i v-if="isProcessing.advanced" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
          <button
            @click="resetFilters"
            class="reset-btn"
            :disabled="isProcessing.reset"
            :class="{ processing: isProcessing.reset }"
          >
            <i class="fas fa-undo-alt"></i>
            <span>Reset All</span>
            <i v-if="isProcessing.reset" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </div>
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
  position: relative;
}

.car-stock-filter.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.filter-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.filter-input {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
  z-index: 2;
}

.filter-input input {
  width: 100%;
  padding: 10px 36px 10px 36px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  border-right: none;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

.search-btn {
  padding: 10px 16px;
  background-color: #3b82f6;
  color: white;
  border: 1px solid #3b82f6;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
  min-width: 44px;
  justify-content: center;
}

.search-btn:hover:not(:disabled) {
  background-color: #2563eb;
  border-color: #2563eb;
}

.search-btn:disabled {
  background-color: #9ca3af;
  border-color: #9ca3af;
  cursor: not-allowed;
}

.search-btn.processing {
  position: relative;
  pointer-events: none;
}

.search-btn.processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

.filter-input .loading-indicator {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
}

.search-help {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 8px 12px;
  background-color: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 4px;
  font-size: 12px;
  color: #0369a1;
}

.search-help i {
  color: #0284c7;
  flex-shrink: 0;
}

.search-help span {
  line-height: 1.4;
}

.filter-actions {
  display: flex;
  gap: 8px;
}

.toggle-btn,
.reset-btn,
.apply-btn {
  padding: 10px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
}

.toggle-btn i,
.reset-btn i,
.apply-btn i {
  font-size: 1.1em;
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

.toggle-btn:hover:not(:disabled) {
  background-color: #4f46e5;
}

.reset-btn:hover:not(:disabled) {
  background-color: #6b7280;
}

.apply-btn:hover:not(:disabled) {
  background-color: #059669;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.processing {
  position: relative;
  pointer-events: none;
}

.processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
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
  display: flex;
  align-items: center;
  gap: 6px;
}

.filter-field label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.filter-field input,
.filter-field select {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.2s ease;
}

.filter-field input:focus,
.filter-field select:focus {
  border-color: #6366f1;
  outline: none;
}

.filter-field input:disabled,
.filter-field select:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

/* Checkbox field styles */
.checkbox-field {
  display: flex;
  align-items: center;
}

.checkbox-field .checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
  font-size: 14px;
  font-weight: 500;
  color: #4b5563;
}

.checkbox-field .checkbox-label input[type='checkbox'] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
}

.checkbox-field .checkbox-label input[type='checkbox']:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.checkbox-field .checkbox-text {
  display: flex;
  align-items: center;
  gap: 6px;
}

.checkbox-field .checkbox-text i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.range-inputs {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 8px;
  align-items: center;
}

.range-inputs span {
  color: #6b7280;
  font-size: 12px;
}

.advanced-filter-actions {
  grid-column: 1 / -1;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fa-spin {
  animation: spin 1s linear infinite;
}

@media (max-width: 640px) {
  .filter-row {
    flex-direction: column;
  }

  .filter-actions {
    width: 100%;
  }

  .toggle-btn,
  .reset-btn {
    flex: 1;
  }

  .filter-input {
    flex-direction: column;
    gap: 8px;
  }

  .filter-input input {
    border-radius: 4px;
    border-right: 1px solid #ddd;
  }

  .search-btn {
    border-radius: 4px;
    width: 100%;
    justify-content: center;
    padding: 12px 16px;
    font-size: 16px;
    min-height: 44px;
  }

  .search-help {
    font-size: 11px;
    padding: 6px 10px;
  }
}

/* Operator choice styles */
.operator-choice {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.operator-checkbox {
  display: flex;
  align-items: center;
  cursor: pointer;
  user-select: none;
  gap: 8px;
}

.operator-checkbox input[type='checkbox'] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--primary-color);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.9rem;
  color: var(--text-color);
}

.checkbox-label i {
  color: var(--primary-color);
  font-size: 0.8rem;
}

.operator-hint {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--primary-color);
  background: rgba(var(--primary-color-rgb), 0.1);
  padding: 4px 8px;
  border-radius: 4px;
  border: 1px solid rgba(var(--primary-color-rgb), 0.2);
}

/* Responsive styles for operator choice */
@media (max-width: 768px) {
  .operator-choice {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .operator-hint {
    align-self: flex-end;
  }
}

@media (max-width: 480px) {
  .operator-choice {
    padding: 6px 8px;
  }

  .checkbox-label {
    font-size: 0.85rem;
  }

  .operator-hint {
    font-size: 0.75rem;
    padding: 3px 6px;
  }
}

/* Search help styles */
</style>
