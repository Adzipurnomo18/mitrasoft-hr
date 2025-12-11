<script>
  import { createEventDispatcher } from "svelte";
  import { API_BASE } from "../api.js";
  import "../styles/login.css";

  const dispatch = createEventDispatcher();

  let email = "";
  let password = "";
  let error = "";
  let loading = false;

  async function handleSubmit() {
    error = "";
    loading = true;

    try {
      const res = await fetch(`${API_BASE}/api/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        let message = "Login failed";
        try {
          const body = await res.json();
          if (body && body.message) message = body.message;
        } catch (_) {}
        throw new Error(message);
      }

      const data = await res.json();
      dispatch("loggedIn", data);
    } catch (e) {
      error = e.message || "Login failed";
    } finally {
      loading = false;
    }
  }
</script>

<div class="login-screen">
  <div class="login-panel">
    <div class="login-brand">
      <div class="brand-dot"></div>
      <span class="brand-name">HR Portal</span>
      <span class="brand-tag">Employee Experience</span>
    </div>

    <div class="login-card">
      <h1 class="login-title">Employee Portal Login</h1>
      <p class="login-subtitle">
        Masuk untuk mengelola kehadiran, cuti, dan aktivitas karyawan.
      </p>

      {#if error}
        <div class="login-alert">{error}</div>
      {/if}

      <form class="login-form" on:submit|preventDefault={handleSubmit}>
        <label class="field">
          <span class="field-label">Email</span>
          <input
            type="email"
            bind:value={email}
            required
            placeholder="you@example.com"
          />
        </label>

        <label class="field">
          <span class="field-label">Password</span>
          <input
            type="password"
            bind:value={password}
            required
            placeholder="••••••••"
          />
        </label>

        <button class="btn-primary" disabled={loading}>
          {#if loading}
            Sedang masuk...
          {:else}
            Login
          {/if}
        </button>
      </form>

      <p class="login-footnote">
        PT Mitrasoft • Secure access &amp; audit ready
      </p>
    </div>
  </div>

  <div class="login-illustration">
    <div class="orb orb-1"></div>
    <div class="orb orb-2"></div>
    <div class="login-stats-card">
      <p class="stats-kicker">Realtime Overview</p>
      <p class="stats-title">Kehadiran &amp; Aktivitas Karyawan</p>
      <div class="stats-row">
        <div>
          <p class="stats-label">On Time</p>
          <p class="stats-value">92%</p>
        </div>
        <div>
          <p class="stats-label">Overtime</p>
          <p class="stats-value">18</p>
        </div>
        <div>
          <p class="stats-label">Leave Today</p>
          <p class="stats-value">6</p>
        </div>
      </div>
    </div>
  </div>
</div>
