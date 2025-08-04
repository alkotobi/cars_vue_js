<script setup>
import { ref, defineEmits, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  isAdmin: {
    type: Boolean,
    default: false,
  },
})

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

// More filters state
const showMoreFilters = ref(false)

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
  sold_date_from: '',
  sold_date_to: '',
  status: '',
  client: '',
  client_id_no: '',
  container_status: '',
  warehouse: '',
  container_ref: '',
  export_lisence_ref: '',
  loading_status: '',
  documents_status: '',
  bl_status: '',
  warehouse_status: '',
  bill_ref: '',
  sell_bill_ref: '',
  tmp_client_status: '',
  exclude_whole_sale: false,
  hidden: '',
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

// Toggle more filters visibility
const toggleMoreFilters = () => {
  showMoreFilters.value = !showMoreFilters.value
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
            :placeholder="t('carStockFilter.searchPlaceholder')"
            @keydown="handleBasicFilterKeydown"
            :disabled="isProcessing.basic || showAdvancedFilter"
          />
          <button
            @click="applyBasicFilter"
            class="search-btn"
            :disabled="isProcessing.basic || showAdvancedFilter"
            :class="{ processing: isProcessing.basic }"
            :title="
              t('carStockFilter.searchButtonTitle', { operator: useAndOperator ? 'AND' : 'OR' })
            "
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
              {{ t('carStockFilter.ifCheckedAllWordsMustMatched') }}
            </span>
          </label>
          <span class="operator-hint">
            {{ useAndOperator ? t('carStockFilter.andMode') : t('carStockFilter.orMode') }}
          </span>
        </div>
        <div class="filter-actions">
          <button
            @click="toggleAdvancedFilter"
            class="toggle-btn"
            :disabled="isProcessing.advanced"
          >
            <i class="fas fa-sliders-h"></i>
            <span>{{
              showAdvancedFilter
                ? t('carStockFilter.hideAdvanced')
                : t('carStockFilter.showAdvanced')
            }}</span>
          </button>
          <button
            @click="resetFilters"
            class="reset-btn"
            :disabled="isProcessing.reset"
            :class="{ processing: isProcessing.reset }"
          >
            <i class="fas fa-undo-alt"></i>
            <span>{{ t('carStockFilter.reset') }}</span>
            <i v-if="isProcessing.reset" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </div>
      </div>
    </div>

    <!-- Advanced Filters -->
    <div v-if="showAdvancedFilter" class="advanced-filters">
      <div class="filter-grid">
        <!-- Car and ID Filter Group -->
        <div class="filter-field combined-car-id-filter">
          <label for="car-name-filter">
            <i class="fas fa-car"></i>
            {{ t('carStockFilter.carAndId') }}
          </label>
          <div class="car-id-filter-inputs">
            <select
              id="car-name-filter"
              v-model="advancedFilters.car_name"
              :disabled="isProcessing.advanced"
              class="car-name-select"
            >
              <option value="">{{ t('carStockFilter.allCars') }}</option>
              <option v-for="car in carNames" :key="car.id" :value="car.car_name">
                {{ car.car_name }}
              </option>
            </select>
            <input
              id="id-filter"
              type="text"
              v-model="advancedFilters.id"
              :placeholder="t('carStockFilter.carId')"
              :disabled="isProcessing.advanced"
              class="car-id-input"
            />
          </div>
        </div>

        <!-- Color and VIN Filter Group -->
        <div class="filter-field combined-color-vin-filter">
          <label for="color-filter">
            <i class="fas fa-palette"></i>
            {{ t('carStockFilter.colorAndVin') }}
          </label>
          <div class="color-vin-filter-inputs">
            <select
              id="color-filter"
              v-model="advancedFilters.color"
              :disabled="isProcessing.advanced"
              class="color-select"
            >
              <option value="">{{ t('carStockFilter.allColors') }}</option>
              <option v-for="color in colors" :key="color.id" :value="color.color">
                {{ color.color }}
              </option>
            </select>
            <input
              id="vin-filter"
              type="text"
              v-model="advancedFilters.vin"
              :placeholder="t('carStockFilter.vinNumber')"
              :disabled="isProcessing.advanced"
              class="vin-input"
            />
          </div>
        </div>

        <!-- Sold Date Range Filter -->
        <div class="filter-field range-field sold-date-group">
          <label>
            <i class="fas fa-dollar-sign"></i>
            {{ t('carStockFilter.soldDateRange') }}
          </label>
          <div class="range-inputs">
            <input
              type="date"
              v-model="advancedFilters.sold_date_from"
              :disabled="isProcessing.advanced"
              placeholder="From"
            />
            <input
              type="date"
              v-model="advancedFilters.sold_date_to"
              :disabled="isProcessing.advanced"
              placeholder="To"
            />
          </div>
        </div>

        <!-- Car Availability Filter -->
        <div class="filter-field status-group">
          <label for="status-filter">
            <i class="fas fa-car"></i>
            {{ t('carStockFilter.carAvailability') }}
          </label>
          <select
            id="status-filter"
            v-model="advancedFilters.status"
            :disabled="isProcessing.advanced"
          >
            <option value="">{{ t('carStockFilter.allCars') }}</option>
            <option value="available">{{ t('carStockFilter.available') }}</option>
            <option value="sold">{{ t('carStockFilter.sold') }}</option>
          </select>
        </div>

        <!-- Client Filter Group -->
        <div class="filter-field combined-client-filter">
          <label for="client-filter">
            <i class="fas fa-user"></i>
            {{ t('carStockFilter.client') }}
          </label>
          <div class="client-filter-inputs">
            <input
              id="client-filter"
              type="text"
              v-model="advancedFilters.client"
              :placeholder="t('carStockFilter.clientName')"
              :disabled="isProcessing.advanced"
              class="client-text-input"
            />
            <input
              id="client-id-no-filter"
              type="text"
              v-model="advancedFilters.client_id_no"
              :placeholder="t('carStockFilter.clientIdNumber')"
              :disabled="isProcessing.advanced"
              class="client-id-input"
            />
          </div>
        </div>

        <!-- Container Reference Filter -->
        <div class="filter-field combined-container-filter">
          <label for="container-ref-filter">
            <i class="fas fa-box"></i>
            {{ t('carStockFilter.containerRef') }}
          </label>
          <div class="container-filter-inputs">
            <input
              id="container-ref-filter"
              type="text"
              v-model="advancedFilters.container_ref"
              :placeholder="t('carStockFilter.containerReference')"
              :disabled="isProcessing.advanced"
              class="container-text-input"
            />
            <div class="container-status-wrapper">
              <label for="container-status-filter" class="container-status-label">
                <i class="fas fa-filter"></i>
                {{ t('carStockFilter.containerStatus') }}
              </label>
              <select
                id="container-status-filter"
                v-model="advancedFilters.container_status"
                :disabled="isProcessing.advanced"
                class="container-status-select"
              >
                <option value="">{{ t('carStockFilter.all') }}</option>
                <option value="has_container">{{ t('carStockFilter.hasContainerRef') }}</option>
                <option value="has_not_container">
                  {{ t('carStockFilter.hasNotContainerRef') }}
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- Loading Status Filter -->
        <div class="filter-field">
          <label for="loading-status-filter">
            <i class="fas fa-truck-loading"></i>
            {{ t('carStockFilter.loadingStatus') }}
          </label>
          <select
            id="loading-status-filter"
            v-model="advancedFilters.loading_status"
            :disabled="isProcessing.advanced"
          >
            <option value="">{{ t('carStockFilter.all') }}</option>
            <option value="loaded">{{ t('carStockFilter.loaded') }}</option>
            <option value="not_loaded">{{ t('carStockFilter.notLoaded') }}</option>
          </select>
        </div>

        <!-- Bill Filter Group -->
        <div class="filter-field combined-bill-filter">
          <label for="bill-ref-filter">
            <i class="fas fa-file-invoice"></i>
            {{ t('carStockFilter.bills') }}
          </label>
          <div class="bill-filter-inputs">
            <input
              id="bill-ref-filter"
              type="text"
              v-model="advancedFilters.bill_ref"
              :placeholder="t('carStockFilter.buyBillReference')"
              :disabled="isProcessing.advanced"
              class="buy-bill-input"
            />
            <input
              id="sell-bill-ref-filter"
              type="text"
              v-model="advancedFilters.sell_bill_ref"
              :placeholder="t('carStockFilter.sellBillReference')"
              :disabled="isProcessing.advanced"
              class="sell-bill-input"
            />
          </div>
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
              {{ t('carStockFilter.excludeWholeSaleCars') }}
            </span>
          </label>
        </div>

        <!-- Hidden Cars Filter (Admin Only) -->
        <div v-if="props.isAdmin" class="filter-field">
          <label for="hidden-filter">
            <i class="fas fa-eye-slash"></i>
            {{ t('carStockFilter.hiddenCars') }}
          </label>
          <select
            id="hidden-filter"
            v-model="advancedFilters.hidden"
            :disabled="isProcessing.advanced"
          >
            <option value="">{{ t('carStockFilter.allCars') }}</option>
            <option value="1">{{ t('carStockFilter.hiddenCarsOnly') }}</option>
            <option value="0">{{ t('carStockFilter.visibleCarsOnly') }}</option>
          </select>
        </div>

        <!-- More Filters Section -->
        <div class="more-filters-section">
          <div class="more-filters-toggle">
            <button
              @click="toggleMoreFilters"
              class="more-filters-btn"
              :class="{ expanded: showMoreFilters }"
            >
              <i class="fas fa-ellipsis-h"></i>
              <span>{{
                showMoreFilters ? t('carStockFilter.showLess') : t('carStockFilter.showMore')
              }}</span>
              <i class="fas fa-chevron-down" :class="{ rotated: showMoreFilters }"></i>
            </button>
          </div>

          <div v-if="showMoreFilters" class="more-filters-content">
            <div class="filter-grid">
              <!-- Loading Port Filter -->
              <div class="filter-field">
                <label for="loading-port-filter">
                  <i class="fas fa-ship"></i>
                  {{ t('carStockFilter.loadingPort') }}
                </label>
                <select
                  id="loading-port-filter"
                  v-model="advancedFilters.loading_port"
                  :disabled="isProcessing.advanced"
                >
                  <option value="">{{ t('carStockFilter.allLoadingPorts') }}</option>
                  <option v-for="port in loadingPorts" :key="port.id" :value="port.loading_port">
                    {{ port.loading_port }}
                  </option>
                </select>
              </div>

              <!-- Discharge Port Filter -->
              <div class="filter-field">
                <label for="discharge-port-filter">
                  <i class="fas fa-anchor"></i>
                  {{ t('carStockFilter.dischargePort') }}
                </label>
                <select
                  id="discharge-port-filter"
                  v-model="advancedFilters.discharge_port"
                  :disabled="isProcessing.advanced"
                >
                  <option value="">{{ t('carStockFilter.allDischargePorts') }}</option>
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
              <div class="filter-field range-field">
                <label>
                  <i class="fas fa-dollar-sign"></i>
                  {{ t('carStockFilter.freightRange') }}
                </label>
                <div class="range-inputs">
                  <input
                    type="number"
                    v-model="advancedFilters.freight_min"
                    :placeholder="t('carStockFilter.min')"
                    :disabled="isProcessing.advanced"
                  />
                  <input
                    type="number"
                    v-model="advancedFilters.freight_max"
                    :placeholder="t('carStockFilter.max')"
                    :disabled="isProcessing.advanced"
                  />
                </div>
              </div>

              <!-- Price Range Filter -->
              <div class="filter-field range-field">
                <label>
                  <i class="fas fa-tags"></i>
                  {{ t('carStockFilter.priceRange') }}
                </label>
                <div class="range-inputs">
                  <input
                    type="number"
                    v-model="advancedFilters.price_min"
                    :placeholder="t('carStockFilter.min')"
                    :disabled="isProcessing.advanced"
                  />
                  <input
                    type="number"
                    v-model="advancedFilters.price_max"
                    :placeholder="t('carStockFilter.max')"
                    :disabled="isProcessing.advanced"
                  />
                </div>
              </div>

              <!-- Loading Date Range Filter -->
              <div class="filter-field range-field loading-date-group">
                <label>
                  <i class="fas fa-calendar-alt"></i>
                  {{ t('carStockFilter.loadingDateRange') }}
                </label>
                <div class="range-inputs">
                  <input
                    type="date"
                    v-model="advancedFilters.loading_date_from"
                    :disabled="isProcessing.advanced"
                    placeholder="From"
                  />
                  <input
                    type="date"
                    v-model="advancedFilters.loading_date_to"
                    :disabled="isProcessing.advanced"
                    placeholder="To"
                  />
                </div>
              </div>

              <!-- Warehouse Filter -->
              <div class="filter-field">
                <label for="warehouse-filter">
                  <i class="fas fa-warehouse"></i>
                  {{ t('carStockFilter.warehouse') }}
                </label>
                <select
                  id="warehouse-filter"
                  v-model="advancedFilters.warehouse"
                  :disabled="isProcessing.advanced"
                >
                  <option value="">{{ t('carStockFilter.allWarehouses') }}</option>
                  <option
                    v-for="warehouse in warehouses"
                    :key="warehouse.id"
                    :value="warehouse.warhouse_name"
                  >
                    {{ warehouse.warhouse_name }}
                  </option>
                </select>
              </div>

              <!-- Export License Filter -->
              <div class="filter-field">
                <label for="export-license-filter">
                  <i class="fas fa-file-signature"></i>
                  {{ t('carStockFilter.exportLicense') }}
                </label>
                <input
                  id="export-license-filter"
                  type="text"
                  v-model="advancedFilters.export_lisence_ref"
                  :placeholder="t('carStockFilter.exportLicenseReference')"
                  :disabled="isProcessing.advanced"
                />
              </div>

              <!-- Documents Status Filter -->
              <div class="filter-field">
                <label for="documents-status-filter">
                  <i class="fas fa-file-alt"></i>
                  {{ t('carStockFilter.documentsStatus') }}
                </label>
                <select
                  id="documents-status-filter"
                  v-model="advancedFilters.documents_status"
                  :disabled="isProcessing.advanced"
                >
                  <option value="">{{ t('carStockFilter.all') }}</option>
                  <option value="received">{{ t('carStockFilter.documentsReceived') }}</option>
                  <option value="not_received">
                    {{ t('carStockFilter.documentsNotReceived') }}
                  </option>
                </select>
              </div>

              <!-- BL Status Filter -->
              <div class="filter-field">
                <label for="bl-status-filter">
                  <i class="fas fa-ship"></i>
                  {{ t('carStockFilter.blStatus') }}
                </label>
                <select
                  id="bl-status-filter"
                  v-model="advancedFilters.bl_status"
                  :disabled="isProcessing.advanced"
                >
                  <option value="">{{ t('carStockFilter.all') }}</option>
                  <option value="received">{{ t('carStockFilter.blReceived') }}</option>
                  <option value="not_received">{{ t('carStockFilter.blNotReceived') }}</option>
                </select>
              </div>

              <!-- Warehouse Status Filter -->
              <div class="filter-field">
                <label for="warehouse-status-filter">
                  <i class="fas fa-warehouse"></i>
                  {{ t('carStockFilter.warehouseStatus') }}
                </label>
                <select
                  id="warehouse-status-filter"
                  v-model="advancedFilters.warehouse_status"
                  :disabled="isProcessing.advanced"
                >
                  <option value="">{{ t('carStockFilter.all') }}</option>
                  <option value="in_warehouse">{{ t('carStockFilter.inWarehouse') }}</option>
                  <option value="not_in_warehouse">{{ t('carStockFilter.notInWarehouse') }}</option>
                </select>
              </div>

              <!-- Temporary Client Status Filter -->
              <div class="filter-field">
                <label for="tmp-client-status-filter">
                  <i class="fas fa-user-clock"></i>
                  {{ t('carStockFilter.temporaryClient') }}
                </label>
                <select
                  id="tmp-client-status-filter"
                  v-model="advancedFilters.tmp_client_status"
                  :disabled="isProcessing.advanced"
                >
                  <option value="">{{ t('carStockFilter.all') }}</option>
                  <option value="tmp">{{ t('carStockFilter.temporaryClientOption') }}</option>
                  <option value="permanent">{{ t('carStockFilter.permanentClientOption') }}</option>
                </select>
              </div>
            </div>
          </div>
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
            <span>{{ t('carStockFilter.applyFilters') }}</span>
            <i v-if="isProcessing.advanced" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
          <button
            @click="resetFilters"
            class="reset-btn"
            :disabled="isProcessing.reset"
            :class="{ processing: isProcessing.reset }"
          >
            <i class="fas fa-undo-alt"></i>
            <span>{{ t('carStockFilter.resetAll') }}</span>
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

.filter-sections {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.filter-section {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid #f3f4f6;
  font-weight: 600;
  color: #374151;
  font-size: 14px;
}

.section-header i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: #fafafa;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  transition: all 0.2s ease;
}

.filter-field:hover {
  background: #f8f9fa;
  border-color: #d1d5db;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.filter-field label {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.filter-field label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
  font-size: 14px;
}

.filter-field input,
.filter-field select {
  padding: 10px 12px;
  border: 2px solid #e5e7eb;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  background-color: white;
}

.filter-field input:focus,
.filter-field select:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  outline: none;
}

