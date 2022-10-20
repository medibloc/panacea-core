package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/medibloc/panacea-core/v2/x/oracle/client/rest"

	"github.com/medibloc/panacea-core/v2/x/oracle/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.CmdUpgradeOracleProposal, rest.ProposalRESTHandler)
