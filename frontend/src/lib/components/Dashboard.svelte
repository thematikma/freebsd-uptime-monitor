<script>
  import { onMount } from 'svelte';

  let dashboardData = {
    total_monitors: 0,
    up_monitors: 0,
    down_monitors: 0,
    avg_uptime: 0
  };

  onMount(async () => {
    await loadDashboard();
    // Refresh every 30 seconds
    setInterval(loadDashboard, 30000);
  });

  async function loadDashboard() {
    try {
      const response = await fetch('/api/v1/dashboard');
      dashboardData = await response.json();
    } catch (error) {
      console.error('Failed to load dashboard:', error);
    }
  }

  $: uptimePercentage = Math.round(dashboardData.avg_uptime || 0);
  $: statusColor = uptimePercentage >= 95 ? '#4ade80' : uptimePercentage >= 80 ? '#fbbf24' : '#ef4444';
</script>

<div class="dashboard">
  <h2>Dashboard</h2>
  
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-value">{dashboardData.total_monitors}</div>
      <div class="stat-label">Total Monitors</div>
    </div>
    
    <div class="stat-card success">
      <div class="stat-value">{dashboardData.up_monitors}</div>
      <div class="stat-label">Up</div>
    </div>
    
    <div class="stat-card error">
      <div class="stat-value">{dashboardData.down_monitors}</div>
      <div class="stat-label">Down</div>
    </div>
    
    <div class="stat-card">
      <div class="stat-value" style="color: {statusColor}">{uptimePercentage}%</div>
      <div class="stat-label">Average Uptime</div>
    </div>
  </div>

  <div class="uptime-overview">
    <h3>System Status</h3>
    <div class="status-indicator">
      <div class="status-dot" class:green={uptimePercentage >= 95} class:yellow={uptimePercentage >= 80 && uptimePercentage < 95} class:red={uptimePercentage < 80}></div>
      <span class="status-text">
        {#if uptimePercentage >= 95}
          All systems operational
        {:else if uptimePercentage >= 80}
          Some services experiencing issues
        {:else}
          Multiple services down
        {/if}
      </span>
    </div>
  </div>
</div>

<style>
  .dashboard {
    max-width: 1200px;
  }

  h2 {
    margin: 0 0 2rem 0;
    color: #333;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .stat-card {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    text-align: center;
    border-left: 4px solid #667eea;
  }

  .stat-card.success {
    border-left-color: #4ade80;
  }

  .stat-card.error {
    border-left-color: #ef4444;
  }

  .stat-value {
    font-size: 2rem;
    font-weight: bold;
    color: #333;
    margin-bottom: 0.5rem;
  }

  .stat-label {
    color: #666;
    font-size: 0.875rem;
  }

  .uptime-overview {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .uptime-overview h3 {
    margin: 0 0 1rem 0;
    color: #333;
  }

  .status-indicator {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .status-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: #666;
  }

  .status-dot.green {
    background: #4ade80;
  }

  .status-dot.yellow {
    background: #fbbf24;
  }

  .status-dot.red {
    background: #ef4444;
  }

  .status-text {
    color: #333;
    font-weight: 500;
  }
</style>