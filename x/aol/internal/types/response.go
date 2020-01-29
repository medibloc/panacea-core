package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ResAddRecord struct {
	OwnerAddress sdk.AccAddress `json:"owner_address"`
	TopicName    string         `json:"topic_name"`
	Offset       uint64         `json:"offset"`
}

func NewMsgAddRecordResponse(owner sdk.AccAddress, topic string, offset uint64) ResAddRecord {
	return ResAddRecord{
		OwnerAddress: owner,
		TopicName:    topic,
		Offset:       offset,
	}
}

func (r ResAddRecord) MustMarshalJSON() []byte {
	return ModuleCdc.MustMarshalJSON(r)
}
