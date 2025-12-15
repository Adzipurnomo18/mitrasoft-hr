<script>
  import { API_BASE } from "./api.js";
  import { user, menus, permissions, currentPage } from "./stores.js";
  import Login from "./pages/Login.svelte";
  import Dashboard from "./pages/Dashboard.svelte";
  import Employees from "./pages/Employees.svelte";
  import Profile from "./pages/Profile.svelte";
  import Requests from "./pages/Requests.svelte";
  import Inbox from "./pages/Inbox.svelte";
  import UserManagement from "./pages/UserManagement.svelte";
  import PermissionsPage from "./pages/Permissions.svelte";
  import Announcements from "./pages/Announcements.svelte";
  import Reports from "./pages/Reports.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import "./styles/app.css";

  // Helper to determine page from URL
  function getPageFromUrl() {
    const path = window.location.pathname;
    if (path === "/" || path === "/dashboard") return "DASHBOARD";
    if (path === "/profile") return "MY_PROFILE";
    if (path === "/inbox") return "INBOX";
    if (path === "/employees") return "EMPLOYEES";
    // Fallback for others
    if (path.length > 1) {
      return path
        .replace(/^\//, "")
        .replace(/\//g, "-")
        .toUpperCase()
        .replace("-", "_");
    }
    return "DASHBOARD";
  }

  let currentUser = null;
  let page = getPageFromUrl();

  // Subscribe to stores
  user.subscribe((value) => (currentUser = value));
  currentPage.subscribe((value) => (page = value));

  async function handleLoggedIn(event) {
    const userData = event.detail;
    user.set(userData);
    currentPage.set("DASHBOARD");

    // Load user permissions and menus after login
    await loadUserData();
  }

  async function loadUserData() {
    try {
      // Load permissions
      const permRes = await fetch(`${API_BASE}/api/me/permissions`, {
        credentials: "include",
      });
      if (permRes.ok) {
        const permData = await permRes.json();
        permissions.set(permData.permissions || []);
      }

      // Load menus
      const menuRes = await fetch(`${API_BASE}/api/me/menus`, {
        credentials: "include",
      });
      if (menuRes.ok) {
        const menuData = await menuRes.json();
        menus.set(menuData || []);
      }
    } catch (e) {
      console.error("Failed to load user data:", e);
    }
  }

  function handleLogout() {
    user.set(null);
    menus.set([]);
    permissions.set([]);
    currentPage.set("DASHBOARD");
  }

  function handleNavigate(event) {
    const path = event.detail;
    // CurrentPage is already updated by Sidebar or the caller.
    // We only need to handle side effects here if any (like history.pushState if we used it)
    try {
      if (typeof path === "string") {
        window.history.replaceState({}, "", path);
      }
    } catch (e) {
      console.error("Navigation side-effect error:", e);
    }
  }
</script>

{#if !currentUser}
  <Login on:loggedIn={handleLoggedIn} />
{:else}
  <div class="app-layout">
    <Sidebar
      activePage={page}
      on:logout={handleLogout}
      on:navigate={handleNavigate}
    />

    <main class="app-main">
      {#if page === "dashboard" || page === "DASHBOARD"}
        <Dashboard />
      {:else if page === "employees" || page === "EMPLOYEES"}
        <Employees {page} />
      {:else if page === "MY_PROFILE" || page === "profile"}
        <Profile />
      {:else if page === "inbox" || page === "INBOX"}
        <Inbox />
      {:else if page === "USER_MANAGEMENT" || page === "user-management"}
        <UserManagement />
      {:else if page === "PERMISSIONS" || page === "permissions"}
        <PermissionsPage />
      {:else if page === "ANNOUNCEMENTS" || page === "announcements"}
        <Announcements />
      {:else if page === "REPORTS" || page === "reports"}
        <Reports />
      {:else if ["SELF_SERVICE", "APPROVALS", "MY_REQUESTS", "LEAVE_REQUEST", "RESIGN_REQUEST", "OVERTIME_REQ", "EXIT_CLEARANCE", "MEDICAL_CLAIM_REQ", "ASSESSMENT"].includes(page)}
        <Requests {page} />
      {:else}
        <div class="coming-soon">
          <h1>🚧 Coming Soon</h1>
          <p>Halaman <strong>{page}</strong> sedang dalam pengembangan.</p>
          <p class="muted">Fitur ini akan tersedia di update selanjutnya.</p>
        </div>
      {/if}
    </main>
  </div>
{/if}
