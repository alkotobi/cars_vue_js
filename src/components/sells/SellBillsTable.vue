<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import { useRouter } from 'vue-router'
import SellBillPrintOption from './SellBillPrintOption.vue'

const { t } = useEnhancedI18n()
const props = defineProps({
  onEdit: Function,
  onDelete: Function,
  onSelect: Function,
  onTask: Function,
  selectedBillId: {
    type: Number,
    default: null,
  },
})
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)
const can_c_all_bills = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_c_other_users_sells')
})
const emit = defineEmits(['refresh', 'select-bill', 'update-selected-bills'])

// Multiple selection state
const selectedBills = ref([])

const router = useRouter()
const { callApi, getAssets } = useApi()
const letterHeadUrl = ref(null)
const sellBills = ref([])
const loading = ref(true)
const error = ref(null)
const selectedBillId = ref(null)
const showPrintOptions = ref(false)
const selectedPrintBillId = ref(null)
const isProcessing = ref(false)

// Add filter states
const filters = ref({
  dateFrom: '',
  dateTo: '',
  broker: '',
  reference: '',
  isBatchSell: null,
  createdBy: '',
  paymentStatus: '',
  loadingStatus: '',
})

// Add allSellBills to store unfiltered data
const allSellBills = ref([])

// Add computed property for payments permission
const can_c_sell_payments = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_c_sell_payments')
})

// Add computed property for delete permission
const can_delete_sell_bill = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_delete_sell_bill')
})

// Add computed property for edit permission
const can_edit_sell_bill = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_sell_bill')
})

const sortConfig = ref({
  field: 'id',
  direction: 'desc',
})

// Add computed property for sorted bills
const sortedBills = computed(() => {
  if (!sellBills.value.length) return []

  return [...sellBills.value].sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle date comparison
    if (sortConfig.value.field === 'date_sell') {
      aValue = new Date(aValue).getTime()
      bValue = new Date(bValue).getTime()
    }

    // Handle numeric fields
    if (sortConfig.value.field === 'id') {
      aValue = Number(aValue)
      bValue = Number(bValue)
    }

    // Handle null values
    if (aValue === null || aValue === undefined) aValue = ''
    if (bValue === null || bValue === undefined) bValue = ''

    // Compare values based on direction
    if (sortConfig.value.direction === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })
})

// Add computed property for sorted and limited bills
const sortedAndLimitedBills = computed(() => {
  if (!sellBills.value.length) return []

  const sorted = [...sellBills.value].sort((a, b) => {
    let aValue = a[sortConfig.value.field]
    let bValue = b[sortConfig.value.field]

    // Handle date comparison
    if (sortConfig.value.field === 'date_sell') {
      aValue = new Date(aValue).getTime()
      bValue = new Date(bValue).getTime()
    }

    // Handle numeric fields
    if (sortConfig.value.field === 'id') {
      aValue = Number(aValue)
      bValue = Number(bValue)
    }

    // Handle null values
    if (aValue === null || aValue === undefined) aValue = ''
    if (bValue === null || bValue === undefined) bValue = ''

    // Compare values based on direction
    if (sortConfig.value.direction === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })

  return sorted
})

// Add computed property for unpaid bills count for current user
const unpaidBillsCount = computed(() => {
  if (!allSellBills.value.length) return 0

  return allSellBills.value.filter((bill) => {
    // Admin counts all unpaid bills, regular users count only their own
    if (!isAdmin.value && bill.id_user !== user.value?.id) {
      return false
    }

    // Exclude batch sell bills
    if (bill.is_batch_sell) {
      return false
    }

    // Check if bill is unpaid (total_paid < total_cfr)
    return bill.total_paid < bill.total_cfr
  }).length
})

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchSellBills()
})

