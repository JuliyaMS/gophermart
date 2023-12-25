package storage

func GetSQL(name string) string {
	var sql string
	switch n := name; n {
	case "Init":
		sql = "CREATE TABLE IF NOT EXISTS Users(id SERIAL PRIMARY KEY, Login varchar(100) NOT NULL, Password varchar(100) NOT NULL);"
		sql += "CREATE TABLE IF NOT EXISTS Orders(id SERIAL PRIMARY KEY, id_user SERIAL, Number varchar(100) NOT NULL, " +
			"Status varchar(100) NOT NULL, Accrual double precision, Uploaded_at TIMESTAMP NOT NULL);"
		sql += "CREATE TABLE IF NOT EXISTS Withdrawals(id SERIAL PRIMARY KEY, id_user SERIAL, Number varchar(100) NOT NULL, " +
			"Withdraw double precision NOT NULL, Uploaded_at TIMESTAMP NOT NULL);"
	case "CheckUser":
		sql = "SELECT Login FROM Users WHERE Login= $1;"
	case "AddUser":
		sql = "INSERT INTO Users(Login, Password) VALUES ($1, $2);"
	case "CheckPassword":
		sql = "SELECT Password FROM Users WHERE Login= $1;"
	case "CheckOrder":
		sql = "SELECT u.Login FROM Orders AS o INNER JOIN Users AS u ON u.id=o.id_user WHERE o.Number= $1;"
	case "AddOrder":
		sql = "INSERT INTO Orders(id_user, Number, Status, Accrual, Uploaded_at) SELECT u.id, $2,'NEW', 10, $3 " +
			"FROM Users as u WHERE u.Login=$1;"
	case "AddWithdraw":
		sql = "INSERT INTO Withdrawals(id_user, Number, Withdraw, Uploaded_at) SELECT u.id, $2, $3, $4 " +
			"FROM Users as u WHERE u.Login=$1;"
	case "GetOrders":
		sql = "SELECT o.Number, o.Status, o.Accrual, o.Uploaded_at FROM Orders AS o " +
			"INNER JOIN Users AS u ON u.id=o.id_user WHERE u.Login= $1;"
	case "GetWithdraws":
		sql = "SELECT w.Number, w.Withdraw, w.Uploaded_at FROM Withdrawals AS w " +
			"INNER JOIN Users AS u ON u.id=w.id_user WHERE u.Login= $1;"

	}
	return sql
}
