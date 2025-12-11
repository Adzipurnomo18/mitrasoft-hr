<script>
    import { createEventDispatcher } from "svelte";
    import { API_BASE } from "../api.js";
    import { menus, permissions, currentPage, user } from "../stores.js"; // Removed theme
    import "../styles/sidebar.css";

    const dispatch = createEventDispatcher();

    export let activePage = "dashboard";

    let menuTree = [];
    let displayMenus = [];
    let userPerms = [];
    let expandedMenus = {};
    let currentUser = null;
    let unreadCount = 0;

    // Subscribe to stores
    menus.subscribe((value) => (menuTree = value));
    permissions.subscribe((value) => (userPerms = value));
    user.subscribe((value) => (currentUser = value));

    $: {
        const selfSvc = (menuTree || []).find((m) => m.code === "SELF_SERVICE");
        const requestsGroup = (menuTree || []).find((m) => m.code === "REQUESTS");
        const others = (menuTree || []).filter((m) => m.code !== "REQUESTS" && m.code !== "SELF_SERVICE");

        const selfChildren = Array.isArray(selfSvc?.children) ? selfSvc.children : [];
        const reqChildren = Array.isArray(requestsGroup?.children) ? requestsGroup.children : [];

        const mergedMap = new Map();
        for (const c of [...selfChildren, ...reqChildren]) {
            mergedMap.set(c.code, c);
        }
        const myRequests = (reqChildren || []).find((c) => c.code === "MY_REQUESTS")
            || { code: "MY_REQUESTS", name: "My Requests", icon: "list", path: "/requests/history" };

        const mergedSelf = selfSvc
            ? { ...selfSvc, children: [myRequests] }
            : { code: "SELF_SERVICE", name: "Self-Service", icon: "briefcase", children: [myRequests] };

        const idxEmp = others.findIndex((m) => m.code === "EMPLOYEES");
        if (idxEmp >= 0) {
            displayMenus = [
                ...others.slice(0, idxEmp + 1),
                mergedSelf,
                ...others.slice(idxEmp + 1),
            ];
        } else {
            displayMenus = [...others, mergedSelf];
        }
    }

    // Load menus from backend
    async function loadMenus() {
        try {
            const res = await fetch(`${API_BASE}/api/me/menus`, {
                credentials: "include",
            });

            if (res.ok) {
                const data = await res.json();
                menus.set(data || []);
            }
        } catch (e) {
            console.error("Failed to load menus:", e);
        }
    }

    // Fetch unread messages
    async function loadUnreadCount() {
        try {
            const res = await fetch(`${API_BASE}/api/inbox`, {
                credentials: "include",
            });
            if (res.ok) {
                const messages = await res.json();
                if (Array.isArray(messages)) {
                    unreadCount = messages.filter((m) => !m.is_read).length;
                }
            }
        } catch (e) {
            console.error("Failed to load unread count", e);
        }
    }

    import { onMount, onDestroy } from "svelte";

    // Polling for unread messages every 30 seconds to keep it updated
    let pollInterval;

    onMount(() => {
        loadMenus();
        loadUnreadCount();
        pollInterval = setInterval(loadUnreadCount, 30000);
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });

    // Toggle submenu expansion
    function toggleMenu(menuCode) {
        expandedMenus[menuCode] = !expandedMenus[menuCode];
        expandedMenus = expandedMenus; // trigger reactivity
    }

    // Navigate to a page
    function navigate(path, menuCode) {
        if (path) {
            currentPage.set(menuCode);
            dispatch("navigate", path);

            // If navigating to inbox, refresh count after a short delay
            if (menuCode === "INBOX") {
                setTimeout(loadUnreadCount, 1000);
            }
        }
    }

    // Icon mapping (using simple text for now, can replace with actual icons)
    const iconMap = {
        home: "🏠",
        user: "👤",
        mail: "📧",
        briefcase: "💼",
        calendar: "📅",
        "log-out": "🚪",
        "file-text": "📄",
        clock: "⏰",
        list: "📋",
        users: "👥",
        "check-circle": "✅",
        megaphone: "📢",
        "bar-chart": "📊",
        settings: "⚙️",
        "user-plus": "➕",
        shield: "🛡️",
    };

    function getIcon(iconName) {
        return iconMap[iconName] || "📁";
    }

    // Check if menu is active
    function isActive(menuCode) {
        if (!activePage || !menuCode) return false;
        // Simple case-insensitive comparison
        const normalizedActivePage = activePage.toLowerCase();
        const normalizedMenuCode = menuCode.toLowerCase();
        const result = normalizedActivePage === normalizedMenuCode;
        if (result || menuCode === "DASHBOARD") {
            // console.log(`isActive check: ${menuCode} vs ${activePage} = ${result}`);
        }
        return result;
    }
</script>

<aside class="sidebar">
    <div class="sidebar-header">
        <div class="sidebar-logo">
            <span class="logo-dot"></span>
            <span class="logo-text">Mitrasoft-HR</span>
        </div>
        <p class="sidebar-company">PT Mitrasoft</p>
    </div>

    {#if currentUser}
        <div class="sidebar-user">
            <p class="sidebar-user-hi">
                Hello, {currentUser.name || "Employee"}!
            </p>
            <p class="sidebar-user-email">{currentUser.email || ""}</p>
            <p class="sidebar-user-role">
                {#if currentUser.roles && currentUser.roles.length > 0}
                    <span class="role-badge">{currentUser.roles[0]}</span>
                {/if}
            </p>
        </div>
    {/if}

    <nav class="sidebar-nav">
        {#each displayMenus as menu}
            {#if menu.children && menu.children.length > 0}
                <!-- Parent menu with children -->
                <div class="nav-group">
                    <button
                        class="nav-item nav-parent"
                        class:expanded={expandedMenus[menu.code]}
                        on:click={() => toggleMenu(menu.code)}
                    >
                        <span class="nav-icon">{getIcon(menu.icon)}</span>
                        <span class="nav-label">{menu.name}</span>
                        <span class="nav-arrow"
                            >{expandedMenus[menu.code] ? "▼" : "▶"}</span
                        >
                    </button>

                    {#if expandedMenus[menu.code]}
                        <div class="nav-children">
                            {#each menu.children as child}
                                <button
                                    class="nav-item nav-child"
                                    class:nav-item-active={activePage.toLowerCase() ===
                                        child.code.toLowerCase()}
                                    on:click={() =>
                                        navigate(child.path, child.code)}
                                >
                                    <span class="nav-icon"
                                        >{getIcon(child.icon)}</span
                                    >
                                    <span class="nav-label">{child.name}</span>
                                    {#if child.code === "INBOX" && unreadCount > 0}
                                        <span class="nav-badge"
                                            >{unreadCount > 99
                                                ? "99+"
                                                : unreadCount}</span
                                        >
                                    {/if}
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>
            {:else if menu.path}
                <!-- Single menu item -->
                <button
                    class="nav-item"
                    class:nav-item-active={activePage.toLowerCase() ===
                        menu.code.toLowerCase()}
                    on:click={() => navigate(menu.path, menu.code)}
                >
                    <span class="nav-icon">{getIcon(menu.icon)}</span>
                    <span class="nav-label">{menu.name}</span>
                    {#if menu.code === "INBOX" && unreadCount > 0}
                        <span class="nav-badge"
                            >{unreadCount > 99 ? "99+" : unreadCount}</span
                        >
                    {/if}
                </button>
            {/if}
        {/each}
    </nav>

    <div class="sidebar-footer">
        <p class="sidebar-footnote">v0.2 • Mitrasoft-HR</p>
    </div>
</aside>
