import { writable } from 'svelte/store';
import { browser } from '$app/environment';

function createThemeStore() {
  // Initialize from localStorage if available
  const initialValue = browser ? localStorage.getItem('darkMode') === 'true' : false;
  const { subscribe, set, update } = writable(initialValue);

  return {
    subscribe,
    toggle: () => {
      update(current => {
        const newValue = !current;
        if (browser) {
          localStorage.setItem('darkMode', newValue.toString());
        }
        return newValue;
      });
    },
    set: (value) => {
      if (browser) {
        localStorage.setItem('darkMode', value.toString());
      }
      set(value);
    },
    init: () => {
      if (browser) {
        const stored = localStorage.getItem('darkMode') === 'true';
        set(stored);
      }
    }
  };
}

export const darkMode = createThemeStore();
