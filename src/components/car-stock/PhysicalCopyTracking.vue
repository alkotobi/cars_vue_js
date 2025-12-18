<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()
const {
  getMyPhysicalCopies,
  checkinPhysicalCopy,
  transferPhysicalCopy,
  getUsersForTransfer,
  getFileTransferHistory,
  getFileUrl,
} = useApi()

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close'])

// State
const loading = ref(false)
const error = ref(null)
const success = ref(null)
const myPhysicalCopies = ref([])
const allPhysicalCopies = ref([]) // For admin view
const transferHistory = ref([])

// Check if current user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Filter options
const filterStatus = ref('all') // all, available, checked_out, my_copies
const searchQuery = ref('')

// Filtered copies
const filteredCopies = computed(() => {
  let copies = isAdmin.value ? allPhysicalCopies.value : myPhysicalCopies.value

  // Filter by status
  if (filterStatus.value === 'available') {
    copies = copies.filter((copy) => copy.physical_status === 'available')
  } else if (filterStatus.value === 'checked_out') {
    copies = copies.filter((copy) => copy.physical_status === 'checked_out')
  } else if (filterStatus.value === 'my_copies') {
    copies = copies.filter((copy) => copy.current_holder_id === currentUser.value?.id)
  }

  // Filter by search query
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    copies = copies.filter(
      (copy) =>
        copy.file_name?.toLowerCase().includes(query) ||
        copy.category_name?.toLowerCase().includes(query) ||
        copy.vin?.toLowerCase().includes(query) ||
        copy.current_holder_username?.toLowerCase().includes(query),
    )
  }

  return copies
})

// Load my physical copies
const loadMyPhysicalCopies = async () => {
  loading.value = true
  error.value = null

  try {
    const copies = await getMyPhysicalCopies()
    myPhysicalCopies.value = copies || []
  } catch (err) {
    error.value = err.message || 'Failed to load physical copies'
    console.error('Load error:', err)
  } finally {
    loading.value = false
  }
}

// Load all physical copies (admin only)
const loadAllPhysicalCopies = async () => {
  if (!isAdmin.value) return

  loading.value = true
  error.value = null

  try {
    const { callApi } = useApi()
    const result = await callApi({
      query: `
        SELECT 
          cf.id,
          cf.car_id,
          cf.file_name,
          cf.file_path,
          cf.uploaded_at,
          cfc.category_name,
          cs.vin,
          cpt.id as tracking_id,
          cpt.current_holder_id,
          u_holder.username as current_holder_username,
          cpt.status as physical_status,
          cpt.checked_out_at,
          cpt.expected_return_date,
          cpt.transfer_notes
        FROM car_file_physical_tracking cpt
        INNER JOIN car_files cf ON cpt.car_file_id = cf.id
        INNER JOIN car_file_categories cfc ON cf.category_id = cfc.id
        INNER JOIN cars_stock cs ON cf.car_id = cs.id
        LEFT JOIN users u_holder ON cpt.current_holder_id = u_holder.id
        WHERE cpt.status IN ('available', 'checked_out')
        ORDER BY cpt.checked_out_at DESC, cpt.status ASC
      `,
      requiresAuth: true,
    })

    if (result.success) {
      allPhysicalCopies.value = result.data || []
    }
  } catch (err) {
    error.value = err.message || 'Failed to load all physical copies'
    console.error('Load error:', err)
  } finally {
    loading.value = false
  }
}

// Load data
const loadData = async () => {
  await loadMyPhysicalCopies()
  if (isAdmin.value) {
    await loadAllPhysicalCopies()
  }
}

// Check in modal
const showCheckinModal = ref(false)
const fileToCheckin = ref(null)
const checkinNotes = ref('')

const openCheckinModal = (file) => {
  fileToCheckin.value = file
  checkinNotes.value = ''
  showCheckinModal.value = true
}

const confirmCheckin = async () => {
  if (!fileToCheckin.value) return

  loading.value = true
  error.value = null

  try {
    await checkinPhysicalCopy(fileToCheckin.value.id, checkinNotes.value || null)
    success.value = 'File checked in successfully'
    showCheckinModal.value = false
    await loadData()
  } catch (err) {
    error.value = err.message || 'Failed to check in file'
  } finally {
    loading.value = false
  }
}

// Transfer modal
const showTransferModal = ref(false)
const fileToTransfer = ref(null)
const transferToUserId = ref(null)
const transferNotes = ref('')
const transferExpectedReturn = ref('')
const availableUsers = ref([])

const openTransferModal = async (file) => {
  fileToTransfer.value = file
  transferToUserId.value = null
  transferNotes.value = ''
  transferExpectedReturn.value = ''

  try {
    const users = await getUsersForTransfer(file.id)
    availableUsers.value = users
    showTransferModal.value = true
  } catch (err) {
    error.value = err.message || 'Failed to load users'
  }
}

