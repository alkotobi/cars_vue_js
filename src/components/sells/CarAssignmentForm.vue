<script setup>
import { ref, onMounted, watch, computed, nextTick } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import { ElSelect, ElOption, ElButton } from 'element-plus'
import 'element-plus/dist/index.css'
import NotesTable from '../shared/NotesTable.vue'
import NotesManagementModal from '../shared/NotesManagementModal.vue'
import CarUpgradesModal from '../shared/CarUpgradesModal.vue'
import AddClientDialog from '../car-stock/AddClientDialog.vue'
import AddItemDialog from '../car-stock/AddItemDialog.vue'

const { t } = useEnhancedI18n()

const props = defineProps({
  carId: {
    type: Number,
    default: null,
  },
  sellBillId: {
    type: Number,
    default: null,
  },
  visible: {
    type: Boolean,
    default: false,
  },
})
const user = ref(JSON.parse(localStorage.getItem('user')))
const isAdmin = computed(() => user.value?.role_id === 1)
const can_assign_to_tmp_clients = computed(
  () =>
    user.value?.permissions?.some((p) => p.permission_name === 'can_assign_to_tmp_clients') ||
    isAdmin.value,
)
const emit = defineEmits(['close', 'assign-success'])

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const clients = ref([])
const filteredClients = ref([])
const dischargePorts = ref([])
const carDetails = ref(null)

// Add loading state for form submission
const isSubmitting = ref(false)
const isProcessing = ref(false)

const formData = ref({
  id_client: null,
  id_port_discharge: null,
  price_cell: null,
  freight: null,
  rate: null,
  is_tmp_client: 0,
})

// Notes management
const carNotes = ref([])
const originalCarNotes = ref([]) // Store original notes to restore on cancel
const allUsers = ref([])
const showNotesModal = ref(false)

// Upgrades management
const showUpgradesModal = ref(false)
const carUpgrades = ref([])
const carUpgradesTotal = ref(0)

// Add ref for CFR DA input
const cfrDaInput = ref(null)

// Add currency selector
const selectedCurrency = ref('DZD') // 'USD' or 'DZD'
const priceInput = ref(null) // Single price input value

// New client functionality
const showAddClientDialog = ref(false)

// New discharge port functionality
const showAddDischargePortDialog = ref(false)
const newDischargePort = ref({
  discharge_port: '',
})

// Add function to calculate price from CFR DA
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

// Add function to calculate CFR DA from price
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
  if (priceInput.value && formData.value.rate) {
    const upgrades = carUpgradesTotal.value || 0
    
    if (selectedCurrency.value === 'USD') {
      // If switching to USD, convert current DZD to USD
      if (cfrDaInput.value) {
        priceInput.value = calculatePriceFromCFRDA(
          cfrDaInput.value,
        formData.value.rate,
        formData.value.freight,
        upgrades,
      )
        formData.value.price_cell = priceInput.value
      }
    } else {
      // If switching to DZD, convert current USD to DZD
      if (formData.value.price_cell) {
        priceInput.value = calculateCFRDAFromPrice(
          formData.value.price_cell,
          formData.value.rate,
          formData.value.freight,
          upgrades,
        )
        cfrDaInput.value = priceInput.value
    }
  }
}
}

// Handle price input change based on selected currency
const handlePriceInputChange = () => {
  if (!priceInput.value || !formData.value.rate) return

  const upgrades = carUpgradesTotal.value || 0

  if (selectedCurrency.value === 'USD') {
    // User entered USD (FOB), calculate DZD (CFR DA)
    formData.value.price_cell = priceInput.value
    cfrDaInput.value = calculateCFRDAFromPrice(
      priceInput.value,
      formData.value.rate,
      formData.value.freight,
      upgrades,
    )
  } else {
    // User entered DZD (CFR DA), calculate USD (FOB)
    cfrDaInput.value = priceInput.value
    formData.value.price_cell = calculatePriceFromCFRDA(
      priceInput.value,
      formData.value.rate,
      formData.value.freight,
      upgrades,
    )
  }
}

