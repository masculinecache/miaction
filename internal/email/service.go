package email

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"
)

type Service struct {
	templates []EmailTemplate
	mu        sync.RWMutex
}

func NewService() *Service {
	s := &Service{}
	s.loadDefaultTemplates()
	return s
}

func (s *Service) loadDefaultTemplates() {
	s.templates = []EmailTemplate{
		{
			ID:          "support-bill",
			Name:        "Support Bill",
			Subject:     "Support for {{.BillNumber}}: {{.BillTitle}}",
			Description: "Template to express support for a bill",
			Body: `Dear {{.RepresentativeName}},

I am writing to express my strong support for {{.BillNumber}}, "{{.BillTitle}}."

{{.PersonalMessage}}

As your constituent, I believe this legislation is important for our community and the state of Michigan. I urge you to vote in favor of this bill when it comes before you.

Thank you for your time and consideration.

Sincerely,
{{.SenderName}}
{{.SenderAddress}}`,
			Variables: []TemplateVariable{
				{Name: "BillNumber", Description: "Bill number (e.g., HB 4001)", Required: true},
				{Name: "BillTitle", Description: "Full title of the bill", Required: true},
				{Name: "RepresentativeName", Description: "Name of the representative", Required: true},
				{Name: "PersonalMessage", Description: "Your personal message about why you support this bill", Required: false, Default: "This bill addresses an important issue that affects our community."},
				{Name: "SenderName", Description: "Your full name", Required: true},
				{Name: "SenderAddress", Description: "Your address", Required: true},
			},
		},
		{
			ID:          "oppose-bill",
			Name:        "Oppose Bill",
			Subject:     "Opposition to {{.BillNumber}}: {{.BillTitle}}",
			Description: "Template to express opposition to a bill",
			Body: `Dear {{.RepresentativeName}},

I am writing to express my strong opposition to {{.BillNumber}}, "{{.BillTitle}}."

{{.PersonalMessage}}

As your constituent, I have serious concerns about this legislation and its potential impact on our community. I urge you to vote against this bill when it comes before you.

Thank you for your time and consideration.

Sincerely,
{{.SenderName}}
{{.SenderAddress}}`,
			Variables: []TemplateVariable{
				{Name: "BillNumber", Description: "Bill number (e.g., HB 4001)", Required: true},
				{Name: "BillTitle", Description: "Full title of the bill", Required: true},
				{Name: "RepresentativeName", Description: "Name of the representative", Required: true},
				{Name: "PersonalMessage", Description: "Your personal message about why you oppose this bill", Required: false, Default: "I have serious concerns about the implications of this legislation."},
				{Name: "SenderName", Description: "Your full name", Required: true},
				{Name: "SenderAddress", Description: "Your address", Required: true},
			},
		},
		{
			ID:          "request-info",
			Name:        "Request Information",
			Subject:     "Request for Information on {{.BillNumber}}",
			Description: "Template to request more information about a bill",
			Body: `Dear {{.RepresentativeName}},

I am writing to request more information about {{.BillNumber}}, "{{.BillTitle}}," which is scheduled for a vote on {{.VoteDate}}.

{{.PersonalMessage}}

Could you please provide clarification on the following:
- The expected fiscal impact of this legislation
- How this bill might affect constituents in our district
- Your current position on this legislation

Thank you for your service to our community. I look forward to your response.

Sincerely,
{{.SenderName}}
{{.SenderAddress}}`,
			Variables: []TemplateVariable{
				{Name: "BillNumber", Description: "Bill number (e.g., HB 4001)", Required: true},
				{Name: "BillTitle", Description: "Full title of the bill", Required: true},
				{Name: "RepresentativeName", Description: "Name of the representative", Required: true},
				{Name: "VoteDate", Description: "Scheduled vote date", Required: true},
				{Name: "PersonalMessage", Description: "Your specific questions or concerns", Required: false, Default: "I would appreciate more information before this vote takes place."},
				{Name: "SenderName", Description: "Your full name", Required: true},
				{Name: "SenderAddress", Description: "Your address", Required: true},
			},
		},
	}
}

func (s *Service) GetTemplates() []EmailTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]EmailTemplate, len(s.templates))
	copy(result, s.templates)
	return result
}

func (s *Service) GetTemplate(id string) *EmailTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.templates {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

func (s *Service) ComposeEmail(templateID string, variables map[string]string) (subject, body string, err error) {
	tmpl := s.GetTemplate(templateID)
	if tmpl == nil {
		return "", "", fmt.Errorf("template not found: %s", templateID)
	}
	
	for _, v := range tmpl.Variables {
		if v.Required {
			if val, ok := variables[v.Name]; !ok || strings.TrimSpace(val) == "" {
				return "", "", fmt.Errorf("required variable missing: %s", v.Name)
			}
		}
	}
	
	subjTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse subject template: %w", err)
	}
	
	var subjBuf bytes.Buffer
	if err := subjTmpl.Execute(&subjBuf, variables); err != nil {
		return "", "", fmt.Errorf("failed to execute subject template: %w", err)
	}
	
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse body template: %w", err)
	}
	
	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, variables); err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %w", err)
	}
	
	return subjBuf.String(), bodyBuf.String(), nil
}

func (s *Service) GenerateMailtoLinks(req EmailRequest, reps []struct{ ID, Email, Name string }) []struct{ RepID, Name, Mailto string } {
	var result []struct{ RepID, Name, Mailto string }
	
	subject, body, err := s.ComposeEmail(req.TemplateID, req.Variables)
	if err != nil {
		return nil
	}
	
	for _, rep := range reps {
		if rep.Email == "" {
			continue
		}
		
		mailto := fmt.Sprintf("mailto:%s?subject=%s&body=%s",
			rep.Email,
			urlEncode(subject),
			urlEncode(body))
		
		result = append(result, struct{ RepID, Name, Mailto string }{
			RepID:  rep.ID,
			Name:   rep.Name,
			Mailto: mailto,
		})
	}
	
	return result
}

func urlEncode(s string) string {
	replacer := strings.NewReplacer(
		" ", "%20",
		"\n", "%0D%0A",
		"&", "%26",
		"?", "%3F",
		"=", "%3D",
		"#", "%23",
		"%", "%25",
	)
	return replacer.Replace(s)
}
