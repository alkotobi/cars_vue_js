<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  groupId: {
    type: Number,
    required: true,
  },
  groupName: {
    type: String,
    required: true,
  },
})

const emit = defineEmits(['message-sent', 'new-messages-received'])

const { callApi, loading, error } = useApi()
const messages = ref([]) // Current group's messages
const messagesByGroup = ref({}) // Store messages for each group: { groupId: [messages] }
const newMessagesCountByGroup = ref({}) // Track new messages count per group: { groupId: count }
const newMessage = ref('')
const currentUser = ref(null)
const isSending = ref(false)
const messagesContainer = ref(null)

// Debug section state
const showDebugSection = ref(false)
const debugInfo = ref('')
const currentNewMessagesCount = ref(0) // Track current newMessagesCount value

const getCurrentUser = () => {
  try {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      currentUser.value = JSON.parse(userStr)
      console.log('Current user loaded for chat:', currentUser.value)
    }
  } catch (err) {
    console.error('Error getting current user:', err)
  }
}

// Watch for changes in groupId to switch message lists
watch(
  [() => props.groupId],
  async ([groupId]) => {
    console.log('ChatMessages watch triggered - groupId:', groupId)
    if (groupId) {
      console.log('Group changed, switching to cached messages or fetching if needed...')
      await getCurrentUser()
      await switchToGroup(groupId)
    } else {
      messages.value = []
    }
  },
  { immediate: true },
)

const switchToGroup = async (groupId) => {
  // Check if we already have messages for this group
  if (messagesByGroup.value[groupId]) {
    console.log(`Using cached messages for group ${groupId}`)
    messages.value = messagesByGroup.value[groupId]
    // Reset new message count when switching to this group (user clicked on group)
    newMessagesCountByGroup.value[groupId] = 0
    currentNewMessagesCount.value = 0
    console.log(`Reset new message count for group ${groupId} (group clicked)`)
    await scrollToBottom()
  } else {
    console.log(`No cached messages for group ${groupId}, fetching all messages...`)
    await fetchAllMessages()
  }
}

const fetchAllMessages = async () => {
  try {
    console.log('Fetching all messages for group:', props.groupId)

    const result = await callApi({
      query: `
        SELECT 
          cm.id,
          cm.id_chat_group,
          cm.message_from_user_id,
          cm.chat_replay_to_message_id,
          cm.message,
          cm.time,
          u.username as sender_username,
          r.role_name as sender_role
        FROM chat_messages cm
        LEFT JOIN users u ON cm.message_from_user_id = u.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE cm.id_chat_group = ?
        ORDER BY cm.id ASC
      `,
      params: [props.groupId],
      requiresAuth: true,
    })

    console.log('All messages result:', result)

    if (result.success && result.data) {
      // Store messages in the group-specific cache
      messagesByGroup.value[props.groupId] = result.data
      // Set current messages
      messages.value = result.data
      // Initialize new messages count to 0 for this group
      newMessagesCountByGroup.value[props.groupId] = 0
      console.log('All messages loaded and cached:', messages.value.length)

      // Scroll to bottom on initial load
      await scrollToBottom()
    } else {
      console.log('No messages found or API failed:', result)
      messagesByGroup.value[props.groupId] = []
      messages.value = []
      newMessagesCountByGroup.value[props.groupId] = 0
    }
  } catch (err) {
    console.error('Error fetching all messages:', err)
    messagesByGroup.value[props.groupId] = []
    messages.value = []
  }
}

