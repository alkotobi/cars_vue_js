<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)
const hasChanges = ref(false)

// Form data
const formData = ref({
  rate: '',
  freight_small: '',
  freight_big: '',
})

const fetchDefaults = async () => {
  loading.value = true
  error.value = null
  try {
    const result = await callApi({
      query: 'SELECT * FROM defaults LIMIT 1',
    })
    if (result.success && result.data.length > 0) {
      const defaults = result.data[0]
      formData.value = {
        rate: defaults.rate?.toString() || '',
        freight_small: defaults.freight_small?.toString() || '',
        freight_big: defaults.freight_big?.toString() || '',
      }
    } else {
      // If no record exists, create one with default values
      await callApi({
        query: `
          INSERT INTO defaults (rate, freight_small, freight_big)
          VALUES (0, 0, 0)
        `,
      })
      formData.value = {
        rate: '0',
        freight_small: '0',
        freight_big: '0',
      }
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const handleInput = () => {
  hasChanges.value = true
}

const handleSubmit = async () => {
  if (isProcessing.value) return
  isProcessing.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        UPDATE defaults 
        SET rate = ?, freight_small = ?, freight_big = ?
        WHERE id = 1
      `,
      params: [
        parseFloat(formData.value.rate) || 0,
        parseFloat(formData.value.freight_small) || 0,
        parseFloat(formData.value.freight_big) || 0,
      ],
    })

    if (result.success) {
      hasChanges.value = false
    } else {
      error.value = 'Failed to save defaults'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    isProcessing.value = false
  }
}

onMounted(() => {
  fetchDefaults()
})
</script>

<template>
  <div class="defaults-form">
    <div class="header">
      <h2>Billing Defaults</h2>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">
      <i class="fas fa-spinner fa-spin"></i>
      Loading defaults...
    </div>

    <!-- Form -->
    <div v-else class="form-container">
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="rate">Default Rate</label>
          <div class="input-group">
            <input
              id="rate"
              v-model="formData.rate"
              type="number"
              step="0.01"
              min="0"
              required
              placeholder="Enter default rate"
              @input="handleInput"
            />
            <span class="currency">DA</span>
          </div>
        </div>

        <div class="form-group">
          <label for="freight_small">Small Car Freight</label>
          <div class="input-group">
            <input
              id="freight_small"
              v-model="formData.freight_small"
              type="number"
              step="0.01"
              min="0"
              required
              placeholder="Enter small car freight"
              @input="handleInput"
            />
            <span class="currency">USD</span>
          </div>
        </div>

        <div class="form-group">
          <label for="freight_big">Big Car Freight</label>
          <div class="input-group">
            <input
              id="freight_big"
              v-model="formData.freight_big"
              type="number"
              step="0.01"
              min="0"
              required
              placeholder="Enter big car freight"
              @input="handleInput"
            />
            <span class="currency">USD</span>
          </div>
        </div>

        <div class="form-actions">
          <button
            type="button"
            class="cancel-btn"
            :disabled="isProcessing || !hasChanges"
            @click="fetchDefaults"
          >
            Cancel
          </button>
          <button type="submit" class="save-btn" :disabled="isProcessing || !hasChanges">
            <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
            {{ isProcessing ? 'Saving...' : 'Save Changes' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.defaults-form {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
}

.header {
  margin-bottom: 24px;
}

.header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #1f2937;
}

.form-container {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #4b5563;
  font-weight: 500;
}

.input-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-group input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 1rem;
  width: 100%;
}

.input-group input:focus {
  outline: none;
  border-color: #3b82f6;
  ring: 2px solid #3b82f6;
}

.currency {
  color: #6b7280;
  font-weight: 500;
  min-width: 40px;
}

.form-actions {
  margin-top: 32px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.cancel-btn {
  padding: 10px 20px;
  background-color: #f3f4f6;
  color: #4b5563;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.cancel-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.save-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.save-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.error-message {
  padding: 12px;
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  border-radius: 6px;
  color: #b91c1c;
  margin-bottom: 16px;
}

.loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  padding: 24px;
  justify-content: center;
}

@media (max-width: 768px) {
  .defaults-form {
    max-width: 100%;
  }

  .form-container {
    padding: 16px;
  }
}
</style>
