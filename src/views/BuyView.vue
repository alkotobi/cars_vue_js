<script setup>
import { ref, computed, onMounted } from 'vue'
import { useEnhancedI18n } from '../composables/useI18n'
import { useApi } from '../composables/useApi'
import BuyBillsTable from '../components/buy/BuyBillsTable.vue'
import BuyDetailsTable from '../components/buy/BuyDetailsTable.vue'
import CarStockTable from '../components/car-stock/CarStockTable.vue'
import TaskForm from '../components/car-stock/TaskForm.vue'
import { useRouter } from 'vue-router'
import EditNotesDialog from '../components/buy/EditNotesDialog.vue'

// API and base setup
const { callApi, uploadFile, getFileUrl, error } = useApi()
const router = useRouter()
const { t } = useEnhancedI18n()

// State management
const buyBills = ref([])
const selectedBill = ref(null)
const buyDetails = ref([])
const suppliers = ref([])
const cars = ref([])
const colors = ref([])
const isLoadingBills = ref(false)

// Component refs for scrolling
const buyBillsTableRef = ref(null)
const buyDetailsTableRef = ref(null)
const carStockTableRef = ref(null)

// Dialog controls
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const showAddDetailDialog = ref(false)
const editingBill = ref(null)

// Add task form state
const showTaskForm = ref(false)
const selectedBillForTask = ref(null)

// Loading states
const isSubmittingPurchase = ref(false)
const isSubmittingDetail = ref(false)
const isUpdatingStock = ref(false)
const isDeletingBill = ref(false)
const isProcessingTask = ref(false)

// User info
const user = ref(JSON.parse(localStorage.getItem('user')))
const isAdmin = computed(() => user.value?.role_id === 1)

// Computed property to check if bill is pending (not stock updated)
const isBillPending = computed(() => {
  if (!selectedBill.value) return false
  const stockUpdated = selectedBill.value.is_stock_updated
  return stockUpdated == 0 || stockUpdated === null || stockUpdated === '0'
})

// Toolbar action methods
const canUpdateStock = (bill) => {
  return bill.amount > 0 && !bill.is_stock_updated
}

const handleUpdateStock = async (bill) => {
  if (!confirm(t('confirm_update_stock'))) {
    return
  }

  if (isUpdatingStock.value) return // Prevent double-click

  try {
    isUpdatingStock.value = true
    const result = await callApi({
      query: `
        UPDATE buy_bill 
        SET is_stock_updated = 1 
        WHERE id = ?
      `,
      params: [bill.id],
    })

    if (result.success) {
      await fetchBuyBills()

      // Update the selected bill data to reflect the new status
      if (selectedBill.value?.id === bill.id) {
        const updatedBill = buyBills.value.find((b) => b.id === bill.id)
        if (updatedBill) {
          selectedBill.value = updatedBill
        }
      }

      // Note: New cars will be added to memory via the cars-created event
      // No need to refresh from database
    } else {
      alert(t('failed_update_stock'))
    }
  } catch (err) {
    console.error('Error updating stock:', err)
    alert(t('failed_update_stock'))
  } finally {
    isUpdatingStock.value = false
  }
}

const handleStockUpdated = async (billId) => {
  // Refresh the bills list to get updated data
  await fetchBuyBills()

  // Update the selected bill data to reflect the new status
  if (selectedBill.value?.id === billId) {
    const updatedBill = buyBills.value.find((b) => b.id === billId)
    if (updatedBill) {
      selectedBill.value = updatedBill
    }
  }
}

const handleCarsCreated = async (newCars) => {
  // Add new cars to CarStockTable memory
  if (carStockTableRef.value) {
    await carStockTableRef.value.addCarsToMemory(newCars)
  }
}

const handleSelectBill = (bill) => {
  selectedBill.value = bill
  fetchBuyDetails(bill.id)

  // Scroll to the appropriate section based on bill status
  setTimeout(() => {
    let element = null

    if (bill.is_stock_updated == 0) {
      // Scroll to buy details table (pending bill)
      element = buyDetailsTableRef.value?.$el
    } else {
      // Scroll to car stock table (updated bill)
      element = carStockTableRef.value?.$el
    }

    // If it's a text node, try to get the parent element
    if (element && element.nodeType === Node.TEXT_NODE) {
      element = element.parentElement
    }

    // Alternative: try to get the element by querying the DOM
    if (!element || !element.getBoundingClientRect) {
      if (bill.is_stock_updated == 0) {
        element =
          document.querySelector('#buy-details-table') ||
          document.querySelector('.buy-details-table-component')
      } else {
        element =
          document.querySelector('#car-stock-table') || document.querySelector('.cars-section')
      }
    }

    if (element && element.getBoundingClientRect) {
      const yOffset = -50 // Increased offset to account for any fixed headers
      const y = element.getBoundingClientRect().top + window.pageYOffset + yOffset
      window.scrollTo({ top: y, behavior: 'smooth' })
    }
  }, 100)

  // Alternative scroll method if the first one doesn't work
  setTimeout(() => {
    let element = null

    if (bill.is_stock_updated == 0) {
      element = buyDetailsTableRef.value?.$el
    } else {
      element = carStockTableRef.value?.$el
    }

    // If it's a text node, try to get the parent element
    if (element && element.nodeType === Node.TEXT_NODE) {
      element = element.parentElement
    }

    // Alternative: try to get the element by querying the DOM
    if (!element || !element.scrollIntoView) {
      if (bill.is_stock_updated == 0) {
        element =
          document.querySelector('#buy-details-table') ||
          document.querySelector('.buy-details-table-component')
      } else {
        element =
          document.querySelector('#car-stock-table') || document.querySelector('.cars-section')
      }
    }

    if (element && element.scrollIntoView) {
      element.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
        inline: 'nearest',
      })
    }
  }, 200)
}

