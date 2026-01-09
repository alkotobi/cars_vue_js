<script setup>
import { ref, onMounted, watch, computed, nextTick } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'
import { ElSelect, ElOption } from 'element-plus'
import SelectWithAddButton from './SelectWithAddButton.vue'
import SearchableSelectWithAddButton from './SearchableSelectWithAddButton.vue'
import AddItemDialog from './AddItemDialog.vue'
import AddColorDialog from './AddColorDialog.vue'
import AddClientDialog from './AddClientDialog.vue'

const { t } = useEnhancedI18n()

const props = defineProps({
  carData: {
    type: Object,
    default: null,
  },
  show: {
    type: Boolean,
    default: false,
  },
})

const user = ref(null)
const can_edit_vin = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_vin')
})

const can_edit_car_client_name = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_car_client_name')
})

const can_edit_cars_sell_price = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_sell_price')
})

const can_edit_cars_sell_rate = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_sell_rate')
})

const can_edit_cars_discharge_port = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_discharge_port')
})

const can_confirm_payment = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_confirm_payment')
})

const emit = defineEmits(['save', 'cancel'])

const { callApi, uploadFile, getFileUrl } = useApi()
const loading = ref(false)
const error = ref(null)

// File upload refs
const documentsFile = ref(null)
const sellPiFile = ref(null)
const buyPiFile = ref(null)
const piFile = ref(null)

// Reference data
const clients = ref([])
const loadingPorts = ref([])
const dischargePorts = ref([])
const buyDetails = ref([])
const warehouses = ref([])
const colors = ref([])
const suppliers = ref([])
const carNames = ref([])
const brands = ref([])
const defaults = ref(null)
const showAddSupplierDialog = ref(false)
const newSupplier = ref({
  name: '',
  contact_info: '',
  notes: '',
})
const showAddCarNameDialog = ref(false)
const newCarName = ref({
  car_name: '',
  id_brand: null,
  notes: '',
})
const showAddBrandDialog = ref(false)
const newBrand = ref({
  brand: '',
})
const showAddColorDialog = ref(false)
const showAddClientDialog = ref(false)

// Buy bill selection mode: 'new' or 'existing'
const buyBillMode = ref('new')
const existingBuyBills = ref([])
const selectedBuyBillId = ref(null)

const formData = ref({
  id: null,
  notes: '',
  id_buy_details: null,
  date_sell: null,
  id_client: null,
  price_cell: null,
  freight: null,
  id_port_loading: null,
  id_port_discharge: null,
  vin: '',
  path_documents: '',
  date_loding: null,
  date_send_documents: null,
  id_sell_pi: '',
  sell_pi_path: '',
  buy_pi_path: '',
  id_sell: null,
  export_lisence_ref: '',
  id_warehouse: null,
  in_wharhouse_date: null,
  date_get_documents_from_supp: null,
  date_get_keys_from_supp: null,
  rate: null,
  payment_confirmed: false,
  is_used_car: 0,
  is_big_car: 0,
  id_color: null,
  // New fields for buy bill creation (only for new cars)
  id_supplier: null,
  bill_ref: '',
  pi_path: '',
  buy_bill_notes: '',
  id_car_name: null,
  purchase_price: null,
  price_sell: null,
})

// Handle file changes
const handleFileChange = (event, fileType) => {
  const file = event.target.files[0]
  if (!file) return

  switch (fileType) {
    case 'documents':
      documentsFile.value = file
      break
    case 'sellPi':
      sellPiFile.value = file
      break
    case 'buyPi':
      buyPiFile.value = file
      break
    case 'pi':
      piFile.value = file
      break
  }
}

// Generate bill_ref for new car
const generateBillRef = () => {
  const date = new Date()
  const year = String(date.getFullYear()).slice(-2) // Last 2 digits of year
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  const userId = user.value?.id || '0'
  const billRef = `INTERNAL-${year}${month}${day}-${hours}${minutes}${seconds}-${userId}`
  return billRef
}

// Upload a single file and return its path
const uploadSingleFile = async (file, folder, customFilename = '') => {
  if (!file) return null

  try {
    // Get original file extension
    const originalExt = file.name.split('.').pop().toLowerCase()

    // Create a meaningful filename based on car details
    let filename = ''
    if (formData.value.vin) {
      // If VIN exists, use it as the primary identifier
      filename = `${formData.value.vin}`
    } else if (formData.value.id) {
      // If no VIN but ID exists, use car ID
      filename = `car_${formData.value.id}`
    } else {
      // Fallback to timestamp if neither exists
      filename = `new_car_${Date.now()}`
    }

    // Add folder-specific prefix
    switch (folder) {
      case 'documents':
        filename = `docs_${filename}`
        break
      case 'sell_pi':
        filename = `sell_pi_${filename}`
        break
      case 'buy_pi':
        filename = `buy_pi_${filename}`
        break
    }

    // Add date for versioning
    const date = new Date()
    const dateStr = `${date.getFullYear()}${String(date.getMonth() + 1).padStart(2, '0')}${String(date.getDate()).padStart(2, '0')}`
    filename = `${filename}_${dateStr}.${originalExt}`

    const result = await uploadFile(file, folder, filename)
    if (result.success) {
      return result.relativePath
    }
    throw new Error(result.message || 'Upload failed')
  } catch (err) {
    console.error(`Error uploading file to ${folder}:`, err)
    throw err
  }
}

