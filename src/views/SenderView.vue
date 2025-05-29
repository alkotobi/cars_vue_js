<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'
import TransferDetails from '../components/TransferDetails.vue'

const router = useRouter()
const { callApi, error } = useApi()
const user = ref(null)
const transfers = ref([])
const showEditDialog = ref(false)
const showAddDialog = ref(false)
const selectedTransfer = ref(null)
const editForm = ref({
  amount_sending_da: '',
  rate: '',
  notes: '',
  id_bank: null,
  date_do_transfer: '',
  ref_pi_transfer: '',
})

const newTransfer = ref({
  amount_sending_da: '',
  rate: '',
  notes: '',
  id_bank: null,
  date_do_transfer: new Date().toISOString().split('T')[0],
  ref_pi_transfer: '',
})

const formError = ref(null)
const banks = ref([])
const showDetailsDialog = ref(false)
const selectedTransferForDetails = ref(null)

const isAdmin = computed(() => user.value?.role_id === 1)

const filters = ref({
  bank: '',
  dateFrom: '',
  dateTo: '',
  status: '', // 'received', 'pending', or ''
  amountMin: '',
  amountMax: '',
})

const sortConfig = ref({
  field: 'date_do_transfer',
  direction: 'desc',
})

const filteredTransfers = computed(() => {
  let result = [...transfers.value]

  // Apply filters
  if (filters.value.bank) {
    result = result.filter(
      (t) =>
        (t.company_name || '').toLowerCase().includes(filters.value.bank.toLowerCase()) ||
        (t.bank_name || '').toLowerCase().includes(filters.value.bank.toLowerCase()),
    )
  }

  if (filters.value.dateFrom) {
    result = result.filter((t) => new Date(t.date_do_transfer) >= new Date(filters.value.dateFrom))
  }

  if (filters.value.dateTo) {
    result = result.filter((t) => new Date(t.date_do_transfer) <= new Date(filters.value.dateTo))
  }

  if (filters.value.status) {
    if (filters.value.status === 'received') {
      result = result.filter((t) => t.date_receive)
    } else if (filters.value.status === 'pending') {
      result = result.filter((t) => !t.date_receive)
    }
  }

  if (filters.value.amountMin) {
    const min = parseFloat(filters.value.amountMin)
    if (!isNaN(min)) {
      result = result.filter((t) => parseFloat(t.amount_sending_da) >= min)
    }
  }

  if (filters.value.amountMax) {
    const max = parseFloat(filters.value.amountMax)
    if (!isNaN(max)) {
      result = result.filter((t) => parseFloat(t.amount_sending_da) <= max)
    }
  }

  // Apply sorting
  result.sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle numeric fields
    if (['amount_sending_da', 'rate', 'amount_received_usd'].includes(sortConfig.value.field)) {
      aValue = parseFloat(aValue) || 0
      bValue = parseFloat(bValue) || 0
    }

    // Handle date fields
    if (['date_do_transfer', 'date_receive'].includes(sortConfig.value.field)) {
      aValue = new Date(aValue || null).getTime()
      bValue = new Date(bValue || null).getTime()
    }

    if (sortConfig.value.direction === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })

  return result
})

const toggleSort = (field) => {
  if (sortConfig.value.field === field) {
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    sortConfig.value.field = field
    sortConfig.value.direction = 'asc'
  }
}

const clearFilters = () => {
  filters.value = {
    bank: '',
    dateFrom: '',
    dateTo: '',
    status: '',
    amountMin: '',
    amountMax: '',
  }
}

const fetchTransfers = async () => {
  const result = await callApi({
    query: `
      SELECT t.*, b.company_name, b.bank_name, b.bank_account, b.swift_code
      FROM transfers t
      LEFT JOIN banks b ON t.id_bank = b.id
      WHERE t.id_user_do_transfer = ? OR t.id_user_receive_transfer = ?
      ORDER BY t.date_do_transfer DESC
    `,
    params: [user.value.id, user.value.id],
  })
  if (result.success) {
    console.log('Fetched transfers:', result.data)
    transfers.value = result.data
  }
}

