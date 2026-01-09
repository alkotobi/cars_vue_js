<script setup>
import { ref, onMounted, watch, computed, nextTick } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import { ElSelect, ElOption } from 'element-plus'
import 'element-plus/dist/index.css'
import NotesTable from './NotesTable.vue'
import NotesManagementModal from './NotesManagementModal.vue'
import AddBrokerDialog from './AddBrokerDialog.vue'

const { t } = useEnhancedI18n()
const props = defineProps({
  mode: {
    type: String,
    default: 'add', // 'add' or 'edit'
    validator: (value) => ['add', 'edit'].includes(value),
  },
  billData: {
    type: Object,
    default: () => ({
      id: null,
      id_broker: null,
      date_sell: new Date().toISOString().split('T')[0],
      notes: '',
      id_user: null,
    }),
  },
})
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)
const can_confirm_payment = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_confirm_payment')
})

const emit = defineEmits(['save', 'cancel'])

// Notes management modal state
const showNotesModal = ref(false)

const { callApi } = useApi()
const brokers = ref([])
const filteredBrokers = ref([])
const users = ref([])
const filteredUsers = ref([])
const loading = ref(false)
const error = ref(null)
const showAddBrokerDialog = ref(false)

// Add loading state for form submission
const isSubmitting = ref(false)

const formData = ref({
  id: null,
  id_broker: null,
  date_sell: new Date().toISOString().split('T')[0],
  notes: '', // This will hold the current note being edited (latest note or new note)
  notesArray: [], // This will hold the full JSON array of notes
  is_batch_sell: false,
  id_user: null,
  payment_confirmed: false,
})

// Watch for changes in billData prop
watch(
  () => props.billData,
  (newData) => {
    if (newData) {
      // Parse notes JSON if it exists
      let notesArray = []
      let currentNote = ''
      
      if (newData.notes) {
        try {
          // Try to parse as JSON (new format)
          if (typeof newData.notes === 'string' && newData.notes.trim().startsWith('[')) {
            notesArray = JSON.parse(newData.notes)
          } else if (Array.isArray(newData.notes)) {
            notesArray = newData.notes
          } else {
            // Old format (plain text) - convert to JSON array
            notesArray = [{
              id_user: null,
              note: newData.notes,
              timestamp: newData.time_created || new Date().toISOString().slice(0, 19).replace('T', ' ')
            }]
          }
          
          // Get the latest note for display
          if (notesArray.length > 0) {
            currentNote = notesArray[notesArray.length - 1].note || ''
          }
        } catch (e) {
          // If parsing fails, treat as old format (plain text)
          currentNote = newData.notes
          notesArray = [{
            id_user: null,
            note: newData.notes,
            timestamp: newData.time_created || new Date().toISOString().slice(0, 19).replace('T', ' ')
          }]
        }
      }
      
      formData.value = {
        ...newData,
        notes: currentNote,
        notesArray: notesArray,
        // Convert is_batch_sell from database format (0/1) to boolean
        is_batch_sell: Boolean(newData.is_batch_sell),
        // Convert payment_confirmed from database format (0/1) to boolean
        payment_confirmed: Boolean(newData.payment_confirmed),
      }

      // Note: id_user should not be changed when editing
      // It represents the original creator of the sell bill
      // For non-admin users creating a new bill, id_user will be set in onMounted

      // Ensure date is in the correct format for the date input
      if (formData.value.date_sell) {
        const date = new Date(formData.value.date_sell)
        if (!isNaN(date)) {
          formData.value.date_sell = date.toISOString().split('T')[0]
        }
      }
    }
  },
  { immediate: true },
)

// Watch for changes in isAdmin to ensure id_user is set correctly (only for new bills)
watch(
  () => isAdmin.value,
  (newIsAdmin) => {
    // For non-admin users creating a new bill, ensure id_user is set to current user's ID
    // Don't change id_user when editing existing bills
    if (!newIsAdmin && user.value?.id && props.mode === 'add') {
      formData.value.id_user = user.value.id
    }
  },
)

