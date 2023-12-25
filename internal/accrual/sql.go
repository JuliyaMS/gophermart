package accrual

import (
	"context"
	"errors"
	"github.com/JuliyaMS/gophermart/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBAccrual struct {
	conn  *pgxpool.Pool
	limit int
}

func NewConnectionDBAccrual() (*DBAccrual, error) {
	if config.DatabaseURI == "" {
		return nil, errors.New("DatabaseURI is empty")
	}

	conn, err := pgxpool.New(context.Background(), config.DatabaseURI)
	if err != nil {
		return nil, err
	}

	return &DBAccrual{
		conn:  conn,
		limit: 10,
	}, nil
}

func (db *DBAccrual) GetNeedOrders() ([]string, error) {
	sql := "SELECT o.Number FROM Orders AS o WHERE o.Status IN ('NEW','PROCESSING') ORDER BY o.Uploaded_at ASC LIMIT $1;"

	var orders []string

	rows, err := db.conn.Query(context.Background(), sql, db.limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			number string
		)
		err = rows.Scan(&number)
		if err != nil {
			return nil, err
		}
		orders = append(orders, number)
	}
	return orders, nil
}

func (db *DBAccrual) UpdateOrders(resp *Response) error {
	sql := "UPDATE Orders SET Status = $1, Accrual=$2 WHERE Number = $3;"

	_, errEx := db.conn.Exec(context.Background(), sql, resp.Status, resp.Accrual, resp.Order)

	if errEx != nil {
		return errEx
	}
	return nil
}
