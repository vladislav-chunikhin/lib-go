package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

func Handler(hc HealthChecker) http.Handler {
	r := chi.NewRouter()
	r.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, hc.CheckLive)
	})
	r.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, hc.CheckReady)
	})
	return r
}

func handle(w http.ResponseWriter, r *http.Request, checks func() map[string]error) {
	checkResults := make(map[string]string)
	status := http.StatusOK
	res := checks()
	for name, err := range res {
		if err != nil {
			status = http.StatusServiceUnavailable
			checkResults[name] = err.Error()
		}
	}

	// write out the response code and content type header
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// unless ?full=1, return an empty body. Kubernetes only cares about the
	// HTTP status code, so we won't waste bytes on the full body.
	if r.URL.Query().Get("full") != "1" {
		w.Write([]byte("{}\n"))
		return
	}

	// otherwise, write the JSON body ignoring any encoding errors (which
	// shouldn't really be possible since we're encoding a map[string]string).
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(checkResults)
}
