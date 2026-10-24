package service

import "petmatch.local/backend/internal/model"

type PreferencesRepository interface {
	SavePreferences(userID int, p model.Preferences) error
	GetPreferences(userID int) (model.Preferences, error)
}

type PreferencesService struct {
	repo PreferencesRepository
}

func NewPreferencesService(r PreferencesRepository) *PreferencesService {
	return &PreferencesService{repo: r}
}

func (s *PreferencesService) Set(userID int, p model.Preferences) (model.Preferences, error) {
	if p.PreferredType != "" && !petTypes[p.PreferredType] {
		return p, invalid("Тип животного: dog, cat или other")
	}
	if p.MinAge < 0 || p.MaxAge < 0 || (p.MaxAge > 0 && p.MinAge > p.MaxAge) {
		return p, invalid("Некорректный диапазон возраста")
	}
	return p, s.repo.SavePreferences(userID, p)
}

// Get возвращает пустые предпочтения, если пользователь их ещё не задавал.
func (s *PreferencesService) Get(userID int) (model.Preferences, error) {
	p, err := s.repo.GetPreferences(userID)
	if isNoRows(err) {
		return model.Preferences{}, nil
	}
	return p, err
}
