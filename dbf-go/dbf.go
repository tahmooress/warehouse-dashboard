package dbf

import (
	"database/sql"
	"fmt"
)

//RunDB is a function to return db connection
func RunDB() (*sql.DB, error) {
	stmt := setDbInfo()
	db, err := sql.Open("postgres", stmt)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

const (
	host     = "localhost"
	port     = 5432
	user     = "mac"
	password = "900587101"
	dbname   = "server"
)

func setDbInfo() string {
	return fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
}

//Env is a wrapper type of a DB to pass to the handlers
type Env struct {
	DB *sql.DB
}

const (
	UserQuery              = `SELECT password, accessibility FROM users WHERE user_name=$1`
	CreateNewUser          = `INSERT INTO users(user_name, password, accessibility) VALUES($1,$2,$3)`
	InsertNewMotor         = `INSERT INTO motors (pelak_number, color, body_number, model_name, model_year) VALUES($1,$2,$3,$4,$5)`
	InsertNewMotorFromList = `INSERT INTO motors (pelak_number, color, body_number, model_name, model_year, list_id) VALUES($1,$2,$3,$4,$5, $6)`
	InsertList             = `INSERT INTO entry_list(provider_name, date, total_amount) VALUES($1, $2, $3) RETURNING list_id`
)

//UpdateReceive is a function to decide to Update table
func UpdateReceive(shop string) string {
	var s string
	switch shop {
	case "shop_a":
		s = "accounts_receivable_a"
	case "shop_b":
		s = "accounts_receivable_b"
	case "shop_c":
		s = "accounts_receivable_c"
	case "warehouse":
		s = "accounts_receivable_warehouse"
	}
	return fmt.Sprintf("UPDATE %s SET status=true WHERE factor_number=$1 AND date=$2", s)
}

//UpdatePayable is a function to decide to Update table
func UpdatePayable(shop string) string {
	var s string
	switch shop {
	case "shop_a":
		s = "accounts_payable_a"
	case "shop_b":
		s = "accounts_payable_b"
	case "shop_c":
		s = "accounts_payable_c"
	case "warehouse":
		s = "accounts_payable_warehouse"
	}
	return fmt.Sprintf("UPDATE %s SET status=true WHERE factor_number=$1 AND date=$2", s)
}

//InsertPayable is a function to decide to insert in which table
func InsertPayable(shop string) string {
	var s string
	switch shop {
	case "shop_a":
		s = "accounts_payable_a"
	case "shop_b":
		s = "accounts_payable_b"
	case "shop_c":
		s = "accounts_payable_c"
	case "warehouse":
		s = "accounts_payable_warehouse"
	}
	return fmt.Sprintf("INSERT INTO %s (factor_number, price, date) VALUES($1, $2, $3)", s)
}

//InsertRecievable is a ...
func InsertRecievable(shop string) string {
	var s string
	switch shop {
	case "shop_a":
		s = "accounts_receivable_a"
	case "shop_b":
		s = "accounts_receivable_b"
	case "shop_c":
		s = "accounts_receivable_c"
	case "warehouse":
		s = "accounts_receivable_warehouse"
	}
	return fmt.Sprintf("INSERT INTO %s (factor_number, price, date) VALUES($1, $2, $3)", s)
}

//InsertSellFactor is a function to decide to insert in which table
func InsertSellFactor(shop string) string {
	var s string
	switch shop {
	case "shop_a":
		s = "sell_factor_a"
	case "shop_b":
		s = "sell_factor_b"
	case "shop_c":
		s = "sell_factor_c"
	case "warehouse":
		s = "sell_factor_warehouse"
	}
	return fmt.Sprintf("INSERT INTO %s (factor_number, pelak_number, price, date, buyer_name, buyer_last_name, buyer_national_code, buyer_mobile) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", s)
}

//InsertBuyFactor is a function to decide to insert in which table
func InsertBuyFactor(shop string) string {
	var s string
	switch shop {
	case "shop_a":
		s = "buy_factor_a"
	case "shop_b":
		s = "buy_factor_b"
	case "shop_c":
		s = "buy_factor_c"
	case "warehouse":
		s = "warehouse"
	}
	return fmt.Sprintf("INSERT INTO %s(factor_number, pelak_number, price, date, customer_name, customer_last_name, customer_national_code, customer_mobile) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", s)
}

//InsertIntoShop decide to insert into which shop
func InsertIntoShop(s string) string {
	return fmt.Sprintf("INSERT INTO %s (pelak_number, buy_factor) VALUES($1,$2)", s)
}

//InsertShopEntry is a function to insert motors from entry list to shop
func InsertShopEntry(s string) string {
	return fmt.Sprintf("INSERT INTO %s (pelak_number) VALUES ($1)", s)
}

//UpdateShopSellFactor is a function that specify the destination shop to update and return the sql query for item that buy from providers
func UpdateShopSellFactor(s string) string {
	return fmt.Sprintf("UPDATE %s SET sell_factor=$1, stock=$2 WHERE pelak_number=$3", s)
}

//RemoveSrc is a...
func RemoveSrc(src string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE pelak_number=$1 RETURNING buy_factor, stock", src)
}

//StockLookUp is a ...s
func StockLookUp(s string) string {
	return fmt.Sprintf("SELECT motors.pelak_number, color, model_name, buy_factor FROM motors INNER JOIN %s ON motors.pelak_number = %s.pelak_number WHERE %s.stock=true", s, s, s)
}

// SaleHistoryLookUp is a...
func SaleHistoryLookUp(shop string) string {
	var s string
	switch shop {
	case "shop_a":
		s = "sell_factor_a"
	case "shop_b":
		s = "sell_factor_b"
	case "shop_c":
		s = "sell_factor_c"
	case "warehouse":
		s = "sell_factor_warehouse"
	}
	return fmt.Sprintf("SELECT motors.pelak_number, motors.color, motors.model_name, %s.price, %s.date, %s.factor_number FROM %s INNER JOIN motors ON %s.pelak_number = motors.pelak_number WHERE %s.date>=$1 AND %s.date<=$2 ORDER BY %s.date DESC", s, s, s, s, s, s, s, s)
}

// UnpayedRecListQuery is a ...
func UnpayedRecListQuery(shop string) string {
	var sellFactor, receiveable string
	switch shop {
	case "shop_a":
		sellFactor = "sell_factor_a"
	case "shop_b":
		sellFactor = "sell_factor_b"
	case "shop_c":
		sellFactor = "sell_factor_c"
	case "warehouse":
		sellFactor = "sell_factor_warehouse"
	}
	switch shop {
	case "shop_a":
		receiveable = "accounts_receivable_a"
	case "shop_b":
		receiveable = "accounts_receivable_b"
	case "shop_c":
		receiveable = "accounts_receivable_c"
	case "warehouse":
		receiveable = "accounts_receivable_warehouse"
	}
	return fmt.Sprintf("SELECT %s.factor_number, %s.price, %s.date ,%s.pelak_number, %s.buyer_name, %s.buyer_last_name, %s.buyer_mobile FROM %s INNER JOIN %s ON %s.factor_number = %s.factor_number WHERE %s.status=false AND %s.date <= (now() + INTERVAL '30 days') ORDER BY %s.date DESC", receiveable, receiveable, receiveable, sellFactor, sellFactor, sellFactor, sellFactor, receiveable, sellFactor, receiveable, sellFactor, receiveable, receiveable, receiveable)
}

// UnpayedPayListQuery is a...
func UnpayedPayListQuery(shop string) string {
	var buyFactor, payable string
	switch shop {
	case "shop_a":
		buyFactor = "buy_factor_a"
	case "shop_b":
		buyFactor = "buy_factor_b"
	case "shop_c":
		buyFactor = "buy_factor_c"
	case "warehouse":
		buyFactor = "buy_factor_warehouse"
	}
	switch shop {
	case "shop_a":
		payable = "accounts_payable_a"
	case "shop_b":
		payable = "accounts_payable_b"
	case "shop_c":
		payable = "accounts_payable_c"
	case "warehouse":
		payable = "accounts_payable_warehouse"
	}
	return fmt.Sprintf("SELECT %s.factor_number, %s.price, %s.date, %s.pelak_number, %s.customer_name, %s.customer_last_name, %s.customer_mobile FROM %s INNER JOIN %s ON %s.factor_number = %s.factor_number WHERE %s.status=false AND %s.date <= (now() +  INTERVAL '30 days') ORDER BY %s.date DESC", payable, payable, payable, buyFactor, buyFactor, buyFactor, buyFactor, payable, buyFactor, payable, buyFactor, payable, payable, payable)
}
