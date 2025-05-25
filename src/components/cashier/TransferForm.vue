<script setup>
import { ref, defineProps, defineEmits, onMounted, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  editData: {
    type: Object,
    default: null
  }
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
  date_transfer: new Date().toISOString().split('T')[0]
})

// Define resetForm before using it in watch
const resetForm = () => {
  formData.value = {
    id: null,
    amount: null,
    from_user_id: currentUser.value?.id || null,
    to_user_id: null,
    notes: '',
    date_transfer: new Date().toISOString().split('T')[0]
  }
  error.value = null
}

// Watch for editData changes
watch(() => props.editData, (newData) => {
  if (newData) {
    formData.value = {
      id: newData.id,
      amount: newData.amount,
      from_user_id: newData.from_user_id,
      to_user_id: newData.to_user_id,
      notes: newData.notes || '',
      date_transfer: newData.date_transfer
    }
  } else {
    resetForm()
  }
}, { immediate: true })

// Fetch users for dropdowns
const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, username FROM users WHERE id != ? ORDER BY username',
      params: [currentUser.value?.id]
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
    let result;
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
          formData.value.id
        ]
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
          formData.value.notes || null
        ]
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
        <h3>{{ formData.id ? 'Edit' : 'New' }} Transfer</h3>
        <button @click="$emit('close')" class="close-btn">&times;</button>
      </div>

      <div v-if="error" class="error-message">{{ error }}</div>

      <form @submit.prevent="saveTransfer" class="form">
        <div class="form-group">
          <label for="amount">Amount: <span class="required">*</span></label>
          <input
            type="number"
            id="amount"
            v-model="formData.amount"
            step="0.01"
            min="0"
            required
          >
        </div>

        <div class="form-group">
          <label for="from_user">From User: <span class="required">*</span></label>
          <select id="from_user" v-model="formData.from_user_id" required :disabled="!formData.id">
            <option :value="currentUser?.id">{{ currentUser?.username }} (You)</option>
          </select>
        </div>

        <div class="form-group">
          <label for="to_user">To User: <span class="required">*</span></label>
          <select id="to_user" v-model="formData.to_user_id" required>
            <option value="">Select User</option>
            <option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.username }}
            </option>
          </select>
        </div>

        <div class="form-group">
          <label for="date">Date: <span class="required">*</span></label>
          <input
            type="date"
            id="date"
            v-model="formData.date_transfer"
            required
          >
        </div>

        <div class="form-group">
          <label for="notes">Notes:</label>
          <textarea
            id="notes"
            v-model="formData.notes"
            rows="3"
          ></textarea>
        </div>

        <div class="form-actions">
          <button type="button" @click="$emit('close')" class="cancel-btn">Cancel</button>
          <button type="submit" class="save-btn" :disabled="loading">
            {{ loading ? (formData.id ? 'Updating...' : 'Saving...') : (formData.id ? 'Update' : 'Save') }}
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
}

.dialog {
  background: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.5rem;
  color: #1a1a1a;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #666;
  cursor: pointer;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.form-group label {
  font-weight: 500;
  color: #374151;
}

.required {
  color: #ef4444;
}

input, select, textarea {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1rem;
}

textarea {
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 10px;
}

.cancel-btn, .save-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.save-btn:disabled {
  background-color: #d1d5db;
  cursor: not-allowed;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 15px;
}
</style> 