package types

import (
	"fmt"
	"math"

	"cosmossdk.io/core/address"
	upstreamnft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type genesisNFTKey struct {
	classID string
	nftID   string
}

// DefaultGenesis returns an empty combined standard and policy state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		NftState:      upstreamnft.DefaultGenesisState(),
		ClassPolicies: []*ClassPolicy{},
		Lifecycles:    []*LifecycleRecord{},
		Tombstones:    []*BurnTombstone{},
	}
}

// ValidateGenesis validates the complete standard and policy state before any
// store writes. minted_count and standard supply are derived from this input.
func ValidateGenesis(
	data GenesisState,
	addressCodec address.Codec,
	unpacker cdctypes.AnyUnpacker,
) error {
	if data.NftState == nil {
		return fmt.Errorf("nft_state must not be nil")
	}
	if addressCodec == nil {
		return fmt.Errorf("nft genesis validation requires an address codec")
	}
	if unpacker == nil {
		return fmt.Errorf("nft genesis validation requires an Any unpacker")
	}
	for index, class := range data.NftState.Classes {
		if class == nil {
			return fmt.Errorf("standard class at index %d must not be nil", index)
		}
	}
	for entryIndex, entry := range data.NftState.Entries {
		if entry == nil {
			return fmt.Errorf("standard nft entry at index %d must not be nil", entryIndex)
		}
		for nftIndex, token := range entry.Nfts {
			if token == nil {
				return fmt.Errorf(
					"standard nft at entry %d index %d must not be nil",
					entryIndex,
					nftIndex,
				)
			}
		}
	}
	if err := upstreamnft.ValidateGenesis(*data.NftState, addressCodec); err != nil {
		return fmt.Errorf("invalid standard nft genesis: %w", err)
	}

	classes := make(map[string]*upstreamnft.Class, len(data.NftState.Classes))
	classCreators := make(map[string]string, len(data.NftState.Classes))
	for _, class := range data.NftState.Classes {
		if _, exists := classes[class.Id]; exists {
			return fmt.Errorf("duplicate standard class %s", class.Id)
		}
		creator, err := validateGenesisClassID(class.Id, addressCodec)
		if err != nil {
			return fmt.Errorf("invalid standard class %s: %w", class.Id, err)
		}
		if err := ValidateClassMetadata(
			class.Name,
			class.Symbol,
			class.Description,
			class.Uri,
			class.UriHash,
		); err != nil {
			return fmt.Errorf("invalid standard class %s metadata: %w", class.Id, err)
		}
		if class.Data != nil {
			return fmt.Errorf("standard class %s data must be nil", class.Id)
		}
		classes[class.Id] = class
		classCreators[class.Id] = creator
	}

	policies := make(map[string]*ClassPolicy, len(data.ClassPolicies))
	for index, policy := range data.ClassPolicies {
		if policy == nil {
			return fmt.Errorf("class policy at index %d must not be nil", index)
		}
		if _, exists := policies[policy.ClassId]; exists {
			return fmt.Errorf("duplicate class policy %s", policy.ClassId)
		}
		creator, hasClass := classCreators[policy.ClassId]
		if !hasClass {
			return fmt.Errorf("class policy %s references missing standard class", policy.ClassId)
		}
		canonicalCreator, _, err := canonicalGenesisAddress(
			addressCodec,
			"class policy creator",
			policy.Creator,
		)
		if err != nil {
			return fmt.Errorf("invalid class policy %s: %w", policy.ClassId, err)
		}
		if canonicalCreator != policy.Creator {
			return fmt.Errorf("class policy %s creator is not canonical", policy.ClassId)
		}
		if canonicalCreator != creator {
			return fmt.Errorf("class policy %s creator does not match class namespace", policy.ClassId)
		}
		canonicalController, _, err := canonicalGenesisAddress(
			addressCodec,
			"class policy controller",
			policy.Controller,
		)
		if err != nil {
			return fmt.Errorf("invalid class policy %s: %w", policy.ClassId, err)
		}
		if canonicalController != policy.Controller {
			return fmt.Errorf("class policy %s controller is not canonical", policy.ClassId)
		}
		if err := ValidateTransferPolicy(policy.TransferPolicy); err != nil {
			return fmt.Errorf("invalid class policy %s: %w", policy.ClassId, err)
		}
		policies[policy.ClassId] = policy
	}
	for _, class := range data.NftState.Classes {
		if _, exists := policies[class.Id]; !exists {
			return fmt.Errorf("standard class %s has no class policy", class.Id)
		}
	}

	liveNFTs := make(map[genesisNFTKey]string)
	mintedCounts := make(map[string]uint64, len(classes))
	seenOwners := make(map[string]struct{}, len(data.NftState.Entries))
	for _, entry := range data.NftState.Entries {
		canonicalOwner, _, err := canonicalGenesisAddress(addressCodec, "nft owner", entry.Owner)
		if err != nil {
			return err
		}
		if canonicalOwner != entry.Owner {
			return fmt.Errorf("nft owner %s is not canonical", entry.Owner)
		}
		if _, exists := seenOwners[entry.Owner]; exists {
			return fmt.Errorf("duplicate standard nft owner entry %s", entry.Owner)
		}
		seenOwners[entry.Owner] = struct{}{}
		if len(entry.Nfts) == 0 {
			return fmt.Errorf("standard nft owner entry %s must not be empty", entry.Owner)
		}
		for _, token := range entry.Nfts {
			if _, exists := classes[token.ClassId]; !exists {
				return fmt.Errorf(
					"standard nft %s/%s references missing class",
					token.ClassId,
					token.Id,
				)
			}
			if err := ValidateNFTID(token.Id); err != nil {
				return fmt.Errorf("invalid standard nft %s/%s: %w", token.ClassId, token.Id, err)
			}
			if err := ValidateURI(token.Uri, token.UriHash); err != nil {
				return fmt.Errorf(
					"invalid standard nft %s/%s URI metadata: %w",
					token.ClassId,
					token.Id,
					err,
				)
			}
			if _, err := CanonicalizeNFTData(unpacker, token.Data); err != nil {
				return fmt.Errorf(
					"invalid standard nft %s/%s data: %w",
					token.ClassId,
					token.Id,
					err,
				)
			}
			key := genesisNFTKey{classID: token.ClassId, nftID: token.Id}
			if _, exists := liveNFTs[key]; exists {
				return fmt.Errorf("duplicate standard nft %s/%s", token.ClassId, token.Id)
			}
			liveNFTs[key] = entry.Owner
			if err := incrementGenesisCount(mintedCounts, token.ClassId); err != nil {
				return err
			}
		}
	}

	lifecycles := make(map[genesisNFTKey]*LifecycleRecord, len(data.Lifecycles))
	for index, lifecycle := range data.Lifecycles {
		if lifecycle == nil {
			return fmt.Errorf("lifecycle at index %d must not be nil", index)
		}
		key := genesisNFTKey{classID: lifecycle.ClassId, nftID: lifecycle.NftId}
		if _, exists := lifecycles[key]; exists {
			return fmt.Errorf("duplicate lifecycle %s/%s", lifecycle.ClassId, lifecycle.NftId)
		}
		if _, exists := classes[lifecycle.ClassId]; !exists {
			return fmt.Errorf(
				"lifecycle %s/%s references missing class",
				lifecycle.ClassId,
				lifecycle.NftId,
			)
		}
		if err := ValidateNFTID(lifecycle.NftId); err != nil {
			return fmt.Errorf("invalid lifecycle %s/%s: %w", lifecycle.ClassId, lifecycle.NftId, err)
		}
		if err := validateGenesisMintRecord(
			addressCodec,
			"lifecycle",
			lifecycle.ClassId,
			lifecycle.NftId,
			lifecycle.Mint,
		); err != nil {
			return err
		}
		if err := validateGenesisRevocation(
			addressCodec,
			"lifecycle",
			lifecycle.ClassId,
			lifecycle.NftId,
			lifecycle.Mint,
			lifecycle.Revocation,
		); err != nil {
			return err
		}
		if lifecycle.Revocation != nil && !policies[lifecycle.ClassId].Revocable {
			return fmt.Errorf(
				"lifecycle %s/%s is revoked under a non-revocable class policy",
				lifecycle.ClassId,
				lifecycle.NftId,
			)
		}
		lifecycles[key] = lifecycle
	}
	for _, entry := range data.NftState.Entries {
		for _, token := range entry.Nfts {
			key := genesisNFTKey{classID: token.ClassId, nftID: token.Id}
			if _, exists := lifecycles[key]; !exists {
				return fmt.Errorf("standard nft %s/%s has no lifecycle", key.classID, key.nftID)
			}
		}
	}
	for _, lifecycle := range data.Lifecycles {
		key := genesisNFTKey{classID: lifecycle.ClassId, nftID: lifecycle.NftId}
		if _, exists := liveNFTs[key]; !exists {
			return fmt.Errorf("lifecycle %s/%s has no standard nft", key.classID, key.nftID)
		}
	}

	tombstones := make(map[genesisNFTKey]struct{}, len(data.Tombstones))
	for index, tombstone := range data.Tombstones {
		if tombstone == nil {
			return fmt.Errorf("tombstone at index %d must not be nil", index)
		}
		key := genesisNFTKey{classID: tombstone.ClassId, nftID: tombstone.NftId}
		if _, exists := tombstones[key]; exists {
			return fmt.Errorf("duplicate tombstone %s/%s", tombstone.ClassId, tombstone.NftId)
		}
		if _, exists := liveNFTs[key]; exists {
			return fmt.Errorf("nft %s/%s is both live and burned", tombstone.ClassId, tombstone.NftId)
		}
		if _, exists := classes[tombstone.ClassId]; !exists {
			return fmt.Errorf(
				"tombstone %s/%s references missing class",
				tombstone.ClassId,
				tombstone.NftId,
			)
		}
		if err := validateGenesisTombstone(addressCodec, unpacker, tombstone); err != nil {
			return err
		}
		if tombstone.Revocation != nil && !policies[tombstone.ClassId].Revocable {
			return fmt.Errorf(
				"tombstone %s/%s contains revocation under a non-revocable class policy",
				tombstone.ClassId,
				tombstone.NftId,
			)
		}
		tombstones[key] = struct{}{}
		if err := incrementGenesisCount(mintedCounts, tombstone.ClassId); err != nil {
			return err
		}
	}

	for _, policy := range data.ClassPolicies {
		mintedCount := mintedCounts[policy.ClassId]
		if policy.MaxSupply != 0 && mintedCount > policy.MaxSupply {
			return fmt.Errorf(
				"class %s minted count %d exceeds max supply %d",
				policy.ClassId,
				mintedCount,
				policy.MaxSupply,
			)
		}
	}
	return nil
}

