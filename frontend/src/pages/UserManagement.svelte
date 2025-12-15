<script>
    import Header from "../components/Header.svelte";
    import { API_BASE } from "../api.js";
    import "../styles/dashboard.css";
    import { onMount } from "svelte";
    let users = [];
    let loading = false;
    let error = "";
    let search = "";

    $: filtered = users.filter((u) => {
        const q = search.toLowerCase().trim();
        if (!q) return true;
        return (
            (u.name || "").toLowerCase().includes(q) ||
            (u.email || "").toLowerCase().includes(q) ||
            (u.employee_code || "").toLowerCase().includes(q)
        );
    });

    async function loadUsers() {
        loading = true;
        error = "";
        try {
            const res = await fetch(`${API_BASE}/api/employees`, { credentials: "include" });
            if (!res.ok) throw new Error("Failed to fetch users");
            const data = await res.json();
            users = Array.isArray(data) ? data : [];
        } catch (e) {
            error = e?.message || "Failed to fetch users";
        } finally {
            loading = false;
        }
    }
    onMount(loadUsers);

    async function updateRoles(user, roles) {
        try {
            const body = {
                employee_code: user.employee_code,
                name: user.name,
                email: user.email,
                branch: user.branch,
                job_title: user.job_title,
                status: user.status,
                department: user.department,
                roles,
            };
            const res = await fetch(`${API_BASE}/api/employees/${user.id}`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify(body),
            });
            if (!res.ok) {
                const msg = await res.text();
                throw new Error(msg || "Failed to update roles");
            }
            await loadUsers();
        } catch (e) {
            alert(e?.message || "Failed to update roles");
        }
    }
    const possibleRoles = ["ADMIN", "HRD", "IT", "EMPLOYEE"];
</script>

<div class="page-wrapper">
    <Header>
        <div>
            <h1>User Management</h1>
            <p>Kelola peran pengguna</p>
        </div>
    </Header>
    <div class="page-content">
        <section class="employees-section">
            <div class="employees-toolbar">
                <div class="employees-toolbar-text">
                    <h2>Users</h2>
                    <p class="muted small">Kelola peran dan hak akses pengguna.</p>
                </div>
                <div class="employees-toolbar-actions">
                    <input
                        type="text"
                        class="input-search"
                        placeholder="Cari nama, email, atau kode…"
                        bind:value={search}
                    />
                </div>
            </div>
        {#if loading}
            <div class="state muted">Loading…</div>
        {:else if error}
            <div class="state error">{error}</div>
        {:else}
            <article class="card employees-card" style="margin-top:16px;">
                <table class="table">
                    <thead>
                        <tr>
                            <th>Code</th>
                            <th>Name</th>
                            <th>Email</th>
                            <th>Roles</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each filtered as u}
                            <tr>
                                <td>{u.employee_code}</td>
                                <td>{u.name}</td>
                                <td>{u.email}</td>
                                <td>{Array.isArray(u.roles) ? u.roles.join(", ") : "-"}</td>
                                <td class="actions">
                                    {#each possibleRoles as r}
                                        <label class="small" style="margin-right:12px;">
                                            <input type="checkbox"
                                                   checked={Array.isArray(u.roles) && u.roles.includes(r)}
                                                   on:change={(e) => {
                                                        const isChecked = e.target.checked;
                                                        const next = new Set(Array.isArray(u.roles) ? u.roles : []);
                                                        if (isChecked) next.add(r); else next.delete(r);
                                                        updateRoles(u, Array.from(next));
                                                   }} />
                                            {r}
                                        </label>
                                    {/each}
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </article>
        {/if}
        </section>
    </div>
</div>
