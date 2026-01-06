<script setup>
import { ref, onMounted, computed, useAttrs } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import CarAssignmentForm from './CarAssignmentForm.vue'

const { t } = useEnhancedI18n()
const attrs = useAttrs()
const emit = defineEmits(['refresh', 'assign-car'])

const { callApi } = useApi()
const unassignedCars = ref([])
const allUnassignedCars = ref([]) // Store all unassigned cars before filtering
const loading = ref(true)
const error = ref(null)
const selectedSellBillId = ref(null)
const showAssignmentForm = ref(false)
const selectedCarId = ref(null)
const isProcessing = ref(false)
const isBatchSell = ref(false)

// Filter states
const filters = ref({
  carName: '',
  color: '',
  vin: '',
  buyBillRef: '',
})
const user = ref(null)
const expandedCarNotes = ref(new Set()) // Track which cars have expanded notes

// Truncate notes display
const getTruncatedCarNotes = (notesText, carId, maxLength = 100) => {
  if (!notesText || notesText === 'N/A' || notesText === '-') return notesText
  if (expandedCarNotes.value.has(carId)) return notesText
  if (notesText.length <= maxLength) return notesText
  return notesText.substring(0, maxLength) + '...'
}

const isCarNotesTruncated = (notesText, carId, maxLength = 100) => {
  if (!notesText || notesText === 'N/A' || notesText === '-') return false
  return notesText.length > maxLength && !expandedCarNotes.value.has(carId)
}

const toggleCarNotesExpansion = (carId) => {
  if (expandedCarNotes.value.has(carId)) {
    expandedCarNotes.value.delete(carId)
  } else {
    expandedCarNotes.value.add(carId)
  }
}

const isAdmin = computed(() => user.value?.role_id === 1)

// Add sort config after the filters
const sortConfig = ref({
  field: 'id',
  direction: 'desc',
})

// Add computed property for sorted cars
const sortedCars = computed(() => {
  if (!unassignedCars.value.length) return []

  return [...unassignedCars.value].sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle date comparison
    if (sortConfig.value.field === 'date_sell') {
      aValue = aValue ? new Date(aValue).getTime() : 0
      bValue = bValue ? new Date(bValue).getTime() : 0
    }

    // Handle numeric fields
    if (['id', 'price_cell', 'buy_price'].includes(sortConfig.value.field)) {
      aValue = Number(aValue) || 0
      bValue = Number(bValue) || 0
    }

    // Handle null values
    if (aValue === null || aValue === undefined) aValue = ''
    if (bValue === null || bValue === undefined) bValue = ''

    // Compare values based on direction
    if (sortConfig.value.direction === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })
})

const handleSort = (field) => {
  if (sortConfig.value.field === field) {
    // Toggle direction if clicking the same field
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    // Set new field and default to descending
    sortConfig.value.field = field
    sortConfig.value.direction = 'desc'
  }
}

