package types

import (
	"bytes"
	"strings"
	"testing"
	"time"

	coreaddress "cosmossdk.io/core/address"
	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/codec/address"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisIsEmptyAndNonNil(t *testing.T) {
	genesis := DefaultGenesis()
	_, cdc := newTestCodec()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.NftState)
	require.Empty(t, genesis.NftState.Classes)
	require.Empty(t, genesis.NftState.Entries)
	require.NotNil(t, genesis.ClassPolicies)
	require.NotNil(t, genesis.Lifecycles)
	require.NotNil(t, genesis.Tombstones)
	require.NoError(t, ValidateGenesis(*genesis, address.NewBech32Codec("panacea"), cdc))
}

func TestValidateGenesisRejectsNilStandardNFTState(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.NftState = nil
	_, cdc := newTestCodec()

	err := ValidateGenesis(*genesis, address.NewBech32Codec("panacea"), cdc)
	require.ErrorContains(t, err, "nft_state must not be nil")
}

func TestValidateGenesisAcceptsConsistentCombinedState(t *testing.T) {
	addressCodec := address.NewBech32Codec("panacea")
	_, cdc := newTestCodec()
	genesis := validCombinedGenesis(t, addressCodec)

	require.NoError(t, ValidateGenesis(*genesis, addressCodec, cdc))
}

func TestValidateGenesisRejectsInconsistentCombinedState(t *testing.T) {
	addressCodec := address.NewBech32Codec("panacea")
	_, cdc := newTestCodec()
	base := validCombinedGenesis(t, addressCodec)

	for _, tc := range []struct {
		name          string
		expectedError string
		mutate        func(genesis *GenesisState)
	}{
		{
			name:          "missing class policy",
			expectedError: "has no class policy",
			mutate:        func(genesis *GenesisState) { genesis.ClassPolicies = nil },
		},
		{
			name:          "creator namespace mismatch",
			expectedError: "creator does not match class namespace",
			mutate: func(genesis *GenesisState) {
				other, err := addressCodec.BytesToString(sdk.AccAddress(bytes.Repeat([]byte{93}, 20)))
				require.NoError(t, err)
				genesis.ClassPolicies[0].Creator = other
			},
		},
		{
			name:          "non-canonical policy creator",
			expectedError: "creator is not canonical",
			mutate: func(genesis *GenesisState) {
				genesis.ClassPolicies[0].Creator = strings.ToUpper(genesis.ClassPolicies[0].Creator)
			},
		},
		{
			name:          "non-canonical owner",
			expectedError: "not canonical",
			mutate: func(genesis *GenesisState) {
				genesis.NftState.Entries[0].Owner = strings.ToUpper(genesis.NftState.Entries[0].Owner)
			},
		},
		{
			name:          "live NFT without lifecycle",
			expectedError: "has no lifecycle",
			mutate:        func(genesis *GenesisState) { genesis.Lifecycles = nil },
		},
		{
			name:          "lifecycle without live NFT",
			expectedError: "has no standard nft",
			mutate: func(genesis *GenesisState) {
				extra := proto.Clone(genesis.Lifecycles[0]).(*LifecycleRecord)
				extra.NftId = "nft-3"
				genesis.Lifecycles = append(genesis.Lifecycles, extra)
			},
		},
		{
			name:          "empty owner entry",
			expectedError: "must not be empty",
			mutate: func(genesis *GenesisState) {
				genesis.NftState.Entries[0].Nfts = nil
				genesis.Lifecycles = nil
			},
		},
		{
			name:          "live and burned overlap",
			expectedError: "is both live and burned",
			mutate: func(genesis *GenesisState) {
				genesis.Tombstones[0].NftId = "nft-1"
			},
		},
		{
			name:          "tombstone references missing class",
			expectedError: "references missing class",
			mutate: func(genesis *GenesisState) {
				genesis.Tombstones[0].ClassId = "missing:nft"
			},
		},
		{
			name:          "tombstone burn predates mint",
			expectedError: "burn predates mint",
			mutate: func(genesis *GenesisState) {
				genesis.Tombstones[0].BurnedAt = genesis.Tombstones[0].Mint.MintedAt.Add(-time.Second)
			},
		},
		{
			name:          "revocation under non-revocable policy",
			expectedError: "non-revocable class policy",
			mutate: func(genesis *GenesisState) {
				genesis.ClassPolicies[0].Revocable = false
				genesis.Tombstones[0].Revocation = &Revocation{
					RevokedAt: genesis.Tombstones[0].Mint.MintedAt.Add(30 * time.Minute),
					RevokedBy: genesis.Tombstones[0].Mint.MintedBy,
				}
			},
		},
		{
			name:          "invalid tombstone data",
			expectedError: "invalid data",
			mutate: func(genesis *GenesisState) {
				genesis.Tombstones[0].Data = &cdctypes.Any{
					TypeUrl: "/panacea.nft.v1.UnknownNFTData",
					Value:   []byte{0x0a, 0x00},
				}
			},
		},
		{
			name:          "max supply exceeded",
			expectedError: "exceeds max supply",
			mutate: func(genesis *GenesisState) {
				genesis.ClassPolicies[0].MaxSupply = 1
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			genesis := proto.Clone(base).(*GenesisState)
			tc.mutate(genesis)

			err := ValidateGenesis(*genesis, addressCodec, cdc)
			require.ErrorContains(t, err, tc.expectedError)
		})
	}
}

