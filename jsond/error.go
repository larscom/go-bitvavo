package jsond

import (
	"fmt"

	"github.com/larscom/go-bitvavo/v2/util"
)

type BitvavoErr struct {
	Code    int    `json:"errorCode"`
	Message string `json:"error"`
	Action  string `json:"action"`
}

func (b *BitvavoErr) Error() string {
	msg := fmt.Sprintf("code %d: %s", b.Code, b.Message)
	return fmt.Sprint(util.IfOrElse(len(b.Action) > 0, func() string { return fmt.Sprintf("%s action: %s", msg, b.Action) }, msg))
}
