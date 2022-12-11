package model

type QuestionData struct {
	Questions []Question `json:"questions"`
}

type Question struct {
	Question string `json:"question"`
	Answer   Answer `json:"answer"`
}

type Answer struct {
	Low    *string `json:"low,omitempty"`
	Medium *string `json:"medium,omitempty"`
	High   *string `json:"high,omitempty"`
}
