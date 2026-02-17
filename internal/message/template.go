package message

func BuildTextTemplate(text string) *TemplateObject {
	return &TemplateObject{
		ObjectType: "text",
		Text:       text,
		Link:       &LinkObject{},
	}
}

func BuildLinkTemplate(text, url string) *TemplateObject {
	return &TemplateObject{
		ObjectType: "text",
		Text:       text,
		Link:       &LinkObject{WebURL: url, MobileWebURL: url},
		Buttons: []Button{
			{
				Title: "Open Link",
				Link:  LinkObject{WebURL: url, MobileWebURL: url},
			},
		},
	}
}