.filter-field input:disabled,
.filter-field select:disabled {
  background-color: #f9fafb;
  border-color: #d1d5db;
  cursor: not-allowed;
  opacity: 0.7;
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
  width: 100%;
}

.range-inputs span {
  color: #6b7280;
  font-size: 12px;
}

/* Ensure range inputs fit within their containers */
.range-field .range-inputs {
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  width: 100%;
}

.range-field input {
  width: 100%;
  box-sizing: border-box;
}

/* Sold Date Group Styles */
.sold-date-group {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 2px solid #0ea5e9;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(14, 165, 233, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.sold-date-group::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #0ea5e9, #3b82f6);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.sold-date-group label {
  color: #0c4a6e;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.sold-date-group label i {
  color: #0ea5e9;
  font-size: 16px;
}

.sold-date-group .range-inputs {
  grid-template-columns: 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 12px;
  width: 100%;
}

.sold-date-group input {
  border: 2px solid #0ea5e9;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.sold-date-group input:focus {
  border-color: #0284c7;
  box-shadow: 0 0 0 2px rgba(14, 165, 233, 0.2);
  outline: none;
}

.sold-date-group input::placeholder {
  color: #6b7280;
  font-weight: 400;
}

/* Loading Date Group Styles */
.loading-date-group {
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  border: 2px solid #22c55e;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(34, 197, 94, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.loading-date-group::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #22c55e, #16a34a);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.loading-date-group label {
  color: #14532d;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.loading-date-group label i {
  color: #22c55e;
  font-size: 16px;
}