const calculateUSD = (amount, rate) => {
  return (amount / rate).toFixed(2)
}

const openEditDialog = (transfer) => {
  selectedTransfer.value = transfer
  editForm.value = {
    amount_sending_da: transfer.amount_sending_da,
    rate: transfer.rate,
    notes: transfer.notes || '',
    id_bank: transfer.id_bank || null,
    date_do_transfer: new Date(transfer.date_do_transfer).toISOString().split('T')[0],
    ref_pi_transfer: transfer.ref_pi_transfer || '',
  }
  showEditDialog.value = true
}

const updateTransfer = async () => {
  if (
    !editForm.value.amount_sending_da ||
    !editForm.value.rate ||
    !editForm.value.date_do_transfer
  ) {
    error.value = 'Please fill all required fields'
    return
  }

  if (!editForm.value.id_bank) {
    error.value = 'Please select a bank'
    return
  }

  const amount_received_usd = (editForm.value.amount_sending_da / editForm.value.rate).toFixed(2)

  const result = await callApi({
    query: `
      UPDATE transfers 
      SET amount_sending_da = ?, 
          rate = ?,
          amount_received_usd = ?,
          notes = ?,
          id_bank = ?,
          date_do_transfer = ?,
          ref_pi_transfer = ?
      WHERE id = ? ${!isAdmin.value ? 'AND date_receive IS NULL' : ''}
    `,
    params: [
      editForm.value.amount_sending_da,
      editForm.value.rate,
      amount_received_usd,
      editForm.value.notes || null,
      editForm.value.id_bank,
      editForm.value.date_do_transfer,
      editForm.value.ref_pi_transfer || null,
      selectedTransfer.value.id,
    ],
  })

  if (result.success) {
    showEditDialog.value = false
    selectedTransfer.value = null
    fetchTransfers()
  }
}

const deleteTransfer = async (transfer) => {
  if (!isAdmin.value) {
    error.value = 'Only admin can delete transfers'
    return
  }

  if (!confirm('Are you sure you want to delete this transfer?')) return

  const result = await callApi({
    query: 'DELETE FROM transfers WHERE id = ?',
    params: [transfer.id],
  })

  if (result.success) {
    fetchTransfers()
  }
}

const fetchBanks = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM banks ORDER BY company_name',
      params: [],
    })
    if (result.success) {
      banks.value = result.data
    }
  } catch (err) {
    console.error('Error fetching banks:', err)
  }
}

const createTransfer = async () => {
  try {
    formError.value = null

    const amount = parseFloat(newTransfer.value.amount_sending_da)
    const rate = parseFloat(newTransfer.value.rate)

    if (isNaN(amount) || amount <= 0) {
      formError.value = 'Please enter a valid amount'
      return
    }

    if (isNaN(rate) || rate <= 0) {
      formError.value = 'Please enter a valid rate'
      return
    }

    if (!newTransfer.value.id_bank) {
      formError.value = 'Please select a bank'
      return
    }

    if (!newTransfer.value.date_do_transfer) {
      formError.value = 'Please select a date'
      return
    }

    const amount_received_usd = (amount / rate).toFixed(2)

    const result = await callApi({
      query: `
        INSERT INTO transfers (
          id_user_do_transfer, date_do_transfer, amount_sending_da, 
          rate, amount_received_usd, notes, id_bank, ref_pi_transfer
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `,
      params: [
        user.value.id,
        newTransfer.value.date_do_transfer,
        amount,
        rate,
        amount_received_usd,
        newTransfer.value.notes || null,
        newTransfer.value.id_bank,
        newTransfer.value.ref_pi_transfer || null,
      ],
    })

    if (result.success) {
      showAddDialog.value = false
      newTransfer.value = {
        amount_sending_da: '',
        rate: '',
        notes: '',
        id_bank: null,
        date_do_transfer: new Date().toISOString().split('T')[0],
        ref_pi_transfer: '',
      }
      await fetchTransfers()
    } else {
      formError.value = result.error || 'Failed to create transfer'
    }
  } catch (err) {
    formError.value = err.message || 'An error occurred while creating the transfer'
    console.error('Create transfer error:', err)
  }
}

