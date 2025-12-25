<script setup>
import { ref, onMounted, computed, watch, onUnmounted, nextTick } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useApi } from '../../composables/useApi'
import VinEditForm from './VinEditForm.vue'
import CarFilesManagement from './CarFilesManagement.vue'
import CarPortsEditForm from './CarPortsEditForm.vue'
import CarMoneyEditForm from './CarMoneyEditForm.vue'
import CarWarehouseForm from './CarWarehouseForm.vue'
import CarDocumentsForm from './CarDocumentsForm.vue'
import CarLoadForm from './CarLoadForm.vue'
import TaskForm from './TaskForm.vue'
import CarStockSwitchBuyBill from './CarStockSwitchBuyBill.vue'
import CarStockToolbar from './CarStockToolbar.vue'
import CarStockPrintOptions from './CarStockPrintOptions.vue'
import CarStockPrintReport from './CarStockPrintReport.vue'
import VinAssignmentModal from './VinAssignmentModal.vue'
import CarPortsBulkEditForm from './CarPortsBulkEditForm.vue'

import CarNotesBulkEditForm from './CarNotesBulkEditForm.vue'
import CarColorBulkEditForm from './CarColorBulkEditForm.vue'
import CarExportLicenseBulkEditForm from './CarExportLicenseBulkEditForm.vue'
import SaveSelectionForm from './SaveSelectionForm.vue'
import SendSelectionForm from './SendSelectionForm.vue'
import ShowSelectionsModal from './ShowSelectionsModal.vue'

import { useRouter } from 'vue-router'

const { t, locale } = useEnhancedI18n()
const isDropdownOpen = ref({})
const props = defineProps({
  cars: Array,
  onEdit: {
    type: Function,
    default: () => {},
  },
  buyBillId: {
    type: [Number, String],
    default: null,
  },
  alertType: {
    type: String,
    default: null, // 'unloaded', 'not_arrived', 'no_license', 'no_docs'
  },
  alertDays: {
    type: Number,
    default: null,
  },
  clientId: {
    type: [Number, String],
    default: null,
  },
  showClientFilter: {
    type: Boolean,
    default: false,
  },
  filters: {
    type: Object,
    default: () => ({
      basic: '',
      advanced: {
        id: '',
        car_name: '',
        color: '',
        vin: '',
        loading_port: '',
        discharge_port: '',
        freight_min: '',
        freight_max: '',
        price_min: '',
        price_max: '',
        loading_date_from: '',
        loading_date_to: '',
        buy_date_from: '',
        buy_date_to: '',
        sold_date_from: '',
        sold_date_to: '',
        status: '',
        client: '',
        client_id_no: '',
        warehouse: '',
        container_ref: '',
        export_lisence_ref: '',
        has_bl: false,
        freight_paid: false,
        has_supplier_docs: false,
        in_warehouse: false,
        has_export_license: false,
        documents_sent: false,
        is_loaded: false,
        has_vin: false,
        loading_status: '',
        documents_status: '',
        bl_status: '',
        warehouse_status: '',
        bill_ref: '',
        sell_bill_ref: '',
        tmp_client_status: '',
        whole_sale_status: '',
      },
    }),
  },
})

// Add watcher for filters
watch(
  () => props.filters,
  (newFilters) => {
    fetchCarsStock()
  },
  { deep: true },
)

// Add watcher for buyBillId
watch(
  () => props.buyBillId,
  (newBuyBillId) => {
    // Apply filters to in-memory backup
    fetchCarsStock()
  },
)

const user = ref(null)
const can_edit_cars_prop = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_prop')
})

const emit = defineEmits(['refresh', 'warehouse-changed'])

const {
  callApi,
  getFileUrl,
  getAssets,
  getCarFiles,
  transferPhysicalCopy,
  getUsersForTransfer,
  getCustomClearanceAgents,
} = useApi()
const letterHeadUrl = ref(null)
const cars = ref([])
const loading = ref(true)
const error = ref(null)
const carFilesCountMap = ref({}) // Map of car_id -> { count: number, hasMissingRequired: boolean }
const isProcessing = ref({
  edit: false,
  vin: false,
  files: false,
  ports: false,
  money: false,
  warehouse: false,
  documents: false,
  load: false,
  task: false,
  notes: false,
  combine: false,
  payment: false,
})

// Add new refs and computed properties for VIN editing
const showVinEditForm = ref(false)
const selectedCar = ref(null)

const can_edit_vin = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_vin')
})

// Add new refs and computed properties for file upload
const showFilesUploadForm = ref(false)

const can_upload_car_files = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_upload_car_files')
})

// Add new refs and computed properties for ports editing
const showPortsEditForm = ref(false)

const can_edit_cars_ports = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_ports')
})

// Add new refs and computed properties for money editing
const showMoneyEditForm = ref(false)

const can_edit_car_money = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_car_money')
})

// Add new refs for warehouse form
const showWarehouseForm = ref(false)

// Add new computed property for warehouse permission
const can_edit_warehouse = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_warehouse')
})

const can_confirm_payment = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_confirm_payment')
})

// Add to the data/refs section
const showDocumentsForm = ref(false)
const selectedCarForDocuments = ref(null)

// Add this computed property after other permission computeds
const can_edit_car_documents = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_car_documents')
})

// Add new refs for load form
const showLoadForm = ref(false)
const selectedCarForLoad = ref(null)

// Add new computed property for load permission
const can_load_car = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_load_car')
})

// Add new refs for task form
const showTaskForm = ref(false)
const selectedCarForTask = ref(null)

// Add new refs for switch buy bill form
const showSwitchBuyBillForm = ref(false)
const selectedCarForSwitchBuyBill = ref(null)

// Add new refs for VIN assignment modal
const showVinAssignmentModal = ref(false)

// Batch transfer modal state
const showBatchTransferModal = ref(false)
const batchTransferToUserId = ref(null)
const batchTransferNotes = ref('')
const batchTransferExpectedReturn = ref('')
const batchTransferAvailableUsers = ref([])
const batchTransferFiles = ref([])
const isBatchTransferring = ref(false)

// Batch checkout modal state
const showBatchCheckoutModal = ref(false)
const batchCheckoutType = ref('user') // 'user', 'client', or 'custom_clearance_agent'
const batchCheckoutUserId = ref(null)
const batchCheckoutAgentId = ref(null)
const batchCheckoutClientId = ref(null)
const batchCheckoutClientName = ref('')
const batchCheckoutClientContact = ref('')
const batchCheckoutNotes = ref('')
const batchCheckoutAvailableUsers = ref([])
const batchCheckoutAvailableAgents = ref([])
const batchCheckoutAvailableClients = ref([])
const batchCheckoutFiles = ref([])
const isBatchCheckouting = ref(false)

// Add new refs for ports bulk edit form
const showPortsBulkEditForm = ref(false)

// Add new refs for warehouse bulk edit form

// Add new refs for notes bulk edit form
const showNotesBulkEditForm = ref(false)

// Add computed property for admin permission
const isAdmin = computed(() => {
  if (!user.value) return false
  return user.value.role_id === 1
})

// Computed property for buy bill ref header translation
const buyBillRefHeader = computed(() => {
  const translation = t('carStock.buy_bill_ref')
  // If translation returns the key or locale code, use fallback
  if (
    translation === 'carStock.buy_bill_ref' ||
    translation === 'en' ||
    translation === locale.value
  ) {
    return 'Buy Bill Ref'
  }
  return translation
})

// Computed property for combining cars translation
const combiningCarsText = computed(() => {
  const translation = t('carStock.combining_cars')
  // If translation returns the key or locale code, use fallback
  if (
    translation === 'carStock.combining_cars' ||
    translation === 'en' ||
    translation === locale.value
  ) {
    return 'Combining cars...'
  }
  return translation
})

// Add to the data/refs section
const showAdvancedFilter = ref(false)
const advancedFilters = ref({})

// Add sorting state
const sortConfig = ref({
  key: 'id',
  direction: 'desc',
})

// Combine mode state
const isCombineMode = ref(false)
const combineModeType = ref(null) // 'buy_ref' or 'container'
const previousSortConfig = ref(null)

// Add sort function
const toggleSort = (key) => {
  if (sortConfig.value.key === key) {
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    sortConfig.value.key = key
    sortConfig.value.direction = 'asc'
  }
}

// Add sorted cars computed property
const sortedCars = computed(() => {
  if (!cars.value) return []

  // If in combine mode, sort by the appropriate field
  if (isCombineMode.value) {
    if (combineModeType.value === 'buy_ref') {
      return [...cars.value].sort((a, b) => {
        const aBuyRef = a.buy_bill_ref || ''
        const bBuyRef = b.buy_bill_ref || ''

        // First sort by buy_bill_ref
        if (aBuyRef !== bBuyRef) {
          if (!aBuyRef) return 1
          if (!bBuyRef) return -1
          return aBuyRef.localeCompare(bBuyRef)
        }

        // Then sort by id
        return a.id - b.id
      })
    } else if (combineModeType.value === 'container') {
      return [...cars.value].sort((a, b) => {
        const aContainer = a.container_ref || ''
        const bContainer = b.container_ref || ''

        // First sort by container_ref
        if (aContainer !== bContainer) {
          if (!aContainer) return 1
          if (!bContainer) return -1
          return aContainer.localeCompare(bContainer)
        }

        // Then sort by id
        return a.id - b.id
      })
    }
  }

  return [...cars.value].sort((a, b) => {
    let aVal = a[sortConfig.value.key]
    let bVal = b[sortConfig.value.key]

    // Handle numeric values
    if (['id', 'price_cell', 'freight', 'rate', 'buy_price'].includes(sortConfig.value.key)) {
      aVal = parseFloat(aVal) || 0
      bVal = parseFloat(bVal) || 0
    }

    // Handle date values
    if (['date_loding', 'date_buy', 'date_sell'].includes(sortConfig.value.key)) {
      aVal = aVal ? new Date(aVal).getTime() : 0
      bVal = bVal ? new Date(bVal).getTime() : 0
    }

    // Handle status field (Sold comes after Available alphabetically, but we want Available first)
    if (sortConfig.value.key === 'status') {
      // Define status priority (Available = 1, Sold = 2)
      const getStatusPriority = (status) => {
        if (status === 'Available') return 1
        if (status === 'Sold') return 2
        return 3 // for any other values
      }

      const aPriority = getStatusPriority(aVal)
      const bPriority = getStatusPriority(bVal)

      if (sortConfig.value.direction === 'asc') {
        return aPriority - bPriority
      } else {
        return bPriority - aPriority
      }
    }

    // Handle null values
    if (aVal == null) return 1
    if (bVal == null) return -1

    // Compare values
    if (sortConfig.value.direction === 'asc') {
      return aVal > bVal ? 1 : aVal < bVal ? -1 : 0
    } else {
      return aVal < bVal ? 1 : aVal > bVal ? -1 : 0
    }
  })
})

// Add selection functionality
const selectedCars = ref(new Set())
const selectAll = ref(false)

// Function to toggle individual car selection
const toggleCarSelection = (carId) => {
  if (selectedCars.value.has(carId)) {
    selectedCars.value.delete(carId)
  } else {
    selectedCars.value.add(carId)
  }
  updateSelectAllState()
}

// Function to select all cars
const selectAllCars = () => {
  if (selectAll.value) {
    selectedCars.value.clear()
    selectAll.value = false
  } else {
    sortedCars.value.forEach((car) => {
      selectedCars.value.add(car.id)
    })
    selectAll.value = true
  }
}

// Function to update select all state
const updateSelectAllState = () => {
  if (sortedCars.value.length === 0) {
    selectAll.value = false
    return
  }

  const allSelected = sortedCars.value.every((car) => selectedCars.value.has(car.id))
  const someSelected = sortedCars.value.some((car) => selectedCars.value.has(car.id))

  if (allSelected) {
    selectAll.value = true
  } else if (someSelected) {
    selectAll.value = false
  } else {
    selectAll.value = false
  }
}

// Watch for changes in sortedCars to update select all state
watch(
  sortedCars,
  () => {
    updateSelectAllState()
  },
  { immediate: true },
)

// Print options modal state
const showPrintOptions = ref(false)
const printOptionsActionType = ref('print') // 'print' or 'loading-order'

// Print handler functions
const handlePrintSelected = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_printing'))
    return
  }

  printOptionsActionType.value = 'print'
  showPrintOptions.value = true
}

const handlePrintOptionsClose = () => {
  showPrintOptions.value = false
}

const handlePrintWithOptions = async (printData) => {
  const { columns, cars, subject, coreContent, groupBy } = printData

  // Generate print content based on action type
  let title = subject || ''
  let contentBeforeTable = ''
  let contentAfterTable = ''

  if (printOptionsActionType.value === 'print') {
    title = subject || 'Car Stock Report'
    contentBeforeTable = `
      ${coreContent ? `<div style="margin: 20px 0; padding: 15px; background-color: #f8f9fa; border-left: 4px solid #3b82f6; border-radius: 4px; white-space: pre-wrap;">${coreContent}</div>` : ''}
    `
    contentAfterTable = `
      <p><strong>Total Cars:</strong> ${cars.length}</p>
      <p><strong>Report Type:</strong> Stock Inventory</p>
    `
  } else if (printOptionsActionType.value === 'loading-order') {
    title = subject || 'Loading Order Report'
    contentBeforeTable = `
      <p>This loading order contains ${cars.length} car${cars.length === 1 ? '' : 's'} to be loaded.</p>
      <p>Loading order generated on ${new Date().toLocaleDateString()}.</p>
      <div style="margin: 20px 0; padding: 15px; background-color: #f8f9fa; border-left: 4px solid #059669; border-radius: 4px;">
        Core Content: ${coreContent || 'NO CORE CONTENT FOUND'}
      </div>
    `
    contentAfterTable = `
      <p><strong>Total Cars for Loading:</strong> ${cars.length}</p>
      <p><strong>Instructions:</strong> Please ensure all vehicles are properly Loaded as soon as possible.</p>
      <div style="margin-top: 40px; border-top: 1px solid #333; padding-top: 20px;">
        <div style="display: flex; justify-content: space-between; align-items: flex-end;">
          <div style="text-align: center;">
            <div style="border-bottom: 1px solid #333; width: 200px; height: 40px; margin-bottom: 5px;"></div>
            <p style="margin: 0; font-size: 12px;"><strong>Receiver Signature</strong></p>
          </div>
          <div style="text-align: center;">
            <div style="border-bottom: 1px solid #333; width: 200px; height: 40px; margin-bottom: 5px;"></div>
            <p style="margin: 0; font-size: 12px;"><strong>Date</strong></p>
          </div>
        </div>
      </div>
    `
  }

  // Create and print the report
  await printReport({
    title,
    cars,
    columns,
    contentBeforeTable,
    contentAfterTable,
    groupBy: groupBy || '',
  })

  // Close the modal after handling the print event
  showPrintOptions.value = false
}

const handleLoadingOrderFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_loading_order'))
    return
  }

  printOptionsActionType.value = 'loading-order'
  showPrintOptions.value = true
}

const handleLoadingOrderWithOptions = async (data) => {
  const { columns, cars, subject, coreContent } = data

  // Generate loading order content based on action type
  let title = subject || 'Loading Order Report'
  let contentBeforeTable = `
    ${coreContent ? `<div style="margin: 20px 0; padding: 15px; background-color: #f8f9fa; border-left: 4px solid #059669; border-radius: 4px; white-space: pre-wrap;">${coreContent}</div>` : ''}
  `
  let contentAfterTable = `
    <p><strong>Total Cars for Loading:</strong> ${cars.length}</p>
    <p><strong>Instructions:</strong> Please ensure all vehicles are properly Loaded as soon as possible.</p>
    <div style="margin-top: 40px; border-top: 1px solid #333; padding-top: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: flex-end;">
        <div style="text-align: center;">
          <div style="border-bottom: 1px solid #333; width: 200px; height: 40px; margin-bottom: 5px;"></div>
          <p style="margin: 0; font-size: 12px;"><strong>Receiver Signature</strong></p>
        </div>
        <div style="text-align: center;">
          <div style="border-bottom: 1px solid #333; width: 200px; height: 40px; margin-bottom: 5px;"></div>
          <p style="margin: 0; font-size: 12px;"><strong>Date</strong></p>
        </div>
      </div>
    </div>
  `

  // Create and print the loading order report
  await printReport({
    title,
    cars,
    columns,
    contentBeforeTable,
    contentAfterTable,
  })

  // Close the modal after handling the loading order event
  showPrintOptions.value = false
}

const handleVinFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_vin_assignment'))
    return
  }

  showVinAssignmentModal.value = true
}

