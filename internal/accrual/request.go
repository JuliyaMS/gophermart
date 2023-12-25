package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JuliyaMS/gophermart/internal/config"
	"github.com/avast/retry-go"
	"io"
	"net/http"
	"time"
)

type Response struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func send(number string) (*Response, error) {

	client := http.Client{Timeout: time.Duration(60) * time.Second}
	var resp Response

	URL := fmt.Sprintf("http://%s/api/orders/%s", config.AccrualURL, number)

	err := retry.Do(func() error {
		r, _ := http.NewRequest("GET", URL, nil)
		res, er := client.Do(r)

		if res != nil {
			var (
				data    []byte
				errResp error
			)
			defer res.Body.Close()
			if res.StatusCode == http.StatusOK {
				if data, errResp = io.ReadAll(res.Body); errResp != nil {
					return errResp
				}
				if errResp = json.Unmarshal(data, &resp); errResp != nil {
					return errResp
				}
			} else {
				return errors.New("status code is not OK")
			}
		}
		return er
	},
		retry.Attempts(3),
		retry.OnRetry(func(n uint, err error) {
			time.Sleep(time.Second * 2)
		}))

	if err != nil {
		return nil, err
	}
	return &resp, nil
}
