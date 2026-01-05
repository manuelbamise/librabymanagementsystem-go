package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
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
