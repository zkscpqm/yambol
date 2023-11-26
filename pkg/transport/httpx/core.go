package httpx

import (
	"net/http"
	"regexp"
	"strings"
)

type Server interface {
	ListenAndServe(port int, key, cert string) error
	ListenAndServeInsecure(port int) error
}

type HandlerFunc = func(w http.ResponseWriter, r *http.Request) Response

var forbiddenQueueNames = []string{
	"broadcast",
}

func normalizeQueueName(name string) string {
	return strings.ToLower(
		strings.TrimPrefix(
			strings.TrimSuffix(
				name,
				"/",
			),
			"/",
		),
	)
}

func IsValidQueueName(name string) bool {
	name = normalizeQueueName(name)
	if strings.TrimSpace(name) == "" {
		return false
	}

	for _, forbiddenName := range forbiddenQueueNames {
		if name == forbiddenName {
			return false
		}
	}

	re := regexp.MustCompile(`^[\w-]+$`)

	return re.MatchString(name)
}
