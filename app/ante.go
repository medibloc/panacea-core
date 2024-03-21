package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcante "github.com/cosmos/ibc-go/v7/modules/core/ante"
)

// setAnteHandler Reference github.com/cosmos/cosmos-sdk/x/auth/ante/ante.go
func (app *App) setAnteHandler(txConfig client.TxConfig) {
	app.SetAnteHandler(
		sdktypes.ChainAnteDecorators(
			ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
			ante.NewExtensionOptionsDecorator(nil),
			ante.NewValidateBasicDecorator(),
			ante.NewTxTimeoutHeightDecorator(),
			ante.NewValidateMemoDecorator(app.AccountKeeper),
			ante.NewConsumeGasForTxSizeDecorator(app.AccountKeeper),
			ante.NewDeductFeeDecorator(app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, nil),
			ante.NewSetPubKeyDecorator(app.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
			ante.NewValidateSigCountDecorator(app.AccountKeeper),
			ante.NewSigGasConsumeDecorator(app.AccountKeeper, ante.DefaultSigVerificationGasConsumer),
			ante.NewSigVerificationDecorator(app.AccountKeeper, txConfig.SignModeHandler()),
			ante.NewIncrementSequenceDecorator(app.AccountKeeper),
			ibcante.NewRedundantRelayDecorator(app.IBCKeeper),
		),
	)
}
