package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const nftBoundaryBasicDataTypeURL = "/panacea.nft.v1.BasicNFTData"

type nftBoundaryJSONNFT struct {
	ClassID string         `json:"class_id"`
	ID      string         `json:"id"`
	URI     string         `json:"uri"`
	URIHash string         `json:"uri_hash"`
	Data    map[string]any `json:"data"`
}

type nftBoundaryJSONClass struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Symbol      string         `json:"symbol"`
	Description string         `json:"description"`
	URI         string         `json:"uri"`
	URIHash     string         `json:"uri_hash"`
	Data        map[string]any `json:"data"`
}

type nftBoundaryCustomClassResponse struct {
	ClassRecord *struct {
		Class nftBoundaryJSONClass `json:"class"`
	} `json:"class_record"`
}

type nftBoundaryCustomNFTResponse struct {
	NFTRecord *struct {
		Live *struct {
			NFT   nftBoundaryJSONNFT `json:"nft"`
			Owner string             `json:"owner"`
		} `json:"live"`
		BurnTombstone *struct {
			ClassID string         `json:"class_id"`
			NFTID   string         `json:"nft_id"`
			URI     string         `json:"uri"`
			URIHash string         `json:"uri_hash"`
			Data    map[string]any `json:"data"`
		} `json:"burn_tombstone"`
	} `json:"nft_record"`
}

type nftBoundaryStandardClassResponse struct {
	Class *nftBoundaryJSONClass `json:"class"`
}

type nftBoundaryStandardNFTResponse struct {
	NFT *nftBoundaryJSONNFT `json:"nft"`
}

type nftBoundaryOwnerResponse struct {
	Owner string `json:"owner"`
}

type nftBoundaryAmountResponse struct {
	Amount string `json:"amount"`
}

type nftBoundaryDataProjection struct {
	TypeURL     string
	Name        string
	Description string
	ImageURI    string
}

type nftBoundaryNFTProjection struct {
	ClassID string
	ID      string
	URI     string
	URIHash string
	Data    nftBoundaryDataProjection
}

type nftBoundaryClassProjection struct {
	ID          string
	Name        string
	Symbol      string
	Description string
	URI         string
	URIHash     string
	Data        nftBoundaryDataProjection
}

func nftBoundaryEncodedPathSegment(identifier string) string {
	return strings.ReplaceAll(url.PathEscape(identifier), ":", "%3A")
}

// assertClassQueryBoundaryParityAtTx pins every query to the transaction's
// committed height. It compares the complete Panacea response across CLI,
// explicit gRPC, and raw/encoded REST routes, then compares the standard
// class fields exposed by both query services across standard CLI, typed gRPC,
// and raw/encoded REST.
func assertClassQueryBoundaryParityAtTx(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	tx *harness.TxResult,
	classID string,
) classRecordQueryResponse {
	t.Helper()
	height := nftBoundaryCommittedHeight(t, tx)
	return assertClassQueryBoundaryParityAtHeight(t, ctx, network, step, height, classID)
}

