package research100
import("time")
type SlippageSensitivity struct{
 EntryAsk float64 `json:"entry_ask"`
 Costs []EntryCost `json:"costs"`
 NonDecreasing bool `json:"non_decreasing"`
 Status string `json:"status"`
}
func AssessSlippageSensitivity(ask float64,at time.Time)SlippageSensitivity{
 out:=SlippageSensitivity{EntryAsk:ask,Costs:[]EntryCost{},NonDecreasing:true,Status:"HYPOTHETICAL_SLIPPAGE_GRID"}
 last:=-1.0
 for _,s:=range []float64{.001,.002,.005}{
  c:=CostToResolution(ask,s,25,2,at)
  out.Costs=append(out.Costs,c)
  if c.Status!="COST_MODEL_ONLY_NOT_EXECUTABLE"{out.Status="INFEASIBLE_SLIPPAGE_SCENARIO";out.NonDecreasing=false;continue}
  if last>=0&&c.ModeledCostUSD<last{out.NonDecreasing=false}
  last=c.ModeledCostUSD
 }
 return out
}
