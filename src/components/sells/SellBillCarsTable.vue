<script setup>
import { ref, watch, onMounted, computed, useAttrs, nextTick } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import { useRouter } from 'vue-router'
import { ElSelect, ElOption } from 'element-plus'
import 'element-plus/dist/index.css'
import SellBillPrintOption from './SellBillPrintOption.vue'
import NotesTable from '../shared/NotesTable.vue'
import NotesManagementModal from '../shared/NotesManagementModal.vue'
import CarUpgradesModal from '../shared/CarUpgradesModal.vue'

const { t } = useEnhancedI18n()
const attrs = useAttrs()
const props = defineProps({
  sellBillId: {
    type: Number,
    default: null,
  },
})

const emit = defineEmits(['refresh'])
const router = useRouter()
const user = ref(null)
const expandedCarNotes = ref(new Set()) // Track which cars have expanded notes
const isAdmin = computed(() => user.value?.role_id === 1)
const can_assign_to_tmp_clients = computed(
  () =>
    user.value?.permissions?.some((p) => p.permission_name === 'can_assign_to_tmp_clients') ||
    isAdmin.value,
)

// Add permission check for unassigning cars
const can_unassign_cars = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_unassign_cars')
})
const { callApi } = useApi()
const cars = ref([])
const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)

// Store upgrades totals for each car (carId -> total)
const carUpgradesTotals = ref(new Map())

// Add edit dialog state and form data
const showEditDialog = ref(false)
const editingCar = ref(null)
const clients = ref([])
const dischargePorts = ref([])
const editFormData = ref({
  id_client: null,
  id_port_discharge: null,
  price_cell: null,
  freight: null,
  rate: null,
  notes: null,
  is_tmp_client: 0,
})

// Notes management for edit dialog
const editCarNotes = ref([])
const originalEditCarNotes = ref([]) // Store original notes to restore on cancel
const allUsers = ref([])
const showEditNotesModal = ref(false)

// Upgrades management for edit dialog
const showEditUpgradesModal = ref(false)
const editCarUpgrades = ref([])
const editCarUpgradesTotal = ref(0)

// Upgrades modal for table (clicking upgrades badge)
const showTableUpgradesModal = ref(false)
const selectedCarForUpgrades = ref(null)

// Add print options dialog state
const showPrintOptions = ref(false)
const selectedCarForPrint = ref(null)

const filteredClients = ref([])
const filteredPorts = ref([])

// Add CFR DA calculation functions
const cfrDaInput = ref(null)

// Add currency selector
const selectedCurrency = ref('DZD') // 'USD' or 'DZD'
const priceInput = ref(null) // Single price input value

const calculatePriceFromCFRDA = (cfrDa, rate, freight, upgrades = 0) => {
  if (!cfrDa || !rate) return null
  const parsedCfrDa = parseFloat(cfrDa)
  const parsedRate = parseFloat(rate)
  const parsedFreight = parseFloat(freight) || 0
  const parsedUpgrades = parseFloat(upgrades) || 0
  // FOB USD = (CFR_DA / rate) - freight + upgrades
  // Upgrades ADD value, so they increase the FOB price
  return (parsedCfrDa / parsedRate - parsedFreight + parsedUpgrades).toFixed(2)
}

const calculateCFRDAFromPrice = (price, rate, freight, upgrades = 0) => {
  if (!price || !rate) return null
  const parsedPrice = parseFloat(price)
  const parsedRate = parseFloat(rate)
  const parsedFreight = parseFloat(freight) || 0
  const parsedUpgrades = parseFloat(upgrades) || 0
  // CFR DA = (price + upgrades + freight) × rate
  // Upgrades ADD value, so they increase both FOB and CFR
  // CFR = (FOB + upgrades + freight) × rate
  const result = ((parsedPrice + parsedUpgrades + parsedFreight) * parsedRate).toFixed(2)
  console.log('[DEBUG] calculateCFRDAFromPrice - Input:', { price: parsedPrice, upgrades: parsedUpgrades, freight: parsedFreight, rate: parsedRate })
  console.log('[DEBUG] calculateCFRDAFromPrice - Calculation:', `(${parsedPrice} + ${parsedUpgrades} + ${parsedFreight}) × ${parsedRate} = ${result}`)
  return result
}

// Handle currency change
const handleCurrencyChange = () => {
  // Recalculate the other currency when currency changes
  if (priceInput.value && editFormData.value.rate) {
    const upgrades = editCarUpgradesTotal.value || 0
    
    if (selectedCurrency.value === 'USD') {
      // If switching to USD, convert current DZD to USD
      if (cfrDaInput.value) {
        priceInput.value = calculatePriceFromCFRDA(
          cfrDaInput.value,
          editFormData.value.rate,
          editFormData.value.freight,
          upgrades,
        )
        editFormData.value.price_cell = priceInput.value
      }
    } else {
      // If switching to DZD, convert current USD to DZD
      if (editFormData.value.price_cell) {
        priceInput.value = calculateCFRDAFromPrice(
          editFormData.value.price_cell,
          editFormData.value.rate,
          editFormData.value.freight,
          upgrades,
        )
        cfrDaInput.value = priceInput.value
      }
    }
  }
}

// Handle price input change based on selected currency
const handlePriceInputChange = () => {
  if (!priceInput.value || !editFormData.value.rate) return

  const upgrades = editCarUpgradesTotal.value || 0

  if (selectedCurrency.value === 'USD') {
    // User entered USD, calculate DZD (CFR DA)
    editFormData.value.price_cell = priceInput.value
    cfrDaInput.value = calculateCFRDAFromPrice(
      priceInput.value,
      editFormData.value.rate,
      editFormData.value.freight,
      upgrades,
    )
  } else {
    // User entered DZD, calculate USD (price_cell)
    cfrDaInput.value = priceInput.value
    editFormData.value.price_cell = calculatePriceFromCFRDA(
      priceInput.value,
      editFormData.value.rate,
      editFormData.value.freight,
      upgrades,
    )
  }
}

// Handle rate or freight change - recalculate based on selected currency
const handleRateOrFreightChange = () => {
  if (!priceInput.value || !editFormData.value.rate) return

  const upgrades = editCarUpgradesTotal.value || 0

  if (selectedCurrency.value === 'USD') {
    // Recalculate DZD from USD
    cfrDaInput.value = calculateCFRDAFromPrice(
      priceInput.value,
      editFormData.value.rate,
      editFormData.value.freight,
      upgrades,
    )
  } else {
    // Recalculate USD from DZD
    editFormData.value.price_cell = calculatePriceFromCFRDA(
      priceInput.value,
      editFormData.value.rate,
      editFormData.value.freight,
      upgrades,
    )
  }
}

// Legacy handlers for backwards compatibility
const handlePriceChange = (event) => {
  // This is now handled by handlePriceInputChange
}

const handleCFRDAChange = (event) => {
  // This is now handled by handlePriceInputChange
}