const fetchReferenceData = async () => {
  try {
    // Fetch clients
    const clientsResult = await callApi({
      query: 'SELECT id, name FROM clients ORDER BY name ASC',
      params: [],
    })
    if (clientsResult.success) {
      clients.value = clientsResult.data
    }

    // Fetch loading ports
    const loadingPortsResult = await callApi({
      query: 'SELECT id, loading_port FROM loading_ports ORDER BY loading_port ASC',
      params: [],
    })
    if (loadingPortsResult.success) {
      loadingPorts.value = loadingPortsResult.data
    }

    // Fetch discharge ports
    const dischargePortsResult = await callApi({
      query: 'SELECT id, discharge_port FROM discharge_ports ORDER BY discharge_port ASC',
      params: [],
    })
    if (dischargePortsResult.success) {
      dischargePorts.value = dischargePortsResult.data
    }

    // Fetch buy details
    const buyDetailsResult = await callApi({
      query: `
        SELECT bd.id, cn.car_name, clr.color 
        FROM buy_details bd
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        ORDER BY bd.id DESC
      `,
      params: [],
    })
    if (buyDetailsResult.success) {
      buyDetails.value = buyDetailsResult.data
    }

    // Fetch existing buy bills (only for new cars)
    if (!props.carData?.id) {
      const buyBillsResult = await callApi({
        query: `
          SELECT bb.id, bb.bill_ref, bb.date_buy, s.name as supplier_name
          FROM buy_bill bb
          LEFT JOIN suppliers s ON bb.id_supplier = s.id
          ORDER BY bb.date_buy DESC, bb.id DESC
        `,
        params: [],
      })
      if (buyBillsResult.success) {
        existingBuyBills.value = buyBillsResult.data
      }
    }

    // Fetch warehouses if needed
    const warehousesResult = await callApi({
      query: 'SELECT id, warhouse_name as name FROM warehouses ORDER BY warhouse_name ASC',
      params: [],
    })
    if (warehousesResult.success) {
      warehouses.value = warehousesResult.data
    }

    // Fetch colors
    const colorsResult = await callApi({
      query: 'SELECT id, color, hexa FROM colors ORDER BY color ASC',
      params: [],
    })
    if (colorsResult.success) {
      colors.value = colorsResult.data
    }

    // Fetch suppliers and car names (only for new car creation)
    if (!formData.value.id) {
      const suppliersResult = await callApi({
        query: 'SELECT id, name FROM suppliers ORDER BY name ASC',
        params: [],
      })
      if (suppliersResult.success) {
        suppliers.value = suppliersResult.data
      }

      const carNamesResult = await callApi({
        query: 'SELECT id, car_name FROM cars_names ORDER BY car_name ASC',
        params: [],
      })
      if (carNamesResult.success) {
        carNames.value = carNamesResult.data
      }

      // Fetch brands
      const brandsResult = await callApi({
        query: 'SELECT id, brand FROM brands ORDER BY brand ASC',
        params: [],
      })
      if (brandsResult.success) {
        brands.value = brandsResult.data
      }
    }

    // Fetch defaults (for rate and freight)
    const defaultsResult = await callApi({
      query: 'SELECT rate, freight_small, freight_big FROM defaults LIMIT 1',
      params: [],
    })
    if (defaultsResult.success && defaultsResult.data && defaultsResult.data.length > 0) {
      defaults.value = defaultsResult.data[0]
    }
  } catch (err) {
    error.value = t('carStockForm.failedToFetchReferenceData')
  }
}

const addSupplier = async () => {
  if (!newSupplier.value.name || newSupplier.value.name.trim() === '') {
    error.value = t('carStockForm.supplierNameRequired') || 'Supplier name is required'
    return
  }

  const result = await callApi({
    query: `
      INSERT INTO suppliers (name, contact_info, notes)
      VALUES (?, ?, ?)
    `,
    params: [
      newSupplier.value.name.trim(),
      newSupplier.value.contact_info || null,
      newSupplier.value.notes || null,
    ],
  })

  if (result.success) {
    // Refresh suppliers list
    const suppliersResult = await callApi({
      query: 'SELECT id, name FROM suppliers ORDER BY name ASC',
      params: [],
    })
    if (suppliersResult.success) {
      suppliers.value = suppliersResult.data
      // Auto-select the newly created supplier
      formData.value.id_supplier = result.lastInsertId
    }

    // Reset form and close dialog
    newSupplier.value = {
      name: '',
      contact_info: '',
      notes: '',
    }
    showAddSupplierDialog.value = false
    error.value = null
  } else {
    error.value = result.error || t('carStockForm.failedToAddSupplier') || 'Failed to add supplier'
  }
}

const openAddSupplierDialog = () => {
  newSupplier.value = {
    name: '',
    contact_info: '',
    notes: '',
  }
  showAddSupplierDialog.value = true
}

const closeAddSupplierDialog = () => {
  showAddSupplierDialog.value = false
  newSupplier.value = {
    name: '',
    contact_info: '',
    notes: '',
  }
}

const addCarName = async () => {
  if (!newCarName.value.car_name || newCarName.value.car_name.trim() === '') {
    error.value = t('carStockForm.carNameRequired') || 'Car name is required'
    return
  }

  const result = await callApi({
    query: `
      INSERT INTO cars_names (car_name, id_brand, notes)
      VALUES (?, ?, ?)
    `,
    params: [
      newCarName.value.car_name.trim(),
      newCarName.value.id_brand || null,
      newCarName.value.notes || null,
    ],
  })

  if (result.success) {
    // Refresh car names list
    const carNamesResult = await callApi({
      query: 'SELECT id, car_name FROM cars_names ORDER BY car_name ASC',
      params: [],
    })
    if (carNamesResult.success) {
      carNames.value = carNamesResult.data
      // Auto-select the newly created car name
      formData.value.id_car_name = result.lastInsertId
    }

    // Reset form and close dialog
    newCarName.value = {
      car_name: '',
      id_brand: null,
      notes: '',
    }
    showAddCarNameDialog.value = false
    error.value = null
  } else {
    error.value = result.error || t('carStockForm.failedToAddCarName') || 'Failed to add car name'
  }
}

