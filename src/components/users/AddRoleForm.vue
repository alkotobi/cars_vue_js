<script setup>
import { ref } from 'vue'
import { useApi } from '../../composables/useApi'

const emit = defineEmits(['role-added', 'close'])

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const formData = ref({
  role_name: '',
  description: '',
})

const handleSubmit = async () => {
  if (!formData.value.role_name.trim()) {
    error.value = 'Role name is required'
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: 'INSERT INTO roles (role_name, description) VALUES (?, ?)',
      params: [formData.value.role_name, formData.value.description || null],
    })

    if (result.success) {
      // Reset form
      formData.value = {
        role_name: '',
        description: '',
      }
      error.value = null
      emit('role-added')
      emit('close')
    } else {
      error.value = result.error || 'Failed to add role'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  formData.value = {
    role_name: '',
    description: '',
  }
  error.value = null
  emit('close')
}
</script>

<template>
  <teleport to="body">
    <div class="modal-overlay" @click="handleCancel">
      <div class="modal-content" @click.stop>
        <div class="form-header">
          <h3>
            <i class="fas fa-user-plus"></i>
            Add New Role
          </h3>
          <button @click="handleCancel" class="close-btn" :disabled="loading">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="handleSubmit" class="add-role-form">
          <div class="form-group">
            <label for="role_name">
              <i class="fas fa-tag"></i>
              Role Name <span class="required">*</span>
            </label>
            <input
              id="role_name"
              v-model="formData.role_name"
              type="text"
              class="form-input"
              placeholder="Enter role name"
              required
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label for="description">
              <i class="fas fa-info-circle"></i>
              Description
            </label>
            <textarea
              id="description"
              v-model="formData.description"
              class="form-textarea"
              placeholder="Enter role description (optional)"
              rows="3"
              :disabled="loading"
            ></textarea>
          </div>

          <div v-if="error" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ error }}
          </div>

          <div class="form-actions">
            <button type="button" @click="handleCancel" class="btn btn-cancel" :disabled="loading">
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button type="submit" class="btn btn-submit" :disabled="loading">
              <i v-if="loading" class="fas fa-spinner fa-spin"></i>
              <i v-else class="fas fa-check"></i>
              {{ loading ? 'Adding...' : 'Add Role' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.18);
  min-width: 400px;
  max-width: 500px;
  width: 90%;
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px;
  border-bottom: 2px solid #e5e7eb;
}

.form-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 12px;
}

.form-header h3 i {
  color: #10b981;
}

.close-btn {
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn:hover:not(:disabled) {
  background: #dc2626;
  transform: scale(1.05);
}

.close-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.add-role-form {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 500;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
}

.required {
  color: #ef4444;
}

.form-input,
.form-textarea {
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s ease;
  font-family: inherit;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:disabled,
.form-textarea:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.error-message {
  padding: 12px;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  color: #dc2626;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.btn-submit {
  background-color: #10b981;
  color: white;
}

.btn-submit:hover:not(:disabled) {
  background-color: #059669;
}
</style>