// Function to fetch all unassigned cars
const fetchUnassignedCars = async () => {
  loading.value = true
  error.value = null

  try {
    let myQuery = `SELECT 
          cs.id,
          cs.vin,
          cs.notes,
          cs.price_cell,
          cs.freight,
          cs.path_documents,
          cs.container_ref,
          cs.hidden,
          cs.hidden_time_stamp,
          cs.hidden_by_user_id,
          hu.username as hidden_by_username,
          bd.amount as buy_price,
          cn.car_name,
          clr.color,
          clr.hexa,
          bb.bill_ref as buy_bill_ref
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
        LEFT JOIN users hu ON cs.hidden_by_user_id = hu.id
        WHERE cs.id_sell IS NULL
        `
    if (!isAdmin.value) {
      myQuery += ' AND cs.hidden = 0'
    }
    myQuery += ' ORDER BY cs.id DESC'

    const result = await callApi({
      query: myQuery,
      params: [],
    })

    if (result.success) {
      // Parse notes JSON and format for display
      const carsData = (result.data || []).map((car) => {
        // Parse notes JSON if it exists
        let notesDisplay = 'N/A'
        if (car.notes) {
          try {
            let notesArray = []
            if (typeof car.notes === 'string' && car.notes.trim().startsWith('[')) {
              notesArray = JSON.parse(car.notes)
            } else if (Array.isArray(car.notes)) {
              notesArray = car.notes
            } else {
              // Old format (plain text) - use as is
              notesDisplay = car.notes
            }
            
            if (Array.isArray(notesArray) && notesArray.length > 0) {
              // Format all notes without username or timestamp
              notesDisplay = notesArray.map((note) => {
                return note.note || ''
              }).join('\n')
            }
          } catch (e) {
            // If parsing fails, treat as old format (plain text)
            notesDisplay = car.notes
          }
        }
        
        return {
          ...car,
          notesDisplay: notesDisplay, // Add formatted display version
        }
      })
      allUnassignedCars.value = carsData
      applyFilters() // Apply filters to the fetched data
    } else {
      error.value = result.error || t('sellBills.failed_to_fetch_unassigned_cars')
    }
  } catch (err) {
    error.value = err.message || t('sellBills.error_occurred')
  } finally {
    loading.value = false
  }
}

// Apply filters to the cars list
const applyFilters = () => {
  unassignedCars.value = allUnassignedCars.value.filter((car) => {
    const matchCarName =
      !filters.value.carName ||
      (car.car_name && car.car_name.toLowerCase().includes(filters.value.carName.toLowerCase()))

    const matchColor =
      !filters.value.color ||
      (car.color && car.color.toLowerCase().includes(filters.value.color.toLowerCase()))

    const matchVin =
      !filters.value.vin ||
      (car.vin && car.vin.toLowerCase().includes(filters.value.vin.toLowerCase()))

    const matchBuyBillRef =
      !filters.value.buyBillRef ||
      (car.buy_bill_ref &&
        car.buy_bill_ref.toLowerCase().includes(filters.value.buyBillRef.toLowerCase()))

    return matchCarName && matchColor && matchVin && matchBuyBillRef
  })
}

// Reset all filters
const resetFilters = () => {
  filters.value = {
    carName: '',
    color: '',
    vin: '',
    buyBillRef: '',
  }
  applyFilters()
}

// Watch for filter changes
const handleFilterChange = () => {
  applyFilters()
}

// Function to check if the selected bill is a batch sell
const checkBatchSell = async (billId) => {
  try {
    const result = await callApi({
      query: `
        SELECT is_batch_sell
        FROM sell_bill
        WHERE id = ?
      `,
      params: [billId],
    })

    if (result.success && result.data.length > 0) {
      isBatchSell.value = result.data[0].is_batch_sell === 1
    }
  } catch (err) {
    console.error('Error checking batch sell:', err)
  }
}

// Add persistent pricing for batch sell
const batchPrice = ref(null)
const batchFreight = ref(null)
const batchRate = ref(null)
const batchDischargePort = ref(null)
const showBatchPricing = ref(false)

// Add discharge ports for batch sell
const dischargePorts = ref([])

// Function to calculate text color based on background color
const getTextColor = (backgroundColor) => {
  if (!backgroundColor) return '#374151'

  // Convert hex to RGB
  const hex = backgroundColor.replace('#', '')
  const r = parseInt(hex.substr(0, 2), 16)
  const g = parseInt(hex.substr(2, 2), 16)
  const b = parseInt(hex.substr(4, 2), 16)

  // Calculate luminance
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255

  // Return white text for dark backgrounds, black text for light backgrounds
  return luminance > 0.5 ? '#000000' : '#ffffff'
}