const handleDeleteBill = async (bill) => {
  if (!confirm(t('confirm_delete_purchase_details'))) {
    return
  }

  if (isDeletingBill.value) return // Prevent double-click

  try {
    isDeletingBill.value = true

    // First check if the bill can be deleted
    if (bill.is_stock_updated) {
      alert(t('cannot_delete_stock_updated_purchase'))
      return
    }

    // Start with deleting related records
    // 1. Delete payments
    const deletePayments = await callApi({
      query: `DELETE FROM buy_payments WHERE id_buy_bill = ?`,
      params: [bill.id],
    })

    if (!deletePayments.success) {
      throw new Error(t('failed_delete_related_payments'))
    }

    // 2. Delete details
    const deleteDetails = await callApi({
      query: `DELETE FROM buy_details WHERE id_buy_bill = ?`,
      params: [bill.id],
    })

    if (!deleteDetails.success) {
      throw new Error(t('failed_delete_purchase_details'))
    }

    // 3. Finally delete the bill
    const deleteBill = await callApi({
      query: `DELETE FROM buy_bill WHERE id = ? AND is_stock_updated = 0`,
      params: [bill.id],
    })

    if (deleteBill.success) {
      if (selectedBill.value?.id === bill.id) {
        selectedBill.value = null
        buyDetails.value = [] // Clear details
      }
      await fetchBuyBills()
    } else {
      throw new Error(t('failed_delete_purchase'))
    }
  } catch (err) {
    console.error('Error deleting purchase:', err)
    alert(err.message || t('failed_delete_purchase'))
  } finally {
    isDeletingBill.value = false
  }
}

const openEditDialog = (bill) => {
  editingBill.value = bill
  newPurchase.value = {
    id_supplier: bill.id_supplier,
    date_buy: bill.date_buy.slice(0, 16), // Format for datetime-local input
    bill_ref: bill.bill_ref || '',
    pi_path: bill.pi_path || '',
    pi_file: null,
    is_ordered: bill.is_ordered,
    notes: bill.notes || '',
  }
  showEditDialog.value = true
}

const openAddDialog = () => {
  editingBill.value = null
  newPurchase.value = {
    id_supplier: null,
    date_buy: new Date().toISOString().slice(0, 16),
    bill_ref: '',
    pi_path: '',
    pi_file: null,
    is_ordered: 1,
    notes: '',
  }
  showAddDialog.value = true
}

// Form models
const newPurchase = ref({
  id_supplier: null,
  date_buy: new Date().toISOString().slice(0, 16),
  amount: 0,
  bill_ref: '',
  pi_path: '',
  pi_file: null,
  is_ordered: 1,
  notes: '',
})

const newDetail = ref({
  id_car_name: null,
  id_color: null,
  amount: 0,
  QTY: 1,
  year: new Date().getFullYear(),
  month: new Date().getMonth() + 1,
  is_used_car: false,
  is_big_car: false,
  id_buy_bill: null,
})

// Handle PI file change
const handlePiFileChange = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // Check if file is PDF or image
  const allowedTypes = ['application/pdf', 'image/jpeg', 'image/png', 'image/gif', 'image/webp']

  if (!allowedTypes.includes(file.type)) {
    alert(t('only_pdf_image_files_allowed'))
    event.target.value = ''
    return
  }

  // Store the file for upload
  newPurchase.value.pi_file = file
}

// Data fetching functions
const fetchSuppliers = async () => {
  const result = await callApi({
    query: 'SELECT * FROM suppliers ORDER BY name ASC',
    params: [],
  })
  if (result.success) {
    suppliers.value = result.data
  }
}

