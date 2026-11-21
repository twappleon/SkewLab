package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
func MomentumSkewSide(s experiment.Sample,asOf time.Time)StrategySide{
 r:=MomentumSide(s,asOf)
 if r.Side=="WAIT"{return r}
 if !finite(s.Skew){return StrategySide{Side:"WAIT",Status:"BAD_SEMIVARIANCE_SKEW"}}
 if r.Side=="UP"&&s.Skew<.2||r.Side=="DOWN"&&s.Skew>-.2{return StrategySide{Side:"WAIT",Status:"MOMENTUM_SKEW_DISAGREE_ABSTAIN"}}
 r.Status="MOMENTUM_AND_SEMIVARIANCE_RESEARCH_ONLY"
 return r
}
