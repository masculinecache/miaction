package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"mibilltracker/internal/api"
	"mibilltracker/internal/bills"
	"mibilltracker/internal/email"
	"mibilltracker/internal/representatives"
)

func templateDict(v ...interface{}) map[string]interface{} {
	d := make(map[string]interface{})
	for i := 0; i < len(v)-1; i += 2 {
		key, ok := v[i].(string)
		if ok {
			d[key] = v[i+1]
		}
	}
	return d
}

func isUrgent(t time.Time) bool {
	return !t.IsZero() && t.Before(time.Now().Add(7*24*time.Hour))
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006")
}

func joinStrings(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

func isSelectedRep(selected []string, id string) bool {
	for _, s := range selected {
		if s == id {
			return true
		}
	}
	return false
}

func getVar(vars map[string]string, key string) string {
	return vars[key]
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	billService := bills.NewService()
	billService.Start()

	repService := representatives.NewService()
	emailService := email.NewService()

	funcMap := template.FuncMap{
		"dict":          templateDict,
		"isUrgent":      isUrgent,
		"formatDate":    formatDate,
		"join":          joinStrings,
		"isSelectedRep": isSelectedRep,
		"getVar":        getVar,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	handler := api.NewHandler(billService, repService, emailService)
	handler.SetTemplates(tmpl)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("MiBillTracker server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
