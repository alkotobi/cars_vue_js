<script setup>
import { ref, defineProps, defineEmits, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  selectedCars: {
    type: Array,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])
const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const success = ref(null)
const isProcessing = ref(false)

// Notes data
const newNotes = ref('')

// Add user data
const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

// Add permission check
const can_edit_car_notes = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true // Admin
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_car_notes')
})

const handleNotesChange = async () => {
  if (isProcessing.value || loading.value) return
  if (!newNotes.value.trim()) {
    error.value = t('carNotesBulkEditForm.pleaseEnterNotes')
    return
  }

  if (
    !confirm(
      t('carNotesBulkEditForm.confirmUpdateNotesForCars', { count: props.selectedCars.length }),
    )
  ) {
    return
  }

  loading.value = true
  isProcessing.value = true
  error.value = null
  success.value = null

  try {
    const carIds = props.selectedCars.map((car) => car.id)
    const result = await callApi({
      query:
        'UPDATE cars_stock SET notes = ? WHERE id IN (' + carIds.map(() => '?').join(',') + ')',
      params: [newNotes.value.trim(), ...carIds],
    })

    if (result.success) {
      success.value = t('carNotesBulkEditForm.successfullyUpdatedNotesForCars', {
        count: props.selectedCars.length,
      })

      // Update the cars data
      const updatedCars = props.selectedCars.map((car) => ({
        ...car,
        notes: newNotes.value.trim(),
      }))

      emit('save', updatedCars)
    } else {
      throw new Error(result.error || t('carNotesBulkEditForm.failedToUpdateNotes'))
    }
  } catch (err) {
    error.value = err.message || t('carNotesBulkEditForm.anErrorOccurred')
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

const handleRevertNotes = async () => {
  if (!props.isAdmin) return

  if (
    !confirm(
      t('carNotesBulkEditForm.confirmRevertNotesForCars', { count: props.selectedCars.length }),
    )
  ) {
    return
  }

  loading.value = true
  isProcessing.value = true
  error.value = null
  success.value = null

  try {
    const carIds = props.selectedCars.map((car) => car.id)
    const result = await callApi({
      query:
        'UPDATE cars_stock SET notes = NULL WHERE id IN (' + carIds.map(() => '?').join(',') + ')',
      params: carIds,
    })

    if (result.success) {
      success.value = t('carNotesBulkEditForm.successfullyRevertedNotesForCars', {
        count: props.selectedCars.length,
      })

      // Update the cars data
      const updatedCars = props.selectedCars.map((car) => ({
        ...car,
        notes: null,
      }))

      emit('save', updatedCars)
    } else {
      throw new Error(result.error || t('carNotesBulkEditForm.failedToRevertNotes'))
    }
  } catch (err) {
    error.value = err.message || t('carNotesBulkEditForm.anErrorOccurred')
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = null
  newNotes.value = ''
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content" :class="{ 'is-processing': isProcessing }">
      <div class="modal-header">
        <h3>
          <i class="fas fa-sticky-note"></i>
          {{ t('carNotesBulkEditForm.bulkNotesEdit') }}
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="info-section">
          <h4>
            <i class="fas fa-info-circle"></i>
            {{ t('carNotesBulkEditForm.selectedCars') }}
          </h4>
          <p>
            {{
              t('carNotesBulkEditForm.youHaveSelectedCarsToUpdateNotes', {
                count: selectedCars.length,
              })
            }}
          </p>
        </div>

        <div class="form-group">
          <label for="notes">
            <i class="fas fa-edit"></i>
            {{ t('carNotesBulkEditForm.notes') }}:
          </label>
          <div class="textarea-wrapper">
            <textarea
              id="notes"
              v-model="newNotes"
              :placeholder="t('carNotesBulkEditForm.enterNotesForSelectedCars')"
              class="notes-textarea"
              :disabled="isProcessing"
              rows="4"
            ></textarea>
          </div>
          <button
            class="update-notes-btn"
            @click="handleNotesChange"
            :disabled="isProcessing || !newNotes.trim()"
            :class="{ 'is-processing': isProcessing }"
          >
            <i class="fas fa-save"></i>
            <span>{{
              isProcessing
                ? t('carNotesBulkEditForm.updating')
                : t('carNotesBulkEditForm.updateNotes')
            }}</span>
            <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </div>

        <!-- Revert Button for Admin -->
        <div v-if="isAdmin" class="form-group">
          <button class="revert-btn" @click="handleRevertNotes" :disabled="isProcessing">
            <i class="fas fa-undo"></i>
            {{ t('carNotesBulkEditForm.revertNotesToNull') }}
          </button>
        </div>

        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>
        <div v-if="success" class="success-message">
          <i class="fas fa-check-circle"></i>
          {{ success }}
        </div>
      </div>

      <div class="modal-footer">
        <button class="close-btn-secondary" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
          {{ t('carNotesBulkEditForm.close') }}
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
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-header h3 i {
  color: #f59e0b;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
}

.info-section {
  margin-bottom: 24px;
  padding: 16px;
  background-color: #f9fafb;
  border-radius: 6px;
}

.info-section h4 {
  margin: 0 0 12px 0;
  color: #374151;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-section h4 i {
  color: #3b82f6;
}

.info-section p {
  margin: 0;
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
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.textarea-wrapper {
  margin-bottom: 12px;
}

.notes-textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  min-height: 100px;
  transition: all 0.2s ease;
}

.notes-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.notes-textarea:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.update-notes-btn,
.revert-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  width: 100%;
}

.update-notes-btn {
  background-color: #f59e0b;
  color: white;
}

.revert-btn {
  background-color: #dc2626;
  color: white;
}

.update-notes-btn:hover:not(:disabled) {
  background-color: #d97706;
  transform: translateY(-1px);
}

.revert-btn:hover:not(:disabled) {
  background-color: #b91c1c;
  transform: translateY(-1px);
}

.update-notes-btn:disabled,
.revert-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.is-processing {
  position: relative;
}

.is-processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

.error-message,
.success-message {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 16px 0;
  padding: 12px;
  border-radius: 4px;
  font-size: 14px;
}

.error-message {
  background-color: #fee2e2;
  color: #ef4444;
}

.success-message {
  background-color: #d1fae5;
  color: #10b981;
}

.close-btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn-secondary:hover:not(:disabled) {
  background-color: #e5e7eb;
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

.modal-content.is-processing {
  opacity: 0.7;
  pointer-events: none;
}
</style>
