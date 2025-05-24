<script setup>
import { ref, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  car: {
    type: Object,
    required: true
  },
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'save'])
const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const success = ref(null)

// File refs
const documentsFile = ref(null)
const sellPiFile = ref(null)
const buyPiFile = ref(null)

const handleFileUpload = async (file, type) => {
  if (!file) return null

  const formData = new FormData()
  formData.append('file', file)
  formData.append('type', type)
  formData.append('car_id', props.car.id)

  try {
    const response = await fetch('/api/upload_car_file.php', {
      method: 'POST',
      body: formData
    })
    
    const result = await response.json()
    if (!result.success) {
      throw new Error(result.error || `Failed to upload ${type} file`)
    }
    
    return result.path
  } catch (err) {
    throw new Error(`Error uploading ${type}: ${err.message}`)
  }
}

const handleSubmit = async () => {
  loading.value = true
  error.value = null
  success.value = null
  
  try {
    const updates = {}
    
    // Handle documents upload
    if (documentsFile.value) {
      const documentsPath = await handleFileUpload(documentsFile.value, 'documents')
      if (documentsPath) updates.path_documents = documentsPath
    }
    
    // Handle sell PI upload
    if (sellPiFile.value) {
      const sellPiPath = await handleFileUpload(sellPiFile.value, 'sell_pi')
      if (sellPiPath) updates.sell_pi_path = sellPiPath
    }
    
    // Handle buy PI upload
    if (buyPiFile.value) {
      const buyPiPath = await handleFileUpload(buyPiFile.value, 'buy_pi')
      if (buyPiPath) updates.buy_pi_path = buyPiPath
    }
    
    // If any files were uploaded, update the database
    if (Object.keys(updates).length > 0) {
      const setClause = Object.keys(updates)
        .map(key => `${key} = ?`)
        .join(', ')
      
      const result = await callApi({
        query: `UPDATE cars_stock SET ${setClause} WHERE id = ?`,
        params: [...Object.values(updates), props.car.id]
      })

      if (result.success) {
        success.value = 'Files uploaded successfully'
        emit('save', { ...props.car, ...updates })
      } else {
        throw new Error(result.error || 'Failed to update file paths')
      }
    } else {
      success.value = 'No files selected for upload'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred during upload'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = null
  documentsFile.value = null
  sellPiFile.value = null
  buyPiFile.value = null
  emit('close')
}

const handleFileChange = (event, fileRef) => {
  const file = event.target.files[0]
  if (file && file.type !== 'application/pdf') {
    alert('Only PDF files are allowed')
    event.target.value = ''
    return
  }
  fileRef.value = file
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Upload Car Files</h3>
        <button class="close-btn" @click="closeModal">&times;</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="documents">Documents:</label>
          <div class="file-input-container">
            <input 
              type="file"
              id="documents"
              accept=".pdf"
              @change="e => handleFileChange(e, documentsFile)"
              class="file-input"
            />
            <span class="current-file" v-if="props.car.path_documents">
              Current: {{ props.car.path_documents.split('/').pop() }}
            </span>
          </div>
        </div>

        <div class="form-group">
          <label for="sell-pi">Sell PI:</label>
          <div class="file-input-container">
            <input 
              type="file"
              id="sell-pi"
              accept=".pdf"
              @change="e => handleFileChange(e, sellPiFile)"
              class="file-input"
            />
            <span class="current-file" v-if="props.car.sell_pi_path">
              Current: {{ props.car.sell_pi_path.split('/').pop() }}
            </span>
          </div>
        </div>

        <div class="form-group">
          <label for="buy-pi">Buy PI:</label>
          <div class="file-input-container">
            <input 
              type="file"
              id="buy-pi"
              accept=".pdf"
              @change="e => handleFileChange(e, buyPiFile)"
              class="file-input"
            />
            <span class="current-file" v-if="props.car.buy_pi_path">
              Current: {{ props.car.buy_pi_path.split('/').pop() }}
            </span>
          </div>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
        <div v-if="success" class="success-message">
          {{ success }}
        </div>
      </div>

      <div class="modal-footer">
        <button 
          class="cancel-btn" 
          @click="closeModal"
          :disabled="loading"
        >
          Cancel
        </button>
        <button 
          class="save-btn" 
          @click="handleSubmit"
          :disabled="loading"
        >
          {{ loading ? 'Uploading...' : 'Upload Files' }}
        </button>
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

.modal-content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.file-input-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-input {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.current-file {
  font-size: 12px;
  color: #6b7280;
}

.error-message {
  color: #ef4444;
  margin-bottom: 16px;
  font-size: 14px;
}

.success-message {
  color: #10b981;
  margin-bottom: 16px;
  font-size: 14px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.cancel-btn, .save-btn {
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  border: none;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
}

.cancel-btn:hover {
  background-color: #e5e7eb;
}

.save-btn:hover {
  background-color: #2563eb;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style> 