const fetchMessages = async (groupId = props.groupId) => {
  try {
    console.log('Fetching new messages for group:', groupId)

    // Get the highest message ID from current messages for this group
    const currentMessages = messagesByGroup.value[groupId] || []
    const currentHighestId =
      currentMessages.length > 0 ? Math.max(...currentMessages.map((msg) => msg.id)) : 0

    console.log('Current highest message ID:', currentHighestId)

    const result = await callApi({
      query: `
        SELECT 
          cm.id,
          cm.id_chat_group,
          cm.message_from_user_id,
          cm.chat_replay_to_message_id,
          cm.message,
          cm.time,
          u.username as sender_username,
          r.role_name as sender_role
        FROM chat_messages cm
        LEFT JOIN users u ON cm.message_from_user_id = u.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE cm.id_chat_group = ?
          AND cm.id > ?
        ORDER BY cm.id ASC
      `,
      params: [groupId, currentHighestId],
      requiresAuth: true,
    })

    console.log('New messages result:', result)

    if (result.success && result.data) {
      const newMessagesCount = result.data.length
      // Update debug variable only if this is the current group
      if (groupId === props.groupId) {
        currentNewMessagesCount.value = newMessagesCount
      }
      console.log(`Found ${newMessagesCount} new messages`)

      if (newMessagesCount > 0) {
        // Add new messages to the group cache
        if (messagesByGroup.value[groupId]) {
          messagesByGroup.value[groupId].push(...result.data)
        } else {
          messagesByGroup.value[groupId] = [...result.data]
        }

        // Update current messages if this is the active group
        if (groupId === props.groupId) {
          messages.value = [...messagesByGroup.value[groupId]]
        }

        // Increment new messages count for this group
        if (newMessagesCountByGroup.value[groupId] !== undefined) {
          newMessagesCountByGroup.value[groupId] += newMessagesCount
        } else {
          newMessagesCountByGroup.value[groupId] = newMessagesCount
        }

        console.log(
          'New messages added to list. Total messages:',
          messagesByGroup.value[groupId]?.length || 0,
        )
        console.log(
          `New messages count for group ${groupId}:`,
          newMessagesCountByGroup.value[groupId],
        )
        console.log('All new message counts:', newMessagesCountByGroup.value)

        // Update debug info if debug section is open and this is the current group
        if (showDebugSection.value && groupId === props.groupId) {
          updateDebugInfo()
        }

        // Notify parent about new messages only if this is the current group
        if (groupId === props.groupId) {
          console.log(`Emitting new-messages-received for group ${groupId}`)
          emit('new-messages-received', groupId)

          // Scroll to bottom when new messages are added
          await scrollToBottom()
        }
      } else {
        // Update debug info even when no new messages (only for current group)
        if (showDebugSection.value && groupId === props.groupId) {
          updateDebugInfo()
        }
      }
    } else {
      console.log('No new messages found or API failed:', result)
      if (groupId === props.groupId) {
        currentNewMessagesCount.value = 0
        if (showDebugSection.value) {
          updateDebugInfo()
        }
      }
    }
  } catch (err) {
    console.error('Error fetching new messages:', err)
  }
}

const sendMessage = async () => {
  if (!newMessage.value.trim() || !currentUser.value) {
    return
  }

  try {
    isSending.value = true
    console.log('Sending message:', newMessage.value)

    const result = await callApi({
      query: `
        INSERT INTO chat_messages (id_chat_group, message_from_user_id, message, time) 
        VALUES (?, ?, ?, UTC_TIMESTAMP())
      `,
      params: [props.groupId, currentUser.value.id, newMessage.value.trim()],
      requiresAuth: true,
    })

    console.log('Send message result:', result)

    if (result.success) {
      // Add the new message to both current list and group cache
      const newMsg = {
        id: result.lastInsertId,
        id_chat_group: props.groupId,
        message_from_user_id: currentUser.value.id,
        message: newMessage.value.trim(),
        time: new Date().toISOString(), // Use UTC time
        sender_username: currentUser.value.username,
        sender_role: currentUser.value.role_name,
      }

      messages.value.push(newMsg)
      messagesByGroup.value[props.groupId] = [...messages.value]
      newMessage.value = ''

      // Emit event to parent
      emit('message-sent', newMsg)

      // Scroll to bottom after message is added
      await nextTick()
      await scrollToBottom()
    } else {
      alert('Failed to send message')
    }
  } catch (err) {
    console.error('Error sending message:', err)
    alert('Error sending message')
  } finally {
    isSending.value = false
  }
}

const scrollToBottom = async () => {
  if (messagesContainer.value) {
    await nextTick()
    // Use multiple timeouts to ensure DOM is fully updated
    setTimeout(() => {
      if (messagesContainer.value) {
        messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
      }
    }, 50)

    // Double-check after a longer delay
    setTimeout(() => {
      if (messagesContainer.value) {
        messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
      }
    }, 200)
  }
}