const confirmTransfer = async () => {
  if (!fileToTransfer.value || !transferToUserId.value) {
    error.value = 'Please select a user to transfer to'
    return
  }

  loading.value = true
  error.value = null

  try {
    await transferPhysicalCopy(
      fileToTransfer.value.id,
      transferToUserId.value,
      transferNotes.value || null,
      transferExpectedReturn.value || null,
    )
    success.value = 'File transferred successfully'
    showTransferModal.value = false
    await loadData()
  } catch (err) {
    error.value = err.message || 'Failed to transfer file'
  } finally {
    loading.value = false
  }
}

// Transfer history modal
const showHistoryModal = ref(false)
const fileForHistory = ref(null)

const openHistoryModal = async (file) => {
  fileForHistory.value = file
  loading.value = true
  error.value = null

  try {
    const history = await getFileTransferHistory(file.id)
    transferHistory.value = history || []
    showHistoryModal.value = true
  } catch (err) {
    error.value = err.message || 'Failed to load transfer history'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = null
  filterStatus.value = 'all'
  searchQuery.value = ''
  showCheckinModal.value = false
  showTransferModal.value = false
  showHistoryModal.value = false
  emit('close')
}

// Format date
const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString()
}

// Check if overdue
const isOverdue = (expectedReturnDate) => {
  if (!expectedReturnDate) return false
  return new Date(expectedReturnDate) < new Date()
}

// Watch for show prop
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      loadData()
    } else {
      closeModal()
    }
  },
  { immediate: true },
)

onMounted(() => {
  if (props.show) {
    loadData()
  }
})
</script>