const getHiddenBadgeText = (car) => {
  if (!car.hidden) return ''
  const user = car.hidden_by_username || t('sellBills.hidden_unknown_user')
  const timestamp = car.hidden_time_stamp ? new Date(car.hidden_time_stamp) : null

  let daysString = t('sellBills.hidden_unknown_days')

  if (timestamp && !Number.isNaN(timestamp.getTime())) {
    const diffMs = Date.now() - timestamp.getTime()
    const diffDays = Math.max(0, Math.floor(diffMs / (1000 * 60 * 60 * 24)))

    daysString =
      diffDays >= 1
        ? t('sellBills.hidden_days_count', { count: diffDays })
        : t('sellBills.hidden_less_than_day')
  }

  return t('sellBills.hidden_by_compact', { user, days: daysString })
}

// Function to fetch discharge ports
const fetchDischargePorts = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, discharge_port
        FROM discharge_ports
        ORDER BY discharge_port ASC
      `,
      params: [],
    })

    if (result.success) {
      dischargePorts.value = result.data
    }
  } catch (err) {
    console.error('Error fetching discharge ports:', err)
  }
}

// Add broker info for batch sell
const brokerInfo = ref(null)

// Function to fetch broker info
const fetchBrokerInfo = async (billId) => {
  try {
    const result = await callApi({
      query: `
        SELECT c.name as broker_name, c.mobiles as broker_phone
        FROM sell_bill sb
        LEFT JOIN clients c ON sb.id_broker = c.id
        WHERE sb.id = ?
      `,
      params: [billId],
    })

    if (result.success && result.data.length > 0) {
      brokerInfo.value = result.data[0]
    }
  } catch (err) {
    console.error('Error fetching broker info:', err)
  }
}

// Function to directly assign car for batch sell with pricing
const assignCarToBatchSell = async (carId) => {
  if (isProcessing.value) return
  isProcessing.value = true

  try {
    // Get the bill_ref and broker_id from the sell_bill
    const billResult = await callApi({
      query: `
        SELECT bill_ref, id_broker
        FROM sell_bill
        WHERE id = ?
      `,
      params: [selectedSellBillId.value],
    })

    if (!billResult.success || !billResult.data.length) {
      error.value = t('sellBills.failed_to_get_bill_reference')
      return
    }

    const billRef = billResult.data[0].bill_ref
    const brokerId = billResult.data[0].id_broker
    const currentDate = new Date().toISOString().split('T')[0]

    // Check if car is still available for assignment
    const checkResult = await callApi({
      query: `
        SELECT id, id_client
        FROM cars_stock
        WHERE id = ?
      `,
      params: [carId],
    })

    if (!checkResult.success || !checkResult.data.length) {
      error.value = 'Car not found'
      return
    }

    if (checkResult.data[0].id_client !== null) {
      error.value =
        'This car has already been assigned to another client. Please refresh and try again.'
      await fetchUnassignedCars()
      return
    }

    const result = await callApi({
      query: `
        UPDATE cars_stock
        SET id_sell = ?,
            id_client = ?,
            id_port_discharge = ?,
            price_cell = ?,
            freight = ?,
            rate = ?,
            date_sell = ?,
            id_sell_pi = ?,
            date_assigned = NOW()
        WHERE id = ? AND id_client IS NULL
      `,
      params: [
        selectedSellBillId.value,
        brokerId, // Set broker as client
        batchDischargePort.value || null,
        batchPrice.value || null,
        batchFreight.value || null,
        batchRate.value || null,
        currentDate,
        billRef,
        carId,
      ],
    })

    if (result.success) {
      if (result.affectedRows === 0) {
        error.value =
          'This car has already been assigned to another client. Please refresh and try again.'
        await fetchUnassignedCars()
      } else {
        await fetchUnassignedCars()
        emit('refresh')
      }
    } else {
      error.value = result.error || t('sellBills.failed_to_assign_car')
    }
  } catch (err) {
    error.value = err.message || t('sellBills.error_occurred')
  } finally {
    isProcessing.value = false
  }
}

// Function to toggle batch pricing visibility
const toggleBatchPricing = () => {
  showBatchPricing.value = !showBatchPricing.value
}

// Function to clear batch pricing
const clearBatchPricing = () => {
  batchPrice.value = null
  batchFreight.value = null
  batchRate.value = null
  batchDischargePort.value = null
}

// Function to set batch pricing from defaults
const setDefaultBatchPricing = async () => {
  try {
    const result = await callApi({
      query: 'SELECT freight_small, freight_big FROM defaults LIMIT 1',
      params: [],
    })

    if (result.success && result.data.length > 0) {
      const defaults = result.data[0]
      // Set default freight (you can choose small or big based on your needs)
      batchFreight.value = defaults.freight_small || null
    }
  } catch (err) {
    console.error('Error fetching default pricing:', err)
  }
}

// Open assignment form for a car
const openAssignmentForm = async (carId) => {
  if (!selectedSellBillId.value) {
    alert(t('sellBills.please_select_sell_bill_first'))
    return
  }

  if (isProcessing.value) return

  // Check if it's a batch sell
  await checkBatchSell(selectedSellBillId.value)

  if (isBatchSell.value) {
    // If it's a batch sell, assign directly without showing the form
    await assignCarToBatchSell(carId)
  } else {
    // If it's not a batch sell, show the assignment form as usual
    selectedCarId.value = carId
    showAssignmentForm.value = true
  }
}

// Handle successful assignment
const handleAssignSuccess = () => {
  showAssignmentForm.value = false
  fetchUnassignedCars()
  emit('refresh')
}

// Set the selected sell bill ID
const setSellBillId = async (billId) => {
  selectedSellBillId.value = billId
  await checkBatchSell(billId)
  if (billId) {
    await fetchBrokerInfo(billId)
  }
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
  setSellBillId,
})

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchUnassignedCars()
  fetchDischargePorts()
})
</script>

<template>
  <div class="unassigned-cars-table" v-bind="attrs" :id="attrs.id">
    <div class="table-header">
      <h3>
        <i class="fas fa-car"></i>
        {{ t('sellBills.unassigned_cars') }}
      </h3>
      <div class="total-info" v-if="unassignedCars.length > 0">
        <span>
          <i class="fas fa-hashtag"></i>
          {{ t('sellBills.total_cars') }}: {{ unassignedCars.length }}
        </span>
      </div>
    </div>
    <!-- Filters Section -->
    <div class="filters-section">
      <div class="filters-header">
        <h3>
          <i class="fas fa-filter"></i>
          {{ t('sellBills.filters') }}
        </h3>
        <button @click="resetFilters" class="reset-btn">
          <i class="fas fa-undo"></i>
          {{ t('sellBills.reset') }}
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>
            <i class="fas fa-car"></i>
            {{ t('sellBills.car_name') }}:
          </label>
          <input
            type="text"
            v-model="filters.carName"
            @input="handleFilterChange"
            :placeholder="t('sellBills.search_car_name')"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-palette"></i>
            {{ t('sellBills.color') }}:
          </label>
          <input
            type="text"
            v-model="filters.color"
            @input="handleFilterChange"
            :placeholder="t('sellBills.search_color')"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-fingerprint"></i>
            {{ t('sellBills.vin') }}:
          </label>
          <input
            type="text"
            v-model="filters.vin"
            @input="handleFilterChange"
            :placeholder="t('sellBills.search_vin')"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-file-alt"></i>
            {{ t('sellBills.buy_bill_ref') }}:
          </label>
          <input
            type="text"
            v-model="filters.buyBillRef"
            @input="handleFilterChange"
            :placeholder="t('sellBills.search_buy_bill_ref')"
          />
        </div>
      </div>
    </div>

    <!-- Batch Pricing Section (only show for batch sell) -->
    <div v-if="isBatchSell && selectedSellBillId" class="batch-pricing-section">
      <div class="batch-pricing-header">
        <h4>
          <i class="fas fa-dollar-sign"></i>
          {{ t('sellBills.batch_sell_pricing') }}
          <span
            v-if="batchPrice || batchFreight || batchRate || batchDischargePort"
            class="pricing-active-indicator"
          >
            <i class="fas fa-check-circle"></i>
            {{ t('sellBills.active') }}
          </span>
        </h4>
        <div class="batch-pricing-actions">
          <button
            @click="toggleBatchPricing"
            class="toggle-pricing-btn"
            :class="{ active: showBatchPricing }"
          >
            <i class="fas fa-cog"></i>
            {{ showBatchPricing ? t('sellBills.hide') : t('sellBills.show') }}
            {{ t('sellBills.pricing') }}
          </button>
          <button
            @click="setDefaultBatchPricing"
            class="default-btn"
            :title="t('sellBills.set_default_freight')"
          >
            <i class="fas fa-magic"></i>
            {{ t('sellBills.defaults') }}
          </button>
          <button
            @click="clearBatchPricing"
            class="clear-btn"
            :title="t('sellBills.clear_all_pricing')"
          >
            <i class="fas fa-eraser"></i>
            {{ t('sellBills.clear') }}
          </button>
        </div>
      </div>

      <!-- Broker Info -->
      <div v-if="brokerInfo" class="broker-info">
        <i class="fas fa-user-tie"></i>
        <span
          ><strong>{{ t('sellBills.broker_client') }}:</strong> {{ brokerInfo.broker_name }}</span
        >
        <span v-if="brokerInfo.broker_phone" class="broker-phone">
          <i class="fas fa-phone"></i>
          {{ brokerInfo.broker_phone }}
        </span>
      </div>

      <div v-if="showBatchPricing" class="batch-pricing-form">
        <div class="pricing-info">
          <i class="fas fa-info-circle"></i>
          <span>{{ t('sellBills.batch_pricing_info') }}</span>
        </div>

        <div class="pricing-grid">
          <div class="pricing-group">
            <label>
              <i class="fas fa-dollar-sign"></i>
              {{ t('sellBills.price_usd') }}:
            </label>
            <input
              type="number"
              v-model="batchPrice"
              :placeholder="t('sellBills.enter_price')"
              step="0.01"
              min="0"
            />
          </div>

          <div class="pricing-group">
            <label>
              <i class="fas fa-ship"></i>
              {{ t('sellBills.freight_usd') }}:
            </label>
            <input
              type="number"
              v-model="batchFreight"
              :placeholder="t('sellBills.enter_freight')"
              step="0.01"
              min="0"
            />
          </div>

          <div class="pricing-group">
            <label>
              <i class="fas fa-exchange-alt"></i>
              {{ t('sellBills.rate_da') }}:
            </label>
            <input
              type="number"
              v-model="batchRate"
              :placeholder="t('sellBills.enter_rate')"
              step="0.01"
              min="0"
            />
          </div>

          <div class="pricing-group">
            <label>
              <i class="fas fa-anchor"></i>
              {{ t('sellBills.discharge_port') }}:
            </label>
            <select v-model="batchDischargePort" class="discharge-port-select">
              <option value="">{{ t('sellBills.select_discharge_port') }}</option>
              <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
                {{ port.discharge_port }}
              </option>
            </select>
          </div>
        </div>

        <div
          class="pricing-summary"
          v-if="batchPrice || batchFreight || batchRate || batchDischargePort"
        >
          <div class="summary-item">
            <span class="label">{{ t('sellBills.price') }}:</span>
            <span class="value">{{
              batchPrice ? '$' + parseFloat(batchPrice).toLocaleString() : t('sellBills.not_set')
            }}</span>
          </div>
          <div class="summary-item">
            <span class="label">{{ t('sellBills.freight') }}:</span>
            <span class="value">{{
              batchFreight
                ? '$' + parseFloat(batchFreight).toLocaleString()
                : t('sellBills.not_set')
            }}</span>
          </div>
          <div class="summary-item">
            <span class="label">{{ t('sellBills.rate') }}:</span>
            <span class="value">{{
              batchRate ? parseFloat(batchRate).toLocaleString() + ' DA' : t('sellBills.not_set')
            }}</span>
          </div>
          <div class="summary-item">
            <span class="label">{{ t('sellBills.discharge_port') }}:</span>
            <span class="value">{{
              batchDischargePort
                ? dischargePorts.find((p) => p.id === batchDischargePort)?.discharge_port ||
                  t('sellBills.selected')
                : t('sellBills.not_set')
            }}</span>
          </div>
          <div class="summary-item total" v-if="batchPrice && batchRate">
            <span class="label">{{ t('sellBills.total_da') }}:</span>
            <span class="value"
              >{{
                (
                  (parseFloat(batchPrice) + (parseFloat(batchFreight) || 0)) *
                  parseFloat(batchRate)
                ).toLocaleString()
              }}
              DA</span
            >
          </div>
        </div>
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>{{ t('sellBills.loading_unassigned_cars') }}</span>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-else-if="unassignedCars.length === 0" class="no-data">
      <i class="fas fa-car fa-2x"></i>
      <p>{{ t('sellBills.no_unassigned_cars_available') }}</p>
    </div>

    <table v-else class="cars-table">
      <thead>
        <tr>
          <th @click="handleSort('id')" class="sortable">
            <i class="fas fa-hashtag"></i> {{ t('sellBills.id') }}
            <i
              v-if="sortConfig.field === 'id'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('car_name')" class="sortable">
            <i class="fas fa-car"></i> {{ t('sellBills.car') }}
            <i
              v-if="sortConfig.field === 'car_name'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('color')" class="sortable">
            <i class="fas fa-palette"></i> {{ t('sellBills.color') }}
            <i
              v-if="sortConfig.field === 'color'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('vin')" class="sortable">
            <i class="fas fa-fingerprint"></i> {{ t('sellBills.vin') }} /
            {{ t('sellBills.container_ref') }}
            <i
              v-if="sortConfig.field === 'vin'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('price_cell')" class="sortable">
            <i class="fas fa-dollar-sign"></i> {{ t('sellBills.sell_price') }}
            <i
              v-if="sortConfig.field === 'price_cell'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th v-if="isAdmin" @click="handleSort('buy_price')" class="sortable">
            <i class="fas fa-tags"></i> {{ t('sellBills.buy_price') }}
            <i
              v-if="sortConfig.field === 'buy_price'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('buy_bill_ref')" class="sortable">
            <i class="fas fa-file-alt"></i> {{ t('sellBills.buy_bill_ref') }}
            <i
              v-if="sortConfig.field === 'buy_bill_ref'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th @click="handleSort('notes')" class="sortable">
            <i class="fas fa-sticky-note"></i> {{ t('sellBills.notes') }}
            <i
              v-if="sortConfig.field === 'notes'"
              :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
            ></i>
          </th>
          <th><i class="fas fa-cog"></i> {{ t('sellBills.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in sortedCars" :key="car.id">
          <td>{{ car.id }}</td>
          <td>
            <div class="car-name-cell">
              <span>{{ car.car_name }}</span>
              <span v-if="car.hidden" class="badge hidden-badge">
                <i class="fas fa-eye-slash"></i>
                <span class="hidden-info-text">{{ getHiddenBadgeText(car) }}</span>
              </span>
            </div>
          </td>
          <td>
            <div class="color-cell">
              <span
                v-if="car.color"
                class="badge color-badge"
                :style="{
                  backgroundColor: car.hexa || '#000000',
                  color: getTextColor(car.hexa || '#000000'),
                }"
              >
                {{ car.color }}
              </span>
            </div>
          </td>
          <td>
            <div class="vin-container-cell">
              <span v-if="car.vin" class="badge vin-badge">
                {{ car.vin }}
              </span>
              <span v-if="car.container_ref" class="badge container-badge">
                {{ car.container_ref }}
              </span>
            </div>
          </td>
          <td>{{ car.price_cell ? '$' + car.price_cell.toLocaleString() : 'N/A' }}</td>
          <td v-if="isAdmin">{{ car.buy_price ? '$' + car.buy_price.toLocaleString() : 'N/A' }}</td>
          <td>{{ car.buy_bill_ref || 'N/A' }}</td>
          <td class="notes-cell">
            <div class="notes-content">
              <div style="white-space: pre-line; max-width: 300px">
                {{ getTruncatedCarNotes(car.notesDisplay, car.id) }}
              </div>
              <button
                v-if="isCarNotesTruncated(car.notesDisplay, car.id)"
                @click.stop="toggleCarNotesExpansion(car.id)"
                class="btn-show-full-notes"
                type="button"
              >
                {{
                  expandedCarNotes.has(car.id)
                    ? t('sellBills.show_less') || 'Show Less'
                    : t('sellBills.show_full') || 'Show Full'
                }}
              </button>
            </div>
          </td>
          <td class="actions">
            <button
              @click="openAssignmentForm(car.id)"
              class="assign-btn"
              :disabled="isProcessing"
              :class="{ processing: isProcessing }"
            >
              <i class="fas fa-link"></i>
              <span>
                {{
                  isBatchSell
                    ? batchPrice
                      ? `${t('sellBills.quick_assign')} ($${parseFloat(batchPrice).toLocaleString()})`
                      : t('sellBills.quick_assign')
                    : t('sellBills.assign')
                }}
                {{ isBatchSell && brokerInfo ? ` → ${brokerInfo.broker_name}` : '' }}
              </span>
              <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Assignment Form Dialog -->
    <div v-if="showAssignmentForm && !isBatchSell" class="dialog-overlay">
      <CarAssignmentForm
        :visible="showAssignmentForm"
        :carId="selectedCarId"
        :sellBillId="selectedSellBillId"
        @close="showAssignmentForm = false"
        @assign-success="handleAssignSuccess"
      />
    </div>
  </div>
</template>

<style scoped>
.unassigned-cars-table {
  position: relative;
  margin: 1.5rem 0;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  z-index: 10;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.table-header h3 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1f2937;
  margin: 0;
}

.total-info {
  display: flex;
  gap: 1.5rem;
}

.total-info span {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  font-weight: 500;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.no-data {
  text-align: center;
  padding: 2rem;
  background-color: #f9fafb;
  border-radius: 0.5rem;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.cars-table {
  width: 100%;
  border-collapse: collapse;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.cars-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #f3f4f6;
}

.cars-table th,
.cars-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.cars-table th {
  background-color: #f3f4f6;
  font-weight: 600;
  color: #374151;
}

.cars-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.cars-table tr:hover {
  background-color: #f9fafb;
}

.cars-table td i {
  margin-right: 0.5rem;
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.badge {
  padding: 4px 10px;
  font-size: 0.95em;
  font-weight: 500;
  display: inline-block;
  border-radius: 4px;
  border: 1px solid #d1d5db;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  width: 100%;
  text-align: center;
}

.color-badge {
  background-color: #fef2f2;
  color: #dc2626;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 500;
  display: inline-block;
  width: 90%;
  text-align: center;
  box-shadow: 0 2px 4px rgba(220, 38, 38, 0.2);
  transition: all 0.2s ease;
}

.color-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(220, 38, 38, 0.3);
}

.vin-container-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
  padding: 8px;
}

.color-cell {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  padding: 8px;
}

.vin-badge {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
  color: white;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 500;
  display: inline-block;
  width: 90%;
  text-align: center;
  box-shadow: 0 2px 4px rgba(139, 92, 246, 0.3);
  transition: all 0.2s ease;
}

.vin-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(139, 92, 246, 0.4);
}

.vin-badge.no-data {
  background: linear-gradient(135deg, #6b7280, #4b5563);
  box-shadow: 0 2px 4px rgba(107, 114, 128, 0.3);
}

.container-badge {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: white;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 500;
  display: inline-block;
  width: 90%;
  text-align: center;
  box-shadow: 0 2px 4px rgba(245, 158, 11, 0.3);
  transition: all 0.2s ease;
}

.container-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(245, 158, 11, 0.4);
}

.container-badge.no-data {
  background: linear-gradient(135deg, #6b7280, #4b5563);
  box-shadow: 0 2px 4px rgba(107, 114, 128, 0.3);
}

.assign-btn {
  background-color: #10b981;
  color: white;
}

.assign-btn:hover:not(:disabled) {
  background-color: #059669;
}

.car-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.hidden-badge {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  padding: 4px 8px;
  width: auto;
}

.hidden-info-text {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  white-space: normal;
  font-size: 0.85rem;
}

/* Add smooth transitions */
.cars-table tr {
  transition: background-color 0.2s;
}

.btn i {
  transition: transform 0.2s;
}

.btn:hover:not(:disabled) i {
  transform: scale(1.1);
}

/* Add filter styles */
.filters-section {
  background-color: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filters-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #1f2937;
  font-size: 1.1rem;
}

.reset-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #e5e7eb;
  border: none;
  border-radius: 4px;
  color: #4b5563;
  cursor: pointer;
  transition: all 0.2s;
}

.reset-btn:hover {
  background-color: #d1d5db;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 0.9rem;
}

.filter-group input {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.filter-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

@media (max-width: 768px) {
  .filters-grid {
    grid-template-columns: 1fr;
  }
}

/* Add sorting styles */
.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 24px; /* Space for sort icon */
}

.sortable i:last-child {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 0.75rem;
  opacity: 0.5;
}

.sortable:hover {
  background-color: #f1f5f9;
}

.sortable:hover i:last-child {
  opacity: 1;
}

.notes-cell {
  max-width: 300px;
  white-space: pre-line;
  word-wrap: break-word;
  vertical-align: top;
  padding: 8px;
}

.notes-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.btn-show-full-notes {
  background: #007bff;
  color: white;
  border: none;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85em;
  align-self: flex-start;
  margin-top: 4px;
  transition: background-color 0.2s;
}

.btn-show-full-notes:hover {
  background: #0056b3;
}

.btn-show-full-notes:active {
  transform: scale(0.98);
}

/* Add batch pricing styles */
.batch-pricing-section {
  background-color: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.batch-pricing-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.batch-pricing-header h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #1f2937;
  font-size: 1.1rem;
}

.pricing-active-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  background-color: #10b981;
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  margin-left: 8px;
}

.pricing-active-indicator i {
  font-size: 0.7rem;
}

.batch-pricing-actions {
  display: flex;
  gap: 8px;
}

.toggle-pricing-btn {
  background-color: #10b981;
  color: white;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-pricing-btn:hover {
  background-color: #059669;
}

.toggle-pricing-btn.active {
  background-color: #059669;
}

.default-btn {
  background-color: #e5e7eb;
  color: #4b5563;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.default-btn:hover {
  background-color: #d1d5db;
}

.clear-btn {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-btn:hover {
  background-color: #f3e8e8;
}

.batch-pricing-form {
  margin-top: 16px;
}

.pricing-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #dbeafe;
  color: #1e40af;
  border-radius: 4px;
  margin-bottom: 16px;
  font-size: 0.9rem;
}

.pricing-info i {
  color: #3b82f6;
}

.pricing-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.pricing-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.pricing-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 0.9rem;
}

.pricing-group input {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.pricing-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.discharge-port-select {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.95rem;
  transition: all 0.2s;
  background-color: white;
}

.discharge-port-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.pricing-summary {
  margin-top: 16px;
  padding: 12px;
  background-color: #f3f4f6;
  border-radius: 4px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.summary-item .label {
  font-weight: 600;
}

.summary-item .value {
  font-weight: 500;
}

.broker-info {
  margin-top: 16px;
  padding: 12px;
  background-color: #f3f4f6;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.broker-info i {
  color: #3b82f6;
}

.broker-info span {
  display: flex;
  align-items: center;
  gap: 6px;
}

.broker-phone {
  font-size: 0.9rem;
  color: #6b7280;
  margin-left: auto;
}
</style>
