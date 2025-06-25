<template>
  <div class="unassigned-cars-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-car"></i>
        Unassigned Cars
      </h3>
      <div class="header-actions" v-if="unassignedCars.length > 0">
        <button
          @click="toggleFilters"
          class="toggle-filters-btn"
          :class="{ active: showFilters }"
          title="Toggle Filters"
        >
          <i class="fas fa-filter"></i>
          <span>{{ showFilters ? 'Hide' : 'Show' }} Filters</span>
        </button>
        <button
          @click="refreshData"
          class="refresh-btn"
          :disabled="loading"
          :class="{ processing: loading }"
        >
          <i class="fas fa-sync-alt"></i>
          <span>Refresh</span>
          <i v-if="loading" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <!-- Filters Section -->
    <div v-if="showFilters" class="filters-section">
      <div class="filters-row">
        <div class="filter-group">
          <label for="carNameFilter">
            <i class="fas fa-car"></i>
            Car Name
          </label>
          <input
            type="text"
            id="carNameFilter"
            v-model="filters.carName"
            placeholder="Filter by car name..."
            @input="applyFilters"
          />
        </div>
        <div class="filter-group">
          <label for="colorFilter">
            <i class="fas fa-palette"></i>
            Color
          </label>
          <input
            type="text"
            id="colorFilter"
            v-model="filters.color"
            placeholder="Filter by color..."
            @input="applyFilters"
          />
        </div>
        <div class="filter-group">
          <label for="vinFilter">
            <i class="fas fa-barcode"></i>
            VIN
          </label>
          <input
            type="text"
            id="vinFilter"
            v-model="filters.vin"
            placeholder="Filter by VIN..."
            @input="applyFilters"
          />
        </div>
      </div>
      <div class="filters-row">
        <div class="filter-group">
          <label for="clientNameFilter">
            <i class="fas fa-user"></i>
            Client Name
          </label>
          <input
            type="text"
            id="clientNameFilter"
            v-model="filters.clientName"
            placeholder="Filter by client name..."
            @input="applyFilters"
          />
        </div>
        <div class="filter-group">
          <label for="clientIdFilter">
            <i class="fas fa-id-card"></i>
            Client ID No
          </label>
          <input
            type="text"
            id="clientIdFilter"
            v-model="filters.clientId"
            placeholder="Filter by client ID..."
            @input="applyFilters"
          />
        </div>
        <div class="filter-actions">
          <button @click="clearFilters" class="clear-filters-btn">
            <i class="fas fa-times"></i>
            Clear Filters
          </button>
        </div>
      </div>
    </div>

    <div class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading...</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="!props.selectedLoadedContainerId" class="empty-state">
        <i class="fas fa-mouse-pointer fa-2x"></i>
        <p>Please select a loading record above to view unassigned cars</p>
      </div>

      <div v-else-if="unassignedCars.length === 0" class="empty-state">
        <i class="fas fa-check-circle fa-2x"></i>
        <p>All cars are assigned to containers</p>
      </div>

      <div v-else class="table-content">
        <table class="unassigned-cars-table">
          <thead>
            <tr>
              <th @click="sortByColumn('id')" class="sortable-header">
                ID
                <i
                  v-if="sortBy === 'id'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('car_name')" class="sortable-header">
                Car Name
                <i
                  v-if="sortBy === 'car_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('color')" class="sortable-header">
                Color
                <i
                  v-if="sortBy === 'color'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('vin')" class="sortable-header">
                VIN
                <i
                  v-if="sortBy === 'vin'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('client_name')" class="sortable-header">
                Client
                <i
                  v-if="sortBy === 'client_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('sell_bill_id')" class="sortable-header">
                Sell Bill ID
                <i
                  v-if="sortBy === 'sell_bill_id'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('sell_bill_date')" class="sortable-header">
                Sell Bill Date
                <i
                  v-if="sortBy === 'sell_bill_date'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('sell_bill_ref')" class="sortable-header">
                Sell Bill Ref
                <i
                  v-if="sortBy === 'sell_bill_ref'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="car in sortedAndFilteredCars"
              :key="car.id"
              class="table-row"
              @click="handleCarClick(car)"
            >
              <td class="id-cell">#{{ car.id }}</td>
              <td class="car-name-cell">{{ car.car_name || '-' }}</td>
              <td class="color-cell">{{ car.color || '-' }}</td>
              <td class="vin-cell">{{ car.vin || '-' }}</td>
              <td class="client-cell">
                <div v-if="car.id_client && car.id_copy_path" class="client-info">
                  <img
                    :src="getFileUrl(car.id_copy_path)"
                    :alt="'ID for ' + (car.client_name || 'Client')"
                    class="client-id-image"
                    @click.stop="openClientId(car.id_copy_path)"
                  />
                  <div class="client-details">
                    <span class="client-name">{{ car.client_name }}</span>
                    <span class="client-id-no">{{ car.client_id_no || 'No ID' }}</span>
                  </div>
                </div>
                <span v-else class="no-client">-</span>
              </td>
              <td class="sell-bill-id-cell">{{ car.sell_bill_id || '-' }}</td>
              <td class="sell-bill-date-cell">{{ formatDate(car.sell_bill_date) }}</td>
              <td class="sell-bill-ref-cell">{{ car.sell_bill_ref || '-' }}</td>
              <td class="actions-cell">
                <div class="action-buttons">
                  <button
                    @click.stop="assignCar(car)"
                    class="action-btn assign-btn"
                    :title="getAssignButtonTitle()"
                    :disabled="isAssigning || props.selectedContainerOnBoard"
                  >
                    <i class="fas fa-link"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="table-footer">
      <span class="record-count">
        Showing {{ sortedAndFilteredCars.length }} of {{ allUnassignedCars.length }} unassigned cars
        <span v-if="hasActiveFilters" class="filter-indicator">(filtered)</span>
      </span>
    </div>

    <!-- VIN Input Dialog -->
    <div v-if="showVinDialog" class="dialog-overlay" @click.self="closeVinDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-car"></i>
            Enter VIN for Car #{{ carToAssign?.id }}
          </h3>
          <button class="close-btn" @click="closeVinDialog" :disabled="isSubmittingVin">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="car-info">
            <p><strong>Car:</strong> {{ carToAssign?.car_name || 'N/A' }}</p>
            <p><strong>Color:</strong> {{ carToAssign?.color || 'N/A' }}</p>
            <p><strong>Client:</strong> {{ carToAssign?.client_name || 'N/A' }}</p>
          </div>

          <div class="form-group">
            <label for="vin-input">
              <i class="fas fa-barcode"></i>
              VIN Number <span style="color: red">*</span>
            </label>
            <input
              type="text"
              id="vin-input"
              v-model="vinInput"
              placeholder="Enter VIN number"
              :disabled="isSubmittingVin"
              @keyup.enter="submitVinAndAssign"
              maxlength="17"
              required
            />
            <div v-if="vinError" class="error-message">{{ vinError }}</div>
          </div>

          <div class="warning-message">
            <i class="fas fa-exclamation-triangle"></i>
            <p><strong>Note:</strong> VIN is required before assigning a car to a container.</p>
          </div>
        </div>

        <div class="dialog-footer">
          <button
            type="button"
            @click="closeVinDialog"
            class="cancel-btn"
            :disabled="isSubmittingVin"
          >
            Cancel
          </button>
          <button
            type="button"
            @click="submitVinAndAssign"
            class="confirm-btn"
            :disabled="isSubmittingVin"
            :class="{ processing: isSubmittingVin }"
          >
            <i v-if="isSubmittingVin" class="fas fa-spinner fa-spin"></i>
            <span>Update VIN & Assign</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '@/composables/useApi'

