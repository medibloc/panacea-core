package types

import (
	"errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

func NewClassFromDenom(cdc codec.BinaryCodec, denom *Denom) (*nft.Class, error) {
	meta, err := codectypes.NewAnyWithValue(
		&DenomMeta{
			Creator: denom.Creator,
			Data:    denom.Data,
		})
	if err != nil {
		return nil, err
	}

	return &nft.Class{
		Id:          denom.Id,
		Name:        denom.Name,
		Symbol:      denom.Symbol,
		Description: denom.Description,
		Uri:         denom.Uri,
		UriHash:     denom.UriHash,
		Data:        meta,
	}, nil
}

func NewDenomFromClass(cdc codec.BinaryCodec, class *nft.Class) (*Denom, error) {
	var meta DenomMeta
	if err := cdc.Unmarshal(class.Data.GetValue(), &meta); err != nil {
		return nil, err
	}

	return &Denom{
		Id:          class.Id,
		Name:        class.Name,
		Symbol:      class.Symbol,
		Description: class.Description,
		Uri:         class.Uri,
		UriHash:     class.UriHash,
		Creator:     meta.Creator,
		Data:        meta.Data,
	}, nil
}

func (d Denom) ValidateBasic() error {
	if d.Id == "" {
		return errors.New("Id cannot be empty.")
	}

	if d.Name == "" {
		return errors.New("Name cannot be empty.")
	}

	if d.Symbol == "" {
		return errors.New("Symbol cannot be empty.")
	}

	if d.Creator == "" {
		return errors.New("Creator cannot be empty.")
	}

	return nil
}
