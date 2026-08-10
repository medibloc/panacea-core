package harness

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	ibcPhasePreUpgrade  = "pre-upgrade"
	ibcPanaceaToOsmosis = "panacea-to-osmosis"
	ibcOsmosisToPanacea = "osmosis-to-panacea"
)

// IBCChannelEndpoint captures one chain's immutable view of the transfer
// channel. Both endpoints are retained so a later upgrade check can prove the
// original client, connection, and channel remain in use.
type IBCChannelEndpoint struct {
	ChainID                  string `json:"chain_id"`
	CounterpartyChainID      string `json:"counterparty_chain_id"`
	ClientID                 string `json:"client_id"`
	CounterpartyClientID     string `json:"counterparty_client_id"`
	ConnectionID             string `json:"connection_id"`
	CounterpartyConnectionID string `json:"counterparty_connection_id"`
	ConnectionState          string `json:"connection_state"`
	PortID                   string `json:"port_id"`
	ChannelID                string `json:"channel_id"`
	CounterpartyChannelID    string `json:"counterparty_channel_id"`
	ChannelState             string `json:"channel_state"`
	Ordering                 string `json:"ordering"`
	Version                  string `json:"version"`
}

// IBCChannelHandshake is the exact same-history link shared by the pre- and
// post-upgrade phases.
type IBCChannelHandshake struct {
	Path    string             `json:"path"`
	Panacea IBCChannelEndpoint `json:"panacea"`
	Osmosis IBCChannelEndpoint `json:"osmosis"`
}

// IBCPacketObservation proves that the destination chain committed the
// packet's MsgRecvPacket. Height is the destination-chain commit height.
type IBCPacketObservation struct {
	Observed bool   `json:"observed"`
	ChainID  string `json:"chain_id"`
	Height   int64  `json:"height"`
}

// IBCAcknowledgementObservation proves that the source chain committed the
// successful ICS-20 acknowledgement. Acknowledgement is the base64 encoding
// of the raw acknowledgement bytes so the artifact remains lossless JSON.
type IBCAcknowledgementObservation struct {
	Observed        bool   `json:"observed"`
	ChainID         string `json:"chain_id"`
	Height          int64  `json:"height"`
	Acknowledgement string `json:"acknowledgement_base64"`
}

// IBCPacketLifecycleEvidence binds a send transaction to its exact packet,
// destination receive, and source acknowledgement.
type IBCPacketLifecycleEvidence struct {
	Direction          string                        `json:"direction"`
	SourceChainID      string                        `json:"source_chain_id"`
	DestinationChainID string                        `json:"destination_chain_id"`
	TxHash             string                        `json:"tx_hash"`
	TxHeight           int64                         `json:"tx_height"`
	Sequence           uint64                        `json:"sequence"`
	SourcePort         string                        `json:"source_port"`
	SourceChannel      string                        `json:"source_channel"`
	DestinationPort    string                        `json:"destination_port"`
	DestinationChannel string                        `json:"destination_channel"`
	Denom              string                        `json:"denom"`
	Amount             string                        `json:"amount"`
	PacketData         string                        `json:"packet_data_base64,omitempty"`
	TimeoutHeight      string                        `json:"timeout_height,omitempty"`
	TimeoutTimestamp   uint64                        `json:"timeout_timestamp,omitempty"`
	Recv               IBCPacketObservation          `json:"recv"`
	Ack                IBCAcknowledgementObservation `json:"ack"`
}

// IBCBalanceEvidence records the exact before/after assertion for a native or
// voucher balance. ExpectedAfter is stored explicitly so an artifact can be
// audited without re-running the test implementation.
type IBCBalanceEvidence struct {
	ChainID       string `json:"chain_id"`
	Address       string `json:"address"`
	Denom         string `json:"denom"`
	Before        string `json:"before"`
	After         string `json:"after"`
	ExpectedAfter string `json:"expected_after"`
}