const props = defineProps({
  selectedLoadedContainerId: {
    type: Number,
    default: null,
  },
  selectedContainerOnBoard: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['car-assigned'])

const { callApi, getFileUrl } = useApi()

const unassignedCars = ref([])
const allUnassignedCars = ref([])
const loading = ref(false)
const error = ref(null)
const selectedCarId = ref(null)
const isAssigning = ref(false)
const filters = ref({
  carName: '',
  color: '',
  vin: '',
  clientName: '',
  clientId: '',
})

// VIN dialog state
const showVinDialog = ref(false)
const vinInput = ref('')
const carToAssign = ref(null)
const isSubmittingVin = ref(false)
const vinError = ref('')

// Sorting state
const sortBy = ref('id')
const sortOrder = ref('asc')

// Filter visibility state
const showFilters = ref(false)

// Computed property to check if any filters are active
const hasActiveFilters = computed(() => {
  return (
    filters.value.carName ||
    filters.value.color ||
    filters.value.vin ||
    filters.value.clientName ||
    filters.value.clientId
  )
})

// Computed property for sorted and filtered cars
const sortedAndFilteredCars = computed(() => {
  let filtered = allUnassignedCars.value

  // Apply filters
  if (filters.value.carName) {
    filtered = filtered.filter((car) =>
      car.car_name?.toLowerCase().includes(filters.value.carName.toLowerCase()),
    )
  }
  if (filters.value.color) {
    filtered = filtered.filter((car) =>
      car.color?.toLowerCase().includes(filters.value.color.toLowerCase()),
    )
  }
  if (filters.value.vin) {
    filtered = filtered.filter((car) =>
      car.vin?.toLowerCase().includes(filters.value.vin.toLowerCase()),
    )
  }
  if (filters.value.clientName) {
    filtered = filtered.filter((car) =>
      car.client_name?.toLowerCase().includes(filters.value.clientName.toLowerCase()),
    )
  }
  if (filters.value.clientId) {
    filtered = filtered.filter((car) =>
      car.client_id_no?.toLowerCase().includes(filters.value.clientId.toLowerCase()),
    )
  }

  // Apply sorting
  return filtered.sort((a, b) => {
    let aVal = a[sortBy.value]
    let bVal = b[sortBy.value]

    // Handle numeric values
    if (['id', 'price_cell', 'sell_bill_id'].includes(sortBy.value)) {
      aVal = parseFloat(aVal) || 0
      bVal = parseFloat(bVal) || 0
    }

    // Handle null values
    if (aVal == null) return 1
    if (bVal == null) return -1

    // Compare values
    if (sortOrder.value === 'asc') {
      return aVal > bVal ? 1 : aVal < bVal ? -1 : 0
    } else {
      return aVal < bVal ? 1 : aVal > bVal ? -1 : 0
    }
  })
})

const fetchUnassignedCars = async () => {
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT
          cs.id,
          cn.car_name,
          c.color,
          cs.vin,
          cs.price_cell,
          cs.id_sell,
          cs.date_sell,
          cs.id_client,
          cl.name as client_name,
          cl.id_copy_path,
          cl.id_no as client_id_no,
          sb.id as sell_bill_id,
          sb.date_sell as sell_bill_date,
          sb.bill_ref as sell_bill_ref
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors c ON bd.id_color = c.id
        LEFT JOIN clients cl ON cs.id_client = cl.id
        LEFT JOIN sell_bill sb ON cs.id_sell = sb.id
        WHERE cs.id_loaded_container IS NULL
        AND cs.date_loding IS NULL
        AND cs.container_ref IS NULL
        AND cs.id_client IS NOT NULL
        ORDER BY cs.id DESC
      `,
      params: [],
    })

    if (result.success) {
      unassignedCars.value = result.data || []
      allUnassignedCars.value = result.data || []
      applyFilters() // Apply any existing filters
    } else {
      error.value = result.error || 'Failed to load unassigned cars'
    }
  } catch (err) {
    error.value = 'Error loading unassigned cars: ' + err.message
  } finally {
    loading.value = false
  }
}

const handleCarClick = (car) => {
  console.log('Unassigned car clicked:', car.id)
  selectedCarId.value = car.id
}

const openClientId = (path) => {
  window.open(getFileUrl(path), '_blank')
}

const getStatusClass = (idSell) => {
  return idSell ? 'status-sold' : 'status-available'
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

const sortByColumn = (column) => {
  if (sortBy.value === column) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = column
    sortOrder.value = 'asc'
  }
}

const refreshData = () => {
  fetchUnassignedCars()
}

const assignCar = async (car) => {
  if (isAssigning.value) return

  if (!props.selectedLoadedContainerId) {
    alert('Please select a container first')
    return
  }

  console.log('Assigning car:', car)
  console.log('Car VIN:', car.vin)
  console.log('VIN check result:', !car.vin || car.vin.trim() === '')

  // Check if car has VIN
  if (!car.vin || car.vin.trim() === '') {
    console.log('VIN is missing, showing dialog')
    // Show VIN input dialog
    carToAssign.value = car
    vinInput.value = ''
    vinError.value = ''
    showVinDialog.value = true
    console.log('Dialog should be visible:', showVinDialog.value)
    return
  }

  console.log('VIN exists, proceeding with normal assignment')

  // Check if the selected container is already on board
  try {
    const containerResult = await callApi({
      query: 'SELECT date_on_board, ref_container FROM loaded_containers WHERE id = ?',
      params: [props.selectedLoadedContainerId],
    })

    if (containerResult.success && containerResult.data && containerResult.data.length > 0) {
      const container = containerResult.data[0]
      if (container.date_on_board) {
        alert(
          `Cannot assign car to this container. Container is already on board since ${container.date_on_board}`,
        )
        return
      }
    }
  } catch (err) {
    error.value = 'Error checking container status: ' + err.message
    return
  }

  if (!confirm(`Are you sure you want to assign car #${car.id} to this container?`)) {
    return
  }

  await performCarAssignment(car)
}

