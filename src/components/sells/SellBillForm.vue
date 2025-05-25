<script setup>
import { ref, onMounted, watch, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'
import { ElSelect, ElOption } from 'element-plus'
import 'element-plus/dist/index.css'

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
const user = ref(null);

const emit = defineEmits(['save', 'cancel'])

const { callApi } = useApi()
const brokers = ref([])
const filteredBrokers = ref([])
const loading = ref(false)
const error = ref(null)

// Add loading state for form submission
const isSubmitting = ref(false)

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
      filteredBrokers.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch brokers'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const remoteMethod = (query) => {
  if (query) {
    // Filter the existing brokers data
    filteredBrokers.value = brokers.value.filter(broker => 
      broker.name.toLowerCase().includes(query.toLowerCase())
    )
  } else {
    // If no query, show all brokers
    filteredBrokers.value = brokers.value
  }
}

const handleBrokerChange = () => {
  // Handle broker change if needed
}

const saveBill = async () => {
  // Prevent multiple submissions
  if (isSubmitting.value) {
    return
  }
  
  try {
    isSubmitting.value = true
    loading.value = true
    error.value = null
    
    let result
    
    if (props.mode === 'add') {
      console.log('User data:', user.value)
      
      // First insert the bill to get the ID
      result = await callApi({
        query: `
          INSERT INTO sell_bill (id_broker, date_sell, notes, id_user)
          VALUES (?, ?, ?, ?)
        `,
        params: [
          formData.value.id_broker || null,
          formData.value.date_sell,
          formData.value.notes || '',
          user.value?.id || null
        ]
      })
      
      console.log('Insert result:', result)
      console.log('Insert result:', result.success && result.lastInsertId)
      if (result.success && result.lastInsertId) {
        // Generate bill_ref
        const username = user.value?.username?.substring(0, 3).toUpperCase() || 'USR'
        const userId = (user.value?.id || 0).toString().padStart(3, '0')
        const date = new Date(formData.value.date_sell)
        const dateStr = date.toLocaleDateString('en-GB', {
          day: '2-digit',
          month: '2-digit',
          year: '2-digit'
        }).replace(/\//g, '')
        const billId = result.lastInsertId.toString().padStart(3, '0')
        const billRef = `${username}${userId}${dateStr}${billId}`
        
        console.log('Generated bill_ref:', billRef)
        
        // Update the bill with the generated bill_ref
        const updateResult = await callApi({
          query: `
            UPDATE sell_bill 
            SET bill_ref = ?
            WHERE id = ?
          `,
          params: [billRef, result.lastInsertId]
        })
        
        console.log('Update result:', updateResult)
        
        if (!updateResult.success) {
          console.error('Failed to update bill_ref:', updateResult.error)
          error.value = 'Failed to update bill reference'
          return
        }

        // Get the complete bill data after update
        const fetchResult = await callApi({
          query: `
            SELECT 
              sb.id,
              sb.id_broker,
              sb.date_sell,
              sb.notes,
              sb.bill_ref,
              c.name as broker_name
            FROM sell_bill sb
            LEFT JOIN clients c ON sb.id_broker = c.id AND c.is_broker = 1
            WHERE sb.id = ?
          `,
          params: [result.lastInsertId]
        })
        
        console.log('Fetch result:', fetchResult)

        if (fetchResult.success && fetchResult.data.length > 0) {
          emit('save', fetchResult.data[0])
        } else {
          error.value = 'Failed to fetch updated bill data'
        }
      } else {
        console.log('Insert failed:', result.error)
        error.value = result.error || 'Failed to create sell bill'
      }
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
    
    if (result.success) {
      emit('save', result.data)
    } else {
      error.value = result.error || `Failed to ${props.mode} sell bill`
      }
    }
  } catch (err) {
    console.error('Error in saveBill:', err)
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    isSubmitting.value = false
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  console.log('User from localStorage:', userStr) // Debug localStorage
  if (userStr) {
    user.value = JSON.parse(userStr)
    console.log('Parsed user:', user.value) // Debug parsed user
    fetchBrokers()
  }
})
</script>

<template>
  <div class="sell-bill-form">
    <h3>{{ mode === 'add' ? 'Add New' : 'Edit' }} Sell Bill</h3>
    
    <div v-if="error" class="error">{{ error }}</div>
    
    <form @submit.prevent="saveBill">
      <div class="form-group">
        <label for="broker">Broker:</label>
        <el-select
          v-model="formData.id_broker"
          filterable
          :loading="loading"
          placeholder="Select or search broker"
          style="width: 100%"
          @change="handleBrokerChange"
        >
          <el-option
            v-for="broker in filteredBrokers"
            :key="broker.id"
            :label="broker.name"
            :value="broker.id"
          />
        </el-select>
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
        <button type="submit" class="save-btn" :disabled="isSubmitting">
          <span v-if="isSubmitting" class="spinner"></span>
          {{ isSubmitting ? 'Saving...' : (mode === 'add' ? 'Add' : 'Update') }}
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
  opacity: 0.6;
}

.error {
  color: red;
  margin-bottom: 15px;
  padding: 10px;
  background-color: #ffebee;
  border-radius: 4px;
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
  to { transform: rotate(360deg); }
}

:deep(.el-select) {
  width: 100%;
}

:deep(.el-input__wrapper) {
  background-color: white;
}

:deep(.el-select .el-input.is-focus .el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset;
}
</style>