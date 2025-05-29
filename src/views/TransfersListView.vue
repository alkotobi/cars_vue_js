<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'
import TransferDetails from '../components/TransferDetails.vue'

const { callApi } = useApi()
const transfers = ref([])
const showEditDialog = ref(false)
const selectedTransfer = ref(null)
const banks = ref([])
const formError = ref(null)
const isLoading = ref(false)
const processingTransferId = ref(null)
const showDetailsDialog = ref(false)
const selectedTransferForDetails = ref(null)
const user = ref(null)

const editForm = ref({
  amount_sending_da: '',
  rate: '',
  notes: '',
  amount_received_usd: '',
  receiver_notes: '',
  id_bank: null,
  ref_pi_transfer: '',
})

const filters = ref({
  sender: '',
  receiver: '',
  bank: '',
  dateFrom: '',
  dateTo: '',
  status: '', // 'received', 'pending', or ''
})

const sortConfig = ref({
  field: 'date_do_transfer',
  direction: 'desc',
})

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

const fetchTransfers = async () => {
  isLoading.value = true
  try {
    const result = await callApi({
      query: `
        SELECT t.*, 
          u_sender.username as sender_name,
          u_receiver.username as receiver_name,
          b.company_name, b.bank_name, b.bank_account, b.swift_code
        FROM transfers t
        LEFT JOIN users u_sender ON t.id_user_do_transfer = u_sender.id
        LEFT JOIN users u_receiver ON t.id_user_receive_transfer = u_receiver.id
        LEFT JOIN banks b ON t.id_bank = b.id
        ORDER BY t.date_do_transfer DESC
      `,
      params: [],
    })
    if (result.success) {
      transfers.value = result.data
    }
  } catch (error) {
    console.error('Error fetching transfers:', error)
  } finally {
    isLoading.value = false
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
    amount_received_usd: transfer.amount_received_usd || '',
    receiver_notes: transfer.receiver_notes || '',
    id_bank: transfer.id_bank || null,
    ref_pi_transfer: transfer.ref_pi_transfer || '',
  }
  showEditDialog.value = true
}

const updateTransfer = async () => {
  if (processingTransferId.value) return // Prevent double submit
  processingTransferId.value = selectedTransfer.value.id
  try {
    formError.value = null

    // Validate inputs
    const amount = parseFloat(editForm.value.amount_sending_da)
    const rate = parseFloat(editForm.value.rate)
    const receivedUsd = editForm.value.amount_received_usd
      ? parseFloat(editForm.value.amount_received_usd)
      : null

    if (isNaN(amount) || amount <= 0) {
      formError.value = 'Please enter a valid amount'
      return
    }

    if (isNaN(rate) || rate <= 0) {
      formError.value = 'Please enter a valid rate'
      return
    }

    if (!editForm.value.id_bank) {
      formError.value = 'Please select a bank'
      return
    }

    if (receivedUsd !== null && isNaN(receivedUsd)) {
      formError.value = 'Please enter a valid received USD amount'
      return
    }

    const amount_received_usd = receivedUsd || (amount / rate).toFixed(2)

    const result = await callApi({
      query: `
        UPDATE transfers SET
          amount_sending_da = ?,
          rate = ?,
          notes = ?,
          amount_received_usd = ?,
          receiver_notes = ?,
          id_bank = ?,
          ref_pi_transfer = ?
        WHERE id = ?
      `,
      params: [
        amount,
        rate,
        editForm.value.notes || null,
        amount_received_usd,
        editForm.value.receiver_notes || null,
        editForm.value.id_bank,
        editForm.value.ref_pi_transfer || null,
        selectedTransfer.value.id,
      ],
    })

    if (result.success) {
      showEditDialog.value = false
      await fetchTransfers()
    } else {
      formError.value = result.error || 'Failed to update transfer'
    }
  } catch (err) {
    console.error('Update transfer error:', err)
    formError.value = err.message || 'An error occurred while updating the transfer'
  } finally {
    processingTransferId.value = null
  }
}

