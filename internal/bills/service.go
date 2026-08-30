package bills

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type billEnrichment struct {
	subject           string
	analysisDocuments []AnalysisDocument
	billDocuments     []BillDocument
}

type billEnrichmentCacheEntry struct {
	enrichment billEnrichment
	fetched    time.Time
}

type Service struct {
	bills              []Bill
	meetings           []CommitteeMeeting
	mu                 sync.RWMutex
	lastFetch          time.Time
	fetchInterval      time.Duration
	enrichmentCache    map[string]billEnrichmentCacheEntry
	enrichmentCacheMu  sync.RWMutex
}

const analysisCacheTTL = 1 * time.Hour

func NewService() *Service {
	return &Service{
		fetchInterval:    15 * time.Minute,
		enrichmentCache:  make(map[string]billEnrichmentCacheEntry),
	}
}

func (s *Service) Start() {
	s.fetchBills()
	go s.backgroundFetch()
}

func (s *Service) backgroundFetch() {
	ticker := time.NewTicker(s.fetchInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.fetchBills()
	}
}

func (s *Service) fetchBills() {
	log.Println("Fetching Michigan legislature data...")
	
	bills := s.scrapeBills()
	meetings := s.scrapeMeetings()
	
	s.mu.Lock()
	s.bills = bills
	s.meetings = meetings
	s.lastFetch = time.Now()
	s.mu.Unlock()
	
	log.Printf("Fetched %d bills and %d meetings", len(bills), len(meetings))
}

func (s *Service) GetBills() []Bill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Bill, len(s.bills))
	copy(result, s.bills)
	return result
}

func (s *Service) GetScheduledBills() []Bill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var scheduled []Bill
	for _, b := range s.bills {
		if b.ScheduledVote != nil && b.ScheduledVote.Date.After(time.Now().Add(-24*time.Hour)) {
			scheduled = append(scheduled, b)
		}
	}
	return scheduled
}

func (s *Service) GetBill(id string) *Bill {
	s.mu.RLock()
	var bill *Bill
	for _, b := range s.bills {
		if b.ID == id {
			bill = &b
			break
		}
	}
	s.mu.RUnlock()

	if bill == nil {
		return nil
	}

	enr, err := s.getBillEnrichment(bill.URL)
	if err != nil {
		log.Printf("Warning: failed to fetch bill page details for %s: %v", bill.Number, err)
	} else {
		if enr.subject != "" {
			bill.Subject = enr.subject
		}
		bill.AnalysisDocuments = enr.analysisDocuments
		bill.BillDocuments = enr.billDocuments
	}

	return bill
}

func (s *Service) getBillEnrichment(billURL string) (*billEnrichment, error) {
	s.enrichmentCacheMu.RLock()
	cached, ok := s.enrichmentCache[billURL]
	valid := ok && time.Since(cached.fetched) < analysisCacheTTL
	s.enrichmentCacheMu.RUnlock()

	if valid {
		return &cached.enrichment, nil
	}

	enr, err := s.fetchBillDetails(billURL)
	if err != nil {
		return nil, err
	}

	s.enrichmentCacheMu.Lock()
	s.enrichmentCache[billURL] = billEnrichmentCacheEntry{
		enrichment: *enr,
		fetched:    time.Now(),
	}
	s.enrichmentCacheMu.Unlock()

	return enr, nil
}

var legislatureBase = "https://www.legislature.mi.gov"

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

