package api

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"mibilltracker/internal/bills"
	"mibilltracker/internal/email"
	"mibilltracker/internal/representatives"
)

type Handler struct {
	billService  *bills.Service
	repService   *representatives.Service
	emailService *email.Service
	templates    *template.Template
}

func NewHandler(billService *bills.Service, repService *representatives.Service, emailService *email.Service) *Handler {
	return &Handler{
		billService:  billService,
		repService:   repService,
		emailService: emailService,
	}
}

func (h *Handler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// API routes (JSON)
	mux.HandleFunc("/api/bills", h.handleBills)
	mux.HandleFunc("/api/bills/scheduled", h.handleScheduledBills)
	mux.HandleFunc("/api/bills/", h.handleBillDetail)
	mux.HandleFunc("/api/meetings", h.handleMeetings)
	mux.HandleFunc("/api/representatives", h.handleFindRepresentatives)
	mux.HandleFunc("/api/representatives/search", h.handleSearchRepresentatives)
	mux.HandleFunc("/api/email/templates", h.handleEmailTemplates)
	mux.HandleFunc("/api/email/compose", h.handleEmailCompose)
	mux.HandleFunc("/api/email/preview", h.handleEmailPreview)

	// HTMX page routes
	mux.HandleFunc("/bills/scheduled", h.handleScheduledFragment)
	mux.HandleFunc("/bills/all", h.handleAllBillsFragment)
	mux.HandleFunc("/bills/meetings", h.handleMeetingsFragment)
	mux.HandleFunc("/bills/filter", h.handleBillsFilterFragment)
	mux.HandleFunc("/bill/", h.handleBillDetailPage)
	mux.HandleFunc("/email/", h.handleEmailPage)

	// Index page (catch-all for SPA-style routing)
	mux.HandleFunc("/", h.handleIndexPage)
}

// ---- Helper functions for template rendering ----

func (h *Handler) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	if h.templates == nil {
		http.Error(w, "Templates not loaded", http.StatusInternalServerError)
		return
	}
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Error rendering template %q: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderFragment(w http.ResponseWriter, name string, data interface{}) {
	if h.templates == nil {
		http.Error(w, "Templates not loaded", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Error rendering fragment %q: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ---- Template data helpers ----

func isScheduledVoteUrgent(sv *bills.ScheduledVote) bool {
	if sv == nil {
		return false
	}
	return time.Until(sv.Date) < 7*24*time.Hour
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006")
}

func formatDateLong(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Monday, January 2, 2006")
}

// ---- HTMX Page Handlers ----

func (h *Handler) handleIndexPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only serve the root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	allBills := h.billService.GetBills()
	scheduled := h.billService.GetScheduledBills()
	meetings := h.billService.GetCommitteeMeetings()

	data := map[string]interface{}{
		"AllBills":        allBills,
		"ScheduledBills":  scheduled,
		"Meetings":        meetings,
		"ScheduledCount":  len(scheduled),
		"AllCount":        len(allBills),
		"MeetingsCount":   len(meetings),
		"Query":           "",
		"Page": "bills",
	}

	h.renderTemplate(w, "base", data)
}

func (h *Handler) handleScheduledFragment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scheduled := h.billService.GetScheduledBills()
	data := map[string]interface{}{
		"ScheduledBills": scheduled,
	}
	h.renderFragment(w, "scheduled_tab", data)
}

func (h *Handler) handleAllBillsFragment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	allBills := h.billService.GetBills()
	data := map[string]interface{}{
		"AllBills": allBills,
		"Query":    "",
	}
	h.renderFragment(w, "all_bills_tab", data)
}

func (h *Handler) handleMeetingsFragment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	meetings := h.billService.GetCommitteeMeetings()
	data := map[string]interface{}{
		"Meetings": meetings,
	}
	h.renderFragment(w, "meetings_tab", data)
}

func (h *Handler) handleBillsFilterFragment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("q"))
	allBills := h.billService.GetBills()

	var filtered []bills.Bill
	if query == "" {
		filtered = allBills
	} else {
		for _, b := range allBills {
			if strings.Contains(strings.ToLower(b.Title), query) ||
				strings.Contains(strings.ToLower(b.Number), query) ||
				strings.Contains(strings.ToLower(b.Sponsor), query) {
				filtered = append(filtered, b)
			}
		}
	}

	data := map[string]interface{}{
		"Bills": filtered,
	}
	h.renderFragment(w, "filter_results", data)
}

