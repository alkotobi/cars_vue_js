<template>
  <div v-if="show" class="modal-overlay" @click="handleCancel">
    <div class="modal-content json-modal" @click.stop>
      <div class="modal-header">
        <h3>Edit db_code.json</h3>
        <button @click="handleCancel" class="modal-close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div v-if="error" class="error-message" style="margin-bottom: 1rem;">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>
        <div v-if="loading" class="loading-state" style="margin-bottom: 1rem;">
          <i class="fas fa-spinner fa-spin"></i>
          <span>{{ loadingMessage }}</span>
        </div>
        <div v-else>
          <p>
            Editing <strong>db_code.json</strong> for database <strong>{{ database?.db_code }}</strong>
          </p>
          <p class="info-text" style="margin-bottom: 1rem;">
            Location: <code>{{ database?.js_dir }}/db_code.json</code>
          </p>
          <div class="form-group">
            <label for="json-content-input">JSON Content *</label>
            <textarea
              id="json-content-input"
              v-model="jsonContent"
              required
              rows="15"
              class="form-input json-textarea"
              placeholder='{"db_code": "your_code_here"}'
              :disabled="saving"
            ></textarea>
            <small class="help-text">Enter valid JSON content. The file will be created if it doesn't exist.</small>
          </div>
        </div>
      </div>
      <div class="modal-actions">
        <button @click="handleCancel" class="btn-cancel" :disabled="saving">Cancel</button>
        <button @click="handleSave" :disabled="!jsonContent.trim() || saving || loading" class="btn-primary">
          <i v-if="saving" class="fas fa-spinner fa-spin"></i>
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  database: {
    type: Object,
    default: null,
  },
  apiBaseUrl: {
    type: String,
    required: true,
  },
})

const emit = defineEmits(['close', 'saved'])

const jsonContent = ref('')
const loading = ref(false)
const loadingMessage = ref('')
const saving = ref(false)
const error = ref('')

// Reset state function (defined before watch to avoid hoisting issues)
const resetState = () => {
  jsonContent.value = ''
  error.value = ''
  loading.value = false
  loadingMessage.value = ''
  saving.value = false
}

// Watch for database changes to load file
watch(
  () => [props.show, props.database],
  async ([newShow, newDatabase]) => {
    if (newShow && newDatabase) {
      await loadJsonFile()
    } else if (!newShow) {
      // Reset state when modal closes
      resetState()
    }
  },
  { immediate: true }
)

const loadJsonFile = async () => {
  if (!props.database || !props.database.js_dir || props.database.js_dir.trim() === '') {
    error.value = 'JS directory (js_dir) is not configured for this database'
    jsonContent.value = JSON.stringify({ db_code: '' }, null, 2)
    return
  }

  loading.value = true
  loadingMessage.value = 'Loading db_code.json...'
  error.value = ''

  try {
    const response = await fetch(`${props.apiBaseUrl}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'read_db_code_json',
        database_id: props.database.id,
      }),
    })

    const result = await response.json()

    if (result.success) {
      // Format JSON content with pretty printing
      jsonContent.value = JSON.stringify(result.data.content, null, 2)
    } else {
      error.value = result.message || 'Failed to load db_code.json'
      // Still show modal with default content
      jsonContent.value = JSON.stringify({ db_code: '' }, null, 2)
    }
  } catch (err) {
    error.value = 'An error occurred while loading the file'
    console.error(err)
    // Show modal with default content
    jsonContent.value = JSON.stringify({ db_code: '' }, null, 2)
  } finally {
    loading.value = false
    loadingMessage.value = ''
  }
}

const handleSave = async () => {
  if (!jsonContent.value.trim()) {
    error.value = 'JSON content cannot be empty'
    return
  }

  // Validate JSON
  let parsedContent
  try {
    parsedContent = JSON.parse(jsonContent.value)
  } catch (err) {
    error.value = 'Invalid JSON format: ' + err.message
    return
  }

  if (!props.database) {
    error.value = 'No database selected'
    return
  }

  saving.value = true
  error.value = ''

  try {
    const response = await fetch(`${props.apiBaseUrl}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'write_db_code_json',
        database_id: props.database.id,
        content: parsedContent,
      }),
    })

    const result = await response.json()

    if (result.success) {
      emit('saved', result.message || 'db_code.json saved successfully')
      handleCancel()
    } else {
      error.value = result.message || 'Failed to save db_code.json'
    }
  } catch (err) {
    error.value = 'An error occurred while saving the file'
    console.error(err)
  } finally {
    saving.value = false
  }
}

const handleCancel = () => {
  resetState()
  emit('close')
}
</script>

<style scoped>
.json-modal {
  max-width: 800px;
  width: 90%;
}

.json-textarea {
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  resize: vertical;
}

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
  z-index: 2000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  max-height: 90vh;
  overflow-y: auto;
  width: 100%;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e4e7ed;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #303133;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #909399;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.modal-close:hover {
  background-color: #f5f7fa;
  color: #606266;
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #606266;
}

.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 1rem;
  transition: border-color 0.3s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #409eff;
}

.form-input:disabled {
  background-color: #f5f7fa;
  cursor: not-allowed;
}

.help-text {
  display: block;
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: #909399;
}

.info-text {
  font-size: 0.875rem;
  color: #606266;
}

.info-text code {
  background-color: #f5f7fa;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid #e4e7ed;
}

.btn-cancel,
.btn-primary {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-cancel {
  background-color: #f5f7fa;
  color: #606266;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #e4e7ed;
}

.btn-cancel:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background-color: #409eff;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #66b1ff;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  background-color: #fef0f0;
  color: #f56c6c;
  padding: 1rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #606266;
  padding: 1rem;
}

.loading-state i {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>

