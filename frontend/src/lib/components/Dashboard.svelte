<script>
  import { onMount } from 'svelte';

  let dashboardData = {
    total_monitors: 0,
    up_monitors: 0,
    down_monitors: 0,
    avg_uptime: 0
  };

  let darkMode = false;

  onMount(async () => {
    // Load dark mode preference
    darkMode = localStorage.getItem('darkMode') === 'true';
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

  function toggleDarkMode() {
    darkMode = !darkMode;
    localStorage.setItem('darkMode', darkMode.toString());
  }

  $: uptimePercentage = Math.round(dashboardData.avg_uptime || 0);
  $: statusColor = uptimePercentage >= 95 ? '#4ade80' : uptimePercentage >= 80 ? '#fbbf24' : '#ef4444';
</script>

<div class="dashboard" class:dark-mode={darkMode}>
  <div class="dashboard-header">
    <h2>Dashboard</h2>
    <button class="theme-toggle" on:click={toggleDarkMode}>
      {darkMode ? '☀️' : '🌙'} {darkMode ? 'Light' : 'Dark'}
    </button>
  </div>
  
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
    transition: all 0.3s ease;
  }

  .dashboard.dark-mode {
    color: #e5e7eb;
  }

  .dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .theme-toggle {
    padding: 0.5rem 1rem;
    border: 1px solid #d1d5db;
    border-radius: 4px;
    background: white;
    cursor: pointer;
    transition: all 0.2s;
    font-size: 0.875rem;
  }

  .dark-mode .theme-toggle {
    background: #374151;
    border-color: #4b5563;
    color: #e5e7eb;
  }

  .theme-toggle:hover {
    background: #f3f4f6;
  }

  .dark-mode .theme-toggle:hover {
    background: #4b5563;
  }

  h2 {
    margin: 0;
    color: #333;
  }

  .dark-mode h2 {
    color: #e5e7eb;
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
    transition: all 0.3s ease;
  }

  .dark-mode .stat-card {
    background: #374151;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
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
    transition: all 0.3s ease;
  }

  .dark-mode .uptime-overview {
    background: #374151;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  }

  .uptime-overview h3 {
    margin: 0 0 1rem 0;
    color: #333;
  }

  .dark-mode .uptime-overview h3 {
    color: #e5e7eb;
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
    transition: all 0.3s ease;
  }

  .status-dot.green {
    background: #4ade80;
    animation: pulse-green 2s infinite;
  }

  .status-dot.yellow {
    background: #fbbf24;
    animation: pulse-yellow 2s infinite;
  }

  .status-dot.red {
    background: #ef4444;
    animation: pulse-red 2s infinite;
  }

  @keyframes pulse-green {
    0%, 100% {
      box-shadow: 0 0 0 0 rgba(74, 222, 128, 0.7);
    }
    50% {
      box-shadow: 0 0 0 8px rgba(74, 222, 128, 0);
    }
  }

  @keyframes pulse-yellow {
    0%, 100% {
      box-shadow: 0 0 0 0 rgba(251, 191, 36, 0.7);
    }
    50% {
      box-shadow: 0 0 0 8px rgba(251, 191, 36, 0);
    }
  }

  @keyframes pulse-red {
    0%, 100% {
      box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7);
    }
    50% {
      box-shadow: 0 0 0 8px rgba(239, 68, 68, 0);
    }
  }

  .status-text {
    color: #333;
    font-weight: 500;
  }

  .dark-mode .status-text {
    color: #e5e7eb;
  }
</style>