package storage

import (
	"context"
	"errors"
	"github.com/JuliyaMS/gophermart/internal/config"
	"github.com/JuliyaMS/gophermart/internal/json"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"sort"
	"time"
)

type Storager interface {
	CheckConnection() error
	Init() error
	CheckUser(user string) error
	AddUser(user, password string) error
	CheckPassword(user string) (string, error)
	CheckOrder(number string) (string, error)
	AddOrder(login, order string) error
	GetOrders(login string) ([]json.Order, error)
	GetBalance(login string) (json.Balance, error)
	AddWithdraw(login, order string, sum float64) error
	GetWithdraws(login string) ([]json.Withdrawal, error)
}

type DB struct {
	conn       *pgx.Conn
	loggerPsql *zap.SugaredLogger
}

func NewConnectionDB(logger *zap.SugaredLogger) *DB {
	if config.DatabaseURI == "" {
		return nil
	}

	logger.Infow("Create context with timeout")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	logger.Infow("Connect to Database...")
	conn, err := pgx.Connect(ctx, config.DatabaseURI)
	if err != nil {
		logger.Error("Get error while connection to database:", err)
		return nil
	}

	logger.Infow("Success to create connection")
	return &DB{
		conn:       conn,
		loggerPsql: logger,
	}
}

func (db *DB) CheckConnection() error {

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	db.loggerPsql.Infow("Check connection to Database")
	err := db.conn.Ping(ctx)
	if err != nil {
		return err
	}
	db.loggerPsql.Infow("Success connection")
	return nil
}

func (db *DB) Init() error {

	db.loggerPsql.Infow("Start creation tables for metrics")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	sql := "CREATE TABLE IF NOT EXISTS Users(id SERIAL PRIMARY KEY, Login varchar(100) NOT NULL, Password varchar(100) NOT NULL);"
	sql += "CREATE TABLE IF NOT EXISTS Orders(id SERIAL PRIMARY KEY, id_user SERIAL, Number varchar(100) NOT NULL, " +
		"Status varchar(100) NOT NULL, Accrual double precision, Uploaded_at TIMESTAMP NOT NULL);"
	sql += "CREATE TABLE IF NOT EXISTS Withdrawals(id SERIAL PRIMARY KEY, id_user SERIAL, Number varchar(100) NOT NULL, " +
		"Withdraw double precision NOT NULL, Uploaded_at TIMESTAMP NOT NULL);"

	_, errEx := db.conn.Exec(ctx, sql)

	if errEx != nil {
		db.loggerPsql.Error("Error while create tables")
		return errEx
	}
	db.loggerPsql.Infow("Tables create successful")
	return nil
}

func (db *DB) CheckUser(user string) error {
	db.loggerPsql.Infow("Check user in table users")

	db.loggerPsql.Infow("Create sql string")
	sql := "SELECT Login FROM Users WHERE Login= $1;"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	db.loggerPsql.Infow("Execute request to check user")
	row := db.conn.QueryRow(ctx, sql, user)

	var login string
	if err := row.Scan(&login); err != nil {
		return err
	}

	return nil

}

func (db *DB) AddUser(user, password string) error {
	db.loggerPsql.Infow("Add new user if login not exist")

	db.loggerPsql.Infow("Create sql string")
	sql := "INSERT INTO Users(Login, Password) VALUES ($1, $2);"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	db.loggerPsql.Infow("Execute request to add user")
	_, errEx := db.conn.Exec(ctx, sql, user, password)

	if errEx != nil {
		db.loggerPsql.Error("Get error while add new user")
		return errEx
	}
	db.loggerPsql.Infow("Add new user successful")
	return nil
}

func (db *DB) CheckPassword(user string) (string, error) {
	db.loggerPsql.Infow("Check user and password in table users")

	db.loggerPsql.Infow("Create sql string")
	sql := "SELECT Password FROM Users WHERE Login= $1;"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	db.loggerPsql.Infow("Execute request to check password")
	row := db.conn.QueryRow(ctx, sql, user)

	var password string
	if err := row.Scan(&password); err != nil {
		return "", err
	}

	return password, nil

}

func (db *DB) CheckOrder(number string) (string, error) {
	db.loggerPsql.Infow("Check order in table Orders")

	db.loggerPsql.Infow("Create sql string")
	sql := "SELECT u.Login FROM Orders AS o INNER JOIN Users AS u ON u.id=o.id_user WHERE o.Number= $1;"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	db.loggerPsql.Infow("Execute request to check order")
	row := db.conn.QueryRow(ctx, sql, number)

	var login string
	if err := row.Scan(&login); err != nil {
		return "", err
	}

	return login, nil
}

