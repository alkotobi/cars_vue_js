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

// Mobile state
const showSidebar = ref(true)
const isMobile = ref(false)

// Set up periodic updates for new message counts
let countsUpdateInterval = null

onMounted(async () => {
  // Check if mobile
  checkMobile()
  window.addEventListener('resize', checkMobile)

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
  window.removeEventListener('resize', checkMobile)
})

const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
  if (isMobile.value && selectedGroup.value) {
    showSidebar.value = false
  }
}

const toggleSidebar = () => {
  showSidebar.value = !showSidebar.value
}

const handleGroupSelected = (group) => {
  selectedGroup.value = group
  // On mobile, hide sidebar when group is selected
  if (isMobile.value) {
    showSidebar.value = false
  }
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
    <!-- Mobile Navigation Bar -->
    <div v-if="isMobile" class="mobile-nav">
      <button v-if="!showSidebar" @click="toggleSidebar" class="mobile-nav-btn" title="Show Groups">
        <i class="fas fa-bars"></i>
        <span>Groups</span>
      </button>
      <button
        v-if="showSidebar && selectedGroup"
        @click="toggleSidebar"
        class="mobile-nav-btn"
        title="Show Chat"
      >
        <i class="fas fa-comments"></i>
        <span>{{ selectedGroup.name }}</span>
      </button>
      <div v-if="!showSidebar && selectedGroup" class="mobile-group-info">
        <h3>{{ selectedGroup.name }}</h3>
        <p>{{ selectedGroup.description || 'No description' }}</p>
      </div>
    </div>

    <!-- Sidebar -->
    <div
      v-show="!isMobile || showSidebar"
      class="chat-sidebar-container"
      :class="{ 'mobile-sidebar': isMobile }"
    >
      <ChatSidebar
        ref="chatSidebarRef"
        :new-messages-counts="newMessagesCounts"
        @group-selected="handleGroupSelected"
      />
    </div>

    <!-- Main Chat Area -->
    <div
      v-show="!isMobile || !showSidebar"
      class="chat-main-container"
      :class="{ 'mobile-main': isMobile }"
    >
      <ChatMain
        ref="chatMainRef"
        :welcome-message="welcomeMessage"
        :selected-group="selectedGroup"
        :on-force-update-counts="forceUpdateCounts"
        @message-sent="handleMessageSent"
        @new-messages-received="handleNewMessagesReceived"
      />
    </div>

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
  position: relative;
}

/* Mobile Navigation */
.mobile-nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background-color: #06b6d4;
  color: white;
  display: flex;
  align-items: center;
  padding: 0 16px;
  z-index: 1000;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.mobile-nav-btn {
  background: none;
  border: none;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background-color 0.2s;
}

.mobile-nav-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

.mobile-nav-btn i {
  font-size: 1.1rem;
}

.mobile-group-info {
  flex: 1;
  margin-left: 16px;
  overflow: hidden;
}

.mobile-group-info h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mobile-group-info p {
  margin: 4px 0 0 0;
  font-size: 0.9rem;
  opacity: 0.9;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Sidebar Container */
.chat-sidebar-container {
  width: 400px;
  background-color: white;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
}

.chat-sidebar-container.mobile-sidebar {
  position: fixed;
  top: 60px;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  z-index: 999;
  border-right: none;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
}

/* Main Chat Container */
.chat-main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.chat-main-container.mobile-main {
  margin-top: 60px;
  height: calc(100vh - 60px);
}

/* Responsive design */
@media (max-width: 768px) {
  .chat-layout {
    flex-direction: column;
  }

  .chat-sidebar-container {
    width: 100%;
  }

  .chat-main-container {
    width: 100%;
  }
}

/* Animation for mobile sidebar */
@keyframes slideIn {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(0);
  }
}

@keyframes slideOut {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(-100%);
  }
}

.mobile-sidebar {
  animation: slideIn 0.3s ease-out;
}

/* Ensure proper z-index layering */
.chat-sidebar-container {
  z-index: 10;
}

.chat-main-container {
  z-index: 5;
}

/* Touch-friendly improvements for mobile */
@media (max-width: 768px) {
  .mobile-nav-btn {
    min-height: 44px;
    min-width: 44px;
  }
}

/* Landscape orientation adjustments */
@media (max-width: 768px) and (orientation: landscape) {
  .mobile-nav {
    height: 50px;
  }

  .chat-main-container.mobile-main {
    margin-top: 50px;
    height: calc(100vh - 50px);
  }

  .chat-sidebar-container.mobile-sidebar {
    top: 50px;
  }
}
</style>