<template>
  <div v-if="show" class="modal-overlay" @click="closeModal">
    <div class="modal-content tracking-modal" @click.stop>
      <div class="modal-header">
        <h3>
          <i class="fas fa-clipboard-list"></i>
          Physical Copy Tracking
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="loading">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <!-- Error Message -->
        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <!-- Success Message -->
        <div v-if="success" class="success-message">
          <i class="fas fa-check-circle"></i>
          {{ success }}
        </div>

        <!-- Filters -->
        <div class="filters-section">
          <div class="filter-group">
            <label>Filter:</label>
            <select v-model="filterStatus" class="filter-select">
              <option value="all">All</option>
              <option value="my_copies">My Copies</option>
              <option value="checked_out">Checked Out</option>
              <option value="available">Available</option>
            </select>
          </div>
          <div class="search-group">
            <input
              v-model="searchQuery"
              type="text"
              class="search-input"
              placeholder="Search by file name, category, VIN, or holder..."
            />
            <i class="fas fa-search search-icon"></i>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="loading && filteredCopies.length === 0" class="loading-state">
          <i class="fas fa-spinner fa-spin"></i>
          <span>Loading...</span>
        </div>

        <!-- Empty State -->
        <div v-else-if="filteredCopies.length === 0" class="empty-state">
          <i class="fas fa-inbox"></i>
          <p>No physical copies found</p>
        </div>

        <!-- Physical Copies Table -->
        <div v-else class="copies-table">
          <table>
            <thead>
              <tr>
                <th>File</th>
                <th>Category</th>
                <th>Car VIN</th>
                <th>Status</th>
                <th>Current Holder</th>
                <th>Checked Out</th>
                <th>Expected Return</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="copy in filteredCopies"
                :key="copy.id"
                :class="{
                  'overdue-row': isOverdue(copy.expected_return_date),
                }"
              >
                <td>
                  <a :href="getFileUrl(copy.file_path)" target="_blank" class="file-link">
                    <i class="fas fa-file-alt"></i>
                    {{ copy.file_name }}
                  </a>
                </td>
                <td>
                  <span class="category-badge">{{ copy.category_name }}</span>
                </td>
                <td>{{ copy.vin || '-' }}</td>
                <td>
                  <span v-if="copy.physical_status === 'available'" class="status-badge available">
                    <i class="fas fa-check-circle"></i> Available
                  </span>
                  <span
                    v-else-if="copy.physical_status === 'checked_out'"
                    class="status-badge checked-out"
                  >
                    <i class="fas fa-user"></i> Checked Out
                  </span>
                </td>
                <td>
                  <span v-if="copy.current_holder_username">
                    <i class="fas fa-user"></i>
                    {{ copy.current_holder_username }}
                  </span>
                  <span v-else class="text-muted">-</span>
                </td>
                <td>{{ formatDate(copy.checked_out_at) }}</td>
                <td>
                  <span
                    v-if="copy.expected_return_date"
                    :class="{ 'overdue-text': isOverdue(copy.expected_return_date) }"
                  >
                    {{ formatDate(copy.expected_return_date) }}
                    <i
                      v-if="isOverdue(copy.expected_return_date)"
                      class="fas fa-exclamation-triangle overdue-icon"
                      title="Overdue"
                    ></i>
                  </span>
                  <span v-else class="text-muted">-</span>
                </td>
                <td>
                  <div class="action-buttons">
                    <button
                      v-if="
                        copy.physical_status === 'checked_out' &&
                        copy.current_holder_id === currentUser?.id
                      "
                      @click="openCheckinModal(copy)"
                      class="btn-action btn-checkin"
                      title="Check in"
                    >
                      <i class="fas fa-sign-in-alt"></i>
                    </button>
                    <button
                      v-if="
                        copy.physical_status === 'checked_out' &&
                        copy.current_holder_id === currentUser?.id
                      "
                      @click="openTransferModal(copy)"
                      class="btn-action btn-transfer"
                      title="Transfer"
                    >
                      <i class="fas fa-exchange-alt"></i>
                    </button>
                    <button
                      @click="openHistoryModal(copy)"
                      class="btn-action btn-history"
                      title="View history"
                    >
                      <i class="fas fa-history"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="modal-footer">
        <div class="summary">
          <span>Total: {{ filteredCopies.length }}</span>
          <span v-if="filterStatus === 'my_copies'">
            | My Copies: {{ filteredCopies.length }}
          </span>
        </div>
        <button class="btn-close" @click="closeModal">Close</button>
      </div>
    </div>

    <!-- Check In Modal -->
    <div v-if="showCheckinModal" class="modal-overlay" @click="showCheckinModal = false">
      <div class="modal-content action-modal" @click.stop>
        <div class="modal-header">
          <h3>Check In Physical Copy</h3>
          <button @click="showCheckinModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>File:</label>
            <p>
              <strong>{{ fileToCheckin?.file_name }}</strong>
            </p>
            <p class="text-muted">
              {{ fileToCheckin?.category_name }} - VIN: {{ fileToCheckin?.vin || 'N/A' }}
            </p>
          </div>
          <div class="form-group">
            <label>Notes (optional):</label>
            <textarea
              v-model="checkinNotes"
              class="form-textarea"
              rows="3"
              placeholder="Add any notes about the check-in..."
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showCheckinModal = false" class="btn-cancel">Cancel</button>
          <button @click="confirmCheckin" :disabled="loading" class="btn-primary">Check In</button>
        </div>
      </div>
    </div>

    <!-- Transfer Modal -->
    <div v-if="showTransferModal" class="modal-overlay" @click="showTransferModal = false">
      <div class="modal-content action-modal" @click.stop>
        <div class="modal-header">
          <h3>Transfer Physical Copy</h3>
          <button @click="showTransferModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>File:</label>
            <p>
              <strong>{{ fileToTransfer?.file_name }}</strong>
            </p>
            <p class="text-muted">
              {{ fileToTransfer?.category_name }} - VIN: {{ fileToTransfer?.vin || 'N/A' }}
            </p>
          </div>
          <div class="form-group">
            <label>Transfer to: *</label>
            <select v-model="transferToUserId" class="form-select" required>
              <option value="">Select user...</option>
              <option v-for="user in availableUsers" :key="user.id" :value="user.id">
                {{ user.username }} ({{ user.email }})
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>Expected Return Date:</label>
            <input v-model="transferExpectedReturn" type="date" class="form-input" />
          </div>
          <div class="form-group">
            <label>Notes:</label>
            <textarea
              v-model="transferNotes"
              class="form-textarea"
              rows="3"
              placeholder="Add transfer notes..."
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showTransferModal = false" class="btn-cancel">Cancel</button>
          <button
            @click="confirmTransfer"
            :disabled="!transferToUserId || loading"
            class="btn-primary"
          >
            Transfer
          </button>
        </div>
      </div>
    </div>

    <!-- Transfer History Modal -->
    <div v-if="showHistoryModal" class="modal-overlay" @click="showHistoryModal = false">
      <div class="modal-content history-modal" @click.stop>
        <div class="modal-header">
          <h3>Transfer History</h3>
          <button @click="showHistoryModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="file-info-header">
            <p>
              <strong>{{ fileForHistory?.file_name }}</strong>
            </p>
            <p class="text-muted">
              {{ fileForHistory?.category_name }} - VIN: {{ fileForHistory?.vin || 'N/A' }}
            </p>
          </div>

          <div v-if="loading" class="loading-state">
            <i class="fas fa-spinner fa-spin"></i>
            <span>Loading history...</span>
          </div>

          <div v-else-if="transferHistory.length === 0" class="empty-state">
            <i class="fas fa-history"></i>
            <p>No transfer history found</p>
          </div>

          <div v-else class="history-list">
            <div
              v-for="(transfer, index) in transferHistory"
              :key="transfer.id"
              class="history-item"
            >
              <div class="history-header">
                <span class="history-number">#{{ transferHistory.length - index }}</span>
                <span class="history-date">{{ formatDate(transfer.transferred_at) }}</span>
              </div>
              <div class="history-content">
                <div class="history-transfer">
                  <i class="fas fa-arrow-right"></i>
                  <span v-if="transfer.from_username">
                    {{ transfer.from_username }}
                  </span>
                  <span v-else class="text-muted">Available</span>
                  <i class="fas fa-arrow-right"></i>
                  <span>{{ transfer.to_username }}</span>
                </div>
                <div v-if="transfer.notes" class="history-notes">
                  <i class="fas fa-comment"></i>
                  {{ transfer.notes }}
                </div>
                <div v-if="transfer.return_expected_date" class="history-return">
                  <i class="fas fa-calendar"></i>
                  Expected return: {{ formatDate(transfer.return_expected_date) }}
                </div>
                <div v-if="transfer.returned_at" class="history-returned">
                  <i class="fas fa-check-circle"></i>
                  Returned: {{ formatDate(transfer.returned_at) }}
                </div>
                <div class="history-by">
                  <i class="fas fa-user"></i>
                  Transferred by: {{ transfer.transferred_by_username }}
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showHistoryModal = false" class="btn-close">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
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

