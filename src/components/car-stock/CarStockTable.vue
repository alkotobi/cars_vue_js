<script setup>
import { ref, onMounted, defineProps, defineEmits, computed, watch, onUnmounted } from 'vue'
import { useApi } from '../../composables/useApi'
import VinEditForm from './VinEditForm.vue'
import CarFilesUploadForm from './CarFilesUploadForm.vue'
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
import CarWarehouseBulkEditForm from './CarWarehouseBulkEditForm.vue'
import CarNotesBulkEditForm from './CarNotesBulkEditForm.vue'

import { useRouter } from 'vue-router'

const isDropdownOpen = ref({})
const props = defineProps({
  cars: Array,
  onEdit: {
    type: Function,
    default: () => {},
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
        status: '',
        client: '',
        client_id_no: '',
        warehouse: '',
        container_ref: '',
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

const user = ref(null)
const can_edit_cars_prop = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_prop')
})

const emit = defineEmits(['refresh'])

const { callApi, getFileUrl } = useApi()
const cars = ref([])
const loading = ref(true)
const error = ref(null)
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

// Add new refs for ports bulk edit form
const showPortsBulkEditForm = ref(false)

// Add new refs for warehouse bulk edit form
const showWarehouseBulkEditForm = ref(false)

// Add new refs for notes bulk edit form
const showNotesBulkEditForm = ref(false)

// Add computed property for admin permission
const isAdmin = computed(() => {
  if (!user.value) return false
  return user.value.role_id === 1
})

// Add to the data/refs section
const showAdvancedFilter = ref(false)
const advancedFilters = ref({})

// Add sorting state
const sortConfig = ref({
  key: 'id',
  direction: 'desc',
})

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
    alert('No cars selected for printing')
    return
  }

  printOptionsActionType.value = 'print'
  showPrintOptions.value = true
}

const handlePrintOptionsClose = () => {
  showPrintOptions.value = false
}