const openAddCarNameDialog = () => {
  newCarName.value = {
    car_name: '',
    id_brand: null,
    notes: '',
  }
  showAddCarNameDialog.value = true
}

const closeAddCarNameDialog = () => {
  showAddCarNameDialog.value = false
  newCarName.value = {
    car_name: '',
    id_brand: null,
    notes: '',
  }
}

const addBrand = async () => {
  if (!newBrand.value.brand || newBrand.value.brand.trim() === '') {
    error.value = t('carStockForm.brandRequired') || 'Brand name is required'
    return
  }

  const result = await callApi({
    query: `
      INSERT INTO brands (brand)
      VALUES (?)
    `,
    params: [newBrand.value.brand.trim()],
  })

  if (result.success) {
    // Refresh brands list
    const brandsResult = await callApi({
      query: 'SELECT id, brand FROM brands ORDER BY brand ASC',
      params: [],
    })
    if (brandsResult.success) {
      brands.value = brandsResult.data
      // Auto-select the newly created brand
      newCarName.value.id_brand = result.lastInsertId
    }

    // Reset form and close dialog
    newBrand.value = {
      brand: '',
    }
    showAddBrandDialog.value = false
    error.value = null
  } else {
    error.value = result.error || t('carStockForm.failedToAddBrand') || 'Failed to add brand'
  }
}

const openAddBrandDialog = () => {
  newBrand.value = {
    brand: '',
  }
  showAddBrandDialog.value = true
}

const closeAddBrandDialog = () => {
  showAddBrandDialog.value = false
  newBrand.value = {
    brand: '',
  }
}

const openAddColorDialog = () => {
  showAddColorDialog.value = true
}

const closeAddColorDialog = () => {
  showAddColorDialog.value = false
}

const handleColorSaved = async (newColor) => {
  // Refresh colors list
  const colorsResult = await callApi({
    query: 'SELECT id, color, hexa FROM colors ORDER BY color ASC',
    params: [],
  })
  if (colorsResult.success) {
    colors.value = colorsResult.data
    // Auto-select the newly created color
    formData.value.id_color = newColor.id
  }
}

const openAddClientDialog = () => {
  showAddClientDialog.value = true
}

const closeAddClientDialog = () => {
  showAddClientDialog.value = false
}

const handleBuyBillModeChange = (mode) => {
  if (mode === 'existing') {
    formData.value.id_supplier = null
    formData.value.bill_ref = ''
    formData.value.buy_bill_notes = ''
    piFile.value = null
  } else {
    selectedBuyBillId.value = null
  }
}

const handleClientSaved = async (newClient) => {
  // Refresh clients list
  const clientsResult = await callApi({
    query: 'SELECT id, name FROM clients ORDER BY name ASC',
    params: [],
  })
  if (clientsResult.success) {
    clients.value = clientsResult.data
    // Auto-select the newly created client
    formData.value.id_client = newClient.id
  }
  closeAddClientDialog()
}

// Check if this is a shipping only car
const isShippingOnly = computed(() => {
  return props.carData?.is_shipping_only === true
})

// Watch for is_big_car changes to update freight
watch(
  () => formData.value.is_big_car,
  (isBigCar) => {
    if (defaults.value && !formData.value.id) {
      // Only update for new cars
      formData.value.freight =
        isBigCar === 1 ? defaults.value.freight_big : defaults.value.freight_small
    }
  },
)

// Watch for changes in carData prop
watch(
  () => props.carData,
  async (newData) => {
    // Only update formData if carData has an id (editing existing car)
    if (newData && newData.id) {
      // Create a new object with all fields from newData
      formData.value = {
        ...newData,
        // Convert rate to number if it exists, otherwise null
        rate: newData.rate ? Number(newData.rate) : null,
        // Convert payment_confirmed from database format (0/1) to boolean
        payment_confirmed: Boolean(newData.payment_confirmed),
      }

      // Format dates for input fields
      const dateFields = [
        'date_send_documents',
        'in_wharhouse_date',
        'date_get_documents_from_supp',
        'date_get_keys_from_supp',
      ]

      dateFields.forEach((field) => {
        if (formData.value[field]) {
          const date = new Date(formData.value[field])
          if (!isNaN(date)) {
            formData.value[field] = date.toISOString().split('T')[0]
          }
        }
      })
    } else {
      // If carData is null or has no id, don't overwrite formData (let the show watch handle it)
    }
  },
  { immediate: true },
)