func assertClassQueryBoundaryParityAtHeight(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	height int64,
	classID string,
) classRecordQueryResponse {
	t.Helper()
	require.Positive(t, height)
	heightFlag := strconv.FormatInt(height, 10)

	customCLI, err := network.FullNodeCLIQuery(
		ctx, step+"-panacea-cli", "nft", "class-record", classID, "--height", heightFlag,
	)
	require.NoError(t, err)
	customGRPC, err := network.FullNodeGRPCQuery(
		ctx, step+"-panacea-grpc", "nft", "class-record", classID, "--height", heightFlag,
	)
	require.NoError(t, err)
	customRESTRaw, customRESTEncoded := nftBoundaryRESTPairAtHeight(
		t,
		ctx,
		network,
		step+"-panacea-class",
		height,
		"/panacea/nft/v1/classes/"+classID,
		"/panacea/nft/v1/classes/"+nftBoundaryEncodedPathSegment(classID),
	)
	nftBoundaryRequireJSONEqual(t, "Panacea class CLI and gRPC", customCLI, customGRPC)
	nftBoundaryRequireJSONEqual(t, "Panacea class CLI and raw REST", customCLI, customRESTRaw)
	nftBoundaryRequireJSONEqual(t, "Panacea class raw and encoded REST", customRESTRaw, customRESTEncoded)

	custom := nftBoundaryDecode[nftBoundaryCustomClassResponse](t, "Panacea class", customCLI)
	require.NotNil(t, custom.ClassRecord)
	require.Equal(t, classID, custom.ClassRecord.Class.ID)
	customProjection, err := nftBoundaryProjectJSONClass(custom.ClassRecord.Class)
	require.NoError(t, err)

	standardCLI, err := network.FullNodeCLIQuery(
		ctx, step+"-standard-cli", "nft", "class", classID, "--height", heightFlag,
	)
	require.NoError(t, err)
	standardCLIResponse := nftBoundaryDecode[nftBoundaryStandardClassResponse](t, "standard class CLI", standardCLI)
	require.NotNil(t, standardCLIResponse.Class)
	standardCLIProjection, err := nftBoundaryProjectJSONClass(*standardCLIResponse.Class)
	require.NoError(t, err)

	typedResponse, err := network.QueryNFTClassGRPC(
		harness.ContextAtHeight(ctx, height), step+"-standard-typed-grpc", classID,
	)
	require.NoError(t, err)
	require.NotNil(t, typedResponse)
	require.NotNil(t, typedResponse.Class)
	typedProjection, err := nftBoundaryProjectTypedClass(typedResponse.Class)
	require.NoError(t, err)

	standardRESTRaw, standardRESTEncoded := nftBoundaryRESTPairAtHeight(
		t,
		ctx,
		network,
		step+"-standard-class",
		height,
		"/cosmos/nft/v1beta1/classes/"+classID,
		"/cosmos/nft/v1beta1/classes/"+nftBoundaryEncodedPathSegment(classID),
	)
	nftBoundaryRequireJSONEqual(t, "standard class raw and encoded REST", standardRESTRaw, standardRESTEncoded)
	standardRESTResponse := nftBoundaryDecode[nftBoundaryStandardClassResponse](t, "standard class REST", standardRESTRaw)
	require.NotNil(t, standardRESTResponse.Class)
	standardRESTProjection, err := nftBoundaryProjectJSONClass(*standardRESTResponse.Class)
	require.NoError(t, err)

	require.Equal(t, customProjection, standardCLIProjection, "Panacea and standard class projections at height %d", height)
	require.Equal(t, customProjection, typedProjection, "Panacea and typed gRPC class projections at height %d", height)
	require.Equal(t, customProjection, standardRESTProjection, "Panacea and standard REST class projections at height %d", height)

	return nftBoundaryDecode[classRecordQueryResponse](t, "Panacea lifecycle class", customCLI)
}

// assertNFTQueryBoundaryParityAtTx compares all custom-module state at one
// committed height. A live record is also projected onto the standard NFT and
// owner contracts and compared across standard CLI, typed gRPC, and raw/
// encoded REST. For a tombstone, the same three standard owner boundaries must
// all report the NFT as ownerless; the caller can separately assert the
// standard NFT endpoint's expected NotFound response without turning an
// expected negative query into a harness failure artifact.
func assertNFTQueryBoundaryParityAtTx(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	tx *harness.TxResult,
	classID string,
	nftID string,
) nftRecordQueryResponse {
	t.Helper()
	height := nftBoundaryCommittedHeight(t, tx)
	return assertNFTQueryBoundaryParityAtHeight(t, ctx, network, step, height, classID, nftID)
}

