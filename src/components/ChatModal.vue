<script setup>
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useEnhancedI18n } from '../composables/useI18n'
import { useApi } from '../composables/useApi'
import ChatSidebar from './chat/ChatSidebar.vue'
import ChatMain from './chat/ChatMain.vue'
import ChatMessages from './chat/ChatMessages.vue'

const { t } = useEnhancedI18n()
const { callApi } = useApi()

// Props for client mode (optional)
const props = defineProps({
  clientId: {
    type: Number,
    default: null,
  },
  clientName: {
    type: String,
    default: null,
  },
})

const welcomeMessage = computed(() => props.clientId 
  ? t('chat.chatWith', { name: props.clientName || 'Client' })
  : t('chat.welcomeToChat'))
const selectedGroup = ref(null)
const chatSidebarRef = ref(null)
const chatMainRef = ref(null)
const chatMessagesRef = ref(null)
const newMessagesCounts = ref({})

// Set up periodic updates for new message counts
let countsUpdateInterval = null

onMounted(async () => {
  // Initialize all groups messages immediately using the hidden ChatMessages component (only for users)
  if (!props.clientId) {
    await initializeAllGroupsMessages()
    // Set up periodic updates for new message counts (only for users)
    countsUpdateInterval = setInterval(updateNewMessageCounts, 5000) // Update every 5 seconds
    // Initial update
    updateNewMessageCounts()
  } else {
    // For clients, automatically select the first available group
    await selectFirstClientGroup()
  }
})

onUnmounted(() => {
  // Clean up intervals and resources
  if (chatSidebarRef.value?.cleanup) {
    chatSidebarRef.value.cleanup()
  }
  if (chatMainRef.value?.cleanup) {
    chatMainRef.value.cleanup()
  }
  if (chatMessagesRef.value?.cleanup) {
    chatMessagesRef.value.cleanup()
  }
  if (countsUpdateInterval) {
    clearInterval(countsUpdateInterval)
  }
})

const handleGroupSelected = (group) => {
  selectedGroup.value = group
  // Trigger an update of the counts when a group is selected
  setTimeout(() => {
    updateNewMessageCounts()
  }, 1000) // Wait 1 second for the group to load

  // Scroll to bottom when a group is selected
  setTimeout(async () => {
    if (chatMainRef.value?.scrollToBottom) {
      await chatMainRef.value.scrollToBottom()
    }
  }, 1500) // Wait a bit longer for messages to load
}

const handleMessageSent = (message) => {
  // Message sent event handler
  // Refresh groups list to update sorting by latest message
  if (chatSidebarRef.value?.refreshGroups) {
    chatSidebarRef.value.refreshGroups()
  }
}

const handleNewMessagesReceived = (groupId) => {
  // Update new message counts from ChatMessages component
  if (chatMainRef.value?.getAllNewMessagesCounts) {
    const counts = chatMainRef.value.getAllNewMessagesCounts()
    newMessagesCounts.value = counts
  }

  // Refresh groups list to update sorting by latest message
  if (chatSidebarRef.value?.refreshGroups) {
    chatSidebarRef.value.refreshGroups()
  }

  // Scroll to bottom when new messages are received
  setTimeout(async () => {
    if (chatMainRef.value?.scrollToBottom) {
      await chatMainRef.value.scrollToBottom()
    }
  }, 100)
}

// Function to manually initialize all groups messages
const initializeAllGroupsMessages = async () => {
  try {
    // Wait for the component to be mounted
    await new Promise((resolve) => setTimeout(resolve, 100))

    if (chatMessagesRef.value?.fetchAllGroupsMessages) {
      await chatMessagesRef.value.fetchAllGroupsMessages()

      // Update counts after initialization
      setTimeout(() => {
        updateNewMessageCounts()
      }, 500)
    } else {
      // Retry after a short delay
      setTimeout(async () => {
        if (chatMessagesRef.value?.fetchAllGroupsMessages) {
          await chatMessagesRef.value.fetchAllGroupsMessages()
          updateNewMessageCounts()
        }
      }, 1000)
    }
  } catch (error) {
    console.error('Error initializing all groups messages:', error)
  }
}

