package server

import (
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"

	"github.com/medibloc/panacea-core/version"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
)

//___________________________________________________________________________________

// PersistentPreRunEFn returns a PersistentPreRunE function for cobra
// that initailizes the passed in context with a properly configured
// logger and config object.
func PersistentPreRunEFn(context *server.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == version.VersionCmd.Name() {
			return nil
		}
		iCfg, err := interceptLoadConfig()
		if err != nil {
			return err
		}
		err = validateConfig(iCfg)
		if err != nil {
			return err
		}
		logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		logger, err = tmflags.ParseLogLevel(iCfg.LogLevel, logger, cfg.DefaultLogLevel())
		if err != nil {
			return err
		}
		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}
		logger = logger.With("module", "main")
		context.Config = iCfg
		context.Logger = logger
		return nil
	}
}

// If a new config is created, change some of the default tendermint settings
func interceptLoadConfig() (conf *cfg.Config, err error) {
	tmpConf := cfg.DefaultConfig()
	err = viper.Unmarshal(tmpConf)
	if err != nil {
		// TODO: Handle with #870
		panic(err)
	}
	rootDir := tmpConf.RootDir
	configFilePath := filepath.Join(rootDir, "config/config.toml")
	// Intercept only if the file doesn't already exist

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// the following parse config is needed to create directories
		conf, _ = tcmd.ParseConfig() // NOTE: ParseConfig() creates dir/files as necessary.
		conf.ProfListenAddress = "localhost:6060"
		conf.P2P.RecvRate = 5120000
		conf.P2P.SendRate = 5120000
		conf.TxIndex.IndexAllTags = true
		conf.Consensus.TimeoutCommit = 5 * time.Second
		cfg.WriteConfigFile(configFilePath, conf)
		// Fall through, just so that its parsed into memory.
	}

	if conf == nil {
		conf, err = tcmd.ParseConfig() // NOTE: ParseConfig() creates dir/files as necessary.
		if err != nil {
			panic(err)
		}
	}

	// create a default panacea config file if it does not exist
	panaceaConfigFilePath := filepath.Join(rootDir, "config/panacead.toml")
	if _, err := os.Stat(panaceaConfigFilePath); os.IsNotExist(err) {
		panaceaConf, _ := config.ParseConfig()
		config.WriteConfigFile(panaceaConfigFilePath, panaceaConf)
	}

	viper.SetConfigName("panacead")
	err = viper.MergeInConfig()

	return
}

// validate the config with the sdk's requirements.
func validateConfig(conf *cfg.Config) error {
	if !conf.Consensus.CreateEmptyBlocks {
		return errors.New("config option CreateEmptyBlocks = false is currently unsupported")
	}
	return nil
}

// add server commands
func AddCommands(
	ctx *server.Context, cdc *codec.Codec,
	rootCmd *cobra.Command,
	appCreator server.AppCreator, appExport server.AppExporter) {

	rootCmd.PersistentFlags().String("log_level", ctx.Config.LogLevel, "Log level")

	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tendermintCmd.AddCommand(
		server.ShowNodeIDCmd(ctx),
		server.ShowValidatorCmd(ctx),
		server.ShowAddressCmd(ctx),
		server.VersionCmd(ctx),
	)

	rootCmd.AddCommand(
		server.StartCmd(ctx, appCreator),
		server.UnsafeResetAllCmd(ctx),
		client.LineBreak,
		tendermintCmd,
		server.ExportCmd(ctx, cdc, appExport),
		client.LineBreak,
		version.VersionCmd,
	)
}

// TrapSignal traps SIGINT and SIGTERM and terminates the server correctly.
func TrapSignal(cleanupFunc func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		switch sig {
		case syscall.SIGTERM:
			defer cleanupFunc()
			os.Exit(128 + int(syscall.SIGTERM))
		case syscall.SIGINT:
			defer cleanupFunc()
			os.Exit(128 + int(syscall.SIGINT))
		}
	}()
}