func assertNFTQueryBoundaryParityAtHeight(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	height int64,
	classID string,
	nftID string,
) nftRecordQueryResponse {
	t.Helper()
	require.Positive(t, height)
	heightFlag := strconv.FormatInt(height, 10)

	customCLI, err := network.FullNodeCLIQuery(
		ctx, step+"-panacea-cli", "nft", "nft-record", classID, nftID, "--height", heightFlag,
	)
	require.NoError(t, err)
	customGRPC, err := network.FullNodeGRPCQuery(
		ctx, step+"-panacea-grpc", "nft", "nft-record", classID, nftID, "--height", heightFlag,
	)
	require.NoError(t, err)
	customRESTRaw, customRESTEncoded := nftBoundaryRESTPairAtHeight(
		t,
		ctx,
		network,
		step+"-panacea-nft",
		height,
		"/panacea/nft/v1/nfts/"+classID+"/"+nftID,
		"/panacea/nft/v1/nfts/"+nftBoundaryEncodedPathSegment(classID)+"/"+nftBoundaryEncodedPathSegment(nftID),
	)
	nftBoundaryRequireJSONEqual(t, "Panacea NFT CLI and gRPC", customCLI, customGRPC)
	nftBoundaryRequireJSONEqual(t, "Panacea NFT CLI and raw REST", customCLI, customRESTRaw)
	nftBoundaryRequireJSONEqual(t, "Panacea NFT raw and encoded REST", customRESTRaw, customRESTEncoded)

	custom := nftBoundaryDecode[nftBoundaryCustomNFTResponse](t, "Panacea NFT", customCLI)
	require.NotNil(t, custom.NFTRecord)
	require.NotEqual(t, custom.NFTRecord.Live == nil, custom.NFTRecord.BurnTombstone == nil, "NFT record must contain exactly one variant")

	wantOwner := ""
	if custom.NFTRecord.Live != nil {
		require.Equal(t, classID, custom.NFTRecord.Live.NFT.ClassID)
		require.Equal(t, nftID, custom.NFTRecord.Live.NFT.ID)
		wantOwner = custom.NFTRecord.Live.Owner
		nftBoundaryAssertLiveStandardProjectionAtHeight(
			t, ctx, network, step, height, custom.NFTRecord.Live.NFT,
		)
	} else {
		require.Equal(t, classID, custom.NFTRecord.BurnTombstone.ClassID)
		require.Equal(t, nftID, custom.NFTRecord.BurnTombstone.NFTID)
		_, err := nftBoundaryProjectJSONNFT(nftBoundaryJSONNFT{
			ClassID: custom.NFTRecord.BurnTombstone.ClassID,
			ID:      custom.NFTRecord.BurnTombstone.NFTID,
			URI:     custom.NFTRecord.BurnTombstone.URI,
			URIHash: custom.NFTRecord.BurnTombstone.URIHash,
			Data:    custom.NFTRecord.BurnTombstone.Data,
		})
		require.NoError(t, err, "burn tombstone metadata projection")
	}
	nftBoundaryAssertStandardOwnerAtHeight(t, ctx, network, step, height, classID, nftID, wantOwner)

	return nftBoundaryDecode[nftRecordQueryResponse](t, "Panacea lifecycle NFT", customCLI)
}

func nftBoundaryAssertLiveStandardProjectionAtHeight(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	height int64,
	customNFT nftBoundaryJSONNFT,
) {
	t.Helper()
	heightFlag := strconv.FormatInt(height, 10)
	customProjection, err := nftBoundaryProjectJSONNFT(customNFT)
	require.NoError(t, err)

	standardCLI, err := network.FullNodeCLIQuery(
		ctx,
		step+"-standard-nft-cli",
		"nft", "nft", customNFT.ClassID, customNFT.ID, "--height", heightFlag,
	)
	require.NoError(t, err)
	standardCLIResponse := nftBoundaryDecode[nftBoundaryStandardNFTResponse](t, "standard NFT CLI", standardCLI)
	require.NotNil(t, standardCLIResponse.NFT)
	standardCLIProjection, err := nftBoundaryProjectJSONNFT(*standardCLIResponse.NFT)
	require.NoError(t, err)

	typedResponse, err := network.QueryNFTGRPC(
		harness.ContextAtHeight(ctx, height),
		step+"-standard-nft-typed-grpc",
		customNFT.ClassID,
		customNFT.ID,
	)
	require.NoError(t, err)
	require.NotNil(t, typedResponse)
	typedProjection, err := nftBoundaryProjectTypedNFT(typedResponse.Nft)
	require.NoError(t, err)

	standardRESTRaw, standardRESTEncoded := nftBoundaryRESTPairAtHeight(
		t,
		ctx,
		network,
		step+"-standard-nft",
		height,
		"/cosmos/nft/v1beta1/nfts/"+customNFT.ClassID+"/"+customNFT.ID,
		"/cosmos/nft/v1beta1/nfts/"+nftBoundaryEncodedPathSegment(customNFT.ClassID)+"/"+nftBoundaryEncodedPathSegment(customNFT.ID),
	)
	nftBoundaryRequireJSONEqual(t, "standard NFT raw and encoded REST", standardRESTRaw, standardRESTEncoded)
	standardRESTResponse := nftBoundaryDecode[nftBoundaryStandardNFTResponse](t, "standard NFT REST", standardRESTRaw)
	require.NotNil(t, standardRESTResponse.NFT)
	standardRESTProjection, err := nftBoundaryProjectJSONNFT(*standardRESTResponse.NFT)
	require.NoError(t, err)

	require.Equal(t, customProjection, standardCLIProjection, "Panacea and standard CLI NFT projections at height %d", height)
	require.Equal(t, customProjection, typedProjection, "Panacea and typed gRPC NFT projections at height %d", height)
	require.Equal(t, customProjection, standardRESTProjection, "Panacea and standard REST NFT projections at height %d", height)
}

