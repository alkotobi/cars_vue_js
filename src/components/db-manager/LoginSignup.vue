<template>
  <div class="login-signup-container">
    <form @submit.prevent="isSignup ? signup() : login()" class="auth-form">
      <h2>
        <i :class="isSignup ? 'fas fa-user-plus' : 'fas fa-sign-in-alt'"></i>
        {{ isSignup ? 'Sign Up' : 'Login' }}
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

      <!-- Password field -->
      <div class="form-group">
        <div class="input-with-icon">
          <i class="fas fa-lock"></i>
          <input
            :type="showPassword ? 'text' : 'password'"
            v-model="password"
            :placeholder="isSignup ? 'Password' : 'Password'"
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

      <!-- Confirm Password field (only for signup) -->
      <div class="form-group" v-if="isSignup">
        <div class="input-with-icon">
          <i class="fas fa-lock"></i>
          <input
            :type="showConfirmPassword ? 'text' : 'password'"
            v-model="confirmPassword"
            placeholder="Confirm Password"
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
                : isSignup
                  ? 'fas fa-user-plus'
                  : 'fas fa-sign-in-alt'
            "
          ></i>
          {{
            isProcessing
              ? 'Processing...'
              : isSignup
                ? 'Sign Up'
                : 'Login'
          }}
        </span>
      </button>

      <!-- Toggle Login/Signup Mode -->
      <button
        type="button"
        class="secondary-btn"
        @click="toggleMode"
        :disabled="isProcessing"
      >
        <i :class="isSignup ? 'fas fa-sign-in-alt' : 'fas fa-user-plus'"></i>
        {{ isSignup ? 'Already have an account? Login' : "Don't have an account? Sign Up" }}
      </button>
    </form>
  </div>
</template>

<script setup>
import { ref, defineEmits } from 'vue'

const emit = defineEmits(['login-success'])

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const successMessage = ref('')
const isProcessing = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const isSignup = ref(false)

// Get API base URL
const getApiBaseUrl = () => {
  const hostname = window.location.hostname
  const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'
  return isLocalhost ? 'http://localhost:8000/api' : 'https://www.merhab.com/api'
}

const toggleMode = () => {
  isSignup.value = !isSignup.value
  error.value = ''
  successMessage.value = ''
  password.value = ''
  confirmPassword.value = ''
}

const login = async () => {
  try {
    if (isProcessing.value) return

    if (!username.value || !password.value) {
      error.value = 'Please enter username and password'
      return
    }

    isProcessing.value = true
    error.value = ''
    successMessage.value = ''

    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'login',
        user: username.value,
        pass: password.value,
      }),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = 'Login successful!'
      // Store user info if needed
      if (result.data) {
        localStorage.setItem('db_manager_user', JSON.stringify(result.data))
      }
      // Emit login success event
      emit('login-success', result.data)
      // Clear form after successful login
      setTimeout(() => {
        username.value = ''
        password.value = ''
      }, 1000)
    } else {
      error.value = result.message || 'Login failed'
    }
  } catch (err) {
    error.value = 'An error occurred during login'
    console.error(err)
  } finally {
    isProcessing.value = false
  }
}

const signup = async () => {
  try {
    if (isProcessing.value) return

    if (!username.value || !password.value) {
      error.value = 'Please enter username and password'
      return
    }

    if (password.value !== confirmPassword.value) {
      error.value = 'Passwords do not match'
      return
    }

    if (password.value.length < 6) {
      error.value = 'Password must be at least 6 characters long'
      return
    }

    isProcessing.value = true
    error.value = ''
    successMessage.value = ''

    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'signup',
        user: username.value,
        pass: password.value,
      }),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = 'Sign up successful! You can now login.'
      // Switch to login mode after successful signup
      setTimeout(() => {
        isSignup.value = false
        username.value = ''
        password.value = ''
        confirmPassword.value = ''
      }, 2000)
    } else {
      error.value = result.message || 'Sign up failed'
    }
  } catch (err) {
    error.value = 'An error occurred during sign up'
    console.error(err)
  } finally {
    isProcessing.value = false
  }
}
</script>

<style scoped>
.login-signup-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 2rem;
}

.auth-form {
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
  z-index: 1;
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
  z-index: 1;
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
  .auth-form {
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
</style>

