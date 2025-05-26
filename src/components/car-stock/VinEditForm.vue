<script setup>
import { ref, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  car: {
    type: Object,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])
const { callApi } = useApi()
const newVin = ref('')
const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)

const handleSubmit = async () => {
  if (isProcessing.value || loading.value) return
  if (!newVin.value.trim()) {
    error.value = 'VIN number is required'
    return
  }

  loading.value = true
  isProcessing.value = true
  error.value = null

  try {
    const result = await callApi({
      query: 'UPDATE cars_stock SET vin = ? WHERE id = ?',
      params: [newVin.value.trim(), props.car.id],
    })

    if (result.success) {
      emit('save', { ...props.car, vin: newVin.value.trim() })
      emit('close')
    } else {
      error.value = result.error || 'Failed to update VIN'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

const closeModal = () => {
  error.value = null
  newVin.value = ''
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content" :class="{ 'is-processing': isProcessing }">
      <div class="modal-header">
        <h3>
          <i class="fas fa-barcode"></i>
          Edit VIN Number
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="current-vin">
            <i class="fas fa-tag"></i>
            Current VIN:
          </label>
          <div class="input-wrapper">
            <input
              type="text"
              id="current-vin"
              :value="car.vin || 'N/A'"
              disabled
              class="input-disabled"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="new-vin">
            <i class="fas fa-edit"></i>
            New VIN:
          </label>
          <div class="input-wrapper">
            <input
              type="text"
              id="new-vin"
              v-model="newVin"
              placeholder="Enter new VIN number"
              class="input-field"
              :disabled="isProcessing"
            />
          </div>
        </div>

        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>
      </div>

      <div class="modal-footer">
        <button class="cancel-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
          Cancel
        </button>
        <button
          class="save-btn"
          @click="handleSubmit"
          :disabled="isProcessing"
          :class="{ 'is-processing': isProcessing }"
        >
          <i class="fas fa-save"></i>
          <span>{{ isProcessing ? 'Saving...' : 'Save Changes' }}</span>
          <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
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
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.input-field,
.input-disabled {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.input-disabled {
  background-color: #f3f4f6;
  color: #6b7280;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ef4444;
  margin-bottom: 16px;
  font-size: 14px;
  padding: 8px 12px;
  background-color: #fee2e2;
  border-radius: 4px;
}

.error-message i {
  font-size: 1.1em;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.cancel-btn,
.save-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  border: none;
  transition: all 0.2s ease;
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

.modal-content.is-processing {
  opacity: 0.7;
  pointer-events: none;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-wrapper i {
  position: absolute;
  right: 12px;
  color: #6b7280;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.save-btn.is-processing {
  position: relative;
}

.save-btn.is-processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

.loading-indicator {
  margin-left: 4px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fa-spin {
  animation: spin 1s linear infinite;
}
</style>