func nftBoundaryAssertStandardOwnerAtHeight(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	height int64,
	classID string,
	nftID string,
	wantOwner string,
) {
	t.Helper()
	heightFlag := strconv.FormatInt(height, 10)
	ownerCLI, err := network.FullNodeCLIQuery(
		ctx, step+"-standard-owner-cli", "nft", "owner", classID, nftID, "--height", heightFlag,
	)
	require.NoError(t, err)
	cliResponse := nftBoundaryDecode[nftBoundaryOwnerResponse](t, "standard owner CLI", ownerCLI)

	typedResponse, err := network.QueryNFTOwnerGRPC(
		harness.ContextAtHeight(ctx, height), step+"-standard-owner-typed-grpc", classID, nftID,
	)
	require.NoError(t, err)
	require.NotNil(t, typedResponse)

	ownerRESTRaw, ownerRESTEncoded := nftBoundaryRESTPairAtHeight(
		t,
		ctx,
		network,
		step+"-standard-owner",
		height,
		"/cosmos/nft/v1beta1/owner/"+classID+"/"+nftID,
		"/cosmos/nft/v1beta1/owner/"+nftBoundaryEncodedPathSegment(classID)+"/"+nftBoundaryEncodedPathSegment(nftID),
	)
	nftBoundaryRequireJSONEqual(t, "standard owner raw and encoded REST", ownerRESTRaw, ownerRESTEncoded)
	restResponse := nftBoundaryDecode[nftBoundaryOwnerResponse](t, "standard owner REST", ownerRESTRaw)

	require.Equal(t, wantOwner, cliResponse.Owner, "standard CLI owner at height %d", height)
	require.Equal(t, wantOwner, typedResponse.Owner, "standard typed gRPC owner at height %d", height)
	require.Equal(t, wantOwner, restResponse.Owner, "standard REST owner at height %d", height)
}

