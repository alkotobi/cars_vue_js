<script setup>
import { ref, watch, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '../../composables/useI18n'

const props = defineProps({
  team: {
    type: Object,
    default: null
  },
  users: {
    type: Array,
    default: () => []
  },
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'saved'])

const { callApi } = useApi()
const { t } = useEnhancedI18n()

const formData = ref({
  name: '',
  team_leader_id: null,
  description: '',
  is_active: 1
})

const isEditMode = computed(() => !!props.team)

// Define resetForm before watchers that use it
const resetForm = () => {
  formData.value = {
    name: '',
    team_leader_id: null,
    description: '',
    is_active: 1
  }
}

// Watch for team changes and populate form
watch(() => props.team, (newTeam) => {
  if (newTeam && newTeam.id) {
    formData.value = {
      name: newTeam.name || '',
      team_leader_id: newTeam.team_leader_id || null,
      description: newTeam.description || '',
      is_active: newTeam.is_active !== undefined ? newTeam.is_active : 1
    }
  } else {
    resetForm()
  }
}, { immediate: true })

watch(() => props.show, (isVisible) => {
  if (!isVisible) {
    resetForm()
  }
})

const saveTeam = async () => {
  if (!formData.value.name || !formData.value.team_leader_id) {
    alert(t('cars.teamNameAndLeaderRequired') || 'Team name and leader are required')
    return
  }

  if (isEditMode.value && (!props.team || !props.team.id)) {
    alert(t('cars.noTeamSelected') || 'No team selected for editing')
    return
  }

  try {
    // Check if leader already leads another team
    const currentTeamId = isEditMode.value && props.team ? props.team.id : null
    const checkLeaderResult = await callApi({
      query: 'SELECT id FROM teams WHERE team_leader_id = ? AND id != ?',
      params: [formData.value.team_leader_id, currentTeamId || 0]
    })
    
    if (checkLeaderResult.success && checkLeaderResult.data && checkLeaderResult.data.length > 0) {
      alert(t('cars.leaderAlreadyLeadsTeam') || 'This user already leads another team. A leader can only lead one team.')
      return
    }

    let result
    let teamId
    
    if (isEditMode.value && props.team && props.team.id) {
      // Update existing team
      const oldLeaderId = props.team.team_leader_id
      const newLeaderId = formData.value.team_leader_id
      
      result = await callApi({
        query: 'UPDATE teams SET name = ?, team_leader_id = ?, description = ?, is_active = ? WHERE id = ?',
        params: [
          formData.value.name,
          formData.value.team_leader_id,
          formData.value.description || null,
          formData.value.is_active,
          props.team.id
        ]
      })
      
      if (result.success) {
        teamId = props.team.id
        
        // If leader changed, update team members
        if (oldLeaderId !== newLeaderId) {
          // Remove old leader from team members if they were a member
          await callApi({
            query: 'DELETE FROM team_members WHERE team_id = ? AND user_id = ?',
            params: [teamId, oldLeaderId]
          })
          
          // Check if new leader is already a member
          const checkMemberResult = await callApi({
            query: 'SELECT id FROM team_members WHERE team_id = ? AND user_id = ?',
            params: [teamId, newLeaderId]
          })
          
          // Add new leader as first member if not already a member
          if (!checkMemberResult.success || !checkMemberResult.data || checkMemberResult.data.length === 0) {
            // First, remove them from any other team (they can only be in one team)
            await callApi({
              query: 'DELETE FROM team_members WHERE user_id = ?',
              params: [newLeaderId]
            })
            
            // Add as first member
            await callApi({
              query: 'INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)',
              params: [teamId, newLeaderId, 'member']
            })
          }
        }
      }
    } else {
      // Create new team
      result = await callApi({
        query: 'INSERT INTO teams (name, team_leader_id, description, is_active) VALUES (?, ?, ?, ?)',
        params: [
          formData.value.name,
          formData.value.team_leader_id,
          formData.value.description || null,
          formData.value.is_active
        ]
      })
      
      if (result.success) {
        teamId = result.insertId
        
        // Add leader as first member automatically
        // First, remove them from any other team (they can only be in one team)
        const removeFromOtherTeamResult = await callApi({
          query: 'DELETE FROM team_members WHERE user_id = ?',
          params: [formData.value.team_leader_id]
        })
        
        if (!removeFromOtherTeamResult.success) {
          console.warn('Warning: Could not remove leader from other team:', removeFromOtherTeamResult.error)
        }
        
        // Add leader as first member
        const addLeaderResult = await callApi({
          query: 'INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)',
          params: [teamId, formData.value.team_leader_id, 'member']
        })
        
        if (!addLeaderResult.success) {
          console.error('Error adding leader to team members:', addLeaderResult.error)
          // Don't fail the whole operation, but log the error
          alert(t('cars.warningLeaderNotAddedToMembers') || 'Team created but leader was not added to members list. Please add manually.')
        }
      }
    }

    if (result.success) {
      emit('saved')
      emit('close')
    } else {
      alert((isEditMode.value ? (t('cars.errorUpdatingTeam') || 'Error updating team') : (t('cars.errorCreatingTeam') || 'Error creating team')) + ': ' + result.error)
    }
  } catch (err) {
    console.error('Error saving team:', err)
    alert((isEditMode.value ? (t('cars.errorUpdatingTeam') || 'Error updating team') : (t('cars.errorCreatingTeam') || 'Error creating team')) + ': ' + err.message)
  }
}

const closeForm = () => {
  emit('close')
}
</script>

<template>
  <div v-if="show" class="form-card">
    <div class="form-header">
      <h3>{{ isEditMode ? (t('cars.editTeam') || 'Edit Team') : (t('cars.addNewTeam') || 'Add New Team') }}</h3>
      <button class="close-button" @click="closeForm">
        <i class="fas fa-times"></i>
      </button>
    </div>
    <div class="form-content">
      <div class="input-wrapper">
        <label>{{ t('cars.teamName') || 'Team Name' }} *</label>
        <input v-model="formData.name" :placeholder="t('cars.enterTeamName') || 'Enter team name'" class="input-field" />
      </div>
      <div class="input-wrapper">
        <label>{{ t('cars.teamLeader') || 'Team Leader' }} *</label>
        <select v-model.number="formData.team_leader_id" class="input-field">
          <option :value="null">{{ t('cars.selectTeamLeader') || 'Select Team Leader' }}</option>
          <option v-for="user in users" :key="user.id" :value="user.id">
            {{ user.username }}
          </option>
        </select>
      </div>
      <div class="input-wrapper">
        <label>{{ t('cars.description') || 'Description' }}</label>
        <textarea v-model="formData.description" :placeholder="t('cars.enterDescription') || 'Enter description'" class="input-field"></textarea>
      </div>
      <div class="input-wrapper">
        <label>
          <input type="checkbox" v-model.number="formData.is_active" :true-value="1" :false-value="0" />
          {{ t('cars.active') || 'Active' }}
        </label>
      </div>
    </div>
    <div class="form-actions">
      <button class="save-button" @click="saveTeam">{{ t('cars.save') || 'Save' }}</button>
      <button class="cancel-button" @click="closeForm">{{ t('cars.cancel') || 'Cancel' }}</button>
    </div>
  </div>
</template>

<style scoped>
.form-card {
  background-color: white;
  border-radius: 0.375rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  margin-bottom: 2rem;
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.form-header h3 {
  margin: 0;
  color: #111827;
}

.close-button {
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
  font-size: 1.25rem;
}

.form-content {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.input-wrapper label {
  font-weight: 500;
  color: #374151;
}

.input-field {
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #e5e7eb;
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

.save-button:hover {
  background-color: #1d4ed8;
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
</style>

