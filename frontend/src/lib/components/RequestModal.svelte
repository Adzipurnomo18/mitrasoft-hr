<script>
    import { createEventDispatcher } from "svelte";
    import { API_BASE } from "../../api.js";
    import "../../styles/components.css";
    import "../../styles/add-employee.css";

    export let isOpen = false;
    export let type = "LEAVE"; // Bindable
    const dispatch = createEventDispatcher();

    let startDate = "";
    let endDate = "";
    let reason = "";
    let loading = false;
    let error = "";

    async function handleSubmit() {
        if (!startDate || !endDate || !reason) {
            error = "Please fill all fields";
            return;
        }

        loading = true;
        error = "";

        try {
            const res = await fetch(`${API_BASE}/api/requests`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify({
                    type,
                    start_date: new Date(startDate + "T00:00:00Z").toISOString(),
                    end_date: new Date(endDate + "T23:59:59Z").toISOString(),
                    reason,
                }),
            });

            if (res.ok) {
                dispatch("created");
                resetForm();
            } else {
                const data = await res.json();
                error = data.error || "Failed to submit request";
            }
        } catch (e) {
            error = "Network error";
        } finally {
            loading = false;
        }
    }

    function resetForm() {
        type = "LEAVE";
        startDate = "";
        endDate = "";
        reason = "";
        error = "";
    }

    function close() {
        isOpen = false;
        resetForm();
    }
</script>

{#if isOpen}
    <div class="modal-overlay" on:click={close}>
        <div class="modal" role="dialog" aria-modal="true" on:click|stopPropagation>
            <header class="modal-header">
                <h2>New Request</h2>
            </header>
            <div class="modal-body">
                <div class="form-row">
                    <label>
                        Type
                        <select bind:value={type} class="form-control">
                            <option value="LEAVE">Leave (Cuti)</option>
                            <option value="OVERTIME">Overtime (Lembur)</option>
                            <option value="PERMIT">Permit (Izin)</option>
                            <option value="EXIT_CLEARANCE">Exit Clearance</option>
                            <option value="MEDICAL_CLAIM">Medical Claim</option>
                        </select>
                    </label>
                </div>
                <div class="form-row">
                    <label>
                        Start Date
                        <input type="date" bind:value={startDate} class="form-control" />
                    </label>
                    <label>
                        End Date
                        <input type="date" bind:value={endDate} class="form-control" />
                    </label>
                </div>
                <div class="form-row">
                    <label>
                        Reason
                        <textarea rows="3" bind:value={reason} class="form-control"></textarea>
                    </label>
                </div>
                {#if error}
                    <div class="alert alert-error">{error}</div>
                {/if}
            </div>
            <footer class="modal-footer">
                <button type="button" class="btn" on:click={close} disabled={loading}>Cancel</button>
                <button type="button" class="btn-primary" on:click={handleSubmit} disabled={loading}>
                    {loading ? "Submitting…" : "Submit"}
                </button>
            </footer>
        </div>
    </div>
{/if}
