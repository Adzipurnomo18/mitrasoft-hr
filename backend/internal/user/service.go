package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"hr-portal-backend/pkg/crypto"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("user not found")
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

// ==========================
// Auth
// ==========================

func (s *Service) Authenticate(email, password string) (*User, error) {
	u, err := s.repo.FindByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := crypto.ComparePassword(u.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (s *Service) GetByID(id int64) (*User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// ==========================
// Employees CRUD
// ==========================

type EmployeeInput struct {
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Branch       string `json:"branch"`
	JobTitle     string `json:"job_title"`
	Status       string `json:"status"`     // ACTIVE / INACTIVE
	Department   string `json:"department"` // ADM, IT, ACC, etc.
}

func (in *EmployeeInput) sanitize() {
	in.EmployeeCode = strings.TrimSpace(in.EmployeeCode)
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Branch = strings.TrimSpace(in.Branch)
	in.JobTitle = strings.TrimSpace(in.JobTitle)
	in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	in.Department = strings.ToUpper(strings.TrimSpace(in.Department))
	if in.Status == "" {
		in.Status = "ACTIVE"
	}
}

func (s *Service) ListEmployees(ctx context.Context, search string) ([]User, error) {
	return s.repo.ListEmployees(ctx, search, 0, 0)
}

func (s *Service) CreateEmployee(ctx context.Context, in EmployeeInput) (*User, error) {
	in.sanitize()

	// default roles untuk user baru
	defaultRoles := []string{"EMPLOYEE"}

	// hash password default
	hash, err := crypto.HashPassword("changeme123")
	if err != nil {
		return nil, err
	}

	u := &User{
		EmployeeCode: in.EmployeeCode,
		Name:         in.Name,
		Email:        in.Email,
		Branch:       in.Branch,
		JobTitle:     in.JobTitle,
		Status:       in.Status,
		Department:   in.Department,
		Roles:        defaultRoles,
		PasswordHash: hash,
	}

	return s.repo.CreateEmployee(ctx, u)
}

func (s *Service) UpdateEmployee(ctx context.Context, id int64, in EmployeeInput) (*User, error) {
	in.sanitize()

	existing, err := s.repo.FindByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	existing.EmployeeCode = in.EmployeeCode
	existing.Name = in.Name
	existing.Email = in.Email
	existing.Branch = in.Branch
	existing.JobTitle = in.JobTitle
	existing.Status = in.Status
	existing.Department = in.Department

	return s.repo.UpdateEmployee(ctx, existing)
}

// GetNextEmployeeCode returns the next available employee code for a department.
func (s *Service) GetNextEmployeeCode(ctx context.Context, department string) (string, error) {
	return s.repo.GetNextEmployeeCode(ctx, department)
}

func (s *Service) DeleteEmployee(ctx context.Context, id int64) error {
	return s.repo.DeleteEmployee(ctx, id)
}

// UpdateProfile updates the current user's own profile
func (s *Service) UpdateProfile(ctx context.Context, userID int64, input *ProfileInput) (*User, error) {
	// Sanitize input
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Address = strings.TrimSpace(input.Address)
	input.Gender = strings.ToUpper(strings.TrimSpace(input.Gender))
	input.EmergencyContact = strings.TrimSpace(input.EmergencyContact)
	input.EmergencyPhone = strings.TrimSpace(input.EmergencyPhone)

	if input.Name == "" {
		return nil, errors.New("name is required")
	}

	return s.repo.UpdateProfile(ctx, userID, input)
}
