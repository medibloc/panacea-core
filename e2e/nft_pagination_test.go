package e2e_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const paginationLocalClassID = "pagination.class"

type nftRecordListItem struct {
	NFT struct {
		ClassID string `json:"class_id"`
		ID      string `json:"id"`
		URI     string `json:"uri"`
		URIHash string `json:"uri_hash"`
	} `json:"nft"`
	Owner  string `json:"owner"`
	Status string `json:"status"`
}

type nftRecordsQueryResponse struct {
	NFTRecords []nftRecordListItem `json:"nft_records"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

type nftPaginationRestartEvidence struct {
	QueryHeight int64                   `json:"query_height"`
	Owner       string                  `json:"owner"`
	Cursor      string                  `json:"cursor"`
	FirstPage   nftRecordsQueryResponse `json:"first_page"`
	CursorPage  nftRecordsQueryResponse `json:"cursor_page"`
}

// runNFTPaginationCompatibility exercises the list-query contract against the
// same live network as the lifecycle. It creates records spanning two classes
// and one owner so every filter and opaque cursor has a non-trivial result.
func runNFTPaginationCompatibility(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	creator ibc.Wallet,
	controller ibc.Wallet,
	owner ibc.Wallet,
	expectedPreexistingClassIDs []string,
) nftPaginationRestartEvidence {
	t.Helper()
	txNode := network.Chain.Validators[0]
	firstClassID := creator.FormattedAddress() + ":" + lifecycleLocalClassID
	secondClassID := creator.FormattedAddress() + ":" + paginationLocalClassID

	created, err := network.BroadcastAndWaitTx(
		ctx,
		"pagination-create-class",
		txNode,
		creator.KeyName(),
		"nft", "create-class",
		paginationLocalClassID,
		"Pagination Class",
		"PAGE",
		"owner-transferable",
		"true",
		"10",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, created, "panacea.nft.v1.EventClassCreated", map[string]string{
		"class_id": secondClassID,
		"creator":  creator.FormattedAddress(),
	})

	first := mintPaginationNFT(t, ctx, network, controller, owner, firstClassID, "page.1")
	second := mintPaginationNFT(t, ctx, network, controller, owner, firstClassID, "page.2")
	third := mintPaginationNFT(t, ctx, network, creator, owner, secondClassID, "cross.1")
	fourth := mintPaginationNFT(t, ctx, network, creator, owner, secondClassID, "cross.2")
	burnCandidate := mintPaginationNFT(t, ctx, network, creator, owner, secondClassID, "cross.burned")
	require.Less(t, first.HeightInt64(), second.HeightInt64())
	require.Less(t, second.HeightInt64(), third.HeightInt64())
	require.Less(t, third.HeightInt64(), fourth.HeightInt64())
	require.Less(t, fourth.HeightInt64(), burnCandidate.HeightInt64())

	burned, err := network.BroadcastAndWaitTx(
		ctx,
		"pagination-burn-excluded",
		txNode,
		owner.KeyName(),
		"nft", "burn", secondClassID, "cross.burned",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, burned, "cosmos.nft.v1beta1.EventBurn", map[string]string{
		"class_id": secondClassID,
		"id":       "cross.burned",
		"owner":    owner.FormattedAddress(),
	})

	wantOwnerKeys := []string{
		firstClassID + "/page.1",
		firstClassID + "/page.2",
		secondClassID + "/cross.1",
		secondClassID + "/cross.2",
	}
	sort.Strings(wantOwnerKeys)

	queryHeight := burned.HeightInt64()
	require.Positive(t, queryHeight)
	pinnedCtx := harness.ContextAtHeight(ctx, queryHeight)

	ownerDefault, err := network.QueryNFTsGRPC(pinnedCtx, "pagination-owner-limit-zero", &upstreamnft.QueryNFTsRequest{
		Owner:      owner.FormattedAddress(),
		Pagination: &query.PageRequest{Limit: 0},
	})
	require.NoError(t, err)
	require.Equal(t, wantOwnerKeys, standardNFTKeys(ownerDefault.Nfts))
	require.NotNil(t, ownerDefault.Pagination)
	require.Empty(t, ownerDefault.Pagination.NextKey)
	require.Zero(t, ownerDefault.Pagination.Total)

	ownerMaximum, err := network.QueryNFTsGRPC(pinnedCtx, "pagination-owner-limit-100", &upstreamnft.QueryNFTsRequest{
		Owner:      owner.FormattedAddress(),
		Pagination: &query.PageRequest{Limit: 100},
	})
	require.NoError(t, err)
	require.Equal(t, wantOwnerKeys, standardNFTKeys(ownerMaximum.Nfts))
	require.Equal(t, wantOwnerKeys, collectStandardNFTOpaquePages(t, pinnedCtx, network, owner.FormattedAddress()))

	classOnly, err := network.QueryNFTsGRPC(pinnedCtx, "pagination-class-filter", &upstreamnft.QueryNFTsRequest{
		ClassId: firstClassID,
	})
	require.NoError(t, err)
	require.Equal(t, []string{firstClassID + "/page.1", firstClassID + "/page.2"}, standardNFTKeys(classOnly.Nfts))

	intersection, err := network.QueryNFTsGRPC(pinnedCtx, "pagination-class-owner-filter", &upstreamnft.QueryNFTsRequest{
		ClassId: firstClassID,
		Owner:   owner.FormattedAddress(),
	})
	require.NoError(t, err)
	require.Equal(t, standardNFTKeys(classOnly.Nfts), standardNFTKeys(intersection.Nfts))

	classesDefault, err := network.QueryNFTClassesGRPC(pinnedCtx, "pagination-classes-limit-zero", &upstreamnft.QueryClassesRequest{
		Pagination: &query.PageRequest{Limit: 0},
	})
	require.NoError(t, err)
	wantClassIDs := expectedStandardClassIDs(
		expectedPreexistingClassIDs,
		firstClassID,
		secondClassID,
	)
	require.Equal(t, wantClassIDs, standardClassIDs(classesDefault.Classes))
	require.NotNil(t, classesDefault.Pagination)
	require.Zero(t, classesDefault.Pagination.Total)

	classesMaximum, err := network.QueryNFTClassesGRPC(pinnedCtx, "pagination-classes-limit-100", &upstreamnft.QueryClassesRequest{
		Pagination: &query.PageRequest{Limit: 100},
	})
	require.NoError(t, err)
	require.Equal(t, wantClassIDs, standardClassIDs(classesMaximum.Classes))
	require.Equal(t, wantClassIDs, collectStandardClassOpaquePages(t, pinnedCtx, network))

	standardClient := upstreamnft.NewQueryClient(network.Chain.FullNodes[0].GrpcConn)
	for name, pagination := range map[string]*query.PageRequest{
		"limit-over-100": {Limit: 101},
		"offset":         {Offset: 1},
		"count-total":    {CountTotal: true},
	} {
		t.Run("standard-nfts-reject-"+name, func(t *testing.T) {
			_, queryErr := standardClient.NFTs(pinnedCtx, &upstreamnft.QueryNFTsRequest{
				Owner:      owner.FormattedAddress(),
				Pagination: pagination,
			})
			require.Equal(t, codes.InvalidArgument, status.Code(queryErr))
		})
		t.Run("standard-classes-reject-"+name, func(t *testing.T) {
			_, queryErr := standardClient.Classes(pinnedCtx, &upstreamnft.QueryClassesRequest{Pagination: pagination})
			require.Equal(t, codes.InvalidArgument, status.Code(queryErr))
		})
	}

	assertPanaceaPaginationContract(
		t,
		pinnedCtx,
		network,
		queryHeight,
		firstClassID,
		secondClassID,
		owner.FormattedAddress(),
		wantOwnerKeys,
	)
	return captureNFTRestartPaginationEvidence(
		t,
		ctx,
		network,
		"pagination-restart-checkpoint",
		owner.FormattedAddress(),
		"",
	)
}

func expectedStandardClassIDs(preexisting []string, created ...string) []string {
	unique := make(map[string]struct{}, len(preexisting)+len(created))
	for _, classID := range preexisting {
		unique[classID] = struct{}{}
	}
	for _, classID := range created {
		unique[classID] = struct{}{}
	}
	classIDs := make([]string, 0, len(unique))
	for classID := range unique {
		classIDs = append(classIDs, classID)
	}
	sort.Strings(classIDs)
	return classIDs
}

func TestExpectedStandardClassIDsIncludesExplicitPreexistingClasses(t *testing.T) {
	require.Equal(t, []string{"legacy", "lifecycle", "pagination"}, expectedStandardClassIDs(
		[]string{"legacy", "lifecycle"},
		"pagination",
		"lifecycle",
	))
}

func captureNFTRestartPaginationEvidence(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	owner string,
	expectedCursor string,
) nftPaginationRestartEvidence {
	t.Helper()
	consensusHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	require.Greater(t, consensusHeight, int64(1))
	queryHeight := consensusHeight - 1
	pinnedCtx := harness.ContextAtHeight(ctx, queryHeight)
	firstPage := queryPanaceaPaginationBoundaries(
		t,
		pinnedCtx,
		network,
		queryHeight,
		step+"-first-page",
		panaceaPaginationSpec{owner: owner, limit: 1},
	)
	require.Len(t, firstPage.NFTRecords, 1)
	require.NotEmpty(t, firstPage.Pagination.NextKey)
	cursor := firstPage.Pagination.NextKey
	if expectedCursor != "" {
		require.Equal(t, expectedCursor, cursor, "pagination cursor changed across restart")
		cursor = expectedCursor
	}
	cursorPage := queryPanaceaPaginationBoundaries(
		t,
		pinnedCtx,
		network,
		queryHeight,
		step+"-cursor-page",
		panaceaPaginationSpec{owner: owner, pageKey: cursor, limit: 1},
	)
	require.Len(t, cursorPage.NFTRecords, 1)
	return nftPaginationRestartEvidence{
		QueryHeight: queryHeight,
		Owner:       owner,
		Cursor:      cursor,
		FirstPage:   firstPage,
		CursorPage:  cursorPage,
	}
}

func mintPaginationNFT(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	controller ibc.Wallet,
	owner ibc.Wallet,
	classID string,
	nftID string,
) *harness.TxResult {
	t.Helper()
	result, err := network.BroadcastAndWaitTx(
		ctx,
		"pagination-mint-"+nftID,
		network.Chain.Validators[0],
		controller.KeyName(),
		"nft", "mint", classID, nftID, owner.FormattedAddress(),
		"--uri", "https://example.test/nfts/"+nftID+".json",
		"--uri-hash", lifecycleNFTURIHash,
		"--data", lifecycleDataJSON,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, result, "cosmos.nft.v1beta1.EventMint", map[string]string{
		"class_id": classID,
		"id":       nftID,
		"owner":    owner.FormattedAddress(),
	})
	return result
}

func collectStandardNFTOpaquePages(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	owner string,
) []string {
	t.Helper()
	var cursor []byte
	var keys []string
	for page := 0; page < 10; page++ {
		response, err := network.QueryNFTsGRPC(ctx, "pagination-owner-cursor-"+strconv.Itoa(page+1), &upstreamnft.QueryNFTsRequest{
			Owner: owner,
			Pagination: &query.PageRequest{
				Key:   cursor,
				Limit: 1,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, response.Pagination)
		require.Len(t, response.Nfts, 1)
		keys = append(keys, standardNFTKeys(response.Nfts)...)
		if len(response.Pagination.NextKey) == 0 {
			return keys
		}
		require.NotEqual(t, cursor, response.Pagination.NextKey)
		cursor = append(cursor[:0], response.Pagination.NextKey...)
	}
	t.Fatalf("owner NFT pagination did not terminate")
	return nil
}

func collectStandardClassOpaquePages(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
) []string {
	t.Helper()
	var cursor []byte
	var classIDs []string
	for page := 0; page < 10; page++ {
		response, err := network.QueryNFTClassesGRPC(ctx, "pagination-class-cursor-"+strconv.Itoa(page+1), &upstreamnft.QueryClassesRequest{
			Pagination: &query.PageRequest{Key: cursor, Limit: 1},
		})
		require.NoError(t, err)
		require.NotNil(t, response.Pagination)
		require.Len(t, response.Classes, 1)
		classIDs = append(classIDs, standardClassIDs(response.Classes)...)
		if len(response.Pagination.NextKey) == 0 {
			return classIDs
		}
		require.NotEqual(t, cursor, response.Pagination.NextKey)
		cursor = append(cursor[:0], response.Pagination.NextKey...)
	}
	t.Fatalf("class pagination did not terminate")
	return nil
}

func standardNFTKeys(nfts []*upstreamnft.NFT) []string {
	keys := make([]string, 0, len(nfts))
	for _, nft := range nfts {
		keys = append(keys, nft.ClassId+"/"+nft.Id)
	}
	return keys
}

func standardClassIDs(classes []*upstreamnft.Class) []string {
	classIDs := make([]string, 0, len(classes))
	for _, class := range classes {
		classIDs = append(classIDs, class.Id)
	}
	return classIDs
}

func panaceaNFTRecordKeys(records []nftRecordListItem) []string {
	keys := make([]string, 0, len(records))
	for _, record := range records {
		keys = append(keys, record.NFT.ClassID+"/"+record.NFT.ID)
	}
	return keys
}

type panaceaPaginationSpec struct {
	classID string
	owner   string
	pageKey string
	limit   uint64
	reverse bool
}

func (spec panaceaPaginationSpec) cliArguments(height int64) []string {
	arguments := []string{"nft", "nft-records"}
	if spec.classID != "" {
		arguments = append(arguments, "--class-id", spec.classID)
	}
	if spec.owner != "" {
		arguments = append(arguments, "--owner", spec.owner)
	}
	if spec.pageKey != "" {
		arguments = append(arguments, "--page-key", spec.pageKey)
	}
	arguments = append(arguments, "--limit", strconv.FormatUint(spec.limit, 10))
	if spec.reverse {
		arguments = append(arguments, "--reverse")
	}
	if height > 0 {
		arguments = append(arguments, "--height", strconv.FormatInt(height, 10))
	}
	return arguments
}

func (spec panaceaPaginationSpec) restPath(basePath string) string {
	values := url.Values{}
	if spec.classID != "" {
		values.Set("class_id", spec.classID)
	}
	if spec.owner != "" {
		values.Set("owner", spec.owner)
	}
	if spec.pageKey != "" {
		values.Set("pagination.key", spec.pageKey)
	}
	values.Set("pagination.limit", strconv.FormatUint(spec.limit, 10))
	if spec.reverse {
		values.Set("pagination.reverse", "true")
	}
	return basePath + "?" + values.Encode()
}

func queryPanaceaPaginationBoundaries(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	height int64,
	step string,
	spec panaceaPaginationSpec,
) nftRecordsQueryResponse {
	t.Helper()
	arguments := spec.cliArguments(height)
	cliRaw, err := network.FullNodeCLIQuery(ctx, step+"-cli", arguments...)
	require.NoError(t, err)
	grpcRaw, err := network.FullNodeGRPCQuery(ctx, step+"-custom-grpc", arguments...)
	require.NoError(t, err)
	require.JSONEq(t, string(cliRaw), string(grpcRaw))

	restPath := spec.restPath("/panacea/nft/v1/nfts")
	if spec.pageKey != "" {
		parsed, parseErr := url.Parse(restPath)
		require.NoError(t, parseErr)
		require.Equal(t, spec.pageKey, parsed.Query().Get("pagination.key"))
		require.Contains(t, restPath, "pagination.key="+url.QueryEscape(spec.pageKey))
	}
	restRaw, err := network.FullNodeRESTGetAtHeight(ctx, nil, step+"-rest", restPath, height)
	require.NoError(t, err)
	require.JSONEq(t, string(cliRaw), string(restRaw))

	var response nftRecordsQueryResponse
	require.NoError(t, json.Unmarshal(cliRaw, &response))
	for _, record := range response.NFTRecords {
		require.NotEmpty(t, record.NFT.ClassID)
		require.NotEmpty(t, record.NFT.ID)
		require.Equal(t, "LIVE_NFT_STATUS_ACTIVE", record.Status)
	}

	var pageKey []byte
	if spec.pageKey != "" {
		pageKey, err = base64.StdEncoding.DecodeString(spec.pageKey)
		require.NoError(t, err)
		require.NotEmpty(t, pageKey)
	}
	standard, err := network.QueryNFTsGRPC(ctx, step+"-standard-grpc", &upstreamnft.QueryNFTsRequest{
		ClassId: spec.classID,
		Owner:   spec.owner,
		Pagination: &query.PageRequest{
			Key:     pageKey,
			Limit:   spec.limit,
			Reverse: spec.reverse,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, standard.Pagination)
	require.Equal(t, panaceaNFTRecordKeys(response.NFTRecords), standardNFTKeys(standard.Nfts))
	require.Equal(t, response.Pagination.NextKey, base64.StdEncoding.EncodeToString(standard.Pagination.NextKey))
	require.Zero(t, standard.Pagination.Total)
	return response
}

func collectPanaceaPaginationPages(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	height int64,
	step string,
	spec panaceaPaginationSpec,
) []string {
	t.Helper()
	var keys []string
	for page := 0; page < 10; page++ {
		response := queryPanaceaPaginationBoundaries(
			t,
			ctx,
			network,
			height,
			step+"-page-"+strconv.Itoa(page+1),
			panaceaPaginationSpec{
				classID: spec.classID,
				owner:   spec.owner,
				pageKey: spec.pageKey,
				limit:   1,
				reverse: spec.reverse,
			},
		)
		require.Len(t, response.NFTRecords, 1)
		keys = append(keys, panaceaNFTRecordKeys(response.NFTRecords)...)
		if response.Pagination.NextKey == "" {
			return keys
		}
		require.NotEqual(t, spec.pageKey, response.Pagination.NextKey)
		decoded, err := base64.StdEncoding.DecodeString(response.Pagination.NextKey)
		require.NoError(t, err)
		require.NotEmpty(t, decoded)
		spec.pageKey = response.Pagination.NextKey
	}
	t.Fatalf("Panacea NFT pagination did not terminate")
	return nil
}

func reversePaginationKeys(keys []string) []string {
	reversed := append([]string(nil), keys...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return reversed
}

func assertPanaceaPaginationContract(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	height int64,
	firstClassID string,
	secondClassID string,
	owner string,
	wantOwnerKeys []string,
) {
	t.Helper()
	wantFirstClassKeys := []string{firstClassID + "/page.1", firstClassID + "/page.2"}
	wantSecondClassKeys := []string{secondClassID + "/cross.1", secondClassID + "/cross.2"}

	ownerDefault := queryPanaceaPaginationBoundaries(
		t, ctx, network, height, "pagination-panacea-owner-limit-zero",
		panaceaPaginationSpec{owner: owner, limit: 0},
	)
	require.Equal(t, wantOwnerKeys, panaceaNFTRecordKeys(ownerDefault.NFTRecords))
	require.Empty(t, ownerDefault.Pagination.NextKey)

	ownerMaximum := queryPanaceaPaginationBoundaries(
		t, ctx, network, height, "pagination-panacea-owner-limit-100",
		panaceaPaginationSpec{owner: owner, limit: 100},
	)
	require.Equal(t, wantOwnerKeys, panaceaNFTRecordKeys(ownerMaximum.NFTRecords))
	require.Empty(t, ownerMaximum.Pagination.NextKey)

	ownerForward := collectPanaceaPaginationPages(
		t, ctx, network, height, "pagination-panacea-owner-forward",
		panaceaPaginationSpec{owner: owner},
	)
	require.Equal(t, wantOwnerKeys, ownerForward)
	ownerReverse := collectPanaceaPaginationPages(
		t, ctx, network, height, "pagination-panacea-owner-reverse",
		panaceaPaginationSpec{owner: owner, reverse: true},
	)
	require.Equal(t, reversePaginationKeys(wantOwnerKeys), ownerReverse)

	firstClassForward := collectPanaceaPaginationPages(
		t, ctx, network, height, "pagination-panacea-first-class-forward",
		panaceaPaginationSpec{classID: firstClassID},
	)
	require.Equal(t, wantFirstClassKeys, firstClassForward)
	secondClassForward := collectPanaceaPaginationPages(
		t, ctx, network, height, "pagination-panacea-second-class-forward",
		panaceaPaginationSpec{classID: secondClassID},
	)
	require.Equal(t, wantSecondClassKeys, secondClassForward)
	burnedKey := secondClassID + "/cross.burned"
	require.NotContains(t, panaceaNFTRecordKeys(ownerMaximum.NFTRecords), burnedKey)
	require.NotContains(t, secondClassForward, burnedKey)

	intersection := collectPanaceaPaginationPages(
		t, ctx, network, height, "pagination-panacea-class-owner-intersection",
		panaceaPaginationSpec{classID: firstClassID, owner: owner},
	)
	require.Equal(t, wantFirstClassKeys, intersection)

	assertNFTPaginationRESTRejections(t, ctx, network, height, secondClassID)
	assertCustomNFTPaginationLimitRejection(t, ctx, network, height, secondClassID)
}

func assertNFTPaginationRESTRejections(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	height int64,
	classID string,
) {
	t.Helper()
	boundaries := []struct {
		name string
		path string
	}{
		{name: "custom", path: "/panacea/nft/v1/nfts"},
		{name: "standard", path: "/cosmos/nft/v1beta1/nfts"},
	}
	testCases := []struct {
		name       string
		field      string
		value      string
		wantDetail string
	}{
		{name: "limit-101", field: "pagination.limit", value: "101", wantDetail: "must not exceed 100"},
		{name: "offset", field: "pagination.offset", value: "1", wantDetail: "offset is not supported"},
		{name: "count-total", field: "pagination.count_total", value: "true", wantDetail: "count_total is not supported"},
	}
	for _, boundary := range boundaries {
		for _, testCase := range testCases {
			path := panaceaPaginationSpec{classID: classID, limit: 100}.restPath(boundary.path)
			parsed, err := url.Parse(path)
			require.NoError(t, err)
			values := parsed.Query()
			values.Set(testCase.field, testCase.value)
			parsed.RawQuery = values.Encode()
			_, err = network.FullNodeRESTGetAtHeight(
				ctx,
				nil,
				"pagination-reject-"+boundary.name+"-"+testCase.name,
				parsed.String(),
				height,
			)
			require.ErrorContains(t, err, "HTTP 400")
			require.ErrorContains(t, err, testCase.wantDetail)
		}
	}

	filterlessPath := panaceaPaginationSpec{limit: 1}.restPath("/panacea/nft/v1/nfts")
	_, err := network.FullNodeRESTGetAtHeight(
		ctx,
		nil,
		"pagination-reject-custom-missing-filter",
		filterlessPath,
		height,
	)
	require.ErrorContains(t, err, "HTTP 400")
	require.ErrorContains(t, err, "must provide at least one of class_id or owner")
}

// Offset and count-total are deliberately absent from the custom CLI. Limit
// 101 reaches the server on both CLI/RPC and explicit gRPC. Interchaintest's
// short-lived CLI container returns command output inside its error on a
// non-zero exit, so include stdout, stderr, and the error in the diagnostic.
func assertCustomNFTPaginationLimitRejection(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	height int64,
	classID string,
) {
	t.Helper()
	fullNode := network.Chain.FullNodes[0]
	baseArguments := panaceaPaginationSpec{classID: classID, limit: 101}.cliArguments(height)
	testCases := []struct {
		name      string
		arguments []string
	}{
		{name: "cli", arguments: append([]string(nil), baseArguments...)},
		{
			name: "custom-grpc",
			arguments: append(
				append([]string(nil), baseArguments...),
				"--grpc-addr", fullNode.HostName()+":9090",
				"--grpc-insecure",
			),
		},
	}
	for _, testCase := range testCases {
		stdout, stderr, err := fullNode.ExecQuery(ctx, testCase.arguments...)
		require.Error(t, err, testCase.name)
		require.Contains(
			t,
			queryRejectionDiagnostic(stdout, stderr, err),
			"pagination limit must not exceed 100",
			testCase.name,
		)
	}
}

func queryRejectionDiagnostic(stdout, stderr []byte, err error) string {
	parts := []string{strings.TrimSpace(string(stdout)), strings.TrimSpace(string(stderr))}
	if err != nil {
		parts = append(parts, err.Error())
	}
	return strings.Join(parts, "\n")
}

func TestPanaceaPaginationRESTPathEscapesOpaqueCursor(t *testing.T) {
	cursor := base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 251, 255})
	path := panaceaPaginationSpec{
		classID: "panacea1creator:pagination.class",
		pageKey: cursor,
		limit:   1,
	}.restPath("/panacea/nft/v1/nfts")
	parsed, err := url.Parse(path)
	require.NoError(t, err)
	require.Equal(t, cursor, parsed.Query().Get("pagination.key"))
	require.Contains(t, path, "%2B")
	require.Contains(t, path, "%2F")
	require.Contains(t, path, "%3D")
}

func TestQueryRejectionDiagnosticIncludesExecError(t *testing.T) {
	diagnostic := queryRejectionDiagnostic(nil, nil, errors.New("pagination limit must not exceed 100"))
	require.Contains(t, diagnostic, "pagination limit must not exceed 100")
}
