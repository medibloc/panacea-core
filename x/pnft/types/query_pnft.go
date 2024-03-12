package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewQueryPNFTsRequest(denomId string) *QueryPNFTsRequest {
	return &QueryPNFTsRequest{
		DenomId: denomId,
	}
}

func (m *QueryPNFTsRequest) ValidateBasic() error {
	if m.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	return nil
}

func NewQueryPNFTsByOwnerRequest(denomId string, ownerId string) *QueryPNFTsByDenomOwnerRequest {
	return &QueryPNFTsByDenomOwnerRequest{
		DenomId: denomId,
		Owner:   ownerId,
	}
}

func (m *QueryPNFTsByDenomOwnerRequest) ValidateBasic() error {
	if m.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	if m.Owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return err
	}

	return nil
}

func NewQueryPNFTRequest(denomId, id string) *QueryPNFTRequest {
	return &QueryPNFTRequest{
		DenomId: denomId,
		Id:      id,
	}
}

func (m *QueryPNFTRequest) ValidateBasic() error {
	if m.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	if m.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	return nil
}
