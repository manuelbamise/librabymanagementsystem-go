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

	rolesTable := `
		CREATE TABLE IF NOT EXISTS roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

	permissionsTable := `
		CREATE TABLE IF NOT EXISTS permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			resource TEXT NOT NULL,
			action TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

	userRolesTable := `
		CREATE TABLE IF NOT EXISTS user_roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			role_id INTEGER NOT NULL,
			assigned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			assigned_by INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (role_id) REFERENCES roles(id),
			FOREIGN KEY (assigned_by) REFERENCES users(id),
			UNIQUE(user_id, role_id)
		)`

	rolePermissionsTable := `
		CREATE TABLE IF NOT EXISTS role_permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (role_id) REFERENCES roles(id),
			FOREIGN KEY (permission_id) REFERENCES permissions(id),
			UNIQUE(role_id, permission_id)
		)`

	tables := []string{userTable, pdfTable, accessTable, rolesTable, permissionsTable, userRolesTable, rolePermissionsTable}

	for _, table := range tables {
		if _, err := d.db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Initialize default roles and permissions
	if err := d.initializeRBAC(); err != nil {
		return fmt.Errorf("failed to initialize RBAC: %w", err)
	}

	// Assign default roles to existing users
	if err := d.assignDefaultRoles(); err != nil {
		return fmt.Errorf("failed to assign default roles: %w", err)
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) CreateUser(username, email, passwordHash string) error {
	// Start transaction
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert user
	query := `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
	result, err := tx.Exec(query, username, email, passwordHash)
	if err != nil {
		return err
	}

	// Get user ID
	userID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Assign default user role
	userRoleID, err := d.getRoleIDByName("user")
	if err != nil {
		return fmt.Errorf("failed to get user role ID: %w", err)
	}

	_, err = tx.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, userID, userRoleID)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
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

func (d *Database) initializeRBAC() error {
	// Create default roles
	roles := []struct {
		name        string
		description string
	}{
		{"admin", "System administrator with full access"},
		{"user", "Regular user with catalog access"},
	}

	for _, role := range roles {
		query := `INSERT OR IGNORE INTO roles (name, description) VALUES (?, ?)`
		if _, err := d.db.Exec(query, role.name, role.description); err != nil {
			return fmt.Errorf("failed to create role %s: %w", role.name, err)
		}
	}

	// Create default permissions
	permissions := []struct {
		name        string
		resource    string
		action      string
		description string
	}{
		{"upload_pdf", "pdf", "create", "Upload PDF files"},
		{"view_pdf", "pdf", "read", "View PDF files"},
		{"edit_pdf", "pdf", "update", "Edit PDF metadata"},
		{"delete_pdf", "pdf", "delete", "Delete PDF files"},
		{"manage_users", "user", "manage", "Manage user accounts"},
		{"manage_roles", "role", "manage", "Manage user roles"},
	}

	for _, perm := range permissions {
		query := `INSERT OR IGNORE INTO permissions (name, resource, action, description) VALUES (?, ?, ?, ?)`
		if _, err := d.db.Exec(query, perm.name, perm.resource, perm.action, perm.description); err != nil {
			return fmt.Errorf("failed to create permission %s: %w", perm.name, err)
		}
	}

	// Assign permissions to roles
	rolePermissions := map[string][]string{
		"admin": {"upload_pdf", "view_pdf", "edit_pdf", "delete_pdf", "manage_users", "manage_roles"},
		"user":  {"view_pdf"},
	}

	for roleName, permNames := range rolePermissions {
		roleID, err := d.getRoleIDByName(roleName)
		if err != nil {
			return fmt.Errorf("failed to get role ID for %s: %w", roleName, err)
		}

		for _, permName := range permNames {
			permID, err := d.getPermissionIDByName(permName)
			if err != nil {
				return fmt.Errorf("failed to get permission ID for %s: %w", permName, err)
			}

			query := `INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)`
			if _, err := d.db.Exec(query, roleID, permID); err != nil {
				return fmt.Errorf("failed to assign permission %s to role %s: %w", permName, roleName, err)
			}
		}
	}

	return nil
}

func (d *Database) getRoleIDByName(name string) (int, error) {
	query := `SELECT id FROM roles WHERE name = ?`
	var id int
	err := d.db.QueryRow(query, name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Database) getPermissionIDByName(name string) (int, error) {
	query := `SELECT id FROM permissions WHERE name = ?`
	var id int
	err := d.db.QueryRow(query, name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// RBAC query methods
func (d *Database) GetUserRoles(userID int) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at 
		FROM roles r 
		INNER JOIN user_roles ur ON r.id = ur.role_id 
		WHERE ur.user_id = ?`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (d *Database) GetUserWithRoles(userID int) (*models.User, error) {
	user, err := d.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	roles, err := d.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	user.Roles = roles
	return user, nil
}

func (d *Database) HasPermission(userID int, permissionName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM user_roles ur 
		INNER JOIN role_permissions rp ON ur.role_id = rp.role_id 
		INNER JOIN permissions p ON rp.permission_id = p.id 
		WHERE ur.user_id = ? AND p.name = ?`

	var count int
	err := d.db.QueryRow(query, userID, permissionName).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (d *Database) AssignRole(userID, roleID int, assignedBy *int) error {
	query := `INSERT OR REPLACE INTO user_roles (user_id, role_id, assigned_by) VALUES (?, ?, ?)`
	_, err := d.db.Exec(query, userID, roleID, assignedBy)
	return err
}

func (d *Database) RemoveRole(userID, roleID int) error {
	query := `DELETE FROM user_roles WHERE user_id = ? AND role_id = ?`
	_, err := d.db.Exec(query, userID, roleID)
	return err
}

func (d *Database) GetAllRoles() ([]models.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles ORDER BY name`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (d *Database) GetRolePermissions(roleID int) ([]models.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at 
		FROM permissions p 
		INNER JOIN role_permissions rp ON p.id = rp.permission_id 
		WHERE rp.role_id = ?`

	rows, err := d.db.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (d *Database) GetAllUsersWithRoles() ([]models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users ORDER BY username`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Get user roles
		roles, err := d.GetUserRoles(user.ID)
		if err != nil {
			return nil, err
		}
		user.Roles = roles

		users = append(users, user)
	}

	return users, nil
}

func (d *Database) DeletePDF(id int) error {
	query := `DELETE FROM pdfs WHERE id = ?`
	_, err := d.db.Exec(query, id)
	return err
}

func (d *Database) assignDefaultRoles() error {
	// Get all users
	query := `SELECT id FROM users`
	rows, err := d.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return err
		}
		userIDs = append(userIDs, userID)
	}

	// Get user role ID
	userRoleID, err := d.getRoleIDByName("user")
	if err != nil {
		return fmt.Errorf("failed to get user role ID: %w", err)
	}

	// Get admin role ID
	adminRoleID, err := d.getRoleIDByName("admin")
	if err != nil {
		return fmt.Errorf("failed to get admin role ID: %w", err)
	}

	// Assign user role to all users who don't have any roles
	for _, userID := range userIDs {
		// Check if user already has roles
		userRoles, err := d.GetUserRoles(userID)
		if err != nil {
			return fmt.Errorf("failed to check user roles for user %d: %w", userID, err)
		}

		if len(userRoles) == 0 {
			// Assign default user role
			if err := d.AssignRole(userID, userRoleID, nil); err != nil {
				return fmt.Errorf("failed to assign user role to user %d: %w", userID, err)
			}
		}
	}

	// Make first user an admin if there are no admins
	query = `SELECT COUNT(*) FROM user_roles WHERE role_id = ?`
	var adminCount int
	err = d.db.QueryRow(query, adminRoleID).Scan(&adminCount)
	if err != nil {
		return fmt.Errorf("failed to check admin count: %w", err)
	}

	if adminCount == 0 && len(userIDs) > 0 {
		// Make first user an admin
		if err := d.AssignRole(userIDs[0], adminRoleID, nil); err != nil {
			return fmt.Errorf("failed to assign admin role to first user: %w", err)
		}
	}

	return nil
}