func (s *Service) fetchBillDetails(billURL string) (*billEnrichment, error) {
	resp, err := httpClient.Get(billURL)
	if err != nil {
		return nil, fmt.Errorf("fetching bill page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bill page returned status %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing bill HTML: %w", err)
	}

	enr := &billEnrichment{}

	subjectDiv := findFirstNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "id", "ObjectSubject")
	})
	if subjectDiv != nil {
		enr.subject = strings.TrimSpace(extractText(subjectDiv))
	}

	analysisDiv := findFirstNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "id", "HFAAnalysisSection")
	})
	if analysisDiv != nil {
		analysisRows := findNodes(analysisDiv, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "class", "billDocRow")
		})
		for _, row := range analysisRows {
			ad := AnalysisDocument{}
			textDiv := findFirstNode(row, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "class", "text")
			})
			if textDiv != nil {
				strongNode := findFirstNode(textDiv, func(n *html.Node) bool {
					return n.Type == html.ElementNode && n.Data == "strong"
				})
				if strongNode != nil {
					ad.Title = extractText(strongNode)
				}
				var textSpans []*html.Node
				for c := textDiv.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "span" && getAttr(c, "class") == "text" {
						textSpans = append(textSpans, c)
					}
				}
				if len(textSpans) >= 2 {
					ad.Description = extractText(textSpans[1])
				}
			}
			if ad.Title != "" {
				enr.analysisDocuments = append(enr.analysisDocuments, ad)
			}
		}
	}

	docSection := findFirstNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "id", "BillDocumentSection")
	})
	if docSection != nil {
		docRows := findNodes(docSection, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "class", "billDocRow")
		})
		for _, row := range docRows {
			bd := BillDocument{}

			pdfDiv := findFirstNode(row, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "class", "pdf")
			})
			if pdfDiv != nil {
				pdfLink := findFirstNode(pdfDiv, func(n *html.Node) bool {
					return n.Type == html.ElementNode && n.Data == "a"
				})
				if pdfLink != nil {
					if href := getAttr(pdfLink, "href"); href != "" {
						bd.PDFURL = legislatureBase + href
					}
				}
			}

			htmlDiv := findFirstNode(row, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "class", "html")
			})
			if htmlDiv != nil {
				htmlLink := findFirstNode(htmlDiv, func(n *html.Node) bool {
					return n.Type == html.ElementNode && n.Data == "a"
				})
				if htmlLink != nil {
					if href := getAttr(htmlLink, "href"); href != "" {
						bd.HTMLURL = legislatureBase + href
					}
				}
			}

			textDiv := findFirstNode(row, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "div" && hasAttr(n, "class", "text")
			})
			if textDiv != nil {
				strongNode := findFirstNode(textDiv, func(n *html.Node) bool {
					return n.Type == html.ElementNode && n.Data == "strong"
				})
				if strongNode != nil {
					bd.Title = extractText(strongNode)
				}
				var descSpans []*html.Node
				for c := textDiv.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "span" {
						descSpans = append(descSpans, c)
					}
				}
				if len(descSpans) >= 2 {
					bd.Description = extractText(descSpans[1])
				}
			}

			if bd.Title != "" {
				enr.billDocuments = append(enr.billDocuments, bd)
			}
		}
	}

	return enr, nil
}

func hasAttr(n *html.Node, key, value string) bool {
	for _, a := range n.Attr {
		if a.Key == key && a.Val == value {
			return true
		}
	}
	return false
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func findFirstNode(n *html.Node, matcher func(*html.Node) bool) *html.Node {
	if matcher(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstNode(c, matcher); found != nil {
			return found
		}
	}
	return nil
}

func (s *Service) GetCommitteeMeetings() []CommitteeMeeting {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]CommitteeMeeting, len(s.meetings))
	copy(result, s.meetings)
	return result
}

func (s *Service) scrapeBills() []Bill {
	return generateSampleBills()
}

func (s *Service) scrapeMeetings() []CommitteeMeeting {
	return generateSampleMeetings()
}

