<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import TransferForm from './TransferForm.vue'

const { callApi } = useApi()
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
    const result = await callApi({
      query: `
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
      `,
      params: [currentUser.value.id, currentUser.value.id]
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
      params: []
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
      params: [transfer.id]
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
      params: [transfer.id, currentUser.value.id]
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
      params: [currentUser.value.id, transfer.id]
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
  return isAdmin.value || (!transfer.date_received)
}

const canDeleteTransfer = (transfer) => {
  if (!transfer) return false
  return isAdmin.value || (!transfer.date_received)
}

// Format date to local string
const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    currentUser.value = JSON.parse(userStr)
    fetchTransfers()
  } else {
    error.value = 'User not logged in'
  }
})
</script>

<template>
  <div class="transfers-table">
    <div class="toolbar">
      <div class="toolbar-left">
        <h2>Internal Transfers</h2>
        <div class="stats">
          <span class="stat-item">
            <span class="stat-label">Total:</span>
            <span class="stat-value">{{ transfers.length }}</span>
          </span>
          <span class="stat-item">
            <span class="stat-label">Pending:</span>
            <span class="stat-value">{{ transfers.filter(t => !t.date_received).length }}</span>
          </span>
          <span class="stat-item" :class="{ 'warning': transfers.filter(t => !t.id_admin_confirm).length > 0 }">
            <span class="stat-label">Needs Admin:</span>
            <span class="stat-value">{{ transfers.filter(t => !t.id_admin_confirm).length }}</span>
          </span>
          <span class="stat-item success">
            <span class="stat-label">Admin Confirmed:</span>
            <span class="stat-value">{{ transfers.filter(t => t.id_admin_confirm).length }}</span>
          </span>
        </div>
      </div>
      <div class="toolbar-right">
        <button @click="handleAddNew" class="toolbar-btn primary">
          <span class="btn-icon">+</span>
          New Transfer
        </button>
        <button 
          v-if="isAdmin"
          @click="handleAdminConfirm(selectedTransfer)" 
          class="toolbar-btn admin"
          :disabled="!selectedTransfer || !canAdminConfirm(selectedTransfer)"
          :title="!selectedTransfer ? 'Select a transfer to confirm' : 
                  selectedTransfer.id_admin_confirm ? 'Transfer already confirmed by admin' :
                  !isAdmin ? 'Only admins can confirm transfers' : 
                  'Confirm this transfer as admin'"
        >
          <span class="btn-icon">👤</span>
          Admin Confirm
        </button>
        <button 
          @click="handleConfirmReceive(selectedTransfer)" 
          class="toolbar-btn success"
          :disabled="!selectedTransfer || !canConfirmReceive(selectedTransfer)"
          :title="!selectedTransfer ? 'Select a transfer to confirm' : 
                  selectedTransfer.date_received ? 'Transfer already received' :
                  selectedTransfer.to_user_id !== currentUser?.id ? 'You are not the receiver of this transfer' : 
                  'Confirm receiving this transfer'"
        >
          <span class="btn-icon">✓</span>
          Confirm Receive
        </button>
        <button 
          @click="handleEdit(selectedTransfer)" 
          class="toolbar-btn secondary"
          :disabled="!selectedTransfer || !canEditTransfer(selectedTransfer)"
          :title="!selectedTransfer ? 'Select a transfer to edit' : 
                  !canEditTransfer(selectedTransfer) ? 'Cannot edit received transfers' : ''"
        >
          <span class="btn-icon">✎</span>
          Edit
        </button>
        <button 
          @click="handleDelete(selectedTransfer)" 
          class="toolbar-btn danger"
          :disabled="!selectedTransfer || !canDeleteTransfer(selectedTransfer)"
          :title="!selectedTransfer ? 'Select a transfer to delete' : 
                  !canDeleteTransfer(selectedTransfer) ? 'Cannot delete received transfers' : ''"
        >
          <span class="btn-icon">🗑</span>
          Delete
        </button>
        <button 
          @click="fetchTransfers" 
          class="toolbar-btn secondary"
          :disabled="loading"
        >
          <span class="btn-icon">↻</span>
          Refresh
        </button>
      </div>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>
    
    <div class="table-container">
      <table v-if="!loading && transfers.length > 0">
        <thead>
          <tr>
            <th>ID</th>
            <th>Date</th>
            <th>From</th>
            <th>To</th>
            <th>Amount</th>
            <th>Status</th>
            <th>Received On</th>
            <th>Admin Confirmed</th>
            <th>Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="transfer in transfers" 
            :key="transfer.id"
            :class="{ 'selected': selectedTransfer?.id === transfer.id }"
            @click="selectedTransfer = transfer"
          >
            <td>#{{ transfer.id }}</td>
            <td>{{ formatDate(transfer.date_transfer) }}</td>
            <td>{{ transfer.from_username || 'N/A' }}</td>
            <td>{{ transfer.to_username || 'N/A' }}</td>
            <td>${{ transfer.amount }}</td>
            <td>
              <span :class="['status', transfer.date_received ? 'received' : 'pending']">
                {{ transfer.date_received ? 'Received' : 'Pending' }}
              </span>
            </td>
            <td>{{ transfer.date_received ? formatDate(transfer.date_received) : '-' }}</td>
            <td>
              <span :class="['admin-status', transfer.id_admin_confirm ? 'confirmed' : 'pending']">
                {{ transfer.admin_username || 'Not confirmed' }}
              </span>
            </td>
            <td>{{ transfer.notes || '-' }}</td>
          </tr>
        </tbody>
      </table>
      
      <div v-else-if="loading" class="loading">
        Loading transfers...
      </div>
      
      <div v-else class="no-data">
        No transfers found
      </div>
    </div>

    <!-- Transfer Form Dialog -->
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
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5e7eb;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 24px;
}

