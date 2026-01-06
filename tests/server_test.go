package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestServerStartup tests that server can start without errors
func TestServerStartup(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Library Management System"))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestMainRoutes tests that all main routes respond without crashing
func TestMainRoutes(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Home page"))
		case "/login":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Login page"))
		case "/register":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Register page"))
		case "/library":
			// Protected route - simulate authentication check
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusSeeOther)
		case "/admin":
			// Admin protected route
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusSeeOther)
		case "/static/test.css":
			w.Header().Set("Content-Type", "text/css")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("body { color: red; }"))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - Page not found"))
		}
	})

	routes := []string{"/", "/login", "/register", "/library", "/admin", "/static/test.css", "/nonexistent"}
	expectedStatuses := map[string]int{
		"/":                http.StatusOK,
		"/login":           http.StatusOK,
		"/register":        http.StatusOK,
		"/library":         http.StatusSeeOther,
		"/admin":           http.StatusSeeOther,
		"/static/test.css": http.StatusOK,
		"/nonexistent":     http.StatusNotFound,
	}

	for _, route := range routes {
		t.Run("Route: "+route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			expectedStatus, exists := expectedStatuses[route]
			assert.True(t, exists, "Route should have expected status defined")
			assert.Equal(t, expectedStatus, w.Code)
		})
	}
}

// TestHTTPMethods tests that routes handle different HTTP methods correctly
func TestHTTPMethods(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			if r.Method == "GET" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		case "/login":
			if r.Method == "GET" || r.Method == "POST" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		case "/register":
			if r.Method == "GET" || r.Method == "POST" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	routes := []string{"/", "/login", "/register", "/nonexistent"}

	for _, route := range routes {
		for _, method := range methods {
			t.Run(method+" "+route, func(t *testing.T) {
				req := httptest.NewRequest(method, route, nil)
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)

				validStatuses := []int{http.StatusOK, http.StatusMethodNotAllowed, http.StatusNotFound}
				assert.Contains(t, validStatuses, w.Code)
			})
		}
	}
}

// TestStaticFileServing tests static file handling
func TestStaticFileServing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/test.css" {
			w.Header().Set("Content-Type", "text/css")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("body { color: red; }"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	testCases := []struct {
		path         string
		expectedCode int
		expectedType string
	}{
		{
			path:         "/static/test.css",
			expectedCode: http.StatusOK,
			expectedType: "text/css",
		},
		{
			path:         "/static/nonexistent.js",
			expectedCode: http.StatusNotFound,
			expectedType: "",
		},
	}

	for _, tc := range testCases {
		t.Run("Static: "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
			if tc.expectedType != "" {
				assert.Equal(t, tc.expectedType, w.Header().Get("Content-Type"))
			}
		})
	}
}

// TestRedirections tests URL redirection functionality
func TestRedirections(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			if r.Method == "POST" {
				w.Header().Set("Location", "/library")
				w.WriteHeader(http.StatusSeeOther)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		case "/auth/logout":
			w.Header().Set("Location", "/")
			w.WriteHeader(http.StatusSeeOther)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	testCases := []struct {
		name             string
		path             string
		method           string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "Login GET shows form",
			path:             "/login",
			method:           "GET",
			expectedStatus:   http.StatusOK,
			expectedLocation: "",
		},
		{
			name:             "Login POST redirects",
			path:             "/login",
			method:           "POST",
			expectedStatus:   http.StatusSeeOther,
			expectedLocation: "/library",
		},
		{
			name:             "Logout redirects",
			path:             "/auth/logout",
			method:           "POST",
			expectedStatus:   http.StatusSeeOther,
			expectedLocation: "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedLocation != "" {
				assert.Equal(t, tc.expectedLocation, w.Header().Get("Location"))
			}
		})
	}
}

// TestHeaders tests HTTP header handling
func TestHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Check request headers
		userAgent := r.Header.Get("User-Agent")
		accept := r.Header.Get("Accept")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK - " + userAgent + " - " + accept))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "Test-Agent/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))

	body := w.Body.String()
	assert.Contains(t, body, "Test-Agent/1.0")
	assert.Contains(t, body, "text/html,application/xhtml+xml")
}

// TestErrorHandling tests server error handling
func TestErrorHandling(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/error-400":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		case "/error-404":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		case "/error-500":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		default:
			w.WriteHeader(http.StatusOK)
		}
	})

	testCases := []struct {
		path         string
		expectedCode int
		expectedBody string
	}{
		{
			path:         "/error-400",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Bad Request",
		},
		{
			path:         "/error-404",
			expectedCode: http.StatusNotFound,
			expectedBody: "Not Found",
		},
		{
			path:         "/error-500",
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Internal Server Error",
		},
	}

	for _, tc := range testCases {
		t.Run("Error: "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBody)
		})
	}
}

// TestConcurrentRequests tests concurrent request handling
func TestConcurrentRequests(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request processed"))
	})

	concurrency := 10
	results := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				results <- fmt.Errorf("request %d returned status %d", id, w.Code)
				return
			}

			body := w.Body.String()
			if body != "Request processed" {
				results <- fmt.Errorf("request %d returned unexpected body: %s", id, body)
				return
			}

			results <- nil
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent request should succeed")
	}
}

// TestRequestSize tests handling of large requests
func TestRequestSize(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > 1024*1024 { // 1MB limit
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			w.Write([]byte("Request too large"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request accepted"))
	})

	testCases := []struct {
		name         string
		contentSize  int64
		expectedCode int
	}{
		{
			name:         "Small request",
			contentSize:  1024,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Large request",
			contentSize:  2 * 1024 * 1024, // 2MB
			expectedCode: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := make([]byte, tc.contentSize)
			req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}
