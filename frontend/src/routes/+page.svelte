<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import MonitorList from '$lib/components/MonitorList.svelte';
  import MonitorForm from '$lib/components/MonitorForm.svelte';
  import MonitorEdit from '$lib/components/MonitorEdit.svelte';
  import Dashboard from '$lib/components/Dashboard.svelte';
  import { authStore } from '$lib/stores/auth.js';

  let currentView = 'dashboard';
  let isAuthenticated = false;
  let editingMonitor = null;
  let loginForm = {
    username: 'admin',
    password: 'password'
  };

  authStore.subscribe(auth => {
    isAuthenticated = auth.isAuthenticated;
  });

  onMount(() => {
    // Check authentication status
    authStore.checkAuth();
  });

  function setView(view) {
    currentView = view;
  }

  function logout() {
    authStore.logout();
  }

  async function handleLogin() {
    const success = await authStore.login(loginForm.username, loginForm.password);
    if (!success) {
      alert('Invalid login credentials');
    }
  }

  function handleEdit(event) {
    editingMonitor = event.detail;
    currentView = 'edit';
  }

  function handleEditSaved() {
    editingMonitor = null;
    currentView = 'monitors';
    // Refresh monitors list
    // This will be handled by the component itself
  }

  function handleEditCancel() {
    editingMonitor = null;
    currentView = 'monitors';
  }
</script>

<main class="app">
  {#if !isAuthenticated}
    <div class="login-container">
      <h1>Uptime Monitor</h1>
      <div class="login-form">
        <h2>Login</h2>
        <form on:submit|preventDefault={handleLogin}>
          <input type="text" bind:value={loginForm.username} placeholder="Username" required />
          <input type="password" bind:value={loginForm.password} placeholder="Password" required />
          <button type="submit">Login</button>
        </form>
        <p class="default-creds">Default: admin / password</p>
      </div>
    </div>
  {:else}
    <nav class="nav">
      <div class="nav-brand">
        <h1>Uptime Monitor</h1>
      </div>
      <div class="nav-links">
        <button class:active={currentView === 'dashboard'} on:click={() => setView('dashboard')}>
          Dashboard
        </button>
        <button class:active={currentView === 'monitors'} on:click={() => setView('monitors')}>
          Monitors
        </button>
        <button class:active={currentView === 'add'} on:click={() => setView('add')}>
          Add Monitor
        </button>
        <button on:click={logout} class="logout-btn">Logout</button>
      </div>
    </nav>

    <div class="content">
      {#if currentView === 'dashboard'}
        <Dashboard />
      {:else if currentView === 'monitors'}
        <MonitorList on:edit={handleEdit} />
      {:else if currentView === 'add'}
        <MonitorForm on:saved={() => setView('monitors')} />
      {:else if currentView === 'edit' && editingMonitor}
        <MonitorEdit monitor={editingMonitor} on:saved={handleEditSaved} on:cancel={handleEditCancel} />
      {/if}
    </div>
  {/if}
</main>

<style>
  .app {
    min-height: 100vh;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }

  .login-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  }

  .login-form {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    width: 100%;
    max-width: 400px;
  }

  .login-form h2 {
    margin: 0 0 1rem 0;
    text-align: center;
  }

  .login-form input {
    width: 100%;
    padding: 0.75rem;
    margin-bottom: 1rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    box-sizing: border-box;
  }

  .login-form button {
    width: 100%;
    padding: 0.75rem;
    background: #667eea;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }

  .default-creds {
    margin-top: 1rem;
    text-align: center;
    color: #666;
    font-size: 0.875rem;
  }

  .nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 2rem;
    background: white;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .nav-brand h1 {
    margin: 0;
    color: #667eea;
  }

  .nav-links {
    display: flex;
    gap: 1rem;
  }

  .nav-links button {
    padding: 0.5rem 1rem;
    border: none;
    background: none;
    cursor: pointer;
    border-radius: 4px;
    transition: background-color 0.2s;
  }

  .nav-links button:hover,
  .nav-links button.active {
    background: #f0f0f0;
  }

  .logout-btn {
    background: #ff6b6b !important;
    color: white !important;
  }

  .content {
    padding: 2rem;
    max-width: 1200px;
    margin: 0 auto;
  }
</style>