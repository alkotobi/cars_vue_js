<script setup>
import { ref,   onMounted, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  editData: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['close', 'save'])

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const users = ref([])
const currentUser = ref(null)

const formData = ref({
  id: null,
  amount: null,
  from_user_id: null,
  to_user_id: null,
  notes: '',
  date_transfer: new Date().toISOString().split('T')[0],
})

// Define resetForm before using it in watch
const resetForm = () => {
  formData.value = {
    id: null,
    amount: null,
    from_user_id: currentUser.value?.id || null,
    to_user_id: null,
    notes: '',
    date_transfer: new Date().toISOString().split('T')[0],
  }
  error.value = null
}

// Watch for editData changes
watch(
  () => props.editData,
  (newData) => {
    if (newData) {
      formData.value = {
        id: newData.id,
        amount: newData.amount,
        from_user_id: newData.from_user_id,
        to_user_id: newData.to_user_id,
        notes: newData.notes || '',
        date_transfer: newData.date_transfer,
      }
    } else {
      resetForm()
    }
  },
  { immediate: true },
)

// Fetch users for dropdowns
const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, username FROM users WHERE id != ? ORDER BY username',
      params: [currentUser.value?.id],
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

const saveTransfer = async () => {
  if (!validateForm()) return

  loading.value = true
  error.value = null

  try {
    let result
    if (formData.value.id) {
      // Update existing transfer
      result = await callApi({
        query: `
          UPDATE transfers_inter 
          SET amount = ?, 
              date_transfer = ?, 
              from_user_id = ?, 
              to_user_id = ?, 
              notes = ?
          WHERE id = ?
        `,
        params: [
          formData.value.amount,
          formData.value.date_transfer,
          formData.value.from_user_id,
          formData.value.to_user_id,
          formData.value.notes || null,
          formData.value.id,
        ],
      })
    } else {
      // Create new transfer
      result = await callApi({
        query: `
          INSERT INTO transfers_inter 
          (amount, date_transfer, from_user_id, to_user_id, notes)
          VALUES (?, ?, ?, ?, ?)
        `,
        params: [
          formData.value.amount,
          formData.value.date_transfer,
          formData.value.from_user_id,
          formData.value.to_user_id,
          formData.value.notes || null,
        ],
      })
    }

    if (result.success) {
      emit('save')
      resetForm()
    } else {
      error.value = result.error || `Failed to ${formData.value.id ? 'update' : 'create'} transfer`
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const validateForm = () => {
  if (!formData.value.amount || formData.value.amount <= 0) {
    error.value = 'Please enter a valid amount'
    return false
  }

  if (!formData.value.from_user_id) {
    error.value = 'Please select the sender'
    return false
  }

  if (!formData.value.to_user_id) {
    error.value = 'Please select the receiver'
    return false
  }

  if (formData.value.from_user_id === formData.value.to_user_id) {
    error.value = 'Sender and receiver cannot be the same user'
    return false
  }

  return true
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    currentUser.value = JSON.parse(userStr)
    formData.value.from_user_id = currentUser.value.id
    fetchUsers()
  }
})
</script>

<template>
  <div v-if="visible" class="dialog-overlay">
    <div class="dialog">
      <div class="dialog-header">
        <h3>
          <i :class="formData.id ? 'fas fa-edit' : 'fas fa-plus-circle'"></i>
          {{ formData.id ? 'Edit' : 'New' }} Transfer
        </h3>
        <button @click="$emit('close')" class="close-btn" :disabled="loading">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div v-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <form @submit.prevent="saveTransfer" class="form" :class="{ 'is-loading': loading }">
        <div class="form-group">
          <label for="amount">
            <i class="fas fa-dollar-sign"></i>
            Amount: <span class="required">*</span>
          </label>
          <div class="input-wrapper">
            <input
              type="number"
              id="amount"
              v-model="formData.amount"
              step="0.01"
              min="0"
              required
              :disabled="loading"
              placeholder="Enter amount"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="from_user">
            <i class="fas fa-user-circle"></i>
            From User: <span class="required">*</span>
          </label>
          <div class="input-wrapper">
            <select
              id="from_user"
              v-model="formData.from_user_id"
              required
              :disabled="!formData.id || loading"
            >
              <option :value="currentUser?.id">{{ currentUser?.username }} (You)</option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label for="to_user">
            <i class="fas fa-user-circle"></i>
            To User: <span class="required">*</span>
          </label>
          <div class="input-wrapper">
            <select id="to_user" v-model="formData.to_user_id" required :disabled="loading">
              <option value="">Select User</option>
              <option v-for="user in users" :key="user.id" :value="user.id">
                {{ user.username }}
              </option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label for="date">
            <i class="fas fa-calendar-alt"></i>
            Date: <span class="required">*</span>
          </label>
          <div class="input-wrapper">
            <input
              type="date"
              id="date"
              v-model="formData.date_transfer"
              required
              :disabled="loading"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="notes">
            <i class="fas fa-sticky-note"></i>
            Notes:
          </label>
          <div class="input-wrapper">
            <textarea
              id="notes"
              v-model="formData.notes"
              rows="3"
              :disabled="loading"
              placeholder="Add any additional notes here"
            ></textarea>
          </div>
        </div>

        <div class="form-actions">
          <button type="button" @click="$emit('close')" class="cancel-btn" :disabled="loading">
            <i class="fas fa-times"></i>
            Cancel
          </button>
          <button type="submit" class="save-btn" :disabled="loading">
            <i
              :class="
                loading ? 'fas fa-spinner fa-spin' : formData.id ? 'fas fa-save' : 'fas fa-plus'
              "
            ></i>
            {{
              loading
                ? formData.id
                  ? 'Updating...'
                  : 'Saving...'
                : formData.id
                  ? 'Update'
                  : 'Save'
            }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
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
  backdrop-filter: blur(4px);
}

.dialog {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  box-shadow:
    0 4px 6px rgba(0, 0, 0, 0.1),
    0 1px 3px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.dialog-header {
  background-color: #f8fafc;
  padding: 16px 24px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-header h3 {
  margin: 0;
  color: #1e293b;
  font-size: 1.25rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dialog-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  color: #64748b;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.close-btn:hover:not(:disabled) {
  color: #ef4444;
  background-color: #fee2e2;
}

.close-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form {
  padding: 24px;
  position: relative;
}

.form.is-loading::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(2px);
}

.error-message {
  margin: 0 24px;
  padding: 12px;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 6px;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 8px;
  animation: slideIn 0.3s ease;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  color: #1e293b;
  font-weight: 500;
}

.form-group label i {
  color: #64748b;
  width: 16px;
}

.required {
  color: #ef4444;
}

.input-wrapper {
  position: relative;
}

input,
select,
textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 0.875rem;
  transition: all 0.2s ease;
  background-color: white;
}

input:focus,
select:focus,
textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

input:disabled,
select:disabled,
textarea:disabled {
  background-color: #f8fafc;
  cursor: not-allowed;
}

textarea {
  resize: vertical;
  min-height: 80px;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.form-actions button {
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.875rem;
}

.cancel-btn {
  background-color: #f1f5f9;
  border: 1px solid #e2e8f0;
  color: #64748b;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e2e8f0;
}

.save-btn {
  background-color: #3b82f6;
  border: none;
  color: white;
  flex: 1;
}

.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.save-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
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

@media (max-width: 640px) {
  .dialog {
    width: 95%;
    margin: 16px;
  }

  .form-actions {
    flex-direction: column;
  }

  .form-actions button {
    width: 100%;
    justify-content: center;
  }
}
</style>