const selectedBankInEdit = computed(() => {
  if (!editForm.value.id_bank) return null
  return banks.value.find((bank) => bank.id === editForm.value.id_bank)
})

const canEditDetails = computed(() => {
  if (!user.value || !selectedTransferForDetails.value) return false
  return (
    user.value.role_id === 1 || // Admin
    user.value.id === selectedTransferForDetails.value.id_user_do_transfer // Transfer creator
  )
})

const openDetailsDialog = (transfer) => {
  console.log('Opening details for transfer:', transfer)
  selectedTransferForDetails.value = transfer
  showDetailsDialog.value = true
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchTransfers()
  fetchBanks()
})

const deleteTransfer = async (transfer) => {
  if (processingTransferId.value) return // Prevent double delete
  if (!confirm('Are you sure you want to delete this transfer permanently?')) return

  processingTransferId.value = transfer.id
  try {
    const result = await callApi({
      query: `DELETE FROM transfers WHERE id = ?`,
      params: [transfer.id],
    })

    if (result.success) {
      await fetchTransfers()
    }
  } finally {
    processingTransferId.value = null
  }
}

// Add isAdmin computed property after other computed properties
const isAdmin = computed(() => user.value?.role_id === 1)

const filteredTransfers = computed(() => {
  let result = [...transfers.value]

  // Apply filters
  if (filters.value.sender) {
    result = result.filter((t) =>
      t.sender_name?.toLowerCase().includes(filters.value.sender.toLowerCase()),
    )
  }

  if (filters.value.receiver) {
    result = result.filter((t) =>
      t.receiver_name?.toLowerCase().includes(filters.value.receiver.toLowerCase()),
    )
  }

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
    sender: '',
    receiver: '',
    bank: '',
    dateFrom: '',
    dateTo: '',
    status: '',
  }
}
</script>

