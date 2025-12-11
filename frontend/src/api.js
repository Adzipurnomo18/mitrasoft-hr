// API Base URL - uses the same hostname as the current page to avoid cross-origin cookie issues
// e.g., if accessed via 127.0.0.1:5173, API will be 127.0.0.1:8080
// if accessed via localhost:5173, API will be localhost:8080
export const API_BASE = `http://${window.location.hostname}:8080`;
