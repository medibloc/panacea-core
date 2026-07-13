package app_test

import (
	"testing"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
)

func TestRuntimeModulesHaveBootstrapBasics(t *testing.T) {
	panaceaapp.SetConfig()
	appOpts := viper.New()
	appOpts.Set(flags.FlagHome, t.TempDir())
	testApp := panaceaapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, appOpts)

	runtimeBasics := testApp.BasicManager()
	moduleNames := make([]string, 0, len(testApp.ModuleManager.Modules))
	for name := range testApp.ModuleManager.Modules {
		moduleNames = append(moduleNames, name)
		require.Contains(t, runtimeBasics, name, "runtime module %q has no derived basic", name)
		require.Contains(t, panaceaapp.ModuleBasics, name, "runtime module %q is missing from bootstrap basics", name)
	}

	runtimeBasicNames := make([]string, 0, len(runtimeBasics))
	for name := range runtimeBasics {
		runtimeBasicNames = append(runtimeBasicNames, name)
	}
	require.ElementsMatch(t, moduleNames, runtimeBasicNames)

	var bootstrapOnly []string
	for name := range panaceaapp.ModuleBasics {
		if _, ok := runtimeBasics[name]; !ok {
			bootstrapOnly = append(bootstrapOnly, name)
		}
	}
	require.ElementsMatch(t, []string{ibctm.ModuleName}, bootstrapOnly)
}
