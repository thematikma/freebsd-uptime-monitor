<script>
  import { onMount } from 'svelte';

  let channels = [];
  let monitors = [];
  let services = [];
  let availableEvents = [];
  let showForm = false;
  let showMonitorAssign = false;
  let editingChannel = null;
  let selectedChannel = null;
  let testing = false;
  let validating = false;
  let validationResult = null;

  let formData = {
    name: '',
    shoutrrr_url: '',
    events: ['monitor_up', 'monitor_down', 'recovery'],
    enabled: true
  };

  let assignedMonitors = [];

  onMount(async () => {
    await Promise.all([
      loadChannels(),
      loadMonitors(),
      loadServices(),
      loadEvents()
    ]);
  });

  async function loadChannels() {
    try {
      const response = await fetch('/api/v1/notifications/channels');
      channels = await response.json();
    } catch (error) {
      console.error('Failed to load notification channels:', error);
    }
  }

  async function loadMonitors() {
    try {
      const response = await fetch('/api/v1/monitors');
      monitors = await response.json();
    } catch (error) {
      console.error('Failed to load monitors:', error);
    }
  }

  async function loadServices() {
    try {
      const response = await fetch('/api/v1/notifications/services');
      services = await response.json();
    } catch (error) {
      console.error('Failed to load services:', error);
    }
  }

  async function loadEvents() {
    try {
      const response = await fetch('/api/v1/notifications/events');
      availableEvents = await response.json();
    } catch (error) {
      console.error('Failed to load events:', error);
    }
  }

  function openForm(channel = null) {
    validationResult = null;
    if (channel) {
      editingChannel = channel;
      let events = ['monitor_up', 'monitor_down', 'recovery'];
      try {
        if (channel.events) {
          events = JSON.parse(channel.events);
        }
      } catch (e) {}
      formData = {
        name: channel.name,
        shoutrrr_url: channel.shoutrrr_url,
        events: events,
        enabled: channel.enabled
      };
    } else {
      editingChannel = null;
      formData = {
        name: '',
        shoutrrr_url: '',
        events: ['monitor_up', 'monitor_down', 'recovery'],
        enabled: true
      };
    }
    showForm = true;
  }

  function closeForm() {
    showForm = false;
    editingChannel = null;
    validationResult = null;
  }

  function handleKeydown(event, closeFunc) {
    if (event.key === 'Escape') {
      closeFunc();
    }
  }

  async function validateURL() {
    if (!formData.shoutrrr_url) return;
    
    validating = true;
    validationResult = null;
    
    try {
      const response = await fetch('/api/v1/notifications/validate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: formData.shoutrrr_url })
      });
      
      const result = await response.json();
      validationResult = result;
    } catch (error) {
      validationResult = { valid: false, error: error.message };
    } finally {
      validating = false;
    }
  }

  async function testURL() {
    if (!formData.shoutrrr_url) return;
    
    testing = true;
    
    try {
      const response = await fetch('/api/v1/notifications/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: formData.shoutrrr_url })
      });
      
      const result = await response.json();
      if (result.success) {
        alert('✅ Test notification sent successfully!');
      } else {
        alert('❌ Test failed: ' + result.error);
      }
    } catch (error) {
      alert('❌ Test failed: ' + error.message);
    } finally {
      testing = false;
    }
  }

  async function saveChannel() {
    const payload = {
      name: formData.name,
      shoutrrr_url: formData.shoutrrr_url,
      events: formData.events,
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
        const error = await response.json();
        throw new Error(error.error || 'Failed to save channel');
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

  async function testChannel(channel) {
    testing = true;
    try {
      const response = await fetch(`/api/v1/notifications/channels/${channel.id}/test`, {
        method: 'POST'
      });
      
      const result = await response.json();
      if (result.success) {
        alert('✅ Test notification sent successfully!');
      } else {
        alert('❌ Test failed: ' + result.error);
      }
    } catch (error) {
      alert('❌ Test failed: ' + error.message);
    } finally {
      testing = false;
    }
  }

  async function openMonitorAssign(channel) {
    selectedChannel = channel;
    
    // Load currently assigned monitors for this channel
    assignedMonitors = [];
    for (const monitor of monitors) {
      try {
        const response = await fetch(`/api/v1/monitors/${monitor.id}/notifications`);
        const monitorChannels = await response.json();
        if (monitorChannels && monitorChannels.some(c => c.id === channel.id)) {
          assignedMonitors = [...assignedMonitors, monitor.id];
        }
      } catch (e) {}
    }
    
    showMonitorAssign = true;
  }

  function closeMonitorAssign() {
    showMonitorAssign = false;
    selectedChannel = null;
    assignedMonitors = [];
  }

  function toggleMonitor(monitorId) {
    if (assignedMonitors.includes(monitorId)) {
      assignedMonitors = assignedMonitors.filter(id => id !== monitorId);
    } else {
      assignedMonitors = [...assignedMonitors, monitorId];
    }
  }

  async function saveMonitorAssignments() {
    if (!selectedChannel) return;

    try {
      // For each monitor, update its notification channels
      for (const monitor of monitors) {
        const isAssigned = assignedMonitors.includes(monitor.id);
        
        // Get current channels for this monitor
        const response = await fetch(`/api/v1/monitors/${monitor.id}/notifications`);
        const currentChannels = await response.json() || [];
        const hasChannel = currentChannels.some(c => c.id === selectedChannel.id);

        if (isAssigned && !hasChannel) {
          // Add channel to monitor
          await fetch(`/api/v1/monitors/${monitor.id}/notifications`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ channel_id: selectedChannel.id })
          });
        } else if (!isAssigned && hasChannel) {
          // Remove channel from monitor
          await fetch(`/api/v1/monitors/${monitor.id}/notifications/${selectedChannel.id}`, {
            method: 'DELETE'
          });
        }
      }

      alert('✅ Monitor assignments saved!');
      closeMonitorAssign();
    } catch (error) {
      alert('Failed to save monitor assignments: ' + error.message);
    }
  }

  function toggleEvent(eventId) {
    if (formData.events.includes(eventId)) {
      formData.events = formData.events.filter(e => e !== eventId);
    } else {
      formData.events = [...formData.events, eventId];
    }
  }

  function getServiceFromURL(url) {
    if (!url) return null;
    const match = url.match(/^([a-z]+):\/\//);
    return match ? match[1] : null;
  }

  function parseEvents(eventsJson) {
    try {
      return JSON.parse(eventsJson);
    } catch (e) {
      return [];
    }
  }

  function getEventLabel(eventId) {
    const event = availableEvents.find(e => e.id === eventId);
    return event ? event.name : eventId;
  }
</script>

<div class="notifications">
  <div class="header">
    <div>
      <h2>Notification Channels</h2>
      <p class="subtitle">Configure where to send alerts using Shoutrrr URLs</p>
    </div>
    <button class="btn btn-primary" on:click={() => openForm()}>
      + Add Channel
    </button>
  </div>

  {#if channels.length === 0}
    <div class="empty-state">
      <div class="empty-icon">🔔</div>
      <h3>No notification channels configured</h3>
      <p>Add a notification channel to get alerts when monitors change status.</p>
      <p class="hint">Supports Discord, Slack, Telegram, Email, Pushover, Gotify, Ntfy, and many more!</p>
      <button class="btn btn-primary" on:click={() => openForm()}>
        Add Your First Channel
      </button>
    </div>
  {:else}
    <div class="channels-list">
      {#each channels as channel}
        {@const service = getServiceFromURL(channel.shoutrrr_url)}
        {@const events = parseEvents(channel.events)}
        <div class="channel-card">
          <div class="channel-info">
            <div class="channel-header">
              <span class="service-badge">{service || 'unknown'}</span>
              <h3>{channel.name}</h3>
            </div>
            <div class="channel-url">
              <code>{channel.shoutrrr_url.substring(0, 50)}{channel.shoutrrr_url.length > 50 ? '...' : ''}</code>
            </div>
            <div class="channel-events">
              {#each events as event}
                <span class="event-tag">{getEventLabel(event)}</span>
              {/each}
            </div>
            <div class="channel-status">
              <span class="status-indicator" class:enabled={channel.enabled}></span>
              <span>{channel.enabled ? 'Enabled' : 'Disabled'}</span>
            </div>
          </div>
          <div class="channel-actions">
            <button class="btn btn-sm btn-secondary" on:click={() => openMonitorAssign(channel)} title="Assign to monitors">
              📋 Monitors
            </button>
            <button class="btn btn-sm btn-secondary" on:click={() => testChannel(channel)} disabled={testing} title="Send test notification">
              {testing ? '...' : '🧪'} Test
            </button>
            <button class="btn btn-sm btn-secondary" on:click={() => openForm(channel)} title="Edit channel">
              ✏️ Edit
            </button>
            <button class="btn btn-sm btn-danger" on:click={() => deleteChannel(channel.id)} title="Delete channel">
              🗑️
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Services Reference -->
  <div class="services-section">
    <h3>Supported Services</h3>
    <p>Click on a service to see the URL format:</p>
    <div class="services-grid">
      {#each services as service}
        <div class="service-card" title={service.description}>
          <div class="service-name">{service.name}</div>
          <code class="service-format">{service.url_format}</code>
        </div>
      {/each}
    </div>
  </div>
</div>

<!-- Add/Edit Channel Modal -->
{#if showForm}
  <div 
    class="modal-overlay" 
    role="button"
    tabindex="0"
    aria-label="Close modal"
    on:click={closeForm}
    on:keydown={(e) => handleKeydown(e, closeForm)}
  >
    <div class="modal modal-lg" role="dialog" aria-modal="true" tabindex="-1" on:click|stopPropagation on:keydown|stopPropagation>
      <div class="modal-header">
        <h3>{editingChannel ? 'Edit' : 'Add'} Notification Channel</h3>
        <button class="close-btn" on:click={closeForm}>×</button>
      </div>
      
      <form on:submit|preventDefault={saveChannel} class="form">
        <div class="form-group">
          <label for="name">Channel Name</label>
          <input
            id="name"
            type="text"
            bind:value={formData.name}
            placeholder="e.g., Discord Alerts, Team Slack"
            required
          />
        </div>

        <div class="form-group">
          <label for="shoutrrr_url">Shoutrrr URL</label>
          <div class="url-input-group">
            <input
              id="shoutrrr_url"
              type="text"
              bind:value={formData.shoutrrr_url}
              placeholder="discord://token@id or slack://token@channel"
              required
              on:blur={validateURL}
            />
            <button type="button" class="btn btn-sm btn-secondary" on:click={validateURL} disabled={validating}>
              {validating ? '...' : 'Validate'}
            </button>
            <button type="button" class="btn btn-sm btn-secondary" on:click={testURL} disabled={testing || !formData.shoutrrr_url}>
              {testing ? '...' : 'Test'}
            </button>
          </div>
          {#if validationResult}
            <div class="validation-result" class:valid={validationResult.valid} class:invalid={!validationResult.valid}>
              {validationResult.valid ? '✅ Valid URL' : '❌ ' + validationResult.error}
            </div>
          {/if}
          <small class="help-text">
            Enter a Shoutrrr URL for your notification service. See supported services below.
          </small>
        </div>

        <div class="form-group">
          <span class="form-label">Trigger Events</span>
          <p class="help-text" style="margin-top: 0;">Select which events should send notifications:</p>
          <div class="events-grid">
            {#each availableEvents as event}
              <label class="event-checkbox" class:selected={formData.events.includes(event.id)}>
                <input
                  type="checkbox"
                  checked={formData.events.includes(event.id)}
                  on:change={() => toggleEvent(event.id)}
                />
                <div class="event-info">
                  <span class="event-name">{event.name}</span>
                  <span class="event-desc">{event.description}</span>
                </div>
              </label>
            {/each}
          </div>
        </div>

        <div class="form-group">
          <label class="checkbox-label">
            <input
              type="checkbox"
              bind:checked={formData.enabled}
            />
            <span>Enable this notification channel</span>
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

<!-- Monitor Assignment Modal -->
{#if showMonitorAssign && selectedChannel}
  <div 
    class="modal-overlay" 
    role="button"
    tabindex="0"
    aria-label="Close modal"
    on:click={closeMonitorAssign}
    on:keydown={(e) => handleKeydown(e, closeMonitorAssign)}
  >
    <div class="modal" role="dialog" aria-modal="true" tabindex="-1" on:click|stopPropagation on:keydown|stopPropagation>
      <div class="modal-header">
        <h3>Assign Monitors to "{selectedChannel.name}"</h3>
        <button class="close-btn" on:click={closeMonitorAssign}>×</button>
      </div>
      
      <div class="form">
        <p class="help-text">Select which monitors should send notifications to this channel:</p>
        
        {#if monitors.length === 0}
          <div class="empty-monitors">
            <p>No monitors configured yet.</p>
          </div>
        {:else}
          <div class="monitors-list">
            {#each monitors as monitor}
              <label class="monitor-checkbox" class:selected={assignedMonitors.includes(monitor.id)}>
                <input
                  type="checkbox"
                  checked={assignedMonitors.includes(monitor.id)}
                  on:change={() => toggleMonitor(monitor.id)}
                />
                <div class="monitor-info">
                  <span class="monitor-name">{monitor.name}</span>
                  <span class="monitor-url">{monitor.url}</span>
                </div>
                <span class="monitor-status" class:up={monitor.current_status === 'up'} class:down={monitor.current_status === 'down'}>
                  {monitor.current_status || 'unknown'}
                </span>
              </label>
            {/each}
          </div>
        {/if}

        <div class="form-actions">
          <button type="button" class="btn btn-secondary" on:click={closeMonitorAssign}>Cancel</button>
          <button type="button" class="btn btn-primary" on:click={saveMonitorAssignments}>
            Save Assignments
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .notifications {
    max-width: 1000px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 2rem;
  }

  .header h2 {
    margin: 0 0 0.25rem 0;
    color: #333;
  }

  .subtitle {
    margin: 0;
    color: #666;
    font-size: 0.875rem;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .empty-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
  }

  .empty-state h3 {
    margin: 0 0 0.5rem 0;
    color: #333;
  }

  .empty-state p {
    margin: 0.5rem 0;
    color: #666;
  }

  .hint {
    font-size: 0.875rem;
    color: #888;
  }

  .channels-list {
    display: grid;
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .channel-card {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
  }

  .channel-info {
    flex: 1;
    min-width: 0;
  }

  .channel-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.5rem;
  }

  .channel-header h3 {
    margin: 0;
    color: #333;
  }

  .service-badge {
    background: #667eea;
    color: white;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
  }

  .channel-url {
    margin-bottom: 0.5rem;
  }

  .channel-url code {
    background: #f3f4f6;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    color: #666;
    word-break: break-all;
  }

  .channel-events {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
    margin-bottom: 0.5rem;
  }

  .event-tag {
    background: #e0f2fe;
    color: #0369a1;
    padding: 0.125rem 0.5rem;
    border-radius: 12px;
    font-size: 0.75rem;
  }

  .channel-status {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: #666;
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
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  /* Services Section */
  .services-section {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .services-section h3 {
    margin: 0 0 0.5rem 0;
    color: #333;
  }

  .services-section > p {
    margin: 0 0 1rem 0;
    color: #666;
    font-size: 0.875rem;
  }

  .services-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 0.75rem;
  }

  .service-card {
    background: #f9fafb;
    padding: 0.75rem;
    border-radius: 6px;
    border: 1px solid #e5e7eb;
    cursor: help;
  }

  .service-name {
    font-weight: 600;
    color: #333;
    font-size: 0.875rem;
    margin-bottom: 0.25rem;
  }

  .service-format {
    background: #e5e7eb;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.7rem;
    color: #666;
    display: block;
    word-break: break-all;
  }

  /* Modal Styles */
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
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    width: 90%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
  }

  .modal-lg {
    max-width: 650px;
  }

  .modal-header {
    padding: 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #e5e7eb;
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
    padding: 1.5rem;
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

  .form-group input[type="text"] {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 4px;
    box-sizing: border-box;
    font-size: 0.875rem;
  }

  .form-label {
    display: block;
    margin-bottom: 0.5rem;
    color: #333;
    font-weight: 500;
  }

  .url-input-group {
    display: flex;
    gap: 0.5rem;
  }

  .url-input-group input {
    flex: 1;
  }

  .validation-result {
    margin-top: 0.5rem;
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.875rem;
  }

  .validation-result.valid {
    background: #d1fae5;
    color: #065f46;
  }

  .validation-result.invalid {
    background: #fee2e2;
    color: #991b1b;
  }

  .help-text {
    color: #666;
    font-size: 0.8rem;
    margin-top: 0.25rem;
    display: block;
  }

  /* Events Grid */
  .events-grid {
    display: grid;
    gap: 0.5rem;
  }

  .event-checkbox {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.75rem;
    background: #f9fafb;
    border: 2px solid #e5e7eb;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .event-checkbox:hover {
    border-color: #d1d5db;
  }

  .event-checkbox.selected {
    background: #eff6ff;
    border-color: #667eea;
  }

  .event-checkbox input {
    margin-top: 2px;
  }

  .event-info {
    display: flex;
    flex-direction: column;
  }

  .event-name {
    font-weight: 500;
    color: #333;
  }

  .event-desc {
    font-size: 0.8rem;
    color: #666;
  }

  /* Monitor Assignment */
  .monitors-list {
    max-height: 400px;
    overflow-y: auto;
    display: grid;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .monitor-checkbox {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    background: #f9fafb;
    border: 2px solid #e5e7eb;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .monitor-checkbox:hover {
    border-color: #d1d5db;
  }

  .monitor-checkbox.selected {
    background: #eff6ff;
    border-color: #667eea;
  }

  .monitor-info {
    flex: 1;
    min-width: 0;
  }

  .monitor-name {
    font-weight: 500;
    color: #333;
    display: block;
  }

  .monitor-url {
    font-size: 0.8rem;
    color: #666;
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .monitor-status {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 500;
    background: #f3f4f6;
    color: #666;
  }

  .monitor-status.up {
    background: #d1fae5;
    color: #065f46;
  }

  .monitor-status.down {
    background: #fee2e2;
    color: #991b1b;
  }

  .empty-monitors {
    text-align: center;
    padding: 2rem;
    color: #666;
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
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid #e5e7eb;
  }

  /* Buttons */
  .btn {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
    transition: background-color 0.2s;
    font-size: 0.875rem;
  }

  .btn-sm {
    padding: 0.5rem 0.75rem;
    font-size: 0.8rem;
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

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>