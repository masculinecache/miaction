package representatives

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type Service struct {
	reps []Representative
	mu   sync.RWMutex
}

func NewService() *Service {
	s := &Service{}
	s.loadSampleData()
	return s
}

func (s *Service) loadSampleData() {
	var reps []Representative
	
	parties := []string{"Republican", "Democratic"}
	
	for i := 1; i <= 110; i++ {
		party := parties[i%2]
		reps = append(reps, Representative{
			ID:       fmt.Sprintf("house-%d", i),
			Name:     fmt.Sprintf("Rep. %s %s", randomFirstName(i), randomLastName(i)),
			Chamber:  "house",
			District: i,
			Party:    party,
			Email:    fmt.Sprintf("rep%d@house.mi.gov", i),
			Phone:    fmt.Sprintf("(517) 373-%04d", 1000+i),
			Office:   fmt.Sprintf("Room %d, House Office Building", 100+i),
			URL:      fmt.Sprintf("https://www.house.mi.gov/repdetail?district=%d", i),
		})
	}
	
	for i := 1; i <= 38; i++ {
		party := parties[(i+1)%2]
		reps = append(reps, Representative{
			ID:       fmt.Sprintf("senate-%d", i),
			Name:     fmt.Sprintf("Sen. %s %s", randomFirstName(i+200), randomLastName(i+200)),
			Chamber:  "senate",
			District: i,
			Party:    party,
			Email:    fmt.Sprintf("sen%d@senate.mi.gov", i),
			Phone:    fmt.Sprintf("(517) 373-%04d", 2000+i),
			Office:   fmt.Sprintf("Room %d, Binsfeld Office Building", 200+i),
			URL:      fmt.Sprintf("https://www.senate.michigan.gov/senatorinfo.php?district=%d", i),
		})
	}
	
	s.reps = reps
}

func (s *Service) FindByLocation(loc Location) ([]Representative, error) {
	houseDistrict, senateDistrict := zipToDistricts(loc.ZIP)
	
	if houseDistrict == 0 || senateDistrict == 0 {
		return nil, fmt.Errorf("unable to determine districts for ZIP code: %s", loc.ZIP)
	}
	
	return s.FindByDistricts(houseDistrict, senateDistrict)
}

func (s *Service) FindByDistricts(houseDistrict, senateDistrict int) ([]Representative, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var result []Representative
	for _, r := range s.reps {
		if r.Chamber == "house" && r.District == houseDistrict {
			result = append(result, r)
		}
		if r.Chamber == "senate" && r.District == senateDistrict {
			result = append(result, r)
		}
	}
	
	if len(result) == 0 {
		return nil, fmt.Errorf("no representatives found for districts House=%d, Senate=%d", houseDistrict, senateDistrict)
	}
	
	return result, nil
}

func (s *Service) GetByID(id string) *Representative {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.reps {
		if r.ID == id {
			return &r
		}
	}
	return nil
}

func (s *Service) GetAll() []Representative {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Representative, len(s.reps))
	copy(result, s.reps)
	return result
}

func zipToDistricts(zip string) (houseDistrict, senateDistrict int) {
	if len(zip) < 3 {
		return 0, 0
	}
	
	prefix := zip[:3]
	num, err := strconv.Atoi(prefix)
	if err != nil {
		return 0, 0
	}
	
	houseDistrict = (num % 110) + 1
	senateDistrict = (num % 38) + 1
	
	return houseDistrict, senateDistrict
}

var firstNames = []string{
	"James", "Mary", "Robert", "Patricia", "John", "Jennifer", "Michael", "Linda",
	"David", "Elizabeth", "William", "Barbara", "Richard", "Susan", "Joseph", "Jessica",
	"Thomas", "Sarah", "Charles", "Karen", "Christopher", "Nancy", "Daniel", "Lisa",
	"Matthew", "Betty", "Anthony", "Margaret", "Mark", "Sandra",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
	"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
	"Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson",
	"White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson",
}

func randomFirstName(seed int) string {
	return firstNames[seed%len(firstNames)]
}

func randomLastName(seed int) string {
	return lastNames[seed%len(lastNames)]
}

func (s *Service) SearchByName(query string) []Representative {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	query = strings.ToLower(query)
	var result []Representative
	for _, r := range s.reps {
		if strings.Contains(strings.ToLower(r.Name), query) {
			result = append(result, r)
		}
	}
	return result
}
