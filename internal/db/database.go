package db

import (
	"database/sql"
	"fmt"
	"librarymanagementsystem/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

func (d *Database) createTables() error {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	pdfTable := `
	CREATE TABLE IF NOT EXISTS pdfs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT,
		description TEXT,
		filename TEXT NOT NULL,
		file_path TEXT NOT NULL,
		uploaded_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (uploaded_by) REFERENCES users(id)
	)`

	accessTable := `
	CREATE TABLE IF NOT EXISTS user_pdf_access (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		pdf_id INTEGER NOT NULL,
		accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (pdf_id) REFERENCES pdfs(id)
	)`

	tables := []string{userTable, pdfTable, accessTable}

	for _, table := range tables {
		if _, err := d.db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) CreateUser(username, email, passwordHash string) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
	_, err := d.db.Exec(query, username, email, passwordHash)
	return err
}

func (d *Database) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE username = ?`
	row := d.db.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *Database) GetUserByID(id int) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id = ?`
	row := d.db.QueryRow(query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *Database) CreatePDF(title, author, description, filename, filePath string, uploadedBy int) error {
	query := `INSERT INTO pdfs (title, author, description, filename, file_path, uploaded_by) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, title, author, description, filename, filePath, uploadedBy)
	return err
}

func (d *Database) GetAllPDFs() ([]models.PDF, error) {
	query := `SELECT id, title, author, description, filename, file_path, uploaded_by, created_at FROM pdfs ORDER BY created_at DESC`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pdfs []models.PDF
	for rows.Next() {
		var pdf models.PDF
		err := rows.Scan(&pdf.ID, &pdf.Title, &pdf.Author, &pdf.Description, &pdf.Filename, &pdf.FilePath, &pdf.UploadedBy, &pdf.CreatedAt)
		if err != nil {
			return nil, err
		}
		pdfs = append(pdfs, pdf)
	}

	return pdfs, nil
}

func (d *Database) GetPDFByID(id int) (*models.PDF, error) {
	query := `SELECT id, title, author, description, filename, file_path, uploaded_by, created_at FROM pdfs WHERE id = ?`
	row := d.db.QueryRow(query, id)

	var pdf models.PDF
	err := row.Scan(&pdf.ID, &pdf.Title, &pdf.Author, &pdf.Description, &pdf.Filename, &pdf.FilePath, &pdf.UploadedBy, &pdf.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &pdf, nil
}

func (d *Database) SearchPDFs(query string) ([]models.PDF, error) {
	searchQuery := `SELECT id, title, author, description, filename, file_path, uploaded_by, created_at FROM pdfs 
	WHERE title LIKE ? OR author LIKE ? OR description LIKE ? ORDER BY created_at DESC`

	searchPattern := "%" + query + "%"
	rows, err := d.db.Query(searchQuery, searchPattern, searchPattern, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pdfs []models.PDF
	for rows.Next() {
		var pdf models.PDF
		err := rows.Scan(&pdf.ID, &pdf.Title, &pdf.Author, &pdf.Description, &pdf.Filename, &pdf.FilePath, &pdf.UploadedBy, &pdf.CreatedAt)
		if err != nil {
			return nil, err
		}
		pdfs = append(pdfs, pdf)
	}

	return pdfs, nil
}

func (d *Database) RecordPDFAccess(userID, pdfID int) error {
	query := `INSERT INTO user_pdf_access (user_id, pdf_id) VALUES (?, ?)`
	_, err := d.db.Exec(query, userID, pdfID)
	return err
}

func (d *Database) GetUserAccessHistory(userID int) ([]models.UserPDFAccess, error) {
	query := `SELECT id, user_id, pdf_id, accessed_at FROM user_pdf_access WHERE user_id = ? ORDER BY accessed_at DESC`
	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var access []models.UserPDFAccess
	for rows.Next() {
		var a models.UserPDFAccess
		err := rows.Scan(&a.ID, &a.UserID, &a.PDFID, &a.AccessedAt)
		if err != nil {
			return nil, err
		}
		access = append(access, a)
	}

	return access, nil
}
