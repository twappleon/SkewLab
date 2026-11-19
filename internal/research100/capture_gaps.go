package research100

import("time")

type GapWindow struct{
 Market string `json:"market"`
 Observations int `json:"observations"`
 ConsecutivePairs int `json:"consecutive_pairs"`
 GapsOverTwoWindows int `json:"gaps_over_two_windows"`
 MaximumGapSeconds *float64 `json:"maximum_gap_seconds,omitempty"`
 Note string `json:"note"`
}
// GapAudit only measures time *between observed candidates*. Selected-side
// capture is not a full market census; absent intervals do NOT prove outages.
func GapAudit(d Dataset,asOf time.Time)[]GapWindow{
 result:=[]GapWindow{
  {Market:"5m",Note:"Candidate-only interarrival gaps; unknown market coverage"},
  {Market:"15m",Note:"Candidate-only interarrival gaps; unknown market coverage"},
 }
 last:=[2]time.Time{}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  i:=0;window:=5*time.Minute
  if o.Market=="15m"{i=1;window=15*time.Minute}
  x:=&result[i];x.Observations++
  if !last[i].IsZero(){
   gap:=o.CapturedAt.Sub(last[i])
   if gap>=0{
    x.ConsecutivePairs++
    if gap>2*window{x.GapsOverTwoWindows++}
    seconds:=gap.Seconds()
    if x.MaximumGapSeconds==nil||seconds>*x.MaximumGapSeconds{v:=seconds;x.MaximumGapSeconds=&v}
   }
  }
  last[i]=o.CapturedAt
 }
 return result
}