func validateGenesisClassID(classID string, addressCodec address.Codec) (string, error) {
	creator, _, err := ParseClassID(classID)
	if err != nil {
		return "", err
	}
	canonicalCreator, _, err := canonicalGenesisAddress(addressCodec, "class creator", creator)
	if err != nil {
		return "", err
	}
	if canonicalCreator != creator {
		return "", fmt.Errorf("class creator must use its canonical address")
	}
	return creator, nil
}

func canonicalGenesisAddress(
	addressCodec address.Codec,
	field string,
	value string,
) (string, []byte, error) {
	addressBytes, err := addressCodec.StringToBytes(value)
	if err != nil {
		return "", nil, fmt.Errorf("invalid %s address: %w", field, err)
	}
	canonical, err := addressCodec.BytesToString(addressBytes)
	if err != nil {
		return "", nil, fmt.Errorf("encode canonical %s address: %w", field, err)
	}
	return canonical, addressBytes, nil
}

func validateGenesisMintRecord(
	addressCodec address.Codec,
	recordType string,
	classID string,
	nftID string,
	mint MintRecord,
) error {
	if mint.MintedAt.IsZero() && mint.MintedBy == "" {
		return fmt.Errorf("%s has no mint record for %s/%s", recordType, classID, nftID)
	}
	if mint.MintedAt.IsZero() {
		return fmt.Errorf("%s has no mint time for %s/%s", recordType, classID, nftID)
	}
	canonicalMinter, _, err := canonicalGenesisAddress(addressCodec, "minted_by", mint.MintedBy)
	if err != nil {
		return fmt.Errorf("%s has invalid minter for %s/%s: %w", recordType, classID, nftID, err)
	}
	if canonicalMinter != mint.MintedBy {
		return fmt.Errorf("%s minter is not canonical for %s/%s", recordType, classID, nftID)
	}
	return nil
}