const performCarAssignment = async (car) => {
  isAssigning.value = true

  try {
    // Get the container reference first
    const containerResult = await callApi({
      query: 'SELECT ref_container FROM loaded_containers WHERE id = ?',
      params: [props.selectedLoadedContainerId],
    })

    if (!containerResult.success || !containerResult.data || containerResult.data.length === 0) {
      error.value = 'Failed to get container reference'
      return
    }

    const containerRef = containerResult.data[0].ref_container

    const result = await callApi({
      query: 'UPDATE cars_stock SET id_loaded_container = ?, container_ref = ? WHERE id = ?',
      params: [props.selectedLoadedContainerId, containerRef, car.id],
    })

    if (result.success) {
      // Refresh the data to update both tables
      await fetchUnassignedCars()
      // Emit an event to refresh assigned cars table
      emit('car-assigned')
    } else {
      error.value = result.error || 'Failed to assign car'
    }
  } catch (err) {
    error.value = 'Error assigning car: ' + err.message
  } finally {
    isAssigning.value = false
  }
}

const submitVinAndAssign = async () => {
  if (isSubmittingVin.value) return

  // Validate VIN input
  if (!vinInput.value || vinInput.value.trim() === '') {
    vinError.value = 'VIN is required'
    return
  }

  vinError.value = ''
  isSubmittingVin.value = true

  try {
    const vinValue = vinInput.value.trim()

    // Update the car's VIN first
    const updateResult = await callApi({
      query: 'UPDATE cars_stock SET vin = ? WHERE id = ?',
      params: [vinValue, carToAssign.value.id],
    })

    if (!updateResult.success) {
      vinError.value = 'Failed to update VIN'
      return
    }

    // Store the car data with updated VIN
    const updatedCar = {
      ...carToAssign.value,
      vin: vinValue,
    }

    // Close the dialog
    showVinDialog.value = false
    carToAssign.value = null
    vinInput.value = ''

    // Now assign the car
    await performCarAssignment(updatedCar)
  } catch (err) {
    vinError.value = 'Error updating VIN: ' + err.message
  } finally {
    isSubmittingVin.value = false
  }
}

