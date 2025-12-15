package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// Repository membungkus akses ke tabel users.
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Kolom yang dipakai di semua SELECT.
// PERHATIAN: Sengaja TANPA created_at / updated_at.
const userSelectColumns = `
	id,
	employee_code,
	name,
	email,
	COALESCE(branch, ''),
	COALESCE(job_title, ''),
	status,
	COALESCE(department, ''),
	COALESCE(roles, '{}'),
	password_hash,
	COALESCE(phone, ''),
	COALESCE(address, ''),
	birth_date,
	COALESCE(gender, ''),
	COALESCE(photo_url, ''),
	join_date,
	COALESCE(emergency_contact, ''),
	COALESCE(emergency_phone, '')
`

// helper untuk scan row menjadi struct User.
func scanUser(row interface{ Scan(dest ...any) error }) (*User, error) {
	var u User
	var roles []string
	var birthDate, joinDate sql.NullTime

	err := row.Scan(
		&u.ID,
		&u.EmployeeCode,
		&u.Name,
		&u.Email,
		&u.Branch,
		&u.JobTitle,
		&u.Status,
		&u.Department,
		pq.Array(&roles),
		&u.PasswordHash,
		&u.Phone,
		&u.Address,
		&birthDate,
		&u.Gender,
		&u.PhotoURL,
		&joinDate,
		&u.EmergencyContact,
		&u.EmergencyPhone,
	)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	if joinDate.Valid {
		u.JoinDate = &joinDate.Time
	}
	return &u, nil
}

// ==========================
// Query single
// ==========================

func (r *Repository) FindByEmail(email string) (*User, error) {
	q := `
		SELECT ` + userSelectColumns + `
		FROM users
		WHERE email = $1
		ORDER BY id DESC
		LIMIT 1
	`
	row := r.db.QueryRow(q, strings.ToLower(strings.TrimSpace(email)))
	return scanUser(row)
}