// Handle rate or freight change - recalculate based on selected currency
const handleRateOrFreightChange = () => {
  if (!priceInput.value || !formData.value.rate) return

  const upgrades = carUpgradesTotal.value || 0

  if (selectedCurrency.value === 'USD') {
    // Recalculate DZD from USD
    cfrDaInput.value = calculateCFRDAFromPrice(
      priceInput.value,
      formData.value.rate,
      formData.value.freight,
      upgrades,
    )
  } else {
    // Recalculate USD from DZD
    formData.value.price_cell = calculatePriceFromCFRDA(
      priceInput.value,
      formData.value.rate,
      formData.value.freight,
      upgrades,
    )
  }
}

// Legacy handlers for backwards compatibility (kept for now but may not be needed)
const handlePriceChange = (event) => {
  // This is now handled by handlePriceInputChange
}

const handleCFRDAChange = (event) => {
  // This is now handled by handlePriceInputChange
}

// Fetch clients for dropdown
const fetchClients = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name
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

const remoteMethod = (query) => {
  if (query) {
    // Filter the existing clients data
    filteredClients.value = clients.value.filter((client) =>
      client.name.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    // If no query, show all clients
    filteredClients.value = clients.value
  }
}

const handleClientChange = async () => {
  if (!formData.value.id_client) return

  try {
    const result = await callApi({
      query: `
        SELECT id_copy_path
        FROM clients
        WHERE id = ?
      `,
      params: [formData.value.id_client],
    })

    if (result.success && result.data.length > 0) {
      if (!result.data[0].id_copy_path) {
        error.value = 'Cannot assign to this client because you did not provide ID card for him'
        formData.value.id_client = null
      }
    }
  } catch (err) {
    console.error('Error checking client ID card:', err)
    error.value = 'Failed to verify client ID card'
    formData.value.id_client = null
  }
}

// Fetch discharge ports for dropdown
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
    } else {
      error.value = result.error || 'Failed to fetch discharge ports'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Add new function to fetch default freight values
const fetchDefaultFreight = async (isSmallCar) => {
  try {
    const result = await callApi({
      query: 'SELECT freight_small, freight_big FROM defaults LIMIT 1',
      params: [],
    })

    if (result.success && result.data.length > 0) {
      const defaultFreight = isSmallCar ? result.data[0].freight_small : result.data[0].freight_big
      formData.value.freight = defaultFreight
    }
  } catch (err) {
    console.error('Error fetching default freight:', err)
  }
}

// Update the fetchCarDetails function to include car size info and set freight
const fetchCarDetails = async () => {
  if (!props.carId) return

  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.price_cell,
          cs.is_big_car,
          cs.notes,
          cn.car_name,
          clr.color,
          bd.price_sell as buy_price
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        WHERE cs.id = ?
      `,
      params: [props.carId],
    })

    if (result.success && result.data.length > 0) {
      carDetails.value = result.data[0]

      // Parse existing notes (JSON format)
      if (carDetails.value.notes) {
        try {
          let notesArray = []
          if (typeof carDetails.value.notes === 'string' && carDetails.value.notes.trim().startsWith('[')) {
            notesArray = JSON.parse(carDetails.value.notes)
          } else if (Array.isArray(carDetails.value.notes)) {
            notesArray = carDetails.value.notes
          } else {
            // Old format (plain text) - convert to JSON
            notesArray = [{
              id_user: user.value?.id || null,
              note: carDetails.value.notes,
              timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
            }]
          }
          carNotes.value = Array.isArray(notesArray) ? notesArray : []
        } catch (e) {
          // If parsing fails, treat as old format (plain text)
          carNotes.value = [{
            id_user: user.value?.id || null,
            note: carDetails.value.notes,
            timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
          }]
        }
      } else {
        carNotes.value = []
      }
      
      // Store original notes to restore on cancel
      originalCarNotes.value = JSON.parse(JSON.stringify(carNotes.value))

      // Fetch defaults first to have rate and freight available
      await fetchDefaultFreight(!carDetails.value.is_big_car)
      await fetchDefaultRate()
      
      // Fetch car upgrades
      await fetchCarUpgrades()

      // Set the price input value based on existing data
      const upgrades = carUpgradesTotal.value || 0
      
      if (carDetails.value.cfr_da) {
        // If car has CFR DA, set currency to DZD and use that value
        selectedCurrency.value = 'DZD'
        priceInput.value = carDetails.value.cfr_da
        cfrDaInput.value = carDetails.value.cfr_da
        // Calculate price_cell from CFR DA
        if (formData.value.rate) {
          formData.value.price_cell = calculatePriceFromCFRDA(
            carDetails.value.cfr_da,
            formData.value.rate,
            formData.value.freight,
            upgrades,
          )
        }
      } else if (carDetails.value.price_cell) {
        // If car has price_cell, set currency to USD and use that value
        selectedCurrency.value = 'USD'
        priceInput.value = carDetails.value.price_cell
        formData.value.price_cell = carDetails.value.price_cell
        // Calculate CFR DA from price_cell
        if (formData.value.rate) {
          cfrDaInput.value = calculateCFRDAFromPrice(
            carDetails.value.price_cell,
            formData.value.rate,
            formData.value.freight,
            upgrades,
          )
        }
      } else {
        // No existing data, default to DZD
        selectedCurrency.value = 'DZD'
        priceInput.value = null
        cfrDaInput.value = null
      }
    } else {
      error.value = result.error || 'Failed to fetch car details'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Update fetchDefaultRate to not calculate initial CFR DA automatically
const fetchDefaultRate = async () => {
  try {
    const result = await callApi({
      query: 'SELECT rate FROM defaults LIMIT 1',
      params: [],
    })

    if (result.success && result.data.length > 0) {
      formData.value.rate = result.data[0].rate
      // CFR DA will be calculated only when user enters it manually
    }
  } catch (err) {
    console.error('Error fetching default rate:', err)
  }
}

// Add computed property for CFR DA calculation
const calculateCFRDA = computed(() => {
  if (!formData.value.price_cell || !formData.value.rate) return t('sellBills.notAvailable')
  const sellPrice = parseFloat(formData.value.price_cell) || 0
  const freight = parseFloat(formData.value.freight) || 0
  const rate = parseFloat(formData.value.rate) || 0
  return ((sellPrice + freight) * rate).toFixed(2)
})

// Add computed property for calculating sell price from CFR DA
const calculateSellPriceFromCFRDA = computed(() => {
  if (!cfrDaInput.value || !formData.value.rate) return 0
  const cfrDa = parseFloat(cfrDaInput.value) || 0
  const rate = parseFloat(formData.value.rate) || 0
  const freight = parseFloat(formData.value.freight) || 0
  return (cfrDa / rate - freight).toFixed(2)
})

// Add computed property to show calculated CFR DA for display purposes
const calculatedCfrDaForDisplay = computed(() => {
  if (carDetails.value?.cfr_da) {
    // If car has existing CFR DA, show it
    return carDetails.value.cfr_da
  } else if (carDetails.value?.price_cell && formData.value.rate && formData.value.freight) {
    // Calculate CFR DA from existing price_cell for display only
    const priceCell = parseFloat(carDetails.value.price_cell) || 0
    const rate = parseFloat(formData.value.rate) || 0
    const freight = parseFloat(formData.value.freight) || 0
    const calculatedCfrDa = ((priceCell + freight) * rate).toFixed(2)
    return calculatedCfrDa
  }
  return null
})

// Assign car with collected data
const assignCar = async () => {
  // Prevent multiple submissions
  if (isSubmitting.value) {
    return
  }

  if (!validateForm()) return

  // Validate client has passport uploaded
  if (formData.value.id_client && !formData.value.is_tmp_client) {
    try {
      const clientCheck = await callApi({
        query: `
          SELECT id_copy_path
          FROM clients
          WHERE id = ?
        `,
        params: [formData.value.id_client],
      })

      if (clientCheck.success && clientCheck.data.length > 0) {
        if (!clientCheck.data[0].id_copy_path) {
          error.value = t('sellBills.cannot_assign_client_no_id') || 'Cannot assign to this client because passport/ID card is not uploaded'
          return
        }
      }
    } catch (err) {
      console.error('Error checking client passport:', err)
      error.value = t('sellBills.failed_to_verify_client_id') || 'Failed to verify client passport'
      return
    }
  }

  try {
    isSubmitting.value = true
    loading.value = true
    error.value = null

    // First get the bill_ref from the sell_bill
    const billResult = await callApi({
      query: `
        SELECT bill_ref
        FROM sell_bill
        WHERE id = ?
      `,
      params: [props.sellBillId],
    })

    if (!billResult.success || !billResult.data.length) {
      error.value = 'Failed to get bill reference'
      return
    }

    const billRef = billResult.data[0].bill_ref
    const currentDate = new Date().toISOString().split('T')[0]

    // Check if car is still available for assignment
    const checkResult = await callApi({
      query: `
        SELECT id, id_client
        FROM cars_stock
        WHERE id = ?
      `,
      params: [props.carId],
    })

    if (!checkResult.success || !checkResult.data.length) {
      error.value = 'Car not found'
      return
    }

    if (checkResult.data[0].id_client !== null) {
      error.value =
        'This car has already been assigned to another client. Please refresh and try again.'
      return
    }

    // Ensure both values are set based on currency selection
    const upgrades = carUpgradesTotal.value || 0
    
    if (selectedCurrency.value === 'USD') {
      // User entered USD (FOB), ensure CFR DA is calculated
      formData.value.price_cell = priceInput.value
      cfrDaInput.value = calculateCFRDAFromPrice(
        priceInput.value,
      formData.value.rate,
      formData.value.freight,
      upgrades,
    )
    } else {
      // User entered DZD (CFR DA), ensure price_cell (FOB) is calculated
      cfrDaInput.value = priceInput.value
      formData.value.price_cell = calculatePriceFromCFRDA(
        priceInput.value,
        formData.value.rate,
        formData.value.freight,
        upgrades,
      )
    }

    const cfrDaValue = cfrDaInput.value
    const calculatedPriceCell = formData.value.price_cell

    // Use current notes (managed separately via NotesManagementModal)
    const notesJson = carNotes.value.length > 0 ? JSON.stringify(carNotes.value) : null

    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = ?,
            id_client = ?,
            id_port_discharge = ?,
            cfr_da = ?,
            price_cell = ?,
            freight = ?,
            date_sell = ?,
            id_sell_pi = ?,
            rate = ?,
            notes = ?,
            is_tmp_client = ?,
            date_assigned = NOW()
        WHERE id = ? AND id_client IS NULL
      `,
      params: [
        props.sellBillId,
        formData.value.id_client,
        formData.value.id_port_discharge,
        cfrDaValue, // Store exact CFR DA value
        calculatedPriceCell, // Store calculated price_cell value
        formData.value.freight || null,
        currentDate,
        billRef,
        formData.value.rate || null,
        notesJson,
        formData.value.is_tmp_client,
        props.carId,
      ],
    })

    if (result.success) {
      if (result.affectedRows === 0) {
        error.value =
          'This car has already been assigned to another client. Please refresh and try again.'
      } else {
        emit('assign-success')
        resetForm()
      }
    } else {
      error.value = result.error || 'Failed to assign car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    isSubmitting.value = false
  }
}

