package auth

import (
	"context"
	"errors"
	"fmt"
	ssov1 "github.com/alsadx/protos/gen/go/sso"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso/internal/domain/models"
	"sso/internal/services/auth"
	"strings"
)

var validate = validator.New()

type Auth interface {
	Login(ctx context.Context, email string, password string, appId int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userId int64, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	input := models.LoginInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		AppID:    int(req.GetAppId()),
	}

	if err := validate.Struct(input); err != nil {
		msgs := ValidationError(err)
		return nil, status.Error(codes.InvalidArgument, strings.Join(msgs, "; "))
	}

	token, err := s.auth.Login(ctx, input.Email, input.Password, input.AppID)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	input := models.RegisterInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	if err := validate.Struct(input); err != nil {
		msgs := ValidationError(err)
		return nil, status.Error(codes.InvalidArgument, strings.Join(msgs, "; "))
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	userID := req.GetUserId()

	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func ValidationError(err error) []string {
	var errs []string

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			field := strings.ToLower(e.Field()) // Email → email
			switch e.Tag() {
			case "required":
				errs = append(errs, fmt.Sprintf("%s is required", field))
			case "email":
				errs = append(errs, fmt.Sprintf("%s must be a valid email", field))
			case "gt":
				errs = append(errs, fmt.Sprintf("%s must be greater than %s", field, e.Param()))
			default:
				errs = append(errs, fmt.Sprintf("%s is invalid", field))
			}
		}
	} else {
		errs = append(errs, err.Error())
	}

	return errs
}