// Watch for show prop to reset form when opening for new car
watch(
  () => props.show,
  async (isVisible) => {
    // Check if carData is null OR if carData.id is null (new car)
    if (isVisible && (!props.carData || !props.carData.id)) {
      // Ensure user is loaded before generating bill_ref
      if (!user.value) {
        const userStr = localStorage.getItem('user')
        if (userStr) {
          user.value = JSON.parse(userStr)
        }
      }

      // Generate bill_ref immediately with current user (or '0' if not loaded yet)
      const billRef = generateBillRef()

      // Reset form for new car
      const isShippingOnlyCar = props.carData?.is_shipping_only === true
      // Default to small car (is_big_car = 0)
      const isBigCar = props.carData?.is_big_car === 1 ? 1 : 0
      // Get freight and rate from defaults (default to small car freight)
      const defaultFreight = defaults.value
        ? isBigCar === 1
          ? defaults.value.freight_big
          : defaults.value.freight_small
        : null
      const defaultRate = defaults.value ? defaults.value.rate : null

      formData.value = {
        id: null,
        notes: '',
        id_buy_details: null,
        price_cell: isShippingOnlyCar ? 0 : null,
        freight: defaultFreight,
        vin: '',
        path_documents: '',
        date_send_documents: null,
        id_sell_pi: '',
        sell_pi_path: '',
        buy_pi_path: '',
        id_sell: null,
        export_lisence_ref: '',
        id_warehouse: null,
        in_wharhouse_date: null,
        date_get_documents_from_supp: null,
        date_get_keys_from_supp: null,
        rate: defaultRate,
        is_used_car: 0,
        is_big_car: isBigCar,
        id_color: null,
        // New fields for buy bill creation
        id_supplier: null,
        bill_ref: billRef, // Set bill_ref immediately - will be visible right away
        pi_path: '',
        buy_bill_notes: '',
        id_car_name: null,
        purchase_price: isShippingOnlyCar ? 0 : null,
        price_sell: isShippingOnlyCar ? 0 : null,
        id_client: null,
        id_port_loading: null,
        id_port_discharge: null,
      }
      documentsFile.value = null
      sellPiFile.value = null
      buyPiFile.value = null
      piFile.value = null

      // Reset buy bill mode for new cars
      buyBillMode.value = 'new'
      selectedBuyBillId.value = null

      // Use nextTick to ensure Vue reactivity updates the input field
      nextTick(() => {
        // Force update if still empty (double-check)
        if (!formData.value.bill_ref || formData.value.bill_ref === '') {
          formData.value.bill_ref = generateBillRef()
        }
      })

      // Fetch reference data for new car (suppliers, car names, colors, defaults)
      await fetchReferenceData()

      // After defaults are loaded, update freight and rate
      if (defaults.value) {
        const isBigCar = formData.value.is_big_car === 1
        formData.value.freight = isBigCar
          ? defaults.value.freight_big
          : defaults.value.freight_small
        formData.value.rate = defaults.value.rate
      }
    }
  },
  { immediate: true }, // Run immediately when component mounts
)

// Watch for user to be loaded and regenerate bill_ref if needed
watch(
  () => user.value?.id,
  (userId) => {
    if (
      userId &&
      props.show &&
      (!props.carData || !props.carData.id) &&
      (!formData.value.bill_ref ||
        formData.value.bill_ref === '' ||
        !formData.value.bill_ref.includes(`-${userId}`))
    ) {
      formData.value.bill_ref = generateBillRef()
    }
  },
)

// Upload PI file for buy bill
const uploadPiFile = async () => {
  if (!piFile.value) return null

  try {
    const date = new Date()
    const timestamp = date.getTime()
    const dateStr = `${date.getFullYear()}${String(date.getMonth() + 1).padStart(2, '0')}${String(date.getDate()).padStart(2, '0')}`
    const timeStr = `${String(date.getHours()).padStart(2, '0')}-${String(date.getMinutes()).padStart(2, '0')}-${String(date.getSeconds()).padStart(2, '0')}`

    const fileExtension = piFile.value.name.split('.').pop().toLowerCase()
    const supplierName =
      suppliers.value.find((s) => s.id === formData.value.id_supplier)?.name || 'supplier'
    const sanitizedSupplierName = supplierName.replace(/[^a-zA-Z0-9]/g, '_')
    const filename = `pi_purchase_${sanitizedSupplierName}_${dateStr}_${timeStr}_${timestamp}.${fileExtension}`

    const result = await uploadFile(piFile.value, 'purchase_pi', filename)
    if (result.success) {
      return `purchase_pi/${filename}`
    }
    throw new Error(result.message || 'Upload failed')
  } catch (err) {
    console.error('Error uploading PI file:', err)
    throw err
  }
}

