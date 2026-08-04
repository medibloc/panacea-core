package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	feegrant "cosmossdk.io/x/feegrant"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeAuthzSendLimit       = int64(10)
	upgradeAuthzExecAmount      = int64(3)
	upgradeFeegrantSpendLimit   = int64(2_000_000)
	upgradeFeegrantConsumedFee  = int64(1_000_000)
	upgradeFeegrantSendAmount   = int64(1)
	upgradeAuthzGeneratedTxPath = "upgrade/authz-feegrant/authz-send.json"
	upgradeBankSendTypeURL      = "/cosmos.bank.v1beta1.MsgSend"
	upgradeSendAuthorizationURL = "/cosmos.bank.v1beta1.SendAuthorization"
	upgradeBasicAllowanceURL    = "/cosmos.feegrant.v1beta1.BasicAllowance"
)

type upgradeAuthzFeegrantFixture struct {
	GranterKeyName      string `json:"granter_key_name"`
	GranterAddress      string `json:"granter_address"`
	GranteeKeyName      string `json:"grantee_key_name"`
	GranteeAddress      string `json:"grantee_address"`
	AuthzRecipient      string `json:"authz_recipient"`
	FeegrantRecipient   string `json:"feegrant_recipient"`
	AuthzMessageTypeURL string `json:"authz_message_type_url"`
}

func (f upgradeAuthzFeegrantFixture) StateObjectIDs() []string {
	return []string{
		fmt.Sprintf("authz:%s/%s/%s", f.GranterAddress, f.GranteeAddress, f.AuthzMessageTypeURL),
		fmt.Sprintf("feegrant:%s/%s", f.GranterAddress, f.GranteeAddress),
	}
}

type upgradeAuthzFeegrantCheckpoint struct {
	Phase                    string                               `json:"phase"`
	RecordedAt               time.Time                            `json:"recorded_at"`
	Height                   int64                                `json:"height"`
	Observation              harness.UpgradeCheckpointObservation `json:"observation"`
	AuthzGrant               *authz.Grant                         `json:"authz_grant,omitempty"`
	FeegrantGrant            *feegrant.Grant                      `json:"feegrant_grant,omitempty"`
	AuthzRecipientBalance    string                               `json:"authz_recipient_balance"`
	FeegrantRecipientBalance string                               `json:"feegrant_recipient_balance"`
	GranterBalance           string                               `json:"granter_balance"`
	GranteeBalance           string                               `json:"grantee_balance"`
}

type upgradeAuthzFeegrantCoinEvidence struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type upgradeAuthzGrantEvidence struct {
	AuthorizationTypeURL string                             `json:"authorization_type_url"`
	SpendLimit           []upgradeAuthzFeegrantCoinEvidence `json:"spend_limit"`
	AllowList            []string                           `json:"allow_list"`
	Expiration           *time.Time                         `json:"expiration,omitempty"`
}

type upgradeFeegrantGrantEvidence struct {
	Granter          string                             `json:"granter"`
	Grantee          string                             `json:"grantee"`
	AllowanceTypeURL string                             `json:"allowance_type_url"`
	SpendLimit       []upgradeAuthzFeegrantCoinEvidence `json:"spend_limit"`
	Expiration       *time.Time                         `json:"expiration,omitempty"`
}

type upgradeAuthzGrantsClient interface {
	Grants(context.Context, *authz.QueryGrantsRequest, ...grpc.CallOption) (*authz.QueryGrantsResponse, error)
}

type upgradeFeegrantAllowanceClient interface {
	Allowance(context.Context, *feegrant.QueryAllowanceRequest, ...grpc.CallOption) (*feegrant.QueryAllowanceResponse, error)
	Allowances(context.Context, *feegrant.QueryAllowancesRequest, ...grpc.CallOption) (*feegrant.QueryAllowancesResponse, error)
}

func (c upgradeAuthzFeegrantCheckpoint) MarshalJSON() ([]byte, error) {
	authzGrant, err := projectUpgradeAuthzGrant(c.AuthzGrant)
	if err != nil {
		return nil, err
	}
	feegrantGrant, err := projectUpgradeFeegrantGrant(c.FeegrantGrant)
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Phase                    string                               `json:"phase"`
		RecordedAt               time.Time                            `json:"recorded_at"`
		Height                   int64                                `json:"height"`
		Observation              harness.UpgradeCheckpointObservation `json:"observation"`
		AuthzGrant               *upgradeAuthzGrantEvidence           `json:"authz_grant,omitempty"`
		FeegrantGrant            *upgradeFeegrantGrantEvidence        `json:"feegrant_grant,omitempty"`
		AuthzRecipientBalance    string                               `json:"authz_recipient_balance"`
		FeegrantRecipientBalance string                               `json:"feegrant_recipient_balance"`
		GranterBalance           string                               `json:"granter_balance"`
		GranteeBalance           string                               `json:"grantee_balance"`
	}{
		Phase:                    c.Phase,
		RecordedAt:               c.RecordedAt,
		Height:                   c.Height,
		Observation:              c.Observation,
		AuthzGrant:               authzGrant,
		FeegrantGrant:            feegrantGrant,
		AuthzRecipientBalance:    c.AuthzRecipientBalance,
		FeegrantRecipientBalance: c.FeegrantRecipientBalance,
		GranterBalance:           c.GranterBalance,
		GranteeBalance:           c.GranteeBalance,
	})
}