func generateSampleBills() []Bill {
	now := time.Now()
	return []Bill{
		{
			ID:             "HB4001-2025",
			Number:         "HB 4001",
			Title:          "Income Tax: retirement benefits; deduction for certain retirement benefits",
			Description:    "To provide for a deduction from the individual income tax for certain retirement benefits received by a taxpayer",
			Chamber:        ChamberHouse,
			Status:         StatusScheduled,
			Sponsor:        "Rep. Jane Smith",
			CoSponsors:     []string{"Rep. John Doe", "Rep. Alice Johnson"},
			Committee:      "House Tax Policy",
			IntroducedDate: now.AddDate(0, -2, 0),
			LastAction:     "Reported with recommendation",
			LastActionDate: now.AddDate(0, 0, -3),
			ScheduledVote: &ScheduledVote{
				Date:        now.AddDate(0, 0, 2),
				Body:        "house",
				VoteType:    "floor",
				Description: "Second reading",
			},
			URL: "https://www.legislature.mi.gov/Search/BillSearch?query=HB4001&defaultTypes=House+Bill,Senate+Bill&session=2025-2026",
		},
		{
			ID:             "SB200-2025",
			Number:         "SB 200",
			Title:          "Education: safety; school safety and security measures",
			Description:    "To establish requirements for school safety and security measures in public schools",
			Chamber:        ChamberSenate,
			Status:         StatusScheduled,
			Sponsor:        "Sen. Robert Wilson",
			CoSponsors:     []string{"Sen. Maria Garcia"},
			Committee:      "Senate Education",
			IntroducedDate: now.AddDate(0, -1, 0),
			LastAction:     "Referred to committee",
			LastActionDate: now.AddDate(0, 0, -1),
			ScheduledVote: &ScheduledVote{
				Date:        now.AddDate(0, 0, 5),
				Body:        "Senate Education",
				VoteType:    "committee",
				Description: "Committee vote on reporting",
			},
			URL: "https://www.legislature.mi.gov/Search/BillSearch?query=SB200&defaultTypes=House+Bill,Senate+Bill&session=2025-2026",
		},
		{
			ID:             "HB4150-2025",
			Number:         "HB 4150",
			Title:          "Transportation: vehicles; autonomous vehicle regulations",
			Description:    "To establish regulations for the operation of autonomous vehicles on public roads",
			Chamber:        ChamberHouse,
			Status:         StatusCommittee,
			Sponsor:        "Rep. Michael Brown",
			CoSponsors:     []string{},
			Committee:      "House Transportation",
			IntroducedDate: now.AddDate(0, -3, 0),
			LastAction:     "Scheduled for hearing",
			LastActionDate: now.AddDate(0, 0, -5),
			URL:            "https://www.legislature.mi.gov/Search/BillSearch?query=HB4150&defaultTypes=House+Bill,Senate+Bill&session=2025-2026",
		},
		{
			ID:             "SB180-2025",
			Number:         "SB 180",
			Title:          "Environment: water; Great Lakes water quality standards",
			Description:    "To establish water quality standards for the protection of the Great Lakes",
			Chamber:        ChamberSenate,
			Status:         StatusScheduled,
			Sponsor:        "Sen. Sarah Davis",
			CoSponsors:     []string{"Sen. James Miller", "Sen. Lisa Chen"},
			Committee:      "Senate Energy and Environment",
			IntroducedDate: now.AddDate(0, -1, -15),
			LastAction:     "Reported favorably",
			LastActionDate: now.AddDate(0, 0, -7),
			ScheduledVote: &ScheduledVote{
				Date:        now.AddDate(0, 0, 3),
				Body:        "senate",
				VoteType:    "floor",
				Description: "Third reading and final passage",
			},
			URL: "https://www.legislature.mi.gov/Search/BillSearch?query=SB180&defaultTypes=House+Bill,Senate+Bill&session=2025-2026",
		},
		{
			ID:             "HB4200-2025",
			Number:         "HB 4200",
			Title:          "Health: insurance; prior authorization reform",
			Description:    "To establish requirements for health insurance prior authorization processes",
			Chamber:        ChamberHouse,
			Status:         StatusCommittee,
			Sponsor:        "Rep. David Lee",
			CoSponsors:     []string{"Rep. Karen White"},
			Committee:      "House Health Policy",
			IntroducedDate: now.AddDate(0, -2, -10),
			LastAction:     "Referred to committee",
			LastActionDate: now.AddDate(0, 0, -10),
			URL:            "https://www.legislature.mi.gov/Search/BillSearch?query=HB4200&defaultTypes=House+Bill,Senate+Bill&session=2025-2026",
		},
		{
			ID:             "SB220-2025",
			Number:         "SB 220",
			Title:          "Labor: wages; minimum wage increase",
			Description:    "To increase the state minimum wage and establish annual adjustments",
			Chamber:        ChamberSenate,
			Status:         StatusScheduled,
			Sponsor:        "Sen. Patricia Taylor",
			CoSponsors:     []string{"Sen. Christopher Anderson"},
			Committee:      "Senate Labor",
			IntroducedDate: now.AddDate(0, -3, -5),
			LastAction:     "Reported with amendments",
			LastActionDate: now.AddDate(0, 0, -2),
			ScheduledVote: &ScheduledVote{
				Date:        now.AddDate(0, 0, 7),
				Body:        "Senate Labor",
				VoteType:    "committee",
				Description: "Committee vote on substitute",
			},
			URL: "https://www.legislature.mi.gov/Search/BillSearch?query=SB220&defaultTypes=House+Bill,Senate+Bill&session=2025-2026",
		},
	}
}

func generateSampleMeetings() []CommitteeMeeting {
	now := time.Now()
	return []CommitteeMeeting{
		{
			Committee: "House Tax Policy",
			Date:      now.AddDate(0, 0, 2),
			Time:      "9:00 AM",
			Location:  "Room 521, House Office Building",
			Bills:     []string{"HB 4001"},
			Chamber:   ChamberHouse,
			URL:       "https://www.legislature.mi.gov/Committees/Meetings",
		},
		{
			Committee: "Senate Education",
			Date:      now.AddDate(0, 0, 5),
			Time:      "12:30 PM",
			Location:  "Room 1100, Binsfeld Office Building",
			Bills:     []string{"SB 200"},
			Chamber:   ChamberSenate,
			URL:       "https://www.legislature.mi.gov/Committees/Meetings",
		},
		{
			Committee: "Senate Energy and Environment",
			Date:      now.AddDate(0, 0, 3),
			Time:      "1:30 PM",
			Location:  "Room 1300, Binsfeld Office Building",
			Bills:     []string{"SB 180"},
			Chamber:   ChamberSenate,
			URL:       "https://www.legislature.mi.gov/Committees/Meetings",
		},
	}
}

func ScrapeLegiScanBills(apiKey string) ([]Bill, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("LegiScan API key required")
	}
	
	url := fmt.Sprintf("https://api.legiscan.com/v1/?key=%s&op=getSessionList&state=22", apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	_ = body
	return nil, fmt.Errorf("LegiScan integration not yet fully implemented")
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var result string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += extractText(c)
	}
	return result
}

func findNodes(n *html.Node, matcher func(*html.Node) bool) []*html.Node {
	var result []*html.Node
	if matcher(n) {
		result = append(result, n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, findNodes(c, matcher)...)
	}
	return result
}
