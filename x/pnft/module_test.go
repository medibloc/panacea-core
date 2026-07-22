package pnft

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	gatewayruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"
)

func TestAppModuleDoesNotImplementGenesisLifecycle(t *testing.T) {
	_, hasGenesisLifecycle := any(NewAppModule(nil)).(module.HasABCIGenesis)
	require.False(t, hasGenesisLifecycle)
}

func TestAppModuleBasicDoesNotRegisterLegacyGatewayRoutes(t *testing.T) {
	mux := gatewayruntime.NewServeMux()
	AppModuleBasic{}.RegisterGRPCGatewayRoutes(client.Context{}, mux)

	for _, path := range []string{
		"/panacea/pnft/v2/denoms",
		"/panacea/pnft/v2/denoms/owners/panacea1owner",
		"/panacea/pnft/v2/denoms/denom",
		"/panacea/pnft/v2/denoms/denom/pnfts",
		"/panacea/pnft/v2/denoms/denom/owners/panacea1owner/pnfts",
		"/panacea/pnft/v2/denoms/denom/pnfts/pnft",
	} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			require.Equal(t, http.StatusNotFound, response.Code)
		})
	}
}
