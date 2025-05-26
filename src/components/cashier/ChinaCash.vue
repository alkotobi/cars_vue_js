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
      params: [],
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
          amount: Number(tx.amount),
        })
        return acc
      }, {})

      // Process each date's transactions
      Object.entries(groupedByDate)
        .sort(([dateA], [dateB]) => dateA.localeCompare(dateB))
        .forEach(([date, dayTransactions]) => {
          // Calculate daily totals
          const dailyIn = dayTransactions
            .filter((tx) => tx.type === 'in')
            .reduce((sum, tx) => sum + tx.amount, 0)

          const dailyOut = dayTransactions
            .filter((tx) => tx.type === 'out')
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
            transactions: dayTransactions.map((tx) => ({
              createdBy: tx.created_by,
              notes: tx.notes,
              amount: tx.amount,
              type: tx.type,
            })),
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
      name: err.name,
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
    day: 'numeric',
  })
}

const formatNumber = (value) => {
  const num = Number(value)
  return !isNaN(num)
    ? num.toLocaleString('en-US', {
        style: 'currency',
        currency: 'USD',
      })
    : 'N/A'
}

// Add computed totals
const totals = computed(() => {
  if (!transactions.value.length) return { totalIn: 0, totalOut: 0, balance: 0 }

  // Get the last transaction to get final balance
  const lastDay = transactions.value[transactions.value.length - 1]
  return {
    totalIn: transactions.value.reduce((sum, day) => sum + day.in, 0),
    totalOut: transactions.value.reduce((sum, day) => sum + day.out, 0),
    balance: lastDay.newBalance,
  }
})

onMounted(() => {
  fetchTransactions()
})
</script>

<template>
  <div class="china-cash">
    <div class="header">
      <h2>
        <i class="fas fa-yen-sign"></i>
        China Cash Management
      </h2>
      <div class="summary">
        <div class="summary-item">
          <i class="fas fa-chart-line"></i>
          <div class="summary-content">
            <span class="label">Total In:</span>
            <span class="value positive">
              <i class="fas fa-arrow-up"></i>
              {{ formatNumber(totals.totalIn) }}
            </span>
          </div>
        </div>
        <div class="summary-item">
          <i class="fas fa-chart-line fa-flip-vertical"></i>
          <div class="summary-content">
            <span class="label">Total Out:</span>
            <span class="value negative">
              <i class="fas fa-arrow-down"></i>
              {{ formatNumber(totals.totalOut) }}
            </span>
          </div>
        </div>
        <div class="summary-item">
          <i class="fas fa-balance-scale"></i>
          <div class="summary-content">
            <span class="label">Current Balance:</span>
            <span
              class="value"
              :class="{ positive: totals.balance >= 0, negative: totals.balance < 0 }"
            >
              <i :class="totals.balance >= 0 ? 'fas fa-plus-circle' : 'fas fa-minus-circle'"></i>
              {{ formatNumber(totals.balance) }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Transactions Table -->
    <div class="transactions-table" :class="{ 'is-loading': loading }">
      <div v-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        <div class="error-content">
          <strong>Error</strong>
          <p>{{ error }}</p>
        </div>
      </div>

      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading transactions...</span>
      </div>

      <table v-if="!loading && transactions.length > 0">
        <thead>
          <tr>
            <th><i class="fas fa-calendar-alt"></i> Date</th>
            <th><i class="fas fa-history"></i> Previous Balance</th>
            <th><i class="fas fa-arrow-up"></i> In</th>
            <th><i class="fas fa-arrow-down"></i> Out</th>
            <th><i class="fas fa-wallet"></i> New Balance</th>
            <th><i class="fas fa-info-circle"></i> Details</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="day in transactions" :key="day.date">
            <tr
              class="day-summary"
              :class="{
                'positive-row': day.newBalance >= 0,
                'negative-row': day.newBalance < 0,
              }"
            >
              <td>
                <span class="date-badge">
                  <i class="fas fa-calendar-day"></i>
                  {{ formatDate(day.date) }}
                </span>
              </td>
              <td
                :class="{ positive: day.previousBalance >= 0, negative: day.previousBalance < 0 }"
              >
                {{ formatNumber(day.previousBalance) }}
              </td>
              <td class="positive">
                <span class="amount-badge in" v-if="day.in > 0">
                  <i class="fas fa-plus"></i>
                  {{ formatNumber(day.in) }}
                </span>
                <span class="empty-amount" v-else>-</span>
              </td>
              <td class="negative">
                <span class="amount-badge out" v-if="day.out > 0">
                  <i class="fas fa-minus"></i>
                  {{ formatNumber(day.out) }}
                </span>
                <span class="empty-amount" v-else>-</span>
              </td>
              <td :class="{ positive: day.newBalance >= 0, negative: day.newBalance < 0 }">
                {{ formatNumber(day.newBalance) }}
              </td>
              <td>
                <div v-for="tx in day.transactions" :key="tx.id" class="transaction-detail">
                  <span class="type-badge" :class="tx.type">
                    <i :class="tx.type === 'in' ? 'fas fa-arrow-right' : 'fas fa-arrow-left'"></i>
                    {{ formatNumber(tx.amount) }}
                  </span>
                  <span class="created-by">
                    <i class="fas fa-user"></i>
                    {{ tx.createdBy }}
                  </span>
                  <span v-if="tx.notes" class="notes" :title="tx.notes">
                    <i class="fas fa-comment-alt"></i>
                    {{ tx.notes }}
                  </span>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>

      <div v-else-if="!loading && !transactions.length" class="no-data">
        <i class="fas fa-inbox fa-3x"></i>
        <p>No transactions found</p>
      </div>
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
  font-size: 1.5rem;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header h2 i {
  color: #3b82f6;
}

