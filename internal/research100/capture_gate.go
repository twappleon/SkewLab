package research100
import("time")
type CaptureGate struct{
 Status string `json:"status"`
 Observations int `json:"observations"`
 PublishedLabels int `json:"published_labels"`
 DistinctDays int `json:"distinct_utc_days"`
 Blockers []string `json:"blockers"`
 TradingAuthorized bool `json:"trading_authorized"`
}
// CheckCaptureReadiness gives an observability REVIEW verdict, not evidence of
// a profitable strategy or permission to create any paper/real positions.
func CheckCaptureReadiness(d Dataset,asOf time.Time)CaptureGate{
 out:=CaptureGate{Status:"INSUFFICIENT_CAPTURE_EVIDENCE",Blockers:[]string{},TradingAuthorized:false}
 a:=AssessAsOfIntegrity(d,asOf)
 q:=AssessQuoteAge(d,asOf)
 lag:=AssessLabelLag(d,asOf)
 days:=map[string]bool{}
 out.Observations=q.Samples;out.PublishedLabels=a.PublishedLabels
 for _,o:=range d.Items{if !o.CapturedAt.After(asOf){days[o.CapturedAt.UTC().Format("2006-01-02")]=true}}
 out.DistinctDays=len(days)
 if out.Observations<100{out.Blockers=append(out.Blockers,"FEWER_THAN_100_OBSERVED_SNAPSHOTS")}
 if out.PublishedLabels<100{out.Blockers=append(out.Blockers,"FEWER_THAN_100_PUBLISHED_LABELS")}
 if out.DistinctDays<10{out.Blockers=append(out.Blockers,"FEWER_THAN_10_DISTINCT_CAPTURE_DAYS")}
 if q.P95Seconds==nil||*q.P95Seconds>5{out.Blockers=append(out.Blockers,"QUOTE_AGE_P95_EXCEEDS_FIVE_SECONDS_OR_UNKNOWN")}
 if a.CapturedInFuture>0||a.FutureOutcomeLabels>0{out.Blockers=append(out.Blockers,"ASOF_CHRONOLOGY_INCOMPLETE")}
 for _,w:=range lag{if w.Overdue30Minutes>0{out.Blockers=append(out.Blockers,"PUBLISHED_OUTCOME_BACKLOG_"+w.Market)}}
 if len(out.Blockers)==0{out.Status="CAPTURE_REVIEW_ONLY_NO_TRADING_AUTHORIZATION"}
 return out
}