func projectUpgradeAuthzGrant(grant *authz.Grant) (*upgradeAuthzGrantEvidence, error) {
	if grant == nil {
		return nil, nil
	}
	if grant.Authorization == nil {
		return nil, errors.New("project authz grant: authorization is missing")
	}
	if grant.Authorization.GetTypeUrl() != upgradeSendAuthorizationURL {
		return nil, fmt.Errorf(
			"project authz grant: authorization type URL %q is not bank send",
			grant.Authorization.GetTypeUrl(),
		)
	}
	var authorization banktypes.SendAuthorization
	if err := authorization.Unmarshal(grant.Authorization.GetValue()); err != nil {
		return nil, fmt.Errorf("project authz grant: decode send authorization: %w", err)
	}

	return &upgradeAuthzGrantEvidence{
		AuthorizationTypeURL: grant.Authorization.GetTypeUrl(),
		SpendLimit:           projectUpgradeAuthzFeegrantCoins(authorization.GetSpendLimit()),
		AllowList:            append([]string{}, authorization.AllowList...),
		Expiration:           grant.Expiration,
	}, nil
}

func projectUpgradeFeegrantGrant(grant *feegrant.Grant) (*upgradeFeegrantGrantEvidence, error) {
	if grant == nil {
		return nil, nil
	}
	if grant.Allowance == nil {
		return nil, errors.New("project feegrant grant: allowance is missing")
	}
	if grant.Allowance.GetTypeUrl() != upgradeBasicAllowanceURL {
		return nil, fmt.Errorf(
			"project feegrant grant: allowance type URL %q is not basic allowance",
			grant.Allowance.GetTypeUrl(),
		)
	}
	var allowance feegrant.BasicAllowance
	if err := allowance.Unmarshal(grant.Allowance.GetValue()); err != nil {
		return nil, fmt.Errorf("project feegrant grant: decode basic allowance: %w", err)
	}

	return &upgradeFeegrantGrantEvidence{
		Granter:          grant.Granter,
		Grantee:          grant.Grantee,
		AllowanceTypeURL: grant.Allowance.GetTypeUrl(),
		SpendLimit:       projectUpgradeAuthzFeegrantCoins(allowance.GetSpendLimit()),
		Expiration:       allowance.Expiration,
	}, nil
}

func projectUpgradeAuthzFeegrantCoins(coins sdk.Coins) []upgradeAuthzFeegrantCoinEvidence {
	evidence := make([]upgradeAuthzFeegrantCoinEvidence, 0, len(coins))
	for _, coin := range coins {
		evidence = append(evidence, upgradeAuthzFeegrantCoinEvidence{
			Denom:  coin.Denom,
			Amount: coin.Amount.String(),
		})
	}
	return evidence
}

type upgradeAuthzFeegrantPreparation struct {
	Fixture          upgradeAuthzFeegrantFixture    `json:"fixture"`
	AuthzGrantTxHash string                         `json:"authz_grant_tx_hash"`
	FeegrantTxHash   string                         `json:"feegrant_grant_tx_hash"`
	Checkpoint       upgradeAuthzFeegrantCheckpoint `json:"checkpoint"`
}

func (p upgradeAuthzFeegrantPreparation) TxHashes() []string {
	return []string{p.AuthzGrantTxHash, p.FeegrantTxHash}
}

type upgradeAuthzFeegrantMutation struct {
	AuthzExecTxHash      string                         `json:"authz_exec_tx_hash"`
	FeegrantUseTxHash    string                         `json:"feegrant_use_tx_hash"`
	AuthzRevokeTxHash    string                         `json:"authz_revoke_tx_hash"`
	FeegrantRevokeTxHash string                         `json:"feegrant_revoke_tx_hash"`
	AfterAuthzExec       upgradeAuthzFeegrantCheckpoint `json:"after_authz_exec"`
	AfterFeegrantUse     upgradeAuthzFeegrantCheckpoint `json:"after_feegrant_use"`
	AfterRevocation      upgradeAuthzFeegrantCheckpoint `json:"after_revocation"`
}

func (m upgradeAuthzFeegrantMutation) TxHashes() []string {
	return []string{
		m.AuthzExecTxHash,
		m.FeegrantUseTxHash,
		m.AuthzRevokeTxHash,
		m.FeegrantRevokeTxHash,
	}
}

