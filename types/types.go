package types

type AttributeKey struct {
	Phase          string
	PaintSeed      string
	FloatPartValue string
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
}

type TransactionsResponse struct {
	Objects []Transaction `json:"objects"`
	Total   int           `json:"total"`
}