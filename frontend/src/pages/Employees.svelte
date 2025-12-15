<script>
    import { onMount } from "svelte";
    import { API_BASE } from "../api.js";
    import { user } from "../stores.js";
    import "../styles/dashboard.css";
    import AddEmployee from "../components/AddEmployee.svelte";
    import Header from "../components/Header.svelte";

    export let page = "employees";

    let currentUser = null;
    user.subscribe((value) => (currentUser = value));

    // ====== DATA EMPLOYEES ======
    let employees = [];
    let loading = true;
    let error = "";
    let search = "";

    // ====== STATE MODAL (Add / Edit) ======
    let showForm = false;
    let formMode = "create";
    let editingEmployee = null;
    let saving = false;

    $: filteredEmployees = employees.filter((e) => {
        const q = search.toLowerCase().trim();
        if (!q) return true;
        return (
            (e.name || "").toLowerCase().includes(q) ||
            (e.email || "").toLowerCase().includes(q) ||
            (e.employee_code || "").toLowerCase().includes(q) ||
            (e.branch || "").toLowerCase().includes(q)
        );
    });

    async function loadEmployees() {
        loading = true;
        error = "";

        try {
            const res = await fetch(`${API_BASE}/api/employees`, {
                credentials: "include",
            });

            if (!res.ok) {
                const body = await res.json().catch(() => ({}));
                throw new Error(body.message || "Failed to fetch employees");
            }

            const data = await res.json();
            employees = Array.isArray(data) ? data : [];
        } catch (e) {
            console.error(e);
            error = e?.message || "Failed to fetch employees";
            employees = [];
        } finally {
            loading = false;
        }
    }

    onMount(loadEmployees);

    function handleAddEmployee() {
        formMode = "create";
        editingEmployee = null;
        showForm = true;
    }

    function handleEditEmployee(emp) {
        formMode = "edit";
        editingEmployee = emp;
        showForm = true;
    }

    async function handleDeleteEmployee(id) {
        const mode = prompt("Ketik 'hard' untuk delete permanen, atau Enter untuk inactive:");
        if (mode === null) return;

        try {
            let res;
            if ((mode || "").toLowerCase() === "hard") {
                res = await fetch(`${API_BASE}/api/employees/${id}/hard`, {
                    method: "DELETE",
                    credentials: "include",
                });
            } else {
                res = await fetch(`${API_BASE}/api/employees/${id}`, {
                    method: "DELETE",
                    credentials: "include",
                });
            }

            if (!res.ok) {
                // Fallback by employee_code when ID-based delete fails
                const emp = employees.find((e) => e.id === id);
                if (emp?.employee_code) {
                    if ((mode || "").toLowerCase() === "hard") {
                        res = await fetch(
                            `${API_BASE}/api/employees/by-code/${encodeURIComponent(emp.employee_code)}/hard`,
                            {
                                method: "DELETE",
                                credentials: "include",
                            },
                        );
                    } else {
                        res = await fetch(
                            `${API_BASE}/api/employees/by-code/${encodeURIComponent(emp.employee_code)}`,
                            {
                                method: "DELETE",
                                credentials: "include",
                            },
                        );
                    }
                }
                if (!res.ok) {
                    const body = await res.json().catch(() => ({}));
                    throw new Error(body.message || "Failed to delete employee");
                }
            }

            await loadEmployees();
        } catch (e) {
            alert(e?.message || "Failed to delete employee");
        }
    }

    function handleCloseForm() {
        showForm = false;
        editingEmployee = null;
    }

    async function handleSubmitEmployee(event) {
        const { mode, id, form } = event.detail;

        const url =
            mode === "create"
                ? `${API_BASE}/api/employees`
                : `${API_BASE}/api/employees/${id}`;

        const method = mode === "create" ? "POST" : "PUT";

        try {
            saving = true;

            // Auto full akses untuk dept IT/HR saat create jika belum set roles
            if (mode === "create") {
                const dept = (form.department || "").toUpperCase();
                if (!Array.isArray(form.roles) || form.roles.length === 0) {
                    if (dept === "IT" || dept === "HR") {
                        form.roles = ["ADMIN", dept === "IT" ? "IT" : "HRD"];
                    } else {
                        form.roles = ["EMPLOYEE"];
                    }
                }
            }

            const res = await fetch(url, {
                method,
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify(form),
            });

            if (!res.ok) {
                const body = await res.json().catch(() => ({}));
                throw new Error(body.message || "Failed to save employee");
            }

            showForm = false;
            editingEmployee = null;
            await loadEmployees();
        } catch (e) {
            alert(e?.message || "Failed to save employee");
        } finally {
            saving = false;
        }
    }
