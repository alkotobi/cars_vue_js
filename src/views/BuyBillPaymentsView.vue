<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../composables/useApi'

const route = useRoute()
const { callApi, uploadFile, getFileUrl } = useApi()

const billId = ref(route.params.id)
const billInfo = ref(null)
const payments = ref([])
const loading = ref(true)
const error = ref(null)
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingPayment = ref(null)
const user = ref(JSON.parse(localStorage.getItem('user')))

// Add computed properties for permissions
const isAdmin = computed(() => user.value?.role_id === 1)

// Form data
const paymentForm = ref({
  amount: '',
  date_payment: new Date().toISOString().slice(0, 16),
  swift_path: '',
  notes: '',
  swift_file: null
})

onMounted(() => {
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
    const amount = Number(payment.amount) || 0
    return sum + amount
  }, 0)
})

// Computed remaining balance
const remainingBalance = computed(() => {
  if (!billInfo.value?.amount) return 'N/A'
  const total = Number(billInfo.value.amount) || 0
  const paid = Number(totalPayments.value) || 0
  return total - paid
})

const fetchBillInfo = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT 
          bb.*,
          s.name as supplier_name,
          COALESCE((
            SELECT SUM(amount * QTY)
            FROM buy_details
            WHERE id_buy_bill = bb.id
          ), 0) as calculated_amount
        FROM buy_bill bb
        LEFT JOIN suppliers s ON bb.id_supplier = s.id
        WHERE bb.id = ?
      `,
      params: [billId.value]
    })
    
    if (result.success && result.data.length > 0) {
      billInfo.value = {
        ...result.data[0],
        amount: Number(result.data[0].calculated_amount) || 0
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
          bp.*,
          u.username as created_by
        FROM buy_payments bp
        LEFT JOIN users u ON bp.id_user = u.id
        WHERE bp.id_buy_bill = ?
        ORDER BY bp.date_payment DESC
      `,
      params: [billId.value]
    })
    
    if (result.success) {
      payments.value = result.data.map(payment => ({
        ...payment,
        amount: Number(payment.amount) || 0
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

const handleAddPayment = () => {
  // Prevent double-click by checking if dialog is already showing
  if (showAddDialog.value || showEditDialog.value) {
    return
  }
  
  error.value = null // Clear any previous errors
  paymentForm.value = {
    amount: '',
    date_payment: new Date().toISOString().slice(0, 16),
    swift_path: '',
    notes: '',
    swift_file: null
  }
  showAddDialog.value = true
}

const handleEditPayment = (payment) => {
  editingPayment.value = payment
  paymentForm.value = {
    amount: payment.amount,
    date_payment: payment.date_payment.slice(0, 16),
    swift_path: payment.swift_path,
    notes: payment.notes,
    swift_file: null
  }
  showEditDialog.value = true
}

const handleDeletePayment = async (payment) => {
  if (!confirm('Are you sure you want to delete this payment?')) return

  try {
    const result = await callApi({
      query: 'DELETE FROM buy_payments WHERE id = ?',
      params: [payment.id]
    })

    if (result.success) {
      await updateBillPaidAmount()
      await fetchPayments()
    } else {
      throw new Error('Failed to delete payment')
    }
  } catch (err) {
    alert(err.message)
  }
}

const updateBillPaidAmount = async () => {
  try {
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
      params: [billId.value, billId.value]
    })
  } catch (err) {
    console.error('Error updating bill paid amount:', err)
  }
}

const savePayment = async (isEdit = false) => {
  try {
    let swiftPath = paymentForm.value.swift_path

    if (paymentForm.value.swift_file) {
      // Generate a unique filename using bill ID, timestamp, and original extension
      const originalExt = paymentForm.value.swift_file.name.split('.').pop()
      const timestamp = new Date().getTime()
      const customFileName = `bill_${billId.value}_${timestamp}.${originalExt}`

      const uploadResult = await uploadFile(
        paymentForm.value.swift_file, 
        'buy_payment_swifts',
        customFileName
      )
      
      if (!uploadResult.success) {
        throw new Error('Failed to upload swift document')
      }
      swiftPath = uploadResult.file_path
      console.log('Uploaded file path:', swiftPath)
    }

    if (isEdit) {
      if (!paymentForm.value.swift_file && !swiftPath) {
        swiftPath = editingPayment.value.swift_path
      }

      console.log('Updating payment with swift path:', swiftPath)
      const result = await callApi({
        query: `
          UPDATE buy_payments 
          SET amount = ?, date_payment = ?, swift_path = ?, notes = ?
          WHERE id = ?
        `,
        params: [
          Number(paymentForm.value.amount),
          paymentForm.value.date_payment,
          swiftPath,
          paymentForm.value.notes,
          editingPayment.value.id
        ]
      })

      if (!result.success) throw new Error('Failed to update payment')
    } else {
      console.log('Creating new payment with swift path:', swiftPath)
      const result = await callApi({
        query: `
          INSERT INTO buy_payments 
          (id_buy_bill, date_payment, amount, swift_path, notes, id_user)
          VALUES (?, ?, ?, ?, ?, ?)
        `,
        params: [
          billId.value,
          paymentForm.value.date_payment,
          Number(paymentForm.value.amount),
          swiftPath,
          paymentForm.value.notes,
          user.value.id
        ]
      })

      if (!result.success) throw new Error('Failed to add payment')
    }

    await updateBillPaidAmount()
    await fetchPayments()
    showAddDialog.value = false
    showEditDialog.value = false
    editingPayment.value = null
    
    // Reset form
    paymentForm.value = {
      amount: '',
      date_payment: new Date().toISOString().slice(0, 16),
      swift_path: '',
      notes: '',
      swift_file: null
    }
  } catch (err) {
    console.error('Error in savePayment:', err)
    alert(err.message)
  }
}

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
  if (!paymentForm.value.swift_file) {
    error.value = 'Swift document is required'
    return false
  }
  return true
}
</script>

<template>
  <div class="payments-view">
    <!-- Bill Info Header -->
    <div class="bill-header" v-if="billInfo">
      <div class="bill-title">
        <h2>Purchase #{{ billInfo.id }} Payments</h2>
        <span class="supplier-name">{{ billInfo.supplier_name }}</span>
        <span class="bill-date">{{ formatDate(billInfo.date_buy) }}</span>
      </div>
      <div class="bill-summary">
        <div class="summary-item">
          <span class="label">Total Amount:</span>
          <span class="value">{{ formatNumber(billInfo.amount) }}</span>
        </div>
        <div class="summary-item">
          <span class="label">Total Paid:</span>
          <span class="value">{{ formatNumber(totalPayments) }}</span>
        </div>
        <div class="summary-item">
          <span class="label">Remaining:</span>
          <span class="value">{{ formatNumber(remainingBalance) }}</span>
        </div>
      </div>
      <button 
        @click="handleAddPayment" 
        class="add-btn"
        :disabled="showAddDialog || showEditDialog"
      >
        Add Payment
      </button>
    </div>

    <!-- Payments Table -->
    <div class="payments-table">
      <table v-if="!loading && payments.length > 0">
        <thead>
          <tr>
            <th class="print-column">Date</th>
            <th class="print-column">Amount</th>
            <th class="print-column">Swift</th>
            <th class="print-column">Notes</th>
            <th class="print-column">Created By</th>
            <th class="no-print">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="payment in payments" :key="payment.id">
            <td class="print-column">{{ formatDate(payment.date_payment) }}</td>
            <td class="print-column">{{ formatNumber(payment.amount) }}</td>
            <td class="print-column">
              <a 
                v-if="payment.swift_path" 
                :href="getFileUrl(payment.swift_path)"
                target="_blank"
                class="swift-link"
              >
                View Swift
              </a>
              <span v-else class="no-swift">No Swift</span>
            </td>
            <td class="print-column">{{ payment.notes || '-' }}</td>
            <td class="print-column">{{ payment.created_by }}</td>
            <td class="no-print">
              <button 
                @click="handleEditPayment(payment)" 
                class="edit-btn"
                v-if="isAdmin"
              >
                Edit
              </button>
              <button 
                @click="handleDeletePayment(payment)" 
                class="delete-btn"
                v-if="isAdmin"
              >
                Delete
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else-if="loading" class="loading">Loading...</div>
      <div v-else class="no-data">No payments found</div>
    </div>

    <!-- Add/Edit Payment Dialog -->
    <div v-if="showAddDialog || showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{{ showEditDialog ? 'Edit' : 'Add' }} Payment</h3>
          <button 
            @click="showAddDialog = showEditDialog = false" 
            class="close-btn"
          >
            &times;
          </button>
        </div>
        <form @submit.prevent="savePayment(showEditDialog)" class="payment-form">
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
            <label>Swift Document <span class="required">*</span></label>
            <input 
              type="file"
              @change="e => paymentForm.swift_file = e.target.files[0]"
              accept=".pdf,image/*"
              required
            >
            <small class="file-hint">Upload PDF or image file (required)</small>
          </div>

          <div class="form-group">
            <label>Notes</label>
            <textarea 
              v-model="paymentForm.notes"
              rows="3"
            ></textarea>
          </div>

          <div class="error" v-if="error">{{ error }}</div>

          <div class="dialog-buttons">
            <button 
              type="button" 
              @click="showAddDialog = showEditDialog = false" 
              class="cancel-btn"
            >
              Cancel
            </button>
            <button type="submit" :disabled="loading" class="submit-btn">
              {{ loading ? 'Adding...' : 'Add Payment' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.payments-view {
  padding: 20px;
}

.bill-header {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.bill-title {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bill-title h2 {
  margin: 0;
  color: #1e293b;
}

.supplier-name {
  color: #64748b;
  font-size: 0.875rem;
}

.bill-date {
  color: #64748b;
  font-size: 0.875rem;
  margin-top: 2px;
}

.bill-summary {
  display: flex;
  gap: 24px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-item .label {
  color: #64748b;
  font-size: 0.875rem;
}

.summary-item .value {
  color: #1e293b;
  font-weight: 600;
  font-size: 1.125rem;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.add-btn:hover:not(:disabled) {
  background-color: #059669;
}

.add-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.payments-table {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
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
  background-color: #f8fafc;
  font-weight: 500;
  color: #475569;
}

.actions {
  display: flex;
  gap: 8px;
}

.edit-btn, .delete-btn {
  padding: 4px 8px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover {
  background-color: #2563eb;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover {
  background-color: #dc2626;
}

.loading, .no-data {
  text-align: center;
  padding: 40px;
  color: #64748b;
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
  animation: fadeIn 0.2s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.dialog {
  background-color: white;
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  animation: slideUp 0.2s ease-out;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
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
  color: #1e293b;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #64748b;
  cursor: pointer;
}

.payment-form {
  padding: 24px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #475569;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.875rem;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn {
  padding: 8px 16px;
  background-color: #e2e8f0;
  color: #475569;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
}

.submit-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
}

.submit-btn:hover {
  background-color: #2563eb;
}

@media (max-width: 768px) {
  .bill-header {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .bill-summary {
    flex-direction: column;
    gap: 12px;
  }
}

/* Print styles */
@media print {
  .payments-view {
    padding: 0;
    background-color: white;
  }

  .bill-header {
    background-color: white;
    box-shadow: none;
    padding: 0;
    margin-bottom: 20px;
    border-bottom: 2px solid #000;
  }

  .bill-title {
    margin-bottom: 10px;
  }

  .bill-title h2 {
    font-size: 18pt;
    margin-bottom: 5px;
  }

  .supplier-name, .bill-date {
    font-size: 12pt;
    color: #000;
  }

  .bill-summary {
    margin: 15px 0;
    border-bottom: 1px solid #000;
    padding-bottom: 10px;
  }

  .summary-item {
    font-size: 12pt;
  }

  .summary-item .label {
    color: #000;
  }

  .summary-item .value {
    font-weight: bold;
    color: #000;
  }

  .add-btn {
    display: none;
  }

  .payments-table {
    width: 100%;
    margin-top: 20px;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th, td {
    padding: 8px;
    text-align: left;
    border-bottom: 1px solid #000;
    font-size: 10pt;
  }

  th {
    background-color: #f0f0f0 !important;
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }

  /* Hide actions column */
  th:last-child,
  td:last-child {
    display: none;
  }

  /* Hide dialog */
  .dialog-overlay {
    display: none;
  }

  /* Page break settings */
  .payments-view {
    page-break-after: always;
  }

  /* Ensure background colors print */
  * {
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }

  /* A4 size adjustments */
  @page {
    size: A4;
    margin: 1.5cm;
  }

  /* Hide any non-essential elements */
  button,
  .loading,
  .no-data {
    display: none;
  }

  /* Improve table readability */
  tbody tr:nth-child(even) {
    background-color: #f8f8f8;
  }

  /* Hide non-printable elements */
  .no-print {
    display: none !important;
  }

  /* Ensure print columns take full width */
  .print-column {
    width: 25%; /* Distribute space evenly among 4 columns */
  }

  .swift-link {
    color: #000;
    text-decoration: none;
  }

  .swift-link:before {
    content: "📄";
  }

  .no-swift {
    color: #666;
  }

  /* Adjust column widths for 5 columns */
  .print-column {
    width: 20%; /* Distribute space evenly among 5 columns */
  }
}

.swift-link {
  color: #2563eb;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.swift-link:hover {
  text-decoration: underline;
}

.swift-link:before {
  content: "📄";
}

.no-swift {
  color: #6b7280;
  font-style: italic;
}

.required {
  color: #dc2626;
  margin-left: 4px;
}

.file-hint {
  display: block;
  margin-top: 4px;
  color: #6b7280;
  font-size: 0.875rem;
}
</style> 