// Validate form before submission
const validateForm = () => {
  if (!formData.value.id_client) {
    error.value = 'Please select a client'
    return false
  }

  if (!formData.value.id_port_discharge) {
    error.value = 'Please select a discharge port'
    return false
  }

  if (!priceInput.value) {
    error.value = `Please enter a price in ${selectedCurrency.value}`
    return false
  }

  if (!formData.value.rate) {
    error.value = 'Please enter a rate'
    return false
  }

  return true
}

// Reset form after submission
const resetForm = () => {
  const currentRate = formData.value.rate // Preserve current rate
  formData.value = {
    id_client: null,
    id_port_discharge: null,
    price_cell: null,
    freight: null,
    rate: currentRate, // Keep the default rate
    is_tmp_client: 0,
  }
  selectedCurrency.value = 'DZD' // Reset to default currency
  priceInput.value = null // Reset price input
  cfrDaInput.value = null // Reset CFR DA input
  error.value = null
  carNotes.value = []
  originalCarNotes.value = []
}

// Fetch all users for notes display
const fetchAllUsers = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, username FROM users ORDER BY username ASC',
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
    carNotes.value = JSON.parse(notesJson)
  } else {
    carNotes.value = []
  }
}

// Handle notes updated from modal
const handleNotesUpdated = (updatedNotes) => {
  carNotes.value = updatedNotes
}

