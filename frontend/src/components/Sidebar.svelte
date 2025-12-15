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
        const upper = (s) => String(s || "").toUpperCase();
        const isAdminMenu = (m) => {
            const code = upper(m.code);
            const name = upper(m.name);
            return code === "ADMINISTRATION" || code === "ADMIN" || name === "ADMINISTRATION";
        };

        const selfSvc = (menuTree || []).find((m) => upper(m.code) === "SELF_SERVICE");
        const requestsGroup = (menuTree || []).find((m) => upper(m.code) === "REQUESTS");
        const others = (menuTree || []).filter((m) => upper(m.code) !== "REQUESTS" && upper(m.code) !== "SELF_SERVICE");

        const selfChildren = Array.isArray(selfSvc?.children) ? selfSvc.children : [];
        const reqChildren = Array.isArray(requestsGroup?.children) ? requestsGroup.children : [];

        const mergedMap = new Map();
        for (const c of [...selfChildren, ...reqChildren]) {
            mergedMap.set(c.code, c);
        }
        const myRequests = (reqChildren || []).find((c) => upper(c.code) === "MY_REQUESTS")
            || { code: "MY_REQUESTS", name: "My Requests", icon: "list", path: "/requests/history" };
        const approvalsRoot = others.find((m) => upper(m.code) === "APPROVALS");
        const approvalsChild =
            approvalsRoot
                ? { code: approvalsRoot.code, name: approvalsRoot.name, icon: approvalsRoot.icon, path: approvalsRoot.path }
                : { code: "APPROVALS", name: "Approvals", icon: "check-circle", path: "/approvals" };

        const privileged =
            Array.isArray(currentUser?.roles) &&
            (currentUser.roles
                .map((r) => String(r).toUpperCase())
                .some((r) => ["ADMIN", "IT", "HRD", "IT_ADMIN", "HR_ADMIN"].includes(r)));

        const mergedSelf = selfSvc
            ? { ...selfSvc, children: privileged ? [myRequests, approvalsChild] : [myRequests] }
            : { code: "SELF_SERVICE", name: "Self-Service", icon: "briefcase", children: privileged ? [myRequests, approvalsChild] : [myRequests] };

        const adminExisting = others.find(isAdminMenu);
        const filteredOthers = others.filter((m) => {
            const codeUp = upper(m.code);
            const isAdmin = isAdminMenu(m);
            // Hapus admin hanya jika kita akan menyisipkan satu instance (privileged)
            if (isAdmin && !privileged) return false; // non-privileged: sembunyikan admin
            if (isAdmin && privileged) return false;  // privileged: buang instance lama, ganti satu di posisi yang benar
            return codeUp !== "APPROVALS";
        });
        const adminGroup = privileged
            ? (adminExisting
                ? adminExisting
                : {
                      code: "ADMINISTRATION",
                      name: "Administration",
                      icon: "settings",
                      children: [
                          { code: "USER_MANAGEMENT", name: "User Management", icon: "user-plus", path: "/user-management" },
                          { code: "PERMISSIONS", name: "Permissions", icon: "shield", path: "/permissions" },
                      ],
                  })
            : null;
        const idxEmployees = filteredOthers.findIndex((m) => upper(m.code) === "EMPLOYEES");
        if (idxEmployees >= 0) {
            displayMenus = [
                ...filteredOthers.slice(0, idxEmployees + 1),
                mergedSelf,
                ...(adminGroup ? [adminGroup] : []),
                ...filteredOthers.slice(idxEmployees + 1),
            ];
        } else {
            displayMenus = [...filteredOthers, mergedSelf, ...(adminGroup ? [adminGroup] : [])];
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
        try {
        const aliasByCode = {
            ADMIN_USERS: "USER_MANAGEMENT",
            USER_MGMT: "USER_MANAGEMENT",
            ADMIN_PERMISSIONS: "PERMISSIONS",
            PERMISSION_MGMT: "PERMISSIONS",
            REQUESTS_HISTORY: "MY_REQUESTS",
            APPROVALS: "APPROVALS",
            MY_PROFILE: "MY_PROFILE",
            DASHBOARD: "DASHBOARD",
            EMPLOYEES: "EMPLOYEES",
            INBOX: "INBOX",
        };
        const aliasByPath = {
            "/user-management": "USER_MANAGEMENT",
            "/permissions": "PERMISSIONS",
            "/admin/users": "USER_MANAGEMENT",
            "/admin/permissions": "PERMISSIONS",
            "/requests/history": "MY_REQUESTS",
            "/approvals": "APPROVALS",
            "/profile": "MY_PROFILE",
            "/dashboard": "DASHBOARD",
            "/employees": "EMPLOYEES",
            "/inbox": "INBOX",
        };

        let target = null;
        if (menuCode) {
            const code = String(menuCode).toUpperCase();
            target = aliasByCode[code] || code;
        } else if (path) {
            target =
                aliasByPath[path] ||
                path.replace(/^\//, "").replace(/\//g, "-").toUpperCase().replace("-", "_");
        }

        if (target) {
            currentPage.set(target);
            if (path) dispatch("navigate", path);
            if (target === "INBOX") {
                setTimeout(loadUnreadCount, 1000);
            }
        }
        } catch (e) {
            console.error("Navigation error:", e);
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
        <p class="sidebar-company">PT Surya Pratama Keramindo</p>
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
