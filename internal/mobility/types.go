package mobility

type DirectionsResult struct {
	TransID string  `json:"trans_id"`
	Routes  []Route `json:"routes"`
}

type Route struct {
	ResultCode int       `json:"result_code"`
	ResultMsg  string    `json:"result_msg"`
	Summary    Summary   `json:"summary"`
	Sections   []Section `json:"sections"`
}

type Summary struct {
	Origin      Location `json:"origin"`
	Destination Location `json:"destination"`
	Distance    int      `json:"distance"`
	Duration    int      `json:"duration"`
	Fare        Fare     `json:"fare"`
}

type Fare struct {
	Taxi int `json:"taxi"`
	Toll int `json:"toll"`
}

type Location struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type Section struct {
	Distance int    `json:"distance"`
	Duration int    `json:"duration"`
	Roads    []Road `json:"roads"`
}

type Road struct {
	Name     string `json:"name"`
	Distance int    `json:"distance"`
	Duration int    `json:"duration"`
}