const fetchBrokers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name, mobiles
        FROM clients
        ORDER BY name ASC
      `,
      params: [],
    })

    if (result.success) {
      brokers.value = result.data
      filteredBrokers.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch brokers'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, username, email
        FROM users
        ORDER BY username ASC
      `,
      params: [],
    })

    if (result.success) {
      users.value = result.data
      filteredUsers.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch users'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const remoteMethod = (query) => {
  if (query) {
    // Filter the existing brokers data
    filteredBrokers.value = brokers.value.filter((broker) =>
      broker.name.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    // If no query, show all brokers
    filteredBrokers.value = brokers.value
  }
}

const remoteMethodUsers = (query) => {
  if (query) {
    // Filter the existing users data
    filteredUsers.value = users.value.filter((user) =>
      user.username.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    // If no query, show all users
    filteredUsers.value = users.value
  }
}

const handleBrokerChange = () => {
  // Handle broker change if needed
}

const openAddBrokerDialog = () => {
  showAddBrokerDialog.value = true
}

const closeAddBrokerDialog = () => {
  showAddBrokerDialog.value = false
}

const handleBrokerSaved = async (newBrokerData) => {
  // Refresh brokers list
  await fetchBrokers()
  // Ensure filtered list includes all brokers (in case there was a previous search)
  filteredBrokers.value = brokers.value
  // Wait for Vue to update the DOM with the new broker list
  await nextTick()
  // Auto-select the newly created broker
  formData.value.id_broker = newBrokerData.id
  closeAddBrokerDialog()
}

const saveBill = async () => {
  // Prevent multiple submissions
  if (isSubmitting.value) {
    return
  }

  // Validate required fields
  if (!formData.value.id_broker) {
    error.value = t('sellBills.brokerRequired') || 'Broker is required'
    return
  }

  try {
    isSubmitting.value = true
    loading.value = true
    error.value = null

    let result

    if (props.mode === 'add') {

      // Prepare notes JSON for new bill
      let notesJson = null
      if (formData.value.notes && formData.value.notes.trim() !== '') {
        const notesArray = [{
          id_user: user.value?.id || null,
          note: formData.value.notes.trim(),
          timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
        }]
        notesJson = JSON.stringify(notesArray)
      }

      // First insert the bill to get the ID
      result = await callApi({
        query: `
          INSERT INTO sell_bill (id_broker, date_sell, notes, id_user, is_batch_sell, payment_confirmed, time_created)
          VALUES (?, ?, ?, ?, ?, ?, NOW())
        `,
        params: [
          formData.value.id_broker || null,
          formData.value.date_sell,
          notesJson,
          formData.value.id_user || null,
          formData.value.is_batch_sell ? 1 : 0,
          can_confirm_payment.value && formData.value.payment_confirmed ? 1 : 0,
        ],
      })

      if (result.success && result.lastInsertId) {
        // Generate bill_ref
        const username = user.value?.username?.substring(0, 3).toUpperCase() || 'USR'
        const userId = (user.value?.id || 0).toString().padStart(3, '0')
        const date = new Date(formData.value.date_sell)
        const dateStr = date
          .toLocaleDateString('en-GB', {
            day: '2-digit',
            month: '2-digit',
            year: '2-digit',
          })
          .replace(/\//g, '')
        const billId = result.lastInsertId.toString().padStart(3, '0')
        const billRef = `${username}${userId}${dateStr}${billId}`


        // Update the bill with the generated bill_ref
        const updateResult = await callApi({
          query: `
            UPDATE sell_bill 
            SET bill_ref = ?
            WHERE id = ?
          `,
          params: [billRef, result.lastInsertId],
        })

        if (!updateResult.success) {
          error.value = 'Failed to update bill reference'
          return
        }

        // Get the complete bill data after update
        const fetchResult = await callApi({
          query: `
            SELECT 
              sb.id,
              sb.id_broker,
              sb.date_sell,
              sb.notes,
              sb.bill_ref,
              c.name as broker_name
            FROM sell_bill sb
            LEFT JOIN clients c ON sb.id_broker = c.id AND c.is_broker = 1
            WHERE sb.id = ?
          `,
          params: [result.lastInsertId],
        })

        if (fetchResult.success && fetchResult.data.length > 0) {
          emit('save', fetchResult.data[0])
        } else {
          error.value = 'Failed to fetch updated bill data'
        }
      } else {
        error.value = result.error || 'Failed to create sell bill'
      }
    } else {
      // When editing, preserve the original id_user (don't update it)
      // id_user should represent the original creator of the sell bill
      // Track who edited the bill in the edited_by JSON column
      
      // Get current edited_by and notes data
      const currentBillResult = await callApi({
        query: 'SELECT edited_by, notes, id_user FROM sell_bill WHERE id = ?',
        params: [formData.value.id],
        requiresAuth: true,
      })
      
      let editedByArray = []
      let notesArray = formData.value.notesArray || []
      
      if (currentBillResult.success && currentBillResult.data.length > 0) {
        // Parse edited_by JSON
        const currentEditedBy = currentBillResult.data[0].edited_by
        try {
          editedByArray = currentEditedBy ? JSON.parse(currentEditedBy) : []
        } catch (e) {
          editedByArray = []
        }
        
        // Parse notes JSON if not already parsed
        if (notesArray.length === 0) {
          const currentNotes = currentBillResult.data[0].notes
          try {
            if (currentNotes) {
              if (typeof currentNotes === 'string' && currentNotes.trim().startsWith('[')) {
                notesArray = JSON.parse(currentNotes)
              } else if (Array.isArray(currentNotes)) {
                notesArray = currentNotes
              }
            }
          } catch (e) {
            notesArray = []
          }
        }
      }
      
      // For edit mode, notes are managed through the modal
      // Only add a new note if notes were changed via the form (for backward compatibility)
      // In practice, notes should be managed through the modal
      const currentNoteText = formData.value.notes?.trim() || ''
      const lastNote = notesArray.length > 0 ? notesArray[notesArray.length - 1] : null
      const lastNoteText = lastNote?.note?.trim() || ''
      
      // Use the notesArray from formData if it exists (updated via modal)
      if (formData.value.notesArray && formData.value.notesArray.length > 0) {
        notesArray = formData.value.notesArray
      } else if (currentNoteText !== lastNoteText && currentNoteText !== '') {
        // Only add note if it's different and not empty (for backward compatibility)
        notesArray.push({
          id_user: user.value?.id || null,
          note: currentNoteText,
          timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
        })
      }
      
      // Add new edit entry
      const newEditEntry = {
        id_user: user.value?.id || null,
        timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
      }
      editedByArray.push(newEditEntry)
      
      // Convert to JSON strings
      const editedByJson = JSON.stringify(editedByArray)
      const notesJson = notesArray.length > 0 ? JSON.stringify(notesArray) : null
      
      // Build UPDATE query - include id_user if admin is editing
      let updateQuery
      let updateParams
      
      if (isAdmin.value) {
        // Admin can update id_user (created by)
        updateQuery = 'UPDATE sell_bill SET id_broker = ?, date_sell = ?, notes = ?, id_user = ?, is_batch_sell = ?, payment_confirmed = ?, edited_by = ? WHERE id = ?'
        updateParams = [
          formData.value.id_broker || null,
          formData.value.date_sell,
          notesJson,
          formData.value.id_user || null,
          formData.value.is_batch_sell ? 1 : 0,
          can_confirm_payment.value && formData.value.payment_confirmed ? 1 : 0,
          editedByJson,
          formData.value.id,
        ]
      } else {
        // Non-admin cannot update id_user
        updateQuery = 'UPDATE sell_bill SET id_broker = ?, date_sell = ?, notes = ?, is_batch_sell = ?, payment_confirmed = ?, edited_by = ? WHERE id = ?'
        updateParams = [
          formData.value.id_broker || null,
          formData.value.date_sell,
          notesJson,
          formData.value.is_batch_sell ? 1 : 0,
          can_confirm_payment.value && formData.value.payment_confirmed ? 1 : 0,
          editedByJson,
          formData.value.id,
        ]
      }
      
      result = await callApi({
        query: updateQuery,
        params: updateParams,
        user_id: user.value?.id || null,
        is_admin: isAdmin.value,
        requiresAuth: true,
      })

      if (result.success) {
        emit('save', result.data)
      } else {
        error.value = result.error || `Failed to ${props.mode} sell bill`
      }
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    isSubmitting.value = false
  }
}

// Handle notes updated from modal
const handleNotesUpdated = (updatedNotes) => {
  formData.value.notesArray = updatedNotes
  // Update the latest note for display compatibility
  if (updatedNotes.length > 0) {
    formData.value.notes = updatedNotes[updatedNotes.length - 1].note || ''
  } else {
    formData.value.notes = ''
  }
}

// Handle Created By change
const handleCreatedByChange = (val) => {
  // v-model automatically updates formData.value.id_user
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)

    // For non-admin users creating a new bill, ensure id_user is set to current user's ID
    // Don't change id_user when editing existing bills (it should preserve the original creator)
    if (!isAdmin.value && user.value?.id && props.mode === 'add') {
      formData.value.id_user = user.value.id
    }

    fetchBrokers()
    fetchUsers()
  }
})
</script>

<template>
  <div class="sell-bill-form">
    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>{{ isSubmitting ? t('sellBills.saving') : t('sellBills.loading') }}</span>
    </div>

    <h3 class="form-title">
      <i class="fas fa-file-invoice-dollar"></i>
      {{ mode === 'add' ? t('sellBills.add_new_sell_bill') : t('sellBills.edit_sell_bill') }}
    </h3>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <form @submit.prevent="saveBill" class="form-content">
      <div class="form-group">
        <label for="broker">
          <i class="fas fa-user-tie"></i>
          {{ t('sellBills.broker') }}: <span class="required">*</span>
        </label>
        <div class="input-with-button">
          <el-select
            id="broker"
            v-model="formData.id_broker"
            filterable
            remote
            :remote-method="remoteMethod"
            :loading="loading"
            :placeholder="t('sellBills.select_broker')"
            class="broker-select"
            style="width: 100%"
          >
            <el-option
              v-for="broker in filteredBrokers"
              :key="broker.id"
              :label="broker.name"
              :value="broker.id"
            >
              <i class="fas fa-user-tie"></i>
              {{ broker.name }}
              <small v-if="broker.mobiles">
                <i class="fas fa-phone"></i>
                {{ broker.mobiles }}
              </small>
            </el-option>
          </el-select>
          <button
            type="button"
            @click="openAddBrokerDialog"
            class="btn-add-broker"
            :title="t('sellBills.addBroker') || 'Add New Broker'"
          >
            <i class="fas fa-plus"></i>
          </button>
        </div>
      </div>

      <div v-if="isAdmin && mode === 'add'" class="form-group">
        <label for="user">
          <i class="fas fa-user"></i>
          {{ t('sellBills.user') }}
        </label>
        <el-select
          v-model="formData.id_user"
          filterable
          remote
          :remote-method="remoteMethodUsers"
          :loading="loading"
          :placeholder="t('sellBills.select_user')"
          class="user-select"
        >
          <el-option
            v-for="user in filteredUsers"
            :key="user.id"
            :label="user.username"
            :value="user.id"
          >
            <i class="fas fa-user"></i>
            {{ user.username }}
            <small v-if="user.email">
              <i class="fas fa-envelope"></i>
              {{ user.email }}
            </small>
          </el-option>
        </el-select>
      </div>
      <div v-if="isAdmin && mode === 'edit'" class="form-group">
        <label for="created_by_user">
          <i class="fas fa-user"></i>
          {{ t('sellBills.created_by') || 'Created By' }}
        </label>
        <el-select
          v-model="formData.id_user"
          filterable
          remote
          :remote-method="remoteMethodUsers"
          :loading="loading"
          :placeholder="t('sellBills.select_user') || 'Select user'"
          class="user-select"
          :disabled="isSubmitting"
          @change="handleCreatedByChange"
        >
          <el-option
            v-for="user in filteredUsers"
            :key="user.id"
            :label="user.username"
            :value="user.id"
          >
            <div style="display: flex; flex-direction: column">
              <span>{{ user.username }}</span>
              <small v-if="user.email" style="color: #6b7280; font-size: 0.75rem">
                <i class="fas fa-envelope"></i>
                {{ user.email }}
              </small>
            </div>
          </el-option>
        </el-select>
        <small class="help-text">{{ t('sellBills.created_by_help') || 'Original creator (admin can change)' }}</small>
      </div>

      <div class="form-group">
        <label for="date">
          <i class="fas fa-calendar"></i>
          {{ t('sellBills.date') }}
        </label>
        <input
          type="date"
          id="date"
          v-model="formData.date_sell"
          :disabled="isSubmitting"
          required
        />
      </div>

      <!-- Notes Table (read-only) -->
      <div v-if="mode === 'edit' && formData.notesArray && formData.notesArray.length > 0" class="form-group">
        <NotesTable
          :notes="formData.notesArray || []"
          :users="filteredUsers"
          :mode="mode"
          @manage-notes="showNotesModal = true"
        />
      </div>
      
      <!-- Notes Textarea (for new bills only) -->
      <div v-if="mode === 'add'" class="form-group">
        <label for="notes">
          <i class="fas fa-sticky-note"></i>
          {{ t('sellBills.notes') }}
        </label>
        <textarea
          id="notes"
          v-model="formData.notes"
          :disabled="isSubmitting"
          :placeholder="t('sellBills.enter_notes_placeholder') || 'Enter notes...'"
          rows="4"
        ></textarea>
      </div>
      
      <!-- Notes Management Modal -->
      <NotesManagementModal
        v-if="mode === 'edit' && formData.id"
        :show="showNotesModal"
        :notes="formData.notesArray || []"
        :users="filteredUsers"
        :current-user-id="user?.id"
        :is-admin="isAdmin"
        :bill-id="formData.id"
        @close="showNotesModal = false"
        @notes-updated="handleNotesUpdated"
      />

      <div class="form-group">
        <div class="checkbox-wrapper">
          <input
            type="checkbox"
            id="is_batch_sell"
            v-model="formData.is_batch_sell"
            class="form-checkbox"
          />
          <label for="is_batch_sell" class="checkbox-label">{{ t('sellBills.batch_sell') }}</label>
        </div>
      </div>

      <div v-if="can_confirm_payment" class="form-group">
        <div class="checkbox-wrapper">
          <input
            type="checkbox"
            id="payment_confirmed"
            v-model="formData.payment_confirmed"
            class="form-checkbox"
          />
          <label for="payment_confirmed" class="checkbox-label">{{ t('sellBills.payment_confirmed') }}</label>
        </div>
      </div>

      <div class="form-actions">
        <button type="button" @click="$emit('cancel')" :disabled="isSubmitting" class="cancel-btn">
          <i class="fas fa-times"></i>
          {{ t('sellBills.cancel') }}
        </button>
        <button type="submit" :disabled="isSubmitting" class="save-btn">
          <i class="fas fa-save"></i>
          {{ isSubmitting ? t('sellBills.saving') : t('sellBills.save') }}
        </button>
      </div>
    </form>

    <!-- Add Broker Dialog -->
    <AddBrokerDialog
      :show="showAddBrokerDialog"
      @close="closeAddBrokerDialog"
      @saved="handleBrokerSaved"
    />
  </div>
</template>

<style scoped>
.sell-bill-form {
  position: relative;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  height: 100%;
  max-height: 90vh;
  overflow: hidden;
}

.loading-overlay {
  position: absolute;
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
  z-index: 10;
}

.form-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1f2937;
  margin-bottom: 1.5rem;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 0.5rem;
  margin-right: -0.5rem;
}

