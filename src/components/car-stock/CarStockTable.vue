<script setup>
import { ref, onMounted, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'


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
      advanced: null
    })
  }
})

const emit = defineEmits(['refresh'])

const { callApi } = useApi()
const cars = ref([])
const loading = ref(true)
const error = ref(null)

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
        c.name as client_name,
        cn.car_name,
        clr.color,
        lp.loading_port,
        dp.discharge_port,
        bd.price_sell as buy_price,
        w.warhouse_name as warehouse_name
      FROM cars_stock cs
      LEFT JOIN clients c ON cs.id_client = c.id
      LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
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

const toggleDropdown = (carId) => {
  isDropdownOpen.value[carId] = !isDropdownOpen.value[carId];
};

const handleVINAction = (car) => {
  alert(`VIN action triggered for car with ID: ${car.id}`);
  // Close the dropdown after action
  isDropdownOpen.value[car.id] = false;
};

onMounted(fetchCarsStock)

// Expose the fetchCarsStock method to parent components
defineExpose({
  fetchCarsStock
})


</script>

<template>
  <div class="cars-stock-table">
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="cars.length === 0" class="empty-state">No cars in stock</div>
    
    <table v-else class="cars-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Car</th>
          <th>Color</th>
          <th>VIN</th>
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
          <td>{{ car.loading_port || 'N/A' }}</td>
          <td>{{ car.discharge_port || 'N/A' }}</td>
          <td>{{ car.freight ? '$' + car.freight : 'N/A' }}</td>
          <td>{{ car.price_cell ? '$' + car.price_cell : 'N/A' }}</td>
          <td>{{ car.date_loding ? new Date(car.date_loding).toLocaleDateString() : 'N/A' }}</td>
          <td :class="car.date_sell || car.client_name ? 'status-sold' : 'status-available'">
            {{ car.date_sell || car.client_name ? 'Sold' : 'Available' }}
          </td>
          <td>{{ car.client_name || 'N/A' }}</td>
          <td>
            <div class="document-links">
              <a v-if="car.path_documents" :href="car.path_documents" target="_blank">Documents</a>
              <a v-if="car.sell_pi_path" :href="car.sell_pi_path" target="_blank">Sell PI</a>
              <a v-if="car.buy_pi_path" :href="car.buy_pi_path" target="_blank">Buy PI</a>
            </div>
          </td>
          <td>
            <button @click="handleEdit(car)" class="edit-btn">Edit</button>
            <div class="dropdown">
            <button @click="toggleDropdown(car.id)" class="dropdown-toggle">Actions</button>
            <ul v-if="isDropdownOpen[car.id]" class="dropdown-menu">
              <li><button @click="handleVINAction(car)">VIN</button></li>
              <!-- You can add more actions here -->
            </ul>
          </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  
</template>

<style scoped>
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

.cars-table {
  width: 100%;
  border-collapse: collapse;
}

.cars-table th,
.cars-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.cars-table th {
  background-color: #f8f9fa;
  font-weight: 600;
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

.document-links {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.document-links a {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.9em;
}

.document-links a:hover {
  text-decoration: underline;
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
</style>