const printReport = async (reportData) => {
  const {
    title,
    cars: reportCars,
    columns,
    contentBeforeTable,
    contentAfterTable,
    groupBy = '',
  } = reportData

  // Load assets if not already loaded
  if (!letterHeadUrl.value) {
    try {
      const assets = await getAssets()
      if (assets && assets.letter_head) {
        letterHeadUrl.value = assets.letter_head
      }
    } catch (err) {
      // Failed to load assets, using default
    }
  }

  // Generate REF number
  const generateRef = () => {
    // Get user name first 3 characters
    const userName = user.value?.name || 'USR'
    const userPrefix = userName.substring(0, 3).toUpperCase()

    // Get button caption first letter of each word
    let buttonCaption = ''
    if (printOptionsActionType.value === 'print') {
      buttonCaption = 'Print Selected'
    } else if (printOptionsActionType.value === 'loading-order') {
      buttonCaption = 'Loading Order'
    }

    const buttonPrefix = buttonCaption
      .split(' ')
      .map((word) => word.charAt(0))
      .join('')
      .toUpperCase()

    // Get current date in dd+mm+yy format
    const now = new Date()
    const day = now.getDate().toString().padStart(2, '0')
    const month = (now.getMonth() + 1).toString().padStart(2, '0')
    const year = now.getFullYear().toString().slice(-2)
    const dateStr = `${day}${month}${year}`

    // Get sequential number from localStorage
    const storageKey = `ref_sequence_${printOptionsActionType.value}_${dateStr}`
    let sequence = parseInt(localStorage.getItem(storageKey) || '0') + 1
    localStorage.setItem(storageKey, sequence.toString())

    // Format sequence with leading zeros
    const sequenceStr = sequence.toString().padStart(3, '0')

    return `${userPrefix}${buttonPrefix}${dateStr}-${sequenceStr}`
  }

  // Create a new window for printing
  const printWindow = window.open('', '_blank', 'width=800,height=600')

  // Create the HTML content for the report
  const reportHTML = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>${title}</title>
      <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
        .print-report { max-width: 210mm; margin: 0 auto; }
        .report-header { text-align: center; margin-bottom: 20px; }
        .letter-head { max-width: 100%; height: auto; max-height: 120px; }
        .report-date { text-align: right; margin-bottom: 20px; font-size: 14px; color: #666; float: right; clear: both; width: 100%; }
        .report-ref { text-align: right; margin-bottom: 20px; font-size: 14px; color: #666; float: right; clear: both; width: 100%; }
        .report-title { text-align: center; margin-bottom: 30px; font-size: 24px; font-weight: bold; color: #333; text-transform: uppercase; clear: both; }
        .content-before-table { margin-bottom: 20px; line-height: 1.6; font-size: 14px; }
        .table-container { margin-bottom: 20px; overflow-x: auto; }
        .report-table { width: 100%; border-collapse: collapse; font-size: 12px; }
        .table-header { background-color: #f8f9fa; border: 1px solid #dee2e6; padding: 8px 12px; text-align: left; font-weight: bold; color: #495057; white-space: nowrap; }
        .table-row:nth-child(even) { background-color: #f8f9fa; }
        .table-cell { border: 1px solid #dee2e6; padding: 8px 12px; text-align: left; vertical-align: top; }
        .group-header-row { background-color: #e0f2fe; }
        .group-header-cell { background-color: #e0f2fe; border: 2px solid #0ea5e9; padding: 10px 12px; font-weight: bold; color: #0c4a6e; }
        .content-after-table { margin-top: 20px; line-height: 1.6; font-size: 14px; }
        @media print { body { padding: 0; margin: 0; } .report-table { font-size: 10px; } .table-header, .table-cell { padding: 6px 8px; } .letter-head { max-height: 80px; } .report-title { font-size: 20px; margin-bottom: 20px; } }
      </style>
    </head>
    <body>
      <div class="print-report">
        <div class="report-header">
          <img 
            src="${letterHeadUrl.value || ''}" 
            alt="Letter Head" 
            class="letter-head"
            style="max-width: 100%; height: auto; max-height: 120px;"
            onerror="this.style.display='none'"
          />
        </div>
        <div class="report-date">
          <span><strong>Date:</strong> ${new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}</span>
        </div>
        <div class="report-ref">
          <span><strong>REF:</strong> ${generateRef()}</span>
        </div>
        <h1 class="report-title">${title}</h1>
        ${contentBeforeTable ? `<div class="content-before-table">${contentBeforeTable}</div>` : ''}
        <div class="table-container">
          <table class="report-table">
            <thead>
              <tr>
                ${columns.map((col) => `<th class="table-header">${col.label}</th>`).join('')}
              </tr>
            </thead>
            <tbody>
              ${reportCars
                .map(
                  (car) => `
                <tr class="table-row">
                  ${columns
                    .map((col) => {
                      const value = car[col.key]
                      let displayValue = '-'
                      if (value !== null && value !== undefined) {
                        if (col.key === 'client_id_picture' && value) {
                          displayValue = `<img src="${getFileUrl(value)}" alt="Client ID" style="max-width: 100px; max-height: 60px; object-fit: contain;" />`
                        } else if (col.key.includes('date') && value) {
                          displayValue = new Date(value).toLocaleDateString()
                        } else if (
                          ['price_cell', 'freight', 'rate', 'cfr_usd', 'cfr_dza'].includes(
                            col.key,
                          ) &&
                          value
                        ) {
                          displayValue = parseFloat(value).toLocaleString()
                        } else {
                          displayValue = value.toString()
                        }
                      }
                      return `<td class="table-cell">${displayValue}</td>`
                    })
                    .join('')}
                </tr>
              `,
                )
                .join('')}
            </tbody>
          </table>
        </div>
        ${contentAfterTable ? `<div class="content-after-table">${contentAfterTable}</div>` : ''}
      </div>
    </body>
    </html>
  `

  printWindow.document.write(reportHTML)
  printWindow.document.close()

  // Wait for content to load then print
  printWindow.onload = () => {
    printWindow.print()
    printWindow.close()
  }
}

// Add new refs for teleport dropdown
const teleportDropdown = ref({
  isOpen: false,
  carId: null,
  position: { x: 0, y: 0 },
  buttonElement: null,
})

const router = useRouter()

// Add defaults data
const defaults = ref(null)

// Add function to fetch defaults
const fetchDefaults = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT freight_small, freight_big, rate
        FROM defaults 
        ORDER BY id ASC
        LIMIT 1
      `,
    })
    if (result.success && result.data.length > 0) {
      defaults.value = result.data[0]
    }
  } catch (error) {
    // Error fetching defaults
  }
}

// Add freight value calculation function
const getFreightValue = (car) => {
  // Convert freight to number for proper comparison
  const freight = parseFloat(car.freight) || 0

  // If freight is greater than 0, return it
  if (freight > 0) {
    return freight
  }

  // Use default freight based on car size
  if (car.is_big_car && defaults.value?.freight_big) {
    const defaultFreight = parseFloat(defaults.value.freight_big) || 0

    return defaultFreight
  }

  if (!car.is_big_car && defaults.value?.freight_small) {
    const defaultFreight = parseFloat(defaults.value.freight_small) || 0

    return defaultFreight
  }

  return 0
}

// Fetch file counts and missing required status for all cars (efficient single query)
const fetchCarFilesCounts = async (carIds) => {
  if (!carIds || carIds.length === 0) return
  
  try {
    // First, get all required category IDs
    const requiredCategoriesResult = await callApi({
      query: `SELECT id FROM car_file_categories WHERE is_required = 1`,
      requiresAuth: true,
    })
    
    const requiredCategoryIds = requiredCategoriesResult.success && requiredCategoriesResult.data
      ? requiredCategoriesResult.data.map((cat) => cat.id)
      : []
    
    if (requiredCategoryIds.length === 0) {
      // No required categories, just count files
      const result = await callApi({
        query: `
          SELECT 
            car_id,
            COUNT(*) as file_count
          FROM car_files
          WHERE car_id IN (${carIds.map(() => '?').join(',')}) AND is_active = 1
          GROUP BY car_id
        `,
        params: carIds,
        requiresAuth: true,
      })
      
      if (result.success && result.data) {
        result.data.forEach((row) => {
          carFilesCountMap.value[row.car_id] = {
            count: parseInt(row.file_count) || 0,
            hasMissingRequired: false,
          }
        })
        // Set 0 for cars with no files
        carIds.forEach((carId) => {
          if (!carFilesCountMap.value[carId]) {
            carFilesCountMap.value[carId] = { count: 0, hasMissingRequired: false }
          }
        })
      }
      return
    }
    
    // Get file counts and existing category IDs per car
    const result = await callApi({
      query: `
        SELECT 
          cf.car_id,
          COUNT(DISTINCT cf.id) as file_count,
          GROUP_CONCAT(DISTINCT cf.category_id) as existing_category_ids
        FROM car_files cf
        WHERE cf.car_id IN (${carIds.map(() => '?').join(',')}) AND cf.is_active = 1
        GROUP BY cf.car_id
      `,
      params: carIds,
      requiresAuth: true,
    })
    
    // Process results
    const countsByCar = {}
    if (result.success && result.data) {
      result.data.forEach((row) => {
        countsByCar[row.car_id] = {
          count: parseInt(row.file_count) || 0,
          existingCategoryIds: row.existing_category_ids 
            ? row.existing_category_ids.split(',').map(Number).filter(Boolean)
            : [],
        }
      })
    }
    
    // Check for missing required files for each car
    carIds.forEach((carId) => {
      const carData = countsByCar[carId] || { count: 0, existingCategoryIds: [] }
      const hasAllRequired = requiredCategoryIds.every((reqId) =>
        carData.existingCategoryIds.includes(reqId)
      )
      
      carFilesCountMap.value[carId] = {
        count: carData.count,
        hasMissingRequired: !hasAllRequired,
      }
    })
  } catch (err) {
    console.error('Error fetching car file counts:', err)
    // Set defaults on error
    carIds.forEach((carId) => {
      carFilesCountMap.value[carId] = { count: 0, hasMissingRequired: false }
    })
  }
}

// Get document count for a car
const getDocumentCount = (car) => {
  const fileData = carFilesCountMap.value[car.id]
  return fileData ? fileData.count : 0
}

// Check if car has missing required documents
const hasMissingRequiredDocuments = (car) => {
  const fileData = carFilesCountMap.value[car.id]
  return fileData ? fileData.hasMissingRequired : false
}

// Handle document link clicks to show better error messages
const handleDocumentClick = async (event, path, documentName) => {
  // Check if the path is valid
  if (!path || path.trim() === '') {
    event.preventDefault()
    alert(t('carStock.document_path_invalid') || `Invalid document path for ${documentName}`)
    return
  }
  
  // Log the path for debugging
  console.log(`Opening document: ${documentName}`, {
    originalPath: path,
    generatedUrl: getFileUrl(path)
  })
  
  // Try to fetch the file first to check if it exists (only in development)
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    try {
      const fileUrl = getFileUrl(path)
      const response = await fetch(fileUrl, { method: 'HEAD' })
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: 'Unknown error' }))
        console.error('Document not found:', {
          documentName,
          path,
          fileUrl,
          error: errorData
        })
        // Don't prevent default - let user see the error page
      }
    } catch (err) {
      // If fetch fails (CORS or network error), let the link open normally
      console.warn('Could not verify document existence:', err)
    }
  }
}

// Add CFR DZA value calculation function
const getCfrDzaValue = (car) => {
  // Use cfr_da field if available, otherwise calculate from price_cell
  if (car.cfr_da) {
    return parseFloat(car.cfr_da).toFixed(2)
  }

  // Fallback calculation for older records
  const fob = parseFloat(car.price_cell) || 0
  const freight = getFreightValue(car)
  const rate = getRateValue(car)

  const cfrUsd = fob + freight
  const cfrDza = cfrUsd * rate

  return cfrDza.toFixed(2)
}

// Add CFR USD value calculation function
const getCfrUsdValue = (car) => {
  const fob = parseFloat(car.price_cell) || 0
  const freight = getFreightValue(car)

  const cfrUsd = fob + freight
  return cfrUsd.toFixed(2)
}

// Add rate value calculation function
const getRateValue = (car) => {
  // If car has a rate and it's greater than 0, use it
  const carRate = parseFloat(car.rate) || 0
  if (carRate > 0) {
    return carRate
  }

  // Otherwise use default rate from defaults table
  if (defaults.value?.rate) {
    return parseFloat(defaults.value.rate) || 0
  }

  return 0
}

// Add hidden time formatting function
const formatHiddenTime = (timestamp) => {
  if (!timestamp) return ''

  try {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now - date
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    const diffMinutes = Math.floor(diffMs / (1000 * 60))

    if (diffDays > 0) {
      return `${diffDays}d ago`
    } else if (diffHours > 0) {
      return `${diffHours}h ago`
    } else if (diffMinutes > 0) {
      return `${diffMinutes}m ago`
    } else {
      return 'Just now'
    }
  } catch (error) {
    return timestamp
  }
}

// Function to calculate text color based on background color
const getTextColor = (backgroundColor) => {
  if (!backgroundColor) return '#374151'

  // Convert hex to RGB
  const hex = backgroundColor.replace('#', '')
  const r = parseInt(hex.substr(0, 2), 16)
  const g = parseInt(hex.substr(2, 2), 16)
  const b = parseInt(hex.substr(4, 2), 16)

  // Calculate luminance
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255

  // Return white text for dark backgrounds, black text for light backgrounds
  return luminance > 0.5 ? '#000000' : '#ffffff'
}

const allCars = ref([])

