package research100
import("time")
type StopLossWitness struct{
 Evaluated bool `json:"evaluated"`
 Triggered bool `json:"triggered"`
 PnLPct float64 `json:"pnl_pct"`
 Status string `json:"status"`
}
func EvaluateStopLossWitness(pair ReplayPair,stopPct float64)StopLossWitness{
 out:=StopLossWitness{Status:"MISSING_OR_TOO_EARLY_WITNESS"}
 if stopPct<=0||stopPct>=100||!finite(stopPct){return out}
 if pair.Exit.QuoteAt.Sub(pair.Entry.CapturedAt)<4*time.Second{return out}
 cost:=CostToResolution(pair.Entry.Ask,.002,25,0,pair.Entry.CapturedAt)
 ex:=CalculatePairedExitCost(pair,.002,25)
 if cost.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"||ex.Status!="MODELED_SINGLE_QUOTE_NOT_PROVEN_EXIT_FILL"{return out}
 out.Evaluated=true
 out.PnLPct=100*ex.ModeledOneShareNetUSD/cost.ModeledCostUSD
 out.Triggered=out.PnLPct<=-stopPct
 out.Status="HYPOTHETICAL_STOP_TRIGGER_ONLY_NOT_GUARANTEED_FILL"
 return out
}