const formatTime = (timeString) => {
  const date = new Date(timeString)
  const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone
  const timeZoneAbbr =
    new Intl.DateTimeFormat('en', { timeZoneName: 'short' })
      .formatToParts(date)
      .find((part) => part.type === 'timeZoneName')?.value || ''

  return date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    timeZoneName: 'short',
  })
}

const formatDate = (timeString) => {
  const date = new Date(timeString)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  } else if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  } else {
    return date.toLocaleDateString()
  }
}

const getRelativeTime = (timeString) => {
  const date = new Date(timeString)
  const now = new Date()
  const diffInMinutes = Math.floor((now - date) / (1000 * 60))

  if (diffInMinutes < 1) return 'Just now'
  if (diffInMinutes < 60) return `${diffInMinutes}m ago`

  const diffInHours = Math.floor(diffInMinutes / 60)
  if (diffInHours < 24) return `${diffInHours}h ago`

  const diffInDays = Math.floor(diffInHours / 24)
  if (diffInDays < 7) return `${diffInDays}d ago`

  return date.toLocaleDateString()
}

const isOwnMessage = (message) => {
  return message.message_from_user_id === currentUser.value?.id
}

const handleKeyPress = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

// Auto-refresh messages every 10 seconds
let refreshInterval = null

onMounted(() => {
  refreshInterval = setInterval(async () => {
    // Fetch messages for all groups that have been loaded
    const groupIds = Object.keys(messagesByGroup.value)
    for (const groupId of groupIds) {
      await fetchMessages(parseInt(groupId))
    }
  }, 10000) // 10 seconds
})

// Clean up interval on unmount
const cleanup = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

// Expose cleanup function and new message counts
defineExpose({
  cleanup,
  getNewMessagesCount: (groupId) => {
    const count = newMessagesCountByGroup.value[groupId] || 0
    console.log(`Exposed getNewMessagesCount for group ${groupId}:`, count)
    return count
  },
  getAllNewMessagesCounts: () => {
    console.log('Exposed getAllNewMessagesCounts:', newMessagesCountByGroup.value)
    return newMessagesCountByGroup.value
  },
})

// Debug functions
const toggleDebugSection = () => {
  showDebugSection.value = !showDebugSection.value
  if (showDebugSection.value) {
    updateDebugInfo()
  }
}

const updateDebugInfo = () => {
  let info = `=== DEBUG: NEW MESSAGES COUNT ===\n`
  info += `Current Group: ${props.groupName} (ID: ${props.groupId})\n`
  info += `Current User: ${currentUser.value?.username} (ID: ${currentUser.value?.id})\n\n`

  info += `=== CURRENT STATE ===\n`
  info += `Current newMessagesCount: ${currentNewMessagesCount.value}\n`
  info += `Total messages in current group: ${messages.value.length}\n`
  info += `Messages cached for this group: ${messagesByGroup.value[props.groupId]?.length || 0}\n\n`

  info += `=== ALL GROUPS COUNTS ===\n`
  if (Object.keys(newMessagesCountByGroup.value).length === 0) {
    info += `No new message counts tracked yet\n`
  } else {
    Object.keys(newMessagesCountByGroup.value).forEach((groupId) => {
      const count = newMessagesCountByGroup.value[groupId]
      info += `Group ${groupId}: ${count} new messages\n`
    })
  }

  info += `\n=== SYSTEM INFO ===\n`
  info += `Last Updated: ${new Date().toLocaleString()}\n`
  info += `Polling Interval: 10 seconds\n`
  info += `Debug Section Active: ${showDebugSection.value}\n`

  debugInfo.value = info
}

const testNewMessagesCount = () => {
  console.log('=== TESTING NEW MESSAGES COUNT ===')
  console.log('Current newMessagesCount:', currentNewMessagesCount.value)
  console.log('All newMessagesCountByGroup:', newMessagesCountByGroup.value)
  updateDebugInfo()
}