const fetchCarsStock = async () => {
  if (allCars.value === null) {
    loading.value = false
    return
  }

  loading.value = true
  error.value = null

  try {
    // Apply filters in-memory to allCars.value
    let filteredCars = [...allCars.value]

    // Apply buy bill filter if provided
    if (props.buyBillId) {
      filteredCars = filteredCars.filter((car) => car.buy_bill_id == props.buyBillId)
    }

    // Apply client filter if provided
    if (props.clientId) {
      filteredCars = filteredCars.filter((car) => car.id_client == props.clientId)
    }

    // Apply filters if they exist
    if (props.filters) {
      // Basic filter (search across multiple fields)
      if (props.filters.basic && props.filters.basic.length > 0) {
        const searchTerms = Array.isArray(props.filters.basic)
          ? props.filters.basic
          : [props.filters.basic]
        const operator = props.filters.basicOperator || 'AND'
        if (searchTerms.length === 1) {
          const term = searchTerms[0].trim().toLowerCase()
          filteredCars = filteredCars.filter((car) =>
            [
              car.id,
              car.car_name,
              car.color,
              car.vin,
              car.loading_port,
              car.discharge_port,
              car.client_name,
              car.client_id_no,
              car.warehouse_name,
              car.buy_bill_ref,
              car.sell_bill_ref,
              car.export_lisence_ref,
            ]
              .map((v) => (v ? v.toString().toLowerCase() : ''))
              .some((v) => v.includes(term)),
          )
        } else {
          filteredCars = filteredCars.filter((car) => {
            const values = [
              car.id,
              car.car_name,
              car.color,
              car.vin,
              car.loading_port,
              car.discharge_port,
              car.client_name,
              car.client_id_no,
              car.warehouse_name,
              car.buy_bill_ref,
              car.sell_bill_ref,
              car.export_lisence_ref,
            ].map((v) => (v ? v.toString().toLowerCase() : ''))
            if (operator === 'AND') {
              return searchTerms.every((term) =>
                values.some((v) => v.includes(term.trim().toLowerCase())),
              )
            } else {
              return searchTerms.some((term) =>
                values.some((v) => v.includes(term.trim().toLowerCase())),
              )
            }
          })
        }
      }
      // Advanced filters
      if (props.filters.advanced) {
        const adv = props.filters.advanced
        if (adv.id && adv.id.trim() !== '') {
          filteredCars = filteredCars.filter((car) => car.id == adv.id.trim())
        }
        if (adv.car_name && adv.car_name.trim() !== '') {
          filteredCars = filteredCars.filter((car) => car.car_name == adv.car_name.trim())
        }
        if (adv.color && adv.color.trim() !== '') {
          filteredCars = filteredCars.filter((car) => car.color == adv.color.trim())
        }
        if (adv.vin && adv.vin.trim() !== '') {
          filteredCars = filteredCars.filter((car) => car.vin && car.vin.includes(adv.vin.trim()))
        }
        if (adv.loading_port && adv.loading_port.trim() !== '') {
          filteredCars = filteredCars.filter((car) => car.loading_port == adv.loading_port.trim())
        }
        if (adv.discharge_port && adv.discharge_port.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.discharge_port == adv.discharge_port.trim(),
          )
        }
        if (adv.freight_min && adv.freight_min.toString().trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => parseFloat(car.freight) >= parseFloat(adv.freight_min),
          )
        }
        if (adv.freight_max && adv.freight_max.toString().trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => parseFloat(car.freight) <= parseFloat(adv.freight_max),
          )
        }
        if (adv.price_min && adv.price_min.toString().trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => parseFloat(car.price_cell) >= parseFloat(adv.price_min),
          )
        }
        if (adv.price_max && adv.price_max.toString().trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => parseFloat(car.price_cell) <= parseFloat(adv.price_max),
          )
        }
        if (adv.loading_date_from && adv.loading_date_from.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.date_loding && car.date_loding >= adv.loading_date_from.trim(),
          )
        }
        if (adv.loading_date_to && adv.loading_date_to.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.date_loding && car.date_loding <= adv.loading_date_to.trim(),
          )
        }
        if (adv.buy_date_from && adv.buy_date_from.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.date_buy && car.date_buy >= adv.buy_date_from.trim(),
          )
        }
        if (adv.buy_date_to && adv.buy_date_to.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.date_buy && car.date_buy <= adv.buy_date_to.trim(),
          )
        }
        if (adv.sold_date_from && adv.sold_date_from.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.sell_bill_date && car.sell_bill_date >= adv.sold_date_from.trim(),
          )
        }
        if (adv.sold_date_to && adv.sold_date_to.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.sell_bill_date && car.sell_bill_date <= adv.sold_date_to.trim(),
          )
        }
        if (adv.status && adv.status.trim() !== '') {
          if (adv.status === 'available') {
            filteredCars = filteredCars.filter((car) => !car.id_sell)
          } else if (adv.status === 'sold') {
            filteredCars = filteredCars.filter((car) => car.id_sell)
          }
        }
        if (adv.client && adv.client.trim() !== '') {
          filteredCars = filteredCars.filter((car) => car.client_name == adv.client.trim())
        }
        if (adv.client_id_no && adv.client_id_no.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.client_id_no && car.client_id_no.includes(adv.client_id_no.trim()),
          )
        }
        if (adv.container_status && adv.container_status.trim() !== '') {
          if (adv.container_status === 'has_container') {
            filteredCars = filteredCars.filter(
              (car) => car.container_ref && car.container_ref.trim() !== '',
            )
          } else if (adv.container_status === 'has_not_container') {
            filteredCars = filteredCars.filter(
              (car) => !car.container_ref || car.container_ref.trim() === '',
            )
          }
        }
        if (adv.warehouse && adv.warehouse.trim() !== '') {
          filteredCars = filteredCars.filter((car) => car.warehouse_name == adv.warehouse.trim())
        }
        if (adv.container_ref && adv.container_ref.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.container_ref && car.container_ref.includes(adv.container_ref.trim()),
          )
        }
        if (adv.export_lisence_ref && adv.export_lisence_ref.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) =>
              car.export_lisence_ref &&
              car.export_lisence_ref.includes(adv.export_lisence_ref.trim()),
          )
        }
        if (adv.loading_status && adv.loading_status.trim() !== '') {
          if (adv.loading_status === 'loaded') {
            filteredCars = filteredCars.filter((car) => car.date_loding)
          } else if (adv.loading_status === 'not_loaded') {
            filteredCars = filteredCars.filter((car) => !car.date_loding && car.is_batch == 0)
          }
        }
        if (adv.documents_status && adv.documents_status.trim() !== '') {
          if (adv.documents_status === 'received') {
            filteredCars = filteredCars.filter((car) => car.date_get_documents_from_supp)
          } else if (adv.documents_status === 'not_received') {
            filteredCars = filteredCars.filter((car) => !car.date_get_documents_from_supp)
          }
        }
        if (adv.bl_status && adv.bl_status.trim() !== '') {
          if (adv.bl_status === 'received') {
            filteredCars = filteredCars.filter((car) => car.date_get_bl)
          } else if (adv.bl_status === 'not_received') {
            filteredCars = filteredCars.filter((car) => !car.date_get_bl)
          }
        }
        if (adv.warehouse_status && adv.warehouse_status.trim() !== '') {
          if (adv.warehouse_status === 'in_warehouse') {
            filteredCars = filteredCars.filter((car) => car.in_wharhouse_date)
          } else if (adv.warehouse_status === 'not_in_warehouse') {
            filteredCars = filteredCars.filter((car) => !car.in_wharhouse_date)
          }
        }
        if (adv.bill_ref && adv.bill_ref.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.buy_bill_ref && car.buy_bill_ref.includes(adv.bill_ref.trim()),
          )
        }
        if (adv.sell_bill_ref && adv.sell_bill_ref.trim() !== '') {
          filteredCars = filteredCars.filter(
            (car) => car.sell_bill_ref && car.sell_bill_ref.includes(adv.sell_bill_ref.trim()),
          )
        }
        if (adv.tmp_client_status && adv.tmp_client_status.trim() !== '') {
          if (adv.tmp_client_status === 'tmp') {
            filteredCars = filteredCars.filter((car) => car.is_tmp_client == 1)
          } else if (adv.tmp_client_status === 'permanent') {
            filteredCars = filteredCars.filter((car) => car.is_tmp_client == 0)
          }
        }
        if (adv.hidden && adv.hidden.trim() !== '') {
          if (adv.hidden === '1') {
            filteredCars = filteredCars.filter((car) => car.hidden == 1)
          } else if (adv.hidden === '0') {
            filteredCars = filteredCars.filter((car) => car.hidden == 0)
          }
        }
        if (adv.whole_sale_status && adv.whole_sale_status.trim() !== '') {
          if (adv.whole_sale_status === 'whole_sale') {
            filteredCars = filteredCars.filter((car) => car.is_batch == 1)
          } else if (adv.whole_sale_status === 'not_whole_sale') {
            filteredCars = filteredCars.filter((car) => car.is_batch == 0)
          }
        }
        if (adv.has_bl) {
          filteredCars = filteredCars.filter((car) => car.date_get_bl)
        }
        if (adv.freight_paid) {
          filteredCars = filteredCars.filter((car) => car.date_pay_freight)
        }
        if (adv.has_supplier_docs) {
          filteredCars = filteredCars.filter((car) => car.date_get_documents_from_supp)
        }
        if (adv.in_warehouse) {
          filteredCars = filteredCars.filter((car) => car.in_wharhouse_date)
        }
        if (adv.has_export_license) {
          filteredCars = filteredCars.filter(
            (car) => car.export_lisence_ref && car.export_lisence_ref !== '',
          )
        }
        if (adv.is_loaded) {
          filteredCars = filteredCars.filter((car) => car.date_loding)
        }
        if (adv.has_vin) {
          filteredCars = filteredCars.filter((car) => car.vin && car.vin !== '')
        }

        // File upload status filtering
        if (adv.bl_file_status && adv.bl_file_status.trim() !== '') {
          if (adv.bl_file_status === 'uploaded') {
            filteredCars = filteredCars.filter(
              (car) => car.path_documents && car.path_documents !== '',
            )
          } else if (adv.bl_file_status === 'not_uploaded') {
            filteredCars = filteredCars.filter(
              (car) => !car.path_documents || car.path_documents === '',
            )
          }
        }

        if (adv.sell_pi_file_status && adv.sell_pi_file_status.trim() !== '') {
          if (adv.sell_pi_file_status === 'uploaded') {
            filteredCars = filteredCars.filter((car) => car.sell_pi_path && car.sell_pi_path !== '')
          } else if (adv.sell_pi_file_status === 'not_uploaded') {
            filteredCars = filteredCars.filter(
              (car) => !car.sell_pi_path || car.sell_pi_path === '',
            )
          }
        }

        if (adv.buy_pi_file_status && adv.buy_pi_file_status.trim() !== '') {
          if (adv.buy_pi_file_status === 'uploaded') {
            filteredCars = filteredCars.filter((car) => car.buy_pi_path && car.buy_pi_path !== '')
          } else if (adv.buy_pi_file_status === 'not_uploaded') {
            filteredCars = filteredCars.filter((car) => !car.buy_pi_path || car.buy_pi_path === '')
          }
        }

        if (adv.coo_file_status && adv.coo_file_status.trim() !== '') {
          if (adv.coo_file_status === 'uploaded') {
            filteredCars = filteredCars.filter((car) => car.path_coo && car.path_coo !== '')
          } else if (adv.coo_file_status === 'not_uploaded') {
            filteredCars = filteredCars.filter((car) => !car.path_coo || car.path_coo === '')
          }
        }

        if (adv.coc_file_status && adv.coc_file_status.trim() !== '') {
          if (adv.coc_file_status === 'uploaded') {
            filteredCars = filteredCars.filter((car) => car.path_coc && car.path_coc !== '')
          } else if (adv.coc_file_status === 'not_uploaded') {
            filteredCars = filteredCars.filter((car) => !car.path_coc || car.path_coc === '')
          }
        }

        // Alert-specific filtering
        if (props.alertType && props.alertDays) {
          const alertDays = props.alertDays
          const alertType = props.alertType
          const currentDate = new Date()

          if (alertType === 'unloaded') {
            // Cars sold more than X days ago that haven't been loaded
            filteredCars = filteredCars.filter((car) => {
              // Must have a sell date (is sold)
              if (!car.date_sell) {
                return false
              }
              // Must not be loaded
              if (car.date_loding) {
                return false
              }
              // Must not have documents sent
              if (car.date_send_documents) {
                return false
              }
              // Must not be hidden
              if (car.hidden == 1) {
                return false
              }
              // Must not be batch
              if (car.is_batch == 1) {
                return false
              }
              // Must not be batch sell
              if (car.is_batch_sell == 1) {
                return false
              }
              // Must not have container ref
              if (car.container_ref && car.container_ref !== '') {
                return false
              }

              const sellDate = new Date(car.date_sell)
              const daysDiff = (currentDate - sellDate) / (1000 * 60 * 60 * 24)
              if (daysDiff <= alertDays) {
                return false
              }

              return true
            })
          } else if (alertType === 'not_arrived') {
            // Cars purchased more than X days ago that haven't arrived at port
            filteredCars = filteredCars.filter((car) => {
              // Must not be loaded (if loaded, it means arrived and has license)
              if (car.date_loding) return false
              // Must not be in warehouse
              if (car.in_wharhouse_date) return false
              // Must not have documents sent
              if (car.date_send_documents) return false
              // Must not be hidden
              if (car.hidden == 1) return false
              // Must not be batch
              if (car.is_batch == 1) return false
              // Must not have container ref
              if (car.container_ref && car.container_ref !== '') return false

              const buyDate = new Date(car.date_buy)
              const daysDiff = (currentDate - buyDate) / (1000 * 60 * 60 * 24)
              return daysDiff > alertDays
            })
          } else if (alertType === 'no_licence') {
            // Cars purchased more than X days ago without export license
            filteredCars = filteredCars.filter((car) => {
              // Must not be loaded (if loaded, it means arrived and has license)
              if (car.date_loding) return false
              // Must not have export license
              if (car.export_lisence_ref && car.export_lisence_ref !== '') return false
              // Must not have documents sent
              if (car.date_send_documents) return false
              // Must not be hidden
              if (car.hidden == 1) return false
              // Must not be batch
              if (car.is_batch == 1) return false
              // Must not have container ref
              if (car.container_ref && car.container_ref !== '') return false

              const buyDate = new Date(car.date_buy)
              const daysDiff = (currentDate - buyDate) / (1000 * 60 * 60 * 24)
              return daysDiff > alertDays
            })
          } else if (alertType === 'no_docs_sent') {
            // Cars sold more than X days ago without documents sent
            filteredCars = filteredCars.filter((car) => {
              // Must have a sell bill (INNER JOIN equivalent)
              if (!car.date_sell) {
                return false
              }
              // Must not have documents sent
              if (car.date_send_documents) {
                return false
              }
              // Must not be hidden
              if (car.hidden == 1) {
                return false
              }
              // Must not be batch
              if (car.is_batch == 1) {
                return false
              }
              // Must not be batch sell
              if (car.is_batch_sell == 1) {
                return false
              }

              const sellDate = new Date(car.date_sell)
              const daysDiff = (currentDate - sellDate) / (1000 * 60 * 60 * 24)
              if (daysDiff <= alertDays) {
                return false
              }

              return true
            })
          }
        }
      }
    }

    cars.value = filteredCars
    
    // Fetch file counts for visible cars
    const carIds = filteredCars.map((car) => car.id)
    if (carIds.length > 0) {
      fetchCarFilesCounts(carIds)
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Add watcher for clientId after fetchCarsStock is defined
watch(
  () => props.clientId,
  (newClientId) => {
    // Only apply filtering if we have data loaded
    if (allCars.value && allCars.value.length > 0) {
      fetchCarsStock()
    }
  },
)

// Add watchers for alertType and alertDays
watch(
  () => [props.alertType, props.alertDays],
  () => {
    // Only apply filtering if we have data loaded
    if (allCars.value && allCars.value.length > 0) {
      fetchCarsStock()
    }
  },
)

const handleEdit = async (car) => {
  if (isProcessing.value.edit) return
  isProcessing.value.edit = true
  try {
    props.onEdit(car)
  } finally {
    isProcessing.value.edit = false
  }
}

const editCfrDa = async (carId, currentValue) => {
  const newValue = prompt(`Enter new CFR DA value for car #${carId}:`, currentValue || '')

  if (newValue === null) return // User cancelled

  const numericValue = parseFloat(newValue)
  if (isNaN(numericValue) || numericValue < 0) {
    alert('Please enter a valid positive number')
    return
  }

  try {
    const result = await callApi({
      query: 'UPDATE cars_stock SET cfr_da = ? WHERE id = ?',
      params: [numericValue, carId],
    })

    if (result.success) {
      // Update the car data in memory
      const car = allCars.value.find((c) => c.id === carId)
      if (car) {
        car.cfr_da = numericValue
      }
      // Refresh the display
      fetchCarsStock()
    } else {
      alert('Failed to update CFR DA value')
    }
  } catch (error) {
    alert('Error updating CFR DA value: ' + error.message)
  }
}

const toggleDropdown = (carId, event) => {
  // Keep existing dropdown logic
  if (isDropdownOpen.value[carId]) {
    isDropdownOpen.value[carId] = false
  } else {
    // Close all other dropdowns
    Object.keys(isDropdownOpen.value).forEach((id) => {
      isDropdownOpen.value[id] = false
    })
    isDropdownOpen.value[carId] = true
  }
}

// Add teleport dropdown functions
const openTeleportDropdown = (carId, event) => {
  const button = event.currentTarget
  const rect = button.getBoundingClientRect()

  teleportDropdown.value = {
    isOpen: true,
    carId: carId,
    position: {
      x: rect.left,
      y: rect.bottom + window.scrollY,
    },
    buttonElement: button,
  }

  // Close the regular dropdown
  isDropdownOpen.value[carId] = false
}

const closeTeleportDropdown = () => {
  teleportDropdown.value.isOpen = false
  teleportDropdown.value.carId = null
}

// Add click outside handler
const handleClickOutside = (event) => {
  // Handle teleport dropdown
  if (teleportDropdown.value.isOpen) {
    const dropdown = document.querySelector('.teleport-dropdown')
    const button = teleportDropdown.value.buttonElement

    if (dropdown && !dropdown.contains(event.target) && button && !button.contains(event.target)) {
      closeTeleportDropdown()
    }
  }

  // Handle regular dropdowns
  const dropdownToggles = document.querySelectorAll('.dropdown-toggle')
  let clickedOnToggle = false

  dropdownToggles.forEach((toggle) => {
    if (toggle.contains(event.target)) {
      clickedOnToggle = true
    }
  })

  if (!clickedOnToggle) {
    // Close all regular dropdowns if clicked outside
    isDropdownOpen.value = {}
  }
}

// Add scroll handler to close dropdown
const handleScroll = () => {
  if (teleportDropdown.value.isOpen) {
    closeTeleportDropdown()
  }
}

// Helper function to get car by ID
const getCarById = (carId) => {
  return cars.value.find((car) => car.id === carId)
}

// Add new methods for VIN editing
const handleVINAction = (car) => {
  closeTeleportDropdown()
  selectedCar.value = car
  showVinEditForm.value = true
}

const handleVinSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for file upload
const handleFilesAction = (car) => {
  closeTeleportDropdown()
  selectedCar.value = car
  showFilesUploadForm.value = true
}

const handleFilesSave = async (updatedCar) => {
  // Refresh file count for this car
  if (updatedCar && updatedCar.id) {
    await fetchCarFilesCounts([updatedCar.id])
  }
  
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for ports editing
const handlePortsAction = (car) => {
  closeTeleportDropdown()
  selectedCar.value = car
  showPortsEditForm.value = true
}

const handlePortsSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for money editing
const handleMoneyAction = (car) => {
  closeTeleportDropdown()
  selectedCar.value = car
  showMoneyEditForm.value = true
}

const handleMoneySave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new method for warehouse action
const handleWarehouseAction = (car) => {
  closeTeleportDropdown()
  selectedCar.value = car
  showWarehouseForm.value = true
}

const handleWarehouseSave = (updatedCarOrCars) => {
  // Handle both single car and bulk cars updates
  if (Array.isArray(updatedCarOrCars)) {
    // Bulk update - update multiple cars
    updatedCarOrCars.forEach((updatedCar) => {
      const index = cars.value.findIndex((c) => c.id === updatedCar.id)
      if (index !== -1) {
        cars.value[index] = { ...cars.value[index], ...updatedCar }
      }
    })
  } else {
    // Single car update
    const index = cars.value.findIndex((c) => c.id === updatedCarOrCars.id)
    if (index !== -1) {
      cars.value[index] = { ...cars.value[index], ...updatedCarOrCars }
    }
  }
}

const handleWarehouseChanged = (updatedCarOrCars) => {
  // Emit the warehouse-changed event to parent components
  emit('warehouse-changed', updatedCarOrCars)
}

// Add new method for documents action
const handleDocumentsClick = (car) => {
  closeTeleportDropdown()
  selectedCarForDocuments.value = car
  showDocumentsForm.value = true
}

const handleDocumentsClose = () => {
  showDocumentsForm.value = false
  selectedCarForDocuments.value = null
}

const handleDocumentsSave = (updatedCar) => {
  // Update the car in the table data
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for load action
const handleLoadClick = (car) => {
  closeTeleportDropdown()
  selectedCarForLoad.value = car
  showLoadForm.value = true
}

const handleLoadClose = () => {
  showLoadForm.value = false
  selectedCarForLoad.value = null
}

const handleLoadSave = (updatedCar) => {
  // Update the car in the table data
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for task form
const handleTaskClose = () => {
  showTaskForm.value = false
  selectedCarForTask.value = null
}

const handleTaskSave = (result) => {
  showTaskForm.value = false
  selectedCarForTask.value = null
  alert(t('carStock.task_created_successfully'))
  fetchCarsStock()
}

// Add new methods for switch buy bill form
const handleSwitchBuyBillClose = () => {
  showSwitchBuyBillForm.value = false
  selectedCarForSwitchBuyBill.value = null
}

const handleSwitchBuyBillSave = async (result) => {
  showSwitchBuyBillForm.value = false
  selectedCarForSwitchBuyBill.value = null

  try {
    const switchResult = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_buy_details = CASE 
          WHEN id = ? THEN (SELECT id_buy_details FROM cars_stock WHERE id = ?)
          WHEN id = ? THEN (SELECT id_buy_details FROM cars_stock WHERE id = ?)
        END
        WHERE id IN (?, ?)
      `,
      params: [
        result.carId1,
        result.carId2,
        result.carId2,
        result.carId1,
        result.carId1,
        result.carId2,
      ],
    })

    if (switchResult.success) {
      alert(t('carStock.purchase_bills_switched_successfully'))
      fetchCarsStock()
    } else {
      alert(t('carStock.failed_to_switch_purchase_bills'))
    }
  } catch (err) {
    alert(t('carStock.error_switching_purchase_bills'))
  }
}

// Add this method after other methods
const closeAllDropdowns = () => {
  isDropdownOpen.value = {}
}

const handleAdvancedFilters = (filters) => {
  advancedFilters.value = filters
  showAdvancedFilter.value = false
  // Trigger the watcher by updating the filters prop
  props.filters.advanced = filters
}

const openSellBillTab = (car) => {
  closeTeleportDropdown()
  if (car.id_sell) {
    // Open in new tab using router.resolve
    const route = router.resolve(`/sell-bills/${car.id_sell}`)
    window.open(route.href, '_blank')
  }
}

const handleSwitchPurchaseBill = (car) => {
  closeTeleportDropdown()
  selectedCarForSwitchBuyBill.value = car
  showSwitchBuyBillForm.value = true
}

const handleTaskAction = (car) => {
  closeTeleportDropdown()
  selectedCarForTask.value = car
  showTaskForm.value = true
}

const showNotesModal = ref(false)
const notesEditCar = ref(null)
const notesEditValue = ref('')

const handleNotesAction = (car) => {
  closeTeleportDropdown()
  notesEditCar.value = car
  notesEditValue.value = car.notes || ''
  showNotesModal.value = true
}

const handleNotesSave = async (result) => {
  try {
    const response = await callApi({
      query: 'UPDATE cars_stock SET notes = ? WHERE id = ?',
      params: [result.notes, result.carId],
    })

    if (response.success) {
      alert(t('carStock.notes_updated_successfully'))
      fetchCarsStock()
    } else {
      alert(result.error || t('carStock.failed_to_update_notes'))
    }
  } catch (err) {
    alert(err.message || t('carStock.error_updating_notes'))
  }
}

const cancelNotes = () => {
  showNotesModal.value = false
}

const handleVinsAssigned = (assignments) => {
  showVinAssignmentModal.value = false
  fetchCarsStock()
}

const handleVinAssignmentModalClose = () => {
  showVinAssignmentModal.value = false
}

const handlePortsFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_ports_editing'))
    return
  }

  const updatedCars = sortedCars.value.filter((car) => selectedCars.value.has(car.id))
  showPortsBulkEditForm.value = true
}

const handlePortsBulkSave = (updatedCars) => {
  showPortsBulkEditForm.value = false
  fetchCarsStock()
}

const handleWarehouseFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_warehouse_editing'))
    return
  }

  const selectedCarsForWarehouse = sortedCars.value.filter((car) => selectedCars.value.has(car.id))
  selectedCar.value = selectedCarsForWarehouse[0]
  showWarehouseForm.value = true
}

const handleNotesFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_notes_editing'))
    return
  }

  const updatedCars = sortedCars.value.filter((car) => selectedCars.value.has(car.id))
  showNotesBulkEditForm.value = true
}

const handleTaskFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_task_creation'))
    return
  }

  const selectedCarsForTask = sortedCars.value.filter((car) => selectedCars.value.has(car.id))
  selectedCarForTask.value = selectedCarsForTask[0]
  showTaskForm.value = true
}

const handleNotesBulkSave = (updatedCars) => {
  showNotesBulkEditForm.value = false
  fetchCarsStock()
}

const handleColorFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_color_editing'))
    return
  }

  const updatedCars = sortedCars.value.filter((car) => selectedCars.value.has(car.id))
  showColorBulkEditForm.value = true
}

const handleColorBulkSave = (updatedCars) => {
  showColorBulkEditForm.value = false
  fetchCarsStock()
}

// Add new refs for color bulk edit form
const showColorBulkEditForm = ref(false)

// Add new refs for selections
const showSaveSelectionForm = ref(false)
const showSendSelectionForm = ref(false)
const showSelectionsModal = ref(false)

// Add computed property for color permission
const can_change_car_color = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_change_car_color')
})

// Add computed property for hide car permission
const can_hide_car = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_hide_car')
})

const showExportLicenseBulkEditForm = ref(false)

const handleExportLicenseFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_export_license_editing'))
    return
  }

  const updatedCars = sortedCars.value.filter((car) => selectedCars.value.has(car.id))
  showExportLicenseBulkEditForm.value = true
}

const handleCfrDaFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_cfr_da_editing'))
    return
  }

  const newValue = prompt(`Enter new CFR DA value for ${selectedCars.value.size} selected cars:`)

  if (newValue === null) return // User cancelled

  const numericValue = parseFloat(newValue)
  if (isNaN(numericValue) || numericValue < 0) {
    alert('Please enter a valid positive number')
    return
  }

  // Update all selected cars
  const selectedCarIds = Array.from(selectedCars.value)
  updateCarsCfrDa(selectedCarIds, numericValue)
}

const updateCarsCfrDa = async (selectedCarIds, numericValue) => {
  try {
    // Update all selected cars in database
    const result = await callApi({
      query:
        'UPDATE cars_stock SET cfr_da = ? WHERE id IN (' +
        selectedCarIds.map(() => '?').join(',') +
        ')',
      params: [numericValue, ...selectedCarIds],
    })

    if (result.success) {
      // Update the cars data in memory
      selectedCarIds.forEach((carId) => {
        const car = allCars.value.find((c) => c.id === carId)
        if (car) {
          car.cfr_da = numericValue
        }
      })

      // Refresh the display
      fetchCarsStock()

      // Clear selection
      selectedCars.value.clear()

      alert(`Successfully updated CFR DA value for ${selectedCarIds.length} cars`)
    } else {
      alert('Failed to update CFR DA values')
    }
  } catch (error) {
    alert('Error updating CFR DA values: ' + error.message)
  }
}

const handleExportLicenseBulkSave = (updatedCars) => {
  // Update the cars in the local array
  updatedCars.forEach((updatedCar) => {
    const index = cars.value.findIndex((c) => c.id === updatedCar.id)
    if (index !== -1) {
      cars.value[index] = { ...cars.value[index], ...updatedCar }
    }
  })
  showExportLicenseBulkEditForm.value = false
}

const loadInitialCarsData = async () => {
  loading.value = true
  error.value = null
  try {
    // Build the WHERE clause conditionally based on admin status
    let whereClause = '1=1' // Default: show all cars
    const currentUserId = user.value?.id
    const isUserAdmin = user.value?.role_id === 1

    if (!isUserAdmin && currentUserId) {
      // Non-admin users can only see:
      // 1. Non-hidden cars (hidden = 0)
      // 2. Cars they hid themselves (hidden = 1 AND hidden_by_user_id = current_user_id)
      whereClause = `(cs.hidden = 0 OR (cs.hidden = 1 AND cs.hidden_by_user_id = ${currentUserId}))`
    } else if (!isUserAdmin && !currentUserId) {
      // If user is not admin and no user ID, show only non-hidden cars
      whereClause = 'cs.hidden = 0'
    }

    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.price_cell,
          cs.cfr_da,
          cs.date_loding,
          sb.date_sell,
          sb.date_sell as sell_bill_date,
          cs.notes,
          cs.freight,
          cs.path_documents,
          cs.sell_pi_path,
          cs.buy_pi_path,
          cs.path_coo,
          cs.path_coc,
          cs.id_client,
          cs.id_port_loading,
          cs.id_port_discharge,
          cs.id_buy_details,
          cs.date_send_documents,
          cs.id_sell_pi,
          cs.id_sell,
          cs.export_lisence_ref,
          cs.id_warehouse,
          cs.in_wharhouse_date,
          cs.date_get_documents_from_supp,
          cs.date_get_keys_from_supp,
          cs.rate,
          cs.date_get_bl,
          cs.date_pay_freight,
          cs.is_batch,
          cs.hidden,
          cs.hidden_by_user_id,
          cs.hidden_time_stamp,
          hu.username as hidden_by_username,
          c.name as client_name,
          cn.car_name,
          clr.color,
          clr.hexa,
          lp.loading_port,
          dp.discharge_port,
          bd.price_sell as buy_price,
          bb.id as buy_bill_id,
          bb.date_buy,
          w.warhouse_name as warehouse_name,
          bb.bill_ref as buy_bill_ref,
          sb.bill_ref as sell_bill_ref,
          sb.is_batch_sell,
          cs.is_used_car,
          cs.is_big_car,
          c.id_no as client_id_no,
          c.id_copy_path as client_id_picture,
          cs.container_ref,
          cs.id_color as car_id_color,
          CASE 
            WHEN cs.id_sell IS NOT NULL THEN 'Sold'
            ELSE 'Available'
          END as status
        FROM cars_stock cs
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
        LEFT JOIN sell_bill sb ON cs.id_sell = sb.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        LEFT JOIN warehouses w ON cs.id_warehouse = w.id
        LEFT JOIN users hu ON cs.hidden_by_user_id = hu.id
        WHERE ${whereClause}
      `,
    })
    if (result.success) {
      // Set default payment_confirmed to 0 if column doesn't exist yet
      const carsData = (result.data || []).map((car) => ({
        ...car,
        payment_confirmed: car.payment_confirmed !== undefined ? car.payment_confirmed : 0,
      }))
      allCars.value = carsData
      cars.value = carsData
      
      // Fetch file counts for all cars
      const carIds = carsData.map((car) => car.id)
      if (carIds.length > 0) {
        fetchCarFilesCounts(carIds)
      }
      
      // Temporary debug - remove after fixing
      if (!result.data || result.data.length === 0) {
        console.warn(
          'CarStockTable: Query returned no data. WHERE clause:',
          whereClause,
          'User:',
          user.value,
        )
      }
    } else {
      error.value = result.error || 'Failed to fetch cars stock'
      allCars.value = []
      cars.value = []
      console.error('CarStockTable: Query failed:', result.error)
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
    allCars.value = []
    cars.value = []
  } finally {
    loading.value = false
  }
}

// Handler for payment confirmation updates
const handlePaymentConfirmedUpdate = (event) => {
  // Refresh cars list when payment confirmation is updated
  // This happens when a sell bill's payment is confirmed/unconfirmed
  fetchCarsStock()
}

onMounted(async () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchDefaults()
  await loadInitialCarsData()
  // Apply initial filtering after data is loaded
  fetchCarsStock()
  // Add event listeners for teleport dropdown
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('scroll', handleScroll, true)
  // Listen for payment confirmation updates from sell bills
  window.addEventListener('payment-confirmed-updated', handlePaymentConfirmedUpdate)
})

onUnmounted(() => {
  // Remove event listeners
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('scroll', handleScroll, true)
  window.removeEventListener('payment-confirmed-updated', handlePaymentConfirmedUpdate)
})

// Expose the fetchCarsStock method to parent components
defineExpose({
  fetchCarsStock,
  addCarsToMemory: async (newCarsData) => {
    // Generate IDs for new cars (since they don't have IDs yet)
    const maxId = allCars.value.length > 0 ? Math.max(...allCars.value.map((car) => car.id)) : 0

    const newCars = newCarsData.map((carData, index) => ({
      id: maxId + index + 1, // Generate unique ID
      vin: null,
      price_cell: carData.price_cell,
      date_loding: null,
      date_sell: null,
      notes: carData.notes,
      freight: null,
      path_documents: null,
      sell_pi_path: null,
      buy_pi_path: null,
      id_client: null,
      id_port_loading: null,
      id_port_discharge: null,
      id_buy_details: carData.id_buy_details,
      date_send_documents: null,
      id_sell_pi: null,
      id_sell: null,
      export_lisence_ref: null,
      id_warehouse: null,
      in_wharhouse_date: null,
      date_get_documents_from_supp: null,
      date_get_keys_from_supp: null,
      rate: null,
      date_get_bl: null,
      date_pay_freight: null,
      is_batch: 0,
      client_name: null,
      car_name: null,
      color: null,
      hexa: null,
      loading_port: null,
      discharge_port: null,
      buy_price: carData.price_cell,
      date_buy: null,
      warehouse_name: null,
      buy_bill_ref: null,
      sell_bill_ref: null,
      is_used_car: carData.is_used_car,
      is_big_car: carData.is_big_car,
      client_id_no: null,
      client_id_picture: null,
      container_ref: null,
      car_id_color: carData.id_color,
      buy_bill_id: carData.buy_bill_id,
      status: 'Available',
    }))

    // Add new cars to allCars
    allCars.value.push(...newCars)

    // Apply current filters to update the display
    fetchCarsStock()
  },
})

// In the handler for the refresh event (called by the toolbar):
const handleRefresh = async () => {
  await loadInitialCarsData()
  fetchCarsStock()
}

// Add after other toolbar handlers
const handleDeleteCars = async () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_deletion'))
    return
  }
  // Find cars that do not meet the deletion criteria
  const notDeletable = sortedCars.value.filter(
    (car) =>
      selectedCars.value.has(car.id) &&
      (car.vin || car.id_client || car.container_ref || car.id_sell_bill),
  )
  if (notDeletable.length > 0) {
    alert(
      t('carStock.cannot_delete_selected_cars') +
        '\n' +
        notDeletable.map((car) => `${car.id}: ${car.car_name || ''}`).join('\n'),
    )
    return
  }
  // Confirm deletion
  if (!confirm(t('carStock.confirm_delete_selected_cars'))) return
  // Get cars to delete
  const carsToDelete = sortedCars.value.filter(
    (car) =>
      selectedCars.value.has(car.id) &&
      !car.vin &&
      !car.id_client &&
      !car.container_ref &&
      !car.id_sell_bill,
  )
  const idsToDelete = carsToDelete.map((car) => car.id)
  // Delete from database
  try {
    const result = await callApi({
      query: `DELETE FROM cars_stock WHERE id IN (${idsToDelete.map(() => '?').join(',')})`,
      params: idsToDelete,
    })
    if (result.success) {
      // Delete cars from in-memory data
      allCars.value = allCars.value.filter((car) => !idsToDelete.includes(car.id))
      selectedCars.value.clear()
      fetchCarsStock()
      alert(t('carStock.deleted_selected_cars'))
    } else {
      alert(t('carStock.failed_to_delete_cars') + ': ' + (result.error || 'Unknown error'))
    }
  } catch (error) {
    alert(t('carStock.failed_to_delete_cars') + ': ' + error.message)
  }
}

