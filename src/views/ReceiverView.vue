<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'
import TransferDetails from '../components/TransferDetails.vue'

const router = useRouter()
const { callApi, error } = useApi()
const user = ref(null)
const transfers = ref([])
const showReceiveDialog = ref(false)
const selectedTransfer = ref(null)
const receiveForm = ref({
  amount_received_usd: '',
  receiver_notes: '',
})

const isLoading = ref(false)
const processingTransferId = ref(null)

const showDetailsDialog = ref(false)
const selectedTransferForDetails = ref(null)

const fetchTransfers = async () => {
  isLoading.value = true
  try {
    const result = await callApi({
      query: `
        SELECT t.*, u.username as sender_name,
          b.company_name, b.bank_name, b.bank_account, b.swift_code
        FROM transfers t
        JOIN users u ON t.id_user_do_transfer = u.id
        LEFT JOIN banks b ON t.id_bank = b.id
        WHERE t.date_receive IS NULL
        ORDER BY t.date_do_transfer DESC
      `,
      params: [],
    })
    if (result.success) {
      transfers.value = result.data
    }
  } catch (err) {
    console.error('Error fetching transfers:', err)
  } finally {
    isLoading.value = false
  }
}

const calculateUSD = (amount, rate) => {
  return (amount / rate).toFixed(2)
}

const openReceiveDialog = (transfer) => {
  selectedTransfer.value = transfer
  receiveForm.value = {
    amount_received_usd: calculateUSD(transfer.amount_sending_da, transfer.rate),
    receiver_notes: '',
  }
  showReceiveDialog.value = true
}

const receiveTransfer = async () => {
  if (processingTransferId.value) return
  if (!receiveForm.value.amount_received_usd) {
    error.value = 'Please fill all required fields'
    return
  }

  processingTransferId.value = selectedTransfer.value.id
  try {
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
        selectedTransfer.value.id,
      ],
    })

    if (result.success) {
      showReceiveDialog.value = false
      selectedTransfer.value = null
      fetchTransfers()
      fetchRecentTransfers()
    } else {
      error.value = result.message
    }
  } finally {
    processingTransferId.value = null
  }
}

const recentTransfers = ref([])

const fetchRecentTransfers = async () => {
  const result = await callApi({
    query: `
      SELECT t.*, u.username as sender_name,
        b.company_name, b.bank_name, b.bank_account, b.swift_code
      FROM transfers t
      JOIN users u ON t.id_user_do_transfer = u.id
      LEFT JOIN banks b ON t.id_bank = b.id
      WHERE t.date_receive IS NOT NULL
      AND t.date_receive >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
      ORDER BY t.date_receive DESC
    `,
    params: [],
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
  if (processingTransferId.value) return
  if (!confirm('Are you sure you want to mark this transfer as unreceived?')) return

  processingTransferId.value = transfer.id
  try {
    const result = await callApi({
      query: `
        UPDATE transfers 
        SET amount_received_usd = NULL,
            receiver_notes = NULL,
            date_receive = NULL,
            id_user_receive_transfer = NULL
        WHERE id = ?
      `,
      params: [transfer.id],
    })

    if (result.success) {
      fetchTransfers()
      fetchRecentTransfers()
    } else {
      error.value = result.message
    }
  } finally {
    processingTransferId.value = null
  }
}

const openDetailsDialog = (transfer) => {
  console.log('Opening details for transfer:', transfer)
  selectedTransferForDetails.value = transfer
  showDetailsDialog.value = true
}
</script>

<template>
  <div class="receiver-view">
    <div class="header">
      <h1><i class="fas fa-exchange-alt"></i> All Pending Transfers</h1>
      <button @click="router.push('/transfers')" class="back-btn">
        <i class="fas fa-arrow-left"></i> Return to Transfers
      </button>
    </div>

    <!-- Add loading overlay -->
    <div v-if="isLoading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Loading transfers...</span>
    </div>

    <!-- Transfers Table -->
    <div class="transfers-table">
      <table>
        <thead>
          <tr>
            <th>Sender</th>
            <th>Date Sent</th>
            <th>Sent USD</th>
            <th>Bank</th>
            <th>Account</th>
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
            <td>${{ transfer.amount_received_usd || '-' }}</td>
            <td>{{ transfer.notes || '-' }}</td>
            <td>{{ transfer.receiver_notes || '-' }}</td>
            <td>
              <button
                @click="openReceiveDialog(transfer)"
                class="btn receive-btn"
                :disabled="processingTransferId === transfer.id"
              >
                <i class="fas fa-check"></i>
                <span v-if="processingTransferId === transfer.id">
                  <i class="fas fa-spinner fa-spin"></i>
                </span>
                <span v-else>Receive</span>
              </button>
              <button
                @click="openDetailsDialog(transfer)"
                class="btn details-btn"
                title="View Details"
              >
                <i class="fas fa-list-ul"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Recent Received Transfers Table -->
    <h2 class="section-title">
      <i class="fas fa-history"></i> Recently Received Transfers (Past Week)
    </h2>
    <div class="transfers-table">
      <table>
        <thead>
          <tr>
            <th>Sender</th>
            <th>Date Received</th>
            <th>Sent USD</th>
            <th>Bank</th>
            <th>Account</th>
            <th>Received USD</th>
            <th>Notes</th>
            <th>Receiver Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="transfer in recentTransfers"
            :key="transfer.id"
            :class="{
              'amount-mismatch':
                calculateUSD(transfer.amount_sending_da, transfer.rate) !==
                transfer.amount_received_usd,
            }"
          >
            <td>{{ transfer.sender_name }}</td>
            <td>{{ new Date(transfer.date_receive).toLocaleString() }}</td>
            <td>${{ calculateUSD(transfer.amount_sending_da, transfer.rate) }}</td>
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
            <td>${{ transfer.amount_received_usd }}</td>
            <td>{{ transfer.notes || '-' }}</td>
            <td>{{ transfer.receiver_notes || '-' }}</td>
            <td>
              <button
                @click="unreceiveTransfer(transfer)"
                class="btn unreceive-btn"
                :disabled="processingTransferId === transfer.id"
              >
                <i class="fas fa-undo"></i>
                <span v-if="processingTransferId === transfer.id">
                  <i class="fas fa-spinner fa-spin"></i>
                </span>
                <span v-else>Unreceive</span>
              </button>
              <button
                @click="openDetailsDialog(transfer)"
                class="btn details-btn"
                title="View Details"
              >
                <i class="fas fa-list-ul"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Receive Dialog -->
    <div v-if="showReceiveDialog" class="dialog-overlay">
      <div class="dialog">
        <h2><i class="fas fa-check-circle"></i> Receive Transfer</h2>
        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>
        <div class="form-group">
          <label><i class="fas fa-dollar-sign"></i> Amount Received (USD):</label>
          <input
            type="number"
            v-model="receiveForm.amount_received_usd"
            class="input-field"
            step="0.01"
          />
        </div>
        <div class="form-group">
          <label><i class="fas fa-comment"></i> Receiver Notes:</label>
          <textarea
            v-model="receiveForm.receiver_notes"
            class="input-field"
            placeholder="Add your notes here"
          ></textarea>
        </div>
        <div class="dialog-actions">
          <button
            @click="receiveTransfer"
            class="btn receive-btn"
            :disabled="processingTransferId === selectedTransfer?.id"
          >
            <i class="fas fa-check"></i>
            <span v-if="processingTransferId === selectedTransfer?.id">
              <i class="fas fa-spinner fa-spin"></i> Processing...
            </span>
            <span v-else>Confirm</span>
          </button>
          <button
            @click="showReceiveDialog = false"
            class="btn cancel-btn"
            :disabled="processingTransferId === selectedTransfer?.id"
          >
            <i class="fas fa-times"></i> Cancel
          </button>
        </div>
      </div>
    </div>

    <TransferDetails
      v-if="showDetailsDialog"
      :transfer-id="selectedTransferForDetails?.id"
      :is-visible="showDetailsDialog"
      :read-only="true"
      @close="showDetailsDialog = false"
    />
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
  width: 10%;
} /* Sender */
th:nth-child(2) {
  width: 12%;
} /* Date */
th:nth-child(3) {
  width: 8%;
} /* Sent USD */
th:nth-child(4) {
  width: 15%;
} /* Bank */
th:nth-child(5) {
  width: 15%;
} /* Account */
th:nth-child(6) {
  width: 8%;
} /* Received USD */
th:nth-child(7) {
  width: 15%;
} /* Notes */
th:nth-child(8) {
  width: 15%;
} /* Receiver Notes */
th:nth-child(9) {
  width: 10%;
} /* Actions */

