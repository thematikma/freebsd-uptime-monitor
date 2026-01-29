<script>
  import { onMount } from 'svelte';

  let channels = [];
  let showForm = false;
  let editingChannel = null;
  let formData = {
    name: '',
    type: 'discord',
    webhook_url: '',
    username: 'Uptime Monitor',
    avatar_url: '',
    enabled: true
  };

  onMount(() => {
    loadChannels();
  });

  async function loadChannels() {
    try {
      const response = await fetch('/api/v1/notifications/channels');
      channels = await response.json();
    } catch (error) {
      console.error('Failed to load notification channels:', error);
    }
  }

  function openForm(channel = null) {
    if (channel) {
      editingChannel = channel;
      const config = JSON.parse(channel.config);
      formData = {
        name: channel.name,
        type: channel.type,
        webhook_url: config.webhook_url || '',
        username: config.username || 'Uptime Monitor',
        avatar_url: config.avatar_url || '',
        enabled: channel.enabled
      };
    } else {
      editingChannel = null;
      formData = {
        name: '',
        type: 'discord',
        webhook_url: '',
        username: 'Uptime Monitor',
        avatar_url: '',
        enabled: true
      };
    }
    showForm = true;
  }

  function closeForm() {
    showForm = false;
    editingChannel = null;
  }

  async function saveChannel() {
    const config = {
      webhook_url: formData.webhook_url,
      username: formData.username,
      avatar_url: formData.avatar_url
    };

    const payload = {
      name: formData.name,
      type: formData.type,
      config: JSON.stringify(config),
      enabled: formData.enabled
    };

    try {
      const url = editingChannel 
        ? `/api/v1/notifications/channels/${editingChannel.id}`
        : '/api/v1/notifications/channels';
      
      const method = editingChannel ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (response.ok) {
        await loadChannels();
        closeForm();
      } else {
        throw new Error('Failed to save channel');
      }
    } catch (error) {
      alert('Failed to save notification channel: ' + error.message);
    }
  }

  async function deleteChannel(id) {
    if (!confirm('Are you sure you want to delete this notification channel?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/notifications/channels/${id}`, {
        method: 'DELETE'
      });

      if (response.ok) {
        await loadChannels();
      } else {
        throw new Error('Failed to delete channel');
      }
    } catch (error) {
      alert('Failed to delete notification channel: ' + error.message);
    }
  }

  async function testNotification(channel) {
    // This would send a test notification
    alert(`Test notification would be sent to ${channel.name}`);
  }
</script>

<div class="notifications">
  <div class="header">
    <h2>Notification Channels</h2>
    <button class="btn btn-primary" on:click={() => openForm()}>
      Add Discord Channel
    </button>
  </div>

  {#if channels.length === 0}
    <div class="empty-state">
      <p>No notification channels configured.</p>
      <p>Add a Discord webhook to get alerts when monitors go down.</p>
    </div>
  {:else}
    <div class="channels-list">
      {#each channels as channel}
        <div class="channel-card">
          <div class="channel-info">
            <h3>{channel.name}</h3>
            <p class="channel-type">{channel.type.toUpperCase()}</p>
            <div class="channel-status">
              <span class="status-indicator" class:enabled={channel.enabled}></span>
              <span>{channel.enabled ? 'Enabled' : 'Disabled'}</span>
            </div>
          </div>
          <div class="channel-actions">
            <button class="btn btn-secondary" on:click={() => testNotification(channel)}>Test</button>
            <button class="btn btn-secondary" on:click={() => openForm(channel)}>Edit</button>
            <button class="btn btn-danger" on:click={() => deleteChannel(channel.id)}>Delete</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if showForm}
  <div class="modal-overlay" on:click={closeForm}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h3>{editingChannel ? 'Edit' : 'Add'} Discord Webhook</h3>
        <button class="close-btn" on:click={closeForm}>×</button>
      </div>
      
      <form on:submit|preventDefault={saveChannel} class="form">
        <div class="form-group">
          <label for="name">Channel Name</label>
          <input
            id="name"
            type="text"
            bind:value={formData.name}
            placeholder="My Discord Channel"
            required
          />
        </div>

        <div class="form-group">
          <label for="webhook_url">Discord Webhook URL</label>
          <input
            id="webhook_url"
            type="url"
            bind:value={formData.webhook_url}
            placeholder="https://discord.com/api/webhooks/..."
            required
          />
          <small class="help-text">
            Get this from your Discord server settings → Integrations → Webhooks
          </small>
        </div>

        <div class="form-group">
          <label for="username">Bot Username (optional)</label>
          <input
            id="username"
            type="text"
            bind:value={formData.username}
            placeholder="Uptime Monitor"
          />
        </div>

        <div class="form-group">
          <label for="avatar_url">Bot Avatar URL (optional)</label>
          <input
            id="avatar_url"
            type="url"
            bind:value={formData.avatar_url}
            placeholder="https://example.com/avatar.png"
          />
        </div>

        <div class="form-group">
          <label class="checkbox-label">
            <input
              type="checkbox"
              bind:checked={formData.enabled}
            />
            <span>Enable notifications</span>
          </label>
        </div>

        <div class="form-actions">
          <button type="button" class="btn btn-secondary" on:click={closeForm}>Cancel</button>
          <button type="submit" class="btn btn-primary">
            {editingChannel ? 'Update' : 'Create'} Channel
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<style>
  .notifications {
    max-width: 800px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .header h2 {
    margin: 0;
    color: #333;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .empty-state p {
    margin: 0.5rem 0;
    color: #666;
  }

  .channels-list {
    display: grid;
    gap: 1rem;
  }

  .channel-card {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .channel-info h3 {
    margin: 0 0 0.5rem 0;
    color: #333;
  }

  .channel-type {
    margin: 0 0 0.5rem 0;
    color: #666;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .channel-status {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
  }

  .status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #6b7280;
  }

  .status-indicator.enabled {
    background: #10b981;
  }

  .channel-actions {
    display: flex;
    gap: 0.5rem;
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: white;
    padding: 0;
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    width: 90%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
  }

  .modal-header {
    padding: 1.5rem 1.5rem 0 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #e5e7eb;
    margin-bottom: 1.5rem;
  }

  .modal-header h3 {
    margin: 0;
    color: #333;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: #666;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    color: #333;
  }

  .form {
    padding: 0 1.5rem 1.5rem 1.5rem;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.5rem;
    color: #333;
    font-weight: 500;
  }

  .form-group input {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 4px;
    box-sizing: border-box;
  }

  .help-text {
    color: #666;
    font-size: 0.875rem;
    margin-top: 0.25rem;
    display: block;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
  }

  .checkbox-label input {
    width: auto;
    margin: 0;
  }

  .form-actions {
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
    margin-top: 2rem;
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

  .btn-primary:hover {
    background: #5a67d8;
  }

  .btn-secondary {
    background: #f3f4f6;
    color: #374151;
  }

  .btn-secondary:hover {
    background: #e5e7eb;
  }

  .btn-danger {
    background: #ef4444;
    color: white;
  }

  .btn-danger:hover {
    background: #dc2626;
  }
</style>