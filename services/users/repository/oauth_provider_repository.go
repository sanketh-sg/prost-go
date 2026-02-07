package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sanketh-sg/prost/services/users/models"
	"github.com/sanketh-sg/prost/shared/db"
)

// OAuthProviderRepository handles OAuth provider database operations
type OAuthProviderRepository struct {
    conn *db.Connection
}


func NewOAuthProviderRepository(conn *db.Connection) *OAuthProviderRepository {
    return &OAuthProviderRepository{
        conn: conn,
    }
}

func (opr *OAuthProviderRepository) GetByProviderSub(ctx context.Context, provider, providerSub string) (*models.OAuthProvider, error) {
    query := `
        SELECT id, user_id, provider, provider_sub, provider_email, created_at, updated_at
        FROM $schema.oauth_providers
        WHERE provider = $1 AND provider_sub = $2
    `
    query = replaceSchema(query, opr.conn.Schema)

    var oauthProvider models.OAuthProvider

    err := opr.conn.QueryRowContext(ctx, query, provider, providerSub).Scan(
        &oauthProvider.ID,
        &oauthProvider.UserID,
        &oauthProvider.Provider,
        &oauthProvider.ProviderSub,
        &oauthProvider.ProviderEmail,
        &oauthProvider.CreatedAt,
        &oauthProvider.UpdatedAt,
    )

    if err != nil {
        log.Printf("Error getting OAuth provider: %v", err)
        return nil, err
    }

    return &oauthProvider, nil
}

// CreateOAuthProvider creates a new OAuth provider link for a user
func (opr *OAuthProviderRepository) CreateOAuthProvider(ctx context.Context, oauthProvider *models.OAuthProvider) error {
    query := `
        INSERT INTO $schema.oauth_providers (id, user_id, provider, provider_sub, provider_email, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, user_id, provider, provider_sub, provider_email, created_at, updated_at
    `
    query = replaceSchema(query, opr.conn.Schema)

    now := time.Now().UTC()
    oauthProvider.ID = uuid.New().String()
    oauthProvider.CreatedAt = now
    oauthProvider.UpdatedAt = now

    err := opr.conn.QueryRowContext(ctx, query,
        oauthProvider.ID,
        oauthProvider.UserID,
        oauthProvider.Provider,
        oauthProvider.ProviderSub,
        oauthProvider.ProviderEmail,
        now,
        now,
    ).Scan(
        &oauthProvider.ID,
        &oauthProvider.UserID,
        &oauthProvider.Provider,
        &oauthProvider.ProviderSub,
        &oauthProvider.ProviderEmail,
        &oauthProvider.CreatedAt,
        &oauthProvider.UpdatedAt,
    )

    if err != nil {
        log.Printf("Error creating OAuth provider: %v", err)
        return fmt.Errorf("failed to create OAuth provider: %w", err)
    }

    return nil
}

// GetByUserID gets all OAuth providers for a user
func (opr *OAuthProviderRepository) GetByUserID(ctx context.Context, userID string) ([]models.OAuthProvider, error) {
    query := `
        SELECT id, user_id, provider, provider_sub, provider_email, created_at, updated_at
        FROM $schema.oauth_providers
        WHERE user_id = $1
    `
    query = replaceSchema(query, opr.conn.Schema)

    rows, err := opr.conn.QueryContext(ctx, query, userID)
    if err != nil {
        log.Printf("Error getting OAuth providers: %v", err)
        return nil, fmt.Errorf("failed to get OAuth providers: %w", err)
    }
    defer rows.Close()

    var providers []models.OAuthProvider
    for rows.Next() {
        var provider models.OAuthProvider
        err := rows.Scan(
            &provider.ID,
            &provider.UserID,
            &provider.Provider,
            &provider.ProviderSub,
            &provider.ProviderEmail,
            &provider.CreatedAt,
            &provider.UpdatedAt,
        )
        if err != nil {
            log.Printf("Error scanning OAuth provider row: %v", err)
            return nil, fmt.Errorf("failed to scan OAuth provider: %w", err)
        }
        providers = append(providers, provider)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error iterating OAuth providers: %v", err)
        return nil, fmt.Errorf("failed to iterate OAuth providers: %w", err)
    }

    return providers, nil
}