// Function to unassign a car from the sell bill
const unassignCar = async (carId) => {
  // Check permission first
  if (!can_unassign_cars.value) {
    error.value = t('sellBills.no_permission_to_unassign_car')
    return
  }

  if (!confirm(t('sellBills.confirm_unassign_car_message'))) {
    return
  }

  isProcessing.value = true
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = NULL, 
            id_client = NULL, 
            id_port_discharge = NULL, 
            freight = NULL,
            id_sell_pi = NULL,
            is_tmp_client = 0,
            is_batch = 0
        WHERE id = ?
      `,
      params: [carId],
    })

    if (result.success) {
      // Refresh the cars list for this component
      await fetchCarsByBillId(props.sellBillId)

      // Emit refresh event to parent to update other components
      emit('refresh')
    } else {
      error.value = result.error || t('sellBills.failed_to_unassign_car')
    }
  } catch (err) {
    error.value = err.message || t('sellBills.error_occurred')
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

// Update client search method
const remoteClientSearch = (query) => {
  if (query) {
    filteredClients.value = clients.value.filter((client) =>
      client.name.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    filteredClients.value = clients.value
  }
}

// Update port search method
const remotePortSearch = (query) => {
  if (query) {
    filteredPorts.value = dischargePorts.value.filter((port) =>
      port.discharge_port.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    filteredPorts.value = dischargePorts.value
  }
}

// Update fetchClients to set filtered clients initially
const fetchClients = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name, mobiles
        FROM clients
        ORDER BY name ASC
      `,
      params: [],
    })

    if (result.success) {
      clients.value = result.data
      filteredClients.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch clients'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Update fetchDischargePorts to set filtered ports initially
const fetchDischargePorts = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, discharge_port
        FROM discharge_ports
        ORDER BY discharge_port ASC
      `,
      params: [],
    })

    if (result.success) {
      dischargePorts.value = result.data
      filteredPorts.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch discharge ports'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Fetch all users for notes display
const fetchAllUsers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, username
        FROM users
        ORDER BY username ASC
      `,
      params: [],
    })

    if (result.success) {
      allUsers.value = result.data
    }
  } catch (err) {
    console.error('Error fetching users:', err)
  }
}

// Save car notes function for NotesManagementModal
const saveCarNotes = async (notesJson, carId) => {
  const result = await callApi({
    query: 'UPDATE cars_stock SET notes = ? WHERE id = ?',
    params: [notesJson, carId],
    requiresAuth: true,
  })
  
  if (!result.success) {
    throw new Error(result.error || 'Failed to save notes')
  }
  
  // Update local notes
  if (notesJson) {
    editCarNotes.value = JSON.parse(notesJson)
  } else {
    editCarNotes.value = []
  }
}

// Handle notes updated from modal
const handleEditNotesUpdated = (updatedNotes) => {
  editCarNotes.value = updatedNotes
}

// Handle manage notes click
const handleManageEditNotes = () => {
  showEditNotesModal.value = true
}

// Handle upgrades click in edit dialog
const handleEditUpgrades = () => {
  showEditUpgradesModal.value = true
}

// Recalculate prices when upgrades change
const recalculatePricesWithUpgrades = () => {
  console.log('[DEBUG] recalculatePricesWithUpgrades called')
  console.log('[DEBUG] priceInput.value:', priceInput.value)
  console.log('[DEBUG] editFormData.value.rate:', editFormData.value.rate)
  console.log('[DEBUG] editingCar.value?.id:', editingCar.value?.id)
  console.log('[DEBUG] editCarUpgradesTotal.value:', editCarUpgradesTotal.value)
  
  if (!priceInput.value || !editFormData.value.rate || !editingCar.value?.id) {
    console.log('[DEBUG] Recalculation skipped - missing required data')
    return
  }
  
  const upgrades = editCarUpgradesTotal.value || 0
  console.log('[DEBUG] Using upgrades value:', upgrades)
  console.log('[DEBUG] selectedCurrency.value:', selectedCurrency.value)
  console.log('[DEBUG] editFormData.value.freight:', editFormData.value.freight)
  
  if (selectedCurrency.value === 'USD') {
    // User has USD entered, recalculate CFR DA with new upgrades
    editFormData.value.price_cell = priceInput.value
    const calculatedCFR = calculateCFRDAFromPrice(
      priceInput.value,
      editFormData.value.rate,
      editFormData.value.freight,
      upgrades,
    )
    console.log('[DEBUG] Calculated CFR DA from USD:', calculatedCFR)
    console.log('[DEBUG] Input values - price:', priceInput.value, 'rate:', editFormData.value.rate, 'freight:', editFormData.value.freight, 'upgrades:', upgrades)
    cfrDaInput.value = calculatedCFR
  } else {
    // User has DZD entered, recalculate price_cell with new upgrades
    cfrDaInput.value = priceInput.value
    const calculatedPrice = calculatePriceFromCFRDA(
      priceInput.value,
      editFormData.value.rate,
      editFormData.value.freight,
      upgrades,
    )
    console.log('[DEBUG] Calculated price_cell from DZD:', calculatedPrice)
    console.log('[DEBUG] Input values - cfrDa:', priceInput.value, 'rate:', editFormData.value.rate, 'freight:', editFormData.value.freight, 'upgrades:', upgrades)
    editFormData.value.price_cell = calculatedPrice
  }
  console.log('[DEBUG] Final values - price_cell:', editFormData.value.price_cell, 'cfr_da:', cfrDaInput.value)
}

// Close upgrades modal in edit dialog
const closeEditUpgradesModal = async () => {
  console.log('[DEBUG] closeEditUpgradesModal called')
  showEditUpgradesModal.value = false
  // Refresh upgrades total when modal closes (in case upgrades were modified)
  if (editingCar.value?.id) {
    console.log('[DEBUG] Refreshing upgrades for car:', editingCar.value.id)
    // Small delay to ensure any pending database operations complete
    await new Promise(resolve => setTimeout(resolve, 100))
    await fetchEditCarUpgrades()
    console.log('[DEBUG] Upgrades refreshed, total:', editCarUpgradesTotal.value)
    // Also refresh upgrades in the table
    fetchCarsUpgrades()
    
    // Recalculate prices with updated upgrades
    // Use nextTick to ensure DOM is updated
    await nextTick()
    console.log('[DEBUG] Calling recalculatePricesWithUpgrades after modal close')
    recalculatePricesWithUpgrades()
  } else {
    console.log('[DEBUG] No car ID, skipping upgrade refresh')
  }
}

// Handle upgrade added event from modal
const handleUpgradeAdded = async () => {
  console.log('[DEBUG] handleUpgradeAdded called')
  // Refresh upgrades when a new one is added
  if (editingCar.value?.id) {
    console.log('[DEBUG] Refreshing upgrades after upgrade added for car:', editingCar.value.id)
    // Small delay to ensure database operation completes
    await new Promise(resolve => setTimeout(resolve, 100))
    await fetchEditCarUpgrades()
    console.log('[DEBUG] Upgrades refreshed after add, total:', editCarUpgradesTotal.value)
    fetchCarsUpgrades()
    // Recalculate prices with updated upgrades
    await nextTick()
    console.log('[DEBUG] Calling recalculatePricesWithUpgrades after upgrade added')
    recalculatePricesWithUpgrades()
  } else {
    console.log('[DEBUG] No car ID, skipping upgrade refresh')
  }
}

// Fetch upgrades for the car being edited
const fetchEditCarUpgrades = async () => {
  console.log('[DEBUG] fetchEditCarUpgrades called for car ID:', editingCar.value?.id)
  if (!editingCar.value?.id) {
    console.log('[DEBUG] No car ID, resetting upgrades')
    editCarUpgrades.value = []
    editCarUpgradesTotal.value = 0
    return
  }

  try {
    const result = await callApi({
      query: `
        SELECT 
          ca.id,
          ca.value
        FROM car_apgrades ca
        WHERE ca.id_car = ?
      `,
      params: [editingCar.value.id],
    })

    if (result.success) {
      editCarUpgrades.value = result.data || []
      console.log('[DEBUG] Fetched upgrades:', editCarUpgrades.value)
      // Calculate total
      editCarUpgradesTotal.value = editCarUpgrades.value.reduce((sum, upgrade) => {
        const value = parseFloat(upgrade.value) || 0
        return sum + value
      }, 0)
      console.log('[DEBUG] Calculated upgrade total:', editCarUpgradesTotal.value)
    } else {
      console.log('[DEBUG] Failed to fetch upgrades:', result.error)
      editCarUpgrades.value = []
      editCarUpgradesTotal.value = 0
    }
  } catch (err) {
    console.error('[DEBUG] Error fetching car upgrades:', err)
    editCarUpgrades.value = []
    editCarUpgradesTotal.value = 0
  }
}

