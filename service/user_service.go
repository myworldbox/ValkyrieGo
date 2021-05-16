package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/sentrionic/valkyrie/model"
	"github.com/sentrionic/valkyrie/model/apperrors"
	"log"
	"mime/multipart"
)

// UserService acts as a struct for injecting an implementation of UserRepository
// for use in service methods
type userService struct {
	UserRepository  model.UserRepository
	ImageRepository model.ImageRepository
	RedisRepository model.RedisRepository
	MailRepository  model.MailRepository
}

// USConfig will hold repositories that will eventually be injected into this
// this service layer
type USConfig struct {
	UserRepository  model.UserRepository
	ImageRepository model.ImageRepository
	RedisRepository model.RedisRepository
	MailRepository  model.MailRepository
}

// NewUserService is a factory function for
// initializing a UserService with its repository layer dependencies
func NewUserService(c *USConfig) model.UserService {
	return &userService{
		UserRepository:  c.UserRepository,
		ImageRepository: c.ImageRepository,
		RedisRepository: c.RedisRepository,
		MailRepository:  c.MailRepository,
	}
}

// Get retrieves a user based on their uuid
func (s *userService) Get(uid string) (*model.User, error) {
	u, err := s.UserRepository.FindByID(uid)

	return u, err
}

// GetByEmail retrieves a user based on their email
func (s *userService) GetByEmail(email string) (*model.User, error) {
	u, err := s.UserRepository.FindByEmail(email)

	return u, err
}

func (s *userService) Register(u *model.User) error {
	pw, err := hashPassword(u.Password)

	if err != nil {
		log.Printf("Unable to signup user for email: %v\n", u.Email)
		return apperrors.NewInternal()
	}

	u.Password = pw

	id, _ := GenerateId()
	u.ID = id
	u.Image = fmt.Sprintf("https://gravatar.com/avatar/%s?d=identicon", getMD5Hash(u.Email))

	if err := s.UserRepository.Create(u); err != nil {
		return err
	}

	return nil
}

// Login reaches our to a UserRepository check if the user exists
// and then compares the supplied password with the provided password
// if a valid email/password combo is provided, u will hold all
// available user fields
func (s *userService) Login(u *model.User) error {
	uFetched, err := s.UserRepository.FindByEmail(u.Email)

	// Will return NotAuthorized to client to omit details of why
	if err != nil {
		return apperrors.NewAuthorization("Invalid email and password combination")
	}

	// verify password - we previously created this method
	match, err := comparePasswords(uFetched.Password, u.Password)

	if err != nil {
		return apperrors.NewInternal()
	}

	if !match {
		return apperrors.NewAuthorization("Invalid email and password combination")
	}

	*u = *uFetched
	return nil
}

func (s *userService) UpdateAccount(u *model.User) error {
	// Update user in UserRepository
	err := s.UserRepository.Update(u)

	if err != nil {
		return err
	}

	return nil
}

func (s *userService) CheckEmail(email string) bool {
	user, _ := s.UserRepository.FindByEmail(email)
	return user.ID != ""
}

func (s *userService) ChangeAvatar(header *multipart.FileHeader, directory string) (string, error) {
	return s.ImageRepository.UploadAvatar(header, directory)
}

func (s *userService) DeleteImage(key string) error {
	return s.ImageRepository.DeleteImage(key)
}

func (s *userService) ChangePassword(password string, u *model.User) error {
	pw, err := hashPassword(password)

	if err != nil {
		log.Printf("Unable to signup user for email: %v\n", u.Email)
		return apperrors.NewInternal()
	}

	u.Password = pw

	if err := s.UserRepository.Update(u); err != nil {
		return err
	}

	return nil
}

func (s *userService) ForgotPassword(ctx context.Context, user *model.User) error {
	token, err := s.RedisRepository.SetResetToken(ctx, user.ID)

	if err != nil {
		return err
	}

	err = s.MailRepository.SendMail(user.Email, token)

	return err
}

func (s *userService) ResetPassword(ctx context.Context, password string, token string) (*model.User, error) {
	id, err := s.RedisRepository.GetIdFromToken(ctx, token)

	if err != nil {
		return nil, err
	}

	user, err := s.UserRepository.FindByID(id)

	if err != nil {
		return nil, err
	}

	pw, err := hashPassword(password)

	if err != nil {
		log.Printf("Unable to reset password")
		return nil, apperrors.NewInternal()
	}

	user.Password = pw

	if err := s.UserRepository.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func getMD5Hash(email string) string {
	hash := md5.Sum([]byte(email))
	return hex.EncodeToString(hash[:])
}
