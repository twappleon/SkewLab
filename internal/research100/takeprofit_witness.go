package research100
import("time")
type TakeProfitWitness struct{
 Evaluated bool `json:"evaluated"`
 Triggered bool `json:"triggered"`
 PnLPct float64 `json:"pnl_pct"`
 Status string `json:"status"`
}
func EvaluateTakeProfitWitness(pair ReplayPair,targetPct float64)TakeProfitWitness{
 out:=TakeProfitWitness{Status:"INSUFFICIENT_SINGLE_EXIT_EVIDENCE"}
 if !finite(targetPct)||targetPct<=0{return out}
 if pair.Exit.QuoteAt.Sub(pair.Entry.CapturedAt)<4*time.Second{return out}
 c:=CostToResolution(pair.Entry.Ask,.002,25,0,pair.Entry.CapturedAt)
 ex:=CalculatePairedExitCost(pair,.002,25)
 if c.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"||ex.Status!="MODELED_SINGLE_QUOTE_NOT_PROVEN_EXIT_FILL"{return out}
 out.Evaluated=true;out.PnLPct=100*ex.ModeledOneShareNetUSD/c.ModeledCostUSD
 out.Triggered=out.PnLPct>=targetPct
 out.Status="HYPOTHETICAL_AT_WITNESS_TIME_NOT_FIRST_TRIGGER"
 return out
}