const closeVinDialog = () => {
  showVinDialog.value = false
  carToAssign.value = null
  vinInput.value = ''
  vinError.value = ''
  isSubmittingVin.value = false
}

const getAssignButtonTitle = () => {
  if (props.selectedContainerOnBoard) {
    return 'Container is already on board'
  }
  return 'Assign to Container'
}

const applyFilters = () => {
  let filteredCars = [...allUnassignedCars.value]

  // Filter by car name
  if (filters.value.carName) {
    filteredCars = filteredCars.filter(
      (car) =>
        car.car_name && car.car_name.toLowerCase().includes(filters.value.carName.toLowerCase()),
    )
  }

  // Filter by color
  if (filters.value.color) {
    filteredCars = filteredCars.filter(
      (car) => car.color && car.color.toLowerCase().includes(filters.value.color.toLowerCase()),
    )
  }

  // Filter by VIN
  if (filters.value.vin) {
    filteredCars = filteredCars.filter(
      (car) => car.vin && car.vin.toLowerCase().includes(filters.value.vin.toLowerCase()),
    )
  }

  // Filter by client name
  if (filters.value.clientName) {
    filteredCars = filteredCars.filter(
      (car) =>
        car.client_name &&
        car.client_name.toLowerCase().includes(filters.value.clientName.toLowerCase()),
    )
  }

  // Filter by client ID number
  if (filters.value.clientId) {
    filteredCars = filteredCars.filter(
      (car) =>
        car.client_id_no &&
        car.client_id_no.toLowerCase().includes(filters.value.clientId.toLowerCase()),
    )
  }

  unassignedCars.value = filteredCars
}

