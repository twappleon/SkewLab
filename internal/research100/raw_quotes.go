package research100
type RawQuoteAnomalies struct{
 Records int `json:"records"`
 CrossedOrLocked int `json:"crossed_or_locked"`
 MissingTopOfBook int `json:"missing_top_of_book"`
 InvalidNumerics int `json:"invalid_numerics"`
 Status string `json:"status"`
}
func AuditRawQuotes(raw []Observation)RawQuoteAnomalies{
 out:=RawQuoteAnomalies{Records:len(raw),Status:"RAW_QUOTES_REVIEWED_ONLY"}
 for _,o:=range raw{
  if !finite(o.Bid)||!finite(o.Ask)||!finite(o.BidDepth)||!finite(o.AskDepth){out.InvalidNumerics++;continue}
  if o.Bid<=0||o.Ask>=1||o.BidDepth<=0||o.AskDepth<=0{out.MissingTopOfBook++;continue}
  if o.Bid>=o.Ask{out.CrossedOrLocked++}
 }
 if out.InvalidNumerics+out.MissingTopOfBook+out.CrossedOrLocked>0{out.Status="RAW_BOOK_ANOMALIES_REQUIRE_REJECTION"}
 return out
}
