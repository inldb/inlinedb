package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

type APIServer struct {
	addr string
}

func NewAPIServer(addr string) *APIServer {
	return &APIServer{
		addr: addr,
	}
}

type wrappedWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}

	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
	w.headerWritten = true
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	router.HandleFunc("/favicon.ico", serveFavicon).Methods("GET")

	// uncomment this when you start working on the client side routes like insert and other stuff
	subrouter := router.PathPrefix("/v1").Subrouter()
	subrouter.HandleFunc("/insert", InsertIntoTable).Methods("POST")
	subrouter.HandleFunc("/select", SelectFromTable).Methods("GET")

	adminRoute := router.PathPrefix("/admin").Subrouter()

	// Health Check
	adminRoute.HandleFunc("/health", HealthCheck).Methods("GET")

	adminRoute.HandleFunc("/tables", GetAllTables).Methods("GET")
	adminRoute.HandleFunc("/table", GetTable).Methods("GET")
	adminRoute.HandleFunc("/table", CreateTable).Methods("POST")
	adminRoute.HandleFunc("/query", RowsAsJson).Methods("POST")
	adminRoute.HandleFunc("/overview", GetOverview).Methods("GET")

	// Dashboard routes
	dashboardRoute := router.PathPrefix("/dashboard").Subrouter()
	dashboardRoute.HandleFunc("/", serveDashboardFile("index.html")).Methods("GET")
	dashboardRoute.HandleFunc("", serveDashboardFile("index.html")).Methods("GET")
	// Logo
	dashboardRoute.HandleFunc("/logo.png", serveDashboardFile("logo.png")).Methods("GET")

	// HTMX navigation endpoints
	dashboardRoute.HandleFunc("/overview", serveDashboardFile("overview.html")).Methods("GET")
	dashboardRoute.HandleFunc("/table-editor", serveDashboardFile("table-editor.html")).Methods("GET")
	dashboardRoute.HandleFunc("/create-table", serveDashboardFile("create-table.html")).Methods("GET")
	dashboardRoute.HandleFunc("/insert-form", serveDashboardFile("insert-form.html")).Methods("GET")
	dashboardRoute.HandleFunc("/sql-editor", serveDashboardFile("sql-editor.html")).Methods("GET")

	// Placeholder endpoints for other navigation items
	dashboardRoute.HandleFunc("/auth", servePlaceholderPage("Authentication")).Methods("GET")
	dashboardRoute.HandleFunc("/storage", servePlaceholderPage("Storage")).Methods("GET")
	dashboardRoute.HandleFunc("/edge-functions", servePlaceholderPage("Edge Functions")).Methods("GET")
	dashboardRoute.HandleFunc("/database", servePlaceholderPage("Database")).Methods("GET")
	dashboardRoute.HandleFunc("/api", servePlaceholderPage("API")).Methods("GET")
	dashboardRoute.HandleFunc("/logs", servePlaceholderPage("Logs")).Methods("GET")
	dashboardRoute.HandleFunc("/settings", servePlaceholderPage("Settings")).Methods("GET")

	middlewareChain := MiddlwareChain(
		RequestLoggerMiddleware,
	)

	server := http.Server{
		Addr:    s.addr,
		Handler: middlewareChain(router),
	}
	log.Printf("Server has started %s", s.addr)
	return server.ListenAndServe()
}

func serveDashboardFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Adjust the path to where your HTML files are located
		filePath := filepath.Join("dashboard", filename)

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading file %s: %v", filePath, err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	}
}

func serveFavicon(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("static", "favicon.ico") // or "favicon.ico" depending on your file name

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading favicon file: %v", err)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/x-icon")
	w.Write(content)
}

// servePlaceholderPage returns a handler for placeholder pages
func servePlaceholderPage(pageName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html := fmt.Sprintf(`
			<div class="p-8">
				<h1 class="text-3xl font-bold mb-4">%s</h1>
				<p class="text-gray-400">Coming soon...</p>
			</div>
		`, pageName)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}
}

func RequestLoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		start := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Try X-Forwarded-For (contains a comma-separated list of IPs)
		var ip string
		xForwardedFor := r.Header.Get("X-Forwarded-For")
		if xForwardedFor != "" {
			// Get first IP in the list (client's original IP)
			ips := strings.Split(xForwardedFor, ",")
			if len(ips) > 0 {
				ip = strings.TrimSpace(ips[0])

			}
		}

		next.ServeHTTP(wrapped, r)

		var color string

		if wrapped.statusCode >= 200 && wrapped.statusCode <= 300 {
			color = Green
		} else {
			color = Red
		}

		log.Printf("%s %s %d %s %s %s %s %s %v", color, "[", wrapped.statusCode, r.Method, "]", Reset, ip, r.URL.Path, time.Since(start))
	}
}

type Middleware func(http.Handler) http.HandlerFunc

func MiddlwareChain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next.ServeHTTP
	}
}