func prepareUpgradeAuthzFeegrant(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
) upgradeAuthzFeegrantPreparation {
	t.Helper()
	granter := buildAndFundNFTWallet(t, ctx, network, "upgrade-authz-feegrant-granter")
	grantee := buildAndFundNFTWallet(t, ctx, network, "upgrade-authz-feegrant-grantee")
	authzRecipient, err := network.BuildWallet(ctx, "upgrade-authz-recipient", "")
	require.NoError(t, err)
	feegrantRecipient, err := network.BuildWallet(ctx, "upgrade-feegrant-recipient", "")
	require.NoError(t, err)
	fixture := upgradeAuthzFeegrantFixture{
		GranterKeyName:      granter.KeyName(),
		GranterAddress:      granter.FormattedAddress(),
		GranteeKeyName:      grantee.KeyName(),
		GranteeAddress:      grantee.FormattedAddress(),
		AuthzRecipient:      authzRecipient.FormattedAddress(),
		FeegrantRecipient:   feegrantRecipient.FormattedAddress(),
		AuthzMessageTypeURL: upgradeBankSendTypeURL,
	}
	node := network.Chain.Validators[0]
	authzGrant, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-v221-authz-grant",
		node,
		fixture.GranterKeyName,
		"authz", "grant", fixture.GranteeAddress, "send",
		"--spend-limit", strconv.FormatInt(upgradeAuthzSendLimit, 10)+"umed",
		"--allow-list", fixture.AuthzRecipient,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	feegrantGrant, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-v221-feegrant-grant",
		node,
		fixture.GranterKeyName,
		"feegrant", "grant", fixture.GranterKeyName, fixture.GranteeAddress,
		"--spend-limit", strconv.FormatInt(upgradeFeegrantSpendLimit, 10)+"umed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	checkpoint, err := captureUpgradeAuthzFeegrantCheckpoint(
		ctx,
		network,
		fixture,
		"v2.2.1-preparation",
		false,
	)
	require.NoError(t, err)
	authzLimit, err := authzSendLimit(checkpoint.AuthzGrant, "umed")
	require.NoError(t, err)
	require.Equal(t, upgradeAuthzSendLimit, authzLimit)
	feeLimit, err := feegrantBasicLimit(checkpoint.FeegrantGrant, "umed")
	require.NoError(t, err)
	require.Equal(t, upgradeFeegrantSpendLimit, feeLimit)
	preparation := upgradeAuthzFeegrantPreparation{
		Fixture:          fixture,
		AuthzGrantTxHash: authzGrant.TxHash,
		FeegrantTxHash:   feegrantGrant.TxHash,
		Checkpoint:       checkpoint,
	}
	require.NoError(t, network.WriteArtifactJSON("upgrade/authz-feegrant/preparation.json", preparation))
	return preparation
}

func captureUpgradeAuthzFeegrantCheckpoint(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeAuthzFeegrantFixture,
	phase string,
	allowMissing bool,
) (upgradeAuthzFeegrantCheckpoint, error) {
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return upgradeAuthzFeegrantCheckpoint{}, errors.New("authz/feegrant checkpoint requires a full node")
	}
	if strings.TrimSpace(phase) == "" || strings.ContainsAny(phase, "/\\") {
		return upgradeAuthzFeegrantCheckpoint{}, fmt.Errorf("invalid authz/feegrant checkpoint phase %q", phase)
	}
	if strings.TrimSpace(fixture.GranterAddress) == "" ||
		strings.TrimSpace(fixture.GranteeAddress) == "" ||
		strings.TrimSpace(fixture.AuthzMessageTypeURL) == "" {
		return upgradeAuthzFeegrantCheckpoint{}, errors.New("authz/feegrant fixture identities are required")
	}
	fullNode := network.Chain.FullNodes[0]
	height, err := fullNode.Height(ctx)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, fmt.Errorf("query authz/feegrant checkpoint height: %w", err)
	}
	observation, err := network.CaptureUpgradeCheckpointObservation(ctx, "upgrade-authz-feegrant-"+phase, fullNode, height)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	pinnedCtx := metadata.AppendToOutgoingContext(ctx, "x-cosmos-block-height", strconv.FormatInt(height, 10))
	authzGrant, err := queryUpgradeAuthzGrant(
		pinnedCtx,
		authz.NewQueryClient(fullNode.GrpcConn),
		&authz.QueryGrantsRequest{
			Granter:    fixture.GranterAddress,
			Grantee:    fixture.GranteeAddress,
			MsgTypeUrl: fixture.AuthzMessageTypeURL,
		},
		allowMissing,
	)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	feegrantGrant, err := queryUpgradeFeegrantGrant(
		pinnedCtx,
		feegrant.NewQueryClient(fullNode.GrpcConn),
		&feegrant.QueryAllowanceRequest{
			Granter: fixture.GranterAddress,
			Grantee: fixture.GranteeAddress,
		},
		allowMissing,
	)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	balance := func(address string) (string, error) {
		amount, balanceErr := network.QueryFullNodeBalance(pinnedCtx, address, "umed")
		if balanceErr != nil {
			return "", balanceErr
		}
		return amount.String(), nil
	}
	authzRecipientBalance, err := balance(fixture.AuthzRecipient)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	feegrantRecipientBalance, err := balance(fixture.FeegrantRecipient)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	granterBalance, err := balance(fixture.GranterAddress)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	granteeBalance, err := balance(fixture.GranteeAddress)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	checkpoint := upgradeAuthzFeegrantCheckpoint{
		Phase:                    phase,
		RecordedAt:               observation.ObservedAt,
		Height:                   height,
		Observation:              observation,
		AuthzGrant:               authzGrant,
		FeegrantGrant:            feegrantGrant,
		AuthzRecipientBalance:    authzRecipientBalance,
		FeegrantRecipientBalance: feegrantRecipientBalance,
		GranterBalance:           granterBalance,
		GranteeBalance:           granteeBalance,
	}
	if err := network.WriteArtifactJSON("upgrade/authz-feegrant/checkpoints/"+phase+".json", checkpoint); err != nil {
		return checkpoint, fmt.Errorf("record authz/feegrant %s checkpoint: %w", phase, err)
	}
	return checkpoint, nil
}

