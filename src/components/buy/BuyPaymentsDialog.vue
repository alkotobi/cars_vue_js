<script setup>
import { ref, computed, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  billId: {
    type: Number,
    required: true,
  },
  show: {
    type: Boolean,
    required: true,
  },
  totalAmount: {
    type: Number,
    required: true,
    validator: (value) => !isNaN(value),
  },
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
  swift_file: null,
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

  const allowedTypes = ['application/pdf', 'image/jpeg', 'image/png', 'image/gif', 'image/webp']

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
      params: [props.billId],
    })

    if (result.success) {
      payments.value = result.data.map((payment) => ({
        ...payment,
        amount: Number(payment.amount) || 0, // Ensure amount is a number with fallback
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
        user.value.id,
      ],
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
        params: [props.billId, props.billId],
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
    swift_file: null,
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
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      fetchPayments()
    } else {
      resetForm()
    }
  },
)
</script>

<template>
  <div v-if="show" class="dialog-overlay">
    <div class="dialog">
      <div class="dialog-header">
        <h3>
          <i class="fas fa-money-check-alt"></i>
          Payment Management
        </h3>
        <button @click="$emit('update:show', false)" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="dialog-content">
        <!-- Payment Summary -->
        <div class="payment-summary">
          <div class="summary-item">
            <i class="fas fa-file-invoice-dollar"></i>
            <span>Total Amount:</span>
            <span class="amount">{{ formatNumber(totalAmount) }}</span>
          </div>
          <div class="summary-item">
            <i class="fas fa-check-circle"></i>
            <span>Total Paid:</span>
            <span class="amount paid">{{ formatNumber(totalPayments) }}</span>
          </div>
          <div class="summary-item">
            <i class="fas fa-balance-scale"></i>
            <span>Remaining:</span>
            <span class="amount remaining">{{ formatNumber(remainingBalance) }}</span>
          </div>
        </div>

        <!-- Add Payment Form -->
        <form @submit.prevent="addPayment" class="payment-form">
          <h4>
            <i class="fas fa-plus-circle"></i>
            Add New Payment
          </h4>

          <div class="form-group">
            <label for="amount">
              <i class="fas fa-money-bill-wave"></i>
              Amount
            </label>
            <input
              type="number"
              id="amount"
              v-model="paymentForm.amount"
              step="0.01"
              required
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label for="date">
              <i class="fas fa-calendar-alt"></i>
              Payment Date
            </label>
            <input
              type="datetime-local"
              id="date"
              v-model="paymentForm.date_payment"
              required
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label for="swift">
              <i class="fas fa-file-alt"></i>
              Swift Document
            </label>
            <input
              type="file"
              id="swift"
              @change="handleSwiftFileChange"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp"
              :disabled="loading"
            />
            <div v-if="paymentForm.swift_path" class="current-file">
              <i class="fas fa-check-circle"></i>
              Current: {{ paymentForm.swift_path.split('/').pop() }}
            </div>
          </div>

          <div class="form-group">
            <label for="notes">
              <i class="fas fa-sticky-note"></i>
              Notes
            </label>
            <textarea
              id="notes"
              v-model="paymentForm.notes"
              rows="3"
              :disabled="loading"
            ></textarea>
          </div>

          <div v-if="error" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ error }}
          </div>

          <div class="form-buttons">
            <button type="submit" class="submit-btn" :disabled="loading">
              <i class="fas fa-save"></i>
              <span>{{ loading ? 'Adding Payment...' : 'Add Payment' }}</span>
              <i v-if="loading" class="fas fa-spinner fa-spin"></i>
            </button>
          </div>
        </form>

        <!-- Payments List -->
        <div class="payments-list">
          <h4>
            <i class="fas fa-history"></i>
            Payment History
          </h4>

          <div v-if="loading" class="loading-state">
            <i class="fas fa-spinner fa-spin"></i>
            Loading payments...
          </div>

          <table v-else class="data-table">
            <thead>
              <tr>
                <th><i class="fas fa-calendar"></i> Date</th>
                <th><i class="fas fa-money-bill-wave"></i> Amount</th>
                <th><i class="fas fa-file-alt"></i> Swift</th>
                <th><i class="fas fa-user"></i> Created By</th>
                <th><i class="fas fa-sticky-note"></i> Notes</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="payment in payments" :key="payment.id">
                <td>{{ formatDate(payment.date_payment) }}</td>
                <td>{{ formatNumber(payment.amount) }}</td>
                <td>
                  <a
                    v-if="payment.swift_path"
                    :href="getFileUrl(payment.swift_path)"
                    target="_blank"
                    class="swift-link"
                  >
                    <i class="fas fa-file-pdf"></i>
                    View Swift
                  </a>
                  <span v-else class="no-document">
                    <i class="fas fa-times-circle"></i>
                    No document
                  </span>
                </td>
                <td>{{ payment.created_by }}</td>
                <td>{{ payment.notes || '-' }}</td>
              </tr>
            </tbody>
          </table>
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
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
  position: sticky;
  top: 0;
  background-color: white;
  z-index: 1;
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dialog-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  font-size: 1.25rem;
  padding: 4px;
  transition: color 0.2s ease;
}

.close-btn:hover {
  color: #374151;
}

.dialog-content {
  padding: 20px;
}

.payment-summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
  background-color: #f8fafc;
  padding: 16px;
  border-radius: 8px;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background-color: white;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.summary-item i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.summary-item .amount {
  margin-left: auto;
  font-weight: 600;
  color: #1e293b;
}

.summary-item .amount.paid {
  color: #059669;
}

.summary-item .amount.remaining {
  color: #dc2626;
}

.payment-form {
  background-color: #f8fafc;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 24px;
}

.payment-form h4 {
  margin: 0 0 16px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.payment-form h4 i {
  color: #3b82f6;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-group input:disabled,
.form-group textarea:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.current-file {
  margin-top: 8px;
  font-size: 0.875rem;
  color: #059669;
  display: flex;
  align-items: center;
  gap: 6px;
}

.error-message {
  color: #ef4444;
  margin: 16px 0;
  padding: 12px;
  background-color: #fef2f2;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-buttons {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}

.submit-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.submit-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.payments-list {
  background-color: #f8fafc;
  padding: 20px;
  border-radius: 8px;
}

.payments-list h4 {
  margin: 0 0 16px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.payments-list h4 i {
  color: #3b82f6;
}

.loading-state {
  text-align: center;
  padding: 40px;
  color: #6b7280;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.data-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  background-color: white;
  border-radius: 8px;
  overflow: hidden;
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #374151;
  white-space: nowrap;
}

.data-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.data-table td {
  color: #4b5563;
}

.swift-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  padding: 6px 12px;
  border-radius: 6px;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: #eff6ff;
}

.swift-link:hover {
  background-color: #dbeafe;
  transform: translateY(-1px);
}

.no-document {
  color: #9ca3af;
  font-size: 0.875rem;
  font-style: italic;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

/* Responsive styles */
@media (max-width: 768px) {
  .payment-summary {
    grid-template-columns: 1fr;
  }

  .data-table {
    display: block;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .dialog {
    width: 95%;
    max-height: 95vh;
  }
}
</style>
