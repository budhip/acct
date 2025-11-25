package gocustomer

type Product struct {
	ProductName    string   `json:"productName"`
	Active         bool     `json:"active"`
	AccountNumbers []string `json:"accountNumbers"`
}

// CustomerEventPayload is payload received from kafka topic `queue.customer`
type CustomerEventPayload struct {
	CustomerNumber string    `json:"customerNumber"`
	Name           string    `json:"name"`
	Products       []Product `json:"products"`
}
