<script>
  import { onMount } from "svelte";
  import { API_BASE } from "../api.js";
  import { user } from "../stores.js";
  import "../styles/dashboard.css";

  let currentUser = null;
  user.subscribe((value) => (currentUser = value));

  let announcements = [];
  let loadingAnnouncements = true;

  onMount(loadAnnouncements);

  async function loadAnnouncements() {
    loadingAnnouncements = true;
    try {
      const res = await fetch(`${API_BASE}/api/announcements`, {
        credentials: "include",
      });
      if (res.ok) {
        announcements = await res.json();
        // Filter only unread or recent? For now show all active
      }
    } catch (e) {
      console.error("Failed to load announcements", e);
    } finally {
      loadingAnnouncements = false;
    }
  }

  import Header from "../components/Header.svelte";

  // ... (currentUser sub) ...
  // ... (announcements logic) ...

  const displayName = () => currentUser?.name || "Employee";
  const email = () => currentUser?.email || "";
  const roles = () =>
    Array.isArray(currentUser?.roles) && currentUser.roles.length
      ? currentUser.roles.join(", ")
      : "-";
</script>

<div class="page-wrapper">
  <Header>
    <div>
      <h1>Dashboard</h1>
      <p>Have a good day, {displayName()}!</p>
    </div>
  </Header>

  <div class="page-content">
    <div class="dashboard-grid">
      <!-- Main Content -->
      <div class="main-column">
        <section class="cards-grid">
          <article class="card">
            <h2>Your Profile</h2>
            <dl class="data-list">
              <div>
                <dt>Email</dt>
                <dd>{email()}</dd>
              </div>
              <div>
                <dt>Roles</dt>
                <dd>{roles()}</dd>
              </div>
            </dl>
          </article>

          <article class="card">
            <h2>Attendance</h2>
            <p class="muted">
              Summary kehadiran akan ditampilkan di sini. Grafik on-time, late,
              dan absence.
            </p>
          </article>

          <article class="card">
            <h2>Leave Balance</h2>
            <p class="muted">
              Annual leave, sick leave, dan sisa cuti akan ditampilkan di sini.
            </p>
          </article>
        </section>
      </div>

      <!-- Sidebar Widget -->
      <aside class="sidebar-column">
        <article class="card announcement-widget">
          <div class="widget-header">
            <h2>📢 Announcements</h2>
          </div>
          {#if loadingAnnouncements}
            <div class="widget-loading">Loading...</div>
          {:else if announcements.length === 0}
            <div class="widget-empty">No active announcements</div>
          {:else}
            <ul class="announcement-list">
              {#each announcements as a}
                <li class="ann-item">
                  <h3 class="ann-title">{a.title}</h3>
                  <p class="ann-content">{a.content}</p>
                  <span class="ann-date"
                    >{new Date(a.created_at).toLocaleDateString()}</span
                  >
                </li>
              {/each}
            </ul>
          {/if}
        </article>

        <article class="card">
          <h2>My Requests</h2>
          <p class="muted">
            Status pengajuan cuti, lembur, dan request lainnya.
          </p>
        </article>
      </aside>
    </div>
  </div>
</div>

<style>
  .dashboard-grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 24px;
  }

  @media (max-width: 1024px) {
    .dashboard-grid {
      grid-template-columns: 1fr;
    }
  }

  .announcement-widget {
    padding: 0;
    overflow: hidden;
  }

  .widget-header {
    padding: 16px;
    background: rgba(15, 23, 42, 0.6); /* Darker transparent bg */
    border-bottom: 1px solid rgba(55, 65, 81, 0.5);
  }

  .widget-header h2 {
    margin: 0;
    font-size: 16px;
    color: #f9fafb;
  }

  .announcement-list {
    list-style: none;
    padding: 0;
    margin: 0;
    max-height: 400px;
    overflow-y: auto;
  }

  .ann-item {
    padding: 16px;
    border-bottom: 1px solid rgba(55, 65, 81, 0.3);
  }

  .ann-item:last-child {
    border-bottom: none;
  }

  .ann-title {
    margin: 0 0 4px;
    font-size: 14px;
    color: #f3f4f6; /* Light text */
    font-weight: 600;
  }

  .ann-content {
    margin: 0 0 8px;
    font-size: 13px;
    color: #9ca3af; /* Gray text */
    line-height: 1.5;
  }

  .ann-date {
    font-size: 11px;
    color: #6b7280;
    display: block;
    text-align: right;
  }

  .widget-empty,
  .widget-loading {
    padding: 24px;
    text-align: center;
    color: #6b7280;
    font-size: 13px;
  }
</style>
