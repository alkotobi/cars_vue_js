<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'
import { useEnhancedI18n } from '../composables/useI18n'
import AddEditTeamForm from '../components/teams/AddEditTeamForm.vue'
import TeamMembersTable from '../components/teams/TeamMembersTable.vue'

const { callApi } = useApi()
const { t } = useEnhancedI18n()
const teams = ref([])
const users = ref([])
const editingTeam = ref(null)
const showAddForm = ref(false)
const showEditForm = ref(false)
const isLoading = ref(true)
const error = ref(null)

// Team members management
const selectedTeam = ref(null)

// Fetch all users for team leader selection (excluding users who already lead a team)
const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT u.id, u.username 
        FROM users u
        WHERE u.id > 0
        ORDER BY u.username
      `,
      params: []
    })
    if (result.success) {
      users.value = result.data
    }
  } catch (err) {
    console.error('Error fetching users:', err)
  }
}

// Fetch all teams
const fetchTeams = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT t.*, u.username as leader_name,
               (SELECT COUNT(*) FROM team_members tm WHERE tm.team_id = t.id) as member_count
        FROM teams t
        LEFT JOIN users u ON t.team_leader_id = u.id
        ORDER BY t.created_at DESC
      `,
      params: []
    })
    if (result.success) {
      teams.value = result.data
    } else {
      error.value = (t('cars.failedToFetchTeams') || 'Failed to fetch teams') + ': ' + result.error
    }
  } catch (err) {
    error.value = (t('cars.errorFetchingTeams') || 'Error fetching teams') + ': ' + err.message
  } finally {
    isLoading.value = false
  }
}

// Edit a team
const startEdit = (team) => {
  editingTeam.value = { ...team }
  showEditForm.value = true
}

// Handle team saved (from AddEditTeamForm)
const handleTeamSaved = async () => {
  editingTeam.value = null
  showAddForm.value = false
  showEditForm.value = false
  await fetchTeams()
  
  // If a team was selected, refresh the members list to show the newly added leader
  if (selectedTeam.value) {
    // Trigger members update by emitting event
    await handleMembersUpdated()
  }
}

// Delete a team
const deleteTeam = async (id) => {
  if (!confirm(t('cars.confirmDeleteTeam') || 'Are you sure you want to delete this team?')) {
    return
  }
  try {
    const result = await callApi({
      query: 'DELETE FROM teams WHERE id = ?',
      params: [id]
    })
    if (result.success) {
      await fetchTeams()
    }
  } catch (err) {
    console.error('Error deleting team:', err)
    alert((t('cars.errorDeletingTeam') || 'Error deleting team') + ': ' + err.message)
  }
}

// Select team to view members
const selectTeam = (team) => {
  selectedTeam.value = team
}

// Handle members updated (from TeamMembersTable)
const handleMembersUpdated = async () => {
  await fetchTeams() // Refresh teams to update member count
  // Keep the selected team updated
  if (selectedTeam.value) {
    const updatedTeam = teams.value.find(t => t.id === selectedTeam.value.id)
    if (updatedTeam) {
      // Create a new object reference to trigger reactivity
      selectedTeam.value = { ...updatedTeam }
    }
  }
}

onMounted(async () => {
  await fetchUsers()
  await fetchTeams()
})
</script>

<template>
  <div class="teams-view">
    <div class="header">
      <h2>{{ t('cars.teams') || 'Teams' }}</h2>
      <button class="add-button" @click="showAddForm = true">
        <i class="fas fa-plus"></i>
        {{ t('cars.addTeam') || 'Add Team' }}
      </button>
    </div>

    <!-- Add/Edit Team Form -->
    <AddEditTeamForm
      :team="editingTeam"
      :users="users"
      :show="showAddForm || showEditForm"
      @close="showAddForm = false; showEditForm = false; editingTeam = null"
      @saved="handleTeamSaved"
    />

    <div class="table-container">
      <div v-if="isLoading">{{ t('cars.loading') || 'Loading teams...' }}</div>
      <div v-else-if="error" class="error-message">{{ error }}</div>
      <table v-else class="modern-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>{{ t('cars.teamName') || 'Team Name' }}</th>
            <th>{{ t('cars.teamLeader') || 'Team Leader' }}</th>
            <th>{{ t('cars.description') || 'Description' }}</th>
            <th>{{ t('cars.status') || 'Status' }}</th>
            <th>{{ t('cars.members') || 'Members' }}</th>
            <th>{{ t('cars.jobsCompleted') || 'Jobs Completed' }}</th>
            <th>{{ t('cars.actions') || 'Actions' }}</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="team in teams" 
            :key="team.id"
            :class="{ 'selected-row': selectedTeam && selectedTeam.id === team.id }"
            @click="selectTeam(team)"
            class="clickable-row"
          >
            <td>{{ team.id }}</td>
            <td>{{ team.name }}</td>
            <td>{{ team.leader_name || '-' }}</td>
            <td>{{ team.description || '-' }}</td>
            <td>
              <span :class="team.is_active ? 'status-active' : 'status-inactive'">
                {{ team.is_active ? (t('cars.active') || 'Active') : (t('cars.inactive') || 'Inactive') }}
              </span>
            </td>
            <td>{{ team.member_count || 0 }}</td>
            <td>{{ team.jobs_completed_count }}</td>
            <td @click.stop>
              <button class="edit-button" @click.stop="startEdit(team)" :title="t('cars.edit') || 'Edit'">
                <i class="fas fa-edit"></i>
              </button>
              <button class="delete-button" @click.stop="deleteTeam(team.id)" :title="t('cars.delete') || 'Delete'">
                <i class="fas fa-trash"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Team Members Table -->
    <TeamMembersTable
      :team="selectedTeam"
      @membersUpdated="handleMembersUpdated"
    />
  </div>
</template>

<style scoped>
.teams-view {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.add-button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background-color: #22c55e;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.add-button:hover {
  background-color: #16a34a;
}


.table-container {
  overflow-x: auto;
}

.modern-table {
  width: 100%;
  border-collapse: collapse;
  background-color: white;
  border-radius: 0.375rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.modern-table th,
.modern-table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.modern-table th {
  background-color: #f9fafb;
  font-weight: 500;
}

.modern-table tbody tr:hover {
  background-color: #f3f4f6;
}

.clickable-row {
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.clickable-row:hover {
  background-color: #e0e7ff;
}

.selected-row {
  background-color: #dbeafe !important;
  border-left: 4px solid #2563eb;
}

.selected-row:hover {
  background-color: #bfdbfe !important;
}

.edit-button,
.delete-button,
.members-button {
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
  padding: 0.5rem;
  margin: 0 0.25rem;
}

.members-button:hover {
  color: #8b5cf6;
}

.edit-button:hover {
  color: #2563eb;
}

.delete-button:hover {
  color: #dc2626;
}

.status-active {
  color: #10b981;
  font-weight: 500;
}

.status-inactive {
  color: #ef4444;
  font-weight: 500;
}

.error-message {
  color: #dc2626;
  padding: 1rem;
  background-color: #fef2f2;
  border-radius: 0.375rem;
  margin-bottom: 1rem;
}
</style>

