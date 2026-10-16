<script setup>
import { onMounted } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useChatGroups } from '../../composables/useChatGroups'
import ChatUsersPopup from './ChatUsersPopup.vue'
import AddGroup from './AddGroup.vue'

const { t } = useEnhancedI18n()

const props = defineProps({
  newMessagesCounts: {
    type: Object,
    default: () => ({}),
  },
  // Client mode props (optional)
  clientId: {
    type: Number,
    default: null,
  },
  clientName: {
    type: String,
    default: null,
  },
})

// Emit events to parent
const emit = defineEmits(['group-selected'])

// All group data/behaviour comes from the shared composable so the desktop
// sidebar and the mobile list can never drift apart.
const {
  chatGroups,
  selectedGroup,
  searchFilter,
  isClientMode,
  showUsersPopup,
  selectedGroupForPopup,
  showAddGroupModal,
  groupsReady,
  loading,
  error,
  init,
  fetchChatGroups,
  selectGroup,
  selectGroupById,
  showUsers,
  closeUsersPopup,
  showAddGroup,
  closeAddGroupModal,
  handleGroupAdded,
  handleSearchInput,
  clearFilter,
  leaveGroup,
  canLeaveGroup,
  getNewMessageCount,
  refreshGroups,
  getChatGroups,
  whenReady,
  isReady,
} = useChatGroups(props, {
  t,
  onGroupSelected: (group) => emit('group-selected', group),
})

onMounted(() => {
  init()
})

// Expose methods to parent
defineExpose({
  cleanup: () => {
    // Cleanup function if needed
  },
  whenReady,
  isReady,
  selectGroupById,
  refreshGroups,
  fetchChatGroups,
  getChatGroups,
})
</script>

