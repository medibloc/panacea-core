package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	lifecycleLocalClassID = "lifecycle.class"
	lifecycleNFTID        = "certificate.1"
	lifecycleClassURI     = "https://example.test/classes/lifecycle.json"
	lifecycleNFTURI       = "https://example.test/nfts/certificate.1.json"
	lifecycleDataJSON     = `{"@type":"/panacea.nft.v1.BasicNFTData","name":"Certificate #1","description":"Real-node lifecycle certificate","image_uri":"https://example.test/images/certificate.1.png"}`
)

var (
	lifecycleClassURIHash = "sha256:" + strings.Repeat("a", 64)
	lifecycleNFTURIHash   = "sha256:" + strings.Repeat("b", 64)
)

type classRecordQueryResponse struct {
	ClassRecord struct {
		Class struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Symbol      string `json:"symbol"`
			Description string `json:"description"`
			URI         string `json:"uri"`
			URIHash     string `json:"uri_hash"`
		} `json:"class"`
		Policy struct {
			ClassID        string `json:"class_id"`
			Creator        string `json:"creator"`
			Controller     string `json:"controller"`
			TransferPolicy string `json:"transfer_policy"`
			Revocable      bool   `json:"revocable"`
			MaxSupply      string `json:"max_supply"`
		} `json:"policy"`
		MintedCount string `json:"minted_count"`
	} `json:"class_record"`
}

type provenanceRecord struct {
	MintedAt string `json:"minted_at"`
	MintedBy string `json:"minted_by"`
}

type revocationRecord struct {
	RevokedAt string `json:"revoked_at"`
	RevokedBy string `json:"revoked_by"`
}

type nftRecordQueryResponse struct {
	NFTRecord struct {
		Live *struct {
			NFT struct {
				ClassID string         `json:"class_id"`
				ID      string         `json:"id"`
				URI     string         `json:"uri"`
				URIHash string         `json:"uri_hash"`
				Data    map[string]any `json:"data"`
			} `json:"nft"`
			Owner      string            `json:"owner"`
			Status     string            `json:"status"`
			Mint       provenanceRecord  `json:"mint"`
			Revocation *revocationRecord `json:"revocation"`
		} `json:"live"`
		BurnTombstone *struct {
			ClassID    string            `json:"class_id"`
			NFTID      string            `json:"nft_id"`
			Mint       provenanceRecord  `json:"mint"`
			URI        string            `json:"uri"`
			URIHash    string            `json:"uri_hash"`
			Data       map[string]any    `json:"data"`
			Revocation *revocationRecord `json:"revocation"`
			BurnedAt   string            `json:"burned_at"`
			BurnedBy   string            `json:"burned_by"`
		} `json:"burn_tombstone"`
	} `json:"nft_record"`
}

type nftLifecycleEvidence struct {
	ClassID    string                       `json:"class_id"`
	NFTID      string                       `json:"nft_id"`
	Class      classRecordQueryResponse     `json:"class"`
	Tombstone  nftRecordQueryResponse       `json:"tombstone"`
	Pagination nftPaginationRestartEvidence `json:"pagination"`
}

