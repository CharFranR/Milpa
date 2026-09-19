package dto

// response
type FuzzySearchDto struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Price          float64 `json:"price"`
	Type           string  `json:"type"`
	ImageURL       string  `json:"image_url"`
	FarmerID       string  `json:"farmer_id"`
	FarmerName     string  `json:"farmer_name"`
	FarmerVerified bool    `json:"farmer_verified"`
	Department     string  `json:"department"`
	Municipality   string  `json:"municipality"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
}

// query
type SearchRequest struct {
	Term         string  `json:"term"`
	Type         string  `json:"type,omitempty"`
	Department   string  `json:"department,omitempty"`
	Municipality string  `json:"municipality,omitempty"`
	PriceMin     float64 `json:"price_min,omitempty"`
	PriceMax     float64 `json:"price_max,omitempty"`
	FarmerID     string  `json:"farmer_id,omitempty"`
	SortBy       string  `json:"sort_by,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	Page         int     `json:"page,omitempty"`
	PageSize     int     `json:"page_size,omitempty"`
}

// Save in elastic search
type IndexOfferingRequest struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Price          float64 `json:"price"`
	Type           string  `json:"type"`
	ImageURL       string  `json:"image_url"`
	UserID         string  `json:"user_id"`
	FarmerName     string  `json:"farmer_name"`
	FarmerVerified bool    `json:"farmer_verified"`
	Department     string  `json:"department"`
	Municipality   string  `json:"municipality"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
}

type ESClient struct {
	Username            string
	Password            string
	Endpoint1           string
	Endpoint2           string
	MaxIdleConnsPerHost int
	Index               string
}
