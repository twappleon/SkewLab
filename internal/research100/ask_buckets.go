package research100
import("math";"time")
type AskBucketAudit struct{
 Market string `json:"market"`
 Observations int `json:"observations"`
 BucketCounts [5]int `json:"ask_20_cent_bucket_counts"`
 Note string `json:"note"`
}
func AssessAskBuckets(d Dataset,asOf time.Time)[]AskBucketAudit{
 out:=[]AskBucketAudit{{Market:"5m",Note:"Quoted prices only; not independent calibrated probabilities"},{Market:"15m",Note:"Quoted prices only; not independent calibrated probabilities"}}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  i:=0;if o.Market=="15m"{i=1}
  bucket:=int(math.Floor(o.Ask*5))
  if bucket<0{bucket=0};if bucket>4{bucket=4}
  out[i].Observations++;out[i].BucketCounts[bucket]++
 }
 return out
}