// Handle manage notes click
const handleManageNotes = () => {
  showNotesModal.value = true
}

// Handle upgrades click
const handleUpgrades = () => {
  showUpgradesModal.value = true
}

// Close upgrades modal
const closeUpgradesModal = () => {
  showUpgradesModal.value = false
  // Refresh upgrades total when modal closes (in case upgrades were modified)
  if (props.carId) {
    fetchCarUpgrades()
  }
}

// Fetch upgrades for the car
const fetchCarUpgrades = async () => {
  if (!props.carId) {
    carUpgrades.value = []
    carUpgradesTotal.value = 0
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
      params: [props.carId],
    })

    if (result.success) {
      carUpgrades.value = result.data || []
      // Calculate total
      carUpgradesTotal.value = carUpgrades.value.reduce((sum, upgrade) => {
        const value = parseFloat(upgrade.value) || 0
        return sum + value
      }, 0)
    } else {
      carUpgrades.value = []
      carUpgradesTotal.value = 0
    }
  } catch (err) {
    console.error('Error fetching car upgrades:', err)
    carUpgrades.value = []
    carUpgradesTotal.value = 0
  }
}

// Format value for display
const formatUpgradeValue = (value) => {
  if (value === null || value === undefined || value === 0) return '0.00'
  return parseFloat(value).toFixed(2)
}

