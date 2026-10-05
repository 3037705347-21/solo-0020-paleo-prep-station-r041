package station

import "example.com/paleo-prep-station/internal/contracts"

func notFoundSpecimen(id string) error {
	return contracts.NotFound("specimen_not_found", "未找到标本 "+id)
}

func notFoundPlan(id string) error {
	return contracts.NotFound("plan_not_found", "未找到修复方案 "+id)
}

func notFoundClimateExcursion(id string) error {
	return contracts.NotFound("climate_excursion_not_found", "未找到环境偏离 "+id)
}

func notFoundCalibration(sensorID, cellID string) error {
	return contracts.Conflict(
		"climate_calibration_missing",
		"保存区 "+cellID+" 的探头 "+sensorID+" 尚无有效校准",
	)
}
