<script>
  import Header from "../components/Header.svelte";
  import { API_BASE } from "../api.js";
  import "../styles/dashboard.css";
  import { onMount } from "svelte";

  let summary = { total: 0, pending: 0, approved: 0, rejected: 0 };
  let attSummary = { present: 0, on_time: 0, late: 0, absent: 0, working_days: 0 };
  let loading = false;
  let error = "";
  let month = new Date().toISOString().slice(0, 7); // YYYY-MM
  let items = [];
  let attDaily = [];
  let monthOptions = [];

  function buildMonthOptions(n) {
    const opts = [];
    const now = new Date();
    for (let i = 0; i < n; i++) {
      const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
      const value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
      const label = d.toLocaleDateString("id-ID", { month: "long", year: "numeric" });
      opts.push({ value, label });
    }
    return opts;
  }

  async function loadSummary() {
    loading = true;
    error = "";
    try {
      const res = await fetch(`${API_BASE}/api/requests/summary?month=${encodeURIComponent(month)}`, {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        summary = {
          total: Number(data.total || 0),
          pending: Number(data.pending || 0),
          approved: Number(data.approved || 0),
          rejected: Number(data.rejected || 0),
        };
        // Attendance summary for the same month
        const [y, m] = month.split("-");
        const from = `${y}-${m}-01`;
        const to = new Date(Number(y), Number(m), 1);
        const toStr = `${to.getFullYear()}-${String(to.getMonth()+1).padStart(2,"0")}-01`;
        const attRes = await fetch(`${API_BASE}/api/attendance/summary?from=${encodeURIComponent(from)}&to=${encodeURIComponent(toStr)}`, { credentials: "include" });
        if (attRes.ok) {
          const s = await attRes.json();
          attSummary = {
            present: Number(s.present || 0),
            on_time: Number(s.on_time || 0),
            late: Number(s.late || 0),
            absent: Number(s.absent || 0),
            working_days: Number(s.working_days || 0),
          };
        }
        const attListRes = await fetch(`${API_BASE}/api/attendance/list?from=${encodeURIComponent(from)}&to=${encodeURIComponent(toStr)}`, { credentials: "include" });
        attDaily = attListRes.ok ? await attListRes.json() : [];
      } else {
        error = "Failed to load summary";
      }
    } catch (e) {
      error = e?.message || "Failed to load summary";
    } finally {
      loading = false;
    }
  }
  async function loadItemsByMonth() {
    try {
      const res = await fetch(`${API_BASE}/api/requests/processed?month=${encodeURIComponent(month)}`, {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        items = Array.isArray(data) ? data : [];
      } else {
        items = [];
      }
    } catch (_) {
      items = [];
    }
  }
  let changing = false;
  async function handleMonthChange(e) {
    if (e && typeof e.stopPropagation === "function") e.stopPropagation();
    if (e && typeof e.preventDefault === "function") e.preventDefault();
    const next = e?.target?.value || month;
    if (changing) return;
    changing = true;
    month = next;
    try {
      await loadSummary();
      await loadItemsByMonth();
      const sb = document.querySelector(".sidebar");
      if (sb) sb.style.pointerEvents = "auto";
      const lm = document.querySelector(".logout-modal-backdrop");
      if (lm) lm.remove();
      // Release focus to avoid native select capturing inputs
      if (document.activeElement && typeof document.activeElement.blur === "function") {
        document.activeElement.blur();
      }
    } finally {
      changing = false;
    }
  }
  async function exportExcel() {
    const url = `${API_BASE}/api/requests/processed/export?month=${encodeURIComponent(month)}`;
    const a = document.createElement("a");
    a.href = url;
    a.download = `requests_${month}.xlsx`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }
  onMount(async () => {
    monthOptions = buildMonthOptions(12);
    await Promise.all([loadSummary(), loadItemsByMonth()]);
  });
</script>

<div class="page-wrapper">
  <Header>
    <div>
      <h1>Reports</h1>
      <p>Ringkasan status request karyawan</p>
    </div>
  </Header>

  <div class="page-content">
    <article class="card">
      {#if loading}
        <div class="state muted">Loading report…</div>
      {:else if error}
        <div class="state error">{error}</div>
      {:else}
        <div class="toolbar" style="display:flex; justify-content:space-between; align-items:center; margin-bottom:12px;">
          <div>
            <label style="font-size:0.85rem; color:var(--text-muted);">Filter Bulan</label>
            <select bind:value={month} on:change={handleMonthChange} class="form-control" style="margin-left:8px; min-width:180px;">
              {#each monthOptions as opt}
                <option value={opt.value}>{opt.label}</option>
              {/each}
            </select>
          </div>
          <div>
            <button class="btn" on:click={exportExcel}>Download Excel</button>
          </div>
        </div>
        <h3 style="margin-top:16px;">Attendance Daily</h3>
        <div class="table-wrap" style="margin-top:8px;">
          <table class="employees-table">
            <thead>
              <tr>
                <th>Date</th>
                <th>Status</th>
                <th>Check-in</th>
                <th>Check-out</th>
              </tr>
            </thead>
            <tbody>
              {#if !attDaily || attDaily.length === 0}
                <tr><td colspan="4" style="text-align:center; color:var(--text-muted);">Tidak ada data</td></tr>
              {:else}
                {#each attDaily as d}
                  <tr>
                    <td>{new Date(d.date).toLocaleDateString()}</td>
                    <td>{d.status}</td>
                    <td>{d.checkin_time ? new Date(d.checkin_time).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) : "-"}</td>
                    <td>{d.checkout_time ? new Date(d.checkout_time).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) : "-"}</td>
                  </tr>
                {/each}
              {/if}
            </tbody>
          </table>
        </div>
        <div class="report-grid">
          <div class="report-card">
            <p class="report-label">Total Requests</p>
            <p class="report-value">{summary.total}</p>
          </div>
          <div class="report-card">
            <p class="report-label">Pending</p>
            <p class="report-value">{summary.pending}</p>
          </div>
          <div class="report-card">
            <p class="report-label">Approved</p>
            <p class="report-value">{summary.approved}</p>
          </div>
          <div class="report-card">
            <p class="report-label">Rejected</p>
            <p class="report-value">{summary.rejected}</p>
          </div>
        </div>
        <h3 style="margin-top:16px;">Attendance Summary</h3>
        <div class="report-grid">
          <div class="report-card">
            <p class="report-label">Working Days</p>
            <p class="report-value">{attSummary.working_days}</p>
          </div>
          <div class="report-card">
            <p class="report-label">Present</p>
            <p class="report-value">{attSummary.present}</p>
          </div>
          <div class="report-card">
            <p class="report-label">On Time</p>
            <p class="report-value">{attSummary.on_time}</p>
          </div>
          <div class="report-card">
            <p class="report-label">Late</p>
            <p class="report-value">{attSummary.late}</p>
          </div>
        </div>
        <div class="table-wrap" style="margin-top:16px;">
          <table class="employees-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Employee</th>
                <th>Type</th>
                <th>Start</th>
                <th>End</th>
                <th>Status</th>
                <th>Approver</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              {#if items.length === 0}
                <tr><td colspan="8" style="text-align:center; color:var(--text-muted);">Tidak ada data</td></tr>
              {:else}
                {#each items as r}
                  <tr>
                    <td>{r.id}</td>
                    <td>{r.user_name}</td>
                    <td>{r.type}</td>
                    <td>{new Date(r.start_date).toLocaleDateString()}</td>
                    <td>{new Date(r.end_date).toLocaleDateString()}</td>
                    <td>{r.status}</td>
                    <td>{r.approver_name}</td>
                    <td>{new Date(r.updated_at).toLocaleString()}</td>
                  </tr>
                {/each}
              {/if}
            </tbody>
          </table>
        </div>
      {/if}
    </article>
  </div>
</div>

<style>
  .report-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 16px;
  }
  .report-card {
    border: 1px solid var(--border-color);
    border-radius: 16px;
    padding: 16px;
    background: var(--bg-card);
  }
  .report-label {
    margin: 0 0 6px;
    color: var(--text-muted);
    font-size: 0.85rem;
  }
  .report-value {
    margin: 0;
    font-size: 1.6rem;
    color: var(--text-primary);
    font-weight: 700;
  }
</style>