func assertStandardNFTAccountingAtTx(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	tx *harness.TxResult,
	classID string,
	wantSupply uint64,
	wantBalances map[string]uint64,
) {
	t.Helper()
	height := nftBoundaryCommittedHeight(t, tx)
	heightFlag := strconv.FormatInt(height, 10)

	supplyCLI, err := network.FullNodeCLIQuery(
		ctx, step+"-supply-cli", "nft", "supply", classID, "--height", heightFlag,
	)
	require.NoError(t, err)
	supplyCLIResponse := nftBoundaryDecode[nftBoundaryAmountResponse](t, "standard supply CLI", supplyCLI)
	supplyTyped, err := network.QueryNFTSupplyGRPC(
		harness.ContextAtHeight(ctx, height), step+"-supply-typed-grpc", classID,
	)
	require.NoError(t, err)
	supplyRESTRaw, supplyRESTEncoded := nftBoundaryRESTPairAtHeight(
		t,
		ctx,
		network,
		step+"-supply-rest",
		height,
		"/cosmos/nft/v1beta1/supply/"+classID,
		"/cosmos/nft/v1beta1/supply/"+nftBoundaryEncodedPathSegment(classID),
	)
	nftBoundaryRequireJSONEqual(t, "standard supply raw and encoded REST", supplyRESTRaw, supplyRESTEncoded)
	supplyRESTResponse := nftBoundaryDecode[nftBoundaryAmountResponse](t, "standard supply REST", supplyRESTRaw)
	require.Equal(t, wantSupply, nftBoundaryAmount(t, supplyCLIResponse.Amount), "CLI supply at height %d", height)
	require.Equal(t, wantSupply, supplyTyped.Amount, "typed gRPC supply at height %d", height)
	require.Equal(t, wantSupply, nftBoundaryAmount(t, supplyRESTResponse.Amount), "REST supply at height %d", height)

	owners := make([]string, 0, len(wantBalances))
	for owner := range wantBalances {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	for index, owner := range owners {
		want := wantBalances[owner]
		balanceStep := step + "-balance-" + strconv.Itoa(index)
		balanceCLI, queryErr := network.FullNodeCLIQuery(
			ctx, balanceStep+"-cli", "nft", "balance", owner, classID, "--height", heightFlag,
		)
		require.NoError(t, queryErr)
		balanceCLIResponse := nftBoundaryDecode[nftBoundaryAmountResponse](t, "standard balance CLI", balanceCLI)
		balanceTyped, queryErr := network.QueryNFTBalanceGRPC(
			harness.ContextAtHeight(ctx, height), balanceStep+"-typed-grpc", classID, owner,
		)
		require.NoError(t, queryErr)
		balanceRESTRaw, balanceRESTEncoded := nftBoundaryRESTPairAtHeight(
			t,
			ctx,
			network,
			balanceStep+"-rest",
			height,
			"/cosmos/nft/v1beta1/balance/"+owner+"/"+classID,
			"/cosmos/nft/v1beta1/balance/"+owner+"/"+nftBoundaryEncodedPathSegment(classID),
		)
		nftBoundaryRequireJSONEqual(t, "standard balance raw and encoded REST", balanceRESTRaw, balanceRESTEncoded)
		balanceRESTResponse := nftBoundaryDecode[nftBoundaryAmountResponse](t, "standard balance REST", balanceRESTRaw)
		require.Equal(t, want, nftBoundaryAmount(t, balanceCLIResponse.Amount), "CLI balance at height %d", height)
		require.Equal(t, want, balanceTyped.Amount, "typed gRPC balance at height %d", height)
		require.Equal(t, want, nftBoundaryAmount(t, balanceRESTResponse.Amount), "REST balance at height %d", height)
	}
}

func nftBoundaryAmount(t *testing.T, value string) uint64 {
	t.Helper()
	if value == "" {
		return 0
	}
	amount, err := strconv.ParseUint(value, 10, 64)
	require.NoError(t, err)
	return amount
}

func nftBoundaryCommittedHeight(t *testing.T, tx *harness.TxResult) int64 {
	t.Helper()
	require.NotNil(t, tx, "query-boundary parity requires a committed transaction")
	height := tx.HeightInt64()
	require.Positive(t, height, "query-boundary parity requires a positive committed height")
	return height
}

func nftBoundaryRESTPairAtHeight(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	height int64,
	rawPath string,
	encodedPath string,
) (json.RawMessage, json.RawMessage) {
	t.Helper()
	rawResponse, err := network.FullNodeRESTGetAtHeight(ctx, nil, step+"-raw-rest", rawPath, height)
	require.NoError(t, err)
	encodedResponse, err := network.FullNodeRESTGetAtHeight(ctx, nil, step+"-encoded-rest", encodedPath, height)
	require.NoError(t, err)
	return rawResponse, encodedResponse
}

func nftBoundaryRequireJSONEqual(t *testing.T, boundary string, expected, actual json.RawMessage) {
	t.Helper()
	require.JSONEq(t, string(expected), string(actual), boundary)
}

func nftBoundaryDecode[T any](t *testing.T, boundary string, raw json.RawMessage) T {
	t.Helper()
	var response T
	require.NoError(t, json.Unmarshal(raw, &response), boundary)
	return response
}

func nftBoundaryProjectJSONClass(class nftBoundaryJSONClass) (nftBoundaryClassProjection, error) {
	data, err := nftBoundaryProjectJSONData(class.Data)
	if err != nil {
		return nftBoundaryClassProjection{}, err
	}
	return nftBoundaryClassProjection{
		ID:          class.ID,
		Name:        class.Name,
		Symbol:      class.Symbol,
		Description: class.Description,
		URI:         class.URI,
		URIHash:     class.URIHash,
		Data:        data,
	}, nil
}

func nftBoundaryProjectTypedClass(class *upstreamnft.Class) (nftBoundaryClassProjection, error) {
	if class == nil {
		return nftBoundaryClassProjection{}, fmt.Errorf("standard class response is empty")
	}
	data, err := nftBoundaryProjectTypedData(class.Data)
	if err != nil {
		return nftBoundaryClassProjection{}, err
	}
	return nftBoundaryClassProjection{
		ID:          class.Id,
		Name:        class.Name,
		Symbol:      class.Symbol,
		Description: class.Description,
		URI:         class.Uri,
		URIHash:     class.UriHash,
		Data:        data,
	}, nil
}

func nftBoundaryProjectJSONNFT(nft nftBoundaryJSONNFT) (nftBoundaryNFTProjection, error) {
	data, err := nftBoundaryProjectJSONData(nft.Data)
	if err != nil {
		return nftBoundaryNFTProjection{}, err
	}
	return nftBoundaryNFTProjection{
		ClassID: nft.ClassID,
		ID:      nft.ID,
		URI:     nft.URI,
		URIHash: nft.URIHash,
		Data:    data,
	}, nil
}

func nftBoundaryProjectJSONData(data map[string]any) (nftBoundaryDataProjection, error) {
	if data == nil {
		return nftBoundaryDataProjection{}, nil
	}
	stringField := func(fields map[string]any, key string) (string, error) {
		value, ok := fields[key]
		if !ok || value == nil {
			return "", nil
		}
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("NFT data field %q has type %T, want string", key, value)
		}
		return text, nil
	}
	typeURL, err := stringField(data, "@type")
	if err != nil {
		return nftBoundaryDataProjection{}, err
	}
	fields := data
	if typeURL == "" {
		typeURL, err = stringField(data, "type")
		if err != nil {
			return nftBoundaryDataProjection{}, err
		}
		if wrapped, ok := data["value"]; ok {
			var valid bool
			fields, valid = wrapped.(map[string]any)
			if !valid {
				return nftBoundaryDataProjection{}, fmt.Errorf("NFT data field %q has type %T, want object", "value", wrapped)
			}
		}
	}
	name, err := stringField(fields, "name")
	if err != nil {
		return nftBoundaryDataProjection{}, err
	}
	description, err := stringField(fields, "description")
	if err != nil {
		return nftBoundaryDataProjection{}, err
	}
	imageURI, err := stringField(fields, "image_uri")
	if err != nil {
		return nftBoundaryDataProjection{}, err
	}
	return nftBoundaryDataProjection{
		TypeURL:     typeURL,
		Name:        name,
		Description: description,
		ImageURI:    imageURI,
	}, nil
}

