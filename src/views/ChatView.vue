<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import ChatSidebar from '../components/chat/ChatSidebar.vue'
import ChatMain from '../components/chat/ChatMain.vue'
import ChatMessages from '../components/chat/ChatMessages.vue'
import ChatMobileList from '../components/chat/mobile/ChatMobileList.vue'
import ChatMobileConversation from '../components/chat/mobile/ChatMobileConversation.vue'

const welcomeMessage = ref('')
const selectedGroup = ref(null)
const chatSidebarRef = ref(null)
const chatMainRef = ref(null)
const chatMessagesRef = ref(null) // Reference to the hidden ChatMessages component
const newMessagesCounts = ref({})
const isMobile = ref(false)
// 1024, not 768: the chat page serves its mobile list/conversation UI up to
// tablet width (iPad portrait 820-1024, phones in landscape). Above that it is
// the desktop sidebar+main layout.
const mobileQuery =
  typeof window !== 'undefined' ? window.matchMedia('(max-width: 1024px)') : null

const setIsMobile = () => {
  isMobile.value = mobileQuery ? mobileQuery.matches : false
}

// Set up periodic updates for new message counts
let countsUpdateInterval = null

onMounted(async () => {
  setIsMobile()
  if (mobileQuery) {
    mobileQuery.addEventListener('change', setIsMobile)
  }

  // Get welcome message from URL parameters
  const urlParams = new URLSearchParams(window.location.search)
  welcomeMessage.value = urlParams.get('welcome') || 'Welcome to Chat!'

  // Check for group parameter in URL
  const groupId = urlParams.get('group')
  const taskId = urlParams.get('task')
  const taskName = urlParams.get('taskName')

  // The hidden ChatMessages is always rendered, so its ref is available now.
  await nextTick()

  // Kick off the unread-count initialization WITHOUT awaiting it. This used to be
  // awaited before selectGroupFromId, so a deep-linked group could not open until
  // two round trips of unread initialization had finished. The counts arrive when
  // they arrive; opening the group no longer waits on them.
  initializeAllGroupsMessages()

  // Set up periodic updates for new message counts. 10s matches the header's own
  // active-chat cadence and the message poll; the previous 5s doubled the idle DB
  // load for badges that rarely change. pollUnreadCounts also skips hidden tabs.
  countsUpdateInterval = setInterval(pollUnreadCounts, 10000)
  // Initial update
  updateNewMessageCounts()

  // If group ID is provided in URL, wait for the sidebar to load its groups
  // (instead of an arbitrary timeout) and then select that group.
  if (groupId) {
    const readyRef = isMobile.value ? mobileListRef.value : chatSidebarRef.value
    if (readyRef?.whenReady) {
      await readyRef.whenReady()
    }
    await selectGroupFromId(groupId, taskId, taskName)
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
  if (mobileConversationRef.value?.cleanup) {
    mobileConversationRef.value.cleanup()
  }
  if (mobileListRef.value?.cleanup) {
    mobileListRef.value.cleanup()
  }
  if (countsUpdateInterval) {
    clearInterval(countsUpdateInterval)
  }
  if (mobileQuery) {
    mobileQuery.removeEventListener('change', setIsMobile)
  }
})

const handleGroupSelected = (group) => {
  selectedGroup.value = group
  // Opening a group marks its messages as read, so immediately clear its badge
  // (the visible ChatMessages persists the last-read id; the hidden engine is
  //  synced here and refreshed on the next poll).
  if (group?.id != null) {
    forceUpdateCounts(group.id)
  }
}

const handleMessageSent = (message) => {
  // Message sent event handler
}

const handleNewMessagesReceived = (groupId) => {
  // Update new message counts from ChatMessages component
  if (chatMainRef.value?.getAllNewMessagesCounts) {
    const counts = chatMainRef.value.getAllNewMessagesCounts()
    if (!sameCounts(newMessagesCounts.value, counts)) {
      newMessagesCounts.value = counts
    }
  }
}

// Function to manually initialize all groups unread counts
const initializeAllGroupsMessages = async () => {
  try {
    if (chatMessagesRef.value?.refreshUnreadCounts) {
      await chatMessagesRef.value.refreshUnreadCounts()
    } else if (chatMessagesRef.value?.fetchAllGroupsMessages) {
      // Fallback for older builds of ChatMessages
      await chatMessagesRef.value.fetchAllGroupsMessages()
    }
    updateNewMessageCounts()
  } catch (error) {
    console.error('Error initializing all groups messages:', error)
  }
}

// Shallow compare of two { groupId: count } maps. The child returns a fresh object
// from every poll, so assigning it unconditionally made ChatSidebar re-render its
// whole group list every tick even when no badge changed.
const sameCounts = (a, b) => {
  const aKeys = Object.keys(a)
  const bKeys = Object.keys(b)
  if (aKeys.length !== bKeys.length) return false
  for (const key of aKeys) {
    if ((a[key] || 0) !== (b[key] || 0)) return false
  }
  return true
}

// Function to update new message counts periodically
const updateNewMessageCounts = () => {
  try {
    if (chatMessagesRef.value?.getAllNewMessagesCounts) {
      const counts = chatMessagesRef.value.getAllNewMessagesCounts()
      if (!sameCounts(newMessagesCounts.value, counts)) {
        newMessagesCounts.value = counts
      }
    }
  } catch (error) {
    console.error('Error updating new message counts:', error)
  }
}

// Recompute unread counts from the backend, then push them to the UI. Guarded so
// a slow poll can't stack behind itself, and skipped entirely while the tab is
// hidden (nothing to show). Once the artificial per-request delay is gone these
// are a couple of cheap COUNT queries.
let unreadPollInFlight = false
const pollUnreadCounts = async () => {
  if (unreadPollInFlight || document.hidden) {
    return
  }
  unreadPollInFlight = true
  try {
    if (chatMessagesRef.value?.refreshUnreadCounts) {
      await chatMessagesRef.value.refreshUnreadCounts()
    }
  } catch (error) {
    console.error('Error polling unread counts:', error)
  } finally {
    unreadPollInFlight = false
  }
  updateNewMessageCounts()
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

// Function to select group from ID and focus message input
const selectGroupFromId = async (groupId, taskId, taskName) => {
  try {
    // Get the group from the sidebar/list component
    const selector = isMobile.value ? mobileListRef.value : chatSidebarRef.value
    if (selector?.selectGroupById) {
      const group = await selector.selectGroupById(groupId)
      if (group) {
        selectedGroup.value = group
        // Wait for ChatMessages to render before focusing the input
        await nextTick()
        focusMessageInput()
      }
    }
  } catch (error) {
    console.error('Error selecting group from ID:', error)
  }
}

// Function to focus the message input
const focusMessageInput = () => {
  if (isMobile.value) {
    if (mobileConversationRef.value?.focusMessageInput) {
      mobileConversationRef.value.focusMessageInput()
    }
  } else {
    if (chatMainRef.value?.focusMessageInput) {
      chatMainRef.value.focusMessageInput()
    }
  }
}

const mobileListRef = ref(null)
const mobileConversationRef = ref(null)

const backToMobileList = () => {
  selectedGroup.value = null
}
</script>

<template>
  <div class="chat-layout" :class="{ 'chat-layout--mobile': isMobile }">
    <!-- Desktop layout: sidebar + main conversation (keep as before) -->
    <template v-if="!isMobile">
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
    </template>

    <!-- Mobile layout: immersive list ↔ full-screen conversation -->
    <template v-else>
      <ChatMobileList
        v-if="!selectedGroup"
        ref="mobileListRef"
        :new-messages-counts="newMessagesCounts"
        @group-selected="handleGroupSelected"
      />
      <ChatMobileConversation
        v-else
        ref="mobileConversationRef"
        :group-id="selectedGroup.id"
        :group-name="selectedGroup.name"
        @back="backToMobileList"
        @message-sent="handleMessageSent"
        @new-messages-received="handleNewMessagesReceived"
        @reset-triggered="forceUpdateCounts(selectedGroup.id)"
      />
    </template>

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




/* Mobile immersive chat: hide chrome is handled at app level. When mobile
 * the layout fills the viewport without the 60px header padding. */
.chat-layout--mobile {
  height: 100dvh;
  height: 100vh; /* fallback */
  max-height: 100dvh;
  max-height: 100vh; /* fallback */
  background-color: #f8fafc;
}


/* Responsive design */
@media (max-width: 768px) {
  .chat-layout {
    flex-direction: column;
  }
}

/* Mobile immersive chat: hide chrome is handled at app level. When mobile
 * the layout fills the viewport without the 60px header padding. */
.chat-layout--mobile {
  height: 100dvh;
  height: 100vh;
  max-height: 100dvh;
  max-height: 100vh;
  background-color: #f8fafc;
}



</style>
