package model

type MatchPet struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Breed string `json:"breed"`
	Photo string `json:"photo"`
}

type MatchPartner struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

// MatchView — мэтч: питомец и второй участник.
type MatchView struct {
	ID        int          `json:"id"`
	CreatedAt string       `json:"created_at"`
	Pet       MatchPet     `json:"pet"`
	Partner   MatchPartner `json:"partner"`
}

type SwipeResult struct {
	Match   bool `json:"match"`
	MatchID int  `json:"match_id,omitempty"`
}