.summary {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 24px;
  background-color: #f8fafc;
  padding: 16px;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
}

.summary-item {
  flex: 1;
  min-width: 200px;
  padding: 16px;
  border-radius: 8px;
  background-color: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: flex-start;
  gap: 12px;
  transition: transform 0.2s ease;
}

.summary-item:hover {
  transform: translateY(-1px);
}

.summary-item > i {
  color: #64748b;
  font-size: 1.25rem;
  margin-top: 4px;
}

.summary-content {
  flex: 1;
}

.summary-item .label {
  display: block;
  color: #64748b;
  margin-bottom: 4px;
  font-size: 0.875rem;
}

.summary-item .value {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.5rem;
  font-weight: 600;
}

.transactions-table {
  background-color: white;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
  position: relative;
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
}

th,
td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #1e293b;
  white-space: nowrap;
}

th i {
  color: #64748b;
  margin-right: 8px;
  width: 16px;
}

.day-summary {
  background-color: #f8fafc;
}

.day-summary:hover {
  background-color: #f1f5f9;
}

.date-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 6px;
  background-color: #e2e8f0;
  color: #475569;
  font-size: 0.875rem;
}

.transaction-detail {
  padding: 8px 0;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px dashed #e2e8f0;
}

.transaction-detail:last-child {
  border-bottom: none;
}

.type-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 6px;
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

.empty-amount {
  color: #9ca3af;
  font-size: 0.875rem;
}

.created-by {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: #64748b;
  font-size: 0.875rem;
}

.notes {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #64748b;
  font-size: 0.875rem;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  border-radius: 8px;
  margin-bottom: 16px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.error-message i {
  margin-top: 2px;
}

.error-content {
  flex: 1;
}

.error-content strong {
  display: block;
  margin-bottom: 4px;
}

.error-content p {
  margin: 0;
  font-size: 0.875rem;
}

.no-data {
  text-align: center;
  padding: 48px;
  color: #64748b;
  background-color: #f8fafc;
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

@media (max-width: 1024px) {
  .summary {
    flex-direction: column;
  }

  .summary-item {
    width: 100%;
  }
}

@media (max-width: 768px) {
  .transaction-detail {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    padding: 12px 0;
  }

  .notes {
    max-width: 100%;
  }

  td,
  th {
    padding: 12px;
  }

  .transactions-table {
    margin: 0 -20px;
    border-radius: 0;
    border-left: none;
    border-right: none;
  }
}
</style>
