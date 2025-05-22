<script setup>
import { ref, onMounted, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  onEdit: Function,
  onDelete: Function,
  onSelect: Function
})

const emit = defineEmits(['refresh', 'select-bill'])

const { callApi } = useApi()
const sellBills = ref([])
const loading = ref(true)
const error = ref(null)
const selectedBillId = ref(null)

const fetchSellBills = async () => {
  loading.value = true
  error.value = null
  
  try {
    const result = await callApi({
      query: `
        SELECT 
          sb.id,
          sb.id_broker,
          sb.date_sell,
          sb.notes,
          c.name as broker_name
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

const handleEdit = (bill) => {
  if (props.onEdit) {
    props.onEdit(bill)
  }
}

const handleDelete = (id) => {
  if (props.onDelete) {
    props.onDelete(id)
  }
}

const selectBill = (bill) => {
  selectedBillId.value = bill.id
  
  // Emit the selected bill ID to the parent component
  emit('select-bill', bill.id)
  
  if (props.onSelect) {
    props.onSelect(bill.id)
  }
}

// Expose the fetchSellBills method to parent component
defineExpose({ fetchSellBills })

onMounted(() => {
  fetchSellBills()
})
</script>

<template>
  <div class="sell-bills-table-component">
    <h3>Sell Bills</h3>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="sellBills.length === 0" class="no-data">
      No sell bills found
    </div>
    <table v-else class="sell-bills-table">
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
            <button @click.stop="handleEdit(bill)" class="edit-btn">Edit</button>
            <button @click.stop="handleDelete(bill.id)" class="delete-btn">Delete</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.sell-bills-table-component {
  margin-bottom: 20px;
}

.sell-bills-table {
  width: 100%;
  border-collapse: collapse;
}

.sell-bills-table th,
.sell-bills-table td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.sell-bills-table th {
  background-color: #f2f2f2;
  font-weight: bold;
}

.sell-bills-table tr:hover {
  background-color: #f5f5f5;
  cursor: pointer;
}

.sell-bills-table tr.selected {
  background-color: #e3f2fd;
}

.loading, .error, .no-data {
  padding: 20px;
  text-align: center;
}

.error {
  color: red;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 5px 10px;
  margin-right: 5px;
  cursor: pointer;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 5px 10px;
  cursor: pointer;
}
</style>