/* Custom scrollbar styling */
.form-content::-webkit-scrollbar {
  width: 8px;
}

.form-content::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.form-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.form-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  font-weight: 500;
}

.form-group label i {
  color: #6b7280;
}

.broker-select {
  width: 100%;
}

.input-with-button {
  display: flex;
  gap: 8px;
  align-items: center;
}

.input-with-button :deep(.el-select) {
  flex: 1;
}

.input-with-button :deep(.el-input__wrapper) {
  width: 100%;
}

.btn-add-broker {
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 38px;
  transition: background-color 0.2s;
}

.btn-add-broker:hover {
  background-color: #059669;
}

.user-select {
  width: 100%;
}

input[type='date'],
textarea {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

input[type='date']:focus,
textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

textarea {
  min-height: 100px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #e5e7eb;
  flex-shrink: 0;
}

.cancel-btn,
.save-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Element Plus Select Customization */
:deep(.el-select-dropdown__item) {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

:deep(.el-select-dropdown__item i) {
  color: #6b7280;
}

:deep(.el-select-dropdown__item small) {
  margin-left: auto;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

/* Add smooth transitions */
.form-group input,
.form-group textarea {
  transition: all 0.2s;
}

button i {
  transition: transform 0.2s;
}

button:hover:not(:disabled) i {
  transform: scale(1.1);
}

.checkbox-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 8px 0;
}

.form-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.checkbox-label {
  cursor: pointer;
  user-select: none;
  color: #374151;
  font-size: 0.95rem;
}

.readonly-input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 1rem;
  background-color: #f3f4f6;
  color: #6b7280;
  cursor: not-allowed;
}

.help-text {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.875rem;
  color: #6b7280;
  font-style: italic;
}

.notes-history {
  margin-bottom: 1rem;
  padding: 1rem;
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  max-height: 200px;
  overflow-y: auto;
}

.notes-history-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
  font-weight: 600;
  color: #374151;
  font-size: 0.9rem;
}

.notes-history-header i {
  color: #6b7280;
}

.notes-history-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.notes-history-item {
  padding: 0.75rem;
  background-color: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  border-left: 3px solid #3b82f6;
}

.note-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
  font-size: 0.75rem;
  color: #6b7280;
}

.note-user {
  font-weight: 500;
  color: #3b82f6;
}

.note-timestamp {
  color: #9ca3af;
  font-style: italic;
}

.note-text {
  color: #1f2937;
  font-size: 0.9rem;
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>
