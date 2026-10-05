package httpapi

import (
	"net/http"

	"example.com/paleo-prep-station/internal/contracts"
)

func (api *API) registerCalibration(w http.ResponseWriter, r *http.Request) {
	var request contracts.RegisterCalibrationRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	value, err := api.service.RegisterCalibration(r.Context(), request)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (api *API) recordClimateReading(w http.ResponseWriter, r *http.Request) {
	var request contracts.ClimateReadingRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	value, err := api.service.RecordClimateReading(r.Context(), r.PathValue("id"), request)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (api *API) resolveClimateExcursion(w http.ResponseWriter, r *http.Request) {
	var request contracts.ResolveClimateExcursionRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	value, err := api.service.ResolveClimateExcursion(
		r.Context(),
		r.PathValue("id"),
		request,
	)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (api *API) getClimateStatus(w http.ResponseWriter, r *http.Request) {
	value, err := api.service.GetClimateStatus(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