.toolbar-left h2 {
  margin: 0;
  color: #1a1a1a;
  font-size: 1.5rem;
}

.stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background-color: #f3f4f6;
  border-radius: 6px;
  font-size: 0.875rem;
}

.stat-label {
  color: #6b7280;
}

.stat-value {
  font-weight: 600;
  color: #1f2937;
}

.toolbar-right {
  display: flex;
  gap: 12px;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.toolbar-btn.primary {
  background-color: #10b981;
  color: white;
}

.toolbar-btn.primary:hover {
  background-color: #059669;
}

.toolbar-btn.secondary {
  background-color: #f3f4f6;
  color: #374151;
}

.toolbar-btn.secondary:hover {
  background-color: #e5e7eb;
}

.toolbar-btn.danger {
  background-color: #ef4444;
  color: white;
}

.toolbar-btn.danger:hover {
  background-color: #dc2626;
}

.toolbar-btn.success {
  background-color: #059669;
  color: white;
}

.toolbar-btn.success:hover {
  background-color: #047857;
}

.toolbar-btn.success:disabled {
  background-color: #059669;
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  font-size: 1.1em;
  line-height: 1;
}

.table-container {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

th {
  background-color: #f9fafb;
  padding: 12px;
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
}

td {
  padding: 12px;
  border-bottom: 1px solid #e5e7eb;
  color: #1f2937;
}

tr {
  cursor: pointer;
  transition: background-color 0.2s;
}

tr:hover {
  background-color: #f3f4f6;
}

tr.selected {
  background-color: #e5e7eb;
}

.status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
}

.status.received {
  background-color: #d1fae5;
  color: #065f46;
}

.status.pending {
  background-color: #fef3c7;
  color: #92400e;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
}

.loading, .no-data {
  text-align: center;
  padding: 40px;
  color: #6b7280;
  font-size: 1.1rem;
}

.stat-item.warning {
  background-color: #fef3c7;
  color: #92400e;
}

.stat-item.warning .stat-label {
  color: #92400e;
}

.stat-item.warning .stat-value {
  color: #92400e;
  font-weight: 700;
}

.stat-item.success {
  background-color: #d1fae5;
  color: #065f46;
}

.stat-item.success .stat-label {
  color: #065f46;
}

.stat-item.success .stat-value {
  color: #065f46;
  font-weight: 700;
}
</style> 