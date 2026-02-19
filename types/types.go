package types

type AccountConfig struct {
	Label           string `json:"label"`
	DMarketKey      string `json:"dmarket_key"`
	CSFloatKey      string `json:"csfloat_key"` // Optional
	TelegramToken   string `json:"telegram_token"`
	TelegramChatID  string `json:"telegram_chat_id"`
	AdvancedBalance bool   `json:"advanced_balance"`
	ProfitPercent   bool   `json:"profit_percent"`
	IgnoreReleased  bool   `json:"ignore_released"`
}

type ChatIDConfig struct {
	Offers       string `json:"offers"`
	Transactions string `json:"transactions"`
}

type Transaction struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	CustomID   string `json:"customId"`
	Emitter    string `json:"emitter"`
	Action     string `json:"action"`
	Subject    string `json:"subject"`
	Contractor struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
	} `json:"contractor"`
	Details struct {
		SettlementTime int64  `json:"-"`
		Image          string `json:"-"`
		ItemID         string `json:"itemId"`
		Extra          struct {
			FloatPartValue string  `json:"floatPartValue"`
			FloatValue     float64 `json:"floatValue"`
			PaintSeed      *int    `json:"paintSeed"` // to compare nil instead of 0
			PhaseTitle     string  `json:"phaseTitle"`
		} `json:"extra"`
	} `json:"details"`
	Changes []struct {
		Money struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"money"`
		ChangeType string `json:"changeType"`
	} `json:"changes"`
	From    string `json:"from"`
	To      string `json:"to"`
	Status  string `json:"status"`
	Balance struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"balance"`
	UpdatedAt int64 `json:"updatedAt"`
	CreatedAt int64 `json:"createdAt"`
}

type TransactionsResponse struct {
	Objects []Transaction `json:"objects"`
	Total   int           `json:"total"`
}

type CostMap map[string]float64

type CSFloatTrade struct {
	ID       string `json:"id"`
	Contract struct {
		Price int `json:"price"` // Price is in CENTS (e.g., 100 = $1.00)
		Item  struct {
			FloatValue float64 `json:"float_value"`
			PaintSeed  *int    `json:"paint_seed"`
			MarketName string  `json:"market_hash_name"`
		} `json:"item"`
	} `json:"contract"`
}

type DMarketInventoryItem struct {
	ItemID string `json:"itemId"`
	Extra  struct {
		FloatValue float64 `json:"floatValue"`
		PaintSeed  *int    `json:"paintSeed"` // DMarket uses int for seed
	} `json:"extra"`
}

// DMarketInventoryResponse represents the API response from /exchange/v1/user/offers
type DMarketInventoryResponse struct {
	Objects []DMarketInventoryItem `json:"objects"`
	Cursor  string                 `json:"cursor"`
}

// CSFloatResponse represents the list of trades
type CSFloatResponse struct {
	Trades []CSFloatTrade `json:"Trades"`
	Count  int            `json:"count"`
}

type UserBalanceResponse struct {
	Usd               string `json:"usd"`               // Available
	UsdTradeProtected string `json:"usdTradeProtected"` // Pending
}

type TargetTrade struct {
	OfferID  string `json:"OfferID"`
	TargetID string `json:"TargetID"`
	AssetID  string `json:"AssetID"` // This corresponds to ItemID
	Price    struct {
		CurrencyCode string  `json:"CurrencyCode"`
		Amount       float64 `json:"Amount"`
	} `json:"Price"`
	Title            string `json:"Title"`
	ClosedAt         string `json:"ClosedAt"`
	Status           string `json:"Status"`
	FinalizationTime string `json:"FinalizationTime"`
}

type UserTargetsClosedResponse struct {
	Trades []TargetTrade `json:"Trades"`
	Total  string        `json:"Total"`  // API returns this as a string number
	Cursor string        `json:"Cursor"` // Used for pagination
}

type AttributeKey struct {
	Phase          string
	PaintSeed      string
	FloatPartValue string
}