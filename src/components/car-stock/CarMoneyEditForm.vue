<script setup>
import { ref, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

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
const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)

// Form fields
const price = ref(props.car.price_cell || '')
const freight = ref(props.car.freight || '')
const rate = ref(props.car.rate || '')

const handleSubmit = async () => {
  if (isProcessing.value || loading.value) return
  // Validate inputs
  if (price.value && isNaN(parseFloat(price.value))) {
    error.value = t('carMoneyEditForm.priceMustBeValidNumber')
    return
  }
  if (freight.value && isNaN(parseFloat(freight.value))) {
    error.value = t('carMoneyEditForm.freightMustBeValidNumber')
    return
  }
  if (rate.value && isNaN(parseFloat(rate.value))) {
    error.value = t('carMoneyEditForm.rateMustBeValidNumber')
    return
  }

  loading.value = true
  isProcessing.value = true
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
      error.value = t('carMoneyEditForm.pleaseFillAtLeastOneField')
      loading.value = false
      return
    }

    params.push(props.car.id)

    const result = await callApi({
      query: `UPDATE cars_stock SET ${updates.join(', ')} WHERE id = ?`,
      params,
    })

    if (result.success) {
      const updatedCar = {
        ...props.car,
        price_cell: price.value !== '' ? parseFloat(price.value) : props.car.price_cell,
        freight: freight.value !== '' ? parseFloat(freight.value) : props.car.freight,
        rate: rate.value !== '' ? parseFloat(rate.value) : props.car.rate,
      }
      emit('save', updatedCar)
      emit('close')
    } else {
      throw new Error(result.error || t('carMoneyEditForm.failedToUpdateMoneyFields'))
    }
  } catch (err) {
    error.value = err.message || t('carMoneyEditForm.anErrorOccurred')
  } finally {
    loading.value = false
    isProcessing.value = false
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
    <div class="modal-content" :class="{ 'is-processing': isProcessing }">
      <div class="modal-header">
        <h3>
          <i class="fas fa-dollar-sign"></i>
          {{ t('carMoneyEditForm.editMoneyFields') }}
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="price">
            <i class="fas fa-tag"></i>
            {{ t('carMoneyEditForm.price') }}:
          </label>
          <div class="input-group">
            <span class="currency-symbol">
              <i class="fas fa-dollar-sign"></i>
            </span>
            <input
              type="number"
              id="price"
              v-model="price"
              :placeholder="t('carMoneyEditForm.enterPrice')"
              step="0.01"
              min="0"
              class="input-field"
              :disabled="isProcessing"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="freight">
            <i class="fas fa-ship"></i>
            {{ t('carMoneyEditForm.freight') }}:
          </label>
          <div class="input-group">
            <span class="currency-symbol">
              <i class="fas fa-dollar-sign"></i>
            </span>
            <input
              type="number"
              id="freight"
              v-model="freight"
              :placeholder="t('carMoneyEditForm.enterFreight')"
              step="0.01"
              min="0"
              class="input-field"
              :disabled="isProcessing"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="rate">
            <i class="fas fa-percentage"></i>
            {{ t('carMoneyEditForm.rate') }}:
          </label>
          <div class="input-group">
            <input
              type="number"
              id="rate"
              v-model="rate"
              :placeholder="t('carMoneyEditForm.enterRate')"
              step="0.01"
              min="0"
              class="input-field"
              :disabled="isProcessing"
            />
            <span class="rate-symbol">
              <i class="fas fa-percent"></i>
            </span>
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
          {{ t('carMoneyEditForm.cancel') }}
        </button>
        <button
          class="save-btn"
          @click="handleSubmit"
          :disabled="isProcessing"
          :class="{ 'is-processing': isProcessing }"
        >
          <i class="fas fa-save"></i>
          <span>{{
            isProcessing ? t('carMoneyEditForm.saving') : t('carMoneyEditForm.saveChanges')
          }}</span>
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
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.input-group {
  display: flex;
  align-items: center;
  width: 100%;
  position: relative;
}

.currency-symbol,
.rate-symbol {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 38px;
  background-color: #f3f4f6;
  border: 1px solid #d1d5db;
  color: #6b7280;
}

.currency-symbol {
  border-right: none;
  border-top-left-radius: 4px;
  border-bottom-left-radius: 4px;
}

.rate-symbol {
  border-left: none;
  border-top-right-radius: 4px;
  border-bottom-right-radius: 4px;
}

.input-field {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  font-size: 14px;
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

.input-field[type='number'] {
  -moz-appearance: textfield;
}

.input-field[type='number']::-webkit-outer-spin-button,
.input-field[type='number']::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ef4444;
  margin: 16px 0;
  padding: 12px;
  background-color: #fee2e2;
  border-radius: 4px;
  font-size: 14px;
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

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
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

.modal-content.is-processing {
  opacity: 0.7;
  pointer-events: none;
}

.modal-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-header h3 i {
  color: #3b82f6;
}
</style>
