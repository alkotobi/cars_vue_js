<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../composables/useApi'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()
const { callApi, uploadFile, getFileUrl } = useApi()

const billId = ref(route.params.id)
const billInfo = ref(null)
const payments = ref([])
const loading = ref(true)
const error = ref(null)
const showPaymentDialog = ref(false)
const editingPayment = ref(null)
const user = ref(null)

// `error` is reserved for load failures: it replaces the table, because there is nothing
// meaningful to show. Action and permission problems must not do that, so they go here
// and render as a dismissible notice on top of the still-usable page.
const actionError = ref(null)
const clearActionError = () => {
  actionError.value = null
}

// Add loading state for form submission
const isSubmittingPayment = ref(false)
const deletingPaymentId = ref(null)

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

// Placeholder for any value that is missing or not a number.
const notAvailable = () => t('sellBillPayments.not_available')

// Display formatter only. The computeds below return numbers or null so that no
// arithmetic is ever performed on a formatted string.
const formatNumber = (value) => {
  if (value === null || value === undefined || value === '') return notAvailable()
  const num = Number(value)
  return isNaN(num) ? notAvailable() : num.toFixed(2)
}

// Single money formatter for every currency tile and the dialog balances, so a null
// balance renders as the placeholder rather than as a misleading 0.00.
const formatMoney = (value, currency = 'USD') => {
  const num = Number(value)
  if (value === null || value === undefined || value === '' || isNaN(num)) {
    return notAvailable()
  }
  return currency === 'DA' ? `${num.toFixed(2)} DA` : `$ ${num.toFixed(2)}`
}

// Format date helper
const formatDate = (dateStr) => {
  if (!dateStr) return notAvailable()
  return new Date(dateStr).toLocaleDateString()
}

// Computed total payments (USD). Number() keeps the sum numeric even if a value
// arrives as a decimal string, which would otherwise concatenate instead of adding.
const totalPayments = computed(() => {
  return payments.value.reduce((sum, payment) => sum + (Number(payment.amount_usd) || 0), 0)
})

// Computed total payments (DA)
const totalPaymentsDa = computed(() => {
  return payments.value.reduce((sum, payment) => sum + (Number(payment.amount_da) || 0), 0)
})

// Computed remaining balance (USD)
const remainingBalance = computed(() => {
  const totalCfr = Number(billInfo.value?.total_cfr)
  if (!totalCfr) return null
  return totalCfr - totalPayments.value
})

// Computed remaining balance (DA)
const remainingBalanceDa = computed(() => {
  const totalCfrDa = Number(billInfo.value?.total_cfr_da)
  if (!totalCfrDa) return null
  return totalCfrDa - totalPaymentsDa.value
})

// Cars with no exchange rate cannot be converted, so the DA totals silently skip them.
// Track this so the UI can say so instead of showing a deceptively clean number.
const totalBillCars = computed(() => billInfo.value?.total_cars || 0)
const carsMissingRate = computed(() => billInfo.value?.cars_missing_rate || 0)
const isDaTotalIncomplete = computed(() => carsMissingRate.value > 0)
const isDaTotalUnavailable = computed(
  () => totalBillCars.value > 0 && carsMissingRate.value >= totalBillCars.value,
)

// Paid DA comes from the payments table and stays valid even when rates are missing,
// but the bill DA total and the remaining balance do not, so they report null and
// formatMoney renders the placeholder for them.
const billTotalCfrDa = computed(() =>
  isDaTotalUnavailable.value ? null : Number(billInfo.value?.total_cfr_da) || 0,
)
const billRemainingDa = computed(() =>
  isDaTotalUnavailable.value ? null : remainingBalanceDa.value,
)

// Balances shown inside the payment dialog. While editing, the payment being edited is
// backed out of the paid total, so the figure reflects what will remain after saving.
const dialogRemainingUsd = computed(() => {
  const totalCfr = Number(billInfo.value?.total_cfr)
  if (!totalCfr) return null
  const paid = totalPayments.value - (Number(editingPayment.value?.amount_usd) || 0)
  return totalCfr - paid
})

