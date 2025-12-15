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
    let lastWorkDate = "";
    let handoverNotes = "";
    let supervisorName = "";
    let assetsReturned = {
        id_card: false,
        laptop: false,
        keys: false,
        email_access_disabled: false,
    };
    let claimDate = "";
    let provider = "";
    let amount = "";
    let claimDescription = "";

    // Datepicker overlay state
    let activePicker = null; // "start" | "end" | "last" | "claim" | null
    let pickerMonth = new Date().getMonth(); // 0-11
    let pickerYear = new Date().getFullYear();

    function parseDDMMYYYY(s) {
        const m = String(s || "").match(/^(\d{2})\/(\d{2})\/(\d{4})$/);
        if (!m) return null;
        const dd = Number(m[1]), mm = Number(m[2]) - 1, yy = Number(m[3]);
        const d = new Date(yy, mm, dd);
        return isNaN(d.getTime()) ? null : d;
    }
    function formatDDMMYYYY(d) {
        const dd = String(d.getDate()).padStart(2, "0");
        const mm = String(d.getMonth() + 1).padStart(2, "0");
        const yy = d.getFullYear();
        return `${dd}/${mm}/${yy}`;
    }
    function daysInMonth(year, month) {
        return new Date(year, month + 1, 0).getDate();
    }
    function openPicker(which) {
        activePicker = which;
        const v = which === "start" ? startDate : endDate;
        const d = parseDDMMYYYY(v) || new Date();
        pickerMonth = d.getMonth();
        pickerYear = d.getFullYear();
    }
    function closePicker() {
        activePicker = null;
    }
    function pickDate(day) {
        const d = new Date(pickerYear, pickerMonth, day);
        const val = formatDDMMYYYY(d);
        if (activePicker === "start") startDate = val;
        else if (activePicker === "end") endDate = val;
        else if (activePicker === "last") lastWorkDate = val;
        else if (activePicker === "claim") claimDate = val;
        closePicker();
    }
    const monthNames = ["Januari","Februari","Maret","April","Mei","Juni","Juli","Agustus","September","Oktober","November","Desember"];
    const dayNames = ["Min","Sen","Sel","Rab","Kam","Jum","Sab"];
    function prevMonth() {
        if (pickerMonth === 0) { pickerMonth = 11; pickerYear -= 1; } else pickerMonth -= 1;
    }
    function nextMonth() {
        if (pickerMonth === 11) { pickerMonth = 0; pickerYear += 1; } else pickerMonth += 1;
    }
    let loading = false;
    let error = "";

    function toISOFromDDMMYYYY(s, endOfDay) {
        const m = String(s || "").match(/^(\d{2})\/(\d{2})\/(\d{4})$/);
        if (!m) return "";
        const d = `${m[3]}-${m[2]}-${m[1]}${endOfDay ? "T23:59:59Z" : "T00:00:00Z"}`;
        const dt = new Date(d);
        return isNaN(dt.getTime()) ? "" : dt.toISOString();
    }

    function normalizeDateInput(value) {
        let v = String(value || "").replace(/[^\d]/g, "");
        if (v.length > 8) v = v.slice(0, 8);
        const a = v.slice(0, 2);
        const b = v.slice(2, 4);
        const c = v.slice(4, 8);
        let out = a;
        if (b) out += "/" + b;
        if (c) out += "/" + c;
        return out;
    }
    function clearDate(which) {
        if (which === "start") startDate = "";
        else if (which === "end") endDate = "";
        else if (which === "last") lastWorkDate = "";
        else if (which === "claim") claimDate = "";
        closePicker();
    }

    async function handleSubmit() {
        if (!startDate || !endDate || !reason) {
            if (type === "EXIT_CLEARANCE") {
                if (!lastWorkDate || !handoverNotes || !supervisorName) {
                    error = "Please fill all fields";
                    return;
                }
            } else if (type === "MEDICAL_CLAIM") {
                if (!claimDate || !provider || !amount) {
                    error = "Please fill all fields";
                    return;
                }
            } else {
                error = "Please fill all fields";
                return;
            }
        }

        loading = true;
        error = "";

        try {
            let body = { type };
            if (type === "EXIT_CLEARANCE") {
                const startISO = toISOFromDDMMYYYY(lastWorkDate, false);
                const endISO = toISOFromDDMMYYYY(lastWorkDate, true);
                if (!startISO || !endISO) {
                    error = "Invalid start date format (RFC3339 required)";
                    loading = false;
                    return;
                }
                const assets = Object.keys(assetsReturned).filter((k) => assetsReturned[k]);
                const reasonText = `Supervisor: ${supervisorName}; Notes: ${handoverNotes}; Assets: ${assets.join(", ")}`;
                body = { type, start_date: startISO, end_date: endISO, reason: reasonText };
            } else if (type === "MEDICAL_CLAIM") {
                const startISO = toISOFromDDMMYYYY(claimDate, false);
                const endISO = toISOFromDDMMYYYY(claimDate, true);
                if (!startISO || !endISO) {
                    error = "Invalid start date format (RFC3339 required)";
                    loading = false;
                    return;
                }
                const amt = Number(amount || 0);
                const reasonText = `Provider: ${provider}; Amount: ${amt}; ${claimDescription || reason || ""}`.trim();
                body = { type, start_date: startISO, end_date: endISO, reason: reasonText };
            } else {
                body = {
                    type,
                    start_date: toISOFromDDMMYYYY(startDate, false),
                    end_date: toISOFromDDMMYYYY(endDate, true),
                    reason,
                };
            }

            const res = await fetch(`${API_BASE}/api/requests`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify(body),
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
        lastWorkDate = "";
        handoverNotes = "";
        supervisorName = "";
        assetsReturned = {
            id_card: false,
            laptop: false,
            keys: false,
            email_access_disabled: false,
        };
        claimDate = "";
        provider = "";
        amount = "";
        claimDescription = "";
        error = "";
    }

    function close() {
        isOpen = false;
        resetForm();
    }
</script>

{#if isOpen}
    <div class="modal-overlay" on:click={close}>
        <div class="modal" role="dialog" aria-modal="true" on:click|stopPropagation style="width: 520px;">
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
                {#if type === "EXIT_CLEARANCE"}
                    <div class="form-row">
                        <label>
                            Last Work Date
                            <div style="position:relative">
                                <input
                                    type="text"
                                    bind:value={lastWorkDate}
                                    class="form-control"
                                    placeholder="dd/mm/yyyy"
                                    on:input={(e) => (lastWorkDate = normalizeDateInput(e.target.value))}
                                    on:focus={() => openPicker("last")}
                                    on:blur={closePicker}
                                />
                                {#if activePicker === "last"}
                                    <div class="datepicker-pop" role="dialog" aria-label="Datepicker" on:click|stopPropagation on:mousedown|preventDefault tabindex="-1">
                                        <div class="dp-header">
                                            <button type="button" on:click={prevMonth}>◀</button>
                                            <span>{monthNames[pickerMonth]} {pickerYear}</span>
                                            <button type="button" on:click={nextMonth}>▶</button>
                                        </div>
                                        <div class="dp-grid">
                                            {#each dayNames as dn}
                                                <div class="dp-dow">{dn}</div>
                                            {/each}
                                            {#each Array(new Date(pickerYear, pickerMonth, 1).getDay()).fill(0) as _}
                                                <div></div>
                                            {/each}
                                            {#each Array(daysInMonth(pickerYear, pickerMonth)).fill(0).map((_,i)=>i+1) as d}
                                                <div class="dp-day" role="button" tabindex="0" on:keydown={(e)=> e.key==='Enter' && pickDate(d)} on:click={() => pickDate(d)}>{d}</div>
                                            {/each}
                                        </div>
                                        <div style="display:flex; justify-content:space-between; margin-top:8px;">
                                            <button type="button" class="btn" on:click={() => clearDate("last")}>Clear</button>
                                            <button type="button" class="btn" on:click={closePicker}>Close</button>
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        </label>
                        <label>
                            Supervisor Name
                            <input type="text" bind:value={supervisorName} class="form-control" />
                        </label>
                    </div>
                    <div class="form-row">
                        <label>
                            Handover Notes
                            <textarea rows="3" bind:value={handoverNotes} class="form-control"></textarea>
                        </label>
                    </div>
                    <div class="form-row">
                        <div class="checkboxes">
                            <label><input type="checkbox" bind:checked={assetsReturned.id_card} /> ID Card</label>
                            <label><input type="checkbox" bind:checked={assetsReturned.laptop} /> Laptop</label>
                            <label><input type="checkbox" bind:checked={assetsReturned.keys} /> Keys</label>
                            <label><input type="checkbox" bind:checked={assetsReturned.email_access_disabled} /> Email Access Disabled</label>
                        </div>
                    </div>
                {:else if type === "MEDICAL_CLAIM"}
                    <div class="form-row">
                        <label>
                            Claim Date
                            <div style="position:relative">
                                <input
                                    type="text"
                                    bind:value={claimDate}
                                    class="form-control"
                                    placeholder="dd/mm/yyyy"
                                    on:input={(e) => (claimDate = normalizeDateInput(e.target.value))}
                                    on:focus={() => openPicker("claim")}
                                    on:blur={closePicker}
                                />
                                {#if activePicker === "claim"}
                                    <div class="datepicker-pop" role="dialog" aria-label="Datepicker" on:click|stopPropagation on:mousedown|preventDefault tabindex="-1">
                                        <div class="dp-header">
                                            <button type="button" on:click={prevMonth}>◀</button>
                                            <span>{monthNames[pickerMonth]} {pickerYear}</span>
                                            <button type="button" on:click={nextMonth}>▶</button>
                                        </div>
                                        <div class="dp-grid">
                                            {#each dayNames as dn}
                                                <div class="dp-dow">{dn}</div>
                                            {/each}
                                            {#each Array(new Date(pickerYear, pickerMonth, 1).getDay()).fill(0) as _}
                                                <div></div>
                                            {/each}
                                            {#each Array(daysInMonth(pickerYear, pickerMonth)).fill(0).map((_,i)=>i+1) as d}
                                                <div class="dp-day" role="button" tabindex="0" on:keydown={(e)=> e.key==='Enter' && pickDate(d)} on:click={() => pickDate(d)}>{d}</div>
                                            {/each}
                                        </div>
                                        <div style="display:flex; justify-content:space-between; margin-top:8px;">
                                            <button type="button" class="btn" on:click={() => clearDate("claim")}>Clear</button>
                                            <button type="button" class="btn" on:click={closePicker}>Close</button>
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        </label>
                        <label>
                            Provider
                            <input type="text" bind:value={provider} class="form-control" />
                        </label>
                    </div>
                    <div class="form-row">
                        <label>
                            Amount
                            <input type="number" min="0" step="1" bind:value={amount} class="form-control" />
                        </label>
                        <label>
                            Description
                            <input type="text" bind:value={claimDescription} class="form-control" />
                        </label>
                    </div>
                {:else}
                    <div class="form-row">
                        <label>
                            Start Date
                            <div style="position:relative">
                                <input
                                    type="text"
                                    bind:value={startDate}
                                    class="form-control"
                                    placeholder="dd/mm/yyyy"
                                    on:input={(e) => (startDate = normalizeDateInput(e.target.value))}
                                    on:focus={() => openPicker("start")}
                                    on:blur={closePicker}
                                />
                                {#if activePicker === "start"}
                                    <div class="datepicker-pop" role="dialog" aria-label="Datepicker" on:click|stopPropagation on:mousedown|preventDefault tabindex="-1">
                                        <div class="dp-header">
                                            <button type="button" on:click={prevMonth}>◀</button>
                                            <span>{monthNames[pickerMonth]} {pickerYear}</span>
                                            <button type="button" on:click={nextMonth}>▶</button>
                                        </div>
                                        <div class="dp-grid">
                                            {#each dayNames as dn}
                                                <div class="dp-dow">{dn}</div>
                                            {/each}
                                            {#each Array(new Date(pickerYear, pickerMonth, 1).getDay()).fill(0) as _}
                                                <div></div>
                                            {/each}
                                            {#each Array(daysInMonth(pickerYear, pickerMonth)).fill(0).map((_,i)=>i+1) as d}
                                                <div class="dp-day" role="button" tabindex="0" on:keydown={(e)=> e.key==='Enter' && pickDate(d)} on:click={() => pickDate(d)}>{d}</div>
                                            {/each}
                                        </div>
                                        <div style="display:flex; justify-content:space-between; margin-top:8px;">
                                            <button type="button" class="btn" on:click={() => clearDate("start")}>Clear</button>
                                            <button type="button" class="btn" on:click={closePicker}>Close</button>
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        </label>
                        <label>
                            End Date
                            <div style="position:relative">
                                <input
                                    type="text"
                                    bind:value={endDate}
                                    class="form-control"
                                    placeholder="dd/mm/yyyy"
                                    on:input={(e) => (endDate = normalizeDateInput(e.target.value))}
                                    on:focus={() => openPicker("end")}
                                    on:blur={closePicker}
                                />
                                {#if activePicker === "end"}
                                    <div class="datepicker-pop" role="dialog" aria-label="Datepicker" on:click|stopPropagation on:mousedown|preventDefault tabindex="-1">
                                        <div class="dp-header">
                                            <button type="button" on:click={prevMonth}>◀</button>
                                            <span>{monthNames[pickerMonth]} {pickerYear}</span>
                                            <button type="button" on:click={nextMonth}>▶</button>
                                        </div>
                                        <div class="dp-grid">
                                            {#each dayNames as dn}
                                                <div class="dp-dow">{dn}</div>
                                            {/each}
                                            {#each Array(new Date(pickerYear, pickerMonth, 1).getDay()).fill(0) as _}
                                                <div></div>
                                            {/each}
                                            {#each Array(daysInMonth(pickerYear, pickerMonth)).fill(0).map((_,i)=>i+1) as d}
                                                <div class="dp-day" role="button" tabindex="0" on:keydown={(e)=> e.key==='Enter' && pickDate(d)} on:click={() => pickDate(d)}>{d}</div>
                                            {/each}
                                        </div>
                                        <div style="display:flex; justify-content:space-between; margin-top:8px;">
                                            <button type="button" class="btn" on:click={() => clearDate("end")}>Clear</button>
                                            <button type="button" class="btn" on:click={closePicker}>Close</button>
                                        </div>
                                    </div>
                                {/if}
                            </div>
                        </label>
                    </div>
                    <div class="form-row">
                        <label>
                            Reason
                            <textarea rows="3" bind:value={reason} class="form-control"></textarea>
                        </label>
                    </div>
                {/if}
                {#if isOpen}
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

<style>
    .datepicker-pop {
        position: absolute;
        z-index: 30;
        margin-top: 8px;
        border: 1px solid var(--border-color);
        border-radius: 12px;
        background: var(--bg-card);
        box-shadow: var(--shadow-card);
        width: 260px;
        padding: 10px;
    }
    .dp-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 8px;
    }
    .dp-header select, .dp-header button {
        background: var(--bg-hover);
        color: var(--text-body);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        padding: 6px 8px;
        font-size: 0.8rem;
    }
    .dp-grid {
        display: grid;
        grid-template-columns: repeat(7, 1fr);
        gap: 4px;
    }
    .dp-day, .dp-dow {
        text-align: center;
        padding: 6px 0;
        border-radius: 8px;
        font-size: 0.8rem;
    }
    .dp-dow {
        color: var(--text-muted);
    }
    .dp-day {
        cursor: pointer;
        border: 1px solid transparent;
    }
    .dp-day:hover {
        background: var(--bg-hover);
        border-color: var(--border-color);
    }
</style>