func queryUpgradeAuthzGrant(
	ctx context.Context,
	client upgradeAuthzGrantsClient,
	request *authz.QueryGrantsRequest,
	allowMissing bool,
) (*authz.Grant, error) {
	response, err := client.Grants(ctx, request)
	if err != nil {
		if !allowMissing {
			return nil, fmt.Errorf("query authz grant: %w", err)
		}
		allRequest := *request
		allRequest.MsgTypeUrl = ""
		allResponse, allErr := client.Grants(ctx, &allRequest)
		if allErr != nil {
			return nil, fmt.Errorf("query authz grant: %w; verify missing grant with unfiltered query: %v", err, allErr)
		}
		if len(allResponse.GetGrants()) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf(
			"query authz grant: %w; unfiltered query returned %d grants, so absence is unproven",
			err,
			len(allResponse.GetGrants()),
		)
	}
	switch len(response.GetGrants()) {
	case 0:
		if !allowMissing {
			return nil, errors.New("authz grant is missing")
		}
		return nil, nil
	case 1:
		return response.GetGrants()[0], nil
	default:
		return nil, fmt.Errorf("authz query returned %d grants, want at most one", len(response.GetGrants()))
	}
}

func queryUpgradeFeegrantGrant(
	ctx context.Context,
	client upgradeFeegrantAllowanceClient,
	request *feegrant.QueryAllowanceRequest,
	allowMissing bool,
) (*feegrant.Grant, error) {
	if request == nil {
		return nil, errors.New("feegrant allowance request is required")
	}
	response, err := client.Allowance(ctx, request)
	if err == nil {
		grant := response.GetAllowance()
		if grant == nil && !allowMissing {
			return nil, errors.New("feegrant allowance response is empty")
		}
		return grant, nil
	}
	if !allowMissing {
		return nil, fmt.Errorf("query feegrant allowance: %w", err)
	}

	// feegrant v0.1.x maps a missing exact allowance to codes.Internal,
	// which is indistinguishable from real keeper failures by status alone.
	// Query the grantee's allowance list and filter the exact pair instead of
	// accepting an Internal error string as proof that the grant is absent.
	allResponse, allErr := client.Allowances(ctx, &feegrant.QueryAllowancesRequest{
		Grantee: request.Grantee,
	})
	if allErr != nil {
		return nil, fmt.Errorf(
			"query feegrant allowance: %w; verify missing allowance with grantee query: %v",
			err,
			allErr,
		)
	}
	var matching *feegrant.Grant
	for _, grant := range allResponse.GetAllowances() {
		if grant == nil || grant.Granter != request.Granter || grant.Grantee != request.Grantee {
			continue
		}
		if matching != nil {
			return nil, fmt.Errorf(
				"feegrant grantee query returned multiple allowances for %s/%s",
				request.Granter,
				request.Grantee,
			)
		}
		matching = grant
	}
	if matching != nil {
		return matching, nil
	}
	if pagination := allResponse.GetPagination(); pagination != nil && len(pagination.NextKey) != 0 {
		return nil, fmt.Errorf(
			"query feegrant allowance: %w; grantee query is paginated, so absence is unproven",
			err,
		)
	}
	return nil, nil
}

func captureAndValidateUpgradeAuthzFeegrantPreserved(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeAuthzFeegrantFixture,
	want upgradeAuthzFeegrantCheckpoint,
) (upgradeAuthzFeegrantCheckpoint, error) {
	got, err := captureUpgradeAuthzFeegrantCheckpoint(ctx, network, fixture, "post-upgrade-preservation", false)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	if !proto.Equal(want.AuthzGrant, got.AuthzGrant) {
		return got, errors.New("authz grant changed across upgrade")
	}
	if !proto.Equal(want.FeegrantGrant, got.FeegrantGrant) {
		return got, errors.New("feegrant allowance changed across upgrade")
	}
	if want.AuthzRecipientBalance != got.AuthzRecipientBalance ||
		want.FeegrantRecipientBalance != got.FeegrantRecipientBalance ||
		want.GranterBalance != got.GranterBalance ||
		want.GranteeBalance != got.GranteeBalance {
		return got, errors.New("authz/feegrant fixture balances changed across upgrade")
	}
	return got, nil
}

