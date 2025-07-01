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

// Add new refs for teleport dropdown
const teleportDropdown = ref({
  isOpen: false,
  carId: null,
  position: { x: 0, y: 0 },
  buttonElement: null,
})

const router = useRouter()

const fetchCarsStock = async () => {
  loading.value = true
  error.value = null

  try {
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
        lp.loading_port,
        dp.discharge_port,
        bd.price_sell as buy_price,
        w.warhouse_name as warehouse_name,
        bb.bill_ref as buy_bill_ref,
        sb.bill_ref as sell_bill_ref,
        cs.is_used_car,
        c.id_no as client_id_no,
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
          console.log('Status filter value:', adv.status) // Debug
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
  if (teleportDropdown.value.isOpen) {
    const dropdown = document.querySelector('.teleport-dropdown')
    const button = teleportDropdown.value.buttonElement

    if (dropdown && !dropdown.contains(event.target) && button && !button.contains(event.target)) {
      closeTeleportDropdown()
    }
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
    console.log('Task created successfully with ID:', result.taskId)
    // You can add any additional logic here, like showing a success message
  }
  handleTaskClose()
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

const handleTaskAction = (car) => {
  closeTeleportDropdown()
  selectedCarForTask.value = car
  showTaskForm.value = true
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
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
    <div v-else class="table-container">
      <table class="cars-table">
        <thead>
          <tr>
            <th @click="toggleSort('id')" class="sortable">
              ID
              <span v-if="sortConfig.key === 'id'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th @click="toggleSort('car_name')" class="sortable">
              Car Details
              <span v-if="sortConfig.key === 'car_name'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th @click="toggleSort('loading_port')" class="sortable">
              Loading Port
              <span v-if="sortConfig.key === 'loading_port'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th @click="toggleSort('discharge_port')" class="sortable">
              Discharge Port
              <span v-if="sortConfig.key === 'discharge_port'" class="sort-indicator">
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
              Price
              <span v-if="sortConfig.key === 'price_cell'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th @click="toggleSort('date_loding')" class="sortable">
              Operation Date
              <span v-if="sortConfig.key === 'date_loding'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th @click="toggleSort('status')" class="sortable">
              Status
              <span v-if="sortConfig.key === 'status'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th @click="toggleSort('notes')" class="sortable">
              Notes
              <span v-if="sortConfig.key === 'notes'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th @click="toggleSort('client_name')" class="sortable">
              Client
              <span v-if="sortConfig.key === 'client_name'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th @click="toggleSort('warehouse_name')" class="sortable">
              Warehouse
              <span v-if="sortConfig.key === 'warehouse_name'" class="sort-indicator">
                {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
              </span>
            </th>
            <th>Documents</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="car in sortedCars" :key="car.id" :class="{ 'used-car': car.is_used_car }">
            <td>#{{ car.id }}</td>
            <td class="car-details-cell">
              <div class="car-name">{{ car.car_name }}</div>
              <div v-if="car.vin" class="info-badge badge-vin">
                <i class="fas fa-barcode"></i>
                {{ car.vin }}
              </div>
              <div v-if="car.color" class="info-badge badge-color">
                <i class="fas fa-palette"></i>
                {{ car.color }}
              </div>
              <div v-if="car.export_lisence_ref" class="info-badge badge-export-license">
                <i class="fas fa-certificate"></i>
                Export License
              </div>
              <div v-if="car.date_pay_freight" class="info-badge badge-freight-paid">
                <i class="fas fa-check-circle"></i>
                Freight Paid
              </div>
              <div v-if="car.buy_bill_ref" class="info-badge badge-buy-bill">
                <i class="fas fa-shopping-cart"></i>
                Buy: {{ car.buy_bill_ref }}
              </div>
              <div v-if="car.sell_bill_ref" class="info-badge badge-sell-bill">
                <i class="fas fa-file-invoice-dollar"></i>
                Sell: {{ car.sell_bill_ref }}
              </div>
            </td>
            <td>
              <div v-if="car.loading_port" class="info-badge badge-loading-port">
                <i class="fas fa-ship"></i>
                {{ car.loading_port }}
              </div>
              <div v-else>-</div>
            </td>
            <td>
              <div>
                {{ car.discharge_port || '-' }}
                <div v-if="car.container_ref" class="info-badge badge-container">
                  <i class="fas fa-box"></i>
                  {{ car.container_ref }}
                </div>
              </div>
            </td>
            <td>${{ car.freight || '0' }}</td>
            <td>${{ car.price_cell || '0' }}</td>
            <td>{{ car.date_loding ? new Date(car.date_loding).toLocaleDateString() : '-' }}</td>
            <td>
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
              <div v-if="car.in_wharhouse_date" class="info-badge badge-in-warehouse">
                <i class="fas fa-warehouse"></i>
                In Warehouse
              </div>
              <div v-if="car.date_get_bl" class="info-badge badge-bl-received">
                <i class="fas fa-file-contract"></i>
                BL Received
              </div>
            </td>
            <td class="notes-cell" :title="car.notes">{{ car.notes || '-' }}</td>
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
                  Documents
                </a>
                <a
                  v-if="car.sell_pi_path"
                  :href="getFileUrl(car.sell_pi_path)"
                  target="_blank"
                  class="document-link"
                >
                  <i class="fas fa-file-invoice-dollar"></i>
                  Sell PI
                </a>
                <a
                  v-if="car.buy_pi_path"
                  :href="getFileUrl(car.buy_pi_path)"
                  target="_blank"
                  class="document-link"
                >
                  <i class="fas fa-file-contract"></i>
                  Buy PI
                </a>
              </div>
            </td>
            <td>
              <div class="dropdown">
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
                      <i
                        v-if="isProcessing.vin"
                        class="fas fa-spinner fa-spin loading-indicator"
                      ></i>
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
                </ul>
              </div>

              <!-- Teleport Dropdown Button (Alternative) -->
              <button
                @click="openTeleportDropdown(car.id, $event)"
                class="teleport-dropdown-toggle"
                title="Actions (Teleport)"
              >
                <i class="fas fa-ellipsis-h"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
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
      </ul>
    </div>
  </teleport>
</template>

<style scoped>
.cars-stock-table {
  width: 100%;
  overflow-x: auto;
  position: relative;
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
}

.cars-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.cars-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #f8f9fa;
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
  margin-left: 8px;
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
  right: 0;
  background-color: white;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  list-style: none;
  padding: 4px;
  margin: 4px 0;
  z-index: 10;
  min-width: 180px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
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
  background-color: #fef3c7;
  color: #92400e;
  border: 1px solid #fde68a;
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

.badge-wholesale {
  background-color: #dbeafe;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}
</style>
