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
const isProcessing = ref(false)
const showPassword = ref(false)
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)

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
    if (isProcessing.value) return // Prevent double-clicks

    if (!username.value || !password.value) {
      error.value = 'Please enter both username and password'
      return
    }

    isProcessing.value = true // Start processing
    error.value = '' // Clear previous errors

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

      if (verifyResult.success) {
        // Changed this line to only check success
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
  } finally {
    isProcessing.value = false // End processing
  }
}

// Update the changePassword function verification check as well
const changePassword = async () => {
  try {
    if (isProcessing.value) return // Prevent double-clicks

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

    isProcessing.value = true // Start processing
    error.value = '' // Clear previous errors

    // First verify current credentials
    const result = await callApi({
      query: `SELECT u.id, u.username, u.role_id, u.password, r.role_name 
              FROM users u 
              JOIN roles r ON u.role_id = r.id 
              WHERE u.username = ?`,
      params: [username.value],
    })

    if (!result.success || !result.data.length) {
      error.value =
        'User not found ' +
        username.value +
        ' ' +
        result.success +
        ' ' +
        result.data.length +
        ' ' +
        result.error
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
      action: 'hash_password', // This tells the API to hash the password before updating
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
  } finally {
    isProcessing.value = false // End processing
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
      <h2>
        <i :class="showChangePassword ? 'fas fa-key' : 'fas fa-sign-in-alt'"></i>
        {{ showChangePassword ? 'Change Password' : 'Login' }}
      </h2>

      <!-- Username field -->
      <div class="form-group">
        <div class="input-with-icon">
          <i class="fas fa-user"></i>
          <input
            type="text"
            v-model="username"
            placeholder="Username"
            required
            :disabled="isProcessing"
            autocomplete="username"
          />
        </div>
      </div>

      <!-- Current Password field -->
      <div class="form-group">
        <div class="input-with-icon">
          <i class="fas fa-lock"></i>
          <input
            :type="showPassword ? 'text' : 'password'"
            v-model="password"
            placeholder="Current Password"
            required
            :disabled="isProcessing"
            autocomplete="current-password"
          />
          <button
            type="button"
            class="toggle-password"
            @click="showPassword = !showPassword"
            :disabled="isProcessing"
          >
            <i :class="showPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
          </button>
        </div>
      </div>

      <!-- Change Password Fields -->
      <template v-if="showChangePassword">
        <div class="form-group">
          <div class="input-with-icon">
            <i class="fas fa-key"></i>
            <input
              :type="showNewPassword ? 'text' : 'password'"
              v-model="newPassword"
              placeholder="New Password"
              required
              :disabled="isProcessing"
              autocomplete="new-password"
            />
            <button
              type="button"
              class="toggle-password"
              @click="showNewPassword = !showNewPassword"
              :disabled="isProcessing"
            >
              <i :class="showNewPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
            </button>
          </div>
        </div>
        <div class="form-group">
          <div class="input-with-icon">
            <i class="fas fa-key"></i>
            <input
              :type="showConfirmPassword ? 'text' : 'password'"
              v-model="confirmPassword"
              placeholder="Confirm New Password"
              required
              :disabled="isProcessing"
              autocomplete="new-password"
            />
            <button
              type="button"
              class="toggle-password"
              @click="showConfirmPassword = !showConfirmPassword"
              :disabled="isProcessing"
            >
              <i :class="showConfirmPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
            </button>
          </div>
        </div>
      </template>

      <!-- Error and Success Messages -->
      <div v-if="error" class="message error" role="alert">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>
      <div v-if="successMessage" class="message success" role="alert">
        <i class="fas fa-check-circle"></i>
        {{ successMessage }}
      </div>

      <!-- Submit Button -->
      <button type="submit" :disabled="isProcessing" :class="{ processing: isProcessing }">
        <span class="button-content">
          <i
            :class="
              isProcessing
                ? 'fas fa-spinner fa-spin'
                : showChangePassword
                  ? 'fas fa-key'
                  : 'fas fa-sign-in-alt'
            "
          ></i>
          {{ isProcessing ? 'Processing...' : showChangePassword ? 'Change Password' : 'Login' }}
        </span>
      </button>

      <!-- Toggle Change Password Mode -->
      <button
        type="button"
        class="secondary-btn"
        @click="showChangePassword = !showChangePassword"
        :disabled="isProcessing"
      >
        <i :class="showChangePassword ? 'fas fa-arrow-left' : 'fas fa-key'"></i>
        {{ showChangePassword ? 'Back to Login' : 'Change Password' }}
      </button>
    </form>
    <div class="copyright">© Merhab Noureddine 2025</div>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.login-form {
  width: 100%;
  max-width: 400px;
  padding: 2rem;
  background: white;
  border-radius: 10px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

h2 {
  text-align: center;
  color: #2c3e50;
  margin-bottom: 1.5rem;
  font-size: 1.8rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.form-group {
  margin-bottom: 1rem;
  position: relative;
}

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.input-with-icon i {
  position: absolute;
  left: 1rem;
  color: #7f8c8d;
  font-size: 1rem;
}

input {
  width: 100%;
  padding: 0.8rem 1rem 0.8rem 2.5rem;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 1rem;
  transition: all 0.3s ease;
}

input:focus {
  border-color: #409eff;
  outline: none;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

input:disabled {
  background-color: #f5f7fa;
  cursor: not-allowed;
}

.toggle-password {
  position: absolute;
  right: 1rem;
  background: none;
  border: none;
  color: #7f8c8d;
  cursor: pointer;
  padding: 0;
}

.toggle-password:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.message {
  padding: 0.8rem;
  margin-bottom: 1rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  font-size: 0.9rem;
}

.message i {
  margin-right: 0.5rem;
}

.error {
  background-color: #fef0f0;
  color: #f56c6c;
  border: 1px solid #fde2e2;
}

.success {
  background-color: #f0f9eb;
  color: #67c23a;
  border: 1px solid #e1f3d8;
}

button {
  width: 100%;
  padding: 0.8rem;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  background-color: #409eff;
  color: white;
}

button:not(:disabled):hover {
  background-color: #66b1ff;
  transform: translateY(-1px);
}

button:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
}

.processing {
  position: relative;
  pointer-events: none;
}

.button-content {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.secondary-btn {
  background-color: #909399;
}

.secondary-btn:not(:disabled):hover {
  background-color: #a6a9ad;
}

@media (max-width: 480px) {
  .login-form {
    margin: 1rem;
    padding: 1.5rem;
  }

  h2 {
    font-size: 1.5rem;
  }

  input {
    font-size: 16px; /* Prevents zoom on mobile */
  }
}

.copyright {
  text-align: center;
  color: #6b7280;
  font-size: 0.875rem;
  margin-top: 1rem;
  position: absolute;
  bottom: 1rem;
  width: 100%;
  left: 0;
}
</style>