const clearFilters = () => {
  filters.value = {
    carName: '',
    color: '',
    vin: '',
    clientName: '',
    clientId: '',
  }
  unassignedCars.value = [...allUnassignedCars.value]
}

const toggleFilters = () => {
  showFilters.value = !showFilters.value
}

onMounted(() => {
  fetchUnassignedCars()
})

// Expose methods to parent component
defineExpose({
  refreshData,
  fetchUnassignedCars,
})
</script>

<style scoped>
.unassigned-cars-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  margin-top: 16px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
}

.table-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.1rem;
}

.table-header h3 i {
  color: #f59e0b;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.toggle-filters-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  background-color: #f8fafc;
  color: #374151;
  border: 1px solid #d1d5db;
}

.toggle-filters-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.toggle-filters-btn.active {
  background-color: #dbeafe;
  color: #1d4ed8;
  border: 1px solid #93c5fd;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  background-color: #f8fafc;
  color: #374151;
  border: 1px solid #d1d5db;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.refresh-btn.processing {
  opacity: 0.7;
}

.loading-indicator {
  margin-left: 4px;
}

.table-wrapper {
  position: relative;
  max-height: 400px;
  overflow-y: auto;
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
  gap: 16px;
  z-index: 10;
}

.loading-overlay i {
  color: #3b82f6;
}

.loading-overlay span {
  color: #6b7280;
  font-weight: 500;
}

.error-message {
  padding: 20px;
  text-align: center;
  color: #dc2626;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  margin: 16px;
}

.error-message i {
  margin-right: 8px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: #6b7280;
}

.empty-state i {
  margin-bottom: 16px;
  color: #10b981;
}

.empty-state p {
  margin: 0;
  font-size: 1rem;
}

.table-content {
  overflow-x: auto;
}

.unassigned-cars-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.unassigned-cars-table th {
  background-color: #f8fafc;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
  position: sticky;
  top: 0;
  z-index: 5;
}

