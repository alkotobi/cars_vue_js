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

const isDropdownOpen = ref({});
const props = defineProps({
  cars: Array,
  onEdit: {
    type: Function,
    default: () => {}
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
        has_vin: false
      }
    })
  }
})

// Add watcher for filters
watch(() => props.filters, (newFilters) => {
  fetchCarsStock()
}, { deep: true })

const user = ref(null)
const can_edit_cars_prop = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'can_edit_cars_prop')
})

const emit = defineEmits(['refresh'])

const { callApi, getFileUrl } = useApi()
const cars = ref([])
const loading = ref(true)
const error = ref(null)

// Add new refs and computed properties for VIN editing
const showVinEditForm = ref(false)
const selectedCar = ref(null)

const can_edit_vin = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'can_edit_vin')
})

// Add new refs and computed properties for file upload
const showFilesUploadForm = ref(false)

const can_upload_car_files = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'can_upload_car_files')
})

// Add new refs and computed properties for ports editing
const showPortsEditForm = ref(false)

const can_edit_cars_ports = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'can_edit_cars_ports')
})

// Add new refs and computed properties for money editing
const showMoneyEditForm = ref(false)

const can_edit_car_money = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'can_edit_car_money')
})

// Add new refs for warehouse form
const showWarehouseForm = ref(false)

// Add new computed property for warehouse permission
const can_edit_warehouse = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'can_edit_warehouse')
})

// Add to the data/refs section
const showDocumentsForm = ref(false)
const selectedCarForDocuments = ref(null)

// Add this computed property after other permission computeds
const can_edit_car_documents = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'can_edit_car_documents')
})

// Add new refs for load form
const showLoadForm = ref(false)
const selectedCarForLoad = ref(null)

// Add new computed property for load permission
const can_load_car = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'can_load_car')
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
        bb.bill_ref as buy_bill_ref
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
      params
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

