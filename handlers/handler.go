package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	dbf "../dbf-go"
	util "../utility"
	"github.com/dgrijalva/jwt-go"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//SECRETKEY is the key for signing token
const SECRETKEY = "secret"

//AuthMiddleWare is middleWare that handles the authentication process
func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header["Authorization"][0], "bearer ")
		if len(authHeader) != 2 {
			fmt.Println("malformed token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("malformed token"))
		}
		token, err := jwt.Parse(authHeader[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(SECRETKEY), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "props", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			fmt.Println(err)
		}
		// fmt.Println(r.Header, len(r.Header))
	})
}

//LoginHandler is responsible for handling login actions
func LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content-Type", "application/json")
		var user util.User
		var res util.LoginResponse
		body, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(body, &user); err != nil {
			res.Err = err.Error()
			json.NewEncoder(w).Encode(res)
			return
		}
		var temp util.User
		err = env.DB.QueryRow(dbf.UserQuery, user.Username).Scan(&temp.Password, pq.Array(&temp.Accessibility))
		if err != nil {
			res.Err = err.Error()
			json.NewEncoder(w).Encode(res)
			// w.WriteHeader(http.StatusForbidden)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(temp.Password), []byte(user.Password))
		if err != nil {
			res.Err = "password is wrong!"
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(res)
			return
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user": user.Username,
			"exp":  time.Now().Add(time.Hour * time.Duration(1)).Unix(),
			"iat":  time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte(SECRETKEY))
		if err != nil {
			res.Err = "error occurd while generating token"
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = tokenString
		res.Accessibility = temp.Accessibility
		json.NewEncoder(w).Encode(res)
	})
}

//CreateUser is handler for creating new admin user in db
func CreateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content-Type", "application/json")
		var user util.User
		var res util.Response
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			res.Err = "error dar ghesmate readAll"
			json.NewEncoder(w).Encode(res)
			return
		}
		err = json.Unmarshal(body, &user)
		if err != nil {
			res.Err = err.Error()
			json.NewEncoder(w).Encode(res)
			return
		}
		if len(user.Accessibility) == 0 {
			res.Err = "accessibility cant be empty, Try again!"
			json.NewEncoder(w).Encode(res)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)
		if err != nil {
			res.Err = "some error occurd during hashing password, Try again!"
			json.NewEncoder(w).Encode(res)
			return
		}
		result, err := env.DB.Exec(dbf.CreateNewUser, user.Username, hash, pq.Array(user.Accessibility))
		if err != nil {
			res.Err = fmt.Sprintf("during inserting data to db %s error occurd and result was %s", err.Error(), result)
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = fmt.Sprintf("user created successfully : %s", result)
		json.NewEncoder(w).Encode(res)
	})
}