func validateGenesisRevocation(
	addressCodec address.Codec,
	recordType string,
	classID string,
	nftID string,
	mint MintRecord,
	revocation *Revocation,
) error {
	if revocation == nil {
		return nil
	}
	if revocation.RevokedAt.IsZero() {
		return fmt.Errorf("%s has no revocation time for %s/%s", recordType, classID, nftID)
	}
	if revocation.RevokedAt.Before(mint.MintedAt) {
		return fmt.Errorf("%s revocation predates mint for %s/%s", recordType, classID, nftID)
	}
	canonicalRevoker, _, err := canonicalGenesisAddress(
		addressCodec,
		"revoked_by",
		revocation.RevokedBy,
	)
	if err != nil {
		return fmt.Errorf("%s has invalid revoker for %s/%s: %w", recordType, classID, nftID, err)
	}
	if canonicalRevoker != revocation.RevokedBy {
		return fmt.Errorf("%s revoker is not canonical for %s/%s", recordType, classID, nftID)
	}
	return nil
}

func validateGenesisTombstone(
	addressCodec address.Codec,
	unpacker cdctypes.AnyUnpacker,
	tombstone *BurnTombstone,
) error {
	if err := ValidateNFTID(tombstone.NftId); err != nil {
		return fmt.Errorf("invalid tombstone %s/%s: %w", tombstone.ClassId, tombstone.NftId, err)
	}
	if err := validateGenesisMintRecord(
		addressCodec,
		"tombstone",
		tombstone.ClassId,
		tombstone.NftId,
		tombstone.Mint,
	); err != nil {
		return err
	}
	if err := validateGenesisRevocation(
		addressCodec,
		"tombstone",
		tombstone.ClassId,
		tombstone.NftId,
		tombstone.Mint,
		tombstone.Revocation,
	); err != nil {
		return err
	}
	if err := ValidateURI(tombstone.Uri, tombstone.UriHash); err != nil {
		return fmt.Errorf(
			"tombstone %s/%s has invalid URI metadata: %w",
			tombstone.ClassId,
			tombstone.NftId,
			err,
		)
	}
	if _, err := CanonicalizeNFTData(unpacker, tombstone.Data); err != nil {
		return fmt.Errorf(
			"tombstone %s/%s has invalid data: %w",
			tombstone.ClassId,
			tombstone.NftId,
			err,
		)
	}
	if tombstone.BurnedAt.IsZero() {
		return fmt.Errorf("tombstone has no burn time for %s/%s", tombstone.ClassId, tombstone.NftId)
	}
	if tombstone.BurnedAt.Before(tombstone.Mint.MintedAt) {
		return fmt.Errorf("tombstone burn predates mint for %s/%s", tombstone.ClassId, tombstone.NftId)
	}
	if tombstone.Revocation != nil && tombstone.BurnedAt.Before(tombstone.Revocation.RevokedAt) {
		return fmt.Errorf(
			"tombstone burn predates revocation for %s/%s",
			tombstone.ClassId,
			tombstone.NftId,
		)
	}
	canonicalBurner, _, err := canonicalGenesisAddress(
		addressCodec,
		"burned_by",
		tombstone.BurnedBy,
	)
	if err != nil {
		return fmt.Errorf(
			"tombstone has invalid burner for %s/%s: %w",
			tombstone.ClassId,
			tombstone.NftId,
			err,
		)
	}
	if canonicalBurner != tombstone.BurnedBy {
		return fmt.Errorf(
			"tombstone burner is not canonical for %s/%s",
			tombstone.ClassId,
			tombstone.NftId,
		)
	}
	return nil
}

func incrementGenesisCount(counts map[string]uint64, classID string) error {
	if counts[classID] == math.MaxUint64 {
		return fmt.Errorf("class %s lifetime mint count overflows uint64", classID)
	}
	counts[classID]++
	return nil
}