func nftBoundaryProjectTypedNFT(nft *upstreamnft.NFT) (nftBoundaryNFTProjection, error) {
	if nft == nil {
		return nftBoundaryNFTProjection{}, fmt.Errorf("standard NFT response is empty")
	}
	data, err := nftBoundaryProjectTypedData(nft.Data)
	if err != nil {
		return nftBoundaryNFTProjection{}, err
	}
	return nftBoundaryNFTProjection{
		ClassID: nft.ClassId,
		ID:      nft.Id,
		URI:     nft.Uri,
		URIHash: nft.UriHash,
		Data:    data,
	}, nil
}

func nftBoundaryProjectTypedData(data *cdctypes.Any) (nftBoundaryDataProjection, error) {
	if data == nil {
		return nftBoundaryDataProjection{}, nil
	}
	if data.TypeUrl != nftBoundaryBasicDataTypeURL {
		return nftBoundaryDataProjection{}, fmt.Errorf("NFT data type URL %q, want %q", data.TypeUrl, nftBoundaryBasicDataTypeURL)
	}
	projection := nftBoundaryDataProjection{TypeURL: data.TypeUrl}
	remaining := data.Value
	for len(remaining) > 0 {
		fieldNumber, wireType, tagLength := protowire.ConsumeTag(remaining)
		if tagLength < 0 {
			return nftBoundaryDataProjection{}, fmt.Errorf("decode BasicNFTData tag: %v", protowire.ParseError(tagLength))
		}
		remaining = remaining[tagLength:]
		if wireType != protowire.BytesType || fieldNumber < 1 || fieldNumber > 3 {
			fieldLength := protowire.ConsumeFieldValue(fieldNumber, wireType, remaining)
			if fieldLength < 0 {
				return nftBoundaryDataProjection{}, fmt.Errorf("decode BasicNFTData field %d: %v", fieldNumber, protowire.ParseError(fieldLength))
			}
			remaining = remaining[fieldLength:]
			continue
		}
		value, valueLength := protowire.ConsumeString(remaining)
		if valueLength < 0 {
			return nftBoundaryDataProjection{}, fmt.Errorf("decode BasicNFTData field %d: %v", fieldNumber, protowire.ParseError(valueLength))
		}
		remaining = remaining[valueLength:]
		switch fieldNumber {
		case 1:
			projection.Name = value
		case 2:
			projection.Description = value
		case 3:
			projection.ImageURI = value
		}
	}
	return projection, nil
}

