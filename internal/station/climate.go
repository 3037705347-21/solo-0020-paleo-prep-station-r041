package station

import (
	"context"
	"time"

	"example.com/paleo-prep-station/internal/contracts"
	"example.com/paleo-prep-station/internal/domain"
	"example.com/paleo-prep-station/internal/repository"
)

type ClimateReadingResult struct {
	Reading   domain.ClimateReading    `json:"reading"`
	Excursion *domain.ClimateExcursion `json:"excursion,omitempty"`
}

type ClimateStatusSnapshot struct {
	SpecimenID string                    `json:"specimen_id"`
	Excursions []domain.ClimateExcursion `json:"excursions"`
}

func (s *Service) RegisterCalibration(
	ctx context.Context,
	request contracts.RegisterCalibrationRequest,
) (domain.SensorCalibration, error) {
	if err := ctx.Err(); err != nil {
		return domain.SensorCalibration{}, err
	}
	validUntil, err := time.Parse(time.RFC3339, request.ValidUntil)
	if err != nil {
		return domain.SensorCalibration{}, contracts.Invalid(
			"invalid_calibration_time",
			"校准截止时间必须使用 RFC3339 格式",
		)
	}
	calibration, err := domain.NewSensorCalibration(request.SensorID, request.CellID, validUntil)
	if err != nil {
		return domain.SensorCalibration{}, err
	}
	err = s.repo.Mutate(func(m *repository.Mutation) error {
		m.PutCalibration(calibration)
		return nil
	})
	return calibration, err
}

func (s *Service) RecordClimateReading(
	ctx context.Context,
	specimenID string,
	request contracts.ClimateReadingRequest,
) (ClimateReadingResult, error) {
	if err := ctx.Err(); err != nil {
		return ClimateReadingResult{}, err
	}
	now := s.now()
	reading, err := domain.NewClimateReading(
		request.SensorID,
		request.CellID,
		request.TemperatureC,
		request.HumidityPct,
		request.RecordedBy,
		now,
	)
	if err != nil {
		return ClimateReadingResult{}, err
	}
	if err := s.repo.Read(func(m *repository.Mutation) error {
		if _, ok := m.GetSpecimen(specimenID); !ok {
			return notFoundSpecimen(specimenID)
		}
		calibration, ok := m.GetCalibration(reading.SensorID, reading.CellID)
		if !ok {
			return notFoundCalibration(reading.SensorID, reading.CellID)
		}
		if !calibration.ValidAt(now) {
			return contracts.Conflict("climate_calibration_expired", "探头校准已经失效")
		}
		return nil
	}); err != nil {
		return ClimateReadingResult{}, err
	}

	result := ClimateReadingResult{Reading: reading}
	if reading.IsWithinLimits() {
		return result, nil
	}

	var excursion domain.ClimateExcursion
	err = s.repo.Mutate(func(m *repository.Mutation) error {
		if existing, ok := m.FindOpenClimateExcursion(specimenID); ok {
			excursion = existing
			return nil
		}
		created, err := domain.NewClimateExcursion(
			s.nextID("climate-excursion"),
			specimenID,
			reading,
			now,
		)
		if err != nil {
			return err
		}
		m.PutClimateExcursion(created)
		excursion = created
		return nil
	})
	if err != nil {
		return ClimateReadingResult{}, err
	}

	// The hold index is refreshed separately so release can make its gate
	// decision without scanning the full excursion history.
	err = s.repo.Mutate(func(m *repository.Mutation) error {
		m.SetClimateHold(specimenID, excursion.ID)
		return nil
	})
	if err != nil {
		return ClimateReadingResult{}, err
	}
	result.Excursion = &excursion
	return result, nil
}

func (s *Service) ResolveClimateExcursion(
	ctx context.Context,
	excursionID string,
	request contracts.ResolveClimateExcursionRequest,
) (domain.ClimateExcursion, error) {
	if err := ctx.Err(); err != nil {
		return domain.ClimateExcursion{}, err
	}
	now := s.now()
	var resolved domain.ClimateExcursion
	err := s.repo.Mutate(func(m *repository.Mutation) error {
		excursion, ok := m.GetClimateExcursion(excursionID)
		if !ok {
			return notFoundClimateExcursion(excursionID)
		}
		if !excursion.IsOpen() {
			return contracts.Conflict("excursion_not_resolvable", "环境偏离已经解除")
		}
		reading, err := domain.NewClimateReading(
			request.SensorID,
			request.CellID,
			request.TemperatureC,
			request.HumidityPct,
			request.ResolvedBy,
			now,
		)
		if err != nil {
			return err
		}
		if reading.SensorID != excursion.SensorID || reading.CellID != excursion.CellID {
			return contracts.Conflict(
				"climate_source_mismatch",
				"解除读数必须来自形成偏离的探头和保存区",
			)
		}
		if !reading.IsWithinLimits() {
			return contracts.Conflict("climate_reading_out_of_range", "当前读数仍超出允许范围")
		}
		if _, ok := m.GetCalibration(reading.SensorID, reading.CellID); !ok {
			return notFoundCalibration(reading.SensorID, reading.CellID)
		}
		completed, err := excursion.WithResolution(request.ResolvedBy, now)
		if err != nil {
			return err
		}
		_ = completed
		m.PutClimateExcursion(excursion)
		m.ClearClimateHold(excursion.SpecimenID)
		resolved = excursion
		return nil
	})
	return resolved, err
}

func (s *Service) GetClimateStatus(
	ctx context.Context,
	specimenID string,
) (ClimateStatusSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return ClimateStatusSnapshot{}, err
	}
	var snapshot ClimateStatusSnapshot
	err := s.repo.Read(func(m *repository.Mutation) error {
		if _, ok := m.GetSpecimen(specimenID); !ok {
			return notFoundSpecimen(specimenID)
		}
		snapshot = ClimateStatusSnapshot{
			SpecimenID: specimenID,
			Excursions: m.ListClimateExcursions(specimenID),
		}
		return nil
	})
	return snapshot, err
}
