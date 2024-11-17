package model

type Experience struct {
	Workplace   string   `json:"workplace"`
	Position    string   `json:"position"`
	Duration    []string `json:"duration"`
	Tech        []string `json:"tech"`
	Description []string `json:"description"`
	URL         string   `json:"url"`
}

func (e *Experience) String() string {
	workplace := e.Workplace
	position := e.Position
	description := e.Description

	text := workplace + " - " + position + "\n"
	for _, desc := range description {
		text += "- " + desc + "\n"
	}
	return text
}

type Project struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Pic         string   `json:"pic"`
	Tech        []string `json:"tech"`
	Links       []Link   `json:"links"`
}

func (p *Project) String() string {
	name := p.Name
	description := p.Description
	tech := p.Tech
	links := p.Links

	text := name + " - " + description + "\n"
	for _, t := range tech {
		text += "- " + t + "\n"
	}
	text += StringifyLinks(links)
	return text
}

type Link struct {
	Label string `json:"label"`
	Icon  string `json:"icon"`
	URL   string `json:"url"`
}

func StringifyLinks(links []Link) string {
	var text string
	for _, l := range links {
		text += "- " + l.Label + "\n"
	}
	return text
}

func (l *Link) String() string {
	return l.Label + ": " + l.URL
}
