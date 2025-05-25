<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi } = useApi()
const transactions = ref([])
const loading = ref(true)
const error = ref(null)

const fetchTransactions = async () => {
  loading.value = true
  error.value = null
  try {
    const result = await callApi({
      query: `
        SELECT 
          CONCAT('T', t.id) as id,
          t.date_receive as date,
          t.amount_received_usd as amount,
          'in' as type,
          COALESCE(t.notes, '') as notes,
          COALESCE(u.username, 'Unknown') as created_by,
          'Transfer' as source
        FROM transfers t
        LEFT JOIN users u ON t.id_user_do_transfer = u.id
        WHERE t.id_user_receive_transfer IS NOT NULL
        AND t.date_receive IS NOT NULL
        
        UNION ALL
        
        SELECT 
          CONCAT('B', bp.id) as id,
          bp.date_payment as date,
          bp.amount as amount,
          'out' as type,
          CONCAT(COALESCE(bp.notes, ''), ' (Bill #', COALESCE(bb.id, ''), ')') as notes,
          COALESCE(u.username, 'Unknown') as created_by,
          'Buy Payment' as source
        FROM buy_payments bp
        LEFT JOIN users u ON bp.id_user = u.id
        LEFT JOIN buy_bill bb ON bp.id_buy_bill = bb.id
        WHERE bp.date_payment IS NOT NULL
        ORDER BY date ASC
      `.trim(),
      params: []
    })

    if (result.success) {
      // Process transactions and calculate running balances
      const processedTransactions = []
      let runningBalance = 0

      // Group transactions by date
      const groupedByDate = result.data.reduce((acc, tx) => {
        const date = new Date(tx.date).toISOString().split('T')[0]
        if (!acc[date]) {
          acc[date] = []
        }
        acc[date].push({
          ...tx,
          amount: Number(tx.amount)
        })
        return acc
      }, {})

      // Process each date's transactions
      Object.entries(groupedByDate)
        .sort(([dateA], [dateB]) => dateA.localeCompare(dateB))
        .forEach(([date, dayTransactions]) => {
          // Calculate daily totals
          const dailyIn = dayTransactions
            .filter(tx => tx.type === 'in')
            .reduce((sum, tx) => sum + tx.amount, 0)
          
          const dailyOut = dayTransactions
            .filter(tx => tx.type === 'out')
            .reduce((sum, tx) => sum + tx.amount, 0)

          // Calculate new balance
          const previousBalance = runningBalance
          const newBalance = previousBalance + dailyIn - dailyOut
          runningBalance = newBalance

          // Create daily summary
          processedTransactions.push({
            date,
            previousBalance,
            in: dailyIn,
            out: dailyOut,
            newBalance,
            transactions: dayTransactions.map(tx => ({
              createdBy: tx.created_by,
              notes: tx.notes,
              amount: tx.amount,
              type: tx.type
            }))
          })
        })

      transactions.value = processedTransactions
    } else {
      throw new Error(`Failed to fetch transactions: ${result.error || 'Unknown error'}`)
    }
  } catch (err) {
    console.error('Full error details:', {
      message: err.message,
      stack: err.stack,
      name: err.name
    })
    error.value = `Error fetching transactions: ${err.message || err}. Please check the console for more details.`
  } finally {
    loading.value = false
  }
}

// Format helpers
const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

const formatNumber = (value) => {
  const num = Number(value)
  return !isNaN(num) ? num.toLocaleString('en-US', {
    style: 'currency',
    currency: 'USD'
  }) : 'N/A'
}

// Add computed totals
const totals = computed(() => {
  if (!transactions.value.length) return { totalIn: 0, totalOut: 0, balance: 0 }
  
  // Get the last transaction to get final balance
  const lastDay = transactions.value[transactions.value.length - 1]
  return {
    totalIn: transactions.value.reduce((sum, day) => sum + day.in, 0),
    totalOut: transactions.value.reduce((sum, day) => sum + day.out, 0),
    balance: lastDay.newBalance
  }
})

onMounted(() => {
  fetchTransactions()
})
</script>

<template>
  <div class="china-cash">
    <div class="header">
      <h2>China Cash Management</h2>
      <div class="summary">
        <div class="summary-item">
          <span class="label">Total In:</span>
          <span class="value positive">{{ formatNumber(totals.totalIn) }}</span>
        </div>
        <div class="summary-item">
          <span class="label">Total Out:</span>
          <span class="value negative">{{ formatNumber(totals.totalOut) }}</span>
        </div>
        <div class="summary-item">
          <span class="label">Current Balance:</span>
          <span class="value" :class="{ 'positive': totals.balance >= 0, 'negative': totals.balance < 0 }">
            {{ formatNumber(totals.balance) }}
          </span>
        </div>
      </div>
    </div>

    <!-- Transactions Table -->
    <div class="transactions-table">
      <div v-if="error" class="error-message">{{ error }}</div>
      
      <table v-if="!loading && transactions.length > 0">
        <thead>
          <tr>
            <th>Date</th>
            <th>Previous Balance</th>
            <th>In</th>
            <th>Out</th>
            <th>New Balance</th>
            <th>Details</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="day in transactions" :key="day.date">
            <tr class="day-summary">
              <td>{{ formatDate(day.date) }}</td>
              <td :class="{ 'positive': day.previousBalance >= 0, 'negative': day.previousBalance < 0 }">
                {{ formatNumber(day.previousBalance) }}
              </td>
              <td class="positive">{{ formatNumber(day.in) }}</td>
              <td class="negative">{{ formatNumber(day.out) }}</td>
              <td :class="{ 'positive': day.newBalance >= 0, 'negative': day.newBalance < 0 }">
                {{ formatNumber(day.newBalance) }}
              </td>
              <td>
                <div v-for="tx in day.transactions" :key="tx.id" class="transaction-detail">
                  <span class="type-badge" :class="tx.type">
                    {{ formatNumber(tx.amount) }}
                  </span>
                  <span class="created-by">{{ tx.createdBy }}</span>
                  <span class="notes">{{ tx.notes }}</span>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
      <div v-else-if="loading" class="loading">Loading...</div>
      <div v-else class="no-data">No transactions found</div>
    </div>
  </div>
</template>

<style scoped>
.china-cash {
  padding: 20px;
}

.header {
  margin-bottom: 24px;
}

.header h2 {
  margin: 0 0 16px 0;
  color: #1e293b;
}

.summary {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 24px;
  background-color: #f8fafc;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.summary-item {
  flex: 1;
  min-width: 200px;
  padding: 12px;
  border-radius: 8px;
  background-color: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.summary-item .label {
  display: block;
  color: #64748b;
  margin-bottom: 4px;
  font-size: 0.875rem;
}

.summary-item .value {
  display: block;
  font-size: 1.5rem;
  font-weight: 600;
}

.transactions-table {
  background-color: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

th {
  background-color: #f8fafc;
  font-weight: 500;
  color: #475569;
}

.day-summary {
  background-color: #f8fafc;
}

.transaction-detail {
  padding: 4px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
}

.type-badge.in {
  background-color: #dcfce7;
  color: #166534;
}

.type-badge.out {
  background-color: #fee2e2;
  color: #991b1b;
}

.created-by {
  font-weight: 500;
  color: #64748b;
}

.notes {
  color: #64748b;
  font-size: 0.875rem;
}

.positive {
  color: #16a34a;
}

.negative {
  color: #dc2626;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #64748b;
}

.no-data {
  text-align: center;
  padding: 40px;
  color: #64748b;
  background-color: #f8fafc;
}

@media (max-width: 768px) {
  .transaction-detail {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}
</style>