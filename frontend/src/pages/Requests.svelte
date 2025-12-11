<script>
    import { onMount } from "svelte";
    import Header from "../components/Header.svelte";
    import { API_BASE } from "../api.js";
    import RequestModal from "../lib/components/RequestModal.svelte";
    import "../styles/dashboard.css";

    export let page = "MY_REQUESTS";

    let activeTab = "my_requests";
    let requests = [];
    let showModal = false;
    let modalType = "LEAVE";
    let loading = false;
    let currentUser = { roles: [] };

    const statusClass = {
        PENDING: "status-pending",
        APPROVED: "status-approved",
        REJECTED: "status-rejected",
    };

    $: {
        if (page === "APPROVALS") {
            activeTab = "approvals";
        } else {
            activeTab = "my_requests";
        }

        if (page === "LEAVE_REQUEST") {
            modalType = "LEAVE";
            showModal = true;
        } else if (page === "OVERTIME_REQ") {
            modalType = "OVERTIME";
            showModal = true;
        }
    }

    $: if (activeTab) {
        if (typeof window !== "undefined") fetchRequests();
    }

    onMount(async () => {
        await loadUser();
    });

    async function loadUser() {
        try {
            const res = await fetch(`${API_BASE}/api/me`, {
                credentials: "include",
            });
            if (res.ok) {
                currentUser = await res.json();
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function fetchRequests() {
        loading = true;
        const endpoint =
            activeTab === "my_requests"
                ? `${API_BASE}/api/requests/my`
                : `${API_BASE}/api/requests/approvals`;

        try {
            const res = await fetch(endpoint, {
                credentials: "include",
            });
            if (res.ok) {
                requests = (await res.json()) || [];
            } else {
                requests = [];
            }
        } catch (e) {
            console.error(e);
            requests = [];
        } finally {
            loading = false;
        }
    }

    function handleTabChange(tab) {
        activeTab = tab;
    }

    function handleRequestCreated() {
        showModal = false;
        fetchRequests();
    }

    async function handleApprove(id) {
        if (!confirm("Approve this request?")) return;
        try {
            const res = await fetch(`${API_BASE}/api/requests/${id}/approve`, {
                method: "POST",
                credentials: "include",
            });
            if (res.ok) {
                fetchRequests();
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function handleReject(id) {
        const reason = prompt("Reason for rejection:");
        if (!reason) return;

        try {
            const res = await fetch(`${API_BASE}/api/requests/${id}/reject`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify({ reason }),
            });
            if (res.ok) {
                fetchRequests();
            }
        } catch (e) {
            console.error(e);
        }
    }

    $: canApprove = activeTab === "approvals";

    $: pageTitle = activeTab === "approvals" ? "Approvals" : "Request History";
</script>
<div class="page-wrapper">
    <Header>
        <div>
            <h1>{pageTitle}</h1>
            <p>Manage leave and overtime requests</p>
        </div>
    </Header>

    <div class="page-content">
        {#if activeTab === "my_requests"}
            <div class="actions">
                <button class="action-btn primary" on:click={() => { modalType = "LEAVE"; showModal = true; }}>+ Request Leave</button>
                <button class="action-btn" on:click={() => { modalType = "OVERTIME"; showModal = true; }}>Overtime Request</button>
                <button class="action-btn" on:click={() => { modalType = "EXIT_CLEARANCE"; showModal = true; }}>Exit Clearance</button>
                <button class="action-btn" on:click={() => { modalType = "MEDICAL_CLAIM"; showModal = true; }}>Medical Claim</button>
            </div>
        {/if}

        <div class="tabs">
            <button
                class="tab-button {activeTab === 'my_requests' ? 'active' : ''}"
                on:click={() => handleTabChange("my_requests")}
            >
                My Requests
            </button>
            <button
                class="tab-button {activeTab === 'approvals' ? 'active' : ''}"
                on:click={() => handleTabChange("approvals")}
            >
                Approvals
            </button>
        </div>

        {#if loading}
            <div class="state muted">Loading...</div>
        {:else}
            <article class="card">
                {#if requests.length === 0}
                    <div class="state muted">No requests found.</div>
                {:else}
                    <div class="employees-table-wrapper">
                        <table class="employees-table">
                            <thead>
                                <tr>
                                    <th>Type</th>
                                    <th>Dates</th>
                                    <th>Reason</th>
                                    {#if activeTab === "approvals"}
                                        <th>Requester</th>
                                    {/if}
                                    <th>Status</th>
                                    {#if canApprove}
                                        <th class="text-right">Actions</th>
                                    {/if}
                                </tr>
                            </thead>
                            <tbody>
                                {#each requests as req}
                                    <tr>
                                        <td>
                                            <span class="type-badge">{req.type}</span>
                                        </td>
                                        <td>
                                            {new Date(req.start_date).toLocaleDateString()} -
                                            {new Date(req.end_date).toLocaleDateString()}
                                        </td>
                                        <td title={req.reason}>{req.reason}</td>
                                        {#if activeTab === "approvals"}
                                            <td>{req.user_name || "Unknown"}</td>
                                        {/if}
                                        <td>
                                            <span class="status-badge {statusClass[req.status] || ''}">{req.status}</span>
                                        </td>
                                        {#if canApprove}
                                            <td class="text-right">
                                                {#if req.status === "PENDING"}
                                                    <button class="btn" on:click={() => handleApprove(req.id)}>Approve</button>
                                                    <button class="btn btn-danger" on:click={() => handleReject(req.id)}>Reject</button>
                                                {/if}
                                            </td>
                                        {/if}
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}
            </article>
        {/if}

        <RequestModal bind:isOpen={showModal} bind:type={modalType} on:created={handleRequestCreated} />
    </div>
</div>

<style>
    .actions {
        display: flex;
        gap: 8px;
        margin-bottom: 16px;
    }
    .action-btn {
        border-radius: 999px;
        padding: 10px 14px;
        font-size: 0.9rem;
        border: 1px solid var(--border-color);
        background: var(--bg-hover);
        color: var(--text-body);
        cursor: pointer;
        transition: all 0.15s ease;
    }
    .action-btn:hover {
        background: var(--border-color);
        color: var(--text-primary);
        transform: translateY(-1px);
    }
    .action-btn.primary {
        border: none;
        background: linear-gradient(90deg, #2563eb, #4f46e5);
        color: white;
        box-shadow: 0 8px 20px rgba(37,99,235,0.35);
    }
    .action-btn.primary:hover {
        filter: brightness(1.05);
    }
    .tabs {
        display: flex;
        gap: 12px;
        border-bottom: 1px solid var(--border-color);
        margin-bottom: 16px;
    }
    .tab-button {
        background: transparent;
        border: none;
        color: var(--text-muted);
        padding: 10px 6px;
        cursor: pointer;
        font-weight: 500;
        border-bottom: 2px solid transparent;
    }
    .tab-button.active {
        color: var(--accent-primary);
        border-bottom-color: var(--accent-primary);
    }
    .type-badge {
        display: inline-flex;
        align-items: center;
        padding: 3px 8px;
        border-radius: 999px;
        font-size: 0.75rem;
        background: var(--bg-hover);
        color: var(--text-body);
        border: 1px solid var(--border-color);
    }
    .status-badge {
        display: inline-flex;
        align-items: center;
        padding: 3px 8px;
        border-radius: 999px;
        font-size: 0.75rem;
        border: 1px solid var(--border-color);
    }
    .status-pending {
        background: rgba(234, 179, 8, 0.15);
        color: #f59e0b;
        border-color: rgba(234, 179, 8, 0.3);
    }
    .status-approved {
        background: rgba(34, 197, 94, 0.15);
        color: #22c55e;
        border-color: rgba(34, 197, 94, 0.3);
    }
    .status-rejected {
        background: rgba(239, 68, 68, 0.15);
        color: #ef4444;
        border-color: rgba(239, 68, 68, 0.3);
    }
    .state.muted {
        color: var(--text-muted);
        padding: 12px 4px;
    }
</style>
