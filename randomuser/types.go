package randomuser

import "encoding/json"

// User is one randomly generated user profile from randomuser.me.
type User struct {
	Gender     string `kit:"id" json:"gender"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Title      string `json:"title"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Nat        string `json:"nationality"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	Postcode   string `json:"postcode"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
	BirthDate  string `json:"birth_date"`
	Age        int    `json:"age"`
}

// rawUser is the wire shape returned by randomuser.me.
type rawUser struct {
	Gender string `json:"gender"`
	Name   struct {
		Title string `json:"title"`
		First string `json:"first"`
		Last  string `json:"last"`
	} `json:"name"`
	Location struct {
		Street struct {
			Number int    `json:"number"`
			Name   string `json:"name"`
		} `json:"street"`
		City     string          `json:"city"`
		State    string          `json:"state"`
		Country  string          `json:"country"`
		Postcode json.RawMessage `json:"postcode"`
		Coordinates struct {
			Latitude  string `json:"latitude"`
			Longitude string `json:"longitude"`
		} `json:"coordinates"`
	} `json:"location"`
	Email string `json:"email"`
	DOB   struct {
		Date string `json:"date"` // "1990-05-12T00:00:00.000Z"
		Age  int    `json:"age"`
	} `json:"dob"`
	Phone string `json:"phone"`
	Nat   string `json:"nat"`
}

// apiResponse is the top-level JSON envelope from randomuser.me.
type apiResponse struct {
	Results []rawUser `json:"results"`
	Info    struct {
		Seed    string `json:"seed"`
		Results int    `json:"results"`
		Page    int    `json:"page"`
		Version string `json:"version"`
	} `json:"info"`
}

// parsePostcode converts the postcode field (string or number) to a string.
func parsePostcode(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// try unquoted string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// number — return as raw digits
	return string(raw)
}
