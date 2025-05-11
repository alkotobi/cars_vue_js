<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const router = useRouter()
const { callApi, error } = useApi()
const user = ref(null)
const transfers = ref([])
const showReceiveDialog = ref(false)
const selectedTransfer = ref(null)
const receiveForm = ref({
  amount_received_usd: '',
  receiver_notes: ''
})

const fetchTransfers = async () => {
  const result = await callApi({
    query: `
      SELECT t.*, u.username as sender_name
      FROM transfers t
      JOIN users u ON t.id_user_do_transfer = u.id
      WHERE t.date_receive IS NULL
      ORDER BY t.date_do_transfer DESC
    `,
    params: []
  })
  if (result.success) {
    transfers.value = result.data
  }
}

const calculateUSD = (amount, rate) => {
  return (amount / rate).toFixed(2)
}

const openReceiveDialog = (transfer) => {
  selectedTransfer.value = transfer
  receiveForm.value = {
    amount_received_usd: calculateUSD(transfer.amount_sending_da, transfer.rate),
    receiver_notes: ''
  }
  showReceiveDialog.value = true
}

const receiveTransfer = async () => {
  if (!receiveForm.value.amount_received_usd) {
    error.value = 'Please fill all required fields'
    return
  }

  const result = await callApi({
    query: `
      UPDATE transfers 
      SET amount_received_usd = ?,
          receiver_notes = ?,
          date_receive = NOW(),
          id_user_receive_transfer = ?
      WHERE id = ? AND date_receive IS NULL
    `,
    params: [
      receiveForm.value.amount_received_usd,
      receiveForm.value.receiver_notes || null,
      user.value.id,
      selectedTransfer.value.id
    ]
  })

  if (result.success) {
    showReceiveDialog.value = false
    selectedTransfer.value = null
    fetchTransfers()
    fetchRecentTransfers()  // Add this line to update recent transfers table
  }
  else {
    error.value = result.message
  }
}

const recentTransfers = ref([])

const fetchRecentTransfers = async () => {
  const result = await callApi({
    query: `
      SELECT t.*, u.username as sender_name
      FROM transfers t
      JOIN users u ON t.id_user_do_transfer = u.id
      WHERE t.date_receive IS NOT NULL
      AND t.date_receive >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
      ORDER BY t.date_receive DESC
    `,
    params: []
  })
  if (result.success) {
    recentTransfers.value = result.data
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchTransfers()
    fetchRecentTransfers()
  }
})

const unreceiveTransfer = async (transfer) => {
  if (!confirm('Are you sure you want to mark this transfer as unreceived?')) return

  const result = await callApi({
    query: `
      UPDATE transfers 
      SET amount_received_usd = NULL,
          receiver_notes = NULL,
          date_receive = NULL,
          id_user_receive_transfer = NULL
      WHERE id = ?
    `,
    params: [transfer.id]
  })

  if (result.success) {
    fetchTransfers()
    fetchRecentTransfers()
  } else {
    error.value = result.message
  }
}
</script>

