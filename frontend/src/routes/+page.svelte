<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import MonitorList from '$lib/components/MonitorList.svelte';
  import MonitorForm from '$lib/components/MonitorForm.svelte';
  import MonitorEdit from '$lib/components/MonitorEdit.svelte';
  import Dashboard from '$lib/components/Dashboard.svelte';
  import Notifications from '$lib/components/Notifications.svelte';
  import { authStore } from '$lib/stores/auth.js';

  let currentView = 'dashboard';
  let isAuthenticated = false;
  let needsSetup = false;
  let authMode = 'login'; // 'login', 'register', 'setup'
  let editingMonitor = null;
  let loading = true;
  let error = '';
  let user = null;

  let loginForm = {
    username: '',
    password: ''
  };

  let registerForm = {
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  };

  authStore.subscribe(auth => {
    isAuthenticated = auth.isAuthenticated;
    needsSetup = auth.needsSetup;
    user = auth.user;
  });

  onMount(async () => {
    // First check setup status
    const setupNeeded = await authStore.checkSetupStatus();
    if (setupNeeded) {
      authMode = 'setup';
      loading = false;
      return;
    }

    // Then check if already authenticated
    await authStore.checkAuth();
    loading = false;
  });

  function setView(view) {
    currentView = view;
  }

  function logout() {
    authStore.logout();
  }

  async function handleLogin() {
    error = '';
    const success = await authStore.login(loginForm.username, loginForm.password);
    if (!success) {
      error = 'Invalid login credentials';
    }
  }

  async function handleSetup() {
    error = '';
    
    if (registerForm.password !== registerForm.confirmPassword) {
      error = 'Passwords do not match';
      return;
    }
    
    if (registerForm.password.length < 6) {
      error = 'Password must be at least 6 characters';
      return;
    }

    const result = await authStore.setup(
      registerForm.username,
      registerForm.email,
      registerForm.password
    );
    
    if (!result.success) {
      error = result.error;
    }
  }

  async function handleRegister() {
    error = '';
    
    if (registerForm.password !== registerForm.confirmPassword) {
      error = 'Passwords do not match';
      return;
    }
    
    if (registerForm.password.length < 6) {
      error = 'Password must be at least 6 characters';
      return;
    }

    const result = await authStore.register(
      registerForm.username,
      registerForm.email,
      registerForm.password
    );
    
    if (!result.success) {
      error = result.error;
    }
  }

  function switchToRegister() {
    error = '';
    authMode = 'register';
    registerForm = { username: '', email: '', password: '', confirmPassword: '' };
  }

  function switchToLogin() {
    error = '';
    authMode = 'login';
    loginForm = { username: '', password: '' };
  }

  function handleEdit(event) {
    editingMonitor = event.detail;
    currentView = 'edit';
  }

  function handleEditSaved() {
    editingMonitor = null;
    currentView = 'monitors';
  }

  function handleEditCancel() {
    editingMonitor = null;
    currentView = 'monitors';
  }
</script>

