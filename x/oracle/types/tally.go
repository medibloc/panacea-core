package types

import (
	"container/heap"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Vote interface {
		codec.ProtoMarshaler

		GetVoterAddress() string

		GetVoteOption() VoteOption

		GetConsensusValue() []byte
	}

	TallyKeeper interface {
		Tally(sdk.Context, sdk.Iterator, Vote) (*TallyResult, error)
	}
)

func NewTally() *Tally {
	return &Tally{
		OracleValidatorInfos: make(map[string]*OracleValidatorInfo),
		Yes:                  make(map[string]*ConsensusTally),
		No:                   sdk.ZeroInt(),
		Total:                sdk.ZeroInt(),
	}
}

// Add puts data to aggregate votes.
func (t *Tally) Add(vote Vote) error {
	oracleValidatorInfo, ok := t.OracleValidatorInfos[vote.GetVoterAddress()]
	if !ok {
		return fmt.Errorf("not found oracle. address: %s", vote.GetVoterAddress())
	}
	if !oracleValidatorInfo.IsPossibleVote() {
		return nil
	}

	bondedTokens := oracleValidatorInfo.BondedTokens
	switch vote.GetVoteOption() {
	case VOTE_OPTION_VALID:
		t.addYes(vote.GetConsensusValue(), bondedTokens)
	case VOTE_OPTION_INVALID:
		t.addNo(bondedTokens)
	default:
		return fmt.Errorf("unsupported voteOption. value: %s", vote.GetVoteOption())
	}
	return nil
}

// addYes defines to be divided and set by ConsensusValue
func (t *Tally) addYes(key []byte, amount sdk.Int) {
	if val, ok := t.Yes[string(key)]; ok {
		val.VotingAmount = val.VotingAmount.Add(amount)
	} else {
		t.Yes[string(key)] = &ConsensusTally{
			ConsensusKey: key,
			VotingAmount: amount,
		}
	}
}

func (t *Tally) addNo(amount sdk.Int) {
	t.No = t.No.Add(amount)
}

func (t *Tally) addTotal(amount sdk.Int) {
	t.Total = t.Total.Add(amount)
}

// calculateTotal calculates the total share based on the registered OracleValidatorInfo.
func (t *Tally) calculateTotal() {
	for _, info := range t.OracleValidatorInfos {
		t.addTotal(info.BondedTokens)
	}
}

// CalculateTallyResult calculates the voting result based on the received quorum, creates and returns a TallyResult.
func (t Tally) CalculateTallyResult(quorum sdk.Dec) *TallyResult {
	t.calculateTotal()

	tallyHeap := NewConsensusTallyMaxHeap()
	for _, tally := range t.Yes {
		heap.Push(&tallyHeap, tally)
	}

	tallyResult := NewTallyResult()

	if tallyHeap.Len() > 0 {
		maxTally := heap.Pop(&tallyHeap).(*ConsensusTally)
		tallyResult.Yes = maxTally.VotingAmount

		voteRate := maxTally.VotingAmount.ToDec().Quo(t.Total.ToDec())
		if voteRate.GTE(quorum) {
			tallyResult.ConsensusValue = maxTally.ConsensusKey
		}

		for tallyHeap.Len() > 0 {
			invalidYesTally := heap.Pop(&tallyHeap).(*ConsensusTally)
			tallyResult.InvalidYes = tallyResult.InvalidYes.Add(invalidYesTally.VotingAmount)
		}

	}

	tallyResult.No = t.No

	tallyResult.Total = t.Total

	return tallyResult
}

// ConsensusTallyMaxHeap implements by referring to the following site.
// https://pkg.go.dev/container/heap
type ConsensusTallyMaxHeap []*ConsensusTally

func NewConsensusTallyMaxHeap() ConsensusTallyMaxHeap {
	return make(ConsensusTallyMaxHeap, 0)
}

func (h ConsensusTallyMaxHeap) Len() int {
	return len(h)
}

func (h ConsensusTallyMaxHeap) Less(i, j int) bool {
	return h[i].VotingAmount.GT(h[j].VotingAmount)
}

func (h ConsensusTallyMaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *ConsensusTallyMaxHeap) Push(x interface{}) {
	tally := x.(*ConsensusTally)
	*h = append(*h, tally)
}

func (h *ConsensusTallyMaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	tally := old[n-1]
	old[n-1] = nil // avoid memory leak
	*h = old[0 : n-1]
	return tally
}

func (o *OracleValidatorInfo) IsPossibleVote() bool {
	return o.OracleActivated && !o.ValidatorJailed
}
