package types

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func NewDeal(dealID uint64, msg *MsgCreateDeal) *Deal {
	dealAddress := NewDealAddress(dealID)

	return &Deal{
		Id:                      dealID,
		Address:                 dealAddress.String(),
		DataSchema:              msg.DataSchema,
		Budget:                  msg.Budget,
		MaxNumData:              msg.MaxNumData,
		CurNumData:              0,
		ConsumerAddress:         msg.ConsumerAddress,
		AgreementTerms:          msg.AgreementTerms,
		Status:                  DEAL_STATUS_ACTIVE,
		PresentationDefinition:  msg.PresentationDefinition,
		ConsumerServiceEndpoint: msg.ConsumerServiceEndpoint,
	}
}

func NewDealAddress(dealID uint64) sdk.AccAddress {
	dealKey := "deal" + strconv.FormatUint(dealID, 10)
	return authtypes.NewModuleAddress(dealKey)
}

func (m *Deal) IsCompleted() bool {
	return m.Status == DEAL_STATUS_COMPLETED
}

func (m *Deal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.ConsumerAddress); err != nil {
		return sdkerrors.Wrapf(err, "consumer address is invalid. address: %s", m.ConsumerAddress)
	}
	if len(m.DataSchema) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "there is no data schema")
	}
	if m.Id <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "ID should be bigger than 0")
	}
	if m.MaxNumData <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "MaxNumData should be bigger than 0")
	}

	if m.Budget == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is empty")
	}

	if !m.Budget.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "budget is not a valid Coin object")
	}

	if m.CurNumData > m.MaxNumData {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "CurNumData can not be bigger than MaxNumData")
	}

	for _, agreementTerm := range m.AgreementTerms {
		if err := agreementTerm.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "invalid agreement term")
		}
	}

	if len(m.ConsumerServiceEndpoint) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "consumer service endpoint is empty")
	}

	return nil
}

func (m *Deal) GetPricePerData() sdk.Dec {
	totalBudget := m.Budget.Amount.ToDec()
	maxNumData := sdk.NewIntFromUint64(m.MaxNumData).ToDec()
	return totalBudget.Quo(maxNumData).TruncateDec()
}

func (m *Deal) IncreaseCurNumData() {
	m.CurNumData += 1
}

func (t *AgreementTerm) ValidateBasic() error {
	if len(t.Title) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "the title of agreement term shouldn't be empty")
	}
	if len(t.Description) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "the description of agreement term shouldn't be empty")
	}
	return nil
}

func ValidatePD(pdBz []byte) error {
	var pd *presexch.PresentationDefinition

	if err := json.Unmarshal(pdBz, &pd); err != nil {
		return fmt.Errorf("failed to unmarshal presentation definition: %w", err)
	}

	if err := pd.ValidateSchema(); err != nil {
		return fmt.Errorf("invalid presentation definition: %w", err)
	}

	return nil
}