const handleToggleHidden = async () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_toggle'))
    return
  }

  // Get selected cars
  const selectedCarsList = sortedCars.value.filter((car) => selectedCars.value.has(car.id))

  // Check if selected cars have mixed hidden status
  const hiddenCars = selectedCarsList.filter((car) => car.hidden === 1)
  const visibleCars = selectedCarsList.filter((car) => car.hidden === 0)

  if (hiddenCars.length > 0 && visibleCars.length > 0) {
    alert(t('carStock.mixed_hidden_status_error'))
    return
  }

  // Determine new hidden status (toggle based on current status)
  const newHiddenStatus = hiddenCars.length > 0 ? 0 : 1
  const actionText = newHiddenStatus === 1 ? t('carStock.hide_cars') : t('carStock.show_cars')

  // Confirm action
  if (!confirm(t('carStock.confirm_toggle_hidden', { action: actionText }))) return

  const carIds = selectedCarsList.map((car) => car.id)

  try {
    const currentUserId = user.value?.id
    const currentTimestamp = new Date().toISOString().slice(0, 19).replace('T', ' ')

    const result = await callApi({
      query: `UPDATE cars_stock SET hidden = ?, hidden_by_user_id = ?, hidden_time_stamp = ? WHERE id IN (${carIds.map(() => '?').join(',')})`,
      params: [
        newHiddenStatus,
        newHiddenStatus === 1 ? currentUserId : null,
        newHiddenStatus === 1 ? currentTimestamp : null,
        ...carIds,
      ],
    })

    if (result.success) {
      // Update in-memory data
      allCars.value = allCars.value.map((car) =>
        carIds.includes(car.id)
          ? {
              ...car,
              hidden: newHiddenStatus,
              hidden_by_user_id: newHiddenStatus === 1 ? currentUserId : null,
              hidden_time_stamp: newHiddenStatus === 1 ? currentTimestamp : null,
              hidden_by_username: newHiddenStatus === 1 ? user.value?.username : null,
            }
          : car,
      )
      selectedCars.value.clear()
      fetchCarsStock()
      alert(t('carStock.toggle_hidden_success', { action: actionText }))
    } else {
      alert(t('carStock.failed_to_toggle_hidden') + ': ' + (result.error || 'Unknown error'))
    }
  } catch (error) {
    alert(t('carStock.failed_to_toggle_hidden') + ': ' + error.message)
  }
}

const handleMarkDelivered = async () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_delivery'))
    return
  }

  const selectedCarsList = sortedCars.value.filter((car) => selectedCars.value.has(car.id))
  const carIds = selectedCarsList.map((car) => car.id)

  if (carIds.length === 0) {
    alert(t('carStock.no_cars_selected_for_delivery'))
    return
  }

  if (!confirm(t('carStock.confirm_mark_delivered', { count: carIds.length }))) {
    return
  }

  try {
    const result = await callApi({
      query: `UPDATE cars_stock 
        SET 
          date_send_documents = NOW(),
          date_loding = COALESCE(date_loding, NOW()),
          export_lisence_ref = COALESCE(NULLIF(export_lisence_ref, ''), 'ref'),
          in_wharhouse_date = COALESCE(in_wharhouse_date, NOW()),
          date_get_documents_from_supp = COALESCE(date_get_documents_from_supp, NOW()),
          date_get_bl = COALESCE(date_get_bl, NOW())
        WHERE id IN (${carIds.map(() => '?').join(',')})`,
      params: [...carIds],
    })

    if (result.success) {
      // Update local data to reflect the delivery without refetching
      allCars.value = allCars.value.filter((car) => !carIds.includes(car.id))
      cars.value = cars.value.filter((car) => !carIds.includes(car.id))

      selectedCars.value.clear()
      alert(t('carStock.marked_delivered_success', { count: carIds.length }))
    } else {
      alert(t('carStock.failed_to_mark_delivered') + (result.error ? `: ${result.error}` : ''))
    }
  } catch (error) {
    alert(t('carStock.failed_to_mark_delivered') + (error.message ? `: ${error.message}` : ''))
  }
}

// Add handlers for combine actions
const handleCombineByBuyRef = async () => {
  // Show loading indicator immediately
  isProcessing.value.combine = true

  // Force browser to render the loading indicator
  // Use setTimeout to yield to the browser's render cycle
  await new Promise((resolve) => setTimeout(resolve, 0))
  await new Promise((resolve) => requestAnimationFrame(resolve))
  await new Promise((resolve) => setTimeout(resolve, 50))

  // Now do the computation
  // Store previous sort config
  previousSortConfig.value = { ...sortConfig.value }

  // Enable combine mode with buy_ref type
  combineModeType.value = 'buy_ref'
  isCombineMode.value = true

  // Keep selection - don't clear it
  updateSelectAllState()

  // Wait for computation to complete
  await nextTick()
  await nextTick()
  await new Promise((resolve) => setTimeout(resolve, 100))

  // Hide loading indicator
  isProcessing.value.combine = false
}

const handleCombineByContainer = async () => {
  // Show loading indicator immediately
  isProcessing.value.combine = true

  // Force browser to render the loading indicator
  // Use setTimeout to yield to the browser's render cycle
  await new Promise((resolve) => setTimeout(resolve, 0))
  await new Promise((resolve) => requestAnimationFrame(resolve))
  await new Promise((resolve) => setTimeout(resolve, 50))

  // Now do the computation
  // Store previous sort config
  previousSortConfig.value = { ...sortConfig.value }

  // Enable combine mode with container type
  combineModeType.value = 'container'
  isCombineMode.value = true

  // Keep selection - don't clear it
  updateSelectAllState()

  // Wait for computation to complete
  await nextTick()
  await nextTick()
  await new Promise((resolve) => setTimeout(resolve, 100))

  // Hide loading indicator
  isProcessing.value.combine = false
}

const togglePaymentConfirmed = async (car) => {
  if (!can_confirm_payment.value) {
    alert(
      t('carStock.no_permission_to_confirm_payment') ||
        'You do not have permission to confirm payments',
    )
    return
  }

  if (isProcessing.value.payment) return

  isProcessing.value.payment = true

  try {
    // Check if payment_confirmed column exists by trying to update it
    // If column doesn't exist, show message to run migration
    const currentStatus = car.payment_confirmed !== undefined ? car.payment_confirmed : 0
    const newStatus = currentStatus ? 0 : 1

    // Get current user info for permission check
    if (!user.value || !user.value.id) {
      alert('User authentication required')
      return
    }

    const result = await callApi({
      query: `
        UPDATE cars_stock
        SET payment_confirmed = ?
        WHERE id = ?
      `,
      params: [newStatus, car.id],
      user_id: user.value.id,
      is_admin: user.value.role_id === 1,
      requiresAuth: true,
    })

    if (result.success) {
      // Update local state
      const carIndex = cars.value.findIndex((c) => c.id === car.id)
      if (carIndex !== -1) {
        cars.value[carIndex].payment_confirmed = newStatus
      }
    } else {
      const errorMsg = result.error || 'Unknown error'
      // Check if it's a missing column error
      if (
        errorMsg.includes("Unknown column 'payment_confirmed'") ||
        errorMsg.includes("doesn't exist")
      ) {
        alert(
          'Payment confirmed column does not exist. Please run migration 009_add_payment_confirmed_to_cars_stock.sql',
        )
      } else {
        alert('Failed to update payment confirmed status: ' + errorMsg)
      }
    }
  } catch (err) {
    alert(
      t('carStock.error_updating_payment_confirmed') ||
        'Error updating payment confirmed status: ' + err.message,
    )
  } finally {
    isProcessing.value.payment = false
  }
}

const handleUncombine = () => {
  // Restore previous sort config if available
  if (previousSortConfig.value) {
    sortConfig.value = { ...previousSortConfig.value }
    previousSortConfig.value = null
  }

  // Disable combine mode
  isCombineMode.value = false
  combineModeType.value = null

  // Keep selection - don't clear it
  updateSelectAllState()
}

// Group cars by buy_bill_ref for combine mode
const groupedCarsByBuyRef = computed(() => {
  if (!isCombineMode.value || !sortedCars.value) return {}

  const groups = {}
  sortedCars.value.forEach((car) => {
    const buyRef = car.buy_bill_ref || 'no_buy_bill'
    if (!groups[buyRef]) {
      groups[buyRef] = []
    }
    groups[buyRef].push(car)
  })

  return groups
})

// Get buy ref for a car (used in template)
const getBuyRefForCar = (car) => {
  return car.buy_bill_ref || null
}

// Check if this is the first car in a group (buy ref or container)
const isFirstCarInGroup = (car, index) => {
  if (!isCombineMode.value) return false
  if (index === 0) return true

  if (combineModeType.value === 'buy_ref') {
    const currentBuyRef = car.buy_bill_ref || null
    const previousCar = sortedCars.value[index - 1]
    const previousBuyRef = previousCar?.buy_bill_ref || null
    return currentBuyRef !== previousBuyRef
  } else if (combineModeType.value === 'container') {
    const currentContainer = car.container_ref || null
    const previousCar = sortedCars.value[index - 1]
    const previousContainer = previousCar?.container_ref || null
    return currentContainer !== previousContainer
  }

  return false
}

// Get row span for group (buy ref or container)
const getGroupRowSpan = (car, index) => {
  if (!isCombineMode.value) return 0

  if (combineModeType.value === 'buy_ref') {
    const buyRef = car.buy_bill_ref || null
    if (!buyRef) return 0

    // Count how many consecutive cars have the same buy ref
    let count = 1
    for (let i = index + 1; i < sortedCars.value.length; i++) {
      if (sortedCars.value[i].buy_bill_ref === buyRef) {
        count++
      } else {
        break
      }
    }
    return count
  } else if (combineModeType.value === 'container') {
    const containerRef = car.container_ref || null
    if (!containerRef) return 0

    // Count how many consecutive cars have the same container ref
    let count = 1
    for (let i = index + 1; i < sortedCars.value.length; i++) {
      if (sortedCars.value[i].container_ref === containerRef) {
        count++
      } else {
        break
      }
    }
    return count
  }

  return 0
}

// Toggle selection for all cars in a group (buy ref or container)
const toggleGroupSelection = (groupValue) => {
  if (!isCombineMode.value) return

  let carsInGroup = []
  if (combineModeType.value === 'buy_ref') {
    carsInGroup = sortedCars.value.filter((car) => car.buy_bill_ref === groupValue)
  } else if (combineModeType.value === 'container') {
    carsInGroup = sortedCars.value.filter((car) => car.container_ref === groupValue)
  }

  const allSelected = carsInGroup.every((car) => selectedCars.value.has(car.id))

  if (allSelected) {
    // Deselect all cars in group
    carsInGroup.forEach((car) => {
      selectedCars.value.delete(car.id)
    })
  } else {
    // Select all cars in group
    carsInGroup.forEach((car) => {
      selectedCars.value.add(car.id)
    })
  }

  updateSelectAllState()
}

// Check if all cars in a group are selected (buy ref or container)
const isGroupSelected = (groupValue) => {
  if (!isCombineMode.value) return false

  let carsInGroup = []
  if (combineModeType.value === 'buy_ref') {
    carsInGroup = sortedCars.value.filter((car) => car.buy_bill_ref === groupValue)
  } else if (combineModeType.value === 'container') {
    carsInGroup = sortedCars.value.filter((car) => car.container_ref === groupValue)
  }

  if (carsInGroup.length === 0) return false
  return carsInGroup.every((car) => selectedCars.value.has(car.id))
}

// Get count of cars in a group (buy ref or container)
const getGroupCount = (groupValue) => {
  if (!isCombineMode.value) return 0

  if (combineModeType.value === 'buy_ref') {
    return sortedCars.value.filter((car) => car.buy_bill_ref === groupValue).length
  } else if (combineModeType.value === 'container') {
    return sortedCars.value.filter((car) => car.container_ref === groupValue).length
  }

  return 0
}

// Get group value for display (buy ref or container)
const getGroupValue = (car) => {
  if (combineModeType.value === 'buy_ref') {
    return car.buy_bill_ref || null
  } else if (combineModeType.value === 'container') {
    return car.container_ref || null
  }
  return null
}

// Get group label for header
const getGroupHeaderLabel = computed(() => {
  if (combineModeType.value === 'buy_ref') {
    return buyBillRefHeader.value
  } else if (combineModeType.value === 'container') {
    const translation = t('carStock.container_ref')
    if (
      translation === 'carStock.container_ref' ||
      translation === 'en' ||
      translation === locale.value
    ) {
      return 'Container Ref'
    }
    return translation
  }
  return ''
})

// Handlers for selections
const handleSaveSelection = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_saving') || 'No cars selected')
    return
  }
  showSaveSelectionForm.value = true
}

const handleSendSelection = () => {
  if (selectedCars.value.size === 0) {
    alert(t('carStock.no_cars_selected_for_sending') || 'No cars selected')
    return
  }
  showSendSelectionForm.value = true
}

const handleShowSelections = () => {
  showSelectionsModal.value = true
}

const handleSelectionSaved = (result) => {
  alert(t('carStock.selection_saved_successfully') || 'Selection saved successfully')
  showSaveSelectionForm.value = false
}

const handleSelectionSent = (result) => {
  alert(t('carStock.selection_sent_successfully') || 'Selection sent successfully')
  showSendSelectionForm.value = false
}

const handleSelectionLoaded = (selection) => {
  // Clear current selection
  selectedCars.value.clear()
  // Select cars from loaded selection
  selection.carIds.forEach((carId) => {
    selectedCars.value.add(carId)
  })
  updateSelectAllState()
  alert(
    `${selection.carIds.length} ${selection.carIds.length === 1 ? t('carStockToolbar.car') || 'car' : t('carStockToolbar.cars') || 'cars'} ${t('carStock.loaded') || 'loaded'}`,
  )
}

const handleTransfer = async () => {
  try {
    // Validate required functions are available
    if (typeof getCarFiles !== 'function') {
      throw new Error('getCarFiles function is not available')
    }
    if (typeof transferPhysicalCopy !== 'function') {
      throw new Error('transferPhysicalCopy function is not available')
    }
    if (typeof getUsersForTransfer !== 'function') {
      throw new Error('getUsersForTransfer function is not available')
    }

    if (!selectedCars || !selectedCars.value || selectedCars.value.size === 0) {
      alert(t('carStock.no_cars_selected_for_transfer') || 'No cars selected for transfer')
      return
    }

    if (!user.value || !user.value.id) {
      alert(t('carStock.user_not_authenticated') || 'User not authenticated. Please log in again.')
      return
    }

    // Get all selected cars - use sortedCars if available, otherwise use cars
    const carsList = sortedCars?.value || cars.value || []
    const selectedCarsList = carsList.filter((car) => selectedCars.value.has(car.id))

    if (selectedCarsList.length === 0) {
      alert(t('carStock.no_cars_selected_for_transfer') || 'No cars selected for transfer')
      return
    }

    // Collect all files from all selected cars
    const allFiles = []
    const currentUserId = user.value.id
    const errors = []

    for (const car of selectedCarsList) {
      try {
        const carFiles = await getCarFiles(car.id)

        if (!Array.isArray(carFiles)) {
          continue
        }

        // Filter eligible files for this car
        const eligibleFiles = carFiles.filter((f) => {
          // Skip files with pending transfers
          if (f.has_pending_transfer) {
            return false
          }

          // Include files checked out to current user
          if (f.physical_status === 'checked_out' && f.current_holder_id === currentUserId) {
            return true
          }

          // Include available files where current user is holder
          if (f.physical_status === 'available' && f.current_holder_id === currentUserId) {
            return true
          }

          // Include files with no tracking record where current user is uploader
          if (f.physical_status === null && f.uploaded_by === currentUserId) {
            return true
          }

          return false
        })

        // Add car info to each file for reference
        eligibleFiles.forEach((file) => {
          allFiles.push({
            ...file,
            car_id: car.id,
            car_name: car.car_name || `Car #${car.id}`,
          })
        })
      } catch (err) {
        const errorMsg = err.message || String(err)
        errors.push(`Car ${car.id}: ${errorMsg}`)
      }
    }

    if (allFiles.length === 0) {
      if (errors.length > 0) {
        alert(
          t('carStock.no_files_available_for_batch_transfer_with_errors') ||
            `No files available for batch transfer. Errors: ${errors.join('; ')}`,
        )
      } else {
        alert(
          t('carStock.no_files_available_for_batch_transfer') ||
            'No files available for batch transfer. Please ensure you are the holder of the documents.',
        )
      }
      return
    }

    // Store files for batch transfer
    batchTransferFiles.value = allFiles
    batchTransferToUserId.value = null
    batchTransferNotes.value = ''
    batchTransferExpectedReturn.value = ''

    // Get users for transfer (use first file to get users)
    try {
      if (!allFiles[0] || !allFiles[0].id) {
        throw new Error('Invalid file data: missing file ID')
      }

      const users = await getUsersForTransfer(allFiles[0].id)

      if (!Array.isArray(users)) {
        throw new Error('Invalid response from getUsersForTransfer: expected array')
      }

      batchTransferAvailableUsers.value = users

      if (users.length === 0) {
        alert(
          t('carStock.no_users_available_for_transfer') ||
            'No users available for transfer. Please contact an administrator.',
        )
        return
      }

      showBatchTransferModal.value = true
    } catch (err) {
      const errorMsg = err.message || String(err)
      alert(
        t('carStock.failed_to_load_users_for_transfer') ||
          'Failed to load users for transfer: ' + errorMsg,
      )
    }
  } catch (err) {
    const errorMsg = err.message || String(err)
    const errorStack = err.stack || ''
    alert(
      (t('carStock.failed_to_prepare_batch_transfer') || 'Failed to prepare batch transfer') +
        ': ' +
        errorMsg +
        (errorStack ? '\n\nCheck console for details.' : ''),
    )
  }
}