<template>
  <div class="chat-sidebar">
    <div class="sidebar-header">
      <h3><i class="fas fa-comments"></i> {{ t('navigation.teams') || 'Groups' }}</h3>
      <div class="header-buttons">
        <button v-if="!isClientMode" @click="showAddGroup" class="add-group-btn" :title="t('chat.addNewGroup')">
          <i class="fas fa-plus"></i>
        </button>
      </div>
    </div>
    
    <!-- Search Filter -->
    <div class="search-filter-container">
      <div class="search-input-wrapper">
        <i class="fas fa-search search-icon"></i>
        <input
          type="text"
          v-model="searchFilter"
          @input="handleSearchInput"
          :placeholder="t('chat.searchGroups')"
          class="search-input"
        />
        <button
          v-if="searchFilter"
          @click="clearFilter"
          class="clear-search-btn"
          :title="t('chat.clearSearch')"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div v-if="searchFilter" class="filter-info">
        {{ chatGroups.length }} / {{ allChatGroups.length }} groups
      </div>
    </div>

    <div class="groups-table">
      <table>
        <thead>
          <tr>
            <th>{{ t('cars.teamName') || 'Group Name' }}</th>
            <th>{{ t('cars.description') || 'Description' }}</th>
            <th>{{ t('cars.actions') || 'Actions' }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="group in chatGroups"
            :key="group.id"
            @click="selectGroup(group)"
            :class="{ selected: selectedGroup?.id === group.id }"
            class="group-row"
          >
            <td>
              <div class="group-name-container">
                <span class="group-name">{{ group.name }}</span>
                <span v-if="getNewMessageCount(group.id) > 0" class="new-messages-badge">
                  {{ getNewMessageCount(group.id) > 99 ? '99+' : getNewMessageCount(group.id) }}
                </span>
              </div>
            </td>
            <td>{{ group.description || (t('cars.description') || 'No description') }}</td>
            <td>
              <div class="action-buttons">
                <button @click="showUsers(group, $event)" class="users-btn" :title="t('chat.members')">
                  <i class="fas fa-users"></i>
                </button>
                <button 
                  v-if="!isClientMode && canLeaveGroup(group)" 
                  @click="leaveGroup(group, $event)" 
                  class="leave-btn" 
                  :title="t('chat.leaveGroup')"
                >
                  <i class="fas fa-sign-out-alt"></i>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <!-- Only the very first load shows the caption. Keying it off `loading`
           re-showed it on every 5s refresh; keying it off groupsReady paints once
           and never blanks the list again. -->
      <div v-if="!groupsReady" class="loading">
        <i class="fas fa-spinner fa-spin"></i> {{ t('chat.loadingGroups') }}
      </div>
      <div v-if="error" class="error"><i class="fas fa-exclamation-triangle"></i> {{ error }}</div>
      <div v-if="!loading && chatGroups.length === 0 && !searchFilter" class="no-groups">
        <i class="fas fa-comments"></i>
        <p>{{ t('chat.noGroupsFound') }}</p>
        <button v-if="!isClientMode" @click="showAddGroup" class="btn btn-primary">
          <i class="fas fa-plus"></i> {{ t('chat.addNewGroup') }}
        </button>
      </div>
      <div v-if="!loading && chatGroups.length === 0 && searchFilter" class="no-results">
        <i class="fas fa-search"></i>
        <p>{{ t('chat.noGroupsFound') }}</p>
        <button @click="clearFilter" class="btn btn-secondary">
          {{ t('chat.clearSearch') }}
        </button>
      </div>
    </div>

    <!-- Users Popup -->
    <ChatUsersPopup
      v-if="selectedGroupForPopup"
      :group-id="selectedGroupForPopup.id"
      :group-name="selectedGroupForPopup.name"
      :is-visible="showUsersPopup"
      @close="closeUsersPopup"
    />

    <!-- Add Group Modal (only for users) -->
    <AddGroup
      v-if="!isClientMode"
      :is-visible="showAddGroupModal"
      @close="closeAddGroupModal"
      @group-added="handleGroupAdded"
    />
  </div>
</template>

<style scoped>
.chat-sidebar {
  width: 400px;
  background-color: white;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
  background-color: #06b6d4;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-filter-container {
  padding: 12px 16px;
  border-bottom: 1px solid #e2e8f0;
  background-color: #f8fafc;
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: #64748b;
  font-size: 14px;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 36px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  color: #374151;
  background-color: white;
  transition: all 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.search-input::placeholder {
  color: #9ca3af;
}

.clear-search-btn {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
  font-size: 12px;
}

.clear-search-btn:hover {
  background-color: #e5e7eb;
  color: #374151;
}

.filter-info {
  margin-top: 8px;
  font-size: 12px;
  color: #64748b;
  text-align: right;
}

.header-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.debug-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.debug-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.add-group-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.add-group-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.groups-table {
  flex: 1;
  overflow-y: auto;
}

.groups-table table {
  width: 100%;
  border-collapse: collapse;
}

.groups-table th {
  background-color: #f1f5f9;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
}

.groups-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f1f5f9;
  color: #374151;
}

.group-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.group-row:hover {
  background-color: #f8fafc;
}

.group-row.selected {
  background-color: #e0f2fe;
  border-left: 3px solid #06b6d4;
}

.group-name-container {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.group-name {
  flex: 1;
}

.new-messages-badge {
  background-color: #ef4444;
  color: white;
  border-radius: 12px;
  padding: 2px 8px;
  font-size: 0.7rem;
  font-weight: 600;
  min-width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.users-btn {
  background-color: #06b6d4;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: background-color 0.2s;
}

.users-btn:hover {
  background-color: #0891b2;
}

.leave-btn {
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: background-color 0.2s;
}

.leave-btn:hover {
  background-color: #dc2626;
}

.loading,
.error,
.no-groups,
.no-results {
  padding: 20px;
  text-align: center;
  color: #64748b;
}

.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.no-results i {
  font-size: 2rem;
  color: #9ca3af;
}

.no-results p {
  margin: 0;
  font-size: 0.9rem;
}

.btn-secondary {
  background-color: #6b7280;
  color: white;
}

.btn-secondary:hover {
  background-color: #4b5563;
}

.error {
  color: #ef4444;
}

.no-groups {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.no-groups i {
  font-size: 3rem;
  color: #06b6d4;
}

.no-groups p {
  margin: 0;
  font-size: 1.1rem;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-primary {
  background-color: #06b6d4;
  color: white;
}

.btn-primary:hover {
  background-color: #0891b2;
}

/* Responsive design */
@media (max-width: 768px) {
  .chat-sidebar {
    width: 100%;
    height: 200px;
  }
}
</style>