.loading-date-group .range-inputs {
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  width: 100%;
}

.loading-date-group input {
  border: 2px solid #22c55e;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.loading-date-group input:focus {
  border-color: #16a34a;
  box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.2);
  outline: none;
}

.loading-date-group input::placeholder {
  color: #6b7280;
  font-weight: 400;
}

/* Status Group Styles */
.status-group {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  border: 2px solid #f59e0b;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.status-group::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #f59e0b, #d97706);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.status-group label {
  color: #92400e;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.status-group label i {
  color: #f59e0b;
  font-size: 16px;
}

.status-group select {
  border: 2px solid #f59e0b;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.status-group select:focus {
  border-color: #d97706;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.2);
  outline: none;
}

/* More Filters Section Styles */
.more-filters-section {
  margin-top: 16px;
}

.more-filters-toggle {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}

.more-filters-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.2);
}

.more-filters-btn:hover {
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.more-filters-btn.expanded {
  background: linear-gradient(135deg, #7c3aed 0%, #a855f7 100%);
}

.more-filters-btn i {
  font-size: 14px;
  transition: transform 0.3s ease;
}

.more-filters-btn .fa-chevron-down {
  transition: transform 0.3s ease;
}

.more-filters-btn .fa-chevron-down.rotated {
  transform: rotate(180deg);
}

.more-filters-content {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-top: 8px;
}

.more-filters-content .filter-grid {
  background: transparent;
  border: none;
  padding: 0;
  box-shadow: none;
}

/* Checkbox Field Styles */
.checkbox-field {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: 500;
  color: #374151;
  user-select: none;
}

.checkbox-input {
  position: absolute;
  opacity: 0;
  cursor: pointer;
  height: 0;
  width: 0;
}

.checkbox-custom {
  position: relative;
  height: 18px;
  width: 18px;
  background-color: #fff;
  border: 2px solid #d1d5db;
  border-radius: 3px;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.checkbox-label:hover .checkbox-custom {
  border-color: #3b82f6;
  background-color: #f0f9ff;
}

.checkbox-input:checked ~ .checkbox-custom {
  background-color: #3b82f6;
  border-color: #3b82f6;
}

.checkbox-input:checked ~ .checkbox-custom::after {
  content: '';
  position: absolute;
  left: 5px;
  top: 2px;
  width: 6px;
  height: 10px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.checkbox-input:disabled ~ .checkbox-custom {
  background-color: #f3f4f6;
  border-color: #d1d5db;
  opacity: 0.6;
}

.checkbox-label i {
  color: #6b7280;
  font-size: 14px;
}

/* Combined Container Filter Styles */
.combined-container-filter {
  margin-bottom: 16px;
}

.container-filter-inputs {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.container-text-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s ease;
}

.container-text-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.container-checkbox-wrapper {
  display: flex;
  align-items: center;
  padding: 8px 0;
}

.container-checkbox-wrapper .checkbox-label {
  margin: 0;
  font-size: 13px;
  color: #6b7280;
}

.container-checkbox-wrapper .checkbox-text {
  font-weight: 500;
  color: #374151;
}

.container-checkbox-wrapper .checkbox-custom {
  height: 16px;
  width: 16px;
}

.container-status-wrapper {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px 0;
}

.container-status-label {
  font-size: 12px;
  font-weight: 500;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 4px;
}

.container-status-label i {
  color: #9ca3af;
  font-size: 11px;
}

.container-status-select {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
  background-color: white;
  color: #374151;
}

.container-status-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

/* Client Filter Group Styles */
.combined-client-filter {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 2px solid #0ea5e9;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(14, 165, 233, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.combined-client-filter::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #0ea5e9, #3b82f6);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.combined-client-filter label {
  color: #0c4a6e;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.combined-client-filter label i {
  color: #0ea5e9;
  font-size: 16px;
}

.client-filter-inputs {
  display: grid;
  grid-template-columns: 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 12px;
  width: 100%;
}

.client-text-input,
.client-id-input {
  border: 2px solid #0ea5e9;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.client-text-input:focus,
.client-id-input:focus {
  border-color: #0284c7;
  box-shadow: 0 0 0 2px rgba(14, 165, 233, 0.2);
  outline: none;
}

.client-text-input::placeholder,
.client-id-input::placeholder {
  color: #6b7280;
  font-weight: 400;
}

/* Bill Filter Group Styles */
.combined-bill-filter {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  border: 2px solid #f59e0b;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.combined-bill-filter::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #f59e0b, #d97706);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.combined-bill-filter label {
  color: #92400e;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.combined-bill-filter label i {
  color: #f59e0b;
  font-size: 16px;
}

.bill-filter-inputs {
  display: grid;
  grid-template-columns: 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 12px;
  width: 100%;
}

.buy-bill-input,
.sell-bill-input {
  border: 2px solid #f59e0b;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.buy-bill-input:focus,
.sell-bill-input:focus {
  border-color: #d97706;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.2);
  outline: none;
}

.buy-bill-input::placeholder,
.sell-bill-input::placeholder {
  color: #6b7280;
  font-weight: 400;
}

/* Car and ID Filter Group Styles */
.combined-car-id-filter {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  border: 2px solid #10b981;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.combined-car-id-filter::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #10b981, #059669);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.combined-car-id-filter label {
  color: #064e3b;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.combined-car-id-filter label i {
  color: #10b981;
  font-size: 16px;
}

.car-id-filter-inputs {
  display: grid;
  grid-template-columns: 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 12px;
  width: 100%;
}

.car-name-select,
.car-id-input {
  border: 2px solid #10b981;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.car-name-select:focus,
.car-id-input:focus {
  border-color: #059669;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
  outline: none;
}

.car-name-select::placeholder,
.car-id-input::placeholder {
  color: #6b7280;
  font-weight: 400;
}

/* Color and VIN Filter Group Styles */
.combined-color-vin-filter {
  background: linear-gradient(135deg, #fdf2f8 0%, #fce7f3 100%);
  border: 2px solid #ec4899;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(236, 72, 153, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.combined-color-vin-filter::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #ec4899, #db2777);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.combined-color-vin-filter label {
  color: #831843;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.combined-color-vin-filter label i {
  color: #ec4899;
  font-size: 16px;
}

.color-vin-filter-inputs {
  display: grid;
  grid-template-columns: 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 12px;
  width: 100%;
}

.color-select,
.vin-input {
  border: 2px solid #ec4899;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.color-select:focus,
.vin-input:focus {
  border-color: #db2777;
  box-shadow: 0 0 0 2px rgba(236, 72, 153, 0.2);
  outline: none;
}

.color-select::placeholder,
.vin-input::placeholder {
  color: #6b7280;
  font-weight: 400;
}

/* Car Availability Filter Styles */
.status-group {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 2px solid #0ea5e9;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(14, 165, 233, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.status-group::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #0ea5e9, #3b82f6);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.status-group label {
  color: #0c4a6e;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.status-group label i {
  color: #0ea5e9;
  font-size: 16px;
}

.status-group select {
  border: 2px solid #0ea5e9;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.status-group select:focus {
  border-color: #0284c7;
  box-shadow: 0 0 0 2px rgba(14, 165, 233, 0.2);
  outline: none;
}

/* Loading Status Filter Styles */
.filter-field:has(#loading-status-filter) {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  border: 2px solid #f59e0b;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.filter-field:has(#loading-status-filter)::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #f59e0b, #d97706);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.filter-field:has(#loading-status-filter) label {
  color: #92400e;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.filter-field:has(#loading-status-filter) label i {
  color: #f59e0b;
  font-size: 16px;
}

.filter-field:has(#loading-status-filter) select {
  border: 2px solid #f59e0b;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.filter-field:has(#loading-status-filter) select:focus {
  border-color: #d97706;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.2);
  outline: none;
}

/* Exclude Whole Sale Filter Styles */
.checkbox-field {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  border: 2px solid #ef4444;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.checkbox-field::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #ef4444, #dc2626);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.checkbox-field .checkbox-label {
  color: #7f1d1d;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.checkbox-field .checkbox-text {
  color: #7f1d1d;
  font-weight: 500;
  font-size: 13px;
}

.checkbox-field .checkbox-text i {
  color: #ef4444;
  font-size: 14px;
}

/* Hidden Cars Filter Styles */
.filter-field:has(#hidden-filter) {
  background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
  border: 2px solid #6b7280;
  border-radius: 12px;
  padding: 16px;
  position: relative;
  box-shadow: 0 4px 12px rgba(107, 114, 128, 0.15);
  margin: 8px 0;
  overflow: hidden;
}

.filter-field:has(#hidden-filter)::before {
  content: '';
  position: absolute;
  top: -2px;
  left: -2px;
  right: -2px;
  bottom: -2px;
  background: linear-gradient(45deg, #6b7280, #4b5563);
  border-radius: 12px;
  z-index: -1;
  opacity: 0.8;
}

.filter-field:has(#hidden-filter) label {
  color: #374151;
  font-weight: 700;
  font-size: 15px;
  margin-bottom: 8px;
  display: block;
}

.filter-field:has(#hidden-filter) label i {
  color: #6b7280;
  font-size: 16px;
}

.filter-field:has(#hidden-filter) select {
  border: 2px solid #6b7280;
  background-color: white;
  border-radius: 6px;
  padding: 8px 10px;
  font-weight: 500;
  width: 100%;
  box-sizing: border-box;
  font-size: 13px;
}

.filter-field:has(#hidden-filter) select:focus {
  border-color: #4b5563;
  box-shadow: 0 0 0 2px rgba(107, 114, 128, 0.2);
  outline: none;
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
