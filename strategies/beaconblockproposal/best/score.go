// Copyright © 2020 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package best

import (
	"context"
	"sort"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
)

// scoreBeaconBlockPropsal generates a score for a beacon block.
// The score is relative to the reward expected by proposing the block.
func scoreBeaconBlockProposal(ctx context.Context, name string, blockProposal *spec.BeaconBlock) float64 {
	if blockProposal == nil {
		return 0
	}
	if blockProposal.Body == nil {
		return 0
	}

	immediateAttestationScore := float64(0)
	attestationScore := float64(0)

	// Add attestation scores.
	for _, attestation := range blockProposal.Body.Attestations {
		inclusionDistance := float64(blockProposal.Slot - attestation.Data.Slot)
		attestationScore += float64(attestation.AggregationBits.Count()) * (float64(0.75) + float64(0.25)/float64(inclusionDistance))
		if inclusionDistance == 1 {
			immediateAttestationScore += float64(attestation.AggregationBits.Count()) * (float64(0.75) + float64(0.25)/float64(inclusionDistance))
		}
	}

	// Add slashing scores.
	// Slashing reward will be at most MAX_EFFECTIVE_BALANCE/WHISTLEBLOWER_REWARD_QUOTIENT,
	// which is 0.0625 Ether.
	// Individual attestation reward at 16K validators will be around 90,000 GWei, or .00009 Ether.
	// So we state that a single slashing event has the same weight as about 700 attestations.
	slashingWeight := float64(700)

	// Add proposer slashing scores.
	proposerSlashingScore := float64(len(blockProposal.Body.ProposerSlashings)) * slashingWeight

	// Add attester slashing scores.
	indicesSlashed := 0
	for i := range blockProposal.Body.AttesterSlashings {
		slashing := blockProposal.Body.AttesterSlashings[i]
		indicesSlashed += len(intersection(slashing.Attestation1.AttestingIndices, slashing.Attestation2.AttestingIndices))
	}
	attesterSlashingScore := slashingWeight * float64(indicesSlashed)

	log.Trace().
		Uint64("slot", blockProposal.Slot).
		Str("provider", name).
		Float64("immediate_attestations", immediateAttestationScore).
		Float64("attestations", attestationScore).
		Float64("proposer_slashings", proposerSlashingScore).
		Float64("attester_slashings", attesterSlashingScore).
		Float64("total", attestationScore+proposerSlashingScore+attesterSlashingScore).
		Msg("Scored block")

	return attestationScore + proposerSlashingScore + attesterSlashingScore
}

// intersection returns a list of items common between the two sets.
func intersection(set1 []uint64, set2 []uint64) []uint64 {
	sort.Slice(set1, func(i, j int) bool { return set1[i] < set1[j] })
	sort.Slice(set2, func(i, j int) bool { return set2[i] < set2[j] })
	res := make([]uint64, 0)

	set1Pos := 0
	set2Pos := 0
	for set1Pos < len(set1) && set2Pos < len(set2) {
		switch {
		case set1[set1Pos] < set2[set2Pos]:
			set1Pos++
		case set2[set2Pos] < set1[set1Pos]:
			set2Pos++
		default:
			res = append(res, set1[set1Pos])
			set1Pos++
			set2Pos++
		}
	}

	return res
}
