package types

const (
	EventTypeDataVerificationVote = "data_verification"
	EventTypeDataDeliveryVote     = "data_delivery"

	AttributeKeyVoteStatus          = "vote_status"
	AttributeKeyDataHash            = "data_hash"
	AttributeKeyDeliveredCID        = "delivered_cid"
	AttributeValueVoteStatusStarted = "started"
	AttributeValueVoteStatusEnded   = "ended"
)