const addPurchase = async () => {
  // Prevent multiple submissions
  if (isSubmittingPurchase.value) {
    return
  }

  try {
    isSubmittingPurchase.value = true

    // Validate that PI document is provided
    if (!newPurchase.value.pi_file) {
      alert(t('pi_document_required'))
      return
    }

    // Handle PI document upload
    try {
      // Create filename using purchase details and original extension
      const date = new Date()
      const timestamp = date.getTime()
      const dateStr = date.toISOString().split('T')[0]
      const timeStr = date.toISOString().split('T')[1].split('.')[0].replace(/:/g, '-')
      const supplierName =
        suppliers.value.find((s) => s.id === newPurchase.value.id_supplier)?.name || 'unknown'
      const fileExt = newPurchase.value.pi_file.name.split('.').pop().toLowerCase()

      // Format: pi_purchase_supplier_date_time_timestamp.ext
      const filename = `pi_purchase_${supplierName.replace(/\s+/g, '_')}_${dateStr}_${timeStr}_${timestamp}.${fileExt}`

      const uploadResult = await uploadFile(newPurchase.value.pi_file, 'purchase_pi', filename)

      if (!uploadResult.success) {
        throw new Error(uploadResult.message || t('failed_upload_pi_document'))
      }

      // Store just the relative path
      newPurchase.value.pi_path = `purchase_pi/${filename}`
    } catch (uploadError) {
      console.error('Upload error:', uploadError)
      throw new Error(`${t('failed_upload_pi_document')}: ${uploadError.message}`)
    }

    const result = await callApi({
      query: `
        INSERT INTO buy_bill (id_supplier, date_buy, bill_ref, pi_path, is_ordered, notes)
        VALUES (?, ?, ?, ?, ?, ?)
      `,
      params: [
        newPurchase.value.id_supplier,
        newPurchase.value.date_buy,
        newPurchase.value.bill_ref,
        newPurchase.value.pi_path,
        newPurchase.value.is_ordered,
        newPurchase.value.notes || null,
      ],
    })

    if (result.success) {
      showAddDialog.value = false

      // Get the new bill ID
      const newBillId = result.lastInsertId

      // Refresh the bills list
      await fetchBuyBills()

      // Automatically select the newly created bill
      if (newBillId) {
        const newBill = buyBills.value.find((b) => b.id == newBillId)
        if (newBill) {
          handleSelectBill(newBill)
        } else {
          // If the bill is not in the list, try to fetch it directly
          const billResult = await callApi({
            query: `
              SELECT 
                bb.*,
                s.name as supplier_name,
                COALESCE((
                  SELECT SUM(amount * QTY)
                  FROM buy_details
                  WHERE id_buy_bill = bb.id
                ), 0) as calculated_amount,
                (
                  SELECT COALESCE(SUM(bp.amount), 0)
                  FROM buy_payments bp
                  WHERE bp.id_buy_bill = bb.id
                ) as total_paid
              FROM buy_bill bb 
              LEFT JOIN suppliers s ON bb.id_supplier = s.id 
              WHERE bb.id = ?
            `,
            params: [newBillId],
          })

          if (billResult.success && billResult.data.length > 0) {
            const newBill = {
              ...billResult.data[0],
              amount: Number(billResult.data[0].calculated_amount),
              payed: Number(billResult.data[0].payed),
              total_paid: Number(billResult.data[0].total_paid) || 0,
            }
            handleSelectBill(newBill)
          }
        }
      }

      // Reset form
      newPurchase.value = {
        id_supplier: null,
        date_buy: new Date().toISOString().slice(0, 16),
        bill_ref: '',
        pi_path: '',
        pi_file: null,
        is_ordered: 1,
        notes: '',
      }
    } else {
      throw new Error(result.error || t('failed_add_purchase'))
    }
  } catch (err) {
    console.error('Error adding purchase:', err)
    alert(err.message)
  } finally {
    isSubmittingPurchase.value = false
  }
}

