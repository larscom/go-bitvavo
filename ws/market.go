package ws

import mapset "github.com/deckarep/golang-set/v2"

func getUniqueMarkets(markets []string) []string {
	return mapset.NewSet(markets...).ToSlice()
}
