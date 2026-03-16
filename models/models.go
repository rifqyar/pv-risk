package models

type Assessment struct {
	ID                int
	TagNumber         string
	YearBuilt         int
	ThicknessActual   float64
	ThicknessMin      float64
	CorrosionRate     float64
	OperatingPressure float64
	FluidType         string
	IsCritical        bool
	DamageMechanism   string
	DamageFactor      int
	InspectionScore   int
	InspectionQuality string
	InventoryVolume   float64
	RemainingLife     float64
	LoF               float64
	CoF               float64
	RiskIndex         float64
	RiskLevel         string
	NextInspection    int
	MatrixRow         int
	MatrixCol         int
}

type ThicknessHistory struct {
	ID         int
	TagNumber  string
	Thickness  float64
	MeasuredAt string
}
