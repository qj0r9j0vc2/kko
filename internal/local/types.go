package local

type Place struct {
	ID           string `json:"id"`
	PlaceName    string `json:"place_name"`
	CategoryName string `json:"category_name"`
	CategoryCode string `json:"category_group_code"`
	Phone        string `json:"phone"`
	AddressName  string `json:"address_name"`
	RoadAddress  string `json:"road_address_name"`
	X            string `json:"x"`
	Y            string `json:"y"`
	PlaceURL     string `json:"place_url"`
	Distance     string `json:"distance"`
}

type SearchResult struct {
	Documents []Place `json:"documents"`
	Meta      Meta    `json:"meta"`
}

type Meta struct {
	TotalCount    int  `json:"total_count"`
	PageableCount int  `json:"pageable_count"`
	IsEnd         bool `json:"is_end"`
}

type AddressResult struct {
	Documents []Address `json:"documents"`
	Meta      Meta      `json:"meta"`
}

type Address struct {
	AddressName string    `json:"address_name"`
	X           string    `json:"x"`
	Y           string    `json:"y"`
	AddressType string    `json:"address_type"`
	RoadAddress *RoadAddr `json:"road_address"`
}

type RoadAddr struct {
	AddressName string `json:"address_name"`
	Region1     string `json:"region_1depth_name"`
	Region2     string `json:"region_2depth_name"`
	Region3     string `json:"region_3depth_name"`
	RoadName    string `json:"road_name"`
	BuildingNo  string `json:"main_building_no"`
}

var CategoryCodes = map[string]string{
	"cafe":        "CE7",
	"convenience": "CS2",
	"pharmacy":    "PM9",
	"food":        "FD6",
	"parking":     "PK6",
	"hospital":    "HP8",
	"bank":        "BK9",
	"gas":         "OL7",
	"subway":      "SW8",
	"restaurant":  "FD6",
	"mart":        "MT1",
	"school":      "SC4",
	"hotel":       "AD5",
	"culture":     "CT1",
}
