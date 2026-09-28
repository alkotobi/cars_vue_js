<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, error } = useApi()
const clientTransfers = ref([])
const isLoading = ref(false)
const showAddClientDialog = ref(false)
const isProcessing = ref(false)
const newClient = ref({
  name: '',
  phone: '',
  email: '',
  address: '',
})

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
          COALESCE(c.name, td.client_name) as client_name,
          td.amount,
          td.rate,
          td.notes,
          t.date_do_transfer,
          t.date_receive,
          t.ref_pi_transfer,
          b.company_name,
          b.bank_name,
          (td.amount * td.rate) as amount_da
        FROM transfer_details td
        JOIN transfers t ON td.id_transfer = t.id
        LEFT JOIN banks b ON t.id_bank = b.id
        LEFT JOIN clients c ON td.id_client = c.id
        ORDER BY t.date_do_transfer DESC
      `,
      params: [],
    })
    if (result.success) {
      clientTransfers.value = result.data.map((transfer) => ({
        ...transfer,
        amount: Number(transfer.amount).toFixed(2),
        amount_da: Number(transfer.amount_da).toFixed(2),
      }))
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

// Add new client function
const addNewClient = async () => {
  if (isProcessing.value) return
  isProcessing.value = true
  try {
    if (!newClient.value.name.trim()) {
      error.value = 'Client name is required'
      return
    }

    const result = await callApi({
      query: `
        INSERT INTO clients (name, phone, email, address)
        VALUES (?, ?, ?, ?)
      `,
      params: [
        newClient.value.name.trim(),
        newClient.value.phone.trim() || null,
        newClient.value.email.trim() || null,
        newClient.value.address.trim() || null,
      ],
    })

    if (result.success) {
      showAddClientDialog.value = false
      newClient.value = {
        name: '',
        phone: '',
        email: '',
        address: '',
      }
      // Optionally refresh the transfers list if it includes client information
      await fetchClientTransfers()
    }
  } catch (err) {
    console.error('Error adding new client:', err)
    error.value = err.message || 'Failed to add new client'
  } finally {
    isProcessing.value = false
  }
}

onMounted(() => {
  fetchClientTransfers()
})
</script>

<template>
  <div class="client-transfers">
    <div class="page-header">
      <div class="title">
        <h2><i class="fas fa-exchange-alt"></i> Client Transfers Details</h2>
      </div>
      <div class="header-actions">
        <button @click="showAddClientDialog = true" class="btn add-client-btn">
          <i class="fas fa-user-plus"></i> Add New Client
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters">
      <div class="filter-group">
        <label><i class="fas fa-user"></i> Client Name:</label>
        <input
          v-model="filters.clientName"
          placeholder="Filter by client name"
          class="filter-input"
        />
      </div>
      <div class="filter-group">
        <label><i class="fas fa-calendar"></i> Date Range:</label>
        <input type="date" v-model="filters.dateFrom" class="filter-input date-input" />
        <span>to</span>
        <input type="date" v-model="filters.dateTo" class="filter-input date-input" />
      </div>
      <div class="filter-group">
        <label><i class="fas fa-dollar-sign"></i> Amount Range (USD):</label>
        <input
          type="number"
          v-model="filters.amountMin"
          placeholder="Min"
          class="filter-input amount-input"
        />
        <span>to</span>
        <input
          type="number"
          v-model="filters.amountMax"
          placeholder="Max"
          class="filter-input amount-input"
        />
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading">
      <i class="fas fa-spinner fa-spin"></i> Loading transfers...
    </div>

    <!-- Table -->
    <div v-else class="transfers-table">
      <table>
        <thead>
          <tr>
            <th @click="toggleSort('client_name')" class="sortable">
              Client Name
              <i
                v-if="sortConfig.field === 'client_name'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="toggleSort('date_do_transfer')" class="sortable">
              Date
              <i
                v-if="sortConfig.field === 'date_do_transfer'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="toggleSort('amount_da')" class="sortable">
              Amount (DA)
              <i
                v-if="sortConfig.field === 'amount_da'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="toggleSort('amount')" class="sortable">
              Amount (USD)
              <i
                v-if="sortConfig.field === 'amount'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="toggleSort('rate')" class="sortable">
              Rate
              <i
                v-if="sortConfig.field === 'rate'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th>Bank</th>
            <th>Reference</th>
            <th>Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="transfer in filteredTransfers" :key="transfer.id">
            <td>{{ transfer.client_name }}</td>
            <td>{{ new Date(transfer.date_do_transfer).toLocaleDateString() }}</td>
            <td class="amount-cell">{{ transfer.amount_da }} DA</td>
            <td class="amount-cell">${{ transfer.amount }}</td>
            <td>{{ transfer.rate }}</td>
            <td>{{ transfer.bank_name }} ({{ transfer.company_name }})</td>
            <td>{{ transfer.ref_pi_transfer || '-' }}</td>
            <td>{{ transfer.notes || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Client Dialog -->
    <div v-if="showAddClientDialog" class="dialog-overlay">
      <div class="dialog">
        <h3><i class="fas fa-user-plus"></i> Add New Client</h3>
        <div v-if="error" class="error-message">{{ error }}</div>

        <form @submit.prevent="addNewClient" class="client-form">
          <div class="form-group">
            <label><i class="fas fa-user"></i> Name *</label>
            <input
              v-model="newClient.name"
              type="text"
              class="input-field"
              placeholder="Enter client name"
              required
            />
          </div>

          <div class="form-group">
            <label><i class="fas fa-phone"></i> Phone</label>
            <input
              v-model="newClient.phone"
              type="tel"
              class="input-field"
              placeholder="Enter phone number"
            />
          </div>

          <div class="form-group">
            <label><i class="fas fa-envelope"></i> Email</label>
            <input
              v-model="newClient.email"
              type="email"
              class="input-field"
              placeholder="Enter email address"
            />
          </div>

          <div class="form-group">
            <label><i class="fas fa-map-marker-alt"></i> Address</label>
            <textarea
              v-model="newClient.address"
              class="input-field"
              placeholder="Enter address"
              rows="3"
            ></textarea>
          </div>

          <div class="dialog-actions">
            <button type="submit" class="btn save-btn" :disabled="isProcessing">
              <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
              <span v-else>Save Client</span>
            </button>
            <button
              type="button"
              @click="showAddClientDialog = false"
              class="btn cancel-btn"
              :disabled="isProcessing"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.client-transfers {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  width: 100%;
}

.title {
  display: flex;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.add-client-btn {
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-size: 0.9em;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
  height: 40px;
}

.add-client-btn:hover {
  background-color: #059669;
}

.filters {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-input {
  padding: 6px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  font-size: 0.9em;
}

.date-input {
  width: 130px;
}

.amount-input {
  width: 100px;
}

.transfers-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 12px;
}

th,
td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #1f2937;
}

.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 20px;
}

.sortable i {
  position: absolute;
  right: 4px;
  top: 50%;
  transform: translateY(-50%);
}

.sortable:hover {
  background-color: #f1f5f9;
}

.amount-cell {
  font-family: monospace;
  text-align: right;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.loading i {
  margin-right: 8px;
  color: #3b82f6;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 8px;
  padding: 24px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 20px 0;
  color: #1f2937;
}

.client-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-group label {
  color: #4b5563;
  font-size: 0.9em;
  display: flex;
  align-items: center;
  gap: 6px;
}

.input-field {
  padding: 8px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  font-size: 0.9em;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

textarea.input-field {
  resize: vertical;
  min-height: 80px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
}

.save-btn:hover {
  background-color: #2563eb;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
}

.cancel-btn:hover {
  background-color: #4b5563;
}

.error-message {
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  color: #b91c1c;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>
