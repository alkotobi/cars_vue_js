<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import { useRouter } from 'vue-router'
import SellBillPrintOption from './SellBillPrintOption.vue'

const { t } = useEnhancedI18n()
const props = defineProps({
  onEdit: Function,
  onDelete: Function,
  onSelect: Function,
  onTask: Function,
  selectedBillId: {
    type: Number,
    default: null,
  },
})
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)
const can_c_all_bills = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_c_other_users_sells')
})
const emit = defineEmits(['refresh', 'select-bill'])

const router = useRouter()
const { callApi } = useApi()
const sellBills = ref([])
const loading = ref(true)
const error = ref(null)
const selectedBillId = ref(null)
const showPrintOptions = ref(false)
const selectedPrintBillId = ref(null)
const isProcessing = ref(false)

// Add filter states
const filters = ref({
  dateFrom: '',
  dateTo: '',
  broker: '',
  reference: '',
  isBatchSell: null,
  createdBy: '',
  paymentStatus: '',
  loadingStatus: '',
})

// Add allSellBills to store unfiltered data
const allSellBills = ref([])

// Add computed property for payments permission
const can_c_sell_payments = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_c_sell_payments')
})

// Add computed property for delete permission
const can_delete_sell_bill = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_delete_sell_bill')
})

// Add computed property for edit permission
const can_edit_sell_bill = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_sell_bill')
})

const sortConfig = ref({
  field: 'id',
  direction: 'desc',
})

