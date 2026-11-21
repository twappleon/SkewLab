package research100
import("time")
type FeeReserveGrid struct{
 EntryAsk float64 `json:"entry_ask"`
 ExtraBps []float64 `json:"extra_bps"`
 Costs []EntryCost `json:"costs"`
 Status string `json:"status"`
}
func AssessFeeReserveGrid(ask float64,at time.Time)FeeReserveGrid{
 out:=FeeReserveGrid{EntryAsk:ask,ExtraBps:[]float64{0,25,100},Costs:[]EntryCost{},Status:"ADDITIONAL_BUFFER_IS_NOT_PLATFORM_FEE"}
 for _,bps:=range out.ExtraBps{out.Costs=append(out.Costs,CostToResolution(ask,.002,bps,2,at))}
 return out
}
