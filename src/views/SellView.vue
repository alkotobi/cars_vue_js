<script setup>
import { ref, onMounted, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi } = useApi()
const sellBills = ref([])
const carsStock = ref([])
const availableCars = ref([])
const loading = ref(true)
const stockLoading = ref(false)
const availableCarsLoading = ref(false)
const error = ref(null)
const selectedBillId = ref(null)
const showAddDialog = ref(false)

// Add loading states for form submissions
const isSubmittingBill = ref(false)
const isSubmittingAssignment = ref(false)
const isSubmittingUpdate = ref(false)

// Add new refs for the assignment dialog
const showAssignDialog = ref(false)
const carToAssign = ref(null)
const loadingPorts = ref([])
const dischargePorts = ref([])
const clients = ref([])

// Add new refs for edit dialog
const showEditDialog = ref(false)
const editingSellBill = ref({
  id: null,
  id_broker: null,
  date_sell: '',
  notes: ''
})

// Add new refs for car edit dialog
const showCarEditDialog = ref(false)
const editingCar = ref(null)

const newSellBill = ref({
  id_broker: null,
  date_sell: new Date().toISOString().split('T')[0], // Today's date in YYYY-MM-DD format
  notes: ''
})

// Add ref for brokers list
const brokers = ref([])

const fetchSellBills = async () => {
  loading.value = true
  error.value = null
  
  try {
    const result = await callApi({
      query: `
        SELECT 
          sb.id,
          sb.date_sell,
          sb.notes,
          c.name as broker_name,
          c.mobiles as broker_mobile
        FROM sell_bill sb
        LEFT JOIN clients c ON sb.id_broker = c.id AND c.is_broker = 1
        ORDER BY sb.date_sell DESC
      `,
      params: []
    })
    
    if (result.success) {
      sellBills.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch sell bills'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const fetchBrokers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name, mobiles
        FROM clients
        WHERE is_broker = 1
        ORDER BY name ASC
      `,
      params: []
    })
    
    if (result.success) {
      brokers.value = result.data
    }
  } catch (err) {
    console.error('Error fetching brokers:', err)
  }
}

const selectBill = async (bill) => {
  selectedBillId.value = bill.id
  await fetchCarStock(bill.id)
}

const fetchCarStock = async (sellBillId) => {
  stockLoading.value = true
  
  try {
    const result = await callApi({
      query: `
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
          c.name as client_name,
          cn.car_name,
          clr.color,
          lp.loading_port,
          dp.discharge_port,
          bd.price_sell as buy_price
        FROM cars_stock cs
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        WHERE cs.id_sell = ? AND cs.hidden = 0
        ORDER BY cs.id DESC
      `,
      params: [sellBillId]
    })
    
    if (result.success) {
      carsStock.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch cars stock'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    stockLoading.value = false
  }
}

const fetchAvailableCars = async () => {
  availableCarsLoading.value = true
  
  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.price_cell,
          cs.date_loding,
          cs.notes,
          cs.freight,
          c.name as client_name,
          cn.car_name,
          clr.color,
          lp.loading_port,
          dp.discharge_port,
          bd.price_sell as buy_price
        FROM cars_stock cs
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        WHERE cs.id_sell IS NULL AND cs.hidden = 0
        ORDER BY cs.id DESC
      `,
      params: []
    })
    
    if (result.success) {
      availableCars.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch available cars'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    availableCarsLoading.value = false
  }
}

