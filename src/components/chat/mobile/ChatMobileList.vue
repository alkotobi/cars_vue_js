<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useEnhancedI18n } from '../../../composables/useI18n'
import { useChatGroups } from '../../../composables/useChatGroups'
import ChatUsersPopup from '../ChatUsersPopup.vue'
import AddGroup from '../AddGroup.vue'

const { t } = useEnhancedI18n()
const router = useRouter()

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

const emit = defineEmits(['group-selected'])

const {
  chatGroups,
  allChatGroups,
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

// Bottom action sheet for a single card (members / leave).
const actionSheetGroup = ref(null)
const openActions = (group) => {
  actionSheetGroup.value = group
}
const closeActions = () => {
  actionSheetGroup.value = null
}

const preview = (group) => {
  const raw = (group.recent_messages || '').trim()
  if (!raw) return group.description || ''
  return raw.length > 80 ? '…' + raw.slice(-79) : raw
}

const formatGroupTime = (group) => {
  const raw = group.latest_message_time
  if (!raw || raw.startsWith('1970')) return ''
  const d = new Date(raw.replace(' ', 'T'))
  if (isNaN(d.getTime())) return ''
  const now = new Date()
  const sameDay =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  if (sameDay) {
    return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString(undefined, { day: '2-digit', month: '2-digit' })
}

const hasUnread = (group) => getNewMessageCount(group.id) > 0
const unreadLabel = (group) => {
  const n = getNewMessageCount(group.id)
  return n > 99 ? '99+' : n
}

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
  <div class="chat-mobile-list">
    <!-- Top bar -->
    <div class="mobile-topbar">
      <button class="topbar-btn" @click="router.push('/dashboard')" :title="t('navigation.dashboard')">
        <i class="fas fa-home"></i>
      </button>
      <h1 class="topbar-title">{{ t('navigation.chat') || 'Chat' }}</h1>
      <button
        v-if="!isClientMode"
        class="topbar-btn"
        @click="showAddGroup"
        :title="t('chat.addNewGroup')"
      >
        <i class="fas fa-plus"></i>
      </button>
      <span v-else class="topbar-spacer"></span>
    </div>

    <!-- Search -->
    <div class="mobile-search">
      <i class="fas fa-search search-icon"></i>
      <input
        type="text"
        v-model="searchFilter"
        @input="handleSearchInput"
        :placeholder="t('chat.searchGroups')"
        class="search-input"
      />
      <button v-if="searchFilter" @click="clearFilter" class="clear-search-btn" :title="t('chat.clearSearch')">
        <i class="fas fa-times"></i>
      </button>
    </div>

    <!-- List -->
    <div class="mobile-list-scroll">
      <div v-if="!groupsReady" class="state-block">
        <i class="fas fa-spinner fa-spin"></i>
        <p>{{ t('chat.loadingGroups') }}</p>
      </div>
      <div v-else-if="error" class="state-block error">
        <i class="fas fa-exclamation-triangle"></i>
        <p>{{ error }}</p>
      </div>
      <div v-else-if="chatGroups.length === 0" class="state-block">
        <i class="fas fa-comments"></i>
        <p>{{ t('chat.noGroupsFound') }}</p>
        <button v-if="!isClientMode" @click="showAddGroup" class="primary-btn">
          <i class="fas fa-plus"></i> {{ t('chat.addNewGroup') }}
        </button>
      </div>

      <button
        v-for="group in chatGroups"
        :key="group.id"
        class="group-card"
        :class="{ unread: hasUnread(group) }"
        @click="selectGroup(group)"
      >
        <span class="group-avatar">{{ (group.name || '?').charAt(0).toUpperCase() }}</span>
        <span class="group-body">
          <span class="group-row-top">
            <span class="group-name">{{ group.name }}</span>
            <span v-if="formatGroupTime(group)" class="group-time">{{ formatGroupTime(group) }}</span>
          </span>
          <span class="group-row-bottom">
            <span class="group-preview">{{ preview(group) }}</span>
            <span v-if="hasUnread(group)" class="unread-badge">{{ unreadLabel(group) }}</span>
          </span>
        </span>
        <span class="card-actions" @click.stop>
          <button class="card-action-btn" @click="openActions(group)" :title="t('chat.actions')">
            <i class="fas fa-ellipsis-v"></i>
          </button>
        </span>
      </button>
    </div>

    <!-- Action sheet -->
    <div v-if="actionSheetGroup" class="sheet-backdrop" @click="closeActions">
      <div class="action-sheet" @click.stop>
        <div class="sheet-title">{{ actionSheetGroup.name }}</div>
        <button class="sheet-item" @click="showUsers(actionSheetGroup, $event); actionSheetGroup = null">
          <i class="fas fa-users"></i> {{ t('chat.members') }}
        </button>
        <button
          v-if="!isClientMode && canLeaveGroup(actionSheetGroup)"
          class="sheet-item danger"
          @click="leaveGroup(actionSheetGroup, $event); actionSheetGroup = null"
        >
          <i class="fas fa-sign-out-alt"></i> {{ t('chat.leaveGroup') }}
        </button>
        <button class="sheet-item sheet-cancel" @click="actionSheetGroup = null">
          {{ t('chat.cancel') }}
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
.chat-mobile-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  background-color: #f8fafc;
}

.mobile-topbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  padding-top: calc(10px + env(safe-area-inset-top, 0px));
  background-color: #06b6d4;
  color: white;
  flex-shrink: 0;
}

