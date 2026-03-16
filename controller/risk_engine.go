package controller

import (
	"math"
	"pv-risk/models"
	"time"
)

func CalculateRisk(pv models.Assessment) models.Assessment {

	// Prevent division by zero
	if pv.CorrosionRate <= 0 {
		pv.CorrosionRate = 0.0001
	}

	// ============ Risk Calculation Logic ============
	// Remaining Life (RL) = (Actual Thickness - Minimum Thickness) / Corrosion Rate
	rl := (pv.ThicknessActual - pv.ThicknessMin) / pv.CorrosionRate
	age := time.Now().Year() - pv.YearBuilt

	// LoF Calculation - combines multiple factors
	rlScore := scoreRemainingLife(rl)
	ageScore := scoreAge(age)
	crScore := scoreCR(pv.CorrosionRate)
	damageScore := scoreDamage(pv.DamageMechanism)

	baseLoF := float64(rlScore+ageScore+crScore+damageScore) / 4.0
	adjustedLoF := baseLoF * inspectionModifier(pv.InspectionQuality)

	if adjustedLoF > 5 {
		adjustedLoF = 5
	}

	// CoF Calculation - combines pressure, fluid, inventory, and criticality
	pressureScore := scorePressure(pv.OperatingPressure)
	fluidScore := scoreFluid(pv.FluidType)
	invScore := scoreInventory(pv.InventoryVolume)

	criticalScore := 2
	if pv.IsCritical {
		criticalScore = 5
	}

	cofFloat :=
		(float64(pressureScore) * 0.3) +
			(float64(fluidScore) * 0.3) +
			(float64(invScore) * 0.2) +
			(float64(criticalScore) * 0.2)

	if cofFloat > 5 {
		cofFloat = 5
	}

	// risk index calculation
	riskIndex := adjustedLoF * cofFloat

	// reranking based on thresholds
	if rl < 2 {
		riskIndex += 2.0
	} else if rl < 5 {
		riskIndex += 1.0
	}

	if pv.IsCritical && pv.FluidType == "hydrocarbon" {
		riskIndex += 1.5
	}

	if riskIndex > 25 {
		riskIndex = 25
	}

	// Matrix-based risk level assignment
	matrixRow := int(math.Ceil(adjustedLoF))
	matrixCol := int(math.Ceil(cofFloat))

	matrixScore := matrixRow * matrixCol
	riskLevel := mapRisk(matrixScore)

	// Next Inspection
	nextInsp := nextInspection(rl, riskLevel)

	pv.RemainingLife = rl
	pv.LoF = adjustedLoF
	pv.CoF = cofFloat
	pv.RiskIndex = riskIndex
	pv.MatrixRow = matrixRow
	pv.MatrixCol = matrixCol
	pv.RiskLevel = riskLevel
	pv.NextInspection = int(nextInsp)
	return pv
}

func scoreRemainingLife(rl float64) int {
	switch {
	case rl > 15:
		return 1
	case rl > 10:
		return 2
	case rl > 5:
		return 3
	case rl > 2:
		return 4
	default:
		return 5
	}
}

func scoreAge(age int) int {
	switch {
	case age < 10:
		return 1
	case age < 20:
		return 2
	case age < 30:
		return 3
	case age < 40:
		return 4
	default:
		return 5
	}
}

func scoreCR(cr float64) int {
	switch {
	case cr < 0.1:
		return 1
	case cr < 0.3:
		return 2
	case cr < 0.6:
		return 3
	case cr < 1:
		return 4
	default:
		return 5
	}
}

func scorePressure(p float64) int {
	switch {
	case p < 5:
		return 1
	case p < 15:
		return 2
	case p < 30:
		return 3
	case p < 60:
		return 4
	default:
		return 5
	}
}

func scoreFluid(f string) int {
	switch f {
	case "water":
		return 1
	case "steam":
		return 2
	case "oil":
		return 4
	case "hydrocarbon":
		return 5
	default:
		return 3
	}
}

func mapRisk(index int) string {
	switch {
	case index <= 5:
		return "Low"
	case index <= 10:
		return "Medium"
	case index <= 15:
		return "High"
	default:
		return "Extreme"
	}
}

func scoreDamage(mech string) int {
	switch mech {
	case "general":
		return 2
	case "pitting":
		return 3
	case "cui":
		return 4
	case "hic", "scc":
		return 5
	case "fatigue":
		return 4
	default:
		return 3
	}
}

func inspectionModifier(q string) float64 {
	switch q {
	case "excellent":
		return 0.7
	case "good":
		return 0.85
	case "fair":
		return 1.0
	case "poor":
		return 1.2
	default:
		return 1.0
	}
}

func scoreInventory(v float64) int {
	switch {
	case v < 1:
		return 1
	case v < 5:
		return 2
	case v < 20:
		return 3
	case v < 50:
		return 4
	default:
		return 5
	}
}

func nextInspection(rl float64, level string) float64 {

	var factor float64

	switch level {
	case "Low":
		factor = 0.8
	case "Medium":
		factor = 0.6
	case "High":
		factor = 0.4
	default:
		factor = 0.2
	}

	return rl * factor
}
