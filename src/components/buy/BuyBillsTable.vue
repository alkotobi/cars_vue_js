<script setup>
import { defineProps, defineEmits } from 'vue'
import { ref, computed, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  buyBills: {
    type: Array,
    required: true,
  },
  selectedBill: {
    type: Object,
    default: null,
  },
  loading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['select-bill'])

const { getFileUrl, callApi } = useApi()

// Add filter states
const filters = ref({
  dateFrom: '',
  dateTo: '',
  supplier: '',
  billRef: '',
  status: '', // 'all', 'updated', 'pending'
  isOrdered: '', // 'all', 'ordered', 'not_ordered'
})

// Add sorting state
const sortConfig = ref({
  field: 'date_buy',
  direction: 'desc',
})

// Filter and sort bills
const filteredAndSortedBills = computed(() => {
  let result = [...props.buyBills]

  // Apply filters
  if (filters.value.dateFrom) {
    result = result.filter((bill) => new Date(bill.date_buy) >= new Date(filters.value.dateFrom))
  }
  if (filters.value.dateTo) {
    result = result.filter((bill) => new Date(bill.date_buy) <= new Date(filters.value.dateTo))
  }
  if (filters.value.supplier) {
    result = result.filter((bill) =>
      bill.supplier_name?.toLowerCase().includes(filters.value.supplier.toLowerCase()),
    )
  }
  if (filters.value.billRef) {
    result = result.filter((bill) =>
      bill.bill_ref?.toLowerCase().includes(filters.value.billRef.toLowerCase()),
    )
  }
  if (filters.value.status) {
    if (filters.value.status === 'updated') {
      result = result.filter((bill) => bill.is_stock_updated)
    } else if (filters.value.status === 'pending') {
      result = result.filter((bill) => !bill.is_stock_updated)
    }
  }
  if (filters.value.isOrdered) {
    if (filters.value.isOrdered === 'ordered') {
      result = result.filter((bill) => bill.is_ordered === 1)
    } else if (filters.value.isOrdered === 'not_ordered') {
      result = result.filter((bill) => bill.is_ordered === 0)
    }
  }

  // Apply sorting
  result.sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle date comparison
    if (sortConfig.value.field === 'date_buy') {
      aValue = new Date(aValue).getTime()
      bValue = new Date(bValue).getTime()
    }

    if (sortConfig.value.direction === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })

  return result
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

const resetFilters = () => {
  filters.value = {
    dateFrom: '',
    dateTo: '',
    supplier: '',
    billRef: '',
    status: '',
    isOrdered: '',
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

const formatNumber = (value) => {
  const num = Number(value)
  return !isNaN(num) ? num.toFixed(2) : 'N/A'
}

// Add computed properties for payment status
const getPaymentStatus = (bill) => {
  const total = bill.amount || 0
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

// Add warehouse car count data
const warehouseCarCounts = ref({})

// Add export license car count data
const exportLicenseCarCounts = ref({})

// Function to get warehouse car count and total car count for a bill
const getWarehouseCarCount = async (billId) => {
  try {
    const result = await callApi({
      query: `
        SELECT 
          COUNT(*) as total_count,
          SUM(CASE WHEN cs.in_wharhouse_date IS NOT NULL THEN 1 ELSE 0 END) as warehouse_count
        FROM cars_stock cs
        INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
        WHERE bd.id_buy_bill = ?
      `,
      params: [billId],
    })

    if (result.success && result.data.length > 0) {
      return {
        total: result.data[0].total_count,
        warehouse: result.data[0].warehouse_count,
      }
    }
    return { total: 0, warehouse: 0 }
  } catch (error) {
    console.error('Error fetching warehouse car count:', error)
    return { total: 0, warehouse: 0 }
  }
}

// Function to get export license car count and total car count for a bill
const getExportLicenseCarCount = async (billId) => {
  try {
    const result = await callApi({
      query: `
        SELECT 
          COUNT(*) as total_count,
          SUM(CASE WHEN cs.export_lisence_ref IS NOT NULL AND cs.export_lisence_ref != '' THEN 1 ELSE 0 END) as export_license_count
        FROM cars_stock cs
        INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
        WHERE bd.id_buy_bill = ?
      `,
      params: [billId],
    })

    if (result.success && result.data.length > 0) {
      return {
        total: result.data[0].total_count,
        exportLicense: result.data[0].export_license_count,
      }
    }
    return { total: 0, exportLicense: 0 }
  } catch (error) {
    console.error('Error fetching export license car count:', error)
    return { total: 0, exportLicense: 0 }
  }
}

// Function to fetch warehouse counts for all bills
const fetchWarehouseCounts = async () => {
  const counts = {}
  for (const bill of props.buyBills) {
    counts[bill.id] = await getWarehouseCarCount(bill.id)
  }
  warehouseCarCounts.value = counts
}

// Function to fetch export license counts for all bills
const fetchExportLicenseCounts = async () => {
  const counts = {}
  for (const bill of props.buyBills) {
    counts[bill.id] = await getExportLicenseCarCount(bill.id)
  }
  exportLicenseCarCounts.value = counts
}

// Watch for changes in buyBills to update warehouse counts
watch(
  () => props.buyBills,
  () => {
    if (props.buyBills.length > 0) {
      fetchWarehouseCounts()
      fetchExportLicenseCounts()
    }
  },
  { immediate: true },
)

// Function to get badge class for warehouse status
const getWarehouseBadgeClass = (billId) => {
  const counts = warehouseCarCounts.value[billId]
  if (!counts) return ''

  // Convert to numbers to handle string values from database
  const total = Number(counts.total) || 0
  const warehouse = Number(counts.warehouse) || 0

  if (total === 0) return 'no-bill-cars'
  if (warehouse === 0) return 'no-cars'
  if (warehouse > 0) return 'has-warehouse-cars'
  return ''
}

// Function to get badge class for export license status
const getExportLicenseBadgeClass = (billId) => {
  const counts = exportLicenseCarCounts.value[billId]
  if (!counts) return ''

  // Convert to numbers to handle string values from database
  const total = Number(counts.total) || 0
  const exportLicense = Number(counts.exportLicense) || 0

  if (total === 0) return 'no-bill-cars'
  if (exportLicense === 0) return 'no-cars'
  if (exportLicense > 0) return 'has-export-license-cars'
  return ''
}

const selectBill = (bill) => {
  emit('select-bill', bill)
}

// Expose functions to parent components
defineExpose({
  fetchWarehouseCounts,
  fetchExportLicenseCounts,
})
</script>

<template>
  <div class="master-section">
    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Loading bills...</span>
    </div>

    <!-- No Selection Message -->
    <div v-if="!loading && buyBills.length === 0" class="no-selection">
      <i class="fas fa-inbox fa-2x"></i>
      <p>No purchase bills found</p>
    </div>

    <div v-else-if="!loading">
      <!-- Filters -->
      <div class="filters">
        <div class="filter-group">
          <label>
            <i class="fas fa-calendar"></i>
            Date From
          </label>
          <input type="date" v-model="filters.dateFrom" class="filter-input" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-calendar"></i>
            Date To
          </label>
          <input type="date" v-model="filters.dateTo" class="filter-input" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-building"></i>
            Supplier
          </label>
          <input
            type="text"
            v-model="filters.supplier"
            placeholder="Search supplier..."
            class="filter-input"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-file-alt"></i>
            Bill Ref
          </label>
          <input
            type="text"
            v-model="filters.billRef"
            placeholder="Search bill ref..."
            class="filter-input"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-info-circle"></i>
            Status
          </label>
          <select v-model="filters.status" class="filter-input">
            <option value="">All</option>
            <option value="updated">Updated</option>
            <option value="pending">Pending</option>
          </select>
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-info-circle"></i>
            Is Ordered
          </label>
          <select v-model="filters.isOrdered" class="filter-input">
            <option value="">All</option>
            <option value="ordered">Ordered</option>
            <option value="not_ordered">Not Ordered</option>
          </select>
        </div>

        <button @click="resetFilters" class="reset-btn">
          <i class="fas fa-undo"></i>
          Reset Filters
        </button>
      </div>

      <!-- Toolbar -->
      <div class="toolbar" v-if="selectedBill">
        <div class="bill-info">
          <div class="bill-id">
            <i class="fas fa-file-invoice-dollar"></i>
            Purchase #{{ selectedBill.id }}
          </div>
          <div class="bill-details">
            <div class="detail-item">
              <i class="fas fa-building"></i>
              <span class="label">Supplier:</span>
              <span class="value">{{ selectedBill.supplier_name }}</span>
            </div>
            <div class="detail-item">
              <i class="fas fa-money-bill-wave"></i>
              <span class="label">Amount:</span>
              <span class="value">{{ formatNumber(selectedBill.amount) }}</span>
            </div>
            <div class="detail-item">
              <i class="fas fa-check-circle"></i>
              <span class="label">Paid:</span>
              <span class="value">{{ formatNumber(selectedBill.payed) }}</span>
            </div>
            <div class="detail-item">
              <i class="fas fa-balance-scale"></i>
              <span class="label">Balance:</span>
              <span class="value">{{
                formatNumber(selectedBill.amount - selectedBill.payed)
              }}</span>
            </div>
            <div class="detail-item">
              <i
                class="fas"
                :class="
                  selectedBill.is_stock_updated ? 'fa-check text-success' : 'fa-clock text-warning'
                "
              ></i>
              <span class="label">Status:</span>
              <span
                class="value"
                :class="selectedBill.is_stock_updated ? 'status-updated' : 'status-pending'"
              >
                {{ selectedBill.is_stock_updated ? 'Updated' : 'Pending' }}
              </span>
            </div>
          </div>
        </div>
        <div class="toolbar-actions">
          <slot name="actions" :bill="selectedBill"></slot>
        </div>
      </div>

      <!-- Table -->
      <table class="data-table" :class="{ 'with-selection': selectedBill }">
        <thead>
          <tr>
            <th @click="handleSort('id')" class="id-column">
              <i class="fas fa-hashtag"></i> ID
              <i
                v-if="sortConfig.field === 'id'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th @click="handleSort('date_buy')" class="sortable date-column">
              <i class="fas fa-calendar"></i> Date
              <i
                v-if="sortConfig.field === 'date_buy'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th @click="handleSort('supplier_name')" class="sortable">
              <i class="fas fa-building"></i> Supplier
              <i
                v-if="sortConfig.field === 'supplier_name'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th @click="handleSort('bill_ref')" class="sortable">
              <i class="fas fa-file-alt"></i> Bill Ref
              <i
                v-if="sortConfig.field === 'bill_ref'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th><i class="fas fa-money-bill-wave"></i> Amount</th>
            <th><i class="fas fa-check-circle"></i> Paid</th>
            <th><i class="fas fa-balance-scale"></i> Balance</th>
            <th><i class="fas fa-money-bill-wave"></i> Payment Status</th>
            <th class="status-column"><i class="fas fa-info-circle"></i> Status</th>
            <th><i class="fas fa-shopping-cart"></i> Order Status</th>
            <th><i class="fas fa-file-pdf"></i> PI Document</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="bill in filteredAndSortedBills"
            :key="bill.id"
            :class="{ selected: selectedBill?.id === bill.id }"
            @click="$emit('select-bill', bill)"
          >
            <td>{{ bill.id }}</td>
            <td>{{ formatDate(bill.date_buy) }}</td>
            <td>{{ bill.supplier_name }}</td>
            <td>{{ bill.bill_ref || 'N/A' }}</td>
            <td>{{ formatNumber(bill.amount) }}</td>
            <td>{{ formatNumber(bill.payed) }}</td>
            <td>{{ formatNumber(bill.amount - bill.payed) }}</td>
            <td>
              <span
                :class="['payment-badge', `payment-${getPaymentStatus(bill).status}`]"
                :title="`Total: $${bill.amount.toFixed(2)} | Paid: $${bill.total_paid.toFixed(2)}`"
              >
                {{ getPaymentStatus(bill).text }}
              </span>
            </td>
            <td class="status-column">
              <div class="status-container">
                <span :class="bill.is_stock_updated ? 'status-updated' : 'status-pending'">
                  <i class="fas" :class="bill.is_stock_updated ? 'fa-check' : 'fa-clock'"></i>
                  {{ bill.is_stock_updated ? 'Updated' : 'Pending' }}
                </span>
                <span
                  class="warehouse-badge"
                  :class="getWarehouseBadgeClass(bill.id)"
                  :title="
                    warehouseCarCounts[bill.id]?.total > 0
                      ? `${warehouseCarCounts[bill.id]?.warehouse} of ${warehouseCarCounts[bill.id]?.total} car${warehouseCarCounts[bill.id]?.total === 1 ? '' : 's'} received in warehouse`
                      : 'No cars in this purchase bill'
                  "
                >
                  <span class="badge-label">Warehouse:</span>
                  <span v-if="warehouseCarCounts[bill.id]?.total > 0">
                    {{ warehouseCarCounts[bill.id]?.warehouse }} of
                    {{ warehouseCarCounts[bill.id]?.total }}
                  </span>
                  <span v-else>No cars</span>
                </span>
                <span
                  class="export-license-badge"
                  :class="getExportLicenseBadgeClass(bill.id)"
                  :title="
                    exportLicenseCarCounts[bill.id]?.total > 0
                      ? `${exportLicenseCarCounts[bill.id]?.exportLicense} of ${exportLicenseCarCounts[bill.id]?.total} car${exportLicenseCarCounts[bill.id]?.total === 1 ? '' : 's'} have export license`
                      : 'No cars in this purchase bill'
                  "
                >
                  <span class="badge-label">Export License:</span>
                  <span v-if="exportLicenseCarCounts[bill.id]?.total > 0">
                    {{ exportLicenseCarCounts[bill.id]?.exportLicense }} of
                    {{ exportLicenseCarCounts[bill.id]?.total }}
                  </span>
                  <span v-else>No cars</span>
                </span>
              </div>
            </td>
            <td>
              <span :class="bill.is_ordered ? 'status-ordered' : 'status-not_ordered'">
                <i class="fas" :class="bill.is_ordered ? 'fa-check' : 'fa-times'"></i>
                {{ bill.is_ordered ? 'Ordered' : 'Not Ordered' }}
              </span>
            </td>
            <td>
              <a
                v-if="bill.pi_path"
                :href="getFileUrl(bill.pi_path)"
                target="_blank"
                class="pi-document-link"
                @click.stop
              >
                <i class="fas fa-file-pdf"></i>
                View PI
              </a>
              <span v-else class="no-document">
                <i class="fas fa-times-circle"></i>
                No document
              </span>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Selection Message -->
      <div v-if="!selectedBill" class="no-selection">
        <i class="fas fa-hand-pointer fa-2x"></i>
        <p>Please select a purchase bill to view its details</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.master-section {
  width: 100%;
  position: relative;
  min-height: 200px;
}

.toolbar {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.bill-info {
  flex: 1;
}

.bill-id {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.bill-id i {
  color: #3b82f6;
}

.bill-details {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: white;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.detail-item i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.label {
  color: #64748b;
  font-size: 0.875rem;
}

.value {
  font-weight: 500;
  color: #1e293b;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.data-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #374151;
  white-space: nowrap;
}

.data-table th.id-column {
  width: 4ch;
  min-width: 4ch;
  max-width: 4ch;
}

.data-table th.date-column {
  width: 10ch;
  min-width: 10ch;
  max-width: 10ch;
}

.data-table th.status-column {
  width: 25ch;
  min-width: 25ch;
  max-width: 25ch;
}

.data-table td.status-column {
  width: 25ch;
  min-width: 25ch;
  max-width: 25ch;
}

.data-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.data-table tbody tr {
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
  border-left: 4px solid transparent;
}

.data-table tbody tr:hover {
  background-color: #f8fafc;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.data-table tbody tr.selected {
  background-color: #eff6ff;
  border-left-color: #3b82f6;
}

.data-table tbody tr.selected::before {
  content: none;
}

.data-table tbody tr.selected td:first-child {
  position: relative;
}

.data-table tbody tr.selected td:first-child::before {
  content: '';
  position: absolute;
  left: -4px;
  top: 0;
  bottom: 0;
  width: 4px;
  background-color: #3b82f6;
}

.data-table td {
  color: #4b5563;
}

.payment-btn {
  position: relative;
  overflow: hidden;
}

.payment-btn:not(:disabled):hover {
  transform: translateY(-1px);
}

.payment-btn:not(:disabled):active {
  transform: translateY(0);
}

.payment-btn i {
  margin-right: 0.5rem;
  transition: transform 0.2s ease;
}

.payment-btn:hover i {
  transform: scale(1.1);
}

.action-btn {
  position: relative;
  overflow: hidden;
}

.action-btn:not(:disabled):hover {
  transform: translateY(-1px);
}

.action-btn:not(:disabled):active {
  transform: translateY(0);
}

.action-btn i {
  margin-right: 0.5rem;
  transition: transform 0.2s ease;
}

.action-btn:hover i {
  transform: scale(1.1);
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 12px;
  z-index: 10;
  backdrop-filter: blur(2px);
}

.loading-overlay i {
  color: #3b82f6;
  animation: pulse 1.5s infinite;
}

.loading-overlay span {
  color: #4b5563;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.data-table.with-selection {
  margin-top: 1rem;
}

.no-selection {
  text-align: center;
  padding: 2rem;
  background-color: #f9fafb;
  border-radius: 0.5rem;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  margin: 1.5rem 0;
  border: 2px dashed #e5e7eb;
  transition: all 0.3s ease;
}

.no-selection:hover {
  border-color: #3b82f6;
  background-color: #f0f7ff;
}

.no-selection i {
  color: #3b82f6;
  animation: bounce 2s infinite;
}

.no-selection p {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 500;
}

@keyframes bounce {
  0%,
  20%,
  50%,
  80%,
  100% {
    transform: translateY(0);
  }
  40% {
    transform: translateY(-10px);
  }
  60% {
    transform: translateY(-5px);
  }
}

@keyframes pulse {
  0% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.7;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

.pi-document-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  padding: 6px 12px;
  border-radius: 6px;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: #eff6ff;
  position: relative;
  overflow: hidden;
}

.pi-document-link:hover {
  background-color: #dbeafe;
  transform: translateY(-1px);
}

.pi-document-link:active {
  transform: translateY(0);
}

.pi-document-link i {
  color: #3b82f6;
  transition: transform 0.2s ease;
}

.pi-document-link:hover i {
  transform: scale(1.1);
}

.status-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-start;
}

.status-updated,
.status-pending {
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.warehouse-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  background-color: #8b5cf6;
  color: white;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  transition: all 0.2s ease;
  cursor: help;
}

.warehouse-badge.has-warehouse-cars {
  background-color: #8b5cf6 !important;
  color: white !important;
}

.warehouse-badge.no-cars {
  background-color: #ef4444 !important;
  color: white !important;
  border: 2px solid #dc2626 !important;
  box-shadow: 0 0 10px #ef4444 !important;
}

.warehouse-badge.no-bill-cars {
  background-color: #9ca3af !important;
  color: #6b7280 !important;
}

.warehouse-badge:hover {
  background-color: #7c3aed;
  transform: scale(1.05);
}

.warehouse-badge.no-cars:hover {
  background-color: #dc2626;
}

.warehouse-badge.no-bill-cars:hover {
  background-color: #6b7280;
}

.warehouse-badge.has-warehouse-cars:hover {
  background-color: #7c3aed;
}

.warehouse-badge .badge-label {
  font-size: 0.7rem;
  font-weight: 600;
  opacity: 0.9;
}

.export-license-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  background-color: #10b981;
  color: white;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  transition: all 0.2s ease;
  cursor: help;
}

.export-license-badge.has-export-license-cars {
  background-color: #10b981 !important;
  color: white !important;
}

.export-license-badge.no-cars {
  background-color: #ef4444 !important;
  color: white !important;
  border: 2px solid #dc2626 !important;
  box-shadow: 0 0 10px #ef4444 !important;
}

.export-license-badge.no-bill-cars {
  background-color: #9ca3af !important;
  color: #6b7280 !important;
}

.export-license-badge:hover {
  background-color: #059669;
  transform: scale(1.05);
}

.export-license-badge.no-cars:hover {
  background-color: #dc2626;
}

.export-license-badge.no-bill-cars:hover {
  background-color: #6b7280;
}

.export-license-badge.has-export-license-cars:hover {
  background-color: #059669;
}

.export-license-badge .badge-label {
  font-size: 0.7rem;
  font-weight: 600;
  opacity: 0.9;
}

.status-updated {
  color: #059669;
  background-color: #ecfdf5;
}

.status-pending {
  color: #d97706;
  background-color: #fffbeb;
}

.status-updated i,
.status-pending i {
  transition: transform 0.2s ease;
}

.status-updated:hover i,
.status-pending:hover i {
  transform: scale(1.1);
}

.status-ordered,
.status-not_ordered {
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.status-ordered {
  color: #059669;
  background-color: #ecfdf5;
}

.status-not_ordered {
  color: #dc2626;
  background-color: #fef2f2;
}

.status-ordered i,
.status-not_ordered i {
  transition: transform 0.2s ease;
}

.status-ordered:hover i,
.status-not_ordered:hover i {
  transform: scale(1.1);
}

/* Responsive styles */
@media (max-width: 1024px) {
  .toolbar {
    flex-direction: column;
    gap: 16px;
  }

  .bill-details {
    flex-direction: column;
    gap: 12px;
  }

  .detail-item {
    width: 100%;
  }

  .toolbar-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

@media (max-width: 768px) {
  .data-table {
    display: block;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }
}

.filters {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-end;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 200px;
}

.filter-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #64748b;
  font-size: 0.875rem;
}

.filter-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.filter-input {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 0.875rem;
  color: #1e293b;
  background-color: white;
  transition: all 0.2s;
}

.filter-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.filter-input::placeholder {
  color: #94a3b8;
}

.reset-btn {
  padding: 8px 16px;
  background-color: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  color: #64748b;
  font-size: 0.875rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
  height: fit-content;
}

.reset-btn:hover {
  background-color: #e2e8f0;
  color: #475569;
}

.reset-btn i {
  font-size: 0.75rem;
}

/* Sorting styles */
.data-table th {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 24px; /* Space for sort icon */
}

.data-table th i:last-child {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 0.75rem;
  opacity: 0.5;
}

.data-table th:hover {
  background-color: #f1f5f9;
}

.data-table th:hover i:last-child {
  opacity: 1;
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

/* Responsive adjustments */
@media (max-width: 1024px) {
  .filters {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-group {
    min-width: 100%;
  }
}
</style>
