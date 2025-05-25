<script setup>
import { ref, computed, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  billId: {
    type: Number,
    required: true
  },
  show: {
    type: Boolean,
    required: true
  },
  totalAmount: {
    type: Number,
    required: true,
    validator: (value) => !isNaN(value)
  }
})

const emit = defineEmits(['update:show', 'payment-added', 'payment-updated'])

const { callApi, uploadFile } = useApi()
const payments = ref([])
const loading = ref(false)
const error = ref(null)

// Form data
const paymentForm = ref({
  amount: '',
  date_payment: new Date().toISOString().slice(0, 16),
  swift_path: '',
  notes: '',
  swift_file: null
})

// User info
const user = ref(JSON.parse(localStorage.getItem('user')))
const isAdmin = computed(() => user.value?.role_id === 1)

// Computed total payments with proper number handling
const totalPayments = computed(() => {
  return payments.value.reduce((sum, payment) => {
    const amount = Number(payment.amount) || 0
    return sum + amount
  }, 0)
})

// Computed remaining balance with proper number handling
const remainingBalance = computed(() => {
  const total = Number(props.totalAmount) || 0
  const paid = Number(totalPayments.value) || 0
  return total - paid
})

// Form validation
const validateForm = () => {
  if (!paymentForm.value.amount) {
    error.value = 'Amount is required'
    return false
  }
  if (!paymentForm.value.date_payment) {
    error.value = 'Date is required'
    return false
  }
  if (!paymentForm.value.swift_file && !paymentForm.value.swift_path) {
    error.value = 'Swift document is required'
    return false
  }
  return true
}

// Handle file change
const handleSwiftFileChange = (event) => {
  const file = event.target.files[0]
  if (!file) return

  const allowedTypes = [
    'application/pdf',
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp'
  ]
  
  if (!allowedTypes.includes(file.type)) {
    alert('Only PDF and image files (JPEG, PNG, GIF, WEBP) are allowed')
    event.target.value = ''
    return
  }

  paymentForm.value.swift_file = file
}