func mutateUpgradeAuthzFeegrant(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeAuthzFeegrantFixture,
) (upgradeAuthzFeegrantMutation, error) {
	node := network.Chain.Validators[0]
	generated, stderr, err := node.Exec(ctx, node.TxCommand(
		fixture.GranterKeyName,
		"bank", "send", fixture.GranterAddress, fixture.AuthzRecipient,
		strconv.FormatInt(upgradeAuthzExecAmount, 10)+"umed",
		"--generate-only",
		"--gas", "500000",
	), node.Chain.Config().Env)
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, fmt.Errorf("generate authz bank send: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	if err := node.WriteFile(ctx, generated, upgradeAuthzGeneratedTxPath); err != nil {
		return upgradeAuthzFeegrantMutation{}, fmt.Errorf("write generated authz transaction: %w", err)
	}
	if err := network.WriteArtifact(upgradeAuthzGeneratedTxPath, generated); err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	authzExec, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-post-authz-exec",
		node,
		fixture.GranteeKeyName,
		"authz", "exec", path.Join(node.HomeDir(), upgradeAuthzGeneratedTxPath),
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	afterAuthz, err := captureUpgradeAuthzFeegrantCheckpoint(ctx, network, fixture, "post-authz-exec", false)
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	authzRemaining, err := authzSendLimit(afterAuthz.AuthzGrant, "umed")
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	if authzRemaining != upgradeAuthzSendLimit-upgradeAuthzExecAmount {
		return upgradeAuthzFeegrantMutation{}, fmt.Errorf("authz remaining spend limit %d, want %d", authzRemaining, upgradeAuthzSendLimit-upgradeAuthzExecAmount)
	}
	if afterAuthz.AuthzRecipientBalance != strconv.FormatInt(upgradeAuthzExecAmount, 10) {
		return upgradeAuthzFeegrantMutation{}, fmt.Errorf("authz recipient balance %s, want %d", afterAuthz.AuthzRecipientBalance, upgradeAuthzExecAmount)
	}
	feegrantUse, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-post-feegrant-use",
		node,
		fixture.GranteeKeyName,
		"bank", "send", fixture.GranteeAddress, fixture.FeegrantRecipient,
		strconv.FormatInt(upgradeFeegrantSendAmount, 10)+"umed",
		"--gas", "200000",
		"--fees", strconv.FormatInt(upgradeFeegrantConsumedFee, 10)+"umed",
		"--fee-granter", fixture.GranterAddress,
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	afterFeegrant, err := captureUpgradeAuthzFeegrantCheckpoint(ctx, network, fixture, "post-feegrant-use", false)
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	feegrantRemaining, err := feegrantBasicLimit(afterFeegrant.FeegrantGrant, "umed")
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	if feegrantRemaining != upgradeFeegrantSpendLimit-upgradeFeegrantConsumedFee {
		return upgradeAuthzFeegrantMutation{}, fmt.Errorf("feegrant remaining spend limit %d, want %d", feegrantRemaining, upgradeFeegrantSpendLimit-upgradeFeegrantConsumedFee)
	}
	if afterFeegrant.FeegrantRecipientBalance != strconv.FormatInt(upgradeFeegrantSendAmount, 10) {
		return upgradeAuthzFeegrantMutation{}, fmt.Errorf("feegrant recipient balance %s, want %d", afterFeegrant.FeegrantRecipientBalance, upgradeFeegrantSendAmount)
	}
	authzRevoke, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-post-authz-revoke",
		node,
		fixture.GranterKeyName,
		"authz", "revoke", fixture.GranteeAddress, fixture.AuthzMessageTypeURL,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	feegrantRevoke, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-post-feegrant-revoke",
		node,
		fixture.GranterKeyName,
		"feegrant", "revoke", fixture.GranterKeyName, fixture.GranteeAddress,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	afterRevocation, err := captureUpgradeAuthzFeegrantCheckpoint(ctx, network, fixture, "post-upgrade-mutation", true)
	if err != nil {
		return upgradeAuthzFeegrantMutation{}, err
	}
	if afterRevocation.AuthzGrant != nil || afterRevocation.FeegrantGrant != nil {
		return upgradeAuthzFeegrantMutation{}, errors.New("authz or feegrant state remains after revocation")
	}
	mutation := upgradeAuthzFeegrantMutation{
		AuthzExecTxHash:      authzExec.TxHash,
		FeegrantUseTxHash:    feegrantUse.TxHash,
		AuthzRevokeTxHash:    authzRevoke.TxHash,
		FeegrantRevokeTxHash: feegrantRevoke.TxHash,
		AfterAuthzExec:       afterAuthz,
		AfterFeegrantUse:     afterFeegrant,
		AfterRevocation:      afterRevocation,
	}
	if err := network.WriteArtifactJSON("upgrade/authz-feegrant/mutation.json", mutation); err != nil {
		return mutation, err
	}
	return mutation, nil
}

func captureAndValidateUpgradeAuthzFeegrantPostRestart(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeAuthzFeegrantFixture,
	want upgradeAuthzFeegrantMutation,
) (upgradeAuthzFeegrantCheckpoint, error) {
	got, err := captureUpgradeAuthzFeegrantCheckpoint(ctx, network, fixture, "post-restart", true)
	if err != nil {
		return upgradeAuthzFeegrantCheckpoint{}, err
	}
	if got.AuthzGrant != nil || got.FeegrantGrant != nil {
		return got, errors.New("revoked authz or feegrant state reappeared after restart")
	}
	if got.AuthzRecipientBalance != want.AfterRevocation.AuthzRecipientBalance ||
		got.FeegrantRecipientBalance != want.AfterRevocation.FeegrantRecipientBalance {
		return got, errors.New("authz/feegrant recipient balances changed after restart")
	}
	return got, nil
}

