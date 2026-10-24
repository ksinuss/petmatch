package service

import "petmatch.local/backend/internal/model"

type SwipeRepository interface {
	CreateSwipe(userID, petID int, isLike bool) error
	CreateOwnerSwipe(ownerID, userID, petID int, isLike bool) error
	HasOwnerLike(userID, petID int) (bool, error)
	HasUserLike(userID, petID int) (bool, error)
	CreateMatch(userID, petID int) (int, error)
	GetPetInfo(petID int) (ownerID int, mode string, err error)
	GetMatches(userID int) ([]model.MatchView, error)
}

type SwipeService struct {
	repo SwipeRepository
}

func NewSwipeService(r SwipeRepository) *SwipeService {
	return &SwipeService{repo: r}
}

// Swipe — основная бизнес-логика: лайк в режиме open даёт мэтч сразу,
// в режиме strict — только если владелец уже лайкнул этого пользователя.
func (s *SwipeService) Swipe(userID, petID int, isLike bool) (model.SwipeResult, error) {
	ownerID, mode, err := s.repo.GetPetInfo(petID)
	if err != nil {
		if isNoRows(err) {
			return model.SwipeResult{}, notFound("Питомец не найден")
		}
		return model.SwipeResult{}, err
	}
	if ownerID == userID {
		return model.SwipeResult{}, forbidden("Нельзя оценивать собственного питомца")
	}

	if err := s.repo.CreateSwipe(userID, petID, isLike); err != nil {
		if isUnique(err) {
			return model.SwipeResult{}, conflict("Вы уже оценили этого питомца")
		}
		return model.SwipeResult{}, err
	}
	if !isLike {
		return model.SwipeResult{}, nil
	}

	if mode == model.AdoptionStrict {
		ok, err := s.repo.HasOwnerLike(userID, petID)
		if err != nil || !ok {
			return model.SwipeResult{}, err
		}
	}
	return s.match(userID, petID)
}

// OwnerSwipe — владелец отвечает на лайк пользователя (режим strict).
func (s *SwipeService) OwnerSwipe(ownerID, userID, petID int, isLike bool) (model.SwipeResult, error) {
	realOwner, _, err := s.repo.GetPetInfo(petID)
	if err != nil {
		if isNoRows(err) {
			return model.SwipeResult{}, notFound("Питомец не найден")
		}
		return model.SwipeResult{}, err
	}
	if realOwner != ownerID {
		return model.SwipeResult{}, forbidden("Это не ваш питомец")
	}

	if err := s.repo.CreateOwnerSwipe(ownerID, userID, petID, isLike); err != nil {
		if isUnique(err) {
			return model.SwipeResult{}, conflict("Вы уже ответили этому пользователю")
		}
		return model.SwipeResult{}, err
	}
	if !isLike {
		return model.SwipeResult{}, nil
	}

	liked, err := s.repo.HasUserLike(userID, petID)
	if err != nil || !liked {
		return model.SwipeResult{}, err
	}
	return s.match(userID, petID)
}

func (s *SwipeService) match(userID, petID int) (model.SwipeResult, error) {
	id, err := s.repo.CreateMatch(userID, petID)
	if err != nil {
		return model.SwipeResult{}, err
	}
	return model.SwipeResult{Match: true, MatchID: id}, nil
}

func (s *SwipeService) GetMatches(userID int) ([]model.MatchView, error) {
	return s.repo.GetMatches(userID)
}
