<script setup>
import { ref,   } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  show: Boolean,
  selectedCars: Array,
  isAdmin: Boolean,
})
const emit = defineEmits(['close', 'save'])
const { callApi } = useApi()

const exportLicense = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

const handleSave = async () => {
  if (!exportLicense.value) {
    error.value = t('carExportLicenseBulkEditForm.pleaseEnterExportLicenseValue')
    return
  }
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    const carIds = props.selectedCars.map((car) => car.id)
    const result = await callApi({
      query:
        'UPDATE cars_stock SET export_lisence_ref = ? WHERE id IN (' +
        carIds.map(() => '?').join(',') +
        ')',
      params: [exportLicense.value, ...carIds],
    })
    if (result.success) {
      success.value = t('carExportLicenseBulkEditForm.exportLicenseUpdatedForCars', {
        count: carIds.length,
      })
      emit(
        'save',
        props.selectedCars.map((car) => ({ ...car, export_lisence_ref: exportLicense.value })),
      )
      emit('close')
    } else {
      throw new Error(result.error || t('carExportLicenseBulkEditForm.failedToUpdateExportLicense'))
    }
  } catch (err) {
    error.value = err.message || t('carExportLicenseBulkEditForm.anErrorOccurred')
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  exportLicense.value = ''
  error.value = ''
  success.value = ''
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>
          <i class="fas fa-file-signature"></i>
          {{ t('carExportLicenseBulkEditForm.bulkExportLicenseEdit') }}
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="loading">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="info-section">
          <h4>
            <i class="fas fa-info-circle"></i> {{ t('carExportLicenseBulkEditForm.selectedCars') }}
          </h4>
          <p>
            {{
              t('carExportLicenseBulkEditForm.youHaveSelectedCarsToUpdateExportLicense', {
                count: selectedCars.length,
              })
            }}
          </p>
        </div>
        <div class="form-group">
          <label for="export-license"
            ><i class="fas fa-id-card"></i>
            {{ t('carExportLicenseBulkEditForm.exportLicense') }}</label
          >
          <input
            id="export-license"
            v-model="exportLicense"
            class="input-field"
            type="text"
            :placeholder="t('carExportLicenseBulkEditForm.enterExportLicenseValue')"
            :disabled="loading"
            @keyup.enter="handleSave"
          />
        </div>
        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i> {{ error }}
        </div>
        <div v-if="success" class="success-message">
          <i class="fas fa-check-circle"></i> {{ success }}
        </div>
        <button class="save-btn" @click="handleSave" :disabled="loading">
          <i class="fas fa-save"></i>
          {{
            loading
              ? t('carExportLicenseBulkEditForm.saving')
              : t('carExportLicenseBulkEditForm.save')
          }}
        </button>
        <button class="close-btn-secondary" @click="closeModal" :disabled="loading">
          <i class="fas fa-times"></i> {{ t('carExportLicenseBulkEditForm.close') }}
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
  max-width: 400px;
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
.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background-color: #f3f4f6;
  color: #6b7280;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 20px;
  box-shadow: none;
  outline: none;
}
.close-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
  color: #374151;
  transform: scale(1.05);
}
.close-btn:focus {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}
.close-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.info-section {
  margin: 24px 0;
  padding: 16px;
  background-color: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #3b82f6;
}
.info-section h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px;
  color: #1f2937;
}
.info-section h4 i {
  color: #3b82f6;
}
.info-section p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}
.form-group {
  margin-bottom: 20px;
}
.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
}
.input-field {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 15px;
  background-color: white;
  transition: all 0.2s ease;
}
.input-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}
.input-field:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}
.save-btn {
  width: 100%;
  padding: 12px 0;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  margin-bottom: 12px;
  transition: all 0.2s ease;
}
.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
}
.save-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
.close-btn-secondary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 12px 0;
  background-color: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
.close-btn-secondary:hover:not(:disabled) {
  background-color: #e5e7eb;
  color: #1f2937;
}
.close-btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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
</style>