const addSellBill = async () => {
  // Prevent multiple submissions
  if (isSubmittingBill.value) {
    return
  }
  
  try {
    isSubmittingBill.value = true
    
    const result = await callApi({
      query: `
        INSERT INTO sell_bill (id_broker, date_sell, notes)
        VALUES (?, ?, ?)
      `,
      params: [
        newSellBill.value.id_broker,
        newSellBill.value.date_sell,
        newSellBill.value.notes
      ]
    })
    
    if (result.success) {
      showAddDialog.value = false
      // Reset form
      newSellBill.value = {
        id_broker: null,
        date_sell: new Date().toISOString().split('T')[0],
        notes: ''
      }
      // Refresh the list
      await fetchSellBills()
    } else {
      error.value = result.error || 'Failed to add sell bill'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isSubmittingBill.value = false
  }
}

// Add function to fetch loading ports
const fetchLoadingPorts = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM loading_ports ORDER BY loading_port ASC',
      params: []
    })
    
    if (result.success) {
      loadingPorts.value = result.data
    }
  } catch (err) {
    console.error('Error fetching loading ports:', err)
  }
}

// Add function to fetch discharge ports
const fetchDischargePorts = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM discharge_ports ORDER BY discharge_port ASC',
      params: []
    })
    
    if (result.success) {
      dischargePorts.value = result.data
    }
  } catch (err) {
    console.error('Error fetching discharge ports:', err)
  }
}

// Add function to fetch clients
const fetchClients = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM clients WHERE is_broker = 0 ORDER BY name ASC',
      params: []
    })
    
    if (result.success) {
      clients.value = result.data
    }
  } catch (err) {
    console.error('Error fetching clients:', err)
  }
}

// Modify the assignCarToBill function to show dialog first
const openAssignDialog = (car) => {
  carToAssign.value = { ...car }
  showAssignDialog.value = true
}

