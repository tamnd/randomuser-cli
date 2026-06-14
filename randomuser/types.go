package randomuser

import "encoding/json"

// User is one randomly generated user profile from randomuser.me.
type User struct {
	UUID     string `kit:"id" json:"uuid"` // login.uuid
	Name     string `json:"name"`          // "First Last"
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`     // dob.age
	City     string `json:"city"`    // location.city
	State    string `json:"state"`   // location.state
	Country  string `json:"country"` // location.country
	Nat      string `json:"nat"`
	Username string `json:"username"` // login.username
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
		City        string          `json:"city"`
		State       string          `json:"state"`
		Country     string          `json:"country"`
		Postcode    json.RawMessage `json:"postcode"`
		Coordinates struct {
			Latitude  string `json:"latitude"`
			Longitude string `json:"longitude"`
		} `json:"coordinates"`
	} `json:"location"`
	Email string `json:"email"`
	Login struct {
		UUID     string `json:"uuid"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"login"`
	DOB struct {
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
	// try quoted string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// number — return as raw digits
	return string(raw)
}
