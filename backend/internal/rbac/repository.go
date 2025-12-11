package rbac

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// Menu represents a sidebar menu item
type Menu struct {
	ID             int    `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Path           string `json:"path,omitempty"`
	ParentCode     string `json:"parent_code,omitempty"`
	PermissionCode string `json:"-"`
	SortOrder      int    `json:"sort_order"`
	Children       []Menu `json:"children,omitempty"`
}

// Repository handles RBAC database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new RBAC repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetPermissionsByRoles returns all permission codes for given roles
func (r *Repository) GetPermissionsByRoles(ctx context.Context, roles []string) ([]string, error) {
	if len(roles) == 0 {
		return []string{}, nil
	}

	// Build query with IN clause
	query := `
		SELECT DISTINCT permission_code 
		FROM role_permissions 
		WHERE role_code = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(roles))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		permissions = append(permissions, code)
	}

	return permissions, rows.Err()
}

// GetMenusByPermissions returns menus that user has access to
func (r *Repository) GetMenusByPermissions(ctx context.Context, permissions []string) ([]Menu, error) {
	if len(permissions) == 0 {
		return []Menu{}, nil
	}

	query := `
		SELECT id, code, name, COALESCE(icon, ''), COALESCE(path, ''), 
		       COALESCE(parent_code, ''), COALESCE(permission_code, ''), sort_order
		FROM menus
		WHERE is_active = true 
		  AND (permission_code = ANY($1) OR permission_code IS NULL OR permission_code = '')
		ORDER BY sort_order, id
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(permissions))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var m Menu
		if err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.Icon, &m.Path,
			&m.ParentCode, &m.PermissionCode, &m.SortOrder); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}

	return menus, rows.Err()
}

// BuildMenuTree converts flat menu list to tree structure
func BuildMenuTree(menus []Menu) []Menu {
	menuMap := make(map[string]*Menu)
	var roots []Menu

	// First pass: create map
	for i := range menus {
		menus[i].Children = []Menu{}
		menuMap[menus[i].Code] = &menus[i]
	}

	// Second pass: build tree
	for i := range menus {
		if menus[i].ParentCode == "" {
			roots = append(roots, menus[i])
		} else if parent, ok := menuMap[menus[i].ParentCode]; ok {
			parent.Children = append(parent.Children, menus[i])
		}
	}

	// Update roots with their children
	for i := range roots {
		if m, ok := menuMap[roots[i].Code]; ok {
			roots[i].Children = m.Children
		}
	}

	return roots
}
