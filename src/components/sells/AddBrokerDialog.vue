<script setup>
import { ref, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useI18n } from 'vue-i18n'
import AddItemDialog from '../car-stock/AddItemDialog.vue'

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'saved'])

const { callApi } = useApi()
const { t } = useI18n()

const newBroker = ref({
  name: '',
  address: '',
  email: '',
  mobiles: '',
  notes: '',
})

const isSubmitting = ref(false)
const validationError = ref('')

// Computed properties for validation status
const isNewBrokerMobileValid = computed(() => {
  return newBroker.value.mobiles && validateAlgerianMobile(newBroker.value.mobiles)
})

const isNewBrokerEmailValid = computed(() => {
  return newBroker.value.email && validateEmail(newBroker.value.email)
})

// Validation functions
const validateEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const validateAlgerianMobile = (mobile) => {
  // Algerian mobile numbers: 05XX, 06XX, 07XX followed by 8 digits
  const mobileRegex = /^(05|06|07)[0-9]{8}$/
  return mobileRegex.test(mobile)
}

const handleMobileInput = (event) => {
  const value = event.target.value.replace(/\D/g, '') // Keep only digits
  newBroker.value.mobiles = value
}

const handleClose = () => {
  newBroker.value = {
    name: '',
    address: '',
    email: '',
    mobiles: '',
    notes: '',
  }
  validationError.value = ''
  emit('close')
}

const addBroker = async () => {
  if (isSubmitting.value) return

  validationError.value = ''

  // Validation
  if (!newBroker.value.name || newBroker.value.name.trim() === '') {
    validationError.value = t('brokers.nameRequired') || 'Name is required'
    return
  }

  if (!newBroker.value.mobiles) {
    validationError.value = t('brokers.mobileRequired') || 'Mobile number is required'
    return
  }

  if (!validateAlgerianMobile(newBroker.value.mobiles)) {
    validationError.value = t('brokers.invalidMobile') || 'Please enter a valid Algerian mobile number (05XX, 06XX, or 07XX followed by 8 digits)'
    return
  }

  if (!newBroker.value.email) {
    validationError.value = t('brokers.emailRequired') || 'Email is required'
    return
  }

  if (!validateEmail(newBroker.value.email)) {
    validationError.value = t('brokers.invalidEmail') || 'Please enter a valid email address'
    return
  }

  isSubmitting.value = true
  try {
    const result = await callApi({
      query: `
        INSERT INTO clients (name, address, email, mobiles, is_broker, is_client, notes)
        VALUES (?, ?, ?, ?, 1, 0, ?)
      `,
      params: [
        newBroker.value.name.trim(),
        newBroker.value.address || null,
        newBroker.value.email.trim(),
        newBroker.value.mobiles,
        newBroker.value.notes || null,
      ],
    })

    if (!result.success) {
      validationError.value = result.error || t('brokers.failedToAddBroker') || 'Failed to add broker'
      return
    }

    emit('saved', { id: result.lastInsertId, name: newBroker.value.name.trim() })
    handleClose()
  } catch (err) {
    validationError.value = err.message || t('brokers.failedToAddBroker') || 'Failed to add broker'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <AddItemDialog
    :show="show"
    :title="t('brokers.add_new_broker') || 'Add New Broker'"
    :loading="isSubmitting || loading"
    @close="handleClose"
    @save="addBroker"
  >
    <div v-if="validationError" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ validationError }}
    </div>

    <div class="form-grid">
      <div class="form-group">
        <label for="broker_name">
          {{ t('brokers.name') || 'Name' }} <span class="required">*</span>:
        </label>
        <input
          id="broker_name"
          v-model="newBroker.name"
          type="text"
          :placeholder="t('brokers.enterName') || 'Enter broker name'"
          required
          :disabled="isSubmitting || loading"
        />
      </div>

      <div class="form-group">
        <label for="broker_email">
          {{ t('brokers.email') || 'Email' }} <span class="required">*</span>:
        </label>
        <input
          id="broker_email"
          v-model="newBroker.email"
          type="email"
          :placeholder="t('brokers.enterEmail') || 'Enter email address'"
          required
          :class="{ 'success-input': isNewBrokerEmailValid }"
          :disabled="isSubmitting || loading"
        />
        <div v-if="isNewBrokerEmailValid" class="field-success">
          <i class="fas fa-check-circle"></i>
          <span>{{ t('brokers.validEmail') || 'Valid email format' }}</span>
        </div>
      </div>

      <div class="form-group">
        <label for="broker_mobiles">
          {{ t('brokers.mobileNumbers') || 'Mobile Numbers' }} <span class="required">*</span>:
        </label>
        <input
          id="broker_mobiles"
          v-model="newBroker.mobiles"
          type="text"
          :placeholder="t('brokers.enterMobileNumbers') || 'Enter mobile number (e.g., 0550123456)'"
          required
          maxlength="10"
          @input="handleMobileInput"
          :class="{ 'success-input': isNewBrokerMobileValid }"
          :disabled="isSubmitting || loading"
        />
        <div v-if="isNewBrokerMobileValid" class="field-success">
          <i class="fas fa-check-circle"></i>
          <span>{{ t('brokers.validMobile') || 'Valid Algerian mobile number' }}</span>
        </div>
        <div v-else-if="newBroker.mobiles && !isNewBrokerMobileValid" class="field-help">
          <i class="fas fa-info-circle"></i>
          <span>{{ t('brokers.mobileFormat') || 'Format: 05XX, 06XX, or 07XX followed by 8 digits' }}</span>
        </div>
      </div>

      <div class="form-group">
        <label for="broker_address">
          {{ t('brokers.address') || 'Address' }}:
        </label>
        <input
          id="broker_address"
          v-model="newBroker.address"
          type="text"
          :placeholder="t('brokers.enterAddress') || 'Enter address'"
          :disabled="isSubmitting || loading"
        />
      </div>

      <div class="form-group full-width">
        <label for="broker_notes">
          {{ t('brokers.notes') || 'Notes' }}:
        </label>
        <textarea
          id="broker_notes"
          v-model="newBroker.notes"
          :placeholder="t('brokers.enterNotes') || 'Enter notes'"
          rows="3"
          :disabled="isSubmitting || loading"
        ></textarea>
      </div>
    </div>
  </AddItemDialog>
</template>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group.full-width {
  grid-column: span 2;
}

.form-group label {
  font-weight: 500;
  color: #374151;
  font-size: 14px;
}

.form-group label .required {
  color: #ef4444;
}

.form-group input,
.form-group textarea {
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

.form-group input:disabled,
.form-group textarea:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.error-message i {
  font-size: 16px;
}

.success-input {
  border-color: #10b981 !important;
  background-color: #f0fdf4 !important;
}

.field-success {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
  padding: 6px 10px;
  background: linear-gradient(135deg, #d1fae5 0%, #ecfdf5 100%);
  border: 1px solid #10b981;
  border-radius: 4px;
  color: #065f46;
  font-size: 12px;
  font-weight: 500;
}

.field-success i {
  color: #10b981;
}

.field-help {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
  padding: 6px 10px;
  background: linear-gradient(135deg, #fef3c7 0%, #fef7cd 100%);
  border: 1px solid #f59e0b;
  border-radius: 4px;
  color: #92400e;
  font-size: 12px;
  font-weight: 500;
}

.field-help i {
  color: #f59e0b;
}

@media (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
  }

  .form-group.full-width {
    grid-column: span 1;
  }
}
</style>