// Format value for display
const formatUpgradeValue = (value) => {
  if (value === null || value === undefined || value === 0) return '0.00'
  return parseFloat(value).toFixed(2)
}

// Fetch upgrades for all cars in the table
const fetchCarsUpgrades = async () => {
  if (!cars.value || cars.value.length === 0) {
    carUpgradesTotals.value.clear()
    return
  }

  try {
    const carIds = cars.value.map(car => car.id)
    if (carIds.length === 0) return

    const result = await callApi({
      query: `
        SELECT 
          ca.id_car,
          SUM(ca.value) as total_upgrades
        FROM car_apgrades ca
        WHERE ca.id_car IN (${carIds.map(() => '?').join(',')})
        GROUP BY ca.id_car
      `,
      params: carIds,
    })

    if (result.success) {
      // Clear existing totals
      carUpgradesTotals.value.clear()
      // Store totals for each car
      result.data.forEach((row) => {
        carUpgradesTotals.value.set(row.id_car, parseFloat(row.total_upgrades) || 0)
      })
    }
  } catch (err) {
    console.error('Error fetching cars upgrades:', err)
  }
}

// Get upgrades total for a car
const getCarUpgradesTotal = (carId) => {
  return carUpgradesTotals.value.get(carId) || 0
}

// Handle clicking upgrades badge in table
const handleTableUpgradesClick = (car) => {
  selectedCarForUpgrades.value = car
  showTableUpgradesModal.value = true
}

// Close upgrades modal from table
const closeTableUpgradesModal = () => {
  showTableUpgradesModal.value = false
  selectedCarForUpgrades.value = null
  // Refresh upgrades totals when modal closes (in case upgrades were modified)
  fetchCarsUpgrades()
}

// Handle client change in edit form - validate passport
const handleEditClientChange = async (clientId) => {
  if (!clientId || editFormData.value.is_tmp_client) return

  try {
    const result = await callApi({
      query: `
        SELECT id_copy_path
        FROM clients
        WHERE id = ?
      `,
      params: [clientId],
    })

    if (result.success && result.data.length > 0) {
      if (!result.data[0].id_copy_path) {
        error.value = t('sellBills.cannot_assign_client_no_id') || 'Cannot assign to this client because passport/ID card is not uploaded'
        editFormData.value.id_client = null
      } else {
        error.value = null
      }
    }
  } catch (err) {
    console.error('Error checking client passport:', err)
    error.value = t('sellBills.failed_to_verify_client_id') || 'Failed to verify client passport'
    editFormData.value.id_client = null
  }
}

// Handle close edit dialog - restore original notes if cancelled
const handleCloseEditDialog = () => {
  // Restore original notes (discard any changes made in the modal)
  editCarNotes.value = JSON.parse(JSON.stringify(originalEditCarNotes.value))
  showEditDialog.value = false
  editingCar.value = null
  editCarNotes.value = []
  originalEditCarNotes.value = []
  // Reset upgrades data
  editCarUpgrades.value = []
  editCarUpgradesTotal.value = 0
}