const dialogRemainingDa = computed(() => {
  const totalCfrDa = Number(billInfo.value?.total_cfr_da)
  if (!totalCfrDa || isDaTotalUnavailable.value) return null
  const paid = totalPaymentsDa.value - (Number(editingPayment.value?.amount_da) || 0)
  return totalCfrDa - paid
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
      formErrors.value.calculation = t('sellBillPayments.err_rate_mismatch', {
        usd,
        rate,
        expectedDa,
      })
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
    formErrors.value.amounts = t('sellBillPayments.err_amount_required')
    return false
  }

  // Check if rate is provided
  if (!paymentForm.value.rate) {
    formErrors.value.rate = t('sellBillPayments.err_rate_required')
    return false
  }

  // Check if swift document is provided for new payments
  if (!editingPayment.value && !paymentForm.value.swift_file && !paymentForm.value.path_swift) {
    formErrors.value.swift = t('sellBillPayments.err_swift_required')
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
    alert(t('sellBillPayments.err_file_type'))
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
            SELECT SUM(
              cs.price_cell + 
              COALESCE(cs.freight, 0) + 
              COALESCE((
                SELECT SUM(ca.value)
                FROM car_apgrades ca
                WHERE ca.id_car = cs.id
              ), 0)
            )
            FROM cars_stock cs
            WHERE cs.id_sell = sb.id
          ) as total_cfr,
          (
            SELECT SUM(
              (cs.price_cell + 
               COALESCE(cs.freight, 0) + 
               COALESCE((
                 SELECT SUM(ca.value)
                 FROM car_apgrades ca
                 WHERE ca.id_car = cs.id
               ), 0)
              ) * COALESCE(cs.rate, 0)
            )
            FROM cars_stock cs
            WHERE cs.id_sell = sb.id AND cs.rate IS NOT NULL AND cs.rate <> 0
          ) as total_cfr_da,
          (
            SELECT COUNT(*)
            FROM cars_stock cs
            WHERE cs.id_sell = sb.id
          ) as total_cars,
          (
            SELECT COUNT(*)
            FROM cars_stock cs
            WHERE cs.id_sell = sb.id AND (cs.rate IS NULL OR cs.rate = 0)
          ) as cars_missing_rate
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
        total_cfr_da: Number(result.data[0].total_cfr_da) || 0,
        total_cars: Number(result.data[0].total_cars) || 0,
        cars_missing_rate: Number(result.data[0].cars_missing_rate) || 0,
      }
    }
  } catch (err) {
    error.value = err.message || t('sellBillPayments.err_load_bill')
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
      error.value = result.error || t('sellBillPayments.err_load_payments')
    }
  } catch (err) {
    error.value = err.message || t('sellBillPayments.err_generic')
  } finally {
    loading.value = false
  }
}

const openAddDialog = () => {
  clearActionError()
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
    actionError.value = t('sellBillPayments.err_no_permission_edit')
    return
  }
  clearActionError()
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
  if (isSubmittingPayment.value) return // prevent double submission

  if (!validateForm()) {
    return
  }

  // Calculate missing amount before submission
  if (!calculateMissingAmount()) {
    actionError.value = t('sellBillPayments.err_calculate_missing')
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
          throw new Error(uploadResult.message || t('sellBillPayments.err_upload_failed'))
        }

        // Store just the relative path without the API endpoint
        paymentForm.value.path_swift = `payments_swift/${filename}`
      } catch (uploadError) {
        console.error('Upload error:', uploadError)
        throw new Error(
          t('sellBillPayments.err_upload_failed_detail', { reason: uploadError.message }),
        )
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
        throw new Error(result.error || t('sellBillPayments.err_update_failed'))
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
        throw new Error(result.error || t('sellBillPayments.err_create_failed'))
      }
    }

    // Reset and refresh
    showPaymentDialog.value = false
    editingPayment.value = null
    clearActionError()
    await fetchPayments()
  } catch (err) {
    console.error('Error in handleSubmit:', err)
    actionError.value = err.message
  } finally {
    isSubmittingPayment.value = false
  }
}

