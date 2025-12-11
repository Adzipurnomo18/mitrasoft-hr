<script>
    import { onMount } from "svelte";
    import { API_BASE } from "../api.js";
    import { user } from "../stores.js";
    import Header from "../components/Header.svelte";
    import "../styles/dashboard.css";
    import "../styles/profile.css";

    let currentUser = null;
    let profile = null;
    let loading = true;
    let saving = false;
    let editing = false;
    let error = "";

    // Form data for editing
    let form = {
        name: "",
        phone: "",
        address: "",
        birth_date: "",
        gender: "",
        emergency_contact: "",
        emergency_phone: "",
    };

    user.subscribe((value) => (currentUser = value));

    onMount(loadProfile);

    async function loadProfile() {
        loading = true;
        error = "";
        try {
            const res = await fetch(`${API_BASE}/api/me`, {
                credentials: "include",
            });

            if (!res.ok) {
                throw new Error("Failed to load profile");
            }

            profile = await res.json();
            initForm();
        } catch (e) {
            error = e.message || "Failed to load profile";
        } finally {
            loading = false;
        }
    }

    function initForm() {
        if (profile) {
            form = {
                name: profile.name || "",
                phone: profile.phone || "",
                address: profile.address || "",
                birth_date: profile.birth_date
                    ? profile.birth_date.split("T")[0]
                    : "",
                gender: profile.gender || "",
                emergency_contact: profile.emergency_contact || "",
                emergency_phone: profile.emergency_phone || "",
            };
        }
    }

    function startEdit() {
        initForm();
        editing = true;
    }

    function cancelEdit() {
        editing = false;
        initForm();
    }

    async function saveProfile() {
        saving = true;
        error = "";

        try {
            const res = await fetch(`${API_BASE}/api/me`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify(form),
            });

            if (!res.ok) {
                const body = await res.json().catch(() => ({}));
                throw new Error(body.message || "Failed to save profile");
            }

            profile = await res.json();
            user.set(profile); // update global store
            editing = false;
        } catch (e) {
            error = e.message || "Failed to save profile";
        } finally {
            saving = false;
        }
    }

    function formatDate(dateStr) {
        if (!dateStr) return "-";
        const d = new Date(dateStr);
        return d.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "long",
            year: "numeric",
        });
    }
</script>