func TestNFTBoundaryEncodedPathSegmentPreservesIdentifierAndEscapesColon(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"panacea1creator%3Alifecycle.class",
		nftBoundaryEncodedPathSegment("panacea1creator:lifecycle.class"),
	)
}

func TestNFTBoundaryProjectionTreatsResolvedJSONAnyAndTypedAnyAsTheSameMetadata(t *testing.T) {
	t.Parallel()

	encoded := protowire.AppendTag(nil, 1, protowire.BytesType)
	encoded = protowire.AppendString(encoded, "Certificate #1")
	encoded = protowire.AppendTag(encoded, 2, protowire.BytesType)
	encoded = protowire.AppendString(encoded, "Lifecycle certificate")
	encoded = protowire.AppendTag(encoded, 3, protowire.BytesType)
	encoded = protowire.AppendString(encoded, "https://example.test/image.png")

	typed, err := nftBoundaryProjectTypedNFT(&upstreamnft.NFT{
		ClassId: "panacea1creator:lifecycle.class",
		Id:      "certificate.1",
		Uri:     "https://example.test/nft.json",
		UriHash: "sha256:abc",
		Data: &cdctypes.Any{
			TypeUrl: "/panacea.nft.v1.BasicNFTData",
			Value:   encoded,
		},
	})
	require.NoError(t, err)

	jsonProjection, err := nftBoundaryProjectJSONNFT(nftBoundaryJSONNFT{
		ClassID: "panacea1creator:lifecycle.class",
		ID:      "certificate.1",
		URI:     "https://example.test/nft.json",
		URIHash: "sha256:abc",
		Data: map[string]any{
			"@type":       "/panacea.nft.v1.BasicNFTData",
			"name":        "Certificate #1",
			"description": "Lifecycle certificate",
			"image_uri":   "https://example.test/image.png",
		},
	})
	require.NoError(t, err)
	require.Equal(t, jsonProjection, typed)

	legacyJSONProjection, err := nftBoundaryProjectJSONNFT(nftBoundaryJSONNFT{
		ClassID: "panacea1creator:lifecycle.class",
		ID:      "certificate.1",
		URI:     "https://example.test/nft.json",
		URIHash: "sha256:abc",
		Data: map[string]any{
			"type": "/panacea.nft.v1.BasicNFTData",
			"value": map[string]any{
				"name":        "Certificate #1",
				"description": "Lifecycle certificate",
				"image_uri":   "https://example.test/image.png",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, jsonProjection, legacyJSONProjection)
}