func (db *DB) AddOrder(login, order string) error {
	db.loggerPsql.Infow("Add new order")

	db.loggerPsql.Infow("Create sql string")
	sql := "INSERT INTO Orders(id_user, Number, Status, Accrual, Uploaded_at) SELECT u.id, $2,'NEW', 0, $3 " +
		"FROM Users as u WHERE u.Login=$1;"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	db.loggerPsql.Infow("Execute request to add order")
	_, errEx := db.conn.Exec(ctx, sql, login, order, time.Now().Format(time.RFC3339))

	if errEx != nil {
		db.loggerPsql.Error("Get error while add new order")
		return errEx
	}
	db.loggerPsql.Infow("Add new order successful")
	return nil
}

func (db *DB) GetOrders(login string) ([]json.Order, error) {
	db.loggerPsql.Info("Get all orders for user:", login)

	var orders []json.Order

	db.loggerPsql.Infow("Create sql string")
	sql := "SELECT o.Number, o.Status, o.Accrual, o.Uploaded_at FROM Orders AS o " +
		"INNER JOIN Users AS u ON u.id=o.id_user WHERE u.Login= $1;"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	db.loggerPsql.Infow("Execute request to get orders")
	rows, err := db.conn.Query(ctx, sql, login)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	db.loggerPsql.Infow("Scan data from rows")
	for rows.Next() {
		var (
			number  string
			status  string
			accrual float64
			dt      time.Time
		)

		err = rows.Scan(&number, &status, &accrual, &dt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, json.Order{Number: number, Status: status, Accrual: accrual, UploadedAt: dt})
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})

	db.loggerPsql.Infow("Get all orders successful")
	return orders, nil
}

func (db *DB) GetBalance(login string) (json.Balance, error) {
	db.loggerPsql.Info("Get balance for user:", login)

	db.loggerPsql.Infow("Create sql string")
	sqlCurrent := "SELECT SUM(o.Accrual) FROM Orders AS o " +
		"INNER JOIN Users AS u ON u.id=o.id_user WHERE u.Login= $1 GROUP BY o.id_user;"

	sqlWithdrawn := "SELECT SUM(w.Withdraw) FROM Withdrawals AS w " +
		"INNER JOIN Users AS u ON u.id=w.id_user WHERE u.Login= $1 GROUP BY w.id_user;"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	var (
		sum       float64
		withdrawn float64
	)

	db.loggerPsql.Infow("Execute request to get current balance")
	rowCurrent := db.conn.QueryRow(ctx, sqlCurrent, login)

	if err := rowCurrent.Scan(&sum); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return json.Balance{}, err
		}
		sum = 0
	}

	db.loggerPsql.Infow("Execute request to get current withdraw")
	rowWithdrawn := db.conn.QueryRow(ctx, sqlWithdrawn, login)

	if err := rowWithdrawn.Scan(&withdrawn); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return json.Balance{}, err
		}
		withdrawn = 0
	}

	return json.Balance{Current: sum - withdrawn, Withdrawn: withdrawn}, nil

}

func (db *DB) AddWithdraw(login, order string, sum float64) error {
	db.loggerPsql.Infow("Add new withdraw")

	db.loggerPsql.Infow("Create sql string")
	sql := "INSERT INTO Withdrawals(id_user, Number, Withdraw, Uploaded_at) SELECT u.id, $2, $3, $4 " +
		"FROM Users as u WHERE u.Login=$1;"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	db.loggerPsql.Infow("Execute request to add withdraw")
	_, errEx := db.conn.Exec(ctx, sql, login, order, sum, time.Now().Format(time.RFC3339))

	if errEx != nil {
		db.loggerPsql.Error("Get error while add new withdraw")
		return errEx
	}
	db.loggerPsql.Infow("Add new withdraw successful")
	return nil
}

func (db *DB) GetWithdraws(login string) ([]json.Withdrawal, error) {
	db.loggerPsql.Info("Get all withdraws for user:", login)

	var orders []json.Withdrawal

	db.loggerPsql.Infow("Create sql string")
	sql := "SELECT w.Number, w.Withdraw, w.Uploaded_at FROM Withdrawals AS w " +
		"INNER JOIN Users AS u ON u.id=w.id_user WHERE u.Login= $1;"

	db.loggerPsql.Infow("Create sql context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	db.loggerPsql.Infow("Execute request to get withdraws")
	rows, err := db.conn.Query(ctx, sql, login)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	db.loggerPsql.Infow("Scan data from rows")
	for rows.Next() {
		var (
			number string
			sum    float64
			dt     time.Time
		)

		err = rows.Scan(&number, &sum, &dt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, json.Withdrawal{Order: number, Sum: sum, ProcessedAt: dt})
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].ProcessedAt.Before(orders[j].ProcessedAt)
	})

	db.loggerPsql.Infow("Get all withdraws successful")
	return orders, nil
}
