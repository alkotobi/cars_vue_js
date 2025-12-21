<script setup>
import { ref, onMounted, watch,   computed } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import { ElSelect, ElOption } from 'element-plus'
import 'element-plus/dist/index.css'

const { t } = useEnhancedI18n()
const props = defineProps({
  mode: {
    type: String,
    default: 'add', // 'add' or 'edit'
    validator: (value) => ['add', 'edit'].includes(value),
  },
  billData: {
    type: Object,
    default: () => ({
      id: null,
      id_broker: null,
      date_sell: new Date().toISOString().split('T')[0],
      notes: '',
      id_user: null,
    }),
  },
})
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)
const can_confirm_payment = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_confirm_payment')
})

const emit = defineEmits(['save', 'cancel'])

const { callApi } = useApi()
const brokers = ref([])
const filteredBrokers = ref([])
const users = ref([])
const filteredUsers = ref([])
const loading = ref(false)
const error = ref(null)

// Add loading state for form submission
const isSubmitting = ref(false)

const formData = ref({
  id: null,
  id_broker: null,
  date_sell: new Date().toISOString().split('T')[0],
  notes: '',
  is_batch_sell: false,
  id_user: null,
  payment_confirmed: false,
})

// Watch for changes in billData prop
watch(
  () => props.billData,
  (newData) => {
    if (newData) {
      formData.value = {
        ...newData,
        // Convert is_batch_sell from database format (0/1) to boolean
        is_batch_sell: Boolean(newData.is_batch_sell),
        // Convert payment_confirmed from database format (0/1) to boolean
        payment_confirmed: Boolean(newData.payment_confirmed),
      }

      // For non-admin users, always set id_user to current user's ID
      if (!isAdmin.value && user.value?.id) {
        formData.value.id_user = user.value.id
      }

      // Ensure date is in the correct format for the date input
      if (formData.value.date_sell) {
        const date = new Date(formData.value.date_sell)
        if (!isNaN(date)) {
          formData.value.date_sell = date.toISOString().split('T')[0]
        }
      }
    }
  },
  { immediate: true },
)

// Watch for changes in isAdmin to ensure id_user is set correctly
watch(
  () => isAdmin.value,
  (newIsAdmin) => {
    // For non-admin users, ensure id_user is set to current user's ID
    if (!newIsAdmin && user.value?.id) {
      formData.value.id_user = user.value.id
    }
  },
)

