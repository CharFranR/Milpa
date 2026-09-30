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

// client input — parsed from query params
type SearchRequest struct {
	Term         string          `json:"term"`
	Type         string          `json:"type,omitempty"`
	CategoryID   string          `json:"category_id,omitempty"`
	Department   string          `json:"department,omitempty"`
	Municipality string          `json:"municipality,omitempty"`
	PriceMin     *float64        `json:"price_min,omitempty"`
	PriceMax     *float64        `json:"price_max,omitempty"`
	FarmerID     string          `json:"farmer_id,omitempty"`
	SortBy       SearchSortField `json:"sort_by,omitempty"`
	Latitude     float64         `json:"latitude,omitempty"`
	Longitude    float64         `json:"longitude,omitempty"`
	Page         int             `json:"page,omitempty"`
	PageSize     int             `json:"page_size,omitempty"`
}

// application-layer query — built by the use case with business rules
type SearchQuery struct {
	Term       string
	Filters    SearchFilters
	Sort       SearchSort
	Pagination SearchPagination
	ScoreRules []ScoreRule
}

type SearchFilters struct {
	Type         string
	CategoryID   string
	Department   string
	Municipality string
	PriceMin     *float64
	PriceMax     *float64
	FarmerID     string
}

type SearchSortField string

const (
	SortRelevance SearchSortField = "relevance"
	SortPriceAsc  SearchSortField = "price_asc"
	SortPriceDesc SearchSortField = "price_desc"
	SortProximity SearchSortField = "proximity"
)

type SearchSort struct {
	Field     SearchSortField
	Latitude  float64
	Longitude float64
}

type SearchPagination struct {
	Page     int
	PageSize int
}

// ScoreRule — relevance decision owned by the application layer.
// The adapter translates this into ES function_score without interpreting it.
type ScoreRule struct {
	Field string  // document field to match (e.g. "farmer_verified")
	Value any     // value to match (e.g. true)
	Boost float64 // score multiplier (e.g. 1.5)
}

// paginated response
type SearchResponse struct {
	Results    []FuzzySearchDto `json:"results"`
	TotalHits  int64            `json:"total_hits"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// Save in elastic search
type IndexOfferingRequest struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Price          float64   `json:"price"`
	Type           string    `json:"type"`
	CategoryID     string    `json:"category_id"`
	ImageURL       string    `json:"image_url"`
	UserID         string    `json:"user_id"`
	FarmerName     string    `json:"farmer_name"`
	FarmerVerified bool      `json:"farmer_verified"`
	Department     string    `json:"department"`
	Municipality   string    `json:"municipality"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Location       *GeoPoint `json:"location,omitempty"`
}

// GeoPoint is the Elasticsearch geo_point object form, built from the
// latitude and longitude the user address carries. It is the field the
// proximity sort resolves against, so it has to be present on every document
// that carries coordinates — otherwise _geo_distance has nothing to measure.
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// NewGeoPoint returns the geo_point for a pair of coordinates, or nil when the
// pair carries no location at all. A zero pair is treated as absent because it
// is indistinguishable from a farmer who never supplied coordinates, and
// indexing it would plant the producer at null island.
func NewGeoPoint(latitude, longitude float64) *GeoPoint {
	if latitude == 0 && longitude == 0 {
		return nil
	}
	return &GeoPoint{Lat: latitude, Lon: longitude}
}

type ESClient struct {
	Username            string
	Password            string
	Endpoint1           string
	Endpoint2           string
	MaxIdleConnsPerHost int
	Index               string
}