const handleEdit = (car) => {
  props.onEdit(car)
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
const handleVINAction = (car) => {
  if (!can_edit_vin.value) {
    alert('You do not have permission to edit VIN numbers')
    return
  }
  selectedCar.value = car
  showVinEditForm.value = true
  isDropdownOpen.value[car.id] = false
}

const handleVinSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex(c => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for file upload
const handleFilesAction = (car) => {
  if (!can_upload_car_files.value) {
    alert('You do not have permission to upload car files')
    return
  }
  selectedCar.value = car
  showFilesUploadForm.value = true
  isDropdownOpen.value[car.id] = false
}

const handleFilesSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex(c => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for ports editing
const handlePortsAction = (car) => {
  if (!can_edit_cars_ports.value) {
    alert('You do not have permission to edit car ports')
    return
  }
  selectedCar.value = car
  showPortsEditForm.value = true
  isDropdownOpen.value[car.id] = false
}

const handlePortsSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex(c => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for money editing
const handleMoneyAction = (car) => {
  if (!can_edit_car_money.value) {
    alert('You do not have permission to edit car money fields')
    return
  }
  selectedCar.value = car
  showMoneyEditForm.value = true
  isDropdownOpen.value[car.id] = false
}

const handleMoneySave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex(c => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new method for warehouse action
const handleWarehouseAction = (car) => {
  if (!can_edit_warehouse.value) {
    alert('You do not have permission to manage warehouse')
    return
  }
  selectedCar.value = car
  showWarehouseForm.value = true
  isDropdownOpen.value[car.id] = false
}

const handleWarehouseSave = (updatedCar) => {
  // Update the car in the local array
  const index = cars.value.findIndex(c => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new method for documents action
const handleDocumentsClick = async (car) => {
  if (!can_edit_car_documents.value) {
    alert('You do not have permission to manage car documents')
    return
  }
  
  try {
    // Fetch the latest car data with all document fields
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.date_get_bl,
          cs.date_pay_freight,
          cs.date_get_documents_from_supp,
          cs.date_send_documents,
          cs.export_lisence_ref
        FROM cars_stock cs
        WHERE cs.id = ?
      `,
      params: [car.id]
    })
    
    if (result.success && result.data.length > 0) {
      // Merge the fetched document data with the existing car data
      selectedCarForDocuments.value = {
        ...car,
        ...result.data[0]
      }
      showDocumentsForm.value = true
    } else {
      throw new Error('Failed to fetch car document data')
    }
  } catch (err) {
    alert(err.message || 'An error occurred while fetching car document data')
  }
}

const handleDocumentsClose = () => {
  showDocumentsForm.value = false
  selectedCarForDocuments.value = null
}

const handleDocumentsSave = (updatedCar) => {
  // Update the car in the table data
  const index = cars.value.findIndex(c => c.id === updatedCar.id)
  if (index !== -1) {
    cars.value[index] = { ...cars.value[index], ...updatedCar }
  }
}

// Add new methods for load action
const handleLoadClick = (car) => {
  selectedCarForLoad.value = car
  showLoadForm.value = true
}

const handleLoadClose = () => {
  showLoadForm.value = false
  selectedCarForLoad.value = null
}

const handleLoadSave = (updatedCar) => {
  // Update the car in the table data
  const index = cars.value.findIndex(c => c.id === updatedCar.id)
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
  
  dropdowns.forEach(dropdown => {
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
  fetchCarsStock
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
        Advanced Filters
      </button>
    </div>
    
    <AdvancedFilter
      :show="showAdvancedFilter"
      :initial-filters="advancedFilters"
      @update:filters="handleAdvancedFilters"
      @close="showAdvancedFilter = false"
    />
    
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="cars.length === 0" class="empty-state">No cars in stock</div>
    
    <div v-else class="table-container">
      <table class="cars-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Car</th>
          <th>Color</th>
          <th>VIN</th>
          <th>Buy Bill Ref</th>
          <th>Loading Port</th>
          <th>Discharge Port</th>
          <th>Freight</th>
          <th>Price Cell</th>
          <th>Loading Date</th>
          <th>Status</th>
          <th>Client</th>
          <th>Documents</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in cars" :key="car.id">
          <td>{{ car.id }}</td>
          <td>{{ car.car_name || 'N/A' }}</td>
          <td>{{ car.color || 'N/A' }}</td>
          <td>{{ car.vin || 'N/A' }}</td>
          <td>{{ car.buy_bill_ref || 'N/A' }}</td>
          <td>{{ car.loading_port || 'N/A' }}</td>
          <td>{{ car.discharge_port || 'N/A' }}</td>
          <td>{{ car.freight ? '$' + car.freight : 'N/A' }}</td>
          <td>{{ car.price_cell ? '$' + car.price_cell : 'N/A' }}</td>
          <td>{{ car.date_loding ? new Date(car.date_loding).toLocaleDateString() : 'N/A' }}</td>
          <td :class="car.date_sell || car.client_name ? 'status-sold' : 'status-available'">
            {{ car.date_sell || car.client_name ? 'Sold' : 'Available' }}
          </td>
          <td>{{ car.client_name || 'N/A' }}</td>
            <td class="documents-cell">
            <div class="document-links">
                <a v-if="car.path_documents" 
                   :href="getFileUrl(car.path_documents)" 
                   target="_blank"
                   class="document-link"
                   title="View Documents">
                  <i class="fas fa-file-pdf"></i>
                  Documents
                </a>
                <a v-if="car.sell_pi_path" 
                   :href="getFileUrl(car.sell_pi_path)" 
                   target="_blank"
                   class="document-link"
                   title="View Sell PI">
                  <i class="fas fa-file-invoice-dollar"></i>
                  Sell PI
                </a>
                <a v-if="car.buy_pi_path" 
                   :href="getFileUrl(car.buy_pi_path)" 
                   target="_blank"
                   class="document-link"
                   title="View Buy PI">
                  <i class="fas fa-file-invoice"></i>
                  Buy PI
                </a>
            </div>
          </td>
          <td>
            <button :disabled="!can_edit_cars_prop" @click="handleEdit(car)" class="edit-btn">Edit</button>
            <div class="dropdown">
                <button @click="toggleDropdown(car.id, $event)" class="dropdown-toggle">Actions</button>
            <ul v-if="isDropdownOpen[car.id]" class="dropdown-menu">
                  <li>
                    <button 
                      @click="handleVINAction(car)"
                      :disabled="!can_edit_vin"
                      :class="{ 'disabled': !can_edit_vin }"
                    >
                      VIN
                    </button>
                  </li>
                  <li>
                    <button 
                      @click="handleFilesAction(car)"
                      :disabled="!can_upload_car_files"
                      :class="{ 'disabled': !can_upload_car_files }"
                    >
                      Files
                    </button>
                  </li>
                  <li>
                    <button 
                      @click="handlePortsAction(car)"
                      :disabled="!can_edit_cars_ports"
                      :class="{ 'disabled': !can_edit_cars_ports }"
                    >
                      Ports
                    </button>
                  </li>
                  <li>
                    <button 
                      @click="handleMoneyAction(car)"
                      :disabled="!can_edit_car_money"
                      :class="{ 'disabled': !can_edit_car_money }"
                    >
                      Money
                    </button>
                  </li>
                  <li>
                    <button 
                      @click="handleWarehouseAction(car)"
                      :disabled="!can_edit_warehouse"
                      :class="{ 'disabled': !can_edit_warehouse }"
                    >
                      Warehouse
                    </button>
                  </li>
                  <li>
                    <button 
                      @click="handleDocumentsClick(car)"
                      :disabled="!can_edit_car_documents"
                      :class="{ 'disabled': !can_edit_car_documents }"
                    >
                      <!-- <v-icon left>mdi-file-document-edit</v-icon> -->
                      Car Documents
                    </button>
                  </li>
                  <li>
                    <button 
                      @click="handleLoadClick(car)"
                      :disabled="!can_load_car"
                      :class="{ 'disabled': !can_load_car }"
                    >
                      Load
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
/* Add styles for frozen header */
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
}

.cars-table td {
  padding: 12px;
  border-bottom: 1px solid #e5e7eb;
  background-color: white;
}

/* Add styles for dropdown items */
.dropdown-menu li button.disabled {
  opacity: 0.6;
  cursor: not-allowed;
  color: #6b7280;
}

button:disabled {
  background-color: #cccccc;
  color: #666666;
  cursor: not-allowed;
}
.cars-stock-table {
  width: 100%;
  overflow-x: auto;
}

.loading, .error, .empty-state {
  padding: 20px;
  text-align: center;
}

.error {
  color: #ef4444;
}

.cars-table tbody tr:hover {
  background-color: #f5f5f5;
}

.status-available {
  color: #10b981;
  font-weight: 500;
}

.status-sold {
  color: #ef4444;
  font-weight: 500;
}

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
  transition: background-color 0.2s;
}

.document-link:hover {
  background-color: #f3f4f6;
  text-decoration: underline;
}

.document-link i {
  font-size: 1em;
  color: #ef4444;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 0.9em;
}

.edit-btn:hover {
  background-color: #2563eb;
}

.dropdown {
  position: relative;
  display: inline-block;
}

.dropdown-toggle {
  background-color: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 4px 8px;
  cursor: pointer;
}

.dropdown-menu {
  position: absolute;
  right: 0;
  background-color: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  list-style: none;
  padding: 8px 0;
  margin: 0;
  z-index: 1;
  min-width: 120px;
}

.dropdown-menu li {
  padding: 8px 16px;
}

.dropdown-menu li button {
  background: none;
  border: none;
  cursor: pointer;
  width: 100%;
  text-align: left;
}

.dropdown-menu li button:hover {
  background-color: #f5f5f5;
}

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
  transition: background-color 0.2s;
}

.advanced-filter-btn:hover {
  background-color: #e5e7eb;
}
</style>