<script setup>
import { ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const router = useRouter()
const { callApi, error: apiError, loading } = useApi()
const username = ref('')
const password = ref('')
const error = ref('')
const showChangePassword = ref(false)
const newPassword = ref('')
const confirmPassword = ref('')
const successMessage = ref('')
const messageTimeout = ref(null)

// Clear messages after a delay
const clearMessages = () => {
  if (messageTimeout.value) {
    clearTimeout(messageTimeout.value)
  }
  messageTimeout.value = setTimeout(() => {
    error.value = ''
    successMessage.value = ''
  }, 3000) // Messages will disappear after 3 seconds
}

// Watch for changes in error or success messages
watch([error, successMessage], () => {
  if (error.value || successMessage.value) {
    clearMessages()
  }
})

const login = async () => {
  try {
    if (!username.value || !password.value) {
      error.value = 'Please enter both username and password'
      return
    }

    // Get user and verify credentials
    const result = await callApi({
      query: `SELECT u.id, u.username, u.role_id, u.password, r.role_name 
              FROM users u 
              JOIN roles r ON u.role_id = r.id 
              WHERE u.username = ?`,
      params: [username.value],
    })
    console.log('User query result:', result)
    
    if (result.success && result.data.length > 0) {
      const user = result.data[0]
      console.log('User data:', user)

      // Verify password
      const verifyResult = await callApi({
        action: 'verify_password',
        password: password.value,
        hash: user.password,
      })
      console.log('Password verification result:', verifyResult)

      if (verifyResult.success) {  // Changed this line to only check success
        // Get user permissions
        const permissionsResult = await callApi({
          query: `SELECT p.permission_name, p.description 
                  FROM permissions p 
                  JOIN role_permissions rp ON p.id = rp.permission_id 
                  WHERE rp.role_id = ?`,
          params: [user.role_id],
        })

        // Store user data without the password hash
        const { password: _, ...userData } = user
        const userInfo = {
          ...userData,
          permissions: permissionsResult.success ? permissionsResult.data : [],
        }

        localStorage.setItem('user', JSON.stringify(userInfo))
        router.push('/dashboard')
      } else {
        error.value = 'Invalid username or password'
      }
    } else {
      error.value = 'Invalid username or password'
    }
  } catch (err) {
    error.value = 'An error occurred during login'
    console.error(err)
  }
}

// Update the changePassword function verification check as well
const changePassword = async () => {
  try {
    if (!username.value || !password.value) {
      error.value = 'Please enter your current username and password'
      return
    }

    if (!newPassword.value || !confirmPassword.value) {
      error.value = 'Please enter and confirm your new password'
      return
    }

    if (newPassword.value !== confirmPassword.value) {
      error.value = 'Passwords do not match'
      return
    }

    // First verify current credentials
    const result = await callApi({
      query: `SELECT u.id, u.username, u.role_id, u.password, r.role_name 
              FROM users u 
              JOIN roles r ON u.role_id = r.id 
              WHERE u.username = ?`,
      params: [username.value],
    })

    if (!result.success || !result.data.length) {
      error.value = 'User not found ' + username.value +' '+result.success+' '+result.data.length+' '+result.error
      return
    }

    const user = result.data[0]

    // Verify current password
    const verifyResult = await callApi({
      action: 'verify_password',
      password: password.value,
      hash: user.password,
    })

    if (!verifyResult.success) {
      error.value = 'Current password is incorrect'
      return
    }

    // Update password with proper hashing
    const updateResult = await callApi({
      query: `UPDATE users SET password = ? WHERE username = ?`,
      params: [newPassword.value, username.value],
      action: 'hash_password'  // This tells the API to hash the password before updating
    })

    if (updateResult.success) {
      successMessage.value = 'Password changed successfully'
      showChangePassword.value = false
      newPassword.value = ''
      confirmPassword.value = ''
      password.value = ''
    } else {
      error.value = 'Failed to change password'
    }
  } catch (err) {
    error.value = 'An error occurred while changing password'
    console.error(err)
  }
}

// Clean up timeout when component is unmounted
onUnmounted(() => {
  if (messageTimeout.value) {
    clearTimeout(messageTimeout.value)
  }
})
</script>

<template>
  <div class="login-container">
    <form @submit.prevent="showChangePassword ? changePassword() : login()" class="login-form">
      <h2>{{ showChangePassword ? 'Change Password' : 'Login' }}</h2>
      <div class="form-group">
        <input type="text" v-model="username" placeholder="Username" required :disabled="loading" />
      </div>
      <div class="form-group">
        <input
          type="password"
          v-model="password"
          placeholder="Current Password"
          required
          :disabled="loading"
        />
      </div>
      <template v-if="showChangePassword">
        <div class="form-group">
          <input
            type="password"
            v-model="newPassword"
            placeholder="New Password"
            required
            :disabled="loading"
          />
        </div>
        <div class="form-group">
          <input
            type="password"
            v-model="confirmPassword"
            placeholder="Confirm New Password"
            required
            :disabled="loading"
          />
        </div>
      </template>
      <div v-if="error" class="error">{{ error }}</div>
      <div v-if="successMessage" class="success">{{ successMessage }}</div>
      <button type="submit" :disabled="loading">
        {{ loading ? 'Processing...' : showChangePassword ? 'Change Password' : 'Login' }}
      </button>
      <button
        type="button"
        class="secondary-btn"
        @click="showChangePassword = !showChangePassword"
        :disabled="loading"
      >
        {{ showChangePassword ? 'Back to Login' : 'Change Password' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

.login-form {
  width: 300px;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 5px;
}

.form-group {
  margin-bottom: 15px;
}

input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

button {
  width: 100%;
  padding: 10px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #45a049;
}

.error {
  color: red;
  margin-bottom: 10px;
  text-align: center;
}

.success {
  color: #4caf50;
  margin-bottom: 10px;
  text-align: center;
}

.secondary-btn {
  width: 100%;
  padding: 10px;
  background-color: #757575;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 10px;
}

.secondary-btn:hover {
  background-color: #616161;
}
</style>
