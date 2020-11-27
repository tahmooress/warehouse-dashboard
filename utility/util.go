package utility

//User is model for active users
type User struct {
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	Accessibility []string `json:"accessibility"`
}

//LoginResponse is a type for login users
type LoginResponse struct {
	Err           string   `json:"err"`
	Result        string   `json:"result"`
	Accessibility []string `json:"accessibility"`
}

//Response is a type for response of handlers
type Response struct {
	Err    string `json:"err"`
	Result string `json:"result"`
}

//Motor is a model for information of motor
type Motor struct {
	PelakNumber string `json:"pelakNumber"`
	BodyNumber  string `json:"bodyNumber"`
	ModelName   string `json:"modelName"`
	ModelYear   string `json:"modelYear"`
	Color       string `json:"color"`
	ListID      string `json:"listID"`
}

//BuyFactor is a model for infromation of a buy factor
type BuyFactor struct {
	FactorNumber string   `json:"factorNumber"`
	Motor        Motor    `json:"motor"`
	Price        string   `json:"price"`
	Date         string   `json:"date"`
	Customer     Customer `json:"customer"`
	Debts        []Debt   `json:"debts"`
	Shop         string   `json:"shop"`
}

//Customer is a model for people who we buy motor from
type Customer struct {
	CustomerName         string `json:"customerName"`
	CustomerLastName     string `json:"customerLastName"`
	CustomerMobile       string `json:"customerMobile"`
	CustomerNationalCode string `json:"customerNationalCode"`
}

//SellFactor is a model for infromation of a sell factor
type SellFactor struct {
	FactorNumber string   `json:"factorNumber"`
	PelakNumber  string   `json:"pelakNumber"`
	Price        string   `json:"price"`
	Date         string   `json:"date"`
	Buyer        Buyer    `json:"buyer"`
	Demands      []Demand `json:"demands"`
	Shop         string   `json:"shop"`
}

//Buyer is a model of buyer
type Buyer struct {
	BuyerName         string `json:"buyerName"`
	BuyerLastName     string `json:"buyerLastName"`
	BuyerMobile       string `json:"buyerMobile"`
	BuyerNationalCode string `json:"buyerNationalCode"`
}

//Debt is a a model for future payable accounts
type Debt struct {
	Date   string `json:"date"`
	Price  string `json:"price"`
	Status bool   `json:"status"`
}

//Demand is a model of future receivable accounts
type Demand struct {
	Date   string `json:"date"`
	Price  string `json:"price"`
	Status bool   `json:"status"`
}

//List is a model for incoming data for motor entry list
type List struct {
	Provider string  `json:"provider"`
	Date     string  `json:"date"`
	Amount   string  `json:"amount"`
	ID       int     `json:"id"`
	Motors   []Motor `json:"motors"`
	Debts    []Debt  `json:"debts"`
	Shop     string  `json:"shop"`
}

//MotorSrc is a ...
type MotorSrc struct {
	PelakNumber string `json:"pelakNumber"`
	Src         string `json:"src"`
	Dst         string `json:"dst"`
	BuyFactor   string `json:"buy_factor"`
	Stock       bool   `json:"stock"`
}

// LookUp is a...
type LookUp struct {
	PelakNumber string `json:"pelakNumber"`
	Color       string `json:"color"`
	ModelName   string `json:"modelName"`
	BuyFactor   string `json:"buyFactor"`
}

// MotorsResult is a ...
type MotorsResult struct {
	Motors []LookUp `json:"motors"`
	Shop   string   `json:"shop"`
}

//LookUpResponse is a ...
type LookUpResponse struct {
	Result []MotorsResult `json:"result"`
	Err    string         `json:"err"`
}

//TimeFilter is a...
type TimeFilter struct {
	Shops []string `json:"shops"`
	From  string   `json:"from"`
	To    string   `json:"to"`
}

// SaleHistory is a...
type SaleHistory struct {
	PelakNumber string `json:"pelakNumber"`
	Color       string `json:"color"`
	ModelName   string `json:"modelName"`
	SellFactor  string `json:"sellFactor"`
	Price       string `json:"price"`
	Date        string `json:"date"`
}

// SaleResult is a...
type SaleResult struct {
	Sales []SaleHistory `json:"sales"`
	Shop  string        `json:"shop"`
}

// SaleHistoryResponse is a ...
type SaleHistoryResponse struct {
	Result []SaleResult `json:"result"`
	Err    string       `json:"err"`
}

// UnpayedListRec is a ...
type UnpayedListRec struct {
	FactorNumber  string `json:"factorNumber"`
	PelakNumber   string `json:"pelakNumber"`
	Price         string `json:"price"`
	Date          string `json:"date"`
	BuyerName     string `json:"buyerName"`
	BuyerLastName string `json:"buyerLastName"`
	Mobile        string `json:"mobile"`
}

// UnpayedRecResult is a...
type UnpayedRecResult struct {
	List []UnpayedListRec `json:"list"`
	Shop string           `json:"shop"`
}

//UnpayedRecResponse is a...
type UnpayedRecResponse struct {
	Result []UnpayedRecResult `json:"result"`
	Err    string             `json:"err"`
}

// UnpayedListPay is a...
type UnpayedListPay struct {
	FactorNumber     string `json:"factroNumber"`
	PelakNumber      string `json:"pelakNumber"`
	Price            string `json:"price"`
	Date             string `json:"date"`
	CustomerName     string `json:"customerName"`
	CustomerLastName string `json:"customerLastName"`
	Mobile           string `json:"mobile"`
}

// UnpayedPayResult is a ...
type UnpayedPayResult struct {
	List []UnpayedListPay `json:"list"`
	Shop string           `json:"shop"`
}

// UnpayedPayResponse is a ...
type UnpayedPayResponse struct {
	Result []UnpayedPayResult `json:"result"`
	Err    string             `json:"err"`
}
