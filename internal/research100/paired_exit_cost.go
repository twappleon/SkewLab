package research100
import("github.com/twappleon/SkewLab/internal/fees";"math")
type PairedExitCost struct{
 ModeledOneShareNetUSD float64 `json:"modeled_one_share_net_usd"`
 VisibleBidDepth float64 `json:"visible_bid_depth"`
 Status string `json:"status"`
}
func CalculatePairedExitCost(pair ReplayPair,slippage,extraBps float64)PairedExitCost{
 out:=PairedExitCost{Status:"INVALID_REPLAY_QUOTE_PAIR"}
 if slippage<0||extraBps<0||!finite(slippage)||!finite(extraBps){return out}
 if !ValidateExitWitness(pair.Entry,pair.Exit,pair.Exit.QuoteAt).Valid{return out}
 costPx:=pair.Entry.Ask+slippage;exitPx:=math.Max(.001,pair.Exit.Bid-slippage)
 if costPx<=0||costPx>=1||exitPx<=0||exitPx>=1{return out}
 entryCost:=fees.BufferedCost(1,costPx,extraBps,pair.Entry.CapturedAt)
 exitProceeds:=fees.BufferedProceeds(1,exitPx,extraBps,pair.Exit.QuoteAt)
 if !finite(entryCost)||!finite(exitProceeds)||entryCost<=0||exitProceeds<=0{return out}
 out.ModeledOneShareNetUSD=exitProceeds-entryCost;out.VisibleBidDepth=pair.Exit.BidDepth
 out.Status="MODELED_SINGLE_QUOTE_NOT_PROVEN_EXIT_FILL"
 return out
}
