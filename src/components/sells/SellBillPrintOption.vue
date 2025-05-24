<script setup>
import { ref, defineProps, defineEmits, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../../composables/useApi'

const router = useRouter()
const { callApi } = useApi()

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  billId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits(['close', 'proceed'])

const formData = ref({
  documentType: 'sell_contract',
  currency: 'dza',
  paymentTerms: 'cfr',
  paymentMode: 'TT',
  bankId: null
})

const banks = ref([])

const fetchBanks = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM banks ORDER BY company_name ASC',
      params: []
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
    query: { options: encodeURIComponent(JSON.stringify(formData.value)) }
  }).href
  
  window.open(printUrl, '_blank')
  
  emit('proceed', {
    billId: props.billId,
    ...formData.value
  })
}

const handleCancel = () => {
  emit('close')
}
</script>

<template>
  <div v-if="visible" class="print-options-overlay">
    <div class="print-options-modal">
      <h3>Print Options</h3>
      
      <form @submit.prevent="handleProceed">
        <div class="form-group">
          <label for="documentType">Document Type:</label>
          <select id="documentType" v-model="formData.documentType">
            <option value="invoice">Invoice</option>
            <option value="proforma_invoice">Proforma Invoice</option>
            <option value="sell_contract" selected>Sell Contract</option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="bank">Bank:</label>
          <select id="bank" v-model="formData.bankId">
            <option v-for="bank in banks" :key="bank.id" :value="bank.id">
              {{ bank.bank_name }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="currency">Currency:</label>
          <select id="currency" v-model="formData.currency">
            <option value="usd">USD</option>
            <option value="dza" selected>DZA</option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="paymentTerms">Payment Terms:</label>
          <select id="paymentTerms" v-model="formData.paymentTerms">
            <option value="fob">FOB</option>
            <option value="cfr" selected>CFR</option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="paymentMode">Payment Mode:</label>
          <select id="paymentMode" v-model="formData.paymentMode">
            <option value="TT" selected>TT</option>
            <option value="LC">LC</option>
          </select>
        </div>
        
        <div class="form-actions">
          <button type="button" class="cancel-btn" @click="handleCancel">Cancel</button>
          <button type="submit" class="proceed-btn">Proceed</button>
        </div>
      </form>
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
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.print-options-modal {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
}

h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #374151;
}

.form-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
  color: #374151;
}

select {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  background-color: white;
}

select:focus {
  outline: none;
  border-color: #6366f1;
  ring: 2px solid #6366f1;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

button {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.proceed-btn {
  background-color: #6366f1;
  color: white;
}

.cancel-btn:hover {
  background-color: #4b5563;
}

.proceed-btn:hover {
  background-color: #4f46e5;
}
</style> 