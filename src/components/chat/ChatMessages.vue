<script setup>
import { ref, onMounted, watch, nextTick, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '../../composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  groupId: {
    type: Number,
    required: true,
  },
  groupName: {
    type: String,
    required: true,
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

const emit = defineEmits(['message-sent', 'new-messages-received', 'reset-triggered'])

const { callApi, uploadFile, getFileUrl, loading, error } = useApi()
const messages = ref([]) // Current group's messages
const messagesByGroup = ref({}) // Store messages for each group: { groupId: [messages] }
const newMessagesCountByGroup = ref({}) // Track new messages count per group: { groupId: count }
const newMessage = ref('')
const currentUser = ref(null)
const currentClient = ref(null)
const isClientMode = computed(() => props.clientId !== null)
const isSending = ref(false)
const messagesContainer = ref(null)
const messageInputRef = ref(null)
const fileInputRef = ref(null)

// Pagination state
const messagesPerPage = 20
const currentPage = ref(1)
const hasMoreMessages = ref(true)
const isLoadingMore = ref(false)
const totalMessages = ref(0)

// File upload state
const isUploading = ref(false)
const uploadProgress = ref(0)
const selectedFile = ref(null)
const fileError = ref('')

// Voice message state
const isRecording = ref(false)
const recordingTime = ref(0)
const recordingInterval = ref(null)
const mediaRecorder = ref(null)
const audioChunks = ref([])
const audioBlob = ref(null)
const audioUrl = ref('')

// File size limit (100MB)
const MAX_FILE_SIZE = 100 * 1024 * 1024 // 100MB in bytes

// Debug section state
const showDebugSection = ref(false)
const debugInfo = ref({
  user: null,
  groupId: null,
  messagesCount: 0,
  apiResponses: [],
  errors: [],
  lastFetch: null,
})
const currentNewMessagesCount = ref(0) // Track current newMessagesCount value

// Emoji picker state
const showEmojiPicker = ref(false)

// Message date editing state
const editingMessageId = ref(null)
const editingMessageDate = ref(null)
const showDatePicker = ref(false)

// Bulk date editing state
const selectionMode = ref(false)
const selectedMessages = ref(new Set())
const showBulkDatePicker = ref(false)
const bulkEditDate = ref(null)

// Message search state
const messageSearch = ref('')

// Track if current user/client is an active member of the current group
const isActiveMember = ref(true)

// Common emojis for the picker
const emojis = [
  '😀',
  '😃',
  '😄',
  '😁',
  '😆',
  '😅',
  '😂',
  '🤣',
  '😊',
  '😇',
  '🙂',
  '🙃',
  '😉',
  '😌',
  '😍',
  '🥰',
  '😘',
  '😗',
  '😙',
  '😚',
  '😋',
  '😛',
  '😝',
  '😜',
  '🤪',
  '🤨',
  '🧐',
  '🤓',
  '😎',
  '🤩',
  '🥳',
  '😏',
  '😒',
  '😞',
  '😔',
  '😟',
  '😕',
  '🙁',
  '☹️',
  '😣',
  '😖',
  '😫',
  '😩',
  '🥺',
  '😢',
  '😭',
  '😤',
  '😠',
  '😡',
  '🤬',
  '🤯',
  '😳',
  '🥵',
  '🥶',
  '😱',
  '😨',
  '😰',
  '😥',
  '😓',
  '🤗',
  '🤔',
  '🤭',
  '🤫',
  '🤥',
  '😶',
  '😐',
  '😑',
  '😯',
  '😦',
  '😧',
  '😮',
  '😲',
  '🥱',
  '😴',
  '🤤',
  '😪',
  '😵',
  '🤐',
  '🥴',
  '🤢',
  '🤮',
  '🤧',
  '😷',
  '🤒',
  '🤕',
  '🤑',
  '🤠',
  '💩',
  '👻',
  '💀',
  '☠️',
  '👽',
  '👾',
  '🤖',
  '😺',
  '😸',
  '😹',
  '😻',
  '😼',
  '😽',
  '🙀',
  '😿',
  '😾',
  '🙈',
  '🙉',
  '🙊',
  '👶',
  '👧',
  '🧒',
  '👦',
  '👩',
  '🧑',
  '👨',
  '👵',
  '🧓',
  '👴',
  '👮‍♀️',
  '👮',
  '👮‍♂️',
  '🕵️‍♀️',
  '🕵️',
  '🕵️‍♂️',
  '💂‍♀️',
  '💂',
  '💂‍♂️',
  '👷‍♀️',
  '👷',
  '👷‍♂️',
  '🤴',
  '👸',
  '👳‍♀️',
  '👳',
  '👳‍♂️',
  '👲',
  '🧕',
  '��‍♀️',
  '🤵',
  '🤵‍♂️',
  '👰‍♀️',
  '👰',
  '👰‍♂️',
  '🤰',
  '🤱',
  '👼',
  '🎅',
  '🤶',
  '🧙‍♀️',
  '🧙',
  '🧙‍♂️',
  '🧝‍♀️',
  '🧝',
  '🧝‍♂️',
  '🧛‍♀️',
  '🧛',
  '🧛‍♂️',
  '🧟‍♀️',
  '🧟',
  '🧟‍♂️',
  '🧞‍♀️',
  '🧞',
  '🧞‍♂️',
  '🧜‍♀️',
  '🧜',
  '🧜‍♂️',
  '🧚‍♀️',
  '🧚',
  '🧚‍♂️',
  '👼',
  '🤰',
  '🤱',
  '❤️',
  '🧡',
  '💛',
  '💚',
  '💙',
  '💜',
  '🖤',
  '🤍',
  '🤎',
  '💔',
  '❣️',
  '💕',
  '💞',
  '💓',
  '💗',
  '💖',
  '💘',
  '💝',
  '💟',
  '☮️',
  '✝️',
  '☪️',
  '🕉️',
  '☸️',
  '✡️',
  '🔯',
  '🕎',
  '☯️',
  '☦️',
  '🛐',
  '⛎',
  '♈',
  '♉',
  '♊',
  '♋',
  '♌',
  '♍',
  '♎',
  '♏',
  '♐',
  '♑',
  '♒',
  '♓',
  '🆔',
  '⚛️',
  '🉑',
  '☢️',
  '☣️',
  '📴',
  '📳',
  '🈶',
  '🈚',
  '🈸',
  '🈺',
  '🈷️',
  '✴️',
  '🆚',
  '💮',
  '🉐',
  '㊙️',
  '㊗️',
  '🈴',
  '🈵',
  '🈹',
  '🈲',
  '🅰️',
  '🅱️',
  '🆎',
  '🆑',
  '🅾️',
  '🆘',
  '❌',
  '⭕',
  '🛑',
  '⛔',
  '📛',
  '🚫',
  '💯',
  '💢',
  '♨️',
  '🚷',
  '🚯',
  '🚳',
  '🚱',
  '🔞',
  '📵',
  '🚭',
  '❗',
  '❕',
  '❓',
  '❔',
  '‼️',
  '⁉️',
  '🔅',
  '🔆',
  '〽️',
  '⚠️',
  '🚸',
  '🔱',
  '⚜️',
  '🔰',
  '♻️',
  '✅',
  '🈯',
  '💹',
  '❇️',
  '✳️',
  '❎',
  '🌐',
  '💠',
  'Ⓜ️',
  '🌀',
  '💤',
  '🏧',
  '🚾',
  '♿',
  '🅿️',
  '🛗',
  '🛂',
  '🛃',
  '🛄',
  '🛅',
  '🚹',
  '🚺',
  '🚼',
  '🚻',
  '🚮',
  '🎦',
  '📶',
  '🈁',
  '🔣',
  'ℹ️',
  '🔤',
  '🔡',
  '🔠',
  '🆖',
  '🆗',
  '🆙',
  '🆒',
  '🆕',
  '🆓',
  '0️⃣',
  '1️⃣',
  '2️⃣',
  '3️⃣',
  '4️⃣',
  '5️⃣',
  '6️⃣',
  '7️⃣',
  '8️⃣',
  '9️⃣',
  '🔟',
  '🔢',
  '#️⃣',
  '*️⃣',
  '⏏️',
  '▶️',
  '⏸️',
  '⏯️',
  '⏹️',
  '⏺️',
  '⏭️',
  '⏮️',
  '⏩',
  '⏪',
  '⏫',
  '⏬',
  '◀️',
  '🔼',
  '🔽',
  '➡️',
  '⬅️',
  '⬆️',
  '⬇️',
  '↗️',
  '↘️',
  '↙️',
  '↖️',
  '↕️',
  '↔️',
  '↪️',
  '↩️',
  '⤴️',
  '⤵️',
  '🔀',
  '🔁',
  '🔂',
  '🔄',
  '🔃',
  '🎵',
  '🎶',
  '➕',
  '➖',
  '➗',
  '✖️',
  '♾️',
  '💲',
  '💱',
  '™️',
  '©️',
  '®️',
  '👁️‍🗨️',
  '🔚',
  '🔙',
  '🔛',
  '🔝',
  '🔜',
  '〰️',
  '➰',
  '➿',
  '✔️',
  '☑️',
  '🔘',
  '🔴',
  '🟠',
  '🟡',
  '🟢',
  '🔵',
  '🟣',
  '⚫',
  '⚪',
  '🟤',
  '🔺',
  '🔻',
  '🔸',
  '🔹',
  '🔶',
  '🔷',
  '🔳',
  '🔲',
  '▪️',
  '▫️',
  '◾',
  '◽',
  '◼️',
  '◻️',
  '🟥',
  '🟧',
  '🟨',
  '🟩',
  '🟦',
  '🟪',
  '⬛',
  '⬜',
  '🟫',
  '🔈',
  '🔇',
  '🔉',
  '🔊',
  '🔔',
  '🔕',
  '📣',
  '📢',
  '💬',
  '💭',
  '🗯️',
  '♠️',
  '♣️',
  '♥️',
  '♦️',
  '🃏',
  '🎴',
  '🀄',
  '🕐',
  '🕑',
  '🕒',
  '🕓',
  '🕔',
  '🕕',
  '🕖',
  '🕗',
  '🕘',
  '🕙',
  '🕚',
  '🕛',
  '🕜',
  '🕝',
  '🕞',
  '🕟',
  '🕠',
  '🕡',
  '🕢',
  '🕣',
  '🕤',
  '🕥',
  '🕦',
  '🕧',
]

// Get current user
const getCurrentUser = () => {
  if (isClientMode.value) {
    // In client mode, set client info
    currentClient.value = { id: props.clientId, name: props.clientName }
    addDebugInfo('clientLoaded', currentClient.value)
    return
  }
  
  try {
    const userStr = localStorage.getItem('user')
    addDebugInfo('getCurrentUser', { userStr })

    if (userStr) {
      currentUser.value = JSON.parse(userStr)
      addDebugInfo('userLoaded', currentUser.value)
      updateDebugInfo()
    } else {
      addError('No user found in localStorage')
      currentUser.value = null
    }
  } catch (err) {
    addError(`Error getting current user: ${err.message}`)
    currentUser.value = null
  }
}

// Reset pagination when switching groups
const resetPagination = () => {
  currentPage.value = 1
  hasMoreMessages.value = true
  isLoadingMore.value = false
}

// Check if user/client is an active member of the group
const checkMembership = async () => {
  if (!props.groupId || props.groupId === 0) {
    isActiveMember.value = false
    return
  }

  try {
    let query, params, requiresAuth

    if (isClientMode.value) {
      if (!currentClient.value) {
        isActiveMember.value = false
        return
      }
      query = `
        SELECT COUNT(*) as count
        FROM chat_users
        WHERE id_chat_group = ? 
          AND id_client = ? 
          AND is_active = 1
      `
      params = [props.groupId, currentClient.value.id]
      requiresAuth = false
    } else {
      if (!currentUser.value) {
        isActiveMember.value = false
        return
      }
      query = `
        SELECT COUNT(*) as count
        FROM chat_users
        WHERE id_chat_group = ? 
          AND id_user = ? 
          AND is_active = 1
      `
      params = [props.groupId, currentUser.value.id]
      requiresAuth = true
    }

    const result = await callApi({
      query,
      params,
      requiresAuth,
    })

    if (result.success && result.data && result.data.length > 0) {
      isActiveMember.value = result.data[0].count > 0
    } else {
      isActiveMember.value = false
    }
  } catch (err) {
    console.error('Error checking membership:', err)
    isActiveMember.value = false
  }
}

const switchToGroup = async (groupId) => {
  // Check if we already have messages for this group
  if (messagesByGroup.value[groupId]) {
    messages.value = messagesByGroup.value[groupId]
    // Reset new message count when switching to this group (user clicked on group)
    await resetNewMessageCount(groupId)
    await scrollToBottom()
  } else {
    await fetchAllMessages()
  }

  // Focus the message input after switching groups
  await nextTick()
  if (messageInputRef.value) {
    messageInputRef.value.focus()
  }
}

const fetchAllMessages = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT 
          cm.id,
          cm.id_chat_group,
          cm.message_from_user_id,
          cm.message_from_client_id,
          cm.chat_replay_to_message_id,
          cm.message,
          cm.time,
          COALESCE(u.username, c.name) as sender_username,
          COALESCE(r.role_name, 'Client') as sender_role,
          CASE WHEN cm.message_from_user_id IS NOT NULL THEN 'user' ELSE 'client' END as sender_type
        FROM chat_messages cm
        LEFT JOIN users u ON cm.message_from_user_id = u.id
        LEFT JOIN roles r ON u.role_id = r.id
        LEFT JOIN clients c ON cm.message_from_client_id = c.id
        WHERE cm.id_chat_group = ?
        ORDER BY cm.id ASC
      `,
      params: [props.groupId],
      requiresAuth: !isClientMode.value,
    })

    if (result.success && result.data) {
      // Store messages in the group-specific cache
      messagesByGroup.value[props.groupId] = result.data
      // Set current messages
      messages.value = result.data
      // Initialize new messages count to 0 for this group
      await resetNewMessageCount(props.groupId)

      // Scroll to bottom on initial load
      await scrollToBottom()

      // Focus the message input after loading messages
      await nextTick()
      if (messageInputRef.value) {
        messageInputRef.value.focus()
      }
    } else {
      messagesByGroup.value[props.groupId] = []
      messages.value = []
      await resetNewMessageCount(props.groupId)
    }
  } catch (err) {
    console.error('Error fetching all messages:', err)
    messagesByGroup.value[props.groupId] = []
    messages.value = []
  }
}

const fetchMessages = async (groupId) => {
  try {
    if (isClientMode.value) {
      if (!currentClient.value) {
        addError(t('chat.noCurrentClient'))
        return
      }
    } else {
      if (!currentUser.value) {
        addError(t('chat.noCurrentUser'))
        return
      }
    }

    addDebugInfo('fetchMessages', { 
      groupId, 
      page: currentPage.value, 
      user: currentUser.value,
      client: currentClient.value,
      isClientMode: isClientMode.value
    })

    const offset = (currentPage.value - 1) * messagesPerPage

    const result = await callApi({
      query: `
        SELECT 
          cm.id,
          cm.id_chat_group,
          cm.message_from_user_id,
          cm.message_from_client_id,
          cm.message,
          cm.time,
          COALESCE(u.username, c.name) as sender_username,
          COALESCE(r.role_name, 'Client') as sender_role,
          CASE WHEN cm.message_from_user_id IS NOT NULL THEN 'user' ELSE 'client' END as sender_type
        FROM chat_messages cm
        LEFT JOIN users u ON cm.message_from_user_id = u.id
        LEFT JOIN roles r ON u.role_id = r.id
        LEFT JOIN clients c ON cm.message_from_client_id = c.id
        WHERE cm.id_chat_group = ?
        ORDER BY cm.id DESC
        LIMIT ${messagesPerPage} OFFSET ${offset}
      `,
      params: [groupId],
      requiresAuth: !isClientMode.value,
    })

    addDebugInfo('fetchMessages_result', result)

    if (result.success && result.data) {
      const newMessages = result.data.reverse() // Reverse to get chronological order

      if (currentPage.value === 1) {
        // First page - replace all messages
        messages.value = newMessages
        messagesByGroup.value[groupId] = [...newMessages]
      } else {
        // Load more - prepend to existing messages
        messages.value = [...newMessages, ...messages.value]
        messagesByGroup.value[groupId] = [...newMessages, ...messagesByGroup.value[groupId]]
      }

      // Check if we have more messages to load
      hasMoreMessages.value = newMessages.length === messagesPerPage

      addDebugInfo('messagesLoaded', {
        count: newMessages.length,
        total: messages.value.length,
        hasMore: hasMoreMessages.value,
      })

      updateDebugInfo()
    } else {
      addError(`No messages found or API failed: ${result.message || 'No error message'}`)
      if (currentPage.value === 1) {
        messages.value = []
        messagesByGroup.value[groupId] = []
      }
      hasMoreMessages.value = false
    }
  } catch (err) {
    addError(`Error fetching messages: ${err.message}`)
    if (currentPage.value === 1) {
      messages.value = []
      messagesByGroup.value[groupId] = []
    }
    hasMoreMessages.value = false
  }
}

const sendMessage = async () => {
  if (isClientMode.value) {
    if (!newMessage.value.trim() || !currentClient.value) {
      return
    }
  } else {
    if (!newMessage.value.trim() || !currentUser.value) {
      return
    }
  }

  // Check if user/client is still an active member before sending
  await checkMembership()
  
  if (!isActiveMember.value) {
    alert(t('chat.youHaveLeftThisGroup') || 'You have left this group and cannot send messages.')
    // Clear message input
    newMessage.value = ''
    // Trigger reset to deselect group
    emit('reset-triggered')
    return
  }

  try {
    isSending.value = true

    let query, params, requiresAuth, newMsg

    if (isClientMode.value) {
      // Client sending message
      query = `
        INSERT INTO chat_messages (id_chat_group, message_from_client_id, message, time) 
        VALUES (?, ?, ?, UTC_TIMESTAMP())
      `
      params = [props.groupId, currentClient.value.id, newMessage.value.trim()]
      requiresAuth = false
    } else {
      // User sending message
      query = `
        INSERT INTO chat_messages (id_chat_group, message_from_user_id, message, time) 
        VALUES (?, ?, ?, UTC_TIMESTAMP())
      `
      params = [props.groupId, currentUser.value.id, newMessage.value.trim()]
      requiresAuth = true
    }

    const result = await callApi({
      query,
      params,
      requiresAuth,
    })

    if (result.success) {
      // Add the new message to both current list and group cache
      if (isClientMode.value) {
        newMsg = {
          id: result.lastInsertId,
          id_chat_group: props.groupId,
          message_from_client_id: currentClient.value.id,
          message: newMessage.value.trim(),
          time: new Date().toISOString(),
          sender_username: currentClient.value.name,
          sender_role: 'Client',
          sender_type: 'client',
        }
      } else {
        newMsg = {
          id: result.lastInsertId,
          id_chat_group: props.groupId,
          message_from_user_id: currentUser.value.id,
          message: newMessage.value.trim(),
          time: new Date().toISOString(),
          sender_username: currentUser.value.username,
          sender_role: currentUser.value.role_name,
          sender_type: 'user',
        }
      }

      messages.value.push(newMsg)
      messagesByGroup.value[props.groupId] = [...messages.value]
      newMessage.value = ''

      // Scroll to bottom after message is added
      await nextTick()
      await scrollToBottom()

      // Emit event to parent
      emit('message-sent', newMsg)

      // Dispatch global event to update header badge
      window.dispatchEvent(new Event('forceUpdateBadge'))
    } else {
      alert(t('chat.failedToSendMessage'))
    }
  } catch (err) {
    console.error('Error sending message:', err)
    alert(t('chat.errorSendingMessage'))
  } finally {
    isSending.value = false

    // Focus the message input after sending is complete and textarea is enabled
    setTimeout(() => {
      requestAnimationFrame(() => {
        if (messageInputRef.value) {
          messageInputRef.value.focus()
        } else {
          // Removed: console.log('Message input ref not available for focus')
        }
      })
    }, 100)
  }
}

const scrollToBottom = async () => {
  if (messagesContainer.value) {
    await nextTick()

    // Multiple attempts with different delays to ensure scrolling works
    const scrollAttempts = [
      () => {
        if (messagesContainer.value) {
          messagesContainer.value.scrollTo({
            top: messagesContainer.value.scrollHeight,
            behavior: 'smooth',
          })
        }
      },
      () => {
        if (messagesContainer.value) {
          messagesContainer.value.scrollTo({
            top: messagesContainer.value.scrollHeight,
            behavior: 'smooth',
          })
        }
      },
      () => {
        if (messagesContainer.value) {
          messagesContainer.value.scrollTo({
            top: messagesContainer.value.scrollHeight,
            behavior: 'smooth',
          })
        }
      },
    ]

    // Execute scroll attempts with different delays
    scrollAttempts.forEach((attempt, index) => {
      setTimeout(attempt, 50 * (index + 1))
    })

    // Final attempt after a longer delay
    setTimeout(() => {
      if (messagesContainer.value) {
        messagesContainer.value.scrollTo({
          top: messagesContainer.value.scrollHeight,
          behavior: 'smooth',
        })
      }
    }, 300)
  } else {
    // Removed: console.log('ChatMessages: messagesContainer ref not available for scrolling')
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
  if (isClientMode.value) {
    if (!currentClient.value) return false
    return message.message_from_client_id === currentClient.value.id
  } else {
    if (!currentUser.value) return false
    return message.message_from_user_id === currentUser.value.id
  }
}

// Check if user can edit message date (admin or message owner)
const canEditMessageDate = (message) => {
  if (isClientMode.value) {
    // Clients cannot edit message dates
    return false
  }
  if (!currentUser.value) return false
  // Admin can edit any message
  if (currentUser.value.role_id === 1) return true
  // Message owner can edit their own messages
  return message.message_from_user_id === currentUser.value.id
}

// Check if user is admin
const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Toggle selection mode (admin only)
const toggleSelectionMode = () => {
  if (!isAdmin.value) return
  selectionMode.value = !selectionMode.value
  if (!selectionMode.value) {
    selectedMessages.value.clear()
  }
}

// Toggle message selection
const toggleMessageSelection = (messageId) => {
  if (!selectionMode.value) return
  if (selectedMessages.value.has(messageId)) {
    selectedMessages.value.delete(messageId)
  } else {
    selectedMessages.value.add(messageId)
  }
}

// Select all messages
const selectAllMessages = () => {
  if (!selectionMode.value) return
  messages.value.forEach((message) => {
    selectedMessages.value.add(message.id)
  })
}

// Deselect all messages
const deselectAllMessages = () => {
  selectedMessages.value.clear()
}

// Open bulk date editor
const openBulkDateEditor = () => {
  if (selectedMessages.value.size === 0) {
    alert('Please select at least one message')
    return
  }
  // Set current date as default (date only, no time)
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  bulkEditDate.value = `${year}-${month}-${day}`
  showBulkDatePicker.value = true
}

// Close bulk date editor
const closeBulkDateEditor = () => {
  showBulkDatePicker.value = false
  bulkEditDate.value = null
}

// Update bulk message dates (preserving original time, only changing date)
const updateBulkMessageDates = async () => {
  if (!bulkEditDate.value || selectedMessages.value.size === 0) return

  try {
    // Get the new date (month and day) from the picker
    const newDate = new Date(bulkEditDate.value)
    const newYear = newDate.getFullYear()
    const newMonth = newDate.getMonth()
    const newDay = newDate.getDate()

    // Get selected messages with their original times
    const selectedMessagesList = messages.value.filter((message) =>
      selectedMessages.value.has(message.id),
    )

    // Update each message individually to preserve its original time
    let successCount = 0
    let failCount = 0

    for (const message of selectedMessagesList) {
      try {
        // Get original message time
        const originalDate = new Date(message.time)
        const originalHours = originalDate.getHours()
        const originalMinutes = originalDate.getMinutes()
        const originalSeconds = originalDate.getSeconds()

        // Create new date with new year/month/day but original time
        const combinedDate = new Date(
          newYear,
          newMonth,
          newDay,
          originalHours,
          originalMinutes,
          originalSeconds,
        )

        // Convert to UTC for database
        const utcDate = combinedDate.toISOString().slice(0, 19).replace('T', ' ')

        // Update this message (only for users, not clients)
        if (isClientMode.value) {
          alert('Clients cannot edit message dates')
          return
        }
        const result = await callApi({
          query: `UPDATE chat_messages SET time = ? WHERE id = ?`,
          params: [utcDate, message.id],
          requiresAuth: true,
        })

        if (result.success) {
          // Update in local state
          message.time = combinedDate.toISOString()
          // Update in group cache as well
          if (messagesByGroup.value[props.groupId]) {
            const groupMessage = messagesByGroup.value[props.groupId].find(
              (m) => m.id === message.id,
            )
            if (groupMessage) {
              groupMessage.time = combinedDate.toISOString()
            }
          }
          successCount++
        } else {
          failCount++
        }
      } catch (err) {
        console.error(`Error updating message ${message.id}:`, err)
        failCount++
      }
    }

    // Clear selection and close modal
    selectedMessages.value.clear()
    selectionMode.value = false
    closeBulkDateEditor()

    if (failCount === 0) {
      alert(`Successfully updated ${successCount} message(s)`)
    } else {
      alert(
        `Updated ${successCount} message(s), ${failCount} failed. Original times were preserved.`,
      )
    }
  } catch (err) {
    console.error('Error updating bulk message dates:', err)
    alert('Error updating messages: ' + err.message)
  }
}

// Open date picker for editing message date
const openDateEditor = (message) => {
  if (!canEditMessageDate(message)) return
  editingMessageId.value = message.id
  // Convert UTC time from database to local time for datetime-local input
  const messageDate = new Date(message.time)
  // Format as YYYY-MM-DDTHH:mm for datetime-local input (in local timezone)
  const year = messageDate.getFullYear()
  const month = String(messageDate.getMonth() + 1).padStart(2, '0')
  const day = String(messageDate.getDate()).padStart(2, '0')
  const hours = String(messageDate.getHours()).padStart(2, '0')
  const minutes = String(messageDate.getMinutes()).padStart(2, '0')
  editingMessageDate.value = `${year}-${month}-${day}T${hours}:${minutes}`
  showDatePicker.value = true
}

// Close date picker
const closeDateEditor = () => {
  editingMessageId.value = null
  editingMessageDate.value = null
  showDatePicker.value = false
}

// Update message date in database
const updateMessageDate = async () => {
  if (!editingMessageId.value || !editingMessageDate.value) return

  try {
    // Convert local datetime to UTC for database
    const localDate = new Date(editingMessageDate.value)
    const utcDate = localDate.toISOString().slice(0, 19).replace('T', ' ')

    const result = await callApi({
      query: `
        UPDATE chat_messages 
        SET time = ?
        WHERE id = ?
      `,
      params: [utcDate, editingMessageId.value],
      requiresAuth: !isClientMode.value,
    })

    if (result.success) {
      // Update the message in local state
      const messageIndex = messages.value.findIndex((m) => m.id === editingMessageId.value)
      if (messageIndex !== -1) {
        // Update the message time
        messages.value[messageIndex].time = new Date(editingMessageDate.value).toISOString()
        // Update in group cache as well
        if (messagesByGroup.value[props.groupId]) {
          const groupMessageIndex = messagesByGroup.value[props.groupId].findIndex(
            (m) => m.id === editingMessageId.value,
          )
          if (groupMessageIndex !== -1) {
            messagesByGroup.value[props.groupId][groupMessageIndex].time = new Date(
              editingMessageDate.value,
            ).toISOString()
          }
        }
      }
      closeDateEditor()
    } else {
      alert('Failed to update message date: ' + (result.error || 'Unknown error'))
    }
  } catch (err) {
    console.error('Error updating message date:', err)
    alert('Error updating message date: ' + err.message)
  }
}

const handleKeyPress = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

// Auto-refresh messages every 10 seconds
let refreshInterval = null

// User activity tracking for adaptive polling
const isUserActive = ref(true)
const lastActivityTime = ref(Date.now())
const inactivityTimeout = 5 * 60 * 1000 // 5 minutes of inactivity
let inactivityTimer = null

// Adaptive polling intervals
const ACTIVE_INTERVALS = {
  messages: 10 * 1000, // 10 seconds when active
}

const INACTIVE_INTERVALS = {
  messages: 30 * 1000, // 30 seconds when inactive
}

// User activity tracking functions
const updateUserActivity = () => {
  lastActivityTime.value = Date.now()
  if (!isUserActive.value) {
    isUserActive.value = true
    restartMessagePolling()
  }
  resetInactivityTimer()
}

const resetInactivityTimer = () => {
  if (inactivityTimer) {
    clearTimeout(inactivityTimer)
  }
  inactivityTimer = setTimeout(() => {
    isUserActive.value = false
    restartMessagePolling()
  }, inactivityTimeout)
}

const restartMessagePolling = () => {
  // Clear existing interval
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }

  // Get current interval based on activity
  const currentInterval = isUserActive.value
    ? ACTIVE_INTERVALS.messages
    : INACTIVE_INTERVALS.messages

  // Restart with new interval
  refreshInterval = setInterval(async () => {
    if (props.groupId && props.groupId > 0) {
      if (isClientMode.value && currentClient.value) {
        try {
          await fetchNewMessages(props.groupId)
        } catch (err) {
          console.error(`Timer: Error fetching new messages for group ${props.groupId}:`, err)
        }
      } else if (!isClientMode.value && currentUser.value) {
        try {
          await fetchNewMessages(props.groupId)
        } catch (err) {
          console.error(`Timer: Error fetching new messages for group ${props.groupId}:`, err)
        }
      }
    }
  }, currentInterval)
}

// Set up activity listeners
const setupActivityListeners = () => {
  const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click']
  events.forEach((event) => {
    document.addEventListener(event, updateUserActivity, { passive: true })
  })

  // Also track visibility changes
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
      updateUserActivity()
    }
  })
}

// Clean up activity listeners
const cleanupActivityListeners = () => {
  const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click']
  events.forEach((event) => {
    document.removeEventListener(event, updateUserActivity)
  })
  document.removeEventListener('visibilitychange', updateUserActivity)
}

onMounted(async () => {
  // Fetch current user and initial messages
  getCurrentUser()

  // If we have a groupId, check membership and load the first page of messages
  if (props.groupId && props.groupId > 0) {
    await checkMembership()
    await fetchMessages(props.groupId)
  }

  // Start polling for new messages (only for the current group)
  // Set up activity listeners
  setupActivityListeners()

  // Initialize adaptive polling with active intervals (user just loaded the page)
  const currentInterval = ACTIVE_INTERVALS.messages
  refreshInterval = setInterval(async () => {
    if (props.groupId && props.groupId > 0 && currentUser.value) {
      try {
        await fetchNewMessages(props.groupId)
      } catch (err) {
        console.error(`Timer: Error fetching new messages for group ${props.groupId}:`, err)
      }
    }
  }, currentInterval)
})

// Clean up interval on unmount
const cleanup = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
  if (inactivityTimer) {
    clearTimeout(inactivityTimer)
    inactivityTimer = null
  }
  cleanupActivityListeners()
}

// Function to reset new message count for a specific group
const resetNewMessageCount = async (groupId) => {
  // Removed: console.log(`Resetting new message count for group ${groupId}`)
  // Removed: console.log(`Before reset - newMessagesCountByGroup for group ${groupId}:`, newMessagesCountByGroup.value[groupId])

  // Get the highest message ID for this group before resetting
  const currentMessages = messagesByGroup.value[groupId] || []
  const lastReadMessageId =
    currentMessages.length > 0 ? Math.max(...currentMessages.map((msg) => msg.id)) : 0

  // Removed: console.log(`Last read message ID for group ${groupId}: ${lastReadMessageId}`)

  // Save the last read message ID to database
  if (currentUser.value && lastReadMessageId > 0) {
    try {
      const result = await callApi({
        query: `
          INSERT INTO chat_last_read_message (id_group, id_user, id_last_read_message) 
          VALUES (?, ?, ?) 
          ON DUPLICATE KEY UPDATE id_last_read_message = VALUES(id_last_read_message)
        `,
        params: [groupId, currentUser.value.id, lastReadMessageId],
        requiresAuth: true,
      })

      if (result.success) {
        // Removed: console.log(`Saved last read message ID ${lastReadMessageId} for group ${groupId}`)
      } else {
        console.error(`Failed to save last read message ID for group ${groupId}`)
      }
    } catch (err) {
      console.error('Error saving last read message ID:', err)
    }
  }

  // Reset the count - FORCE IT TO 0
  newMessagesCountByGroup.value[groupId] = 0
  // Removed: console.log(`After reset - newMessagesCountByGroup for group ${groupId}:`, newMessagesCountByGroup.value[groupId])

  // Also reset current count if this is the active group
  if (groupId === props.groupId) {
    currentNewMessagesCount.value = 0
    // Removed: console.log(`Reset currentNewMessagesCount to 0 for active group ${groupId}`)
  }

  // Force reactive update by triggering a change
  newMessagesCountByGroup.value = { ...newMessagesCountByGroup.value }
  // Removed: console.log(`Final newMessagesCountByGroup after force update:`, newMessagesCountByGroup.value)

  // Update debug info if debug section is open and this is the current group
  if (showDebugSection.value && groupId === props.groupId) {
    updateDebugInfo()
  }

  // Removed: console.log(`✓ Reset completed for group ${groupId}`)

  // Emit reset event to trigger parent update
  emit('reset-triggered', groupId)

  // Also force an immediate update of the hidden component's counts
  // by calling fetchAllGroupsMessages to recalculate unread counts
  if (props.groupId === 0) {
    // This is the hidden component, so we need to recalculate all counts
    // Removed: console.log('Hidden component: Recalculating all unread counts after reset')
    await fetchAllGroupsMessages()
  }
}

// Function to fetch messages for all groups at once
const fetchAllGroupsMessages = async () => {
  try {
    // Removed: console.log('Fetching messages for all groups...')

    // Skip fetchAllGroupsMessages in client mode (only for users)
    if (isClientMode.value) {
      return
    }

    if (!currentUser.value) {
      // Removed: console.log('No current user available')
      return
    }

    // Get all groups this user is in
    const groupsResult = await callApi({
      query: `
        SELECT DISTINCT
          cg.id, 
          cg.name, 
          cg.description, 
          cg.id_user_owner, 
          cg.is_active 
        FROM chat_groups cg
        INNER JOIN chat_users cu ON cg.id = cu.id_chat_group
        WHERE cg.is_active = 1 
          AND cu.is_active = 1
          AND cu.id_user = ?
        ORDER BY cg.name ASC
      `,
      params: [currentUser.value.id],
      requiresAuth: true,
    })

    if (!groupsResult.success || !groupsResult.data) {
      // Removed: console.log('No groups found or API failed:', groupsResult)
      return
    }

    // Removed: console.log(`Found ${groupsResult.data.length} groups for user`)

    // Fetch messages and count unread messages for each group
    for (const group of groupsResult.data) {
      // Removed: console.log(`Fetching messages for group: ${group.name} (ID: ${group.id})`)

      // Get the last read message ID for this group and user
      const lastReadResult = await callApi({
        query: `
          SELECT id_last_read_message 
          FROM chat_last_read_message 
          WHERE id_group = ? AND id_user = ?
        `,
        params: [group.id, currentUser.value.id],
        requiresAuth: true,
      })

      const lastReadMessageId =
        lastReadResult.success && lastReadResult.data && lastReadResult.data.length > 0
          ? lastReadResult.data[0].id_last_read_message
          : 0

      // Removed: console.log(`Last read message ID for group ${group.name}: ${lastReadMessageId}`)

      // Count unread messages (messages with ID > last_read_message_id and not from current user)
      // Include both user messages (from other users) and client messages
      const unreadCountResult = await callApi({
        query: `
          SELECT COUNT(*) as unread_count
          FROM chat_messages 
          WHERE id_chat_group = ? 
            AND id > ?
            AND (message_from_user_id IS NULL OR message_from_user_id != ?)
        `,
        params: [group.id, lastReadMessageId, currentUser.value.id],
        requiresAuth: true,
      })

      const unreadCount =
        unreadCountResult.success && unreadCountResult.data && unreadCountResult.data.length > 0
          ? parseInt(unreadCountResult.data[0].unread_count)
          : 0

      // Removed: console.log(`Unread messages count for group ${group.name}: ${unreadCount}`)

      // Fetch all messages for this group
      const messagesResult = await callApi({
        query: `
          SELECT 
            cm.id,
            cm.id_chat_group,
            cm.message_from_user_id,
            cm.message_from_client_id,
            cm.chat_replay_to_message_id,
            cm.message,
            cm.time,
            COALESCE(u.username, c.name) as sender_username,
            COALESCE(r.role_name, 'Client') as sender_role,
            CASE WHEN cm.message_from_user_id IS NOT NULL THEN 'user' ELSE 'client' END as sender_type
          FROM chat_messages cm
          LEFT JOIN users u ON cm.message_from_user_id = u.id
          LEFT JOIN roles r ON u.role_id = r.id
          LEFT JOIN clients c ON cm.message_from_client_id = c.id
          WHERE cm.id_chat_group = ?
          ORDER BY cm.id ASC
        `,
        params: [group.id],
        requiresAuth: true,
      })

      if (messagesResult.success && messagesResult.data) {
        // Store messages in group cache
        messagesByGroup.value[group.id] = messagesResult.data
        // Removed: console.log(`Loaded ${messagesResult.data.length} messages for group ${group.name}`)

        // Set unread count for this group
        newMessagesCountByGroup.value[group.id] = unreadCount
        // Removed: console.log(`Set unread count for group ${group.name}: ${unreadCount}`)
      } else {
        // Removed: console.log(`No messages found for group ${group.name}`)
        messagesByGroup.value[group.id] = []
        newMessagesCountByGroup.value[group.id] = 0
      }
    }

    // Removed: console.log('All groups messages loaded successfully')
    // Removed: console.log('Groups loaded:', Object.keys(messagesByGroup.value))
    // Removed: console.log('Total groups with messages:', Object.keys(messagesByGroup.value).length)
    // Removed: console.log('Unread counts by group:', newMessagesCountByGroup.value)

    // If there's a selected group, set its messages as current
    if (props.groupId && messagesByGroup.value[props.groupId]) {
      messages.value = messagesByGroup.value[props.groupId]
      // Removed: console.log(`Set current messages for group ${props.groupId}: ${messages.value.length} messages`)
    }
  } catch (err) {
    console.error('Error fetching all groups messages:', err)
  }
}

// Click handlers to reset new message count
const handleMessagesContainerClick = async () => {
  // Removed: console.log(`Chat area clicked for group ${props.groupId}`)
  // Removed: console.log(`Current newMessagesCountByGroup for this group:`, newMessagesCountByGroup.value[props.groupId])
  // Removed: console.log(`All newMessagesCountByGroup:`, newMessagesCountByGroup.value)

  // Always reset the count when chat area is clicked
  // Removed: console.log(`Reset new message count for group ${props.groupId} (chat area clicked)`)
  await resetNewMessageCount(props.groupId)
}

const handleInputClick = async () => {
  // Removed: console.log(`Input clicked for group ${props.groupId}`)
  // Removed: console.log(`Current newMessagesCountByGroup for this group:`, newMessagesCountByGroup.value[props.groupId])
  // Removed: console.log(`All newMessagesCountByGroup:`, newMessagesCountByGroup.value)

  // Always reset the count when input is clicked
  // Removed: console.log(`Reset new message count for group ${props.groupId} (input clicked)`)
  await resetNewMessageCount(props.groupId)
}

// Debug functions
const toggleDebugSection = () => {
  showDebugSection.value = !showDebugSection.value
  if (showDebugSection.value) {
    // Add some test debug info
    addDebugInfo('debugOpened', {
      message: 'Debug section opened',
      timestamp: new Date().toISOString(),
      groupId: props.groupId,
      user: currentUser.value,
    })
    updateDebugInfo()
  }
}

// Function to add debug info
const addDebugInfo = (type, data) => {
  const timestamp = new Date().toLocaleTimeString()
  debugInfo.value.apiResponses.push({
    timestamp,
    type,
    data: JSON.stringify(data, null, 2),
  })

  // Keep only last 10 responses
  if (debugInfo.value.apiResponses.length > 10) {
    debugInfo.value.apiResponses.shift()
  }
}

// Function to add error
const addError = (error) => {
  const timestamp = new Date().toLocaleTimeString()
  debugInfo.value.errors.push({
    timestamp,
    error: error.toString(),
  })

  // Keep only last 5 errors
  if (debugInfo.value.errors.length > 5) {
    debugInfo.value.errors.shift()
  }
}

// Function to update debug info
const updateDebugInfo = () => {
  debugInfo.value.user = currentUser.value
  debugInfo.value.groupId = props.groupId
  debugInfo.value.messagesCount = messages.value.length
  debugInfo.value.lastFetch = new Date().toLocaleTimeString()
}

const testNewMessagesCount = () => {
  // Removed: console.log('=== TESTING NEW MESSAGES COUNT ===')
  // Removed: console.log('Current newMessagesCount:', currentNewMessagesCount.value)
  // Removed: console.log('All newMessagesCountByGroup:', newMessagesCountByGroup.value)
  updateDebugInfo()
}

// Expose cleanup function and new message counts
defineExpose({
  cleanup,
  fetchAllGroupsMessages,
  getNewMessagesCount: (groupId) => {
    const count = newMessagesCountByGroup.value[groupId] || 0
    // Removed: console.log(`Exposed getNewMessagesCount for group ${groupId}:`, count)
    return count
  },
  getAllNewMessagesCounts: () => {
    // Removed: console.log('Exposed getAllNewMessagesCounts:', newMessagesCountByGroup.value)
    return newMessagesCountByGroup.value
  },
  resetGroupCount: (groupId) => {
    // Removed: console.log(`Exposed resetGroupCount called for group ${groupId}`)
    newMessagesCountByGroup.value[groupId] = 0
    // Force reactive update
    newMessagesCountByGroup.value = { ...newMessagesCountByGroup.value }
    // Removed: console.log(`Exposed resetGroupCount completed for group ${groupId}`)
  },
  focusMessageInput: () => {
    // Removed: console.log('ChatMessages: focusMessageInput called')
    if (messageInputRef.value) {
      messageInputRef.value.focus()
    } else {
      // Removed: console.log('Message input ref not available')
    }
  },
  scrollToBottom: async () => {
    // Removed: console.log('ChatMessages: scrollToBottom called')
    await scrollToBottom()
  },
})

const toggleEmojiPicker = () => {
  showEmojiPicker.value = !showEmojiPicker.value
}

const addEmoji = (emoji) => {
  newMessage.value += emoji
  // Close the emoji picker after adding emoji
  showEmojiPicker.value = false
  // Focus back to the input after adding emoji
  nextTick(() => {
    if (messageInputRef.value) {
      messageInputRef.value.focus()
    }
  })
}

const closeEmojiPicker = () => {
  showEmojiPicker.value = false
}

// File upload functions
const selectFile = () => {
  fileInputRef.value?.click()
}

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // Reset error state
  fileError.value = ''

  // Check file size
  if (file.size > MAX_FILE_SIZE) {
    fileError.value = `File size (${(file.size / 1024 / 1024).toFixed(2)}MB) exceeds the 100MB limit`
    return
  }

  selectedFile.value = file
  uploadFileToChat()
}

const uploadFileToChat = async () => {
  if (!selectedFile.value) return

  // Check if user/client is still an active member before uploading
  await checkMembership()
  
  if (!isActiveMember.value) {
    alert(t('chat.youHaveLeftThisGroup') || 'You have left this group and cannot send messages.')
    selectedFile.value = null
    fileInputRef.value.value = ''
    return
  }

  isUploading.value = true
  uploadProgress.value = 0
  fileError.value = ''

  try {
    // Create folder name for this chat group
    const chatFolder = `chat_files/group_${props.groupId}`

    // Generate unique filename with timestamp
    const timestamp = Date.now()
    const fileExtension = selectedFile.value.name.split('.').pop()
    const customFilename = `${timestamp}_${selectedFile.value.name}`

    // Removed: console.log('Uploading file to chat:', { fileName: selectedFile.value.name, fileSize: selectedFile.value.size, chatFolder, customFilename })

    // Upload the file
    const uploadResult = await uploadFile(selectedFile.value, chatFolder, customFilename)

    if (uploadResult.success) {
      // Create a file message with special prefix
      let query, params, requiresAuth
      
      if (isClientMode.value) {
        query = `
          INSERT INTO chat_messages (id_chat_group, message_from_client_id, message, time) 
          VALUES (?, ?, ?, UTC_TIMESTAMP())
        `
        params = [props.groupId, currentClient.value.id, `[FILE]${uploadResult.relativePath}|${selectedFile.value.name}|${selectedFile.value.size}|${selectedFile.value.type}`]
        requiresAuth = false
      } else {
        query = `
          INSERT INTO chat_messages (id_chat_group, message_from_user_id, message, time) 
          VALUES (?, ?, ?, UTC_TIMESTAMP())
        `
        params = [props.groupId, currentUser.value.id, `[FILE]${uploadResult.relativePath}|${selectedFile.value.name}|${selectedFile.value.size}|${selectedFile.value.type}`]
        requiresAuth = true
      }

      // Send the file message to the database
      const result = await callApi({
        query,
        params,
        requiresAuth,
      })

      if (result.success) {
        // Add the new message to both current list and group cache
        let newMsg
        if (isClientMode.value) {
          newMsg = {
            id: result.lastInsertId,
            id_chat_group: props.groupId,
            message_from_client_id: currentClient.value.id,
            message: `[FILE]${uploadResult.relativePath}|${selectedFile.value.name}|${selectedFile.value.size}|${selectedFile.value.type}`,
            time: new Date().toISOString(),
            sender_username: currentClient.value.name,
            sender_role: 'Client',
            sender_type: 'client',
          }
        } else {
          newMsg = {
            id: result.lastInsertId,
            id_chat_group: props.groupId,
            message_from_user_id: currentUser.value.id,
            message: `[FILE]${uploadResult.relativePath}|${selectedFile.value.name}|${selectedFile.value.size}|${selectedFile.value.type}`,
            time: new Date().toISOString(),
            sender_username: currentUser.value.username,
            sender_role: currentUser.value.role_name,
            sender_type: 'user',
          }
        }

        messages.value.push(newMsg)
        messagesByGroup.value[props.groupId] = [...messages.value]

        // Emit event to parent
        emit('message-sent', newMsg)

        // Dispatch global event to update header badge
        window.dispatchEvent(new Event('forceUpdateBadge'))

        // Scroll to bottom after message is added
        await nextTick()
        await scrollToBottom()

        // Removed: console.log('File message sent successfully:', newMsg)
      } else {
        throw new Error('Failed to save file message to database')
      }
    } else {
      throw new Error(uploadResult.message || 'File upload failed')
    }
  } catch (err) {
    console.error('Error uploading file:', err)
    fileError.value = err.message || 'Failed to upload file'
  } finally {
    isUploading.value = false
    uploadProgress.value = 0
    selectedFile.value = null
    // Clear the file input
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
  }
}

const downloadFile = async (filePath, fileName) => {
  try {
    const fileUrl = getFileUrl(filePath)
    const response = await fetch(fileUrl)

    if (!response.ok) {
      throw new Error('Failed to download file')
    }

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = fileName
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (err) {
    console.error('Error downloading file:', err)
    alert('Failed to download file')
  }
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getFileIcon = (fileType) => {
  if (fileType.startsWith('image/')) return '🖼️'
  if (fileType.startsWith('video/')) return '🎥'
  if (fileType.startsWith('audio/')) return '🎵'
  if (fileType.includes('pdf')) return '📄'
  if (fileType.includes('word') || fileType.includes('document')) return '📝'
  if (fileType.includes('excel') || fileType.includes('spreadsheet')) return '📊'
  if (fileType.includes('powerpoint') || fileType.includes('presentation')) return '📽️'
  if (fileType.includes('zip') || fileType.includes('rar') || fileType.includes('archive'))
    return '📦'
  return '📎'
}

// Parse file message to extract file information
const parseFileMessage = (message) => {
  if (!message.startsWith('[FILE]')) return null

  const fileData = message.substring(6) // Remove '[FILE]' prefix
  const parts = fileData.split('|')

  if (parts.length >= 4) {
    return {
      filePath: parts[0],
      fileName: parts[1],
      fileSize: parseInt(parts[2]),
      fileType: parts[3],
    }
  }

  return null
}

// Check if message is a file message
const isFileMessage = (message) => {
  return message.startsWith('[FILE]')
}

// Check if message is a voice message
const isVoiceMessage = (message) => {
  return message.startsWith('[VOICE]')
}

// Parse voice message to extract voice information
const parseVoiceMessage = (message) => {
  if (!message.startsWith('[VOICE]')) return null

  const voiceData = message.substring(7) // Remove '[VOICE]' prefix
  const parts = voiceData.split('|')

  if (parts.length >= 5) {
    return {
      filePath: parts[0],
      fileName: parts[1],
      fileSize: parseInt(parts[2]),
      fileType: parts[3],
      duration: parseInt(parts[4]),
    }
  }

  return null
}

// Check if file is an image
const isImageFile = (fileType) => {
  return fileType.startsWith('image/')
}

// Check if file is a video
const isVideoFile = (fileType) => {
  return fileType.startsWith('video/')
}

// Check if file is an audio file
const isAudioFile = (fileType) => {
  return fileType.startsWith('audio/')
}

// Get file URL for display
const getFileDisplayUrl = (filePath) => {
  return getFileUrl(filePath)
}

// Voice recording functions
const startRecording = async () => {
  try {
    // Check if we're on HTTPS (required for microphone access)
    if (window.location.protocol !== 'https:' && window.location.hostname !== 'localhost') {
      throw new Error('Microphone access requires HTTPS. Please use https://cars.merhab.com')
    }

    // Check if MediaRecorder is supported
    if (!window.MediaRecorder) {
      throw new Error('MediaRecorder is not supported in this browser')
    }

    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    mediaRecorder.value = new MediaRecorder(stream)
    audioChunks.value = []

    mediaRecorder.value.ondataavailable = (event) => {
      audioChunks.value.push(event.data)
    }

    mediaRecorder.value.onstop = () => {
      audioBlob.value = new Blob(audioChunks.value, { type: 'audio/wav' })
      audioUrl.value = URL.createObjectURL(audioBlob.value)

      // Stop all tracks to release microphone
      stream.getTracks().forEach((track) => track.stop())
    }

    mediaRecorder.value.start()
    isRecording.value = true
    recordingTime.value = 0

    // Start timer
    recordingInterval.value = setInterval(() => {
      recordingTime.value++
    }, 1000)
  } catch (err) {
    console.error('Error starting recording:', err)

    let errorMessage = 'Could not access microphone. '

    if (err.name === 'NotAllowedError') {
      errorMessage += 'Please allow microphone access in your browser settings.'
    } else if (err.name === 'NotFoundError') {
      errorMessage += 'No microphone found. Please connect a microphone and try again.'
    } else if (err.name === 'NotSupportedError') {
      errorMessage += 'Microphone access is not supported in this browser.'
    } else if (err.message.includes('HTTPS')) {
      errorMessage = err.message
    } else {
      errorMessage += 'Please check permissions and try again.'
    }

    alert(errorMessage)
  }
}

const stopRecording = () => {
  if (mediaRecorder.value && isRecording.value) {
    mediaRecorder.value.stop()
    isRecording.value = false

    if (recordingInterval.value) {
      clearInterval(recordingInterval.value)
      recordingInterval.value = null
    }
  }
}

const sendVoiceMessage = async () => {
  if (!audioBlob.value) return

  // Check if user/client is still an active member before sending voice message
  await checkMembership()
  
  if (!isActiveMember.value) {
    alert(t('chat.youHaveLeftThisGroup') || 'You have left this group and cannot send messages.')
    cancelRecording()
    return
  }

  isUploading.value = true
  fileError.value = ''

  try {
    // Create a file from the audio blob
    const audioFile = new File([audioBlob.value], `voice_message_${Date.now()}.wav`, {
      type: 'audio/wav',
    })

    // Create folder name for this chat group
    const chatFolder = `chat_files/group_${props.groupId}`

    // Generate unique filename with timestamp
    const timestamp = Date.now()
    const customFilename = `${timestamp}_voice_message.wav`

    // Removed: console.log('Uploading voice message:', { fileName: audioFile.name, fileSize: audioFile.size, chatFolder, customFilename })

    // Upload the voice message
    const uploadResult = await uploadFile(audioFile, chatFolder, customFilename)

    if (uploadResult.success) {
      // Create a voice message with special prefix
      let query, params, requiresAuth
      
      if (isClientMode.value) {
        const voiceMessage = {
          id_chat_group: props.groupId,
          message_from_client_id: currentClient.value.id,
          message: `[VOICE]${uploadResult.relativePath}|voice_message.wav|${audioFile.size}|audio/wav|${recordingTime.value}`,
        }
        query = `
          INSERT INTO chat_messages (id_chat_group, message_from_client_id, message, time) 
          VALUES (?, ?, ?, UTC_TIMESTAMP())
        `
        params = [
          voiceMessage.id_chat_group,
          voiceMessage.message_from_client_id,
          voiceMessage.message,
        ]
        requiresAuth = false
      } else {
        const voiceMessage = {
          id_chat_group: props.groupId,
          message_from_user_id: currentUser.value.id,
          message: `[VOICE]${uploadResult.relativePath}|voice_message.wav|${audioFile.size}|audio/wav|${recordingTime.value}`,
        }
        query = `
          INSERT INTO chat_messages (id_chat_group, message_from_user_id, message, time) 
          VALUES (?, ?, ?, UTC_TIMESTAMP())
        `
        params = [
          voiceMessage.id_chat_group,
          voiceMessage.message_from_user_id,
          voiceMessage.message,
        ]
        requiresAuth = true
      }

      // Send the voice message to the database
      const result = await callApi({
        query,
        params,
        requiresAuth,
      })

      if (result.success) {
        // Add the new message to both current list and group cache
        let newMsg
        const voiceMessageText = `[VOICE]${uploadResult.relativePath}|voice_message.wav|${audioFile.size}|audio/wav|${recordingTime.value}`
        
        if (isClientMode.value) {
          newMsg = {
            id: result.lastInsertId,
            id_chat_group: props.groupId,
            message_from_client_id: currentClient.value.id,
            message: voiceMessageText,
            time: new Date().toISOString(),
            sender_username: currentClient.value.name,
            sender_role: 'Client',
            sender_type: 'client',
          }
        } else {
          newMsg = {
            id: result.lastInsertId,
            id_chat_group: props.groupId,
            message_from_user_id: currentUser.value.id,
            message: voiceMessageText,
            time: new Date().toISOString(),
            sender_username: currentUser.value.username,
            sender_role: currentUser.value.role_name,
            sender_type: 'user',
          }
        }

        messages.value.push(newMsg)
        messagesByGroup.value[props.groupId] = [...messages.value]

        // Emit event to parent
        emit('message-sent', newMsg)

        // Dispatch global event to update header badge
        window.dispatchEvent(new Event('forceUpdateBadge'))

        // Scroll to bottom after message is added
        await nextTick()
        await scrollToBottom()

        // Removed: console.log('Voice message sent successfully:', newMsg)
      } else {
        throw new Error('Failed to save voice message to database')
      }
    } else {
      throw new Error(uploadResult.message || 'Voice message upload failed')
    }
  } catch (err) {
    console.error('Error sending voice message:', err)
    fileError.value = err.message || t('chat.failedToSendVoiceMessage')
  } finally {
    isUploading.value = false
    // Clean up
    audioBlob.value = null
    audioUrl.value = ''
    recordingTime.value = 0
  }
}

const cancelRecording = () => {
  stopRecording()
  audioBlob.value = null
  audioUrl.value = ''
  recordingTime.value = 0
}

const formatRecordingTime = (seconds) => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

// Load more messages function
const loadMoreMessages = async () => {
  if (isLoadingMore.value || !hasMoreMessages.value) return

  isLoadingMore.value = true

  try {
    currentPage.value++
    await fetchMessages(props.groupId)

    // Scroll to maintain position after loading more messages
    await nextTick()
    if (messagesContainer.value) {
      const oldHeight = messagesContainer.value.scrollHeight
      setTimeout(() => {
        const newHeight = messagesContainer.value.scrollHeight
        const heightDifference = newHeight - oldHeight
        messagesContainer.value.scrollTo({
          top: heightDifference,
          behavior: 'smooth',
        })
      }, 100)
    }
  } catch (err) {
    console.error('Error loading more messages:', err)
    currentPage.value-- // Revert page increment on error
  } finally {
    isLoadingMore.value = false
  }
}

// Fetch only new messages for polling (doesn't affect pagination)
const fetchNewMessages = async (groupId) => {
  try {
    if (!currentUser.value) return

    // Get the highest message ID from current messages
    const currentMessages = messages.value
    const currentHighestId =
      currentMessages.length > 0 ? Math.max(...currentMessages.map((msg) => msg.id)) : 0

    const result = await callApi({
      query: `
        SELECT 
          cm.id,
          cm.id_chat_group,
          cm.message_from_user_id,
          cm.message_from_client_id,
          cm.message,
          cm.time,
          COALESCE(u.username, c.name) as sender_username,
          COALESCE(r.role_name, 'Client') as sender_role,
          CASE WHEN cm.message_from_user_id IS NOT NULL THEN 'user' ELSE 'client' END as sender_type
        FROM chat_messages cm
        LEFT JOIN users u ON cm.message_from_user_id = u.id
        LEFT JOIN roles r ON u.role_id = r.id
        LEFT JOIN clients c ON cm.message_from_client_id = c.id
        WHERE cm.id_chat_group = CAST(? AS UNSIGNED)
          AND cm.id > CAST(? AS UNSIGNED)
        ORDER BY cm.id ASC
      `,
      params: [groupId, currentHighestId],
      requiresAuth: !isClientMode.value,
    })

    if (result.success && result.data && result.data.length > 0) {
      // Add new messages to the end
      messages.value.push(...result.data)
      messagesByGroup.value[groupId] = [...messages.value]

      // Scroll to bottom for new messages
      await nextTick()
      await scrollToBottom()

      addDebugInfo('newMessagesAdded', { count: result.data.length, groupId })
    }
  } catch (err) {
    addError(`Error fetching new messages: ${err.message}`)
  }
}

// Watch for changes in selected group
watch(
  () => props.groupId,
  async (newGroupId) => {
    addDebugInfo('groupChanged', { newGroupId })

    if (newGroupId && newGroupId > 0) {
      currentPage.value = 1
      hasMoreMessages.value = true
      // Check membership when group changes
      await checkMembership()
      await fetchMessages(newGroupId)
      // Clear search when switching groups
      messageSearch.value = ''
      // Scroll to bottom for new group
      await nextTick()
      await scrollToBottom()
    } else {
      addDebugInfo('groupCleared', { reason: 'No group selected or invalid group ID' })
      messages.value = []
      hasMoreMessages.value = false
      messageSearch.value = ''
      isActiveMember.value = false
    }
  },
  { immediate: true },
)

// Filtered messages based on search
const filteredMessages = computed(() => {
  if (!messageSearch.value.trim()) {
    return messages.value
  }

  const searchTerm = messageSearch.value.trim().toLowerCase()
  
  return messages.value.filter((message) => {
    // Search in message content
    const messageContent = message.message?.toLowerCase() || ''
    
    // Search in sender username
    const senderName = message.sender_username?.toLowerCase() || ''
    
    // Search in sender role
    const senderRole = message.sender_role?.toLowerCase() || ''
    
    return (
      messageContent.includes(searchTerm) ||
      senderName.includes(searchTerm) ||
      senderRole.includes(searchTerm)
    )
  })
})

// Clear search
const clearMessageSearch = () => {
  messageSearch.value = ''
}
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
        <button
          v-if="isAdmin"
          @click="toggleSelectionMode"
          :class="['selection-mode-btn', { active: selectionMode }]"
          :title="selectionMode ? t('chat.exitSelectionMode') : t('chat.enterSelectionMode')"
        >
          <i :class="selectionMode ? 'fas fa-check-square' : 'fas fa-square'"></i>
          {{ selectionMode ? t('chat.selectionMode') : t('chat.selectMessages') }}
        </button>
        <button
          v-if="selectionMode && selectedMessages.size > 0"
          @click="openBulkDateEditor"
          class="bulk-edit-btn"
          :title="`Edit date for ${selectedMessages.size} selected message(s)`"
        >
          <i class="fas fa-calendar-alt"></i>
          {{ t('chat.editDate') }} ({{ selectedMessages.size }})
        </button>
        <button
          v-if="selectionMode && selectedMessages.size > 0"
          @click="deselectAllMessages"
          class="deselect-all-btn"
          :title="t('chat.deselectAll')"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
    </div>

    <!-- Message Search Box -->
    <div class="message-search-container">
      <div class="message-search-wrapper">
        <i class="fas fa-search message-search-icon"></i>
        <input
          type="text"
          v-model="messageSearch"
          :placeholder="t('chat.searchMessages')"
          class="message-search-input"
        />
        <button
          v-if="messageSearch"
          @click="clearMessageSearch"
          class="clear-message-search-btn"
          :title="t('chat.clearSearch')"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div v-if="messageSearch && filteredMessages.length > 0" class="search-results-info">
        {{ filteredMessages.length }} / {{ messages.length }} messages
      </div>
      <div v-if="messageSearch && filteredMessages.length === 0 && messages.length > 0" class="no-search-results">
        No messages found matching "{{ messageSearch }}"
      </div>
    </div>

    <div class="messages-container" ref="messagesContainer" @click="handleMessagesContainerClick">
      <!-- Load More Button -->
      <div v-if="hasMoreMessages" class="load-more-container">
        <button @click="loadMoreMessages" :disabled="isLoadingMore" class="load-more-btn">
          <i :class="isLoadingMore ? 'fas fa-spinner fa-spin' : 'fas fa-chevron-up'"></i>
          {{ isLoadingMore ? 'Loading...' : 'Load More Messages' }}
        </button>
      </div>

      <div v-if="loading && messages.length === 0" class="loading-messages">
        <i class="fas fa-spinner fa-spin"></i> {{ t('chat.loadingMessages') }}
      </div>

      <div v-else-if="error" class="error-messages">
        <i class="fas fa-exclamation-triangle"></i> {{ error }}
      </div>

      <div v-else-if="messages.length === 0" class="no-messages">
        <i class="fas fa-comment-slash"></i>
        <p>{{ t('chat.noMessagesYet') }}</p>
        <p class="sub-text">{{ t('chat.startConversation') }}</p>
      </div>

      <div v-else-if="messageSearch && filteredMessages.length === 0" class="no-search-results-in-list">
        <i class="fas fa-search"></i>
        <p>{{ t('chat.noMessagesFound', { search: messageSearch }) }}</p>
        <button @click="clearMessageSearch" class="clear-search-link-btn">
          Clear search
        </button>
      </div>

      <div v-else class="messages-list">
        <!-- Select All button (shown in selection mode) -->
        <div v-if="selectionMode" class="select-all-container">
          <button @click="selectAllMessages" class="select-all-btn">
            <i class="fas fa-check-double"></i>
            Select All Messages
          </button>
          <span v-if="selectedMessages.size > 0" class="selected-count">
            {{ selectedMessages.size }} selected
          </span>
        </div>
        <div
          v-for="(message, index) in filteredMessages"
          :key="message.id"
          class="message-wrapper"
          :class="{ 'own-message': isOwnMessage(message), 'selected': selectedMessages.has(message.id) }"
        >
          <!-- Date separator -->
          <div
            v-if="index === 0 || formatDate(message.time) !== formatDate(filteredMessages[index - 1].time)"
            class="date-separator"
          >
            {{ formatDate(message.time) }}
          </div>

          <div class="message">
            <!-- Selection checkbox (shown in selection mode) -->
            <div v-if="selectionMode" class="message-checkbox">
              <input
                type="checkbox"
                :checked="selectedMessages.has(message.id)"
                @change="toggleMessageSelection(message.id)"
                class="message-select-checkbox"
              />
            </div>
            <div class="message-header">
              <span class="sender-name">{{ message.sender_username }}</span>
              <div class="time-info">
                <span class="relative-time">{{ getRelativeTime(message.time) }}</span>
                <span class="absolute-time">{{ formatTime(message.time) }}</span>
                <button
                  v-if="canEditMessageDate(message)"
                  @click="openDateEditor(message)"
                  class="edit-date-btn"
                  :title="t('chat.editMessageDate')"
                >
                  <i class="fas fa-edit"></i>
                </button>
              </div>
            </div>

            <div class="message-content">
              <!-- Show voice message if it's a voice message -->
              <div v-if="isVoiceMessage(message.message)" class="voice-message">
                <div class="voice-message-content">
                  <i class="fas fa-microphone voice-icon"></i>
                  <audio
                    :src="getFileDisplayUrl(parseVoiceMessage(message.message)?.filePath)"
                    controls
                    class="voice-player"
                    preload="metadata"
                  >
                    Your browser does not support the audio tag.
                  </audio>
                  <div class="voice-info">
                    <span class="voice-duration">{{
                      formatRecordingTime(parseVoiceMessage(message.message)?.duration)
                    }}</span>
                    <button
                      @click="
                        downloadFile(
                          parseVoiceMessage(message.message)?.filePath,
                          parseVoiceMessage(message.message)?.fileName,
                        )
                      "
                      class="download-btn"
                      :title="t('chat.downloadVoiceMessage')"
                    >
                      <i class="fas fa-download"></i>
                    </button>
                  </div>
                </div>
              </div>

              <!-- Show file attachment if it's a file message -->
              <div v-else-if="isFileMessage(message.message)" class="file-attachment">
                <!-- Image display -->
                <div
                  v-if="isImageFile(parseFileMessage(message.message)?.fileType)"
                  class="image-attachment"
                >
                  <img
                    :src="getFileDisplayUrl(parseFileMessage(message.message)?.filePath)"
                    :alt="parseFileMessage(message.message)?.fileName"
                    class="chat-image"
                    @click="
                      downloadFile(
                        parseFileMessage(message.message)?.filePath,
                        parseFileMessage(message.message)?.fileName,
                      )
                    "
                    :title="t('chat.clickToDownload')"
                  />
                  <div class="file-info-overlay">
                    <span class="file-name">{{ parseFileMessage(message.message)?.fileName }}</span>
                    <span class="file-size">{{
                      formatFileSize(parseFileMessage(message.message)?.fileSize)
                    }}</span>
                  </div>
                </div>

                <!-- Video display -->
                <div
                  v-else-if="isVideoFile(parseFileMessage(message.message)?.fileType)"
                  class="video-attachment"
                >
                  <video
                    :src="getFileDisplayUrl(parseFileMessage(message.message)?.filePath)"
                    controls
                    class="chat-video"
                    preload="metadata"
                  >
                    Your browser does not support the video tag.
                  </video>
                  <div class="file-info-overlay">
                    <span class="file-name">{{ parseFileMessage(message.message)?.fileName }}</span>
                    <span class="file-size">{{
                      formatFileSize(parseFileMessage(message.message)?.fileSize)
                    }}</span>
                    <button
                      @click="
                        downloadFile(
                          parseFileMessage(message.message)?.filePath,
                          parseFileMessage(message.message)?.fileName,
                        )
                      "
                      class="download-btn"
                      :title="t('chat.downloadVideo')"
                    >
                      <i class="fas fa-download"></i>
                    </button>
                  </div>
                </div>

                <!-- Audio display -->
                <div
                  v-else-if="isAudioFile(parseFileMessage(message.message)?.fileType)"
                  class="audio-attachment"
                >
                  <div class="audio-player">
                    <i class="fas fa-music audio-icon"></i>
                    <audio
                      :src="getFileDisplayUrl(parseFileMessage(message.message)?.filePath)"
                      controls
                      class="chat-audio"
                      preload="metadata"
                    >
                      Your browser does not support the audio tag.
                    </audio>
                  </div>
                  <div class="file-info-overlay">
                    <span class="file-name">{{ parseFileMessage(message.message)?.fileName }}</span>
                    <span class="file-size">{{
                      formatFileSize(parseFileMessage(message.message)?.fileSize)
                    }}</span>
                    <button
                      @click="
                        downloadFile(
                          parseFileMessage(message.message)?.filePath,
                          parseFileMessage(message.message)?.fileName,
                        )
                      "
                      class="download-btn"
                      :title="t('chat.downloadAudio')"
                    >
                      <i class="fas fa-download"></i>
                    </button>
                  </div>
                </div>

                <!-- Other file types -->
                <div v-else class="file-info">
                  <span class="file-icon">{{
                    getFileIcon(parseFileMessage(message.message)?.fileType)
                  }}</span>
                  <div class="file-details">
                    <span class="file-name">{{ parseFileMessage(message.message)?.fileName }}</span>
                    <span class="file-size">{{
                      formatFileSize(parseFileMessage(message.message)?.fileSize)
                    }}</span>
                  </div>
                  <button
                    @click="
                      downloadFile(
                        parseFileMessage(message.message)?.filePath,
                        parseFileMessage(message.message)?.fileName,
                      )
                    "
                    class="download-btn"
                    :title="t('chat.downloadFile')"
                  >
                    <i class="fas fa-download"></i>
                  </button>
                </div>
              </div>

              <!-- Show regular text message if it's not a file or voice message -->
              <div v-else>
                {{ message.message }}
              </div>
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
        <button 
          @click="toggleEmojiPicker" 
          class="emoji-button" 
          :title="t('chat.addEmoji')" 
          type="button"
          :disabled="!isActiveMember"
        >
          <i class="fas fa-smile"></i>
        </button>
        <button
          @click="selectFile"
          class="file-button"
          :title="t('chat.attachFile')"
          type="button"
          :disabled="isUploading || !isActiveMember"
        >
          <i :class="isUploading ? 'fas fa-spinner fa-spin' : 'fas fa-paperclip'"></i>
        </button>
        <button
          @click="isRecording ? stopRecording() : startRecording()"
          class="voice-button"
          :class="{ recording: isRecording }"
          :title="t('chat.recordVoiceMessage')"
          type="button"
          :disabled="isUploading || !isActiveMember"
        >
          <i :class="isRecording ? 'fas fa-stop' : 'fas fa-microphone'"></i>
        </button>
        <textarea
          v-model="newMessage"
          @keypress="handleKeyPress"
          @click="handleInputClick"
          :placeholder="isActiveMember ? t('chat.typeYourMessage') : t('chat.youHaveLeftThisGroup')"
          class="message-input"
          :disabled="isSending || !isActiveMember"
          rows="1"
          maxlength="1000"
          ref="messageInputRef"
        ></textarea>
        <button
          @click="
            (e) => {
              e.preventDefault()
              sendMessage()
            }
          "
          :disabled="!newMessage.trim() || isSending || !isActiveMember"
          class="send-button"
          :title="!isActiveMember ? t('chat.youHaveLeftThisGroup') : ''"
        >
          <i :class="isSending ? 'fas fa-spinner fa-spin' : 'fas fa-paper-plane'"></i>
        </button>
      </div>

      <!-- File upload progress and error -->
      <div v-if="isUploading" class="upload-progress">
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: uploadProgress + '%' }"></div>
        </div>
        <span class="progress-text">{{ t('chat.uploadingFile') }}</span>
      </div>

      <!-- Voice recording interface -->
      <div v-if="isRecording" class="voice-recording">
        <div class="recording-indicator">
          <i class="fas fa-microphone"></i>
          <span class="recording-text">{{ t('chat.recording') }}</span>
          <span class="recording-time">{{ formatRecordingTime(recordingTime) }}</span>
        </div>
        <div class="recording-actions">
          <button @click="cancelRecording" class="cancel-btn" :title="t('chat.cancelRecording')">
            <i class="fas fa-times"></i>
          </button>
        </div>
      </div>

      <!-- Voice message preview -->
      <div v-if="audioUrl && !isRecording" class="voice-preview">
        <div class="voice-preview-content">
          <audio :src="audioUrl" controls class="voice-audio"></audio>
          <div class="voice-actions">
            <button @click="sendVoiceMessage" class="send-voice-btn" :title="t('chat.sendVoiceMessage')">
              <i class="fas fa-paper-plane"></i>
            </button>
            <button @click="cancelRecording" class="cancel-voice-btn" :title="t('chat.cancel')">
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>
      </div>

      <div v-if="fileError" class="file-error">
        <i class="fas fa-exclamation-triangle"></i>
        {{ fileError }}
      </div>

      <div class="input-footer">
        <span class="char-count">{{ newMessage.length }}/1000</span>
        <span class="send-hint">{{ t('chat.pressEnterToSend') }}</span>
      </div>

      <!-- Hidden file input -->
      <input
        ref="fileInputRef"
        type="file"
        @change="handleFileSelect"
        style="display: none"
        accept="*/*"
      />
    </div>

    <!-- Emoji Picker Modal -->
    <div v-if="showEmojiPicker" class="emoji-picker-modal" @click="closeEmojiPicker">
      <div class="emoji-picker-content" @click.stop>
        <div class="emoji-picker-header">
          <h4>Select Emoji</h4>
          <button @click="closeEmojiPicker" class="close-emoji-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="emoji-grid">
          <button
            v-for="emoji in emojis"
            :key="emoji"
            @click="addEmoji(emoji)"
            class="emoji-item"
            :title="emoji"
          >
            {{ emoji }}
          </button>
        </div>
      </div>
    </div>

    <!-- Date Picker Modal -->
    <div v-if="showDatePicker" class="date-picker-modal" @click.self="closeDateEditor">
      <div class="date-picker-content">
        <div class="date-picker-header">
          <h3>Edit Message Date</h3>
          <button @click="closeDateEditor" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="date-picker-body">
          <label for="message-date-input">Select Date & Time:</label>
          <input
            id="message-date-input"
            type="datetime-local"
            v-model="editingMessageDate"
            class="date-input"
          />
        </div>
        <div class="date-picker-footer">
          <button @click="closeDateEditor" class="cancel-btn">Cancel</button>
          <button @click="updateMessageDate" class="save-btn">Save</button>
        </div>
      </div>
    </div>

    <!-- Bulk Date Picker Modal -->
    <div v-if="showBulkDatePicker" class="date-picker-modal" @click.self="closeBulkDateEditor">
      <div class="date-picker-content">
        <div class="date-picker-header">
          <h3>Edit Date for {{ selectedMessages.size }} Message(s)</h3>
          <button @click="closeBulkDateEditor" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="date-picker-body">
          <label for="bulk-date-input">Select Date (Time will be preserved):</label>
          <input
            id="bulk-date-input"
            type="date"
            v-model="bulkEditDate"
            class="date-input"
          />
          <p class="bulk-edit-info">
            This will update the date (month and day) for all {{ selectedMessages.size }} selected
            message(s). The original time of each message will be preserved.
          </p>
        </div>
        <div class="date-picker-footer">
          <button @click="closeBulkDateEditor" class="cancel-btn">Cancel</button>
          <button @click="updateBulkMessageDates" class="save-btn">Update All</button>
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
  min-height: 0;
  max-height: 100%;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  position: relative;
}

.messages-header {
  padding: 8px 16px;
  background-color: #06b6d4;
  color: white;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  flex-shrink: 0;
  position: relative;
  z-index: 2;
}

.messages-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.95rem;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.group-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.message-search-container {
  padding: 8px 16px;
  background-color: #f1f5f9;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
  position: relative;
  z-index: 2;
}

.message-search-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.message-search-icon {
  position: absolute;
  left: 12px;
  color: #64748b;
  font-size: 14px;
  pointer-events: none;
}

.message-search-input {
  width: 100%;
  padding: 8px 12px 8px 36px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  color: #374151;
  background-color: white;
  transition: all 0.2s ease;
}

.message-search-input:focus {
  outline: none;
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.message-search-input::placeholder {
  color: #9ca3af;
}

.clear-message-search-btn {
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

.clear-message-search-btn:hover {
  background-color: #e5e7eb;
  color: #374151;
}

.search-results-info {
  margin-top: 8px;
  font-size: 12px;
  color: #64748b;
  text-align: right;
}

.no-search-results {
  margin-top: 8px;
  padding: 8px;
  text-align: center;
  font-size: 12px;
  color: #64748b;
  font-style: italic;
}

.no-search-results-in-list {
  padding: 40px 20px;
  text-align: center;
  color: #64748b;
}

.no-search-results-in-list i {
  font-size: 2rem;
  color: #9ca3af;
  margin-bottom: 12px;
}

.no-search-results-in-list p {
  margin: 0 0 12px 0;
  font-size: 0.9rem;
}

.clear-search-link-btn {
  background: none;
  border: none;
  color: #06b6d4;
  cursor: pointer;
  font-size: 0.9rem;
  text-decoration: underline;
  padding: 4px 8px;
}

.clear-search-link-btn:hover {
  color: #0891b2;
}

.timezone-info {
  font-size: 0.75rem;
  opacity: 0.9;
  display: flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

.member-count {
  font-size: 0.75rem;
  opacity: 0.9;
  display: flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

.messages-container {
  flex: 1 1 auto;
  min-height: 0;
  max-height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 20px;
  background-color: #f8fafc;
  scroll-behavior: smooth;
  position: relative;
  -webkit-overflow-scrolling: touch;
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
  position: relative;
}

.message-wrapper.own-message {
  align-items: flex-end;
}

.message-wrapper.selected .message {
  border: 2px solid #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.2);
}

.select-all-container {
  padding: 12px 16px;
  background-color: #f0f9ff;
  border: 1px solid #06b6d4;
  border-radius: 8px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.select-all-btn {
  background-color: #06b6d4;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: background-color 0.2s;
}

.select-all-btn:hover {
  background-color: #0891b2;
}

.selected-count {
  color: #0369a1;
  font-weight: 500;
  font-size: 0.9rem;
}

.message-checkbox {
  position: absolute;
  left: -30px;
  top: 12px;
  z-index: 10;
}

@media (max-width: 768px) {
  .message-checkbox {
    left: -25px;
  }
}

.message-select-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #06b6d4;
}

.selection-mode-btn {
  background-color: transparent;
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}

.selection-mode-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.selection-mode-btn.active {
  background-color: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.5);
}

.bulk-edit-btn {
  background-color: #10b981;
  border: none;
  color: white;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: background-color 0.2s;
}

.bulk-edit-btn:hover {
  background-color: #059669;
}

.deselect-all-btn {
  background-color: transparent;
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.deselect-all-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
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

.edit-date-btn {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 4px 6px;
  margin-left: 6px;
  opacity: 0.6;
  transition: opacity 0.2s;
  font-size: 0.7rem;
  display: inline-flex;
  align-items: center;
  border-radius: 4px;
}

.edit-date-btn:hover {
  opacity: 1;
  background-color: rgba(0, 0, 0, 0.05);
}

.own-message .edit-date-btn {
  color: rgba(255, 255, 255, 0.8);
}

.own-message .edit-date-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

/* Date Picker Modal */
.date-picker-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.date-picker-content {
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  width: 90%;
  max-width: 400px;
  overflow: hidden;
}

.date-picker-header {
  padding: 16px 20px;
  background-color: #06b6d4;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.date-picker-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.close-btn {
  background: transparent;
  border: none;
  color: white;
  cursor: pointer;
  font-size: 1.2rem;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.close-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

.date-picker-body {
  padding: 20px;
}

.date-picker-body label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.date-input {
  width: 100%;
  padding: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 1rem;
  outline: none;
  transition: border-color 0.2s;
}

.date-input:focus {
  border-color: #06b6d4;
}

.bulk-edit-info {
  margin-top: 12px;
  padding: 8px;
  background-color: #f0f9ff;
  border-left: 3px solid #06b6d4;
  border-radius: 4px;
  font-size: 0.85rem;
  color: #0369a1;
}

.date-picker-footer {
  padding: 16px 20px;
  background-color: #f8fafc;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.cancel-btn,
.save-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background-color 0.2s;
}

.cancel-btn {
  background-color: #e2e8f0;
  color: #374151;
}

.cancel-btn:hover {
  background-color: #cbd5e1;
}

.save-btn {
  background-color: #06b6d4;
  color: white;
}

.save-btn:hover {
  background-color: #0891b2;
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
  flex-shrink: 0;
  flex-grow: 0;
  padding: 12px 16px;
  background-color: #ffffff;
  border-top: 1px solid #e2e8f0;
  z-index: 10;
  width: 100%;
  box-sizing: border-box;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.05);
  position: relative;
  -webkit-transform: translateZ(0);
  transform: translateZ(0);
}

.input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.emoji-button {
  background-color: transparent;
  border: none;
  color: #64748b;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  height: 44px;
}

.emoji-button:hover {
  background-color: #f1f5f9;
  color: #06b6d4;
}

.emoji-button:hover {
  background-color: #f1f5f9;
  color: #06b6d4;
}

.file-button {
  background-color: transparent;
  border: none;
  color: #64748b;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  height: 44px;
}

.file-button:hover {
  background-color: #f1f5f9;
  color: #06b6d4;
}

.file-button:hover {
  background-color: #f1f5f9;
  color: #06b6d4;
}

.voice-button {
  background-color: transparent;
  border: none;
  color: #64748b;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  height: 44px;
}

.voice-button:hover {
  background-color: #f1f5f9;
  color: #06b6d4;
}

.message-input {
  flex: 1;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 0.95rem;
  resize: none;
  outline: none;
  transition: border-color 0.2s;
  font-family: inherit;
  line-height: 1.5;
  max-height: 120px;
  min-height: 44px;
  background-color: #ffffff;
}

.message-input:focus {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.message-input:disabled {
  background-color: #f8fafc;
  color: #64748b;
  cursor: not-allowed;
}

.send-button {
  background-color: #06b6d4;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 10px 16px;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 500;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  height: 44px;
  box-shadow: 0 2px 4px rgba(6, 182, 212, 0.2);
}

.send-button:hover:not(:disabled) {
  background-color: #0891b2;
  box-shadow: 0 4px 8px rgba(6, 182, 212, 0.3);
  transform: translateY(-1px);
}

.send-button:hover:not(:disabled) {
  background-color: #0891b2;
}

.send-button:disabled {
  background-color: #cbd5e1;
  cursor: not-allowed;
}

.input-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6px;
  font-size: 0.75rem;
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
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.debug-info-content {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  max-width: 800px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.debug-info-content h3 {
  margin: 0 0 20px 0;
  color: #06b6d4;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 8px;
}

.debug-section {
  margin-bottom: 24px;
}

.debug-section h4 {
  margin: 0 0 12px 0;
  color: #374151;
  font-size: 1rem;
}

.debug-item {
  margin-bottom: 8px;
  padding: 8px;
  background-color: #f8fafc;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.9rem;
}

.debug-response {
  margin-bottom: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  overflow: hidden;
}

.debug-response-header {
  background-color: #f1f5f9;
  padding: 8px 12px;
  font-weight: 600;
  font-size: 0.9rem;
  border-bottom: 1px solid #e2e8f0;
}

.debug-response-data {
  margin: 0;
  padding: 12px;
  background-color: #f8fafc;
  font-size: 0.8rem;
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.debug-error {
  margin-bottom: 12px;
  border: 1px solid #fecaca;
  border-radius: 4px;
  overflow: hidden;
}

.debug-error-header {
  background-color: #fef2f2;
  padding: 8px 12px;
  font-weight: 600;
  font-size: 0.9rem;
  border-bottom: 1px solid #fecaca;
  color: #dc2626;
}

.debug-error-message {
  padding: 12px;
  background-color: #fef2f2;
  font-size: 0.9rem;
  color: #dc2626;
  word-break: break-word;
}

.debug-buttons {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.debug-buttons button {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background-color 0.2s;
}

.refresh-btn {
  background-color: #06b6d4;
  color: white;
}

.refresh-btn:hover {
  background-color: #0891b2;
}

.debug-buttons button:last-child {
  background-color: #64748b;
  color: white;
}

.debug-buttons button:last-child:hover {
  background-color: #475569;
}

.test-btn {
  background-color: #10b981 !important;
}

.test-btn:hover {
  background-color: #059669 !important;
}

.reset-btn {
  background-color: #ef4444 !important;
}

.reset-btn:hover {
  background-color: #dc2626 !important;
}

/* File attachment styles */
.file-attachment {
  margin-top: 8px;
}

.image-attachment {
  position: relative;
  display: inline-block;
  max-width: 100%;
}

.chat-image {
  max-width: 400px;
  max-height: 400px;
  width: auto;
  height: auto;
  object-fit: contain;
  border-radius: 8px;
  cursor: pointer;
  transition: opacity 0.2s;
  display: block;
}

.chat-image:hover {
  opacity: 0.9;
}

.video-attachment {
  position: relative;
  display: inline-block;
  max-width: 100%;
}

.chat-video {
  max-width: 400px;
  max-height: 400px;
  width: auto;
  height: auto;
  border-radius: 8px;
  display: block;
}

.audio-attachment {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.audio-player {
  display: flex;
  align-items: center;
  gap: 8px;
}

.audio-icon {
  font-size: 1.5rem;
  color: #06b6d4;
}

.chat-audio {
  flex: 1;
  max-width: 300px;
}

.file-info-overlay {
  margin-top: 8px;
  padding: 8px;
  background-color: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 0.85rem;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.file-size {
  color: #64748b;
  font-size: 0.8rem;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.file-icon {
  font-size: 1.5rem;
}

.file-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.download-btn {
  background-color: transparent;
  border: 1px solid #e2e8f0;
  color: #64748b;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 4px;
}

.download-btn:hover {
  background-color: #06b6d4;
  border-color: #06b6d4;
  color: white;
}

/* Responsive design */
@media (max-width: 768px) {
  .message {
    max-width: 85%;
  }

  .messages-header {
    padding: 6px 12px;
  }

  .messages-header h3 {
    font-size: 0.85rem;
  }

  .timezone-info,
  .member-count {
    font-size: 0.7rem;
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

  .chat-image {
    max-width: 100%;
    max-height: 300px;
  }

  .chat-video {
    max-width: 100%;
    max-height: 300px;
  }
}
</style>
