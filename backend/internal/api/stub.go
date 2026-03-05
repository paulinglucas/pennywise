package api

import "net/http"

type StubHandler struct{}

func (s *StubHandler) ListAccounts(w http.ResponseWriter, r *http.Request, params ListAccountsParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) DeleteAccount(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetAccount(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) UpdateAccount(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ListAlerts(w http.ResponseWriter, r *http.Request, params ListAlertsParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) MarkAlertRead(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ListAssets(w http.ResponseWriter, r *http.Request, params ListAssetsParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetAssetAllocation(w http.ResponseWriter, r *http.Request, params GetAssetAllocationParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) DeleteAsset(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetAsset(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) UpdateAsset(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetAssetHistory(w http.ResponseWriter, r *http.Request, id IdParam, params GetAssetHistoryParams) {
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

func (s *StubHandler) ListTransactions(w http.ResponseWriter, r *http.Request, params ListTransactionsParams) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) ImportTransactions(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) GetTransaction(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}

func (s *StubHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request, id IdParam) {
	WriteError(w, http.StatusNotImplemented, INTERNALERROR, "Not implemented", "")
}
