package domain

import (
	"math"
	"time"

	"example.com/paleo-prep-station/internal/contracts"
)

const (
	minSafeTemperature = 15.0
	maxSafeTemperature = 30.0
	minSafeHumidity    = 20.0
	maxSafeHumidity    = 70.0
)

type SensorCalibration struct {
	SensorID   string    `json:"sensor_id"`
	CellID     string    `json:"cell_id"`
	ValidUntil time.Time `json:"valid_until"`
}

func NewSensorCalibration(
	sensorID string,
	cellID string,
	validUntil time.Time,
) (SensorCalibration, error) {
	sensor, err := requiredText("探头标识", sensorID, 64)
	if err != nil {
		return SensorCalibration{}, err
	}
	cell, err := requiredText("保存区", cellID, 64)
	if err != nil {
		return SensorCalibration{}, err
	}
	if validUntil.IsZero() {
		return SensorCalibration{}, contracts.Invalid("invalid_calibration", "校准截止时间不能为空")
	}
	return SensorCalibration{
		SensorID:   sensor,
		CellID:     cell,
		ValidUntil: validUntil.UTC(),
	}, nil
}

func (c SensorCalibration) ValidAt(at time.Time) bool {
	return !at.UTC().After(c.ValidUntil)
}

type ClimateReading struct {
	SensorID     string    `json:"sensor_id"`
	CellID       string    `json:"cell_id"`
	TemperatureC float64   `json:"temperature_c"`
	HumidityPct  float64   `json:"humidity_pct"`
	RecordedBy   string    `json:"recorded_by"`
	RecordedAt   time.Time `json:"recorded_at"`
}

func NewClimateReading(
	sensorID string,
	cellID string,
	temperatureC float64,
	humidityPct float64,
	recordedBy string,
	now time.Time,
) (ClimateReading, error) {
	sensor, err := requiredText("探头标识", sensorID, 64)
	if err != nil {
		return ClimateReading{}, err
	}
	cell, err := requiredText("保存区", cellID, 64)
	if err != nil {
		return ClimateReading{}, err
	}
	recorder, err := requiredText("记录人", recordedBy, 80)
	if err != nil {
		return ClimateReading{}, err
	}
	if math.IsNaN(temperatureC) || math.IsInf(temperatureC, 0) ||
		math.IsNaN(humidityPct) || math.IsInf(humidityPct, 0) {
		return ClimateReading{}, contracts.Invalid("invalid_climate_reading", "环境读数必须是有效数字")
	}
	if temperatureC < -50 || temperatureC > 80 || humidityPct < 0 || humidityPct > 100 {
		return ClimateReading{}, contracts.Invalid("invalid_climate_reading", "环境读数超出仪器量程")
	}
	return ClimateReading{
		SensorID:     sensor,
		CellID:       cell,
		TemperatureC: temperatureC,
		HumidityPct:  humidityPct,
		RecordedBy:   recorder,
		RecordedAt:   now.UTC(),
	}, nil
}

func (r ClimateReading) IsWithinLimits() bool {
	return r.TemperatureC >= minSafeTemperature &&
		r.TemperatureC <= maxSafeTemperature &&
		r.HumidityPct >= minSafeHumidity &&
		r.HumidityPct <= maxSafeHumidity
}

type ClimateExcursionStatus string

const (
	ClimateOpen     ClimateExcursionStatus = "open"
	ClimateResolved ClimateExcursionStatus = "resolved"
)

type ClimateExcursion struct {
	ID         string                 `json:"id"`
	SpecimenID string                 `json:"specimen_id"`
	CellID     string                 `json:"cell_id"`
	SensorID   string                 `json:"sensor_id"`
	Status     ClimateExcursionStatus `json:"status"`
	OpenedAt   time.Time              `json:"opened_at"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy string                 `json:"resolved_by,omitempty"`
	Version    int                    `json:"version"`
}

func NewClimateExcursion(
	id string,
	specimenID string,
	reading ClimateReading,
	now time.Time,
) (ClimateExcursion, error) {
	if id == "" {
		return ClimateExcursion{}, contracts.Internal("missing_generated_id", "环境偏离标识未生成", nil)
	}
	specimen, err := requiredText("标本标识", specimenID, 80)
	if err != nil {
		return ClimateExcursion{}, err
	}
	return ClimateExcursion{
		ID:         id,
		SpecimenID: specimen,
		CellID:     reading.CellID,
		SensorID:   reading.SensorID,
		Status:     ClimateOpen,
		OpenedAt:   now.UTC(),
		Version:    1,
	}, nil
}

func (e ClimateExcursion) IsOpen() bool {
	return e.Status == ClimateOpen
}

func (e ClimateExcursion) WithResolution(
	resolvedBy string,
	at time.Time,
) (ClimateExcursion, error) {
	if !e.IsOpen() {
		return ClimateExcursion{}, contracts.Conflict("excursion_not_resolvable", "环境偏离已经解除")
	}
	actor, err := requiredText("处置人", resolvedBy, 80)
	if err != nil {
		return ClimateExcursion{}, err
	}
	stamp := at.UTC()
	resolved := e
	resolved.Status = ClimateResolved
	resolved.ResolvedAt = &stamp
	resolved.ResolvedBy = actor
	resolved.Version++
	return resolved, nil
}