func authzSendLimit(grant *authz.Grant, denom string) (int64, error) {
	if grant == nil || grant.Authorization == nil {
		return 0, errors.New("send authorization is missing")
	}
	if grant.Authorization.GetTypeUrl() != "/cosmos.bank.v1beta1.SendAuthorization" {
		return 0, fmt.Errorf("authorization type URL %q is not bank send", grant.Authorization.GetTypeUrl())
	}
	var authorization banktypes.SendAuthorization
	if err := authorization.Unmarshal(grant.Authorization.GetValue()); err != nil {
		return 0, fmt.Errorf("decode send authorization: %w", err)
	}
	amount := authorization.GetSpendLimit().AmountOf(denom)
	if !amount.IsInt64() {
		return 0, fmt.Errorf("send authorization %s amount does not fit int64", denom)
	}
	return amount.Int64(), nil
}

func feegrantBasicLimit(grant *feegrant.Grant, denom string) (int64, error) {
	if grant == nil || grant.GetAllowance() == nil {
		return 0, errors.New("feegrant allowance is missing")
	}
	if grant.GetAllowance().GetTypeUrl() != "/cosmos.feegrant.v1beta1.BasicAllowance" {
		return 0, fmt.Errorf("feegrant type URL %q is not basic allowance", grant.GetAllowance().GetTypeUrl())
	}
	var allowance feegrant.BasicAllowance
	if err := allowance.Unmarshal(grant.GetAllowance().GetValue()); err != nil {
		return 0, fmt.Errorf("decode basic fee allowance: %w", err)
	}
	amount := allowance.GetSpendLimit().AmountOf(denom)
	if !amount.IsInt64() {
		return 0, fmt.Errorf("feegrant %s amount does not fit int64", denom)
	}
	return amount.Int64(), nil
}

func TestUpgradeAuthzFeegrantLimitDecoders(t *testing.T) {
	t.Parallel()
	authzValue := &banktypes.SendAuthorization{SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("umed", 10))}
	authzAny, err := cdctypes.NewAnyWithValue(authzValue)
	require.NoError(t, err)
	authzLimit, err := authzSendLimit(&authz.Grant{Authorization: authzAny}, "umed")
	require.NoError(t, err)
	require.Equal(t, int64(10), authzLimit)

	feeValue := &feegrant.BasicAllowance{SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("umed", 2_000_000))}
	feeAny, err := cdctypes.NewAnyWithValue(feeValue)
	require.NoError(t, err)
	feeLimit, err := feegrantBasicLimit(&feegrant.Grant{Allowance: feeAny}, "umed")
	require.NoError(t, err)
	require.Equal(t, int64(2_000_000), feeLimit)
}

type upgradeAuthzGrantsClientFunc func(
	context.Context,
	*authz.QueryGrantsRequest,
	...grpc.CallOption,
) (*authz.QueryGrantsResponse, error)

func (f upgradeAuthzGrantsClientFunc) Grants(
	ctx context.Context,
	request *authz.QueryGrantsRequest,
	options ...grpc.CallOption,
) (*authz.QueryGrantsResponse, error) {
	return f(ctx, request, options...)
}

type upgradeFeegrantAllowanceClientFunc struct {
	allowance func(
		context.Context,
		*feegrant.QueryAllowanceRequest,
		...grpc.CallOption,
	) (*feegrant.QueryAllowanceResponse, error)
	allowances func(
		context.Context,
		*feegrant.QueryAllowancesRequest,
		...grpc.CallOption,
	) (*feegrant.QueryAllowancesResponse, error)
}

func (f upgradeFeegrantAllowanceClientFunc) Allowance(
	ctx context.Context,
	request *feegrant.QueryAllowanceRequest,
	options ...grpc.CallOption,
) (*feegrant.QueryAllowanceResponse, error) {
	return f.allowance(ctx, request, options...)
}

func (f upgradeFeegrantAllowanceClientFunc) Allowances(
	ctx context.Context,
	request *feegrant.QueryAllowancesRequest,
	options ...grpc.CallOption,
) (*feegrant.QueryAllowancesResponse, error) {
	return f.allowances(ctx, request, options...)
}

func TestQueryUpgradeAuthzGrantAllowsSDKExactMissingResponse(t *testing.T) {
	t.Parallel()

	calls := 0
	client := upgradeAuthzGrantsClientFunc(func(
		_ context.Context,
		request *authz.QueryGrantsRequest,
		_ ...grpc.CallOption,
	) (*authz.QueryGrantsResponse, error) {
		calls++
		switch calls {
		case 1:
			require.Equal(t, upgradeBankSendTypeURL, request.MsgTypeUrl)
			return nil, status.Error(codes.Unknown, "codespace authz code 2: authorization not found")
		case 2:
			require.Empty(t, request.MsgTypeUrl)
			return &authz.QueryGrantsResponse{}, nil
		default:
			require.FailNow(t, "unexpected authz query", "call %d", calls)
			return nil, nil
		}
	})

	grant, err := queryUpgradeAuthzGrant(context.Background(), client, &authz.QueryGrantsRequest{
		Granter:    "panacea1granter",
		Grantee:    "panacea1grantee",
		MsgTypeUrl: upgradeBankSendTypeURL,
	}, true)
	require.NoError(t, err)
	require.Nil(t, grant)
	require.Equal(t, 2, calls)
}

