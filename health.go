// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

func (c *clientImpl) Health(ctx context.Context) (*domain.HealthCheck, error) {
	params := &domain.GetHealthParams{}
	response, err := c.apiClient.GetHealthWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	if response.JSON503 != nil {
		//unhealthy server
		return response.JSON503, nil
	}

	return response.JSON200, nil
}