// IBCEscrowBalanceEvidence proves the native token movement in the ICS-20
// channel escrow account itself. ExpectedDelta is signed: sends lock a
// positive amount, timeouts release a negative amount, and restart checks use
// zero. Recording the derived address makes the assertion independently
// auditable from chain state.
type IBCEscrowBalanceEvidence struct {
	Phase         string `json:"phase"`
	ChainID       string `json:"chain_id"`
	PortID        string `json:"port_id"`
	ChannelID     string `json:"channel_id"`
	Address       string `json:"address"`
	Denom         string `json:"denom"`
	Before        string `json:"before"`
	After         string `json:"after"`
	ExpectedDelta string `json:"expected_delta"`
	ExpectedAfter string `json:"expected_after"`
}

// IBCPreUpgradeTransferEvidence is the complete pre-upgrade ICS-20 proof on a
// single handshake: one successful packet in each direction and the four
// resulting source-native/destination-voucher balance assertions.
type IBCPreUpgradeTransferEvidence struct {
	Phase          string                       `json:"phase"`
	Channel        IBCChannelHandshake          `json:"channel"`
	Transfers      []IBCPacketLifecycleEvidence `json:"transfers"`
	FinalBalances  []IBCBalanceEvidence         `json:"final_balances"`
	EscrowBalances []IBCEscrowBalanceEvidence   `json:"escrow_balances"`
	DenomTraces    IBCDenomTraceSnapshot        `json:"denom_traces"`
}

// Validate rejects a partial or ambiguously paired channel snapshot.
func (h IBCChannelHandshake) Validate() error {
	var validationErrors []error
	if strings.TrimSpace(h.Path) == "" {
		validationErrors = append(validationErrors, errors.New("IBC path is required"))
	}
	if err := h.Panacea.validate("Panacea"); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := h.Osmosis.validate("Osmosis"); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if h.Panacea.ChainID == h.Osmosis.ChainID {
		validationErrors = append(validationErrors, errors.New("IBC endpoints must use distinct chain IDs"))
	}
	if h.Panacea.CounterpartyChainID != h.Osmosis.ChainID || h.Osmosis.CounterpartyChainID != h.Panacea.ChainID {
		validationErrors = append(validationErrors, errors.New("IBC endpoint counterparty chain IDs do not match"))
	}
	if h.Panacea.CounterpartyChannelID != h.Osmosis.ChannelID || h.Osmosis.CounterpartyChannelID != h.Panacea.ChannelID {
		validationErrors = append(validationErrors, errors.New("IBC endpoint counterparty channel IDs do not match"))
	}
	if h.Panacea.CounterpartyConnectionID != h.Osmosis.ConnectionID || h.Osmosis.CounterpartyConnectionID != h.Panacea.ConnectionID {
		validationErrors = append(validationErrors, errors.New("IBC endpoint counterparty connection IDs do not match"))
	}
	if h.Panacea.CounterpartyClientID != h.Osmosis.ClientID || h.Osmosis.CounterpartyClientID != h.Panacea.ClientID {
		validationErrors = append(validationErrors, errors.New("IBC endpoint counterparty client IDs do not match"))
	}
	return errors.Join(validationErrors...)
}

