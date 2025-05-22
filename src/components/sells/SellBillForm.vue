<script setup>
import { ref, onMounted, watch, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  mode: {
    type: String,
    default: 'add', // 'add' or 'edit'
    validator: (value) => ['add', 'edit'].includes(value)
  },
  billData: {
    type: Object,
    default: () => ({
      id: null,
      id_broker: null,
      date_sell: new Date().toISOString().split('T')[0],
      notes: ''
    })
  }
})

const emit = defineEmits(['save', 'cancel'])

const { callApi } = useApi()
const brokers = ref([])
const loading = ref(false)
const error = ref(null)

const formData = ref({
  id: null,
  id_broker: null,
  date_sell: new Date().toISOString().split('T')[0],
  notes: ''
})

// Watch for changes in billData prop
watch(() => props.billData, (newData) => {
  if (newData) {
    formData.value = { ...newData }
    // Ensure date is in the correct format for the date input
    if (formData.value.date_sell) {
      const date = new Date(formData.value.date_sell)
      if (!isNaN(date)) {
        formData.value.date_sell = date.toISOString().split('T')[0]
      }
    }
  }
}, { immediate: true })

const fetchBrokers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name
        FROM clients
        WHERE is_broker = 1
        ORDER BY name ASC
      `,
      params: []
    })
    
    if (result.success) {
      brokers.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch brokers'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const saveBill = async () => {
  loading.value = true
  error.value = null
  
  try {
    let result
    
    if (props.mode === 'add') {
      result = await callApi({
        query: `
          INSERT INTO sell_bill (id_broker, date_sell, notes)
          VALUES (?, ?, ?)
        `,
        params: [
          formData.value.id_broker || null,
          formData.value.date_sell,
          formData.value.notes || ''
        ]
      })
    } else {
      result = await callApi({
        query: `
          UPDATE sell_bill
          SET id_broker = ?, date_sell = ?, notes = ?
          WHERE id = ?
        `,
        params: [
          formData.value.id_broker || null,
          formData.value.date_sell,
          formData.value.notes || '',
          formData.value.id
        ]
      })
    }
    
    if (result.success) {
      emit('save', result.data)
    } else {
      error.value = result.error || `Failed to ${props.mode} sell bill`
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchBrokers()
})
</script>

<template>
  <div class="sell-bill-form">
    <h3>{{ mode === 'add' ? 'Add New' : 'Edit' }} Sell Bill</h3>
    
    <div v-if="error" class="error">{{ error }}</div>
    
    <form @submit.prevent="saveBill">
      <div class="form-group">
        <label for="broker">Broker:</label>
        <select id="broker" v-model="formData.id_broker">
          <option value="">Select Broker</option>
          <option v-for="broker in brokers" :key="broker.id" :value="broker.id">
            {{ broker.name }}
          </option>
        </select>
      </div>
      
      <div class="form-group">
        <label for="date">Date:</label>
        <input type="date" id="date" v-model="formData.date_sell" required>
      </div>
      
      <div class="form-group">
        <label for="notes">Notes:</label>
        <textarea id="notes" v-model="formData.notes"></textarea>
      </div>
      
      <div class="form-actions">
        <button type="button" @click="$emit('cancel')" class="cancel-btn">Cancel</button>
        <button type="submit" class="save-btn" :disabled="loading">
          {{ loading ? 'Saving...' : (mode === 'add' ? 'Add' : 'Update') }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.sell-bill-form {
  max-width: 600px;
  margin: 0 auto;
}

.form-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

input, select, textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

textarea {
  height: 100px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

button {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.save-btn:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.error {
  color: red;
  margin-bottom: 15px;
  padding: 10px;
  background-color: #ffebee;
  border-radius: 4px;
}
</style>