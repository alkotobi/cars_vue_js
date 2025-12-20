<script setup>
import { ref, computed, onMounted } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useApi } from '../../composables/useApi'

const { t } = useEnhancedI18n()
const { callApi } = useApi()

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  selectedCarIds: {
    type: Array,
    default: () => [],
  },
  selectionId: {
    type: Number,
    default: null,
  },
})

const emit = defineEmits(['close', 'sent'])

const users = ref([])
const teams = ref([])
const selectedUsers = ref([])
const selectedTeam = ref(null)
const deadline = ref('')
const priority = ref('medium')
const loading = ref(false)
const isSending = ref(false)
const error = ref(null)

const isValid = computed(() => {
  return (
    props.selectedCarIds.length > 0 &&
    (selectedUsers.value.length > 0 || selectedTeam.value !== null)
  )
})

onMounted(async () => {
  await loadUsers()
  await loadTeams()
})

const loadUsers = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, username, email FROM users WHERE id > 0 ORDER BY username',
    })
    if (result.success) {
      users.value = result.data
    }
  } catch (err) {
    console.error('Error loading users:', err)
  }
}

const loadTeams = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT t.id, t.name, t.team_leader_id, u.username as leader_name
        FROM teams t
        LEFT JOIN users u ON t.team_leader_id = u.id
        WHERE t.is_active = 1
        ORDER BY t.name
      `,
    })
    if (result.success) {
      teams.value = result.data
    }
  } catch (err) {
    console.error('Error loading teams:', err)
  }
}

const toggleUser = (userId) => {
  const index = selectedUsers.value.indexOf(userId)
  if (index > -1) {
    selectedUsers.value.splice(index, 1)
  } else {
    selectedUsers.value.push(userId)
  }
  // Clear team selection if selecting users
  if (selectedUsers.value.length > 0) {
    selectedTeam.value = null
  }
}

const handleTeamChange = (teamId) => {
  selectedTeam.value = teamId
  // Clear user selection if selecting team
  if (teamId !== null) {
    selectedUsers.value = []
  }
}

const handleSend = async () => {
  if (!isValid.value) return

  isSending.value = true
  error.value = null

  try {
    const userStr = localStorage.getItem('user')
    const user = userStr ? JSON.parse(userStr) : null

    if (!user || !user.id) {
      throw new Error('User not found')
    }

    let selectionId = props.selectionId

    // If updating existing selection, get current owned_by
    let currentOwnedBy = []
    if (selectionId) {
      // Update existing selection - get current owned_by
      const getResult = await callApi({
        query: 'SELECT owned_by FROM car_selections WHERE id = ?',
        params: [selectionId],
      })
      if (getResult.success && getResult.data && getResult.data.length > 0) {
        const ownedByData = getResult.data[0].owned_by
        currentOwnedBy = ownedByData ? JSON.parse(ownedByData) : []
      }
    } else if (selectedUsers.value.length > 0) {
      // When sending to users (not team) and no selectionId provided, 
      // try to find existing selection with same car IDs from this user
      const findResult = await callApi({
        query: `
          SELECT id, owned_by 
          FROM car_selections 
          WHERE user_create_selection = ? 
            AND JSON_CONTAINS(selection_data, ?)
            AND assigned_to_team IS NULL
          ORDER BY created_at DESC
          LIMIT 1
        `,
        params: [user.id, JSON.stringify(props.selectedCarIds)],
      })
      
      if (findResult.success && findResult.data && findResult.data.length > 0) {
        // Found existing selection, use it
        selectionId = findResult.data[0].id
        const ownedByData = findResult.data[0].owned_by
        currentOwnedBy = ownedByData ? JSON.parse(ownedByData) : []
      } else {
        // No existing selection found, create new one
        const createResult = await callApi({
          query: `
            INSERT INTO car_selections (
              name,
              description,
              user_create_selection,
              selection_data,
              owned_by,
              sent_by_user_id,
              status,
              priority,
              deadline,
              status_changed_at
            ) VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?, NOW())
          `,
          params: [
            `${t('carStock.selection_sent')} ${new Date().toLocaleDateString()}`,
            null,
            user.id,
            JSON.stringify(props.selectedCarIds),
            JSON.stringify([user.id]), // Creator is added to owned_by
            user.id,
            priority.value,
            deadline.value.trim() || null,
          ],
        })

        if (!createResult.success) {
          throw new Error(createResult.error || 'Failed to create selection')
        }

        selectionId = createResult.insertId
        currentOwnedBy = [user.id] // Start with creator
      }
    } else {
      // Sending to team or no users selected - create new selection
      const createResult = await callApi({
        query: `
          INSERT INTO car_selections (
            name,
            description,
            user_create_selection,
            selection_data,
            assigned_to_team,
            assigned_to_team_from_user_id,
            assigned_at,
            sent_by_user_id,
            status,
            priority,
            deadline,
            status_changed_at
          ) VALUES (?, ?, ?, ?, ?, ?, NOW(), ?, 'pending', ?, ?, NOW())
        `,
        params: [
          `${t('carStock.selection_sent')} ${new Date().toLocaleDateString()}`,
          null,
          user.id,
          JSON.stringify(props.selectedCarIds),
          selectedTeam.value,
          selectedTeam.value ? user.id : null,
          user.id,
          priority.value,
          deadline.value.trim() || null,
        ],
      })

      if (!createResult.success) {
        throw new Error(createResult.error || 'Failed to create selection')
      }

      selectionId = createResult.insertId
    }

    // Build list of users to add to owned_by (start with existing users)
    const usersToAdd = [...currentOwnedBy]
    
    // Ensure creator is included
    if (!usersToAdd.includes(user.id)) {
      usersToAdd.push(user.id)
    }
    
    if (selectedUsers.value.length > 0) {
      // Add selected users (avoid duplicates)
      for (const userId of selectedUsers.value) {
        if (!usersToAdd.includes(userId)) {
          usersToAdd.push(userId)
        }
      }
    }

    // If sent to team, get all team members
    if (selectedTeam.value) {
      const teamMembersResult = await callApi({
        query: 'SELECT user_id FROM team_members WHERE team_id = ?',
        params: [selectedTeam.value],
      })
      
      if (teamMembersResult.success && teamMembersResult.data) {
        for (const member of teamMembersResult.data) {
          if (!usersToAdd.includes(member.user_id)) {
            usersToAdd.push(member.user_id)
          }
        }
      }
    }

    // Update owned_by with all users who can see this selection
    await callApi({
      query: 'UPDATE car_selections SET owned_by = ? WHERE id = ?',
      params: [JSON.stringify(usersToAdd), selectionId],
    })

    // Create ownership history records for newly added users
    for (const userId of selectedUsers.value) {
      await callApi({
        query: `
          INSERT INTO selection_ownership_history (
            selection_id,
            user_id,
            action,
            from_user_id,
            to_user_id,
            created_at
          ) VALUES (?, ?, 'sent', ?, ?, NOW())
        `,
        params: [selectionId, userId, user.id, userId],
      })
    }

    emit('sent', {
      selectionId,
      users: selectedUsers.value,
      team: selectedTeam.value,
    })

    // Reset form
    selectedUsers.value = []
    selectedTeam.value = null
    deadline.value = ''
    priority.value = 'medium'
    emit('close')
  } catch (err) {
    error.value = err.message || t('carStock.failed_to_send_selection')
  } finally {
    isSending.value = false
  }
}

const handleClose = () => {
  selectedUsers.value = []
  selectedTeam.value = null
  deadline.value = ''
  priority.value = 'medium'
  error.value = null
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="handleClose">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>{{ t('carStock.send_selection') }}</h3>
        <button @click="handleClose" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div class="selection-info">
          <i class="fas fa-car"></i>
          <span>
            {{ selectedCarIds.length }}
            {{ selectedCarIds.length === 1 ? t('carStockToolbar.car') : t('carStockToolbar.cars') }}
            {{ t('carStock.selected') }}
          </span>
        </div>

        <div class="form-section">
          <h4>{{ t('carStock.send_to') }}</h4>

          <div class="form-group">
            <label>{{ t('carStock.users') }}</label>
            <div class="users-list">
              <label
                v-for="user in users"
                :key="user.id"
                class="user-checkbox"
                :class="{ active: selectedUsers.includes(user.id) }"
              >
                <input
                  type="checkbox"
                  :checked="selectedUsers.includes(user.id)"
                  @change="toggleUser(user.id)"
                  :disabled="selectedTeam !== null || isSending"
                />
                <span>{{ user.username }}</span>
                <span class="user-email">{{ user.email }}</span>
              </label>
            </div>
          </div>

          <div class="divider">
            <span>{{ t('carStock.or') }}</span>
          </div>

          <div class="form-group">
            <label>{{ t('carStock.team') }}</label>
            <select
              v-model="selectedTeam"
              class="form-select"
              :disabled="selectedUsers.length > 0 || isSending"
              @change="handleTeamChange(selectedTeam)"
            >
              <option :value="null">{{ t('carStock.select_team') }}</option>
              <option v-for="team in teams" :key="team.id" :value="team.id">
                {{ team.name }} ({{ t('carStock.leader') }}: {{ team.leader_name || '-' }})
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>{{ t('carStock.deadline') || 'Deadline' }}</label>
            <input
              v-model="deadline"
              type="datetime-local"
              class="form-select"
              :disabled="isSending"
            />
          </div>

          <div class="form-group">
            <label>{{ t('carStock.priority') || 'Priority' }}</label>
            <select
              v-model="priority"
              class="form-select"
              :disabled="isSending"
            >
              <option value="low">{{ t('carStock.priority_low') || 'Low' }}</option>
              <option value="medium">{{ t('carStock.priority_medium') || 'Medium' }}</option>
              <option value="high">{{ t('carStock.priority_high') || 'High' }}</option>
              <option value="urgent">{{ t('carStock.priority_urgent') || 'Urgent' }}</option>
            </select>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button @click="handleClose" class="btn btn-cancel" :disabled="isSending">
          {{ t('carStock.cancel') }}
        </button>
        <button
          @click="handleSend"
          class="btn btn-send"
          :disabled="!isValid || isSending"
        >
          <i v-if="isSending" class="fas fa-spinner fa-spin"></i>
          <i v-else class="fas fa-paper-plane"></i>
          {{ t('carStock.send') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.modal-body {
  padding: 24px;
  flex: 1;
}

.error-message {
  background-color: #fef2f2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.selection-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background-color: #f0f9ff;
  border-radius: 6px;
  color: #0c4a6e;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 20px;
}

.selection-info i {
  color: #3b82f6;
}

.form-section h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #374151;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.users-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 8px;
}

.user-checkbox {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.user-checkbox:hover {
  background-color: #f9fafb;
}

.user-checkbox.active {
  background-color: #eff6ff;
}

.user-checkbox input[type='checkbox'] {
  cursor: pointer;
}

.user-checkbox span:first-of-type {
  font-weight: 500;
  color: #1f2937;
}

.user-email {
  margin-left: auto;
  font-size: 12px;
  color: #6b7280;
}

.divider {
  display: flex;
  align-items: center;
  margin: 20px 0;
  text-align: center;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  border-bottom: 1px solid #e5e7eb;
}

.divider span {
  padding: 0 12px;
  color: #6b7280;
  font-size: 14px;
  font-weight: 500;
}

.form-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background-color: white;
  cursor: pointer;
  transition: all 0.2s ease;
}

.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-select:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
}

.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.btn-send {
  background-color: #10b981;
  color: white;
}

.btn-send:hover:not(:disabled) {
  background-color: #059669;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>

