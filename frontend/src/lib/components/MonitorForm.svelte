<script>
  import { createEventDispatcher } from 'svelte';
  import { createMonitor } from '$lib/stores/monitors.js';
  import { darkMode } from '$lib/stores/theme.js';

  const dispatch = createEventDispatcher();

  let isDarkMode = false;
  darkMode.subscribe(value => {
    isDarkMode = value;
  });

  let formData = {
    name: '',
    url: '',
    type: 'http',
    interval: 60,
    timeout: 30,
    max_retries: 3,
    active: true
  };

  let isSubmitting = false;
  let errors = {};

  function validateForm() {
    errors = {};

    if (!formData.name.trim()) {
      errors.name = 'Name is required';
    }

    if (!formData.url.trim()) {
      errors.url = 'URL is required';
    } else if (!isValidUrl(formData.url)) {
      errors.url = 'Please enter a valid URL';
    }

    if (formData.interval < 10) {
      errors.interval = 'Interval must be at least 10 seconds';
    }

    if (formData.timeout < 5) {
      errors.timeout = 'Timeout must be at least 5 seconds';
    }

    return Object.keys(errors).length === 0;
  }

  function isValidUrl(url) {
    if (formData.type === 'tcp') {
      // TCP should be in format host:port or tcp://host:port
      const tcpRegex = /^(tcp:\/\/)?.+:\d+$/;
      return tcpRegex.test(url);
    }
    
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }

  async function handleSubmit() {
    if (!validateForm()) {
      return;
    }

    isSubmitting = true;

    try {
      await createMonitor(formData);
      
      // Reset form
      formData = {
        name: '',
        url: '',
        type: 'http',
        interval: 60,
        timeout: 30,
        max_retries: 3,
        active: true
      };

      dispatch('saved');
    } catch (error) {
      alert('Failed to create monitor: ' + error.message);
    } finally {
      isSubmitting = false;
    }
  }

  function getPlaceholderUrl() {
    switch (formData.type) {
      case 'http':
      case 'https':
        return 'https://example.com';
      case 'tcp':
        return 'example.com:80 or tcp://example.com:80';
      case 'ping':
        return 'ping://example.com';
      default:
        return 'https://example.com';
    }
  }
</script>

<div class="monitor-form" class:dark-mode={isDarkMode}>
  <div class="header">
    <h2>Add New Monitor</h2>
    <p>Configure a new endpoint to monitor</p>
  </div>

  <form on:submit|preventDefault={handleSubmit} class="form">
    <div class="form-group">
      <label for="name">Monitor Name *</label>
      <input
        id="name"
        type="text"
        bind:value={formData.name}
        placeholder="My Website"
        class:error={errors.name}
        required
      />
      {#if errors.name}
        <span class="error-message">{errors.name}</span>
      {/if}
    </div>

    <div class="form-group">
      <label for="type">Monitor Type</label>
      <select id="type" bind:value={formData.type}>
        <option value="http">HTTP</option>
        <option value="https">HTTPS</option>
        <option value="tcp">TCP</option>
        <option value="ping">Ping</option>
      </select>
    </div>

    <div class="form-group">
      <label for="url">URL *</label>
      <input
        id="url"
        type="url"
        bind:value={formData.url}
        placeholder={getPlaceholderUrl()}
        class:error={errors.url}
        required
      />
      {#if errors.url}
        <span class="error-message">{errors.url}</span>
      {/if}
      {#if formData.type === 'tcp'}
        <small class="help-text">For TCP monitoring, use format: host:port (e.g., example.com:80)</small>
      {/if}
    </div>

    <div class="form-row">
      <div class="form-group">
        <label for="interval">Check Interval (seconds)</label>
        <input
          id="interval"
          type="number"
          bind:value={formData.interval}
          min="10"
          class:error={errors.interval}
          required
        />
        {#if errors.interval}
          <span class="error-message">{errors.interval}</span>
        {/if}
      </div>

      <div class="form-group">
        <label for="timeout">Timeout (seconds)</label>
        <input
          id="timeout"
          type="number"
          bind:value={formData.timeout}
          min="5"
          class:error={errors.timeout}
          required
        />
        {#if errors.timeout}
          <span class="error-message">{errors.timeout}</span>
        {/if}
      </div>

      <div class="form-group">
        <label for="retries">Max Retries</label>
        <input
          id="retries"
          type="number"
          bind:value={formData.max_retries}
          min="0"
          max="10"
          required
        />
      </div>
    </div>

    <div class="form-group">
      <label class="checkbox-label">
        <input
          type="checkbox"
          bind:checked={formData.active}
        />
        <span class="checkmark"></span>
        Enable monitoring
      </label>
    </div>

    <div class="form-actions">
      <button type="submit" class="btn btn-primary" disabled={isSubmitting}>
        {isSubmitting ? 'Creating...' : 'Create Monitor'}
      </button>
    </div>
  </form>
</div>

<style>
  .monitor-form {
    max-width: 600px;
  }

  .header {
    margin-bottom: 2rem;
  }

  .header h2 {
    margin: 0 0 0.5rem 0;
    color: #333;
  }

  .dark-mode .header h2 {
    color: #e5e7eb;
  }

  .header p {
    margin: 0;
    color: #666;
  }

  .dark-mode .header p {
    color: #9ca3af;
  }

  .form {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: background-color 0.3s;
  }

  .dark-mode .form {
    background: #374151;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 1rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    color: #333;
    font-weight: 500;
  }

  .dark-mode label {
    color: #e5e7eb;
  }

  input[type="text"],
  input[type="url"],
  input[type="number"],
  select {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 4px;
    box-sizing: border-box;
    transition: border-color 0.2s, background-color 0.3s, color 0.3s;
  }

  .dark-mode input[type="text"],
  .dark-mode input[type="url"],
  .dark-mode input[type="number"],
  .dark-mode select {
    background: #1f2937;
    border-color: #4b5563;
    color: #e5e7eb;
  }

  input:focus,
  select:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }

  input.error {
    border-color: #ef4444;
  }

  .error-message {
    color: #ef4444;
    font-size: 0.875rem;
    margin-top: 0.25rem;
    display: block;
  }

  .help-text {
    color: #666;
    font-size: 0.875rem;
    margin-top: 0.25rem;
    display: block;
  }

  .dark-mode .help-text {
    color: #9ca3af;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    cursor: pointer;
    margin-bottom: 0;
  }

  .checkbox-label input[type="checkbox"] {
    margin-right: 0.75rem;
    width: auto;
  }

  .form-actions {
    padding-top: 1rem;
    border-top: 1px solid #e5e7eb;
  }

  .dark-mode .form-actions {
    border-top-color: #4b5563;
  }

  .btn {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
    transition: background-color 0.2s;
  }

  .btn-primary {
    background: #667eea;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #5a67d8;
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  @media (max-width: 768px) {
    .form-row {
      grid-template-columns: 1fr;
    }
  }
</style>