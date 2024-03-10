package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func NewQueryServiceDenomsRequest(pagination *query.PageRequest) *QueryServiceDenomsRequest {
	return &QueryServiceDenomsRequest{
		Pagination: pagination,
	}
}

func (m *QueryServiceDenomsRequest) ValidateBasic() error {
	return nil
}

func NewQueryServiceDenomsByOwnerRequest(owner string) *QueryServiceDenomsByOwnerRequest {
	return &QueryServiceDenomsByOwnerRequest{
		Owner: owner,
	}
}

func (m *QueryServiceDenomsByOwnerRequest) ValidateBasic() error {
	if m.Owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return err
	}

	return nil
}

func NewQueryServiceDenomRequest(id string) *QueryServiceDenomRequest {
	return &QueryServiceDenomRequest{
		Id: id,
	}
}

func (m *QueryServiceDenomRequest) ValidateBasic() error {
	if m.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	return nil
}
