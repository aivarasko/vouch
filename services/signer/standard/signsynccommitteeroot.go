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

package standard

import (
	"context"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"go.opentelemetry.io/otel"
)

// SignSyncCommitteeRoots returns root signatures.
// This signs a beacon block root with the "sync committee" domain.
func (s *Service) SignSyncCommitteeRoots(ctx context.Context,
	accounts []e2wtypes.Account,
	epoch phase0.Epoch,
	root phase0.Root,
) (
	[]phase0.BLSSignature,
	error,
) {
	ctx, span := otel.Tracer("attestantio.vouch.services.signer.standard").Start(ctx, "SignSyncCommitteeRoots")
	defer span.End()

	if s.syncCommitteeDomainType == nil {
		return []phase0.BLSSignature{}, errors.New("no sync committee domain type available; cannot sign")
	}

	// Calculate the domain.
	domain, err := s.domainProvider.Domain(ctx, *s.syncCommitteeDomainType, epoch)
	if err != nil {
		return []phase0.BLSSignature{}, errors.Wrap(err, "failed to obtain signature domain for sync committee")
	}

	roots := make([]phase0.Root, len(accounts))
	for i := range accounts {
		roots[i] = root
	}

	sigs, err := s.signRootsByAccountType(ctx, accounts, roots, domain)
	if err != nil {
		return []phase0.BLSSignature{}, errors.Wrap(err, "failed to sign sync committee roots")
	}

	return sigs, nil
}