// Handle close - restore original notes if cancelled
const handleClose = () => {
  // Restore original notes (discard any changes made in the modal)
  carNotes.value = JSON.parse(JSON.stringify(originalCarNotes.value))
  resetForm()
  // Reset upgrades data
  carUpgrades.value = []
  carUpgradesTotal.value = 0
  emit('close')
}

// Calculate potential profit
const calculateProfit = () => {
  if (!carDetails.value?.buy_price || !formData.value.price_cell) return t('sellBills.notAvailable')

  const buyPrice = parseFloat(carDetails.value.buy_price)
  const sellPrice = parseFloat(formData.value.price_cell)
  const freightCost = formData.value.freight ? parseFloat(formData.value.freight) : 0

  return (sellPrice - buyPrice - freightCost).toFixed(2)
}

// New client functionality
const openAddClientDialog = () => {
  showAddClientDialog.value = true
}

const closeAddClientDialog = () => {
  showAddClientDialog.value = false
}

const handleClientSaved = async (newClientData) => {
  // Refresh clients list
  await fetchClients()
  // Wait for Vue to update the DOM
  await nextTick()
  // Auto-select the newly created client
  formData.value.id_client = newClientData.id
  closeAddClientDialog()
}

const openAddDischargePortDialog = () => {
  showAddDischargePortDialog.value = true
}

const closeAddDischargePortDialog = () => {
  showAddDischargePortDialog.value = false
  newDischargePort.value = {
    discharge_port: '',
  }
}

const addDischargePort = async () => {
  if (isSubmitting.value) return

  if (!newDischargePort.value.discharge_port || newDischargePort.value.discharge_port.trim() === '') {
    error.value = t('sellBills.dischargePortNameRequired') || 'Discharge port name is required'
    return
  }

  try {
    isSubmitting.value = true
    const result = await callApi({
      query: 'INSERT INTO discharge_ports (discharge_port) VALUES (?)',
      params: [newDischargePort.value.discharge_port.trim()],
    })

    if (result.success) {
      // Refresh discharge ports list
      await fetchDischargePorts()
      // Wait for Vue to update the DOM
      await nextTick()
      // Auto-select the newly created discharge port
      formData.value.id_port_discharge = result.lastInsertId
      closeAddDischargePortDialog()
    } else {
      error.value = result.error || t('sellBills.failedToAddDischargePort') || 'Failed to add discharge port'
    }
  } catch (err) {
    error.value = err.message || t('sellBills.failedToAddDischargePort') || 'Failed to add discharge port'
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  fetchClients()
  fetchDischargePorts()
  fetchDefaultRate()
  fetchAllUsers()
  if (props.carId) {
    fetchCarDetails()
  }
})

// Watch for visible prop to reload data when modal opens
watch(() => props.visible, (newVal) => {
  if (newVal && props.carId) {
    fetchCarDetails()
  }
})
</script>