func (h *Handler) handleBillDetailPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/bill/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	bill := h.billService.GetBill(id)
	data := map[string]interface{}{
		"Bill":            bill,
		"Page": "bill_detail",
	}
	if bill != nil && bill.ScheduledVote != nil {
		data["ScheduledVoteDate"] = formatDateLong(bill.ScheduledVote.Date)
	}

	h.renderTemplate(w, "base", data)
}

func (h *Handler) handleEmailPage(w http.ResponseWriter, r *http.Request) {
	// Parse the bill ID from path: /email/{billId} or /email/{billId}/action
	path := strings.TrimPrefix(r.URL.Path, "/email/")
	parts := strings.SplitN(path, "/", 2)
	billID := parts[0]

	if billID == "" {
		http.NotFound(w, r)
		return
	}

	// Check for sub-actions
	if len(parts) > 1 {
		action := parts[1]
		switch action {
		case "find-reps":
			if r.Method == http.MethodPost {
				h.handleEmailFindReps(w, r, billID)
				return
			}
		case "compose":
			if r.Method == http.MethodPost {
				h.handleEmailComposeAction(w, r, billID)
				return
			}
		case "preview":
			if r.Method == http.MethodPost {
				h.handleEmailPreviewAction(w, r, billID)
				return
			}
		}
		http.NotFound(w, r)
		return
	}

	// Handle GET for the initial email page (step 1: location)
	bill := h.billService.GetBill(billID)
	templates := h.emailService.GetTemplates()

	data := map[string]interface{}{
		"Bill":            bill,
		"BillID":          billID,
		"EmailStep":       "location",
		"Templates":       templates,
		"Page": "email",
	}
	h.renderTemplate(w, "base", data)
}

func (h *Handler) handleEmailFindReps(w http.ResponseWriter, r *http.Request, billID string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	zipCode := r.FormValue("zip")
	if zipCode == "" {
		http.Error(w, "ZIP code required", http.StatusBadRequest)
		return
	}

	bill := h.billService.GetBill(billID)
	if bill == nil {
		http.Error(w, "Bill not found", http.StatusNotFound)
		return
	}

	reps, err := h.repService.FindByLocation(representatives.Location{ZIP: zipCode})
	if err != nil {
		// Return to location step with error
		data := map[string]interface{}{
			"Bill":      bill,
			"BillID":    billID,
			"EmailStep": "location",
			"ZipCode":   zipCode,
			"Error":     "Failed to find representatives. Please check your ZIP code.",
		}
		h.renderFragment(w, "email_step_location", data)
		return
	}

	// Pre-select all reps
	var repIDs []string
	for _, r := range reps {
		repIDs = append(repIDs, r.ID)
	}

	tmpls := h.emailService.GetTemplates()
	currentTemplateID := ""
	if len(tmpls) > 0 {
		currentTemplateID = tmpls[0].ID
	}

	variables := initEmailVariables(tmpls, currentTemplateID, bill)

	data := map[string]interface{}{
		"Bill":             bill,
		"BillID":           billID,
		"EmailStep":        "compose",
		"ZipCode":          zipCode,
		"Reps":             reps,
		"SelectedReps":     repIDs,
		"Templates":        tmpls,
		"SelectedTemplate": currentTemplateID,
		"CurrentTemplate":  h.emailService.GetTemplate(currentTemplateID),
		"Variables":        variables,
		"TemplateVariables": func() []email.TemplateVariable {
			t := h.emailService.GetTemplate(currentTemplateID)
			if t != nil {
				return t.Variables
			}
			return nil
		}(),
	}

	h.renderFragment(w, "email_step_compose", data)
}

