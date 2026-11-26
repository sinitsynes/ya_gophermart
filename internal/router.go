package internal

import "net/http"

func NewRouter(muxes ...*http.ServeMux) *http.ServeMux {
	mux := http.NewServeMux()
	for _, m := range muxes {
		mux.Handle("/api", m)
	}

	return mux
}
