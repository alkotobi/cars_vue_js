<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, error } = useApi()
const clientTransfers = ref([])
const isLoading = ref(false)

// Add sort configuration
const sortConfig = ref({
  field: 'date_do_transfer',
  direction: 'desc',
})

// Filter states
const filters = ref({
  clientName: '',
  dateFrom: '',
  dateTo: '',
  amountMin: '',
  amountMax: '',
})

// Add toggle sort function
const toggleSort = (field) => {
  if (sortConfig.value.field === field) {
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    sortConfig.value.field = field
    sortConfig.value.direction = 'asc'
  }
}

// Fetch client transfers with their details
const fetchClientTransfers = async () => {
  isLoading.value = true
  try {
    const result = await callApi({
      query: `
        SELECT 
          td.id,
          td.client_name,
          td.amount,
          td.rate,
          td.notes,
          t.date_do_transfer,
          t.date_receive,
          t.ref_pi_transfer,
          b.company_name,
          b.bank_name
        FROM transfer_details td
        JOIN transfers t ON td.id_transfer = t.id
        LEFT JOIN banks b ON t.id_bank = b.id
        ORDER BY t.date_do_transfer DESC
      `,
      params: [],
    })
    if (result.success) {
      clientTransfers.value = result.data
    }
  } catch (err) {
    console.error('Error fetching client transfers:', err)
  } finally {
    isLoading.value = false
  }
}

// Filtered and computed transfers
const filteredTransfers = computed(() => {
  let result = [...clientTransfers.value]

  // Apply filters
  if (filters.value.clientName) {
    result = result.filter((t) =>
      t.client_name.toLowerCase().includes(filters.value.clientName.toLowerCase()),
    )
  }

  if (filters.value.dateFrom) {
    result = result.filter((t) => new Date(t.date_do_transfer) >= new Date(filters.value.dateFrom))
  }

  if (filters.value.dateTo) {
    result = result.filter((t) => new Date(t.date_do_transfer) <= new Date(filters.value.dateTo))
  }

  if (filters.value.amountMin) {
    const min = parseFloat(filters.value.amountMin)
    if (!isNaN(min)) {
      result = result.filter((t) => parseFloat(t.amount) >= min)
    }
  }

  if (filters.value.amountMax) {
    const max = parseFloat(filters.value.amountMax)
    if (!isNaN(max)) {
      result = result.filter((t) => parseFloat(t.amount) <= max)
    }
  }

  // Apply sorting
  result.sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Special handling for USD amount
    if (sortConfig.value.field === 'amount_usd') {
      aValue = Number(a.amount) / Number(a.rate)
      bValue = Number(b.amount) / Number(b.rate)
    }
    // Special handling for numeric fields
    else if (['amount', 'rate'].includes(sortConfig.value.field)) {
      aValue = Number(aValue) || 0
      bValue = Number(bValue) || 0
    }
    // Special handling for dates
    else if (sortConfig.value.field === 'date_do_transfer') {
      aValue = new Date(aValue).getTime()
      bValue = new Date(bValue).getTime()
    }
    // Default string comparison
    else {
      aValue = String(aValue || '').toLowerCase()
      bValue = String(bValue || '').toLowerCase()
    }

    if (sortConfig.value.direction === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })

  return result
})

const clearFilters = () => {
  filters.value = {
    clientName: '',
    dateFrom: '',
    dateTo: '',
    amountMin: '',
    amountMax: '',
  }
}

onMounted(() => {
  fetchClientTransfers()
})
</script>