// Fetch payments with proper number handling
const fetchPayments = async () => {
  loading.value = true
  error.value = null
  
  try {
    const result = await callApi({
      query: `
        SELECT 
          bp.*,
          u.username as created_by
        FROM buy_payments bp
        LEFT JOIN users u ON bp.id_user = u.id
        WHERE bp.id_buy_bill = ?
        ORDER BY bp.date_payment DESC
      `,
      params: [props.billId]
    })
    
    if (result.success) {
      payments.value = result.data.map(payment => ({
        ...payment,
        amount: Number(payment.amount) || 0 // Ensure amount is a number with fallback
      }))
    } else {
      error.value = result.error || 'Failed to fetch payments'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Add payment with proper number handling
const addPayment = async () => {
  if (!validateForm()) return
  
  loading.value = true
  error.value = null

  try {
    let swiftPath = paymentForm.value.swift_path

    // Upload file if provided
    if (paymentForm.value.swift_file) {
      const uploadResult = await uploadFile(paymentForm.value.swift_file)
      if (!uploadResult.success) {
        throw new Error('Failed to upload swift document')
      }
      swiftPath = uploadResult.path
    }

    const result = await callApi({
      query: `
        INSERT INTO buy_payments 
        (id_buy_bill, date_payment, amount, swift_path, notes, id_user)
        VALUES (?, ?, ?, ?, ?, ?)
      `,
      params: [
        props.billId,
        paymentForm.value.date_payment,
        Number(paymentForm.value.amount), // Ensure amount is a number
        swiftPath,
        paymentForm.value.notes,
        user.value.id
      ]
    })

    if (result.success) {
      // Update the payed amount in buy_bill
      await callApi({
        query: `
          UPDATE buy_bill 
          SET payed = (
            SELECT COALESCE(SUM(amount), 0)
            FROM buy_payments
            WHERE id_buy_bill = ?
          )
          WHERE id = ?
        `,
        params: [props.billId, props.billId]
      })

      await fetchPayments()
      emit('payment-added')
      resetForm()
    } else {
      throw new Error(result.error || 'Failed to add payment')
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Reset form
const resetForm = () => {
  paymentForm.value = {
    amount: '',
    date_payment: new Date().toISOString().slice(0, 16),
    swift_path: '',
    notes: '',
    swift_file: null
  }
  error.value = null
}

// Format date
const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

// Format number
const formatNumber = (value) => {
  const num = Number(value)
  return !isNaN(num) ? num.toFixed(2) : 'N/A'
}

// Watch for dialog show/hide
watch(() => props.show, (newVal) => {
  if (newVal) {
    fetchPayments()
  } else {
    resetForm()
  }
})
</script>

<template>
  <div v-if="show" class="dialog-overlay">
    <div class="dialog">
      <div class="dialog-header">
        <h3>Payment Management</h3>
        <button @click="$emit('update:show', false)" class="close-btn">&times;</button>
      </div>

      <div class="dialog-content">
        <!-- Payment Summary -->
        <div class="payment-summary">
          <div class="summary-item">
            <span>Total Amount:</span>
            <span>{{ formatNumber(totalAmount) }}</span>
          </div>
          <div class="summary-item">
            <span>Total Paid:</span>
            <span>{{ formatNumber(totalPayments) }}</span>
          </div>
          <div class="summary-item">
            <span>Remaining:</span>
            <span>{{ formatNumber(remainingBalance) }}</span>
          </div>
        </div>

        <!-- Payment Form -->
        <form @submit.prevent="addPayment" class="payment-form">
          <div class="form-group">
            <label>Amount</label>
            <input 
              type="number" 
              v-model="paymentForm.amount"
              step="0.01"
              required
            >
          </div>

          <div class="form-group">
            <label>Date</label>
            <input 
              type="datetime-local" 
              v-model="paymentForm.date_payment"
              required
            >
          </div>

          <div class="form-group">
            <label>Swift Document</label>
            <input 
              type="file"
              @change="handleSwiftFileChange"
              accept=".pdf,image/*"
            >
          </div>

          <div class="form-group">
            <label>Notes</label>
            <textarea 
              v-model="paymentForm.notes"
              rows="3"
            ></textarea>
          </div>

          <div class="error" v-if="error">{{ error }}</div>

          <button type="submit" :disabled="loading" class="submit-btn">
            {{ loading ? 'Adding...' : 'Add Payment' }}
          </button>
        </form>

        <!-- Payments List -->
        <div class="payments-list">
          <h4>Payment History</h4>
          <div v-if="loading" class="loading">Loading...</div>
          <table v-else>
            <thead>
              <tr>
                <th>Date</th>
                <th>Amount</th>
                <th>Notes</th>
                <th>Created By</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="payment in payments" :key="payment.id">
                <td>{{ formatDate(payment.date_payment) }}</td>
                <td>{{ formatNumber(payment.amount) }}</td>
                <td>{{ payment.notes || '-' }}</td>
                <td>{{ payment.created_by }}</td>
              </tr>
            </tbody>
          </table>
          <div v-if="!loading && payments.length === 0" class="no-data">
            No payments found
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dialog-overlay {
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

.dialog {
  background-color: white;
  border-radius: 8px;
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #111827;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
}

.dialog-content {
  padding: 24px;
}

.payment-summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
  padding: 16px;
  background-color: #f9fafb;
  border-radius: 6px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-item span:first-child {
  color: #6b7280;
  font-size: 0.875rem;
}

.summary-item span:last-child {
  font-size: 1.25rem;
  font-weight: 600;
  color: #111827;
}

.payment-form {
  margin-bottom: 24px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #2563eb;
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.1);
}

.error {
  color: #dc2626;
  margin-bottom: 16px;
  font-size: 0.875rem;
}

.submit-btn {
  background-color: #2563eb;
  color: white;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
}

.submit-btn:disabled {
  background-color: #93c5fd;
  cursor: not-allowed;
}

.payments-list {
  margin-top: 24px;
}

.payments-list h4 {
  margin-bottom: 16px;
  color: #111827;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

th {
  background-color: #f9fafb;
  font-weight: 500;
  color: #374151;
}

.loading {
  text-align: center;
  padding: 24px;
  color: #6b7280;
}

.no-data {
  text-align: center;
  padding: 24px;
  color: #6b7280;
  background-color: #f9fafb;
  border-radius: 6px;
}
</style> 