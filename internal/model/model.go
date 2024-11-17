package model

type Experience struct {
	Workplace   string   `json:"workplace"`
	Position    string   `json:"position"`
	Duration    []string `json:"duration"`
	Tech        []string `json:"tech"`
	Description []string `json:"description"`
	URL         string   `json:"url"`
}

type Project struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Pic         string   `json:"pic"`
	Tech        []string `json:"tech"`
	Links       []Link   `json:"links"`
}

type Link struct {
	Label string `json:"label"`
	Icon  string `json:"icon"`
	URL   string `json:"url"`
}
