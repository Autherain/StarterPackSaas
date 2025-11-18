package modext

import (
	"github.com/autherain/test/internal/user"
	"github.com/autherain/test/internal/user/store/models"
)

func MapUser(m *models.User) *user.User {
	if m == nil {
		return nil
	}

	return &user.User{
		ID:               m.ID,
		Created:          m.Created,
		Email:            m.Email,
		HashedPassword:   m.HashedPassword,
		SubscriptionTier: m.SubscriptionTier,
	}
}

func MapUsers(s []*models.User) []*user.User {
	if s == nil {
		return nil
	}

	result := make([]*user.User, 0, len(s))
	for _, e := range s {
		result = append(result, MapUser(e))
	}

	return result
}

func ToModel(u *user.User) *models.User {
	if u == nil {
		return nil
	}

	return &models.User{
		ID:               u.ID,
		Created:          u.Created,
		Email:            u.Email,
		HashedPassword:   u.HashedPassword,
		SubscriptionTier: u.SubscriptionTier,
	}
}
