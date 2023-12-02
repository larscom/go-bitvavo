package jsond

import "fmt"

type BitvavoErr struct {
	Code    int    `json:"errorCode"`
	Message string `json:"error"`
	Action  string `json:"action"`
}

func (b *BitvavoErr) Error() string {
	return fmt.Sprintf("Error %d: %s. Action: %s", b.Code, b.Message, b.Action)
}
