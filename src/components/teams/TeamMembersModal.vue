<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '../../composables/useI18n'

const props = defineProps({
  team: {
    type: Object,
    default: null
  },
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'membersUpdated'])

const { callApi } = useApi()
const { t } = useEnhancedI18n()

const teamMembers = ref([])
const availableUsers = ref([])
const selectedUserToAdd = ref(null)
const isLoadingMembers = ref(false)

// Watch for team changes and load members
watch(() => props.team, async (newTeam) => {
  if (newTeam && props.show) {
    await loadTeamMembers(newTeam.id)
    await loadAvailableUsers(newTeam.id)
  }
}, { immediate: true })

watch(() => props.show, async (isVisible) => {
  if (isVisible && props.team) {
    await loadTeamMembers(props.team.id)
    await loadAvailableUsers(props.team.id)
  } else {
    // Reset when closing
    teamMembers.value = []
    availableUsers.value = []
    selectedUserToAdd.value = null
  }
})

// Load team members
const loadTeamMembers = async (teamId) => {
  isLoadingMembers.value = true
  try {
    const result = await callApi({
      query: `
        SELECT tm.*, u.username, u.email
        FROM team_members tm
        JOIN users u ON tm.user_id = u.id
        WHERE tm.team_id = ?
        ORDER BY tm.joined_at DESC
      `,
      params: [teamId]
    })
    if (result.success) {
      teamMembers.value = result.data
    }
  } catch (err) {
    console.error('Error loading team members:', err)
    alert((t('cars.errorLoadingMembers') || 'Error loading team members') + ': ' + err.message)
  } finally {
    isLoadingMembers.value = false
  }
}

// Load available users (users not in any team or not in this team)
const loadAvailableUsers = async (teamId) => {
  try {
    const result = await callApi({
      query: `
        SELECT u.id, u.username, u.email
        FROM users u
        WHERE u.id > 0
          AND u.id NOT IN (
            SELECT tm.user_id 
            FROM team_members tm 
            WHERE tm.team_id = ?
          )
        ORDER BY u.username
      `,
      params: [teamId]
    })
    if (result.success) {
      availableUsers.value = result.data
    }
  } catch (err) {
    console.error('Error loading available users:', err)
  }
}

// Add user to team
const addUserToTeam = async () => {
  if (!selectedUserToAdd.value) {
    alert(t('cars.pleaseSelectUser') || 'Please select a user to add')
    return
  }
  try {
    const result = await callApi({
      query: 'INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)',
      params: [props.team.id, selectedUserToAdd.value, 'member']
    })
    if (result.success) {
      selectedUserToAdd.value = null
      await loadTeamMembers(props.team.id)
      await loadAvailableUsers(props.team.id)
      emit('membersUpdated') // Notify parent to refresh teams list
    } else {
      alert((t('cars.errorAddingMember') || 'Error adding member') + ': ' + result.error)
    }
  } catch (err) {
    console.error('Error adding user to team:', err)
    if (err.message && err.message.includes('Duplicate entry')) {
      alert(t('cars.userAlreadyInTeam') || 'This user is already in a team')
    } else {
      alert((t('cars.errorAddingMember') || 'Error adding member') + ': ' + err.message)
    }
  }
}

// Remove user from team
const removeUserFromTeam = async (memberId, username) => {
  if (!confirm((t('cars.confirmRemoveMember') || 'Are you sure you want to remove') + ' ' + username + ' ' + (t('cars.fromTeam') || 'from this team') + '?')) {
    return
  }
  try {
    const result = await callApi({
      query: 'DELETE FROM team_members WHERE id = ?',
      params: [memberId]
    })
    if (result.success) {
      await loadTeamMembers(props.team.id)
      await loadAvailableUsers(props.team.id)
      emit('membersUpdated') // Notify parent to refresh teams list
    } else {
      alert((t('cars.errorRemovingMember') || 'Error removing member') + ': ' + result.error)
    }
  } catch (err) {
    console.error('Error removing user from team:', err)
    alert((t('cars.errorRemovingMember') || 'Error removing member') + ': ' + err.message)
  }
}

const closeModal = () => {
  emit('close')
}
</script>

<template>
  <div v-if="show && team" class="modal-overlay" @click.self="closeModal">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ t('cars.manageMembers') || 'Manage Members' }} - {{ team.name }}</h3>
        <button class="close-button" @click="closeModal">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <!-- Add Member Form -->
        <div class="add-member-section">
          <h4>{{ t('cars.addMember') || 'Add Member' }}</h4>
          <div class="add-member-form">
            <select v-model="selectedUserToAdd" class="input-field">
              <option :value="null">{{ t('cars.selectUser') || 'Select a user' }}</option>
              <option v-for="user in availableUsers" :key="user.id" :value="user.id">
                {{ user.username }} {{ user.email ? '(' + user.email + ')' : '' }}
              </option>
            </select>
            <button class="add-member-button" @click="addUserToTeam" :disabled="!selectedUserToAdd">
              <i class="fas fa-plus"></i>
              {{ t('cars.add') || 'Add' }}
            </button>
          </div>
        </div>

        <!-- Current Members List -->
        <div class="members-list-section">
          <h4>{{ t('cars.currentMembers') || 'Current Members' }} ({{ teamMembers.length }})</h4>
          <div v-if="isLoadingMembers" class="loading-members">
            {{ t('cars.loading') || 'Loading...' }}
          </div>
          <div v-else-if="teamMembers.length === 0" class="no-members">
            {{ t('cars.noMembers') || 'No members in this team' }}
          </div>
          <div v-else class="members-list">
            <div v-for="member in teamMembers" :key="member.id" class="member-item">
              <div class="member-info">
                <span class="member-name">{{ member.username }}</span>
                <span class="member-email" v-if="member.email">{{ member.email }}</span>
                <span class="member-role">{{ member.role === 'deputy_leader' ? (t('cars.deputyLeader') || 'Deputy Leader') : (t('cars.member') || 'Member') }}</span>
              </div>
              <button class="remove-member-button" @click="removeUserFromTeam(member.id, member.username)" :title="t('cars.removeMember') || 'Remove Member'">
                <i class="fas fa-times"></i>
              </button>
            </div>
          </div>
        </div>
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
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background-color: white;
  border-radius: 0.5rem;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  color: #111827;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.close-button {
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
  font-size: 1.25rem;
}

.add-member-section {
  margin-bottom: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.add-member-section h4 {
  margin: 0 0 1rem 0;
  color: #374151;
}

.add-member-form {
  display: flex;
  gap: 0.5rem;
}

.input-field {
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  flex: 1;
}

.add-member-button {
  background-color: #22c55e;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.3s ease;
}

.add-member-button:hover:not(:disabled) {
  background-color: #16a34a;
}

.add-member-button:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.members-list-section h4 {
  margin: 0 0 1rem 0;
  color: #374151;
}

.loading-members,
.no-members {
  padding: 1rem;
  text-align: center;
  color: #6b7280;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.member-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background-color: #f9fafb;
  border-radius: 0.375rem;
  border: 1px solid #e5e7eb;
}

.member-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.member-name {
  font-weight: 500;
  color: #111827;
}

.member-email {
  font-size: 0.875rem;
  color: #6b7280;
}

.member-role {
  font-size: 0.75rem;
  color: #9ca3af;
  text-transform: capitalize;
}

.remove-member-button {
  background: none;
  border: none;
  cursor: pointer;
  color: #dc2626;
  padding: 0.5rem;
  border-radius: 0.25rem;
  transition: background-color 0.3s ease;
}

.remove-member-button:hover {
  background-color: #fef2f2;
}
</style>