.topbar-title {
  flex: 1;
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  text-align: left;
}

.topbar-btn {
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.18);
  color: white;
  font-size: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}

.topbar-btn:active {
  background-color: rgba(255, 255, 255, 0.32);
}

.topbar-spacer {
  width: 40px;
}

.mobile-search {
  position: relative;
  display: flex;
  align-items: center;
  padding: 10px 12px;
  background-color: #f1f5f9;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.search-icon {
  position: absolute;
  left: 24px;
  color: #64748b;
  font-size: 14px;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 10px 36px;
  border: 1px solid #d1d5db;
  border-radius: 22px;
  font-size: 15px;
  color: #374151;
  background-color: white;
}

.search-input:focus {
  outline: none;
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.clear-search-btn {
  position: absolute;
  right: 22px;
  background: none;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 6px;
  border-radius: 50%;
  display: flex;
  align-items: center;
}

.mobile-list-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  padding: 6px 0 16px;
}

.group-card {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 14px;
  border: none;
  background-color: white;
  border-bottom: 1px solid #f1f5f9;
  text-align: left;
  cursor: pointer;
  font-family: inherit;
}

.group-card:active {
  background-color: #f1f5f9;
}

.group-card.unread {
  background-color: #ecfeff;
}

.group-avatar {
  width: 46px;
  height: 46px;
  border-radius: 50%;
  background-color: #06b6d4;
  color: white;
  font-weight: 600;
  font-size: 1.15rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  text-transform: uppercase;
}

.group-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.group-row-top {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.group-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
  color: #0f172a;
  font-size: 0.95rem;
}

.group-time {
  font-size: 0.7rem;
  color: #94a3b8;
  flex-shrink: 0;
}

.group-row-bottom {
  display: flex;
  align-items: center;
  gap: 8px;
}

.group-preview {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.82rem;
  color: #64748b;
}

.unread-badge {
  background-color: #06b6d4;
  color: white;
  border-radius: 12px;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  font-size: 0.72rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.card-actions {
  flex-shrink: 0;
}

.card-action-btn {
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  color: #94a3b8;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-action-btn:active {
  background-color: #e2e8f0;
}

.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 56px 24px;
  text-align: center;
  color: #64748b;
}

.state-block i {
  font-size: 2.4rem;
  color: #06b6d4;
}

.state-block.error i {
  color: #ef4444;
}

.primary-btn {
  padding: 10px 18px;
  border: none;
  border-radius: 22px;
  background-color: #06b6d4;
  color: white;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.sheet-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(15, 23, 42, 0.45);
  display: flex;
  align-items: flex-end;
  z-index: 10050;
}

.action-sheet {
  width: 100%;
  background-color: white;
  border-radius: 16px 16px 0 0;
  padding: 8px 0 calc(8px + env(safe-area-inset-bottom, 0px));
  animation: sheet-up 0.18s ease-out;
}

@keyframes sheet-up {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}

.sheet-title {
  padding: 12px 20px;
  font-weight: 600;
  color: #0f172a;
  border-bottom: 1px solid #f1f5f9;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sheet-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 15px 20px;
  border: none;
  background: white;
  color: #334155;
  font-size: 1rem;
  font-family: inherit;
  text-align: left;
  cursor: pointer;
}

.sheet-item:active {
  background-color: #f1f5f9;
}

.sheet-item.danger {
  color: #ef4444;
}

.sheet-cancel {
  justify-content: center;
  color: #64748b;
  border-top: 1px solid #f1f5f9;
  margin-top: 4px;
}
</style>