func runNFTLifecycle(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	expectedPreexistingClassIDs ...string,
) nftLifecycleEvidence {
	t.Helper()

	creator := buildAndFundNFTWallet(t, ctx, network, "nft-creator")
	controller := buildAndFundNFTWallet(t, ctx, network, "nft-controller")
	firstOwner := buildAndFundNFTWallet(t, ctx, network, "nft-owner-first")
	finalOwner := buildAndFundNFTWallet(t, ctx, network, "nft-owner-final")
	txNode := network.Chain.Validators[0]
	httpClient := &http.Client{Timeout: 5 * time.Second}

	classID := creator.FormattedAddress() + ":" + lifecycleLocalClassID
	createTx, err := network.BroadcastAndWaitTx(
		ctx,
		"nft-create-class",
		txNode,
		creator.KeyName(),
		"nft", "create-class",
		lifecycleLocalClassID,
		"Lifecycle Certificate",
		"LIFE",
		"owner-transferable",
		"true",
		"10",
		"--description", "Real-node lifecycle class",
		"--uri", lifecycleClassURI,
		"--uri-hash", lifecycleClassURIHash,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, createTx, "panacea.nft.v1.EventClassCreated", map[string]string{
		"class_id": classID,
		"creator":  creator.FormattedAddress(),
	})

	createdClass := assertClassQueryBoundaryParityAtTx(t, ctx, network, "class-after-create", createTx, classID)
	require.Equal(t, classID, createdClass.ClassRecord.Class.ID)
	require.Equal(t, "Lifecycle Certificate", createdClass.ClassRecord.Class.Name)
	require.Equal(t, "LIFE", createdClass.ClassRecord.Class.Symbol)
	require.Equal(t, lifecycleClassURI, createdClass.ClassRecord.Class.URI)
	require.Equal(t, lifecycleClassURIHash, createdClass.ClassRecord.Class.URIHash)
	require.Equal(t, classID, createdClass.ClassRecord.Policy.ClassID)
	require.Equal(t, creator.FormattedAddress(), createdClass.ClassRecord.Policy.Creator)
	require.Equal(t, creator.FormattedAddress(), createdClass.ClassRecord.Policy.Controller)
	require.Equal(t, "TRANSFER_POLICY_OWNER_TRANSFERABLE", createdClass.ClassRecord.Policy.TransferPolicy)
	require.True(t, createdClass.ClassRecord.Policy.Revocable)
	require.Equal(t, "10", createdClass.ClassRecord.Policy.MaxSupply)
	require.Equal(t, "0", createdClass.ClassRecord.MintedCount)

	standardClass, err := network.QueryNFTClassGRPC(ctx, "standard-class-after-create", classID)
	require.NoError(t, err)
	require.NotNil(t, standardClass.Class)
	require.Equal(t, classID, standardClass.Class.Id)
	standardSupply, err := network.QueryNFTSupplyGRPC(ctx, "standard-supply-after-create", classID)
	require.NoError(t, err)
	require.Zero(t, standardSupply.Amount)
	assertStandardNFTAccountingAtTx(t, ctx, network, "accounting-after-create", createTx, classID, 0, nil)
	assertRawAndEncodedRESTEqual(
		t,
		ctx,
		network,
		httpClient,
		"class-after-create",
		"/panacea/nft/v1/classes/"+classID,
		"/panacea/nft/v1/classes/"+percentEncodeClassDelimiter(classID),
	)

	updateTx, err := network.BroadcastAndWaitTx(
		ctx,
		"nft-update-controller",
		txNode,
		creator.KeyName(),
		"nft", "update-controller", classID, controller.FormattedAddress(),
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, updateTx, "panacea.nft.v1.EventControllerUpdated", map[string]string{
		"class_id":       classID,
		"old_controller": creator.FormattedAddress(),
		"new_controller": controller.FormattedAddress(),
	})
	updatedClass := assertClassQueryBoundaryParityAtTx(t, ctx, network, "class-after-controller-update", updateTx, classID)
	require.Equal(t, createdClass.ClassRecord.Class, updatedClass.ClassRecord.Class)
	require.Equal(t, creator.FormattedAddress(), updatedClass.ClassRecord.Policy.Creator)
	require.Equal(t, controller.FormattedAddress(), updatedClass.ClassRecord.Policy.Controller)
	require.Equal(t, classID, updatedClass.ClassRecord.Class.ID)
	require.Equal(t, createdClass.ClassRecord.Policy.TransferPolicy, updatedClass.ClassRecord.Policy.TransferPolicy)
	require.Equal(t, createdClass.ClassRecord.Policy.Revocable, updatedClass.ClassRecord.Policy.Revocable)
	require.Equal(t, createdClass.ClassRecord.Policy.MaxSupply, updatedClass.ClassRecord.Policy.MaxSupply)
	require.Equal(t, "0", updatedClass.ClassRecord.MintedCount)

	mintTx, err := network.BroadcastAndWaitTx(
		ctx,
		"nft-mint",
		txNode,
		controller.KeyName(),
		"nft", "mint", classID, lifecycleNFTID, firstOwner.FormattedAddress(),
		"--uri", lifecycleNFTURI,
		"--uri-hash", lifecycleNFTURIHash,
		"--data", lifecycleDataJSON,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, mintTx, "cosmos.nft.v1beta1.EventMint", map[string]string{
		"class_id": classID,
		"id":       lifecycleNFTID,
		"owner":    firstOwner.FormattedAddress(),
	})

	minted := assertNFTQueryBoundaryParityAtTx(t, ctx, network, "nft-after-mint", mintTx, classID, lifecycleNFTID)
	require.NotNil(t, minted.NFTRecord.Live)
	require.Nil(t, minted.NFTRecord.BurnTombstone)
	require.Equal(t, firstOwner.FormattedAddress(), minted.NFTRecord.Live.Owner)
	require.Equal(t, "LIVE_NFT_STATUS_ACTIVE", minted.NFTRecord.Live.Status)
	require.Equal(t, controller.FormattedAddress(), minted.NFTRecord.Live.Mint.MintedBy)
	require.NotEmpty(t, minted.NFTRecord.Live.Mint.MintedAt)
	require.Nil(t, minted.NFTRecord.Live.Revocation)
	require.Equal(t, lifecycleNFTURI, minted.NFTRecord.Live.NFT.URI)
	require.Equal(t, lifecycleNFTURIHash, minted.NFTRecord.Live.NFT.URIHash)
	require.Equal(t, "/panacea.nft.v1.BasicNFTData", minted.NFTRecord.Live.NFT.Data["@type"])
	require.Equal(t, "Certificate #1", minted.NFTRecord.Live.NFT.Data["name"])
	mintProvenance := minted.NFTRecord.Live.Mint
	mintedNFT := minted.NFTRecord.Live.NFT
	classAfterMint := assertClassQueryBoundaryParityAtTx(t, ctx, network, "class-after-mint", mintTx, classID)
	require.Equal(t, "1", classAfterMint.ClassRecord.MintedCount)

	assertStandardNFTState(t, ctx, network, "after-mint", classID, lifecycleNFTID, firstOwner.FormattedAddress(), 1, 1)
	assertStandardNFTAccountingAtTx(t, ctx, network, "accounting-after-mint", mintTx, classID, 1, map[string]uint64{
		firstOwner.FormattedAddress(): 1,
	})
	assertRawAndEncodedRESTEqual(
		t,
		ctx,
		network,
		httpClient,
		"nft-after-mint",
		"/panacea/nft/v1/nfts/"+classID+"/"+lifecycleNFTID,
		"/panacea/nft/v1/nfts/"+percentEncodeClassDelimiter(classID)+"/"+lifecycleNFTID,
	)

	sendTx, err := network.BroadcastAndWaitTx(
		ctx,
		"nft-send",
		txNode,
		firstOwner.KeyName(),
		"nft", "send", classID, lifecycleNFTID, finalOwner.FormattedAddress(),
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, sendTx, "cosmos.nft.v1beta1.EventSend", map[string]string{
		"class_id": classID,
		"id":       lifecycleNFTID,
		"sender":   firstOwner.FormattedAddress(),
		"receiver": finalOwner.FormattedAddress(),
	})
	sent := assertNFTQueryBoundaryParityAtTx(t, ctx, network, "nft-after-send", sendTx, classID, lifecycleNFTID)
	require.NotNil(t, sent.NFTRecord.Live)
	require.Equal(t, finalOwner.FormattedAddress(), sent.NFTRecord.Live.Owner)
	require.Equal(t, "LIVE_NFT_STATUS_ACTIVE", sent.NFTRecord.Live.Status)
	require.Equal(t, mintProvenance, sent.NFTRecord.Live.Mint)
	require.Equal(t, mintedNFT, sent.NFTRecord.Live.NFT)
	require.Nil(t, sent.NFTRecord.Live.Revocation)
	assertStandardNFTState(t, ctx, network, "after-send", classID, lifecycleNFTID, finalOwner.FormattedAddress(), 1, 1)
	firstOwnerBalance, err := network.QueryNFTBalanceGRPC(ctx, "first-owner-balance-after-send", classID, firstOwner.FormattedAddress())
	require.NoError(t, err)
	require.Zero(t, firstOwnerBalance.Amount)
	assertStandardNFTAccountingAtTx(t, ctx, network, "accounting-after-send", sendTx, classID, 1, map[string]uint64{
		firstOwner.FormattedAddress(): 0,
		finalOwner.FormattedAddress(): 1,
	})

	revokeTx, err := network.BroadcastAndWaitTx(
		ctx,
		"nft-revoke",
		txNode,
		controller.KeyName(),
		"nft", "revoke", classID, lifecycleNFTID,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, revokeTx, "panacea.nft.v1.EventNFTRevoked", map[string]string{
		"class_id":   classID,
		"nft_id":     lifecycleNFTID,
		"controller": controller.FormattedAddress(),
	})
	revoked := assertNFTQueryBoundaryParityAtTx(t, ctx, network, "nft-after-revoke", revokeTx, classID, lifecycleNFTID)
	require.NotNil(t, revoked.NFTRecord.Live)
	require.Equal(t, finalOwner.FormattedAddress(), revoked.NFTRecord.Live.Owner)
	require.Equal(t, "LIVE_NFT_STATUS_REVOKED", revoked.NFTRecord.Live.Status)
	require.Equal(t, mintProvenance, revoked.NFTRecord.Live.Mint)
	require.Equal(t, mintedNFT, revoked.NFTRecord.Live.NFT)
	require.NotNil(t, revoked.NFTRecord.Live.Revocation)
	require.Equal(t, controller.FormattedAddress(), revoked.NFTRecord.Live.Revocation.RevokedBy)
	require.NotEmpty(t, revoked.NFTRecord.Live.Revocation.RevokedAt)
	revocation := *revoked.NFTRecord.Live.Revocation
	assertStandardNFTState(t, ctx, network, "after-revoke", classID, lifecycleNFTID, finalOwner.FormattedAddress(), 1, 1)
	assertStandardNFTAccountingAtTx(t, ctx, network, "accounting-after-revoke", revokeTx, classID, 1, map[string]uint64{
		firstOwner.FormattedAddress(): 0,
		finalOwner.FormattedAddress(): 1,
	})

	burnTx, err := network.BroadcastAndWaitTx(
		ctx,
		"nft-burn",
		txNode,
		finalOwner.KeyName(),
		"nft", "burn", classID, lifecycleNFTID,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, burnTx, "cosmos.nft.v1beta1.EventBurn", map[string]string{
		"class_id": classID,
		"id":       lifecycleNFTID,
		"owner":    finalOwner.FormattedAddress(),
	})
	burned := assertNFTQueryBoundaryParityAtTx(t, ctx, network, "nft-after-burn", burnTx, classID, lifecycleNFTID)
	require.Nil(t, burned.NFTRecord.Live)
	require.NotNil(t, burned.NFTRecord.BurnTombstone)
	require.Equal(t, classID, burned.NFTRecord.BurnTombstone.ClassID)
	require.Equal(t, lifecycleNFTID, burned.NFTRecord.BurnTombstone.NFTID)
	require.Equal(t, mintProvenance, burned.NFTRecord.BurnTombstone.Mint)
	require.Equal(t, revocation, *burned.NFTRecord.BurnTombstone.Revocation)
	require.Equal(t, lifecycleNFTURI, burned.NFTRecord.BurnTombstone.URI)
	require.Equal(t, lifecycleNFTURIHash, burned.NFTRecord.BurnTombstone.URIHash)
	require.Equal(t, mintedNFT.Data, burned.NFTRecord.BurnTombstone.Data)
	require.Equal(t, finalOwner.FormattedAddress(), burned.NFTRecord.BurnTombstone.BurnedBy)
	require.NotEmpty(t, burned.NFTRecord.BurnTombstone.BurnedAt)

	ownerAfterBurn, err := network.QueryNFTOwnerGRPC(ctx, "standard-owner-after-burn", classID, lifecycleNFTID)
	require.NoError(t, err)
	require.Empty(t, ownerAfterBurn.Owner)
	supplyAfterBurn, err := network.QueryNFTSupplyGRPC(ctx, "standard-supply-after-burn", classID)
	require.NoError(t, err)
	require.Zero(t, supplyAfterBurn.Amount)
	finalOwnerBalance, err := network.QueryNFTBalanceGRPC(ctx, "final-owner-balance-after-burn", classID, finalOwner.FormattedAddress())
	require.NoError(t, err)
	require.Zero(t, finalOwnerBalance.Amount)
	firstOwnerBalanceAfterBurn, err := network.QueryNFTBalanceGRPC(ctx, "first-owner-balance-after-burn", classID, firstOwner.FormattedAddress())
	require.NoError(t, err)
	require.Zero(t, firstOwnerBalanceAfterBurn.Amount)
	classAfterBurn := assertClassQueryBoundaryParityAtTx(t, ctx, network, "class-after-burn", burnTx, classID)
	require.Equal(t, "1", classAfterBurn.ClassRecord.MintedCount)
	assertStandardNFTAccountingAtTx(t, ctx, network, "accounting-after-burn", burnTx, classID, 0, map[string]uint64{
		firstOwner.FormattedAddress(): 0,
		finalOwner.FormattedAddress(): 0,
	})
	assertRawAndEncodedRESTEqual(
		t,
		ctx,
		network,
		httpClient,
		"tombstone-after-burn",
		"/panacea/nft/v1/nfts/"+classID+"/"+lifecycleNFTID,
		"/panacea/nft/v1/nfts/"+percentEncodeClassDelimiter(classID)+"/"+lifecycleNFTID,
	)
	_, err = network.FullNodeRESTGet(
		ctx,
		httpClient,
		"standard-nft-missing-after-burn",
		"/cosmos/nft/v1beta1/nfts/"+percentEncodeClassDelimiter(classID)+"/"+lifecycleNFTID,
	)
	require.ErrorContains(t, err, "HTTP 404")

	remintTx, err := network.BroadcastAndWaitTxExpectDeliverFailure(
		ctx,
		"nft-remint-burned-id",
		txNode,
		controller.KeyName(),
		"panacea_nft",
		6,
		"nft", "mint", classID, lifecycleNFTID, firstOwner.FormattedAddress(),
		"--uri", lifecycleNFTURI,
		"--uri-hash", lifecycleNFTURIHash,
		"--data", lifecycleDataJSON,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	require.Contains(t, remintTx.RawLog, "permanently used")
	_, hasMintEvent := remintTx.FindEvent("cosmos.nft.v1beta1.EventMint")
	require.False(t, hasMintEvent)

	burnedAfterRemint := assertNFTQueryBoundaryParityAtTx(t, ctx, network, "tombstone-after-remint-rejection", remintTx, classID, lifecycleNFTID)
	require.Equal(t, burned.NFTRecord.BurnTombstone, burnedAfterRemint.NFTRecord.BurnTombstone)
	classAfterRemint := assertClassQueryBoundaryParityAtTx(t, ctx, network, "class-after-remint-rejection", remintTx, classID)
	require.Equal(t, "1", classAfterRemint.ClassRecord.MintedCount)
	supplyAfterRemint, err := network.QueryNFTSupplyGRPC(ctx, "standard-supply-after-remint-rejection", classID)
	require.NoError(t, err)
	require.Zero(t, supplyAfterRemint.Amount)
	firstOwnerBalanceAfterRemint, err := network.QueryNFTBalanceGRPC(ctx, "first-owner-balance-after-remint-rejection", classID, firstOwner.FormattedAddress())
	require.NoError(t, err)
	require.Zero(t, firstOwnerBalanceAfterRemint.Amount)
	finalOwnerBalanceAfterRemint, err := network.QueryNFTBalanceGRPC(ctx, "final-owner-balance-after-remint-rejection", classID, finalOwner.FormattedAddress())
	require.NoError(t, err)
	require.Zero(t, finalOwnerBalanceAfterRemint.Amount)
	assertStandardNFTAccountingAtTx(t, ctx, network, "accounting-after-remint-rejection", remintTx, classID, 0, map[string]uint64{
		firstOwner.FormattedAddress(): 0,
		finalOwner.FormattedAddress(): 0,
	})

	pagination := runNFTPaginationCompatibility(
		t,
		ctx,
		network,
		creator,
		controller,
		firstOwner,
		expectedPreexistingClassIDs,
	)
	finalClass := assertClassQueryBoundaryParityAtHeight(
		t, ctx, network, "class-restart-checkpoint", pagination.QueryHeight, classID,
	)
	finalTombstone := assertNFTQueryBoundaryParityAtHeight(
		t, ctx, network, "tombstone-restart-checkpoint", pagination.QueryHeight, classID, lifecycleNFTID,
	)
	return nftLifecycleEvidence{
		ClassID:    classID,
		NFTID:      lifecycleNFTID,
		Class:      finalClass,
		Tombstone:  finalTombstone,
		Pagination: pagination,
	}
}

func buildAndFundNFTWallet(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	keyName string,
) ibc.Wallet {
	t.Helper()
	wallet, err := network.BuildWallet(ctx, keyName, "")
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx,
		"fund-"+keyName,
		network.Chain.Validators[0],
		interchaintest.FaucetAccountKeyName,
		"bank", "send",
		interchaintest.FaucetAccountKeyName,
		wallet.FormattedAddress(),
		sdkmath.NewInt(20_000_000).String()+"umed",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	return wallet
}

func queryNFTRecord(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	classID string,
	nftID string,
) nftRecordQueryResponse {
	t.Helper()
	raw, err := network.FullNodeCLIQuery(ctx, step, "nft", "nft-record", classID, nftID)
	require.NoError(t, err)
	var response nftRecordQueryResponse
	require.NoError(t, json.Unmarshal(raw, &response))
	return response
}

func assertStandardNFTState(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	classID string,
	nftID string,
	owner string,
	wantSupply uint64,
	wantBalance uint64,
) {
	t.Helper()
	nftResponse, err := network.QueryNFTGRPC(ctx, "standard-nft-"+step, classID, nftID)
	require.NoError(t, err)
	require.NotNil(t, nftResponse.Nft)
	require.Equal(t, classID, nftResponse.Nft.ClassId)
	require.Equal(t, nftID, nftResponse.Nft.Id)
	require.Equal(t, lifecycleNFTURI, nftResponse.Nft.Uri)
	require.Equal(t, lifecycleNFTURIHash, nftResponse.Nft.UriHash)
	ownerResponse, err := network.QueryNFTOwnerGRPC(ctx, "standard-owner-"+step, classID, nftID)
	require.NoError(t, err)
	require.Equal(t, owner, ownerResponse.Owner)
	supplyResponse, err := network.QueryNFTSupplyGRPC(ctx, "standard-supply-"+step, classID)
	require.NoError(t, err)
	require.Equal(t, wantSupply, supplyResponse.Amount)
	balanceResponse, err := network.QueryNFTBalanceGRPC(ctx, "standard-balance-"+step, classID, owner)
	require.NoError(t, err)
	require.Equal(t, wantBalance, balanceResponse.Amount)
}

func assertTxEvent(t *testing.T, result *harness.TxResult, eventType string, attributes map[string]string) {
	t.Helper()
	require.NotNil(t, result)
	matching := make([]harness.TxEvent, 0, 1)
	for _, event := range result.Events {
		if event.Type == eventType {
			matching = append(matching, event)
		}
	}
	require.Len(t, matching, 1, "expected exactly one %s event in tx %s", eventType, result.TxHash)
	event := matching[0]
	for key, want := range attributes {
		require.Equal(t, want, event.Attribute(key), "%s attribute %s", eventType, key)
	}
}

func assertRawAndEncodedRESTEqual(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	client *http.Client,
	step string,
	rawPath string,
	encodedPath string,
) {
	t.Helper()
	rawResponse, err := network.FullNodeRESTGet(ctx, client, step+"-raw-rest", rawPath)
	require.NoError(t, err)
	encodedResponse, err := network.FullNodeRESTGet(ctx, client, step+"-encoded-rest", encodedPath)
	require.NoError(t, err)
	require.JSONEq(t, string(rawResponse), string(encodedResponse))
}

func percentEncodeClassDelimiter(classID string) string {
	return strings.Replace(classID, ":", "%3A", 1)
}
