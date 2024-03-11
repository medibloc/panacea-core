package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewQueryServicePNFTsRequest(denomId string) *QueryServicePNFTsRequest {
	return &QueryServicePNFTsRequest{
		DenomId: denomId,
	}
}

func (m *QueryServicePNFTsRequest) ValidateBasic() error {
	if m.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	return nil
}

func NewQueryServicePNFTsByOwnerRequest(denomId string, ownerId string) *QueryServicePNFTsByDenomOwnerRequest {
	return &QueryServicePNFTsByDenomOwnerRequest{
		DenomId: denomId,
		Owner:   ownerId,
	}
}

func (m *QueryServicePNFTsByDenomOwnerRequest) ValidateBasic() error {
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

func NewQueryServicePNFTRequest(denomId, id string) *QueryServicePNFTRequest {
	return &QueryServicePNFTRequest{
		DenomId: denomId,
		Id:      id,
	}
}

func (m *QueryServicePNFTRequest) ValidateBasic() error {
	if m.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	if m.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	return nil
}
