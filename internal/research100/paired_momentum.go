package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
type StrategySide struct{
 Side string `json:"side"`
 Status string `json:"status"`
}
func MomentumSide(s experiment.Sample,asOf time.Time)StrategySide{
 out:=StrategySide{Side:"WAIT",Status:"INVALID_PAIRED_CAPTURE"}
 if !ValidatePairedSnapshot(s,asOf).Valid{return out}
 if !finite(s.MomentumPct){return out}
 out.Status="NO_MOMENTUM_EDGE_SIGNAL"
 if s.MomentumPct>=.02{out.Side="UP"}else if s.MomentumPct<=-.02{out.Side="DOWN"}
 if out.Side!="WAIT"{out.Status="MOMENTUM_RULE_ONLY_NOT_TRADING_PERMISSION"}
 return out
}
