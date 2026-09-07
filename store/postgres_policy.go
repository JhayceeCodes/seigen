package store

import (
	"database/sql"
	"errors"

	"github.com/JhayceeCodes/seigen/model"
)

type PostgresPolicyRepository struct {
	db *sql.DB
}

type PostgresPolicyGroupRepository struct {
	db *sql.DB
}

type PostgresPolicyGroupMemberRepository struct {
	db *sql.DB
}

// For individual policies
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

// For policy groups
func NewPostgresPolicyGroupRepository(db *sql.DB) *PostgresPolicyGroupRepository {
	return &PostgresPolicyGroupRepository{
		db: db,
	}
}

func (g *PostgresPolicyGroupRepository) Set(group model.PolicyGroup) error {
	if err := group.Validate(); err != nil {
		return err
	}

	config, err := serializeLimiterConfig(group.Limiter)
	if err != nil {
		return err
	}

	_, err = g.db.Exec(
		`
		INSERT INTO seigen_policy_groups (name, limiter_config)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET limiter_config = EXCLUDED.limiter_config
		`,
		group.Name,
		config,
	)

	return err
}

func (g *PostgresPolicyGroupRepository) Get(name string) (model.PolicyGroup, error) {
	row := g.db.QueryRow(
		`
		SELECT name, limiter_config
		FROM seigen_policy_groups
		WHERE name = $1
		`,
		name,
	)

	var (
		group  model.PolicyGroup
		config []byte
	)

	if err := row.Scan(&group.Name, &config); err != nil {
		if err == sql.ErrNoRows {
			return model.PolicyGroup{}, ErrPolicyGroupNotFound
		}

		return model.PolicyGroup{}, err
	}

	limiterConfig, err := deserializeLimiterConfig(config)
	if err != nil {
		return model.PolicyGroup{}, err
	}

	group.Limiter = limiterConfig
	return group, nil
}

func (g *PostgresPolicyGroupRepository) Delete(name string) error {
	_, err := g.db.Exec(
		`
		DELETE FROM seigen_policy_groups
		WHERE name = $1
		`,
		name,
	)

	return err
}

// Policy group member
func NewPostgresPolicyGroupMemberRepository(db *sql.DB) *PostgresPolicyGroupMemberRepository {
	return &PostgresPolicyGroupMemberRepository{
		db: db,
	}
}

func (r *PostgresPolicyGroupMemberRepository) AddMember(
	groupID int64,
	identifier model.Identifier,
) error {
	member := model.PolicyGroupMember{
		GroupID:    groupID,
		Identifier: identifier,
	}

	if err := member.Validate(); err != nil {
		return err
	}

	_, err := r.db.Exec(
		`
		INSERT INTO seigen_policy_group_members (group_id, identifier)
		VALUES ($1, $2)
		`,
		member.GroupID,
		member.Identifier,
	)

	return err
}

func (r *PostgresPolicyGroupMemberRepository) RemoveMember(
	identifier model.Identifier,
) error {
	_, err := r.db.Exec(
		`
		DELETE FROM seigen_policy_group_members
		WHERE identifier = $1
		`,
		identifier,
	)

	return err
}

func (r *PostgresPolicyGroupMemberRepository) GetGroup(
	identifier model.Identifier,
) (model.PolicyGroup, error) {
	row := r.db.QueryRow(
		`
		SELECT g.name, g.limiter_config
		FROM seigen_policy_group_members AS m
		JOIN seigen_policy_groups AS g
			ON g.id = m.group_id
		WHERE m.identifier = $1
		`,
		identifier,
	)

	var (
		group  model.PolicyGroup
		config []byte
	)

	if err := row.Scan(&group.Name, &config); err != nil {
		if err == sql.ErrNoRows {
			return model.PolicyGroup{}, ErrPolicyGroupNotFound
		}

		return model.PolicyGroup{}, err
	}

	limiterConfig, err := deserializeLimiterConfig(config)
	if err != nil {
		return model.PolicyGroup{}, err
	}

	group.Limiter = limiterConfig
	return group, nil
}