// Validate rejects incomplete evidence and, critically, evidence assembled
// from a different channel than the recorded handshake.
func (e IBCPreUpgradeTransferEvidence) Validate() error {
	var validationErrors []error
	if e.Phase != ibcPhasePreUpgrade {
		validationErrors = append(validationErrors, fmt.Errorf("IBC evidence phase = %q, want %q", e.Phase, ibcPhasePreUpgrade))
	}
	if err := e.Channel.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("validate IBC channel handshake: %w", err))
	}
	if len(e.Transfers) != 2 {
		validationErrors = append(validationErrors, fmt.Errorf("IBC evidence has %d transfers, want exactly 2", len(e.Transfers)))
	}

	transferByDirection := make(map[string]IBCPacketLifecycleEvidence, len(e.Transfers))
	for i, transfer := range e.Transfers {
		if _, exists := transferByDirection[transfer.Direction]; exists {
			validationErrors = append(validationErrors, fmt.Errorf("IBC transfer %d duplicates direction %q", i, transfer.Direction))
			continue
		}
		transferByDirection[transfer.Direction] = transfer
	}
	for _, expected := range []struct {
		direction   string
		source      IBCChannelEndpoint
		destination IBCChannelEndpoint
	}{
		{ibcPanaceaToOsmosis, e.Channel.Panacea, e.Channel.Osmosis},
		{ibcOsmosisToPanacea, e.Channel.Osmosis, e.Channel.Panacea},
	} {
		transfer, ok := transferByDirection[expected.direction]
		if !ok {
			validationErrors = append(validationErrors, fmt.Errorf("IBC evidence is missing %s transfer", expected.direction))
			continue
		}
		if err := transfer.validate(expected.source, expected.destination); err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("validate %s transfer: %w", expected.direction, err))
		}
	}

	if len(e.FinalBalances) != 4 {
		validationErrors = append(validationErrors, fmt.Errorf("IBC evidence has %d final balances, want exactly 4", len(e.FinalBalances)))
	}
	chainBalanceCounts := map[string]int{
		e.Channel.Panacea.ChainID: 0,
		e.Channel.Osmosis.ChainID: 0,
	}
	seenBalances := make(map[string]struct{}, len(e.FinalBalances))
	for i, balance := range e.FinalBalances {
		if err := balance.validate(); err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("validate IBC balance %d: %w", i, err))
		}
		if _, ok := chainBalanceCounts[balance.ChainID]; !ok {
			validationErrors = append(validationErrors, fmt.Errorf("IBC balance %d uses unknown chain %q", i, balance.ChainID))
		} else {
			chainBalanceCounts[balance.ChainID]++
		}
		key := strings.Join([]string{balance.ChainID, balance.Address, balance.Denom}, "\x00")
		if _, exists := seenBalances[key]; exists {
			validationErrors = append(validationErrors, fmt.Errorf("IBC balance %d duplicates chain/address/denom", i))
		}
		seenBalances[key] = struct{}{}
	}
	for _, chainID := range []string{e.Channel.Panacea.ChainID, e.Channel.Osmosis.ChainID} {
		if chainBalanceCounts[chainID] != 2 {
			validationErrors = append(validationErrors, fmt.Errorf("IBC evidence has %d balances for chain %q, want 2", chainBalanceCounts[chainID], chainID))
		}
	}
	if err := validateEscrowBalanceSet(e.EscrowBalances, e.Channel, e.Transfers, ibcPhasePreUpgrade); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.DenomTraces.validate(e.Channel, ibcPhasePreUpgrade); err != nil {
		validationErrors = append(validationErrors, err)
	}

	return errors.Join(validationErrors...)
}

func (e IBCPacketLifecycleEvidence) validate(source, destination IBCChannelEndpoint) error {
	var validationErrors []error
	if e.SourceChainID != source.ChainID {
		validationErrors = append(validationErrors, fmt.Errorf("source chain = %q, want %q", e.SourceChainID, source.ChainID))
	}
	if e.DestinationChainID != destination.ChainID {
		validationErrors = append(validationErrors, fmt.Errorf("destination chain = %q, want %q", e.DestinationChainID, destination.ChainID))
	}
	if strings.TrimSpace(e.TxHash) == "" {
		validationErrors = append(validationErrors, errors.New("send transaction hash is required"))
	}
	if e.TxHeight < 1 {
		validationErrors = append(validationErrors, fmt.Errorf("send transaction height = %d, want positive", e.TxHeight))
	}
	if e.Sequence == 0 {
		validationErrors = append(validationErrors, errors.New("packet sequence must be positive"))
	}
	if e.SourcePort != source.PortID || e.SourceChannel != source.ChannelID {
		validationErrors = append(validationErrors, fmt.Errorf("source packet endpoint = %s/%s, want %s/%s", e.SourcePort, e.SourceChannel, source.PortID, source.ChannelID))
	}
	if e.DestinationPort != destination.PortID || e.DestinationChannel != destination.ChannelID {
		validationErrors = append(validationErrors, fmt.Errorf("destination packet endpoint = %s/%s, want %s/%s", e.DestinationPort, e.DestinationChannel, destination.PortID, destination.ChannelID))
	}
	if strings.TrimSpace(e.Denom) == "" {
		validationErrors = append(validationErrors, errors.New("packet denomination is required"))
	}
	if !isPositiveInteger(e.Amount) {
		validationErrors = append(validationErrors, fmt.Errorf("packet amount = %q, want positive integer", e.Amount))
	}
	if !e.Recv.Observed {
		validationErrors = append(validationErrors, errors.New("destination receive was not observed"))
	}
	if e.Recv.ChainID != destination.ChainID {
		validationErrors = append(validationErrors, fmt.Errorf("receive chain = %q, want %q", e.Recv.ChainID, destination.ChainID))
	}
	if e.Recv.Height < 1 {
		validationErrors = append(validationErrors, fmt.Errorf("receive height = %d, want positive", e.Recv.Height))
	}
	if !e.Ack.Observed {
		validationErrors = append(validationErrors, errors.New("source acknowledgement was not observed"))
	}
	if e.Ack.ChainID != source.ChainID {
		validationErrors = append(validationErrors, fmt.Errorf("acknowledgement chain = %q, want %q", e.Ack.ChainID, source.ChainID))
	}
	if e.Ack.Height < 1 {
		validationErrors = append(validationErrors, fmt.Errorf("acknowledgement height = %d, want positive", e.Ack.Height))
	}
	if err := validateSuccessfulAcknowledgement(e.Ack.Acknowledgement); err != nil {
		validationErrors = append(validationErrors, err)
	}
	return errors.Join(validationErrors...)
}

