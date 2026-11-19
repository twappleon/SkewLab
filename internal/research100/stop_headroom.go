package research100
import("github.com/twappleon/SkewLab/internal/fees";"math";"time")
type StopHeadroomReview struct{
 Observations int `json:"observations"`
 AlreadyExceedsStop int `json:"already_exceeds_stop"`
 MinHeadroomPct *float64 `json:"minimum_headroom_pct,omitempty"`
 Status string `json:"status"`
}
func AssessStopHeadroom(d Dataset,asOf time.Time,stopPct float64)StopHeadroomReview{
 out:=StopHeadroomReview{Status:"INVALID_STOP_LIMIT"}
 if !finite(stopPct)||stopPct<=0||stopPct>=100{return out}
 out.Status="MODELED_STOP_IS_TRIGGER_NOT_GUARANTEED_FILL"
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  px:=o.Ask+.002;exit:=math.Max(.001,o.Bid-.002)
  if px>=1||exit<=0{continue}
  buy:=fees.BufferedCost(1,px,25,o.CapturedAt)
  sell:=fees.BufferedProceeds(1,exit,25,o.CapturedAt)
  if buy<=0||sell<=0||!finite(buy)||!finite(sell){continue}
  loss:=100*(buy-sell)/buy;room:=stopPct-loss
  out.Observations++
  if room<=0{out.AlreadyExceedsStop++}
  if out.MinHeadroomPct==nil||room<*out.MinHeadroomPct{v:=room;out.MinHeadroomPct=&v}
 }
 return out
}
