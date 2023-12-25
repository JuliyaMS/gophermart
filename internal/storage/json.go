package storage

import (
	"encoding/json"
	"io"
	"time"
)

type AuthData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}

func Decode(data any, r io.Reader) error {
	dec := json.NewDecoder(r)

	if err := dec.Decode(data); err != nil {
		return err
	}
	return nil
}

func Encode(data any, w io.Writer) error {
	enc := json.NewEncoder(w)

	if err := enc.Encode(data); err != nil {
		return err
	}
	return nil
}