func validateSuccessfulAcknowledgement(encoded string) error {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("decode acknowledgement base64: %w", err)
	}
	var acknowledgement struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(raw, &acknowledgement); err != nil {
		return fmt.Errorf("decode acknowledgement JSON: %w", err)
	}
	if acknowledgement.Error != "" {
		return fmt.Errorf("IBC acknowledgement reports error: %s", acknowledgement.Error)
	}
	if strings.TrimSpace(acknowledgement.Result) == "" {
		return errors.New("IBC acknowledgement has no success result")
	}
	if _, err := base64.StdEncoding.DecodeString(acknowledgement.Result); err != nil {
		return fmt.Errorf("decode acknowledgement result: %w", err)
	}
	return nil
}

func (e IBCBalanceEvidence) validate() error {
	var validationErrors []error
	for field, value := range map[string]string{
		"chain ID": e.ChainID,
		"address":  e.Address,
		"denom":    e.Denom,
	} {
		if strings.TrimSpace(value) == "" {
			validationErrors = append(validationErrors, fmt.Errorf("balance %s is required", field))
		}
	}
	for field, value := range map[string]string{
		"before":         e.Before,
		"after":          e.After,
		"expected after": e.ExpectedAfter,
	} {
		if !isNonNegativeInteger(value) {
			validationErrors = append(validationErrors, fmt.Errorf("balance %s = %q, want non-negative integer", field, value))
		}
	}
	if e.After != e.ExpectedAfter {
		validationErrors = append(validationErrors, fmt.Errorf("balance after = %q, want expected %q", e.After, e.ExpectedAfter))
	}
	return errors.Join(validationErrors...)
}

func (e IBCEscrowBalanceEvidence) validate(endpoint IBCChannelEndpoint, phase string) error {
	var validationErrors []error
	if e.Phase != phase || e.ChainID != endpoint.ChainID || e.PortID != endpoint.PortID || e.ChannelID != endpoint.ChannelID {
		validationErrors = append(validationErrors, errors.New("IBC escrow phase or channel identity is invalid"))
	}
	if strings.TrimSpace(e.Address) == "" || strings.TrimSpace(e.Denom) == "" {
		validationErrors = append(validationErrors, errors.New("IBC escrow address and denomination are required"))
	}
	before, beforeOK := new(big.Int).SetString(e.Before, 10)
	after, afterOK := new(big.Int).SetString(e.After, 10)
	delta, deltaOK := new(big.Int).SetString(e.ExpectedDelta, 10)
	expected, expectedOK := new(big.Int).SetString(e.ExpectedAfter, 10)
	if !beforeOK || !afterOK || !deltaOK || !expectedOK || before.Sign() < 0 || after.Sign() < 0 || expected.Sign() < 0 {
		validationErrors = append(validationErrors, errors.New("IBC escrow balances or delta are invalid integers"))
	} else {
		calculated := new(big.Int).Add(new(big.Int).Set(before), delta)
		if calculated.Cmp(expected) != 0 || after.Cmp(expected) != 0 {
			validationErrors = append(validationErrors, fmt.Errorf("IBC escrow balance after = %s, want %s from delta %s", e.After, calculated.String(), e.ExpectedDelta))
		}
	}
	return errors.Join(validationErrors...)
}