// Add computed property for sorted bills
const sortedBills = computed(() => {
  if (!sellBills.value.length) return []

  return [...sellBills.value].sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle date comparison
    if (sortConfig.value.field === 'date_sell') {
      aValue = new Date(aValue).getTime()
      bValue = new Date(bValue).getTime()
    }

    // Handle numeric fields
    if (sortConfig.value.field === 'id') {
      aValue = Number(aValue)
      bValue = Number(bValue)
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

// Add computed property for sorted and limited bills
const sortedAndLimitedBills = computed(() => {
  if (!sellBills.value.length) return []

  const sorted = [...sellBills.value].sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle date comparison
    if (sortConfig.value.field === 'date_sell') {
      aValue = new Date(aValue).getTime()
      bValue = new Date(bValue).getTime()
    }

    // Handle numeric fields
    if (sortConfig.value.field === 'id') {
      aValue = Number(aValue)
      bValue = Number(bValue)
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

  return sorted
})

// Add computed property for unpaid bills count for current user
const unpaidBillsCount = computed(() => {
  if (!allSellBills.value.length) return 0

  return allSellBills.value.filter((bill) => {
    // Admin counts all unpaid bills, regular users count only their own
    if (!isAdmin.value && bill.id_user !== user.value?.id) {
      return false
    }

    // Exclude batch sell bills
    if (bill.is_batch_sell) {
      return false
    }

    // Check if bill is unpaid (total_paid < total_cfr)
    return bill.total_paid < bill.total_cfr
  }).length
})

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchSellBills()
})

const fetchSellBills = async () => {
  loading.value = true
  error.value = null

  try {
    // Different query based on admin status - now includes payment and loading information
    const query = can_c_all_bills.value
      ? `
      SELECT 
        sb.*,
        c.name as broker_name,
        u.username as created_by,
        (
          SELECT GROUP_CONCAT(DISTINCT cl.name SEPARATOR ', ')
          FROM cars_stock cs
          LEFT JOIN clients cl ON cs.id_client = cl.id
          WHERE cs.id_sell = sb.id AND cl.name IS NOT NULL
        ) as client_names,
        (
          SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cfr,
        (
          SELECT COALESCE(SUM(sp.amount_usd), 0)
          FROM sell_payments sp
          WHERE sp.id_sell_bill = sb.id
        ) as total_paid,
        (
          SELECT COUNT(*)
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cars,
        (
          SELECT COUNT(*)
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id AND cs.container_ref IS NOT NULL
        ) as loaded_cars
      FROM sell_bill sb
      LEFT JOIN clients c ON sb.id_broker = c.id
      LEFT JOIN users u ON sb.id_user = u.id
      ORDER BY sb.date_sell DESC
    `
      : `
      SELECT 
        sb.*,
        c.name as broker_name,
        u.username as created_by,
        (
          SELECT GROUP_CONCAT(DISTINCT cl.name SEPARATOR ', ')
          FROM cars_stock cs
          LEFT JOIN clients cl ON cs.id_client = cl.id
          WHERE cs.id_sell = sb.id AND cl.name IS NOT NULL
        ) as client_names,
        (
          SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cfr,
        (
          SELECT COALESCE(SUM(sp.amount_usd), 0)
          FROM sell_payments sp
          WHERE sp.id_sell_bill = sb.id
        ) as total_paid,
        (
          SELECT COUNT(*)
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cars,
        (
          SELECT COUNT(*)
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id AND cs.container_ref IS NOT NULL
        ) as loaded_cars
      FROM sell_bill sb
      LEFT JOIN clients c ON sb.id_broker = c.id
      LEFT JOIN users u ON sb.id_user = u.id
      WHERE sb.id_user = ?
      ORDER BY sb.date_sell DESC
    `

    const params = can_c_all_bills.value ? [] : [user.value?.id]
    const result = await callApi({ query, params })

    if (result.success) {
      console.log(result.data.length)
      allSellBills.value = result.data.map((bill) => ({
        ...bill,
        total_cfr: Number(bill.total_cfr) || 0,
        total_paid: Number(bill.total_paid) || 0,
        total_cars: Number(bill.total_cars) || 0,
        loaded_cars: Number(bill.loaded_cars) || 0,
      }))
      applyFilters() // Apply filters after fetching
    } else {
      error.value = result.error || 'Failed to fetch sell bills'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Add filter functions
const applyFilters = () => {
  sellBills.value = allSellBills.value.filter((bill) => {
    // Date range filter
    if (filters.value.dateFrom && new Date(bill.date_sell) < new Date(filters.value.dateFrom)) {
      return false
    }
    if (filters.value.dateTo && new Date(bill.date_sell) > new Date(filters.value.dateTo)) {
      return false
    }

    // NAME filter - searches both broker and client names
    if (filters.value.broker) {
      const searchTerm = filters.value.broker.toLowerCase()

      // Check if broker name matches
      const brokerMatch = bill.broker_name && bill.broker_name.toLowerCase().includes(searchTerm)

      // Check if any client name matches
      const clientMatch = bill.client_names && bill.client_names.toLowerCase().includes(searchTerm)

      // Bill must match either broker OR at least one client
      if (!brokerMatch && !clientMatch) {
        return false
      }
    }

    // Reference filter
    if (
      filters.value.reference &&
      (!bill.bill_ref ||
        !bill.bill_ref.toLowerCase().includes(filters.value.reference.toLowerCase()))
    ) {
      return false
    }

    // Batch sell filter
    if (filters.value.isBatchSell !== null) {
      if (filters.value.isBatchSell && !bill.is_batch_sell) return false
      if (!filters.value.isBatchSell && bill.is_batch_sell) return false
    }

    // Created by filter
    if (
      filters.value.createdBy &&
      (!bill.created_by ||
        !bill.created_by.toLowerCase().includes(filters.value.createdBy.toLowerCase()))
    ) {
      return false
    }

    // Payment status filter
    if (filters.value.paymentStatus) {
      const isPaid = bill.total_paid >= bill.total_cfr
      const isPartiallyPaid = bill.total_paid > 0 && bill.total_paid < bill.total_cfr
      const isUnpaid = bill.total_paid === 0

      if (filters.value.paymentStatus === 'paid' && !isPaid) return false
      if (filters.value.paymentStatus === 'partially_paid' && !isPartiallyPaid) return false
      if (filters.value.paymentStatus === 'unpaid' && !isUnpaid) return false
    }

    // Loading status filter
    if (filters.value.loadingStatus) {
      const isFullyLoaded = bill.loaded_cars === bill.total_cars && bill.total_cars > 0
      const isPartiallyLoaded = bill.loaded_cars > 0 && bill.loaded_cars < bill.total_cars
      const isNotLoaded = bill.loaded_cars === 0

      if (filters.value.loadingStatus === 'fully_loaded' && !isFullyLoaded) return false
      if (filters.value.loadingStatus === 'partially_loaded' && !isPartiallyLoaded) return false
      if (filters.value.loadingStatus === 'not_loaded' && !isNotLoaded) return false
    }

    return true
  })
}

// Watch for filter changes
watch(
  filters,
  () => {
    applyFilters()
  },
  { deep: true },
)

const resetFilters = () => {
  filters.value = {
    dateFrom: '',
    dateTo: '',
    broker: '',
    reference: '',
    isBatchSell: null,
    createdBy: '',
    paymentStatus: '',
    loadingStatus: '',
  }
}

const handleEdit = (bill) => {
  if (props.onEdit) {
    props.onEdit(bill)
  }
}

const handleDelete = (id) => {
  if (props.onDelete) {
    props.onDelete(id)
  }
}

const handlePrint = (id) => {
  selectedPrintBillId.value = id
  showPrintOptions.value = true
}

const handlePrintClose = () => {
  showPrintOptions.value = false
  selectedPrintBillId.value = null
}

const handlePrintProceed = (options) => {
  console.log('Print options:', options)
  showPrintOptions.value = false
  selectedPrintBillId.value = null
}

const handlePayments = (billId) => {
  // Open payments in a new tab using router.resolve
  const route = router.resolve({
    name: 'sell-bill-payments',
    params: { id: billId },
  })
  window.open(route.href, '_blank')
}

const handleTask = (bill) => {
  if (props.onTask) {
    props.onTask(bill)
  }
}

const selectBill = (bill) => {
  console.log('selectBill called with bill:', bill)
  selectedBillId.value = bill.id

  // Emit the selected bill ID to the parent component
  console.log('Emitting select-bill event with billId:', bill.id)
  emit('select-bill', bill.id)

  if (props.onSelect) {
    console.log('Calling props.onSelect with billId:', bill.id)
    props.onSelect(bill.id)
  } else {
    console.log('props.onSelect is not available')
  }
}

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

// Watch for changes in the selectedBillId prop
watch(
  () => props.selectedBillId,
  (newId) => {
    selectedBillId.value = newId
  },
  { immediate: true },
)

// Expose the fetchSellBills method and unpaidBillsCount to parent component
defineExpose({ fetchSellBills, unpaidBillsCount })

// Add computed properties for payment status
const getPaymentStatus = (bill) => {
  const total = bill.total_cfr || 0
  const paid = bill.total_paid || 0
  const remaining = total - paid

  if (total === 0) {
    return { status: 'no-amount', text: t('sellBills.no_amount'), color: 'gray' }
  } else if (paid === 0) {
    return { status: 'not-paid', text: t('sellBills.not_paid'), color: 'red' }
  } else if (remaining <= 0) {
    return { status: 'paid', text: t('sellBills.paid'), color: 'green' }
  } else {
    return {
      status: 'partial',
      text: `${t('sellBills.left')}: $${remaining.toFixed(2)}`,
      color: 'orange',
    }
  }
}

// Add computed properties for loading status
const getLoadingStatus = (bill) => {
  const totalCars = bill.total_cars || 0
  const loadedCars = bill.loaded_cars || 0

  if (totalCars === 0) {
    return { status: 'no-cars', text: t('sellBills.no_vehicles'), color: 'gray' }
  } else if (loadedCars === 0) {
    return { status: 'not-loaded', text: t('sellBills.awaiting_loading'), color: 'red' }
  } else if (loadedCars === totalCars) {
    return { status: 'fully-loaded', text: t('sellBills.fully_loaded'), color: 'green' }
  } else {
    return {
      status: 'partially-loaded',
      text: `${loadedCars}/${totalCars} ${t('sellBills.loaded')}`,
      color: 'orange',
    }
  }
}
</script>

<template>
  <div class="sell-bills-table-component">
    <!-- Filters Section -->
    <div class="filters-section">
      <div class="filters-header">
        <div class="filters-title-section">
          <h3>
            <i class="fas fa-filter"></i>
            {{ t('sellBills.filters') }}
          </h3>
          <div class="unpaid-bills-info">
            <span class="unpaid-badge">
              <i class="fas fa-exclamation-triangle"></i>
              {{ t('sellBills.unpaid_bills_count', { count: unpaidBillsCount }) }}
            </span>
          </div>
        </div>
        <button @click="resetFilters" class="reset-btn">
          <i class="fas fa-undo"></i>
          {{ t('sellBills.reset') }}
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>
            <i class="fas fa-calendar"></i>
            {{ t('sellBills.date_from') }}
          </label>
          <input type="date" v-model="filters.dateFrom" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-calendar"></i>
            {{ t('sellBills.date_to') }}
          </label>
          <input type="date" v-model="filters.dateTo" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-user-tie"></i>
            {{ t('sellBills.name') }}
          </label>
          <input type="text" v-model="filters.broker" :placeholder="t('sellBills.search_name')" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-barcode"></i>
            {{ t('sellBills.reference') }}
          </label>
          <input
            type="text"
            v-model="filters.reference"
            :placeholder="t('sellBills.search_reference')"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-layer-group"></i>
            {{ t('sellBills.batch_sell') }}
          </label>
          <select v-model="filters.isBatchSell">
            <option :value="null">{{ t('sellBills.all') }}</option>
            <option :value="true">{{ t('sellBills.yes') }}</option>
            <option :value="false">{{ t('sellBills.no') }}</option>
          </select>
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-user"></i>
            {{ t('sellBills.created_by') }}
          </label>
          <input
            type="text"
            v-model="filters.createdBy"
            :placeholder="t('sellBills.search_created_by')"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-credit-card"></i>
            {{ t('sellBills.payment_status') }}
          </label>
          <select v-model="filters.paymentStatus">
            <option value="">{{ t('sellBills.all') }}</option>
            <option value="paid">{{ t('sellBills.paid') }}</option>
            <option value="partially_paid">{{ t('sellBills.partially_paid') }}</option>
            <option value="unpaid">{{ t('sellBills.unpaid') }}</option>
          </select>
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-shipping-fast"></i>
            {{ t('sellBills.loading_status') }}
          </label>
          <select v-model="filters.loadingStatus">
            <option value="">{{ t('sellBills.all') }}</option>
            <option value="fully_loaded">{{ t('sellBills.fully_loaded') }}</option>
            <option value="partially_loaded">{{ t('sellBills.partially_loaded') }}</option>
            <option value="not_loaded">{{ t('sellBills.not_loaded') }}</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>{{ t('sellBills.loading_bills') }}</span>
    </div>

    <div v-if="error" class="error">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-else-if="sortedAndLimitedBills.length === 0" class="no-data">
      <i class="fas fa-inbox fa-2x"></i>
      <p>{{ t('sellBills.no_bills_found') }}</p>
    </div>

    <div v-else class="table-container">
      <table class="sell-bills-table">
        <thead>
          <tr>
            <th @click="handleSort('id')" class="sortable">
              <i class="fas fa-hashtag"></i> {{ t('sellBills.id') }}
              <i
                v-if="sortConfig.field === 'id'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('bill_ref')" class="sortable">
              <i class="fas fa-barcode"></i> {{ t('sellBills.reference') }}
              <i
                v-if="sortConfig.field === 'bill_ref'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('date_sell')" class="sortable">
              <i class="fas fa-calendar"></i> {{ t('sellBills.date') }}
              <i
                v-if="sortConfig.field === 'date_sell'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('broker_name')" class="sortable">
              <i class="fas fa-user-tie"></i> {{ t('sellBills.name') }}
              <i
                v-if="sortConfig.field === 'broker_name'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('created_by')" class="sortable">
              <i class="fas fa-user"></i> {{ t('sellBills.created_by') }}
              <i
                v-if="sortConfig.field === 'created_by'"
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
            <th><i class="fas fa-money-bill-wave"></i> {{ t('sellBills.payment_status') }}</th>
            <th><i class="fas fa-shipping-fast"></i> {{ t('sellBills.loading_status') }}</th>
            <th><i class="fas fa-cog"></i> {{ t('sellBills.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="bill in sortedAndLimitedBills"
            :key="bill.id"
            @click="selectBill(bill)"
            :class="{ selected: selectedBillId === bill.id }"
          >
            <td>{{ bill.id }}</td>
            <td>{{ bill.bill_ref || 'N/A' }}</td>
            <td>
              <div class="date-badges">
                <span class="badge sell-date-badge">
                  <i class="fas fa-calendar"></i>
                  {{ t('sellBills.sell') }} {{ new Date(bill.date_sell).toLocaleDateString() }}
                </span>
                <span v-if="isAdmin" class="badge created-date-badge">
                  <i class="fas fa-clock"></i>
                  {{ t('sellBills.created') }}
                  {{ bill.time_created ? new Date(bill.time_created).toLocaleDateString() : 'N/A' }}
                </span>
              </div>
            </td>
            <td>{{ bill.broker_name || 'N/A' }}</td>
            <td>{{ bill.created_by || 'N/A' }}</td>
            <td>{{ bill.notes || 'N/A' }}</td>
            <td>
              <span
                :class="['payment-badge', `payment-${getPaymentStatus(bill).status}`]"
                :title="`${t('sellBills.total')} $${bill.total_cfr.toFixed(2)} | ${t('sellBills.paid_amount')} $${bill.total_paid.toFixed(2)}`"
              >
                {{ getPaymentStatus(bill).text }}
              </span>
            </td>
            <td>
              <span
                :class="['loading-badge', `loading-${getLoadingStatus(bill).status}`]"
                :title="`${bill.loaded_cars} ${t('sellBills.vehicles_loaded', { total: bill.total_cars })}`"
              >
                {{ getLoadingStatus(bill).text }}
              </span>
            </td>
            <td class="actions">
              <button
                v-if="can_edit_sell_bill"
                @click.stop="handleEdit(bill)"
                :disabled="isProcessing"
                class="btn edit-btn"
                :title="t('sellBills.edit_bill')"
              >
                <i class="fas fa-edit"></i>
              </button>
              <button
                v-if="can_delete_sell_bill"
                @click.stop="handleDelete(bill.id)"
                :disabled="isProcessing"
                class="btn delete-btn"
                :title="t('sellBills.delete_bill')"
              >
                <i class="fas fa-trash-alt"></i>
              </button>
              <button
                @click.stop="handlePrint(bill.id)"
                :disabled="isProcessing"
                class="btn print-btn"
                :title="t('sellBills.print_bill')"
              >
                <i class="fas fa-print"></i>
              </button>
              <button
                v-if="can_c_sell_payments"
                @click.stop="handlePayments(bill.id)"
                :disabled="isProcessing"
                class="btn payment-btn"
                :title="t('sellBills.view_payments')"
              >
                <i class="fas fa-money-bill-wave"></i>
              </button>
              <button
                @click.stop="handleTask(bill)"
                :disabled="isProcessing"
                class="btn task-btn"
                :title="t('sellBills.add_new_task')"
              >
                <i class="fas fa-tasks"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <SellBillPrintOption
      :visible="showPrintOptions"
      :billId="selectedPrintBillId"
      @close="handlePrintClose"
      @proceed="handlePrintProceed"
    />
  </div>
</template>

<style scoped>
.sell-bills-table-component {
  position: relative;
  margin-bottom: 20px;
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

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f8f9fa;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.selected-bill-info {
  font-size: 1rem;
  color: #374151;
}

.bill-id {
  font-weight: 600;
  color: #3b82f6;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
}

.sell-bills-table {
  width: 100%;
  border-collapse: collapse;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.sell-bills-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #f3f4f6;
}

.sell-bills-table th,
.sell-bills-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.sell-bills-table th {
  background-color: #f3f4f6;
  font-weight: 600;
  color: #374151;
}

.sell-bills-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.sell-bills-table tr:hover {
  background-color: #f9fafb;
  cursor: pointer;
}

.sell-bills-table tr.selected {
  background-color: #e5edff;
}

.loading,
.error,
.no-data {
  padding: 2rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  color: #6b7280;
}

.error {
  color: #dc2626;
  background-color: #fee2e2;
  border-radius: 8px;
  padding: 1rem;
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

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.print-btn {
  background-color: #10b981;
  color: white;
}

.print-btn:hover:not(:disabled) {
  background-color: #059669;
}

.payment-btn {
  background-color: #10b981;
  color: white;
}

.payment-btn:hover:not(:disabled) {
  background-color: #059669;
}

.task-btn {
  background-color: #8b5cf6;
  color: white;
}

.task-btn:hover:not(:disabled) {
  background-color: #7c3aed;
}

/* Add smooth transitions */
.sell-bills-table tr {
  transition: background-color 0.2s;
}

.btn i {
  transition: transform 0.2s;
}

.btn:hover:not(:disabled) i {
  transform: scale(1.1);
}

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

.filters-title-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.filters-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #1f2937;
  font-size: 1.1rem;
}

.unpaid-bills-info {
  display: flex;
  align-items: center;
}

.unpaid-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 6px;
  color: #92400e;
  font-size: 0.9rem;
  font-weight: 500;
}

.unpaid-badge i {
  color: #f59e0b;
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

.filter-group input,
.filter-group select {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.filter-group input:focus,
.filter-group select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.filter-group select {
  cursor: pointer;
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

.table-container {
  max-height: 400px; /* Limit the height */
  overflow-y: auto; /* Make it scrollable */
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

/* Payment status badges */
.payment-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  text-align: center;
  min-width: 80px;
  cursor: help;
}

.payment-paid {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.payment-not-paid {
  background-color: #fee2e2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.payment-partial {
  background-color: #fed7aa;
  color: #92400e;
  border: 1px solid #fdba74;
}

.payment-no-amount {
  background-color: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
}

/* Loading status badges */
.loading-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  text-align: center;
  min-width: 80px;
  cursor: help;
}

.loading-fully-loaded {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.loading-not-loaded {
  background-color: #fee2e2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.loading-partially-loaded {
  background-color: #fed7aa;
  color: #92400e;
  border: 1px solid #fdba74;
}

.loading-no-cars {
  background-color: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
}

/* New styles for date badges */
.date-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.sell-date-badge,
.created-date-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  font-family: inherit;
  text-align: center;
  width: 100%;
  box-sizing: border-box;
  justify-content: center;
}

.sell-date-badge {
  background-color: #eff6ff;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}

.created-date-badge {
  background-color: #f3e8ff;
  color: #6b21a8;
  border: 1px solid #d9b8ff;
}
</style>
