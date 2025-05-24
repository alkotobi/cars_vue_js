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

// Form fields
const price = ref(props.car.price_cell || '')
const freight = ref(props.car.freight || '')
const rate = ref(props.car.rate || '')

const handleSubmit = async () => {
  // Validate inputs
  if (price.value && isNaN(parseFloat(price.value))) {
    error.value = 'Price must be a valid number'
    return
  }
  if (freight.value && isNaN(parseFloat(freight.value))) {
    error.value = 'Freight must be a valid number'
    return
  }
  if (rate.value && isNaN(parseFloat(rate.value))) {
    error.value = 'Rate must be a valid number'
    return
  }

  loading.value = true
  error.value = null

  try {
    const updates = []
    const params = []
    
    if (price.value !== '') {
      updates.push('price_cell = ?')
      params.push(parseFloat(price.value))
    }
    
    if (freight.value !== '') {
      updates.push('freight = ?')
      params.push(parseFloat(freight.value))
    }
    
    if (rate.value !== '') {
      updates.push('rate = ?')
      params.push(parseFloat(rate.value))
    }
    
    if (updates.length === 0) {
      error.value = 'Please fill at least one field to update'
      loading.value = false
      return
    }
    
    params.push(props.car.id)

    const result = await callApi({
      query: `UPDATE cars_stock SET ${updates.join(', ')} WHERE id = ?`,
      params
    })

    if (result.success) {
      const updatedCar = { 
        ...props.car,
        price_cell: price.value !== '' ? parseFloat(price.value) : props.car.price_cell,
        freight: freight.value !== '' ? parseFloat(freight.value) : props.car.freight,
        rate: rate.value !== '' ? parseFloat(rate.value) : props.car.rate
      }
      emit('save', updatedCar)
      emit('close')
    } else {
      throw new Error(result.error || 'Failed to update money fields')
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  price.value = props.car.price_cell || ''
  freight.value = props.car.freight || ''
  rate.value = props.car.rate || ''
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Edit Money Fields</h3>
        <button class="close-btn" @click="closeModal">&times;</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="price">Price:</label>
          <div class="input-group">
            <span class="currency-symbol">$</span>
            <input 
              type="number"
              id="price"
              v-model="price"
              placeholder="Enter price"
              step="0.01"
              min="0"
              class="input-field"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="freight">Freight:</label>
          <div class="input-group">
            <span class="currency-symbol">$</span>
            <input 
              type="number"
              id="freight"
              v-model="freight"
              placeholder="Enter freight"
              step="0.01"
              min="0"
              class="input-field"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="rate">Rate:</label>
          <div class="input-group">
            <input 
              type="number"
              id="rate"
              v-model="rate"
              placeholder="Enter rate"
              step="0.01"
              min="0"
              class="input-field"
            />
          </div>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
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
          {{ loading ? 'Saving...' : 'Save Changes' }}
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

.input-group {
  display: flex;
  align-items: center;
  width: 100%;
}

.currency-symbol {
  padding: 8px 12px;
  background-color: #f3f4f6;
  border: 1px solid #d1d5db;
  border-right: none;
  border-radius: 4px 0 0 4px;
  color: #6b7280;
}

.input-field {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 0 4px 4px 0;
  font-size: 14px;
}

.input-field:not(.currency-input) {
  border-radius: 4px;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.error-message {
  color: #ef4444;
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