const assignCarToBill = async () => {
  // Prevent multiple submissions
  if (isSubmittingAssignment.value) {
    return
  }
  
  if (!selectedBillId.value) {
    alert('Please select a bill first')
    return
  }
  
  // Validate required fields
  if (!carToAssign.value.id_client) {
    alert('Client name is required')
    return
  }
  
  if (!carToAssign.value.id_port_discharge) {
    alert('Discharge port is required')
    return
  }
  
  if (!carToAssign.value.price_cell) {
    alert('Selling price is required')
    return
  }
  
  try {
    isSubmittingAssignment.value = true
    
    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = ?,
            id_client = ?,
            vin = ?,
            id_port_loading = ?,
            id_port_discharge = ?,
            price_cell = ?
        WHERE id = ?
      `,
      params: [
        selectedBillId.value, 
        carToAssign.value.id_client,
        carToAssign.value.vin,
        carToAssign.value.id_port_loading,
        carToAssign.value.id_port_discharge,
        carToAssign.value.price_cell,
        carToAssign.value.id
      ]
    })
    
    if (result.success) {
      showAssignDialog.value = false
      // Refresh both car lists
      await fetchCarStock(selectedBillId.value)
      await fetchAvailableCars()
    } else {
      error.value = result.error || 'Failed to assign car to bill'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isSubmittingAssignment.value = false
  }
}

// Add this new function to unassign a car from a bill
const unassignCar = async (carId) => {
  try {
    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_client = NULL,
            id_port_discharge = NULL,
            id_sell = NULL
        WHERE id = ?
      `,
      params: [carId]
    })
    
    if (result.success) {
      // Refresh both car lists
      await fetchCarStock(selectedBillId.value)
      await fetchAvailableCars()
    } else {
      error.value = result.error || 'Failed to unassign car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Add new function to open edit dialog
const openEditDialog = (bill, event) => {
  // Stop the click event from propagating to the row
  event.stopPropagation()
  
  // Set the editing bill data
  editingSellBill.value = {
    id: bill.id,
    id_broker: bill.id_broker,
    date_sell: bill.date_sell.split('T')[0], // Format date for input
    notes: bill.notes || ''
  }
  
  showEditDialog.value = true
}

// Add new function to update sell bill
const updateSellBill = async () => {
  // Prevent multiple submissions
  if (isSubmittingUpdate.value) {
    return
  }
  
  try {
    isSubmittingUpdate.value = true
    
    const result = await callApi({
      query: `
        UPDATE sell_bill 
        SET id_broker = ?, 
            date_sell = ?, 
            notes = ?
        WHERE id = ?
      `,
      params: [
        editingSellBill.value.id_broker,
        editingSellBill.value.date_sell,
        editingSellBill.value.notes,
        editingSellBill.value.id
      ]
    })
    
    if (result.success) {
      showEditDialog.value = false
      // Refresh the list
      await fetchSellBills()
      
      // If the edited bill was selected, refresh its data
      if (selectedBillId.value === editingSellBill.value.id) {
        await fetchCarStock(selectedBillId.value)
      }
    } else {
      error.value = result.error || 'Failed to update sell bill'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isSubmittingUpdate.value = false
  }
}

// Add new function to delete sell bill
const deleteSellBill = async (billId, event) => {
  // Stop the click event from propagating to the row
  event.stopPropagation()
  
  // Confirm deletion
  if (!confirm('Are you sure you want to delete this sell bill? This will also unassign all cars from this bill.')) {
    return
  }
  
  try {
    // First, unassign all cars from this bill
    const unassignResult = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = NULL 
        WHERE id_sell = ?
      `,
      params: [billId]
    })
    
    if (!unassignResult.success) {
      error.value = unassignResult.error || 'Failed to unassign cars from bill'
      return
    }
    
    // Then delete the bill
    const deleteResult = await callApi({
      query: `
        DELETE FROM sell_bill 
        WHERE id = ?
      `,
      params: [billId]
    })
    
    if (deleteResult.success) {
      // If the deleted bill was selected, clear selection
      if (selectedBillId.value === billId) {
        selectedBillId.value = null
        carsStock.value = []
      }
      
      // Refresh the lists
      await fetchSellBills()
      await fetchAvailableCars()
    } else {
      error.value = deleteResult.error || 'Failed to delete sell bill'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

onMounted(async () => {
  await fetchSellBills()
  await fetchBrokers()
  await fetchAvailableCars()
  await fetchLoadingPorts()
  await fetchDischargePorts()
  await fetchClients()
})

// Add function to edit car
const editCar = (car) => {
  editingCar.value = { ...car }
  showCarEditDialog.value = true
}

// Add function to update car
const updateCar = async () => {
  if (!editingCar.value) return
  
  // Prevent multiple submissions
  if (isSubmittingUpdate.value) {
    return
  }
  
  try {
    isSubmittingUpdate.value = true
    
    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_client = ?,
            vin = ?,
            id_port_loading = ?,
            id_port_discharge = ?,
            price_cell = ?,
            notes = ?
        WHERE id = ?
      `,
      params: [
        editingCar.value.id_client,
        editingCar.value.vin,
        editingCar.value.id_port_loading,
        editingCar.value.id_port_discharge,
        editingCar.value.price_cell,
        editingCar.value.notes,
        editingCar.value.id
      ]
    })
    
    if (result.success) {
      showCarEditDialog.value = false
      await fetchCarStock(selectedBillId.value)
    } else {
      error.value = result.error || 'Failed to update car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    isSubmittingUpdate.value = false
  }
}
</script>

<template>
  <div class="sell-view">
    <div class="header">
      <h2>Sell Bills Management</h2>
      <button class="add-btn" @click="showAddDialog = true">
        Add New Sell Bill
      </button>
    </div>
    
    <div v-if="selectedBillId" class="selected-bill">
      Selected Bill ID: <span class="bill-id">{{ selectedBillId }}</span>
    </div>
    
    <div class="content">
      <div v-if="loading" class="loading">Loading...</div>
      
      <div v-else-if="error" class="error">{{ error }}</div>
      
      <div v-else class="tables-container">
        <!-- Sell Bills Table -->
        <div class="sell-bills-table">
          <h3>Sell Bills</h3>
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Date</th>
                <th>Broker</th>
                <th>Notes</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr 
                v-for="bill in sellBills" 
                :key="bill.id" 
                @click="selectBill(bill)"
                :class="{ 'selected': selectedBillId === bill.id }"
              >
                <td>{{ bill.id }}</td>
                <td>{{ new Date(bill.date_sell).toLocaleDateString() }}</td>
                <td>{{ bill.broker_name || 'N/A' }}</td>
                <td>{{ bill.notes || 'N/A' }}</td>
                <td>
                  <button @click="openEditDialog(bill, $event)" class="btn edit-btn">Edit</button>
                  <button class="btn delete-btn" @click="deleteSellBill(bill.id, $event)">Delete</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        
        <!-- Car Stock Table for Selected Bill -->
        <div class="car-stock-table" v-if="selectedBillId">
          <h3>Cars in Selected Bill</h3>
          <div v-if="stockLoading" class="loading">Loading cars...</div>
          <div v-else-if="carsStock.length === 0" class="no-data">
            No cars found for this sell bill
          </div>
          <table v-else>
            <thead>
              <tr>
                <th>ID</th>
                <th>Car</th>
                <th>Color</th>
                <th>VIN</th>
                <th>Client</th>
                <th>Loading Port</th>
                <th>Discharge Port</th>
                <th>Buy Price</th>
                <th>Sell Price</th>
                <th>Profit</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="car in carsStock" :key="car.id">
                <td>{{ car.id }}</td>
                <td>{{ car.car_name || 'N/A' }}</td>
                <td>{{ car.color || 'N/A' }}</td>
                <td>{{ car.vin || 'N/A' }}</td>
                <td>{{ car.client_name || 'N/A' }}</td>
                <td>{{ car.loading_port || 'N/A' }}</td>
                <td>{{ car.discharge_port || 'N/A' }}</td>
                <td>${{ car.buy_price || 0 }}</td>
                <td>${{ car.price_cell || 0 }}</td>
                <td>${{ (car.price_cell - car.buy_price) || 0 }}</td>
                <td>
                  <button @click="editCar(car)" class="edit-btn">Edit</button>
                  <button @click="unassignCar(car.id)" class="unassign-btn">Unassign</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        
        <!-- Available Cars Table -->
        <div class="available-cars-table" v-if="selectedBillId">
          <h3>Available Cars</h3>
          <div v-if="availableCarsLoading" class="loading">Loading...</div>
          <div v-else-if="availableCars.length === 0" class="no-data">No available cars</div>
          <table v-else>
            <thead>
              <tr>
                <th>Car</th>
                <th>Color</th>
                <th>VIN</th>
                <th>Buy Price</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="car in availableCars" :key="car.id">
                <td>{{ car.car_name || 'N/A' }}</td>
                <td>{{ car.color || 'N/A' }}</td>
                <td>{{ car.vin || 'N/A' }}</td>
                <td>${{ car.buy_price || 0 }}</td>
                <td>${{ car.price_cell || 0 }}</td>
                <td>${{ (car.price_cell - car.buy_price) || 0 }}</td>
                <td>
                  <button @click="openAssignDialog(car)" class="assign-btn">Assign to Bill</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
    
    <!-- Add Assignment Dialog -->
    <div class="dialog-overlay" v-if="showAssignDialog">
      <div class="dialog">
        <h2>Assign Car to Bill</h2>
        <div class="form-group">
          <label for="client">Client: <span class="required">*</span></label>
          <select id="client" v-model="carToAssign.id_client" required>
            <option value="">Select Client</option>
            <option v-for="client in clients" :key="client.id" :value="client.id">{{ client.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label for="vin">VIN:</label>
          <input type="text" id="vin" v-model="carToAssign.vin">
        </div>
        <div class="form-group">
          <label for="loading-port">Loading Port:</label>
          <select id="loading-port" v-model="carToAssign.id_port_loading">
            <option value="">Select Loading Port</option>
            <option v-for="port in loadingPorts" :key="port.id" :value="port.id">{{ port.loading_port }}</option>
          </select>
        </div>
        <div class="form-group">
          <label for="discharge-port">Discharge Port: <span class="required">*</span></label>
          <select id="discharge-port" v-model="carToAssign.id_port_discharge" required>
            <option value="">Select Discharge Port</option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.id">{{ port.discharge_port }}</option>
          </select>
        </div>
        <div class="form-group">
          <label for="sell-price">Sell Price: <span class="required">*</span></label>
          <input type="number" id="sell-price" v-model="carToAssign.price_cell" required>
        </div>
        <div class="dialog-buttons">
          <button @click="assignCarToBill" class="primary">Assign</button>
          <button @click="showAssignDialog = false">Cancel</button>
        </div>
      </div>
    </div>
    
    <!-- Add Sell Bill Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Sell Bill</h3>
        
        <div class="form-group">
          <label for="broker">Broker:</label>
          <select id="broker" v-model="newSellBill.id_broker">
            <option value="">Select Broker</option>
            <option v-for="broker in brokers" :key="broker.id" :value="broker.id">
              {{ broker.name }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="date">Date:</label>
          <input type="date" id="date" v-model="newSellBill.date_sell">
        </div>
        
        <div class="form-group">
          <label for="notes">Notes:</label>
          <textarea id="notes" v-model="newSellBill.notes"></textarea>
        </div>
        
        <div class="dialog-buttons">
          <button @click="showAddDialog = false">Cancel</button>
          <button @click="addSellBill" class="primary">Add</button>
        </div>
      </div>
    </div>
    <!-- Car Edit Dialog -->
    <div v-if="showCarEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Car Details</h3>
        <div class="form-group">
          <label for="edit-car-client">Client:</label>
          <select id="edit-car-client" v-model="editingCar.id_client">
            <option value="">Select Client</option>
            <option v-for="client in clients" :key="client.id" :value="client.id">
              {{ client.name }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label for="edit-car-vin">VIN:</label>
          <input type="text" id="edit-car-vin" v-model="editingCar.vin">
        </div>
        <div class="form-group">
          <label for="edit-car-loading-port">Loading Port:</label>
          <select id="edit-car-loading-port" v-model="editingCar.id_port_loading">
            <option value="">Select Loading Port</option>
            <option v-for="port in loadingPorts" :key="port.id" :value="port.id">
              {{ port.loading_port }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label for="edit-car-discharge-port">Discharge Port:</label>
          <select id="edit-car-discharge-port" v-model="editingCar.id_port_discharge">
            <option value="">Select Discharge Port</option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
              {{ port.discharge_port }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label for="edit-car-price">Selling Price:</label>
          <input type="number" id="edit-car-price" v-model="editingCar.price_cell">
        </div>
        <div class="form-group">
          <label for="edit-car-notes">Notes:</label>
          <textarea id="edit-car-notes" v-model="editingCar.notes"></textarea>
        </div>
        <div class="dialog-buttons">
          <button @click="updateCar" class="primary">Update</button>
          <button @click="showCarEditDialog = false">Cancel</button>
        </div>
      </div>
    </div>
    
    <!-- Assign Car Dialog -->
    <div class="dialog-overlay" v-if="showAssignDialog">
      <div class="dialog">
        <h2>Assign Car to Bill</h2>
        <div class="form-group">
          <label for="client">Client: <span class="required">*</span></label>
          <select id="client" v-model="carToAssign.id_client" required>
            <option value="">Select Client</option>
            <option v-for="client in clients" :key="client.id" :value="client.id">{{ client.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label for="vin">VIN:</label>
          <input type="text" id="vin" v-model="carToAssign.vin">
        </div>
        <div class="form-group">
          <label for="loading-port">Loading Port:</label>
          <select id="loading-port" v-model="carToAssign.id_port_loading">
            <option value="">Select Loading Port</option>
            <option v-for="port in loadingPorts" :key="port.id" :value="port.id">{{ port.loading_port }}</option>
          </select>
        </div>
        <div class="form-group">
          <label for="discharge-port">Discharge Port: <span class="required">*</span></label>
          <select id="discharge-port" v-model="carToAssign.id_port_discharge" required>
            <option value="">Select Discharge Port</option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.id">{{ port.discharge_port }}</option>
          </select>
        </div>
        <div class="form-group">
          <label for="sell-price">Sell Price: <span class="required">*</span></label>
          <input type="number" id="sell-price" v-model="carToAssign.price_cell" required>
        </div>
        <div class="dialog-buttons">
          <button @click="assignCarToBill" class="primary">Assign</button>
          <button @click="showAssignDialog = false">Cancel</button>
        </div>
      </div>
    </div>
    
    <!-- Add Sell Bill Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Sell Bill</h3>
        
        <div class="form-group">
          <label for="broker">Broker:</label>
          <select id="broker" v-model="newSellBill.id_broker">
            <option value="">Select Broker</option>
            <option v-for="broker in brokers" :key="broker.id" :value="broker.id">
              {{ broker.name }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="date">Date:</label>
          <input type="date" id="date" v-model="newSellBill.date_sell">
        </div>
        
        <div class="form-group">
          <label for="notes">Notes:</label>
          <textarea id="notes" v-model="newSellBill.notes"></textarea>
        </div>
        
        <div class="dialog-buttons">
          <button @click="showAddDialog = false">Cancel</button>
          <button @click="addSellBill" class="primary">Add</button>
        </div>
      </div>
    </div>
    <!-- Replace this edit dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Sell Bill</h3>
        
        <div class="form-group">
          <label for="edit-broker">Broker:</label>
          <select id="edit-broker" v-model="editingSellBill.id_broker">
            <option value="">Select Broker</option>
            <option v-for="broker in brokers" :key="broker.id" :value="broker.id">
              {{ broker.name }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="edit-date">Date:</label>
          <input type="date" id="edit-date" v-model="editingSellBill.date_sell">
        </div>
        
        <div class="form-group">
          <label for="edit-notes">Notes:</label>
          <textarea id="edit-notes" v-model="editingSellBill.notes"></textarea>
        </div>
        
        <div class="dialog-buttons">
          <button @click="showEditDialog = false">Cancel</button>
          <button @click="updateSellBill" class="primary">Update</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sell-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.add-btn {
  background-color: #4CAF50;
  color: white;
  border: none;
  padding: 10px 15px;
  border-radius: 4px;
  cursor: pointer;
}

.selected-bill {
  margin-bottom: 15px;
  padding: 10px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.bill-id {
  font-weight: bold;
  color: #007bff;
}

.tables-container {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}

th, td {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}

th {
  background-color: #f2f2f2;
}

tr:hover {
  background-color: #f5f5f5;
}

tr.selected {
  background-color: #e0f7fa;
}

.view-btn {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
}

.no-selection, .no-data {
  padding: 20px;
  text-align: center;
  color: #6c757d;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.loading, .error {
  padding: 20px;
  text-align: center;
}

.error {
  color: #dc3545;
  margin-left: 3px;
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
  padding: 20px;
  border-radius: 8px;
  width: 500px;
  max-width: 90%;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.dialog-buttons button {
  padding: 8px 15px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.dialog-buttons button.primary {
  background-color: #4CAF50;
  color: white;
}
.edit-btn {
  background-color: #3b82f6;
  color: white;
}
.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
}
.delete-btn {
  background-color: #ef4444;
  color: white;
}

.available-cars-table {
  margin-top: 30px;
}

.assign-btn {
  background-color: #ff9800;
  color: white;
  border: none;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
}

.assign-btn:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.assign-btn:hover {
  background-color: #f57c00;
}

.unassign-btn {
  background-color: #680707;
  color: white;
  border: none;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
}

.unassign-btn:hover {
  background-color: #f50303;
}


</style>