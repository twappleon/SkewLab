package research100
import("time")
type PriorLabels struct{
 Market string `json:"market"`
 Side string `json:"side"`
 AsOf time.Time `json:"as_of"`
 PublishedBefore int `json:"published_before"`
 CorrectBefore int `json:"correct_before"`
 DistinctUTCDays int `json:"distinct_utc_days"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func AssessPriorLabels(d Dataset,market,side string,at time.Time)PriorLabels{
 out:=PriorLabels{Market:market,Side:side,AsOf:at,TradingAuthorized:false}
 days:=map[string]bool{}
 if at.IsZero()||(market!="5m"&&market!="15m")||(side!="UP"&&side!="DOWN"){return out}
 for _,o:=range d.Items{
  if o.Market!=market||o.Side!=side||o.Winner==""||!o.ResolvedAt.Before(at)||!o.CapturedAt.Before(at){continue}
  out.PublishedBefore++
  if o.Winner==side{out.CorrectBefore++}
  days[o.CapturedAt.UTC().Format("2006-01-02")]=true
 }
 out.DistinctUTCDays=len(days)
 return out
}