</script>

<div class="page-wrapper">
    <Header>
        <div>
            <h1>Employees</h1>
            <p>Manage employee profiles, roles, dan status keaktifan.</p>
        </div>
    </Header>

    <div class="page-content">
        <section class="employees-section">
            <div class="employees-toolbar">
                <div class="employees-toolbar-text">
                    <h2>Employees list</h2>
                    <p class="muted small">Data karyawan aktif & non-aktif.</p>
                </div>

                <div class="employees-toolbar-actions">
                    <input
                        type="text"
                        class="input-search"
                        placeholder="Search name, email, or code…"
                        bind:value={search}
                        style="margin-bottom:12px;"
                    />
                    <button
                        type="button"
                        class="btn-primary"
                        on:click={handleAddEmployee}
                    >
                        + Add Employee
                    </button>
                </div>
            </div>

            <article class="card employees-card">
                {#if loading}
                    <div class="state muted">Loading employees…</div>
                {:else if error}
                    <div class="state error">{error}</div>
                {:else if filteredEmployees.length === 0}
                    <div class="state muted">
                        Tidak ada data karyawan yang cocok dengan pencarian.
                    </div>
                {:else}
                    <div class="employees-table-wrapper">
                        <table class="employees-table">
                            <thead>
                                <tr>
                                    <th>Code</th>
                                    <th>Name</th>
                                    <th>Email</th>
                                    <th>Department</th>
                                    <th>Job Title</th>
                                    <th>Branch</th>
                                    <th>Status</th>
                                    <th style="width: 1%"></th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filteredEmployees as e}
                                    <tr>
                                        <td class="mono">{e.employee_code}</td>
                                        <td>{e.name}</td>
                                        <td class="email-cell">{e.email}</td>
                                        <td>
                                            {#if e.department}
                                                <span class="dept-badge"
                                                    >{e.department}</span
                                                >
                                            {:else}
                                                -
                                            {/if}
                                        </td>
                                        <td>{e.job_title}</td>
                                        <td>{e.branch}</td>
                                        <td>
                                            <span
                                                class="pill-status {e.status ===
                                                'ACTIVE'
                                                    ? 'pill-status-active'
                                                    : 'pill-status-inactive'}"
                                            >
                                                {e.status === "ACTIVE"
                                                    ? "Active"
                                                    : "Inactive"}
                                            </span>
                                        </td>

                                        <td class="actions">
                                            <button
                                                type="button"
                                                class="table-action-btn btn-edit"
                                                on:click={() =>
                                                    handleEditEmployee(e)}
                                            >
                                                Edit
                                            </button>

                                            <button
                                                type="button"
                                                class="table-action-btn btn-delete"
                                                on:click={() =>
                                                    handleDeleteEmployee(e.id)}
                                            >
                                                Delete
                                            </button>
                                        </td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}
            </article>
        </section>

        <AddEmployee
            open={showForm}
            mode={formMode}
            employee={editingEmployee}
            {saving}
            on:close={handleCloseForm}
            on:submit={handleSubmitEmployee}
        />
    </div>
</div>

<style>
    .dept-badge {
        display: inline-block;
        padding: 2px 8px;
        font-size: 11px;
        font-weight: 600;
        background: rgba(0, 212, 255, 0.2);
        color: #00d4ff;
        border-radius: 4px;
    }
</style>
