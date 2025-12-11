package user

import "time"

// User adalah representasi row di tabel "users".
type User struct {
	ID           int64    `json:"id"`
	EmployeeCode string   `json:"employee_code"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Branch       string   `json:"branch"`
	JobTitle     string   `json:"job_title"`
	Status       string   `json:"status"`
	Department   string   `json:"department"`
	Roles        []string `json:"roles"`
	PasswordHash string   `json:"-"` // jangan dibocorkan ke JSON

	// Profile fields
	Phone            string     `json:"phone,omitempty"`
	Address          string     `json:"address,omitempty"`
	BirthDate        *time.Time `json:"birth_date,omitempty"`
	Gender           string     `json:"gender,omitempty"`
	PhotoURL         string     `json:"photo_url,omitempty"`
	JoinDate         *time.Time `json:"join_date,omitempty"`
	EmergencyContact string     `json:"emergency_contact,omitempty"`
	EmergencyPhone   string     `json:"emergency_phone,omitempty"`
}

// ProfileInput is used for updating user's own profile
type ProfileInput struct {
	Name             string `json:"name"`
	Phone            string `json:"phone"`
	Address          string `json:"address"`
	BirthDate        string `json:"birth_date"` // YYYY-MM-DD format
	Gender           string `json:"gender"`
	EmergencyContact string `json:"emergency_contact"`
	EmergencyPhone   string `json:"emergency_phone"`
}
