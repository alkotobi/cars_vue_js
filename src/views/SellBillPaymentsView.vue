<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../composables/useApi'

const route = useRoute()
const { callApi, uploadFile, getFileUrl } = useApi()

const billId = ref(route.params.id)
const billInfo = ref(null)
const payments = ref([])
const loading = ref(true)
const error = ref(null)
const showPaymentDialog = ref(false)
const editingPayment = ref(null)
const user = ref(null)

// Add loading state for form submission
const isSubmittingPayment = ref(false)

// Add computed properties for permissions
const can_edit_sell_payments = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_sell_payments')
})

const can_delete_sell_payments = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_delete_sell_payments')
})

// Form data
const paymentForm = ref({
  amount_usd: '',
  amount_da: '',
  rate: '',
  date: new Date().toISOString().split('T')[0],
  path_swift: '',
  notes: '',
  swift_file: null,
})

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchBillInfo()
  fetchPayments()
})

// Helper function to format number safely
const formatNumber = (value) => {
  const num = Number(value)
  return !isNaN(num) ? num.toFixed(2) : 'N/A'
}

// Format date helper
const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

// Computed total payments
const totalPayments = computed(() => {
  return payments.value.reduce((sum, payment) => {
    return sum + (payment.amount_usd || 0)
  }, 0)
})

// Computed remaining balance
const remainingBalance = computed(() => {
  if (!billInfo.value?.total_cfr) return 'N/A'
  return billInfo.value.total_cfr - totalPayments.value
})

// Form validation
const formErrors = ref({
  amounts: '',
  rate: '',
  calculation: '',
  swift: '',
})

// Verify if calculations are correct
const verifyCalculations = () => {
  const usd = Number(paymentForm.value.amount_usd)
  const da = Number(paymentForm.value.amount_da)
  const rate = Number(paymentForm.value.rate)

  // If all three values are provided
  if (usd && da && rate) {
    const expectedDa = (usd * rate).toFixed(2)
    const expectedUsd = (da / rate).toFixed(2)
    const tolerance = 0.02 // 2% tolerance for rounding differences

    // Check if the actual DA amount is within tolerance of expected DA
    const daError = Math.abs(da - expectedDa) / expectedDa
    // Check if the actual USD amount is within tolerance of expected USD
    const usdError = Math.abs(usd - expectedUsd) / expectedUsd

    if (daError > tolerance || usdError > tolerance) {
      formErrors.value.calculation = `The values don't match the rate. Expected: ${usd} USD × ${rate} = ${expectedDa} DA`
      return false
    }
  }

  formErrors.value.calculation = ''
  return true
}

// Calculate missing amount based on rate and existing amount
const calculateMissingAmount = () => {
  const rate = Number(paymentForm.value.rate)
  const usd = Number(paymentForm.value.amount_usd)
  const da = Number(paymentForm.value.amount_da)

  if (!rate) return false

  // If both amounts are missing, validation will catch it
  if (!usd && !da) return true

  // If both amounts exist, verify calculations
  if (usd && da) {
    return verifyCalculations()
  }

  // Calculate missing amount
  if (usd) {
    paymentForm.value.amount_da = (usd * rate).toFixed(2)
  } else if (da) {
    paymentForm.value.amount_usd = (da / rate).toFixed(2)
  }

  return true
}

// Form validation
const validateForm = () => {
  formErrors.value = {
    amounts: '',
    rate: '',
    calculation: '',
    swift: '',
  }

  // Check if at least one amount is provided
  if (!paymentForm.value.amount_usd && !paymentForm.value.amount_da) {
    formErrors.value.amounts = 'Either USD or DA amount is required'
    return false
  }

  // Check if rate is provided
  if (!paymentForm.value.rate) {
    formErrors.value.rate = 'Rate is required'
    return false
  }

  // Check if swift document is provided for new payments
  if (!editingPayment.value && !paymentForm.value.swift_file && !paymentForm.value.path_swift) {
    formErrors.value.swift = 'Swift document is required'
    return false
  }

  return true
}

// Handle file change for swift document
const handleSwiftFileChange = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // Check if file is PDF or image
  const allowedTypes = ['application/pdf', 'image/jpeg', 'image/png', 'image/gif', 'image/webp']

  if (!allowedTypes.includes(file.type)) {
    alert('Only PDF and image files (JPEG, PNG, GIF, WEBP) are allowed')
    event.target.value = ''
    return
  }

  // Store the file for upload
  paymentForm.value.swift_file = file
}

