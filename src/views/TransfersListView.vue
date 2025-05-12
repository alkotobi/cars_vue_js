<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi } = useApi()
const transfers = ref([])
const showEditDialog = ref(false)
const selectedTransfer = ref(null)
const editForm = ref({
  amount_sending_da: '',
  rate: '',
  notes: '',
  amount_received_usd: '',
  receiver_notes: ''
})

const fetchTransfers = async () => {
  const result = await callApi({
    query: `
      SELECT t.*, 
        u_sender.username as sender_name,
        u_receiver.username as receiver_name
      FROM transfers t
      LEFT JOIN users u_sender ON t.id_user_do_transfer = u_sender.id
      LEFT JOIN users u_receiver ON t.id_user_receive_transfer = u_receiver.id
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

const openEditDialog = (transfer) => {
  selectedTransfer.value = transfer
  editForm.value = {
    amount_sending_da: transfer.amount_sending_da,
    rate: transfer.rate,
    notes: transfer.notes || '',
    amount_received_usd: transfer.amount_received_usd || '',
    receiver_notes: transfer.receiver_notes || ''
  }
  showEditDialog.value = true
}

const updateTransfer = async () => {
  const result = await callApi({
    query: `
      UPDATE transfers SET
        amount_sending_da = ?,
        rate = ?,
        notes = ?,
        amount_received_usd = ?,
        receiver_notes = ?
      WHERE id = ?
    `,
    params: [
      editForm.value.amount_sending_da,
      editForm.value.rate,
      editForm.value.notes,
      editForm.value.amount_received_usd,
      editForm.value.receiver_notes,
      selectedTransfer.value.id
    ]
  })

  if (result.success) {
    showEditDialog.value = false
    fetchTransfers()
  }
}
onMounted(() => {
  fetchTransfers()
})
const deleteTransfer = async (transfer) => {
  if (!confirm('Are you sure you want to delete this transfer permanently?')) return
  
  const result = await callApi({
    query: `DELETE FROM transfers WHERE id = ?`,
    params: [transfer.id]
  })

  if (result.success) {
    fetchTransfers()
  }
}
</script>

<template>
  <div class="transfers-list-view">
    <h1>All Transfers List</h1>
    <button @click="$router.push('/transfers')" class="back-btn">← Return to Transfers</button>

    <div class="transfers-table">
      <table>
        <thead>
          <tr>
            <th>Sender</th>
            <th>Date Sent</th>
            <th>Amount DA</th>
            <th>Rate</th>
            <th>USD Sent</th>
            <th>Notes</th>
            <th>Receiver</th>
            <th>Date Received</th>
            <th>Received USD</th>
            <th>Receiver Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="transfer in transfers" 
              :key="transfer.id"
              :class="{
                'not-received': !transfer.date_receive,
                'amount-mismatch': transfer.date_receive && calculateUSD(transfer.amount_sending_da, transfer.rate) !== transfer.amount_received_usd
              }">
            <td>{{ transfer.sender_name }}</td>
            <td>{{ new Date(transfer.date_do_transfer).toLocaleString() }}</td>
            <td>{{ transfer.amount_sending_da }}</td>
            <td>{{ transfer.rate }}</td>
            <td>${{ calculateUSD(transfer.amount_sending_da, transfer.rate) }}</td>
            <td>{{ transfer.notes || '-' }}</td>
            <td>{{ transfer.receiver_name || '-' }}</td>
            <td>{{ transfer.date_receive ? new Date(transfer.date_receive).toLocaleString() : '-' }}</td>
            <td>{{ transfer.amount_received_usd ? `$${transfer.amount_received_usd}` : '-' }}</td>
            <td>{{ transfer.receiver_notes || '-' }}</td>
            <td>
              <div class="action-buttons">
                <button @click="openEditDialog(transfer)" class="btn edit-btn">
                  Edit
                </button>
                <button @click="deleteTransfer(transfer)" class="btn delete-btn">
                  Delete
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
        <div class="form-group">
          <label>Amount Sending (DA):</label>
          <input type="number" v-model="editForm.amount_sending_da" step="0.01"/>
        </div>
        <div class="form-group">
          <label>Rate:</label>
          <input type="number" v-model="editForm.rate" step="0.0001"/>
        </div>
        <div class="form-group">
          <label>Notes:</label>
          <textarea v-model="editForm.notes"/>
        </div>
        <div class="form-group">
          <label>Received USD:</label>
          <input type="number" v-model="editForm.amount_received_usd" step="0.01"/>
        </div>
        <div class="form-group">
          <label>Receiver Notes:</label>
          <textarea v-model="editForm.receiver_notes"/>
        </div>
        <div class="dialog-actions">
          <button @click="updateTransfer" class="btn save-btn">Save</button>
          <button @click="showEditDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.transfers-list-view {
  padding: 20px;
  width: 90vw;
}

.transfers-table {
  overflow-x: auto;
  margin: 20px 0;
}

table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

th, td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
  word-wrap: break-word;
}

th {
  background-color: #f8f9fa;
  font-weight: 600;
}

/* Add to existing styles */
.action-buttons {
  display: flex;
  gap: 8px;
}

.btn.delete-btn {
  background-color: #DC2626;
  color: white;
}

/* Adjust action column width */
th:nth-child(11) { width: 9%; }  /* Actions */

/* Adjusted column widths */
th:nth-child(1) { width: 9%; }   /* Sender */
th:nth-child(2) { width: 10%; }  /* Date Sent */
th:nth-child(3) { width: 7%; }   /* Amount DA */
th:nth-child(4) { width: 6%; }   /* Rate */
th:nth-child(5) { width: 7%; }   /* USD Sent */
th:nth-child(6) { width: 12%; }  /* Notes */
th:nth-child(7) { width: 8%; }   /* Receiver */
th:nth-child(8) { width: 10%; }  /* Date Received */
th:nth-child(9) { width: 7%; }   /* Received USD */
th:nth-child(10) { width: 9%; }  /* Receiver Notes */
th:nth-child(11) { width: 7%; }  /* Actions */

.back-btn {
  padding: 8px 16px;
  background-color: #4b5563;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 20px;
}

.btn.edit-btn {
  background-color: #3b82f6;
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
  min-width: 400px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
}

.form-group input,
.form-group textarea {
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

.btn.save-btn {
  background-color: #10b981;
}

/* Add status highlighting */
.not-received {
  background-color: #FFEBEE;  /* Light red for unreceived */
}

.amount-mismatch {
  background-color: #FFCDD2;  /* Darker red for amount mismatch */
}

/* Keep existing styles */
.form-group label {
  display: block;
  margin-bottom: 5px;
}

.form-group input,
.form-group textarea {
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

.btn.save-btn {
  background-color: #10b981;
}

@media print {
  .action-buttons,
  th:nth-child(6),
  td:nth-child(6),
  th:nth-child(10),
  td:nth-child(10),
  th:nth-child(11),
  td:nth-child(11) {
    display: none !important;
  }

  /* Adjusted print column widths */
  th:nth-child(1) { width: 15% !important; }
  th:nth-child(2) { width: 16% !important; }
  th:nth-child(3) { width: 10% !important; }
  th:nth-child(4) { width: 8% !important; }
  th:nth-child(5) { width: 10% !important; }
  th:nth-child(7) { width: 12% !important; }
  th:nth-child(8) { width: 15% !important; }
  th:nth-child(9) { width: 14% !important; }
}
</style>

