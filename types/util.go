package types

import (
	"github.com/larscom/go-bitvavo/v2/util"
)

func getOrEmpty[T any](key string, data map[string]any) T {
	var empty T
	value, exist := data[key]
	return util.IfOrElse(exist && value != nil, func() T { return value.(T) }, empty)
}