const saveCar = async () => {
  loading.value = true
  error.value = null

  try {
    // Determine if this is a new car (INSERT) or existing car (UPDATE)
    const isNewCar = !formData.value.id

    let result
    if (isNewCar) {
      // Validate required fields for new car
      if (!formData.value.id_car_name) {
        error.value = t('carStockForm.carNameRequired') || 'Car Name is required'
        loading.value = false
        return
      }
      if (!formData.value.id_color) {
        error.value = t('carStockForm.colorRequired') || 'Color is required'
        loading.value = false
        return
      }
      // Validation for shipping only cars
      if (isShippingOnly.value) {
        if (!formData.value.id_client) {
          error.value =
            t('carStockForm.clientRequired') || 'Client is required for shipping only cars'
          loading.value = false
          return
        }
        if (
          formData.value.freight === null ||
          formData.value.freight === undefined ||
          formData.value.freight === ''
        ) {
          error.value =
            t('carStockForm.freightRequired') || 'Freight is required for shipping only cars'
          loading.value = false
          return
        }
        if (!formData.value.id_port_loading) {
          error.value =
            t('carStockForm.loadingPortRequired') ||
            'Loading Port is required for shipping only cars'
          loading.value = false
          return
        }
        if (!formData.value.id_port_discharge) {
          error.value =
            t('carStockForm.dischargePortRequired') ||
            'Discharge Port is required for shipping only cars'
          loading.value = false
          return
        }
        if (
          formData.value.rate === null ||
          formData.value.rate === undefined ||
          formData.value.rate === ''
        ) {
          error.value = t('carStockForm.rateRequired') || 'Rate is required for shipping only cars'
          loading.value = false
          return
        }
      } else {
        // Validation for regular cars
        if (!formData.value.purchase_price) {
          error.value = t('carStockForm.purchasePriceRequired') || 'Purchase Price is required'
          loading.value = false
          return
        }
        if (!formData.value.price_sell) {
          error.value = t('carStockForm.priceSellRequired') || 'Price Sell is required'
          loading.value = false
          return
        }
      }

      // Validate buy bill selection
      let buyBillId = null
      if (buyBillMode.value === 'existing') {
        if (!selectedBuyBillId.value) {
          error.value = t('carStockForm.buyBillRequired') || 'Please select an existing buy bill'
          loading.value = false
          return
        }
        buyBillId = selectedBuyBillId.value
      } else {
        // New buy bill validation
        if (!formData.value.id_supplier) {
          error.value = t('carStockForm.supplierRequired') || 'Supplier is required'
          loading.value = false
          return
        }
        if (!formData.value.bill_ref || formData.value.bill_ref.trim() === '') {
          error.value = t('carStockForm.billRefRequired') || 'Bill Reference is required'
          loading.value = false
          return
        }

        // Step 1: Upload PI file if provided
        let piPath = null
        if (piFile.value) {
          piPath = await uploadPiFile()
        }

        // Step 2: Create buy_bill
        const buyBillResult = await callApi({
          query: `
            INSERT INTO buy_bill (id_supplier, date_buy, bill_ref, pi_path, is_ordered, is_stock_updated, notes)
            VALUES (?, NOW(), ?, ?, 1, 1, ?)
          `,
          params: [
            formData.value.id_supplier,
            formData.value.bill_ref.trim(),
            piPath,
            formData.value.buy_bill_notes || null,
          ],
        })

        if (!buyBillResult.success) {
          throw new Error(buyBillResult.error || t('carStockForm.failedToCreateBuyBill'))
        }

        buyBillId = buyBillResult.lastInsertId
      }

      // Step 3: Create buy_detail
      const currentDate = new Date()
      const buyDetailResult = await callApi({
        query: `
          INSERT INTO buy_details 
          (id_car_name, id_color, amount, QTY, year, month, is_used_car, id_buy_bill, price_sell, is_big_car, notes)
          VALUES (?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?)
        `,
        params: [
          formData.value.id_car_name,
          formData.value.id_color,
          isShippingOnly.value ? 0 : formData.value.purchase_price,
          currentDate.getFullYear(),
          currentDate.getMonth() + 1,
          formData.value.is_used_car || 0,
          buyBillId,
          isShippingOnly.value ? 0 : formData.value.price_sell,
          formData.value.is_big_car || 0,
          formData.value.notes || null,
        ],
      })

      if (!buyDetailResult.success) {
        throw new Error(buyDetailResult.error || t('carStockForm.failedToCreateBuyDetail'))
      }

      const buyDetailId = buyDetailResult.lastInsertId

      // Step 4: Create car in cars_stock
      result = await callApi({
        query: `
          INSERT INTO cars_stock
          (id_buy_details, price_cell, notes, is_used_car, is_big_car, id_color, id_client, id_port_loading, id_port_discharge, freight, rate)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `,
        params: [
          buyDetailId,
          isShippingOnly.value ? 0 : formData.value.price_sell,
          formData.value.notes || null,
          formData.value.is_used_car || 0,
          formData.value.is_big_car || 0,
          formData.value.id_color,
          formData.value.id_client || null,
          formData.value.id_port_loading || null,
          formData.value.id_port_discharge || null,
          formData.value.freight || null,
          formData.value.rate || null,
        ],
      })
    } else {
      // UPDATE existing car
      result = await callApi({
        query: `
        UPDATE cars_stock
        SET 
          notes = ?,
          id_buy_details = ?,
          price_cell = ?,
          freight = ?,
          vin = ?,
          path_documents = ?,
          date_send_documents = ?,
          id_sell_pi = ?,
          sell_pi_path = ?,
          buy_pi_path = ?,
          id_sell = ?,
          export_lisence_ref = ?,
          id_warehouse = ?,
          in_wharhouse_date = ?,
          date_get_documents_from_supp = ?,
          date_get_keys_from_supp = ?,
            rate = ?
        WHERE id = ?
      `,
        params: [
          formData.value.notes || null,
          formData.value.id_buy_details || null,
          formData.value.price_cell || null,
          formData.value.freight || null,
          formData.value.vin || null,
          formData.value.path_documents || null,
          formData.value.date_send_documents || null,
          formData.value.id_sell_pi || null,
          formData.value.sell_pi_path || null,
          formData.value.buy_pi_path || null,
          formData.value.id_sell || null,
          formData.value.export_lisence_ref || null,
          formData.value.id_warehouse || null,
          formData.value.in_wharhouse_date || null,
          formData.value.date_get_documents_from_supp || null,
          formData.value.date_get_keys_from_supp || null,
          formData.value.rate || null,
          formData.value.id,
        ],
      })
    }

    if (result.success) {
      // Emit the car ID if it's a new car, or the existing ID for updates
      const carId = isNewCar ? result.lastInsertId : formData.value.id
      emit('save', carId)
    } else {
      throw new Error(result.error || t('carStockForm.failedToSaveCar'))
    }
  } catch (err) {
    error.value = err.message
    console.error('Error saving car:', err)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  await fetchReferenceData()

  // If form is already visible for new car, update freight and rate from defaults
  if (props.show && !formData.value.id && defaults.value) {
    const isBigCar = formData.value.is_big_car === 1
    formData.value.freight = isBigCar ? defaults.value.freight_big : defaults.value.freight_small
    formData.value.rate = defaults.value.rate
  }

  // If form is already visible for new car, regenerate bill_ref with user ID
  if (
    props.show &&
    (!props.carData || !props.carData.id) &&
    (!formData.value.bill_ref || formData.value.bill_ref === '')
  ) {
    formData.value.bill_ref = generateBillRef()
  }
})
</script>

