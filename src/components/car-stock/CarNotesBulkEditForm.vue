<script setup>
import { ref, watch } from 'vue'
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

// Notes data - new note to add to all cars
const newNote = ref('')

// Store existing notes for each car (to append new notes to)
const carsNotesMap = ref({})

// Add user data
const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

// Parse notes for a car
const parseCarNotes = (notes) => {
  if (!notes) return []
  try {
    if (typeof notes === 'string' && notes.trim().startsWith('[')) {
      return JSON.parse(notes)
    } else if (Array.isArray(notes)) {
      return notes
    } else {
      // Old format (plain text) - convert to JSON
      return [{
        id_user: user.value?.id || null,
        note: notes,
        timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
      }]
    }
  } catch (e) {
    // If parsing fails, treat as old format
    return [{
      id_user: user.value?.id || null,
      note: notes,
      timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
    }]
  }
}

// Load notes for all selected cars
const loadCarsNotes = () => {
  carsNotesMap.value = {}
  props.selectedCars.forEach((car) => {
    carsNotesMap.value[car.id] = parseCarNotes(car.notes)
  })
}

// Watch for modal opening to load notes
watch(() => props.show, (isOpen) => {
  if (isOpen) {
    loadCarsNotes()
  }
})

const handleNotesChange = async () => {
  if (isProcessing.value || loading.value) return
  if (!newNote.value.trim()) {
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
    const userId = user.value?.id || null
    const timestamp = new Date().toISOString().slice(0, 19).replace('T', ' ')
    const newNoteEntry = {
      id_user: userId,
      note: newNote.value.trim(),
      timestamp: timestamp
    }

    // Update each car's notes by appending the new note
    const updatePromises = carIds.map(async (carId) => {
      const existingNotes = carsNotesMap.value[carId] || []
      const updatedNotes = [...existingNotes, newNoteEntry]
      const notesJson = JSON.stringify(updatedNotes)

      return callApi({
        query: 'UPDATE cars_stock SET notes = ? WHERE id = ?',
        params: [notesJson, carId],
      })
    })

    const results = await Promise.all(updatePromises)
    const allSuccess = results.every((result) => result.success)

    if (allSuccess) {
      success.value = t('carNotesBulkEditForm.successfullyUpdatedNotesForCars', {
        count: props.selectedCars.length,
      })

      // Update the cars data
      const updatedCars = props.selectedCars.map((car) => {
        const existingNotes = carsNotesMap.value[car.id] || []
        const updatedNotes = [...existingNotes, newNoteEntry]
        return {
          ...car,
          notes: JSON.stringify(updatedNotes),
        }
      })

      emit('save', updatedCars)
      newNote.value = '' // Clear the input
      loadCarsNotes() // Reload notes for next addition
    } else {
      const failedResult = results.find((r) => !r.success)
      throw new Error(failedResult?.error || t('carNotesBulkEditForm.failedToUpdateNotes'))
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
  newNote.value = ''
  carsNotesMap.value = {}
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
            {{ t('carNotesBulkEditForm.addNewNote') || 'Add New Note' }}:
          </label>
          <div class="textarea-wrapper">
            <textarea
              id="notes"
              v-model="newNote"
              :placeholder="t('carNotesBulkEditForm.enterNotesForSelectedCars')"
              class="notes-textarea"
              :disabled="isProcessing"
              rows="4"
            ></textarea>
          </div>
          <small class="help-text">
            {{ t('carNotesBulkEditForm.noteWillBeAddedToAllCars') || 'This note will be added to all selected cars' }}
          </small>
          <button
            class="update-notes-btn"
            @click="handleNotesChange"
            :disabled="isProcessing || !newNote.trim()"
            :class="{ 'is-processing': isProcessing }"
          >
            <i class="fas fa-save"></i>
            <span>{{
              isProcessing
                ? t('carNotesBulkEditForm.updating')
                : t('carNotesBulkEditForm.addNoteToAllCars') || 'Add Note to All Cars'
            }}</span>
            <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
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
  z-index: 2000;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
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

.update-notes-btn {
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
  background-color: #f59e0b;
  color: white;
}

.update-notes-btn:hover:not(:disabled) {
  background-color: #d97706;
  transform: translateY(-1px);
}

.update-notes-btn:disabled {
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

.help-text {
  display: block;
  margin-top: 8px;
  margin-bottom: 12px;
  color: #6b7280;
  font-size: 0.875rem;
  font-style: italic;
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
