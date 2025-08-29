package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user models.User, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (app models.App, err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// New returns a new instance of the Auth service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Login checks if user with given credentials exists in the system and returns acess token.
//
// If user exists, but the password is incorrect, returns error.
// If user doesn't exists, returns error.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (token string, err error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to get user", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid password", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("app not found", slog.String("error", err.Error()))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to get app", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully logged in", slog.String("email", email))

	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to create token", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string, password string,
) (userID int64, err error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err = a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("user already exists", slog.String("error", err.Error()))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save new user", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("new user registered")

	return userID, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error) {
	const op = "auth.IsAdmin"

	log := a.log.With(slog.String("op", op), slog.Int64("userID", userID))

	log.Info("checking if user is admin")

	isAdmin, err = a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("app not found", slog.String("error", err.Error()))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		a.log.Error("failed to check if user is admin", slog.String("error", err.Error()))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}
