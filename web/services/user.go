package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	db "oauth/database/generated"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markbates/goth"
	"github.com/rs/zerolog"
)

type UserService struct {
	pool    *pgxpool.Pool
	logger  *zerolog.Logger
	queries *db.Queries
}

func NewUserService(pool *pgxpool.Pool, l *zerolog.Logger) *UserService {
	return &UserService{
		pool:    pool,
		logger:  l,
		queries: db.New(pool),
	}
}

type CreateUserWithOAuthRequest struct {
	Email     pgtype.Text
	FullName  pgtype.Text
	AvatarURL pgtype.Text
	Provider  string
	IDToken   []byte
	RawData   []byte
}

type UserWithOAuth struct {
	User          db.User
	OAuthIdentity db.OauthIdentity
}

func (s *UserService) CreateUserWithOAuth(ctx context.Context, req CreateUserWithOAuthRequest) (*UserWithOAuth, error) {
	var result UserWithOAuth

	// Start transaction
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		// Create queries instance with transaction
		qtx := s.queries.WithTx(tx)

		// Check if user already exists
		existingUser, err := qtx.GetUserByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		var user db.User
		if errors.Is(err, pgx.ErrNoRows) {
			// Create new user
			user, err = qtx.CreateUser(ctx, db.CreateUserParams{
				Email:     req.Email,
				FullName:  req.FullName,
				AvatarUrl: req.AvatarURL,
			})
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			user = existingUser
		}

		// Check if OAuth identity already exists
		_, err = qtx.GetOAuthIdentityByProviderAndToken(ctx, db.GetOAuthIdentityByProviderAndTokenParams{
			Provider: req.Provider,
			IDToken:  req.IDToken,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to check existing OAuth identity: %w", err)
		}

		var oauthIdentity db.OauthIdentity
		s.logger.Info().Interface("id_token", req.IDToken).Str("email", req.Email.String).Msg("Inserting oauth identity")
		if errors.Is(err, pgx.ErrNoRows) {
			// Create OAuth identity
			oauthIdentity, err = qtx.CreateOAuthIdentity(ctx, db.CreateOAuthIdentityParams{
				UserID:   user.ID,
				Provider: req.Provider,
				IDToken:  req.IDToken,
				Email:    req.Email.String,
				RawData:  req.RawData,
			})
			if err != nil {
				return fmt.Errorf("failed to create OAuth identity: %w", err)
			}
		}

		result.User = user
		result.OAuthIdentity = oauthIdentity
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateUserFromGothUser
// Helper function to convert goth.User to CreateUserWithOAuthRequest
func (s *UserService) CreateUserFromGothUser(gothUser goth.User) CreateUserWithOAuthRequest {
	email := pgtype.Text{
		String: gothUser.Email,
		Valid:  true,
	}
	fullName := pgtype.Text{
		String: gothUser.Name,
		Valid:  true,
	}
	avatarURL := pgtype.Text{
		String: gothUser.AvatarURL,
		Valid:  true,
	}
	// Use JSON encoding instead of gob for JSONB compatibility
	rawData, _ := mapToJSONBytes(gothUser.RawData)

	return CreateUserWithOAuthRequest{
		Email:     email,
		FullName:  fullName,
		AvatarURL: avatarURL,
		Provider:  gothUser.Provider,
		IDToken:   []byte(gothUser.IDToken),
		RawData:   rawData,
	}
}

// mapToJSONBytes converts map to JSON bytes instead of gob
func mapToJSONBytes(m map[string]interface{}) ([]byte, error) {
	return json.Marshal(m)
}
