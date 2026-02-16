<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useInvoiceCompanyInfo } from '../../composables/useInvoiceCompanyInfo'
import UserTransactionsTable from './UserTransactionsTable.vue'

const { callApi, getAssets } = useApi()
const { getCompanyLogoUrl } = useInvoiceCompanyInfo()
const letterHeadUrl = ref(null)
const users = ref([])
const loading = ref(false)
const error = ref(null)
const selectedUser = ref(null)
const showFilters = ref(false)
const searchTerm = ref('')
const sortField = ref('username')
const sortDirection = ref('asc')

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
    const query = isAdmin.value
      ? `
      SELECT 
        u.id,
        u.username,
        COALESCE(
          (SELECT SUM(CAST(amount_da AS DECIMAL(20,2))) 
           FROM sell_payments
           WHERE id_user = u.id
          ), 0
        ) + COALESCE(
          (SELECT SUM(CAST(amount AS DECIMAL(20,2))) 
           FROM transfers_inter
           WHERE to_user_id = u.id AND id_admin_confirm IS NOT NULL
          ), 0
        ) as total_money_in,
        COALESCE(
          (SELECT SUM(CAST(amount AS DECIMAL(20,2)))
           FROM transfers_inter
           WHERE from_user_id = u.id
          ), 0
        ) + COALESCE(
          (SELECT SUM(CAST(amount_sending_da AS DECIMAL(20,2)))
           FROM transfers
           WHERE id_user_do_transfer = u.id AND id_user_receive_transfer IS NOT NULL
          ), 0
        ) as total_money_out
      FROM users u
      ORDER BY u.username
    `
      : `
      SELECT 
        u.id,
        u.username,
        COALESCE(
          (SELECT SUM(CAST(amount_da AS DECIMAL(20,2))) 
           FROM sell_payments
           WHERE id_user = u.id
          ), 0
        ) + COALESCE(
          (SELECT SUM(CAST(amount AS DECIMAL(20,2))) 
           FROM transfers_inter
           WHERE to_user_id = u.id AND id_admin_confirm IS NOT NULL
          ), 0
        ) as total_money_in,
        COALESCE(
          (SELECT SUM(CAST(amount AS DECIMAL(20,2)))
           FROM transfers_inter
           WHERE from_user_id = u.id
          ), 0
        ) + COALESCE(
          (SELECT SUM(CAST(amount_sending_da AS DECIMAL(20,2)))
           FROM transfers
           WHERE id_user_do_transfer = u.id AND id_user_receive_transfer IS NOT NULL
          ), 0
        ) as total_money_out
      FROM users u
      WHERE u.id = ?
      ORDER BY u.username
    `

    const params = isAdmin.value ? [] : [currentUser.value?.id]
    const result = await callApi({ query, params })

    if (result.success) {
      users.value = result.data.map((user) => {
        // Ensure we're working with numbers, not strings
        const moneyIn = parseFloat(user.total_money_in) || 0
        const moneyOut = parseFloat(user.total_money_out) || 0

        return {
          ...user,
          total_money_in: moneyIn,
          total_money_out: moneyOut,
          balance: moneyIn - moneyOut,
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
    maximumFractionDigits: 2,
  }).format(numValue)
}

const totalBalance = computed(() => {
  // Ensure we're adding numbers, not strings
  return users.value.reduce((sum, user) => {
    const balance = parseFloat(user.balance) || 0
    return sum + balance
  }, 0)
})

// Filtered and sorted users
const filteredUsers = computed(() => {
  let filtered = [...users.value]

  // Search filter
  if (searchTerm.value) {
    const term = searchTerm.value.toLowerCase()
    filtered = filtered.filter((user) => user.username.toLowerCase().includes(term))
  }

  // Sort users
  filtered.sort((a, b) => {
    let aValue, bValue

    switch (sortField.value) {
      case 'username':
        aValue = a.username.toLowerCase()
        bValue = b.username.toLowerCase()
        break
      case 'total_money_in':
        aValue = parseFloat(a.total_money_in) || 0
        bValue = parseFloat(b.total_money_in) || 0
        break
      case 'total_money_out':
        aValue = parseFloat(a.total_money_out) || 0
        bValue = parseFloat(b.total_money_out) || 0
        break
      case 'balance':
        aValue = parseFloat(a.balance) || 0
        bValue = parseFloat(b.balance) || 0
        break
      default:
        aValue = a.username.toLowerCase()
        bValue = b.username.toLowerCase()
    }

    if (sortDirection.value === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })

  return filtered
})

const handleUserSelect = (user) => {
  // Only allow selection if admin or if it's the current user
  if (isAdmin.value || user.id === currentUser.value?.id) {
    selectedUser.value = selectedUser.value?.id === user.id ? null : user
  }
}

const handleSort = (field) => {
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortDirection.value = 'asc'
  }
}

const getSortIcon = (field) => {
  if (sortField.value !== field) return 'fas fa-sort'
  return sortDirection.value === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'
}

const printUsersList = async () => {
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    alert('Popup blocked. Please allow popups for this site.')
    return
  }

  // Load assets if not already loaded
  if (!letterHeadUrl.value) {
    try {
      letterHeadUrl.value = (await getCompanyLogoUrl()) || ''
    } catch (err) {
      console.error('Failed to load company logo:', err)
    }
  }

  const usersListHtml = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Users List</title>
      <style>
        @page {
          size: A4;
          margin: 15mm;
        }
        
        body {
          font-family: Arial, sans-serif;
          font-size: 11px;
          line-height: 1.3;
          color: #333;
          margin: 0;
          padding: 0;
        }
        
        .letterhead {
          text-align: center;
          margin-bottom: 20px;
        }
        
        .letterhead img {
          width: 100%;
          height: auto;
          max-height: 50px;
          object-fit: contain;
        }
        
        .report-header {
          text-align: center;
          margin-bottom: 20px;
          border-bottom: 2px solid #333;
          padding-bottom: 10px;
        }
        
        .report-title {
          font-size: 18px;
          font-weight: bold;
          margin-bottom: 5px;
        }
        
        .report-date {
          font-size: 12px;
          color: #666;
        }
        
        .summary-info {
          display: flex;
          justify-content: space-between;
          margin-bottom: 20px;
          padding: 10px;
          background-color: #f8f9fa;
          border-radius: 6px;
        }
        
        .summary-item {
          text-align: center;
        }
        
        .summary-label {
          font-size: 10px;
          color: #666;
          display: block;
          margin-bottom: 2px;
        }
        
        .summary-value {
          font-size: 14px;
          font-weight: bold;
          color: #333;
        }
        
        .users-table {
          width: 100%;
          border-collapse: collapse;
          margin-bottom: 20px;
        }
        
        .users-table th {
          background-color: #f8f9fa;
          padding: 8px 6px;
          border: 1px solid #ddd;
          font-weight: bold;
          font-size: 10px;
          text-align: left;
        }
        
        .users-table td {
          padding: 6px;
          border: 1px solid #ddd;
          font-size: 10px;
        }
        
        .users-table tr:nth-child(even) {
          background-color: #f9f9f9;
        }
        
        .amount-cell {
          text-align: right;
          font-weight: bold;
        }
        
        .money-in {
          color: #16a34a;
        }
        
        .money-out {
          color: #dc2626;
        }
        
        .positive {
          color: #16a34a;
        }
        
        .negative {
          color: #dc2626;
        }
        
        .footer {
          margin-top: 20px;
          text-align: center;
          font-size: 9px;
          color: #666;
          border-top: 1px solid #eee;
          padding-top: 10px;
        }
        
        @media print {
          body {
            -webkit-print-color-adjust: exact;
            print-color-adjust: exact;
          }
        }
      </style>
    </head>
    <body>
      <div class="letterhead">
        <img src="${letterHeadUrl.value}" alt="Company Letterhead">
      </div>
      
      <div class="report-header">
        <div class="report-title">USERS LIST</div>
        <div class="report-date">Generated on ${new Date().toLocaleDateString()}</div>
      </div>
      
      <div class="summary-info">
        <div class="summary-item">
          <span class="summary-label">Total Users</span>
          <span class="summary-value">${filteredUsers.value.length}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">System Balance</span>
          <span class="summary-value ${totalBalance.value >= 0 ? 'positive' : 'negative'}">${formatCurrency(totalBalance.value)}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Total Money In</span>
          <span class="summary-value positive">${formatCurrency(filteredUsers.value.reduce((sum, user) => sum + (parseFloat(user.total_money_in) || 0), 0))}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Total Money Out</span>
          <span class="summary-value negative">${formatCurrency(filteredUsers.value.reduce((sum, user) => sum + (parseFloat(user.total_money_out) || 0), 0))}</span>
        </div>
      </div>
      
      <table class="users-table">
        <thead>
          <tr>
            <th style="width: 8%;">#</th>
            <th style="width: 25%;">User</th>
            <th style="width: 20%;">Money In</th>
            <th style="width: 20%;">Money Out</th>
            <th style="width: 27%;">Balance</th>
          </tr>
        </thead>
        <tbody>
          ${filteredUsers.value
            .map(
              (user, index) => `
            <tr>
              <td>${index + 1}</td>
              <td>${user.username}</td>
              <td class="amount-cell money-in">${formatCurrency(user.total_money_in)}</td>
              <td class="amount-cell money-out">${formatCurrency(user.total_money_out)}</td>
              <td class="amount-cell ${user.balance >= 0 ? 'positive' : 'negative'}">${formatCurrency(user.balance)}</td>
            </tr>
          `,
            )
            .join('')}
        </tbody>
      </table>
      
      <div class="footer">
        <p>This is a computer-generated users list.</p>
        <p>Generated on ${new Date().toLocaleString()}</p>
        ${searchTerm.value ? `<p><strong>Note:</strong> This list shows filtered results for "${searchTerm.value}".</p>` : ''}
      </div>
    </body>
    </html>
  `

  printWindow.document.write(usersListHtml)
  printWindow.document.close()

  // Wait for images to load, then print
  printWindow.onload = () => {
    setTimeout(() => {
      printWindow.print()
      printWindow.close()
    }, 500)
  }
}

const formatDate = (date) => {
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

onMounted(() => {
  fetchUserBalances()
})
</script>

<template>
  <div class="money-movements" :class="{ 'is-loading': loading }">
    <div class="header" :data-print-date="formatDate(new Date())">
      <div class="header-content">
        <h2>
          <i class="fas fa-money-bill-wave pulse"></i>
          Money Movements
        </h2>
        <button
          class="refresh-btn"
          @click="fetchUserBalances"
          :disabled="loading"
          :class="{ 'is-loading': loading }"
        >
          <i class="fas fa-sync-alt"></i>
          Refresh
        </button>
      </div>

      <div class="summary" v-if="isAdmin">
        <div
          class="summary-item"
          :class="{ 'positive-card': totalBalance >= 0, 'negative-card': totalBalance < 0 }"
        >
          <i class="fas fa-balance-scale-right"></i>
          <div class="summary-content">
            <span class="label">Total System Balance</span>
            <span
              class="value"
              :class="{ positive: totalBalance >= 0, negative: totalBalance < 0 }"
            >
              <i
                :class="totalBalance >= 0 ? 'fas fa-arrow-up bounce' : 'fas fa-arrow-down bounce'"
              ></i>
              {{ formatCurrency(totalBalance) }}
            </span>
          </div>
        </div>
        <div class="summary-item">
          <i class="fas fa-users-cog"></i>
          <div class="summary-content">
            <span class="label">Active Users</span>
            <span class="value">
              <i class="fas fa-user-check"></i>
              {{ users.length }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-triangle fa-lg shake"></i>
      <div class="error-content">
        <strong>Error Occurred</strong>
        <p>{{ error }}</p>
        <button class="retry-btn" @click="fetchUserBalances">
          <i class="fas fa-redo"></i>
          Try Again
        </button>
      </div>
    </div>

    <div class="movements-table-container">
      <div class="table-overlay" v-if="loading">
        <div class="loading-spinner">
          <i class="fas fa-circle-notch fa-spin fa-2x"></i>
          <span>Loading Data...</span>
        </div>
      </div>

      <!-- Search and Controls -->
      <div class="controls-section" v-if="!loading && users.length > 0">
        <div class="search-controls">
          <div class="search-box">
            <i class="fas fa-search"></i>
            <input
              v-model="searchTerm"
              type="text"
              placeholder="Search by username..."
              class="search-input"
            />
          </div>
          <button @click="printUsersList" class="print-btn" :disabled="filteredUsers.length === 0">
            <i class="fas fa-print"></i>
            Print List
          </button>
        </div>
      </div>

      <div class="table-wrapper" v-if="!loading && users.length > 0">
        <table>
          <thead>
            <tr>
              <th @click="handleSort('username')" class="sortable">
                <i class="fas fa-user-tag"></i> User
                <i :class="getSortIcon('username')" class="sort-icon"></i>
              </th>
              <th @click="handleSort('total_money_in')" class="sortable">
                <i class="fas fa-arrow-circle-down"></i> Money In
                <i :class="getSortIcon('total_money_in')" class="sort-icon"></i>
              </th>
              <th @click="handleSort('total_money_out')" class="sortable">
                <i class="fas fa-arrow-circle-up"></i> Money Out
                <i :class="getSortIcon('total_money_out')" class="sort-icon"></i>
              </th>
              <th @click="handleSort('balance')" class="sortable">
                <i class="fas fa-wallet"></i> Balance
                <i :class="getSortIcon('balance')" class="sort-icon"></i>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="user in filteredUsers"
              :key="user.id"
              @click="handleUserSelect(user)"
              :class="{
                'row-selected': selectedUser?.id === user.id,
                'row-clickable': isAdmin || user.id === currentUser?.id,
                'row-disabled': !isAdmin && user.id !== currentUser?.id,
              }"
            >
              <td>
                <div class="user-cell">
                  <i class="fas fa-user-circle"></i>
                  <span>{{ user.username }}</span>
                  <i
                    v-if="user.id === currentUser?.id"
                    class="fas fa-star current-user-star"
                    title="Current User"
                  ></i>
                </div>
              </td>
              <td class="money-in">
                <div class="amount-cell">
                  <i class="fas fa-plus-circle pulse-success"></i>
                  <span>{{ formatCurrency(user.total_money_in) }}</span>
                </div>
              </td>
              <td class="money-out">
                <div class="amount-cell">
                  <i class="fas fa-minus-circle pulse-danger"></i>
                  <span>{{ formatCurrency(user.total_money_out) }}</span>
                </div>
              </td>
              <td :class="{ positive: user.balance >= 0, negative: user.balance < 0 }">
                <div class="amount-cell">
                  <i
                    :class="
                      user.balance >= 0
                        ? 'fas fa-arrow-up bounce-soft'
                        : 'fas fa-arrow-down bounce-soft'
                    "
                  ></i>
                  <span>{{ formatCurrency(user.balance) }}</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-else-if="!loading && !users.length" class="empty-state">
        <i class="fas fa-users-slash fa-3x pulse"></i>
        <h3>No Users Found</h3>
        <p>There are currently no users in the system.</p>
        <button class="retry-btn" @click="fetchUserBalances">
          <i class="fas fa-sync-alt"></i>
          Refresh Data
        </button>
      </div>
    </div>

    <!-- User Transactions Table -->
    <UserTransactionsTable
      v-if="selectedUser"
      :selectedUser="selectedUser"
      class="transactions-table-wrapper"
    />
  </div>
</template>

<style scoped>
.money-movements {
  padding: 24px;
  background: white;
  border-radius: 12px;
  box-shadow:
    0 4px 6px -1px rgba(0, 0, 0, 0.1),
    0 2px 4px -1px rgba(0, 0, 0, 0.06);
  transition: all 0.3s ease;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h2 {
  margin: 0;
  color: #1e293b;
  font-size: 1.75rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header h2 i {
  color: #3b82f6;
  font-size: 1.5em;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  background: #f1f5f9;
  color: #475569;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: #e2e8f0;
  color: #1e293b;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-btn.is-loading i {
  animation: spin 1s linear infinite;
}

.summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.summary-item {
  background: #f8fafc;
  padding: 20px;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  transition: all 0.3s ease;
}

.summary-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.summary-item.positive-card {
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.summary-item.negative-card {
  background: #fef2f2;
  border-color: #fecaca;
}

.summary-item > i {
  color: #64748b;
  font-size: 1.5rem;
  margin-top: 4px;
}

.summary-content {
  flex: 1;
}

.summary-item .label {
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 500;
  display: block;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.summary-item .value {
  color: #1e293b;
  font-size: 1.5rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.movements-table-container {
  position: relative;
  background: white;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

.table-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10;
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #3b82f6;
}

.loading-spinner span {
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 500;
}

.table-wrapper {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

th {
  background: #f8fafc;
  padding: 16px;
  font-weight: 600;
  color: #475569;
  border-bottom: 2px solid #e2e8f0;
  text-transform: uppercase;
  font-size: 0.75rem;
  letter-spacing: 0.05em;
}

th i {
  color: #64748b;
  margin-right: 8px;
  width: 16px;
}

td {
  padding: 16px;
  border-bottom: 1px solid #e2e8f0;
  color: #1e293b;
  font-size: 0.875rem;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-cell i {
  color: #64748b;
  font-size: 1.25rem;
}

.current-user-star {
  color: #eab308 !important;
  margin-left: auto;
}

.amount-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.money-in {
  color: #059669;
}

.money-out {
  color: #dc2626;
}

.positive {
  color: #059669;
}

.negative {
  color: #dc2626;
}

.error-message {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
  padding: 16px;
  border-radius: 12px;
  margin-bottom: 24px;
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.error-content {
  flex: 1;
}

.error-content strong {
  display: block;
  margin-bottom: 4px;
  font-size: 0.875rem;
}

.error-content p {
  margin: 0 0 12px 0;
  color: #ef4444;
  font-size: 0.875rem;
}

.retry-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 1px solid currentColor;
  border-radius: 6px;
  background: transparent;
  color: inherit;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.retry-btn:hover {
  background: rgba(220, 38, 38, 0.1);
}

.empty-state {
  text-align: center;
  padding: 48px 24px;
  color: #64748b;
}

.empty-state i {
  color: #94a3b8;
  margin-bottom: 16px;
}

.empty-state h3 {
  margin: 0 0 8px 0;
  color: #475569;
  font-size: 1.25rem;
}

.empty-state p {
  margin: 0 0 20px 0;
  color: #64748b;
  font-size: 0.875rem;
}

.row-clickable {
  cursor: pointer;
  transition: all 0.2s ease;
}

.row-clickable:hover {
  background: #f8fafc;
}

.row-selected {
  background: #f1f5f9;
}

.row-disabled {
  opacity: 0.6;
  cursor: not-allowed;
  pointer-events: none;
}

/* Animations */
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
  100% {
    transform: scale(1);
  }
}

@keyframes shake {
  0%,
  100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-2px);
  }
  75% {
    transform: translateX(2px);
  }
}

@keyframes bounce {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-3px);
  }
}

@keyframes bounce-soft {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-2px);
  }
}

.pulse {
  animation: pulse 2s infinite;
}

.shake {
  animation: shake 0.5s infinite;
}

.bounce {
  animation: bounce 2s infinite;
}

.bounce-soft {
  animation: bounce-soft 2s infinite;
}

.pulse-success {
  color: #059669;
  animation: pulse 2s infinite;
}

.pulse-danger {
  color: #dc2626;
  animation: pulse 2s infinite;
}

/* Responsive Design */
@media (max-width: 1024px) {
  .money-movements {
    padding: 20px;
  }

  .summary {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .refresh-btn {
    width: 100%;
    justify-content: center;
  }

  .movements-table-container {
    margin: 0 -20px;
    border-radius: 0;
    border-left: none;
    border-right: none;
  }

  .table-wrapper {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  td,
  th {
    padding: 12px;
    font-size: 0.813rem;
  }

  .user-cell {
    gap: 8px;
  }

  .amount-cell {
    gap: 6px;
  }

  .search-controls {
    flex-direction: column;
    align-items: stretch;
  }

  .search-box {
    min-width: auto;
  }
}

/* Search and Controls */
.controls-section {
  margin-bottom: 20px;
}

.search-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 15px;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  flex: 1;
  min-width: 200px;
}

.search-box i {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
  font-size: 14px;
}

.search-input {
  width: 100%;
  padding: 10px 12px 10px 35px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.3s ease;
  background-color: #f9fafb;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  background-color: white;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.print-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

.print-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #1d4ed8, #1e40af);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
}

.print-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* Sortable Headers */
.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  transition: background-color 0.2s ease;
}

.sortable:hover {
  background-color: #f3f4f6;
}

.sort-icon {
  margin-left: 8px;
  font-size: 12px;
  opacity: 0.6;
  transition: opacity 0.2s ease;
}

.sortable:hover .sort-icon {
  opacity: 1;
}

/* Print Styles - Keep existing print styles */
</style>