<main class="app">
  {#if loading}
    <div class="loading-container">
      <div class="spinner"></div>
      <p>Loading...</p>
    </div>
  {:else if !isAuthenticated}
    <div class="login-container">
      <h1>Uptime Monitor</h1>
      
      {#if authMode === 'setup'}
        <!-- Initial Setup Form -->
        <div class="auth-form">
          <h2>🚀 Initial Setup</h2>
          <p class="setup-info">Create your administrator account to get started.</p>
          
          {#if error}
            <div class="error-message">{error}</div>
          {/if}
          
          <form on:submit|preventDefault={handleSetup}>
            <input 
              type="text" 
              bind:value={registerForm.username} 
              placeholder="Admin Username" 
              required 
              autocomplete="username"
            />
            <input 
              type="email" 
              bind:value={registerForm.email} 
              placeholder="Email Address" 
              required 
              autocomplete="email"
            />
            <input 
              type="password" 
              bind:value={registerForm.password} 
              placeholder="Password (min 6 chars)" 
              required 
              minlength="6"
              autocomplete="new-password"
            />
            <input 
              type="password" 
              bind:value={registerForm.confirmPassword} 
              placeholder="Confirm Password" 
              required 
              minlength="6"
              autocomplete="new-password"
            />
            <button type="submit" class="primary-btn">Create Admin Account</button>
          </form>
        </div>
      {:else if authMode === 'register'}
        <!-- Registration Form -->
        <div class="auth-form">
          <h2>Create Account</h2>
          
          {#if error}
            <div class="error-message">{error}</div>
          {/if}
          
          <form on:submit|preventDefault={handleRegister}>
            <input 
              type="text" 
              bind:value={registerForm.username} 
              placeholder="Username" 
              required 
              autocomplete="username"
            />
            <input 
              type="email" 
              bind:value={registerForm.email} 
              placeholder="Email Address" 
              required 
              autocomplete="email"
            />
            <input 
              type="password" 
              bind:value={registerForm.password} 
              placeholder="Password (min 6 chars)" 
              required 
              minlength="6"
              autocomplete="new-password"
            />
            <input 
              type="password" 
              bind:value={registerForm.confirmPassword} 
              placeholder="Confirm Password" 
              required 
              minlength="6"
              autocomplete="new-password"
            />
            <button type="submit" class="primary-btn">Register</button>
          </form>
          <p class="auth-switch">
            Already have an account? 
            <button type="button" class="link-btn" on:click={switchToLogin}>Login</button>
          </p>
        </div>
      {:else}
        <!-- Login Form -->
        <div class="auth-form">
          <h2>Login</h2>
          
          {#if error}
            <div class="error-message">{error}</div>
          {/if}
          
          <form on:submit|preventDefault={handleLogin}>
            <input 
              type="text" 
              bind:value={loginForm.username} 
              placeholder="Username" 
              required 
              autocomplete="username"
            />
            <input 
              type="password" 
              bind:value={loginForm.password} 
              placeholder="Password" 
              required 
              autocomplete="current-password"
            />
            <button type="submit" class="primary-btn">Login</button>
          </form>
          <p class="auth-switch">
            Don't have an account? 
            <button type="button" class="link-btn" on:click={switchToRegister}>Register</button>
          </p>
        </div>
      {/if}
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
        <button class:active={currentView === 'notifications'} on:click={() => setView('notifications')}>
          Notifications
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
      {:else if currentView === 'notifications'}
        <Notifications />
      {/if}
    </div>
  {/if}
</main>

<style>
  .app {
    min-height: 100vh;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .login-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  }

  .login-container > h1 {
    color: white;
    margin-bottom: 1.5rem;
    font-size: 2.5rem;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  }

  .auth-form {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
    width: 100%;
    max-width: 400px;
  }

  .auth-form h2 {
    margin: 0 0 1rem 0;
    text-align: center;
    color: #333;
  }

  .setup-info {
    text-align: center;
    color: #666;
    margin-bottom: 1.5rem;
    font-size: 0.9rem;
  }

  .error-message {
    background: #ffe0e0;
    color: #c00;
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 1rem;
    text-align: center;
    font-size: 0.9rem;
  }

  .auth-form input {
    width: 100%;
    padding: 0.75rem;
    margin-bottom: 1rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    box-sizing: border-box;
    font-size: 1rem;
  }

  .auth-form input:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
  }

  .primary-btn {
    width: 100%;
    padding: 0.75rem;
    background: #667eea;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 1rem;
    font-weight: 500;
    transition: background 0.2s;
  }

  .primary-btn:hover {
    background: #5a6fd6;
  }

  .auth-switch {
    margin-top: 1.5rem;
    text-align: center;
    color: #666;
    font-size: 0.9rem;
  }

  .link-btn {
    background: none;
    border: none;
    color: #667eea;
    cursor: pointer;
    font-size: 0.9rem;
    text-decoration: underline;
    padding: 0;
  }

  .link-btn:hover {
    color: #5a6fd6;
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
    align-items: center;
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

  .logout-btn:hover {
    background: #e55555 !important;
  }

  .content {
    padding: 2rem;
    max-width: 1200px;
    margin: 0 auto;
  }
</style>