// Click handlers to reset new message count
const handleMessagesContainerClick = () => {
  if (newMessagesCountByGroup.value[props.groupId] > 0) {
    console.log(`Reset new message count for group ${props.groupId} (chat area clicked)`)
    newMessagesCountByGroup.value[props.groupId] = 0
    currentNewMessagesCount.value = 0
    if (showDebugSection.value) {
      updateDebugInfo()
    }
  }
}

const handleInputClick = () => {
  if (newMessagesCountByGroup.value[props.groupId] > 0) {
    console.log(`Reset new message count for group ${props.groupId} (input clicked)`)
    newMessagesCountByGroup.value[props.groupId] = 0
    currentNewMessagesCount.value = 0
    if (showDebugSection.value) {
      updateDebugInfo()
    }
  }
}
</script>

<template>
  <div class="chat-messages">
    <div class="messages-header">
      <h3><i class="fas fa-comments"></i> {{ groupName }}</h3>
      <div class="group-info">
        <button @click="toggleDebugSection" class="debug-btn" title="Debug New Messages Count">
          <i class="fas fa-bug"></i>
        </button>
        <span class="timezone-info">
          <i class="fas fa-clock"></i> {{ Intl.DateTimeFormat().resolvedOptions().timeZone }}
        </span>
        <span class="member-count">
          <i class="fas fa-users"></i> {{ messages.length }} messages
        </span>
      </div>
    </div>

    <div class="messages-container" ref="messagesContainer" @click="handleMessagesContainerClick">
      <div v-if="loading && messages.length === 0" class="loading-messages">
        <i class="fas fa-spinner fa-spin"></i> Loading messages...
      </div>

      <div v-else-if="error" class="error-messages">
        <i class="fas fa-exclamation-triangle"></i> {{ error }}
      </div>

      <div v-else-if="messages.length === 0" class="no-messages">
        <i class="fas fa-comment-slash"></i>
        <p>No messages yet</p>
        <p class="sub-text">Start the conversation!</p>
      </div>

      <div v-else class="messages-list">
        <div
          v-for="(message, index) in messages"
          :key="message.id"
          class="message-wrapper"
          :class="{ 'own-message': isOwnMessage(message) }"
        >
          <!-- Date separator -->
          <div
            v-if="index === 0 || formatDate(message.time) !== formatDate(messages[index - 1].time)"
            class="date-separator"
          >
            {{ formatDate(message.time) }}
          </div>

          <div class="message">
            <div class="message-header">
              <span class="sender-name">{{ message.sender_username }}</span>
              <div class="time-info">
                <span class="relative-time">{{ getRelativeTime(message.time) }}</span>
                <span class="absolute-time">{{ formatTime(message.time) }}</span>
              </div>
            </div>

            <div class="message-content">
              {{ message.message }}
            </div>

            <div class="message-footer">
              <span class="sender-role">{{ message.sender_role }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="message-input-container">
      <div class="input-wrapper">
        <textarea
          v-model="newMessage"
          @keypress="handleKeyPress"
          @click="handleInputClick"
          placeholder="Type your message..."
          class="message-input"
          :disabled="isSending"
          rows="1"
          maxlength="1000"
        ></textarea>
        <button
          @click="sendMessage"
          :disabled="!newMessage.trim() || isSending"
          class="send-button"
        >
          <i :class="isSending ? 'fas fa-spinner fa-spin' : 'fas fa-paper-plane'"></i>
        </button>
      </div>
      <div class="input-footer">
        <span class="char-count">{{ newMessage.length }}/1000</span>
        <span class="send-hint">Press Enter to send, Shift+Enter for new line</span>
      </div>
    </div>

    <!-- Debug Info Modal -->
    <div v-if="showDebugSection" class="debug-info-modal">
      <div class="debug-info-content">
        <h3>Debug: New Messages Count</h3>
        <pre>{{ debugInfo }}</pre>
        <div class="debug-buttons">
          <button @click="testNewMessagesCount" class="test-btn">Test Count</button>
          <button @click="toggleDebugSection">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-messages {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.messages-header {
  padding: 16px 20px;
  background-color: #06b6d4;
  color: white;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.messages-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.1rem;
}

.group-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.timezone-info {
  font-size: 0.9rem;
  opacity: 0.9;
  display: flex;
  align-items: center;
  gap: 4px;
}

.member-count {
  font-size: 0.9rem;
  opacity: 0.9;
  display: flex;
  align-items: center;
  gap: 4px;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background-color: #f8fafc;
}

.loading-messages,
.error-messages,
.no-messages {
  text-align: center;
  padding: 40px 20px;
  color: #64748b;
}

.error-messages {
  color: #ef4444;
}

.no-messages i {
  font-size: 3rem;
  color: #06b6d4;
  margin-bottom: 16px;
  display: block;
}

.sub-text {
  font-size: 0.9rem;
  color: #9ca3af;
  margin-top: 8px;
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-wrapper {
  display: flex;
  flex-direction: column;
}

.message-wrapper.own-message {
  align-items: flex-end;
}

.date-separator {
  text-align: center;
  margin: 20px 0 12px 0;
  color: #64748b;
  font-size: 0.8rem;
  font-weight: 500;
  position: relative;
}

.date-separator::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background-color: #e2e8f0;
  z-index: 1;
}

.date-separator span {
  background-color: #f8fafc;
  padding: 0 12px;
  position: relative;
  z-index: 2;
}

.message {
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 12px;
  background-color: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  border: 1px solid #e2e8f0;
}

.own-message .message {
  background-color: #06b6d4;
  color: white;
  border-color: #06b6d4;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.sender-name {
  font-weight: 600;
  font-size: 0.9rem;
}

.own-message .sender-name {
  color: white;
}

.time-info {
  font-size: 0.8rem;
  opacity: 0.7;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.own-message .time-info {
  color: rgba(255, 255, 255, 0.8);
}

.relative-time {
  font-size: 0.75rem;
  font-weight: 500;
}

.absolute-time {
  font-size: 0.7rem;
  opacity: 0.8;
}

.message-content {
  line-height: 1.4;
  word-wrap: break-word;
  white-space: pre-wrap;
}

.message-footer {
  margin-top: 4px;
}

.sender-role {
  font-size: 0.7rem;
  opacity: 0.6;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.own-message .sender-role {
  color: rgba(255, 255, 255, 0.7);
}

.message-input-container {
  padding: 16px 20px;
  background-color: white;
  border-top: 1px solid #e2e8f0;
}

.input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.message-input {
  flex: 1;
  padding: 12px 16px;
  border: 1px solid #d1d5db;
  border-radius: 20px;
  font-size: 1rem;
  resize: none;
  outline: none;
  transition: border-color 0.2s;
  font-family: inherit;
  line-height: 1.4;
}

.message-input:focus {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.message-input:disabled {
  background-color: #f5f7fa;
  cursor: not-allowed;
}

.send-button {
  background-color: #06b6d4;
  color: white;
  border: none;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  flex-shrink: 0;
}

.send-button:hover:not(:disabled) {
  background-color: #0891b2;
}

.send-button:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.input-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
  font-size: 0.8rem;
  color: #64748b;
}

.char-count {
  font-weight: 500;
}

.send-hint {
  font-style: italic;
}

.debug-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 0.8rem;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 4px;
}

.debug-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.debug-info-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.debug-info-content {
  background-color: white;
  padding: 24px;
  border-radius: 12px;
  width: 90%;
  max-width: 700px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.debug-info-content h3 {
  margin-top: 0;
  margin-bottom: 16px;
  color: #06b6d4;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 8px;
}

.debug-info-content pre {
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  background-color: #f8fafc;
  padding: 16px;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
  max-height: 400px;
  overflow-y: auto;
}

.debug-buttons {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}

.debug-info-content button {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  background-color: #06b6d4;
  color: white;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.debug-info-content button:hover {
  background-color: #0891b2;
}

.test-btn {
  background-color: #10b981 !important;
}

.test-btn:hover {
  background-color: #059669 !important;
}

/* Responsive design */
@media (max-width: 768px) {
  .message {
    max-width: 85%;
  }

  .messages-header {
    padding: 12px 16px;
  }

  .messages-container {
    padding: 16px;
  }

  .message-input-container {
    padding: 12px 16px;
  }

  .input-footer {
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
  }
}
</style>
