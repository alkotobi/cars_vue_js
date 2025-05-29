<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

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
})

const newTransfer = ref({
  amount_sending_da: '',
  rate: '',
  notes: '',
})

const formError = ref(null)

const isAdmin = computed(() => user.value?.role_id === 1)

const fetchTransfers = async () => {
  const result = await callApi({
    query: `
      SELECT t.* 
      FROM transfers t
      WHERE t.id_user_do_transfer = ? OR t.id_user_receive_transfer = ?
      ORDER BY t.date_do_transfer DESC
    `,
    params: [user.value.id, user.value.id],
  })
  if (result.success) {
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
  }
  showEditDialog.value = true
}

const updateTransfer = async () => {
  if (!editForm.value.amount_sending_da || !editForm.value.rate) {
    error.value = 'Please fill all required fields'
    return
  }

  const amount_received_usd = (editForm.value.amount_sending_da / editForm.value.rate).toFixed(2)

  const result = await callApi({
    query: `
      UPDATE transfers 
      SET amount_sending_da = ?, 
          rate = ?,
          amount_received_usd = ?,
          notes = ?
      WHERE id = ? ${!isAdmin.value ? 'AND date_receive IS NULL' : ''}
    `,
    params: [
      editForm.value.amount_sending_da,
      editForm.value.rate,
      amount_received_usd,
      editForm.value.notes || null,
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

    const amount_received_usd = (amount / rate).toFixed(2)

    const result = await callApi({
      query: `
        INSERT INTO transfers (
          id_user_do_transfer, date_do_transfer, amount_sending_da, 
          rate, amount_received_usd, notes
        ) VALUES (?, NOW(), ?, ?, ?, ?)
      `,
      params: [user.value.id, amount, rate, amount_received_usd, newTransfer.value.notes || null],
    })

    if (result.success) {
      showAddDialog.value = false
      newTransfer.value = { amount_sending_da: '', rate: '', notes: '' }
      await fetchTransfers()
    } else {
      formError.value = result.error || 'Failed to create transfer'
    }
  } catch (err) {
    formError.value = err.message || 'An error occurred while creating the transfer'
    console.error('Create transfer error:', err)
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchTransfers()
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

    <!-- Transfers Table -->
    <div class="transfers-table">
      <table>
        <thead>
          <tr>
            <th>Date Sent</th>
            <th>Amount (DA)</th>
            <th>Rate</th>
            <th>USD Value</th>
            <th>Notes</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="transfer in transfers" :key="transfer.id">
            <td>{{ new Date(transfer.date_do_transfer).toLocaleString() }}</td>
            <td>{{ transfer.amount_sending_da }}</td>
            <td>{{ transfer.rate }}</td>
            <td>${{ calculateUSD(transfer.amount_sending_da, transfer.rate) }}</td>
            <td>{{ transfer.notes || '-' }}</td>
            <td class="status-cell">
              <span v-if="transfer.date_receive" class="status received">✅</span>
              <span v-else class="status pending">⏳</span>
            </td>
            <td>
              <button
                v-if="isAdmin || !transfer.date_receive"
                @click="openEditDialog(transfer)"
                class="btn update-btn"
              >
                Edit
              </button>
              <button v-if="isAdmin" @click="deleteTransfer(transfer)" class="btn delete-btn">
                Delete
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Edit Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h2>Edit Transfer</h2>
        <div class="form-group">
          <label>Amount (DA):</label>
          <input type="number" v-model="editForm.amount_sending_da" class="input-field" />
        </div>
        <div class="form-group">
          <label>Rate:</label>
          <input type="number" v-model="editForm.rate" class="input-field" />
        </div>
        <div class="form-group">
          <label>Notes:</label>
          <textarea
            v-model="editForm.notes"
            class="input-field"
            placeholder="Optional notes"
          ></textarea>
        </div>
        <div class="dialog-actions">
          <button @click="updateTransfer" class="btn update-btn">Save</button>
          <button @click="showEditDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
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
            <label>Notes:</label>
            <textarea
              v-model="newTransfer.notes"
              class="input-field"
              placeholder="Optional notes"
            ></textarea>
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
  width: 15%;
} /* Date */
th:nth-child(2) {
  width: 12%;
} /* Amount DA */
th:nth-child(3) {
  width: 10%;
} /* Rate */
th:nth-child(4) {
  width: 12%;
} /* USD Value */
th:nth-child(5) {
  width: 25%;
} /* Notes - wider column */
th:nth-child(6) {
  width: 12%;
} /* Status */
th:nth-child(7) {
  width: 14%;
} /* Actions */

/* Notes cell specific styling */
td:nth-child(5) {
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
</style>
