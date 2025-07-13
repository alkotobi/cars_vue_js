<script setup>
import { ref, onMounted, defineProps, defineEmits, computed, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import { useRouter } from 'vue-router'
import SellBillPrintOption from './SellBillPrintOption.vue'

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
    // Different query based on admin status - now includes payment information
    const query = isAdmin.value
      ? `
      SELECT 
        sb.*,
        c.name as broker_name,
        u.username as created_by,
        (
          SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cfr,
        (
          SELECT COALESCE(SUM(sp.amount_usd), 0)
          FROM sell_payments sp
          WHERE sp.id_sell_bill = sb.id
        ) as total_paid
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
          SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cfr,
        (
          SELECT COALESCE(SUM(sp.amount_usd), 0)
          FROM sell_payments sp
          WHERE sp.id_sell_bill = sb.id
        ) as total_paid
      FROM sell_bill sb
      LEFT JOIN clients c ON sb.id_broker = c.id
      LEFT JOIN users u ON sb.id_user = u.id
      WHERE sb.id_user = ?
      ORDER BY sb.date_sell DESC
    `

    const params = isAdmin.value ? [] : [user.value?.id]
    const result = await callApi({ query, params })

    if (result.success) {
      console.log(result.data.length)
      allSellBills.value = result.data.map((bill) => ({
        ...bill,
        total_cfr: Number(bill.total_cfr) || 0,
        total_paid: Number(bill.total_paid) || 0,
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

    // Broker filter
    if (
      filters.value.broker &&
      (!bill.broker_name ||
        !bill.broker_name.toLowerCase().includes(filters.value.broker.toLowerCase()))
    ) {
      return false
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

// Expose the fetchSellBills method to parent component
defineExpose({ fetchSellBills })

// Add computed properties for payment status
const getPaymentStatus = (bill) => {
  const total = bill.total_cfr || 0
  const paid = bill.total_paid || 0
  const remaining = total - paid

  if (total === 0) {
    return { status: 'no-amount', text: 'No Amount', color: 'gray' }
  } else if (paid === 0) {
    return { status: 'not-paid', text: 'Not Paid', color: 'red' }
  } else if (remaining <= 0) {
    return { status: 'paid', text: 'Paid', color: 'green' }
  } else {
    return {
      status: 'partial',
      text: `Left: $${remaining.toFixed(2)}`,
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
        <h3>
          <i class="fas fa-filter"></i>
          Filters
        </h3>
        <button @click="resetFilters" class="reset-btn">
          <i class="fas fa-undo"></i>
          Reset
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>
            <i class="fas fa-calendar"></i>
            Date From:
          </label>
          <input type="date" v-model="filters.dateFrom" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-calendar"></i>
            Date To:
          </label>
          <input type="date" v-model="filters.dateTo" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-user-tie"></i>
            Broker:
          </label>
          <input type="text" v-model="filters.broker" placeholder="Search broker..." />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-barcode"></i>
            Reference:
          </label>
          <input type="text" v-model="filters.reference" placeholder="Search reference..." />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-layer-group"></i>
            Batch Sell:
          </label>
          <select v-model="filters.isBatchSell">
            <option :value="null">All</option>
            <option :value="true">Yes</option>
            <option :value="false">No</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Loading bills...</span>
    </div>

    <div v-if="error" class="error">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-else-if="sortedAndLimitedBills.length === 0" class="no-data">
      <i class="fas fa-inbox fa-2x"></i>
      <p>No sell bills found</p>
    </div>

    <div v-else class="table-container">
      <table class="sell-bills-table">
        <thead>
          <tr>
            <th @click="handleSort('id')" class="sortable">
              <i class="fas fa-hashtag"></i> ID
              <i
                v-if="sortConfig.field === 'id'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('bill_ref')" class="sortable">
              <i class="fas fa-barcode"></i> Reference
              <i
                v-if="sortConfig.field === 'bill_ref'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('date_sell')" class="sortable">
              <i class="fas fa-calendar"></i> Date
              <i
                v-if="sortConfig.field === 'date_sell'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('broker_name')" class="sortable">
              <i class="fas fa-user-tie"></i> Broker
              <i
                v-if="sortConfig.field === 'broker_name'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('created_by')" class="sortable">
              <i class="fas fa-user"></i> Created By
              <i
                v-if="sortConfig.field === 'created_by'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('notes')" class="sortable">
              <i class="fas fa-sticky-note"></i> Notes
              <i
                v-if="sortConfig.field === 'notes'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th><i class="fas fa-money-bill-wave"></i> Payment Status</th>
            <th><i class="fas fa-cog"></i> Actions</th>
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
            <td>{{ new Date(bill.date_sell).toLocaleDateString() }}</td>
            <td>{{ bill.broker_name || 'N/A' }}</td>
            <td>{{ bill.created_by || 'N/A' }}</td>
            <td>{{ bill.notes || 'N/A' }}</td>
            <td>
              <span
                :class="['payment-badge', `payment-${getPaymentStatus(bill).status}`]"
                :title="`Total: $${bill.total_cfr.toFixed(2)} | Paid: $${bill.total_paid.toFixed(2)}`"
              >
                {{ getPaymentStatus(bill).text }}
              </span>
            </td>
            <td class="actions">
              <button
                v-if="can_edit_sell_bill"
                @click.stop="handleEdit(bill)"
                :disabled="isProcessing"
                class="btn edit-btn"
                title="Edit Bill"
              >
                <i class="fas fa-edit"></i>
              </button>
              <button
                v-if="can_delete_sell_bill"
                @click.stop="handleDelete(bill.id)"
                :disabled="isProcessing"
                class="btn delete-btn"
                title="Delete Bill"
              >
                <i class="fas fa-trash-alt"></i>
              </button>
              <button
                @click.stop="handlePrint(bill.id)"
                :disabled="isProcessing"
                class="btn print-btn"
                title="Print Bill"
              >
                <i class="fas fa-print"></i>
              </button>
              <button
                v-if="can_c_sell_payments"
                @click.stop="handlePayments(bill.id)"
                :disabled="isProcessing"
                class="btn payment-btn"
                title="View Payments"
              >
                <i class="fas fa-money-bill-wave"></i>
              </button>
              <button
                @click.stop="handleTask(bill)"
                :disabled="isProcessing"
                class="btn task-btn"
                title="Add New Task"
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
</style>