func TestQueryUpgradeAuthzGrantFailsClosedWhenMissingIsUnproven(t *testing.T) {
	t.Parallel()

	t.Run("unfiltered query still has grants", func(t *testing.T) {
		calls := 0
		client := upgradeAuthzGrantsClientFunc(func(
			_ context.Context,
			_ *authz.QueryGrantsRequest,
			_ ...grpc.CallOption,
		) (*authz.QueryGrantsResponse, error) {
			calls++
			if calls == 1 {
				return nil, status.Error(codes.Unknown, "codespace authz code 2: authorization not found")
			}
			return &authz.QueryGrantsResponse{Grants: []*authz.Grant{{}}}, nil
		})

		grant, err := queryUpgradeAuthzGrant(context.Background(), client, &authz.QueryGrantsRequest{
			Granter:    "panacea1granter",
			Grantee:    "panacea1grantee",
			MsgTypeUrl: upgradeBankSendTypeURL,
		}, true)
		require.ErrorContains(t, err, "absence is unproven")
		require.Nil(t, grant)
		require.Equal(t, 2, calls)
	})

	t.Run("missing is not allowed", func(t *testing.T) {
		calls := 0
		client := upgradeAuthzGrantsClientFunc(func(
			_ context.Context,
			_ *authz.QueryGrantsRequest,
			_ ...grpc.CallOption,
		) (*authz.QueryGrantsResponse, error) {
			calls++
			return nil, status.Error(codes.Unknown, "codespace authz code 2: authorization not found")
		})

		grant, err := queryUpgradeAuthzGrant(context.Background(), client, &authz.QueryGrantsRequest{
			Granter:    "panacea1granter",
			Grantee:    "panacea1grantee",
			MsgTypeUrl: upgradeBankSendTypeURL,
		}, false)
		require.ErrorContains(t, err, "authorization not found")
		require.Nil(t, grant)
		require.Equal(t, 1, calls)
	})
}

func TestQueryUpgradeFeegrantGrantConfirmsSDKExactMissingResponse(t *testing.T) {
	t.Parallel()

	client := upgradeFeegrantAllowanceClientFunc{
		allowance: func(
			_ context.Context,
			request *feegrant.QueryAllowanceRequest,
			_ ...grpc.CallOption,
		) (*feegrant.QueryAllowanceResponse, error) {
			require.Equal(t, "panacea1granter", request.Granter)
			require.Equal(t, "panacea1grantee", request.Grantee)
			return nil, status.Error(codes.Internal, "fee-grant not found: not found")
		},
		allowances: func(
			_ context.Context,
			request *feegrant.QueryAllowancesRequest,
			_ ...grpc.CallOption,
		) (*feegrant.QueryAllowancesResponse, error) {
			require.Equal(t, "panacea1grantee", request.Grantee)
			return &feegrant.QueryAllowancesResponse{}, nil
		},
	}

	grant, err := queryUpgradeFeegrantGrant(context.Background(), client, &feegrant.QueryAllowanceRequest{
		Granter: "panacea1granter",
		Grantee: "panacea1grantee",
	}, true)
	require.NoError(t, err)
	require.Nil(t, grant)
}

func TestQueryUpgradeFeegrantGrantDoesNotHideFallbackEvidenceOrFailure(t *testing.T) {
	t.Parallel()

	t.Run("matching grant remains", func(t *testing.T) {
		remaining := &feegrant.Grant{
			Granter: "panacea1granter",
			Grantee: "panacea1grantee",
		}
		client := upgradeFeegrantAllowanceClientFunc{
			allowance: func(
				context.Context,
				*feegrant.QueryAllowanceRequest,
				...grpc.CallOption,
			) (*feegrant.QueryAllowanceResponse, error) {
				return nil, status.Error(codes.Internal, "fee-grant not found: not found")
			},
			allowances: func(
				context.Context,
				*feegrant.QueryAllowancesRequest,
				...grpc.CallOption,
			) (*feegrant.QueryAllowancesResponse, error) {
				return &feegrant.QueryAllowancesResponse{Allowances: []*feegrant.Grant{remaining}}, nil
			},
		}

		grant, err := queryUpgradeFeegrantGrant(context.Background(), client, &feegrant.QueryAllowanceRequest{
			Granter: "panacea1granter",
			Grantee: "panacea1grantee",
		}, true)
		require.NoError(t, err)
		require.Same(t, remaining, grant)
	})

	t.Run("fallback query fails", func(t *testing.T) {
		client := upgradeFeegrantAllowanceClientFunc{
			allowance: func(
				context.Context,
				*feegrant.QueryAllowanceRequest,
				...grpc.CallOption,
			) (*feegrant.QueryAllowanceResponse, error) {
				return nil, status.Error(codes.Internal, "fee-grant not found: not found")
			},
			allowances: func(
				context.Context,
				*feegrant.QueryAllowancesRequest,
				...grpc.CallOption,
			) (*feegrant.QueryAllowancesResponse, error) {
				return nil, status.Error(codes.Unavailable, "query transport unavailable")
			},
		}

		grant, err := queryUpgradeFeegrantGrant(context.Background(), client, &feegrant.QueryAllowanceRequest{
			Granter: "panacea1granter",
			Grantee: "panacea1grantee",
		}, true)
		require.ErrorContains(t, err, "verify missing allowance")
		require.ErrorContains(t, err, "query transport unavailable")
		require.Nil(t, grant)
	})
}

