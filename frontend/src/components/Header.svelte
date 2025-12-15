<script>
    import { createEventDispatcher, onMount } from "svelte";
    import { theme, user, menus, permissions, currentPage } from "../stores.js"; // Import stores
    import { API_BASE } from "../api.js";

    const dispatch = createEventDispatcher();

    // Theme Logic
    let currentTheme = "dark";
    theme.subscribe((val) => (currentTheme = val));

    function toggleTheme() {
        theme.set(currentTheme === "dark" ? "light" : "dark");
    }

    // Logout Logic
    let showLogoutConfirm = false;

    function confirmLogout() {
        showLogoutConfirm = true;
    }

    function cancelLogout() {
        showLogoutConfirm = false;
    }

    async function handleLogout() {
        showLogoutConfirm = false;
        try {
            await fetch(`${API_BASE}/api/auth/logout`, {
                method: "POST",
                credentials: "include",
            });
        } catch (e) {
            console.error("Logout failed", e);
        } finally {
            // Clean up state directly
            user.set(null);
            menus.set([]);
            permissions.set([]);
            currentPage.set("DASHBOARD");
            dispatch("logout"); // Optional, if parent wants to know
        }
    }

    // Attendance actions
    let checking = false;
    let attInfo = { checkin_time: "", checkout_time: "", status: "" };
    function fmtTime(s) {
        if (!s) return "-";
        const d = new Date(s);
        const hh = String(d.getHours()).padStart(2, "0");
        const mm = String(d.getMinutes()).padStart(2, "0");
        return `${hh}:${mm}`;
    }
    function attText() {
        const ci = fmtTime(attInfo.checkin_time);
        const co = fmtTime(attInfo.checkout_time);
        const d = new Date();
        const dateStr = d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
        return `${dateStr} · CI ${ci} · CO ${co}`;
    }
    async function loadToday() {
        try {
            const now = new Date();
            const y = now.getFullYear();
            const m = String(now.getMonth() + 1).padStart(2, "0");
            const d = String(now.getDate()).padStart(2, "0");
            const from = `${y}-${m}-${d}`;
            const toD = new Date(y, now.getMonth(), now.getDate() + 1);
            const to = `${toD.getFullYear()}-${String(toD.getMonth() + 1).padStart(2, "0")}-${String(toD.getDate()).padStart(2, "0")}`;
            const res = await fetch(`${API_BASE}/api/attendance/list?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`, {
                credentials: "include",
            });
            if (res.ok) {
                const items = await res.json();
                if (Array.isArray(items) && items.length > 0) {
                    const today = items.find((r) => r.date && new Date(r.date).toDateString() === new Date(from).toDateString()) || items[0];
                    attInfo = {
                        checkin_time: today.checkin_time || "",
                        checkout_time: today.checkout_time || "",
                        status: today.status || "",
                    };
                }
            }
        } catch (_) {}
    }
    onMount(() => {
        loadToday();
    });
    async function checkin() {
        if (checking) return;
        checking = true;
        try {
            const res = await fetch(`${API_BASE}/api/attendance/checkin`, {
                method: "POST",
                credentials: "include",
            });
            if (!res.ok) {
                const msg = await res.text();
                throw new Error(msg || "Check-in failed");
            }
            const rec = await res.json();
            attInfo = {
                checkin_time: rec.checkin_time || "",
                checkout_time: rec.checkout_time || "",
                status: rec.status || "",
            };
            alert("Check-in berhasil");
            // optional: refresh metrics via custom event
            dispatch("refresh");
        } catch (e) {
            console.error(e);
            alert(e?.message || "Check-in failed");
        } finally {
            checking = false;
        }
    }
    async function checkout() {
        if (checking) return;
        checking = true;
        try {
            const res = await fetch(`${API_BASE}/api/attendance/checkout`, {
                method: "POST",
                credentials: "include",
            });
            if (!res.ok) {
                const msg = await res.text();
                throw new Error(msg || "Check-out failed");
            }
            const rec = await res.json();
            attInfo = {
                checkin_time: rec.checkin_time || attInfo.checkin_time || "",
                checkout_time: rec.checkout_time || "",
                status: rec.status || "",
            };
            alert("Check-out berhasil");
            dispatch("refresh");
        } catch (e) {
            console.error(e);
            alert(e?.message || "Check-out failed");
        } finally {
            checking = false;
        }
    }
</script>

<header class="main-header">
    <div class="header-content">
        <slot></slot>
        <!-- Page Title/Subtitle -->
    </div>

    <div class="header-actions">
        <!-- Attendance -->
        <div class="attendance-actions">
            <button class="btn-icon" on:click={checkin} title="Check-in" disabled={checking}>🟢</button>
            <button class="btn-icon" on:click={checkout} title="Check-out" disabled={checking}>🔴</button>
            <span class="att-time" id="att-time">{attText()}</span>
        </div>

        <!-- Secure Session Pill -->
        <div class="header-pill">
            <span class="pill-dot"></span>
            <span>Secure Session</span>
        </div>

        <!-- Theme Toggle -->
        <button
            class="btn-icon"
            on:click={toggleTheme}
            title={currentTheme === "dark"
                ? "Switch to Light Mode"
                : "Switch to Dark Mode"}
        >
            {currentTheme === "dark" ? "☀️" : "🌙"}
        </button>

        <!-- Logout -->
        <button
            class="btn-icon btn-logout-icon"
            on:click={confirmLogout}
            title="Logout"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><path d="M18.36 6.64a9 9 0 1 1-12.73 0"></path><line
                    x1="12"
                    y1="2"
                    x2="12"
                    y2="12"
                ></line></svg
            >
        </button>
    </div>