.tracking-modal {
  background: white;
  border-radius: 8px;
  width: 95%;
  max-width: 1200px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
  padding: 0.5rem;
}

.close-btn:hover {
  color: #374151;
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.error-message,
.success-message {
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.error-message {
  background-color: #fef2f2;
  color: #dc2626;
}

.success-message {
  background-color: #f0fdf4;
  color: #10b981;
}

.filters-section {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background-color: #f9fafb;
  border-radius: 4px;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-group label {
  font-weight: 500;
  color: #374151;
}

.filter-select {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
}

.search-group {
  flex: 1;
  position: relative;
  min-width: 300px;
}

.search-input {
  width: 100%;
  padding: 0.5rem 2.5rem 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
}

.search-icon {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #6b7280;
}

.empty-state i {
  font-size: 3rem;
  color: #d1d5db;
  margin-bottom: 1rem;
}

.copies-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

thead {
  background-color: #f9fafb;
}

th {
  padding: 0.75rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  font-size: 0.875rem;
  border-bottom: 2px solid #e5e7eb;
}

td {
  padding: 0.75rem;
  border-bottom: 1px solid #e5e7eb;
  font-size: 0.875rem;
}

tr:hover {
  background-color: #f9fafb;
}

.overdue-row {
  background-color: #fef2f2;
}

.file-link {
  color: #3b82f6;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.file-link:hover {
  text-decoration: underline;
}

.category-badge {
  padding: 0.25rem 0.5rem;
  background-color: #e0f2fe;
  color: #0369a1;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.875rem;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.status-badge.available {
  background-color: #f0fdf4;
  color: #10b981;
}

.status-badge.checked-out {
  background-color: #fef2f2;
  color: #dc2626;
}

.text-muted {
  color: #6b7280;
}

.overdue-text {
  color: #dc2626;
  font-weight: 600;
}

.overdue-icon {
  margin-left: 0.25rem;
  color: #dc2626;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  background: none;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 0.5rem;
  cursor: pointer;
  color: #6b7280;
  transition: all 0.2s;
}

.btn-action:hover {
  background-color: #f9fafb;
  border-color: #d1d5db;
}

.btn-action.btn-checkin {
  color: #10b981;
}

.btn-action.btn-transfer {
  color: #3b82f6;
}

.btn-action.btn-history {
  color: #8b5cf6;
}

.modal-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.summary {
  color: #6b7280;
  font-size: 0.875rem;
}

.btn-close,
.btn-cancel,
.btn-primary {
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  border: none;
}

.btn-close,
.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-close:hover,
.btn-cancel:hover {
  background-color: #e5e7eb;
}

.btn-primary {
  background-color: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-modal,
.history-modal {
  max-width: 600px;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #374151;
}

.form-select,
.form-input,
.form-textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1rem;
}

.form-textarea {
  resize: vertical;
}

.file-info-header {
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e5e7eb;
}

.history-list {
  max-height: 400px;
  overflow-y: auto;
}

.history-item {
  padding: 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  margin-bottom: 0.75rem;
  background-color: #ffffff;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.history-number {
  font-weight: 600;
  color: #3b82f6;
}

.history-date {
  color: #6b7280;
  font-size: 0.875rem;
}

.history-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.history-transfer {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}

.history-notes,
.history-return,
.history-returned,
.history-by {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.history-returned {
  color: #10b981;
}
</style>
