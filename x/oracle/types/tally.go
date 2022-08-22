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
		Tally(sdk.Context, sdk.Iterator, Vote, func(Vote) error) (*TallyResult, error)
	}
)

type Tally struct {
	OracleValidatorInfos map[string]*OracleValidatorInfo
	Yes                  map[string]*ConsensusTally
	No                   sdk.Int
	Total                sdk.Int
}

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
	// is not error. However, it is not included in the voting.
	if !oracleValidatorInfo.IsPossibleVote() {
		return nil
	}

	bondedTokens := oracleValidatorInfo.BondedTokens
	switch vote.GetVoteOption() {
	case VOTE_OPTION_YES:
		t.addYes(vote.GetConsensusValue(), bondedTokens)
	case VOTE_OPTION_NO:
		t.addNo(bondedTokens)
	default:
		return fmt.Errorf("unsupported voteOption. value: %s", vote.GetVoteOption())
	}
	return nil
}

// addYes defines to be divided and set by ConsensusValue
func (t *Tally) addYes(consensusValue []byte, amount sdk.Int) {
	if val, ok := t.Yes[string(consensusValue)]; ok {
		val.VotingAmount = val.VotingAmount.Add(amount)
	} else {
		t.Yes[string(consensusValue)] = &ConsensusTally{
			ConsensusValue: consensusValue,
			VotingAmount:   amount,
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
		if info.IsPossibleVote() {
			t.addTotal(info.BondedTokens)
		}
	}
}

// CalculateTallyResult calculates the voting result based on the received quorum, creates and returns a TallyResult.
func (t Tally) CalculateTallyResult(quorum sdk.Dec) *TallyResult {
	t.calculateTotal()

	tallyHeap := NewConsensusTallyMaxHeap()
	for _, tally := range t.Yes {
		tallyHeap.PushConsensusTally(tally)
	}

	tallyResult := NewTallyResult()

	if tallyHeap.Len() > 0 {
		maxTally := tallyHeap.PopConsensusTally()

		voteRate := maxTally.VotingAmount.ToDec().Quo(t.Total.ToDec())
		if voteRate.GTE(quorum) {
			tallyResult.Yes = maxTally.VotingAmount
			tallyResult.ConsensusValue = maxTally.ConsensusValue
		} else {
			tallyResult.AddInvalidYes(maxTally)
		}

		for tallyHeap.Len() > 0 {
			invalidYesTally := tallyHeap.PopConsensusTally()
			tallyResult.AddInvalidYes(invalidYesTally)
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

func (h *ConsensusTallyMaxHeap) PushConsensusTally(tally *ConsensusTally) {
	heap.Push(h, tally)
}

func (h *ConsensusTallyMaxHeap) PopConsensusTally() *ConsensusTally {
	return heap.Pop(h).(*ConsensusTally)
}

type OracleValidatorInfo struct {
	Address         string
	OracleActivated bool
	BondedTokens    sdk.Int
	ValidatorJailed bool
}

func (o *OracleValidatorInfo) IsPossibleVote() bool {
	return o.OracleActivated && !o.ValidatorJailed
}
