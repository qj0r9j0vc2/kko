package message

type TemplateObject struct {
	ObjectType string      `json:"object_type"`
	Text       string      `json:"text"`
	Link       *LinkObject `json:"link"`
	Buttons    []Button    `json:"buttons,omitempty"`
}

type LinkObject struct {
	WebURL       string `json:"web_url,omitempty"`
	MobileWebURL string `json:"mobile_web_url,omitempty"`
}

type Button struct {
	Title string     `json:"title"`
	Link  LinkObject `json:"link"`
}

type SendResult struct {
	ResultCode int `json:"result_code"`
}