const handlePrintWithOptions = (printData) => {
  const { columns, cars } = printData
  console.log('Printing with options:', { columns, cars })

  // Generate print content based on action type
  let title = ''
  let contentBeforeTable = ''
  let contentAfterTable = ''

  if (printOptionsActionType.value === 'print') {
    title = 'Car Stock Report'
    contentBeforeTable = `
      <p>This report contains information about ${cars.length} car${cars.length === 1 ? '' : 's'} in stock.</p>
      <p>Report generated on ${new Date().toLocaleDateString()}.</p>
    `
    contentAfterTable = `
      <p><strong>Total Cars:</strong> ${cars.length}</p>
      <p><strong>Report Type:</strong> Stock Inventory</p>
    `
  } else if (printOptionsActionType.value === 'loading-order') {
    title = 'Loading Order Report'
    contentBeforeTable = `
      <p>This loading order contains ${cars.length} car${cars.length === 1 ? '' : 's'} to be loaded.</p>
      <p>Loading order generated on ${new Date().toLocaleDateString()}.</p>
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
  printReport({
    title,
    cars,
    columns,
    contentBeforeTable,
    contentAfterTable,
  })

  // Close the modal after handling the print event
  showPrintOptions.value = false
}

const handleLoadingOrderFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert('No cars selected for loading order')
    return
  }

  printOptionsActionType.value = 'loading-order'
  showPrintOptions.value = true
}

const handleLoadingOrderWithOptions = (data) => {
  const { columns, cars } = data
  console.log('Loading order with options:', { columns, cars })

  // Generate loading order content based on action type
  let title = 'Loading Order Report'
  let contentBeforeTable = `
    <p>This loading order contains ${cars.length} car${cars.length === 1 ? '' : 's'} to be loaded.</p>
    <p>Loading order generated on ${new Date().toLocaleDateString()}.</p>
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
  printReport({
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
    alert('No cars selected for VIN assignment')
    return
  }

  console.log('Opening VIN assignment modal')
  showVinAssignmentModal.value = true
  console.log('Modal state after opening:', showVinAssignmentModal.value)
}

const printReport = (reportData) => {
  const { title, cars: reportCars, columns, contentBeforeTable, contentAfterTable } = reportData

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
        .content-after-table { margin-top: 20px; line-height: 1.6; font-size: 14px; }
        @media print { body { padding: 0; margin: 0; } .report-table { font-size: 10px; } .table-header, .table-cell { padding: 6px 8px; } .letter-head { max-height: 80px; } .report-title { font-size: 20px; margin-bottom: 20px; } }
      </style>
    </head>
    <body>
      <div class="print-report">
        <div class="report-header">
          <img 
            src="/src/assets/letter_head.png" 
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
      console.log('Defaults loaded:', defaults.value)
    } else {
      console.log('No defaults found or API failed:', result)
    }
  } catch (error) {
    console.error('Error fetching defaults:', error)
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

// Add CFR DZA value calculation function
const getCfrDzaValue = (car) => {
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

const fetchCarsStock = async () => {
  loading.value = true
  error.value = null

  try {
    // Fetch defaults first if not already loaded
    if (!defaults.value) {
      await fetchDefaults()
    }

    // Build the base query
    let query = `
      SELECT 
        cs.id,
        cs.vin,
        cs.price_cell,
        cs.date_loding,
        cs.date_sell,
        cs.notes,
        cs.freight,
        cs.path_documents,
        cs.sell_pi_path,
        cs.buy_pi_path,
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
        c.name as client_name,
        cn.car_name,
        clr.color,
        clr.hexa,
        lp.loading_port,
        dp.discharge_port,
        bd.price_sell as buy_price,
        bb.date_buy,
        w.warhouse_name as warehouse_name,
        bb.bill_ref as buy_bill_ref,
        sb.bill_ref as sell_bill_ref,
        cs.is_used_car,
        cs.is_big_car,
        c.id_no as client_id_no,
        c.id_copy_path as client_id_picture,
        cs.container_ref,
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
      LEFT JOIN colors clr ON bd.id_color = clr.id
      LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
      LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
      LEFT JOIN warehouses w ON cs.id_warehouse = w.id
      WHERE cs.hidden = 0   
      AND cs.date_send_documents IS NULL
    `

    const params = []

    // Apply filters if they exist
    if (props.filters) {
      // Basic filter (search across multiple fields)
      if (props.filters.basic && props.filters.basic.length > 0) {
        // Handle both string (backward compatibility) and array of words
        const searchTerms = Array.isArray(props.filters.basic)
          ? props.filters.basic
          : [props.filters.basic]

        // Determine the operator to use (default to AND if not specified)
        const operator = props.filters.basicOperator || 'AND'

        if (searchTerms.length === 1) {
          // Single word search - use OR across all fields
          const term = `%${searchTerms[0].trim()}%`
          query += `
            AND (
              cs.id LIKE ? OR
              cn.car_name LIKE ? OR
              clr.color LIKE ? OR
              cs.vin LIKE ? OR
              lp.loading_port LIKE ? OR
              dp.discharge_port LIKE ? OR
              c.name LIKE ? OR
              c.id_no LIKE ? OR
              w.warhouse_name LIKE ? OR
              bb.bill_ref LIKE ? OR
              sb.bill_ref LIKE ?
            )
          `
          // Add the search parameter 11 times (once for each field)
          for (let i = 0; i < 11; i++) {
            params.push(term)
          }
        } else {
          // Multiple words - use the specified operator
          const conditions = searchTerms.map((searchTerm) => {
            const term = `%${searchTerm.trim()}%`
            return `(
              cs.id LIKE ? OR
              cn.car_name LIKE ? OR
              clr.color LIKE ? OR
              cs.vin LIKE ? OR
              lp.loading_port LIKE ? OR
              dp.discharge_port LIKE ? OR
              c.name LIKE ? OR
              c.id_no LIKE ? OR
              w.warhouse_name LIKE ? OR
              bb.bill_ref LIKE ? OR
              sb.bill_ref LIKE ?
            )`
          })

          query += ` AND (${conditions.join(` ${operator} `)})`

          // Add parameters for each search term (11 parameters per term)
          searchTerms.forEach((searchTerm) => {
            const term = `%${searchTerm.trim()}%`
            for (let i = 0; i < 11; i++) {
              params.push(term)
            }
          })
        }
      }

      // Advanced filters
      if (props.filters.advanced) {
        const adv = props.filters.advanced

        // ID filter
        if (adv.id && adv.id.trim() !== '') {
          query += ` AND cs.id = ?`
          params.push(adv.id.trim())
        }

        // Car name filter
        if (adv.car_name && adv.car_name.trim() !== '') {
          query += ` AND cn.car_name = ?`
          params.push(adv.car_name.trim())
        }

        // Color filter
        if (adv.color && adv.color.trim() !== '') {
          query += ` AND clr.color = ?`
          params.push(adv.color.trim())
        }

        // VIN filter
        if (adv.vin && adv.vin.trim() !== '') {
          query += ` AND cs.vin LIKE ?`
          params.push(`%${adv.vin.trim()}%`)
        }

        // Loading port filter
        if (adv.loading_port && adv.loading_port.trim() !== '') {
          query += ` AND lp.loading_port = ?`
          params.push(adv.loading_port.trim())
        }

        // Discharge port filter
        if (adv.discharge_port && adv.discharge_port.trim() !== '') {
          query += ` AND dp.discharge_port = ?`
          params.push(adv.discharge_port.trim())
        }

        // Freight range filter
        if (adv.freight_min && adv.freight_min.toString().trim() !== '') {
          query += ` AND cs.freight >= ?`
          params.push(parseFloat(adv.freight_min))
        }
        if (adv.freight_max && adv.freight_max.toString().trim() !== '') {
          query += ` AND cs.freight <= ?`
          params.push(parseFloat(adv.freight_max))
        }

        // Price range filter
        if (adv.price_min && adv.price_min.toString().trim() !== '') {
          query += ` AND cs.price_cell >= ?`
          params.push(parseFloat(adv.price_min))
        }
        if (adv.price_max && adv.price_max.toString().trim() !== '') {
          query += ` AND cs.price_cell <= ?`
          params.push(parseFloat(adv.price_max))
        }

        // Loading date range filter
        if (adv.loading_date_from && adv.loading_date_from.trim() !== '') {
          query += ` AND cs.date_loding >= ?`
          params.push(adv.loading_date_from.trim())
        }
        if (adv.loading_date_to && adv.loading_date_to.trim() !== '') {
          query += ` AND cs.date_loding <= ?`
          params.push(adv.loading_date_to.trim())
        }

        // Status filter (Available/Sold)
        if (adv.status && adv.status.trim() !== '') {
          if (adv.status === 'available') {
            query += ` AND cs.id_sell IS NULL`
          } else if (adv.status === 'sold') {
            query += ` AND cs.id_sell IS NOT NULL`
          }
        }

        // Client filter
        if (adv.client && adv.client.trim() !== '') {
          query += ` AND c.name = ?`
          params.push(adv.client.trim())
        }

        // Client ID Number filter
        if (adv.client_id_no && adv.client_id_no.trim() !== '') {
          query += ` AND c.id_no LIKE ?`
          params.push(`%${adv.client_id_no.trim()}%`)
        }

        // Warehouse filter
        if (adv.warehouse && adv.warehouse.trim() !== '') {
          query += ` AND w.warhouse_name = ?`
          params.push(adv.warehouse.trim())
        }

        // Container Reference filter
        if (adv.container_ref && adv.container_ref.trim() !== '') {
          query += ` AND cs.container_ref LIKE ?`
          params.push(`%${adv.container_ref.trim()}%`)
        }

        // Loading Status filter
        if (adv.loading_status && adv.loading_status.trim() !== '') {
          if (adv.loading_status === 'loaded') {
            query += ` AND cs.date_loding IS NOT NULL`
          } else if (adv.loading_status === 'not_loaded') {
            query += ` AND cs.date_loding IS NULL AND cs.is_batch = 0`
          }
        }

        // Documents Status filter
        if (adv.documents_status && adv.documents_status.trim() !== '') {
          if (adv.documents_status === 'received') {
            query += ` AND cs.date_get_documents_from_supp IS NOT NULL`
          } else if (adv.documents_status === 'not_received') {
            query += ` AND cs.date_get_documents_from_supp IS NULL`
          }
        }

        // BL Status filter
        if (adv.bl_status && adv.bl_status.trim() !== '') {
          if (adv.bl_status === 'received') {
            query += ` AND cs.date_get_bl IS NOT NULL`
          } else if (adv.bl_status === 'not_received') {
            query += ` AND cs.date_get_bl IS NULL`
          }
        }

        // Warehouse Status filter
        if (adv.warehouse_status && adv.warehouse_status.trim() !== '') {
          if (adv.warehouse_status === 'in_warehouse') {
            query += ` AND cs.in_wharhouse_date IS NOT NULL`
          } else if (adv.warehouse_status === 'not_in_warehouse') {
            query += ` AND cs.in_wharhouse_date IS NULL`
          }
        }

        // Bill Reference filter
        if (adv.bill_ref && adv.bill_ref.trim() !== '') {
          query += ` AND bb.bill_ref LIKE ?`
          params.push(`%${adv.bill_ref.trim()}%`)
        }

        // Sell Bill Reference filter
        if (adv.sell_bill_ref && adv.sell_bill_ref.trim() !== '') {
          query += ` AND sb.bill_ref LIKE ?`
          params.push(`%${adv.sell_bill_ref.trim()}%`)
        }

        // Temporary Client Status filter
        if (adv.tmp_client_status && adv.tmp_client_status.trim() !== '') {
          if (adv.tmp_client_status === 'tmp') {
            query += ` AND cs.is_tmp_client = 1`
          } else if (adv.tmp_client_status === 'permanent') {
            query += ` AND cs.is_tmp_client = 0`
          }
        }

        // Exclude Whole Sale filter
        if (adv.exclude_whole_sale) {
          query += ` AND cs.is_batch = 0`
        }

        if (adv.has_bl) {
          query += ` AND cs.date_get_bl IS NOT NULL`
        }

        if (adv.freight_paid) {
          query += ` AND cs.date_pay_freight IS NOT NULL`
        }

        if (adv.has_supplier_docs) {
          query += ` AND cs.date_get_documents_from_supp IS NOT NULL`
        }

        if (adv.in_warehouse) {
          query += ` AND cs.in_wharhouse_date IS NOT NULL`
        }

        if (adv.has_export_license) {
          query += ` AND cs.export_lisence_ref IS NOT NULL AND cs.export_lisence_ref != ''`
        }

        if (adv.is_loaded) {
          query += ` AND cs.date_loding IS NOT NULL`
        }

        if (adv.has_vin) {
          query += ` AND cs.vin IS NOT NULL AND cs.vin != ''`
        }
      }
    }

    const result = await callApi({
      query,
      params,
    })

    if (result.success) {
      cars.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch cars stock'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const handleEdit = async (car) => {
  if (isProcessing.value.edit) return
  isProcessing.value.edit = true
  try {
    props.onEdit(car)
  } finally {
    isProcessing.value.edit = false
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

const handleFilesSave = (updatedCar) => {
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

const handleWarehouseSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
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
  if (result.success) {
    handleTaskClose()
  }
}

// Add new methods for switch buy bill form
const handleSwitchBuyBillClose = () => {
  showSwitchBuyBillForm.value = false
  selectedCarForSwitchBuyBill.value = null
}

const handleSwitchBuyBillSave = async (result) => {
  console.log('Switch result received:', result)

  if (result.success) {
    try {
      console.log('Starting switch process for cars:', result.carId1, 'and', result.carId2)

      // Use the original SQL with direct ID insertion
      const switchResult = await callApi({
        action: 'execute_multi_sql',
        query: `
          SET @id1 = ${result.carId1};
          SET @id2 = ${result.carId2};

          -- Start transaction to ensure data integrity
          START TRANSACTION;

          -- Get the current id_buy_details values
          SET @buy_details_1 = (SELECT id_buy_details FROM cars_stock WHERE id = @id1);
          SET @buy_details_2 = (SELECT id_buy_details FROM cars_stock WHERE id = @id2);

          -- Update the first record with the second record's id_buy_details
          UPDATE cars_stock 
          SET id_buy_details = @buy_details_2 
          WHERE id = @id1;

          -- Update the second record with the first record's id_buy_details
          UPDATE cars_stock 
          SET id_buy_details = @buy_details_1 
          WHERE id = @id2;

          -- Commit the transaction
          COMMIT;

          -- Optional: Verify the switch was successful
          SELECT 
              id,
              id_buy_details,
              CASE 
                  WHEN id = @id1 THEN 'First Record (was @buy_details_1, now @buy_details_2)'
                  WHEN id = @id2 THEN 'Second Record (was @buy_details_2, now @buy_details_1)'
              END as switch_info
          FROM cars_stock 
          WHERE id IN (@id1, @id2)
          ORDER BY id;
        `,
        params: [],
      })

      console.log('Switch result:', switchResult)

      if (switchResult.success) {
        alert('Purchase bills switched successfully!')
        handleSwitchBuyBillClose()
        // Refresh the cars data
        await fetchCarsStock()
      } else {
        console.error('Switch failed:', switchResult)
        alert('Failed to switch purchase bills')
      }
    } catch (err) {
      console.error('Error switching purchase bills:', err)
      alert('Error switching purchase bills')
    }
  } else {
    console.log('Switch result was not successful:', result)
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

const saveNotes = async () => {
  if (!notesEditCar.value) return
  isProcessing.notes = true
  try {
    const result = await callApi({
      query: 'UPDATE cars_stock SET notes = ? WHERE id = ?',
      params: [notesEditValue.value, notesEditCar.value.id],
    })
    if (result.success) {
      notesEditCar.value.notes = notesEditValue.value
      // Also update in cars array if needed
      if (cars.value && Array.isArray(cars.value)) {
        const carIdx = cars.value.findIndex((c) => c.id === notesEditCar.value.id)
        if (carIdx !== -1) cars.value[carIdx].notes = notesEditValue.value
      }
      showNotesModal.value = false
    } else {
      alert(result.error || 'Failed to update notes')
    }
  } catch (err) {
    alert(err.message || 'Error updating notes')
  } finally {
    isProcessing.notes = false
  }
}

const cancelNotes = () => {
  showNotesModal.value = false
}

const handleVinsAssigned = (assignments) => {
  console.log('VINs assigned:', assignments)
  showVinAssignmentModal.value = false
  fetchCarsStock()
}

// Add a separate close handler for debugging
const handleVinAssignmentModalClose = () => {
  console.log('Closing VIN assignment modal from parent')
  showVinAssignmentModal.value = false
}

const handlePortsFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert('No cars selected for ports editing')
    return
  }

  showPortsBulkEditForm.value = true
}

const handlePortsBulkSave = (updatedCars) => {
  console.log('Ports updated for cars:', updatedCars)
  showPortsBulkEditForm.value = false
  fetchCarsStock()
}

const handleWarehouseFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert('No cars selected for warehouse editing')
    return
  }

  showWarehouseBulkEditForm.value = true
}

const handleWarehouseBulkSave = (updatedCars) => {
  console.log('Warehouse updated for cars:', updatedCars)
  showWarehouseBulkEditForm.value = false
  fetchCarsStock()
}

const handleNotesFromToolbar = () => {
  if (selectedCars.value.size === 0) {
    alert('No cars selected for notes editing')
    return
  }

  showNotesBulkEditForm.value = true
}

const handleNotesBulkSave = (updatedCars) => {
  console.log('Notes updated for cars:', updatedCars)
  showNotesBulkEditForm.value = false
  fetchCarsStock()
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchDefaults()
  fetchCarsStock()

  // Add event listeners for teleport dropdown
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('scroll', handleScroll, true)
})

onUnmounted(() => {
  // Remove event listeners
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('scroll', handleScroll, true)
})

// Expose the fetchCarsStock method to parent components
defineExpose({
  fetchCarsStock,
})
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

    <CarFilesUploadForm
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
      :show="showWarehouseForm"
      @close="showWarehouseForm = false"
      @save="handleWarehouseSave"
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
      Loading...
    </div>
    <div v-else-if="error" class="error">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>
    <div v-else-if="cars.length === 0" class="empty-state">
      <i class="fas fa-car fa-2x"></i>
      <p>No cars in stock</p>
    </div>
    <div v-else>
      <!-- Toolbar -->
      <CarStockToolbar
        :selected-cars="selectedCars"
        :total-cars="sortedCars.length"
        @print-selected="handlePrintSelected"
        @loading-order="handleLoadingOrderFromToolbar"
        @vin="handleVinFromToolbar"
        @ports="handlePortsFromToolbar"
        @warehouse="handleWarehouseFromToolbar"
        @notes="handleNotesFromToolbar"
      />

      <div class="table-container">
        <table class="cars-table">
          <thead>
            <tr>
              <th class="select-all-header">
                <input
                  type="checkbox"
                  :checked="selectAll"
                  @change="selectAllCars"
                  :indeterminate="selectedCars.size > 0 && selectedCars.size < sortedCars.length"
                  title="Select All Cars"
                />
              </th>
              <th>Actions</th>
              <th @click="toggleSort('id')" class="sortable">
                ID
                <span v-if="sortConfig.key === 'id'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('date_buy')" class="sortable">
                Date Buy
                <span v-if="sortConfig.key === 'date_buy'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('date_sell')" class="sortable">
                Date Sell
                <span v-if="sortConfig.key === 'date_sell'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('car_name')" class="sortable">
                Car Details
                <span v-if="sortConfig.key === 'car_name'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('client_name')" class="sortable">
                Client
                <span v-if="sortConfig.key === 'client_name'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('loading_port')" class="sortable">
                Ports
                <span v-if="sortConfig.key === 'loading_port'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('freight')" class="sortable">
                Freight
                <span v-if="sortConfig.key === 'freight'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('price_cell')" class="sortable">
                FOB
                <span v-if="sortConfig.key === 'price_cell'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('rate')" class="sortable">
                CFR
                <span v-if="sortConfig.key === 'rate'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('status')" class="sortable">
                Status
                <span v-if="sortConfig.key === 'status'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th @click="toggleSort('warehouse_name')" class="sortable">
                Warehouse
                <span v-if="sortConfig.key === 'warehouse_name'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
              <th>BL</th>
              <th @click="toggleSort('notes')" class="sortable">
                Notes
                <span v-if="sortConfig.key === 'notes'" class="sort-indicator">
                  {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="car in sortedCars"
              :key="car.id"
              :class="{ 'used-car': car.is_used_car, selected: selectedCars.has(car.id) }"
            >
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
                  <div class="car-name">{{ car.car_name }}</div>
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
                  <div v-if="car.export_lisence_ref" class="car-detail-item">
                    <div class="info-badge badge-export-license">
                      <i class="fas fa-certificate"></i>
                      Export License
                    </div>
                  </div>
                  <div v-if="car.buy_bill_ref" class="car-detail-item">
                    <div class="info-badge badge-buy-bill">
                      <i class="fas fa-shopping-cart"></i>
                      Buy: {{ car.buy_bill_ref }}
                    </div>
                  </div>
                  <div v-if="car.sell_bill_ref" class="car-detail-item">
                    <div class="info-badge badge-sell-bill">
                      <i class="fas fa-file-invoice-dollar"></i>
                      Sell: {{ car.sell_bill_ref }}
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
                    Whole Sale
                  </div>
                </div>
              </td>
              <td>
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
                    Freight Paid
                  </div>
                </div>
              </td>
              <td>${{ getFreightValue(car) }}</td>
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
              <td>
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
                      {{ car.id_sell ? 'Sold' : 'Available' }}
                      {{ car.is_used_car ? '(Used)' : '' }}
                    </span>
                  </div>
                  <div v-if="car.in_wharhouse_date" class="status-item">
                    <div class="info-badge badge-in-warehouse">
                      <i class="fas fa-warehouse"></i>
                      In Warehouse
                    </div>
                  </div>
                  <div v-if="car.date_get_bl" class="status-item">
                    <div class="info-badge badge-bl-received">
                      <i class="fas fa-file-contract"></i>
                      BL Received
                    </div>
                  </div>
                </div>
              </td>
              <td>
                <div v-if="car.warehouse_name" class="info-badge badge-warehouse">
                  <i class="fas fa-building"></i>
                  {{ car.warehouse_name }}
                </div>
                <div v-else>-</div>
              </td>
              <td class="documents-cell">
                <div class="document-links">
                  <a
                    v-if="car.path_documents"
                    :href="getFileUrl(car.path_documents)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-pdf"></i>
                    BL
                  </a>
                  <a
                    v-if="car.sell_pi_path"
                    :href="getFileUrl(car.sell_pi_path)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-invoice-dollar"></i>
                    INVOICE
                  </a>
                  <a
                    v-if="car.buy_pi_path"
                    :href="getFileUrl(car.buy_pi_path)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-contract"></i>
                    PACKING LIST
                  </a>
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
                    <span>Task</span>
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
                    <span>VIN</span>
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
                    <span>Files</span>
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
                    <span>Ports</span>
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
                    <span>Money</span>
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
                    <span>Warehouse</span>
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
                    <span>Car Documents</span>
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
                    <span>Load</span>
                    <i
                      v-if="isProcessing.load"
                      class="fas fa-spinner fa-spin loading-indicator"
                    ></i>
                  </button>
                </li>
                <li>
                  <button @click="openSellBillTab(car)" :disabled="!car.id_sell">
                    <i class="fas fa-file-invoice-dollar"></i>
                    <span>Sell Bill</span>
                  </button>
                </li>
                <li>
                  <button
                    @click="handleNotesAction(car)"
                    :disabled="isProcessing.notes"
                    :class="{ processing: isProcessing.notes }"
                  >
                    <i class="fas fa-sticky-note"></i>
                    <span>Notes</span>
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
                        ? 'Only administrators can switch purchase bills'
                        : 'Switch purchase bill with another car'
                    "
                  >
                    <i class="fas fa-exchange-alt"></i>
                    <span>Switch Purchase Bill</span>
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
                Export License
              </div>
              <div v-if="car.buy_bill_ref" class="info-badge badge-buy-bill">
                <i class="fas fa-shopping-cart"></i>
                Buy: {{ car.buy_bill_ref }}
              </div>
              <div v-if="car.sell_bill_ref" class="info-badge badge-sell-bill">
                <i class="fas fa-file-invoice-dollar"></i>
                Sell: {{ car.sell_bill_ref }}
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
                Whole Sale
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
              Freight Paid
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
                {{ car.id_sell ? 'Sold' : 'Available' }}
                {{ car.is_used_car ? '(Used)' : '' }}
              </span>
            </div>
            <div v-if="car.in_wharhouse_date" class="info-badge badge-in-warehouse">
              <i class="fas fa-warehouse"></i>
              In Warehouse
            </div>
            <div v-if="car.date_get_bl" class="info-badge badge-bl-received">
              <i class="fas fa-file-contract"></i>
              BL Received
            </div>
          </div>
        </div>

        <!-- Warehouse Section -->
        <div class="card-section">
          <div class="section-title">Warehouse</div>
          <div v-if="car.warehouse_name" class="info-badge badge-warehouse">
            <i class="fas fa-building"></i>
            {{ car.warehouse_name }}
          </div>
          <div v-else class="card-value">-</div>
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
              INVOICE
            </a>
            <a
              v-if="car.buy_pi_path"
              :href="getFileUrl(car.buy_pi_path)"
              target="_blank"
              class="card-document-link"
            >
              <i class="fas fa-file-contract"></i>
              PACKING LIST
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
                ? 'Only administrators can switch purchase bills'
                : 'Switch purchase bill with another car'
            "
          >
            <i class="fas fa-exchange-alt"></i>
            <span>Switch Purchase Bill</span>
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

  <!-- Warehouse Bulk Edit Modal -->
  <CarWarehouseBulkEditForm
    :show="showWarehouseBulkEditForm"
    :selected-cars="sortedCars.filter((car) => selectedCars.has(car.id))"
    :is-admin="isAdmin"
    @close="showWarehouseBulkEditForm = false"
    @save="handleWarehouseBulkSave"
  />

  <!-- Notes Bulk Edit Modal -->
  <CarNotesBulkEditForm
    :show="showNotesBulkEditForm"
    :selected-cars="sortedCars.filter((car) => selectedCars.has(car.id))"
    :is-admin="isAdmin"
    @close="showNotesBulkEditForm = false"
    @save="handleNotesBulkSave"
  />
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
  flex-direction: column;
  gap: 4px;
}

.document-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.85em;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.document-link:hover {
  background-color: #f3f4f6;
  text-decoration: underline;
}

.document-link i {
  color: #ef4444;
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
}

.badge-client-id {
  background-color: #f0fdf4;
  color: #15803d;
  border: 1px solid #bbf7d0;
}

.badge-export-license {
  background-color: #fef3c7;
  color: #92400e;
  border: 1px solid #fde68a;
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
}

.badge-cfr-dza {
  background-color: #fef3c7;
  color: #92400e;
  border: 1px solid #fde68a;
}

.badge-cfr-rate {
  background-color: #dbeafe;
  color: #1e40af;
  border: 1px solid #bfdbfe;
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
</style>
