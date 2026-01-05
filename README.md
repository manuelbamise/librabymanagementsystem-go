# Library Management System

A modern web application for uploading, organizing, and reading PDF documents built with Go, Templ, and HTMX.

## 🚀 Features

- **User Authentication**: Secure login/registration with bcrypt password hashing
- **PDF Management**: Upload and organize PDF documents with metadata
- **Library Catalog**: Browse and search through your PDF collection
- **Online Viewer**: Read PDFs directly in your browser
- **Real-time Search**: Instant search across titles, authors, and descriptions
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Session Management**: Secure session-based authentication
- **Access Tracking**: Records user interactions with PDFs

## 🛠️ Technology Stack

### Backend
- **Go**: Programming language and standard library
- **SQLite**: Lightweight database for data storage
- **bcrypt**: Password hashing for security
- **net/http**: HTTP server and client implementation

### Frontend
- **Templ**: Go-based HTML templating language
- **HTMX**: Modern browser interactions without JavaScript complexity
- **CSS**: Responsive design with mobile-first approach

### Architecture
- **MVC Pattern**: Clear separation of concerns
- **Component-based Templates**: Reusable UI components
- **Middleware**: Request authentication and validation
- **Session Storage**: In-memory session management

## 📁 Project Structure

```
librarymanagementsystem/
├── main.go                 # Entry point and HTTP server
├── go.mod                  # Go module dependencies
├── library.db              # SQLite database (auto-created)
├── internal/
│   ├── auth/              # Authentication handlers
│   │   ├── session.go     # Session management
│   │   └── password.go    # Password utilities
│   ├── db/                # Database layer
│   │   └── database.go    # Database operations
│   ├── handlers/          # HTTP route handlers
│   │   ├── auth.go        # Authentication endpoints
│   │   └── library.go     # Library functionality
│   └── models/            # Data structures
│       └── models.go       # User, PDF, and Access models
├── templates/
│   └── base.templ         # All HTML templates (Templ)
├── static/
│   ├── css/
│   │   └── style.css     # Complete styling
│   ├── js/               # JavaScript files (if needed)
│   └── uploads/          # PDF file storage
└── README.md             # This file
```

## 🗄️ Database Schema

The application uses SQLite with three main tables:

### Users
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### PDFs
```sql
CREATE TABLE pdfs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT,
    description TEXT,
    filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    uploaded_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (uploaded_by) REFERENCES users(id)
);
```

### User PDF Access
```sql
CREATE TABLE user_pdf_access (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    pdf_id INTEGER NOT NULL,
    accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (pdf_id) REFERENCES pdfs(id)
);
```

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or later
- Basic understanding of web development

### Installation

#### Quick Start (Recommended)
```bash
# Clone or create the project directory
mkdir librarymanagementsystem
cd librarymanagementsystem

# Run the quickstart script
./quickstart.sh

# Start the server
go run .
```

#### Manual Installation
1. **Clone or create the project directory**
   ```bash
   mkdir librarymanagementsystem
   cd librarymanagementsystem
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Install templ CLI**
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

4. **Generate templates**
   ```bash
   templ generate
   ```

5. **Run application**
   ```bash
   go run .
   ```

6. **Access the application**
   Open your browser and navigate to `http://localhost:8080`

## 🌐 API Endpoints

### Authentication
- `GET /` - Landing page
- `GET /login` - Login form
- `POST /auth/login` - Process login
- `GET /register` - Registration form
- `POST /auth/register` - Process registration
- `POST /auth/logout` - Logout user

### Library
- `GET /library` - Main library catalog (protected)
- `GET /library/search` - Search PDFs (protected)
- `GET /library/view/{id}` - View PDF (protected)
- `GET /upload` - Upload form (protected)
- `POST /library/upload` - Process PDF upload (protected)

### Static Files
- `/static/*` - Static file serving (CSS, JS, uploaded PDFs)

## 🔒 Security Features

- **Password Security**: bcrypt hashing with default cost factor
- **Session Management**: Secure HTTP-only cookies
- **CSRF Protection**: SameSite cookie attributes
- **SQL Injection Prevention**: Prepared statements for all database queries
- **File Upload Validation**: PDF file type restrictions
- **Input Sanitization**: Form validation and sanitization
- **Authentication Middleware**: Protected route enforcement

## 📱 User Interface

### Landing Page
- Feature overview and call-to-action
- Login and registration access points

### Library Catalog
- Grid layout of PDF documents
- Real-time search functionality
- Responsive design for all screen sizes

### PDF Viewer
- In-browser PDF viewing with iframe
- Navigation back to catalog
- Metadata display

### Upload Form
- Multi-field form with file upload
- Client and server-side validation
- Automatic catalog refresh after upload

## 🎨 Styling Features

- **Responsive Design**: Mobile-first approach
- **Modern CSS**: Flexbox and Grid layouts
- **Interactive Elements**: Hover effects and transitions
- **Component-based CSS**: Reusable styling patterns
- **Accessibility**: Semantic HTML and proper labeling

## 🔄 HTMX Integration

### Dynamic Interactions
- Search without page reloads
- Form submissions with targeted updates
- Smooth navigation between views
- Real-time content updates

### HTMX Features Used
- `hx-get` for dynamic content loading
- `hx-post` for form submissions
- `hx-target` for precise DOM updates
- `hx-push-url` for proper browser history
- `hx-trigger` for event-driven interactions

## 🧪 Development

### Code Generation
```bash
# Generate templates after changes
templ generate

# Build the application
go build .

# Run with live reload (if using air)
air
```

### Database Operations
The SQLite database (`library.db`) is automatically created on first run with proper table structure.

### File Uploads
PDF files are stored in `static/uploads/` with unique filenames to prevent conflicts.

## 🔧 Configuration

### Environment Variables
Currently using in-memory sessions. For production:
- Consider Redis for session storage
- Configure database connection strings
- Set proper cookie security settings

### Production Considerations
- Enable HTTPS
- Configure reverse proxy (nginx/Apache)
- Set up proper logging
- Implement rate limiting
- Add database backups

## 🤝 Contributing

This is a demonstration project showcasing Go web development with modern tools. Feel free to extend or modify:

1. Add user roles and permissions
2. Implement PDF annotations
3. Add collection/folder organization
4. Include PDF sharing features
5. Add advanced search filters

## 📝 License

This project is provided as an educational example. Feel free to use and modify for your own purposes.

## 📖 Usage Guide

### First Time Setup
1. Register a new account using the registration form
2. Login with your credentials
3. Upload your first PDF using the "Upload PDF" button
4. Browse your library catalog
5. Click on any PDF to view it in the browser

### Key Workflows

**Uploading PDFs**
- Navigate to `/upload` or click "Upload PDF" in the catalog
- Fill in the title (required), author, and description
- Select a PDF file from your device
- Click "Upload PDF" to add it to your library

**Searching PDFs**
- Use the search bar in the catalog to filter PDFs
- Search works across title, author, and description
- Results update in real-time as you type

**Viewing PDFs**
- Click on any PDF card in the catalog
- PDF opens in an embedded viewer
- Use the "Back to Catalog" button to return

**Managing Account**
- Use the logout button to end your session
- Login again to access your library
- Sessions expire after 24 hours

## 🙏 Acknowledgments

- **Templ**: Modern Go templating library
- **HTMX**: Simplified web interactions
- **Go Standard Library**: Comprehensive web development tools
- **SQLite**: Reliable embedded database