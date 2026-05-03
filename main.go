package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/supabase-community/supabase-go"
)

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

	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)

	http.HandleFunc("/enroll", handleEnrollment)

	port := "8080"
	fmt.Printf("Server starting at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func handleEnrollment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodPost {
		return
	}

	age, _ := strconv.Atoi(r.FormValue("age"))
	newEntry := map[string]interface{}{
		"full_name":   r.FormValue("name"),
		"age":         age,
		"phone":       r.FormValue("phone"),
		"course_slug": r.FormValue("course"),
	}

	_, _, err := client.From("enrollments").Insert(newEntry, false, "", "", "").Execute()

	w.Header().Set("Content-Type", "text/html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Supabase Error: %v\n", err)
		w.Write([]byte("❌ Error: " + err.Error()))
		return
	}

	w.Write([]byte("✅ Spot reserved! We will call you within 24h."))
}
