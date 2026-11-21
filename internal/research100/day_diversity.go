package research100

import("time")
type DayDiversity struct{
 Market string `json:"market"`
 Observations int `json:"observations"`
 UTCDateCount int `json:"utc_date_count"`
 MaximumSingleDayCount int `json:"maximum_single_day_count"`
 Concentrated bool `json:"concentrated"`
 Status string `json:"status"`
}
// DayConcentration is only a sample-dependence warning. Ten distinct dates
// do not prove outcomes independent or profitability.
func DayConcentration(d Dataset,asOf time.Time)[]DayDiversity{
 out:=[]DayDiversity{{Market:"5m",Status:"INSUFFICIENT_DAY_DIVERSITY"},{Market:"15m",Status:"INSUFFICIENT_DAY_DIVERSITY"}}
 counts:=[2]map[string]int{make(map[string]int),make(map[string]int)}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1}
  counts[i][o.CapturedAt.UTC().Format("2006-01-02")]++
  out[i].Observations++
 }
 for i:=range out{
  out[i].UTCDateCount=len(counts[i])
  for _,n:=range counts[i]{if n>out[i].MaximumSingleDayCount{out[i].MaximumSingleDayCount=n}}
  if out[i].Observations>0{out[i].Concentrated=out[i].MaximumSingleDayCount*2>out[i].Observations}
  if out[i].UTCDateCount>=10&&!out[i].Concentrated{out[i].Status="DIVERSE_DATES_NOT_INDEPENDENCE_PROOF"}
 }
 return out
}