const confirmBatchTransfer = async () => {
  if (!batchTransferToUserId.value) {
    alert(t('carStock.please_select_user_to_transfer_to') || 'Please select a user to transfer to')
    return
  }

  if (batchTransferFiles.value.length === 0) {
    alert(t('carStock.no_files_to_transfer') || 'No files to transfer')
    return
  }

  isBatchTransferring.value = true

  try {
    const results = []
    const errors = []
    const currentUserId = user.value?.id

    // Loop through all files and transfer each one
    for (const file of batchTransferFiles.value) {
      try {
        // Double-check eligibility before transferring
        const isEligible =
          !file.has_pending_transfer &&
          ((file.physical_status === 'checked_out' && file.current_holder_id === currentUserId) ||
            (file.physical_status === 'available' && file.current_holder_id === currentUserId) ||
            (file.physical_status === null && file.uploaded_by === currentUserId))

        if (!isEligible) {
          errors.push({
            file: file.file_name,
            car: file.car_name,
            error: 'File is not eligible for transfer',
          })
          continue
        }

        await transferPhysicalCopy(
          file.id,
          batchTransferToUserId.value,
          batchTransferNotes.value || null,
          batchTransferExpectedReturn.value || null,
        )
        results.push({ file: file.file_name, car: file.car_name })
      } catch (err) {
        errors.push({
          file: file.file_name,
          car: file.car_name,
          error: err.message || 'Failed to transfer',
        })
      }
    }

    // Show results
    if (errors.length > 0) {
      const errorMsg =
        t('carStock.batch_transfer_completed_with_errors') ||
        `Batch transfer completed with errors. ${results.length} succeeded, ${errors.length} failed.`
      alert(errorMsg)
    } else {
      const successMsg =
        t('carStock.batch_transfer_success') || `Successfully transferred ${results.length} file(s)`
      alert(successMsg)
    }

    // Close modal and refresh
    showBatchTransferModal.value = false
    batchTransferFiles.value = []
    fetchCarsStock()
  } catch (err) {
    alert(
      t('carStock.failed_to_perform_batch_transfer') ||
        'Failed to perform batch transfer: ' + err.message,
    )
  } finally {
    isBatchTransferring.value = false
  }
}

const closeBatchTransferModal = () => {
  showBatchTransferModal.value = false
  batchTransferFiles.value = []
  batchTransferToUserId.value = null
  batchTransferNotes.value = ''
  batchTransferExpectedReturn.value = ''
}

const handleCheckout = async () => {
  try {
    if (selectedCars.value.size === 0) {
      alert(t('carStock.no_cars_selected_for_checkout') || 'No cars selected for checkout')
      return
    }

    if (!user.value || !user.value.id) {
      alert(t('carStock.user_not_authenticated') || 'User not authenticated. Please log in again.')
      return
    }

    // Get all selected cars - use sortedCars if available, otherwise use cars
    const carsList = sortedCars?.value || cars.value || []
    const selectedCarsList = carsList.filter((car) => car && selectedCars.value.has(car.id))

    if (selectedCarsList.length === 0) {
      alert(t('carStock.no_cars_selected_for_checkout') || 'No cars selected for checkout')
      return
    }

    // Collect all files from all selected cars
    const allFiles = []
    const currentUserId = user.value.id
    const errors = []

    for (const car of selectedCarsList) {
      if (!car || !car.id) {
        continue
      }

      try {
        if (!getCarFiles) {
          throw new Error('getCarFiles function is not available')
        }

        const carFiles = await getCarFiles(car.id)

        if (!Array.isArray(carFiles)) {
          continue
        }

        // Filter eligible files for this car
        const eligibleFiles = carFiles.filter((f) => {
          // Skip files with pending transfers
          if (f.has_pending_transfer) {
            return false
          }

          // Include files checked out to current user
          if (f.physical_status === 'checked_out' && f.current_holder_id === currentUserId) {
            return true
          }

          // Include available files where current user is holder
          if (f.physical_status === 'available' && f.current_holder_id === currentUserId) {
            return true
          }

          // Include files with no tracking record where current user is uploader
          if (f.physical_status === null && f.uploaded_by === currentUserId) {
            return true
          }

          return false
        })

        // Add car info to each file for reference
        eligibleFiles.forEach((file) => {
          allFiles.push({
            ...file,
            car_id: car.id,
            car_name: car.car_name || `Car #${car.id}`,
          })
        })
      } catch (err) {
        const errorMsg = err.message || String(err)
        errors.push(`Car ${car.id}: ${errorMsg}`)
      }
    }

    if (allFiles.length === 0) {
      if (errors.length > 0) {
        alert(
          t('carStock.no_files_available_for_batch_checkout_with_errors') ||
            `No files available for batch checkout. Errors: ${errors.join('; ')}`,
        )
      } else {
        alert(
          t('carStock.no_files_available_for_batch_checkout') ||
            'No files available for batch checkout. Please ensure you are the holder of the documents.',
        )
      }
      return
    }

    // Store files for batch checkout
    batchCheckoutFiles.value = allFiles
    batchCheckoutType.value = 'user'
    batchCheckoutUserId.value = null
    batchCheckoutAgentId.value = null
    batchCheckoutClientId.value = null
    batchCheckoutClientName.value = ''
    batchCheckoutClientContact.value = ''
    batchCheckoutNotes.value = ''

    // Load users, clients, and agents
    try {
      if (!allFiles[0] || !allFiles[0].id) {
        throw new Error('Invalid file data: missing file ID')
      }

      // Load users
      if (getUsersForTransfer) {
        const users = await getUsersForTransfer(allFiles[0].id)
        batchCheckoutAvailableUsers.value = Array.isArray(users) ? users : []
      }

      // Load clients
      const clientsResult = await callApi({
        query:
          'SELECT id, name, email, mobiles, address FROM clients WHERE is_client = 1 ORDER BY name ASC',
        params: [],
        requiresAuth: true,
      })
      if (clientsResult.success) {
        batchCheckoutAvailableClients.value = clientsResult.data || []
      }

      // Load agents
      if (getCustomClearanceAgents) {
        const agents = await getCustomClearanceAgents()
        batchCheckoutAvailableAgents.value = Array.isArray(agents) ? agents : []
      }

      if (
        batchCheckoutAvailableUsers.value.length === 0 &&
        batchCheckoutAvailableClients.value.length === 0 &&
        batchCheckoutAvailableAgents.value.length === 0
      ) {
        alert(
          t('carStock.no_checkout_options_available') ||
            'No checkout options available. Please contact an administrator.',
        )
        return
      }

      showBatchCheckoutModal.value = true
    } catch (err) {
      const errorMsg = err.message || String(err)
      alert(
        t('carStock.failed_to_load_checkout_data') || 'Failed to load checkout data: ' + errorMsg,
      )
    }
  } catch (err) {
    const errorMsg = err.message || String(err)
    const errorStack = err.stack || ''
    alert(
      (t('carStock.failed_to_prepare_batch_checkout') || 'Failed to prepare batch checkout') +
        ': ' +
        errorMsg +
        (errorStack ? '\n\nCheck console for details.' : ''),
    )
  }
}

const confirmBatchCheckout = async () => {
  // Validate based on checkout type
  if (batchCheckoutType.value === 'user' && !batchCheckoutUserId.value) {
    alert(t('carStock.please_select_user_for_checkout') || 'Please select a user')
    return
  }
  if (batchCheckoutType.value === 'custom_clearance_agent' && !batchCheckoutAgentId.value) {
    alert(
      t('carStock.please_select_agent_for_checkout') || 'Please select a custom clearance agent',
    )
    return
  }
  if (
    batchCheckoutType.value === 'client' &&
    !batchCheckoutClientId.value &&
    !batchCheckoutClientName.value
  ) {
    alert(
      t('carStock.please_select_client_for_checkout') ||
        'Please select a client or enter client name',
    )
    return
  }

  if (batchCheckoutFiles.value.length === 0) {
    alert(t('carStock.no_files_to_checkout') || 'No files to checkout')
    return
  }

  isBatchCheckouting.value = true

  try {
    const results = []
    const errors = []
    const currentUserId = user.value?.id

    // Loop through all files and checkout each one
    for (const file of batchCheckoutFiles.value) {
      try {
        // Double-check eligibility before checking out
        const isEligible =
          !file.has_pending_transfer &&
          ((file.physical_status === 'checked_out' && file.current_holder_id === currentUserId) ||
            (file.physical_status === 'available' && file.current_holder_id === currentUserId) ||
            (file.physical_status === null && file.uploaded_by === currentUserId))

        if (!isEligible) {
          errors.push({
            file: file.file_name,
            car: file.car_name,
            error: 'File is not eligible for checkout',
          })
          continue
        }

        const result = await callApi({
          action: 'checkout_physical_copy',
          file_id: file.id,
          checkout_type: batchCheckoutType.value,
          user_id: batchCheckoutType.value === 'user' ? batchCheckoutUserId.value : null,
          agent_id:
            batchCheckoutType.value === 'custom_clearance_agent'
              ? batchCheckoutAgentId.value
              : null,
          client_id: batchCheckoutType.value === 'client' ? batchCheckoutClientId.value : null,
          client_name: batchCheckoutType.value === 'client' ? batchCheckoutClientName.value : null,
          client_contact:
            batchCheckoutType.value === 'client' ? batchCheckoutClientContact.value : null,
          notes: batchCheckoutNotes.value || null,
          performed_by: currentUserId,
          requiresAuth: true,
        })

        if (result.success) {
          results.push({ file: file.file_name, car: file.car_name })
        } else {
          errors.push({
            file: file.file_name,
            car: file.car_name,
            error: result.error || 'Failed to checkout',
          })
        }
      } catch (err) {
        errors.push({
          file: file.file_name,
          car: file.car_name,
          error: err.message || 'Failed to checkout',
        })
      }
    }

    // Show results
    if (errors.length > 0) {
      const errorMsg =
        t('carStock.batch_checkout_completed_with_errors') ||
        `Batch checkout completed with errors. ${results.length} succeeded, ${errors.length} failed.`
      alert(errorMsg)
    } else {
      const successMsg =
        t('carStock.batch_checkout_success') || `Successfully checked out ${results.length} file(s)`
      alert(successMsg)
    }

    // Close modal and refresh
    showBatchCheckoutModal.value = false
    batchCheckoutFiles.value = []
    fetchCarsStock()
  } catch (err) {
    alert(
      t('carStock.failed_to_perform_batch_checkout') ||
        'Failed to perform batch checkout: ' + err.message,
    )
  } finally {
    isBatchCheckouting.value = false
  }
}

const handleClientChange = () => {
  if (batchCheckoutClientId.value) {
    const client = batchCheckoutAvailableClients.value.find(
      (c) => c.id === batchCheckoutClientId.value,
    )
    if (client) {
      batchCheckoutClientName.value = client.name || ''
      batchCheckoutClientContact.value = client.mobiles || client.email || ''
    }
  }
}

const closeBatchCheckoutModal = () => {
  showBatchCheckoutModal.value = false
  batchCheckoutFiles.value = []
  batchCheckoutType.value = 'user'
  batchCheckoutUserId.value = null
  batchCheckoutAgentId.value = null
  batchCheckoutClientId.value = null
  batchCheckoutClientName.value = ''
  batchCheckoutClientContact.value = ''
  batchCheckoutNotes.value = ''
}
</script>

