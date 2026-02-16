<script setup>
import { ref, watch, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useInvoiceCompanyInfo } from '../../composables/useInvoiceCompanyInfo'

const props = defineProps({
  selectedUser: {
    type: Object,
    required: true,
  },
})

const { callApi, getAssets } = useApi()
const { getCompanyLogoUrl } = useInvoiceCompanyInfo()
const letterHeadUrl = ref(null)
const sellPayments = ref([])
const receivedTransfers = ref([])
const sentTransfers = ref([])
const externalTransfers = ref([])
const loading = ref(false)
const error = ref(null)
const showFilters = ref(false)

// Filters
const filters = ref({
  dateFrom: '',
  dateTo: '',
  amountFrom: '',
  amountTo: '',
  transactionType: '',
  source: '',
  notesSearch: '',
})

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
const allTransactions = computed(() => {
  // Combine all transactions
  const transactions = [
    ...sellPayments.value,
    ...receivedTransfers.value,
    ...sentTransfers.value,
    ...externalTransfers.value,
  ]

  // Sort by date (oldest first)
  transactions.sort((a, b) => {
    return new Date(a.transaction_date) - new Date(b.transaction_date)
  })

  let runningBalance = 0
  let previousBalance = 0

  // Process transactions and calculate balances
  return transactions.map((t, index) => {
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
  })
})