func validateEscrowBalanceSet(
	escrows []IBCEscrowBalanceEvidence,
	channel IBCChannelHandshake,
	transfers []IBCPacketLifecycleEvidence,
	phase string,
) error {
	if len(escrows) != 2 {
		return fmt.Errorf("IBC escrow evidence has %d entries, want 2", len(escrows))
	}
	amountByChain := make(map[string]string, len(transfers))
	for _, transfer := range transfers {
		amountByChain[transfer.SourceChainID] = transfer.Amount
	}
	seen := make(map[string]struct{}, 2)
	for _, escrow := range escrows {
		var endpoint IBCChannelEndpoint
		switch escrow.ChainID {
		case channel.Panacea.ChainID:
			endpoint = channel.Panacea
		case channel.Osmosis.ChainID:
			endpoint = channel.Osmosis
		default:
			return fmt.Errorf("IBC escrow uses unknown chain %q", escrow.ChainID)
		}
		if _, duplicate := seen[escrow.ChainID]; duplicate {
			return fmt.Errorf("IBC escrow chain %q is duplicated", escrow.ChainID)
		}
		seen[escrow.ChainID] = struct{}{}
		if err := escrow.validate(endpoint, phase); err != nil {
			return err
		}
		if escrow.ExpectedDelta != amountByChain[escrow.ChainID] {
			return fmt.Errorf("IBC escrow delta for %q = %s, want transfer amount %s", escrow.ChainID, escrow.ExpectedDelta, amountByChain[escrow.ChainID])
		}
	}
	return nil
}

func isPositiveInteger(value string) bool {
	number, ok := new(big.Int).SetString(value, 10)
	return ok && number.Sign() > 0 && number.String() == value
}

func isNonNegativeInteger(value string) bool {
	number, ok := new(big.Int).SetString(value, 10)
	return ok && number.Sign() >= 0 && number.String() == value
}

func (e IBCChannelEndpoint) validate(name string) error {
	var validationErrors []error
	required := map[string]string{
		"chain ID":                   e.ChainID,
		"counterparty chain ID":      e.CounterpartyChainID,
		"client ID":                  e.ClientID,
		"counterparty client ID":     e.CounterpartyClientID,
		"connection ID":              e.ConnectionID,
		"counterparty connection ID": e.CounterpartyConnectionID,
		"channel ID":                 e.ChannelID,
		"counterparty channel ID":    e.CounterpartyChannelID,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			validationErrors = append(validationErrors, fmt.Errorf("%s %s is required", name, field))
		}
	}
	if !strings.HasPrefix(e.ClientID, "07-tendermint-") {
		validationErrors = append(validationErrors, fmt.Errorf("%s client %q is not 07-tendermint", name, e.ClientID))
	}
	if e.PortID != "transfer" {
		validationErrors = append(validationErrors, fmt.Errorf("%s port = %q, want transfer", name, e.PortID))
	}
	if e.Version != "ics20-1" {
		validationErrors = append(validationErrors, fmt.Errorf("%s channel version = %q, want ics20-1", name, e.Version))
	}
	if !ibcStateEquals(e.ConnectionState, "OPEN") {
		validationErrors = append(validationErrors, fmt.Errorf("%s connection state = %q, want open", name, e.ConnectionState))
	}
	if !ibcStateEquals(e.ChannelState, "OPEN") {
		validationErrors = append(validationErrors, fmt.Errorf("%s channel state = %q, want open", name, e.ChannelState))
	}
	if !ibcStateEquals(e.Ordering, "UNORDERED") {
		validationErrors = append(validationErrors, fmt.Errorf("%s channel ordering = %q, want unordered", name, e.Ordering))
	}
	return errors.Join(validationErrors...)
}

func ibcStateEquals(value, want string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	for _, prefix := range []string{"STATE_", "ORDER_", "ORDERING_"} {
		normalized = strings.TrimPrefix(normalized, prefix)
	}
	return normalized == want
}