<div class="page-wrapper">
    <Header>
        <div>
            <h1>My Profile</h1>
            <p>View and update your personal information</p>
        </div>
    </Header>

    <div class="page-content">
        {#if loading}
            <div class="loading-state">Loading profile...</div>
        {:else if error}
            <div class="error-state">{error}</div>
        {:else if profile}
            <div class="profile-container">
                <!-- Profile Header Card -->
                <div class="profile-header-card">
                    <div class="profile-avatar">
                        {#if profile.photo_url}
                            <img src={profile.photo_url} alt={profile.name} />
                        {:else}
                            <span class="avatar-initials">
                                {profile.name
                                    ? profile.name.charAt(0).toUpperCase()
                                    : "?"}
                            </span>
                        {/if}
                    </div>
                    <div class="profile-header-info">
                        <h2>{profile.name}</h2>
                        <p class="job-title">{profile.job_title || "-"}</p>
                        <p class="dept-info">
                            {#if profile.department}
                                <span class="dept-badge"
                                    >{profile.department}</span
                                >
                            {/if}
                            <span class="emp-code">{profile.employee_code}</span
                            >
                        </p>
                    </div>
                    <div class="profile-actions">
                        {#if !editing}
                            <button class="btn-primary" on:click={startEdit}
                                >Edit Profile</button
                            >
                        {/if}
                    </div>
                </div>

                {#if editing}
                    <!-- Edit Form -->
                    <div class="profile-form card">
                        <h3>Edit Profile</h3>
                        <div class="form-grid">
                            <label>
                                <span>Full Name</span>
                                <input type="text" bind:value={form.name} />
                            </label>
                            <label>
                                <span>Phone</span>
                                <input
                                    type="tel"
                                    bind:value={form.phone}
                                    placeholder="+62..."
                                />
                            </label>
                            <label>
                                <span>Birth Date</span>
                                <input
                                    type="date"
                                    bind:value={form.birth_date}
                                />
                            </label>
                            <label>
                                <span>Gender</span>
                                <select bind:value={form.gender}>
                                    <option value="">Select...</option>
                                    <option value="M">Male</option>
                                    <option value="F">Female</option>
                                </select>
                            </label>
                            <label class="full-width">
                                <span>Address</span>
                                <textarea bind:value={form.address} rows="2"
                                ></textarea>
                            </label>
                            <label>
                                <span>Emergency Contact Name</span>
                                <input
                                    type="text"
                                    bind:value={form.emergency_contact}
                                />
                            </label>
                            <label>
                                <span>Emergency Contact Phone</span>
                                <input
                                    type="tel"
                                    bind:value={form.emergency_phone}
                                />
                            </label>
                        </div>
                        <div class="form-actions">
                            <button
                                class="btn-secondary"
                                on:click={cancelEdit}
                                disabled={saving}>Cancel</button
                            >
                            <button
                                class="btn-primary"
                                on:click={saveProfile}
                                disabled={saving}
                            >
                                {saving ? "Saving..." : "Save Changes"}
                            </button>
                        </div>
                    </div>
                {:else}
                    <!-- View Mode -->
                    <div class="profile-details-grid">
                        <div class="card profile-card">
                            <h3>Personal Information</h3>
                            <dl class="info-list">
                                <div>
                                    <dt>Email</dt>
                                    <dd>{profile.email}</dd>
                                </div>
                                <div>
                                    <dt>Phone</dt>
                                    <dd>{profile.phone || "-"}</dd>
                                </div>
                                <div>
                                    <dt>Birth Date</dt>
                                    <dd>{formatDate(profile.birth_date)}</dd>
                                </div>
                                <div>
                                    <dt>Gender</dt>
                                    <dd>
                                        {profile.gender === "M"
                                            ? "Male"
                                            : profile.gender === "F"
                                              ? "Female"
                                              : "-"}
                                    </dd>
                                </div>
                                <div class="full-width">
                                    <dt>Address</dt>
                                    <dd>{profile.address || "-"}</dd>
                                </div>
                            </dl>
                        </div>

                        <div class="card profile-card">
                            <h3>Employment Information</h3>
                            <dl class="info-list">
                                <div>
                                    <dt>Employee Code</dt>
                                    <dd class="mono">
                                        {profile.employee_code}
                                    </dd>
                                </div>
                                <div>
                                    <dt>Department</dt>
                                    <dd>{profile.department || "-"}</dd>
                                </div>
                                <div>
                                    <dt>Job Title</dt>
                                    <dd>{profile.job_title || "-"}</dd>
                                </div>
                                <div>
                                    <dt>Branch</dt>
                                    <dd>{profile.branch || "-"}</dd>
                                </div>
                                <div>
                                    <dt>Join Date</dt>
                                    <dd>{formatDate(profile.join_date)}</dd>
                                </div>
                                <div>
                                    <dt>Status</dt>
                                    <dd>
                                        <span
                                            class="status-badge {profile.status ===
                                            'ACTIVE'
                                                ? 'active'
                                                : 'inactive'}"
                                        >
                                            {profile.status}
                                        </span>
                                    </dd>
                                </div>
                            </dl>
                        </div>

                        <div class="card profile-card">
                            <h3>Emergency Contact</h3>
                            <dl class="info-list">
                                <div>
                                    <dt>Contact Name</dt>
                                    <dd>{profile.emergency_contact || "-"}</dd>
                                </div>
                                <div>
                                    <dt>Contact Phone</dt>
                                    <dd>{profile.emergency_phone || "-"}</dd>
                                </div>
                            </dl>
                        </div>
                    </div>
                {/if}
            </div>
        {/if}
    </div>
</div>
