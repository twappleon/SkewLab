package research100
import("time")
type CalibrationReview struct{
 Status string `json:"status"`
 WindowScores []WindowCalibration `json:"window_scores"`
 Reliability []ReliabilityBin `json:"reliability_bins"`
 PairedBaseline PairedBrier `json:"paired_brier"`
 FlatBaseline FlatBrierBaseline `json:"coin_flip_baseline"`
 ScoreRange BidAskScoreRange `json:"bid_ask_score_range"`
 Uncertainty []CohortUncertainty `json:"cohort_uncertainty"`
 Gate CalibrationGate `json:"gate"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func ReviewCalibrationEvidence(d Dataset,asOf time.Time)CalibrationReview{
 out:=CalibrationReview{TradingAuthorized:false}
 out.WindowScores=AssessWindowCalibration(d,asOf)
 out.Reliability=AssessReliabilityBins(d,asOf)
 out.PairedBaseline=AssessPairedBrier(d,asOf)
 out.FlatBaseline=AssessFlatBrierBaseline(d,asOf)
 out.ScoreRange=AssessBidAskScoreRange(d,asOf)
 out.Uncertainty=AssessCohortUncertainty(d,asOf)
 out.Gate=CheckCalibrationGate(d,asOf)
 out.Status=out.Gate.Status
 return out
}