const fetchBrokers = async () => {
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
      brokers.value = result.data
      filteredBrokers.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch brokers'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, username, email
        FROM users
        ORDER BY username ASC
      `,
      params: [],
    })

    if (result.success) {
      users.value = result.data
      filteredUsers.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch users'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const remoteMethod = (query) => {
  if (query) {
    // Filter the existing brokers data
    filteredBrokers.value = brokers.value.filter((broker) =>
      broker.name.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    // If no query, show all brokers
    filteredBrokers.value = brokers.value
  }
}

const remoteMethodUsers = (query) => {
  if (query) {
    // Filter the existing users data
    filteredUsers.value = users.value.filter((user) =>
      user.username.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    // If no query, show all users
    filteredUsers.value = users.value
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
          INSERT INTO sell_bill (id_broker, date_sell, notes, id_user, is_batch_sell, payment_confirmed, time_created)
          VALUES (?, ?, ?, ?, ?, ?, NOW())
        `,
        params: [
          formData.value.id_broker || null,
          formData.value.date_sell,
          formData.value.notes || '',
          formData.value.id_user || null,
          formData.value.is_batch_sell ? 1 : 0,
          can_confirm_payment.value && formData.value.payment_confirmed ? 1 : 0,
        ],
      })

      console.log('Insert result:', result)
      console.log('Insert result:', result.success && result.lastInsertId)
      if (result.success && result.lastInsertId) {
        // Generate bill_ref
        const username = user.value?.username?.substring(0, 3).toUpperCase() || 'USR'
        const userId = (user.value?.id || 0).toString().padStart(3, '0')
        const date = new Date(formData.value.date_sell)
        const dateStr = date
          .toLocaleDateString('en-GB', {
            day: '2-digit',
            month: '2-digit',
            year: '2-digit',
          })
          .replace(/\//g, '')
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
          params: [billRef, result.lastInsertId],
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
          params: [result.lastInsertId],
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
          SET id_broker = ?, date_sell = ?, notes = ?, is_batch_sell = ?, id_user = ?, payment_confirmed = ?
          WHERE id = ?
        `,
        params: [
          formData.value.id_broker || null,
          formData.value.date_sell,
          formData.value.notes || '',
          formData.value.is_batch_sell ? 1 : 0,
          formData.value.id_user || null,
          can_confirm_payment.value && formData.value.payment_confirmed ? 1 : 0,
          formData.value.id,
        ],
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

    // For non-admin users, ensure id_user is set to current user's ID
    if (!isAdmin.value && user.value?.id) {
      formData.value.id_user = user.value.id
    }

    fetchBrokers()
    fetchUsers()
  }
})
</script>

<template>
  <div class="sell-bill-form">
    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>{{ isSubmitting ? t('sellBills.saving') : t('sellBills.loading') }}</span>
    </div>

    <h3 class="form-title">
      <i class="fas fa-file-invoice-dollar"></i>
      {{ mode === 'add' ? t('sellBills.add_new_sell_bill') : t('sellBills.edit_sell_bill') }}
    </h3>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <form @submit.prevent="saveBill" class="form-content">
      <div class="form-group">
        <label for="broker">
          <i class="fas fa-user-tie"></i>
          {{ t('sellBills.broker') }}
        </label>
        <el-select
          v-model="formData.id_broker"
          filterable
          remote
          :remote-method="remoteMethod"
          :loading="loading"
          :placeholder="t('sellBills.select_broker')"
          class="broker-select"
        >
          <el-option
            v-for="broker in filteredBrokers"
            :key="broker.id"
            :label="broker.name"
            :value="broker.id"
          >
            <i class="fas fa-user-tie"></i>
            {{ broker.name }}
            <small v-if="broker.mobiles">
              <i class="fas fa-phone"></i>
              {{ broker.mobiles }}
            </small>
          </el-option>
        </el-select>
      </div>

      <div v-if="isAdmin" class="form-group">
        <label for="user">
          <i class="fas fa-user"></i>
          {{ t('sellBills.user') }}
        </label>
        <el-select
          v-model="formData.id_user"
          filterable
          remote
          :remote-method="remoteMethodUsers"
          :loading="loading"
          :placeholder="t('sellBills.select_user')"
          class="user-select"
        >
          <el-option
            v-for="user in filteredUsers"
            :key="user.id"
            :label="user.username"
            :value="user.id"
          >
            <i class="fas fa-user"></i>
            {{ user.username }}
            <small v-if="user.email">
              <i class="fas fa-envelope"></i>
              {{ user.email }}
            </small>
          </el-option>
        </el-select>
      </div>

      <div class="form-group">
        <label for="date">
          <i class="fas fa-calendar"></i>
          {{ t('sellBills.date') }}
        </label>
        <input
          type="date"
          id="date"
          v-model="formData.date_sell"
          :disabled="isSubmitting"
          required
        />
      </div>

      <div class="form-group">
        <label for="notes">
          <i class="fas fa-sticky-note"></i>
          {{ t('sellBills.notes') }}
        </label>
        <textarea
          id="notes"
          v-model="formData.notes"
          :disabled="isSubmitting"
          :placeholder="t('sellBills.enter_notes_placeholder')"
        ></textarea>
      </div>

      <div class="form-group">
        <div class="checkbox-wrapper">
          <input
            type="checkbox"
            id="is_batch_sell"
            v-model="formData.is_batch_sell"
            class="form-checkbox"
          />
          <label for="is_batch_sell" class="checkbox-label">{{ t('sellBills.batch_sell') }}</label>
        </div>
      </div>

      <div v-if="can_confirm_payment" class="form-group">
        <div class="checkbox-wrapper">
          <input
            type="checkbox"
            id="payment_confirmed"
            v-model="formData.payment_confirmed"
            class="form-checkbox"
          />
          <label for="payment_confirmed" class="checkbox-label">{{ t('sellBills.payment_confirmed') }}</label>
        </div>
      </div>

      <div class="form-actions">
        <button type="button" @click="$emit('cancel')" :disabled="isSubmitting" class="cancel-btn">
          <i class="fas fa-times"></i>
          {{ t('sellBills.cancel') }}
        </button>
        <button type="submit" :disabled="isSubmitting" class="save-btn">
          <i class="fas fa-save"></i>
          {{ isSubmitting ? t('sellBills.saving') : t('sellBills.save') }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.sell-bill-form {
  position: relative;
  padding: 1.5rem;
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

.broker-select {
  width: 100%;
}

.user-select {
  width: 100%;
}

input[type='date'],
textarea {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

input[type='date']:focus,
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
  background-color: #3b82f6;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
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
.form-group textarea {
  transition: all 0.2s;
}

button i {
  transition: transform 0.2s;
}

button:hover:not(:disabled) i {
  transform: scale(1.1);
}

.checkbox-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 8px 0;
}

.form-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.checkbox-label {
  cursor: pointer;
  user-select: none;
  color: #374151;
  font-size: 0.95rem;
}
</style>
