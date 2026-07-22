package nft

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	upstreamnft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogogateway"
	gatewayruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/medibloc/panacea-core/v2/x/nft/keeper"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

func TestNFTRoutesRoundTripClassColonAndNFTDot(t *testing.T) {
	moduleKeeper, sdkContext, addressCodec, cdc := newModuleTestKeeper(t)
	sdkContext = sdkContext.WithBlockTime(time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC))
	creator, err := addressCodec.BytesToString(sdk.AccAddress(bytes.Repeat([]byte{112}, 20)))
	require.NoError(t, err)
	owner, err := addressCodec.BytesToString(sdk.AccAddress(bytes.Repeat([]byte{113}, 32)))
	require.NoError(t, err)

	classResponse, err := keeper.NewMsgServer(moduleKeeper).CreateClass(
		sdk.WrapSDKContext(sdkContext),
		&types.MsgCreateClassRequest{
			Creator:        creator,
			LocalClassId:   "rest.class",
			Name:           "REST Class",
			Symbol:         "REST",
			Description:    "REST route test class",
			Uri:            "https://example.test/rest-class.json",
			UriHash:        "sha256:" + strings.Repeat("a", 64),
			TransferPolicy: types.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE,
			Revocable:      true,
			MaxSupply:      1,
		},
	)
	require.NoError(t, err)
	classID := classResponse.ClassId
	require.Contains(t, classID, ":")
	encodedClassID := strings.Replace(classID, ":", "%3A", 1)
	require.Contains(t, encodedClassID, "%3A")

	const nftID = "nft.1"
	data, err := cdctypes.NewAnyWithValue(&types.BasicNFTData{
		Name:        "Gateway metadata",
		Description: "Metadata resolved through the REST gateway",
	})
	require.NoError(t, err)
	_, err = keeper.NewMsgServer(moduleKeeper).Mint(
		sdk.WrapSDKContext(sdkContext),
		&types.MsgMintRequest{
			ClassId:    classID,
			NftId:      nftID,
			Controller: creator,
			Recipient:  owner,
			Uri:        "https://example.test/nft.1.json",
			UriHash:    "sha256:" + strings.Repeat("b", 64),
			Data:       data,
		},
	)
	require.NoError(t, err)

	// Match server/api.New: the default grpc-gateway JSONPb cannot marshal
	// gogoproto stdtime fields used by Panacea lifecycle records.
	mux := gatewayruntime.NewServeMux(gatewayruntime.WithMarshalerOption(
		gatewayruntime.MIMEWildcard,
		&gateway.JSONPb{
			EmitDefaults: true,
			OrigName:     true,
			AnyResolver:  cdc.InterfaceRegistry(),
		},
	))
	require.NoError(t, upstreamnft.RegisterQueryHandlerServer(
		context.Background(),
		mux,
		keeper.NewStandardQueryServer(moduleKeeper),
	))
	require.NoError(t, types.RegisterQueryHandlerServer(
		context.Background(),
		mux,
		keeper.NewQueryServer(moduleKeeper),
	))

	for _, testCase := range []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "standard class",
			path:     "/cosmos/nft/v1beta1/classes/" + classID,
			expected: []string{classID},
		},
		{
			name:     "standard class with encoded colon",
			path:     "/cosmos/nft/v1beta1/classes/" + encodedClassID,
			expected: []string{classID},
		},
		{
			name: "standard nft",
			path: "/cosmos/nft/v1beta1/nfts/" + classID + "/" + nftID,
			expected: []string{
				classID,
				nftID,
				types.BasicNFTDataTypeURL,
				"Gateway metadata",
			},
		},
		{
			name:     "standard supply",
			path:     "/cosmos/nft/v1beta1/supply/" + classID,
			expected: []string{`"amount":"1"`},
		},
		{
			name:     "standard balance",
			path:     "/cosmos/nft/v1beta1/balance/" + owner + "/" + classID,
			expected: []string{`"amount":"1"`},
		},
		{
			name:     "standard owner",
			path:     "/cosmos/nft/v1beta1/owner/" + classID + "/" + nftID,
			expected: []string{owner},
		},
		{
			name:     "panacea class record",
			path:     "/panacea/nft/v1/classes/" + classID,
			expected: []string{classID},
		},
		{
			name:     "panacea class record with encoded colon",
			path:     "/panacea/nft/v1/classes/" + encodedClassID,
			expected: []string{classID},
		},
		{
			name: "panacea nft record",
			path: "/panacea/nft/v1/nfts/" + classID + "/" + nftID,
			expected: []string{
				classID,
				nftID,
				types.BasicNFTDataTypeURL,
				"Gateway metadata",
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			request = request.WithContext(sdk.WrapSDKContext(sdkContext))
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			for _, expected := range testCase.expected {
				require.Contains(t, response.Body.String(), expected)
			}
		})
	}
}