<template>
  <div class="cars-stock-table">
    <!-- Add all form components -->
    <VinEditForm
      v-if="showVinEditForm"
      :car="selectedCar"
      :show="showVinEditForm"
      @close="showVinEditForm = false"
      @save="handleVinSave"
    />

    <CarFilesManagement
      v-if="showFilesUploadForm"
      :car="selectedCar"
      :show="showFilesUploadForm"
      @close="showFilesUploadForm = false"
      @save="handleFilesSave"
    />

    <CarPortsEditForm
      v-if="showPortsEditForm"
      :car="selectedCar"
      :show="showPortsEditForm"
      @close="showPortsEditForm = false"
      @save="handlePortsSave"
    />

    <CarMoneyEditForm
      v-if="showMoneyEditForm"
      :car="selectedCar"
      :show="showMoneyEditForm"
      @close="showMoneyEditForm = false"
      @save="handleMoneySave"
    />

    <CarWarehouseForm
      v-if="showWarehouseForm"
      :car="selectedCar"
      :cars="
        selectedCars.size > 1 ? sortedCars.filter((car) => selectedCars.has(car.id)) : undefined
      "
      :show="showWarehouseForm"
      :is-admin="isAdmin"
      @close="showWarehouseForm = false"
      @save="handleWarehouseSave"
      @warehouse-changed="handleWarehouseChanged"
    />

    <CarDocumentsForm
      v-if="showDocumentsForm"
      :car="selectedCarForDocuments"
      :show="showDocumentsForm"
      @close="handleDocumentsClose"
      @save="handleDocumentsSave"
      @documents-sent="fetchCarsStock"
    />

    <CarLoadForm
      v-if="showLoadForm"
      :car="selectedCarForLoad"
      :show="showLoadForm"
      @close="handleLoadClose"
      @save="handleLoadSave"
    />

    <TaskForm
      v-if="selectedCarForTask"
      :entity-data="selectedCarForTask"
      entity-type="car"
      :is-visible="showTaskForm"
      @task-created="handleTaskSave"
      @cancel="handleTaskClose"
    />

    <CarStockSwitchBuyBill
      v-if="selectedCarForSwitchBuyBill"
      :car="selectedCarForSwitchBuyBill"
      :show="showSwitchBuyBillForm"
      @close="handleSwitchBuyBillClose"
      @save="handleSwitchBuyBillSave"
    />

    <CarStockPrintOptions
      :show="showPrintOptions"
      :selected-cars="sortedCars.filter((car) => selectedCars.has(car.id))"
      :action-type="printOptionsActionType"
      @close="handlePrintOptionsClose"
      @print="handlePrintWithOptions"
      @loading-order="handleLoadingOrderWithOptions"
    />

    <!-- <div class="table-actions">
      <button class="advanced-filter-btn" @click="showAdvancedFilter = true">
        <i class="fas fa-filter"></i>
        Advanced Filters
      </button>
    </div> -->

    <!-- <AdvancedFilter
      :show="showAdvancedFilter"
      :initial-filters="advancedFilters"
      @update:filters="handleAdvancedFilters"
      @close="showAdvancedFilter = false"
    /> -->

    <div v-if="loading" class="loading">
      <i class="fas fa-spinner fa-spin"></i>
      {{ t('carStock.loading') }}
    </div>
    <div v-else-if="error" class="error">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>
    <div v-else-if="cars.length === 0" class="empty-state">
      <i class="fas fa-car fa-2x"></i>
      <p>{{ t('carStock.no_cars_in_stock') }}</p>
    </div>
    <div v-else>
      <!-- Toolbar -->
      <CarStockToolbar
        :selected-cars="selectedCars"
        :total-cars="sortedCars.length"
        :can-change-color="can_change_car_color"
        :can-hide-car="can_hide_car"
        :is-admin="user?.role_id === 1"
        :is-combine-mode="isCombineMode"
        @print-selected="handlePrintSelected"
        @loading-order="handleLoadingOrderFromToolbar"
        @vin="handleVinFromToolbar"
        @ports="handlePortsFromToolbar"
        @warehouse="handleWarehouseFromToolbar"
        @notes="handleNotesFromToolbar"
        @task="handleTaskFromToolbar"
        @color="handleColorFromToolbar"
        @export-license="handleExportLicenseFromToolbar"
        @cfr-da="handleCfrDaFromToolbar"
        @mark-delivered="handleMarkDelivered"
        @refresh="handleRefresh"
        @delete-cars="handleDeleteCars"
        @toggle-hidden="handleToggleHidden"
        @combine-by-buy-ref="handleCombineByBuyRef"
        @combine-by-container="handleCombineByContainer"
        @uncombine="handleUncombine"
        @save-selection="handleSaveSelection"
        @send-selection="handleSendSelection"
        @show-selections="handleShowSelections"
        @transfer="handleTransfer"
        @checkout="handleCheckout"
      />

      <div class="table-container" :class="{ 'processing-combine': isProcessing.combine }">
        <table class="cars-table">
          <thead>
            <tr>
              <th v-if="isCombineMode" class="buy-ref-group-header frozen-column">
                {{ getGroupHeaderLabel }}
              </th>
              <th class="select-all-header">
                <input
                  type="checkbox"
                  :checked="selectAll"
                  @change="selectAllCars"
                  :indeterminate="selectedCars.size > 0 && selectedCars.size < sortedCars.length"
                  :title="t('carStock.select_all_cars')"
                />
              </th>
              <th>{{ t('carStock.actions') }}</th>
              <th @click="toggleSort('id')" class="sortable">
                {{ t('carStock.id') }}
                <span v-if="sortConfig.key === 'id'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>

              <th @click="toggleSort('date_buy')" class="sortable">
                {{ t('carStock.date_buy') }}
                <span v-if="sortConfig.key === 'date_buy'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('date_sell')" class="sortable">
                {{ t('carStock.date_sell') }}
                <span v-if="sortConfig.key === 'date_sell'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('car_name')" class="sortable">
                {{ t('carStock.car_details') }}
                <span v-if="sortConfig.key === 'car_name'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('client_name')" class="sortable">
                {{ t('carStock.client') }}
                <span v-if="sortConfig.key === 'client_name'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('loading_port')" class="sortable ports-column">
                {{ t('carStock.ports') }}
                <span v-if="sortConfig.key === 'loading_port'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('freight')" class="sortable freight-column">
                {{ t('carStock.freight') }}
                <span v-if="sortConfig.key === 'freight'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('price_cell')" class="sortable">
                {{ t('carStock.fob') }}
                <span v-if="sortConfig.key === 'price_cell'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('rate')" class="sortable">
                {{ t('carStock.cfr') }}
                <span v-if="sortConfig.key === 'rate'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>

              <th @click="toggleSort('status')" class="sortable">
                {{ t('carStock.status') }}
                <span v-if="sortConfig.key === 'status'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th>{{ t('carStock.payment_confirmed') }}</th>
              <th>{{ t('carStock.documents') }}</th>
              <th @click="toggleSort('notes')" class="sortable">
                {{ t('carStock.notes') }}
                <span v-if="sortConfig.key === 'notes'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(car, index) in sortedCars"
              :key="car.id"
              :class="{
                'used-car': car.is_used_car,
                selected: selectedCars.has(car.id),
                'hidden-car': car.hidden,
                'buy-ref-group-row':
                  isCombineMode &&
                  ((combineModeType.value === 'buy_ref' && car.buy_bill_ref) ||
                    (combineModeType.value === 'container' && car.container_ref)),
              }"
            >
              <td
                v-if="isCombineMode"
                v-show="isFirstCarInGroup(car, index)"
                :rowspan="getGroupRowSpan(car, index)"
                class="buy-ref-group-cell frozen-column"
                :class="{ 'group-selected': isGroupSelected(getGroupValue(car)) }"
              >
                <div class="buy-ref-group-content">
                  <input
                    type="checkbox"
                    :checked="isGroupSelected(getGroupValue(car))"
                    @change="toggleGroupSelection(getGroupValue(car))"
                    :title="`Select all cars in ${getGroupValue(car) || (combineModeType.value === 'buy_ref' ? t('carStock.no_buy_bill') : t('carStock.no_container_ref'))}`"
                    class="group-checkbox"
                  />
                  <div class="buy-ref-text">
                    <div class="buy-ref-label">
                      {{
                        getGroupValue(car) ||
                        (combineModeType.value === 'buy_ref'
                          ? t('carStock.no_buy_bill')
                          : t('carStock.no_container_ref'))
                      }}
                    </div>
                    <div class="buy-ref-count">
                      {{ getGroupCount(getGroupValue(car)) }}
                      {{
                        getGroupCount(getGroupValue(car)) === 1
                          ? t('carStockToolbar.car')
                          : t('carStockToolbar.cars')
                      }}
                    </div>
                  </div>
                </div>
              </td>
              <td class="select-cell">
                <input
                  type="checkbox"
                  :checked="selectedCars.has(car.id)"
                  @change="toggleCarSelection(car.id)"
                  :title="`Select car #${car.id}`"
                />
              </td>
              <td>
                <!-- Teleport Dropdown Button -->
                <button
                  @click="openTeleportDropdown(car.id, $event)"
                  class="teleport-dropdown-toggle"
                  title="Actions"
                >
                  <i class="fas fa-ellipsis-h"></i>
                </button>
              </td>
              <td>{{ car.id }}</td>

              <td>{{ car.date_buy ? new Date(car.date_buy).toLocaleDateString() : '-' }}</td>
              <td>{{ car.date_sell ? new Date(car.date_sell).toLocaleDateString() : '-' }}</td>
              <td class="car-details-cell">
                <div class="car-details-container">
                  <div class="car-name">
                    {{ car.car_name }}
                  </div>
                  <div v-if="car.vin" class="car-detail-item">
                    <div class="info-badge badge-vin">
                      <i class="fas fa-barcode"></i>
                      {{ car.vin }}
                    </div>
                  </div>
                  <div v-if="car.color" class="car-detail-item">
                    <div
                      class="info-badge badge-color"
                      :style="{
                        backgroundColor: car.hexa || '#000000',
                        color: getTextColor(car.hexa || '#000000'),
                      }"
                    >
                      <i class="fas fa-palette"></i>
                      {{ car.color }}
                    </div>
                  </div>
                  <div v-if="car.export_lisence_ref" class="info-badge badge-export-license">
                    <i class="fas fa-certificate"></i>
                    {{ car.export_lisence_ref }}
                  </div>

                  <div v-if="car.buy_bill_ref" class="car-detail-item">
                    <div class="info-badge badge-buy-bill">
                      <i class="fas fa-shopping-cart"></i>
                      {{ t('carStock.buy') }}: {{ car.buy_bill_ref }}
                    </div>
                  </div>
                  <div v-if="car.sell_bill_ref" class="car-detail-item">
                    <div class="info-badge badge-sell-bill">
                      <i class="fas fa-file-invoice-dollar"></i>
                      {{ t('carStock.sell') }}: {{ car.sell_bill_ref }}
                    </div>
                  </div>
                </div>
              </td>
              <td>
                <div class="client-info">
                  <div>{{ car.client_name || '-' }}</div>
                  <div v-if="car.client_id_no" class="info-badge badge-client-id">
                    <i class="fas fa-id-card"></i>
                    {{ car.client_id_no }}
                  </div>
                  <div v-if="car.is_batch" class="info-badge badge-wholesale">
                    <i class="fas fa-layer-group"></i>
                    {{ t('carStock.whole_sale') }}
                  </div>
                </div>
              </td>
              <td class="ports-column">
                <div class="ports-container">
                  <div class="port-item">
                    <div v-if="car.loading_port" class="info-badge badge-loading-port">
                      <i class="fas fa-ship"></i>
                      {{ car.loading_port }}
                    </div>
                    <div v-else class="port-empty">-</div>
                  </div>
                  <div class="port-item">
                    <div v-if="car.discharge_port" class="info-badge badge-discharge-port">
                      <i class="fas fa-anchor"></i>
                      {{ car.discharge_port }}
                    </div>
                    <div v-else class="port-empty">-</div>
                  </div>
                  <div v-if="car.container_ref" class="info-badge badge-container">
                    <i class="fas fa-box"></i>
                    {{ car.container_ref }}
                  </div>
                  <div v-if="car.date_pay_freight" class="info-badge badge-freight-paid">
                    <i class="fas fa-check-circle"></i>
                    {{ t('carStock.freight_paid') }}
                  </div>
                </div>
              </td>
              <td class="freight-column">${{ getFreightValue(car) }}</td>
              <td>${{ car.price_cell || '0' }}</td>
              <td>
                <div class="cfr-container">
                  <div class="info-badge badge-cfr-usd">
                    <i class="fas fa-dollar-sign"></i>
                    USD: ${{ getCfrUsdValue(car) }}
                  </div>
                  <div class="info-badge badge-cfr-dza">
                    <i class="fas fa-coins"></i>
                    DZA: {{ getCfrDzaValue(car) }}
                  </div>
                  <div class="info-badge badge-cfr-rate">
                    <i class="fas fa-percentage"></i>
                    Rate: {{ getRateValue(car) }}
                  </div>
                </div>
              </td>

              <td class="status-column">
                <div class="status-container">
                  <div class="status-item">
                    <span
                      :class="{
                        'status-sold': car.id_sell,
                        'status-available': !car.id_sell,
                        'status-used': car.is_used_car,
                      }"
                    >
                      <i
                        :class="[
                          car.id_sell ? 'fas fa-check-circle' : 'fas fa-clock',
                          car.is_used_car ? 'fas fa-history' : '',
                        ]"
                      ></i>
                      {{ car.id_sell ? t('carStock.sold') : t('carStock.available') }}
                      {{ car.is_used_car ? `(${t('carStock.used')})` : '' }}
                    </span>
                    <i
                      v-if="car.hidden"
                      class="fas fa-eye-slash hidden-icon"
                      :title="
                        car.hidden_by_username
                          ? t('carStock.hidden_by_user', {
                              user: car.hidden_by_username,
                              date: car.hidden_time_stamp,
                            })
                          : t('carStock.hidden_car')
                      "
                    ></i>
                  </div>
                  <div v-if="car.hidden && car.hidden_by_username" class="status-item hidden-info">
                    <div class="info-badge badge-hidden">
                      <i class="fas fa-user-slash"></i>
                      {{ t('carStock.hidden_by') }}: {{ car.hidden_by_username }}
                    </div>
                    <div class="info-badge badge-hidden-time">
                      <i class="fas fa-clock"></i>
                      {{ formatHiddenTime(car.hidden_time_stamp) }}
                    </div>
                  </div>
                  <div v-if="car.in_wharhouse_date" class="status-item">
                    <div class="info-badge badge-in-warehouse">
                      <i class="fas fa-warehouse"></i>
                      {{ t('carStock.in_warehouse') }}
                    </div>
                  </div>
                  <div v-if="car.date_get_bl" class="status-item">
                    <div class="info-badge badge-bl-received">
                      <i class="fas fa-file-contract"></i>
                      {{ t('carStock.bl_received') }}
                    </div>
                  </div>
                </div>
              </td>

              <td class="payment-confirmed-column">
                <button
                  @click="can_confirm_payment ? togglePaymentConfirmed(car) : null"
                  :disabled="isProcessing.payment || !can_confirm_payment"
                  :class="[
                    'payment-confirmed-btn',
                    car.payment_confirmed || 0 ? 'confirmed' : 'not-confirmed',
                    !can_confirm_payment ? 'read-only' : '',
                  ]"
                  :title="
                    can_confirm_payment
                      ? car.payment_confirmed || 0
                        ? t('carStock.payment_confirmed_yes')
                        : t('carStock.payment_confirmed_no')
                      : t('carStock.no_permission_to_confirm_payment')
                  "
                >
                  <i
                    :class="car.payment_confirmed || 0 ? 'fas fa-check-circle' : 'far fa-circle'"
                  ></i>
                  {{
                    car.payment_confirmed || 0
                      ? t('carStock.confirmed')
                      : t('carStock.not_confirmed')
                  }}
                </button>
              </td>

              <td class="documents-cell">
                <div class="document-links">
                  <div class="documents-summary">
                    <div class="document-count-badge" :title="t('carStock.documents_count', { count: getDocumentCount(car) }) || `${getDocumentCount(car)} document(s)`">
                      <i class="fas fa-folder-open"></i>
                      {{ getDocumentCount(car) }}
                    </div>
                    <div 
                      v-if="hasMissingRequiredDocuments(car)" 
                      class="missing-required-indicator" 
                      :title="t('carStock.missing_required_documents') || 'Missing required documents'"
                    >
                      <i class="fas fa-exclamation-triangle"></i>
                    </div>
                  </div>
                </div>
              </td>
              <td class="notes-cell" :title="car.notes">{{ car.notes || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Mobile Cards Container -->
    <div class="mobile-cards-container">
      <div
        v-for="car in sortedCars"
        :key="car.id"
        class="car-card"
        :class="{ 'used-car': car.is_used_car }"
      >
        <!-- Card Header -->
        <div class="card-header">
          <div class="card-id">#{{ car.id }}</div>
          <div class="card-actions">
            <div class="dropdown">
              <button @click="toggleDropdown(car.id)" class="dropdown-toggle" title="Actions">
                <i class="fas fa-ellipsis-v"></i>
              </button>
              <ul v-if="isDropdownOpen[car.id]" class="dropdown-menu">
                <li>
                  <button
                    @click="handleTaskAction(car)"
                    :disabled="isProcessing.task"
                    :class="{ processing: isProcessing.task }"
                  >
                    <i class="fas fa-tasks"></i>
                    <span>{{ t('carStock.task') }}</span>
                    <i
                      v-if="isProcessing.task"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleVINAction(car)"
                    :disabled="!can_edit_vin || isProcessing.vin"
                    :class="{ disabled: !can_edit_vin, processing: isProcessing.vin }"
                  >
                    <i class="fas fa-barcode"></i>
                    <span>{{ t('carStock.vin') }}</span>
                    <i v-if="isProcessing.vin" class="fas fa-spinner fa-spin loading-indicator"></i>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleFilesAction(car)"
                    :disabled="!can_upload_car_files || isProcessing.files"
                    :class="{ disabled: !can_upload_car_files, processing: isProcessing.files }"
                  >
                    <i class="fas fa-file-upload"></i>
                    <span>{{ t('carStock.files') }}</span>
                    <i
                      v-if="isProcessing.files"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button
                    @click="handlePortsAction(car)"
                    :disabled="!can_edit_cars_ports || isProcessing.ports"
                    :class="{ disabled: !can_edit_cars_ports, processing: isProcessing.ports }"
                  >
                    <i class="fas fa-anchor"></i>
                    <span>{{ t('carStock.ports') }}</span>
                    <i
                      v-if="isProcessing.ports"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleMoneyAction(car)"
                    :disabled="!can_edit_car_money || isProcessing.money"
                    :class="{ disabled: !can_edit_car_money, processing: isProcessing.money }"
                  >
                    <i class="fas fa-dollar-sign"></i>
                    <span>{{ t('carStock.money') }}</span>
                    <i
                      v-if="isProcessing.money"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleWarehouseAction(car)"
                    :disabled="!can_edit_warehouse || isProcessing.warehouse"
                    :class="{ disabled: !can_edit_warehouse, processing: isProcessing.warehouse }"
                  >
                    <i class="fas fa-warehouse"></i>
                    <span>{{ t('carStock.warehouse') }}</span>
                    <i
                      v-if="isProcessing.warehouse"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleDocumentsClick(car)"
                    :disabled="!can_edit_car_documents || isProcessing.documents"
                    :class="{
                      disabled: !can_edit_car_documents,
                      processing: isProcessing.documents,
                    }"
                  >
                    <i class="fas fa-file-alt"></i>
                    <span>{{ t('carStock.car_documents') }}</span>
                    <i
                      v-if="isProcessing.documents"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleLoadClick(car)"
                    :disabled="!can_load_car || isProcessing.load"
                    :class="{ disabled: !can_load_car, processing: isProcessing.load }"
                  >
                    <i class="fas fa-truck-loading"></i>
                    <span>{{ t('carStock.load') }}</span>
                    <i
                      v-if="isProcessing.load"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button @click="openSellBillTab(car)" :disabled="!car.id_sell">
                    <i class="fas fa-file-invoice-dollar"></i>
                    <span>{{ t('carStock.sell_bill') }}</span>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleNotesAction(car)"
                    :disabled="isProcessing.notes"
                    :class="{ processing: isProcessing.notes }"
                  >
                    <i class="fas fa-sticky-note"></i>
                    <span>{{ t('carStock.notes') }}</span>
                    <i
                      v-if="isProcessing.notes"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleSwitchPurchaseBill(car)"
                    :disabled="!isAdmin"
                    :title="
                      !isAdmin
                        ? t('carStock.only_administrators_can_switch_purchase_bills')
                        : t('carStock.switch_purchase_bill_with_another_car')
                    "
                  >
                    <i class="fas fa-exchange-alt"></i>
                    <span>{{ t('carStock.switch_purchase_bill') }}</span>
                  </button>
                </li>
              </ul>
            </div>

            <!-- Teleport Dropdown Button -->
          </div>
        </div>

        <!-- Date Buy Section -->
        <div class="card-section">
          <div class="section-title">Date Buy</div>
          <div class="card-value">
            {{ car.date_buy ? new Date(car.date_buy).toLocaleDateString() : '-' }}
          </div>
        </div>

        <!-- Car Details Section -->
        <div class="card-section">
          <div class="section-title">Car Details</div>
          <div class="card-details">
            <div class="card-name">{{ car.car_name }}</div>
            <div class="card-badges">
              <div v-if="car.vin" class="info-badge badge-vin">
                <i class="fas fa-barcode"></i>
                {{ car.vin }}
              </div>
              <div
                v-if="car.color"
                class="info-badge badge-color"
                :style="{
                  backgroundColor: car.hexa || '#000000',
                  color: getTextColor(car.hexa || '#000000'),
                }"
              >
                <i class="fas fa-palette"></i>
                {{ car.color }}
              </div>
              <div v-if="car.export_lisence_ref" class="info-badge badge-export-license">
                <i class="fas fa-certificate"></i>
                {{ car.export_lisence_ref }}
              </div>
              <div v-else class="info-badge badge-export-license-empty">
                <i class="fas fa-certificate"></i>
                -
              </div>
              <div v-if="car.buy_bill_ref" class="info-badge badge-buy-bill">
                <i class="fas fa-shopping-cart"></i>
                {{ t('carStock.buy') }}: {{ car.buy_bill_ref }}
              </div>
              <div v-if="car.sell_bill_ref" class="info-badge badge-sell-bill">
                <i class="fas fa-file-invoice-dollar"></i>
                {{ t('carStock.sell') }}: {{ car.sell_bill_ref }}
              </div>
            </div>
          </div>
        </div>

        <!-- Client Section -->
        <div class="card-section">
          <div class="section-title">Client</div>
          <div class="card-details">
            <div class="card-value">{{ car.client_name || '-' }}</div>
            <div class="card-badges">
              <div v-if="car.client_id_no" class="info-badge badge-client-id">
                <i class="fas fa-id-card"></i>
                {{ car.client_id_no }}
              </div>
              <div v-if="car.is_batch" class="info-badge badge-wholesale">
                <i class="fas fa-layer-group"></i>
                {{ t('carStock.whole_sale') }}
              </div>
            </div>
          </div>
        </div>

        <!-- Ports Section -->
        <div class="card-section">
          <div class="section-title">Ports</div>
          <div class="card-ports">
            <div class="card-port-item">
              <div class="card-port-label">Loading:</div>
              <div v-if="car.loading_port" class="info-badge badge-loading-port">
                <i class="fas fa-ship"></i>
                {{ car.loading_port }}
              </div>
              <div v-else class="port-empty">-</div>
            </div>
            <div class="card-port-item">
              <div class="card-port-label">Discharge:</div>
              <div v-if="car.discharge_port" class="info-badge badge-discharge-port">
                <i class="fas fa-anchor"></i>
                {{ car.discharge_port }}
              </div>
              <div v-else class="port-empty">-</div>
            </div>
            <div v-if="car.container_ref" class="info-badge badge-container">
              <i class="fas fa-box"></i>
              {{ car.container_ref }}
            </div>
            <div v-if="car.date_pay_freight" class="info-badge badge-freight-paid">
              <i class="fas fa-check-circle"></i>
              {{ t('carStock.freight_paid') }}
            </div>
          </div>
        </div>

        <!-- Financial Section -->
        <div class="card-section">
          <div class="section-title">Financial</div>
          <div class="card-row">
            <div class="card-label">Freight:</div>
            <div class="card-value">${{ getFreightValue(car) }}</div>
          </div>
          <div class="card-row">
            <div class="card-label">FOB:</div>
            <div class="card-value">${{ car.price_cell || '0' }}</div>
          </div>
          <div class="card-cfr">
            <div class="info-badge badge-cfr-usd">
              <i class="fas fa-dollar-sign"></i>
              USD: ${{ getCfrUsdValue(car) }}
            </div>
            <div class="info-badge badge-cfr-dza">
              <i class="fas fa-coins"></i>
              DZA: {{ getCfrDzaValue(car) }}
            </div>
            <div class="info-badge badge-cfr-rate">
              <i class="fas fa-percentage"></i>
              Rate: {{ getRateValue(car) }}
            </div>
          </div>
        </div>

        <!-- Status Section -->
        <div class="card-section">
          <div class="section-title">Status</div>
          <div class="card-status">
            <div class="card-status-item">
              <span
                :class="{
                  'status-sold': car.id_sell,
                  'status-available': !car.id_sell,
                  'status-used': car.is_used_car,
                }"
              >
                <i
                  :class="[
                    car.id_sell ? 'fas fa-check-circle' : 'fas fa-clock',
                    car.is_used_car ? 'fas fa-history' : '',
                  ]"
                ></i>
                {{ car.id_sell ? t('carStock.sold') : t('carStock.available') }}
                {{ car.is_used_car ? `(${t('carStock.used')})` : '' }}
              </span>
            </div>
            <div v-if="car.in_wharhouse_date" class="info-badge badge-in-warehouse">
              <i class="fas fa-warehouse"></i>
              {{ t('carStock.in_warehouse') }}
            </div>
            <div v-if="car.date_get_bl" class="info-badge badge-bl-received">
              <i class="fas fa-file-contract"></i>
              {{ t('carStock.bl_received') }}
            </div>
          </div>
        </div>

        <!-- Documents Section -->
        <div class="card-section">
          <div class="section-title">Documents</div>
          <div class="card-documents">
            <a
              v-if="car.path_documents"
              :href="getFileUrl(car.path_documents)"
              target="_blank"
              class="card-document-link"
            >
              <i class="fas fa-file-pdf"></i>
              BL
            </a>
            <a
              v-if="car.sell_pi_path"
              :href="getFileUrl(car.sell_pi_path)"
              target="_blank"
              class="card-document-link"
            >
              <i class="fas fa-file-invoice-dollar"></i>
              {{ t('carStock.invoice') }}
            </a>
            <a
              v-if="car.buy_pi_path"
              :href="getFileUrl(car.buy_pi_path)"
              target="_blank"
              class="card-document-link"
            >
              <i class="fas fa-file-contract"></i>
              {{ t('carStock.packing_list') }}
            </a>
            <a
              v-if="car.path_coo"
              :href="getFileUrl(car.path_coo)"
              target="_blank"
              class="card-document-link"
            >
              <i class="fas fa-certificate"></i>
              COO
            </a>
            <a
              v-if="car.path_coc"
              :href="getFileUrl(car.path_coc)"
              target="_blank"
              class="card-document-link"
            >
              <i class="fas fa-award"></i>
              COC
            </a>
          </div>
        </div>

        <!-- Notes Section -->
        <div class="card-section">
          <div class="section-title">Notes</div>
          <div class="card-notes">{{ car.notes || '-' }}</div>
        </div>
      </div>
    </div>
  </div>

  <!-- Combine Loading Overlay -->
  <teleport to="body">
    <div v-show="isProcessing.combine" class="combine-loading-overlay">
      <div class="combine-loading-content">
        <i class="fas fa-spinner fa-spin"></i>
        <span>{{ combiningCarsText }}</span>
      </div>
    </div>
  </teleport>

  <!-- Teleport Dropdown Menu -->
  <teleport to="body">
    <div
      v-if="teleportDropdown.isOpen"
      class="teleport-dropdown"
      :style="{
        position: 'absolute',
        left: teleportDropdown.position.x + 'px',
        top: teleportDropdown.position.y + 'px',
        zIndex: 9999,
      }"
    >
      <ul class="teleport-dropdown-menu">
        <li>
          <button
            @click="handleTaskAction(getCarById(teleportDropdown.carId))"
            :disabled="isProcessing.task"
            :class="{ processing: isProcessing.task }"
          >
            <i class="fas fa-tasks"></i>
            <span>Task</span>
            <i v-if="isProcessing.task" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>
        <li>
          <button
            @click="handleVINAction(getCarById(teleportDropdown.carId))"
            :disabled="!can_edit_vin || isProcessing.vin"
            :class="{ disabled: !can_edit_vin, processing: isProcessing.vin }"
          >
            <i class="fas fa-barcode"></i>
            <span>VIN</span>
            <i v-if="isProcessing.vin" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>
        <li>
          <button
            @click="handleFilesAction(getCarById(teleportDropdown.carId))"
            :disabled="!can_upload_car_files || isProcessing.files"
            :class="{ disabled: !can_upload_car_files, processing: isProcessing.files }"
          >
            <i class="fas fa-file-upload"></i>
            <span>Files</span>
            <i v-if="isProcessing.files" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>
        <li>
          <button
            @click="handlePortsAction(getCarById(teleportDropdown.carId))"
            :disabled="!can_edit_cars_ports || isProcessing.ports"
            :class="{ disabled: !can_edit_cars_ports, processing: isProcessing.ports }"
          >
            <i class="fas fa-anchor"></i>
            <span>Ports</span>
            <i v-if="isProcessing.ports" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>
        <li>
          <button
            @click="handleMoneyAction(getCarById(teleportDropdown.carId))"
            :disabled="!can_edit_car_money || isProcessing.money"
            :class="{ disabled: !can_edit_car_money, processing: isProcessing.money }"
          >
            <i class="fas fa-dollar-sign"></i>
            <span>Money</span>
            <i v-if="isProcessing.money" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>
        <li>
          <button
            @click="handleWarehouseAction(getCarById(teleportDropdown.carId))"
            :disabled="!can_edit_warehouse || isProcessing.warehouse"
            :class="{ disabled: !can_edit_warehouse, processing: isProcessing.warehouse }"
          >
            <i class="fas fa-warehouse"></i>
            <span>Warehouse</span>
            <i v-if="isProcessing.warehouse" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>
        <li>
          <button
            @click="handleDocumentsClick(getCarById(teleportDropdown.carId))"
            :disabled="!can_edit_car_documents || isProcessing.documents"
            :class="{
              disabled: !can_edit_car_documents,
              processing: isProcessing.documents,
            }"
          >
            <i class="fas fa-file-alt"></i>
            <span>Car Documents</span>
            <i v-if="isProcessing.documents" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>
        <li>
          <button
            @click="handleLoadClick(getCarById(teleportDropdown.carId))"
            :disabled="!can_load_car || isProcessing.load"
            :class="{ disabled: !can_load_car, processing: isProcessing.load }"
          >
            <i class="fas fa-truck-loading"></i>
            <span>Load</span>
            <i v-if="isProcessing.load" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>

        <li>
          <button
            @click="openSellBillTab(getCarById(teleportDropdown.carId))"
            :disabled="!getCarById(teleportDropdown.carId)?.id_sell"
          >
            <i class="fas fa-file-invoice-dollar"></i>
            <span>Sell Bill</span>
          </button>
        </li>
        <li>
          <button
            @click="handleNotesAction(getCarById(teleportDropdown.carId))"
            :disabled="isProcessing.notes"
            :class="{ processing: isProcessing.notes }"
          >
            <i class="fas fa-sticky-note"></i>
            <span>Notes</span>
            <i v-if="isProcessing.notes" class="fas fa-spinner fa-spin loading-indicator"></i>
          </button>
        </li>
        <li>
          <button
            @click="handleSwitchPurchaseBill(getCarById(teleportDropdown.carId))"
            :disabled="!isAdmin"
            :title="
              !isAdmin
                ? t('carStock.only_administrators_can_switch_purchase_bills')
                : t('carStock.switch_purchase_bill_with_another_car')
            "
          >
            <i class="fas fa-exchange-alt"></i>
            <span>{{ t('carStock.switch_purchase_bill') }}</span>
          </button>
        </li>
      </ul>
    </div>
  </teleport>

  <!-- Notes Edit Modal -->
  <div v-if="showNotesModal" class="modal-overlay">
    <div class="modal-content">
      <h3>Edit Notes</h3>
      <textarea
        v-model="notesEditValue"
        rows="6"
        class="notes-textarea"
        placeholder="Enter notes..."
      ></textarea>
      <div class="modal-actions">
        <button @click="saveNotes" :disabled="isProcessing.notes" class="btn save-btn">
          <i class="fas fa-save"></i> Save
        </button>
        <button @click="cancelNotes" :disabled="isProcessing.notes" class="btn cancel-btn">
          <i class="fas fa-times"></i> Cancel
        </button>
      </div>
    </div>
  </div>

  <!-- VIN Assignment Modal -->
  <VinAssignmentModal
    :is-visible="showVinAssignmentModal"
    :selected-cars="sortedCars.filter((car) => selectedCars.has(car.id))"
    @close="handleVinAssignmentModalClose"
    @vins-assigned="handleVinsAssigned"
  />

  <!-- Ports Bulk Edit Modal -->
  <CarPortsBulkEditForm
    :show="showPortsBulkEditForm"
    :selected-cars="sortedCars.filter((car) => selectedCars.has(car.id))"
    :is-admin="isAdmin"
    @close="showPortsBulkEditForm = false"
    @save="handlePortsBulkSave"
  />

  <!-- Notes Bulk Edit Modal -->
  <CarNotesBulkEditForm
    :show="showNotesBulkEditForm"
    :selected-cars="sortedCars.filter((car) => selectedCars.has(car.id))"
    :is-admin="isAdmin"
    @close="showNotesBulkEditForm = false"
    @save="handleNotesBulkSave"
  />

  <!-- Color Bulk Edit Modal -->
  <CarColorBulkEditForm
    :show="showColorBulkEditForm"
    :selected-cars="sortedCars.filter((car) => selectedCars.has(car.id))"
    :is-admin="isAdmin"
    @close="showColorBulkEditForm = false"
    @save="handleColorBulkSave"
  />

  <CarExportLicenseBulkEditForm
    :show="showExportLicenseBulkEditForm"
    :selected-cars="sortedCars.filter((car) => selectedCars.has(car.id))"
    :is-admin="isAdmin"
    @close="showExportLicenseBulkEditForm = false"
    @save="handleExportLicenseBulkSave"
  />

  <!-- Selections Components -->
  <SaveSelectionForm
    :show="showSaveSelectionForm"
    :selected-car-ids="Array.from(selectedCars)"
    @close="showSaveSelectionForm = false"
    @saved="handleSelectionSaved"
  />

  <SendSelectionForm
    :show="showSendSelectionForm"
    :selected-car-ids="Array.from(selectedCars)"
    @close="showSendSelectionForm = false"
    @sent="handleSelectionSent"
  />

  <ShowSelectionsModal
    :show="showSelectionsModal"
    @close="showSelectionsModal = false"
    @selection-loaded="handleSelectionLoaded"
  />

  <!-- Batch Transfer Modal -->
  <div v-if="showBatchTransferModal" class="modal-overlay" @click="closeBatchTransferModal">
    <div class="modal-content batch-transfer-modal" @click.stop>
      <div class="modal-header">
        <h3>{{ t('carStock.batch_transfer_documents') || 'Batch Transfer Documents' }}</h3>
        <button class="close-btn" @click="closeBatchTransferModal">&times;</button>
      </div>

      <div class="modal-body">
        <div class="info-section">
          <i class="fas fa-info-circle"></i>
          <span>
            {{
              t('carStock.batch_transfer_info', {
                count: batchTransferFiles.length,
                cars: selectedCars.size,
              }) ||
              `Transferring ${batchTransferFiles.length} file(s) from ${selectedCars.size} selected car(s)`
            }}
          </span>
        </div>

        <div class="form-group">
          <label>{{ t('carStock.transfer_to_user') || 'Transfer To User' }} *</label>
          <select
            v-model="batchTransferToUserId"
            class="form-select"
            :disabled="isBatchTransferring"
          >
            <option value="">{{ t('carStock.select_user') || 'Select a user...' }}</option>
            <option
              v-for="userOption in batchTransferAvailableUsers"
              :key="userOption.id"
              :value="userOption.id"
            >
              {{ userOption.name || userOption.username }}
            </option>
          </select>
        </div>

        <div class="form-group">
          <label
            >{{ t('carStock.notes') || 'Notes' }} ({{
              t('carStock.optional') || 'Optional'
            }})</label
          >
          <textarea
            v-model="batchTransferNotes"
            class="form-textarea"
            rows="3"
            :disabled="isBatchTransferring"
            :placeholder="
              t('carStock.transfer_notes_placeholder') || 'Add any notes about this transfer...'
            "
          ></textarea>
        </div>

        <div class="form-group">
          <label
            >{{ t('carStock.expected_return_date') || 'Expected Return Date' }} ({{
              t('carStock.optional') || 'Optional'
            }})</label
          >
          <input
            v-model="batchTransferExpectedReturn"
            type="date"
            class="form-input"
            :disabled="isBatchTransferring"
          />
        </div>

        <div v-if="batchTransferFiles.length > 0" class="files-preview">
          <h4>
            {{ t('carStock.files_to_transfer') || 'Files to Transfer' }} ({{
              batchTransferFiles.length
            }})
          </h4>
          <div class="files-list">
            <div v-for="(file, index) in batchTransferFiles" :key="index" class="file-item">
              <i class="fas fa-file"></i>
              <span>{{ file.category_name || 'Unknown Category' }}</span>
              <span class="car-name">({{ file.car_name }})</span>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button @click="closeBatchTransferModal" class="cancel-btn" :disabled="isBatchTransferring">
          {{ t('carStock.cancel') || 'Cancel' }}
        </button>
        <button
          @click="confirmBatchTransfer"
          class="confirm-btn"
          :disabled="isBatchTransferring || !batchTransferToUserId"
        >
          <i v-if="isBatchTransferring" class="fas fa-spinner fa-spin"></i>
          {{ t('carStock.confirm_transfer') || 'Confirm Transfer' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Batch Checkout Modal -->
  <div v-if="showBatchCheckoutModal" class="modal-overlay" @click="closeBatchCheckoutModal">
    <div class="modal-content batch-checkout-modal" @click.stop>
      <div class="modal-header">
        <h3>{{ t('carStock.batch_checkout_documents') || 'Batch Checkout Documents' }}</h3>
        <button class="close-btn" @click="closeBatchCheckoutModal">&times;</button>
      </div>

      <div class="modal-body">
        <div class="info-section">
          <i class="fas fa-info-circle"></i>
          <span>
            {{
              t('carStock.batch_checkout_info', {
                count: batchCheckoutFiles.length,
                cars: selectedCars.size,
              }) ||
              `Checking out ${batchCheckoutFiles.length} file(s) from ${selectedCars.size} selected car(s)`
            }}
          </span>
        </div>

        <div class="form-group">
          <label>{{ t('carStock.checkout_to') || 'Check Out To' }} *</label>
          <select v-model="batchCheckoutType" class="form-select" :disabled="isBatchCheckouting">
            <option value="user">{{ t('carStock.user') || 'User' }}</option>
            <option value="client">{{ t('carStock.client') || 'Client' }}</option>
            <option value="custom_clearance_agent">
              {{ t('carStock.custom_clearance_agent') || 'Custom Clearance Agent' }}
            </option>
          </select>
        </div>

        <!-- User Selection -->
        <div v-if="batchCheckoutType === 'user'" class="form-group">
          <label>{{ t('carStock.select_user') || 'Select User' }} *</label>
          <select v-model="batchCheckoutUserId" class="form-select" :disabled="isBatchCheckouting">
            <option value="">{{ t('carStock.select_user') || 'Select a user...' }}</option>
            <option
              v-for="userOption in batchCheckoutAvailableUsers"
              :key="userOption.id"
              :value="userOption.id"
            >
              {{ userOption.name || userOption.username }}
            </option>
          </select>
        </div>

        <!-- Agent Selection -->
        <div v-if="batchCheckoutType === 'custom_clearance_agent'" class="form-group">
          <label>{{ t('carStock.select_agent') || 'Select Custom Clearance Agent' }} *</label>
          <select v-model="batchCheckoutAgentId" class="form-select" :disabled="isBatchCheckouting">
            <option value="">{{ t('carStock.select_agent') || 'Select an agent...' }}</option>
            <option v-for="agent in batchCheckoutAvailableAgents" :key="agent.id" :value="agent.id">
              {{ agent.name }}
            </option>
          </select>
        </div>

        <!-- Client Selection -->
        <div v-if="batchCheckoutType === 'client'" class="form-group">
          <label>{{ t('carStock.select_client') || 'Select Client' }} *</label>
          <select
            v-model="batchCheckoutClientId"
            class="form-select"
            :disabled="isBatchCheckouting"
            @change="handleClientChange"
          >
            <option value="">{{ t('carStock.select_client') || 'Select a client...' }}</option>
            <option
              v-for="client in batchCheckoutAvailableClients"
              :key="client.id"
              :value="client.id"
            >
              {{ client.name }}
            </option>
          </select>
        </div>

        <!-- Client Name (for new clients) -->
        <div v-if="batchCheckoutType === 'client' && !batchCheckoutClientId" class="form-group">
          <label>{{ t('carStock.client_name') || 'Client Name' }} *</label>
          <input
            v-model="batchCheckoutClientName"
            type="text"
            class="form-input"
            :disabled="isBatchCheckouting"
            :placeholder="t('carStock.enter_client_name') || 'Enter client name...'"
          />
        </div>

        <!-- Client Contact (for new clients) -->
        <div v-if="batchCheckoutType === 'client' && !batchCheckoutClientId" class="form-group">
          <label
            >{{ t('carStock.client_contact') || 'Client Contact' }} ({{
              t('carStock.optional') || 'Optional'
            }})</label
          >
          <input
            v-model="batchCheckoutClientContact"
            type="text"
            class="form-input"
            :disabled="isBatchCheckouting"
            :placeholder="t('carStock.enter_client_contact') || 'Enter client contact...'"
          />
        </div>

        <div class="form-group">
          <label
            >{{ t('carStock.notes') || 'Notes' }} ({{
              t('carStock.optional') || 'Optional'
            }})</label
          >
          <textarea
            v-model="batchCheckoutNotes"
            class="form-textarea"
            rows="3"
            :disabled="isBatchCheckouting"
            :placeholder="
              t('carStock.checkout_notes_placeholder') || 'Add any notes about this checkout...'
            "
          ></textarea>
        </div>

        <div v-if="batchCheckoutFiles.length > 0" class="files-preview">
          <h4>
            {{ t('carStock.files_to_checkout') || 'Files to Checkout' }} ({{
              batchCheckoutFiles.length
            }})
          </h4>
          <div class="files-list">
            <div v-for="(file, index) in batchCheckoutFiles" :key="index" class="file-item">
              <i class="fas fa-file"></i>
              <span>{{ file.file_name }}</span>
              <span class="car-name">({{ file.car_name }})</span>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button @click="closeBatchCheckoutModal" class="cancel-btn" :disabled="isBatchCheckouting">
          {{ t('carStock.cancel') || 'Cancel' }}
        </button>
        <button
          @click="confirmBatchCheckout"
          class="confirm-btn"
          :disabled="
            isBatchCheckouting ||
            (batchCheckoutType === 'user' && !batchCheckoutUserId) ||
            (batchCheckoutType === 'client' &&
              !batchCheckoutClientId &&
              !batchCheckoutClientName) ||
            (batchCheckoutType === 'custom_clearance_agent' && !batchCheckoutAgentId)
          "
        >
          <i v-if="isBatchCheckouting" class="fas fa-spinner fa-spin"></i>
          {{ t('carStock.confirm_checkout') || 'Confirm Checkout' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cars-stock-table {
  width: 100%;
  overflow-x: auto;
  position: relative;
  overflow: visible;
}

.loading,
.error,
.empty-state {
  padding: 40px;
  text-align: center;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  background-color: #f9fafb;
  border-radius: 8px;
  margin: 20px 0;
}

.loading i {
  font-size: 2em;
  color: #3b82f6;
}

.error {
  color: #ef4444;
}

.error i {
  font-size: 2em;
}

.empty-state i {
  color: #9ca3af;
}

.empty-state p {
  margin: 0;
  font-size: 1.1em;
}

/* Table styles */
.table-container {
  position: relative;
  max-height: calc(100vh - 250px);
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow-x: auto;
}

.cars-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  position: relative;
  table-layout: fixed;
}

.cars-table thead {
  position: sticky !important;
  top: 0 !important;
  z-index: 100 !important;
  background-color: #f8f9fa !important;
  border-bottom: 2px solid #e5e7eb;
}

.cars-table thead th {
  position: sticky !important;
  top: 0 !important;
  background-color: #f8f9fa !important;
  z-index: 100 !important;
  border-bottom: 2px solid #e5e7eb;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.cars-table th {
  padding: 12px;
  text-align: left;
  border-bottom: 2px solid #e5e7eb;
  font-weight: 600;
  white-space: nowrap;
  color: #374151;
}

.cars-table td {
  padding: 12px;
  border-bottom: 1px solid #e5e7eb;
  background-color: white;
}

.cars-table tbody tr:hover {
  background-color: #f9fafb;
}

/* ID column width */
.cars-table th:nth-child(3),
.cars-table td:nth-child(3) {
  width: 4ch;
  min-width: 4ch;
  max-width: 4ch;
  text-align: center;
}

/* Select column width - make it compact */
.cars-table th:nth-child(1),
.cars-table td:nth-child(1) {
  width: 2ch;
  min-width: 2ch;
  max-width: 2ch;
  text-align: center;
  padding: 8px 4px;
}

/* Car details column - make it compact */
.cars-table th:nth-child(6),
.cars-table td:nth-child(6) {
  width: 20ch;
  min-width: 20ch;
  max-width: 20ch;
}

/* Freight column - make it compact */
.cars-table th:nth-child(8),
.cars-table td:nth-child(8) {
  width: 8ch;
  min-width: 8ch;
  max-width: 8ch;
  text-align: center;
}

/* Date columns - make them compact */
.cars-table th:nth-child(4),
.cars-table td:nth-child(4),
.cars-table th:nth-child(5),
.cars-table td:nth-child(5) {
  width: 10ch;
  min-width: 10ch;
  max-width: 10ch;
  text-align: center;
}

/* Actions column - make it compact */
.cars-table th:nth-child(2),
.cars-table td:nth-child(2) {
  width: 3ch;
  min-width: 3ch;
  max-width: 3ch;
  text-align: center;
  padding: 8px 4px;
}

/* Status styles */
.status-available,
.status-sold {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.9em;
  font-weight: 500;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}

.status-available:hover,
.status-sold:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.status-available {
  color: #10b981;
  background-color: #d1fae5;
  border: 1px solid #a7f3d0;
}

.status-sold {
  color: #ef4444;
  background-color: #fee2e2;
  border: 1px solid #fca5a5;
}

/* Document styles */
.documents-cell {
  min-width: 120px;
}

.document-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.documents-summary {
  display: flex;
  align-items: center;
  gap: 8px;
}

.document-count-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background-color: #10b981;
  color: white;
  font-size: 0.75em;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 12px;
}

.document-count-badge i {
  color: white;
  font-size: 0.85em;
}

.missing-required-indicator {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background-color: #fef2f2;
  color: #dc2626;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  font-size: 0.9em;
  box-shadow: 0 2px 4px rgba(220, 38, 38, 0.2);
  animation: pulse-warning 2s infinite;
  cursor: help;
}

@keyframes pulse-warning {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.05);
  }
}