const updatePurchase = async () => {
  // Prevent multiple submissions
  if (isSubmittingPurchase.value) {
    return
  }

  try {
    isSubmittingPurchase.value = true

    // Validate that PI document exists (either existing or new file)
    if (!newPurchase.value.pi_file && !newPurchase.value.pi_path) {
      alert(t('pi_document_required'))
      return
    }

    // Handle PI document upload if there's a new file
    if (newPurchase.value.pi_file) {
      try {
        // Create filename using purchase details and original extension
        const date = new Date()
        const timestamp = date.getTime()
        const dateStr = date.toISOString().split('T')[0]
        const timeStr = date.toISOString().split('T')[1].split('.')[0].replace(/:/g, '-')
        const supplierName =
          suppliers.value.find((s) => s.id === newPurchase.value.id_supplier)?.name || 'unknown'
        const fileExt = newPurchase.value.pi_file.name.split('.').pop().toLowerCase()

        // Format: pi_purchase_supplier_date_time_timestamp.ext
        const filename = `pi_purchase_${supplierName.replace(/\s+/g, '_')}_${dateStr}_${timeStr}_${timestamp}.${fileExt}`

        const uploadResult = await uploadFile(newPurchase.value.pi_file, 'purchase_pi', filename)

        if (!uploadResult.success) {
          throw new Error(uploadResult.message || t('failed_upload_pi_document'))
        }

        // Store just the relative path
        newPurchase.value.pi_path = `purchase_pi/${filename}`
      } catch (uploadError) {
        console.error('Upload error:', uploadError)
        throw new Error(`${t('failed_upload_pi_document')}: ${uploadError.message}`)
      }
    }

    const result = await callApi({
      query: `
        UPDATE buy_bill 
        SET id_supplier = ?, date_buy = ?, bill_ref = ?, pi_path = ?, is_ordered = ?, notes = ?
        WHERE id = ?
      `,
      params: [
        newPurchase.value.id_supplier,
        newPurchase.value.date_buy,
        newPurchase.value.bill_ref,
        newPurchase.value.pi_path,
        newPurchase.value.is_ordered,
        newPurchase.value.notes || null,
        editingBill.value.id,
      ],
    })

    if (result.success) {
      showEditDialog.value = false
      editingBill.value = null
      await fetchBuyBills()
      // Reset form
      newPurchase.value = {
        id_supplier: null,
        date_buy: new Date().toISOString().slice(0, 16),
        bill_ref: '',
        pi_path: '',
        pi_file: null,
        is_ordered: 1,
        notes: '',
      }
    } else {
      throw new Error(result.error || t('failed_update_purchase'))
    }
  } catch (err) {
    console.error('Error updating purchase:', err)
    alert(err.message)
  } finally {
    isSubmittingPurchase.value = false
  }
}

const fetchBuyBills = async () => {
  isLoadingBills.value = true
  const result = await callApi({
    query: `
      SELECT 
        bb.*,
        s.name as supplier_name,
        COALESCE((
          SELECT SUM(amount * QTY)
          FROM buy_details
          WHERE id_buy_bill = bb.id
        ), 0) as calculated_amount,
        (
          SELECT COALESCE(SUM(bp.amount), 0)
          FROM buy_payments bp
          WHERE bp.id_buy_bill = bb.id
        ) as total_paid
      FROM buy_bill bb 
      LEFT JOIN suppliers s ON bb.id_supplier = s.id 
      ORDER BY bb.date_buy DESC
    `,
    params: [],
  })
  if (result.success) {
    buyBills.value = result.data.map((bill) => ({
      ...bill,
      amount: Number(bill.calculated_amount),
      payed: Number(bill.payed),
      total_paid: Number(bill.total_paid) || 0,
    }))
  }
  isLoadingBills.value = false
}

const fetchBuyDetails = async (billId) => {
  const result = await callApi({
    query: `
      SELECT bd.*, cn.car_name, c.color, bb.is_stock_updated
      FROM buy_details bd
      LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
      LEFT JOIN colors c ON bd.id_color = c.id
      LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
      WHERE bd.id_buy_bill = ?
    `,
    params: [billId],
  })
  if (result.success) {
    buyDetails.value = result.data
  } else {
    console.error('Error fetching details:', result.error)
  }
}

const selectBill = (bill) => {
  selectedBill.value = bill
  fetchBuyDetails(bill.id)
}

const fetchCars = async () => {
  const result = await callApi({
    query: 'SELECT * FROM cars_names ORDER BY car_name ASC',
    params: [],
  })
  if (result.success) {
    cars.value = result.data
  }
}

const fetchColors = async () => {
  const result = await callApi({
    query: 'SELECT * FROM colors ORDER BY color ASC',
    params: [],
  })
  if (result.success) {
    colors.value = result.data
  }
}

const updateBillAmount = async (billId) => {
  const result = await callApi({
    query: `
      UPDATE buy_bill 
      SET amount = (
        SELECT COALESCE(SUM(amount * QTY), 0)
        FROM buy_details
        WHERE id_buy_bill = ?
      )
      WHERE id = ?
    `,
    params: [billId, billId],
  })

  if (!result.success) {
    console.error('Error updating bill amount:', result.error)
  }
}

