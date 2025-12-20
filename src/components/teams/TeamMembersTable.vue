<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '../../composables/useI18n'

const props = defineProps({
  team: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['membersUpdated'])

const { callApi } = useApi()
const { t } = useEnhancedI18n()

const teamMembers = ref([])
const isLoading = ref(false)
const availableUsers = ref([])
const selectedUserToAdd = ref(null)
const showAddForm = ref(false)

// Watch for team changes and load members
watch(() => props.team, async (newTeam) => {
  if (newTeam && newTeam.id) {
    await loadTeamMembers(newTeam.id)
    await loadAvailableUsers(newTeam.id)
  } else {
    teamMembers.value = []
    availableUsers.value = []
    selectedUserToAdd.value = null
    showAddForm.value = false
  }
}, { immediate: true })

// Load team members (leader first, then others by join date)
const loadTeamMembers = async (teamId) => {
  isLoading.value = true
  try {
    // First get the team leader
    const teamResult = await callApi({
      query: 'SELECT team_leader_id FROM teams WHERE id = ?',
      params: [teamId]
    })
    
    const leaderId = teamResult.success && teamResult.data && teamResult.data.length > 0 
      ? teamResult.data[0].team_leader_id 
      : null
    
    // Get all team members
    const result = await callApi({
      query: `
        SELECT tm.*, u.username, u.email,
               CASE WHEN tm.user_id = ? THEN 0 ELSE 1 END as is_leader_sort
        FROM team_members tm
        JOIN users u ON tm.user_id = u.id
        WHERE tm.team_id = ?
        ORDER BY is_leader_sort ASC, tm.joined_at ASC
      `,
      params: [leaderId, teamId]
    })
    if (result.success) {
      teamMembers.value = result.data
    }
  } catch (err) {
    console.error('Error loading team members:', err)
    alert((t('cars.errorLoadingMembers') || 'Error loading team members') + ': ' + err.message)
  } finally {
    isLoading.value = false
  }
}

// Load available users (users not in this team and not leading another team)
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
          AND u.id NOT IN (
            SELECT t.team_leader_id
            FROM teams t
            WHERE t.id != ?
          )
        ORDER BY u.username
      `,
      params: [teamId, teamId || 0]
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
  if (!props.team || !props.team.id) {
    alert(t('cars.noTeamSelected') || 'No team selected')
    return
  }
  try {
    const result = await callApi({
      query: 'INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)',
      params: [props.team.id, selectedUserToAdd.value, 'member']
    })
    if (result.success) {
      selectedUserToAdd.value = null
      showAddForm.value = false
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
  if (!props.team || !props.team.id) {
    alert(t('cars.noTeamSelected') || 'No team selected')
    return
  }
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
</script>

<template>
  <div v-if="team" class="team-members-table-container">
    <div class="table-header">
      <h3>{{ t('cars.teamMembers') || 'Team Members' }} - {{ team.name }}</h3>
      <button class="add-member-btn" @click="showAddForm = !showAddForm">
        <i class="fas fa-plus"></i>
        {{ t('cars.addMember') || 'Add Member' }}
      </button>
    </div>

    <!-- Add Member Form -->
    <div v-if="showAddForm" class="add-member-form-card">
      <div class="form-content">
        <select v-model="selectedUserToAdd" class="input-field">
          <option :value="null">{{ t('cars.selectUser') || 'Select a user' }}</option>
          <option v-for="user in availableUsers" :key="user.id" :value="user.id">
            {{ user.username }} {{ user.email ? '(' + user.email + ')' : '' }}
          </option>
        </select>
        <div class="form-actions">
          <button class="save-button" @click="addUserToTeam" :disabled="!selectedUserToAdd">
            {{ t('cars.add') || 'Add' }}
          </button>
          <button class="cancel-button" @click="showAddForm = false; selectedUserToAdd = null">
            {{ t('cars.cancel') || 'Cancel' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Members Table -->
    <div class="table-container">
      <div v-if="isLoading" class="loading-state">
        <i class="fas fa-spinner fa-spin"></i>
        {{ t('cars.loading') || 'Loading...' }}
      </div>
      <div v-else-if="teamMembers.length === 0" class="no-members">
        {{ t('cars.noMembers') || 'No members in this team' }}
      </div>
      <table v-else class="members-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>{{ t('cars.username') || 'Username' }}</th>
            <th>{{ t('cars.email') || 'Email' }}</th>
            <th>{{ t('cars.role') || 'Role' }}</th>
            <th>{{ t('cars.joinedAt') || 'Joined At' }}</th>
            <th>{{ t('cars.actions') || 'Actions' }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="member in teamMembers" :key="member.id">
            <td>{{ member.id }}</td>
            <td>{{ member.username }}</td>
            <td>{{ member.email || '-' }}</td>
            <td>
              <span v-if="team && member.user_id === team.team_leader_id" class="role-badge role-leader">
                {{ t('cars.teamLeader') || 'Team Leader' }}
              </span>
              <span v-else class="role-badge" :class="member.role === 'deputy_leader' ? 'role-deputy' : 'role-member'">
                {{ member.role === 'deputy_leader' ? (t('cars.deputyLeader') || 'Deputy Leader') : (t('cars.member') || 'Member') }}
              </span>
            </td>
            <td>{{ new Date(member.joined_at).toLocaleDateString() }}</td>
            <td>
              <button 
                v-if="!team || member.user_id !== team.team_leader_id"
                class="remove-button" 
                @click="removeUserFromTeam(member.id, member.username)" 
                :title="t('cars.removeMember') || 'Remove Member'"
              >
                <i class="fas fa-times"></i>
              </button>
              <span v-else class="cannot-remove">{{ t('cars.leaderCannotBeRemoved') || 'Leader' }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
  <div v-else class="no-team-selected">
    <i class="fas fa-hand-pointer"></i>
    <p>{{ t('cars.selectTeamToViewMembers') || 'Select a team from above to view its members' }}</p>
  </div>
</template>

<style scoped>
.team-members-table-container {
  margin-top: 2rem;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.table-header h3 {
  margin: 0;
  color: #111827;
}

.add-member-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background-color: #22c55e;
  color: white;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.add-member-btn:hover {
  background-color: #16a34a;
}

.add-member-form-card {
  background-color: white;
  border-radius: 0.375rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 1rem;
  padding: 1rem;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.input-field {
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
}

.save-button {
  background-color: #2563eb;
  color: white;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.save-button:hover:not(:disabled) {
  background-color: #1d4ed8;
}

.save-button:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.cancel-button {
  background-color: #f3f4f6;
  color: #374151;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.cancel-button:hover {
  background-color: #e5e7eb;
}

.table-container {
  overflow-x: auto;
}

.loading-state,
.no-members {
  padding: 2rem;
  text-align: center;
  color: #6b7280;
  background-color: white;
  border-radius: 0.375rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.loading-state i {
  margin-right: 0.5rem;
}

.members-table {
  width: 100%;
  border-collapse: collapse;
  background-color: white;
  border-radius: 0.375rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.members-table th,
.members-table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.members-table th {
  background-color: #f9fafb;
  font-weight: 500;
  color: #374151;
}

.members-table tbody tr:hover {
  background-color: #f3f4f6;
}

.role-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
}

.role-member {
  background-color: #dbeafe;
  color: #1e40af;
}

.role-deputy {
  background-color: #fef3c7;
  color: #92400e;
}

.role-leader {
  background-color: #dbeafe;
  color: #1e3a8a;
  font-weight: 600;
}

.cannot-remove {
  color: #9ca3af;
  font-size: 0.875rem;
  font-style: italic;
}

.remove-button {
  background: none;
  border: none;
  cursor: pointer;
  color: #dc2626;
  padding: 0.5rem;
  border-radius: 0.25rem;
  transition: background-color 0.3s ease;
}

.remove-button:hover {
  background-color: #fef2f2;
}

.no-team-selected {
  margin-top: 2rem;
  padding: 3rem;
  text-align: center;
  background-color: #f9fafb;
  border-radius: 0.375rem;
  border: 2px dashed #d1d5db;
  color: #6b7280;
}

.no-team-selected i {
  font-size: 3rem;
  margin-bottom: 1rem;
  color: #9ca3af;
}

.no-team-selected p {
  margin: 0;
  font-size: 1.1rem;
}
</style>

