package store

import (
	"database/sql"
	"errors"

	"github.com/JhayceeCodes/seigen/internal/model"
)

type PostgresPolicyRepository struct {
	db *sql.DB
}

func NewPostgresPolicyRepository(db *sql.DB) *PostgresPolicyRepository {
	return &PostgresPolicyRepository{
		db: db,
	}
}

func (r *PostgresPolicyRepository) Get(identifier model.Identifier) (model.Policy, error) {

	var (
		storedIdentifier string
		configData       []byte
	)

	row := r.db.QueryRow(
		`
		SELECT identifier, limiter_config
		FROM seigen_policies
		WHERE identifier = $1
		`,
		identifier,
	)

	if err := row.Scan(&storedIdentifier, &configData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Policy{}, ErrPolicyNotFound
		}

		return model.Policy{}, err
	}

	limiterConfig, err := deserializeLimiterConfig(configData)
	if err != nil {
		return model.Policy{}, err
	}

	policy := model.Policy{
		Identifier: model.Identifier(storedIdentifier),
		Limiter:    limiterConfig,
	}

	if err := policy.Validate(); err != nil {
		return model.Policy{}, err
	}

	return policy, nil
}

func (r *PostgresPolicyRepository) Set(policy model.Policy) error {
	if err := policy.Validate(); err != nil {
		return err
	}

	config, err := serializeLimiterConfig(policy.Limiter)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`
		INSERT INTO seigen_policies (identifier, limiter_config)
		VALUES ($1, $2)
		ON CONFLICT (identifier)
		DO UPDATE SET limiter_config = EXCLUDED.limiter_config
		`,
		policy.Identifier,
		config,
	)

	return err
}

func (r *PostgresPolicyRepository) Delete(identifier model.Identifier) error {
	result, err := r.db.Exec(
		`
		DELETE FROM seigen_policies 
		WHERE identifier = $1
		`,
		identifier,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrPolicyNotFound
	}

	return nil
}
