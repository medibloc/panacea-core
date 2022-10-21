package types

const (
	EventTypeRegistrationVote = "oracle_registration"
	EventTypeUpgradeVote      = "oracle_upgrade"
	EventTypeOracleReward     = "oracle_reward"

	AttributeKeyUniqueID            = "unique_id"
	AttributeKeyVoteStatus          = "vote_status"
	AttributeKeyOracleAddress       = "oracle_address"
	AttributeValueVoteStatusStarted = "started"
	AttributeValueVoteStatusEnded   = "ended"
)