<template>
  <div class="car-stock-form">
    <h3>{{ formData.id ? t('carStockForm.editCarStock') : t('carStockForm.addCarStock') }}</h3>

    <div class="form-grid">
      <!-- Buy Bill Information (only for new cars) -->
      <div v-if="!formData.id" class="form-section">
        <h4>{{ t('carStockForm.buyBillInformation') || 'Buy Bill Information' }}</h4>

        <!-- Buy Bill Mode Selection -->
        <div class="form-group">
          <label>{{ t('carStockForm.buyBillMode') || 'Buy Bill' }}:</label>
          <div class="radio-group">
            <label class="radio-label">
              <input
                type="radio"
                v-model="buyBillMode"
                value="new"
                @change="handleBuyBillModeChange('new')"
              />
              <span>{{ t('carStockForm.createNewBuyBill') || 'Create New Buy Bill' }}</span>
            </label>
            <label class="radio-label">
              <input
                type="radio"
                v-model="buyBillMode"
                value="existing"
                @change="handleBuyBillModeChange('existing')"
              />
              <span>{{
                t('carStockForm.addToExistingBuyBill') || 'Add to Existing Buy Bill'
              }}</span>
            </label>
          </div>
        </div>

        <!-- Existing Buy Bill Selection -->
        <div v-if="buyBillMode === 'existing'" class="form-group">
          <label for="selected_buy_bill"
            >{{ t('carStockForm.selectBuyBill') || 'Select Buy Bill' }}
            <span class="required">*</span>:</label
          >
          <el-select
            id="selected_buy_bill"
            v-model="selectedBuyBillId"
            :placeholder="
              t('carStockForm.selectBuyBillPlaceholder') || 'Select an existing buy bill'
            "
            filterable
            style="width: 100%"
            :required="true"
          >
            <el-option
              v-for="bill in existingBuyBills"
              :key="bill.id"
              :label="`${bill.bill_ref || 'N/A'} - ${bill.supplier_name || 'No Supplier'} (${new Date(bill.date_buy).toLocaleDateString()})`"
              :value="bill.id"
            />
          </el-select>
        </div>

        <!-- New Buy Bill Fields -->
        <template v-if="buyBillMode === 'new'">
          <div class="form-group">
            <label for="id_supplier"
              >{{ t('carStockForm.supplier') || 'Supplier' }}
              <span class="required">*</span>:</label
            >
            <SearchableSelectWithAddButton
              id="id_supplier"
              v-model="formData.id_supplier"
              :options="suppliers"
              option-value="id"
              option-label="name"
              :placeholder="t('carStockForm.selectSupplier') || 'Select Supplier'"
              :required="true"
              :add-button-title="t('carStockForm.addSupplier') || 'Add New Supplier'"
              @add="openAddSupplierDialog"
            />
          </div>

          <div class="form-group">
            <label for="bill_ref"
              >{{ t('carStockForm.billReference') || 'Bill Reference' }}
              <span class="required">*</span>:</label
            >
            <input
              type="text"
              id="bill_ref"
              v-model="formData.bill_ref"
              required
              :placeholder="t('carStockForm.billRefPlaceholder') || 'Bill Reference'"
            />
          </div>

          <div class="form-group">
            <label for="pi_file"
              >{{ t('carStockForm.piDocument') || 'PI Document' }} ({{
                t('carStockForm.pdfOrImage') || 'PDF or Image'
              }}):</label
            >
            <input
              type="file"
              id="pi_file"
              @change="handleFileChange($event, 'pi')"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp"
            />
          </div>

          <div class="form-group">
            <label for="buy_bill_notes"
              >{{ t('carStockForm.buyBillNotes') || 'Buy Bill Notes' }}:</label
            >
            <textarea id="buy_bill_notes" v-model="formData.buy_bill_notes"></textarea>
          </div>
        </template>
      </div>

      <!-- Basic Information -->
      <div class="form-section">
        <h4>{{ t('carStockForm.basicInformation') }}</h4>

        <!-- Car Name dropdown (only for new cars) -->
        <div v-if="!formData.id" class="form-group">
          <label for="id_car_name"
            >{{ t('carStockForm.carName') || 'Car Name' }} <span class="required">*</span>:</label
          >
          <SearchableSelectWithAddButton
            id="id_car_name"
            v-model="formData.id_car_name"
            :options="carNames"
            option-value="id"
            option-label="car_name"
            :placeholder="t('carStockForm.selectCarName') || 'Select Car Name'"
            :required="true"
            :add-button-title="t('carStockForm.addCarName') || 'Add New Car Name'"
            @add="openAddCarNameDialog"
          />
        </div>

        <div class="form-group">
          <label for="vin">{{ t('carStockForm.vin') }}:</label>
          <input :disabled="!can_edit_vin" type="text" id="vin" v-model="formData.vin" />
        </div>

        <!-- Buy Details dropdown (only for existing cars) -->
        <div v-if="formData.id" class="form-group">
          <label for="id_buy_details">{{ t('carStockForm.buyDetails') }}:</label>
          <select id="id_buy_details" v-model="formData.id_buy_details">
            <option value="">{{ t('carStockForm.selectBuyDetails') }}</option>
            <option v-for="detail in buyDetails" :key="detail.id" :value="detail.id">
              {{ detail.car_name }} - {{ detail.color }} (ID: {{ detail.id }})
            </option>
          </select>
        </div>

        <div class="form-group">
          <label for="id_color"
            >{{ t('carStockForm.color') || 'Color' }} <span class="required">*</span>:</label
          >
          <SelectWithAddButton
            id="id_color"
            v-model="formData.id_color"
            :options="colors"
            option-value="id"
            option-label="color"
            :placeholder="t('carStockForm.selectColor') || 'Select Color'"
            :required="true"
            :add-button-title="t('carStockForm.addColor') || 'Add New Color'"
            @add="openAddColorDialog"
          />
        </div>

        <div class="form-group">
          <label for="notes">{{ t('carStockForm.notes') }}:</label>
          <textarea id="notes" v-model="formData.notes"></textarea>
        </div>

        <div class="form-group">
          <div class="checkbox-wrapper">
            <input
              type="checkbox"
              id="is_big_car"
              v-model="formData.is_big_car"
              :true-value="1"
              :false-value="0"
            />
            <label for="is_big_car" class="checkbox-label">{{
              t('carStockForm.bigCar') || 'Big Car'
            }}</label>
          </div>
        </div>
      </div>

      <!-- Pricing and Client -->
      <div class="form-section">
        <h4>{{ t('carStockForm.pricingAndClient') }}</h4>

        <!-- Purchase Price (only for new cars, not shipping only) -->
        <div v-if="!formData.id && !isShippingOnly" class="form-group">
          <label for="purchase_price"
            >{{ t('carStockForm.purchasePrice') || 'Purchase Price' }}
            <span class="required">*</span>:</label
          >
          <input
            type="number"
            id="purchase_price"
            v-model="formData.purchase_price"
            step="0.01"
            required
          />
        </div>

        <!-- Price Sell (only for new cars, not shipping only) -->
        <div v-if="!formData.id && !isShippingOnly" class="form-group">
          <label for="price_sell"
            >{{ t('carStockForm.priceSell') || 'Price Sell' }}
            <span class="required">*</span>:</label
          >
          <input type="number" id="price_sell" v-model="formData.price_sell" step="0.01" required />
        </div>

        <!-- Client Name (for shipping only) -->
        <div v-if="!formData.id && isShippingOnly" class="form-group">
          <label for="id_client"
            >{{ t('carStockForm.client') || 'Client' }} <span class="required">*</span>:</label
          >
          <div class="input-with-button">
            <el-select
              id="id_client"
              v-model="formData.id_client"
              :placeholder="t('carStockForm.selectClient') || 'Select Client'"
              filterable
              style="width: 100%"
            >
              <el-option
                v-for="client in clients"
                :key="client.id"
                :label="client.name"
                :value="client.id"
              />
            </el-select>
            <button
              type="button"
              @click="openAddClientDialog"
              class="btn-add-supplier"
              title="Add New Client"
            >
              <i class="fas fa-plus"></i>
            </button>
          </div>
        </div>

        <!-- Loading Port (for shipping only) -->
        <div v-if="!formData.id && isShippingOnly" class="form-group">
          <label for="id_port_loading"
            >{{ t('carStockForm.loadingPort') || 'Loading Port' }}
            <span class="required">*</span>:</label
          >
          <select id="id_port_loading" v-model="formData.id_port_loading">
            <option :value="null">
              {{ t('carStockForm.selectLoadingPort') || 'Select Loading Port' }}
            </option>
            <option v-for="port in loadingPorts" :key="port.id" :value="port.id">
              {{ port.loading_port }}
            </option>
          </select>
        </div>

        <!-- Discharge Port (for shipping only) -->
        <div v-if="!formData.id && isShippingOnly" class="form-group">
          <label for="id_port_discharge"
            >{{ t('carStockForm.dischargePort') || 'Discharge Port' }}
            <span class="required">*</span>:</label
          >
          <select id="id_port_discharge" v-model="formData.id_port_discharge">
            <option :value="null">
              {{ t('carStockForm.selectDischargePort') || 'Select Discharge Port' }}
            </option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
              {{ port.discharge_port }}
            </option>
          </select>
        </div>

        <!-- Freight (for all new cars, required for shipping only) -->
        <div v-if="!formData.id" class="form-group">
          <label for="freight"
            >{{ t('carStockForm.freight') || 'Freight' }}
            <span v-if="isShippingOnly" class="required">*</span>:</label
          >
          <input type="number" id="freight" v-model="formData.freight" step="0.01" />
        </div>

        <!-- Rate (for all new cars, required for shipping only) -->
        <div v-if="!formData.id" class="form-group">
          <label for="rate"
            >{{ t('carStockForm.rate') || 'Rate' }}
            <span v-if="isShippingOnly" class="required">*</span>:</label
          >
          <input type="number" id="rate" v-model="formData.rate" step="0.01" />
        </div>

        <!-- Price Cell (only for existing cars) -->
        <div v-if="formData.id" class="form-group">
          <label for="price_cell">{{ t('carStockForm.priceCell') }}:</label>
          <input
            :disabled="!can_edit_cars_sell_price"
            type="number"
            id="price_cell"
            v-model="formData.price_cell"
            step="0.01"
          />
        </div>
      </div>
    </div>

    <div class="form-actions">
      <button @click="saveCar" :disabled="loading" class="btn save-btn">
        {{ loading ? t('carStockForm.saving') : t('carStockForm.save') }}
      </button>
      <button @click="$emit('cancel')" class="btn cancel-btn">
        {{ t('carStockForm.cancel') }}
      </button>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>
  </div>

  <!-- Add Supplier Dialog -->
  <AddItemDialog
    :show="showAddSupplierDialog"
    :title="t('carStockForm.addSupplier') || 'Add New Supplier'"
    :loading="loading"
    @close="closeAddSupplierDialog"
    @save="addSupplier"
  >
    <div class="form-group">
      <label for="supplier_name"
        >{{ t('carStockForm.supplierName') || 'Name' }} <span class="required">*</span>:</label
      >
      <input
        id="supplier_name"
        v-model="newSupplier.name"
        type="text"
        :placeholder="t('carStockForm.supplierNamePlaceholder') || 'Supplier Name'"
        required
      />
    </div>
    <div class="form-group">
      <label for="supplier_contact_info"
        >{{ t('carStockForm.contactInfo') || 'Contact Information' }}:</label
      >
      <textarea
        id="supplier_contact_info"
        v-model="newSupplier.contact_info"
        :placeholder="t('carStockForm.contactInfoPlaceholder') || 'Contact Information'"
        rows="3"
      ></textarea>
    </div>
    <div class="form-group">
      <label for="supplier_notes">{{ t('carStockForm.notes') || 'Notes' }}:</label>
      <textarea
        id="supplier_notes"
        v-model="newSupplier.notes"
        :placeholder="t('carStockForm.notesPlaceholder') || 'Notes'"
        rows="3"
      ></textarea>
    </div>
  </AddItemDialog>

  <!-- Add Car Name Dialog -->
  <AddItemDialog
    :show="showAddCarNameDialog"
    :title="t('carStockForm.addCarName') || 'Add New Car Name'"
    :loading="loading"
    @close="closeAddCarNameDialog"
    @save="addCarName"
  >
    <div class="form-group">
      <label for="car_name"
        >{{ t('carStockForm.carName') || 'Car Name' }} <span class="required">*</span>:</label
      >
      <input
        id="car_name"
        v-model="newCarName.car_name"
        type="text"
        :placeholder="t('carStockForm.carNamePlaceholder') || 'Car Name'"
        required
      />
    </div>
    <div class="form-group">
      <label for="car_name_brand">{{ t('carStockForm.brand') || 'Brand' }}:</label>
      <SelectWithAddButton
        id="car_name_brand"
        v-model="newCarName.id_brand"
        :options="brands"
        option-value="id"
        option-label="brand"
        :placeholder="t('carStockForm.selectBrand') || 'Select Brand'"
        :add-button-title="t('carStockForm.addBrand') || 'Add New Brand'"
        @add="openAddBrandDialog"
      />
    </div>
    <div class="form-group">
      <label for="car_name_notes">{{ t('carStockForm.notes') || 'Notes' }}:</label>
      <textarea
        id="car_name_notes"
        v-model="newCarName.notes"
        :placeholder="t('carStockForm.notesPlaceholder') || 'Notes'"
        rows="3"
      ></textarea>
    </div>
  </AddItemDialog>

  <!-- Add Brand Dialog -->
  <AddItemDialog
    :show="showAddBrandDialog"
    :title="t('carStockForm.addBrand') || 'Add New Brand'"
    :loading="loading"
    @close="closeAddBrandDialog"
    @save="addBrand"
  >
    <div class="form-group">
      <label for="brand_name"
        >{{ t('carStockForm.brand') || 'Brand' }} <span class="required">*</span>:</label
      >
      <input
        id="brand_name"
        v-model="newBrand.brand"
        type="text"
        :placeholder="t('carStockForm.brandPlaceholder') || 'Brand Name'"
        required
      />
    </div>
  </AddItemDialog>

  <!-- Add Color Dialog -->
  <AddColorDialog
    :show="showAddColorDialog"
    @close="closeAddColorDialog"
    @saved="handleColorSaved"
  />

  <!-- Add Client Dialog -->
  <AddClientDialog
    :show="showAddClientDialog"
    @close="closeAddClientDialog"
    @saved="handleClientSaved"
  />
