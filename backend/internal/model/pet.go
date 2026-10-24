package model

// Допустимые значения полей питомца.
const (
	AdoptionOpen   = "open"   // мэтч сразу после лайка пользователя
	AdoptionStrict = "strict" // мэтч только если владелец тоже лайкнул пользователя
)

type Pet struct {
	ID           int      `json:"id"`
	OwnerID      int      `json:"owner_id"`
	OwnerName    string   `json:"owner_name"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Breed        string   `json:"breed"`
	Age          int      `json:"age"`
	Description  string   `json:"description"`
	AdoptionMode string   `json:"adoption_mode"`
	Status       string   `json:"status"`
	Photos       []string `json:"photos"`
}

// Preferences — предпочтения пользователя для ранжирования ленты.
type Preferences struct {
	PreferredType     string `json:"preferred_type"`
	MinAge            int    `json:"min_age"`
	MaxAge            int    `json:"max_age"`
	PreferredBreed    string `json:"preferred_breed"`
	PreferredLocation string `json:"preferred_location"`
}
