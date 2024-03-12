package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func NewQueryDenomsRequest(pagination *query.PageRequest) *QueryDenomsRequest {
	return &QueryDenomsRequest{
		Pagination: pagination,
	}
}

func (m *QueryDenomsRequest) ValidateBasic() error {
	return nil
}

func NewQueryDenomsByOwnerRequest(owner string) *QueryDenomsByOwnerRequest {
	return &QueryDenomsByOwnerRequest{
		Owner: owner,
	}
}

func (m *QueryDenomsByOwnerRequest) ValidateBasic() error {
	if m.Owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return err
	}

	return nil
}

func NewQueryDenomRequest(id string) *QueryDenomRequest {
	return &QueryDenomRequest{
		Id: id,
	}
}

func (m *QueryDenomRequest) ValidateBasic() error {
	if m.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	return nil
}
