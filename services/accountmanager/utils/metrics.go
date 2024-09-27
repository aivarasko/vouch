// Copyright © 2024 Attestant Limited.
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

package utils

import (
	"context"
	"errors"

	"github.com/attestantio/vouch/services/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var accountManagerAccounts *prometheus.GaugeVec

// RegisterMetrics registers metrics if appropriate for the monitor type.
func RegisterMetrics(ctx context.Context, monitor metrics.Service) error {
	if monitor == nil {
		// No monitor.
		return nil
	}
	if monitor.Presenter() == "prometheus" {
		return registerPrometheusMetrics(ctx)
	}
	return nil
}

func registerPrometheusMetrics(_ context.Context) error {
	accountManagerAccounts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "vouch",
		Subsystem: "accountmanager",
		Name:      "accounts_total",
		Help:      "The number of accounts managed by Vouch.",
	}, []string{"state"})
	if err := prometheus.Register(accountManagerAccounts); err != nil {
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if ok := errors.As(err, &alreadyRegisteredError); ok {
			accountManagerAccounts = alreadyRegisteredError.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return err
		}
	}

	return nil
}

// MonitorAccounts is used to update the account manager metrics.
func MonitorAccounts(state string, count uint64) {
	if accountManagerAccounts == nil {
		return
	}
	accountManagerAccounts.WithLabelValues(state).Set(float64(count))
}
