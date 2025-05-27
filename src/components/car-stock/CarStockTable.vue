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
import AdvancedFilter from './AdvancedFilter.vue'

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
        warehouse: '',
        has_bl: false,
        freight_paid: false,
        has_supplier_docs: false,
        in_warehouse: false,
        has_export_license: false,
        documents_sent: false,
        is_loaded: false,
        has_vin: false,
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

// Add to the data/refs section
const showAdvancedFilter = ref(false)
const advancedFilters = ref({})

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
        c.name as client_name,
        cn.car_name,
        clr.color,
        lp.loading_port,
        dp.discharge_port,
        bd.price_sell as buy_price,
        w.warhouse_name as warehouse_name,
        bb.bill_ref as buy_bill_ref,
        cs.is_used_car
      FROM cars_stock cs
      LEFT JOIN clients c ON cs.id_client = c.id
      LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
      LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
      LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
      LEFT JOIN colors clr ON bd.id_color = clr.id
      LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
      LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
      LEFT JOIN warehouses w ON cs.id_warehouse = w.id
      WHERE cs.hidden = 0
    `

    const params = []

    // Apply filters if they exist
    if (props.filters) {
      // Basic filter (search across multiple fields)
      if (props.filters.basic && props.filters.basic.trim() !== '') {
        const searchTerm = `%${props.filters.basic.trim()}%`
        query += `
          AND (
            cs.id LIKE ? OR
            cn.car_name LIKE ? OR
            clr.color LIKE ? OR
            cs.vin LIKE ? OR
            lp.loading_port LIKE ? OR
            dp.discharge_port LIKE ? OR
            c.name LIKE ? OR
            w.warhouse_name LIKE ?
          )
        `
        // Add the search parameter 8 times (once for each field)
        for (let i = 0; i < 8; i++) {
          params.push(searchTerm)
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
        if (adv.freight_min && adv.freight_min.trim() !== '') {
          query += ` AND cs.freight >= ?`
          params.push(parseFloat(adv.freight_min.trim()))
        }
        if (adv.freight_max && adv.freight_max.trim() !== '') {
          query += ` AND cs.freight <= ?`
          params.push(parseFloat(adv.freight_max.trim()))
        }

        // Price range filter
        if (adv.price_min && adv.price_min.trim() !== '') {
          query += ` AND cs.price_cell >= ?`
          params.push(parseFloat(adv.price_min.trim()))
        }
        if (adv.price_max && adv.price_max.trim() !== '') {
          query += ` AND cs.price_cell <= ?`
          params.push(parseFloat(adv.price_max.trim()))
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
            query += ` AND (cs.date_sell IS NULL AND cs.id_client IS NULL)`
          } else if (adv.status === 'sold') {
            query += ` AND (cs.date_sell IS NOT NULL OR cs.id_client IS NOT NULL)`
          }
        }

        // Client filter
        if (adv.client && adv.client.trim() !== '') {
          query += ` AND c.name = ?`
          params.push(adv.client.trim())
        }

        // Warehouse filter
        if (adv.warehouse && adv.warehouse.trim() !== '') {
          query += ` AND w.name = ?`
          params.push(adv.warehouse.trim())
        }

        // Document status filters
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

        if (adv.documents_sent) {
          query += ` AND cs.date_send_documents IS NOT NULL`
        }

        if (adv.is_loaded) {
          query += ` AND cs.date_loding IS NOT NULL`
        }

        if (adv.has_vin) {
          query += ` AND cs.vin IS NOT NULL AND cs.vin != ''`
        }
      }
    }

    // Add the ORDER BY clause
    query += ` ORDER BY cs.id DESC`

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
  // Prevent the click from bubbling up to the document
  event.stopPropagation()

  // Close all other dropdowns
  const newDropdownState = {}
  newDropdownState[carId] = !isDropdownOpen.value[carId]
  isDropdownOpen.value = newDropdownState
}

// Add new methods for VIN editing
const handleVINAction = async (car) => {
  if (!can_edit_vin.value) {
    alert('You do not have permission to edit VIN numbers')
    return
  }
  if (isProcessing.value.vin) return
  isProcessing.value.vin = true
  try {
    selectedCar.value = car
    showVinEditForm.value = true
    isDropdownOpen.value[car.id] = false
  } finally {
    isProcessing.value.vin = false
  }
}

const handleVinSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for file upload
const handleFilesAction = async (car) => {
  if (!can_upload_car_files.value) {
    alert('You do not have permission to upload car files')
    return
  }
  if (isProcessing.value.files) return
  isProcessing.value.files = true
  try {
    selectedCar.value = car
    showFilesUploadForm.value = true
    isDropdownOpen.value[car.id] = false
  } finally {
    isProcessing.value.files = false
  }
}

const handleFilesSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for ports editing
const handlePortsAction = async (car) => {
  if (!can_edit_cars_ports.value) {
    alert('You do not have permission to edit car ports')
    return
  }
  if (isProcessing.value.ports) return
  isProcessing.value.ports = true
  try {
    selectedCar.value = car
    showPortsEditForm.value = true
    isDropdownOpen.value[car.id] = false
  } finally {
    isProcessing.value.ports = false
  }
}

const handlePortsSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for money editing
const handleMoneyAction = async (car) => {
  if (!can_edit_car_money.value) {
    alert('You do not have permission to edit car money fields')
    return
  }
  if (isProcessing.value.money) return
  isProcessing.value.money = true
  try {
    selectedCar.value = car
    showMoneyEditForm.value = true
    isDropdownOpen.value[car.id] = false
  } finally {
    isProcessing.value.money = false
  }
}

const handleMoneySave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new method for warehouse action
const handleWarehouseAction = async (car) => {
  if (!can_edit_warehouse.value) {
    alert('You do not have permission to edit warehouse')
    return
  }
  if (isProcessing.value.warehouse) return
  isProcessing.value.warehouse = true
  try {
    selectedCar.value = car
    showWarehouseForm.value = true
    isDropdownOpen.value[car.id] = false
  } finally {
    isProcessing.value.warehouse = false
  }
}

const handleWarehouseSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex((c) => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new method for documents action
const handleDocumentsClick = async (car) => {
  if (!can_edit_car_documents.value) {
    alert('You do not have permission to edit car documents')
    return
  }
  if (isProcessing.value.documents) return
  isProcessing.value.documents = true
  try {
    selectedCarForDocuments.value = car
    showDocumentsForm.value = true
    isDropdownOpen.value[car.id] = false
  } finally {
    isProcessing.value.documents = false
  }
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
const handleLoadClick = async (car) => {
  if (!can_load_car.value) {
    alert('You do not have permission to load cars')
    return
  }
  if (isProcessing.value.load) return
  isProcessing.value.load = true
  try {
    selectedCarForLoad.value = car
    showLoadForm.value = true
    isDropdownOpen.value[car.id] = false
  } finally {
    isProcessing.value.load = false
  }
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

// Add this method after other methods
const closeAllDropdowns = () => {
  isDropdownOpen.value = {}
}

const handleClickOutside = (event) => {
  const dropdowns = document.querySelectorAll('.dropdown')
  let clickedOutside = true

  dropdowns.forEach((dropdown) => {
    if (dropdown.contains(event.target)) {
      clickedOutside = false
    }
  })

  if (clickedOutside) {
    closeAllDropdowns()
  }
}

const handleAdvancedFilters = (filters) => {
  advancedFilters.value = filters
  showAdvancedFilter.value = false
  // Trigger the watcher by updating the filters prop
  props.filters.advanced = filters
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchCarsStock()
  }

  // Add click outside listener
  document.addEventListener('click', handleClickOutside)
})

// Add onUnmounted to clean up the event listener
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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
    />

    <CarLoadForm
      v-if="showLoadForm"
      :car="selectedCarForLoad"
      :show="showLoadForm"
      @close="handleLoadClose"
      @save="handleLoadSave"
    />

    <div class="table-actions">
      <button class="advanced-filter-btn" @click="showAdvancedFilter = true">
        <i class="fas fa-filter"></i>
        Advanced Filters
      </button>
    </div>

    <AdvancedFilter
      :show="showAdvancedFilter"
      :initial-filters="advancedFilters"
      @update:filters="handleAdvancedFilters"
      @close="showAdvancedFilter = false"
    />

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
            <th>ID</th>
            <th>Car</th>
            <th>Color</th>
            <th>VIN</th>
            <th>Loading Port</th>
            <th>Discharge Port</th>
            <th>Freight</th>
            <th>Price</th>
            <th>Loading Date</th>
            <th>Status</th>
            <th>Notes</th>
            <th>Client</th>
            <th>Warehouse</th>
            <th>Documents</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="car in cars" :key="car.id" :class="{ 'used-car': car.is_used_car }">
            <td>#{{ car.id }}</td>
            <td>{{ car.car_name }}</td>
            <td>{{ car.color }}</td>
            <td>{{ car.vin || '-' }}</td>
            <td>{{ car.loading_port || '-' }}</td>
            <td>{{ car.discharge_port || '-' }}</td>
            <td>${{ car.freight || '0' }}</td>
            <td>${{ car.price_cell || '0' }}</td>
            <td>{{ car.date_loding ? new Date(car.date_loding).toLocaleDateString() : '-' }}</td>
            <td>
              <span
                :class="{
                  'status-sold': car.date_sell || car.id_client,
                  'status-available': !car.date_sell && !car.id_client,
                  'status-used': car.is_used_car,
                }"
              >
                <i
                  :class="[
                    car.date_sell || car.id_client ? 'fas fa-check-circle' : 'fas fa-clock',
                    car.is_used_car ? 'fas fa-history' : '',
                  ]"
                ></i>
                {{ car.date_sell || car.id_client ? 'Sold' : 'Available' }}
                {{ car.is_used_car ? '(Used)' : '' }}
              </span>
            </td>
            <td class="notes-cell" :title="car.notes">{{ car.notes || '-' }}</td>
            <td>{{ car.client_name || '-' }}</td>
            <td>{{ car.warehouse_name || '-' }}</td>
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
              <button
                :disabled="!can_edit_cars_prop || isProcessing.edit"
                @click="handleEdit(car)"
                class="edit-btn"
                :class="{ processing: isProcessing.edit }"
              >
                <i class="fas fa-edit"></i>
                <span>Edit</span>
                <i v-if="isProcessing.edit" class="fas fa-spinner fa-spin loading-indicator"></i>
              </button>
              <div class="dropdown">
                <button @click="toggleDropdown(car.id, $event)" class="dropdown-toggle">
                  <i class="fas fa-ellipsis-v"></i>
                  Actions
                </button>
                <ul v-if="isDropdownOpen[car.id]" class="dropdown-menu">
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
                </ul>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
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
  border-radius: 4px;
  font-size: 0.9em;
  font-weight: 500;
}

.status-available {
  color: #10b981;
  background-color: #d1fae5;
}

.status-sold {
  color: #ef4444;
  background-color: #fee2e2;
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
</style>