// Apply filters to transactions
const transactionsWithBalance = computed(() => {
  let filtered = [...allTransactions.value]

  // Date range filter
  if (filters.value.dateFrom) {
    filtered = filtered.filter(
      (transaction) => new Date(transaction.transaction_date) >= new Date(filters.value.dateFrom),
    )
  }

  if (filters.value.dateTo) {
    filtered = filtered.filter(
      (transaction) => new Date(transaction.transaction_date) <= new Date(filters.value.dateTo),
    )
  }

  // Amount range filter
  if (filters.value.amountFrom) {
    filtered = filtered.filter(
      (transaction) => parseFloat(transaction.amount) >= parseFloat(filters.value.amountFrom),
    )
  }

  if (filters.value.amountTo) {
    filtered = filtered.filter(
      (transaction) => parseFloat(transaction.amount) <= parseFloat(filters.value.amountTo),
    )
  }

  // Transaction type filter
  if (filters.value.transactionType) {
    filtered = filtered.filter((transaction) => transaction.type === filters.value.transactionType)
  }

  // Source filter
  if (filters.value.source) {
    filtered = filtered.filter((transaction) => transaction.source === filters.value.source)
  }

  // Notes search filter
  if (filters.value.notesSearch) {
    const searchTerm = filters.value.notesSearch.toLowerCase()
    filtered = filtered.filter(
      (transaction) => transaction.notes && transaction.notes.toLowerCase().includes(searchTerm),
    )
  }

  return filtered
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

const clearFilters = () => {
  filters.value = {
    dateFrom: '',
    dateTo: '',
    amountFrom: '',
    amountTo: '',
    transactionType: '',
    source: '',
    notesSearch: '',
  }
}

const hasActiveFilters = computed(() => {
  return (
    filters.value.dateFrom ||
    filters.value.dateTo ||
    filters.value.amountFrom ||
    filters.value.amountTo ||
    filters.value.transactionType ||
    filters.value.source ||
    filters.value.notesSearch
  )
})

const printTransactionsList = async () => {
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

  // Calculate totals for print
  const printTotalIn = transactionsWithBalance.value.reduce(
    (sum, t) => (t.type === 'in' ? sum + t.amount : sum),
    0,
  )
  const printTotalOut = transactionsWithBalance.value.reduce(
    (sum, t) => (t.type === 'out' ? sum + t.amount : sum),
    0,
  )

  const transactionsListHtml = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Transactions List</title>
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
        
        .transactions-table {
          width: 100%;
          border-collapse: collapse;
          margin-bottom: 20px;
        }
        
        .transactions-table th {
          background-color: #f8f9fa;
          padding: 8px 6px;
          border: 1px solid #ddd;
          font-weight: bold;
          font-size: 10px;
          text-align: left;
        }
        
        .transactions-table td {
          padding: 6px;
          border: 1px solid #ddd;
          font-size: 10px;
        }
        
        .transactions-table tr:nth-child(even) {
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
        <div class="report-title">TRANSACTIONS LIST</div>
        <div class="report-date">Generated on ${new Date().toLocaleDateString()}</div>
      </div>
      
      <div class="summary-info">
        <div class="summary-item">
          <span class="summary-label">User</span>
          <span class="summary-value">${props.selectedUser.username}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Total Transactions</span>
          <span class="summary-value">${transactionsWithBalance.value.length}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Money In</span>
          <span class="summary-value positive">${formatCurrency(printTotalIn)}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Money Out</span>
          <span class="summary-value negative">${formatCurrency(printTotalOut)}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Current Balance</span>
          <span class="summary-value ${transactionsWithBalance.value.length > 0 ? (transactionsWithBalance.value[transactionsWithBalance.value.length - 1].balance >= 0 ? 'positive' : 'negative') : ''}">
            ${transactionsWithBalance.value.length > 0 ? formatCurrency(transactionsWithBalance.value[transactionsWithBalance.value.length - 1].balance) : 'DZD 0.00'}
          </span>
        </div>
      </div>
      
      <table class="transactions-table">
        <thead>
          <tr>
            <th style="width: 8%;">#</th>
            <th style="width: 15%;">Date</th>
            <th style="width: 15%;">Previous Balance</th>
            <th style="width: 15%;">Money In</th>
            <th style="width: 15%;">Money Out</th>
            <th style="width: 15%;">Current Balance</th>
            <th style="width: 17%;">Notes</th>
          </tr>
        </thead>
        <tbody>
          ${transactionsWithBalance.value
            .map(
              (transaction, index) => `
            <tr>
              <td>${index + 1}</td>
              <td>${formatDate(transaction.transaction_date)}</td>
              <td class="amount-cell ${transaction.previousBalance >= 0 ? 'positive' : 'negative'}">${formatCurrency(transaction.previousBalance)}</td>
              <td class="money-in">
                ${transaction.type === 'in' ? formatCurrency(transaction.amount) : '-'}
              </td>
              <td class="money-out">
                ${transaction.type === 'out' ? formatCurrency(transaction.amount) : '-'}
              </td>
              <td class="amount-cell ${transaction.balance >= 0 ? 'positive' : 'negative'}">${formatCurrency(transaction.balance)}</td>
              <td class="notes-cell">${transaction.notes || 'No notes'}</td>
            </tr>
          `,
            )
            .join('')}
          <tr style="border-top: 2px solid #333; font-weight: bold; background-color: #f8f9fa;">
            <td colspan="3"><strong>TOTAL</strong></td>
            <td class="money-in"><strong>${formatCurrency(printTotalIn)}</strong></td>
            <td class="money-out"><strong>${formatCurrency(printTotalOut)}</strong></td>
            <td class="amount-cell ${transactionsWithBalance.value.length > 0 ? (transactionsWithBalance.value[transactionsWithBalance.value.length - 1].balance >= 0 ? 'positive' : 'negative') : ''}">
              <strong>${transactionsWithBalance.value.length > 0 ? formatCurrency(transactionsWithBalance.value[transactionsWithBalance.value.length - 1].balance) : 'DZD 0.00'}</strong>
            </td>
            <td></td>
          </tr>
        </tbody>
      </table>
      
      <div class="footer">
        <p>This is a computer-generated transactions list.</p>
        <p>Generated on ${new Date().toLocaleString()}</p>
        ${hasActiveFilters.value ? '<p><strong>Note:</strong> This list shows filtered results only.</p>' : ''}
      </div>
    </body>
    </html>
  `

  printWindow.document.write(transactionsListHtml)
  printWindow.document.close()

  // Wait for images to load, then print
  printWindow.onload = () => {
    setTimeout(() => {
      printWindow.print()
      printWindow.close()
    }, 500)
  }
}

const printTransactionReceipt = async (transaction) => {
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
      <title>Transaction Receipt</title>
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
          text-align: center;
          margin-bottom: 15px;
        }
        
        .letterhead img {
          width: 100%;
          height: auto;
          max-height: 40px;
          object-fit: contain;
        }
        
        .receipt-title {
          text-align: center;
          font-size: 16px;
          font-weight: bold;
          margin-bottom: 15px;
          color: #1e293b;
        }
        
        .receipt-details {
          background: #f8f9fa;
          padding: 12px;
          border-radius: 8px;
          margin-bottom: 15px;
        }
        
        .detail-row {
          display: flex;
          justify-content: space-between;
          margin-bottom: 6px;
          font-size: 11px;
        }
        
        .detail-label {
          font-weight: bold;
          color: #374151;
        }
        
        .detail-value {
          color: #1e293b;
        }
        
        .amount-section {
          text-align: center;
          margin: 20px 0;
          padding: 15px;
          background: ${transaction.type === 'in' ? '#dcfce7' : '#fee2e2'};
          border-radius: 8px;
          border: 2px solid ${transaction.type === 'in' ? '#16a34a' : '#dc2626'};
        }
        
        .amount-label {
          font-size: 12px;
          color: #374151;
          margin-bottom: 5px;
        }
        
        .amount-value {
          font-size: 14px;
          font-weight: bold;
          color: ${transaction.type === 'in' ? '#16a34a' : '#dc2626'};
        }
        
        .notes-section {
          margin-top: 15px;
          padding: 10px;
          background: #f1f5f9;
          border-radius: 6px;
        }
        
        .notes-label {
          font-size: 10px;
          font-weight: bold;
          color: #374151;
          margin-bottom: 5px;
        }
        
        .notes-content {
          font-size: 10px;
          color: #1e293b;
          line-height: 1.3;
        }
        
        .footer {
          margin-top: 20px;
          text-align: center;
          font-size: 8px;
          color: #6b7280;
          border-top: 1px solid #e5e7eb;
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
      
      <div class="receipt-title">TRANSACTION RECEIPT</div>
      
      <div class="receipt-details">
        <div class="detail-row">
          <span class="detail-label">Receipt ID:</span>
          <span class="detail-value">#${transaction.id}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Date:</span>
          <span class="detail-value">${formatDate(transaction.transaction_date)}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">User:</span>
          <span class="detail-value">${props.selectedUser.username}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Type:</span>
          <span class="detail-value">${transaction.type === 'in' ? 'Money In' : 'Money Out'}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Source:</span>
          <span class="detail-value">${getSourceLabel(transaction.source)}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Previous Balance:</span>
          <span class="detail-value">${formatCurrency(transaction.previousBalance)}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Current Balance:</span>
          <span class="detail-value">${formatCurrency(transaction.balance)}</span>
        </div>
      </div>
      
      <div class="amount-section">
        <div class="amount-label">Transaction Amount</div>
        <div class="amount-value">
          ${transaction.type === 'in' ? '+' : '-'}${formatCurrency(transaction.amount)}
        </div>
      </div>
      
      ${
        transaction.notes
          ? `
        <div class="notes-section">
          <div class="notes-label">Notes:</div>
          <div class="notes-content">${transaction.notes}</div>
        </div>
      `
          : ''
      }
      
      <div class="footer">
        <p>This is a computer-generated transaction receipt.</p>
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
</script>

<template>
  <div class="user-transactions" v-if="selectedUser?.id">
    <div class="header">
      <h3>
        <i class="fas fa-history"></i>
        Transactions for {{ selectedUser.username }}
      </h3>
      <div class="header-actions">
        <button
          @click="printTransactionsList"
          class="print-list-btn"
          :disabled="transactionsWithBalance.length === 0"
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
            >{{ transactionsWithBalance.length }}/{{ allTransactions.length }}</span
          >
        </button>
      </div>
    </div>

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
            transactionsWithBalance.reduce((sum, t) => (t.type === 'in' ? sum + t.amount : sum), 0),
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

    <!-- Filters Panel -->
    <div v-if="showFilters" class="filters-panel">
      <div class="filters-header">
        <h4>
          <i class="fas fa-filter"></i>
          Filter Transactions
        </h4>
        <button @click="clearFilters" class="clear-filters-btn" :disabled="!hasActiveFilters">
          <i class="fas fa-times"></i>
          Clear All
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>Date From</label>
          <input v-model="filters.dateFrom" type="date" class="filter-input" />
        </div>

        <div class="filter-group">
          <label>Date To</label>
          <input v-model="filters.dateTo" type="date" class="filter-input" />
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
          />
        </div>

        <div class="filter-group full-width">
          <label>Search in Notes</label>
          <input
            v-model="filters.notesSearch"
            type="text"
            placeholder="Search notes..."
            class="filter-input"
          />
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
            <th><i class="fas fa-history"></i> Previous Balance</th>
            <th><i class="fas fa-arrow-down"></i> Money In</th>
            <th><i class="fas fa-arrow-up"></i> Money Out</th>
            <th><i class="fas fa-wallet"></i> Current Balance</th>
            <th><i class="fas fa-sticky-note"></i> Notes</th>
            <th><i class="fas fa-cogs"></i> Actions</th>
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
            <td class="notes">
              <span v-if="transaction.notes" class="notes-content" :title="transaction.notes">
                <i class="fas fa-comment-alt"></i>
                {{ transaction.notes }}
              </span>
              <span v-else class="no-notes">
                <i class="fas fa-minus"></i>
              </span>
            </td>
            <td class="actions">
              <button
                @click="printTransactionReceipt(transaction)"
                class="btn btn-print"
                title="Print receipt"
              >
                <i class="fas fa-print"></i>
              </button>
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
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.filters-header h4 {
  margin: 0;
  color: #1e293b;
  font-size: 1.125rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.filters-header h4 i {
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

/* Actions Column */
.actions {
  text-align: center;
  width: 80px;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 8px;
  border: none;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  text-decoration: none;
  gap: 4px;
}

.btn-print {
  background-color: #3b82f6;
  color: white;
}

.btn-print:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.btn-print i {
  font-size: 0.75rem;
}
</style>
