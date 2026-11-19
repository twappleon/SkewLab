package research100

import ("time")

// CaptureHorizon provides observed seconds-to-expiry by market horizon.
// This is not a measure of all tradable market slots or uptime.
type HorizonStat struct{
 Market string `json:"market"`
 Count int `json:"count"`
 MinSeconds *float64 `json:"min_seconds,omitempty"`
 MaxSeconds *float64 `json:"max_seconds,omitempty"`
}
func AssessCaptureHorizon(d Dataset,asOf time.Time) []HorizonStat{
 result:=[]HorizonStat{{Market:"5m"},{Market:"15m"}}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1}
  lead:=o.ExpiresAt.Sub(o.CapturedAt).Seconds()
  if lead<=0{continue}
  r:=&result[i];r.Count++
  if r.MinSeconds==nil||lead<*r.MinSeconds{v:=lead;r.MinSeconds=&v}
  if r.MaxSeconds==nil||lead>*r.MaxSeconds{v:=lead;r.MaxSeconds=&v}
 }
 return result
}
