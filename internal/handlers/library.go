package handlers

import (
	"context"
	"fmt"
	"io"
	"librarymanagementsystem/internal/auth"
	"librarymanagementsystem/internal/db"
	"librarymanagementsystem/internal/models"
	"librarymanagementsystem/templates"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type LibraryHandler struct {
	db             *db.Database
	sessionManager *auth.SessionManager
}

func NewLibraryHandler(database *db.Database, sessionManager *auth.SessionManager) *LibraryHandler {
	return &LibraryHandler{
		db:             database,
		sessionManager: sessionManager,
	}
}

func (h *LibraryHandler) Index(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r.Context())

	pdfs, err := h.db.GetAllPDFs()
	if err != nil {
		http.Error(w, "Failed to fetch PDFs", http.StatusInternalServerError)
		return
	}

	templates.LibraryIndex(pdfs, user).Render(r.Context(), w)
}

func (h *LibraryHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		query = r.FormValue("search")
	}

	var pdfs []models.PDF
	var err error

	if query == "" {
		pdfs, err = h.db.GetAllPDFs()
	} else {
		pdfs, err = h.db.SearchPDFs(query)
	}

	if err != nil {
		http.Error(w, "Failed to search PDFs", http.StatusInternalServerError)
		return
	}

	templates.PDFList(pdfs).Render(r.Context(), w)
}

func (h *LibraryHandler) ViewPDF(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r.Context())

	idStr := strings.TrimPrefix(r.URL.Path, "/library/view/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid PDF ID", http.StatusBadRequest)
		return
	}

	pdf, err := h.db.GetPDFByID(id)
	if err != nil {
		http.Error(w, "PDF not found", http.StatusNotFound)
		return
	}

	// Record access
	err = h.db.RecordPDFAccess(user.ID, pdf.ID)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to record PDF access: %v\n", err)
	}

	templates.PDFViewer(*pdf, user).Render(r.Context(), w)
}

func (h *LibraryHandler) UploadForm(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r.Context())
	templates.UploadPDF(user).Render(r.Context(), w)
}

func (h *LibraryHandler) UploadPDF(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r.Context())

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 32MB)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check if file is PDF
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", user.ID, header.Filename)
	filePath := fmt.Sprintf("static/uploads/%s", filename)

	// Save file
	err = h.saveUploadedFile(file, filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Create PDF record in database
	err = h.db.CreatePDF(title, author, description, filename, filePath, user.ID)
	if err != nil {
		http.Error(w, "Failed to create PDF record", http.StatusInternalServerError)
		return
	}

	// Redirect to library
	http.Redirect(w, r, "/library", http.StatusSeeOther)
}

func (h *LibraryHandler) saveUploadedFile(file multipart.File, filePath string) error {
	// Create uploads directory if it doesn't exist
	uploadsDir := "static/uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create uploads directory: %w", err)
	}

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	_, err = io.Copy(dst, file)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

func (h *LibraryHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionToken, err := auth.GetSessionToken(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := h.sessionManager.ValidateSession(sessionToken)
		if err != nil {
			auth.ClearSessionCookie(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to context
		user, err := h.db.GetUserByID(session.UserID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *LibraryHandler) getUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value("user").(*models.User); ok {
		return user
	}
	return nil
}
