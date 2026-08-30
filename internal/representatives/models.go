package representatives

type Representative struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Chamber     string   `json:"chamber"` // "house" or "senate"
	District    int      `json:"district"`
	Party       string   `json:"party"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Office      string   `json:"office"`
	URL         string   `json:"url"`
	PhotoURL    string   `json:"photoUrl,omitempty"`
}

type Location struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZIP        string `json:"zip"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
}

type DistrictInfo struct {
	HouseDistrict  int `json:"houseDistrict"`
	SenateDistrict int `json:"senateDistrict"`
}
