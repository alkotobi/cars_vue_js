<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useInvoiceCompanyInfo } from '../../composables/useInvoiceCompanyInfo'
import TransferForm from './TransferForm.vue'

const { callApi, getAssets } = useApi()
const { getCompanyLogoUrl } = useInvoiceCompanyInfo()
const letterHeadUrl = ref(null)
const transfers = ref([])
const users = ref([])
const loading = ref(false)
const error = ref(null)
const showTransferForm = ref(false)
const selectedTransfer = ref(null)
const currentUser = ref(null)

// Fetch all transfers with user details
const fetchTransfers = async () => {
  loading.value = true
  error.value = null

  try {
    // Build query based on user role
    const isAdmin = currentUser.value?.role_id === 1
    let query, params

    if (isAdmin) {
      // Admin can see all transfers
      query = `
        SELECT 
          t.*,
          from_user.username as from_username,
          to_user.username as to_username,
          admin.username as admin_username
        FROM transfers_inter t
        LEFT JOIN users from_user ON t.from_user_id = from_user.id
        LEFT JOIN users to_user ON t.to_user_id = to_user.id
        LEFT JOIN users admin ON t.id_admin_confirm = admin.id
        ORDER BY t.date_transfer DESC
      `
      params = []
    } else {
      // Regular users only see their own transfers
      query = `
        SELECT 
          t.*,
          from_user.username as from_username,
          to_user.username as to_username,
          admin.username as admin_username
        FROM transfers_inter t
        LEFT JOIN users from_user ON t.from_user_id = from_user.id
        LEFT JOIN users to_user ON t.to_user_id = to_user.id
        LEFT JOIN users admin ON t.id_admin_confirm = admin.id
        WHERE t.from_user_id = ? OR t.to_user_id = ?
        ORDER BY t.date_transfer DESC
      `
      params = [currentUser.value.id, currentUser.value.id]
    }

    const result = await callApi({
      query,
      params,
    })

    if (result.success) {
      transfers.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch transfers'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Fetch users for dropdowns
const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, username FROM users ORDER BY username',
      params: [],
    })

    if (result.success) {
      users.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch users'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const handleEdit = (transfer) => {
  selectedTransfer.value = transfer
  showTransferForm.value = true
}

const handleDelete = async (transfer) => {
  if (!confirm(`Are you sure you want to delete transfer #${transfer.id}?`)) {
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: 'DELETE FROM transfers_inter WHERE id = ?',
      params: [transfer.id],
    })

    if (result.success) {
      await fetchTransfers()
    } else {
      error.value = result.error || 'Failed to delete transfer'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const handleTransferSaved = () => {
  showTransferForm.value = false
  selectedTransfer.value = null
  fetchTransfers() // Refresh the list
}

const handleAddNew = () => {
  selectedTransfer.value = null
  showTransferForm.value = true
}

const handleConfirmReceive = async (transfer) => {
  if (!confirm('Are you sure you want to confirm receiving this transfer?')) {
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        UPDATE transfers_inter 
        SET date_received = CURRENT_TIMESTAMP
        WHERE id = ? AND to_user_id = ? AND date_received IS NULL
      `,
      params: [transfer.id, currentUser.value.id],
    })

    if (result.success) {
      await fetchTransfers()
    } else {
      error.value = result.error || 'Failed to confirm transfer'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const canConfirmReceive = (transfer) => {
  return transfer.to_user_id === currentUser.value?.id && !transfer.date_received
}

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

const handleAdminConfirm = async (transfer) => {
  if (!confirm('Are you sure you want to confirm this transfer as an admin?')) {
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        UPDATE transfers_inter 
        SET id_admin_confirm = ?
        WHERE id = ? AND id_admin_confirm IS NULL
      `,
      params: [currentUser.value.id, transfer.id],
    })

    if (result.success) {
      await fetchTransfers()
    } else {
      error.value = result.error || 'Failed to confirm transfer'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const canAdminConfirm = (transfer) => {
  return isAdmin.value && transfer && !transfer.id_admin_confirm
}

// Add new helper functions after canAdminConfirm
const canEditTransfer = (transfer) => {
  if (!transfer) return false
  return isAdmin.value || !transfer.date_received
}

const canDeleteTransfer = (transfer) => {
  if (!transfer) return false
  return isAdmin.value || !transfer.date_received
}

// Format date to local string
const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

// Format date with time
const formatDateTime = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleString()
}

const printTransferReceipt = async (transfer) => {
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    alert('Popup blocked. Please allow popups for this site.')
    return
  }

  // Load company logo for header (same as xlsx invoice)
  if (!letterHeadUrl.value) {
    try {
      letterHeadUrl.value = (await getCompanyLogoUrl()) || ''
    } catch (err) {
      console.error('Failed to load company logo:', err)
    }
  }

  const receiptHtml = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Transfer Receipt #${transfer.id}</title>
      <style>
        @page {
          size: A4;
          margin: 10mm;
        }
        
        body {
          font-family: Arial, sans-serif;
          font-size: 11px;
          line-height: 1.4;
          color: #333;
          margin: 0;
          padding: 0;
        }
        
        .letterhead {
          margin-bottom: 15px;
        }
        
        .letterhead img {
          width: 100%;
          height: auto;
          max-height: 40px;
          object-fit: contain;
        }
        
        .receipt-title {
          text-align: center;
          font-size: 18px;
          font-weight: bold;
          margin-bottom: 15px;
          color: #1e293b;
        }
        
        .receipt-info {
          background-color: #f8f9fa;
          padding: 12px;
          border-radius: 6px;
          margin-bottom: 15px;
        }
        
        .info-row {
          display: flex;
          justify-content: space-between;
          padding: 6px 0;
          border-bottom: 1px solid #e5e7eb;
        }
        
        .info-row:last-child {
          border-bottom: none;
        }
        
        .info-label {
          font-weight: bold;
          color: #64748b;
        }
        
        .info-value {
          color: #1e293b;
          font-weight: 500;
        }
        
        .amount-section {
          background: linear-gradient(135deg, #8b5cf6, #7c3aed);
          color: white;
          padding: 20px;
          border-radius: 8px;
          text-align: center;
          margin: 20px 0;
        }
        
        .amount-label {
          font-size: 12px;
          opacity: 0.9;
          margin-bottom: 5px;
        }
        
        .amount-value {
          font-size: 28px;
          font-weight: bold;
        }
        
        .status-section {
          padding: 12px;
          margin-bottom: 15px;
          border-radius: 6px;
        }
        
        .status-received {
          background-color: #ecfdf5;
          color: #059669;
          border: 1px solid #bbf7d0;
        }
        
        .status-pending {
          background-color: #fef3c7;
          color: #d97706;
          border: 1px solid #fde68a;
        }
        
        .status-admin {
          background-color: #f5f3ff;
          color: #7c3aed;
          border: 1px solid #e9d5ff;
        }
        
        .status-label {
          font-weight: bold;
          margin-bottom: 5px;
        }
        
        .footer {
          margin-top: 20px;
          text-align: center;
          font-size: 9px;
          color: #64748b;
          border-top: 1px solid #e5e7eb;
          padding-top: 10px;
        }
        
        @media print {
          body {
            -webkit-print-color-adjust: exact;
            print-color-adjust: exact;
          }
        }
      </style>
    </head>
    <body>
      <div class="letterhead">
        <img src="${letterHeadUrl.value}" alt="Company Letterhead">
      </div>
      
      <div class="receipt-title">INTERNAL TRANSFER RECEIPT</div>
      
      <div class="receipt-info">
        <div class="info-row">
          <span class="info-label">Transfer ID:</span>
          <span class="info-value">Transfer #${transfer.id}</span>
        </div>
        <div class="info-row">
          <span class="info-label">Date:</span>
          <span class="info-value">${formatDateTime(transfer.date_transfer)}</span>
        </div>
        <div class="info-row">
          <span class="info-label">From User:</span>
          <span class="info-value">${transfer.from_username}</span>
        </div>
        <div class="info-row">
          <span class="info-label">To User:</span>
          <span class="info-value">${transfer.to_username}</span>
        </div>
        ${
          transfer.notes
            ? `
        <div class="info-row">
          <span class="info-label">Notes:</span>
          <span class="info-value">${transfer.notes}</span>
        </div>
        `
            : ''
        }
      </div>
      
      <div class="amount-section">
        <div class="amount-label">Transfer Amount</div>
        <div class="amount-value">$${parseFloat(transfer.amount).toFixed(2)}</div>
      </div>
      
      ${
        transfer.date_received
          ? `
      <div class="status-section status-received">
        <div class="status-label">✓ Received Confirmation</div>
        <div>Confirmed on: ${formatDateTime(transfer.date_received)}</div>
      </div>
      `
          : `
      <div class="status-section status-pending">
        <div class="status-label">⏱ Pending Receipt</div>
        <div>Waiting for receiver confirmation</div>
      </div>
      `
      }
      
      ${
        transfer.id_admin_confirm
          ? `
      <div class="status-section status-admin">
        <div class="status-label">🛡 Admin Confirmed</div>
        <div>Confirmed by: ${transfer.admin_username || 'Admin'}</div>
      </div>
      `
          : `
      <div class="status-section status-pending">
        <div class="status-label">⏱ Pending Admin Approval</div>
        <div>Waiting for admin confirmation</div>
      </div>
      `
      }
      
      <div class="footer">
        <p>This is a computer-generated receipt for internal transfer.</p>
        <p>Generated on ${new Date().toLocaleString()}</p>
      </div>
    </body>
    </html>
  `

  printWindow.document.write(receiptHtml)
  printWindow.document.close()

  // Wait for images to load, then print
  printWindow.onload = () => {
    setTimeout(() => {
      printWindow.print()
      printWindow.close()
    }, 500)
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    currentUser.value = JSON.parse(userStr)
    fetchTransfers()
    fetchUsers()
  } else {
    error.value = 'User not logged in'
  }
})
</script>

<template>
  <div class="transfers-table">
    <div class="toolbar">
      <div class="toolbar-left">
        <h2>
          <i class="fas fa-exchange-alt"></i>
          Internal Transfers
        </h2>
        <div class="stats">
          <span class="stat-item">
            <i class="fas fa-list"></i>
            <span class="stat-label">Total:</span>
            <span class="stat-value">{{ transfers.length }}</span>
          </span>
          <span class="stat-item pending">
            <i class="fas fa-clock"></i>
            <span class="stat-label">Pending:</span>
            <span class="stat-value">{{ transfers.filter((t) => !t.date_received).length }}</span>
          </span>
          <span
            class="stat-item warning"
            :class="{ warning: transfers.filter((t) => !t.id_admin_confirm).length > 0 }"
          >
            <i class="fas fa-user-shield"></i>
            <span class="stat-label">Needs Admin:</span>
            <span class="stat-value">{{
              transfers.filter((t) => !t.id_admin_confirm).length
            }}</span>
          </span>
          <span class="stat-item success">
            <i class="fas fa-check-circle"></i>
            <span class="stat-label">Admin Confirmed:</span>
            <span class="stat-value">{{ transfers.filter((t) => t.id_admin_confirm).length }}</span>
          </span>
        </div>
      </div>
      <div class="toolbar-right">
        <button @click="handleAddNew" class="toolbar-btn primary" :disabled="loading">
          <i class="fas fa-plus"></i>
          New Transfer
        </button>
        <button
          v-if="isAdmin"
          @click="handleAdminConfirm(selectedTransfer)"
          class="toolbar-btn admin"
          :disabled="loading || !selectedTransfer || !canAdminConfirm(selectedTransfer)"
          :title="
            !selectedTransfer
              ? 'Select a transfer to confirm'
              : selectedTransfer.id_admin_confirm
                ? 'Transfer already confirmed by admin'
                : !isAdmin
                  ? 'Only admins can confirm transfers'
                  : 'Confirm this transfer as admin'
          "
        >
          <i class="fas fa-user-shield"></i>
          Admin Confirm
        </button>
      </div>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div class="table-container" :class="{ 'is-loading': loading }">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading transfers...</span>
      </div>

      <table v-if="transfers.length > 0">
        <thead>
          <tr>
            <th><i class="fas fa-hashtag"></i> ID</th>
            <th><i class="fas fa-calendar-alt"></i> Date</th>
            <th><i class="fas fa-dollar-sign"></i> Amount</th>
            <th><i class="fas fa-user-circle"></i> From</th>
            <th><i class="fas fa-user-circle"></i> To</th>
            <th><i class="fas fa-sticky-note"></i> Notes</th>
            <th><i class="fas fa-check-circle"></i> Received</th>
            <th><i class="fas fa-user-shield"></i> Admin</th>
            <th><i class="fas fa-cog"></i> Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="transfer in transfers"
            :key="transfer.id"
            @click="selectedTransfer = transfer"
            :class="{ selected: selectedTransfer?.id === transfer.id }"
            style="cursor: pointer"
          >
            <td>#{{ transfer.id }}</td>
            <td>{{ formatDate(transfer.date_transfer) }}</td>
            <td class="amount">${{ transfer.amount }}</td>
            <td>{{ transfer.from_username }}</td>
            <td>{{ transfer.to_username }}</td>
            <td class="notes">{{ transfer.notes || '-' }}</td>
            <td :class="{ confirmed: transfer.date_received }">
              <span v-if="transfer.date_received" class="status-badge received">
                <i class="fas fa-check"></i>
                {{ formatDate(transfer.date_received) }}
              </span>
              <button
                v-else-if="canConfirmReceive(transfer)"
                @click="handleConfirmReceive(transfer)"
                class="action-btn confirm-btn"
                :disabled="loading"
              >
                <i class="fas fa-check"></i>
                Confirm Receipt
              </button>
              <span v-else class="status-badge pending">
                <i class="fas fa-clock"></i>
                Pending
              </span>
            </td>
            <td :class="{ confirmed: transfer.id_admin_confirm }">
              <span v-if="transfer.id_admin_confirm" class="status-badge admin">
                <i class="fas fa-user-shield"></i>
                {{ transfer.admin_username }}
              </span>
              <button
                v-else-if="isAdmin"
                @click.stop="handleAdminConfirm(transfer)"
                class="action-btn admin-btn"
                :disabled="loading"
                title="Confirm this transfer as admin"
              >
                <i class="fas fa-user-shield"></i>
                Confirm
              </button>
              <span v-else class="status-badge pending">
                <i class="fas fa-clock"></i>
                Pending
              </span>
            </td>
            <td class="actions">
              <button
                @click="printTransferReceipt(transfer)"
                class="action-btn print-btn"
                :disabled="loading"
                title="Print receipt"
              >
                <i class="fas fa-print"></i>
              </button>
              <button
                v-if="canEditTransfer(transfer)"
                @click="handleEdit(transfer)"
                class="action-btn edit-btn"
                :disabled="loading"
                title="Edit transfer"
              >
                <i class="fas fa-edit"></i>
              </button>
              <button
                v-if="canDeleteTransfer(transfer)"
                @click="handleDelete(transfer)"
                class="action-btn delete-btn"
                :disabled="loading"
                title="Delete transfer"
              >
                <i class="fas fa-trash-alt"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="!loading" class="empty-state">
        <i class="fas fa-inbox fa-3x"></i>
        <p>No transfers found</p>
        <button @click="handleAddNew" class="toolbar-btn primary" :disabled="loading">
          <i class="fas fa-plus"></i>
          Create New Transfer
        </button>
      </div>
    </div>

    <TransferForm
      :visible="showTransferForm"
      :edit-data="selectedTransfer"
      @close="showTransferForm = false"
      @save="handleTransferSaved"
    />
  </div>
</template>

<style scoped>
.transfers-table {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.toolbar {
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  border-bottom: 1px solid #e2e8f0;
  background-color: #f8fafc;
}

.toolbar-left h2 {
  margin: 0 0 16px 0;
  color: #1e293b;
  font-size: 1.5rem;
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-left h2 i {
  color: #3b82f6;
}

.stats {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  font-size: 0.875rem;
}

.stat-item i {
  color: #64748b;
}

.stat-label {
  color: #64748b;
  font-weight: 500;
}

.stat-value {
  color: #1e293b;
  font-weight: 600;
}

.stat-item.warning {
  border-color: #fbbf24;
  background-color: #fef3c7;
}

.stat-item.warning i {
  color: #d97706;
}

.stat-item.success {
  border-color: #34d399;
  background-color: #ecfdf5;
}

.stat-item.success i {
  color: #059669;
}

.stat-item.pending {
  border-color: #60a5fa;
  background-color: #eff6ff;
}

.stat-item.pending i {
  color: #2563eb;
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.toolbar-btn {
  padding: 10px 16px;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.875rem;
  transition: all 0.2s ease;
  border: none;
}

.toolbar-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.toolbar-btn.primary {
  background-color: #3b82f6;
  color: white;
}

.toolbar-btn.primary:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.toolbar-btn.admin {
  background-color: #8b5cf6;
  color: white;
}

.toolbar-btn.admin:hover:not(:disabled) {
  background-color: #7c3aed;
  transform: translateY(-1px);
}

.error-message {
  margin: 16px 20px;
  padding: 12px;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 8px;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 8px;
  animation: slideIn 0.3s ease;
}

.table-container {
  position: relative;
  overflow-x: auto;
}

.table-container.is-loading {
  min-height: 200px;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 12px;
  z-index: 10;
}

.loading-overlay i {
  color: #3b82f6;
}

.loading-overlay span {
  color: #4b5563;
  font-size: 0.875rem;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  background-color: #f8fafc;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #1e293b;
  font-size: 0.875rem;
  border-bottom: 2px solid #e2e8f0;
  white-space: nowrap;
}

th i {
  color: #64748b;
  margin-right: 8px;
  width: 16px;
}

td {
  padding: 12px 16px;
  border-bottom: 1px solid #e2e8f0;
  color: #1e293b;
  font-size: 0.875rem;
}

tr:hover {
  background-color: #f8fafc;
}

tr.selected {
  background-color: #eff6ff;
}

.amount {
  font-weight: 600;
  font-family: monospace;
}

.notes {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge.received {
  background-color: #ecfdf5;
  color: #059669;
}

.status-badge.admin {
  background-color: #f5f3ff;
  color: #7c3aed;
}

.status-badge.pending {
  background-color: #f3f4f6;
  color: #6b7280;
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.action-btn {
  padding: 6px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  color: white;
  width: 32px;
  height: 32px;
}

.action-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.action-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.edit-btn {
  background-color: #3b82f6;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.delete-btn {
  background-color: #ef4444;
}

.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.confirm-btn {
  background-color: #10b981;
  padding: 6px 12px;
  width: auto;
  font-size: 0.75rem;
  font-weight: 500;
}

.confirm-btn:hover:not(:disabled) {
  background-color: #059669;
}

.admin-btn {
  background-color: #8b5cf6;
  padding: 6px 12px;
  width: auto;
  font-size: 0.75rem;
  font-weight: 500;
}

.admin-btn:hover:not(:disabled) {
  background-color: #7c3aed;
}

.print-btn {
  background-color: #64748b;
}

.print-btn:hover:not(:disabled) {
  background-color: #475569;
}

.empty-state {
  padding: 48px;
  text-align: center;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.empty-state i {
  color: #94a3b8;
}

.empty-state p {
  margin: 0;
  font-size: 0.875rem;
}

@keyframes slideIn {
  from {
    transform: translateY(-10px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@media (max-width: 1024px) {
  .toolbar {
    flex-direction: column;
  }

  .toolbar-right {
    width: 100%;
  }

  .toolbar-btn {
    flex: 1;
    justify-content: center;
  }

  .stats {
    flex-direction: column;
    width: 100%;
  }

  .stat-item {
    width: 100%;
    justify-content: space-between;
  }
}

@media (max-width: 768px) {
  .table-container {
    margin: 0 -20px;
  }

  td,
  th {
    padding: 12px;
  }

  .notes {
    max-width: 150px;
  }
}
</style>
