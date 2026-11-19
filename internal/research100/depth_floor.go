package research100
import("time")
type DepthFloorAudit struct{
 Market string `json:"market"`
 Observations int `json:"observations"`
 BelowBidFloor int `json:"below_bid_floor"`
 BelowAskFloor int `json:"below_ask_floor"`
 MeetsBoth int `json:"meets_both"`
 Status string `json:"status"`
}
func AssessDepthFloors(d Dataset,asOf time.Time,bidMin,askMin float64)[]DepthFloorAudit{
 out:=[]DepthFloorAudit{{Market:"5m",Status:"VISIBLE_TOP_BOOK_ONLY"},{Market:"15m",Status:"VISIBLE_TOP_BOOK_ONLY"}}
 if !finite(bidMin)||!finite(askMin)||bidMin<=0||askMin<=0{
  for i:=range out{out[i].Status="INVALID_DEPTH_THRESHOLD"};return out
 }
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1};x:=&out[i];x.Observations++
  if o.BidDepth<bidMin{x.BelowBidFloor++}
  if o.AskDepth<askMin{x.BelowAskFloor++}
  if o.BidDepth>=bidMin&&o.AskDepth>=askMin{x.MeetsBoth++}
 }
 return out
}
