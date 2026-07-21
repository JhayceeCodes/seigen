package model

type APIKey struct {
	Key  string
	Tier string
}

const (
	TierFree    = "free"
	TierPremium = "premium"
)
