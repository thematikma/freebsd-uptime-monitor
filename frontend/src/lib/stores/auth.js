import { writable } from 'svelte/store';
import { browser } from '$app/environment';

/**
 * @typedef {Object} AuthState
 * @property {boolean} isAuthenticated
 * @property {any} user
 * @property {string|null} token
 * @property {boolean} needsSetup
 */

const initialState = {
  isAuthenticated: false,
  user: null,
  token: null,
  needsSetup: false
};

function createAuthStore() {
  const { subscribe, set, update } = writable(initialState);

  return {
    subscribe,
    
    // Check if the system needs initial setup (no users exist)
    checkSetupStatus: async () => {
      try {
        const response = await fetch('/api/v1/auth/setup-status');
        if (response.ok) {
          const data = await response.json();
          update(state => ({
            ...state,
            needsSetup: data.needs_setup
          }));
          return data.needs_setup;
        }
      } catch (error) {
        console.error('Setup status check error:', error);
      }
      return false;
    },

    // Setup first admin user
    setup: async (username, email, password) => {
      try {
        const response = await fetch('/api/v1/auth/setup', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, email, password })
        });

        if (response.ok) {
          const data = await response.json();
          const authState = {
            isAuthenticated: true,
            user: data.user,
            token: data.token,
            needsSetup: false
          };
          
          if (browser) {
            localStorage.setItem('auth_token', data.token);
          }
          
          set(authState);
          return { success: true };
        } else {
          const error = await response.json();
          return { success: false, error: error.error || 'Setup failed' };
        }
      } catch (error) {
        console.error('Setup error:', error);
        return { success: false, error: 'Connection error' };
      }
    },

    // Register a new user
    register: async (username, email, password) => {
      try {
        const response = await fetch('/api/v1/auth/register', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, email, password })
        });

        if (response.ok) {
          const data = await response.json();
          const authState = {
            isAuthenticated: true,
            user: data.user,
            token: data.token,
            needsSetup: false
          };
          
          if (browser) {
            localStorage.setItem('auth_token', data.token);
          }
          
          set(authState);
          return { success: true };
        } else {
          const error = await response.json();
          return { success: false, error: error.error || 'Registration failed' };
        }
      } catch (error) {
        console.error('Registration error:', error);
        return { success: false, error: 'Connection error' };
      }
    },

    login: async (username, password) => {
      try {
        const response = await fetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password })
        });

        if (response.ok) {
          const data = await response.json();
          const authState = {
            isAuthenticated: true,
            user: data.user,
            token: data.token,
            needsSetup: false
          };
          
          if (browser) {
            localStorage.setItem('auth_token', data.token);
          }
          
          set(authState);
          return true;
        }
        return false;
      } catch (error) {
        console.error('Login error:', error);
        return false;
      }
    },
    
    logout: () => {
      if (browser) {
        localStorage.removeItem('auth_token');
      }
      set(initialState);
    },
    
    checkAuth: async () => {
      if (browser) {
        const token = localStorage.getItem('auth_token');
        if (token) {
          try {
            // Verify token with the server
            const response = await fetch('/api/v1/auth/profile', {
              headers: {
                'Authorization': `Bearer ${token}`
              }
            });
            
            if (response.ok) {
              const user = await response.json();
              update(state => ({
                ...state,
                isAuthenticated: true,
                token,
                user,
                needsSetup: false
              }));
              return true;
            } else {
              // Token is invalid, remove it
              localStorage.removeItem('auth_token');
            }
          } catch (error) {
            console.error('Auth check error:', error);
          }
        }
      }
      return false;
    }
  };
}

export const authStore = createAuthStore();