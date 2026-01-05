package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Roles        []Role    `json:"roles"`
	CreatedAt    time.Time `json:"created_at"`
}

type PDF struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Filename    string    `json:"filename"`
	FilePath    string    `json:"file_path"`
	UploadedBy  int       `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserPDFAccess struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	PDFID      int       `json:"pdf_id"`
	AccessedAt time.Time `json:"accessed_at"`
}

type Role struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
}

type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserRole struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	RoleID     int       `json:"role_id"`
	AssignedAt time.Time `json:"assigned_at"`
	AssignedBy *int      `json:"assigned_by"`
}
