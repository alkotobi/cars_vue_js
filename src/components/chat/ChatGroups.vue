<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi, loading, error } = useApi()
const chatGroups = ref([])
const selectedGroup = ref(null)

// Emit events to parent
const emit = defineEmits(['group-selected'])

onMounted(async () => {
  await fetchChatGroups()
})

const fetchChatGroups = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name, description, id_user_owner, is_active 
        FROM chat_groups 
        WHERE is_active = 1 
        ORDER BY name ASC
      `,
      requiresAuth: true,
    })

    if (result.success && result.data) {
      chatGroups.value = result.data
    }
  } catch (err) {
    console.error('Error fetching chat groups:', err)
  }
}

const selectGroup = (group) => {
  selectedGroup.value = group
  emit('group-selected', group)
}
</script>

<template>
  <div class="chat-groups">
    <div class="groups-header">
      <h3><i class="fas fa-users"></i> Available Groups</h3>
    </div>
    <div class="groups-list">
      <div
        v-for="group in chatGroups"
        :key="group.id"
        @click="selectGroup(group)"
        :class="{ selected: selectedGroup?.id === group.id }"
        class="group-item"
      >
        <div class="group-name">{{ group.name }}</div>
        <div class="group-description">{{ group.description || 'No description' }}</div>
      </div>
      <div v-if="loading" class="loading">
        <i class="fas fa-spinner fa-spin"></i> Loading groups...
      </div>
      <div v-if="error" class="error"><i class="fas fa-exclamation-triangle"></i> {{ error }}</div>
      <div v-if="!loading && chatGroups.length === 0" class="no-groups">
        No chat groups available
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-groups {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.groups-header {
  padding: 16px 20px;
  background-color: #06b6d4;
  color: white;
}

.groups-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.1rem;
}

.groups-list {
  max-height: 400px;
  overflow-y: auto;
}

.group-item {
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
  cursor: pointer;
  transition: background-color 0.2s;
}

.group-item:hover {
  background-color: #f8fafc;
}

.group-item.selected {
  background-color: #e0f2fe;
  border-left: 3px solid #06b6d4;
}

.group-name {
  font-weight: 600;
  color: #374151;
  margin-bottom: 4px;
}

.group-description {
  font-size: 0.9rem;
  color: #64748b;
}

.loading,
.error,
.no-groups {
  padding: 20px;
  text-align: center;
  color: #64748b;
}

.error {
  color: #ef4444;
}
</style>