//HandleBuy is a handler for inserting new buy factor in db
func HandleBuy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content_Type", "application/json")
		var factor util.BuyFactor
		var res util.Response
		body, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(body, &factor); err != nil {
			res.Err = fmt.Sprintf("during parsing json this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		fmt.Println("from pay: ", factor)
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = fmt.Sprintf("during beginig a transactions this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		_, err = tx.ExecContext(ctx, dbf.InsertNewMotor, factor.Motor.PelakNumber, factor.Motor.Color, factor.Motor.BodyNumber, factor.Motor.ModelName, factor.Motor.ModelYear)
		if err != nil {
			res.Err = fmt.Sprintf("during inserting motors to db this error occurd : %s ", err.Error())
			tx.Rollback()
			json.NewEncoder(w).Encode(res)
			return
		}
		_, err = tx.ExecContext(ctx, dbf.InsertBuyFactor(factor.Shop), factor.FactorNumber, factor.Motor.PelakNumber, factor.Price, factor.Date, factor.Customer.CustomerName, factor.Customer.CustomerLastName, factor.Customer.CustomerNationalCode, factor.Customer.CustomerMobile)
		if err != nil {
			res.Err = fmt.Sprintf("during inserting data to buy-factor this error occurd: %s", err)
			tx.Rollback()
			json.NewEncoder(w).Encode(res)
			return
		}
		_, err = tx.ExecContext(ctx, dbf.InsertIntoShop(factor.Shop), factor.Motor.PelakNumber, factor.FactorNumber)
		if err != nil {
			res.Err = fmt.Sprintf("during inserting motors into specified shop this error occurd : %s ", err.Error())
			tx.Rollback()
			json.NewEncoder(w).Encode(res)
			return
		}
		if len(factor.Debts) > 0 {
			for _, debt := range factor.Debts {
				_, err := tx.ExecContext(ctx, dbf.InsertPayable(factor.Shop), factor.FactorNumber, debt.Price, debt.Date)
				fmt.Println("come here")
				if err != nil {
					res.Err = fmt.Sprintf("during inserting payable accounts this error occurd : %s ", err.Error())
					tx.Rollback()
					json.NewEncoder(w).Encode(res)
					return
				}
			}
		}
		err = tx.Commit()
		if err != nil {
			res.Err = fmt.Sprintf("during commiting the transactions this error happend : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "Successfully inserted"
		json.NewEncoder(w).Encode(res)
	})
}

//HandleSell is responsible for handling the incoming sell factor
func HandleSell() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content_Type", "application/json")
		var factor util.SellFactor
		var res util.Response
		body, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(body, &factor); err != nil {
			res.Err = "something went wrong during parsing json request"
			json.NewEncoder(w).Encode(res)
			return
		}
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = fmt.Sprintf("during beginig the transaction this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		fmt.Println(factor, "factor is this")
		_, err = tx.ExecContext(ctx, dbf.InsertSellFactor(factor.Shop), factor.FactorNumber, factor.PelakNumber, factor.Price, factor.Date, factor.Buyer.BuyerName, factor.Buyer.BuyerLastName, factor.Buyer.BuyerNationalCode, factor.Buyer.BuyerMobile)
		if err != nil {
			res.Err = fmt.Sprintf("during inserting new sell factor this error occurd: %s", err.Error())
			tx.Rollback()
			json.NewEncoder(w).Encode(res)
			return
		}
		if len(factor.Demands) > 0 {
			for _, demand := range factor.Demands {
				_, err := tx.ExecContext(ctx, dbf.InsertRecievable(factor.Shop), factor.FactorNumber, demand.Price, demand.Date)
				if err != nil {
					res.Err = fmt.Sprintf("during inserting recievable accounts into db this error: %s", err.Error())
					tx.Rollback()
					json.NewEncoder(w).Encode(res)
					return
				}
			}
		}
		_, err = tx.ExecContext(ctx, dbf.UpdateShopSellFactor(factor.Shop), factor.FactorNumber, false, factor.PelakNumber)
		if err != nil {
			res.Err = fmt.Sprintf("During updating specified shop this error occurd : %s", err.Error())
			tx.Rollback()
			json.NewEncoder(w).Encode(res)
			return
		}
		err = tx.Commit()
		if err != nil {
			res.Err = fmt.Sprintf("during commiting the transactions this error happend : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = fmt.Sprintf("sell factor successfully created")
		json.NewEncoder(w).Encode(res)
	})
}

//HandleList is a handler for inserting new list of motors to specified shop
func HandleList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content_Type", "application/json")
		body, _ := ioutil.ReadAll(r.Body)
		var res util.Response
		var list util.List
		err = json.Unmarshal(body, &list)
		if err != nil {
			res.Err = fmt.Sprintf("during parsing json data this error occurd %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = fmt.Sprintf("during beginig the transaction this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		err = tx.QueryRow(dbf.InsertList, list.Provider, list.Date, list.Amount).Scan(&list.ID)
		if err != nil {
			res.Err = fmt.Sprintf("during inserting into entry_list this error occurd: %s", err.Error())
			tx.Rollback()
			json.NewEncoder(w).Encode(res)
			return
		}
		var wg sync.WaitGroup
		for _, motor := range list.Motors {
			wg.Add(1)
			go func(motor util.Motor) {
				defer wg.Done()
				_, err := tx.ExecContext(ctx, dbf.InsertNewMotorFromList, motor.PelakNumber, motor.Color, motor.BodyNumber, motor.ModelName, motor.ModelYear, list.ID)
				if err != nil {
					res.Err = fmt.Sprintf("during inserting motors to db this error: %s", err.Error())
					tx.Rollback()
					// json.NewEncoder(w).Encode(res)
					return
				}
				_, err = tx.ExecContext(ctx, dbf.InsertShopEntry(list.Shop), motor.PelakNumber)
				if err != nil {
					res.Err = fmt.Sprintf("during inserting motors to specified shop this error: %s", err.Error())
					tx.Rollback()
					// json.NewEncoder(w).Encode(res)
					return
				}
			}(motor)

		}
		wg.Wait()
		if len(res.Err) > 0 {
			json.NewEncoder(w).Encode(res)
			return
		}
		err = tx.Commit()
		if err != nil {
			res.Err = fmt.Sprintf("during commiting transactions to db this error happend: %s", err.Error())
			return
		}
		res.Result = "entry list successfully created"
		json.NewEncoder(w).Encode(res)
	})
}

//UpdateReceive is a handler for updating receivable accounts
func UpdateReceive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content_Type", "application/json")
		var factor util.SellFactor
		var res util.Response
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println("here")
		err = json.Unmarshal(body, &factor)
		if err != nil {
			res.Err = fmt.Sprintf("during parsing json this error occurd %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		result, err := env.DB.Exec(dbf.UpdateReceive(factor.Shop), factor.FactorNumber, factor.Date)
		if err != nil {
			res.Err = fmt.Sprintf("during updating receivable accounts this error occurd : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		if s, _ := result.RowsAffected(); s == 0 {
			res.Err = fmt.Sprintf("no row updated, please check your inputs are correct and try again!")
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "recive account successfully updated"
		json.NewEncoder(w).Encode(res)
		return
	})
}

//UpdatePayable is a Handler for updating payable accounts
func UpdatePayable() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content-Type", "application/json")
		var res util.Response
		var factor util.SellFactor
		body, _ := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(body, &factor)
		if err != nil {
			res.Err = fmt.Sprintf("during parsing json this error occurd : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		result, err := env.DB.Exec(dbf.UpdatePayable(factor.Shop), factor.FactorNumber, factor.Date)
		if err != nil {
			res.Err = fmt.Sprintf("during updating db this error happend : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		s, _ := result.RowsAffected()
		if s == 0 {
			res.Err = fmt.Sprintf("no rows updated, check your input and try again!")
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = fmt.Sprintf("payable accounts updated successfully")
		json.NewEncoder(w).Encode(res)
		return
	})
}

//HandleSwap is a handler for swaping motors between diffrent locations
func HandleSwap() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content-Type", "application/json")
		var motors []util.MotorSrc
		var res util.Response
		body, _ := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(body, &motors)
		if err != nil {
			res.Err = fmt.Sprintf("during parsing json this error %s happend", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		fmt.Println("here!", motors)
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = fmt.Sprintf("during begining the transaction this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		var wg sync.WaitGroup
		for _, m := range motors {
			wg.Add(1)
			go func(m util.MotorSrc) {
				defer wg.Done()
				var temp = m
				var fs sql.NullString
				if err := tx.QueryRow(dbf.RemoveSrc(m.Src), m.PelakNumber).Scan(&fs, &temp.Stock); err != nil {
					res.Err = fmt.Sprintf("during remove from source this error occurd : %s", err.Error())
					tx.Rollback()
					fmt.Println(temp, "from error")
					// json.NewEncoder(w).Encode(res)
					return
				}
				if !temp.Stock {
					res.Err = fmt.Sprintf("motor with pelak_number : %s is not in stock", temp.PelakNumber)
					// json.NewEncoder(w).Encode(res)
					tx.Rollback()
					return
				}
				if fs.Valid {
					temp.BuyFactor = fs.String
					if _, err := tx.ExecContext(ctx, dbf.InsertIntoShop(m.Dst), temp.PelakNumber, temp.BuyFactor); err != nil {
						res.Err = fmt.Sprintf("during inserting into destination %s this error: %s ", m.Dst, err.Error())
						tx.Rollback()
						// json.NewEncoder(w).Encode(res)
						return
					}
				} else {
					if _, err := tx.ExecContext(ctx, dbf.InsertShopEntry(m.Dst), temp.PelakNumber); err != nil {
						res.Err = fmt.Sprintf("during inserting into destination %s this error: %s ", m.Dst, err.Error())
						tx.Rollback()
						// json.NewEncoder(w).Encode(res)
						return
					}
				}
				return
			}(m)
		}
		wg.Wait()
		if len(res.Err) > 0 {
			json.NewEncoder(w).Encode(res)
			return
		}
		err = tx.Commit()
		if err != nil {
			res.Err = fmt.Sprintf("during commiting transactions this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "successfully swapped"
		json.NewEncoder(w).Encode(res)
	})
}

//StockHandle is a handler for returning the specified shops inventory at current time
func StockHandle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content_Type", "application/json")
		var locations struct {
			Shops []string `json:"shops"`
		}
		var res util.LookUpResponse
		body, _ := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(body, &locations)
		if err != nil {
			res.Err = fmt.Sprintf("during parsing json this error occurd : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = fmt.Sprintf("during beginig the transactions this error happend : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		var wg sync.WaitGroup
		for _, l := range locations.Shops {
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				var mres util.MotorsResult
				mres.Shop = l
				stms := dbf.StockLookUp(l)
				rows, err := tx.QueryContext(ctx, stms)
				defer rows.Close()
				if err != nil {
					res.Err = fmt.Sprintf("during query to db for location %s this error occurd : %s", l, err.Error())
					tx.Rollback()
					// json.NewEncoder(w).Encode(res)
					return
				}
				for rows.Next() {
					var temp util.LookUp
					var bf sql.NullString
					err := rows.Scan(&temp.PelakNumber, &temp.Color, &temp.ModelName, &bf)
					if err != nil {
						res.Err = fmt.Sprintf("during reading rows from db this error happend: %s", err.Error())
						tx.Rollback()
						// json.NewEncoder(w).Encode(res)
						return
					}
					if bf.Valid {
						temp.BuyFactor = bf.String
					}
					err = rows.Err()
					if err != nil {
						res.Err = fmt.Sprintf("rows.err is : %s", err)
						tx.Rollback()
						json.NewEncoder(w).Encode(res)
						return
					}
					mres.Motors = append(mres.Motors, temp)
				}
				if len(mres.Motors) > 0 {
					res.Result = append(res.Result, mres)
				}
				return
			}(l)

		}
		wg.Wait()
		if len(res.Err) > 0 {
			json.NewEncoder(w).Encode(res)
			return
		}
		err = tx.Commit()
		if err != nil {
			res.Result = []util.MotorsResult{}
			res.Err = fmt.Sprintf("during commiting transactions this error haapend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		json.NewEncoder(w).Encode(res)
	})
}

//HandleSaleHistory is a handler for returning the sales of specified shops in certain time
// func HandleSaleHistory(env *dbf.Env) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content_Type", "application/json")
// 		var filter util.TimeFilter
// 		var res util.SaleHistoryResponse
// 		body, _ := ioutil.ReadAll(r.Body)
// 		err := json.Unmarshal(body, &filter)
// 		if err != nil {
// 			res.Err = fmt.Sprintf("during parsing json this error happend : %s", err.Error())
// 			json.NewEncoder(w).Encode(res)
// 			return
// 		}
// 		fmt.Println(filter)
// 		ctx := context.Background()
// 		tx, err := env.DB.BeginTx(ctx, nil)
// 		if err != nil {
// 			res.Err = fmt.Sprintf("during begining transaction this error happend: %s", err.Error())
// 			json.NewEncoder(w).Encode(res)
// 			return
// 		}
// 		var wg sync.WaitGroup
// 		for _, s := range filter.Shops {
// 			wg.Add(1)
// 			go func(s string) {
// 				defer wg.Done()
// 				var saleRes util.SaleResult
// 				saleRes.Shop = s
// 				rows, err := tx.QueryContext(ctx, dbf.SaleHistoryLookUp(s), filter.From, filter.To)
// 				if err != nil {
// 					res.Err = fmt.Sprintf("during query to db for location %s this error happed: %s", s, err.Error())
// 					tx.Rollback()
// 					// json.NewEncoder(w).Encode(res)
// 					return
// 				}
// 				defer rows.Close()
// 				for rows.Next() {
// 					var temp util.SaleHistory
// 					if err := rows.Scan(&temp.PelakNumber, &temp.Color, &temp.ModelName, &temp.Price, &temp.Date, &temp.SellFactor); err != nil {
// 						res.Err = fmt.Sprintf("during scaning rows from db this err occurd :%s", err.Error())
// 						tx.Rollback()
// 						// json.NewEncoder(w).Encode(res)
// 						return
// 					}
// 					saleRes.Sales = append(saleRes.Sales, temp)
// 				}
// 				// res.Result = append(res.Result, saleRes)
// 				if len(saleRes.Sales) > 0 {
// 					res.Result = append(res.Result, saleRes)
// 				}
// 				return
// 			}(s)
// 		}
// 		wg.Wait()
// 		if len(res.Err) > 0 {
// 			json.NewEncoder(w).Encode(res)
// 			return
// 		}
// 		err = tx.Commit()
// 		if err != nil {
// 			res.Err = fmt.Sprintf("during commiting transactions this error happend : %s", err.Error())
// 			res.Result = []util.SaleResult{}
// 			return
// 		}
// 		json.NewEncoder(w).Encode(res)
// 		return
// 	})
// }

// HandleUnpayedRec is a Handler for returning unpayed receivable accounts for spesified shops
func HandleUnpayedRec() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content_Type", "application/json")
		body, _ := ioutil.ReadAll(r.Body)
		var res util.UnpayedRecResponse
		var locations struct {
			Shops []string `json:"shops"`
		}
		err = json.Unmarshal(body, &locations)
		if err != nil {
			res.Err = fmt.Sprintf("during parsing json this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		var wg sync.WaitGroup
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = fmt.Sprintf("during begining transactions this error happend : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		for _, l := range locations.Shops {
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				var temp util.UnpayedRecResult
				temp.Shop = l
				fmt.Println("shops", l)
				statm := dbf.UnpayedRecListQuery(l)
				rows, err := tx.QueryContext(ctx, statm)
				if err != nil {
					res.Err = fmt.Sprintf("during query to db for unpayed list for %s location this error %s happend", l, err.Error())
					// json.NewEncoder(w).Encode(res)
					tx.Rollback()
					return
				}
				defer rows.Close()
				for rows.Next() {
					var t util.UnpayedListRec
					fmt.Println("here3")
					if err := rows.Scan(&t.FactorNumber, &t.Price, &t.Date, &t.PelakNumber, &t.BuyerName, &t.BuyerLastName, &t.Mobile); err != nil {
						res.Err = fmt.Sprintf("during scaning row this error happend : %s", err.Error())
						// json.NewEncoder(w).Encode(res)
						tx.Rollback()
						return
					}
					temp.List = append(temp.List, t)
				}
				if len(temp.List) > 0 {
					res.Result = append(res.Result, temp)
				}
				return
			}(l)
		}
		wg.Wait()
		if len(res.Err) > 0 {
			json.NewEncoder(w).Encode(res)
			return
		}
		err = tx.Commit()
		if err != nil {
			res.Err = fmt.Sprintf("during commiting transactions this error happend : %s", err.Error())
			res.Result = []util.UnpayedRecResult{}
			json.NewEncoder(w).Encode(res)
			return
		}
		json.NewEncoder(w).Encode(res)
		return
	})
}

// HandleUnpayedPay is a Handler for returning unpayed payable accounts for speciefied shops
func HandleUnpayedPay() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content_type", "application/json")
		var locations struct {
			Shops []string `json:"shops"`
		}
		var res util.UnpayedPayResponse
		body, _ := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(body, &locations)
		if err != nil {
			res.Err = fmt.Sprintf("during parsing json this error happend : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		var wg sync.WaitGroup
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = fmt.Sprintf("during begining the transactions this error happend : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		for _, l := range locations.Shops {
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				var temp util.UnpayedPayResult
				temp.Shop = l
				stm := dbf.UnpayedPayListQuery(l)
				rows, err := tx.QueryContext(ctx, stm)
				if err != nil {
					res.Err = fmt.Sprintf("during query to db for %s location this error happend : %s", l, err.Error())
					tx.Rollback()
					// json.NewEncoder(w).Encode(res)
					return
				}
				defer rows.Close()
				for rows.Next() {
					var t util.UnpayedListPay
					if err := rows.Scan(&t.FactorNumber, &t.Price, &t.Date, &t.PelakNumber, &t.CustomerName, &t.CustomerLastName, &t.Mobile); err != nil {
						res.Err = fmt.Sprintf("during scaning rows this error happend: %s", err.Error())
						tx.Rollback()
						// json.NewEncoder(w).Encode(res)
						return
					}
					temp.List = append(temp.List, t)
				}
				// if len(temp.List) > 0 {
				// 	res.Result = append(res.Result, temp)
				// }
				res.Result = append(res.Result, temp)
				return
			}(l)

		}
		wg.Wait()
		if len(res.Err) > 0 {
			json.NewEncoder(w).Encode(res)
			return
		}
		err = tx.Commit()
		if err != nil {
			res.Err = fmt.Sprintf("during commiting transactions this error happend : %s", err.Error())
			res.Result = []util.UnpayedPayResult{}
			json.NewEncoder(w).Encode(res)
			return
		}
		json.NewEncoder(w).Encode(res)
		return
	})
}

//HandleSaleHistory is a....
func HandleSaleHistory() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := dbf.RunDB()
		if err != nil {
			log.Fatal(err)
		}
		var env dbf.Env
		env.DB = db
		defer db.Close()
		w.Header().Set("Content_Type", "application/json")
		var filter util.TimeFilter
		var res util.SaleHistoryResponse
		body, _ := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(body, &filter)
		if err != nil {
			res.Err = fmt.Sprintf("during parsing json this error happend : %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		fmt.Println(filter)
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = fmt.Sprintf("during begining transaction this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		var wg sync.WaitGroup
		for _, s := range filter.Shops {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				var saleRes util.SaleResult
				saleRes.Shop = s
				stm := dbf.SaleHistoryLookUp(s)
				rows, err := tx.QueryContext(ctx, stm, filter.From, filter.To)
				if err != nil {
					res.Err = fmt.Sprintf("during query to db for location %s this error happed: %s", s, err.Error())
					tx.Rollback()
					// json.NewEncoder(w).Encode(res)
					return
				}
				defer rows.Close()
				for rows.Next() {
					var temp util.SaleHistory
					if err := rows.Scan(&temp.PelakNumber, &temp.Color, &temp.ModelName, &temp.Price, &temp.Date, &temp.SellFactor); err != nil {
						res.Err = fmt.Sprintf("during scaning rows from db this err occurd :%s", err.Error())
						tx.Rollback()
						// json.NewEncoder(w).Encode(res)
						return
					}
					saleRes.Sales = append(saleRes.Sales, temp)
				}
				// res.Result = append(res.Result, saleRes)
				if len(saleRes.Sales) > 0 {
					res.Result = append(res.Result, saleRes)
				}
				return
			}(s)
		}
		wg.Wait()
		if len(res.Err) > 0 {
			json.NewEncoder(w).Encode(res)
			return
		}
		err = tx.Commit()
		if err != nil {
			res.Err = fmt.Sprintf("during commiting transactions this error happend : %s", err.Error())
			res.Result = []util.SaleResult{}
			return
		}
		json.NewEncoder(w).Encode(res)
		return
	})
}
