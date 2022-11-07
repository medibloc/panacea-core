package types

const (
	EventTypeRegistrationVote = "oracle_registration_vote"
	EventTypeUpgradeVote      = "oracle_upgrade_vote"
	EventTypeOracleUpgraded   = "oracle_upgraded"
	EventTypeOracleReward     = "oracle_reward"

	AttributeKeyUniqueID            = "unique_id"
	AttributeKeyVoteStatus          = "vote_status"
	AttributeKeyOracleAddress       = "oracle_address"
	AttributeValueVoteStatusStarted = "started"
	AttributeValueVoteStatusEnded   = "ended"
)