// Function to update new message counts periodically (only for users)
const updateNewMessageCounts = () => {
  if (props.clientId) {
    // Skip for clients - they don't need unread counts
    return
  }
  
  try {
    if (chatMessagesRef.value?.getAllNewMessagesCounts) {
      const counts = chatMessagesRef.value.getAllNewMessagesCounts()
      newMessagesCounts.value = counts
    }
    
    // Refresh groups list to update sorting by latest message
    if (chatSidebarRef.value?.refreshGroups) {
      chatSidebarRef.value.refreshGroups()
    }
  } catch (error) {
    console.error('Error updating new message counts:', error)
  }
}

// Function to force update counts after reset
const forceUpdateCounts = async (groupId) => {
  // Directly reset the count in the hidden component
  if (chatMessagesRef.value?.resetGroupCount) {
    chatMessagesRef.value.resetGroupCount(groupId)
  }

  // Then update the UI counts
  updateNewMessageCounts()
}

// Watch for group selection changes and scroll to bottom
watch(selectedGroup, async (newGroup, oldGroup) => {
  if (newGroup && newGroup !== oldGroup) {
    // Wait for messages to load and then scroll to bottom
    setTimeout(async () => {
      if (chatMainRef.value?.scrollToBottom) {
        await chatMainRef.value.scrollToBottom()
      }
    }, 1000) // Wait for messages to load
  }
})

// For clients: automatically select the first available group
const selectFirstClientGroup = async () => {
  if (!props.clientId) return

  try {
    // Fetch groups for this client
    const result = await callApi({
      query: `
        SELECT DISTINCT
          cg.id,
          cg.name,
          cg.description,
          cg.id_user_owner,
          cg.id_client_owner,
          cg.is_active,
          COALESCE(MAX(cm.time), '1970-01-01 00:00:00') as latest_message_time
        FROM chat_groups cg
        INNER JOIN chat_users cu_member ON cg.id = cu_member.id_chat_group
        LEFT JOIN chat_messages cm ON cg.id = cm.id_chat_group
        WHERE cg.is_active = 1
          AND cu_member.is_active = 1
          AND cu_member.id_client = ?
        GROUP BY cg.id, cg.name, cg.description, cg.id_user_owner, cg.id_client_owner, cg.is_active
        ORDER BY latest_message_time DESC, cg.name ASC
        LIMIT 1
      `,
      params: [props.clientId],
      requiresAuth: false,
    })

    if (result.success && result.data && result.data.length > 0) {
      const group = result.data[0]
      selectedGroup.value = group
      console.log('Auto-selected first group for client:', group)
    }
  } catch (err) {
    console.error('Error selecting first client group:', err)
  }
}

// Expose method to select group by ID (for external calls)
const selectGroupById = async (groupId) => {
  if (chatSidebarRef.value?.selectGroupById) {
    return await chatSidebarRef.value.selectGroupById(groupId)
  }
  return null
}

// Expose methods for parent components
defineExpose({
  selectGroupById,
})
</script>

<template>
  <div class="chat-modal-layout">
    <ChatSidebar
      v-if="!clientId"
      ref="chatSidebarRef"
      :new-messages-counts="newMessagesCounts"
      :client-id="clientId"
      :client-name="clientName"
      @group-selected="handleGroupSelected"
    />
    <ChatMain
      ref="chatMainRef"
      :welcome-message="welcomeMessage"
      :selected-group="selectedGroup"
      :on-force-update-counts="forceUpdateCounts"
      :client-id="clientId"
      :client-name="clientName"
      @message-sent="handleMessageSent"
      @new-messages-received="handleNewMessagesReceived"
    />

    <!-- Hidden ChatMessages component for initializing unread counts (only for users) -->
    <div v-if="!clientId" style="display: none">
      <ChatMessages
        ref="chatMessagesRef"
        :group-id="0"
        :group-name="'init'"
        @message-sent="handleMessageSent"
        @new-messages-received="handleNewMessagesReceived"
      />
    </div>
  </div>
</template>

<style scoped>
.chat-modal-layout {
  display: flex;
  height: 100%;
  min-height: 0;
  max-height: 100%;
  background-color: #f8fafc;
  overflow: hidden;
}


/* Responsive design for modal */
@media (max-width: 768px) {
  .chat-modal-layout {
    flex-direction: column;
    max-height: calc(100vh - 80px); /* Smaller header on mobile */
  }
}
</style>