func (h *Handler) handleEmailComposeAction(w http.ResponseWriter, r *http.Request, billID string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	bill := h.billService.GetBill(billID)
	if bill == nil {
		http.Error(w, "Bill not found", http.StatusNotFound)
		return
	}

	templateID := r.FormValue("template_id")
	repIDs := r.Form["rep_ids"]

	if len(repIDs) == 0 {
		http.Error(w, "No representatives selected", http.StatusBadRequest)
		return
	}

	// Build variables map
	variables := make(map[string]string)
	for key, vals := range r.Form {
		if strings.HasPrefix(key, "var_") {
			varName := strings.TrimPrefix(key, "var_")
			if len(vals) > 0 {
				variables[varName] = vals[0]
			}
		}
	}

	// Look up representatives
	var repEmails []struct{ ID, Email, Name string }
	for _, repID := range repIDs {
		rep := h.repService.GetByID(repID)
		if rep != nil {
			repEmails = append(repEmails, struct{ ID, Email, Name string }{
				ID:    rep.ID,
				Email: rep.Email,
				Name:  rep.Name,
			})
		}
	}

	if len(repEmails) == 0 {
		http.Error(w, "No valid representatives found", http.StatusBadRequest)
		return
	}

	mailtoLinks := h.emailService.GenerateMailtoLinks(
		email.EmailRequest{
			TemplateID:      templateID,
			Variables:       variables,
			Representatives: repIDs,
		},
		repEmails,
	)

	data := map[string]interface{}{
		"Bill":        bill,
		"BillID":      billID,
		"MailtoLinks": mailtoLinks,
		"EmailStep":   "send",
	}
	h.renderFragment(w, "email_step_send", data)
}

func (h *Handler) handleEmailPreviewAction(w http.ResponseWriter, r *http.Request, billID string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	templateID := r.FormValue("template_id")

	// Build variables map
	variables := make(map[string]string)
	for key, vals := range r.Form {
		if strings.HasPrefix(key, "var_") {
			varName := strings.TrimPrefix(key, "var_")
			if len(vals) > 0 {
				variables[varName] = vals[0]
			}
		}
	}

	subject, body, err := h.emailService.ComposeEmail(templateID, variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"Subject": subject,
		"Body":    body,
	}
	h.renderFragment(w, "email_preview", data)
}

func initEmailVariables(tmpls []email.EmailTemplate, templateID string, bill *bills.Bill) map[string]string {
	variables := make(map[string]string)
	for _, t := range tmpls {
		if t.ID == templateID {
			for _, v := range t.Variables {
				switch v.Name {
				case "BillNumber":
					variables[v.Name] = bill.Number
				case "BillTitle":
					variables[v.Name] = bill.Title
				default:
					variables[v.Name] = v.Default
				}
			}
			break
		}
	}
	return variables
}

// ---- Existing JSON Handlers (unchanged) ----

func (h *Handler) handleBills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bills := h.billService.GetBills()
	respondJSON(w, bills)
}

func (h *Handler) handleScheduledBills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bills := h.billService.GetScheduledBills()
	respondJSON(w, bills)
}

func (h *Handler) handleBillDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/api/bills/"):]
	if id == "" {
		http.Error(w, "Bill ID required", http.StatusBadRequest)
		return
	}

	bill := h.billService.GetBill(id)
	if bill == nil {
		http.Error(w, "Bill not found", http.StatusNotFound)
		return
	}

	respondJSON(w, bill)
}

func (h *Handler) handleMeetings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	meetings := h.billService.GetCommitteeMeetings()
	respondJSON(w, meetings)
}

func (h *Handler) handleFindRepresentatives(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loc representatives.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if loc.ZIP == "" {
		http.Error(w, "ZIP code required", http.StatusBadRequest)
		return
	}

	reps, err := h.repService.FindByLocation(loc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, reps)
}

func (h *Handler) handleSearchRepresentatives(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' required", http.StatusBadRequest)
		return
	}

	reps := h.repService.SearchByName(query)
	respondJSON(w, reps)
}

func (h *Handler) handleEmailTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templates := h.emailService.GetTemplates()
	respondJSON(w, templates)
}

func (h *Handler) handleEmailCompose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req email.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var repEmails []struct{ ID, Email, Name string }
	for _, repID := range req.Representatives {
		rep := h.repService.GetByID(repID)
		if rep != nil {
			repEmails = append(repEmails, struct{ ID, Email, Name string }{
				ID:    rep.ID,
				Email: rep.Email,
				Name:  rep.Name,
			})
		}
	}

	if len(repEmails) == 0 {
		http.Error(w, "No valid representatives found", http.StatusBadRequest)
		return
	}

	mailtoLinks := h.emailService.GenerateMailtoLinks(req, repEmails)
	respondJSON(w, mailtoLinks)
}

func (h *Handler) handleEmailPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TemplateID string            `json:"templateId"`
		Variables  map[string]string `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subject, body, err := h.emailService.ComposeEmail(req.TemplateID, req.Variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]string{
		"subject": subject,
		"body":    body,
	})
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