const fetchBillInfo = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT 
          sb.*,
          c.name as broker_name,
          u.username as created_by,
          (
            SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
            FROM cars_stock cs
            WHERE cs.id_sell = sb.id
          ) as total_cfr
        FROM sell_bill sb
        LEFT JOIN clients c ON sb.id_broker = c.id
        LEFT JOIN users u ON sb.id_user = u.id
        WHERE sb.id = ?
      `,
      params: [billId.value],
    })

    if (result.success && result.data.length > 0) {
      billInfo.value = {
        ...result.data[0],
        total_cfr: Number(result.data[0].total_cfr) || 0,
      }
    }
  } catch (err) {
    error.value = err.message || 'Failed to fetch bill information'
  }
}

const fetchPayments = async () => {
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          sp.*,
          u.username as created_by
        FROM sell_payments sp
        LEFT JOIN users u ON sp.id_user = u.id
        WHERE sp.id_sell_bill = ?
        ORDER BY sp.date DESC
      `,
      params: [billId.value],
    })

    if (result.success) {
      // Convert string numbers to actual numbers
      payments.value = result.data.map((payment) => ({
        ...payment,
        amount_usd: payment.amount_usd ? Number(payment.amount_usd) : null,
        amount_da: payment.amount_da ? Number(payment.amount_da) : null,
        rate: payment.rate ? Number(payment.rate) : null,
      }))
    } else {
      error.value = result.error || 'Failed to fetch payments'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const openAddDialog = () => {
  editingPayment.value = null
  paymentForm.value = {
    amount_usd: '',
    amount_da: '',
    rate: '',
    date: new Date().toISOString().split('T')[0],
    path_swift: '',
    notes: '',
    swift_file: null,
  }
  showPaymentDialog.value = true
}

const openEditDialog = (payment) => {
  if (!can_edit_sell_payments.value) {
    error.value = 'You do not have permission to edit payments'
    return
  }
  editingPayment.value = payment
  paymentForm.value = {
    amount_usd: payment.amount_usd,
    amount_da: payment.amount_da,
    rate: payment.rate,
    date: payment.date.split('T')[0],
    path_swift: payment.path_swift || '',
    notes: payment.notes || '',
    swift_file: null,
  }
  showPaymentDialog.value = true
}

const handleSubmit = async () => {
  // Prevent multiple submissions
  if (isSubmittingPayment.value) {
    return
  }

  if (!validateForm()) {
    return
  }

  // Calculate missing amount before submission
  if (!calculateMissingAmount()) {
    error.value = 'Failed to calculate missing amount. Please check the rate.'
    return
  }

  try {
    isSubmittingPayment.value = true

    // Handle swift document upload first if there's a new file
    if (paymentForm.value.swift_file) {
      try {
        // Create filename using payment details and original extension
        const date = new Date()
        const timestamp = date.getTime() // Add timestamp for uniqueness
        const dateStr = date.toISOString().split('T')[0]
        const timeStr = date.toISOString().split('T')[1].split('.')[0].replace(/:/g, '-')
        const amountStr = paymentForm.value.amount_usd
          ? `_${paymentForm.value.amount_usd}USD`
          : `_${paymentForm.value.amount_da}DA`
        const fileExt = paymentForm.value.swift_file.name.split('.').pop().toLowerCase()

        // Format: swift_payment_billId_date_time_amount_timestamp.ext
        const filename = `swift_payment_${billId.value}_${dateStr}_${timeStr}${amountStr}_${timestamp}.${fileExt}`

        const uploadResult = await uploadFile(
          paymentForm.value.swift_file,
          'payments_swift',
          filename,
        )

        if (!uploadResult.success) {
          throw new Error(uploadResult.message || 'Failed to upload swift document')
        }

        // Store just the relative path without the API endpoint
        paymentForm.value.path_swift = `payments_swift/${filename}`
      } catch (uploadError) {
        console.error('Upload error:', uploadError)
        throw new Error(`Failed to upload swift document: ${uploadError.message}`)
      }
    }

    if (editingPayment.value) {
      // Update existing payment
      const result = await callApi({
        query: `
          UPDATE sell_payments 
          SET amount_usd = ?,
              amount_da = ?,
              rate = ?,
              date = ?,
              path_swift = ?,
              notes = ?
          WHERE id = ?
        `,
        params: [
          paymentForm.value.amount_usd || null,
          paymentForm.value.amount_da || null,
          paymentForm.value.rate,
          paymentForm.value.date,
          paymentForm.value.path_swift,
          paymentForm.value.notes,
          editingPayment.value.id,
        ],
      })

      if (!result.success) {
        throw new Error(result.error || 'Failed to update payment')
      }
    } else {
      // Create new payment
      const result = await callApi({
        query: `
          INSERT INTO sell_payments 
          (id_sell_bill, amount_usd, amount_da, rate, date, path_swift, notes, id_user)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `,
        params: [
          billId.value,
          paymentForm.value.amount_usd || null,
          paymentForm.value.amount_da || null,
          paymentForm.value.rate,
          paymentForm.value.date,
          paymentForm.value.path_swift,
          paymentForm.value.notes,
          user.value?.id,
        ],
      })

      if (!result.success) {
        throw new Error(result.error || 'Failed to create payment')
      }
    }

    // Reset and refresh
    showPaymentDialog.value = false
    editingPayment.value = null
    await fetchPayments()
  } catch (err) {
    console.error('Error in handleSubmit:', err)
    error.value = err.message
  } finally {
    isSubmittingPayment.value = false
  }
}

const handleDelete = async (paymentId) => {
  if (!can_delete_sell_payments.value) {
    error.value = 'You do not have permission to delete payments'
    return
  }

  if (!confirm('Are you sure you want to delete this payment?')) {
    return
  }

  try {
    const result = await callApi({
      query: 'DELETE FROM sell_payments WHERE id = ?',
      params: [paymentId],
    })

    if (result.success) {
      await fetchPayments()
    } else {
      throw new Error(result.error || 'Failed to delete payment')
    }
  } catch (err) {
    error.value = err.message
  }
}
</script>

<template>
  <div class="sell-bill-payments">
    <div class="header">
      <h2>Payments for Sell Bill #{{ billId }}</h2>

      <div v-if="billInfo" class="bill-info">
        <div class="info-grid">
          <div class="info-item">
            <span class="label">Reference:</span>
            <span class="value">{{ billInfo.bill_ref || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="label">Date:</span>
            <span class="value">{{ formatDate(billInfo.date_sell) }}</span>
          </div>
          <div class="info-item">
            <span class="label">Broker:</span>
            <span class="value">{{ billInfo.broker_name || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="label">Created By:</span>
            <span class="value">{{ billInfo.created_by || 'N/A' }}</span>
          </div>
        </div>

        <div class="financial-summary">
          <div class="summary-item">
            <span class="label">Total CFR:</span>
            <span class="value amount">$ {{ formatNumber(billInfo.total_cfr) }}</span>
          </div>
          <div class="summary-item">
            <span class="label">Total Paid:</span>
            <span class="value amount">$ {{ formatNumber(totalPayments) }}</span>
          </div>
          <div class="summary-item">
            <span class="label">Remaining:</span>
            <span class="value amount">$ {{ formatNumber(remainingBalance) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="actions">
      <button @click="openAddDialog" class="add-btn">Add Payment</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="payments.length === 0" class="no-data">No payments found for this bill</div>
    <table v-else class="payments-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Date</th>
          <th>Amount (USD)</th>
          <th>Amount (DA)</th>
          <th>Rate</th>
          <th>Created By</th>
          <th>Swift Document</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="payment in payments" :key="payment.id">
          <td>{{ payment.id }}</td>
          <td>{{ formatDate(payment.date) }}</td>
          <td>{{ formatNumber(payment.amount_usd) }}</td>
          <td>{{ formatNumber(payment.amount_da) }}</td>
          <td>{{ formatNumber(payment.rate) }}</td>
          <td>{{ payment.created_by || 'N/A' }}</td>
          <td>
            <a v-if="payment.path_swift" :href="getFileUrl(payment.path_swift)" target="_blank"
              >View Swift</a
            >
            <span v-else>No document</span>
          </td>
          <td class="actions-cell">
            <button
              @click="openEditDialog(payment)"
              class="edit-btn"
              :disabled="!can_edit_sell_payments"
              :class="{ disabled: !can_edit_sell_payments }"
            >
              Edit
            </button>
            <button
              @click="handleDelete(payment.id)"
              class="delete-btn"
              :disabled="!can_delete_sell_payments"
              :class="{ disabled: !can_delete_sell_payments }"
            >
              Delete
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Payment Dialog -->
    <div v-if="showPaymentDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>{{ editingPayment ? 'Edit Payment' : 'Add Payment' }}</h3>

        <form @submit.prevent="handleSubmit" class="payment-form">
          <div class="form-group">
            <label for="amount_usd">Amount (USD):</label>
            <input
              type="number"
              id="amount_usd"
              v-model="paymentForm.amount_usd"
              step="0.01"
              placeholder="Enter USD amount"
            />
          </div>

          <div class="form-group">
            <label for="amount_da">Amount (DA):</label>
            <input
              type="number"
              id="amount_da"
              v-model="paymentForm.amount_da"
              step="0.01"
              placeholder="Enter DA amount"
            />
          </div>

          <div v-if="formErrors.amounts" class="error-message">
            {{ formErrors.amounts }}
          </div>

          <div class="form-group">
            <label for="rate">Rate:</label>
            <input
              type="number"
              id="rate"
              v-model="paymentForm.rate"
              step="0.01"
              required
              placeholder="Enter rate"
            />
            <div v-if="formErrors.rate" class="error-message">
              {{ formErrors.rate }}
            </div>
          </div>

          <div v-if="formErrors.calculation" class="error-message calculation-error">
            {{ formErrors.calculation }}
          </div>

          <div class="form-group">
            <label for="date">Date:</label>
            <input type="date" id="date" v-model="paymentForm.date" required />
          </div>

          <div class="form-group">
            <label for="path_swift">Swift Document (PDF or Image):</label>
            <input
              type="file"
              id="path_swift"
              @change="handleSwiftFileChange"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp"
              :required="!editingPayment"
            />
            <div v-if="formErrors.swift" class="error-message">
              {{ formErrors.swift }}
            </div>
            <a
              v-if="paymentForm.path_swift"
              :href="getFileUrl(paymentForm.path_swift)"
              target="_blank"
              class="current-file-link"
            >
              View Current Swift Document
            </a>
          </div>

          <div class="form-group">
            <label for="notes">Notes:</label>
            <textarea id="notes" v-model="paymentForm.notes" rows="3"></textarea>
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="showPaymentDialog = false" class="cancel-btn">
              Cancel
            </button>
            <button type="submit" class="submit-btn" :disabled="isSubmittingPayment">
              <span v-if="isSubmittingPayment" class="spinner"></span>
              {{ isSubmittingPayment ? 'Saving...' : editingPayment ? 'Update' : 'Add' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sell-bill-payments {
  padding: 20px;
}

.header {
  margin-bottom: 20px;
}

.header h2 {
  color: #1f2937;
  font-size: 1.5rem;
  margin-bottom: 1rem;
}

.actions {
  margin-bottom: 20px;
}

.add-btn {
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-weight: 500;
}

.add-btn:hover {
  background-color: #059669;
}

.bill-info {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 1.5rem;
  margin-top: 1rem;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.financial-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #e2e8f0;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.label {
  font-size: 0.875rem;
  color: #64748b;
  font-weight: 500;
}

.value {
  font-size: 1rem;
  color: #1f2937;
  font-weight: 500;
}

.amount {
  font-size: 1.25rem;
  color: #0f172a;
  font-weight: 600;
}

.payments-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.payments-table th,
.payments-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.payments-table th {
  background-color: #f3f4f6;
  font-weight: 600;
  color: #374151;
}

.payments-table tr:hover {
  background-color: #f9fafb;
}

.actions-cell {
  display: flex;
  gap: 8px;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 4px 8px;
  cursor: pointer;
  font-size: 0.875rem;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 4px 8px;
  cursor: pointer;
  font-size: 0.875rem;
}

.edit-btn:hover {
  background-color: #2563eb;
}

.delete-btn:hover {
  background-color: #dc2626;
}

.loading,
.error,
.no-data {
  padding: 20px;
  text-align: center;
  color: #6b7280;
}

.error {
  color: #ef4444;
}

a {
  color: #3b82f6;
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}

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
  padding: 24px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1);
}

.dialog h3 {
  margin-bottom: 20px;
  color: #1f2937;
  font-size: 1.25rem;
}

.payment-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-group label {
  font-size: 0.875rem;
  color: #374151;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1rem;
}

.form-group input:focus,
.form-group textarea:focus {
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
  background-color: #9ca3af;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-weight: 500;
}

.submit-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-weight: 500;
}

.cancel-btn:hover {
  background-color: #6b7280;
}

.submit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.submit-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid #ffffff;
  border-radius: 50%;
  border-top-color: transparent;
  animation: spin 1s ease-in-out infinite;
  margin-right: 8px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-message {
  color: #ef4444;
  font-size: 0.875rem;
  margin-top: 0.25rem;
}

.calculation-error {
  margin: 1rem 0;
  padding: 0.5rem;
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  border-radius: 4px;
}

.current-file-link {
  color: #3b82f6;
  text-decoration: none;
}

.current-file-link:hover {
  text-decoration: underline;
}

.edit-btn.disabled,
.delete-btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.edit-btn:disabled,
.delete-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
