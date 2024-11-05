package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"
)

type UserStorage interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, id int) (models.App, error)
}

type Auth struct {
	log         *slog.Logger
	usrStorage  UserStorage
	AppProvider AppProvider
	tokenTTL    time.Duration
}

func New(log *slog.Logger, usrStorage UserStorage, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:         log,
		usrStorage:  usrStorage,
		AppProvider: appProvider,
		tokenTTL:    tokenTTL,
	}

}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error) {
	const op = "auth.Auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("registering new user")
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := a.usrStorage.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (token string, err error) {
	const op = "auth.Auth.Login"
	log := a.log.With(slog.String("op", op), slog.String("email", email))
	log.Info("logging in")
	usr, err := a.usrStorage.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Warn("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(usr.PassHash, []byte(password)); err != nil {
		a.log.Warn("invalid credentials", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.AppProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err = jwt.NewToken(usr, app, a.tokenTTL)
	if err != nil {
		a.log.Warn("failed to create token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}