<template>
  <div class="client-transfers">
    <div class="header">
      <h1>Client Transfer Details</h1>
    </div>

    <!-- Filters Section -->
    <div class="filters-section">
      <div class="filters-grid">
        <div class="filter-group">
          <label><i class="fas fa-user"></i> Client Name</label>
          <input
            type="text"
            v-model="filters.clientName"
            class="filter-input"
            placeholder="Filter by client name..."
          />
        </div>
        <div class="filter-group">
          <label><i class="fas fa-calendar"></i> From Date</label>
          <input type="date" v-model="filters.dateFrom" class="filter-input" />
        </div>
        <div class="filter-group">
          <label><i class="fas fa-calendar"></i> To Date</label>
          <input type="date" v-model="filters.dateTo" class="filter-input" />
        </div>
        <div class="filter-group">
          <label><i class="fas fa-money-bill"></i> Min Amount</label>
          <input
            type="number"
            v-model="filters.amountMin"
            class="filter-input"
            placeholder="Minimum amount..."
            min="0"
            step="0.01"
          />
        </div>
        <div class="filter-group">
          <label><i class="fas fa-money-bill"></i> Max Amount</label>
          <input
            type="number"
            v-model="filters.amountMax"
            class="filter-input"
            placeholder="Maximum amount..."
            min="0"
            step="0.01"
          />
        </div>
        <button @click="clearFilters" class="btn clear-btn">
          <i class="fas fa-times"></i> Clear Filters
        </button>
      </div>
    </div>

    <!-- Transfers Table -->
    <div class="transfers-table">
      <div v-if="isLoading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading client transfers...</span>
      </div>
      <table>
        <thead>
          <tr>
            <th @click="toggleSort('client_name')" class="sortable">
              Client Name
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'client_name'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th @click="toggleSort('date_do_transfer')" class="sortable">
              Date
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'date_do_transfer'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th @click="toggleSort('amount')" class="sortable">
              Amount (DA)
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'amount'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th @click="toggleSort('amount_usd')" class="sortable">
              Amount ($)
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'amount_usd'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th @click="toggleSort('rate')" class="sortable">
              Rate
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'rate'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th @click="toggleSort('company_name')" class="sortable">
              Bank
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'company_name'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th @click="toggleSort('ref_pi_transfer')" class="sortable">
              PI Reference
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'ref_pi_transfer'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th>Status</th>
            <th>Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="transfer in filteredTransfers" :key="transfer.id">
            <td>{{ transfer.client_name }}</td>
            <td>{{ new Date(transfer.date_do_transfer).toLocaleDateString() }}</td>
            <td class="amount-cell">{{ Number(transfer.amount).toFixed(2) }} DA</td>
            <td class="amount-cell">
              ${{ (Number(transfer.amount) / Number(transfer.rate)).toFixed(2) }}
            </td>
            <td>{{ transfer.rate }}</td>
            <td>
              <div v-if="transfer.company_name" class="bank-cell">
                <strong>{{ transfer.company_name }}</strong
                ><br />
                <small>{{ transfer.bank_name }}</small>
              </div>
              <span v-else>-</span>
            </td>
            <td>{{ transfer.ref_pi_transfer || '-' }}</td>
            <td class="status-cell">
              <span v-if="transfer.date_receive" class="status received">✅</span>
              <span v-else class="status pending">⏳</span>
            </td>
            <td>{{ transfer.notes || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.client-transfers {
  padding: 20px;
}

.header {
  margin-bottom: 20px;
}

.header h1 {
  color: #1f2937;
  font-size: 1.8em;
}

.filters-section {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  align-items: end;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  color: #4b5563;
  font-size: 0.9em;
  display: flex;
  align-items: center;
  gap: 4px;
}

.filter-input {
  padding: 8px;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  font-size: 0.9em;
}

.clear-btn {
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  height: 38px;
  align-self: flex-end;
}

.clear-btn:hover {
  background-color: #4b5563;
}

.transfers-table {
  position: relative;
  overflow-x: auto;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

th {
  background-color: #f8f9fa;
  font-weight: 600;
  color: #4b5563;
}

.amount-cell {
  font-family: monospace;
  text-align: right;
}

.bank-cell {
  font-size: 0.9em;
  line-height: 1.3;
}

.bank-cell strong {
  color: #1f2937;
}

.bank-cell small {
  color: #6b7280;
}

.status-cell {
  text-align: center;
}

.status {
  font-size: 1.2em;
}

.received {
  color: #10b981;
}

.pending {
  color: #f59e0b;
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
  gap: 12px;
  font-size: 1.1em;
  color: #4b5563;
}

.loading-overlay i {
  color: #3b82f6;
}

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
  opacity: 0.5;
}

.sortable:hover i {
  opacity: 1;
}

.fa-sort-up,
.fa-sort-down {
  opacity: 1;
  color: #2563eb;
}
</style>