<template>
  <div class="receiver-view">
    <div class="header">
      <h1>All Pending Transfers</h1>
      <button @click="router.push('/transfers')" class="back-btn">← Return to Transfers</button>
    </div>

    <!-- Transfers Table -->
    <div class="transfers-table">
      <table>
        <thead>
          <tr>
            <th>Sender</th>
            <th>Date Sent</th>
            <th>Sent USD</th>
            <th>Received USD</th>
            <th>Notes</th>
            <th>Receiver Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="transfer in transfers" :key="transfer.id">
            <td>{{ transfer.sender_name }}</td>
            <td>{{ new Date(transfer.date_do_transfer).toLocaleString() }}</td>
            <td>${{ calculateUSD(transfer.amount_sending_da, transfer.rate) }}</td>
            <td>${{ transfer.amount_received_usd || '-' }}</td>
            <td>{{ transfer.notes || '-' }}</td>
            <td>{{ transfer.receiver_notes || '-' }}</td>
            <td>
              <button 
                @click="openReceiveDialog(transfer)" 
                class="btn receive-btn"
              >
                Receive
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Recent Received Transfers Table -->
    <h2 class="section-title">Recently Received Transfers (Past Week)</h2>
    <div class="transfers-table">
      <table>
        <thead>
          <tr>
            <th>Sender</th>
            <th>Date Received</th>
            <th>Sent USD</th>
            <th>Received USD</th>
            <th>Notes</th>
            <th>Receiver Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="transfer in recentTransfers" 
              :key="transfer.id"
              :class="{ 'amount-mismatch': calculateUSD(transfer.amount_sending_da, transfer.rate) !== transfer.amount_received_usd }"
          >
            <td>{{ transfer.sender_name }}</td>
            <td>{{ new Date(transfer.date_receive).toLocaleString() }}</td>
            <td>${{ calculateUSD(transfer.amount_sending_da, transfer.rate) }}</td>
            <td>${{ transfer.amount_received_usd }}</td>
            <td>{{ transfer.notes || '-' }}</td>
            <td>{{ transfer.receiver_notes || '-' }}</td>
            <td>
              <button 
                @click="unreceiveTransfer(transfer)" 
                class="btn unreceive-btn"
              >
                Unreceive
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Receive Dialog -->
    <div v-if="showReceiveDialog" class="dialog-overlay">
      <div class="dialog">
        <h2>Receive Transfer</h2>
        <div class="form-group">
          <label>Amount Received (USD):</label>
          <input 
            type="number" 
            v-model="receiveForm.amount_received_usd"
            class="input-field"
            step="0.01"
          />
        </div>
        <div class="form-group">
          <label>Receiver Notes:</label>
          <textarea 
            v-model="receiveForm.receiver_notes" 
            class="input-field"
            placeholder="Add your notes here"
          ></textarea>
        </div>
        <div class="dialog-actions">
          <button @click="receiveTransfer" class="btn receive-btn">Confirm</button>
          <button @click="showReceiveDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.receiver-view {
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

th, td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

/* Column widths */
th:nth-child(1) { width: 12%; } /* Sender */
th:nth-child(2) { width: 15%; } /* Date */
th:nth-child(3) { width: 10%; } /* Sent USD */
th:nth-child(4) { width: 10%; } /* Received USD */
th:nth-child(5) { width: 20%; } /* Notes */
th:nth-child(6) { width: 20%; } /* Receiver Notes */
th:nth-child(7) { width: 13%; } /* Actions */

/* Notes cell specific styling */
td:nth-child(5),
td:nth-child(6) {
  white-space: pre-wrap;
  min-width: 150px;
}

th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
}

.btn:hover {
  opacity: 0.9;
}

.receive-btn {
  background-color: #4CAF50;
  color: white;
}

.unreceive-btn {
  background-color: #DC2626;
  color: white;
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

textarea.input-field {
  min-height: 100px;
  resize: vertical;
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

.section-title {
  margin: 30px 0 20px;
  color: #374151;
  font-size: 1.5rem;
}

/* Recent transfers table specific widths */
.transfers-table:nth-of-type(2) table th:nth-child(1) { width: 15%; } /* Sender */
.transfers-table:nth-of-type(2) table th:nth-child(2) { width: 15%; } /* Date */
.transfers-table:nth-of-type(2) table th:nth-child(3) { width: 12%; } /* Sent USD */
.transfers-table:nth-of-type(2) table th:nth-child(4) { width: 12%; } /* Received USD */
.transfers-table:nth-of-type(2) table th:nth-child(5) { width: 23%; } /* Notes */
.transfers-table:nth-of-type(2) table th:nth-child(6) { width: 23%; } /* Receiver Notes */

/* Notes cell specific styling for recent transfers */
.transfers-table:nth-of-type(2) table td:nth-child(5),
.transfers-table:nth-of-type(2) table td:nth-child(6) {
  white-space: pre-wrap;
  min-width: 150px;
}

/* Add this new style for amount mismatch highlighting */
.amount-mismatch {
  background-color: #FEF2F2;  /* Light red background */
  color: #991B1B;  /* Darker red text for contrast */
}

.amount-mismatch td {
  border-bottom-color: #FCA5A5;  /* Reddish border */
}


@media print {
  .header button,
  .btn {
    display: none !important;
  }

  table {
    width: 100% !important;
    font-size: 10px !important;
    table-layout: auto !important;
  }

  th:nth-child(7),
  td:nth-child(7) {
    display: none !important;
  }

  .amount-mismatch {
    background-color: transparent !important;
    border-left: 4px solid #DC2626 !important;
  }

  .receiver-view {
    width: 100% !important;
    padding: 0 !important;
  }

  .transfers-table {
    overflow: visible !important;
    page-break-inside: avoid;
  }
}
</style>