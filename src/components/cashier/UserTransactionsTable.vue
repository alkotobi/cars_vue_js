<script setup>
import { ref, watch, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  selectedUser: {
    type: Object,
    required: true,
  },
})

const { callApi } = useApi()
const sellPayments = ref([])
const receivedTransfers = ref([])
const sentTransfers = ref([])
const externalTransfers = ref([])
const loading = ref(false)
const error = ref(null)

const fetchAllData = async () => {
  if (!props.selectedUser?.id) return

  loading.value = true
  error.value = null

  try {
    console.log('Fetching data for user:', props.selectedUser.id)

    // Fetch sell payments (money in)
    const sellPaymentsResult = await callApi({
      query: `
        SELECT 
          id,
          date as transaction_date,
          amount_da as amount,
          notes
        FROM sell_payments
        WHERE id_user = ?
      `,
      params: [props.selectedUser.id],
    })
    console.log('Sell payments result:', sellPaymentsResult)

    // Fetch received transfers (money in)
    const receivedTransfersResult = await callApi({
      query: `
        SELECT 
          id,
          date_transfer as transaction_date,
          amount,
          notes
        FROM transfers_inter
        WHERE to_user_id = ? AND id_admin_confirm IS NOT NULL
      `,
      params: [props.selectedUser.id],
    })
    console.log('Received transfers result:', receivedTransfersResult)

    // Fetch sent transfers (money out)
    const sentTransfersResult = await callApi({
      query: `
        SELECT 
          id,
          date_transfer as transaction_date,
          amount,
          notes
        FROM transfers_inter
        WHERE from_user_id = ?
      `,
      params: [props.selectedUser.id],
    })
    console.log('Sent transfers result:', sentTransfersResult)

    // Fetch external transfers (money out)
    const externalTransfersResult = await callApi({
      query: `
        SELECT 
          id,
          date_do_transfer as transaction_date,
          amount_sending_da as amount,
          notes
        FROM transfers
        WHERE id_user_do_transfer = ? AND id_user_receive_transfer IS NOT NULL
      `,
      params: [props.selectedUser.id],
    })
    console.log('External transfers result:', externalTransfersResult)

    // Reset arrays before updating with new data
    sellPayments.value = []
    receivedTransfers.value = []
    sentTransfers.value = []
    externalTransfers.value = []

    if (sellPaymentsResult?.success) {
      sellPayments.value = sellPaymentsResult.data.map((payment) => ({
        ...payment,
        type: 'in',
        source: 'sell_payment',
      }))
    }

    if (receivedTransfersResult?.success) {
      receivedTransfers.value = receivedTransfersResult.data.map((transfer) => ({
        ...transfer,
        type: 'in',
        source: 'transfer_received',
      }))
    }

    if (sentTransfersResult?.success) {
      sentTransfers.value = sentTransfersResult.data.map((transfer) => ({
        ...transfer,
        type: 'out',
        source: 'transfer_sent',
      }))
    }

    if (externalTransfersResult?.success) {
      externalTransfers.value = externalTransfersResult.data.map((transfer) => ({
        ...transfer,
        type: 'out',
        source: 'external_transfer',
      }))
    }

    console.log('Data fetching complete', {
      sellPayments: sellPayments.value,
      receivedTransfers: receivedTransfers.value,
      sentTransfers: sentTransfers.value,
      externalTransfers: externalTransfers.value,
    })
  } catch (err) {
    console.error('Error fetching data:', err)
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Watch for changes in selected user
watch(
  () => props.selectedUser?.id,
  (newUserId) => {
    console.log('Selected user changed:', newUserId)
    if (newUserId) {
      fetchAllData()
    } else {
      sellPayments.value = []
      receivedTransfers.value = []
      sentTransfers.value = []
      externalTransfers.value = []
      loading.value = false
    }
  },
  { immediate: true },
)

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

const formatCurrency = (value) => {
  // Ensure value is a valid number
  const numValue = parseFloat(value) || 0
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'DZD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(numValue)
}

// Combine all transactions and calculate running balance
const transactionsWithBalance = computed(() => {
  // Combine all transactions
  const allTransactions = [
    ...sellPayments.value,
    ...receivedTransfers.value,
    ...sentTransfers.value,
    ...externalTransfers.value,
  ]

  // Sort by date (oldest first)
  allTransactions.sort((a, b) => {
    return new Date(a.transaction_date) - new Date(b.transaction_date)
  })

  let runningBalance = 0
  let previousBalance = 0

  // Process transactions and calculate balances
  return allTransactions.map((t, index) => {
    // Store the previous balance before updating
    previousBalance = runningBalance

    // Ensure amount is a valid number
    const amount = t.amount ? parseFloat(t.amount) : 0
    if (!isNaN(amount)) {
      if (t.type === 'in') {
        runningBalance += amount
      } else {
        runningBalance -= amount
      }
    }

    return {
      ...t,
      amount: amount || 0,
      previousBalance,
      balance: runningBalance || 0,
    }
  }) // Remove reverse() to show oldest first
})

const getSourceLabel = (source) => {
  switch (source) {
    case 'sell_payment':
      return 'Sell Payment'
    case 'transfer_received':
      return 'Transfer Received'
    case 'transfer_sent':
      return 'Transfer Sent'
    case 'external_transfer':
      return 'External Transfer'
    default:
      return source
  }
}
</script>

<template>
  <div class="user-transactions" v-if="selectedUser?.id">
    <div class="header">
      <h3>
        <i class="fas fa-history"></i>
        Transactions for {{ selectedUser.username }}
      </h3>
      <div class="header-stats">
        <div class="stat-item">
          <i class="fas fa-exchange-alt"></i>
          <span class="stat-label">Total Transactions:</span>
          <span class="stat-value">{{ transactionsWithBalance.length }}</span>
        </div>
        <div class="stat-item">
          <i class="fas fa-arrow-down"></i>
          <span class="stat-label">Money In:</span>
          <span class="stat-value positive">{{
            formatCurrency(
              transactionsWithBalance.reduce(
                (sum, t) => (t.type === 'in' ? sum + t.amount : sum),
                0,
              ),
            )
          }}</span>
        </div>
        <div class="stat-item">
          <i class="fas fa-arrow-up"></i>
          <span class="stat-label">Money Out:</span>
          <span class="stat-value negative">{{
            formatCurrency(
              transactionsWithBalance.reduce(
                (sum, t) => (t.type === 'out' ? sum + t.amount : sum),
                0,
              ),
            )
          }}</span>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div class="transactions-table" :class="{ 'is-loading': loading }">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading transactions...</span>
      </div>

      <table v-if="!loading && transactionsWithBalance.length > 0">
        <thead>
          <tr>
            <th><i class="fas fa-calendar-alt"></i> Date</th>
            <th><i class="fas fa-exchange-alt"></i> Type</th>
            <th><i class="fas fa-history"></i> Previous Balance</th>
            <th><i class="fas fa-arrow-down"></i> Money In</th>
            <th><i class="fas fa-arrow-up"></i> Money Out</th>
            <th><i class="fas fa-wallet"></i> Current Balance</th>
            <th><i class="fas fa-tag"></i> Source</th>
            <th><i class="fas fa-sticky-note"></i> Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(transaction, index) in transactionsWithBalance"
            :key="index"
            :class="{
              'positive-row': transaction.balance >= 0,
              'negative-row': transaction.balance < 0,
            }"
          >
            <td>{{ formatDate(transaction.transaction_date) }}</td>
            <td>
              <span class="transaction-type" :class="transaction.type">
                <i :class="transaction.type === 'in' ? 'fas fa-arrow-down' : 'fas fa-arrow-up'"></i>
              </span>
            </td>
            <td
              :class="{
                positive: transaction.previousBalance >= 0,
                negative: transaction.previousBalance < 0,
              }"
            >
              {{ formatCurrency(transaction.previousBalance) }}
            </td>
            <td class="money-in">
              <span v-if="transaction.type === 'in'" class="amount-badge in">
                <i class="fas fa-plus"></i>
                {{ formatCurrency(transaction.amount) }}
              </span>
              <span v-else>-</span>
            </td>
            <td class="money-out">
              <span v-if="transaction.type === 'out'" class="amount-badge out">
                <i class="fas fa-minus"></i>
                {{ formatCurrency(transaction.amount) }}
              </span>
              <span v-else>-</span>
            </td>
            <td :class="{ positive: transaction.balance >= 0, negative: transaction.balance < 0 }">
              {{ formatCurrency(transaction.balance) }}
            </td>
            <td>
              <span class="source-badge" :class="transaction.source">
                <i
                  :class="{
                    'fas fa-shopping-cart': transaction.source === 'sell_payment',
                    'fas fa-arrow-right': transaction.source === 'transfer_received',
                    'fas fa-arrow-left': transaction.source === 'transfer_sent',
                    'fas fa-globe': transaction.source === 'external_transfer',
                  }"
                ></i>
                {{ getSourceLabel(transaction.source) }}
              </span>
            </td>
            <td class="notes">
              <span v-if="transaction.notes" class="notes-content" :title="transaction.notes">
                <i class="fas fa-comment-alt"></i>
                {{ transaction.notes }}
              </span>
              <span v-else class="no-notes">
                <i class="fas fa-minus"></i>
              </span>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="!loading && !transactionsWithBalance.length" class="no-data">
        <i class="fas fa-inbox fa-3x"></i>
        <p>No transactions found</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-transactions {
  margin-top: 24px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header {
  margin-bottom: 24px;
}

.header h3 {
  margin: 0 0 16px 0;
  color: #1e293b;
  font-size: 1.5rem;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header h3 i {
  color: #3b82f6;
}

.header-stats {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  font-size: 0.875rem;
}

.stat-item i {
  color: #64748b;
}

.stat-label {
  color: #64748b;
  font-weight: 500;
}

.stat-value {
  color: #1e293b;
  font-weight: 600;
}

.transactions-table {
  position: relative;
  overflow-x: auto;
}

.transactions-table.is-loading {
  min-height: 200px;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 12px;
  z-index: 10;
}

.loading-overlay i {
  color: #3b82f6;
}

.loading-overlay span {
  color: #4b5563;
  font-size: 0.875rem;
}

table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

th {
  background-color: #f8fafc;
  padding: 12px 16px;
  font-weight: 600;
  color: #1e293b;
  border-bottom: 2px solid #e2e8f0;
  white-space: nowrap;
}

th i {
  color: #64748b;
  margin-right: 8px;
  width: 16px;
}

td {
  padding: 12px 16px;
  border-bottom: 1px solid #e2e8f0;
  color: #1e293b;
}

tr:hover {
  background-color: #f8fafc;
}

.transaction-type {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  font-weight: bold;
  transition: transform 0.2s ease;
}

.transaction-type:hover {
  transform: scale(1.1);
}

.transaction-type.in {
  background-color: #dcfce7;
  color: #059669;
}

.transaction-type.out {
  background-color: #fee2e2;
  color: #dc2626;
}

.amount-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 6px;
  font-weight: 500;
  font-size: 0.875rem;
}

.amount-badge.in {
  background-color: #dcfce7;
  color: #059669;
}

.amount-badge.out {
  background-color: #fee2e2;
  color: #dc2626;
}

.source-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 0.875rem;
  background-color: #f1f5f9;
  color: #475569;
}

.source-badge i {
  color: #64748b;
}

.notes {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notes-content {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 0.875rem;
}

.no-notes {
  color: #9ca3af;
  font-size: 0.875rem;
}

.money-in,
.positive {
  color: #059669;
  font-weight: 600;
}

.money-out,
.negative {
  color: #dc2626;
  font-weight: 600;
}

.error-message {
  margin-bottom: 16px;
  padding: 12px;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 8px;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 8px;
  animation: slideIn 0.3s ease;
}

.no-data {
  padding: 48px;
  text-align: center;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.no-data i {
  color: #94a3b8;
}

.no-data p {
  margin: 0;
  font-size: 0.875rem;
}

@keyframes slideIn {
  from {
    transform: translateY(-10px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@media (max-width: 1024px) {
  .header-stats {
    flex-direction: column;
  }

  .stat-item {
    width: 100%;
    justify-content: space-between;
  }
}

@media (max-width: 768px) {
  .transactions-table {
    margin: 0 -20px;
  }

  td,
  th {
    padding: 12px;
  }

  .notes {
    max-width: 150px;
  }
}
</style>
