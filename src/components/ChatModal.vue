<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import ChatSidebar from './chat/ChatSidebar.vue'
import ChatMain from './chat/ChatMain.vue'
import ChatMessages from './chat/ChatMessages.vue'

const welcomeMessage = ref('Welcome to Team Chat!')
const selectedGroup = ref(null)
const chatSidebarRef = ref(null)
const chatMainRef = ref(null)
const chatMessagesRef = ref(null)
const newMessagesCounts = ref({})

// Set up periodic updates for new message counts
let countsUpdateInterval = null

onMounted(async () => {
  // Initialize all groups messages immediately using the hidden ChatMessages component
  await initializeAllGroupsMessages()

  // Set up periodic updates for new message counts
  countsUpdateInterval = setInterval(updateNewMessageCounts, 5000) // Update every 5 seconds
  // Initial update
  updateNewMessageCounts()
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
}

const handleNewMessagesReceived = (groupId) => {
  // Update new message counts from ChatMessages component
  if (chatMainRef.value?.getAllNewMessagesCounts) {
    const counts = chatMainRef.value.getAllNewMessagesCounts()
    newMessagesCounts.value = counts
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

// Function to update new message counts periodically
const updateNewMessageCounts = () => {
  try {
    if (chatMessagesRef.value?.getAllNewMessagesCounts) {
      const counts = chatMessagesRef.value.getAllNewMessagesCounts()
      newMessagesCounts.value = counts
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
</script>

<template>
  <div class="chat-modal-layout">
    <ChatSidebar
      ref="chatSidebarRef"
      :new-messages-counts="newMessagesCounts"
      @group-selected="handleGroupSelected"
    />
    <ChatMain
      ref="chatMainRef"
      :welcome-message="welcomeMessage"
      :selected-group="selectedGroup"
      :on-force-update-counts="forceUpdateCounts"
      @message-sent="handleMessageSent"
      @new-messages-received="handleNewMessagesReceived"
    />

    <!-- Hidden ChatMessages component for initializing unread counts -->
    <div style="display: none">
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
  max-height: calc(100vh - 140px); /* Account for modal header */
  background-color: #f8fafc;
  overflow: hidden;
}

/* Responsive design for modal */
@media (max-width: 768px) {
  .chat-modal-layout {
    flex-direction: column;
    max-height: calc(100vh - 120px); /* Smaller header on mobile */
  }
}
</style>
