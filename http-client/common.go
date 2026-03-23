package http_client

type PayloadBodyType int

const (
	JsonPayload PayloadBodyType = iota
	FormPayload
)

func (c PayloadBodyType) getContentType() string {
	return [...]string{
		"application/json",
		"application/x-www-form-urlencoded",
	}[c]
}