const handleDelete = async (paymentId) => {
  if (deletingPaymentId.value) return
  if (!can_delete_sell_payments.value) {
    actionError.value = t('sellBillPayments.err_no_permission_delete')
    return
  }

  if (!confirm(t('sellBillPayments.confirm_delete_payment'))) {
    return
  }

  deletingPaymentId.value = paymentId
  try {
    const result = await callApi({
      query: 'DELETE FROM sell_payments WHERE id = ?',
      params: [paymentId],
    })

    if (result.success) {
      clearActionError()
      await fetchPayments()
    } else {
      throw new Error(result.error || t('sellBillPayments.err_delete_failed'))
    }
  } catch (err) {
    actionError.value = err.message
  } finally {
    deletingPaymentId.value = null
  }
}
</script>

<template>
  <div class="sell-bill-payments">
    <div class="header">
      <h2>{{ t('sellBillPayments.title', { id: billId }) }}</h2>

      <div v-if="billInfo" class="bill-info">
        <div class="info-grid">
          <div class="info-item">
            <span class="label">{{ t('sellBillPayments.lbl_reference') }}</span>
            <span class="value">{{ billInfo.bill_ref || notAvailable() }}</span>
          </div>
          <div class="info-item">
            <span class="label">{{ t('sellBillPayments.lbl_date') }}</span>
            <span class="value">{{ formatDate(billInfo.date_sell) }}</span>
          </div>
          <div class="info-item">
            <span class="label">{{ t('sellBillPayments.lbl_broker') }}</span>
            <span class="value">{{ billInfo.broker_name || notAvailable() }}</span>
          </div>
          <div class="info-item">
            <span class="label">{{ t('sellBillPayments.lbl_created_by') }}</span>
            <span class="value">{{ billInfo.created_by || notAvailable() }}</span>
          </div>
        </div>

        <div
          v-if="isDaTotalIncomplete"
          class="rate-warning"
          :class="{ fatal: isDaTotalUnavailable }"
        >
          <i
            class="fas"
            :class="isDaTotalUnavailable ? 'fa-exclamation-triangle' : 'fa-exclamation-circle'"
          ></i>
          <div class="rate-warning-text">
            <p v-if="isDaTotalUnavailable" class="rate-warning-title">
              {{
                t('sellBillPayments.da_total_unavailable', {
                  total: totalBillCars,
                  missing: carsMissingRate,
                })
              }}
            </p>
            <p v-else class="rate-warning-title">
              {{
                t('sellBillPayments.da_total_incomplete', {
                  total: totalBillCars,
                  missing: carsMissingRate,
                })
              }}
            </p>
            <p class="rate-warning-hint">{{ t('sellBillPayments.da_total_hint') }}</p>
          </div>
        </div>

        <div class="financial-summary">
          <div class="summary-item">
            <span class="label">{{ t('sellBillPayments.lbl_total_cfr_usd') }}</span>
            <span class="value amount">{{ formatMoney(billInfo.total_cfr) }}</span>
          </div>
          <div class="summary-item">
            <span class="label">{{ t('sellBillPayments.lbl_total_paid_usd') }}</span>
            <span class="value amount">{{ formatMoney(totalPayments) }}</span>
          </div>
          <div class="summary-item">
            <span class="label">{{ t('sellBillPayments.lbl_remaining_usd') }}</span>
            <span class="value amount">{{ formatMoney(remainingBalance) }}</span>
          </div>
          <div class="summary-item" :class="{ 'summary-item-flagged': isDaTotalIncomplete }">
            <span class="label">
              {{ t('sellBillPayments.lbl_total_cfr_da') }}
              <i v-if="isDaTotalIncomplete" class="fas fa-exclamation-triangle flag-icon"></i>
            </span>
            <span class="value amount">{{ formatMoney(billTotalCfrDa, 'DA') }}</span>
          </div>
          <div class="summary-item">
            <span class="label">{{ t('sellBillPayments.lbl_total_paid_da') }}</span>
            <span class="value amount">{{ formatMoney(totalPaymentsDa, 'DA') }}</span>
          </div>
          <div class="summary-item" :class="{ 'summary-item-flagged': isDaTotalIncomplete }">
            <span class="label">
              {{ t('sellBillPayments.lbl_remaining_da') }}
              <i v-if="isDaTotalIncomplete" class="fas fa-exclamation-triangle flag-icon"></i>
            </span>
            <span class="value amount">{{ formatMoney(billRemainingDa, 'DA') }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="actions">
      <button @click="openAddDialog" class="add-btn">
        {{ t('sellBillPayments.add_payment') }}
      </button>
    </div>

    <div v-if="actionError && !showPaymentDialog" class="action-error">
      <i class="fas fa-exclamation-circle"></i>
      <span>{{ actionError }}</span>
      <button type="button" class="action-error-close" @click="clearActionError">
        <i class="fas fa-times"></i>
      </button>
    </div>

    <div v-if="loading" class="loading">{{ t('sellBillPayments.loading') }}</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="payments.length === 0" class="no-data">
      {{ t('sellBillPayments.no_payments') }}
    </div>
    <table v-else class="payments-table">
      <thead>
        <tr>
          <th>{{ t('sellBillPayments.col_id') }}</th>
          <th>{{ t('sellBillPayments.col_date') }}</th>
          <th>{{ t('sellBillPayments.col_amount_usd') }}</th>
          <th>{{ t('sellBillPayments.col_amount_da') }}</th>
          <th>{{ t('sellBillPayments.col_rate') }}</th>
          <th>{{ t('sellBillPayments.col_created_by') }}</th>
          <th>{{ t('sellBillPayments.col_swift') }}</th>
          <th>{{ t('sellBillPayments.col_actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="payment in payments" :key="payment.id">
          <td>{{ payment.id }}</td>
          <td>{{ formatDate(payment.date) }}</td>
          <td>{{ formatNumber(payment.amount_usd) }}</td>
          <td>{{ formatNumber(payment.amount_da) }}</td>
          <td>{{ formatNumber(payment.rate) }}</td>
          <td>{{ payment.created_by || notAvailable() }}</td>
          <td>
            <a v-if="payment.path_swift" :href="getFileUrl(payment.path_swift)" target="_blank">
              {{ t('sellBillPayments.view_swift') }}
            </a>
            <span v-else>{{ t('sellBillPayments.no_document') }}</span>
          </td>
          <td class="actions-cell">
            <button
              @click="openEditDialog(payment)"
              class="edit-btn"
              :disabled="!can_edit_sell_payments"
              :class="{ disabled: !can_edit_sell_payments }"
            >
              {{ t('sellBillPayments.edit') }}
            </button>
            <button
              @click="handleDelete(payment.id)"
              class="delete-btn"
              :disabled="!can_delete_sell_payments || deletingPaymentId !== null"
              :class="{ disabled: !can_delete_sell_payments || deletingPaymentId !== null }"
            >
              {{
                deletingPaymentId === payment.id
                  ? t('sellBillPayments.deleting')
                  : t('sellBillPayments.delete')
              }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Payment Dialog -->
    <div v-if="showPaymentDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>
          {{
            editingPayment ? t('sellBillPayments.edit_payment') : t('sellBillPayments.add_payment')
          }}
        </h3>

        <form @submit.prevent="handleSubmit" class="payment-form">
          <div v-if="actionError" class="error-message">{{ actionError }}</div>

          <div v-if="billInfo" class="dialog-balance">
            <div class="balance-item">
              <span class="balance-label">{{ t('sellBillPayments.remaining_usd') }}</span>
              <span class="balance-value">{{ formatMoney(dialogRemainingUsd) }}</span>
            </div>
            <div class="balance-item" :class="{ 'balance-item-flagged': isDaTotalIncomplete }">
              <span class="balance-label">
                {{ t('sellBillPayments.remaining_da') }}
                <i v-if="isDaTotalIncomplete" class="fas fa-exclamation-triangle balance-flag"></i>
              </span>
              <span class="balance-value">
                {{ formatMoney(dialogRemainingDa, 'DA') }}
              </span>
            </div>
          </div>
          <p v-if="editingPayment && billInfo" class="balance-hint">
            {{ t('sellBillPayments.remaining_excludes_edit') }}
          </p>

          <div class="form-group">
            <label for="amount_usd">{{ t('sellBillPayments.lbl_amount_usd') }}</label>
            <input
              type="number"
              id="amount_usd"
              v-model="paymentForm.amount_usd"
              step="0.01"
              :placeholder="t('sellBillPayments.ph_amount_usd')"
            />
          </div>

          <div class="form-group">
            <label for="amount_da">{{ t('sellBillPayments.lbl_amount_da') }}</label>
            <input
              type="number"
              id="amount_da"
              v-model="paymentForm.amount_da"
              step="0.01"
              :placeholder="t('sellBillPayments.ph_amount_da')"
            />
          </div>

          <div v-if="formErrors.amounts" class="error-message">
            {{ formErrors.amounts }}
          </div>

          <div class="form-group">
            <label for="rate">{{ t('sellBillPayments.lbl_rate') }}</label>
            <input
              type="number"
              id="rate"
              v-model="paymentForm.rate"
              step="0.01"
              required
              :placeholder="t('sellBillPayments.ph_rate')"
            />
            <div v-if="formErrors.rate" class="error-message">
              {{ formErrors.rate }}
            </div>
          </div>

          <div v-if="formErrors.calculation" class="error-message calculation-error">
            {{ formErrors.calculation }}
          </div>

          <div class="form-group">
            <label for="date">{{ t('sellBillPayments.lbl_date') }}</label>
            <input type="date" id="date" v-model="paymentForm.date" required />
          </div>

          <div class="form-group">
            <label for="path_swift">{{ t('sellBillPayments.lbl_swift') }}</label>
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
              {{ t('sellBillPayments.view_current_swift') }}
            </a>
          </div>

          <div class="form-group">
            <label for="notes">{{ t('sellBillPayments.lbl_notes') }}</label>
            <textarea id="notes" v-model="paymentForm.notes" rows="3"></textarea>
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="showPaymentDialog = false" class="cancel-btn">
              {{ t('sellBillPayments.cancel') }}
            </button>
            <button type="submit" class="submit-btn" :disabled="isSubmittingPayment">
              <span v-if="isSubmittingPayment" class="spinner"></span>
              {{
                isSubmittingPayment
                  ? t('sellBillPayments.saving')
                  : editingPayment
                    ? t('sellBillPayments.update')
                    : t('sellBillPayments.add')
              }}
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

.action-error {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 16px;
  padding: 0.75rem 1rem;
  background-color: #fef2f2;
  border: 1px solid #fca5a5;
  border-left: 4px solid #dc2626;
  border-radius: 6px;
  color: #991b1b;
  font-size: 0.9rem;
}

.action-error-close {
  margin-left: auto;
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 4px 6px;
  border-radius: 4px;
  opacity: 0.7;
}

.action-error-close:hover {
  background-color: #fee2e2;
  opacity: 1;
}

.rate-warning {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  margin-top: 1rem;
  padding: 0.85rem 1rem;
  background-color: #fffbeb;
  border: 1px solid #fcd34d;
  border-left: 4px solid #f59e0b;
  border-radius: 6px;
  color: #92400e;
}

.rate-warning.fatal {
  background-color: #fef2f2;
  border-color: #fca5a5;
  border-left-color: #dc2626;
  color: #991b1b;
}

.rate-warning > i {
  font-size: 1.05rem;
  line-height: 1.4;
  flex-shrink: 0;
}

.rate-warning-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.rate-warning-title {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
}

.rate-warning-hint {
  margin: 0;
  font-size: 0.8rem;
  opacity: 0.85;
}

.summary-item-flagged {
  padding: 0.5rem 0.6rem;
  margin: -0.5rem -0.6rem;
  background-color: #fffbeb;
  border-radius: 6px;
}

.flag-icon {
  margin-left: 4px;
  color: #f59e0b;
  font-size: 0.75rem;
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

.dialog-balance {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  padding: 0.75rem 0.9rem;
  margin-bottom: 4px;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
}

.balance-item {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.balance-item-flagged {
  padding: 0.3rem 0.45rem;
  margin: -0.3rem -0.45rem;
  background-color: #fffbeb;
  border-radius: 5px;
}

.balance-label {
  font-size: 0.72rem;
  color: #64748b;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.balance-flag {
  margin-left: 3px;
  color: #f59e0b;
  font-size: 0.65rem;
}

.balance-value {
  font-size: 1.05rem;
  font-weight: 700;
  color: #0f172a;
  font-variant-numeric: tabular-nums;
}

.balance-hint {
  margin: -8px 0 0;
  font-size: 0.72rem;
  color: #94a3b8;
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
