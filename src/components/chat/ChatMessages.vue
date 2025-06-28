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

const emit = defineEmits(['message-sent'])

const { callApi, loading, error } = useApi()
const messages = ref([])
const newMessage = ref('')
const currentUser = ref(null)
const isSending = ref(false)
const messagesContainer = ref(null)

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

// Watch for changes in groupId to fetch messages
watch(
  [() => props.groupId],
  async ([groupId]) => {
    console.log('ChatMessages watch triggered - groupId:', groupId)
    if (groupId) {
      console.log('Group changed, fetching messages...')
      await getCurrentUser()
      await fetchMessages()
    } else {
      messages.value = []
    }
  },
  { immediate: true },
)

const fetchMessages = async () => {
  try {
    console.log('Fetching messages for group:', props.groupId)

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

    console.log('Messages result:', result)

    if (result.success && result.data) {
      const previousCount = messages.value.length
      messages.value = result.data
      console.log('Messages loaded:', messages.value)

      // Always scroll to bottom on initial load or when new messages are added
      if (previousCount === 0 || result.data.length > previousCount) {
        await scrollToBottom()
      }
    } else {
      console.log('No messages found or API failed:', result)
      messages.value = []
    }
  } catch (err) {
    console.error('Error fetching messages:', err)
    messages.value = []
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
      // Add the new message to the local array with UTC time
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
    if (props.groupId) {
      await fetchMessages()
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

// Expose cleanup function
defineExpose({ cleanup })
</script>

<template>
  <div class="chat-messages">
    <div class="messages-header">
      <h3><i class="fas fa-comments"></i> {{ groupName }}</h3>
      <div class="group-info">
        <span class="timezone-info">
          <i class="fas fa-clock"></i> {{ Intl.DateTimeFormat().resolvedOptions().timeZone }}
        </span>
        <span class="member-count">
          <i class="fas fa-users"></i> {{ messages.length }} messages
        </span>
      </div>
    </div>

    <div class="messages-container" ref="messagesContainer">
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
