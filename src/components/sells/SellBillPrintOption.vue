<script setup>
import { ref, defineProps, defineEmits, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../../composables/useApi'

const router = useRouter()
const { callApi } = useApi()

const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  billId: {
    type: Number,
    default: null,
  },
})

const emit = defineEmits(['close', 'proceed'])

const formData = ref({
  documentType: 'sell_contract',
  currency: 'dza',
  paymentTerms: 'cfr',
  paymentMode: 'TT',
  bankId: null,
})

const banks = ref([])
const isProcessing = ref(false)

const fetchBanks = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM banks ORDER BY company_name ASC',
      params: [],
    })
    if (result.success && result.data.length > 0) {
      banks.value = result.data
      // Set the first bank as default
      formData.value.bankId = result.data[0].id
    }
  } catch (err) {
    console.error('Error fetching banks:', err)
  }
}

onMounted(() => {
  fetchBanks()
})

const handleProceed = () => {
  // Open print page in new tab
  const printUrl = router.resolve({
    name: 'print',
    params: { billId: props.billId },
    query: { options: encodeURIComponent(JSON.stringify(formData.value)) },
  }).href

  window.open(printUrl, '_blank')

  emit('proceed', {
    billId: props.billId,
    ...formData.value,
  })
}

const handleCancel = () => {
  emit('close')
}
</script>

<template>
  <div v-if="visible" class="print-options-overlay">
    <div class="print-options-dialog">
      <!-- Loading Overlay -->
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>{{ isProcessing ? 'Generating document...' : 'Loading...' }}</span>
      </div>

      <div class="dialog-header">
        <h3>
          <i class="fas fa-print"></i>
          Print Options
        </h3>
        <button @click="$emit('close')" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div v-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div class="options-content">
        <div class="option-group">
          <label>
            <i class="fas fa-file-alt"></i>
            Document Type:
          </label>
          <div class="radio-group">
            <label class="radio-label">
              <input
                type="radio"
                v-model="selectedOption"
                value="invoice"
                :disabled="isProcessing"
              />
              <i class="fas fa-file-invoice-dollar"></i>
              Invoice
            </label>
            <label class="radio-label">
              <input
                type="radio"
                v-model="selectedOption"
                value="proforma"
                :disabled="isProcessing"
              />
              <i class="fas fa-file-contract"></i>
              Proforma
            </label>
          </div>
        </div>

        <div class="option-group">
          <label>
            <i class="fas fa-language"></i>
            Language:
          </label>
          <div class="radio-group">
            <label class="radio-label">
              <input type="radio" v-model="language" value="en" :disabled="isProcessing" />
              <i class="fas fa-flag-usa"></i>
              English
            </label>
            <label class="radio-label">
              <input type="radio" v-model="language" value="ar" :disabled="isProcessing" />
              <i class="fas fa-flag"></i>
              Arabic
            </label>
          </div>
        </div>
      </div>

      <div class="dialog-actions">
        <button @click="$emit('close')" :disabled="isProcessing" class="cancel-btn">
          <i class="fas fa-times"></i>
          Cancel
        </button>
        <button @click="handleProceed" :disabled="isProcessing" class="proceed-btn">
          <i class="fas fa-print"></i>
          {{ isProcessing ? 'Generating...' : 'Generate Document' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.print-options-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.print-options-dialog {
  position: relative;
  background: white;
  border-radius: 0.5rem;
  width: 90%;
  max-width: 500px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  z-index: 10;
  border-radius: 0.5rem;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.dialog-header h3 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1f2937;
  margin: 0;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.5rem;
  transition: color 0.2s;
}

.close-btn:hover:not(:disabled) {
  color: #374151;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.options-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.option-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.option-group > label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  font-weight: 500;
}

.option-group > label i {
  color: #6b7280;
}

.radio-group {
  display: flex;
  gap: 1rem;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  cursor: pointer;
}

.radio-label input[type='radio'] {
  cursor: pointer;
}

.radio-label i {
  color: #6b7280;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1.5rem;
}

.cancel-btn,
.proceed-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.proceed-btn {
  background-color: #10b981;
  color: white;
}

.proceed-btn:hover:not(:disabled) {
  background-color: #059669;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Add smooth transitions */
button i {
  transition: transform 0.2s;
}

button:hover:not(:disabled) i {
  transform: scale(1.1);
}
</style>