<template>
  <div class="transfers-list-view">
    <h1>All Transfers List</h1>
    <button @click="$router.push('/transfers')" class="back-btn">← Return to Transfers</button>

    <div class="filters-section">
      <div class="filters-grid">
        <div class="filter-group">
          <label><i class="fas fa-user"></i> Sender</label>
          <input
            type="text"
            v-model="filters.sender"
            class="filter-input"
            placeholder="Filter by sender..."
          />
        </div>
        <div class="filter-group">
          <label><i class="fas fa-user-check"></i> Receiver</label>
          <input
            type="text"
            v-model="filters.receiver"
            class="filter-input"
            placeholder="Filter by receiver..."
          />
        </div>
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
        <button @click="clearFilters" class="btn clear-btn">
          <i class="fas fa-times"></i> Clear Filters
        </button>
      </div>
    </div>

    <div class="transfers-table">
      <div v-if="isLoading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin"></i>
        Loading transfers...
      </div>
      <table>
        <thead>
          <tr>
            <th @click="toggleSort('sender_name')" class="sortable">
              Sender
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'sender_name'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
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
              Amount DA
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
            <th>Sender Notes</th>
            <th @click="toggleSort('receiver_name')" class="sortable">
              Receiver
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'receiver_name'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th @click="toggleSort('date_receive')" class="sortable">
              Date Received
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'date_receive'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th @click="toggleSort('amount_received_usd')" class="sortable">
              Received USD
              <i
                :class="[
                  'fas',
                  sortConfig.field === 'amount_received_usd'
                    ? sortConfig.direction === 'asc'
                      ? 'fa-sort-up'
                      : 'fa-sort-down'
                    : 'fa-sort',
                ]"
              ></i>
            </th>
            <th>Receiver Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="transfer in filteredTransfers"
            :key="transfer.id"
            :class="{
              'not-received': !transfer.date_receive,
              'amount-mismatch':
                transfer.date_receive &&
                calculateUSD(transfer.amount_sending_da, transfer.rate) !==
                  transfer.amount_received_usd,
            }"
          >
            <td>{{ transfer.sender_name }}</td>
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
            <td>{{ transfer.receiver_name || '-' }}</td>
            <td>
              {{ transfer.date_receive ? new Date(transfer.date_receive).toLocaleString() : '-' }}
            </td>
            <td>{{ transfer.amount_received_usd ? `$${transfer.amount_received_usd}` : '-' }}</td>
            <td>{{ transfer.receiver_notes || '-' }}</td>
            <td>
              <div class="action-buttons">
                <button
                  :disabled="!isAdmin"
                  @click="openEditDialog(transfer)"
                  class="btn edit-btn"
                  :class="{ disabled: !isAdmin }"
                >
                  <i class="fas fa-edit"></i>
                  <span v-if="processingTransferId === transfer.id">
                    <i class="fas fa-spinner fa-spin"></i>
                  </span>
                  <span v-else>Edit</span>
                </button>
                <button
                  :disabled="!isAdmin"
                  @click="deleteTransfer(transfer)"
                  class="btn delete-btn"
                  :class="{ disabled: !isAdmin }"
                >
                  <i class="fas fa-trash"></i>
                  <span v-if="processingTransferId === transfer.id">
                    <i class="fas fa-spinner fa-spin"></i>
                  </span>
                  <span v-else>Delete</span>
                </button>
                <button
                  @click="openDetailsDialog(transfer)"
                  class="btn details-btn"
                  title="View Details"
                >
                  <i class="fas fa-list-ul"></i>
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
        <div v-if="formError" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ formError }}
        </div>
        <form @submit.prevent="updateTransfer">
          <div class="form-grid">
            <!-- Left Column -->
            <div class="form-column">
              <div class="form-group">
                <label>Amount (DA):</label>
                <input
                  type="number"
                  v-model.number="editForm.amount_sending_da"
                  step="0.01"
                  min="0.01"
                  required
                  class="input-field"
                />
              </div>
              <div class="form-group">
                <label>Rate:</label>
                <input
                  type="number"
                  v-model.number="editForm.rate"
                  step="0.0001"
                  min="0.0001"
                  required
                  class="input-field"
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
              <div class="form-group">
                <label>Notes:</label>
                <textarea v-model="editForm.notes" class="input-field" />
              </div>
            </div>

            <!-- Right Column -->
            <div class="form-column">
              <div class="form-group">
                <label>Received USD:</label>
                <input
                  type="number"
                  v-model.number="editForm.amount_received_usd"
                  step="0.01"
                  min="0"
                  class="input-field"
                />
              </div>
              <div class="form-group">
                <label>Receiver Notes:</label>
                <textarea v-model="editForm.receiver_notes" class="input-field" />
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
            </div>
          </div>

          <!-- Bank Details Section - Full Width -->
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

          <div class="dialog-actions">
            <button
              type="submit"
              class="btn save-btn"
              :disabled="processingTransferId === selectedTransfer?.id"
            >
              <i class="fas fa-save"></i>
              <span v-if="processingTransferId === selectedTransfer?.id">
                <i class="fas fa-spinner fa-spin"></i> Saving...
              </span>
              <span v-else>Save</span>
            </button>
            <button
              type="button"
              @click="showEditDialog = false"
              class="btn cancel-btn"
              :disabled="processingTransferId === selectedTransfer?.id"
            >
              <i class="fas fa-times"></i> Cancel
            </button>
          </div>
        </form>
      </div>
    </div>

    <TransferDetails
      v-if="showDetailsDialog"
      :transfer-id="selectedTransferForDetails?.id"
      :is-visible="showDetailsDialog"
      :read-only="!canEditDetails"
      @close="showDetailsDialog = false"
    />
  </div>
</template>

<style scoped>
.transfers-list-view {
  padding: 30px;
  width: 95%;
  max-width: 1600px;
  margin: 0 auto;
}

h1 {
  color: #1f2937;
  font-size: 2em;
  margin-bottom: 1.5rem;
  font-weight: 600;
}

