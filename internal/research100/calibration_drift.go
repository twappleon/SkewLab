package research100
import("time")
type CalibrationDrift struct{
 Earlier int `json:"earlier"`
 Later int `json:"later"`
 EarlierBrier *float64 `json:"earlier_brier,omitempty"`
 LaterBrier *float64 `json:"later_brier,omitempty"`
 Status string `json:"status"`
}
func AssessCalibrationDrift(d Dataset,cutoff,asOf time.Time)CalibrationDrift{
 out:=CalibrationDrift{Status:"INSUFFICIENT_PREDECLARED_TIME_SPLIT"}
 if cutoff.IsZero()||cutoff.After(asOf){return out}
 a,b:=0.0,0.0
 for _,o:=range d.Items{
  if o.Winner==""||o.CapturedAt.After(asOf)||o.ResolvedAt.After(asOf){continue}
  p:=(o.Bid+o.Ask)/2;y:=0.0;if o.Winner==o.Side{y=1}
  score:=(p-y)*(p-y)
  if o.CapturedAt.Before(cutoff){out.Earlier++;a+=score}else{out.Later++;b+=score}
 }
 if out.Earlier>0{x:=a/float64(out.Earlier);out.EarlierBrier=&x}
 if out.Later>0{x:=b/float64(out.Later);out.LaterBrier=&x}
 if out.Earlier>=30&&out.Later>=30{out.Status="CHRONOLOGICAL_DRIFT_RESEARCH_ONLY"}
 return out
}
