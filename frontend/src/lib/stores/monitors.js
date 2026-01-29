import { writable } from 'svelte/store';

export const monitors = writable([]);

export async function fetchMonitors() {
  try {
    const response = await fetch('/api/v1/monitors');
    const data = await response.json();
    monitors.set(data);
    return data;
  } catch (error) {
    console.error('Failed to fetch monitors:', error);
    return [];
  }
}

export async function createMonitor(monitor) {
  try {
    const response = await fetch('/api/v1/monitors', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(monitor)
    });
    
    if (response.ok) {
      const newMonitor = await response.json();
      monitors.update(list => [...list, newMonitor]);
      return newMonitor;
    }
    throw new Error('Failed to create monitor');
  } catch (error) {
    console.error('Failed to create monitor:', error);
    throw error;
  }
}

export async function deleteMonitor(id) {
  try {
    const response = await fetch(`/api/v1/monitors/${id}`, {
      method: 'DELETE'
    });
    
    if (response.ok) {
      monitors.update(list => list.filter(m => m.id !== id));
      return true;
    }
    throw new Error('Failed to delete monitor');
  } catch (error) {
    console.error('Failed to delete monitor:', error);
    throw error;
  }
}