const selectedBank = computed(() => {
  if (!newTransfer.value.id_bank) return null
  return banks.value.find((bank) => bank.id === newTransfer.value.id_bank)
})

const selectedBankInEdit = computed(() => {
  if (!editForm.value.id_bank) return null
  return banks.value.find((bank) => bank.id === editForm.value.id_bank)
})

const openDetailsDialog = (transfer) => {
  console.log('Opening details for transfer:', transfer)
  console.log('Transfer ID:', transfer.id, typeof transfer.id)
  selectedTransferForDetails.value = transfer
  showDetailsDialog.value = true
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchTransfers()
    fetchBanks()
  }
})
</script>

<template>
  <div class="sender-view">
    <div class="header">
      <h1>My Transfers</h1>
      <div class="header-buttons">
        <button @click="showAddDialog = true" class="btn create-btn">New Transfer</button>
        <button @click="router.push('/transfers')" class="back-btn">← Return to Transfers</button>
      </div>
    </div>

    <!-- Add filters section before table -->
    <div class="filters-section">
      <div class="filters-grid">
        <div class="filter-group">
          <label><i class="fas fa-university"></i> Bank</label>
          <input
            type="text"
            v-model="filters.bank"
            class="filter-input"
            placeholder="Filter by bank..."
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
          <label><i class="fas fa-tasks"></i> Status</label>
          <select v-model="filters.status" class="filter-input">
            <option value="">All</option>
            <option value="received">Received</option>
            <option value="pending">Pending</option>
          </select>
        </div>
        <div class="filter-group">
          <label><i class="fas fa-money-bill"></i> Min Amount (DA)</label>
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
          <label><i class="fas fa-money-bill"></i> Max Amount (DA)</label>
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
      <table>
        <thead>
          <tr>
            <th @click="toggleSort('date_do_transfer')" class="sortable">
              Date Sent
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
            <th @click="toggleSort('amount_sending_da')" class="sortable">
              Amount (DA)
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'amount_sending_da'
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
            <th>Bank</th>
            <th>Account</th>
            <th>Notes</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="transfer in filteredTransfers" :key="transfer.id">
            <td>{{ new Date(transfer.date_do_transfer).toLocaleString() }}</td>
            <td>{{ transfer.ref_pi_transfer || '-' }}</td>
            <td>{{ transfer.amount_sending_da }}</td>
            <td>{{ transfer.rate }}</td>
            <td>
              <div v-if="transfer.company_name" class="bank-cell">
                <strong>{{ transfer.company_name }}</strong
                ><br />
                <small>{{ transfer.bank_name }}</small>
              </div>
              <span v-else>-</span>
            </td>
            <td>
              <div v-if="transfer.bank_account" class="bank-cell">
                {{ transfer.bank_account }}<br />
                <small class="swift">{{ transfer.swift_code }}</small>
              </div>
              <span v-else>-</span>
            </td>
            <td>{{ transfer.notes || '-' }}</td>
            <td class="status-cell">
              <span v-if="transfer.date_receive" class="status received">✅</span>
              <span v-else class="status pending">⏳</span>
            </td>
            <td>
              <div class="action-buttons">
                <button
                  @click="openDetailsDialog(transfer)"
                  class="btn details-btn"
                  title="View Details"
                >
                  <i class="fas fa-list-ul"></i>
                </button>
                <button
                  v-if="!transfer.date_receive || isAdmin"
                  @click="openEditDialog(transfer)"
                  class="btn edit-btn"
                  title="Edit Transfer"
                >
                  <i class="fas fa-edit"></i>
                </button>
                <button
                  v-if="isAdmin"
                  @click="deleteTransfer(transfer)"
                  class="btn delete-btn"
                  title="Delete Transfer"
                >
                  <i class="fas fa-trash"></i>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Edit Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h2>Edit Transfer</h2>
        <div v-if="error" class="error-message">
          {{ error }}
        </div>
        <form @submit.prevent="updateTransfer">
          <div class="form-group">
            <label>Date:</label>
            <input type="date" v-model="editForm.date_do_transfer" class="input-field" required />
          </div>
          <div class="form-group">
            <label>Amount (DA):</label>
            <input
              type="number"
              v-model.number="editForm.amount_sending_da"
              class="input-field"
              step="0.01"
              min="0.01"
              required
            />
          </div>
          <div class="form-group">
            <label>Rate:</label>
            <input
              type="number"
              v-model.number="editForm.rate"
              class="input-field"
              step="0.0001"
              min="0.0001"
              required
            />
          </div>
          <div class="form-group">
            <label>Bank Account:</label>
            <select v-model="editForm.id_bank" class="input-field" required>
              <option value="">Select a bank account</option>
              <option v-for="bank in banks" :key="bank.id" :value="bank.id">
                {{ bank.company_name }} - {{ bank.bank_name }} ({{ bank.bank_account }})
              </option>
            </select>
          </div>

          <div v-if="editForm.id_bank" class="bank-details">
            <div class="bank-info">
              <p><strong>Selected Bank Details:</strong></p>
              <p v-if="selectedBankInEdit">
                Company: {{ selectedBankInEdit.company_name }}<br />
                Bank: {{ selectedBankInEdit.bank_name }}<br />
                Account: {{ selectedBankInEdit.bank_account }}<br />
                Swift: {{ selectedBankInEdit.swift_code }}<br />
                Address: {{ selectedBankInEdit.bank_address }}
              </p>
            </div>
          </div>

          <div class="form-group">
            <label>Notes:</label>
            <textarea
              v-model="editForm.notes"
              class="input-field"
              placeholder="Optional notes"
            ></textarea>
          </div>
          <div class="form-group">
            <label><i class="fas fa-hashtag"></i> PI Reference:</label>
            <input
              type="text"
              v-model="editForm.ref_pi_transfer"
              class="input-field"
              placeholder="Enter PI reference..."
            />
          </div>
          <div class="dialog-actions">
            <button type="submit" class="btn update-btn">Save</button>
            <button type="button" @click="showEditDialog = false" class="btn cancel-btn">
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
    <!-- Add Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h2>New Transfer</h2>
        <div v-if="formError" class="error-message">
          {{ formError }}
        </div>
        <form @submit.prevent="createTransfer">
          <div class="form-group">
            <label>Date:</label>
            <input
              type="date"
              v-model="newTransfer.date_do_transfer"
              class="input-field"
              required
            />
          </div>
          <div class="form-group">
            <label>Amount (DA):</label>
            <input
              type="number"
              v-model.number="newTransfer.amount_sending_da"
              class="input-field"
              placeholder="Enter amount in DA"
              step="0.01"
              min="0.01"
              required
            />
          </div>
          <div class="form-group">
            <label>Rate:</label>
            <input
              type="number"
              v-model.number="newTransfer.rate"
              class="input-field"
              placeholder="Enter exchange rate"
              step="0.0001"
              min="0.0001"
              required
            />
          </div>
          <div class="form-group">
            <label>Bank Account:</label>
            <select v-model="newTransfer.id_bank" class="input-field" required>
              <option value="">Select a bank account</option>
              <option v-for="bank in banks" :key="bank.id" :value="bank.id">
                {{ bank.company_name }} - {{ bank.bank_name }} ({{ bank.bank_account }})
              </option>
            </select>
          </div>

          <div v-if="newTransfer.id_bank" class="bank-details">
            <div class="bank-info">
              <p><strong>Selected Bank Details:</strong></p>
              <p v-if="selectedBank">
                Company: {{ selectedBank.company_name }}<br />
                Bank: {{ selectedBank.bank_name }}<br />
                Account: {{ selectedBank.bank_account }}<br />
                Swift: {{ selectedBank.swift_code }}<br />
                Address: {{ selectedBank.bank_address }}
              </p>
            </div>
          </div>

          <div class="form-group">
            <label>Notes:</label>
            <textarea
              v-model="newTransfer.notes"
              class="input-field"
              placeholder="Optional notes"
            ></textarea>
          </div>
          <div class="form-group">
            <label><i class="fas fa-hashtag"></i> PI Reference:</label>
            <input
              type="text"
              v-model="newTransfer.ref_pi_transfer"
              class="input-field"
              placeholder="Enter PI reference..."
            />
          </div>
          <div class="dialog-actions">
            <button type="submit" class="btn create-btn">Create</button>
            <button type="button" @click="showAddDialog = false" class="btn cancel-btn">
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
    <TransferDetails
      v-if="showDetailsDialog"
      :transfer-id="selectedTransferForDetails?.id"
      :is-visible="showDetailsDialog"
      @close="showDetailsDialog = false"
    />
  </div>
