<script setup>
import { ref, defineProps, defineEmits, onMounted, watch } from 'vue'
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
  <div v-if="visible" class="print-options-dialog">
    <div class="dialog-content">
      <h3>Print Options</h3>

      <div class="form-group">
        <label>Document Type:</label>
        <select v-model="formData.documentType">
          <option v-for="type in documentTypes" :key="type.value" :value="type.value">
            {{ type.label }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label>Payment Terms:</label>
        <select v-model="formData.paymentTerms">
          <option v-for="term in paymentTermsOptions" :key="term.value" :value="term.value">
            {{ term.label }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label>Payment Type:</label>
        <select v-model="formData.paymentMode">
          <option v-for="type in paymentTypes" :key="type.value" :value="type.value">
            {{ type.label }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label>Currency:</label>
        <select v-model="formData.currency">
          <option v-for="currency in currencyOptions" :key="currency.value" :value="currency.value">
            {{ currency.label }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label>Bank:</label>
        <select v-model="formData.bankId" :disabled="isLoadingBanks">
          <option v-if="isLoadingBanks" value="">Loading banks...</option>
          <option v-else value="">Select Bank</option>
          <option v-for="bank in banks" :key="bank.id" :value="bank.id">
            {{ bank.company_name }}
          </option>
        </select>
      </div>

      <div class="dialog-buttons">
        <button @click="handleCancel" class="cancel-btn">Cancel</button>
        <button @click="handleProceed" class="proceed-btn">Print</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.print-options-dialog {
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

.dialog-content {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1);
}

h3 {
  margin-top: 0;
  margin-bottom: 24px;
  color: #1f2937;
  font-size: 1.25rem;
}

.form-group {
  margin-bottom: 16px;
}

label {
  display: block;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
}

select {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1rem;
}

select:focus {
  outline: none;
  border-color: #3b82f6;
  ring: 2px solid #3b82f6;
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn {
  padding: 8px 16px;
  background-color: #9ca3af;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
}

.proceed-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
}

.cancel-btn:hover {
  background-color: #6b7280;
}

.proceed-btn:hover {
  background-color: #2563eb;
}
</style>
