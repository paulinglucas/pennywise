package api

import "net/http"

type StubHandler struct{}

func (s *StubHandler) PostVitals(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}