/* Notes cell specific styling */
td:nth-child(7),
td:nth-child(8) {
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
  background-color: #4caf50;
  color: white;
}

.unreceive-btn {
  background-color: #dc2626;
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
.transfers-table:nth-of-type(2) table th:nth-child(1) {
  width: 10%;
} /* Sender */
.transfers-table:nth-of-type(2) table th:nth-child(2) {
  width: 12%;
} /* Date */
.transfers-table:nth-of-type(2) table th:nth-child(3) {
  width: 8%;
} /* Sent USD */
.transfers-table:nth-of-type(2) table th:nth-child(4) {
  width: 15%;
} /* Bank */
.transfers-table:nth-of-type(2) table th:nth-child(5) {
  width: 15%;
} /* Account */
.transfers-table:nth-of-type(2) table th:nth-child(6) {
  width: 8%;
} /* Received USD */
.transfers-table:nth-of-type(2) table th:nth-child(7) {
  width: 15%;
} /* Notes */
.transfers-table:nth-of-type(2) table th:nth-child(8) {
  width: 15%;
} /* Receiver Notes */
.transfers-table:nth-of-type(2) table th:nth-child(9) {
  width: 10%;
} /* Actions */

/* Notes cell specific styling for recent transfers */
.transfers-table:nth-of-type(2) table td:nth-child(7),
.transfers-table:nth-of-type(2) table td:nth-child(8) {
  white-space: pre-wrap;
  min-width: 150px;
}

/* Add this new style for amount mismatch highlighting */
.amount-mismatch {
  background-color: #fef2f2; /* Light red background */
  color: #991b1b; /* Darker red text for contrast */
}

.amount-mismatch td {
  border-bottom-color: #fca5a5; /* Reddish border */
}

.bank-cell {
  font-size: 0.95em;
  line-height: 1.5;
  white-space: normal;
}

.bank-cell strong {
  color: #1f2937;
  font-weight: 600;
  display: block;
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
  display: inline-block;
  margin-top: 2px;
}

.loading-overlay {
  position: fixed;
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
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn i {
  font-size: 0.9em;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background-color: #fef2f2;
  border: 1px solid #fee2e2;
  border-radius: 4px;
  color: #dc2626;
  margin-bottom: 15px;
}

h1 i,
h2 i {
  margin-right: 8px;
}

.form-group label i {
  margin-right: 4px;
  color: #6b7280;
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

  th:nth-child(9),
  td:nth-child(9) {
    display: none !important;
  }

  .amount-mismatch {
    background-color: transparent !important;
    border-left: 4px solid #dc2626 !important;
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