.back-btn {
  padding: 10px 20px;
  background-color: #4b5563;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  margin: 20px 0;
  font-size: 0.95em;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.back-btn:hover {
  background-color: #374151;
  transform: translateX(-2px);
}

.transfers-table {
  position: relative;
  background: white;
  border-radius: 12px;
  box-shadow:
    0 4px 6px -1px rgba(0, 0, 0, 0.1),
    0 2px 4px -1px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  margin: 20px 0;
}

table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 0.95em;
}

th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #1f2937;
  padding: 16px;
  text-align: left;
  border-bottom: 2px solid #e5e7eb;
  white-space: nowrap;
}

td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
  color: #4b5563;
}

tr:hover td {
  background-color: #f8fafc;
}

.bank-cell {
  font-size: 0.95em;
  line-height: 1.5;
}

.bank-cell strong {
  color: #1f2937;
  font-weight: 600;
}

.bank-cell small {
  color: #6b7280;
}

.bank-cell .swift {
  color: #64748b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  background: #f1f5f9;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.9em;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9em;
  font-weight: 500;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.btn.edit-btn:hover {
  background-color: #2563eb;
}

.btn.delete-btn {
  background-color: #ef4444;
  color: white;
}

.btn.delete-btn:hover {
  background-color: #dc2626;
}

.btn.save-btn {
  background-color: #10b981;
  color: white;
}

.btn.save-btn:hover {
  background-color: #059669;
}

.btn.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.btn.cancel-btn:hover {
  background-color: #4b5563;
}

.not-received {
  background-color: #fef2f2;
}

.amount-mismatch {
  background-color: #fff1f2;
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
  backdrop-filter: blur(4px);
}

.dialog {
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  overflow-y: auto;
  background-color: white;
  padding: 30px;
  border-radius: 12px;
  box-shadow:
    0 20px 25px -5px rgba(0, 0, 0, 0.1),
    0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.dialog h2 {
  color: #1f2937;
  font-size: 1.5em;
  margin-bottom: 1.5rem;
  font-weight: 600;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
}

.form-column {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #4b5563;
  font-weight: 500;
}

.input-field {
  width: 100%;
  padding: 8px 12px;
}

textarea.input-field {
  min-height: 100px;
  resize: vertical;
}

.bank-details {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
  margin: 20px 0;
}

.bank-info p {
  margin: 0 0 12px 0;
  line-height: 1.6;
}

.error-message {
  background-color: #fef2f2;
  border: 1px solid #fee2e2;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: #dc2626;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 12px;
  font-size: 1.1em;
  color: #4b5563;
  backdrop-filter: blur(2px);
}

.loading-overlay i {
  font-size: 2.5em;
  color: #3b82f6;
}

/* Money values styling */
td:nth-child(3), /* Amount DA */
td:nth-child(5), /* USD Sent */
td:nth-child(11) /* Received USD */ {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-weight: 500;
}

/* Date styling */
td:nth-child(2), /* Date Sent */
td:nth-child(10) /* Date Received */ {
  white-space: nowrap;
  color: #64748b;
}

@media (max-width: 1400px) {
  .transfers-list-view {
    width: 98%;
    padding: 20px;
  }

  td,
  th {
    padding: 12px;
  }
}

@media print {
  .transfers-list-view {
    width: 100%;
    padding: 0;
  }

  .back-btn,
  .action-buttons,
  .loading-overlay {
    display: none !important;
  }

  .transfers-table {
    box-shadow: none;
  }

  table {
    font-size: 11pt;
  }
}

.details-btn {
  background: #6366f1;
  color: white;
  padding: 6px 10px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn i {
  font-size: 0.9em;
}

.btn.disabled {
  opacity: 0.6;
  cursor: not-allowed;
  pointer-events: none;
}

.btn.edit-btn.disabled {
  background-color: #3b82f6;
}

.btn.delete-btn.disabled {
  background-color: #ef4444;
}

.btn.disabled:hover {
  transform: none;
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

  .form-grid {
    grid-template-columns: 1fr;
    gap: 15px;
  }

  .dialog {
    padding: 20px;
    width: 95%;
  }
}
</style>
