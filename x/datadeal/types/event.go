package types

const (
	EventTypeDataVerificationVote = "data_verification"
	EventTypeDataDeliveryVote     = "data_delivery"

	AttributeKeyVoteStatus          = "vote_status"
	AttributeKeyVerifiableCID       = "verifiable_cid"
	AttributeKeyDeliveredCID        = "delivered_cid"
	AttributeValueVoteStatusStarted = "started"
	AttributeValueVoteStatusEnded   = "ended"
)
