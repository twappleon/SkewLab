package research100
import("time";"github.com/twappleon/SkewLab/internal/experiment")
type PairedSnapshotCheck struct{
 Valid bool `json:"valid"`
 Reason string `json:"reason"`
}
func ValidatePairedSnapshot(s experiment.Sample,asOf time.Time)PairedSnapshotCheck{
 bad:=func(s string)PairedSnapshotCheck{return PairedSnapshotCheck{Reason:s}}
 if s.Slug==""||s.ConditionID==""||s.Policy!=experiment.SamplingPolicy||(s.WindowSeconds!=300&&s.WindowSeconds!=900){return bad("BAD_MARKET_ID_OR_SAMPLING_POLICY")}
 if s.CapturedAt.IsZero()||s.MarketEnd.IsZero()||s.CapturedAt.After(asOf){return bad("MISSING_OR_FUTURE_CAPTURE")}
 elapsed:=s.CapturedAt.Sub(s.MarketEnd.Add(-time.Duration(s.WindowSeconds)*time.Second))
 if elapsed<90*time.Second||elapsed>120*time.Second||!s.CapturedAt.Before(s.MarketEnd){return bad("OUTSIDE_FIXED_CAPTURE_WINDOW")}
 for _,qt:=range []time.Time{s.UpQuoteAt,s.DownQuoteAt}{
  if qt.IsZero()||qt.After(s.CapturedAt)||s.CapturedAt.Sub(qt)>10*time.Second{return bad("INVALID_BID_ASK_QUOTE_TIMESTAMP")}
 }
 if s.UpQuoteAt.Sub(s.DownQuoteAt)>5*time.Second||s.DownQuoteAt.Sub(s.UpQuoteAt)>5*time.Second{return bad("UNSYNCED_UP_DOWN_BOOKS")}
 if !finite(s.UpBid)||!finite(s.UpAsk)||!finite(s.DownBid)||!finite(s.DownAsk)||!finite(s.UpAskDepth)||!finite(s.DownAskDepth){return bad("INVALID_BOOK_NUMBER")}
 if s.UpBid<=0||s.DownBid<=0||s.UpAsk>=1||s.DownAsk>=1||s.UpBid>s.UpAsk||s.DownBid>s.DownAsk||s.UpAskDepth<=0||s.DownAskDepth<=0{return bad("MISSING_OR_CROSSED_BOOK")}
 if s.Winner!=""&&(s.Winner!="UP"&&s.Winner!="DOWN"||s.ResolvedAt.Before(s.MarketEnd)){return bad("INVALID_OUTCOME_CHRONOLOGY")}
 if s.Winner==""&&!s.ResolvedAt.IsZero(){return bad("UNLABELLED_HAS_RESOLUTION")}
 return PairedSnapshotCheck{Valid:true,Reason:"PRESET_PAIRED_OBSERVATION_ONLY"}
}
