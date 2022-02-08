package protocol

type WSMessage[T any] struct {
	Header  *WSHeader `json:"header"`
	Message *T        `json:"message"`
}

type WSHeader struct {
	Id        string `json:"id"`
	Version   string `json:"version"`
	UserId    string `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
}
