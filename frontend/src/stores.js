import { writable } from 'svelte/store';

// User store - holds current logged-in user data
export const user = writable(null);

// Menus store - holds dynamic menu tree from backend
export const menus = writable([]);

// Permissions store - holds user's permission codes
export const permissions = writable([]);

// Current page store
export const currentPage = writable('dashboard');

// Loading state
export const isLoading = writable(false);

// Theme Store
const initialTheme = localStorage.getItem('theme') || 'dark';
export const theme = writable(initialTheme);

theme.subscribe(val => {
    localStorage.setItem('theme', val);
    if (typeof document !== 'undefined') {
        document.documentElement.setAttribute('data-theme', val);
    }
});