<template>
  <div v-if="visible" class="car-assignment-form-overlay">
    <div class="car-assignment-form">
      <!-- Loading Overlay -->
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>{{ isProcessing ? 'Assigning car...' : 'Loading...' }}</span>
      </div>

      <div class="form-header">
        <h3>
          <i class="fas fa-link"></i>
          {{ t('sellBills.assign_car_to_bill') }}
        </h3>
        <button @click="handleClose" class="close-btn">&times;</button>
      </div>

      <div v-if="error" class="error">{{ error }}</div>

      <div v-if="carDetails" class="car-details">
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-car"></i>
            {{ t('sellBills.car') }}:
          </span>
          <span class="value">{{ carDetails.car_name }}</span>
        </div>
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-palette"></i>
            {{ t('sellBills.color') }}:
          </span>
          <span class="value">{{ carDetails.color }}</span>
        </div>
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-fingerprint"></i>
            {{ t('sellBills.vin') }}:
          </span>
          <span class="value">{{ carDetails.vin || t('sellBills.notAvailable') }}</span>
        </div>
        <div v-if="isAdmin" class="detail-item">
          <span class="label">
            <i class="fas fa-dollar-sign"></i>
            {{ t('sellBills.buy_price') }}:
          </span>
          <span class="value">${{ carDetails.buy_price }}</span>
        </div>
      </div>

      <form @submit.prevent="assignCar" class="form-content">
        <div class="form-row">
          <div class="form-group">
            <label for="client">
              <i class="fas fa-user"></i>
              {{ t('sellBills.client') }}: <span class="required">*</span>
            </label>
            <div class="client-select-wrapper">
              <el-select
                v-model="formData.id_client"
                filterable
                remote
                :remote-method="remoteMethod"
                :loading="loading"
                :placeholder="t('sellBills.select_or_search_client')"
                style="width: 100%"
                @change="handleClientChange"
              >
                <el-option
                  v-for="client in filteredClients"
                  :key="client.id"
                  :label="client.name"
                  :value="client.id"
                >
                  <i class="fas fa-user"></i>
                  {{ client.name }}
                  <small v-if="client.mobiles">
                    <i class="fas fa-phone"></i>
                    {{ client.mobiles }}
                  </small>
                </el-option>
              </el-select>
              <el-button type="primary" @click="openAddClientDialog" class="new-client-btn">
                <i class="fas fa-plus"></i>
                {{ t('sellBills.new') }}
              </el-button>
            </div>
          </div>

          <div class="form-group">
            <label for="discharge-port">
              <i class="fas fa-anchor"></i>
              {{ t('sellBills.discharge_port') }}: <span class="required">*</span>
            </label>
            <div class="input-with-button">
              <select id="discharge-port" v-model="formData.id_port_discharge" required>
                <option value="">{{ t('sellBills.select_discharge_port') }}</option>
                <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
                  {{ port.discharge_port }}
                </option>
              </select>
              <button
                type="button"
                @click="openAddDischargePortDialog"
                class="btn-add-port"
                :title="t('sellBills.addDischargePort')"
              >
                <i class="fas fa-plus"></i>
              </button>
            </div>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="currency">
              <i class="fas fa-money-bill-wave"></i>
              {{ t('sellBills.currency') }}: <span class="required">*</span>
            </label>
            <select id="currency" v-model="selectedCurrency" @change="handleCurrencyChange">
              <option value="USD">USD</option>
              <option value="DZD">DZD (DA)</option>
            </select>
          </div>

          <div class="form-group">
            <label for="price-input">
              <i class="fas fa-calculator"></i>
              {{ selectedCurrency === 'USD' ? t('sellBills.price') : t('sellBills.cfr_da') }}: <span class="required">*</span>
            </label>
            <div class="input-with-info">
              <input
                type="number"
                id="price-input"
                v-model="priceInput"
                step="0.01"
                min="0"
                required
                @input="handlePriceInputChange"
                :placeholder="selectedCurrency === 'USD' ? t('sellBills.enterPriceUSD') : t('sellBills.enterPriceDZD')"
              />
            </div>
            <span class="info-text" v-if="selectedCurrency === 'USD' && formData.rate && priceInput">
              {{ t('sellBills.calculatedDZD') }}: {{ cfrDaInput ? parseFloat(cfrDaInput).toLocaleString() : t('sellBills.notAvailable') }} DA
            </span>
            <span class="info-text" v-else-if="selectedCurrency === 'DZD' && formData.rate && priceInput">
              {{ t('sellBills.calculatedUSD') }}: {{ formData.price_cell ? '$' + parseFloat(formData.price_cell).toLocaleString() : t('sellBills.notAvailable') }}
            </span>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="freight">
              <i class="fas fa-ship"></i>
              {{ t('sellBills.freight_cost') }}:
            </label>
            <div class="input-with-info">
              <input
                type="number"
                id="freight"
                v-model="formData.freight"
                step="0.01"
                min="0"
                @input="handleRateOrFreightChange"
                :class="{ 'default-value': formData.freight !== null }"
              />
              <span v-if="formData.freight !== null" class="default-badge">
                {{ t('sellBills.default') }} ({{
                  carDetails?.is_big_car
                    ? t('sellBills.default_big_car')
                    : t('sellBills.default_small_car')
                }})
              </span>
            </div>
          </div>

          <div class="form-group">
            <label for="rate">
              <i class="fas fa-exchange-alt"></i>
              {{ t('sellBills.rate_required') }}:
            </label>
            <div class="input-with-info">
              <input
                type="number"
                id="rate"
                v-model="formData.rate"
                step="0.01"
                min="0"
                required
                @input="handleRateOrFreightChange"
                :class="{ 'default-value': formData.rate !== null }"
              />
              <span v-if="formData.rate !== null" class="default-badge">{{
                t('sellBills.default')
              }}</span>
            </div>
          </div>
        </div>

        <div class="form-row full-width">
          <div class="form-group">
            <label>
              <i class="fas fa-sticky-note"></i>
              {{ t('sellBills.notes') }}:
            </label>
            <div class="notes-section">
              <NotesTable
                v-if="carNotes && carNotes.length > 0"
                :notes="carNotes"
                :users="allUsers"
              />
              <button
                type="button"
                @click="handleManageNotes"
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
                  @click="handleUpgrades"
                  class="btn-manage-upgrades"
                  :disabled="isProcessing || !carId"
                >
                  <i class="fas fa-wrench"></i>
                  {{ t('carStock.upgrades') || 'Manage Upgrades' }}
                </button>
                <div v-if="carUpgrades.length > 0" class="upgrades-total-display">
                  <span class="upgrades-count">
                    <i class="fas fa-list"></i>
                    {{ carUpgrades.length }} {{ t('carUpgrades.total_upgrades') || 'upgrades' }}
                  </span>
                  <span class="upgrades-total">
                    <i class="fas fa-calculator"></i>
                    {{ t('carUpgrades.total_amount') || 'Total' }}: {{ formatUpgradeValue(carUpgradesTotal) }}
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
                v-model="formData.is_tmp_client"
                :true-value="1"
                :false-value="0"
                :disabled="isSubmitting"
              />
              <i class="fas fa-clock"></i>
              {{ t('sellBills.temporary_client') }}
            </label>
            <small class="help-text">{{ t('sellBills.help_text_temporary_client') }}</small>
          </div>
        </div>

        <div
          v-if="formData.price_cell && carDetails?.buy_price && isAdmin"
          class="form-row full-width"
        >
          <div class="profit-calculation">
            <span class="label">
              <i class="fas fa-dollar-sign"></i>
              {{ t('sellBills.estimated_profit') }}:
            </span>
            <span class="value profit">${{ calculateProfit() }}</span>
          </div>
        </div>

        <div class="form-row full-width">
          <div class="form-actions">
            <button type="button" @click="handleClose" class="cancel-btn">
              <i class="fas fa-times"></i>
              {{ t('sellBills.cancel') }}
            </button>
            <button type="submit" class="assign-btn" :disabled="isProcessing">
              <i class="fas fa-save"></i>
              {{ isProcessing ? t('sellBills.assigning_car') : t('sellBills.assign_car') }}
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>

  <!-- Add Client Dialog -->
  <AddClientDialog
    :show="showAddClientDialog"
    @close="closeAddClientDialog"
    @saved="handleClientSaved"
  />

  <!-- Add Discharge Port Dialog -->
  <AddItemDialog
    :show="showAddDischargePortDialog"
    :title="t('sellBills.addDischargePort')"
    :loading="isSubmitting"
    @close="closeAddDischargePortDialog"
    @save="addDischargePort"
  >
    <div class="form-group">
      <label for="discharge_port_name">
        {{ t('sellBills.dischargePortName') }}
        <span class="required">*</span>:
      </label>
      <input
        id="discharge_port_name"
        v-model="newDischargePort.discharge_port"
        type="text"
        :placeholder="t('sellBills.enterDischargePortName')"
        required
        :disabled="isSubmitting"
      />
    </div>
  </AddItemDialog>

  <!-- Notes Management Modal -->
  <NotesManagementModal
    :show="showNotesModal"
    :notes="carNotes"
    :users="allUsers"
    :current-user-id="user?.id"
    :is-admin="isAdmin"
    :entity-id="props.carId"
    :save-function="saveCarNotes"
    :save-immediately="false"
    @close="showNotesModal = false"
    @notes-updated="handleNotesUpdated"
  />

  <!-- Car Upgrades Modal -->
  <CarUpgradesModal
    :show="showUpgradesModal"
    :carId="props.carId"
    :title="carDetails ? `${t('carStock.upgrades') || 'Upgrades'} - ${carDetails.car_name}` : ''"
    :users="allUsers"
    @close="closeUpgradesModal"
  />
