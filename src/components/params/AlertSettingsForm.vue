<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)
const hasChanges = ref(false)

// Form data - alert settings
const formData = ref({
  alert_unloaded_after_days: '',
  alert_not_arrived_after_days: '',
  alert_no_licence_after_days: '',
  alert_no_docs_sent_after_days: '',
  max_unpayed_sell_bills: '',
})

const fetchAlertSettings = async () => {
  loading.value = true
  error.value = null
  try {
    const result = await callApi({
      query: 'SELECT * FROM defaults LIMIT 1',
    })
    if (result.success && result.data.length > 0) {
      const defaults = result.data[0]
      formData.value = {
        alert_unloaded_after_days: defaults.alert_unloaded_after_days?.toString() || '30',
        alert_not_arrived_after_days: defaults.alert_not_arrived_after_days?.toString() || '60',
        alert_no_licence_after_days: defaults.alert_no_licence_after_days?.toString() || '90',
        alert_no_docs_sent_after_days: defaults.alert_no_docs_sent_after_days?.toString() || '45',
        max_unpayed_sell_bills: defaults.max_unpayed_sell_bills?.toString() || '3',
      }
    } else {
      // If no record exists, create one with default values
      await callApi({
        query: `
          INSERT INTO defaults (rate, freight_small, freight_big, alert_unloaded_after_days, alert_not_arrived_after_days, alert_no_licence_after_days, alert_no_docs_sent_after_days, max_unpayed_sell_bills)
          VALUES (0, 0, 0, 30, 60, 90, 45, 3)
        `,
      })
      formData.value = {
        alert_unloaded_after_days: '30',
        alert_not_arrived_after_days: '60',
        alert_no_licence_after_days: '90',
        alert_no_docs_sent_after_days: '45',
        max_unpayed_sell_bills: '3',
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
        SET alert_unloaded_after_days = ?, alert_not_arrived_after_days = ?, 
            alert_no_licence_after_days = ?, alert_no_docs_sent_after_days = ?, 
            max_unpayed_sell_bills = ?
        WHERE id = 1
      `,
      params: [
        parseInt(formData.value.alert_unloaded_after_days) || 30,
        parseInt(formData.value.alert_not_arrived_after_days) || 60,
        parseInt(formData.value.alert_no_licence_after_days) || 90,
        parseInt(formData.value.alert_no_docs_sent_after_days) || 45,
        parseInt(formData.value.max_unpayed_sell_bills) || 3,
      ],
    })

    if (result.success) {
      hasChanges.value = false
    } else {
      error.value = 'Failed to save alert settings'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    isProcessing.value = false
  }
}

onMounted(() => {
  fetchAlertSettings()
})
</script>

<template>
  <div class="alert-settings-form">
    <div class="header">
      <h2>Alert Settings</h2>
      <p class="description">Configure alert thresholds and limits for the system</p>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">
      <i class="fas fa-spinner fa-spin"></i>
      Loading alert settings...
    </div>

    <!-- Form -->
    <div v-else class="form-container">
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="alert_unloaded_after_days">Alert Unloaded After (Days)</label>
          <div class="input-group">
            <input
              id="alert_unloaded_after_days"
              v-model="formData.alert_unloaded_after_days"
              type="number"
              min="1"
              required
              placeholder="Enter days"
              @input="handleInput"
            />
            <span class="unit">days</span>
          </div>
          <small class="help-text">Number of days after which to alert about unloaded cars</small>
        </div>

        <div class="form-group">
          <label for="alert_not_arrived_after_days">Alert Not Arrived After (Days)</label>
          <div class="input-group">
            <input
              id="alert_not_arrived_after_days"
              v-model="formData.alert_not_arrived_after_days"
              type="number"
              min="1"
              required
              placeholder="Enter days"
              @input="handleInput"
            />
            <span class="unit">days</span>
          </div>
          <small class="help-text"
            >Number of days after which to alert about cars that haven't arrived</small
          >
        </div>

        <div class="form-group">
          <label for="alert_no_licence_after_days">Alert No License After (Days)</label>
          <div class="input-group">
            <input
              id="alert_no_licence_after_days"
              v-model="formData.alert_no_licence_after_days"
              type="number"
              min="1"
              required
              placeholder="Enter days"
              @input="handleInput"
            />
            <span class="unit">days</span>
          </div>
          <small class="help-text"
            >Number of days after which to alert about cars without license</small
          >
        </div>

        <div class="form-group">
          <label for="alert_no_docs_sent_after_days">Alert No Docs Sent After (Days)</label>
          <div class="input-group">
            <input
              id="alert_no_docs_sent_after_days"
              v-model="formData.alert_no_docs_sent_after_days"
              type="number"
              min="1"
              required
              placeholder="Enter days"
              @input="handleInput"
            />
            <span class="unit">days</span>
          </div>
          <small class="help-text"
            >Number of days after which to alert about cars with no documents sent</small
          >
        </div>

        <div class="form-group">
          <label for="max_unpayed_sell_bills">Max Unpaid Sell Bills</label>
          <div class="input-group">
            <input
              id="max_unpayed_sell_bills"
              v-model="formData.max_unpayed_sell_bills"
              type="number"
              min="1"
              required
              placeholder="Enter max unpaid bills"
              @input="handleInput"
            />
            <span class="unit">bills</span>
          </div>
          <small class="help-text"
            >Maximum number of unpaid sell bills allowed before preventing new bill creation</small
          >
        </div>

        <div class="form-actions">
          <button
            type="button"
            class="cancel-btn"
            :disabled="isProcessing || !hasChanges"
            @click="fetchAlertSettings"
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
.alert-settings-form {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
}

.header {
  margin-bottom: 24px;
}

.header h2 {
  margin: 0 0 8px 0;
  font-size: 1.5rem;
  color: #1f2937;
}

.description {
  margin: 0;
  color: #6b7280;
  font-size: 0.9rem;
}

.form-container {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 24px;
}

.form-group {
  margin-bottom: 24px;
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

.unit {
  color: #6b7280;
  font-weight: 500;
  min-width: 40px;
}

.help-text {
  display: block;
  margin-top: 4px;
  color: #6b7280;
  font-size: 0.8rem;
  line-height: 1.4;
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
  .alert-settings-form {
    max-width: 100%;
  }

  .form-container {
    padding: 16px;
  }
}
</style>