func (r *Repository) FindByID(id int64) (*User, error) {
	q := `
		SELECT ` + userSelectColumns + `
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRow(q, id)
	return scanUser(row)
}

// ==========================
// Employees list (untuk page Employees)
// ==========================

func (r *Repository) ListEmployees(ctx context.Context, search string, limit, offset int) ([]User, error) {
	search = strings.TrimSpace(search)

	// default paging kalau mau nanti diaktifkan
	if limit <= 0 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var (
		rows *sql.Rows
		err  error
	)

	if search == "" {
		q := `
			SELECT ` + userSelectColumns + `
			FROM users
			ORDER BY employee_code ASC, id ASC
			LIMIT $1 OFFSET $2
		`
		rows, err = r.db.QueryContext(ctx, q, limit, offset)
	} else {
		like := "%" + strings.ToLower(search) + "%"
		q := `
			SELECT ` + userSelectColumns + `
			FROM users
			WHERE
				LOWER(name) LIKE $1
				OR LOWER(email) LIKE $1
				OR LOWER(employee_code) LIKE $1
				OR LOWER(COALESCE(branch, '')) LIKE $1
			ORDER BY employee_code ASC, id ASC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.QueryContext(ctx, q, like, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// ==========================
// Employees CRUD
// ==========================

func (r *Repository) CreateEmployee(ctx context.Context, u *User) (*User, error) {
	q := `
		INSERT INTO users (
			employee_code,
			name,
			email,
			branch,
			job_title,
			status,
			department,
			roles,
			password_hash
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING ` + userSelectColumns + `
	`
	row := r.db.QueryRowContext(
		ctx,
		q,
		u.EmployeeCode,
		u.Name,
		strings.ToLower(u.Email),
		u.Branch,
		u.JobTitle,
		u.Status,
		u.Department,
		pq.Array(u.Roles),
		u.PasswordHash,
	)
	return scanUser(row)
}

func (r *Repository) UpdateEmployee(ctx context.Context, u *User) (*User, error) {
	q := `
		UPDATE users
		SET
			employee_code = $1,
			name          = $2,
			email         = $3,
			branch        = $4,
			job_title     = $5,
			status        = $6,
			department    = $7,
			roles         = $8,
			password_hash = COALESCE(NULLIF($9, ''), password_hash)
		WHERE id = $10
		RETURNING ` + userSelectColumns + `
	`
	row := r.db.QueryRowContext(
		ctx,
		q,
		u.EmployeeCode,
		u.Name,
		strings.ToLower(u.Email),
		u.Branch,
		u.JobTitle,
		u.Status,
		u.Department,
		pq.Array(u.Roles),
		u.PasswordHash,
		u.ID,
	)
	return scanUser(row)
}

func (r *Repository) DeleteEmployee(ctx context.Context, id int64) error {
	// Soft delete: mark as INACTIVE to avoid FK constraint issues
	q := `UPDATE users SET status = 'INACTIVE' WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmployeeByCode performs soft delete by employee code
func (r *Repository) DeleteEmployeeByCode(ctx context.Context, code string) error {
	q := `UPDATE users SET status = 'INACTIVE' WHERE employee_code = $1`
	res, err := r.db.ExecContext(ctx, q, code)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HardDeleteEmployee permanently deletes a user by id
func (r *Repository) HardDeleteEmployee(ctx context.Context, id int64) error {
	q := `DELETE FROM users WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HardDeleteEmployeeByCode permanently deletes a user by employee code
func (r *Repository) HardDeleteEmployeeByCode(ctx context.Context, code string) error {
	q := `DELETE FROM users WHERE employee_code = $1`
	res, err := r.db.ExecContext(ctx, q, code)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetNextEmployeeCode generates the next employee code for a department.
// Format: PREFIX + 3-digit number (e.g., IT001, ADM002)
func (r *Repository) GetNextEmployeeCode(ctx context.Context, department string) (string, error) {
	department = strings.ToUpper(strings.TrimSpace(department))
	if department == "" {
		department = "EMP" // default prefix
	}

	// Find the max number for this department prefix
	q := `
		SELECT COALESCE(MAX(
			CAST(SUBSTRING(employee_code FROM LENGTH($1) + 1) AS INTEGER)
		), 0)
		FROM users
		WHERE employee_code LIKE $1 || '%'
		  AND employee_code ~ ('^' || $1 || '[0-9]+$')
	`

	var maxNum int
	err := r.db.QueryRowContext(ctx, q, department).Scan(&maxNum)
	if err != nil {
		return "", err
	}

	// Generate next code with 3-digit padding
	nextCode := department + fmt.Sprintf("%03d", maxNum+1)
	return nextCode, nil
}

// UpdateProfile updates user's own profile fields
func (r *Repository) UpdateProfile(ctx context.Context, userID int64, input *ProfileInput) (*User, error) {
	q := `
		UPDATE users
		SET
			name = $1,
			phone = $2,
			address = $3,
			birth_date = $4,
			gender = $5,
			emergency_contact = $6,
			emergency_phone = $7
		WHERE id = $8
		RETURNING ` + userSelectColumns + `
	`

	// Parse birth_date
	var birthDate interface{}
	if input.BirthDate != "" {
		birthDate = input.BirthDate
	} else {
		birthDate = nil
	}

	row := r.db.QueryRowContext(
		ctx,
		q,
		input.Name,
		input.Phone,
		input.Address,
		birthDate,
		input.Gender,
		input.EmergencyContact,
		input.EmergencyPhone,
		userID,
	)
	return scanUser(row)
}

// GetDepartment returns only the department code for a user
func (r *Repository) GetDepartment(ctx context.Context, userID int64) (string, error) {
	var dept sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT department FROM users WHERE id = $1", userID).Scan(&dept)
	if err != nil {
		return "", err
	}
	return dept.String, nil
}