</template>

<style scoped>
.car-stock-form {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
}

.error {
  color: #ef4444;
  margin-bottom: 16px;
  padding: 8px;
  background-color: #fee2e2;
  border-radius: 4px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}

.form-section {
  background-color: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.form-section h4 {
  margin-top: 0;
  margin-bottom: 16px;
  color: #4b5563;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 8px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
  color: #4b5563;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.form-group textarea {
  min-height: 80px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #4b5563;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
}

.save-btn:hover {
  background-color: #2563eb;
}

.save-btn:disabled {
  background-color: #93c5fd;
  cursor: not-allowed;
}

.file-upload-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-input {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  width: 100%;
}

.current-file-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.9em;
  display: inline-block;
  padding: 4px 0;
}

.current-file-link:hover {
  text-decoration: underline;
}

.error-message {
  color: #ef4444;
  margin-top: 16px;
  font-size: 0.9em;
}

.input-with-button {
  display: flex;
  gap: 8px;
  align-items: center;
}

.input-with-button select {
  flex: 1;
}

.input-with-button :deep(.el-select) {
  flex: 1;
}

.input-with-button :deep(.el-input__wrapper) {
  width: 100%;
}

.btn-add-supplier {
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
}

.btn-add-supplier:hover {
  background-color: #059669;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3000;
}

.dialog {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.dialog-close-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #6b7280;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-close-btn:hover {
  color: #374151;
}

.dialog-body {
  padding: 20px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px;
  border-top: 1px solid #e5e7eb;
}

.checkbox-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.checkbox-wrapper input[type='checkbox'] {
  width: auto;
  margin: 0;
  cursor: pointer;
}

.checkbox-label {
  margin: 0;
  cursor: pointer;
  font-weight: normal;
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: normal;
}

.radio-label input[type='radio'] {
  width: auto;
  margin: 0;
  cursor: pointer;
}

.radio-label span {
  user-select: none;
}
</style>
