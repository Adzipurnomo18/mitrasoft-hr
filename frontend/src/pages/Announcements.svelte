<script>
  import Header from "../components/Header.svelte";
  import { API_BASE } from "../api.js";
  import "../styles/dashboard.css";
  import { onMount } from "svelte";

  let items = [];
  let loading = false;
  let error = "";
  let search = "";
  let showModal = false;
  let editing = null;
  let form = { title: "", content: "" };

  $: filtered = items
    .filter((a) =>
      search
        ? (a.title || "").toLowerCase().includes(search.toLowerCase()) ||
          (a.content || "").toLowerCase().includes(search.toLowerCase())
        : true,
    )
    .sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

  async function loadData() {
    loading = true;
    error = "";
    try {
      const res = await fetch(`${API_BASE}/api/announcements`, {
        credentials: "include",
      });
      if (!res.ok) throw new Error("Failed to fetch announcements");
      const serverItems = await res.json();
      const localItems = JSON.parse(localStorage.getItem("announcements_local") || "[]");
      const map = new Map();
      [...serverItems, ...localItems].forEach((a) => {
        const id = a.id || a.local_id;
        const prev = map.get(id);
        if (!prev || (a.updated_at && (!prev.updated_at || a.updated_at > prev.updated_at))) {
          map.set(id, a);
        }
      });
      items = Array.from(map.values());
    } catch (e) {
      error = e?.message || "Failed to fetch announcements";
      // Fallback to local storage only
      items = JSON.parse(localStorage.getItem("announcements_local") || "[]");
    } finally {
      loading = false;
    }
  }
  onMount(loadData);

  function openCreate() {
    editing = null;
    form = { title: "", content: "" };
    showModal = true;
  }
  function openEdit(a) {
    editing = a;
    form = { title: a.title || "", content: a.content || "" };
    showModal = true;
  }
  async function deleteAnnouncement(a) {
    // Try server delete (soft deactivate); fallback remove from local
    try {
      const res = await fetch(`${API_BASE}/api/announcements/${a.id}`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!res.ok) throw new Error("server delete failed");
      await loadData();
    } catch (_) {
      const local = JSON.parse(localStorage.getItem("announcements_local") || "[]");
      const next = local.filter((x) => (x.id || x.local_id) !== (a.id || a.local_id));
      localStorage.setItem("announcements_local", JSON.stringify(next));
      await loadData();
    }
  }
  async function saveAnnouncement() {
    try {
      const method = editing ? "PUT" : "POST";
      const url = editing ? `${API_BASE}/api/announcements/${editing.id}` : `${API_BASE}/api/announcements`;
      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(form),
      });
      if (res.ok) {
        await loadData();
        showModal = false;
      } else {
        // Fallback: store locally so UI tetap bekerja
        const now = new Date().toISOString();
        const local = JSON.parse(localStorage.getItem("announcements_local") || "[]");
        if (editing && editing.id) {
          const idx = local.findIndex((a) => a.id === editing.id || a.local_id === editing.id);
          if (idx >= 0) {
            local[idx] = { ...local[idx], title: form.title, content: form.content, updated_at: now };
          } else {
            local.push({ local_id: `local-${Date.now()}`, title: form.title, content: form.content, created_at: now, updated_at: now });
          }
        } else {
          local.push({ local_id: `local-${Date.now()}`, title: form.title, content: form.content, created_at: now, updated_at: now });
        }
        localStorage.setItem("announcements_local", JSON.stringify(local));
        await loadData();
        showModal = false;
      }
    } catch (e) {
      // Fallback jika jaringan error
      const now = new Date().toISOString();
      const local = JSON.parse(localStorage.getItem("announcements_local") || "[]");
      local.push({ local_id: `local-${Date.now()}`, title: form.title, content: form.content, created_at: now, updated_at: now });
      localStorage.setItem("announcements_local", JSON.stringify(local));
      await loadData();
      showModal = false;
    }
  }
</script>

<div class="page-wrapper">
  <Header>
    <div>
      <h1>Announcements</h1>
      <p>Informasi dan pengumuman terbaru perusahaan</p>
    </div>
  </Header>

  <div class="page-content">
    <section class="employees-section">
      <div class="employees-toolbar">
        <div class="employees-toolbar-text">
          <h2>Latest Announcements</h2>
          <p class="muted small">Cari pengumuman berdasarkan judul atau isi.</p>
        </div>
        <div class="employees-toolbar-actions">
          <input
            type="text"
            class="input-search"
            placeholder="Search announcements…"
            bind:value={search}
            style="margin-bottom: 12px;"
          />
        </div>
      </div>

      <article class="card employees-card">
        {#if loading}
          <div class="state muted">Loading…</div>
        {:else if error}
          <div class="state error">{error}</div>
        {:else if filtered.length === 0}
          <div class="state muted">No announcements found.</div>
        {:else}
          <div class="toolbar-actions" style="display:flex; justify-content:flex-end; margin-bottom:12px;">
            <button class="btn-primary" on:click={openCreate}>+ New Announcement</button>
          </div>
          <ul class="announcement-list">
            {#each filtered as a}
              <li class="ann-item">
                <h3 class="ann-title">{a.title}</h3>
                <p class="ann-content">{a.content}</p>
                <span class="ann-date"
                  >{new Date(a.created_at).toLocaleString()}</span
                >
                <div style="margin-top:8px;">
                  <button class="btn" on:click={() => openEdit(a)}>Edit</button>
                  <button class="btn danger" on:click={() => deleteAnnouncement(a)} style="margin-left:8px;">Delete</button>
                </div>
              </li>
            {/each}
          </ul>
        {/if}
      </article>
    </section>
  </div>
</div>

{#if showModal}
  <div class="modal-overlay" on:click={() => (showModal = false)}>
    <div class="modal" role="dialog" aria-modal="true" on:click|stopPropagation>
      <header class="modal-header">
        <h2>{editing ? "Edit Announcement" : "New Announcement"}</h2>
      </header>
      <div class="modal-body">
        <div class="form-row">
          <label>
            Title
            <input type="text" class="form-control" bind:value={form.title} />
          </label>
        </div>
        <div class="form-row">
          <label>
            Content
            <textarea rows="4" class="form-control" bind:value={form.content}></textarea>
          </label>
        </div>
      </div>
      <footer class="modal-footer">
        <button class="btn" on:click={() => (showModal = false)}>Cancel</button>
        <button class="btn-primary" on:click={saveAnnouncement}>Save</button>
      </footer>
    </div>
  </div>
{/if}
