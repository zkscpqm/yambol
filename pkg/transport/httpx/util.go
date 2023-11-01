package httpx

import (
	"strconv"
	"strings"
)

func UrlJoin(url string, components ...any) string {
	url = strings.TrimSuffix(url, "/")
	for _, component := range components {

		switch component.(type) {
		case string:
			sComp := strings.TrimSuffix(component.(string), "/")
			url += "/" + strings.TrimPrefix(sComp, "/")
		case int:
			url += "/" + strconv.Itoa(component.(int))
		}
	}
	return url
}
