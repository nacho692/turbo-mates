package discovery

import (
	"encoding/json"
	"net/http"
)

func Handler(discovery *Discovery) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			marshal, err := json.Marshal(discovery.Stats())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			_, _ = w.Write(marshal)
		})
}
