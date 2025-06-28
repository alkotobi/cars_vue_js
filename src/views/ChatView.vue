<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import ChatSidebar from '../components/chat/ChatSidebar.vue'
import ChatMain from '../components/chat/ChatMain.vue'
import ChatMessages from '../components/chat/ChatMessages.vue'

const welcomeMessage = ref('')
const selectedGroup = ref(null)
const chatSidebarRef = ref(null)
const chatMainRef = ref(null)
const chatMessagesRef = ref(null) // Reference to the hidden ChatMessages component
const newMessagesCounts = ref({})

// Set up periodic updates for new message counts
let countsUpdateInterval = null

onMounted(async () => {
  // Get welcome message from URL parameters
  const urlParams = new URLSearchParams(window.location.search)
  welcomeMessage.value = urlParams.get('welcome') || 'Welcome to Chat!'

  // Check for group parameter in URL
  const groupId = urlParams.get('group')
  const taskId = urlParams.get('task')
  const taskName = urlParams.get('taskName')

  // Initialize all groups messages immediately using the hidden ChatMessages component
  await initializeAllGroupsMessages()

  // Set up periodic updates for new message counts
  countsUpdateInterval = setInterval(updateNewMessageCounts, 5000) // Update every 5 seconds
  // Initial update
  updateNewMessageCounts()

  // If group ID is provided in URL, select that group after a short delay
  if (groupId) {
    setTimeout(async () => {
      await selectGroupFromId(groupId, taskId, taskName)
    }, 1000) // Wait for groups to load
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
}

const handleMessageSent = (message) => {
  console.log('Message sent:', message)
}

const handleNewMessagesReceived = (groupId) => {
  console.log('New messages received for group:', groupId)
  console.log('chatMainRef.value:', chatMainRef.value)

  // Update new message counts from ChatMessages component
  if (chatMainRef.value?.getAllNewMessagesCounts) {
    const counts = chatMainRef.value.getAllNewMessagesCounts()
    console.log('Retrieved counts from new messages event:', counts)
    newMessagesCounts.value = counts
    console.log('Updated newMessagesCounts from event:', newMessagesCounts.value)
  } else {
    console.log('getAllNewMessagesCounts method not available in event handler')
  }
}

// Function to manually initialize all groups messages
const initializeAllGroupsMessages = async () => {
  try {
    console.log('Manually initializing all groups messages...')

    // Wait for the component to be mounted
    await new Promise((resolve) => setTimeout(resolve, 100))

    if (chatMessagesRef.value?.fetchAllGroupsMessages) {
      console.log('Calling fetchAllGroupsMessages from hidden component...')
      await chatMessagesRef.value.fetchAllGroupsMessages()
      console.log('fetchAllGroupsMessages completed')

      // Update counts after initialization
      setTimeout(() => {
        updateNewMessageCounts()
      }, 500)
    } else {
      console.log('fetchAllGroupsMessages method not available, retrying...')
      // Retry after a short delay
      setTimeout(async () => {
        if (chatMessagesRef.value?.fetchAllGroupsMessages) {
          await chatMessagesRef.value.fetchAllGroupsMessages()
          updateNewMessageCounts()
        } else {
          console.error('fetchAllGroupsMessages method still not available')
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
    console.log('Updating new message counts...')
    console.log('chatMessagesRef.value:', chatMessagesRef.value)

    if (chatMessagesRef.value?.getAllNewMessagesCounts) {
      const counts = chatMessagesRef.value.getAllNewMessagesCounts()
      console.log('Retrieved counts:', counts)
      newMessagesCounts.value = counts
      console.log('Updated newMessagesCounts:', newMessagesCounts.value)
    } else {
      console.log('getAllNewMessagesCounts method not available')
    }
  } catch (error) {
    console.error('Error updating new message counts:', error)
  }
}

// Function to force update counts after reset
const forceUpdateCounts = async (groupId) => {
  console.log('Force updating counts after reset for group:', groupId)

  // Directly reset the count in the hidden component
  if (chatMessagesRef.value?.resetGroupCount) {
    console.log('Directly resetting count in hidden component...')
    chatMessagesRef.value.resetGroupCount(groupId)
  }

  // Then update the UI counts
  updateNewMessageCounts()
}

// Function to select group from ID and focus message input
const selectGroupFromId = async (groupId, taskId, taskName) => {
  try {
    console.log('Selecting group from ID:', groupId, 'Task:', taskId, 'TaskName:', taskName)

    // Get the group from the sidebar component
    if (chatSidebarRef.value?.selectGroupById) {
      const group = await chatSidebarRef.value.selectGroupById(groupId)
      if (group) {
        selectedGroup.value = group
        console.log('Group selected:', group)

        // Focus the message input after a short delay
        setTimeout(() => {
          focusMessageInput()
        }, 500)
      } else {
        console.log('Group not found with ID:', groupId)
      }
    } else {
      console.log('selectGroupById method not available in sidebar')
    }
  } catch (error) {
    console.error('Error selecting group from ID:', error)
  }
}

// Function to focus the message input
const focusMessageInput = () => {
  console.log('Focusing message input...')
  if (chatMainRef.value?.focusMessageInput) {
    chatMainRef.value.focusMessageInput()
  } else {
    console.log('focusMessageInput method not available in chat main')
  }
}
</script>

<template>
  <div class="chat-layout">
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
.chat-layout {
  display: flex;
  height: 100vh;
  background-color: #f8fafc;
}

/* Responsive design */
@media (max-width: 768px) {
  .chat-layout {
    flex-direction: column;
  }
}
</style>