// Handle edit button click
const handleEdit = async (car) => {
  isProcessing.value = true
  try {
    // Fetch full car details including client, port, car name, and VIN
    const result = await callApi({
      query: `
        SELECT 
          cs.*,
          c.name as client_name,
          dp.discharge_port,
          cn.car_name,
          cs.vin
        FROM cars_stock cs
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        WHERE cs.id = ?
      `,
      params: [car.id],
    })

    if (result.success && result.data.length > 0) {
      const carData = result.data[0]
      editingCar.value = carData
      editFormData.value = {
        id_client: carData.id_client,
        id_port_discharge: carData.id_port_discharge,
        price_cell: carData.price_cell,
        freight: carData.freight,
        rate: carData.rate,
        notes: carData.notes,
        is_tmp_client: carData.is_tmp_client,
      }

      // Parse existing notes (JSON format)
      if (carData.notes) {
        try {
          let notesArray = []
          if (typeof carData.notes === 'string' && carData.notes.trim().startsWith('[')) {
            notesArray = JSON.parse(carData.notes)
          } else if (Array.isArray(carData.notes)) {
            notesArray = carData.notes
          } else {
            // Old format (plain text) - convert to JSON
            const userStr = localStorage.getItem('user')
            const currentUser = userStr ? JSON.parse(userStr) : null
            notesArray = [{
              id_user: currentUser?.id || null,
              note: carData.notes,
              timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
            }]
          }
          editCarNotes.value = Array.isArray(notesArray) ? notesArray : []
        } catch (e) {
          // If parsing fails, treat as old format (plain text)
          const userStr = localStorage.getItem('user')
          const currentUser = userStr ? JSON.parse(userStr) : null
          editCarNotes.value = [{
            id_user: currentUser?.id || null,
            note: carData.notes,
            timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
          }]
        }
      } else {
        editCarNotes.value = []
      }
      
      // Store original notes to restore on cancel
      originalEditCarNotes.value = JSON.parse(JSON.stringify(editCarNotes.value))

      // Fetch necessary data for dropdowns and upgrades BEFORE calculating prices
      await Promise.all([fetchClients(), fetchDischargePorts(), fetchAllUsers(), fetchEditCarUpgrades()])

      // Set the price input value based on existing data (after upgrades are fetched)
      const upgrades = editCarUpgradesTotal.value || 0
      console.log('[DEBUG] handleEdit - Initializing prices with upgrades:', upgrades)
      console.log('[DEBUG] handleEdit - carData.price_cell:', carData.price_cell)
      console.log('[DEBUG] handleEdit - carData.cfr_da:', carData.cfr_da)
      console.log('[DEBUG] handleEdit - editFormData.value.rate:', editFormData.value.rate)
      console.log('[DEBUG] handleEdit - editFormData.value.freight:', editFormData.value.freight)
      
      // Always prefer price_cell as source of truth if it exists, as it's the base FOB price
      // CFR DA might have been calculated without upgrades in the past
      if (carData.price_cell) {
        // If car has price_cell, use it as the source of truth
        selectedCurrency.value = 'USD'
        priceInput.value = carData.price_cell
        editFormData.value.price_cell = carData.price_cell
        // Recalculate CFR DA from price_cell with current upgrades
        if (editFormData.value.rate) {
          const recalculatedCFR = calculateCFRDAFromPrice(
            carData.price_cell,
            editFormData.value.rate,
            editFormData.value.freight,
            upgrades,
          )
          console.log('[DEBUG] handleEdit - Recalculated CFR DA from price_cell:', recalculatedCFR, 'with upgrades:', upgrades)
          cfrDaInput.value = recalculatedCFR
        }
      } else if (carData.cfr_da) {
        // If only CFR DA exists, calculate price_cell from it with current upgrades
        // Note: This CFR DA might have been calculated without upgrades, so we recalculate
        selectedCurrency.value = 'DZD'
        priceInput.value = carData.cfr_da
        cfrDaInput.value = carData.cfr_da
        // Calculate price_cell from CFR DA with current upgrades
        if (editFormData.value.rate) {
          const recalculatedPrice = calculatePriceFromCFRDA(
            carData.cfr_da,
            editFormData.value.rate,
            editFormData.value.freight,
            upgrades,
          )
          console.log('[DEBUG] handleEdit - Calculated price_cell from CFR DA:', recalculatedPrice, 'with upgrades:', upgrades)
          editFormData.value.price_cell = recalculatedPrice
        }
      } else {
        // No existing data, default to DZD
        selectedCurrency.value = 'DZD'
        priceInput.value = null
        cfrDaInput.value = null
      }
      console.log('[DEBUG] handleEdit - Final initialized values - price_cell:', editFormData.value.price_cell, 'cfr_da:', cfrDaInput.value, 'priceInput:', priceInput.value, 'currency:', selectedCurrency.value)

      showEditDialog.value = true
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isProcessing.value = false
  }
}

// Handle save edit
const handleSaveEdit = async () => {
  if (!editingCar.value) return

  isProcessing.value = true
  error.value = null

  // Validate client has passport uploaded (if client changed and not temporary)
  if (editFormData.value.id_client && !editFormData.value.is_tmp_client) {
    try {
      const clientCheck = await callApi({
        query: `
          SELECT id_copy_path
          FROM clients
          WHERE id = ?
        `,
        params: [editFormData.value.id_client],
      })

      if (clientCheck.success && clientCheck.data.length > 0) {
        if (!clientCheck.data[0].id_copy_path) {
          error.value = t('sellBills.cannot_assign_client_no_id') || 'Cannot assign to this client because passport/ID card is not uploaded'
          isProcessing.value = false
          return
        }
      }
    } catch (err) {
      console.error('Error checking client passport:', err)
      error.value = t('sellBills.failed_to_verify_client_id') || 'Failed to verify client passport'
      isProcessing.value = false
      return
    }
  }

  try {
    // Ensure both values are set based on currency selection
    const upgrades = editCarUpgradesTotal.value || 0
    console.log('[DEBUG] handleSaveEdit - Using upgrades:', upgrades)
    console.log('[DEBUG] handleSaveEdit - priceInput.value:', priceInput.value)
    console.log('[DEBUG] handleSaveEdit - selectedCurrency.value:', selectedCurrency.value)
    console.log('[DEBUG] handleSaveEdit - rate:', editFormData.value.rate)
    console.log('[DEBUG] handleSaveEdit - freight:', editFormData.value.freight)
    
    if (selectedCurrency.value === 'USD') {
      // User entered USD, ensure CFR DA is calculated
      editFormData.value.price_cell = priceInput.value
      const calculatedCFR = calculateCFRDAFromPrice(
        priceInput.value,
        editFormData.value.rate,
        editFormData.value.freight,
        upgrades,
      )
      console.log('[DEBUG] handleSaveEdit - Calculated CFR DA:', calculatedCFR)
      cfrDaInput.value = calculatedCFR
    } else {
      // User entered DZD, ensure price_cell is calculated
      cfrDaInput.value = priceInput.value
      const calculatedPrice = calculatePriceFromCFRDA(
        priceInput.value,
        editFormData.value.rate,
        editFormData.value.freight,
        upgrades,
      )
      console.log('[DEBUG] handleSaveEdit - Calculated price_cell:', calculatedPrice)
      editFormData.value.price_cell = calculatedPrice
    }
    console.log('[DEBUG] handleSaveEdit - Final values before save - price_cell:', editFormData.value.price_cell, 'cfr_da:', cfrDaInput.value)

    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_client = ?,
            id_port_discharge = ?,
            price_cell = ?,
            freight = ?,
            rate = ?,
            cfr_da = ?,
            notes = ?,
            is_tmp_client = ?
        WHERE id = ?
      `,
      params: [
        editFormData.value.id_client,
        editFormData.value.id_port_discharge,
        editFormData.value.price_cell,
        editFormData.value.freight,
        editFormData.value.rate,
        cfrDaInput.value || null,
        editFormData.value.notes || null,
        editFormData.value.is_tmp_client,
        editingCar.value.id,
      ],
    })

    if (result.success) {
      showEditDialog.value = false
      editingCar.value = null
      editCarNotes.value = []
      originalEditCarNotes.value = []
      await fetchCarsByBillId(props.sellBillId)
      emit('refresh')
    } else {
      error.value = result.error || 'Failed to update car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isProcessing.value = false
  }
}

// Add handleUnassign function
const handleUnassign = async (carId) => {
  await unassignCar(carId)
}

// Add handlePrint function
const handlePrint = (car) => {
  selectedCarForPrint.value = car
  showPrintOptions.value = true
}

// Handle print options dialog close
const handlePrintOptionsClose = () => {
  showPrintOptions.value = false
  selectedCarForPrint.value = null
}

// Handle print options dialog proceed
const handlePrintOptionsProceed = (options) => {
  showPrintOptions.value = false
  selectedCarForPrint.value = null
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
})

const fetchCarsByBillId = async (billId) => {
  if (!billId) {
    cars.value = []
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.notes,
          cs.price_cell,
          cs.cfr_da,
          cs.freight,
          cs.rate,
          cs.date_loding,
          cs.path_documents,
          cs.is_big_car,
          cs.is_used_car,
          cs.container_ref,
          cs.date_assigned,
          lp.loading_port,
          c.name as client_name,
          c.mobiles as client_mobiles,
          dp.discharge_port,
          bd.amount as buy_price,
          cn.car_name,
          clr.color,
          clr.hexa
        FROM cars_stock cs
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        WHERE cs.id_sell = ?
      `,
      params: [billId],
    })

    if (result.success) {
      // Parse notes JSON and format for display
      const carsData = (result.data || []).map((car) => {
        // Parse notes JSON if it exists
        let notesDisplay = 'N/A'
        if (car.notes) {
          try {
            let notesArray = []
            if (typeof car.notes === 'string' && car.notes.trim().startsWith('[')) {
              notesArray = JSON.parse(car.notes)
            } else if (Array.isArray(car.notes)) {
              notesArray = car.notes
            } else {
              // Old format (plain text) - use as is
              notesDisplay = car.notes
            }
            
            if (Array.isArray(notesArray) && notesArray.length > 0) {
              // Format all notes without username or timestamp
              notesDisplay = notesArray.map((note) => {
                return note.note || ''
              }).join('\n')
            }
          } catch (e) {
            // If parsing fails, treat as old format (plain text)
            notesDisplay = car.notes
          }
        }
        
        return {
          ...car,
          notesDisplay: notesDisplay, // Add formatted display version
        }
      })
      cars.value = carsData
      
      // Fetch upgrades for all cars
      await fetchCarsUpgrades()
    } else {
      error.value = result.error || 'Failed to fetch cars'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Watch for changes in the sellBillId prop
// Watch for upgrades total changes and recalculate prices
watch(
  () => editCarUpgradesTotal.value,
  (newValue, oldValue) => {
    console.log('[DEBUG] editCarUpgradesTotal watcher triggered')
    console.log('[DEBUG] Old value:', oldValue, 'New value:', newValue)
    console.log('[DEBUG] showEditDialog.value:', showEditDialog.value)
    console.log('[DEBUG] priceInput.value:', priceInput.value)
    console.log('[DEBUG] editFormData.value.rate:', editFormData.value.rate)
    console.log('[DEBUG] editingCar.value?.id:', editingCar.value?.id)
    
    // Only recalculate if edit dialog is open and we have the necessary data
    if (showEditDialog.value && priceInput.value && editFormData.value.rate && editingCar.value?.id) {
      console.log('[DEBUG] Conditions met, calling recalculatePricesWithUpgrades from watcher')
      recalculatePricesWithUpgrades()
    } else {
      console.log('[DEBUG] Conditions not met, skipping recalculation')
    }
  }
)

watch(
  () => props.sellBillId,
  (newId) => {
    if (newId) {
      fetchCarsByBillId(newId)
    } else {
      cars.value = []
    }
  },
  { immediate: true },
)

// Truncate notes display
const getTruncatedCarNotes = (notesText, carId, maxLength = 100) => {
  if (!notesText || notesText === 'N/A' || notesText === '-') return notesText
  if (expandedCarNotes.value.has(carId)) return notesText
  if (notesText.length <= maxLength) return notesText
  return notesText.substring(0, maxLength) + '...'
}

const isCarNotesTruncated = (notesText, carId, maxLength = 100) => {
  if (!notesText || notesText === 'N/A' || notesText === '-') return false
  return notesText.length > maxLength && !expandedCarNotes.value.has(carId)
}

const toggleCarNotesExpansion = (carId) => {
  if (expandedCarNotes.value.has(carId)) {
    expandedCarNotes.value.delete(carId)
  } else {
    expandedCarNotes.value.add(carId)
  }
}

// Calculate profit
const calculateProfit = (sellPrice, buyPrice, freight) => {
  if (!sellPrice || !buyPrice) return 'N/A'
  const sell = parseFloat(sellPrice)
  const buy = parseFloat(buyPrice)
  const freightCost = freight ? parseFloat(freight) : 0
  return (sell - buy - freightCost).toFixed(2)
}

// Add totalValue computed property - now in CFR DA
const totalValue = computed(() => {
  if (!cars.value || cars.value.length === 0) return 0
  return cars.value.reduce((total, car) => {
    // Use cfr_da field if available, otherwise calculate from price_cell
    if (car.cfr_da) {
      return total + parseFloat(car.cfr_da)
    }

    const price = parseFloat(car.price_cell) || 0
    const freight = parseFloat(car.freight) || 0
    const upgrades = getCarUpgradesTotal(car.id) || 0
    const rate = parseFloat(car.rate) || 0

    // Calculate CFR value (price + upgrades + freight) and convert to DA
    const cfrValue = price + upgrades + freight
    const daValue = cfrValue * rate

    return total + daValue
  }, 0)
})

// Add computed function for CFR DA calculation
const calculateCFRDA = (car) => {
  // Use cfr_da field if available, otherwise calculate from price_cell
  if (car.cfr_da) {
    return parseFloat(car.cfr_da).toFixed(2)
  }
  if (!car.price_cell || !car.rate) return 'N/A'
  const sellPrice = parseFloat(car.price_cell) || 0
  const freight = parseFloat(car.freight) || 0
  const upgrades = getCarUpgradesTotal(car.id) || 0
  const rate = parseFloat(car.rate) || 0
  // CFR DA = (price_cell + upgrades + freight) × rate
  return ((sellPrice + upgrades + freight) * rate).toFixed(2)
}

// Add filter refs
const filters = ref({
  carName: '',
  clientName: '',
  minPrice: '',
  maxPrice: '',
})

// Add computed property for filtered cars
const filteredCars = computed(() => {
  return cars.value.filter((car) => {
    // Car name filter
    if (
      filters.value.carName &&
      !car.car_name?.toLowerCase().includes(filters.value.carName.toLowerCase())
    ) {
      return false
    }

    // Client name filter
    if (
      filters.value.clientName &&
      !car.client_name?.toLowerCase().includes(filters.value.clientName.toLowerCase())
    ) {
      return false
    }

    // Price range filter
    const price = parseFloat(car.price_cell)
    if (filters.value.minPrice && price < parseFloat(filters.value.minPrice)) {
      return false
    }
    if (filters.value.maxPrice && price > parseFloat(filters.value.maxPrice)) {
      return false
    }

    return true
  })
})

// Add function to clear filters
const clearFilters = () => {
  filters.value = {
    carName: '',
    clientName: '',
    minPrice: '',
    maxPrice: '',
  }
}

// Expose methods to parent component
defineExpose({
  fetchCarsByBillId,
})

// Add this method in <script setup>
const calculateCFRUSD = (car) => {
  if (!car.price_cell || !car.freight) return 'N/A'
  const price = parseFloat(car.price_cell) || 0
  const freight = parseFloat(car.freight) || 0
  const upgrades = getCarUpgradesTotal(car.id) || 0
  // CFR USD = price_cell + upgrades + freight
  return (price + upgrades + freight).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

// Add totalCfrUsd computed property
const totalCfrUsd = computed(() => {
  if (!cars.value || cars.value.length === 0) return 0
  return filteredCars.value.reduce((total, car) => {
    const price = parseFloat(car.price_cell) || 0
    const freight = parseFloat(car.freight) || 0
    const upgrades = getCarUpgradesTotal(car.id) || 0
    // CFR USD = price_cell + upgrades + freight
    return total + price + upgrades + freight
  }, 0)
})

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

// Function to get container status
const getContainerStatus = (car) => {
  if (car.container_ref) {
    return { status: 'loaded', text: car.container_ref, color: 'blue' }
  } else {
    return { status: 'not-loaded', text: 'Not Loaded', color: 'gray' }
  }
}

// Function to format date
const formatDate = (dateString) => {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('ar-DZ', {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
  })
}
</script>

<template>
  <div class="cars-table-component" v-bind="attrs" :id="attrs.id">
    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>{{ t('sellBills.loading_cars') }}</span>
    </div>

    <div class="table-header">
      <h3>
        <i class="fas fa-car"></i>
        {{ t('sellBills.cars_in_bill') }}
      </h3>
      <div class="total-info" v-if="cars.length > 0">
        <span>
          <i class="fas fa-hashtag"></i>
          {{ t('sellBills.total_cars') }}: {{ filteredCars.length }} / {{ cars.length }}
        </span>
        <span v-if="totalValue > 0">
          <i class="fas fa-money-bill-wave"></i>
          {{ t('sellBills.total_cfr_da') }}: {{ totalValue.toLocaleString() }} DZD
        </span>
        <span v-if="totalCfrUsd > 0">
          <i class="fas fa-dollar-sign"></i>
          {{ t('sellBills.total_cfr_usd') }}:
          {{
            totalCfrUsd.toLocaleString(undefined, {
              minimumFractionDigits: 2,
              maximumFractionDigits: 2,
            })
          }}
          USD
        </span>
      </div>
    </div>

    <!-- Add Filters Section -->
    <div class="filters-section">
      <div class="filters-header">
        <h4>
          <i class="fas fa-filter"></i>
          {{ t('sellBills.filters') }}
        </h4>
        <button @click="clearFilters" class="clear-filters-btn">
          <i class="fas fa-times"></i>
          {{ t('sellBills.clear_filters') }}
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>
            <i class="fas fa-car"></i>
            {{ t('sellBills.car_name') }}
          </label>
          <input
            type="text"
            v-model="filters.carName"
            :placeholder="t('sellBills.filter_by_car_name')"
          />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-user"></i>
            {{ t('sellBills.client_name') }}
          </label>
          <input
            type="text"
            v-model="filters.clientName"
            :placeholder="t('sellBills.filter_by_client_name')"
          />
        </div>

        <div class="filter-group price-range">
          <label>
            <i class="fas fa-dollar-sign"></i>
            {{ t('sellBills.price_range') }}
          </label>
          <div class="price-inputs">
            <input
              type="number"
              v-model="filters.minPrice"
              :placeholder="t('sellBills.min')"
              step="0.01"
              min="0"
            />
            <span class="separator">-</span>
            <input
              type="number"
              v-model="filters.maxPrice"
              :placeholder="t('sellBills.max')"
              step="0.01"
              min="0"
            />
          </div>
        </div>
      </div>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-else-if="!sellBillId" class="no-selection">
      <i class="fas fa-hand-pointer fa-2x"></i>
      <p>{{ t('sellBills.please_select_sell_bill') }}</p>
    </div>

    <div v-else-if="cars.length === 0" class="no-data">
      <i class="fas fa-car fa-2x"></i>
      <p>{{ t('sellBills.no_cars_found_in_bill') }}</p>
    </div>

    <div v-else-if="filteredCars.length === 0" class="no-data">
      <i class="fas fa-filter fa-2x"></i>
      <p>{{ t('sellBills.no_cars_match_filters') }}</p>
    </div>

    <table v-else class="cars-table">
      <thead>
        <tr>
          <th><i class="fas fa-hashtag"></i> {{ t('sellBills.id') }}</th>
          <th><i class="fas fa-car"></i> {{ t('sellBills.car') }}</th>
          <th><i class="fas fa-user"></i> {{ t('sellBills.client') }}</th>
          <th><i class="fas fa-shipping-fast"></i> {{ t('sellBills.container') }}</th>
          <th><i class="fas fa-dollar-sign"></i> {{ t('sellBills.price') }}</th>
          <th><i class="fas fa-ship"></i> {{ t('sellBills.freight') }}</th>
          <th><i class="fas fa-calculator"></i> {{ t('sellBills.cfr') }}</th>
          <th><i class="fas fa-sticky-note"></i> {{ t('sellBills.notes') }}</th>
          <th><i class="fas fa-cog"></i> {{ t('sellBills.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in filteredCars" :key="car.id">
          <td>{{ car.id }}</td>
          <td class="car-cell">
            <div class="car-name">{{ car.car_name }}</div>
            <div class="car-vin" v-if="car.vin">{{ car.vin }}</div>
            <div class="car-badges">
              <span
                v-if="car.color"
                class="badge color-badge"
                :style="{
                  backgroundColor: car.hexa || '#000000',
                  color: getTextColor(car.hexa || '#000000'),
                }"
              >
                {{ car.color }}
              </span>
              <span
                v-if="isAdmin"
                class="badge assigned-date-badge"
                :class="{ 'not-assigned': !car.date_assigned }"
              >
                <i class="fas fa-clock"></i>
                {{
                  car.date_assigned
                    ? `${t('sellBills.timestamp')}: ${formatDate(car.date_assigned)}`
                    : `${t('sellBills.timestamp')}: ${t('sellBills.not_set')}`
                }}
              </span>
            </div>
          </td>
          <td>
            <div class="client-info">
              <div>{{ car.client_name || 'N/A' }}</div>
              <div v-if="car.client_mobiles && car.client_mobiles !== 'please provide mobile'" class="client-mobile">
                <i class="fas fa-phone"></i>
                {{ car.client_mobiles }}
              </div>
            </div>
          </td>
          <td>
            <span
              :class="['container-badge', `container-badge-${getContainerStatus(car).status}`]"
              :title="getContainerStatus(car).text"
            >
              <i class="fas fa-shipping-fast"></i>
              {{ getContainerStatus(car).text }}
            </span>
          </td>
          <td>{{ car.price_cell ? '$' + car.price_cell.toLocaleString() : 'N/A' }}</td>
          <td>{{ car.freight ? '$' + car.freight.toLocaleString() : 'N/A' }}</td>
          <td class="cfr-cell">
            <div class="cfr-badges">
              <span class="badge cfr-da-badge"
                >{{ t('sellBills.da') }}: {{ calculateCFRDA(car) }}</span
              >
              <span class="badge cfr-usd-badge"
                >{{ t('sellBills.usd') }}: {{ calculateCFRUSD(car) }}</span
              >
              <span class="badge cfr-rate-badge"
                >{{ t('sellBills.rate') }}: {{ car.rate || 'N/A' }}</span
              >
              <span 
                v-if="getCarUpgradesTotal(car.id) > 0" 
                class="badge cfr-upgrades-badge clickable"
                @click.stop="handleTableUpgradesClick(car)"
                :title="t('carStock.upgrades') || 'Click to view upgrades'"
              >
                <i class="fas fa-wrench"></i> {{ t('carStock.upgrades') || 'Upgrades' }}: ${{ formatUpgradeValue(getCarUpgradesTotal(car.id)) }}
              </span>
            </div>
          </td>
          <td class="notes-cell">
            <div class="notes-content">
              <div style="white-space: pre-line; max-width: 300px">
                {{ getTruncatedCarNotes(car.notesDisplay, car.id) }}
              </div>
              <button
                v-if="isCarNotesTruncated(car.notesDisplay, car.id)"
                @click.stop="toggleCarNotesExpansion(car.id)"
                class="btn-show-full-notes"
                type="button"
              >
                {{
                  expandedCarNotes.has(car.id)
                    ? t('sellBills.show_less') || 'Show Less'
                    : t('sellBills.show_full') || 'Show Full'
                }}
              </button>
            </div>
          </td>
          <td class="actions">
            <button
              @click="handleEdit(car)"
              :disabled="isProcessing"
              class="btn edit-btn"
              :title="t('sellBills.edit_car')"
            >
              <i class="fas fa-edit"></i>
            </button>
            <button
              @click="handlePrint(car)"
              :disabled="isProcessing"
              class="btn print-btn"
              :title="t('sellBills.print_car_document')"
            >
              <i class="fas fa-print"></i>
            </button>
            <button
              v-if="can_unassign_cars"
              @click="handleUnassign(car.id)"
              :disabled="isProcessing"
              class="btn unassign-btn"
              :title="t('sellBills.unassign_car')"
            >
              <i class="fas fa-unlink"></i>
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Edit Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-edit"></i>
            {{ t('sellBills.edit_car_details') }}
          </h3>
          <button @click="handleCloseEditDialog" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div v-if="editingCar" class="car-details">
          <p>
            <strong>{{ t('sellBills.car') }}:</strong> {{ editingCar.car_name }}
          </p>
          <p>
            <strong>{{ t('sellBills.vin') }}:</strong> {{ editingCar.vin }}
          </p>
        </div>

        <form @submit.prevent="handleSaveEdit" class="edit-form">
          <div class="form-row">
            <div class="form-group">
              <label>{{ t('sellBills.client') }}: <span class="required">*</span></label>
              <el-select
                v-model="editFormData.id_client"
                filterable
                remote
                :remote-method="remoteClientSearch"
                :loading="isProcessing"
                :placeholder="t('sellBills.search_client')"
                class="custom-select"
                required
                @change="handleEditClientChange"
              >
                <el-option
                  v-for="client in filteredClients"
                  :key="client.id"
                  :label="client.name"
                  :value="client.id"
                >
                  <div class="custom-option">
                    <i class="fas fa-user"></i>
                    <span>{{ client.name }}</span>
                    <small v-if="client.mobiles">
                      <i class="fas fa-phone"></i>
                      {{ client.mobiles }}
                    </small>
                  </div>
                </el-option>
              </el-select>
            </div>

            <div class="form-group">
              <label>{{ t('sellBills.discharge_port') }}: <span class="required">*</span></label>
              <el-select
                v-model="editFormData.id_port_discharge"
                filterable
                remote
                :remote-method="remotePortSearch"
                :loading="isProcessing"
                :placeholder="t('sellBills.search_port')"
                class="custom-select"
                required
              >
                <el-option
                  v-for="port in filteredPorts"
                  :key="port.id"
                  :label="port.discharge_port"
                  :value="port.id"
                >
                  <div class="custom-option">
                    <i class="fas fa-anchor"></i>
                    <span>{{ port.discharge_port }}</span>
                  </div>
                </el-option>
              </el-select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>
                <i class="fas fa-money-bill-wave"></i>
                {{ t('sellBills.currency') }}: <span class="required">*</span>
              </label>
              <select v-model="selectedCurrency" @change="handleCurrencyChange">
                <option value="USD">USD</option>
                <option value="DZD">DZD (DA)</option>
              </select>
            </div>

            <div class="form-group">
              <label>
                <i class="fas fa-calculator"></i>
                {{ selectedCurrency === 'USD' ? t('sellBills.price') : t('sellBills.cfr_da') }}: <span class="required">*</span>
              </label>
              <div class="input-with-info">
                <input
                  type="number"
                  v-model="priceInput"
                  step="0.01"
                  min="0"
                  required
                  @input="handlePriceInputChange"
                  :placeholder="selectedCurrency === 'USD' ? 'Enter price in USD' : 'Enter price in DZD'"
                />
              </div>
              <span class="info-text" v-if="selectedCurrency === 'USD' && editFormData.rate && priceInput">
                Calculated DZD: {{ cfrDaInput ? parseFloat(cfrDaInput).toLocaleString() : 'N/A' }} DA
              </span>
              <span class="info-text" v-else-if="selectedCurrency === 'DZD' && editFormData.rate && priceInput">
                Calculated USD: {{ editFormData.price_cell ? '$' + parseFloat(editFormData.price_cell).toLocaleString() : 'N/A' }}
              </span>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>
                <i class="fas fa-ship"></i>
                {{ t('sellBills.freight') }}:
              </label>
              <div class="input-with-info">
                <input
                  type="number"
                  v-model="editFormData.freight"
                  step="0.01"
                  min="0"
                  @input="handleRateOrFreightChange"
                  :class="{ 'default-value': editFormData.freight === editingCar?.freight }"
                />
              </div>
            </div>

            <div class="form-group">
              <label for="edit-rate">{{ t('sellBills.rate') }}:</label>
              <input
                type="number"
                id="edit-rate"
                v-model="editFormData.rate"
                step="0.01"
                min="0"
                @input="handleRateOrFreightChange"
                :disabled="isProcessing"
              />
            </div>
          </div>

          <div class="form-row full-width">
            <div class="form-group">
              <label for="edit-notes">
                <i class="fas fa-sticky-note"></i>
                {{ t('sellBills.notes') }}:
              </label>
              <div class="notes-section">
                <NotesTable v-if="editCarNotes && editCarNotes.length > 0" :notes="editCarNotes" :users="allUsers" />
                <button
                  type="button"
                  @click="handleManageEditNotes"
                  class="btn-manage-notes"
                  :disabled="isProcessing"
                >
                  <i class="fas fa-edit"></i>
                  {{ t('sellBills.manage_notes') }}
                </button>
              </div>
            </div>
          </div>

          <div class="form-row full-width">
            <div class="form-group">
              <label>
                <i class="fas fa-wrench"></i>
                {{ t('carStock.upgrades') || 'Upgrades' }}:
              </label>
              <div class="upgrades-section">
                <div class="upgrades-info">
                  <button
                    type="button"
                    @click="handleEditUpgrades"
                    class="btn-manage-upgrades"
                    :disabled="isProcessing || !editingCar?.id"
                  >
                    <i class="fas fa-wrench"></i>
                    {{ t('carStock.upgrades') || 'Manage Upgrades' }}
                  </button>
                  <div v-if="editCarUpgrades.length > 0" class="upgrades-total-display">
                    <span class="upgrades-count">
                      <i class="fas fa-list"></i>
                      {{ editCarUpgrades.length }} {{ t('carUpgrades.total_upgrades') || 'upgrades' }}
                    </span>
                    <span class="upgrades-total">
                      <i class="fas fa-calculator"></i>
                      {{ t('carUpgrades.total_amount') || 'Total' }}: {{ formatUpgradeValue(editCarUpgradesTotal) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="can_assign_to_tmp_clients" class="form-row full-width">
            <div class="form-group">
              <label class="checkbox-label">
                <input
                  type="checkbox"
                  v-model="editFormData.is_tmp_client"
                  :true-value="1"
                  :false-value="0"
                  :disabled="isProcessing"
                />
                <i class="fas fa-clock"></i>
                {{ t('sellBills.temporary_client') }}
              </label>
              <small class="help-text">{{ t('sellBills.help_text_temporary_client') }}</small>
            </div>
          </div>

          <div class="form-row full-width">
            <div class="form-actions">
              <button
                type="button"
                @click="handleCloseEditDialog"
                class="cancel-btn"
                :disabled="isProcessing"
              >
                <i class="fas fa-times"></i>
                {{ t('sellBills.cancel') }}
              </button>
              <button type="submit" class="save-btn" :disabled="isProcessing">
                <i class="fas fa-save"></i>
                {{ isProcessing ? t('sellBills.saving') : t('sellBills.save_changes') }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>

    <!-- Notes Management Modal for Edit Dialog -->
    <NotesManagementModal
      :show="showEditNotesModal"
      :notes="editCarNotes"
      :users="allUsers"
      :current-user-id="user?.id"
      :is-admin="isAdmin"
      :entity-id="editingCar?.id"
      :save-function="saveCarNotes"
      :save-immediately="false"
      @close="showEditNotesModal = false"
      @notes-updated="handleEditNotesUpdated"
    />

    <!-- Car Upgrades Modal for Edit Dialog -->
    <CarUpgradesModal
      :show="showEditUpgradesModal"
      :carId="editingCar?.id"
      :title="editingCar ? `${t('carStock.upgrades') || 'Upgrades'} - ${editingCar.car_name}` : ''"
      :users="allUsers"
      @close="closeEditUpgradesModal"
      @upgrade-added="handleUpgradeAdded"
    />

    <!-- Car Upgrades Modal for Table (clicking upgrades badge) -->
    <CarUpgradesModal
      :show="showTableUpgradesModal"
      :carId="selectedCarForUpgrades?.id"
      :title="selectedCarForUpgrades ? `${t('carStock.upgrades') || 'Upgrades'} - ${selectedCarForUpgrades.car_name}` : ''"
      :users="allUsers"
      @close="closeTableUpgradesModal"
    />
  </div>

  <!-- Print Options Dialog -->
  <SellBillPrintOption
    :visible="showPrintOptions"
    :billId="sellBillId"
    :carId="selectedCarForPrint?.id"
    documentType="car"
    @close="handlePrintOptionsClose"
    @proceed="handlePrintOptionsProceed"
  />
</template>

<style scoped>
.cars-table-component {
  position: relative;
  margin: 1.5rem 0;
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

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.table-header h3 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1f2937;
  margin: 0;
}

.total-info {
  display: flex;
  gap: 1.5rem;
}

.total-info span {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  font-weight: 500;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.no-selection,
.no-data {
  text-align: center;
  padding: 2rem;
  background-color: #f9fafb;
  border-radius: 0.5rem;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.cars-table {
  width: 100%;
  border-collapse: collapse;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border: 1px solid #d1d5db;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.cars-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #f3f4f6;
}

.cars-table th,
.cars-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1.5px solid #cbd5e1;
  position: static;
  background: none;
}

.cars-table th {
  background-color: #f3f4f6;
  font-weight: 600;
  color: #374151;
  border-bottom: 3px solid #3b82f6;
}

.cars-table tbody tr,
.cars-table tbody tr:not(:last-child) td,
.cars-table tbody tr:first-child td {
  border-top: none;
  border-bottom: 1.5px solid #cbd5e1;
  box-shadow: none;
}

.cars-table tbody tr:last-child td {
  border-bottom: none;
}

.cars-table tbody tr:nth-child(even) {
  background-color: #fafafa;
}

.cars-table tbody tr:hover {
  background-color: #f8fafc;
}

.cars-table th::after {
  display: none;
}

.cars-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.cars-table tr:hover {
  background-color: #f9fafb;
}

.cars-table td i {
  margin-right: 0.5rem;
}

.actions {
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

.print-btn {
  background-color: #10b981;
  color: white;
}

.print-btn:hover:not(:disabled) {
  background-color: #059669;
}

.unassign-btn {
  background-color: #ef4444;
  color: white;
}

.unassign-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

/* Add smooth transitions */
.cars-table tr {
  transition: background-color 0.2s;
}

.btn i {
  transition: transform 0.2s;
}

.btn:hover:not(:disabled) i {
  transform: scale(1.1);
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
  z-index: 2000;
}

.dialog {
  background: white;
  border-radius: 8px;
  padding: 24px;
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.dialog-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  color: #666;
}

.car-details {
  background: #f9fafb;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 20px;
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.edit-form .form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.edit-form .form-row.full-width {
  grid-column: 1 / -1;
}

@media (max-width: 768px) {
  .edit-form .form-row {
    grid-template-columns: 1fr;
  }
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.form-group label {
  font-weight: 500;
}

.form-group select,
.form-group input {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 0;
  width: 100%;
}

.cancel-btn,
.save-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
}

.cancel-btn {
  background: #f3f4f6;
  color: #374151;
}

.save-btn {
  background: #3b82f6;
  color: white;
}

.cancel-btn:hover:not(:disabled) {
  background: #e5e7eb;
}

.save-btn:hover:not(:disabled) {
  background: #2563eb;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.custom-select {
  width: 100%;
}

.custom-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.custom-option i {
  color: #6b7280;
  width: 16px;
}

.custom-option small {
  margin-left: auto;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 4px;
}

.required {
  color: #ef4444;
  margin-left: 2px;
}

:deep(.el-select) {
  width: 100%;
}

:deep(.el-select .el-input__wrapper) {
  background-color: white;
}

:deep(.el-select-dropdown__item) {
  padding: 0 12px;
}

:deep(.el-select-dropdown__item.selected) {
  background-color: #e5edff;
  color: #3b82f6;
}

:deep(.el-select-dropdown__item:hover) {
  background-color: #f3f4f6;
}

.input-with-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-with-info input {
  flex: 1;
}

.calculated-value {
  background-color: #f3f4f6;
  color: #1f2937;
  font-weight: 500;
  font-family: monospace;
}

.default-value {
  border-color: #10b981;
}

.filters-section {
  background-color: #f9fafb;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
  border: 1px solid #e5e7eb;
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filters-header h4 {
  margin: 0;
  font-size: 1rem;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.clear-filters-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  background-color: #ef4444;
  color: white;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.clear-filters-btn:hover {
  background-color: #dc2626;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #374151;
  font-weight: 500;
  font-size: 0.875rem;
}

.filter-group input {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
  transition: border-color 0.2s;
}

.filter-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.price-range .price-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.price-range .separator {
  color: #6b7280;
}

.price-range input {
  width: 100%;
}

/* Add transition for filtered rows */
.cars-table tbody tr {
  transition: opacity 0.2s;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
  font-weight: 500;
  color: #374151;
}

.checkbox-label input[type='checkbox'] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
}

.checkbox-label input[type='checkbox']:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.help-text {
  color: #6b7280;
  font-size: 0.875rem;
  margin-top: 0.25rem;
  margin-left: 24px;
}

.car-cell {
  flex-direction: column;
  align-items: flex-start;
}

.car-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.car-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.badge {
  padding: 4px 10px;
  font-size: 0.875rem;
  font-weight: 500;
  font-family: inherit;
  display: inline-block;
  width: 100%;
  text-align: center;
  box-sizing: border-box;
  border-radius: 6px;
}

.color-badge {
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #d1d5db;
  color: #374151;
  font-weight: 500;
  font-family: inherit;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  width: 100%;
  text-align: center;
  box-sizing: border-box;
  border-radius: 6px;
}

.vin-badge {
  background-color: #f0f9ff;
  color: #0284c7;
}

.car-vin {
  font-size: 0.875rem;
  color: #6b7280;
  margin-bottom: 4px;
  font-family: inherit;
  background-color: #f3f4f6;
  padding: 4px 8px;
  border-radius: 6px;
  display: inline-block;
  border: 1px solid #d1d5db;
  box-shadow: none;
  width: 100%;
  text-align: center;
  box-sizing: border-box;
}

.cfr-cell {
  padding-top: 8px;
  padding-bottom: 8px;
}

.cfr-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.cfr-da-badge {
  background-color: #fef9c3;
  color: #b45309;
  border-radius: 8px;
}

.cfr-usd-badge {
  background-color: #dbeafe;
  color: #1d4ed8;
  border-radius: 8px;
}

.cfr-rate-badge {
  background-color: #fce7f3;
  color: #be185d;
}

.cfr-upgrades-badge {
  background-color: #f0f9ff;
  color: #1e40af;
  border: 1px solid #3b82f6;
  font-weight: 600;
  border-radius: 8px;
}

.cfr-upgrades-badge.clickable {
  cursor: pointer;
  transition: all 0.2s;
}

.cfr-upgrades-badge.clickable:hover {
  background-color: #dbeafe;
  border-color: #2563eb;
  transform: scale(1.05);
}

.cfr-upgrades-badge i {
  margin-right: 4px;
}

.notes-cell {
  max-width: 300px;
  white-space: pre-line;
  word-wrap: break-word;
  vertical-align: top;
  padding: 8px;
}

.notes-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.btn-show-full-notes {
  background: #007bff;
  color: white;
  border: none;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85em;
  align-self: flex-start;
  margin-top: 4px;
  transition: background-color 0.2s;
}

.btn-show-full-notes:hover {
  background: #0056b3;
}

.btn-show-full-notes:active {
  transform: scale(0.98);
}

.badge {
  padding: 4px 10px;
  font-size: 0.875rem;
  font-weight: 500;
  font-family: inherit;
  display: inline-block;
  width: 100%;
  text-align: center;
  box-sizing: border-box;
  border-radius: 6px;
}

/* Container badge styles */
.container-badge {
  display: inline-block;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  text-align: center;
  min-width: 100px;
  cursor: help;
}

.container-badge-loaded {
  background-color: #dbeafe;
  color: #1d4ed8;
  border: 1px solid #93c5fd;
}

.container-badge-not-loaded {
  background-color: #f3f4f6;
  color: #6b7280;
  border: 1px solid #d1d5db;
}

.assigned-date-badge {
  background-color: #eff6ff;
  color: #1e40af;
  border: 1px solid #bfdbfe;
  font-size: 0.875rem;
  padding: 4px 10px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
  font-weight: 500;
  font-family: inherit;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  width: 100%;
  justify-content: center;
}

.assigned-date-badge i {
  font-size: 0.7rem;
  color: #3b82f6;
}

.assigned-date-badge.not-assigned {
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.assigned-date-badge.not-assigned i {
  color: #dc2626;
}

.notes-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.btn-manage-notes {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background-color 0.2s;
  align-self: flex-start;
}

.btn-manage-notes:hover {
  background-color: #0056b3;
}

.btn-manage-notes:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.upgrades-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.upgrades-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.upgrades-total-display {
  display: flex;
  gap: 1.5rem;
  padding: 0.75rem;
  background-color: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 6px;
  flex-wrap: wrap;
}

.upgrades-count,
.upgrades-total {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: #1e40af;
}

.upgrades-count i,
.upgrades-total i {
  color: #3b82f6;
}

.upgrades-total {
  font-weight: 600;
}

.upgrades-total {
  font-family: monospace;
}

.btn-manage-upgrades {
  background-color: #f59e0b;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background-color 0.2s;
  align-self: flex-start;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-manage-upgrades:hover:not(:disabled) {
  background-color: #d97706;
}

.btn-manage-upgrades:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.client-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.client-mobile {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  color: #1e40af;
  font-weight: 500;
}

.client-mobile i {
  font-size: 0.7rem;
  color: #3b82f6;
}
</style>
