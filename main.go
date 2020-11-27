package main

import (
	"log"
	"net/http"

	h "./handlers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	r := mux.NewRouter()
	method := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	origin := handlers.AllowedOrigins([]string{"*"})
	headers := handlers.AllowedHeaders([]string{"X-Request-With", "Content-Type", "Authorization"})
	// db, err := dbf.RunDB()
	// if err != nil {
	// 	panic(err)
	// }
	// env := &dbf.Env{DB: db}
	r.Handle("/login", h.LoginHandler()).Methods("POST")
	r.Handle("/create-user", h.CreateUser()).Methods("POST")
	r.Handle("/buy-factor", h.AuthMiddleWare(h.HandleBuy())).Methods("POST")
	r.Handle("/sell-factor", h.AuthMiddleWare(h.HandleSell())).Methods("POST")
	r.Handle("/entry-list", h.AuthMiddleWare(h.HandleList())).Methods("POST")
	r.Handle("/stock-lookup", h.AuthMiddleWare(h.StockHandle())).Methods("POST")
	r.Handle("/sales-history", h.AuthMiddleWare(h.HandleSaleHistory())).Methods("POST")
	r.Handle("/update-recive", h.AuthMiddleWare(h.UpdateReceive())).Methods("PUT")
	r.Handle("/update-payable", h.AuthMiddleWare(h.UpdatePayable())).Methods("PUT")
	r.Handle("/swap-inventory", h.AuthMiddleWare(h.HandleSwap())).Methods("PUT")
	r.Handle("/unrec-list", h.AuthMiddleWare(h.HandleUnpayedRec())).Methods("POST")
	r.Handle("/unpay-list", h.AuthMiddleWare(h.HandleUnpayedPay())).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(headers, method, origin)(r)))
	// defer db.Close()
}
