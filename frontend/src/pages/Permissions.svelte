<script>
    import Header from "../components/Header.svelte";
    import { API_BASE } from "../api.js";
    import "../styles/dashboard.css";
    import { onMount } from "svelte";
    let roles = [];
    let permissions = [];
    let loading = false;
    let error = "";
    let search = "";

    $: filteredPerms = permissions
        .filter((p) => (search ? String(p).toLowerCase().includes(search.toLowerCase()) : true))
        .sort((a, b) => String(a).localeCompare(String(b)));

    function groupPermissions(perms) {
        const groups = new Map();
        for (const p of perms) {
            const key = String(p).split("_")[0] || "GENERAL";
            if (!groups.has(key)) groups.set(key, []);
            groups.get(key).push(p);
        }
        return Array.from(groups.entries()).map(([k, v]) => ({ code: k, items: v }));
    }

    $: grouped = groupPermissions(filteredPerms);

    async function loadData() {
        loading = true;
        error = "";
        try {
            const res = await fetch(`${API_BASE}/api/me/permissions`, { credentials: "include" });
            if (!res.ok) throw new Error("Failed to fetch permissions");
            const data = await res.json();
            roles = data.roles || [];
            permissions = data.permissions || [];
        } catch (e) {
            error = e?.message || "Failed to fetch permissions";
        } finally {
            loading = false;
        }
    }
    onMount(loadData);
</script>

<div class="page-wrapper">
    <Header>
        <div>
            <h1>Permissions</h1>
            <p>Kelola dan lihat izin yang dimiliki pengguna saat ini</p>
        </div>
    </Header>
    <div class="page-content">
        {#if loading}
            <div class="state muted">Loading…</div>
        {:else if error}
            <div class="state error">{error}</div>
        {:else}
            <section class="employees-section">
                <div class="employees-toolbar">
                    <div class="employees-toolbar-text">
                        <h2>Permission Overview</h2>
                        <p class="muted small">Roles: {roles.length === 0 ? "-" : roles.join(", ")}</p>
                    </div>
                    <div class="employees-toolbar-actions">
                        <input
                            type="text"
                            class="input-search"
                            placeholder="Search permission…"
                            bind:value={search}
                        />
                    </div>
                </div>

                {#if grouped.length === 0}
                    <div class="state muted">No permissions</div>
                {:else}
                    {#each grouped as grp}
                        <article class="card employees-card" style="margin-bottom:12px;">
                            <div class="card-header">
                                <h3>{grp.code}</h3>
                            </div>
                            <div class="card-body">
                                <div class="perm-grid">
                                    {#each grp.items as p}
                                        <span class="pill-perm">{p}</span>
                                    {/each}
                                </div>
                            </div>
                        </article>
                    {/each}
                {/if}
            </section>
        {/if}
    </div>
</div>

<style>
    .card-header h3 {
        margin: 0;
        font-size: 1rem;
        color: var(--text-primary);
    }
    .perm-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
        gap: 8px;
    }
    .pill-perm {
        display: inline-flex;
        align-items: center;
        border: 1px solid var(--border-color);
        border-radius: 999px;
        padding: 6px 10px;
        font-size: 0.8rem;
        color: var(--text-body);
        background: var(--bg-hover);
    }
</style>
