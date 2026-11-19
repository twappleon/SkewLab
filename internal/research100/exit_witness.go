package research100
import("time")
type ExitWitness struct{
 ID string `json:"id"`
 QuoteAt time.Time `json:"quote_at"`
 Bid float64 `json:"bid"`
 BidDepth float64 `json:"bid_depth"`
 Source string `json:"source"`
}
type ExitWitnessValidation struct{
 Valid bool `json:"valid"`
 Reason string `json:"reason"`
}
func ValidateExitWitness(entry Observation,w ExitWitness,asOf time.Time)ExitWitnessValidation{
 if w.ID==""||entry.ID!=w.ID{return ExitWitnessValidation{Reason:"WRONG_OR_MISSING_MARKET_ID"}}
 if w.QuoteAt.IsZero()||!w.QuoteAt.After(entry.CapturedAt)||!w.QuoteAt.Before(entry.ExpiresAt)||w.QuoteAt.After(asOf){return ExitWitnessValidation{Reason:"EXIT_QUOTE_OUTSIDE_POST_ENTRY_WINDOW"}}
 if w.Source==""||!finite(w.Bid)||!finite(w.BidDepth)||w.Bid<=0||w.Bid>=1||w.BidDepth<=0{return ExitWitnessValidation{Reason:"INVALID_EXIT_BID_OR_DEPTH"}}
 return ExitWitnessValidation{Valid:true,Reason:"VISIBLE_BID_ONLY_NOT_EXECUTABLE_FILL"}
}
