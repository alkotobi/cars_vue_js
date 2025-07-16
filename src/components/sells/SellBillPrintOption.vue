<script setup>
import { ref, defineProps, defineEmits, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'

const router = useRouter()
const { callApi } = useApi()
const { t } = useEnhancedI18n()

const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  billId: {
    type: Number,
    default: null,
  },
  carId: {
    type: Number,
    default: null,
  },
  documentType: {
    type: String,
    default: 'sell-bill', // 'sell-bill' or 'car'
  },
})

const emit = defineEmits(['close', 'proceed'])

const formData = ref({
  documentType: 'contract',
  paymentTerms: 'FOB',
  paymentMode: 'TT',
  currency: 'USD',
  bankId: null,
})

const banks = ref([])
const isProcessing = ref(false)
const isLoadingBanks = ref(false)

// Add document type options
const documentTypes = [
  { value: 'contract', label: 'Contract' },
  { value: 'invoice', label: 'Invoice' },
  { value: 'proforma', label: 'Proforma' },
]

// Add payment terms options
const paymentTermsOptions = [
  { value: 'FOB', label: 'FOB' },
  { value: 'CFR', label: 'CFR' },
]

// Add payment type options
const paymentTypes = [
  { value: 'LC', label: 'LC' },
  { value: 'TT', label: 'TT' },
]

// Add currency options
const currencyOptions = [
  { value: 'USD', label: 'USD' },
  { value: 'DA', label: 'DA' },
]

const fetchBanks = async () => {
  try {
    isLoadingBanks.value = true
    const result = await callApi({
      query: 'SELECT * FROM banks ORDER BY company_name ASC',
      params: [],
    })
    if (result.success && result.data.length > 0) {
      banks.value = result.data
      // Set the first bank as default if not already set
      if (!formData.value.bankId) {
        formData.value.bankId = result.data[0].id
      }
    }
  } catch (err) {
    console.error('Error fetching banks:', err)
  } finally {
    isLoadingBanks.value = false
  }
}

// Watch for dialog visibility
watch(
  () => props.visible,
  (newValue) => {
    if (newValue) {
      fetchBanks()
    }
  },
)

onMounted(() => {
  if (props.visible) {
    fetchBanks()
  }
})

const handleProceed = () => {
  let printUrl

  if (props.documentType === 'car') {
    // Open car print page in new tab
    printUrl = router.resolve({
      name: 'print-car',
      params: { carId: props.carId },
      query: {
        billId: props.billId,
        options: encodeURIComponent(JSON.stringify(formData.value)),
      },
    }).href
  } else {
    // Open sell bill print page in new tab
    printUrl = router.resolve({
      name: 'print',
      params: { billId: props.billId },
      query: { options: encodeURIComponent(JSON.stringify(formData.value)) },
    }).href
  }

  window.open(printUrl, '_blank')

  emit('proceed', {
    billId: props.billId,
    carId: props.carId,
    documentType: props.documentType,
    ...formData.value,
  })
}

const handleCancel = () => {
  emit('close')
}
</script>

