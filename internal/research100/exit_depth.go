package research100
import("time")
type ExitDepthStudy struct{
 VisibleCandidates int `json:"visible_candidates"`
 EnoughTopBidShares int `json:"enough_top_bid_shares"`
 InsufficientTopBidShares int `json:"insufficient_top_bid_shares"`
 RequestedShares float64 `json:"requested_shares"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func AssessPotentialExitDepth(d Dataset,asOf time.Time,qty float64)ExitDepthStudy{
 out:=ExitDepthStudy{RequestedShares:qty,Status:"INVALID_HYPOTHETICAL_EXIT_SIZE",TradingAuthorized:false}
 if qty<=0||!finite(qty){return out}
 out.Status="VISIBLE_BID_DEPTH_NOT_EXECUTABLE_FILL"
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  out.VisibleCandidates++
  if o.BidDepth>=qty{out.EnoughTopBidShares++}else{out.InsufficientTopBidShares++}
 }
 return out
}
