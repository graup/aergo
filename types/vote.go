package types

import "math/big"

const AergoSystem = "aergo.system"
const AergoName = "aergo.name"

func (vl VoteList) Len() int { return len(vl.Votes) }
func (vl VoteList) Less(i, j int) bool {
	result := new(big.Int).SetBytes(vl.Votes[i].Amount).Cmp(new(big.Int).SetBytes(vl.Votes[j].Amount))
	if result == -1 {
		return true
	} else if result == 0 {
		return new(big.Int).SetBytes(vl.Votes[i].Candidate[7:]).Cmp(new(big.Int).SetBytes(vl.Votes[j].Candidate[7:])) > 0
	}
	return false
}
func (vl VoteList) Swap(i, j int) { vl.Votes[i], vl.Votes[j] = vl.Votes[j], vl.Votes[i] }

func (v Vote) GetAmountBigInt() *big.Int {
	return new(big.Int).SetBytes(v.Amount)
}
