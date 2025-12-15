<script>
    import { createEventDispatcher } from "svelte";
    import { API_BASE } from "../api.js";
    import "../styles/add-employee.css";

    const dispatch = createEventDispatcher();

    // props dari parent
    export let open = false; // kontrol buka/tutup modal
    export let mode = "create"; // "create" | "edit"
    export let employee = null; // data karyawan yg sedang diedit (bisa null)
    export let saving = false; // state loading tombol

    // Daftar departemen yang tersedia
    const DEPARTMENTS = [
        { code: "ADM", name: "Administrasi" },
        { code: "IT", name: "Information Technology" },
        { code: "ACC", name: "Accounting" },
        { code: "TAX", name: "Pajak" },
        { code: "FIN", name: "Finance" },
        { code: "HR", name: "Human Resources" },
        { code: "MKT", name: "Marketing" },
        { code: "OPS", name: "Operations" },
    ];

    const EMPTY_FORM = {
        employee_code: "",
        name: "",
        email: "",
        branch: "",
        job_title: "",
        status: "ACTIVE",
        department: "",
        password: "",
    };

    let form = { ...EMPTY_FORM };
    let loadingCode = false;

    // flag supaya form hanya di-init saat modal baru dibuka
    let initialized = false;
    let lastMode = null;
    let lastEmpId = null;

    // ====== SYNC FORM DENGAN PROPS SAAT MODAL DIBUKA ======
    $: {
        const empId = employee?.id ?? null;
        if (open && (!initialized || mode !== lastMode || empId !== lastEmpId)) {
            initialized = true;
            lastMode = mode;
            lastEmpId = empId;

            if (mode === "edit" && employee) {
                form = {
                    employee_code: employee.employee_code ?? "",
                    name: employee.name ?? "",
                    email: employee.email ?? "",
                    branch: employee.branch ?? "",
                    job_title: employee.job_title ?? "",
                    status: employee.status ?? "ACTIVE",
                    department: employee.department ?? "",
                    password: "",
                };
            } else {
                // mode create
                form = { ...EMPTY_FORM };
            }
        }

        // ketika modal ditutup, reset flag & form
        if (!open && initialized) {
            initialized = false;
            lastMode = null;
            lastEmpId = null;
            form = { ...EMPTY_FORM };
        }
    }

    // ====== FETCH NEXT EMPLOYEE CODE SAAT DEPARTMENT DIPILIH ======
    async function handleDepartmentChange(event) {
        const dept = event.target.value;
        form.department = dept;

        // Hanya auto-generate code untuk mode create
        if (mode === "create" && dept) {
            loadingCode = true;
            try {
                const res = await fetch(
                    `${API_BASE}/api/employees/next-code?department=${encodeURIComponent(dept)}`,
                    { credentials: "include" },
                );

                if (res.ok) {
                    const data = await res.json();
                    form.employee_code = data.employee_code || "";
                } else {
                    console.error("Failed to fetch next code");
                }
            } catch (e) {
                console.error("Error fetching next code:", e);
            } finally {
                loadingCode = false;
            }
        }
    }

    function close() {
        dispatch("close");
    }

    function handleSubmit() {
        dispatch("submit", {
            mode,
            id: employee?.id ?? null,
            form,
        });
    }
</script>

{#if open}
    <div class="modal-backdrop" on:click={close}>
        <div class="modal" on:click|stopPropagation>
            <header class="modal-header">
                <h2>{mode === "edit" ? "Edit Employee" : "Add Employee"}</h2>
            </header>

            <div class="modal-body">
                <!-- Department dropdown (untuk mode create, harus dipilih dulu) -->
                <div class="form-row">
                    <label>
                        Department
                        <select
                            value={form.department}
                            on:change={handleDepartmentChange}
                            disabled={mode === "edit"}
                        >
                            <option value="">-- Pilih Department --</option>
                            {#each DEPARTMENTS as dept}
                                <option value={dept.code}
                                    >{dept.code} - {dept.name}</option
                                >
                            {/each}
                        </select>
                    </label>
                </div>

                <div class="form-row">
                    <label>
                        Employee Code
                        <input
                            bind:value={form.employee_code}
                            readonly={mode === "create"}
                            placeholder={loadingCode
                                ? "Generating..."
                                : mode === "create"
                                  ? "Pilih department dulu"
                                  : ""}
                            class:readonly-input={mode === "create"}
                        />
                    </label>

                    <label>
                        Name
                        <input bind:value={form.name} />
                    </label>
                </div>

                <div class="form-row">
                    <label>
                        Email
                        <input type="email" bind:value={form.email} placeholder="" autocomplete="off" />
                    </label>
                    <label>
                        Password
                        <input type="password" bind:value={form.password} placeholder="" autocomplete="off" />
                    </label>
                </div>

                <div class="form-row">
                    <label>
                        Job Title
                        <input bind:value={form.job_title} />
                    </label>

                    <label>
                        Branch
                        <input bind:value={form.branch} />
                    </label>
                </div>

                <div class="form-row">
                    <label>
                        Status
                        <select bind:value={form.status}>
                            <option value="ACTIVE">ACTIVE</option>
                            <option value="INACTIVE">INACTIVE</option>
                        </select>
                    </label>
                </div>

                <!-- note dihilangkan sesuai permintaan -->
            </div>

            <footer class="modal-footer">
                <button
                    type="button"
                    class="btn-secondary"
                    on:click={close}
                    disabled={saving}
                >
                    Cancel
                </button>

                <button
                    type="button"
                    class="btn-primary"
                    on:click|preventDefault={handleSubmit}
                    disabled={saving ||
                        (mode === "create" && !form.employee_code)}
                >
                    {#if saving}
                        {mode === "edit" ? "Updating…" : "Saving…"}
                    {:else}
                        {mode === "edit" ? "Update" : "Save"}
                    {/if}
                </button>
            </footer>
        </div>
    </div>
{/if}
