package picosynth

type FilterMode uint

const (
	FilterNoFilter FilterMode = iota
	FilterChamberlinLowPass
	FilterChamberlinHighPass
	FilterChamberlinBandPass
	FilterChamberlinNotch
)
