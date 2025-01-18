package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AccrualServiceClient struct {
	client  *http.Client
	address string
}

func NewAccrualServiceClient(address string) AccrualServiceClient {
	return AccrualServiceClient{
		client:  http.DefaultClient,
		address: address,
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

	var result AccrualResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("can not unmarshal accrual service response: %w", err)
	}

	return &result, nil
}
