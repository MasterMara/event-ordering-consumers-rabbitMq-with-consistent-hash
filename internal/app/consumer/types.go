package consumer

import "time"

//Todo: Change Here

type OrderLineEvent struct {
	Message       string    `json:"message,omitempty"`
	CorrelationId string    `json:"correlationId"`
	SentTime      time.Time `json:"sentTime"`
	MessageType   []string  `json:"messageType"`
}

type Line struct { // Todo: Configure et
	name string
}
