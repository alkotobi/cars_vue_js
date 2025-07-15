<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)
const hasChanges = ref(false)

// Form data - only alert settings
const formData = ref({
  alert_unloaded_after_days: '',
  alert_not_arrived_after_days: '',
  alert_no_licence_after_days: '',
  alert_no_docs_sent_after_days: '',
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
        alert_unloaded_after_days: defaults.alert_unloaded_after_days?.toString() || '',
        alert_not_arrived_after_days: defaults.alert_not_arrived_after_days?.toString() || '',
        alert_no_licence_after_days: defaults.alert_no_licence_after_days?.toString() || '',
        alert_no_docs_sent_after_days: defaults.alert_no_docs_sent_after_days?.toString() || '',
      }
    } else {
      // If no record exists, create one with default values
      await callApi({
        query: `
          INSERT INTO defaults (alert_unloaded_after_days, alert_not_arrived_after_days, alert_no_licence_after_days, alert_no_docs_sent_after_days)
          VALUES (30, 30, 30, 30)
        `,
      })
      formData.value = {
        alert_unloaded_after_days: '30',
        alert_not_arrived_after_days: '30',
        alert_no_licence_after_days: '30',
        alert_no_docs_sent_after_days: '30',
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
            alert_no_licence_after_days = ?, alert_no_docs_sent_after_days = ?
        WHERE id = 1
      `,
      params: [
        parseInt(formData.value.alert_unloaded_after_days) || 30,
        parseInt(formData.value.alert_not_arrived_after_days) || 30,
        parseInt(formData.value.alert_no_licence_after_days) || 30,
        parseInt(formData.value.alert_no_docs_sent_after_days) || 30,
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
  fetchDefaults()
})
</script>

<template>
  <div class="alert-settings-form">
    <div class="header">
      <h2>Alert Settings</h2>
      <p class="description">Configure when alerts should be triggered for overdue cars</p>
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
          <label for="alert_unloaded_after_days">Not Loaded Cars Alert (Days)</label>
          <div class="input-group">
            <input
              id="alert_unloaded_after_days"
              v-model="formData.alert_unloaded_after_days"
              type="number"
              min="1"
              max="365"
              required
              placeholder="Enter days"
              @input="handleInput"
            />
            <span class="unit">days after sell date</span>
          </div>
          <small class="help-text">Alert for cars sold but not loaded after this many days</small>
        </div>

        <div class="form-group">
          <label for="alert_not_arrived_after_days">Not Arrived Cars Alert (Days)</label>
          <div class="input-group">
            <input
              id="alert_not_arrived_after_days"
              v-model="formData.alert_not_arrived_after_days"
              type="number"
              min="1"
              max="365"
              required
              placeholder="Enter days"
              @input="handleInput"
            />
            <span class="unit">days after buy date</span>
          </div>
          <small class="help-text"
            >Alert for cars purchased but not arrived at warehouse after this many days</small
          >
        </div>

        <div class="form-group">
          <label for="alert_no_licence_after_days">No License Cars Alert (Days)</label>
          <div class="input-group">
            <input
              id="alert_no_licence_after_days"
              v-model="formData.alert_no_licence_after_days"
              type="number"
              min="1"
              max="365"
              required
              placeholder="Enter days"
              @input="handleInput"
            />
            <span class="unit">days after buy date</span>
          </div>
          <small class="help-text"
            >Alert for cars purchased but without export license after this many days</small
          >
        </div>

        <div class="form-group">
          <label for="alert_no_docs_sent_after_days">No Docs Sent Alert (Days)</label>
          <div class="input-group">
            <input
              id="alert_no_docs_sent_after_days"
              v-model="formData.alert_no_docs_sent_after_days"
              type="number"
              min="1"
              max="365"
              required
              placeholder="Enter days"
              @input="handleInput"
            />
            <span class="unit">days after sell date</span>
          </div>
          <small class="help-text"
            >Alert for cars sold but documents not sent after this many days</small
          >
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

.unit {
  color: #6b7280;
  font-size: 0.875rem;
  min-width: 120px;
}

.help-text {
  display: block;
  margin-top: 4px;
  color: #6b7280;
  font-size: 0.75rem;
  font-style: italic;
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

  .input-group {
    flex-direction: column;
    align-items: stretch;
  }

  .unit {
    min-width: auto;
    text-align: center;
  }
}
</style>
