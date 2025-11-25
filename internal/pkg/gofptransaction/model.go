package gofptransaction

const SERVICE_NAME string = "go-fp-transaction"

type UpdateBySubCategory struct {
	Code            string
	ProductTypeName *string
	Currency        *string
}

type GetAllOrderTypesOut struct {
	Kind      string          `json:"kind"`
	Contents  []OrderTypesOut `json:"contents"`
	TotalRows int             `json:"total_rows"`
}

type OrderTypesOut struct {
	Kind             string               `json:"kind"`
	OrderTypeCode    string               `json:"orderTypeCode"`
	OrderTypeName    string               `json:"orderTypeName"`
	TransactionTypes []TransactionTypeOut `json:"transactionTypes"`
}

type TransactionTypeOut struct {
	Kind                string `json:"kind"`
	TransactionTypeCode string `json:"transactionTypeCode"`
	TransactionTypeName string `json:"transactionTypeName"`
}
