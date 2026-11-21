package research100
import("time")
type BookGate struct{
 Observations int `json:"observations"`
 RejectedOldQuote int `json:"rejected_old_quote"`
 RejectedWideSpread int `json:"rejected_wide_spread"`
 RejectedShallowBook int `json:"rejected_shallow_book"`
 PassedVisibleFilters int `json:"passed_visible_filters"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func AssessBookGate(d Dataset,asOf time.Time)BookGate{
 out:=BookGate{Status:"VISIBLE_QUOTE_FILTER_ONLY_NEVER_TRADE_PERMISSION",TradingAuthorized:false}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  out.Observations++
  switch{
  case o.CapturedAt.Sub(o.QuoteAt)>5*time.Second:out.RejectedOldQuote++
  case o.Ask-o.Bid>.03+1e-9:out.RejectedWideSpread++
  case o.AskDepth<20||o.BidDepth<20:out.RejectedShallowBook++
  default:out.PassedVisibleFilters++
  }
 }
 return out
}
