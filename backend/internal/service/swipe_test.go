package service

import (
	"database/sql"
	"errors"
	"testing"

	"petmatch.local/backend/internal/model"
)

type fakeSwipeRepo struct {
	pets map[int]struct {
		owner int
		mode  string
	}
	swipes     map[[2]int]bool
	ownerLikes map[[2]int]bool
	matches    map[[2]int]int
}

func newFakeSwipeRepo() *fakeSwipeRepo {
	return &fakeSwipeRepo{
		pets: map[int]struct {
			owner int
			mode  string
		}{
			1: {owner: 100, mode: model.AdoptionOpen},
			2: {owner: 100, mode: model.AdoptionStrict},
		},
		swipes: map[[2]int]bool{}, ownerLikes: map[[2]int]bool{}, matches: map[[2]int]int{},
	}
}

func (f *fakeSwipeRepo) CreateSwipe(u, p int, like bool) error {
	if _, ok := f.swipes[[2]int{u, p}]; ok {
		return errors.New("UNIQUE constraint failed: swipes.user_id, swipes.pet_id")
	}
	f.swipes[[2]int{u, p}] = like
	return nil
}
func (f *fakeSwipeRepo) CreateOwnerSwipe(o, u, p int, like bool) error {
	f.ownerLikes[[2]int{u, p}] = like
	return nil
}
func (f *fakeSwipeRepo) HasOwnerLike(u, p int) (bool, error) { return f.ownerLikes[[2]int{u, p}], nil }
func (f *fakeSwipeRepo) HasUserLike(u, p int) (bool, error)  { return f.swipes[[2]int{u, p}], nil }
func (f *fakeSwipeRepo) CreateMatch(u, p int) (int, error) {
	k := [2]int{u, p}
	if id, ok := f.matches[k]; ok {
		return id, nil
	}
	f.matches[k] = len(f.matches) + 1
	return f.matches[k], nil
}
func (f *fakeSwipeRepo) GetPetInfo(p int) (int, string, error) {
	pet, ok := f.pets[p]
	if !ok {
		return 0, "", sql.ErrNoRows
	}
	return pet.owner, pet.mode, nil
}
func (f *fakeSwipeRepo) GetMatches(int) ([]model.MatchView, error) { return nil, nil }

func TestSwipeOpenModeMatchesOnLike(t *testing.T) {
	s := NewSwipeService(newFakeSwipeRepo())
	res, err := s.Swipe(1, 1, true)
	if err != nil || !res.Match || res.MatchID == 0 {
		t.Fatalf("лайк в open-режиме должен дать мэтч, получили %+v, err=%v", res, err)
	}
}

func TestSwipeDislikeNeverMatches(t *testing.T) {
	s := NewSwipeService(newFakeSwipeRepo())
	res, err := s.Swipe(1, 1, false)
	if err != nil || res.Match {
		t.Fatalf("дизлайк не должен давать мэтч: %+v, err=%v", res, err)
	}
}

func TestSwipeStrictModeNeedsOwnerLike(t *testing.T) {
	s := NewSwipeService(newFakeSwipeRepo())

	res, err := s.Swipe(1, 2, true)
	if err != nil || res.Match {
		t.Fatalf("в strict-режиме мэтча без ответа владельца быть не должно: %+v, err=%v", res, err)
	}
	res, err = s.OwnerSwipe(100, 1, 2, true)
	if err != nil || !res.Match {
		t.Fatalf("лайк владельца должен завершить мэтч: %+v, err=%v", res, err)
	}
}

func TestSwipeStrictOwnerFirst(t *testing.T) {
	s := NewSwipeService(newFakeSwipeRepo())
	if res, _ := s.OwnerSwipe(100, 1, 2, true); res.Match {
		t.Fatal("пока пользователь не лайкнул, мэтча быть не должно")
	}
	res, err := s.Swipe(1, 2, true)
	if err != nil || !res.Match {
		t.Fatalf("лайк пользователя после лайка владельца должен дать мэтч: %+v, err=%v", res, err)
	}
}

func TestSwipeErrors(t *testing.T) {
	s := NewSwipeService(newFakeSwipeRepo())

	if _, err := s.Swipe(100, 1, true); !errors.Is(err, ErrForbidden) {
		t.Errorf("свой питомец: ожидали ErrForbidden, получили %v", err)
	}
	if _, err := s.Swipe(1, 999, true); !errors.Is(err, ErrNotFound) {
		t.Errorf("нет питомца: ожидали ErrNotFound, получили %v", err)
	}
	_, _ = s.Swipe(1, 1, true)
	if _, err := s.Swipe(1, 1, false); !errors.Is(err, ErrConflict) {
		t.Errorf("повторный свайп: ожидали ErrConflict, получили %v", err)
	}
	if _, err := s.OwnerSwipe(555, 1, 1, true); !errors.Is(err, ErrForbidden) {
		t.Errorf("чужой питомец в owner-swipe: ожидали ErrForbidden, получили %v", err)
	}
}