func TestUpgradeAuthzFeegrantEvidenceMarshalsWithoutRawProtoAny(t *testing.T) {
	t.Parallel()
	expiration := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	authzValue := &banktypes.SendAuthorization{
		SpendLimit: sdk.NewCoins(
			sdk.NewInt64Coin("umed", 10),
			sdk.NewInt64Coin("uzzz", 11),
		),
		AllowList: []string{"panacea1recipient"},
	}
	authzAny, err := cdctypes.NewAnyWithValue(authzValue)
	require.NoError(t, err)
	feeValue := &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(
			sdk.NewInt64Coin("umed", 2_000_000),
			sdk.NewInt64Coin("uzzz", 3_000_000),
		),
		Expiration: &expiration,
	}
	feeAny, err := cdctypes.NewAnyWithValue(feeValue)
	require.NoError(t, err)
	checkpoint := upgradeAuthzFeegrantCheckpoint{
		Phase:      "v2.2.1-preparation",
		RecordedAt: time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC),
		Height:     42,
		AuthzGrant: &authz.Grant{
			Authorization: authzAny,
			Expiration:    &expiration,
		},
		FeegrantGrant: &feegrant.Grant{
			Granter:   "panacea1granter",
			Grantee:   "panacea1grantee",
			Allowance: feeAny,
		},
		AuthzRecipientBalance:    "3",
		FeegrantRecipientBalance: "1",
		GranterBalance:           "100",
		GranteeBalance:           "99",
	}

	for name, evidence := range map[string]any{
		"checkpoint": checkpoint,
		"preparation": upgradeAuthzFeegrantPreparation{
			Fixture:          upgradeAuthzFeegrantFixture{GranterAddress: "panacea1granter"},
			AuthzGrantTxHash: "AUTHZ",
			FeegrantTxHash:   "FEEGRANT",
			Checkpoint:       checkpoint,
		},
		"mutation": upgradeAuthzFeegrantMutation{
			AuthzExecTxHash:  "EXEC",
			AfterAuthzExec:   checkpoint,
			AfterFeegrantUse: checkpoint,
			AfterRevocation:  upgradeAuthzFeegrantCheckpoint{Phase: "post-upgrade-mutation"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := json.Marshal(evidence)
			require.NoError(t, err)
		})
	}

	raw, err := json.Marshal(checkpoint)
	require.NoError(t, err)
	var projected struct {
		AuthzGrant struct {
			AuthorizationTypeURL string `json:"authorization_type_url"`
			SpendLimit           []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"spend_limit"`
			AllowList  []string   `json:"allow_list"`
			Expiration *time.Time `json:"expiration"`
		} `json:"authz_grant"`
		FeegrantGrant struct {
			Granter          string `json:"granter"`
			Grantee          string `json:"grantee"`
			AllowanceTypeURL string `json:"allowance_type_url"`
			SpendLimit       []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"spend_limit"`
			Expiration *time.Time `json:"expiration"`
		} `json:"feegrant_grant"`
	}
	require.NoError(t, json.Unmarshal(raw, &projected))
	require.Equal(t, "/cosmos.bank.v1beta1.SendAuthorization", projected.AuthzGrant.AuthorizationTypeURL)
	require.Equal(t, []string{"panacea1recipient"}, projected.AuthzGrant.AllowList)
	require.Len(t, projected.AuthzGrant.SpendLimit, 2)
	require.Equal(t, "umed", projected.AuthzGrant.SpendLimit[0].Denom)
	require.Equal(t, "10", projected.AuthzGrant.SpendLimit[0].Amount)
	require.Equal(t, "uzzz", projected.AuthzGrant.SpendLimit[1].Denom)
	require.Equal(t, "11", projected.AuthzGrant.SpendLimit[1].Amount)
	require.NotNil(t, projected.AuthzGrant.Expiration)
	require.Equal(t, expiration, *projected.AuthzGrant.Expiration)
	require.Equal(t, "panacea1granter", projected.FeegrantGrant.Granter)
	require.Equal(t, "panacea1grantee", projected.FeegrantGrant.Grantee)
	require.Equal(t, "/cosmos.feegrant.v1beta1.BasicAllowance", projected.FeegrantGrant.AllowanceTypeURL)
	require.Len(t, projected.FeegrantGrant.SpendLimit, 2)
	require.Equal(t, "umed", projected.FeegrantGrant.SpendLimit[0].Denom)
	require.Equal(t, "2000000", projected.FeegrantGrant.SpendLimit[0].Amount)
	require.Equal(t, "uzzz", projected.FeegrantGrant.SpendLimit[1].Denom)
	require.Equal(t, "3000000", projected.FeegrantGrant.SpendLimit[1].Amount)
	require.NotNil(t, projected.FeegrantGrant.Expiration)
	require.Equal(t, expiration, *projected.FeegrantGrant.Expiration)
	require.NotContains(t, string(raw), `"value"`)
}
