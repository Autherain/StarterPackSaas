package params

import (
	"time"

	"github.com/autherain/test/internal/user"
)

type CreateUserParams struct {
	Email            string `json:"Email"`
	Password         string `json:"Password"`
	SubscriptionTier string `json:"SubscriptionTier,omitempty"`
}

func (p *CreateUserParams) Map(hashedPassword string) *user.User {
	tier := p.SubscriptionTier
	if tier == "" {
		tier = "free"
	}

	return &user.User{
		Email:            p.Email,
		HashedPassword:   hashedPassword,
		SubscriptionTier: tier,
		Created:          time.Now(),
	}
}

type UpdateUserParams struct {
	ID               int    `json:"-"`
	Email            string `json:"Email,omitempty"`
	SubscriptionTier string `json:"SubscriptionTier,omitempty"`
}

func (p *UpdateUserParams) Map(existingUser *user.User) *user.User {
	updated := &user.User{
		ID:               p.ID,
		Email:            existingUser.Email,
		HashedPassword:   existingUser.HashedPassword,
		SubscriptionTier: existingUser.SubscriptionTier,
		Created:          existingUser.Created,
	}

	if p.Email != "" {
		updated.Email = p.Email
	}

	if p.SubscriptionTier != "" {
		updated.SubscriptionTier = p.SubscriptionTier
	}

	return updated
}

type CreateAuthenticationTokenParams struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}