func validCombinedGenesis(t *testing.T, addressCodec coreaddress.Codec) *GenesisState {
	t.Helper()
	creator, err := addressCodec.BytesToString(sdk.AccAddress(bytes.Repeat([]byte{91}, 20)))
	require.NoError(t, err)
	owner, err := addressCodec.BytesToString(sdk.AccAddress(bytes.Repeat([]byte{92}, 32)))
	require.NoError(t, err)
	classID := creator + ":certificate"
	mintedAt := time.Date(2026, time.July, 25, 0, 0, 0, 0, time.UTC)
	data, err := cdctypes.NewAnyWithValue(&BasicNFTData{Name: "metadata"})
	require.NoError(t, err)

	return &GenesisState{
		NftState: &upstreamnft.GenesisState{
			Classes: []*upstreamnft.Class{{
				Id:          classID,
				Name:        "Certificate",
				Symbol:      "CERT",
				Description: "Completion certificate",
				Uri:         "https://example.test/class.json",
				UriHash:     "sha256:" + strings.Repeat("a", 64),
			}},
			Entries: []*upstreamnft.Entry{{
				Owner: owner,
				Nfts: []*upstreamnft.NFT{{
					ClassId: classID,
					Id:      "nft-1",
					Uri:     "https://example.test/nft-1.json",
					UriHash: "sha256:" + strings.Repeat("b", 64),
					Data:    data,
				}},
			}},
		},
		ClassPolicies: []*ClassPolicy{{
			ClassId:        classID,
			Creator:        creator,
			Controller:     creator,
			TransferPolicy: TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
			Revocable:      true,
			MaxSupply:      2,
		}},
		Lifecycles: []*LifecycleRecord{{
			ClassId: classID,
			NftId:   "nft-1",
			Mint: &MintRecord{
				MintedAt: mintedAt,
				MintedBy: creator,
			},
		}},
		Tombstones: []*BurnTombstone{{
			ClassId: classID,
			NftId:   "nft-2",
			Mint: &MintRecord{
				MintedAt: mintedAt,
				MintedBy: creator,
			},
			Uri:      "https://example.test/nft-2.json",
			UriHash:  "sha256:" + strings.Repeat("c", 64),
			Data:     data,
			BurnedAt: mintedAt.Add(time.Hour),
			BurnedBy: owner,
		}},
	}
}
