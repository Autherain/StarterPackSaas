package store

import (
	"context"
	"errors"
	"time"

	"github.com/autherain/test/internal/user"
	"github.com/autherain/test/internal/user/store/models"
	"github.com/autherain/test/internal/user/store/modext"
	"gorm.io/gorm"
)

const defaultTimeout = 3 * time.Second

type baseStore interface {
	GetDB() *gorm.DB
}

type usersStore struct {
	baseStore baseStore
}

func New(baseStore baseStore) user.UsersReadWriter {
	return &usersStore{baseStore: baseStore}
}

func (s *usersStore) CreateUser(u *user.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	model := modext.ToModel(u)
	if err := s.baseStore.GetDB().WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	return nil
}

func (s *usersStore) ReadUser(selector *user.UserSelector) (*user.User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var model models.User
	query := s.baseStore.GetDB().WithContext(ctx).Model(&models.User{})

	query = buildUserQuery(query, selector)

	err := query.First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return modext.MapUser(&model), true, nil
}

func buildUserQuery(query *gorm.DB, selector *user.UserSelector) *gorm.DB {
	if selector.ID != 0 {
		query = query.Where("id = ?", selector.ID)
	}
	if selector.Email != "" {
		query = query.Where("LOWER(email) = LOWER(?)", selector.Email)
	}
	return query
}

func (s *usersStore) UpdateUser(u *user.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	model := modext.ToModel(u)
	result := s.baseStore.GetDB().WithContext(ctx).Model(&models.User{}).
		Where("id = ?", u.ID).
		Updates(model)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (s *usersStore) DeleteUser(selector *user.UserSelector) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := s.baseStore.GetDB().WithContext(ctx).Model(&models.User{})
	query = buildUserQuery(query, selector)

	result := query.Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
