package contracts

type RegisterSpecimenRequest struct {
	CatalogNumber string `json:"catalog_number"`
	FieldLabel    string `json:"field_label"`
	TaxonHint     string `json:"taxon_hint"`
	MatrixKind    string `json:"matrix_kind"`
	JacketNotes   string `json:"jacket_notes"`
}

type CreatePlanRequest struct {
	Objective      string `json:"objective"`
	Method         string `json:"method"`
	Adhesive       string `json:"adhesive"`
	TargetExposure int    `json:"target_exposure"`
}

type ApprovalRequest struct {
	ApprovedBy string `json:"approved_by"`
}

type ProgressRequest struct {
	Stage      string `json:"stage"`
	Percent    int    `json:"percent"`
	Note       string `json:"note"`
	RecordedBy string `json:"recorded_by"`
}

type ReleaseRequest struct {
	Purpose    string `json:"purpose"`
	Recipient  string `json:"recipient"`
	ApprovedBy string `json:"approved_by"`
}

type RegisterCalibrationRequest struct {
	SensorID   string `json:"sensor_id"`
	CellID     string `json:"cell_id"`
	ValidUntil string `json:"valid_until"`
}

type ClimateReadingRequest struct {
	SensorID     string  `json:"sensor_id"`
	CellID       string  `json:"cell_id"`
	TemperatureC float64 `json:"temperature_c"`
	HumidityPct  float64 `json:"humidity_pct"`
	RecordedBy   string  `json:"recorded_by"`
}

type ResolveClimateExcursionRequest struct {
	SensorID     string  `json:"sensor_id"`
	CellID       string  `json:"cell_id"`
	TemperatureC float64 `json:"temperature_c"`
	HumidityPct  float64 `json:"humidity_pct"`
	ResolvedBy   string  `json:"resolved_by"`
}
