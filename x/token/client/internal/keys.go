package internal

import (
	"fmt"

	"github.com/medibloc/panacea-core/x/token/types"
)

var (
	RouteToken      = fmt.Sprintf("custom/%s/token", types.RouterKey)
	RouteListTokens = fmt.Sprintf("custom/%s/listTokens", types.RouterKey)
)
