package research100
import("time")
type CalibrationGate struct{
 Published int `json:"published"`
 UTCObservationDays int `json:"utc_observation_days"`
 QuoteProxyBrier *float64 `json:"quote_proxy_brier,omitempty"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func CheckCalibrationGate(d Dataset,asOf time.Time)CalibrationGate{
 out:=CalibrationGate{Status:"INSUFFICIENT_CALIBRATION_EVIDENCE",TradingAuthorized:false}
 days:=map[string]bool{}
 for _,o:=range d.Items{
  if o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.After(asOf){continue}
  days[o.CapturedAt.UTC().Format("2006-01-02")]=true
  out.Published++
 }
 out.UTCObservationDays=len(days)
 score:=ScoreCalibration(d,asOf);out.QuoteProxyBrier=score.Brier
 if out.Published<100||out.UTCObservationDays<10||out.QuoteProxyBrier==nil{return out}
 if *out.QuoteProxyBrier>=.25{out.Status="QUOTE_PROXY_NO_GAIN_OVER_COIN_FLIP"}
 if *out.QuoteProxyBrier<.25{out.Status="RESEARCH_ONLY_POTENTIAL_IN_SAMPLE_GAIN"}
 return out
}