</header>

{#if showLogoutConfirm}
    <div class="logout-modal-backdrop">
        <div class="logout-modal">
            <div class="logout-icon-wrapper">🚪</div>
            <h3>Logout Confirmation</h3>
            <p>Are you sure you want to end your secure session?</p>
            <div class="logout-actions">
                <button class="btn-cancel" on:click={cancelLogout}
                    >Cancel</button
                >
                <button class="btn-confirm-logout" on:click={handleLogout}
                    >Yes, Logout</button
                >
            </div>
        </div>
    </div>
{/if}

<style>
    .main-header {
        flex-shrink: 0;
        z-index: 10;
        background: var(--bg-card);
        backdrop-filter: var(--glass-blur);
        -webkit-backdrop-filter: var(--glass-blur);
        padding: 20px 24px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        border-bottom: 1px solid var(--border-color);
        transition:
            background 0.3s,
            border-color 0.3s;
    }

    .header-content {
        display: flex;
        flex-direction: column;
    }

    .header-actions {
        display: flex;
        align-items: center;
        gap: 12px;
    }
    .attendance-actions {
        display: inline-flex;
        gap: 8px;
        align-items: center;
    }
    .att-time {
        margin-left: 4px;
        font-size: 12px;
        color: var(--text-muted);
    }

    /* Pill Styles */
    .header-pill {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        border-radius: 999px;
        padding: 6px 12px;
        background: var(--bg-hover);
        border: 1px solid var(--border-color);
        color: var(--text-body);
        font-size: 0.75rem;
        font-weight: 500;
    }

    .pill-dot {
        width: 6px;
        height: 6px;
        border-radius: 999px;
        background: var(--accent-success);
        box-shadow: 0 0 8px var(--accent-success);
    }

    /* Action Buttons */
    .btn-icon {
        background: var(--bg-hover);
        border: 1px solid var(--border-color);
        color: var(--text-body);
        width: 36px;
        height: 36px;
        border-radius: 8px;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        font-size: 1.1rem;
        transition: all 0.2s ease;
    }

    .btn-icon:hover {
        background: var(--border-color);
        color: var(--text-primary);
        transform: translateY(-1px);
    }

    .btn-logout-icon {
        background: rgba(239, 68, 68, 0.1);
        border-color: rgba(239, 68, 68, 0.3);
        color: var(--accent-danger);
    }

    .btn-logout-icon:hover {
        background: var(--accent-danger);
        color: white;
        border-color: var(--accent-danger);
        box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
    }

    /* Modal Styles (Scoped) */
    .logout-modal-backdrop {
        position: fixed;
        top: 0;
        left: 0;
        width: 100vw;
        height: 100vh;
        background: rgba(0, 0, 0, 0.6);
        backdrop-filter: blur(4px);
        z-index: 1000;
        display: flex;
        align-items: center;
        justify-content: center;
        animation: fadeIn 0.2s ease;
    }

    .logout-modal {
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        width: 400px;
        max-width: 90%;
        border-radius: 20px;
        padding: 32px;
        text-align: center;
        box-shadow: var(--shadow-card);
        animation: scaleIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    }

    .logout-icon-wrapper {
        font-size: 48px;
        margin-bottom: 16px;
    }

    .logout-modal h3 {
        color: var(--text-primary);
        font-size: 1.5rem;
        margin: 0 0 12px;
        font-weight: 700;
    }

    .logout-modal p {
        color: var(--text-muted);
        font-size: 1rem;
        margin: 0 0 32px;
        line-height: 1.5;
    }

    .logout-actions {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 16px;
    }

    .btn-cancel {
        background: transparent;
        border: 1px solid var(--text-muted);
        color: var(--text-body);
        padding: 12px;
        border-radius: 12px;
        font-size: 0.95rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-cancel:hover {
        background: var(--bg-hover);
        border-color: var(--text-primary);
        color: var(--text-primary);
    }

    .btn-confirm-logout {
        background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
        border: none;
        color: white;
        padding: 12px;
        border-radius: 12px;
        font-size: 0.95rem;
        font-weight: 600;
        cursor: pointer;
        box-shadow: 0 4px 6px -1px rgba(220, 38, 38, 0.2);
        transition: all 0.2s ease;
    }

    .btn-confirm-logout:hover {
        transform: translateY(-2px);
        box-shadow: 0 10px 15px -3px rgba(220, 38, 38, 0.3);
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }

    @keyframes scaleIn {
        from {
            transform: scale(0.95);
            opacity: 0;
        }
        to {
            transform: scale(1);
            opacity: 1;
        }
    }
</style>
