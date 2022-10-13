package types

const (
	EventTypeRegistrationVote = "oracle_registration"
	EventTypeUpgradeVote      = "oracle_upgrade"

	AttributeKeyUniqueID            = "unique_id"
	AttributeKeyVoteStatus          = "vote_status"
	AttributeKeyOracleAddress       = "oracle_address"
	AttributeValueVoteStatusStarted = "started"
	AttributeValueVoteStatusEnded   = "ended"
)
