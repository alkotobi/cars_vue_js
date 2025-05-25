<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import UserTransactionsTable from './UserTransactionsTable.vue'

const { callApi } = useApi()
const users = ref([])
const loading = ref(false)
const error = ref(null)
const selectedUser = ref(null)

// Get current user from localStorage
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

const fetchUserBalances = async () => {
  loading.value = true
  error.value = null
  
  try {
    // If not admin, only fetch current user's data
    const query = isAdmin.value ? `
      SELECT 
        u.id,
        u.username,
        COALESCE(
          (SELECT SUM(CAST(total_in AS DECIMAL(20,2))) 
           FROM (
             SELECT CAST(amount_da AS DECIMAL(20,2)) as total_in
             FROM sell_payments
             WHERE id_user = u.id
             UNION ALL
             SELECT CAST(amount AS DECIMAL(20,2)) as total_in
             FROM transfers_inter
             WHERE to_user_id = u.id AND id_admin_confirm IS NOT NULL
           ) as money_in
          ), 0
        ) as total_money_in,
        COALESCE(
          (SELECT SUM(CAST(total_out AS DECIMAL(20,2)))
           FROM (
             SELECT CAST(amount AS DECIMAL(20,2)) as total_out
             FROM transfers_inter
             WHERE from_user_id = u.id
             UNION ALL
             SELECT CAST(amount_sending_da AS DECIMAL(20,2)) as total_out
             FROM transfers
             WHERE id_user_do_transfer = u.id AND id_user_receive_transfer IS NOT NULL
           ) as money_out
          ), 0
        ) as total_money_out
      FROM users u
      ORDER BY u.username
    ` : `
      SELECT 
        u.id,
        u.username,
        COALESCE(
          (SELECT SUM(CAST(total_in AS DECIMAL(20,2))) 
           FROM (
             SELECT CAST(amount_da AS DECIMAL(20,2)) as total_in
             FROM sell_payments
             WHERE id_user = u.id
             UNION ALL
             SELECT CAST(amount AS DECIMAL(20,2)) as total_in
             FROM transfers_inter
             WHERE to_user_id = u.id AND id_admin_confirm IS NOT NULL
           ) as money_in
          ), 0
        ) as total_money_in,
        COALESCE(
          (SELECT SUM(CAST(total_out AS DECIMAL(20,2)))
           FROM (
             SELECT CAST(amount AS DECIMAL(20,2)) as total_out
             FROM transfers_inter
             WHERE from_user_id = u.id
             UNION ALL
             SELECT CAST(amount_sending_da AS DECIMAL(20,2)) as total_out
             FROM transfers
             WHERE id_user_do_transfer = u.id AND id_user_receive_transfer IS NOT NULL
           ) as money_out
          ), 0
        ) as total_money_out
      FROM users u
      WHERE u.id = ?
      ORDER BY u.username
    `

    const params = isAdmin.value ? [] : [currentUser.value?.id]
    const result = await callApi({ query, params })
    
    if (result.success) {
      users.value = result.data.map(user => {
        // Ensure we're working with numbers, not strings
        const moneyIn = parseFloat(user.total_money_in) || 0
        const moneyOut = parseFloat(user.total_money_out) || 0
        
        return {
          ...user,
          total_money_in: moneyIn,
          total_money_out: moneyOut,
          balance: moneyIn - moneyOut
        }
      })

      // If not admin, auto-select the current user
      if (!isAdmin.value && users.value.length > 0) {
        selectedUser.value = users.value[0]
      }
    } else {
      error.value = result.error || 'Failed to fetch user balances'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
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

const totalBalance = computed(() => {
  // Ensure we're adding numbers, not strings
  return users.value.reduce((sum, user) => {
    const balance = parseFloat(user.balance) || 0
    return sum + balance
  }, 0)
})

const handleUserSelect = (user) => {
  // Only allow selection if admin or if it's the current user
  if (isAdmin.value || user.id === currentUser.value?.id) {
    selectedUser.value = selectedUser.value?.id === user.id ? null : user
  }
}

const formatDate = (date) => {
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

onMounted(() => {
  fetchUserBalances()
})
</script>

<template>
  <div class="money-movements">
    <div class="header" :data-print-date="formatDate(new Date())">
      <h2>Money Movements</h2>
      <div class="summary" v-if="isAdmin">
        <div class="summary-item">
          <span class="label">Total System Balance:</span>
          <span class="value" :class="{ 'positive': totalBalance >= 0, 'negative': totalBalance < 0 }">
            {{ formatCurrency(totalBalance) }}
          </span>
        </div>
        <div class="summary-item">
          <span class="label">Total Users:</span>
          <span class="value">{{ users.length }}</span>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>

    <div class="movements-table">
      <table v-if="!loading && users.length > 0">
        <thead>
          <tr>
            <th>User</th>
            <th>Money In</th>
            <th>Money Out</th>
            <th>Balance</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="user in users" 
            :key="user.id"
            @click="handleUserSelect(user)"
            :class="{ 
              'selected': selectedUser?.id === user.id,
              'clickable': isAdmin || user.id === currentUser?.id
            }"
          >
            <td>{{ user.username }}</td>
            <td class="money-in">{{ formatCurrency(user.total_money_in) }}</td>
            <td class="money-out">{{ formatCurrency(user.total_money_out) }}</td>
            <td :class="{ 'positive': user.balance >= 0, 'negative': user.balance < 0 }">
              {{ formatCurrency(user.balance) }}
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="loading" class="loading">
        Loading user balances...
      </div>

      <div v-else class="no-data">
        No users found
      </div>
    </div>

    <!-- User Transactions Table -->
    <UserTransactionsTable 
      v-if="selectedUser"
      :selectedUser="selectedUser"
    />
  </div>
</template>

<style scoped>
.money-movements {
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header {
  margin-bottom: 24px;
}

.header h2 {
  margin: 0 0 16px 0;
  color: #1a1a1a;
  font-size: 1.5rem;
}

.summary {
  display: flex;
  gap: 24px;
  margin-bottom: 20px;
}

.summary-item {
  background-color: #f3f4f6;
  padding: 12px 20px;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-item .label {
  color: #6b7280;
  font-size: 0.875rem;
}

.summary-item .value {
  color: #1f2937;
  font-size: 1.25rem;
  font-weight: 600;
}

.movements-table {
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

tr {
  transition: background-color 0.2s;
}

tr.clickable {
  cursor: pointer;
}

tr.clickable:hover {
  background-color: #f3f4f6;
}

tr.selected {
  background-color: #e5e7eb;
}

tr:not(.clickable) {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>

<style>
@media print {
  /* Hide only specific UI elements */
  nav,
  aside,
  .sidebar,
  .toolbar,
  button:not(.print-button),
  .btn:not(.print-button),
  .actions,
  .action-buttons {
    display: none !important;
  }

  /* Show the main content */
  .money-movements {
    display: block !important;
    visibility: visible !important;
    padding: 0 !important;
    margin: 0 !important;
    box-shadow: none !important;
    background: white !important;
    width: 100% !important;
    position: absolute !important;
    left: 0 !important;
    top: 0 !important;
  }

  /* Style both tables containers */
  .movements-table,
  .user-transactions {
    display: block !important;
    visibility: visible !important;
    width: 100% !important;
    margin: 0 !important;
    overflow: visible !important;
    page-break-inside: avoid !important;
  }

  /* Add section title for transactions detail */
  .user-transactions::before {
    content: "Transaction Details" !important;
    display: block !important;
    font-size: 16pt !important;
    font-weight: bold !important;
    margin: 20px 0 !important;
    text-align: center !important;
    border-top: 2px solid #000 !important;
    padding-top: 20px !important;
  }

  /* Style all tables */
  table {
    display: table !important;
    visibility: visible !important;
    width: 100% !important;
    border-collapse: collapse !important;
    font-size: 11pt !important;
    margin-bottom: 20px !important;
    page-break-inside: auto !important;
  }

  thead {
    display: table-header-group !important;
  }

  tbody {
    display: table-row-group !important;
  }

  tr {
    display: table-row !important;
    page-break-inside: avoid !important;
  }

  th, td {
    display: table-cell !important;
    padding: 8px !important;
    border: 1px solid #000 !important;
    text-align: left !important;
  }

  /* Keep colors for money in/out but make them printer-friendly */
  .money-in {
    color: #006400 !important; /* Dark green */
  }

  .money-out {
    color: #8B0000 !important; /* Dark red */
  }

  .positive {
    color: #006400 !important;
  }

  .negative {
    color: #8B0000 !important;
  }

  /* Format header for print */
  .header {
    display: block !important;
    text-align: center !important;
    margin-bottom: 20px !important;
    padding: 0 !important;
  }

  .header h2, .header h3 {
    display: block !important;
    font-size: 18pt !important;
    margin-bottom: 10px !important;
  }

  .summary {
    display: flex !important;
    justify-content: space-around !important;
    margin-bottom: 20px !important;
    border-bottom: 2px solid #000 !important;
    padding-bottom: 10px !important;
  }

  .summary-item {
    display: block !important;
    background: none !important;
    padding: 5px !important;
  }

  /* Add date to print */
  .header::after {
    content: "Printed on " attr(data-print-date) !important;
    display: block !important;
    font-size: 10pt !important;
    margin-top: 5px !important;
    color: #666 !important;
  }

  /* Remove interactive styles */
  tr.clickable, tr:not(.clickable) {
    opacity: 1 !important;
    cursor: default !important;
  }

  tr:hover, tr.selected {
    background: none !important;
  }

  /* Set A4 page margins */
  @page {
    size: A4;
    margin: 1.5cm;
  }

  /* Ensure user transactions table starts on new page */
  .user-transactions {
    page-break-before: always !important;
    display: block !important;
  }

  /* Style transaction type indicators */
  .transaction-type {
    font-weight: bold !important;
    padding: 2px 6px !important;
    border-radius: 0 !important;
    background: none !important;
  }

  .transaction-type.in {
    color: #006400 !important;
  }

  .transaction-type.out {
    color: #8B0000 !important;
  }
}
</style> 