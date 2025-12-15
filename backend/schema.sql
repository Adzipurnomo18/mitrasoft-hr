-- Minimal schema (users, roles, user_roles)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    employee_code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    branch VARCHAR(50),
    job_title VARCHAR(100),
    status VARCHAR(20) DEFAULT 'ACTIVE',
    department VARCHAR(10),
    roles TEXT[] DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- =============================================
-- RBAC: Permissions & Menus
-- =============================================

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    module VARCHAR(50) -- group permissions by module
);

-- Role permissions mapping
CREATE TABLE IF NOT EXISTS role_permissions (
    role_code VARCHAR(50) NOT NULL,
    permission_code VARCHAR(50) NOT NULL,
    PRIMARY KEY (role_code, permission_code)
);

-- Menus table for dynamic sidebar
CREATE TABLE IF NOT EXISTS menus (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(50),
    path VARCHAR(100),
    parent_code VARCHAR(50),
    permission_code VARCHAR(50),
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true
);

-- =============================================
-- SEED DATA: Roles
-- =============================================
INSERT INTO roles (code, name) VALUES 
    ('EMPLOYEE', 'Karyawan'),
    ('HRD', 'HRD Staff'),
    ('IT_ADMIN', 'IT Administrator')
ON CONFLICT (code) DO NOTHING;

-- =============================================
-- SEED DATA: Permissions
-- =============================================
INSERT INTO permissions (code, name, description, module) VALUES 
    -- Dashboard
    ('VIEW_DASHBOARD', 'View Dashboard', 'Akses halaman dashboard', 'dashboard'),
    
    -- Profile
    ('VIEW_PROFILE', 'View Profile', 'Lihat profil sendiri', 'profile'),
    ('EDIT_PROFILE', 'Edit Profile', 'Edit profil sendiri', 'profile'),
    
    -- Inbox
    ('VIEW_INBOX', 'View Inbox', 'Lihat pesan masuk', 'inbox'),
    
    -- Self Service
    ('REQUEST_LEAVE', 'Request Leave', 'Ajukan cuti', 'self_service'),
    ('REQUEST_RESIGN', 'Request Resignation', 'Ajukan resign', 'self_service'),
    
    -- Requests
    ('REQUEST_OVERTIME', 'Request Overtime', 'Ajukan lembur', 'requests'),
    ('VIEW_REQUESTS', 'View Requests', 'Lihat status pengajuan', 'requests'),
    
    -- Employee Management (HRD/Admin)
    ('VIEW_EMPLOYEES', 'View Employees', 'Lihat daftar karyawan', 'employees'),
    ('MANAGE_EMPLOYEES', 'Manage Employees', 'Tambah/edit/hapus karyawan', 'employees'),
    
    -- Approvals (HRD)
    ('APPROVE_LEAVE', 'Approve Leave', 'Approve/reject cuti', 'approvals'),
    ('APPROVE_OVERTIME', 'Approve Overtime', 'Approve/reject lembur', 'approvals'),
    
    -- Admin
    ('MANAGE_PERMISSIONS', 'Manage Permissions', 'Kelola permission roles', 'admin'),
    ('MANAGE_USERS', 'Manage Users', 'Kelola user accounts', 'admin'),
    ('VIEW_REPORTS', 'View Reports', 'Lihat laporan', 'reports'),
    
    -- Announcements
    ('VIEW_ANNOUNCEMENTS', 'View Announcements', 'Lihat pengumuman', 'announcements'),
    ('CREATE_ANNOUNCEMENTS', 'Create Announcements', 'Buat pengumuman', 'announcements')
ON CONFLICT (code) DO NOTHING;

-- =============================================
-- SEED DATA: Role Permissions
-- =============================================

-- EMPLOYEE permissions
INSERT INTO role_permissions (role_code, permission_code) VALUES 
    ('EMPLOYEE', 'VIEW_DASHBOARD'),
    ('EMPLOYEE', 'VIEW_PROFILE'),
    ('EMPLOYEE', 'EDIT_PROFILE'),
    ('EMPLOYEE', 'VIEW_INBOX'),
    ('EMPLOYEE', 'REQUEST_LEAVE'),
    ('EMPLOYEE', 'REQUEST_OVERTIME'),
    ('EMPLOYEE', 'VIEW_REQUESTS'),
    ('EMPLOYEE', 'VIEW_ANNOUNCEMENTS')
ON CONFLICT DO NOTHING;

-- HRD permissions (includes all EMPLOYEE + more)
INSERT INTO role_permissions (role_code, permission_code) VALUES 
    ('HRD', 'VIEW_DASHBOARD'),
    ('HRD', 'VIEW_PROFILE'),
    ('HRD', 'EDIT_PROFILE'),
    ('HRD', 'VIEW_INBOX'),
    ('HRD', 'REQUEST_LEAVE'),
    ('HRD', 'REQUEST_OVERTIME'),
    ('HRD', 'VIEW_REQUESTS'),
    ('HRD', 'VIEW_ANNOUNCEMENTS'),
    ('HRD', 'VIEW_EMPLOYEES'),
    ('HRD', 'MANAGE_EMPLOYEES'),
    ('HRD', 'APPROVE_LEAVE'),
    ('HRD', 'APPROVE_OVERTIME'),
    ('HRD', 'CREATE_ANNOUNCEMENTS'),
    ('HRD', 'VIEW_REPORTS')
