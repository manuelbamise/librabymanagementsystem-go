# Library Management System

A modern web application for uploading, organizing, and reading PDF documents built with Go, Templ, and HTMX.

## 🚀 Features

- **User Authentication**: Secure login/registration with bcrypt password hashing
- **PDF Management**: Upload and organize PDF documents with metadata
- **Library Catalog**: Browse and search through your PDF collection
- **Session Management**: Secure session-based authentication

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
- **Middleware**: Request authentication and validation
- **Session Storage**: In-memory session management

## 🚀 Getting Started
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

### Database Operations
The SQLite database (`library.db`) is automatically created on first run with proper table structure.

### File Uploads
PDF files are stored in `static/uploads/` with unique filenames to prevent conflicts.

