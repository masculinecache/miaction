package email

type EmailTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	Description string            `json:"description"`
	Variables   []TemplateVariable `json:"variables"`
}

type TemplateVariable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
}

type EmailRequest struct {
	TemplateID      string            `json:"templateId"`
	Variables       map[string]string `json:"variables"`
	Representatives []string          `json:"representativeIds"` // IDs of reps to email
	SenderName      string            `json:"senderName"`
	SenderEmail     string            `json:"senderEmail"`
}

type EmailResult struct {
	RepresentativeID string `json:"representativeId"`
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
}
