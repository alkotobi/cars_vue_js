<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useInvoiceCompanyInfo } from '../../composables/useInvoiceCompanyInfo'

const { callApi, getAssets } = useApi()
const { getCompanyLogoUrl } = useInvoiceCompanyInfo()
const letterHeadUrl = ref(null)
const expenses = ref([])
const filteredExpenses = ref([])
const loading = ref(false)
const error = ref(null)
const showAddForm = ref(false)
const showFilters = ref(false)
const newExpense = ref({
  amount: '',
  date_transfer: new Date().toISOString().split('T')[0],
  notes: '',
})

// Filters
const filters = ref({
  dateFrom: '',
  dateTo: '',
  amountFrom: '',
  amountTo: '',
  notesSearch: '',
  userId: '',
})

// Users list for admin filter
const users = ref([])

// Get current user
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

// Check if user is admin
const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1 || currentUser.value?.role === 'admin'
})

const fetchUsers = async () => {
  if (!isAdmin.value) return

  try {
    const result = await callApi({
      query: 'SELECT id, username FROM users ORDER BY username',
      params: [],
    })

    if (result.success) {
      users.value = result.data
    }
  } catch (err) {
    console.error('Failed to fetch users:', err)
  }
}

const fetchExpenses = async () => {
  loading.value = true
  error.value = null

  try {
    let query, params

    if (isAdmin.value) {
      // Admin can see all expenses
      query = `
        SELECT 
          ti.id,
          ti.amount,
          ti.date_transfer,
          ti.notes,
          ti.date_received,
          ti.from_user_id,
          u.username
        FROM transfers_inter ti
        LEFT JOIN users u ON ti.from_user_id = u.id
        WHERE ti.from_user_id = ti.to_user_id
        ORDER BY ti.date_transfer DESC
      `
      params = []
    } else {
      // Non-admin can only see their own expenses
      query = `
        SELECT 
          ti.id,
          ti.amount,
          ti.date_transfer,
          ti.notes,
          ti.date_received,
          ti.from_user_id,
          u.username
        FROM transfers_inter ti
        LEFT JOIN users u ON ti.from_user_id = u.id
        WHERE ti.from_user_id = ? AND ti.to_user_id = ?
        ORDER BY ti.date_transfer DESC
      `
      params = [currentUser.value?.id, currentUser.value?.id]
    }

    const result = await callApi({
      query,
      params,
    })

    if (result.success) {
      expenses.value = result.data
      applyFilters()
    } else {
      error.value = result.error || 'Failed to fetch expenses'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const addExpense = async () => {
  if (!newExpense.value.amount || !newExpense.value.date_transfer) {
    error.value = 'Amount and date are required'
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        INSERT INTO transfers_inter (amount, date_transfer, from_user_id, to_user_id, notes)
        VALUES (?, ?, ?, ?, ?)
      `,
      params: [
        newExpense.value.amount,
        newExpense.value.date_transfer,
        currentUser.value?.id,
        currentUser.value?.id,
        newExpense.value.notes || null,
      ],
    })

    if (result.success) {
      await fetchExpenses()
      showAddForm.value = false
      newExpense.value = {
        amount: '',
        date_transfer: new Date().toISOString().split('T')[0],
        notes: '',
      }
    } else {
      error.value = result.error || 'Failed to add expense'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const deleteExpense = async (expenseId) => {
  if (!confirm('Are you sure you want to delete this expense?')) {
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: 'DELETE FROM transfers_inter WHERE id = ? AND from_user_id = ? AND to_user_id = ?',
      params: [expenseId, currentUser.value?.id, currentUser.value?.id],
    })

    if (result.success) {
      await fetchExpenses()
    } else {
      error.value = result.error || 'Failed to delete expense'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const printExpensesList = async () => {
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    alert('Popup blocked. Please allow popups for this site.')
    return
  }

  // Load company logo for header (same as xlsx invoice)
  if (!letterHeadUrl.value) {
    try {
      letterHeadUrl.value = (await getCompanyLogoUrl()) || ''
    } catch (err) {
      console.error('Failed to load company logo:', err)
    }
  }

  // Calculate total for print
  const printTotal = filteredExpenses.value.reduce((sum, expense) => {
    return sum + (parseFloat(expense.amount) || 0)
  }, 0)

  // Get filtered expenses count
  const printCount = filteredExpenses.value.length

  const expensesListHtml = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Expenses List</title>
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
        
        .expenses-table {
          width: 100%;
          border-collapse: collapse;
          margin-bottom: 20px;
        }
        
        .expenses-table th {
          background-color: #f8f9fa;
          padding: 8px 6px;
          border: 1px solid #ddd;
          font-weight: bold;
          font-size: 10px;
          text-align: left;
        }
        
        .expenses-table td {
          padding: 6px;
          border: 1px solid #ddd;
          font-size: 10px;
        }
        
        .expenses-table tr:nth-child(even) {
          background-color: #f9f9f9;
        }
        
        .amount-cell {
          text-align: right;
          font-weight: bold;
          color: #dc2626;
        }
        
        .notes-cell {
          max-width: 150px;
          word-wrap: break-word;
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
        <div class="report-title">EXPENSES LIST</div>
        <div class="report-date">Generated on ${new Date().toLocaleDateString()}</div>
      </div>
      
      <div class="summary-info">
        <div class="summary-item">
          <span class="summary-label">Total Expenses</span>
          <span class="summary-value">${formatCurrency(printTotal)}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Number of Expenses</span>
          <span class="summary-value">${printCount}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">User</span>
          <span class="summary-value">${currentUser.value?.username || 'N/A'}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Date Range</span>
          <span class="summary-value">
            ${filters.value.dateFrom ? formatDate(filters.value.dateFrom) : 'All'} - 
            ${filters.value.dateTo ? formatDate(filters.value.dateTo) : 'All'}
          </span>
        </div>
      </div>
      
      <table class="expenses-table">
        <thead>
          <tr>
            <th style="width: 8%;">#</th>
            <th style="width: 15%;">Date</th>
            ${isAdmin.value ? '<th style="width: 15%;">User</th>' : ''}
            <th style="width: 20%;">Amount (DA)</th>
            <th style="width: ${isAdmin.value ? '30%' : '45%'};">Notes</th>
            <th style="width: 12%;">ID</th>
          </tr>
        </thead>
        <tbody>
          ${filteredExpenses.value
            .map(
              (expense, index) => `
            <tr>
              <td>${index + 1}</td>
              <td>${formatDate(expense.date_transfer)}</td>
              ${isAdmin.value ? `<td>${expense.username || 'Unknown'}</td>` : ''}
              <td class="amount-cell">${formatCurrency(expense.amount)}</td>
              <td class="notes-cell">${expense.notes || 'No notes'}</td>
              <td>#${expense.id}</td>
            </tr>
          `,
            )
            .join('')}
          <tr style="border-top: 2px solid #333; font-weight: bold; background-color: #f8f9fa;">
            <td colspan="${isAdmin.value ? '3' : '2'}"><strong>TOTAL</strong></td>
            <td class="amount-cell"><strong>${formatCurrency(printTotal)}</strong></td>
            <td colspan="2"></td>
          </tr>
        </tbody>
      </table>
      
      <div class="footer">
        <p>This is a computer-generated expenses list.</p>
        <p>Generated on ${new Date().toLocaleString()}</p>
        ${hasActiveFilters.value ? '<p><strong>Note:</strong> This list shows filtered results only.</p>' : ''}
      </div>
    </body>
    </html>
  `

  printWindow.document.write(expensesListHtml)
  printWindow.document.close()

  // Wait for images to load, then print
  printWindow.onload = () => {
    setTimeout(() => {
      printWindow.print()
      printWindow.close()
    }, 500)
  }
}

const printReceipt = async (expense) => {
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

  const receiptHtml = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Expense Receipt</title>
      <style>
        @page {
          size: A4;
          margin: 10mm;
        }
        
        body {
          font-family: Arial, sans-serif;
          font-size: 10px;
          line-height: 1.2;
          color: #333;
          margin: 0;
          padding: 0;
          max-height: 140mm;
        }
        
        .letterhead {
          width: 100%;
          margin-bottom: 10px;
        }
        
        .letterhead img {
          width: 100%;
          height: auto;
          max-height: 40px;
          object-fit: contain;
        }
        
        .receipt-header {
          text-align: center;
          margin-bottom: 15px;
          border-bottom: 1px solid #333;
          padding-bottom: 8px;
        }
        
        .receipt-title {
          font-size: 16px;
          font-weight: bold;
          margin-bottom: 5px;
        }
        
        .receipt-date {
          font-size: 10px;
          color: #666;
        }
        
        .expense-details {
          margin-bottom: 15px;
        }
        
        .detail-row {
          display: flex;
          justify-content: space-between;
          margin-bottom: 5px;
          padding: 3px 0;
          border-bottom: 1px solid #eee;
        }
        
        .detail-label {
          font-weight: bold;
          color: #555;
          font-size: 9px;
        }
        
        .detail-value {
          color: #333;
          font-size: 9px;
        }
        
        .amount-highlight {
          background-color: #f8f9fa;
          padding: 8px;
          border: 1px solid #dc2626;
          border-radius: 4px;
          text-align: center;
          margin: 10px 0;
        }
        
        .amount-value {
          font-size: 14px;
          font-weight: bold;
          color: #dc2626;
        }
        
        .notes-section {
          margin-top: 10px;
          padding: 8px;
          background-color: #f8f9fa;
          border-radius: 4px;
        }
        
        .notes-label {
          font-weight: bold;
          margin-bottom: 4px;
          font-size: 9px;
        }
        
        .notes-content {
          font-style: italic;
          color: #666;
          font-size: 8px;
        }
        
        .footer {
          margin-top: 15px;
          text-align: center;
          font-size: 8px;
          color: #666;
          border-top: 1px solid #eee;
          padding-top: 5px;
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
        <img src="${letterHeadUrl.value || ''}" alt="Company Letterhead">
      </div>
      
      <div class="receipt-header">
        <div class="receipt-title">EXPENSE RECEIPT</div>
        <div class="receipt-date">Generated on ${new Date().toLocaleDateString()}</div>
      </div>
      
      <div class="expense-details">
        <div class="detail-row">
          <span class="detail-label">Receipt ID:</span>
          <span class="detail-value">#${expense.id}</span>
        </div>
        
        <div class="detail-row">
          <span class="detail-label">Date:</span>
          <span class="detail-value">${formatDate(expense.date_transfer)}</span>
        </div>
        
        <div class="detail-row">
          <span class="detail-label">User:</span>
          <span class="detail-value">${currentUser.value?.username || 'N/A'}</span>
        </div>
      </div>
      
      <div class="amount-highlight">
        <div style="font-size: 14px; margin-bottom: 5px;">EXPENSE AMOUNT</div>
        <div class="amount-value">${formatCurrency(expense.amount)}</div>
      </div>
      
      ${
        expense.notes
          ? `
        <div class="notes-section">
          <div class="notes-label">Notes:</div>
          <div class="notes-content">${expense.notes}</div>
        </div>
      `
          : ''
      }
      
      <div class="footer">
        <p>This is a computer-generated receipt for personal expense tracking.</p>
        <p>Generated on ${new Date().toLocaleString()}</p>
      </div>
    </body>
    </html>
  `

  printWindow.document.write(receiptHtml)
  printWindow.document.close()

  // Wait for images to load, then print
  printWindow.onload = () => {
    setTimeout(() => {
      printWindow.print()
      printWindow.close()
    }, 500)
  }
}

const formatCurrency = (value) => {
  const numValue = parseFloat(value) || 0
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'DZD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(numValue)
}

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

const totalExpenses = computed(() => {
  return filteredExpenses.value.reduce((sum, expense) => {
    return sum + (parseFloat(expense.amount) || 0)
  }, 0)
})

const applyFilters = () => {
  let filtered = [...expenses.value]

  // Date range filter
  if (filters.value.dateFrom) {
    filtered = filtered.filter(
      (expense) => new Date(expense.date_transfer) >= new Date(filters.value.dateFrom),
    )
  }

  if (filters.value.dateTo) {
    filtered = filtered.filter(
      (expense) => new Date(expense.date_transfer) <= new Date(filters.value.dateTo),
    )
  }

  // Amount range filter
  if (filters.value.amountFrom) {
    filtered = filtered.filter(
      (expense) => parseFloat(expense.amount) >= parseFloat(filters.value.amountFrom),
    )
  }

  if (filters.value.amountTo) {
    filtered = filtered.filter(
      (expense) => parseFloat(expense.amount) <= parseFloat(filters.value.amountTo),
    )
  }

  // Notes search filter
  if (filters.value.notesSearch) {
    const searchTerm = filters.value.notesSearch.toLowerCase()
    filtered = filtered.filter(
      (expense) => expense.notes && expense.notes.toLowerCase().includes(searchTerm),
    )
  }

  // User filter (admin only)
  if (filters.value.userId && isAdmin.value) {
    filtered = filtered.filter((expense) => expense.from_user_id == filters.value.userId)
  }

  filteredExpenses.value = filtered
}

const clearFilters = () => {
  filters.value = {
    dateFrom: '',
    dateTo: '',
    amountFrom: '',
    amountTo: '',
    notesSearch: '',
    userId: '',
  }
  applyFilters()
}

const hasActiveFilters = computed(() => {
  return (
    filters.value.dateFrom ||
    filters.value.dateTo ||
    filters.value.amountFrom ||
    filters.value.amountTo ||
    filters.value.notesSearch ||
    filters.value.userId
  )
})

onMounted(() => {
  fetchExpenses()
  fetchUsers()
})
</script>

<template>
  <div class="expenses">
    <div class="header">
      <h2>
        <i class="fas fa-receipt"></i>
        Expenses Management
      </h2>
      <div class="header-actions">
        <button
          @click="printExpensesList"
          class="print-list-btn"
          :disabled="filteredExpenses.length === 0"
        >
          <i class="fas fa-print"></i>
          Print List
        </button>
        <button
          @click="showFilters = !showFilters"
          class="filter-btn"
          :class="{ active: showFilters }"
        >
          <i class="fas fa-filter"></i>
          Filters
          <span v-if="hasActiveFilters" class="filter-badge"
            >{{ filteredExpenses.length }}/{{ expenses.length }}</span
          >
        </button>
        <button
          class="refresh-btn"
          @click="fetchExpenses"
          :disabled="loading"
          :class="{ 'is-loading': loading }"
        >
          <i class="fas fa-sync-alt"></i>
          Refresh
        </button>
      </div>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-triangle"></i>
      <div class="error-content">
        <strong>Error Occurred</strong>
        <p>{{ error }}</p>
        <button class="retry-btn" @click="fetchExpenses">
          <i class="fas fa-redo"></i>
          Try Again
        </button>
      </div>
    </div>

    <!-- Filters Panel -->
    <div v-if="showFilters" class="filters-panel">
      <div class="filters-header">
        <h3>
          <i class="fas fa-filter"></i>
          Filter Expenses
        </h3>
        <button @click="clearFilters" class="clear-filters-btn" :disabled="!hasActiveFilters">
          <i class="fas fa-times"></i>
          Clear All
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>Date From</label>
          <input
            v-model="filters.dateFrom"
            type="date"
            class="filter-input"
            @change="applyFilters"
          />
        </div>

        <div class="filter-group">
          <label>Date To</label>
          <input v-model="filters.dateTo" type="date" class="filter-input" @change="applyFilters" />
        </div>

        <div class="filter-group">
          <label>Amount From (DA)</label>
          <input
            v-model.number="filters.amountFrom"
            type="number"
            step="0.01"
            min="0"
            placeholder="Min amount"
            class="filter-input"
            @input="applyFilters"
          />
        </div>

        <div class="filter-group">
          <label>Amount To (DA)</label>
          <input
            v-model.number="filters.amountTo"
            type="number"
            step="0.01"
            min="0"
            placeholder="Max amount"
            class="filter-input"
            @input="applyFilters"
          />
        </div>

        <div class="filter-group full-width">
          <label>Search in Notes</label>
          <input
            v-model="filters.notesSearch"
            type="text"
            placeholder="Search notes..."
            class="filter-input"
            @input="applyFilters"
          />
        </div>

        <div v-if="isAdmin" class="filter-group full-width">
          <label>Filter by User</label>
          <select v-model="filters.userId" class="filter-input" @change="applyFilters">
            <option value="">All Users</option>
            <option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.username }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Summary Cards -->
    <div class="summary-cards">
      <div class="summary-card">
        <i class="fas fa-receipt"></i>
        <div class="summary-content">
          <span class="label">Total Expenses</span>
          <span class="value">{{ formatCurrency(totalExpenses) }}</span>
        </div>
      </div>
      <div class="summary-card">
        <i class="fas fa-list"></i>
        <div class="summary-content">
          <span class="label">Number of Expenses</span>
          <span class="value">{{ filteredExpenses.length }}</span>
        </div>
      </div>
    </div>

    <!-- Add Expense Form -->
    <div v-if="showAddForm" class="add-form">
      <h3>
        <i class="fas fa-plus"></i>
        Add New Expense
      </h3>
      <div class="form-row">
        <div class="form-group">
          <label>Amount (DA)</label>
          <input
            v-model.number="newExpense.amount"
            type="number"
            step="0.01"
            min="0"
            placeholder="Enter amount"
            class="form-input"
            :disabled="loading"
          />
        </div>
        <div class="form-group">
          <label>Date</label>
          <input
            v-model="newExpense.date_transfer"
            type="date"
            class="form-input"
            :disabled="loading"
          />
        </div>
      </div>
      <div class="form-group">
        <label>Notes</label>
        <textarea
          v-model="newExpense.notes"
          placeholder="Enter notes (optional)"
          class="form-textarea"
          :disabled="loading"
        ></textarea>
      </div>
      <div class="form-actions">
        <button @click="addExpense" :disabled="loading" class="btn btn-primary">
          <i class="fas" :class="loading ? 'fa-spinner fa-spin' : 'fa-save'"></i>
          {{ loading ? 'Adding...' : 'Add Expense' }}
        </button>
        <button @click="showAddForm = false" class="btn btn-secondary" :disabled="loading">
          <i class="fas fa-times"></i>
          Cancel
        </button>
      </div>
    </div>

    <div class="expenses-content">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading expenses...</span>
      </div>

      <div v-else-if="filteredExpenses.length === 0" class="empty-state">
        <i class="fas fa-receipt fa-3x"></i>
        <h3 v-if="hasActiveFilters">No Expenses Match Your Filters</h3>
        <h3 v-else>No Expenses Found</h3>
        <p v-if="hasActiveFilters">Try adjusting your filters or clear them to see all expenses.</p>
        <p v-else>You haven't recorded any expenses yet.</p>
        <button v-if="!hasActiveFilters" class="add-btn" @click="showAddForm = true">
          <i class="fas fa-plus"></i>
          Add Your First Expense
        </button>
        <button v-else class="add-btn" @click="clearFilters">
          <i class="fas fa-times"></i>
          Clear Filters
        </button>
      </div>

      <div v-else class="expenses-list">
        <div class="list-header">
          <h3>Your Expenses</h3>
          <button @click="showAddForm = true" class="add-btn">
            <i class="fas fa-plus"></i>
            Add Expense
          </button>
        </div>

        <!-- Admin Table -->
        <div v-if="isAdmin" class="expenses-table">
          <table>
            <thead>
              <tr>
                <th style="width: 15%"><i class="fas fa-calendar"></i> Date</th>
                <th style="width: 20%"><i class="fas fa-user"></i> User</th>
                <th style="width: 20%"><i class="fas fa-money-bill"></i> Amount</th>
                <th style="width: 35%"><i class="fas fa-comment"></i> Notes</th>
                <th style="width: 10%"><i class="fas fa-cogs"></i> Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="expense in filteredExpenses" :key="expense.id">
                <td>{{ formatDate(expense.date_transfer) }}</td>
                <td class="user-cell">
                  <i class="fas fa-user"></i>
                  {{ expense.username || 'Unknown' }}
                </td>
                <td class="amount-cell">
                  <i class="fas fa-minus-circle"></i>
                  {{ formatCurrency(expense.amount) }}
                </td>
                <td class="notes-cell">
                  <span v-if="expense.notes">{{ expense.notes }}</span>
                  <span v-else class="no-notes">No notes</span>
                </td>
                <td class="actions-cell">
                  <div class="action-buttons">
                    <button
                      @click="printReceipt(expense)"
                      class="btn btn-print"
                      title="Print receipt"
                    >
                      <i class="fas fa-print"></i>
                    </button>
                    <button
                      @click="deleteExpense(expense.id)"
                      class="btn btn-danger"
                      :disabled="loading"
                      title="Delete expense"
                    >
                      <i class="fas fa-trash"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Non-Admin Table -->
        <div v-else class="expenses-table">
          <table>
            <thead>
              <tr>
                <th style="width: 20%"><i class="fas fa-calendar"></i> Date</th>
                <th style="width: 25%"><i class="fas fa-money-bill"></i> Amount</th>
                <th style="width: 45%"><i class="fas fa-comment"></i> Notes</th>
                <th style="width: 10%"><i class="fas fa-cogs"></i> Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="expense in filteredExpenses" :key="expense.id">
                <td>{{ formatDate(expense.date_transfer) }}</td>
                <td class="amount-cell">
                  <i class="fas fa-minus-circle"></i>
                  {{ formatCurrency(expense.amount) }}
                </td>
                <td class="notes-cell">
                  <span v-if="expense.notes">{{ expense.notes }}</span>
                  <span v-else class="no-notes">No notes</span>
                </td>
                <td class="actions-cell">
                  <div class="action-buttons">
                    <button
                      @click="printReceipt(expense)"
                      class="btn btn-print"
                      title="Print receipt"
                    >
                      <i class="fas fa-print"></i>
                    </button>
                    <button
                      @click="deleteExpense(expense.id)"
                      class="btn btn-danger"
                      :disabled="loading"
                      title="Delete expense"
                    >
                      <i class="fas fa-trash"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.expenses {
  padding: 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.header {
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
  color: #7c3aed;
  font-size: 1.5em;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 1px solid #7c3aed;
  border-radius: 8px;
  background: white;
  color: #7c3aed;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.filter-btn:hover {
  background: #7c3aed;
  color: white;
}

.filter-btn.active {
  background: #7c3aed;
  color: white;
}

.filter-badge {
  background: #dc2626;
  color: white;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 10px;
  margin-left: 4px;
}

.print-list-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 1px solid #059669;
  border-radius: 8px;
  background: white;
  color: #059669;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.print-list-btn:hover:not(:disabled) {
  background: #059669;
  color: white;
}

.print-list-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.expenses-content {
  position: relative;
}

.loading-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px;
  color: #7c3aed;
}

.loading-overlay span {
  color: #64748b;
  font-size: 0.875rem;
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

.add-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  background: #7c3aed;
  color: white;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.add-btn:hover {
  background: #6d28d9;
  transform: translateY(-1px);
}

/* Filters Panel */
.filters-panel {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
  animation: slideDown 0.3s ease;
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filters-header h3 {
  margin: 0;
  color: #1e293b;
  font-size: 1.125rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.filters-header h3 i {
  color: #7c3aed;
}

.clear-filters-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid #dc2626;
  border-radius: 6px;
  background: white;
  color: #dc2626;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.clear-filters-btn:hover:not(:disabled) {
  background: #dc2626;
  color: white;
}

.clear-filters-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-group.full-width {
  grid-column: 1 / -1;
}

.filter-group label {
  color: #374151;
  font-size: 0.875rem;
  font-weight: 500;
}

.filter-input {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.filter-input:focus {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.filter-input select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='m6 8 4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 8px center;
  background-repeat: no-repeat;
  background-size: 16px;
  padding-right: 32px;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Summary Cards */
.summary-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.summary-card {
  background: #f8fafc;
  padding: 20px;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s ease;
}

.summary-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.summary-card i {
  color: #7c3aed;
  font-size: 1.5rem;
}

.summary-content {
  flex: 1;
}

.summary-content .label {
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 500;
  display: block;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.summary-content .value {
  color: #1e293b;
  font-size: 1.5rem;
  font-weight: 600;
}

/* Add Form */
.add-form {
  background: #f8fafc;
  padding: 24px;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  margin-bottom: 24px;
}

.add-form h3 {
  margin: 0 0 20px 0;
  color: #1e293b;
  font-size: 1.25rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.add-form h3 i {
  color: #7c3aed;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  color: #374151;
  font-size: 0.875rem;
  font-weight: 500;
}

.form-input,
.form-textarea {
  padding: 12px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #7c3aed;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.form-input:disabled,
.form-textarea:disabled {
  background-color: #f9fafb;
  cursor: not-allowed;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border: none;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #7c3aed;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #6d28d9;
  transform: translateY(-1px);
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background: #4b5563;
  transform: translateY(-1px);
}

.btn-danger {
  background: #ef4444;
  color: white;
  padding: 8px 12px;
  font-size: 0.75rem;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
  transform: translateY(-1px);
}

.btn-print {
  background: #3b82f6;
  color: white;
  padding: 8px 12px;
  font-size: 0.75rem;
}

.btn-print:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
}

/* List Header */
.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.list-header h3 {
  margin: 0;
  color: #1e293b;
  font-size: 1.25rem;
}

/* Expenses Table */
.expenses-table {
  background: white;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

.expenses-table table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.expenses-table th {
  background: #f8fafc;
  padding: 16px;
  font-weight: 600;
  color: #475569;
  border-bottom: 2px solid #e2e8f0;
  text-align: left;
}

.expenses-table th i {
  color: #64748b;
  margin-right: 8px;
}

.expenses-table td {
  padding: 16px;
  border-bottom: 1px solid #e2e8f0;
  color: #1e293b;
}

.expenses-table tr:hover {
  background: #f8fafc;
}

.amount-cell {
  color: #dc2626;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.amount-cell i {
  color: #dc2626;
}

.notes-cell {
  max-width: 200px;
}

.no-notes {
  color: #9ca3af;
  font-style: italic;
}

.actions-cell {
  text-align: center;
}

.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Responsive */
@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .summary-cards {
    grid-template-columns: 1fr;
  }

  .list-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .expenses-table {
    overflow-x: auto;
  }
}
</style>