ON CONFLICT DO NOTHING;

-- IT_ADMIN permissions (full access)
INSERT INTO role_permissions (role_code, permission_code) VALUES 
    ('IT_ADMIN', 'VIEW_DASHBOARD'),
    ('IT_ADMIN', 'VIEW_PROFILE'),
    ('IT_ADMIN', 'EDIT_PROFILE'),
    ('IT_ADMIN', 'VIEW_INBOX'),
    ('IT_ADMIN', 'REQUEST_LEAVE'),
    ('IT_ADMIN', 'REQUEST_OVERTIME'),
    ('IT_ADMIN', 'VIEW_REQUESTS'),
    ('IT_ADMIN', 'VIEW_ANNOUNCEMENTS'),
    ('IT_ADMIN', 'VIEW_EMPLOYEES'),
    ('IT_ADMIN', 'MANAGE_EMPLOYEES'),
    ('IT_ADMIN', 'APPROVE_LEAVE'),
    ('IT_ADMIN', 'APPROVE_OVERTIME'),
    ('IT_ADMIN', 'CREATE_ANNOUNCEMENTS'),
    ('IT_ADMIN', 'VIEW_REPORTS'),
    ('IT_ADMIN', 'MANAGE_PERMISSIONS'),
    ('IT_ADMIN', 'MANAGE_USERS')
ON CONFLICT DO NOTHING;

-- =============================================
-- SEED DATA: Menus
-- =============================================
INSERT INTO menus (code, name, icon, path, parent_code, permission_code, sort_order) VALUES 
    -- Main menus
    ('DASHBOARD', 'Dashboard', 'home', '/dashboard', NULL, 'VIEW_DASHBOARD', 1),
    ('MY_PROFILE', 'My Profile', 'user', '/profile', NULL, 'VIEW_PROFILE', 2),
    ('INBOX', 'Inbox', 'mail', '/inbox', NULL, 'VIEW_INBOX', 3),
    
    -- Self Service (parent)
    ('SELF_SERVICE', 'Self-Service', 'briefcase', NULL, NULL, 'REQUEST_LEAVE', 4),
    ('LEAVE_REQUEST', 'Leave Request', 'calendar', '/self-service/leave', 'SELF_SERVICE', 'REQUEST_LEAVE', 1),
    ('RESIGN_REQUEST', 'Resignation', 'log-out', '/self-service/resign', 'SELF_SERVICE', 'REQUEST_RESIGN', 2),
    
    -- Requests (parent)
    ('REQUESTS', 'Requests', 'file-text', NULL, NULL, 'VIEW_REQUESTS', 5),
    ('OVERTIME_REQ', 'Overtime Request', 'clock', '/requests/overtime', 'REQUESTS', 'REQUEST_OVERTIME', 1),
    ('MY_REQUESTS', 'My Requests', 'list', '/requests/history', 'REQUESTS', 'VIEW_REQUESTS', 2),
    
    -- HRD/Admin menus
    ('EMPLOYEES', 'Employees', 'users', '/employees', NULL, 'VIEW_EMPLOYEES', 10),
    ('APPROVALS', 'Approvals', 'check-circle', '/approvals', NULL, 'APPROVE_LEAVE', 11),
    ('ANNOUNCEMENTS', 'Announcements', 'megaphone', '/announcements', NULL, 'CREATE_ANNOUNCEMENTS', 12),
    ('REPORTS', 'Reports', 'bar-chart', '/reports', NULL, 'VIEW_REPORTS', 13),
    
    -- Admin only
    ('ADMIN', 'Administration', 'settings', NULL, NULL, 'MANAGE_PERMISSIONS', 20),
    ('USER_MGMT', 'User Management', 'user-plus', '/admin/users', 'ADMIN', 'MANAGE_USERS', 1),
    ('PERMISSION_MGMT', 'Permissions', 'shield', '/admin/permissions', 'ADMIN', 'MANAGE_PERMISSIONS', 2)
ON CONFLICT (code) DO NOTHING;
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    sender_id INT REFERENCES users(id),
    receiver_id INT REFERENCES users(id),
    subject VARCHAR(200),
    body TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    parent_id INT REFERENCES messages(id)
);

CREATE TABLE IF NOT EXISTS announcements (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    target_departments VARCHAR(50)[],
    target_roles VARCHAR(50)[],
    created_by INT REFERENCES users(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS announcement_reads (
    user_id INT REFERENCES users(id),
    announcement_id INT REFERENCES announcements(id),
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, announcement_id)
);

CREATE TABLE IF NOT EXISTS requests (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    type VARCHAR(50) NOT NULL, -- LEAVE, OVERTIME, PERMIT
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    reason TEXT,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED
    approver_id INT REFERENCES users(id), -- Who approved/rejected
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Attendance table
CREATE TABLE IF NOT EXISTS attendance (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    checkin_time TIMESTAMP,
    checkout_time TIMESTAMP,
    status VARCHAR(20) DEFAULT 'ABSENT', -- ABSENT, ON_TIME, LATE
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, date)
);