</template>

<style scoped>
.car-assignment-form-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
}

.car-assignment-form {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.form-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #666;
}

.car-details {
  background-color: #f9fafb;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  margin-bottom: 8px;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.label {
  font-weight: 600;
  width: 100px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
}

.required {
  color: #ef4444;
}

select,
input {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.9rem;
}

.profit-calculation {
  background-color: #ecfdf5;
  padding: 10px;
  border-radius: 4px;
  margin: 15px 0;
  display: flex;
  align-items: center;
}

.profit {
  color: #10b981;
  font-weight: 600;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 0;
  width: 100%;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}

.assign-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}

.assign-btn:disabled {
  background-color: #93c5fd;
  cursor: not-allowed;
  opacity: 0.6;
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

.error {
  color: #ef4444;
  background-color: #fee2e2;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
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

.form-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1f2937;
  margin-bottom: 1.5rem;
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

.form-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-content .form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.form-content .form-row.full-width {
  grid-column: 1 / -1;
}

@media (max-width: 768px) {
  .form-content .form-row {
    grid-template-columns: 1fr;
  }
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  font-weight: 500;
}

.form-group label i {
  color: #6b7280;
}

.car-details {
  background-color: #f9fafb;
  padding: 1rem;
  border-radius: 0.5rem;
  border: 1px solid #e5e7eb;
}

.car-details p {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.5rem 0;
  color: #374151;
}

.car-details i {
  color: #6b7280;
}

.client-select {
  width: 100%;
}

input[type='number'],
select,
textarea {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

input[type='number']:focus,
select:focus,
textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

textarea {
  min-height: 100px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}

.cancel-btn,
.save-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #059669;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Element Plus Select Customization */
:deep(.el-select-dropdown__item) {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

:deep(.el-select-dropdown__item i) {
  color: #6b7280;
}

:deep(.el-select-dropdown__item small) {
  margin-left: auto;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

/* Add smooth transitions */
.form-group input,
.form-group select,
.form-group textarea {
  transition: all 0.2s;
}

button i {
  transition: transform 0.2s;
}

button:hover:not(:disabled) i {
  transform: scale(1.1);
}

.calculated-value {
  background-color: #f3f4f6;
  color: #1f2937;
  font-weight: 500;
  font-family: monospace;
}

.base-value {
  background-color: #dbeafe;
  color: #1e40af;
  font-weight: 600;
  border: 2px solid #3b82f6;
}

.info-text {
  color: #6b7280;
  font-size: 0.75rem;
  font-style: italic;
  margin-top: 4px;
}

.input-with-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-with-info input {
  flex: 1;
}

.client-select-wrapper {
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
}

.new-client-btn {
  white-space: nowrap;
  height: 32px;
  padding: 0 12px;
}

.input-with-button {
  display: flex;
  gap: 8px;
  align-items: center;
}

.input-with-button select {
  flex: 1;
}

.btn-add-port {
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 38px;
  transition: background-color 0.2s;
}

.btn-add-port:hover {
  background-color: #059669;
}

.file-input {
  width: 100%;
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
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

.calculated-display {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background-color: #fef3c7;
  color: #92400e;
  border: 1px solid #f59e0b;
  border-radius: 4px;
  font-size: 0.875rem;
  margin-top: 4px;
}

.calculated-display i {
  color: #f59e0b;
  font-size: 0.75rem;
}
</style>
