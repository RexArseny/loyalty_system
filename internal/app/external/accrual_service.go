package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type AccrualServiceClient struct {
	client  *http.Client
	logger  *zap.Logger
	address string
}

func NewAccrualServiceClient(logger *zap.Logger, address string) AccrualServiceClient {
	return AccrualServiceClient{
		client:  http.DefaultClient,
		address: address,
		logger:  logger,
	}
}

func (c *AccrualServiceClient) GetData(ctx context.Context, order *string) (*AccrualResponse, error) {
	response, err := c.client.Get(fmt.Sprintf("%s/api/orders/%s", c.address, *order))
	if err != nil {
		return nil, fmt.Errorf("can not make request to accrual service: %w", err)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("can not read accrual service response: %w", err)
	}
	defer func() {
		err = response.Body.Close()
		if err != nil {
			c.logger.Error("Can not close response body", zap.Error(err))
		}
	}()

	var result AccrualResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal accrual service response: %w", err)
	}

	return &result, nil
}