.missing-required-indicator:hover {
  background-color: #fee2e2;
  box-shadow: 0 2px 6px rgba(220, 38, 38, 0.3);
}

.file-version {
  display: inline-block;
  margin-left: 4px;
  font-size: 0.7em;
  color: #6b7280;
  font-weight: normal;
}

.no-documents {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #9ca3af;
  font-size: 0.85em;
  padding: 4px 8px;
  font-style: italic;
}

.no-documents i {
  color: #d1d5db;
}

/* Button styles */
.edit-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 0.9em;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.edit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.edit-btn.processing {
  position: relative;
  pointer-events: none;
}

.edit-btn.processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

/* Dropdown styles */
.dropdown {
  position: relative;
  display: inline-block;
}

.dropdown-toggle {
  background-color: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 6px 12px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.dropdown-toggle:hover {
  background-color: #e5e7eb;
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background: white;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  min-width: 160px;
  z-index: 1000;
  list-style: none;
  margin: 0;
  padding: 8px 0;
}

.dropdown-menu li {
  margin: 2px 0;
}

.dropdown-menu li button {
  background: none;
  border: none;
  width: 100%;
  text-align: left;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #374151;
  transition: all 0.2s ease;
}

.dropdown-menu li button:not(.disabled):hover {
  background-color: #f3f4f6;
}

.dropdown-menu li button.disabled {
  opacity: 0.6;
  cursor: not-allowed;
  color: #6b7280;
}

.dropdown-menu li button i:not(.loading-indicator) {
  width: 16px;
  text-align: center;
  font-size: 1em;
}

.dropdown-menu li button span {
  flex: 1;
}

.dropdown-menu li button.processing {
  position: relative;
  pointer-events: none;
}

.dropdown-menu li button.processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

/* Loading indicator */
.loading-indicator {
  font-size: 0.9em;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fa-spin {
  animation: spin 1s linear infinite;
}

/* Advanced filter button */
.table-actions {
  margin-bottom: 16px;
}

.advanced-filter-btn {
  padding: 8px 16px;
  background-color: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.advanced-filter-btn:hover {
  background-color: #e5e7eb;
}

.advanced-filter-btn i {
  font-size: 1.1em;
}

/* Responsive adjustments */
@media (max-width: 1024px) {
  .table-container {
    max-height: none;
    overflow-x: auto;
  }

  .cars-table {
    min-width: 1200px;
  }
}

/* Used car styling */
.used-car {
  background-color: #fef3c7;
}

.used-car:hover {
  background-color: #fde68a;
}

/* Status styling */
.status-used {
  background-color: #fef3c7 !important;
  color: #92400e !important;
  border-color: #f59e0b !important;
}

.status-used i {
  color: #d97706;
}

/* Notes cell styling */
.notes-cell {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  position: relative;
}

.notes-cell:hover {
  white-space: normal;
  overflow: visible;
  background-color: #f8fafc;
  z-index: 1;
  position: relative;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border-radius: 4px;
  padding: 8px;
  margin: -8px;
}

.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 20px;
}

.sortable:hover {
  background-color: #f0f0f0;
}

.sort-indicator {
  position: absolute;
  right: 5px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 0.8em;
  color: #0d6efd;
}

.client-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.client-id {
  font-size: 0.85em;
  color: #666;
}

.car-name {
  font-weight: 600;
}

.car-vin {
  font-size: 0.85em;
  color: #666;
}

.car-color {
  font-size: 0.85em;
  color: #666;
}

.car-buy-bill {
  font-size: 0.85em;
  color: #059669;
  font-weight: 500;
}

.car-sell-bill {
  font-size: 0.85em;
  color: #dc2626;
  font-weight: 500;
}

/* Loading indicator styles */
.loading-indicator {
  margin-left: 8px;
  color: #409eff;
}

/* Teleport Dropdown Styles */
.teleport-dropdown-toggle {
  background: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 6px 10px;
  cursor: pointer;
  margin-left: 8px;
  transition: all 0.3s;
}

.teleport-dropdown-toggle:hover {
  background: #ecf5ff;
  border-color: #409eff;
  color: #409eff;
}

.teleport-dropdown {
  background: white;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  min-width: 160px;
  z-index: 9999;
}

.teleport-dropdown-menu {
  list-style: none;
  margin: 0;
  padding: 8px 0;
}

.teleport-dropdown-menu li {
  margin: 0;
}

.teleport-dropdown-menu button {
  width: 100%;
  text-align: left;
  padding: 8px 16px;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
  transition: background-color 0.2s;
}

.teleport-dropdown-menu button:hover:not(:disabled) {
  background-color: #f5f7fa;
  color: #409eff;
}

.teleport-dropdown-menu button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.teleport-dropdown-menu button.processing {
  opacity: 0.7;
  cursor: not-allowed;
}

.teleport-dropdown-menu button.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.teleport-dropdown-menu .loading-indicator {
  margin-left: auto;
}

/* Info Badge Styles */
.info-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.8em;
  font-weight: 500;
  margin: 2px 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}

.info-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.badge-vin {
  background-color: #dbeafe;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}

.badge-color {
  border: 1px solid #d1d5db;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  width: 100%;
  text-align: center;
  justify-content: center;
}

.badge-buy-bill {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.badge-sell-bill {
  background-color: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.badge-container {
  background-color: #f3e8ff;
  color: #7c3aed;
  border: 1px solid #e9d5ff;
  width: 100%;
  text-align: center;
  justify-content: center;
}

.badge-client-id {
  background-color: #f0fdf4;
  color: #15803d;
  border: 1px solid #bbf7d0;
}

.badge-export-license {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: #3b82f6;
  color: #fff;
  border-radius: 8px;
  padding: 2px 10px;
  font-size: 13px;
  font-weight: 500;
}

.badge-export-license-empty {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: #e5e7eb;
  color: #6b7280;
  border-radius: 8px;
  padding: 2px 10px;
  font-size: 13px;
  font-weight: 500;
}

.badge-freight-paid {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.badge-in-warehouse {
  background-color: #dbeafe;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}

.badge-bl-received {
  background-color: #f3e8ff;
  color: #7c3aed;
  border: 1px solid #e9d5ff;
}

.badge-warehouse {
  background-color: #f0fdf4;
  color: #15803d;
  border: 1px solid #bbf7d0;
}

.badge-loading-port {
  background-color: #fef3c7;
  color: #92400e;
  border: 1px solid #fde68a;
}

.badge-discharge-port {
  background-color: #f0f9ff;
  color: #0c4a6e;
  border: 1px solid #bae6fd;
}

.badge-wholesale {
  background-color: #dbeafe;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}

/* Ports Container Styles */
.ports-container {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ports-column {
  width: 13ch !important;
  min-width: 13ch !important;
  max-width: 13ch !important;
}

.freight-column {
  width: 8ch !important;
  min-width: 8ch !important;
  max-width: 8ch !important;
}

.port-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.port-label {
  font-size: 0.8em;
  font-weight: 600;
  color: #6b7280;
  min-width: 70px;
}

.port-empty {
  font-size: 0.85em;
  color: #9ca3af;
  font-style: italic;
}

/* Car Details Styles */
.car-details-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.car-detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-label {
  font-size: 0.8em;
  font-weight: 600;
  color: #6b7280;
  min-width: 60px;
}

/* Status Container Styles */
.status-container {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-label {
  font-size: 0.8em;
  font-weight: 600;
  color: #6b7280;
  min-width: 70px;
}

/* CFR Container Styles */
.cfr-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.badge-cfr-usd {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
  width: fit-content;
  display: inline-block;
}

.badge-cfr-dza {
  background-color: #fef3c7;
  color: #92400e;
  border: 1px solid #fde68a;
  width: fit-content;
  display: inline-block;
}

.badge-cfr-rate {
  background-color: #dbeafe;
  color: #1e40af;
  border: 1px solid #bfdbfe;
  width: fit-content;
  display: inline-block;
}

/* Mobile Responsive Styles */
@media (max-width: 768px) {
  .table-container {
    display: none;
  }

  .mobile-cards-container {
    display: block;
    padding: 16px;
  }

  .car-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    margin-bottom: 16px;
    padding: 16px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
  }

  .car-card:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
    transform: translateY(-2px);
  }

  .car-card.used-car {
    border-left: 4px solid #f59e0b;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #f3f4f6;
  }

  .card-id {
    font-size: 1.1em;
    font-weight: 700;
    color: #1f2937;
  }

  .card-actions {
    display: flex;
    gap: 8px;
  }

  .card-section {
    margin-bottom: 16px;
  }

  .card-section:last-child {
    margin-bottom: 0;
  }

  .section-title {
    font-size: 0.9em;
    font-weight: 600;
    color: #6b7280;
    margin-bottom: 8px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .card-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0;
    border-bottom: 1px solid #f9fafb;
  }

  .card-row:last-child {
    border-bottom: none;
  }

  .card-label {
    font-size: 0.85em;
    font-weight: 500;
    color: #6b7280;
    min-width: 80px;
  }

  .card-value {
    font-size: 0.9em;
    color: #1f2937;
    text-align: right;
    flex: 1;
  }

  .card-badges {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 4px;
  }

  .card-details {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .card-detail-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-detail-label {
    font-size: 0.8em;
    font-weight: 600;
    color: #6b7280;
    min-width: 60px;
  }

  .card-ports {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .card-port-item {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .card-port-label {
    font-size: 0.8em;
    font-weight: 600;
    color: #6b7280;
    min-width: 70px;
  }

  .card-status {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .card-status-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-status-label {
    font-size: 0.8em;
    font-weight: 600;
    color: #6b7280;
    min-width: 70px;
  }

  .card-cfr {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .card-documents {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .card-document-link {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background-color: #f3f4f6;
    color: #374151;
    text-decoration: none;
    border-radius: 8px;
    font-size: 0.85em;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  .card-document-link:hover {
    background-color: #e5e7eb;
    color: #1f2937;
  }

  .card-notes {
    font-size: 0.9em;
    color: #6b7280;
    font-style: italic;
    line-height: 1.4;
  }

  /* Mobile dropdown adjustments */
  .mobile-cards-container .dropdown {
    position: relative;
  }

  .mobile-cards-container .dropdown-menu {
    position: absolute;
    top: 100%;
    right: 0;
    z-index: 1000;
    min-width: 160px;
    background: white;
    border: 1px solid #dcdfe6;
    border-radius: 6px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    list-style: none;
    margin: 0;
    padding: 8px 0;
  }

  .mobile-cards-container .teleport-dropdown-toggle {
    padding: 8px 12px;
    font-size: 0.9em;
  }

  .dropdown-toggle {
    background: #f5f7fa;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    padding: 8px 12px;
    cursor: pointer;
    transition: all 0.3s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .dropdown-toggle:hover {
    background: #ecf5ff;
    border-color: #409eff;
    color: #409eff;
  }

  .mobile-cards-container .dropdown-toggle {
    padding: 10px 14px;
    font-size: 1em;
    background: #ffffff;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .mobile-cards-container .dropdown-toggle:hover {
    background: #f9fafb;
    border-color: #d1d5db;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
  }

  /* Ensure dropdowns appear above everything */
  .mobile-cards-container {
    position: relative;
    z-index: 1;
  }

  .mobile-cards-container .dropdown {
    z-index: 1001;
  }
}

/* Hide mobile cards on desktop */
@media (min-width: 769px) {
  .mobile-cards-container {
    display: none;
  }
}

/* Disabled button styles */
.dropdown-menu button:disabled,
.teleport-dropdown-menu button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background-color: #f3f4f6 !important;
  color: #9ca3af !important;
  border-color: #e5e7eb !important;
}

.dropdown-menu button:disabled:hover,
.teleport-dropdown-menu button:disabled:hover {
  background-color: #f3f4f6 !important;
  color: #9ca3af !important;
  transform: none !important;
  box-shadow: none !important;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}
.modal-content {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  min-width: 320px;
  max-width: 90vw;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.18);
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.notes-textarea {
  width: 100%;
  min-height: 120px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  padding: 8px;
  font-size: 1em;
  resize: vertical;
}
.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}
.save-btn {
  background: #10b981;
  color: #fff;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}
.cancel-btn {
  background: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}
.save-btn:disabled,
.cancel-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* Selection checkbox styles */
.select-all-header,
.select-cell {
  width: 40px;
  text-align: center;
  vertical-align: middle;
}

.select-all-header input[type='checkbox'],
.select-cell input[type='checkbox'] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
}

.select-all-header input[type='checkbox']:disabled,
.select-cell input[type='checkbox']:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.select-all-header input[type='checkbox']:indeterminate {
  accent-color: #f59e0b;
}

/* Highlight selected rows */
.cars-table tbody tr.selected {
  background-color: #eff6ff;
  border-left: 3px solid #3b82f6;
}

.cars-table tbody tr.selected:hover {
  background-color: #dbeafe;
}

/* Hidden car indicator styles */
.hidden-icon {
  color: #ef4444;
  font-size: 14px;
  cursor: help;
  margin-left: 8px;
}

.hidden-icon:hover {
  color: #dc2626;
  transform: scale(1.1);
}

/* Hidden car row styles */
.cars-table tbody tr.hidden-car {
  background-color: #fef2f2;
  border-left: 3px solid #ef4444;
  opacity: 0.7;
}

.cars-table tbody tr.hidden-car:hover {
  background-color: #fee2e2;
  opacity: 0.8;
}

.cars-table tbody tr.hidden-car.selected {
  background-color: #fecaca;
  border-left: 3px solid #dc2626;
}

.cars-table tbody tr.hidden-car.selected:hover {
  background-color: #fca5a5;
}

/* Hidden info badges */
.badge-hidden {
  background: #fef2f2 !important;
  color: #dc2626 !important;
  border: 1px solid #fecaca !important;
  font-size: 10px !important;
  padding: 2px 4px !important;
  margin: 1px 0 !important;
  border-radius: 3px !important;
  white-space: nowrap !important;
}

.badge-hidden-time {
  background: #f3f4f6 !important;
  color: #6b7280 !important;
  border: 1px solid #d1d5db !important;
  font-size: 10px !important;
  padding: 2px 4px !important;
  margin: 1px 0 !important;
  border-radius: 3px !important;
  white-space: nowrap !important;
}

.hidden-info {
  margin-top: 2px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

/* Payment confirmed column */
.payment-confirmed-column {
  text-align: center;
  padding: 8px;
}

.payment-confirmed-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  background-color: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
}

.payment-confirmed-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.payment-confirmed-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.payment-confirmed-btn.confirmed {
  background-color: #dcfce7;
  color: #166534;
  border-color: #bbf7d0;
}

.payment-confirmed-btn.confirmed:hover:not(:disabled) {
  background-color: #bbf7d0;
}

.payment-confirmed-btn.not-confirmed {
  background-color: #fee2e2;
  color: #991b1b;
  border-color: #fecaca;
}

.payment-confirmed-btn.not-confirmed:hover:not(:disabled) {
  background-color: #fecaca;
}

.payment-confirmed-btn i {
  font-size: 1rem;
}

.payment-confirmed-btn.read-only {
  cursor: not-allowed;
  opacity: 0.7;
}

.payment-confirmed-btn.read-only:hover {
  transform: none;
  box-shadow: none;
}

/* Status column multiline support */
.status-column {
  white-space: normal !important;
  vertical-align: top !important;
  min-width: 200px;
  max-width: 250px;
}

.status-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
}

.status-item span {
  white-space: nowrap;
}

/* Frozen Buy Ref Group Column Styles */
.frozen-column {
  position: sticky;
  left: 0;
  z-index: 10;
  background-color: #f8f9fa;
  border-right: 2px solid #e5e7eb;
  box-shadow: 2px 0 4px rgba(0, 0, 0, 0.1);
}

.buy-ref-group-header {
  min-width: 280px !important;
  max-width: 280px !important;
  width: 280px !important;
  padding: 12px;
  text-align: center;
  font-weight: 600;
  background-color: #e0f2fe;
  border-bottom: 2px solid #0ea5e9;
  color: #0c4a6e;
}

.buy-ref-group-cell {
  min-width: 280px !important;
  max-width: 280px !important;
  width: 280px !important;
  padding: 12px;
  vertical-align: top;
  background-color: #f0f9ff;
  border-right: 2px solid #0ea5e9;
  border-bottom: 1px solid #bae6fd;
}

.buy-ref-group-cell.group-selected {
  background-color: #dbeafe;
  border-right-color: #3b82f6;
}

.buy-ref-group-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 8px;
  width: 100%;
}

.group-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #3b82f6;
}

.buy-ref-text {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  text-align: center;
}

.buy-ref-label {
  font-weight: 600;
  font-size: 13px;
  color: #1f2937;
  word-break: break-word;
  max-width: 100%;
  overflow-wrap: break-word;
}

.buy-ref-count {
  font-size: 11px;
  color: #6b7280;
  font-weight: 500;
}

.buy-ref-group-row {
  background-color: #fafafa;
}

.buy-ref-group-row:hover {
  background-color: #f0f9ff;
}

.buy-ref-group-row.selected {
  background-color: #e0f2fe;
}

/* Ensure frozen column stays on top of other columns when scrolling horizontally */
.table-container {
  position: relative;
}

.cars-table thead th.frozen-column {
  z-index: 101;
}

.cars-table tbody td.frozen-column {
  z-index: 10;
}

/* Combine Loading Overlay */
.combine-loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  backdrop-filter: blur(2px);
}

.combine-loading-content {
  background: white;
  padding: 24px 32px;
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 16px;
  font-weight: 500;
  color: #374151;
}

.combine-loading-content i {
  font-size: 20px;
  color: #3b82f6;
}

.table-container.processing-combine {
  opacity: 0.6;
  pointer-events: none;
}

/* Batch Transfer Modal Styles */
.batch-transfer-modal {
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.batch-transfer-modal .modal-body {
  padding: 20px;
}

.batch-transfer-modal .info-section {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  border-radius: 8px;
  margin-bottom: 20px;
  color: #1e40af;
  font-weight: 500;
}

.batch-transfer-modal .info-section i {
  color: #3b82f6;
}

.batch-transfer-modal .form-group {
  margin-bottom: 20px;
}

.batch-transfer-modal .form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.batch-transfer-modal .form-select,
.batch-transfer-modal .form-input,
.batch-transfer-modal .form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
  transition: all 0.2s ease;
}

.batch-transfer-modal .form-select:focus,
.batch-transfer-modal .form-input:focus,
.batch-transfer-modal .form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.batch-transfer-modal .form-select:disabled,
.batch-transfer-modal .form-input:disabled,
.batch-transfer-modal .form-textarea:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.batch-transfer-modal .form-textarea {
  resize: vertical;
  min-height: 80px;
}

.batch-transfer-modal .files-preview {
  margin-top: 20px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.batch-transfer-modal .files-preview h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.batch-transfer-modal .files-list {
  max-height: 200px;
  overflow-y: auto;
}

.batch-transfer-modal .file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: white;
  border-radius: 4px;
  margin-bottom: 6px;
  font-size: 13px;
  color: #374151;
}

.batch-transfer-modal .file-item i {
  color: #6b7280;
}

.batch-transfer-modal .file-item .car-name {
  color: #9ca3af;
  font-size: 12px;
  margin-left: auto;
}

.batch-transfer-modal .modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px;
  border-top: 1px solid #e5e7eb;
}

.batch-transfer-modal .cancel-btn,
.batch-transfer-modal .confirm-btn {
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
  border: none;
}

.batch-transfer-modal .cancel-btn {
  background: #f3f4f6;
  color: #374151;
}

.batch-transfer-modal .cancel-btn:hover:not(:disabled) {
  background: #e5e7eb;
}

.batch-transfer-modal .confirm-btn {
  background: #3b82f6;
  color: white;
}

.batch-transfer-modal .confirm-btn:hover:not(:disabled) {
  background: #2563eb;
}

.batch-transfer-modal .confirm-btn:disabled,
.batch-transfer-modal .cancel-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Batch Checkout Modal Styles - reuse transfer modal styles */
.batch-checkout-modal {
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.batch-checkout-modal .modal-body {
  padding: 20px;
}

.batch-checkout-modal .info-section {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  border-radius: 8px;
  margin-bottom: 20px;
  color: #1e40af;
  font-weight: 500;
}

.batch-checkout-modal .info-section i {
  color: #3b82f6;
}

.batch-checkout-modal .form-group {
  margin-bottom: 20px;
}

.batch-checkout-modal .form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.batch-checkout-modal .form-select,
.batch-checkout-modal .form-input,
.batch-checkout-modal .form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
  transition: all 0.2s ease;
}

.batch-checkout-modal .form-select:focus,
.batch-checkout-modal .form-input:focus,
.batch-checkout-modal .form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.batch-checkout-modal .form-select:disabled,
.batch-checkout-modal .form-input:disabled,
.batch-checkout-modal .form-textarea:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.batch-checkout-modal .form-textarea {
  resize: vertical;
  min-height: 80px;
}

.batch-checkout-modal .files-preview {
  margin-top: 20px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.batch-checkout-modal .files-preview h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.batch-checkout-modal .files-list {
  max-height: 200px;
  overflow-y: auto;
}

.batch-checkout-modal .file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: white;
  border-radius: 4px;
  margin-bottom: 6px;
  font-size: 13px;
  color: #374151;
}

.batch-checkout-modal .file-item i {
  color: #6b7280;
}

.batch-checkout-modal .file-item .car-name {
  color: #9ca3af;
  font-size: 12px;
  margin-left: auto;
}

.batch-checkout-modal .modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px;
  border-top: 1px solid #e5e7eb;
}

.batch-checkout-modal .cancel-btn,
.batch-checkout-modal .confirm-btn {
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
  border: none;
}

.batch-checkout-modal .cancel-btn {
  background: #f3f4f6;
  color: #374151;
}

.batch-checkout-modal .cancel-btn:hover:not(:disabled) {
  background: #e5e7eb;
}

.batch-checkout-modal .confirm-btn {
  background: #3b82f6;
  color: white;
}

.batch-checkout-modal .confirm-btn:hover:not(:disabled) {
  background: #2563eb;
}

.batch-checkout-modal .confirm-btn:disabled,
.batch-checkout-modal .cancel-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