const addDetail = async () => {
  // Prevent multiple submissions
  if (isSubmittingDetail.value) {
    return
  }

  try {
    isSubmittingDetail.value = true

    const result = await callApi({
      query: `
      INSERT INTO buy_details 
      (id_car_name, id_color, amount, QTY, year, month, is_used_car, id_buy_bill, price_sell, is_big_car)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
      params: [
        newDetail.value.id_car_name,
        newDetail.value.id_color,
        newDetail.value.amount,
        newDetail.value.QTY,
        newDetail.value.year,
        newDetail.value.month,
        newDetail.value.is_used_car ? 1 : 0,
        selectedBill.value.id,
        newDetail.value.price_sell,
        newDetail.value.is_big_car ? 1 : 0,
      ],
    })

    if (result.success) {
      await updateBillAmount(selectedBill.value.id)
      showAddDetailDialog.value = false
      await fetchBuyDetails(selectedBill.value.id)
      await fetchBuyBills()

      // Reset form
      newDetail.value = {
        id_car_name: null,
        id_color: null,
        amount: 0,
        QTY: 1,
        year: new Date().getFullYear(),
        month: new Date().getMonth() + 1,
        is_used_car: false,
        is_big_car: false,
        id_buy_bill: null,
      }
    } else {
      console.error('Error adding detail:', result.error)
    }
  } catch (err) {
    console.error('Error adding detail:', err)
    alert(t('failed_add_detail'))
  } finally {
    isSubmittingDetail.value = false
  }
}

// Lifecycle hooks
onMounted(() => {
  fetchBuyBills()
  fetchSuppliers()
  fetchCars()
  fetchColors()
})

const handleDeleteDetail = async (detailId) => {
  const result = await callApi({
    query: 'DELETE FROM buy_details WHERE id = ?',
    params: [detailId],
  })

  if (result.success) {
    await updateBillAmount(selectedBill.value.id) // Update bill amount
    await fetchBuyDetails(selectedBill.value.id)
    await fetchBuyBills() // Refresh bills to show updated amount
  } else {
    alert(t('failed_delete_detail'))
    console.error('Error deleting detail:', result.error)
  }
}

const handleUpdateDetail = async (updatedDetail) => {
  const result = await callApi({
    query: `
      UPDATE buy_details 
      SET QTY = ?, amount = ?, year = ?, month = ?, price_sell = ?, is_big_car = ?
      WHERE id = ?
    `,
    params: [
      updatedDetail.QTY,
      updatedDetail.amount,
      updatedDetail.year,
      updatedDetail.month,
      updatedDetail.price_sell,
      updatedDetail.is_big_car ? 1 : 0,
      updatedDetail.id,
    ],
  })

  if (result.success) {
    await updateBillAmount(selectedBill.value.id)
    await fetchBuyDetails(selectedBill.value.id)
    await fetchBuyBills()
  } else {
    alert(t('failed_update_detail'))
    console.error('Error updating detail:', result.error)
  }
}

// Get base path helper (same logic as router and useApi)
const getBasePath = () => {
  let baseUrl = import.meta.env.BASE_URL || './'
  if (baseUrl === './' || baseUrl.startsWith('./')) {
    const pathname = window.location.pathname
    // If pathname is like '/mig_26/login', extract '/mig_26/'
    // If pathname is like '/login', use '/'
    const match = pathname.match(/^(\/[^/]+\/)/)
    return match ? match[1] : '/'
  }
  return baseUrl
}

// Replace handlePaymentClick with new method to open in new tab
const openPayments = (bill) => {
  // No double-click prevention needed for opening new tab
  const basePath = getBasePath()
  window.open(`${basePath}buy-payments/${bill.id}`, '_blank')
}

// Add task handling methods
const openTaskForBill = (bill) => {
  if (isProcessingTask.value) return // Prevent double-click
  selectedBillForTask.value = bill
  showTaskForm.value = true
}

const handleTaskCreated = () => {
  isProcessingTask.value = false
  showTaskForm.value = false
  // Don't set selectedBillForTask to null to avoid prop validation errors
  // Optionally refresh data if needed
}

const handleWarehouseChanged = (updatedCar) => {
  // Refresh the warehouse counts in BuyBillsTable when a car's warehouse status changes
  if (buyBillsTableRef.value) {
    // Call the fetchWarehouseCounts function to refresh the warehouse counts
    buyBillsTableRef.value.fetchWarehouseCounts()
  }
}

const showEditNotesDialog = ref(false)
const notesEditValue = ref('')
const isSavingNotes = ref(false)

const openEditNotesDialog = () => {
  if (!selectedBill.value) return
  notesEditValue.value = selectedBill.value.notes || ''
  showEditNotesDialog.value = true
}

const saveNotes = async (newNotes) => {
  if (!selectedBill.value) return
  isSavingNotes.value = true
  try {
    const result = await callApi({
      query: 'UPDATE buy_bill SET notes = ? WHERE id = ?',
      params: [newNotes, selectedBill.value.id],
    })
    if (result.success) {
      selectedBill.value.notes = newNotes
      await fetchBuyBills()
      showEditNotesDialog.value = false
    } else {
      alert(t('failed_update_notes'))
    }
  } catch (err) {
    alert(t('failed_update_notes'))
  } finally {
    isSavingNotes.value = false
  }
}
</script>

<template>
  <div class="buy-view">
    <div class="header">
      <h2>{{ t('buyView.buyManagement') }}</h2>
      <button @click="openAddDialog" class="add-btn">{{ t('buyView.addNewPurchase') }}</button>
    </div>

    <div class="content">
      <div class="master-detail-vertical">
        <BuyBillsTable
          ref="buyBillsTableRef"
          :buyBills="buyBills"
          :selectedBill="selectedBill"
          :loading="isLoadingBills"
          @select-bill="handleSelectBill"
        >
          <!-- Toolbar actions -->
          <template #actions="{ bill }">
            <button @click.stop="openPayments(bill)" class="payment-btn">
              {{ t('buyView.payments') }}
            </button>
            <button
              @click.stop="openEditDialog(bill)"
              class="action-btn edit-btn"
              :disabled="bill.is_stock_updated"
            >
              <i class="fas fa-edit"></i>
              {{ t('buyView.edit') }}
            </button>
            <button
              v-if="false"
              @click.stop="handleUpdateStock(bill)"
              class="action-btn update-btn"
              :disabled="!canUpdateStock(bill)"
            >
              {{ t('buyView.updateStock') }}
            </button>
            <button
              v-if="isAdmin"
              @click.stop="handleDeleteBill(bill)"
              class="action-btn delete-btn"
              :disabled="bill.is_stock_updated || isDeletingBill"
            >
              <i v-if="isDeletingBill" class="fas fa-spinner fa-spin"></i>
              {{ isDeletingBill ? t('buyView.deleting') : t('buyView.delete') }}
            </button>
            <button
              @click.stop="openTaskForBill(bill)"
              class="action-btn task-btn"
              :disabled="isProcessingTask"
              :title="t('buyView.addNewTask')"
            >
              <i v-if="isProcessingTask" class="fas fa-spinner fa-spin"></i>
              <i v-else class="fas fa-tasks"></i>
            </button>
            <button
              v-if="bill"
              class="action-btn notes-btn"
              @click.stop="openEditNotesDialog"
              :title="t('buy.billsTable.editNotes')"
            >
              <i class="fas fa-sticky-note"></i>
              {{ t('buy.billsTable.editNotes') }}
            </button>
          </template>
        </BuyBillsTable>

        <!-- Purchase Details: Only show if bill is pending -->
        <BuyDetailsTable
          ref="buyDetailsTableRef"
          id="buy-details-table"
          v-if="selectedBill && isBillPending"
          :buyDetails="buyDetails"
          :isAdmin="isAdmin"
          @add-detail="showAddDetailDialog = true"
          @delete-detail="handleDeleteDetail"
          @update-detail="handleUpdateDetail"
          @stock-updated="handleStockUpdated"
          @cars-created="handleCarsCreated"
        />

        <!-- Cars in Purchase Bill: Only show if bill is updated -->
        <div
          v-if="selectedBill && selectedBill.is_stock_updated == 1"
          class="cars-section"
          id="car-stock-table"
        >
          <div class="section-header">
            <h3>
              <i class="fas fa-car"></i>
              {{ t('buyView.carsInPurchaseBill') }} #{{ selectedBill.id }}
            </h3>
            <p class="section-description">
              {{ t('buyView.showingAllCarsAssociatedWithPurchaseBill') }}
            </p>
          </div>
          <CarStockTable
            ref="carStockTableRef"
            :buyBillId="selectedBill.id"
            :filters="{ basic: '', advanced: {} }"
            @warehouse-changed="handleWarehouseChanged"
          />
        </div>
      </div>
    </div>

    <!-- Dialogs -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>{{ t('buyView.addNewPurchase') }}</h3>
        <form @submit.prevent="addPurchase">
          <div class="form-group">
            <label>{{ t('buyView.supplier') }}</label>
            <select v-model="newPurchase.id_supplier" required>
              <option value="">{{ t('buyView.selectSupplier') }}</option>
              <option v-for="supplier in suppliers" :key="supplier.id" :value="supplier.id">
                {{ supplier.name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>{{ t('buyView.date') }}</label>
            <input type="datetime-local" v-model="newPurchase.date_buy" required />
          </div>

          <div class="form-group">
            <label>{{ t('buyView.billReference') }} <span class="required">*</span></label>
            <input
              type="text"
              v-model="newPurchase.bill_ref"
              :placeholder="t('buyView.enterBillReferenceNumber')"
              required
            />
          </div>

          <div class="form-group">
            <label
              >{{ t('buyView.piDocument') }} ({{ t('buyView.pdfOrImage') }})
              <span class="required">*</span></label
            >
            <input
              type="file"
              @change="handlePiFileChange"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp"
              required
            />
            <a
              v-if="newPurchase.pi_path"
              :href="getFileUrl(newPurchase.pi_path)"
              target="_blank"
              class="current-file-link"
            >
              {{ t('buyView.viewCurrentPiDocument') }}
            </a>
          </div>

          <div v-if="isAdmin" class="form-group checkbox">
            <label>
              <input
                type="checkbox"
                v-model="newPurchase.is_ordered"
                :true-value="1"
                :false-value="0"
              />
              {{ t('buyView.orderConfirmed') }}
            </label>
            <small class="help-text">{{ t('buyView.markThisPurchaseAsConfirmedOrder') }}</small>
          </div>

          <div class="form-group">
            <label>{{ t('buyView.notes') }}</label>
            <textarea
              v-model="newPurchase.notes"
              rows="3"
              :placeholder="t('buyView.enterNotes')"
            ></textarea>
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="showAddDialog = false" class="cancel-btn">
              {{ t('buyView.cancel') }}
            </button>
            <button type="submit" class="submit-btn" :disabled="isSubmittingPurchase">
              <span v-if="isSubmittingPurchase" class="spinner"></span>
              {{ isSubmittingPurchase ? t('buyView.adding') : t('buyView.addPurchase') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Edit Purchase Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>{{ t('buyView.editPurchase') }}</h3>
        <form @submit.prevent="updatePurchase">
          <div class="form-group">
            <label>{{ t('buyView.supplier') }}</label>
            <select v-model="newPurchase.id_supplier" required>
              <option value="">{{ t('buyView.selectSupplier') }}</option>
              <option v-for="supplier in suppliers" :key="supplier.id" :value="supplier.id">
                {{ supplier.name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>{{ t('buyView.date') }}</label>
            <input type="datetime-local" v-model="newPurchase.date_buy" required />
          </div>

          <div class="form-group">
            <label>{{ t('buyView.billReference') }} <span class="required">*</span></label>
            <input
              type="text"
              v-model="newPurchase.bill_ref"
              :placeholder="t('buyView.enterBillReferenceNumber')"
              required
            />
          </div>

          <div class="form-group">
            <label
              >{{ t('buyView.piDocument') }} ({{ t('buyView.pdfOrImage') }})
              <span class="required">*</span></label
            >
            <input
              type="file"
              @change="handlePiFileChange"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp"
            />
            <a
              v-if="newPurchase.pi_path"
              :href="getFileUrl(newPurchase.pi_path)"
              target="_blank"
              class="current-file-link"
            >
              {{ t('buyView.viewCurrentPiDocument') }}
            </a>
            <div v-if="!newPurchase.pi_path && !newPurchase.pi_file" class="validation-message">
              {{ t('buyView.piDocumentRequired') }}
            </div>
          </div>

          <div v-if="isAdmin" class="form-group checkbox">
            <label>
              <input
                type="checkbox"
                v-model="newPurchase.is_ordered"
                :true-value="1"
                :false-value="0"
              />
              {{ t('buyView.orderConfirmed') }}
            </label>
            <small class="help-text">{{ t('buyView.markThisPurchaseAsConfirmedOrder') }}</small>
          </div>

          <div class="form-group">
            <label>{{ t('buyView.notes') }}</label>
            <textarea
              v-model="newPurchase.notes"
              rows="3"
              :placeholder="t('buyView.enterNotes')"
            ></textarea>
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="showEditDialog = false" class="cancel-btn">
              {{ t('buyView.cancel') }}
            </button>
            <button type="submit" class="submit-btn" :disabled="isSubmittingPurchase">
              <span v-if="isSubmittingPurchase" class="spinner"></span>
              {{ isSubmittingPurchase ? t('buyView.updating') : t('buyView.updatePurchase') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showAddDetailDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>{{ t('buyView.addPurchaseDetail') }}</h3>
        <form @submit.prevent="addDetail">
          <div class="form-group">
            <label>{{ t('buyView.car') }}</label>
            <select v-model="newDetail.id_car_name" required>
              <option value="">{{ t('buyView.selectCar') }}</option>
              <option v-for="car in cars" :key="car.id" :value="car.id">
                {{ car.car_name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>{{ t('buyView.color') }}</label>
            <select v-model="newDetail.id_color" required>
              <option value="">{{ t('buyView.selectColor') }}</option>
              <option v-for="color in colors" :key="color.id" :value="color.id">
                {{ color.color }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>{{ t('buyView.quantity') }}</label>
            <input type="number" v-model="newDetail.QTY" min="1" required />
          </div>

          <div class="form-group">
            <label>{{ t('buyView.amount') }}</label>
            <input type="number" v-model="newDetail.amount" step="0.01" required />
          </div>

          <div class="form-row">
            <div class="form-group half">
              <label>{{ t('buyView.year') }}</label>
              <input type="number" v-model="newDetail.year" required />
            </div>

            <div class="form-group half">
              <label>{{ t('buyView.month') }}</label>
              <input type="number" v-model="newDetail.month" min="1" max="12" required />
            </div>
          </div>

          <div class="form-group">
            <label>{{ t('buyView.priceSell') }}</label>
            <input type="number" v-model="newDetail.price_sell" step="0.01" required />
          </div>

          <div class="form-group checkbox">
            <label>
              <input type="checkbox" v-model="newDetail.is_used_car" />
              {{ t('buyView.usedCar') }}
            </label>
          </div>

          <div class="form-group checkbox">
            <label>
              <input type="checkbox" v-model="newDetail.is_big_car" />
              {{ t('buyView.bigCar') }}
            </label>
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="showAddDetailDialog = false" class="cancel-btn">
              {{ t('buyView.cancel') }}
            </button>
            <button type="submit" class="submit-btn" :disabled="isSubmittingDetail">
              <span v-if="isSubmittingDetail" class="spinner"></span>
              {{ isSubmittingDetail ? t('buyView.adding') : t('buyView.addDetail') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Task Form -->
    <TaskForm
      v-if="selectedBillForTask"
      :entityType="'buy'"
      :entityData="selectedBillForTask"
      :isVisible="showTaskForm"
      @task-created="handleTaskCreated"
      @cancel="showTaskForm = false"
    />

    <EditNotesDialog
      v-if="showEditNotesDialog"
      v-model="notesEditValue"
      :visible="showEditNotesDialog"
      :title="t('buy.billsTable.editNotes')"
      @save="saveNotes"
      @cancel="showEditNotesDialog = false"
    />
  </div>
</template>

<style scoped>
.buy-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
}

.master-detail-vertical {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Make only the table scrollable with height limit for 10 rows */
.master-detail-vertical :deep(.data-table) {
  max-height: 400px; /* Height for approximately 10 rows */
  overflow-y: auto;
  display: block;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.master-detail-vertical :deep(.data-table thead) {
  position: sticky;
  top: 0;
  background-color: #f8fafc;
  z-index: 10;
  display: table;
  width: 100%;
  table-layout: fixed;
}

.master-detail-vertical :deep(.data-table tbody) {
  display: table;
  width: 100%;
  table-layout: fixed;
}

.master-detail-vertical :deep(.data-table thead th) {
  background-color: #f8fafc;
}

.master-section {
  width: 100%;
}

.detail-section {
  width: 100%;
  border-top: 1px solid #e5e7eb;
  padding-top: 20px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.data-table tbody tr {
  cursor: pointer;
  transition: background-color 0.2s;
}

.data-table tbody tr:hover {
  background-color: #f3f4f6;
}

.data-table tbody tr.selected {
  background-color: #e5e7eb;
}

.empty-detail {
  text-align: center;
  color: #6b7280;
  padding: 40px;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.add-btn:hover {
  background-color: #059669;
}

h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #374151;
}

.selected-id {
  background-color: #f0fdf4;
  color: #047857;
  padding: 8px 16px;
  border-radius: 4px;
  margin-bottom: 16px;
  border: 1px solid #10b981;
  font-weight: 500;
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
  padding: 24px;
  border-radius: 8px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #10b981;
  ring: 2px #10b981;
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn {
  padding: 8px 16px;
  background-color: #e5e7eb;
  color: #374151;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.cancel-btn:hover {
  background-color: #d1d5db;
}

.submit-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.submit-btn:hover:not(:disabled) {
  background-color: #059669;
}

.submit-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-group.half {
  flex: 1;
}

.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  resize: vertical;
}

.form-group.checkbox {
  display: flex;
  align-items: center;
}

.form-group.checkbox input {
  width: auto;
  margin-right: 8px;
}

.form-group.checkbox label {
  display: flex;
  align-items: center;
  margin: 0;
}

.payment-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s;
}

.payment-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.payment-btn:before {
  content: '💰';
  font-size: 1rem;
}

.payment-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.update-btn {
  background-color: #059669;
}

.update-btn:hover:not(:disabled) {
  background-color: #047857;
}

.update-btn:before {
  content: '📦';
  font-size: 1rem;
}

.delete-btn {
  background-color: #dc3545;
  color: white;
}

.delete-btn:hover {
  background-color: #c82333;
}

.delete-btn:before {
  content: '🗑';
  font-size: 1rem;
}

.edit-btn {
  background-color: #3b82f6;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.edit-btn:before {
  content: '✏️';
  font-size: 1rem;
}

.current-file-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.875rem;
  margin-top: 0.5rem;
  display: inline-block;
}

.current-file-link:hover {
  text-decoration: underline;
}

.required {
  color: #ef4444;
  font-weight: 600;
}

.validation-message {
  color: #ef4444;
  font-size: 0.875rem;
  margin-top: 0.25rem;
  font-style: italic;
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

.help-text {
  color: #6b7280;
  font-size: 0.875rem;
  margin-top: 0.25rem;
  margin-left: 24px;
  display: block;
}

.task-btn {
  background-color: #8b5cf6;
  color: white;
}

.task-btn:hover {
  background-color: #7c3aed;
}

.task-btn i {
  font-size: 0.9rem;
}

/* Cars section styles */
.cars-section {
  margin-top: 20px;
  border-top: 2px solid #e5e7eb;
  padding-top: 20px;
}

.section-header {
  margin-bottom: 20px;
}

.section-header h3 {
  color: #1e293b;
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header h3 i {
  color: #3b82f6;
}

.section-description {
  color: #64748b;
  font-size: 0.875rem;
  margin: 0;
}
</style>
