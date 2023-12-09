package util

func GetOrEmptyString(key string, data map[string]any) string {
	value, exist := data[key]
	return IfOrElse(exist, func() string { return value.(string) }, "")
}