.unassigned-cars-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f3f4f6;
  vertical-align: middle;
}

.table-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.table-row:hover {
  background-color: #f9fafb;
}

.table-row:active {
  background-color: #f3f4f6;
}

.id-cell {
  font-weight: 600;
  color: #6b7280;
  font-family: monospace;
}

.car-name-cell {
  font-weight: 500;
  color: #1f2937;
}

.color-cell {
  color: #374151;
}

.vin-cell {
  font-family: monospace;
  color: #6b7280;
  font-size: 0.85rem;
}

.client-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.client-id-image {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid #e5e7eb;
  cursor: pointer;
  transition: all 0.2s ease;
}

.client-id-image:hover {
  transform: scale(1.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.client-details {
  display: flex;
  flex-direction: column;
}

.client-name {
  font-weight: 500;
  color: #374151;
  font-size: 0.9rem;
}

.client-id-no {
  color: #6b7280;
  font-size: 0.8rem;
}

.no-client {
  color: #6b7280;
  font-style: italic;
}

.price-cell {
  font-weight: 600;
  color: #059669;
}

.status-cell {
  text-align: center;
}

.status-available {
  background-color: #d1fae5;
  color: #065f46;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-sold {
  background-color: #fee2e2;
  color: #991b1b;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.sell-bill-id-cell {
  text-align: center;
  font-size: 0.9rem;
  color: #6b7280;
}

.sell-bill-date-cell {
  text-align: center;
  font-size: 0.9rem;
  color: #6b7280;
}

.sell-bill-ref-cell {
  text-align: center;
  font-size: 0.9rem;
  color: #6b7280;
  font-family: monospace;
}

.actions-cell {
  text-align: center;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.9rem;
}

.action-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.assign-btn {
  background-color: #dbeafe;
  color: #1d4ed8;
  border: 1px solid #93c5fd;
}

.assign-btn:hover:not(:disabled) {
  background-color: #bfdbfe;
  color: #1e40af;
}

.assign-btn:disabled {
  background-color: #e5e7eb;
  color: #9ca3af;
  border-color: #d1d5db;
  cursor: not-allowed;
}

.assign-btn:disabled:hover {
  background-color: #e5e7eb;
  color: #9ca3af;
}

.table-footer {
  padding: 12px 20px;
  border-top: 1px solid #e5e7eb;
  background-color: #f8fafc;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.record-count {
  color: #6b7280;
  font-size: 0.9rem;
}

.filters-section {
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
}

.filters-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  font-weight: 600;
  color: #374151;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 6px;
}

.filter-group label i {
  color: #6b7280;
}

.filter-group input {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  transition: border-color 0.2s;
}

.filter-group input:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.filter-actions {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.clear-filters-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  background-color: #f8fafc;
  color: #374151;
  border: 1px solid #d1d5db;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 6px;
}

.clear-filters-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.clear-filters-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.filter-indicator {
  color: #3498db;
  font-weight: 500;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.dialog {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  max-width: 400px;
  width: 100%;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.dialog-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.1rem;
}

.dialog-header h3 i {
  color: #f59e0b;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
}

.dialog-content {
  margin-bottom: 16px;
}

.car-info {
  margin-bottom: 16px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #374151;
}

.form-group input {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
}

.error-message {
  color: #dc2626;
  font-size: 0.8rem;
}

.warning-message {
  background-color: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 16px;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.warning-message i {
  color: #d97706;
  margin-top: 2px;
}

.warning-message p {
  margin: 0;
  font-size: 0.9rem;
  color: #92400e;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.cancel-btn,
.confirm-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.9rem;
}

.cancel-btn {
  background-color: #f8fafc;
  color: #374151;
}

.confirm-btn {
  background-color: #3498db;
  color: white;
}

.confirm-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.confirm-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.confirm-btn.processing {
  opacity: 0.7;
}
</style>
