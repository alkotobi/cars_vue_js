<template>
  <div class="invitations-table">
    <div class="table-header">
      <h3>{{ t('invitations.title') }}</h3>
      <button @click="showAddDialog = true" class="add-btn">
        <i class="fas fa-plus"></i> {{ t('invitations.addInvitation') }}
      </button>
    </div>

    <!-- Summary Section -->
    <div class="summary-section">
      <div class="summary-card">
        <div class="summary-icon">
          <i class="fas fa-calculator"></i>
        </div>
        <div class="summary-content">
          <div class="summary-label">{{ t('invitations.valueMinusPayment') }}</div>
          <div class="summary-value">{{ formatNumber(summaryStats.valueMinusPayment) }}</div>
        </div>
      </div>
    </div>

    <!-- Filters Section -->
    <div class="filters-section">
      <div class="filters-row">
        <div class="filter-group">
          <label>{{ t('invitations.searchName') }}</label>
          <input
            v-model="filters.name"
            type="text"
            :placeholder="t('invitations.searchNamePlaceholder')"
            @input="applyFilters"
          />
        </div>

        <div class="filter-group">
          <label>{{ t('invitations.searchPass') }}</label>
          <input
            v-model="filters.pass"
            type="text"
            :placeholder="t('invitations.searchPassPlaceholder')"
            @input="applyFilters"
          />
        </div>

        <div class="filter-group">
          <label>{{ t('invitations.dateFrom') }}</label>
          <input v-model="filters.dateFrom" type="date" @change="applyFilters" />
        </div>

        <div class="filter-group">
          <label>{{ t('invitations.dateTo') }}</label>
          <input v-model="filters.dateTo" type="date" @change="applyFilters" />
        </div>

        <div class="filter-group">
          <label>{{ t('invitations.valueRange') }}</label>
          <div class="range-inputs">
            <input
              v-model.number="filters.valueMin"
              type="number"
              :placeholder="t('invitations.min')"
              @input="applyFilters"
            />
            <span>-</span>
            <input
              v-model.number="filters.valueMax"
              type="number"
              :placeholder="t('invitations.max')"
              @input="applyFilters"
            />
          </div>
        </div>

        <div class="filter-group">
          <button @click="clearFilters" class="clear-filters-btn">
            <i class="fas fa-times"></i> {{ t('invitations.clearFilters') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="loading">
      {{ t('invitations.loading') }}
    </div>

    <div v-else-if="error" class="error">
      {{ error }}
    </div>

    <div v-else class="table-container">
      <table>
        <thead>
          <tr>
            <th @click="sortBy('id')" class="sortable">
              {{ t('invitations.id') }}
              <i :class="getSortIcon('id')"></i>
            </th>
            <th @click="sortBy('name')" class="sortable">
              {{ t('invitations.name') }}
              <i :class="getSortIcon('name')"></i>
            </th>
            <th @click="sortBy('pass')" class="sortable">
              {{ t('invitations.pass') }}
              <i :class="getSortIcon('pass')"></i>
            </th>
            <th @click="sortBy('dateInv')" class="sortable">
              {{ t('invitations.dateInv') }}
              <i :class="getSortIcon('dateInv')"></i>
            </th>
            <th @click="sortBy('value')" class="sortable">
              {{ t('invitations.value') }}
              <i :class="getSortIcon('value')"></i>
            </th>
            <th @click="sortBy('payment')" class="sortable">
              {{ t('invitations.payment') }}
              <i :class="getSortIcon('payment')"></i>
            </th>
            <th @click="sortBy('rate')" class="sortable">
              {{ t('invitations.rate') }}
              <i :class="getSortIcon('rate')"></i>
            </th>
            <th @click="sortBy('rem')" class="sortable">
              {{ t('invitations.rem') }}
              <i :class="getSortIcon('rem')"></i>
            </th>
            <th @click="sortBy('balance')" class="sortable">
              {{ t('invitations.balance') }}
              <i :class="getSortIcon('balance')"></i>
            </th>
            <th>{{ t('invitations.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="invitation in filteredAndSortedInvitations" :key="invitation.id">
            <td>{{ invitation.id }}</td>
            <td>{{ invitation.name }}</td>
            <td>{{ invitation.pass }}</td>
            <td>{{ formatDate(invitation.dateInv) }}</td>
            <td>{{ formatNumber(invitation.value) }}</td>
            <td>{{ formatNumber(invitation.payment) }}</td>
            <td>{{ formatNumber(invitation.rate) }}</td>
            <td>{{ invitation.rem }}</td>
            <td>{{ formatNumber(invitation.balance) }}</td>
            <td>
              <button @click="editInvitation(invitation)" class="edit-btn">
                <i class="fas fa-edit"></i>
              </button>
              <button @click="deleteInvitation(invitation.id)" class="delete-btn">
                <i class="fas fa-trash"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Results Summary -->
      <div class="results-summary">
        <p>
          {{
            t('invitations.showingResults', {
              shown: filteredAndSortedInvitations.length,
              total: invitations.length,
            })
          }}
        </p>
      </div>
    </div>

    <!-- Add/Edit Dialog -->
    <div v-if="showAddDialog || showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>
          {{ showEditDialog ? t('invitations.editInvitation') : t('invitations.addInvitation') }}
        </h3>
        <form @submit.prevent="saveInvitation">
          <div class="form-group">
            <label>{{ t('invitations.name') }}</label>
            <input v-model="form.name" type="text" required />
          </div>

          <div class="form-group">
            <label>{{ t('invitations.pass') }}</label>
            <input v-model="form.pass" type="text" maxlength="20" />
          </div>

          <div class="form-group">
            <label>{{ t('invitations.dateInv') }}</label>
            <input v-model="form.dateInv" type="datetime-local" required />
          </div>

          <div class="form-group">
            <label>{{ t('invitations.value') }}</label>
            <input v-model.number="form.value" type="number" step="0.01" required />
          </div>

          <div class="form-group">
            <label>{{ t('invitations.payment') }}</label>
            <input v-model.number="form.payment" type="number" step="0.01" required />
          </div>

          <div class="form-group">
            <label>{{ t('invitations.rate') }}</label>
            <input v-model.number="form.rate" type="number" step="0.01" required />
          </div>

          <div class="form-group">
            <label>{{ t('invitations.rem') }}</label>
            <textarea v-model="form.rem" rows="3"></textarea>
          </div>

          <div class="form-group">
            <label>{{ t('invitations.balance') }}</label>
            <input v-model.number="form.balance" type="number" step="0.01" required />
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="closeDialog" class="cancel-btn">
              {{ t('invitations.cancel') }}
            </button>
            <button type="submit" class="save-btn" :disabled="saving">
              <span v-if="saving" class="spinner"></span>
              {{ saving ? t('invitations.saving') : t('invitations.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useEnhancedI18n } from '../composables/useI18n'
import { useApi } from '../composables/useApi'

const { t } = useEnhancedI18n()
const { callApi } = useApi()

const invitations = ref([])
const loading = ref(false)
const error = ref(null)
const saving = ref(false)
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingId = ref(null)

// Sorting state
const sortField = ref('dateInv')
const sortDirection = ref('desc')

// Filters state
const filters = ref({
  name: '',
  pass: '',
  dateFrom: '',
  dateTo: '',
  valueMin: null,
  valueMax: null,
})

const form = ref({
  name: '',
  pass: '',
  dateInv: '',
  value: 0,
  payment: 0,
  rate: 0,
  rem: '',
  balance: 0,
})

// Computed properties for filtering and sorting
const filteredInvitations = computed(() => {
  return invitations.value.filter((invitation) => {
    // Name filter
    if (
      filters.value.name &&
      !invitation.name?.toLowerCase().includes(filters.value.name.toLowerCase())
    ) {
      return false
    }

    // Pass filter
    if (
      filters.value.pass &&
      !invitation.pass?.toLowerCase().includes(filters.value.pass.toLowerCase())
    ) {
      return false
    }

    // Date range filter
    if (filters.value.dateFrom) {
      const invitationDate = new Date(invitation.dateInv)
      const fromDate = new Date(filters.value.dateFrom)
      if (invitationDate < fromDate) {
        return false
      }
    }

    if (filters.value.dateTo) {
      const invitationDate = new Date(invitation.dateInv)
      const toDate = new Date(filters.value.dateTo)
      toDate.setHours(23, 59, 59) // Include the entire day
      if (invitationDate > toDate) {
        return false
      }
    }

    // Value range filter
    if (filters.value.valueMin !== null && invitation.value < filters.value.valueMin) {
      return false
    }

    if (filters.value.valueMax !== null && invitation.value > filters.value.valueMax) {
      return false
    }

    return true
  })
})

const filteredAndSortedInvitations = computed(() => {
  const sorted = [...filteredInvitations.value].sort((a, b) => {
    let aValue = a[sortField.value]
    let bValue = b[sortField.value]

    // Handle null/undefined values
    if (aValue === null || aValue === undefined) aValue = ''
    if (bValue === null || bValue === undefined) bValue = ''

    // Handle numeric values
    if (['value', 'payment', 'rate', 'balance', 'id'].includes(sortField.value)) {
      aValue = Number(aValue) || 0
      bValue = Number(bValue) || 0
    }

    // Handle date values
    if (sortField.value === 'dateInv') {
      aValue = new Date(aValue || 0)
      bValue = new Date(bValue || 0)
    }

    // Handle string values
    if (typeof aValue === 'string') {
      aValue = aValue.toLowerCase()
      bValue = bValue.toLowerCase()
    }

    if (sortDirection.value === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })

  return sorted
})

// Summary statistics computed property
const summaryStats = computed(() => {
  const stats = {
    valueMinusPayment: 0,
  }

  if (filteredInvitations.value.length > 0) {
    stats.valueMinusPayment = filteredInvitations.value.reduce((sum, invitation) => {
      const value = Number(invitation.value) || 0
      const payment = Number(invitation.payment) || 0
      return sum + (value - payment)
    }, 0)
  }

  return stats
})

const sortBy = (field) => {
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortDirection.value = 'asc'
  }
}

const getSortIcon = (field) => {
  if (sortField.value !== field) {
    return 'fas fa-sort'
  }
  return sortDirection.value === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'
}

const applyFilters = () => {
  // Filters are applied automatically through computed properties
}

const clearFilters = () => {
  filters.value = {
    name: '',
    pass: '',
    dateFrom: '',
    dateTo: '',
    valueMin: null,
    valueMax: null,
  }
}

const fetchInvitations = async () => {
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: 'SELECT * FROM merhab_invitations.invitations ORDER BY dateInv DESC',
      params: [],
    })

    if (result.success) {
      invitations.value = result.data
    } else {
      error.value = result.error || t('invitations.fetchError')
    }
  } catch (err) {
    error.value = t('invitations.fetchError')
    console.error('Error fetching invitations:', err)
  } finally {
    loading.value = false
  }
}

const saveInvitation = async () => {
  saving.value = true

  try {
    if (showEditDialog.value) {
      // Update existing invitation
      const result = await callApi({
        query: `
          UPDATE merhab_invitations.invitations 
          SET name = ?, pass = ?, dateInv = ?, value = ?, payment = ?, rate = ?, rem = ?, balance = ?
          WHERE id = ?
        `,
        params: [
          form.value.name,
          form.value.pass,
          form.value.dateInv,
          form.value.value,
          form.value.payment,
          form.value.rate,
          form.value.rem,
          form.value.balance,
          editingId.value,
        ],
      })

      if (result.success) {
        await fetchInvitations()
        closeDialog()
      } else {
        alert(result.error || t('invitations.saveError'))
      }
    } else {
      // Insert new invitation
      const result = await callApi({
        query: `
          INSERT INTO merhab_invitations.invitations (name, pass, dateInv, value, payment, rate, rem, balance)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `,
        params: [
          form.value.name,
          form.value.pass,
          form.value.dateInv,
          form.value.value,
          form.value.payment,
          form.value.rate,
          form.value.rem,
          form.value.balance,
        ],
      })

      if (result.success) {
        await fetchInvitations()
        closeDialog()
      } else {
        alert(result.error || t('invitations.saveError'))
      }
    }
  } catch (err) {
    alert(t('invitations.saveError'))
    console.error('Error saving invitation:', err)
  } finally {
    saving.value = false
  }
}

const editInvitation = (invitation) => {
  editingId.value = invitation.id
  form.value = { ...invitation }
  form.value.dateInv = invitation.dateInv ? invitation.dateInv.slice(0, 16) : ''
  showEditDialog.value = true
}

const deleteInvitation = async (id) => {
  if (!confirm(t('invitations.confirmDelete'))) {
    return
  }

  try {
    const result = await callApi({
      query: 'DELETE FROM merhab_invitations.invitations WHERE id = ?',
      params: [id],
    })

    if (result.success) {
      await fetchInvitations()
    } else {
      alert(result.error || t('invitations.deleteError'))
    }
  } catch (err) {
    alert(t('invitations.deleteError'))
    console.error('Error deleting invitation:', err)
  }
}

const closeDialog = () => {
  showAddDialog.value = false
  showEditDialog.value = false
  editingId.value = null
  form.value = {
    name: '',
    pass: '',
    dateInv: '',
    value: 0,
    payment: 0,
    rate: 0,
    rem: '',
    balance: 0,
  }
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString()
}

const formatNumber = (number, decimals = 0) => {
  if (number === null || number === undefined) return '0'
  return Number(number).toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  })
}

onMounted(() => {
  fetchInvitations()
})
</script>

<style scoped>
.invitations-table {
  padding: 20px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.table-header h3 {
  margin: 0;
  color: #2c3e50;
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.add-btn:hover {
  background-color: #45a049;
}

.summary-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.summary-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 15px;
  border-left: 4px solid #4caf50;
}

.summary-icon {
  width: 50px;
  height: 50px;
  background: linear-gradient(135deg, #4caf50, #45a049);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
}

.summary-content {
  flex: 1;
}

.summary-label {
  font-size: 12px;
  color: #6c757d;
  font-weight: 500;
  margin-bottom: 5px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.summary-value {
  font-size: 24px;
  font-weight: bold;
  color: #2c3e50;
}

.filters-section {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid #e9ecef;
}

.filters-row {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  align-items: end;
}

.filter-group {
  display: flex;
  flex-direction: column;
  min-width: 150px;
}

.filter-group label {
  font-weight: bold;
  margin-bottom: 5px;
  color: #495057;
  font-size: 12px;
}

.filter-group input {
  padding: 8px;
  border: 1px solid #ced4da;
  border-radius: 4px;
  font-size: 14px;
}

.range-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.range-inputs input {
  flex: 1;
  min-width: 80px;
}

.range-inputs span {
  color: #6c757d;
  font-weight: bold;
}

.clear-filters-btn {
  padding: 8px 12px;
  background-color: #6c757d;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 5px;
}

.clear-filters-btn:hover {
  background-color: #5a6268;
}

.loading,
.error {
  text-align: center;
  padding: 20px;
  font-size: 16px;
}

.error {
  color: #f44336;
}

.table-container {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

th,
td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

th {
  background-color: #f5f5f5;
  font-weight: bold;
  color: #333;
  cursor: pointer;
  user-select: none;
}

th.sortable:hover {
  background-color: #e9ecef;
}

th i {
  margin-left: 5px;
  color: #6c757d;
}

tr:hover {
  background-color: #f9f9f9;
}

.edit-btn,
.delete-btn {
  padding: 4px 8px;
  margin: 0 2px;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
}

.edit-btn {
  background-color: #2196f3;
  color: white;
}

.edit-btn:hover {
  background-color: #1976d2;
}

.delete-btn {
  background-color: #f44336;
  color: white;
}

.delete-btn:hover {
  background-color: #d32f2f;
}

.results-summary {
  margin-top: 15px;
  padding: 10px;
  background-color: #f8f9fa;
  border-radius: 4px;
  text-align: center;
  color: #6c757d;
  font-size: 14px;
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
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
  color: #333;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-group textarea {
  resize: vertical;
  min-height: 60px;
}

.dialog-buttons {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
}

.cancel-btn,
.save-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.cancel-btn {
  background-color: #6c757d;
  color: white;
}

.cancel-btn:hover {
  background-color: #5a6268;
}

.save-btn {
  background-color: #28a745;
  color: white;
}

.save-btn:hover {
  background-color: #218838;
}

.save-btn:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}

.spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid #ffffff;
  border-radius: 50%;
  border-top-color: transparent;
  animation: spin 1s ease-in-out infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .filters-row {
    flex-direction: column;
  }

  .filter-group {
    min-width: 100%;
  }

  .table-header {
    flex-direction: column;
    gap: 10px;
    align-items: stretch;
  }
}
</style>