<template>
  <div v-if="visible" class="print-options-dialog">
    <div class="dialog-content">
      <div class="dialog-header">
        <h3><i class="fas fa-print"></i> {{ t('sellBills.print_options') }}</h3>
        <p class="dialog-subtitle">
          {{ t('sellBills.configure_document_settings') }}
        </p>
      </div>

      <div class="form-container">
        <div class="form-section">
          <h4><i class="fas fa-file-alt"></i> {{ t('sellBills.document_settings') }}</h4>
          <div class="form-group">
            <label>{{ t('sellBills.document_type') }}:</label>
            <select v-model="formData.documentType" class="form-select">
              <option v-for="type in documentTypes" :key="type.value" :value="type.value">
                {{ type.label }}
              </option>
            </select>
          </div>
        </div>

        <div class="form-section">
          <h4><i class="fas fa-credit-card"></i> {{ t('sellBills.payment_settings') }}</h4>
          <div class="form-row">
            <div class="form-group">
              <label>{{ t('sellBills.payment_terms') }}:</label>
              <select v-model="formData.paymentTerms" class="form-select">
                <option v-for="term in paymentTermsOptions" :key="term.value" :value="term.value">
                  {{ term.label }}
                </option>
              </select>
            </div>

            <div class="form-group">
              <label>{{ t('sellBills.payment_mode') }}:</label>
              <select v-model="formData.paymentMode" class="form-select">
                <option v-for="type in paymentTypes" :key="type.value" :value="type.value">
                  {{ type.label }}
                </option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label>{{ t('sellBills.currency') }}:</label>
            <select v-model="formData.currency" class="form-select">
              <option
                v-for="currency in currencyOptions"
                :key="currency.value"
                :value="currency.value"
              >
                {{ currency.label }}
              </option>
            </select>
          </div>
        </div>

        <div class="form-section">
          <h4><i class="fas fa-university"></i> {{ t('sellBills.bank_information') }}</h4>
          <div class="form-group">
            <label>{{ t('sellBills.select_bank') }}:</label>
            <select v-model="formData.bankId" :disabled="isLoadingBanks" class="form-select">
              <option v-if="isLoadingBanks" value="">
                <i class="fas fa-spinner fa-spin"></i> {{ t('sellBills.loading_banks') }}
              </option>
              <option v-else value="">{{ t('sellBills.select_bank') }}</option>
              <option v-for="bank in banks" :key="bank.id" :value="bank.id">
                {{ bank.company_name }}
              </option>
            </select>
          </div>
        </div>
      </div>

      <div class="dialog-buttons">
        <button @click="handleCancel" class="cancel-btn">
          <i class="fas fa-times"></i> {{ t('sellBills.cancel') }}
        </button>
        <button @click="handleProceed" class="proceed-btn">
          <i class="fas fa-print"></i> {{ t('sellBills.print_document') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Professional Print Options Dialog */
.print-options-dialog {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.dialog-content {
  background: linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
  border-radius: 12px;
  padding: 32px;
  width: 100%;
  max-width: 600px;
  max-height: calc(100vh - 40px);
  overflow-y: auto;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  border: 1px solid #e2e8f0;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.dialog-header {
  margin-bottom: 32px;
  text-align: center;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 20px;
}

h3 {
  margin: 0 0 8px 0;
  color: #1e293b;
  font-size: 1.5rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

h3 i {
  color: #3498db;
  font-size: 1.25rem;
}

.dialog-subtitle {
  color: #64748b;
  font-size: 0.95rem;
  margin: 0;
  font-weight: 400;
}

.form-container {
  margin-bottom: 32px;
}

.form-section {
  margin-bottom: 24px;
  background: #ffffff;
  border-radius: 8px;
  padding: 20px;
  border: 1px solid #e2e8f0;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.form-section h4 {
  margin: 0 0 16px 0;
  color: #1e293b;
  font-weight: 600;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e2e8f0;
}

.form-section h4 i {
  color: #3498db;
  font-size: 1rem;
}

.form-group {
  margin-bottom: 20px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 600;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.form-select {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 1rem;
  background: #ffffff;
  color: #1e293b;
  transition: all 0.2s ease;
  font-weight: 500;
}

.form-select:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
  background: #ffffff;
}

.form-select:hover {
  border-color: #cbd5e1;
}

.form-select:disabled {
  background: #f1f5f9;
  color: #64748b;
  cursor: not-allowed;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

@media (max-width: 640px) {
  .form-row {
    grid-template-columns: 1fr;
  }
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 16px;
  margin-top: 32px;
  padding-top: 24px;
  border-top: 2px solid #e2e8f0;
}

.cancel-btn,
.proceed-btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
  min-width: 120px;
  justify-content: center;
}

.cancel-btn {
  background: linear-gradient(135deg, #f1f5f9 0%, #e2e8f0 100%);
  color: #64748b;
  border: 1px solid #cbd5e1;
}

.cancel-btn:hover {
  background: linear-gradient(135deg, #e2e8f0 0%, #cbd5e1 100%);
  color: #475569;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.proceed-btn {
  background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
  color: white;
  border: 1px solid #2980b9;
}

.proceed-btn:hover {
  background: linear-gradient(135deg, #2980b9 0%, #1f5f8b 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(52, 152, 219, 0.3);
}

.proceed-btn:active {
  transform: translateY(0);
}

.cancel-btn:active,
.proceed-btn:active {
  transform: translateY(0);
}

/* Loading state */
.form-select:disabled {
  opacity: 0.7;
}

/* Responsive design */
@media (max-width: 768px) {
  .dialog-content {
    margin: 20px;
    padding: 24px;
    max-width: calc(100vw - 40px);
    max-height: calc(100vh - 40px);
    overflow-y: auto;
  }

  .dialog-buttons {
    flex-direction: column;
  }

  .cancel-btn,
  .proceed-btn {
    width: 100%;
  }
}

/* Small screens and low height screens */
@media (max-height: 700px) {
  .dialog-content {
    margin: 15px;
    padding: 20px;
    max-height: calc(100vh - 30px);
  }

  .dialog-header {
    margin-bottom: 20px;
    padding-bottom: 15px;
  }

  .form-container {
    margin-bottom: 20px;
  }

  .form-section {
    margin-bottom: 15px;
    padding: 15px;
  }

  .form-group {
    margin-bottom: 15px;
  }

  .dialog-buttons {
    margin-top: 20px;
    padding-top: 20px;
  }
}

/* Low resolution screens */
@media (max-height: 600px) {
  .dialog-content {
    margin: 10px;
    padding: 16px;
    max-height: calc(100vh - 20px);
    overflow-y: auto;
  }

  .dialog-header {
    margin-bottom: 16px;
    padding-bottom: 12px;
  }

  .form-container {
    margin-bottom: 16px;
  }

  .form-section {
    margin-bottom: 12px;
    padding: 12px;
  }

  .form-group {
    margin-bottom: 12px;
  }

  .dialog-buttons {
    margin-top: 16px;
    padding-top: 16px;
    position: sticky;
    bottom: 0;
    background: linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
    border-top: 2px solid #e2e8f0;
    padding-bottom: 8px;
  }
}

/* Very low resolution screens */
@media (max-width: 480px) and (max-height: 500px) {
  .dialog-content {
    margin: 5px;
    padding: 12px;
    max-height: calc(100vh - 10px);
  }

  h3 {
    font-size: 1.2rem;
  }

  .dialog-subtitle {
    font-size: 0.8rem;
  }

  .form-section h4 {
    font-size: 1rem;
    margin-bottom: 8px;
  }

  .form-select {
    padding: 8px 12px;
    font-size: 0.9rem;
  }

  .cancel-btn,
  .proceed-btn {
    padding: 8px 16px;
    font-size: 0.9rem;
    min-width: 100px;
  }
}
</style>
