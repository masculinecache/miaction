package bills

import "time"

type BillStatus string

const (
	StatusIntroduced   BillStatus = "introduced"
	StatusCommittee    BillStatus = "committee"
	StatusScheduled    BillStatus = "scheduled"
	StatusPassed       BillStatus = "passed"
	StatusFailed       BillStatus = "failed"
	StatusSigned       BillStatus = "signed"
	StatusVetoed       BillStatus = "vetoed"
)

type Chamber string

const (
	ChamberHouse  Chamber = "house"
	ChamberSenate Chamber = "senate"
)

type Bill struct {
	ID                 string             `json:"id"`
	Number             string             `json:"number"`
	Title              string             `json:"title"`
	Description        string             `json:"description"`
	Subject            string             `json:"subject,omitempty"`
	Chamber            Chamber            `json:"chamber"`
	Status             BillStatus         `json:"status"`
	Sponsor            string             `json:"sponsor"`
	CoSponsors         []string           `json:"coSponsors"`
	Committee          string             `json:"committee,omitempty"`
	IntroducedDate     time.Time          `json:"introducedDate"`
	LastAction         string             `json:"lastAction"`
	LastActionDate     time.Time          `json:"lastActionDate"`
	ScheduledVote      *ScheduledVote     `json:"scheduledVote,omitempty"`
	URL                string             `json:"url"`
	AnalysisDocuments  []AnalysisDocument `json:"analysisDocuments,omitempty"`
	BillDocuments      []BillDocument     `json:"billDocuments,omitempty"`
}

type AnalysisDocument struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type BillDocument struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	PDFURL      string `json:"pdfUrl,omitempty"`
	HTMLURL     string `json:"htmlUrl,omitempty"`
}

type ScheduledVote struct {
	Date        time.Time `json:"date"`
	Body        string    `json:"body"` // "house", "senate", or committee name
	VoteType    string    `json:"voteType"` // "committee", "floor"
	Description string    `json:"description"`
}

type CommitteeMeeting struct {
	Committee   string    `json:"committee"`
	Date        time.Time `json:"date"`
	Time        string    `json:"time"`
	Location    string    `json:"location"`
	Bills       []string  `json:"bills"` // Bill numbers
	Chamber     Chamber   `json:"chamber"`
	URL         string    `json:"url"`
}
