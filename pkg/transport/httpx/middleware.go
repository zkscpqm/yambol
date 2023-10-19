package httpx

import (
	"fmt"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log details of the request
		fmt.Printf("Connection from %s at %s: %s %s\n", r.RemoteAddr, time.Now().Format(time.RFC3339), r.Method, r.URL.Path)

		// Call the next handler (which can be another middleware or the final handler)
		next.ServeHTTP(w, r)
	})
}
