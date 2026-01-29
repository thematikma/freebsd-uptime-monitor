<script>
  import { onMount } from 'svelte';
  import { monitors, fetchMonitors, deleteMonitor } from '$lib/stores/monitors.js';

  let monitorList = [];

  monitors.subscribe(value => {
    monitorList = value;
  });

  onMount(() => {
    fetchMonitors();
  });

  async function handleDelete(id) {
    if (confirm('Are you sure you want to delete this monitor?')) {
      try {
        await deleteMonitor(id);
      } catch (error) {
        alert('Failed to delete monitor');
      }
    }
  }

  function getStatusColor(status) {
    switch (status) {
      case 'up': return '#4ade80';
      case 'down': return '#ef4444';
      default: return '#6b7280';
    }
  }

  function formatDate(dateString) {
    return new Date(dateString).toLocaleString();
  }
</script>

<div class="monitor-list">
  <div class="header">
    <h2>Monitors</h2>
    <p>Manage your monitoring endpoints</p>
  </div>

  {#if monitorList.length === 0}
    <div class="empty-state">
      <p>No monitors configured yet.</p>
      <p>Add your first monitor to get started.</p>
    </div>
  {:else}
    <div class="monitors-grid">
      {#each monitorList as monitor}
        <div class="monitor-card">
          <div class="monitor-header">
            <div class="monitor-info">
              <h3>{monitor.name}</h3>
              <p class="monitor-url">{monitor.url}</p>
            </div>
            <div class="monitor-status">
              <span class="status-indicator" style="background-color: {getStatusColor('up')}"></span>
              <span class="status-text">Online</span>
            </div>
          </div>
          
          <div class="monitor-details">
            <div class="detail-item">
              <span class="detail-label">Type:</span>
              <span class="detail-value">{monitor.type.toUpperCase()}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Interval:</span>
              <span class="detail-value">{monitor.interval}s</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Timeout:</span>
              <span class="detail-value">{monitor.timeout}s</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Created:</span>
              <span class="detail-value">{formatDate(monitor.created_at)}</span>
            </div>
          </div>

          <div class="monitor-actions">
            <button class="btn btn-secondary" on:click={() => {}}>Edit</button>
            <button class="btn btn-danger" on:click={() => handleDelete(monitor.id)}>Delete</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .monitor-list {
    max-width: 1200px;
  }

  .header {
    margin-bottom: 2rem;
  }

  .header h2 {
    margin: 0 0 0.5rem 0;
    color: #333;
  }

  .header p {
    margin: 0;
    color: #666;
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

  .monitors-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }

  .monitor-card {
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    padding: 1.5rem;
    transition: transform 0.2s;
  }

  .monitor-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  }

  .monitor-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 1rem;
  }

  .monitor-info h3 {
    margin: 0 0 0.25rem 0;
    color: #333;
    font-size: 1.1rem;
  }

  .monitor-url {
    margin: 0;
    color: #666;
    font-size: 0.875rem;
    word-break: break-all;
  }

  .monitor-status {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .status-text {
    font-size: 0.875rem;
    color: #333;
    font-weight: 500;
  }

  .monitor-details {
    margin-bottom: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e5e7eb;
  }

  .detail-item {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.5rem;
  }

  .detail-label {
    color: #666;
    font-size: 0.875rem;
  }

  .detail-value {
    color: #333;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .monitor-actions {
    display: flex;
    gap: 0.75rem;
  }

  .btn {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: background-color 0.2s;
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