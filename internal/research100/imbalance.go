package research100
import("time")
type ImbalanceAudit struct{
 Observations int `json:"observations"`
 BidDominant int `json:"bid_dominant"`
 AskDominant int `json:"ask_dominant"`
 Neutral int `json:"neutral"`
 Status string `json:"status"`
}
func AssessTopBookImbalance(d Dataset,asOf time.Time)ImbalanceAudit{
 out:=ImbalanceAudit{Status:"TOP_BOOK_IMBALANCE_NOT_DIRECTION_PROOF"}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){continue}
  total:=o.BidDepth+o.AskDepth
  if total<=0||!finite(total){continue}
  out.Observations++
  ratio:=o.BidDepth/total
  if ratio>=.8{out.BidDominant++}else if ratio<=.2{out.AskDominant++}else{out.Neutral++}
 }
 return out
}
