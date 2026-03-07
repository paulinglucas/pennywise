package api

import "net/http"

type StubHandler struct{}

func (s *StubHandler) ListAlerts(w http.ResponseWriter, r *http.Request, params ListAlertsParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) MarkAlertRead(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetNetWorthHistory(w http.ResponseWriter, r *http.Request, params GetNetWorthHistoryParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ExportCsv(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ListGoals(w http.ResponseWriter, r *http.Request, params ListGoalsParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ReorderGoals(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) DeleteGoal(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetGoal(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) UpdateGoal(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ComputeProjection(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ListRecurring(w http.ResponseWriter, r *http.Request, params ListRecurringParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) CreateRecurring(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) DeleteRecurring(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) UpdateRecurring(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) PostVitals(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}
