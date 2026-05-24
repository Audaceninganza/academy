package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/supabase-community/supabase-go"
)

type EnrollmentRequest struct {
	Name             string `json:"name"`
	Age              int    `json:"age"`
	Phone            string `json:"phone"`
	Course           string `json:"course"`
	PaymentStatus    string `json:"payment_status"`
	RegistrationFees bool   `json:"registration_fees"`
	FullPackage      bool   `json:"full_package"`
}

var client *supabase.Client

func main() {
	var err error
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")

	if url == "" || key == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_KEY environment variables are required")
	}

	client, err = supabase.NewClient(url, key, nil)
	if err != nil {
		log.Fatalf("cannot initialize supabase client: %v", err)
	}

	// ── Public routes ────────────────────────────────────────────────────────
	fs := http.FileServer(http.Dir("../"))
	http.Handle("/", fs)
	http.HandleFunc("/enroll", handleEnrollment)

	// ── Admin routes (protected by ADMIN_SECRET) ─────────────────────────────
	http.HandleFunc("/admin/students", adminOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleListStudents(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	http.HandleFunc("/admin/students/", adminOnly(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handleUpdateStudent(w, r)
		case http.MethodDelete:
			handleDeleteStudent(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// ── Serve the admin dashboard HTML ───────────────────────────────────────
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../dashboard.html")
	})

	port := "3000"
	fmt.Printf("Server starting at http://localhost:%s\n", port)
	fmt.Printf("Admin dashboard: http://localhost:%s/admin?secret=YOUR_SECRET\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func handleEnrollment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Default payment_status to "pending" if not provided
	if req.PaymentStatus == "" {
		req.PaymentStatus = "pending"
	}

	newEntry := map[string]interface{}{
		"full_name":         req.Name,
		"age":               req.Age,
		"phone":             req.Phone,
		"course_slug":       req.Course,
		"payment_status":    req.PaymentStatus,
		"registration_fees": req.RegistrationFees,
		"full_package":      req.FullPackage,
	}

	_, _, err := client.From("enrollments").Insert(newEntry, false, "", "", "").Execute()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Printf("Supabase Error: %v\n", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Database error",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