const fetchSellBills = async () => {
  loading.value = true
  error.value = null

  try {
    // Different query based on admin status - now includes payment and loading information
    const query = can_c_all_bills.value
      ? `
      SELECT 
        sb.*,
        c.name as broker_name,
        u.username as created_by,
        (
          SELECT GROUP_CONCAT(DISTINCT cl.name SEPARATOR ', ')
          FROM cars_stock cs
          LEFT JOIN clients cl ON cs.id_client = cl.id
          WHERE cs.id_sell = sb.id AND cl.name IS NOT NULL
        ) as client_names,
        (
          SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cfr,
        (
          SELECT COALESCE(SUM(sp.amount_usd), 0)
          FROM sell_payments sp
          WHERE sp.id_sell_bill = sb.id
        ) as total_paid,
        (
          SELECT COUNT(*)
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cars,
        (
          SELECT COUNT(*)
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id AND cs.container_ref IS NOT NULL
        ) as loaded_cars
      FROM sell_bill sb
      LEFT JOIN clients c ON sb.id_broker = c.id
      LEFT JOIN users u ON sb.id_user = u.id
      ORDER BY sb.date_sell DESC
    `
      : `
      SELECT 
        sb.*,
        c.name as broker_name,
        u.username as created_by,
        (
          SELECT GROUP_CONCAT(DISTINCT cl.name SEPARATOR ', ')
          FROM cars_stock cs
          LEFT JOIN clients cl ON cs.id_client = cl.id
          WHERE cs.id_sell = sb.id AND cl.name IS NOT NULL
        ) as client_names,
        (
          SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cfr,
        (
          SELECT COALESCE(SUM(sp.amount_usd), 0)
          FROM sell_payments sp
          WHERE sp.id_sell_bill = sb.id
        ) as total_paid,
        (
          SELECT COUNT(*)
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cars,
        (
          SELECT COUNT(*)
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id AND cs.container_ref IS NOT NULL
        ) as loaded_cars
      FROM sell_bill sb
      LEFT JOIN clients c ON sb.id_broker = c.id
      LEFT JOIN users u ON sb.id_user = u.id
      WHERE sb.id_user = ?
      ORDER BY sb.date_sell DESC
    `

    const params = can_c_all_bills.value ? [] : [user.value?.id]
    const result = await callApi({ query, params })

    if (result.success) {
      console.log(result.data.length)
      allSellBills.value = result.data.map((bill) => ({
        ...bill,
        total_cfr: Number(bill.total_cfr) || 0,
        total_paid: Number(bill.total_paid) || 0,
        total_cars: Number(bill.total_cars) || 0,
        loaded_cars: Number(bill.loaded_cars) || 0,
      }))
      applyFilters() // Apply filters after fetching
    } else {
      error.value = result.error || 'Failed to fetch sell bills'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Add filter functions
const applyFilters = () => {
  sellBills.value = allSellBills.value.filter((bill) => {
    // Date range filter
    if (filters.value.dateFrom && new Date(bill.date_sell) < new Date(filters.value.dateFrom)) {
      return false
    }
    if (filters.value.dateTo && new Date(bill.date_sell) > new Date(filters.value.dateTo)) {
      return false
    }

    // NAME filter - searches both broker and client names
    if (filters.value.broker) {
      const searchTerm = filters.value.broker.toLowerCase()

      // Check if broker name matches
      const brokerMatch = bill.broker_name && bill.broker_name.toLowerCase().includes(searchTerm)

      // Check if any client name matches
      const clientMatch = bill.client_names && bill.client_names.toLowerCase().includes(searchTerm)

      // Bill must match either broker OR at least one client
      if (!brokerMatch && !clientMatch) {
        return false
      }
    }

    // Reference filter
    if (
      filters.value.reference &&
      (!bill.bill_ref ||
        !bill.bill_ref.toLowerCase().includes(filters.value.reference.toLowerCase()))
    ) {
      return false
    }

    // Batch sell filter
    if (filters.value.isBatchSell !== null) {
      if (filters.value.isBatchSell && !bill.is_batch_sell) return false
      if (!filters.value.isBatchSell && bill.is_batch_sell) return false
    }

    // Created by filter
    if (
      filters.value.createdBy &&
      (!bill.created_by ||
        !bill.created_by.toLowerCase().includes(filters.value.createdBy.toLowerCase()))
    ) {
      return false
    }

    // Payment status filter
    if (filters.value.paymentStatus) {
      const isPaid = bill.total_paid >= bill.total_cfr
      const isPartiallyPaid = bill.total_paid > 0 && bill.total_paid < bill.total_cfr
      const isUnpaid = bill.total_paid === 0

      if (filters.value.paymentStatus === 'paid' && !isPaid) return false
      if (filters.value.paymentStatus === 'partially_paid' && !isPartiallyPaid) return false
      if (filters.value.paymentStatus === 'unpaid' && !isUnpaid) return false
    }

    // Loading status filter
    if (filters.value.loadingStatus) {
      const isFullyLoaded = bill.loaded_cars === bill.total_cars && bill.total_cars > 0
      const isPartiallyLoaded = bill.loaded_cars > 0 && bill.loaded_cars < bill.total_cars
      const isNotLoaded = bill.loaded_cars === 0

      if (filters.value.loadingStatus === 'fully_loaded' && !isFullyLoaded) return false
      if (filters.value.loadingStatus === 'partially_loaded' && !isPartiallyLoaded) return false
      if (filters.value.loadingStatus === 'not_loaded' && !isNotLoaded) return false
    }

    return true
  })
}

// Watch for filter changes
watch(
  filters,
  () => {
    applyFilters()
  },
  { deep: true },
)

const resetFilters = () => {
  filters.value = {
    dateFrom: '',
    dateTo: '',
    broker: '',
    reference: '',
    isBatchSell: null,
    createdBy: '',
    paymentStatus: '',
    loadingStatus: '',
  }
}

const handleEdit = (bill) => {
  if (props.onEdit) {
    props.onEdit(bill)
  }
}

const handleDelete = (id) => {
  if (props.onDelete) {
    props.onDelete(id)
  }
}

const handlePrint = (id) => {
  selectedPrintBillId.value = id
  showPrintOptions.value = true
}

const handlePrintClose = () => {
  showPrintOptions.value = false
  selectedPrintBillId.value = null
}

const handlePrintProceed = (options) => {
  console.log('Print options:', options)
  showPrintOptions.value = false
  selectedPrintBillId.value = null
}

const handlePayments = (billId) => {
  // Open payments in a new tab using router.resolve
  const route = router.resolve({
    name: 'sell-bill-payments',
    params: { id: billId },
  })
  window.open(route.href, '_blank')
}

const handleTask = (bill) => {
  if (props.onTask) {
    props.onTask(bill)
  }
}

const selectBill = (bill) => {
  console.log('selectBill called with bill:', bill)
  selectedBillId.value = bill.id

  // Emit the selected bill ID to the parent component
  console.log('Emitting select-bill event with billId:', bill.id)
  emit('select-bill', bill.id)

  if (props.onSelect) {
    console.log('Calling props.onSelect with billId:', bill.id)
    props.onSelect(bill.id)
  } else {
    console.log('props.onSelect is not available')
  }
}

const handleSort = (field) => {
  if (sortConfig.value.field === field) {
    // Toggle direction if clicking the same field
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    // Set new field and default to descending
    sortConfig.value.field = field
    sortConfig.value.direction = 'desc'
  }
}

// Watch for changes in the selectedBillId prop
watch(
  () => props.selectedBillId,
  (newId) => {
    selectedBillId.value = newId
  },
  { immediate: true },
)

// Multiple selection methods
const toggleBillSelection = (billId, event) => {
  event.stopPropagation() // Prevent row click
  const index = selectedBills.value.indexOf(billId)
  if (index > -1) {
    selectedBills.value.splice(index, 1)
  } else {
    selectedBills.value.push(billId)
  }
  emitSelectedBills()
}

const isBillSelected = (billId) => {
  return selectedBills.value.includes(billId)
}

// Check if all visible bills are selected
const allVisibleSelected = computed(() => {
  if (sortedAndLimitedBills.value.length === 0) return false
  return sortedAndLimitedBills.value.every((bill) => selectedBills.value.includes(bill.id))
})

// Select all visible bills (filtered)
const toggleSelectAllVisible = () => {
  if (allVisibleSelected.value) {
    // Deselect all visible bills
    sortedAndLimitedBills.value.forEach((bill) => {
      const index = selectedBills.value.indexOf(bill.id)
      if (index > -1) {
        selectedBills.value.splice(index, 1)
      }
    })
  } else {
    // Select all visible bills (don't clear existing selections)
    sortedAndLimitedBills.value.forEach((bill) => {
      if (!selectedBills.value.includes(bill.id)) {
        selectedBills.value.push(bill.id)
      }
    })
  }
  emitSelectedBills()
}

// Clear all selections
const clearAllSelections = () => {
  selectedBills.value = []
  emitSelectedBills()
}

// Emit selected bills to parent
const emitSelectedBills = () => {
  emit('update-selected-bills', [...selectedBills.value])
}

// Batch delete functionality
const isDeletingBatch = ref(false)
const showBatchDeleteConfirm = ref(false)

// Batch print functionality
const isPrintingBatch = ref(false)

const handleBatchDelete = () => {
  if (!isAdmin.value) {
    alert(t('sellBills.only_admins_can_delete'))
    return
  }

  if (selectedBills.value.length === 0) {
    return
  }

  showBatchDeleteConfirm.value = true
}

const confirmBatchDelete = async () => {
  showBatchDeleteConfirm.value = false
  isDeletingBatch.value = true

  const billsToDelete = [...selectedBills.value]
  const errors = []
  let successCount = 0

  try {
    for (const billId of billsToDelete) {
      try {
        // First, unassign all cars from this bill
        const unassignResult = await callApi({
          query: `
            UPDATE cars_stock 
            SET id_sell = NULL,
                id_client = NULL,
                id_port_discharge = NULL,
                freight = NULL,
                id_sell_pi = NULL,
                is_tmp_client = 0,
                is_batch = 0
            WHERE id_sell = ?
          `,
          params: [billId],
        })

        if (!unassignResult.success) {
          errors.push({ billId, error: `Failed to unassign cars: ${unassignResult.error}` })
          continue
        }

        // Then delete the bill
        const deleteResult = await callApi({
          query: 'DELETE FROM sell_bill WHERE id = ?',
          params: [billId],
        })

        if (deleteResult.success) {
          successCount++
        } else {
          errors.push({ billId, error: deleteResult.error })
        }
      } catch (err) {
        errors.push({ billId, error: err.message })
      }
    }

    // Show results
    if (errors.length === 0) {
      alert(t('sellBills.batch_delete_success', { count: successCount }))
      clearAllSelections()
    } else if (successCount > 0) {
      alert(
        t('sellBills.batch_delete_partial', {
          success: successCount,
          failed: errors.length,
        }),
      )
    } else {
      alert(t('sellBills.batch_delete_failed'))
    }

    // Refresh the table
    await fetchSellBills()

    // Clear selection if all succeeded
    if (errors.length === 0) {
      selectedBills.value = []
      emitSelectedBills()
    }
  } catch (err) {
    console.error('Batch delete error:', err)
    alert(t('sellBills.batch_delete_error', { error: err.message }))
  } finally {
    isDeletingBatch.value = false
  }
}

const cancelBatchDelete = () => {
  showBatchDeleteConfirm.value = false
}

// Batch print functionality
const handleBatchPrint = async () => {
  if (selectedBills.value.length === 0) {
    return
  }

  console.log('Starting batch print for bills:', selectedBills.value)

  isPrintingBatch.value = true

  try {
    // Fetch full details for all selected bills
    const billsData = []

    for (const billId of selectedBills.value) {
      console.log('Fetching bill ID:', billId)

      // Fetch bill details
      const billResult = await callApi({
        query: `
          SELECT 
            sb.*,
            c.name as broker_name,
            c.address as broker_address,
            c.mobiles as broker_mobile,
            u.username as created_by
          FROM sell_bill sb
          LEFT JOIN clients c ON sb.id_broker = c.id
          LEFT JOIN users u ON sb.id_user = u.id
          WHERE sb.id = ?
        `,
        params: [billId],
      })

      console.log('Bill result for ID', billId, ':', billResult)

      if (!billResult.success || billResult.data.length === 0) {
        console.warn('Skipping bill', billId, '- not found or error')
        continue
      }

      const bill = billResult.data[0]
      console.log('Fetched bill:', bill)

      // Fetch cars for this bill
      const carsResult = await callApi({
        query: `
          SELECT 
            cs.*,
            cn.car_name,
            col.color as color_name,
            cl.name as client_name,
            dp.discharge_port as port_name
          FROM cars_stock cs
          LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
          LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
          LEFT JOIN colors col ON cs.id_color = col.id
          LEFT JOIN clients cl ON cs.id_client = cl.id
          LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
          WHERE cs.id_sell = ?
          ORDER BY cs.id
        `,
        params: [billId],
      })

      console.log('Cars result for bill', billId, ':', carsResult)

      billsData.push({
        bill,
        cars: carsResult.success ? carsResult.data : [],
      })
    }

    console.log('Bills data for printing:', billsData)

    if (billsData.length === 0) {
      console.error('No bills data collected!')
      alert(t('sellBills.no_bills_to_print'))
      return
    }

    // Generate and print the report
    generateBatchPrintReport(billsData)
  } catch (err) {
    console.error('Batch print error:', err)
    alert(t('sellBills.batch_print_error', { error: err.message }))
  } finally {
    isPrintingBatch.value = false
  }
}

const generateBatchPrintReport = async (billsData) => {
  console.log('generateBatchPrintReport called with:', billsData.length, 'bills')

  // Load assets if not already loaded
  if (!letterHeadUrl.value) {
    try {
      const assets = await getAssets()
      letterHeadUrl.value = assets?.letter_head || ''
    } catch (err) {
      console.error('Failed to load assets, using default:', err)
    }
  }

  // Pre-translate all strings
  const translations = {
    batchPrintTitle: t('sellBills.batch_print_title'),
    sellBill: t('sellBills.sell_bill'),
    reference: t('sellBills.reference'),
    date: t('sellBills.date'),
    broker: t('sellBills.broker'),
    address: t('sellBills.address'),
    mobile: t('sellBills.mobile'),
    createdBy: t('sellBills.created_by'),
    carsCount: t('sellBills.cars_count'),
    car: t('sellBills.car'),
    color: t('sellBills.color'),
    vin: t('sellBills.vin'),
    client: t('sellBills.client'),
    port: t('sellBills.port'),
    fobDa: t('sellBills.fob_da'),
    freight: t('sellBills.freight'),
    cfrDa: t('sellBills.cfr_da'),
    subtotal: t('sellBills.subtotal'),
    summary: t('sellBills.summary'),
    description: t('sellBills.description'),
    amountDa: t('sellBills.amount_da'),
    amountUsd: t('sellBills.amount_usd'),
    totalFob: t('sellBills.total_fob'),
    totalFreight: t('sellBills.total_freight'),
    totalCfr: t('sellBills.total_cfr'),
    totalBills: t('sellBills.total_bills'),
    totalCars: t('sellBills.total_cars'),
  }

  // Calculate totals
  let totalFobDa = 0
  let totalFobUsd = 0
  let totalFreightUsd = 0
  let totalFreightDa = 0
  let totalCfrDa = 0
  let totalCfrUsd = 0
  let totalCarsCount = 0

  billsData.forEach(({ cars }) => {
    cars.forEach((car) => {
      const priceUsd = Number(car.price_cell) || 0 // price_cell is in USD
      const freightUsd = Number(car.freight) || 0 // freight is in USD
      const rate = Number(car.rate) || 1
      const cfrDa = Number(car.cfr_da) || 0 // Use database cfr_da value

      totalFobUsd += priceUsd
      totalFobDa += priceUsd * rate
      totalFreightUsd += freightUsd
      totalFreightDa += freightUsd * rate
      totalCfrUsd += priceUsd + freightUsd
      totalCfrDa += cfrDa // Use the database value
      totalCarsCount++
    })
  })

  // Create print window
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    alert(t('sellBills.popup_blocked'))
    return
  }

  // Load assets if not already loaded
  if (!letterHeadUrl.value) {
    try {
      const assets = await getAssets()
      letterHeadUrl.value = assets?.letter_head || ''
    } catch (err) {
      console.error('Failed to load assets, using default:', err)
    }
  }

  // Generate HTML
  let html = `
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <title>${translations.batchPrintTitle}</title>
      <style>
        @page {
          size: A4;
          margin: 15mm;
        }
        
        * {
          margin: 0;
          padding: 0;
          box-sizing: border-box;
        }
        
        body {
          font-family: 'Arial', sans-serif;
          font-size: 10pt;
          line-height: 1.4;
          color: #000;
        }
        
        .letter-head {
          width: 100%;
          max-height: 100px;
          object-fit: contain;
          margin-bottom: 20px;
        }
        
        .page-break {
          page-break-after: always;
        }
        
        .bill-section {
          margin-bottom: 25px;
          padding-bottom: 15px;
          border-bottom: 2px dashed #ccc;
        }
        
        .bill-section:last-of-type {
          border-bottom: none;
        }
        
        .bill-header {
          background: #f0f0f0;
          padding: 8px;
          margin-bottom: 8px;
          border: 1px solid #ccc;
        }
        
        .bill-header h2 {
          font-size: 12pt;
          margin-bottom: 4px;
        }
        
        .bill-info {
          display: flex;
          justify-content: space-between;
          margin-bottom: 8px;
          font-size: 9pt;
        }
        
        .info-group {
          flex: 1;
        }
        
        .info-row {
          margin-bottom: 2px;
        }
        
        .info-label {
          font-weight: bold;
          display: inline-block;
          min-width: 70px;
          font-size: 9pt;
        }
        
        table {
          width: 100%;
          border-collapse: collapse;
          margin-bottom: 15px;
          table-layout: fixed;
        }
        
        th, td {
          border: 1px solid #000;
          padding: 4px;
          text-align: left;
          font-size: 8pt;
          word-wrap: break-word;
          overflow: hidden;
        }
        
        th {
          background-color: #e0e0e0;
          font-weight: bold;
          font-size: 8pt;
        }
        
        .text-right {
          text-align: right;
        }
        
        .text-center {
          text-align: center;
        }
        
        /* Column widths */
        .col-num {
          width: 4%;
        }
        
        .col-car {
          width: 18%;
        }
        
        .col-color {
          width: 10%;
        }
        
        .col-vin {
          width: 15%;
        }
        
        .col-client {
          width: 12%;
        }
        
        .col-port {
          width: 12%;
        }
        
        .col-price {
          width: 9%;
        }
        
        .col-freight {
          width: 9%;
        }
        
        .col-cfr {
          width: 9%;
        }
        
        .summary-section {
          margin-top: 30px;
          padding: 15px;
          border: 2px solid #000;
          background: #f9f9f9;
        }
        
        .summary-title {
          font-size: 14pt;
          font-weight: bold;
          text-align: center;
          margin-bottom: 15px;
          text-transform: uppercase;
        }
        
        .summary-table {
          margin-top: 10px;
        }
        
        .summary-table th {
          background-color: #d0d0d0;
          font-size: 9pt;
          padding: 6px;
        }
        
        .summary-table td {
          font-size: 9pt;
          font-weight: bold;
          padding: 6px;
        }
        
        .total-row {
          background-color: #f0f0f0;
          font-weight: bold;
        }
        
        @media print {
          body {
            print-color-adjust: exact;
            -webkit-print-color-adjust: exact;
          }
        }
      </style>
    </head>
    <body>
      <img src="${letterHeadUrl.value}" class="letter-head" alt="Letter Head" />
  `

  // Add each bill
  billsData.forEach(({ bill, cars }, index) => {
    console.log(
      `Processing bill ${index + 1}/${billsData.length}:`,
      bill.id,
      'with',
      cars.length,
      'cars',
    )

    html += `
      <div class="bill-section">
        
        <div class="bill-header">
          <h2>${translations.sellBill} #${bill.id}</h2>
          <div style="display: flex; justify-content: space-between;">
            <span><strong>${translations.reference}:</strong> ${bill.bill_ref || 'N/A'}</span>
            <span><strong>${translations.date}:</strong> ${new Date(bill.date_sell).toLocaleDateString()}</span>
          </div>
        </div>
        
        <div class="bill-info">
          <div class="info-group">
            <div class="info-row">
              <span class="info-label">${translations.broker}:</span>
              <span>${bill.broker_name || 'N/A'}</span>
            </div>
            <div class="info-row">
              <span class="info-label">${translations.address}:</span>
              <span>${bill.broker_address || 'N/A'}</span>
            </div>
            <div class="info-row">
              <span class="info-label">${translations.mobile}:</span>
              <span>${bill.broker_mobile || 'N/A'}</span>
            </div>
          </div>
          <div class="info-group">
            <div class="info-row">
              <span class="info-label">${translations.createdBy}:</span>
              <span>${bill.created_by || 'N/A'}</span>
            </div>
            <div class="info-row">
              <span class="info-label">${translations.carsCount}:</span>
              <span>${cars.length}</span>
            </div>
          </div>
        </div>
        
        <table>
          <thead>
            <tr>
              <th class="text-center col-num">#</th>
              <th class="col-car">${translations.car}</th>
              <th class="col-color">${translations.color}</th>
              <th class="col-vin">${translations.vin}</th>
              <th class="col-client">${translations.client}</th>
              <th class="col-port">${translations.port}</th>
              <th class="text-right col-price">FOB USD</th>
              <th class="text-right col-freight">Freight</th>
              <th class="text-right col-cfr">CFR DA</th>
            </tr>
          </thead>
          <tbody>
    `

    let billFobUsd = 0
    let billFreightUsd = 0
    let billCfrDa = 0

    console.log('Rendering', cars.length, 'cars for bill', bill.id)

    if (cars.length === 0) {
      html += `
        <tr>
          <td colspan="9" class="text-center" style="padding: 20px; color: #999;">No cars in this bill</td>
        </tr>
      `
    } else {
      cars.forEach((car, carIndex) => {
        const priceUsd = Number(car.price_cell) || 0 // USD
        const freightUsd = Number(car.freight) || 0 // USD
        const cfrDa = Number(car.cfr_da) || 0 // Use database cfr_da value

        billFobUsd += priceUsd
        billFreightUsd += freightUsd
        billCfrDa += cfrDa

        console.log(
          'Car',
          carIndex + 1,
          ':',
          car.car_name,
          'VIN:',
          car.vin,
          'Price USD:',
          priceUsd,
          'CFR DA:',
          cfrDa,
        )

        html += `
          <tr>
            <td class="text-center col-num">${carIndex + 1}</td>
            <td class="col-car">${car.car_name || 'N/A'}</td>
            <td class="col-color">${car.color_name || 'N/A'}</td>
            <td class="col-vin">${car.vin || 'N/A'}</td>
            <td class="col-client">${car.client_name || 'N/A'}</td>
            <td class="col-port">${car.port_name || 'N/A'}</td>
            <td class="text-right col-price">$${priceUsd.toFixed(2)}</td>
            <td class="text-right col-freight">$${freightUsd.toFixed(2)}</td>
            <td class="text-right col-cfr">${cfrDa.toFixed(2)}</td>
          </tr>
        `
      })
    }

    html += `
            <tr class="total-row">
              <td colspan="6" class="text-right"><strong>${translations.subtotal}:</strong></td>
              <td class="text-right"><strong>$${billFobUsd.toFixed(2)}</strong></td>
              <td class="text-right"><strong>$${billFreightUsd.toFixed(2)}</strong></td>
              <td class="text-right"><strong>${billCfrDa.toFixed(2)}</strong></td>
            </tr>
          </tbody>
        </table>
      </div>
    `
  })

  // Add summary section
  html += `
    <div class="summary-section page-break">
      <div class="summary-title">${translations.summary}</div>
      
      <table class="summary-table">
        <thead>
          <tr>
            <th>${translations.description}</th>
            <th class="text-right">${translations.amountDa}</th>
            <th class="text-right">${translations.amountUsd}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>${translations.totalFob}</td>
            <td class="text-right">${totalFobDa.toFixed(2)}</td>
            <td class="text-right">$${totalFobUsd.toFixed(2)}</td>
          </tr>
          <tr>
            <td>${translations.totalFreight}</td>
            <td class="text-right">${totalFreightDa.toFixed(2)}</td>
            <td class="text-right">$${totalFreightUsd.toFixed(2)}</td>
          </tr>
          <tr class="total-row">
            <td><strong>${translations.totalCfr}</strong></td>
            <td class="text-right"><strong>${totalCfrDa.toFixed(2)}</strong></td>
            <td class="text-right"><strong>$${totalCfrUsd.toFixed(2)}</strong></td>
          </tr>
          <tr>
            <td colspan="3">&nbsp;</td>
          </tr>
          <tr>
            <td><strong>${translations.totalBills}:</strong></td>
            <td colspan="2" class="text-center"><strong>${billsData.length}</strong></td>
          </tr>
          <tr>
            <td><strong>${translations.totalCars}:</strong></td>
            <td colspan="2" class="text-center"><strong>${totalCarsCount}</strong></td>
          </tr>
        </tbody>
      </table>
    </div>
  `

  html += `
    </body>
    </html>
  `

  console.log('Generated HTML length:', html.length, 'characters')
  console.log('HTML preview (first 500 chars):', html.substring(0, 500))

  printWindow.document.write(html)
  printWindow.document.close()

  console.log('Document written to print window')

  // Wait for images to load before printing
  printWindow.onload = () => {
    setTimeout(() => {
      printWindow.focus()
      printWindow.print()
    }, 500)
  }
}

// Expose the fetchSellBills method and unpaidBillsCount to parent component
defineExpose({ fetchSellBills, unpaidBillsCount, clearAllSelections })

// Add computed properties for payment status
const getPaymentStatus = (bill) => {
  const total = bill.total_cfr || 0
  const paid = bill.total_paid || 0
  const remaining = total - paid

  if (total === 0) {
    return { status: 'no-amount', text: t('sellBills.no_amount'), color: 'gray' }
  } else if (paid === 0) {
    return { status: 'not-paid', text: t('sellBills.not_paid'), color: 'red' }
  } else if (remaining <= 0) {
    return { status: 'paid', text: t('sellBills.paid'), color: 'green' }
  } else {
    return {
      status: 'partial',
      text: `${t('sellBills.left')}: $${remaining.toFixed(2)}`,
      color: 'orange',
    }
  }
}

// Add computed properties for loading status
const getLoadingStatus = (bill) => {
  const totalCars = bill.total_cars || 0
  const loadedCars = bill.loaded_cars || 0

  if (totalCars === 0) {
    return { status: 'no-cars', text: t('sellBills.no_vehicles'), color: 'gray' }
  } else if (loadedCars === 0) {
    return { status: 'not-loaded', text: t('sellBills.awaiting_loading'), color: 'red' }
  } else if (loadedCars === totalCars) {
    return { status: 'fully-loaded', text: t('sellBills.fully_loaded'), color: 'green' }
  } else {
    return {
      status: 'partially-loaded',
      text: `${loadedCars}/${totalCars} ${t('sellBills.loaded')}`,
      color: 'orange',
    }
  }
}
</script>

<template>
  <div class="sell-bills-table-component">
    <!-- Filters Section -->
    <div class="filters-section">
      <div class="filters-header">
        <div class="filters-title-section">
          <h3>
            <i class="fas fa-filter"></i>
            {{ t('sellBills.filters') }}
          </h3>
          <div class="unpaid-bills-info">
            <span class="unpaid-badge">
              <i class="fas fa-exclamation-triangle"></i>
              {{ t('sellBills.unpaid_bills_count', { count: unpaidBillsCount }) }}
            </span>
          </div>
          <div v-if="selectedBills.length > 0" class="selected-bills-info">
            <span class="selected-badge">
              <i class="fas fa-check-square"></i>
              {{ t('sellBills.selected_bills_count', { count: selectedBills.length }) }}
            </span>
            <button
              @click="clearAllSelections"
              class="clear-selection-btn"
              :title="t('sellBills.clear_all_selections')"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>
        <button @click="resetFilters" class="reset-btn">
          <i class="fas fa-undo"></i>
          {{ t('sellBills.reset') }}
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>
            <i class="fas fa-calendar"></i>
            {{ t('sellBills.date_from') }}
          </label>
          <input type="date" v-model="filters.dateFrom" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-calendar"></i>
            {{ t('sellBills.date_to') }}
          </label>
          <input type="date" v-model="filters.dateTo" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-user-tie"></i>
            {{ t('sellBills.name') }}
          </label>
          <input type="text" v-model="filters.broker" :placeholder="t('sellBills.search_name')" />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-barcode"></i>
            {{ t('sellBills.reference') }}
          </label>
          <input
            type="text"
            v-model="filters.reference"
            :placeholder="t('sellBills.search_reference')"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-layer-group"></i>
            {{ t('sellBills.batch_sell') }}
          </label>
          <select v-model="filters.isBatchSell">
            <option :value="null">{{ t('sellBills.all') }}</option>
            <option :value="true">{{ t('sellBills.yes') }}</option>
            <option :value="false">{{ t('sellBills.no') }}</option>
          </select>
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-user"></i>
            {{ t('sellBills.created_by') }}
          </label>
          <input
            type="text"
            v-model="filters.createdBy"
            :placeholder="t('sellBills.search_created_by')"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-credit-card"></i>
            {{ t('sellBills.payment_status') }}
          </label>
          <select v-model="filters.paymentStatus">
            <option value="">{{ t('sellBills.all') }}</option>
            <option value="paid">{{ t('sellBills.paid') }}</option>
            <option value="partially_paid">{{ t('sellBills.partially_paid') }}</option>
            <option value="unpaid">{{ t('sellBills.unpaid') }}</option>
          </select>
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-shipping-fast"></i>
            {{ t('sellBills.loading_status') }}
          </label>
          <select v-model="filters.loadingStatus">
            <option value="">{{ t('sellBills.all') }}</option>
            <option value="fully_loaded">{{ t('sellBills.fully_loaded') }}</option>
            <option value="partially_loaded">{{ t('sellBills.partially_loaded') }}</option>
            <option value="not_loaded">{{ t('sellBills.not_loaded') }}</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Batch Actions Toolbar -->
    <div v-if="selectedBills.length > 0" class="batch-actions-toolbar">
      <div class="toolbar-left">
        <span class="toolbar-selected-count">
          <i class="fas fa-check-circle"></i>
          {{ t('sellBills.selected_bills_count', { count: selectedBills.length }) }}
        </span>
      </div>
      <div class="toolbar-actions">
        <button @click="handleBatchPrint" :disabled="isPrintingBatch" class="batch-print-btn">
          <i class="fas fa-print"></i>
          {{ isPrintingBatch ? t('sellBills.printing') : t('sellBills.print_selected') }}
        </button>
        <button
          v-if="isAdmin"
          @click="handleBatchDelete"
          :disabled="isDeletingBatch"
          class="batch-delete-btn"
        >
          <i class="fas fa-trash-alt"></i>
          {{ isDeletingBatch ? t('sellBills.deleting') : t('sellBills.delete_selected') }}
        </button>
        <button @click="clearAllSelections" class="batch-clear-btn">
          <i class="fas fa-times"></i>
          {{ t('sellBills.clear_selection') }}
        </button>
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>{{ t('sellBills.loading_bills') }}</span>
    </div>

    <div v-if="error" class="error">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-else-if="sortedAndLimitedBills.length === 0" class="no-data">
      <i class="fas fa-inbox fa-2x"></i>
      <p>{{ t('sellBills.no_bills_found') }}</p>
    </div>

    <div v-else class="table-container">
      <table class="sell-bills-table">
        <thead>
          <tr>
            <th class="checkbox-column">
              <input
                type="checkbox"
                :checked="allVisibleSelected"
                @change="toggleSelectAllVisible"
                :title="t('sellBills.select_all_visible')"
                class="select-all-checkbox"
              />
            </th>
            <th @click="handleSort('id')" class="sortable">
              <i class="fas fa-hashtag"></i> {{ t('sellBills.id') }}
              <i
                v-if="sortConfig.field === 'id'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('bill_ref')" class="sortable">
              <i class="fas fa-barcode"></i> {{ t('sellBills.reference') }}
              <i
                v-if="sortConfig.field === 'bill_ref'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('date_sell')" class="sortable">
              <i class="fas fa-calendar"></i> {{ t('sellBills.date') }}
              <i
                v-if="sortConfig.field === 'date_sell'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('broker_name')" class="sortable">
              <i class="fas fa-user-tie"></i> {{ t('sellBills.name') }}
              <i
                v-if="sortConfig.field === 'broker_name'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('created_by')" class="sortable">
              <i class="fas fa-user"></i> {{ t('sellBills.created_by') }}
              <i
                v-if="sortConfig.field === 'created_by'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th @click="handleSort('notes')" class="sortable">
              <i class="fas fa-sticky-note"></i> {{ t('sellBills.notes') }}
              <i
                v-if="sortConfig.field === 'notes'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              ></i>
            </th>
            <th><i class="fas fa-money-bill-wave"></i> {{ t('sellBills.payment_status') }}</th>
            <th><i class="fas fa-shipping-fast"></i> {{ t('sellBills.loading_status') }}</th>
            <th><i class="fas fa-cog"></i> {{ t('sellBills.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="bill in sortedAndLimitedBills"
            :key="bill.id"
            @click="selectBill(bill)"
            :class="{
              selected: selectedBillId === bill.id,
              'multi-selected': isBillSelected(bill.id),
            }"
          >
            <td class="checkbox-column" @click.stop>
              <input
                type="checkbox"
                :checked="isBillSelected(bill.id)"
                @change="(e) => toggleBillSelection(bill.id, e)"
                class="row-checkbox"
              />
            </td>
            <td>{{ bill.id }}</td>
            <td>{{ bill.bill_ref || 'N/A' }}</td>
            <td>
              <div class="date-badges">
                <span class="badge sell-date-badge">
                  <i class="fas fa-calendar"></i>
                  {{ t('sellBills.sell') }} {{ new Date(bill.date_sell).toLocaleDateString() }}
                </span>
                <span v-if="isAdmin" class="badge created-date-badge">
                  <i class="fas fa-clock"></i>
                  {{ t('sellBills.created') }}
                  {{ bill.time_created ? new Date(bill.time_created).toLocaleDateString() : 'N/A' }}
                </span>
              </div>
            </td>
            <td>{{ bill.broker_name || 'N/A' }}</td>
            <td>{{ bill.created_by || 'N/A' }}</td>
            <td>{{ bill.notes || 'N/A' }}</td>
            <td>
              <span
                :class="['payment-badge', `payment-${getPaymentStatus(bill).status}`]"
                :title="`${t('sellBills.total')} $${bill.total_cfr.toFixed(2)} | ${t('sellBills.paid_amount')} $${bill.total_paid.toFixed(2)}`"
              >
                {{ getPaymentStatus(bill).text }}
              </span>
            </td>
            <td>
              <span
                :class="['loading-badge', `loading-${getLoadingStatus(bill).status}`]"
                :title="`${bill.loaded_cars} ${t('sellBills.vehicles_loaded', { total: bill.total_cars })}`"
              >
                {{ getLoadingStatus(bill).text }}
              </span>
            </td>
            <td class="actions">
              <button
                v-if="can_edit_sell_bill"
                @click.stop="handleEdit(bill)"
                :disabled="isProcessing"
                class="btn edit-btn"
                :title="t('sellBills.edit_bill')"
              >
                <i class="fas fa-edit"></i>
              </button>
              <button
                v-if="can_delete_sell_bill"
                @click.stop="handleDelete(bill.id)"
                :disabled="isProcessing"
                class="btn delete-btn"
                :title="t('sellBills.delete_bill')"
              >
                <i class="fas fa-trash-alt"></i>
              </button>
              <button
                @click.stop="handlePrint(bill.id)"
                :disabled="isProcessing"
                class="btn print-btn"
                :title="t('sellBills.print_bill')"
              >
                <i class="fas fa-print"></i>
              </button>
              <button
                v-if="can_c_sell_payments"
                @click.stop="handlePayments(bill.id)"
                :disabled="isProcessing"
                class="btn payment-btn"
                :title="t('sellBills.view_payments')"
              >
                <i class="fas fa-money-bill-wave"></i>
              </button>
              <button
                @click.stop="handleTask(bill)"
                :disabled="isProcessing"
                class="btn task-btn"
                :title="t('sellBills.add_new_task')"
              >
                <i class="fas fa-tasks"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <SellBillPrintOption
      :visible="showPrintOptions"
      :billId="selectedPrintBillId"
      @close="handlePrintClose"
      @proceed="handlePrintProceed"
    />

    <!-- Batch Delete Confirmation Dialog -->
    <div v-if="showBatchDeleteConfirm" class="dialog-overlay" @click.self="cancelBatchDelete">
      <div class="dialog batch-delete-dialog">
        <div class="dialog-header">
          <i class="fas fa-exclamation-triangle warning-icon"></i>
          <h3>{{ t('sellBills.confirm_batch_delete') }}</h3>
        </div>
        <div class="dialog-body">
          <p>{{ t('sellBills.batch_delete_message', { count: selectedBills.length }) }}</p>
          <p class="warning-text">{{ t('sellBills.batch_delete_warning') }}</p>
        </div>
        <div class="dialog-actions">
          <button @click="cancelBatchDelete" class="dialog-btn cancel-btn">
            <i class="fas fa-times"></i>
            {{ t('sellBills.cancel') }}
          </button>
          <button @click="confirmBatchDelete" class="dialog-btn confirm-delete-btn">
            <i class="fas fa-trash-alt"></i>
            {{ t('sellBills.confirm_delete') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sell-bills-table-component {
  position: relative;
  margin-bottom: 20px;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  z-index: 10;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f8f9fa;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.selected-bill-info {
  font-size: 1rem;
  color: #374151;
}

.bill-id {
  font-weight: 600;
  color: #3b82f6;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
}

.sell-bills-table {
  width: 100%;
  border-collapse: collapse;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.sell-bills-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #f3f4f6;
}

.sell-bills-table th,
.sell-bills-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.sell-bills-table th {
  background-color: #f3f4f6;
  font-weight: 600;
  color: #374151;
}

.sell-bills-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.sell-bills-table tr:hover {
  background-color: #f9fafb;
  cursor: pointer;
}

.sell-bills-table tr.selected {
  background-color: #e5edff;
}

.sell-bills-table tr.multi-selected {
  background-color: #f0fdf4;
  border-left: 3px solid #22c55e;
}

.loading,
.error,
.no-data {
  padding: 2rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  color: #6b7280;
}

.error {
  color: #dc2626;
  background-color: #fee2e2;
  border-radius: 8px;
  padding: 1rem;
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.print-btn {
  background-color: #10b981;
  color: white;
}

.print-btn:hover:not(:disabled) {
  background-color: #059669;
}

.payment-btn {
  background-color: #10b981;
  color: white;
}

.payment-btn:hover:not(:disabled) {
  background-color: #059669;
}

.task-btn {
  background-color: #8b5cf6;
  color: white;
}

.task-btn:hover:not(:disabled) {
  background-color: #7c3aed;
}

/* Add smooth transitions */
.sell-bills-table tr {
  transition: background-color 0.2s;
}

.btn i {
  transition: transform 0.2s;
}

.btn:hover:not(:disabled) i {
  transform: scale(1.1);
}

.filters-section {
  background-color: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filters-title-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.filters-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #1f2937;
  font-size: 1.1rem;
}

.unpaid-bills-info {
  display: flex;
  align-items: center;
}

.unpaid-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 6px;
  color: #92400e;
  font-size: 0.9rem;
  font-weight: 500;
}

.unpaid-badge i {
  color: #f59e0b;
}

.reset-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #e5e7eb;
  border: none;
  border-radius: 4px;
  color: #4b5563;
  cursor: pointer;
  transition: all 0.2s;
}

.reset-btn:hover {
  background-color: #d1d5db;
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

.filter-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 0.9rem;
}

.filter-group input,
.filter-group select {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.filter-group input:focus,
.filter-group select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.filter-group select {
  cursor: pointer;
}

@media (max-width: 768px) {
  .filters-grid {
    grid-template-columns: 1fr;
  }
}

/* Add sorting styles */
.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 24px; /* Space for sort icon */
}

.sortable i:last-child {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 0.75rem;
  opacity: 0.5;
}

.sortable:hover {
  background-color: #f1f5f9;
}

.sortable:hover i:last-child {
  opacity: 1;
}

.table-container {
  max-height: 400px; /* Limit the height */
  overflow-y: auto; /* Make it scrollable */
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

/* Payment status badges */
.payment-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  text-align: center;
  min-width: 80px;
  cursor: help;
}

.payment-paid {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.payment-not-paid {
  background-color: #fee2e2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.payment-partial {
  background-color: #fed7aa;
  color: #92400e;
  border: 1px solid #fdba74;
}

.payment-no-amount {
  background-color: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
}

/* Loading status badges */
.loading-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  text-align: center;
  min-width: 80px;
  cursor: help;
}

.loading-fully-loaded {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.loading-not-loaded {
  background-color: #fee2e2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.loading-partially-loaded {
  background-color: #fed7aa;
  color: #92400e;
  border: 1px solid #fdba74;
}

.loading-no-cars {
  background-color: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
}

/* New styles for date badges */
.date-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.sell-date-badge,
.created-date-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  font-family: inherit;
  text-align: center;
  width: 100%;
  box-sizing: border-box;
  justify-content: center;
}

.sell-date-badge {
  background-color: #eff6ff;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}

.created-date-badge {
  background-color: #f3e8ff;
  color: #6b21a8;
  border: 1px solid #d9b8ff;
}

/* Checkbox column styles */
.checkbox-column {
  width: 40px;
  text-align: center;
  padding: 8px !important;
}

.select-all-checkbox,
.row-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #22c55e;
}

.select-all-checkbox:hover,
.row-checkbox:hover {
  transform: scale(1.1);
}

/* Selected bills info styles */
.selected-bills-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selected-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #d1fae5;
  border: 1px solid #22c55e;
  border-radius: 6px;
  color: #065f46;
  font-size: 0.9rem;
  font-weight: 500;
}

.selected-badge i {
  color: #22c55e;
}

.clear-selection-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  border-radius: 4px;
  color: #dc2626;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-selection-btn:hover {
  background-color: #fecaca;
  transform: scale(1.05);
}

.clear-selection-btn i {
  font-size: 0.9rem;
}

/* Batch Actions Toolbar */
.batch-actions-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  animation: slideDown 0.3s ease-out;
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

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-selected-count {
  display: flex;
  align-items: center;
  gap: 8px;
  color: white;
  font-weight: 600;
  font-size: 1rem;
}

.toolbar-selected-count i {
  font-size: 1.2rem;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
}

.batch-print-btn,
.batch-delete-btn,
.batch-clear-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.95rem;
}

.batch-print-btn {
  background-color: #10b981;
  color: white;
}

.batch-print-btn:hover:not(:disabled) {
  background-color: #059669;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(16, 185, 129, 0.3);
}

.batch-print-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.batch-delete-btn {
  background-color: #ef4444;
  color: white;
}

.batch-delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(239, 68, 68, 0.3);
}

.batch-delete-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.batch-clear-btn {
  background-color: white;
  color: #4b5563;
  border: 1px solid #e5e7eb;
}

.batch-clear-btn:hover {
  background-color: #f3f4f6;
  transform: translateY(-1px);
}

/* Batch Delete Dialog */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.batch-delete-dialog {
  background: white;
  border-radius: 12px;
  box-shadow:
    0 20px 25px -5px rgba(0, 0, 0, 0.1),
    0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 500px;
  width: 90%;
  animation: scaleIn 0.2s ease-out;
}

@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 1.25rem;
}

.warning-icon {
  color: #f59e0b;
  font-size: 1.5rem;
}

.dialog-body {
  padding: 24px;
}

.dialog-body p {
  margin: 0 0 12px 0;
  color: #4b5563;
  line-height: 1.6;
}

.warning-text {
  color: #dc2626;
  font-weight: 500;
  background-color: #fee2e2;
  padding: 12px;
  border-radius: 6px;
  border-left: 4px solid #dc2626;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  background-color: #f9fafb;
  border-radius: 0 0 12px 12px;
}

.dialog-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.95rem;
}

.cancel-btn {
  background-color: #e5e7eb;
  color: #4b5563;
}

.cancel-btn:hover {
  background-color: #d1d5db;
}

.confirm-delete-btn {
  background-color: #ef4444;
  color: white;
}

.confirm-delete-btn:hover {
  background-color: #dc2626;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(239, 68, 68, 0.3);
}
</style>
