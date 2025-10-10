package credit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HttpCreditClient struct {
	BaseURL string
	Client  *http.Client
}

func NewHttpCreditClient(baseURL string) *HttpCreditClient {
	return &HttpCreditClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

type payload struct {
	OrgID  string `json:"org_id"`
	Amount int    `json:"amount"`
}

type balanceResponse struct {
	Balance int `json:"balance"`
}

// Debit sends POST /credit/debit
func (h *HttpCreditClient) Debit(ctx context.Context, orgId string, amount int) (int, error) {
	reqBody, _ := json.Marshal(payload{OrgID: orgId, Amount: amount})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/debit", h.BaseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return 0, fmt.Errorf("debit failed: %s", string(body))
	}

	var resp balanceResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return 0, err
	}

	return resp.Balance, nil
}
