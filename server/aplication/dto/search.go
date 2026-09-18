package dto

type FuzzySearchDto struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type ESClient struct {
	Username            string
	Password            string
	Endpoint1           string
	Endpoint2           string
	MaxIdleConnsPerHost int
}