</template>

<style scoped>
.sender-view {
  width: 80vw;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.back-btn {
  padding: 8px 16px;
  background-color: #4b5563;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.transfers-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
  table-layout: fixed;
}

th,
td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

/* Column widths */
th:nth-child(1) {
  width: 12%;
} /* Date */
th:nth-child(2) {
  width: 10%;
} /* PI Reference */
th:nth-child(3) {
  width: 10%;
} /* Amount DA */
th:nth-child(4) {
  width: 8%;
} /* Rate */
th:nth-child(5) {
  width: 15%;
} /* Bank */
th:nth-child(6) {
  width: 15%;
} /* Account */
th:nth-child(7) {
  width: 15%;
} /* Notes */
th:nth-child(8) {
  width: 7%;
} /* Status */
th:nth-child(9) {
  width: 10%;
} /* Actions */

/* Notes cell specific styling */
td:nth-child(7) {
  white-space: pre-wrap;
  min-width: 200px;
}

th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.status-cell {
  text-align: center;
}

.status {
  font-size: 1.2em;
}

.received {
  color: #4caf50;
}

.pending {
  color: #ffa500;
}

.edit-input {
  width: 100px;
  padding: 4px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
}

.update-btn {
  background-color: #2196f3;
  color: white;
}

.delete-btn {
  background-color: #f44336;
  color: white;
}

.btn:hover {
  opacity: 0.9;
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
}

.dialog {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  min-width: 300px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  color: #4b5563;
}

.input-field {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}
.header-buttons {
  display: flex;
  gap: 10px;
}

.create-btn {
  background-color: #4caf50;
  color: white;
}

textarea.input-field {
  min-height: 100px;
  resize: vertical;
}

.error-message {
  padding: 12px;
  margin-bottom: 16px;
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  border-radius: 4px;
  color: #b91c1c;
}

.bank-details {
  margin: 15px 0;
  padding: 15px;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
}

.bank-info {
  font-size: 0.9em;
  line-height: 1.5;
}

.bank-info p {
  margin: 0 0 10px 0;
}

.bank-info strong {
  color: #4b5563;
}

select.input-field {
  appearance: none;
  background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 10px center;
  background-size: 1em;
  padding-right: 40px;
}

.bank-cell {
  font-size: 0.9em;
  line-height: 1.4;
}

.bank-cell strong {
  color: #1f2937;
}

.bank-cell small {
  color: #6b7280;
}

.bank-cell .swift {
  color: #9ca3af;
  font-family: monospace;
}

.details-btn {
  background: #6366f1;
  color: white;
  padding: 6px 10px;
}

.action-buttons {
  display: flex;
  gap: 6px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn i {
  font-size: 0.9em;
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
  height: 38px;
  align-self: flex-end;
}

.clear-btn:hover {
  background-color: #4b5563;
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

@media (max-width: 768px) {
  .filters-grid {
    grid-template-columns: 1fr;
  }
}
</style>
