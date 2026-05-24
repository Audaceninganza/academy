package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	postgrest "github.com/supabase-community/postgrest-go"
)

// ── Auth middleware ──────────────────────────────────────────────────────────

func adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("ADMIN_SECRET")
		if secret == "" {
			http.Error(w, "Admin access not configured", http.StatusInternalServerError)
			return
		}

		provided := r.Header.Get("X-Admin-Secret")
		if provided == "" {
			provided = r.URL.Query().Get("secret")
		}

		if provided != secret {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Secret")

		if r.Method == http.MethodOptions {
			return
		}

		next(w, r)
	}
}

// ── Student type ─────────────────────────────────────────────────────────────

type Student struct {
	ID               interface{} `json:"id"`
	FullName         string      `json:"full_name"`
	Age              int         `json:"age"`
	Phone            string      `json:"phone"`
	CourseSlug       string      `json:"course_slug"`
	PaymentStatus    string      `json:"payment_status"`
	RegistrationFees bool        `json:"registration_fees"`
	FullPackage      bool        `json:"full_package"`
	CreatedAt        interface{} `json:"created_at,omitempty"`
}

// ── GET /admin/students ───────────────────────────────────────────────────────

func handleListStudents(w http.ResponseWriter, r *http.Request) {
	data, _, err := client.From("enrollments").
		Select("*", "exact", false).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Execute()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Printf("Supabase list error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch students"})
		return
	}

	w.Write(data)
}

// ── PUT /admin/students/{id} ──────────────────────────────────────────────────

func handleUpdateStudent(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/students/")
	if id == "" {
		http.Error(w, "Missing student ID", http.StatusBadRequest)
		return
	}

	var req EnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	update := map[string]interface{}{
		"full_name":         req.Name,
		"age":               req.Age,
		"phone":             req.Phone,
		"course_slug":       req.Course,
		"payment_status":    req.PaymentStatus,
		"registration_fees": req.RegistrationFees,
		"full_package":      req.FullPackage,
	}

	_, _, err := client.From("enrollments").
		Update(update, "", "").
		Eq("id", id).
		Execute()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Printf("Supabase update error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Update failed"})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// ── DELETE /admin/students/{id} ───────────────────────────────────────────────

func handleDeleteStudent(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/students/")
	if id == "" {
		http.Error(w, "Missing student ID", http.StatusBadRequest)
		return
	}

	_, _, err := client.From("enrollments").
		Delete("", "").
		Eq("id", id).
		Execute()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Printf("Supabase delete error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Delete failed"})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
