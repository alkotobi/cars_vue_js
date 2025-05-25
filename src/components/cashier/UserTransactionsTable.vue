<script setup>
import { ref, watch, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  selectedUser: {
    type: Object,
    required: true
  }
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
      params: [props.selectedUser.id]
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
      params: [props.selectedUser.id]
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
      params: [props.selectedUser.id]
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
      params: [props.selectedUser.id]
    })
    console.log('External transfers result:', externalTransfersResult)

    // Reset arrays before updating with new data
    sellPayments.value = []
    receivedTransfers.value = []
    sentTransfers.value = []
    externalTransfers.value = []

    if (sellPaymentsResult?.success) {
      sellPayments.value = sellPaymentsResult.data.map(payment => ({
        ...payment,
        type: 'in',
        source: 'sell_payment'
      }))
    }

    if (receivedTransfersResult?.success) {
      receivedTransfers.value = receivedTransfersResult.data.map(transfer => ({
        ...transfer,
        type: 'in',
        source: 'transfer_received'
      }))
    }

    if (sentTransfersResult?.success) {
      sentTransfers.value = sentTransfersResult.data.map(transfer => ({
        ...transfer,
        type: 'out',
        source: 'transfer_sent'
      }))
    }

    if (externalTransfersResult?.success) {
      externalTransfers.value = externalTransfersResult.data.map(transfer => ({
        ...transfer,
        type: 'out',
        source: 'external_transfer'
      }))
    }

    console.log('Data fetching complete', {
      sellPayments: sellPayments.value,
      receivedTransfers: receivedTransfers.value,
      sentTransfers: sentTransfers.value,
      externalTransfers: externalTransfers.value
    })

  } catch (err) {
    console.error('Error fetching data:', err)
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Watch for changes in selected user
watch(() => props.selectedUser?.id, (newUserId) => {
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
}, { immediate: true })

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
    maximumFractionDigits: 2
  }).format(numValue)
}

// Combine all transactions and calculate running balance
const transactionsWithBalance = computed(() => {
  // Combine all transactions
  const allTransactions = [
    ...sellPayments.value,
    ...receivedTransfers.value,
    ...sentTransfers.value,
    ...externalTransfers.value
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
      balance: runningBalance || 0
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
      <h3>Transactions for {{ selectedUser.username }}</h3>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>

    <div class="transactions-table">
      <table v-if="!loading && transactionsWithBalance.length > 0">
        <thead>
          <tr>
            <th>Date</th>
            <th>Type</th>
            <th>Previous Balance</th>
            <th>Money In</th>
            <th>Money Out</th>
            <th>Current Balance</th>
            <th>Source</th>
            <th>Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(transaction, index) in transactionsWithBalance" :key="index">
            <td>{{ formatDate(transaction.transaction_date) }}</td>
            <td>
              <span class="transaction-type" :class="transaction.type">
                {{ transaction.type === 'in' ? '↓' : '↑' }}
              </span>
            </td>
            <td :class="{ 'positive': transaction.previousBalance >= 0, 'negative': transaction.previousBalance < 0 }">
              {{ formatCurrency(transaction.previousBalance) }}
            </td>
            <td class="money-in">
              {{ transaction.type === 'in' ? formatCurrency(transaction.amount) : '-' }}
            </td>
            <td class="money-out">
              {{ transaction.type === 'out' ? formatCurrency(transaction.amount) : '-' }}
            </td>
            <td :class="{ 'positive': transaction.balance >= 0, 'negative': transaction.balance < 0 }">
              {{ formatCurrency(transaction.balance) }}
            </td>
            <td>{{ getSourceLabel(transaction.source) }}</td>
            <td>{{ transaction.notes || '-' }}</td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="loading" class="loading">
        Loading transactions...
      </div>

      <div v-else class="no-data">
        No transactions found
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-transactions {
  margin-top: 24px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header {
  margin-bottom: 20px;
}

.header h3 {
  margin: 0;
  color: #1a1a1a;
  font-size: 1.25rem;
}

.transactions-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

th {
  background-color: #f9fafb;
  padding: 12px;
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
}

td {
  padding: 12px;
  border-bottom: 1px solid #e5e7eb;
  color: #1f2937;
}

.transaction-type {
  display: inline-block;
  width: 24px;
  height: 24px;
  line-height: 24px;
  text-align: center;
  border-radius: 50%;
  font-weight: bold;
}

.transaction-type.in {
  background-color: #d1fae5;
  color: #059669;
}

.transaction-type.out {
  background-color: #fee2e2;
  color: #dc2626;
}

.money-in {
  color: #059669;
}

.money-out {
  color: #dc2626;
}

.positive {
  color: #059669;
  font-weight: 600;
}

.negative {
  color: #dc2626;
  font-weight: 600;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
}

.loading, .no-data {
  text-align: center;
  padding: 40px;
  color: #6b7280;
  font-size: 1.1rem;
}
</style> 