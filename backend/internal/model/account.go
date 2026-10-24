package model

// Account — пользователь сервиса. Роль "user" ищет питомца,
// роль "shelter" — приют или частный владелец, который отдаёт животное.
type Account struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	DisplayName  string `json:"display_name"`
	Bio          string `json:"bio"`
	Location     string `json:"location"`
}
