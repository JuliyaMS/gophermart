package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JuliyaMS/gophermart/internal/config"
	"github.com/JuliyaMS/gophermart/internal/logger"
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

	log := logger.NewLogger()

	client := http.Client{}
	var resp Response

	URL := fmt.Sprintf("%s/api/orders/%s", config.AccrualURL, number)

	log.Info("Send data to address: ", URL)
	err := retry.Do(func() error {
		r, _ := http.NewRequest("GET", URL, nil)
		res, er := client.Do(r)

		if res != nil {
			var (
				data    []byte
				errResp error
			)
			defer res.Body.Close()
			log.Infow("Check status ...")
			if res.StatusCode == http.StatusOK {
				log.Infow("Read data ...")
				if data, errResp = io.ReadAll(res.Body); errResp != nil {
					return errResp
				}
				log.Infow("Decode data ...")
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
			log.Infow("Retry send data ...")
			time.Sleep(time.Second * 2)
		}))

	if err != nil {
		log.Infow("Get error")
		return nil, err
	}
	return &resp, nil
}
