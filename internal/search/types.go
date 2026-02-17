package search

type Meta struct {
	TotalCount    int  `json:"total_count"`
	PageableCount int  `json:"pageable_count"`
	IsEnd         bool `json:"is_end"`
}

type WebResult struct {
	Title    string `json:"title"`
	Contents string `json:"contents"`
	URL      string `json:"url"`
	Datetime string `json:"datetime"`
}

type BlogResult struct {
	Title     string `json:"title"`
	Contents  string `json:"contents"`
	URL       string `json:"url"`
	Blogname  string `json:"blogname"`
	Datetime  string `json:"datetime"`
	Thumbnail string `json:"thumbnail"`
}

type CafeResult struct {
	Title     string `json:"title"`
	Contents  string `json:"contents"`
	URL       string `json:"url"`
	Cafename  string `json:"cafename"`
	Datetime  string `json:"datetime"`
	Thumbnail string `json:"thumbnail"`
}

type WebSearchResponse struct {
	Documents []WebResult `json:"documents"`
	Meta      Meta        `json:"meta"`
}

type BlogSearchResponse struct {
	Documents []BlogResult `json:"documents"`
	Meta      Meta         `json:"meta"`
}

type CafeSearchResponse struct {
	Documents []CafeResult `json:"documents"`
	Meta      Meta         `json:"meta"`
}
