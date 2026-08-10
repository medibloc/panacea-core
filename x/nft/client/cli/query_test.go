package cli_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/medibloc/panacea-core/v2/x/nft/client/cli"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

func TestClassRecordQueryCommandBindsRequestAndDecodesResponse(t *testing.T) {
	queryServer := classRecordQueryServer{}
	connection := newQueryTestConnection(t, &queryServer)

	clientCtx, cdc := newQueryTestClientContext(connection)

	output, err := clitestutil.ExecTestCLICmd(
		clientCtx,
		cli.NewClassRecordQueryCmd(),
		[]string{"class-id", "--output", "json"},
	)
	require.NoError(t, err)

	var response types.QueryClassRecordResponse
	require.NoError(t, cdc.UnmarshalJSON(output.Bytes(), &response))
	require.Equal(t, "class-id", response.ClassRecord.Policy.ClassId)
	require.Equal(t, uint64(7), response.ClassRecord.MintedCount)
}

func TestNFTRecordsQueryCommandBindsFiltersAndCursor(t *testing.T) {
	queryServer := nftRecordsQueryServer{}
	connection := newQueryTestConnection(t, &queryServer)
	clientCtx, cdc := newQueryTestClientContext(connection)

	output, err := clitestutil.ExecTestCLICmd(
		clientCtx,
		cli.NewNFTRecordsQueryCmd(),
		[]string{
			"--class-id", "class-id",
			"--owner", "owner",
			"--page-key", "AQI=",
			"--limit", "25",
			"--reverse",
			"--output", "json",
		},
	)
	require.NoError(t, err)

	var response types.QueryNFTRecordsResponse
	require.NoError(t, cdc.UnmarshalJSON(output.Bytes(), &response))
	require.Empty(t, response.NftRecords)
}

type classRecordQueryServer struct {
	types.UnimplementedQueryServer
}

func (classRecordQueryServer) ClassRecord(
	_ context.Context,
	request *types.QueryClassRecordRequest,
) (*types.QueryClassRecordResponse, error) {
	if request.ClassId != "class-id" {
		return nil, fmt.Errorf("unexpected class ID %q", request.ClassId)
	}
	return &types.QueryClassRecordResponse{
		ClassRecord: &types.ClassRecord{
			Policy:      &types.ClassPolicy{ClassId: request.ClassId},
			MintedCount: 7,
		},
	}, nil
}

type nftRecordsQueryServer struct {
	types.UnimplementedQueryServer
}

func (nftRecordsQueryServer) NFTRecords(
	_ context.Context,
	request *types.QueryNFTRecordsRequest,
) (*types.QueryNFTRecordsResponse, error) {
	switch {
	case request.ClassId != "class-id":
		return nil, fmt.Errorf("unexpected class ID %q", request.ClassId)
	case request.Owner != "owner":
		return nil, fmt.Errorf("unexpected owner %q", request.Owner)
	case request.Pagination == nil:
		return nil, fmt.Errorf("missing pagination")
	case string(request.Pagination.Key) != string([]byte{1, 2}):
		return nil, fmt.Errorf("unexpected page key %v", request.Pagination.Key)
	case request.Pagination.Limit != 25:
		return nil, fmt.Errorf("unexpected limit %d", request.Pagination.Limit)
	case !request.Pagination.Reverse:
		return nil, fmt.Errorf("reverse is false")
	}
	return &types.QueryNFTRecordsResponse{}, nil
}

func newQueryTestClientContext(connection *grpc.ClientConn) (client.Context, codec.Codec) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	return client.Context{}.
		WithCodec(cdc).
		WithInterfaceRegistry(interfaceRegistry).
		WithGRPCClient(connection).
		WithOutputFormat("json"), cdc
}

func newQueryTestConnection(t *testing.T, queryServer types.QueryServer) *grpc.ClientConn {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	types.RegisterQueryServer(server, queryServer)
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(func() {
		server.Stop()
		require.NoError(t, listener.Close())
	})

	connection, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, connection.Close())
	})
	return connection
}
