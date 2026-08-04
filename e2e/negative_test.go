package e2e_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestNFTNegativePagination(t *testing.T) {
	if os.Getenv("PANACEA_E2E") != "1" {
		t.Skip("set PANACEA_E2E=1 or use ./scripts/e2e/run.sh negative")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 1,
		NumFullNodes:  1,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()

	creator := buildAndFundPaginationWallet(t, ctx, network, "negative-pagination-creator")
	owner := buildAndFundPaginationWallet(t, ctx, network, "negative-pagination-owner")
	classID := creator.FormattedAddress() + ":negative.pagination"
	node := network.Chain.Validators[0]
	_, err = network.BroadcastAndWaitTx(
		ctx,
		"negative-pagination-create",
		node,
		creator.KeyName(),
		"nft", "create-class",
		"negative.pagination", "Negative Pagination", "PAGE", "owner-transferable", "true", "2",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx,
		"negative-pagination-mint",
		node,
		creator.KeyName(),
		"nft", "mint", classID, "page.1", owner.FormattedAddress(),
		"--data", `{"@type":"/panacea.nft.v1.BasicNFTData","name":"Pagination fixture"}`,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)

	snapshotRequest := harness.NFTStateSnapshotRequest{
		ClassID: classID,
		NFTIDs:  []string{"page.1"},
		Owners: []string{
			creator.FormattedAddress(),
			owner.FormattedAddress(),
		},
	}
	before, err := network.SnapshotNFTState(ctx, "negative-pagination-before", snapshotRequest)
	require.NoError(t, err)
	t.Run("no-op controller update is committed and atomic", func(t *testing.T) {
		result, txErr := network.BroadcastAndWaitTxExpectDeliverFailure(
			ctx,
			"negative-no-op-controller-update",
			node,
			creator.KeyName(),
			"sdk",
			18,
			"nft", "update-controller", classID, creator.FormattedAddress(),
			"--gas", "500000",
			"--broadcast-mode", "sync",
		)
		require.NoError(t, txErr)
		require.NotNil(t, result)
		require.Positive(t, result.HeightInt64(), "DeliverTx failure must be committed")
		require.Contains(t, result.RawLog, "new controller must differ from current controller")
		for _, eventType := range []string{
			"panacea.nft.v1.EventControllerUpdated",
			"panacea.nft.v1.EventClassCreated",
			"cosmos.nft.v1beta1.EventMint",
			"cosmos.nft.v1beta1.EventSend",
			"panacea.nft.v1.EventNFTRevoked",
			"cosmos.nft.v1beta1.EventBurn",
		} {
			_, found := result.FindEvent(eventType)
			require.False(t, found, "failed transaction emitted mutation event %s", eventType)
		}

		afterFailure, snapshotErr := network.SnapshotNFTState(
			ctx,
			"negative-no-op-controller-update-after",
			snapshotRequest,
		)
		require.NoError(t, snapshotErr)
		require.Equal(t, before, afterFailure, "failed DeliverTx changed semantic NFT state")
	})

	for _, testCase := range []struct {
		name       string
		endpoint   string
		parameters url.Values
		message    string
	}{
		{
			name:       "panacea limit above hard maximum",
			endpoint:   "/panacea/nft/v1/nfts",
			parameters: url.Values{"class_id": {classID}, "pagination.limit": {"101"}},
			message:    "pagination limit must not exceed 100",
		},
		{
			name:       "panacea offset is unsupported",
			endpoint:   "/panacea/nft/v1/nfts",
			parameters: url.Values{"class_id": {classID}, "pagination.offset": {"1"}},
			message:    "pagination offset is not supported",
		},
		{
			name:       "panacea total count is unsupported",
			endpoint:   "/panacea/nft/v1/nfts",
			parameters: url.Values{"class_id": {classID}, "pagination.count_total": {"true"}},
			message:    "pagination count_total is not supported",
		},
		{
			name:       "standard limit above hard maximum",
			endpoint:   "/cosmos/nft/v1beta1/nfts",
			parameters: url.Values{"class_id": {classID}, "pagination.limit": {"101"}},
			message:    "pagination limit must not exceed 100",
		},
		{
			name:       "standard offset is unsupported",
			endpoint:   "/cosmos/nft/v1beta1/nfts",
			parameters: url.Values{"class_id": {classID}, "pagination.offset": {"1"}},
			message:    "pagination offset is not supported",
		},
		{
			name:       "standard total count is unsupported",
			endpoint:   "/cosmos/nft/v1beta1/nfts",
			parameters: url.Values{"class_id": {classID}, "pagination.count_total": {"true"}},
			message:    "pagination count_total is not supported",
		},
		{
			name:       "panacea list requires a filter",
			endpoint:   "/panacea/nft/v1/nfts",
			parameters: url.Values{"pagination.limit": {"1"}},
			message:    "must provide at least one of class_id or owner",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, queryErr := network.FullNodeRESTGet(
				ctx,
				&http.Client{Timeout: 5 * time.Second},
				"negative-pagination-"+paginationStepSlug(testCase.name),
				testCase.endpoint+"?"+testCase.parameters.Encode(),
			)
			require.ErrorContains(t, queryErr, "HTTP 400")
			require.ErrorContains(t, queryErr, testCase.message)
		})
	}

	after, err := network.SnapshotNFTState(ctx, "negative-pagination-after", snapshotRequest)
	require.NoError(t, err)
	require.Equal(t, before, after, "rejected queries changed semantic NFT state")
}

func buildAndFundPaginationWallet(
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
		sdkmath.NewInt(30_000_000).String()+"umed",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	return wallet
}

func paginationStepSlug(value string) string {
	result := make([]byte, 0, len(value))
	for index := range len(value) {
		character := value[index]
		if character >= 'a' && character <= 'z' {
			result = append(result, character)
			continue
		}
		if len(result) > 0 && result[len(result)-1] != '-' {
			result = append(result, '-')
		}
	}
	return string(result)
}
