import { writable } from 'svelte/store';
import { browser } from '$app/environment';

interface AuthState {
  isAuthenticated: boolean;
  user: any | null;
  token: string | null;
}

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  return {
    subscribe,
    login: async (username: string, password: string) => {
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
            token: data.token
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
    checkAuth: () => {
      if (browser) {
        const token = localStorage.getItem('auth_token');
        if (token) {
          // In a real app, you'd verify the token with the server
          update(state => ({
            ...state,
            isAuthenticated: true,
            token
          }));
        }
      }
    }
  };
}

export const authStore = createAuthStore();