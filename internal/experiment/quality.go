package experiment

import (
    "math"
    "time"
)

// WindowQuality measures theoretical observation slots only after the first
// valid fixed-time capture. Unknown periods before that are NOT counted.
type WindowQuality struct {
    Window string `json:"window"`
    ExpectedSlots *int `json:"expected_slots,omitempty"`
    Captured int `json:"captured"`
    MissingSlots *int `json:"missing_slots,omitempty"`
    CaptureRatePct *float64 `json:"capture_rate_pct,omitempty"`
    Resolved int `json:"resolved"`
    Pending int `json:"pending"`
    Matured int `json:"matured"`
    ResolvedMatured int `json:"resolved_matured"`
    SettlementRatePct *float64 `json:"settlement_rate_pct,omitempty"`
    Overdue int `json:"overdue"`
    Invalid int `json:"invalid"`
    FirstCaptureAt *time.Time `json:"first_capture_at,omitempty"`
    LastCaptureAt *time.Time `json:"last_capture_at,omitempty"`
    ForwardResolved int `json:"forward_resolved"`
    MissingReason string `json:"missing_reason"`
}

type CoverageQuality struct {
    Status string `json:"status"`
    AsOf time.Time `json:"as_of"`
    Windows []WindowQuality `json:"windows"`
    ArchiveError bool `json:"archive_error"`
    TradingAuthorized bool `json:"trading_authorized"`
    Interpretation string `json:"interpretation"`
}
func coveragePercent(n,d int)*float64 {
    if d<=0{return nil}
    x:=math.Round(float64(n)*10000/float64(d))/100
    return &x
}

// AssessCoverage never backfills unknown samples or predicts official results.
// The theoretical denominator includes service downtime and is NOT feed uptime.
func AssessCoverage(samples []Sample,now time.Time,archiveError bool)CoverageQuality{
    out:=CoverageQuality{AsOf:now,Status:"INSUFFICIENT_FORWARD_EVIDENCE",
        ArchiveError:archiveError,TradingAuthorized:false,
        Windows:[]WindowQuality{{Window:"5m"},{Window:"15m"}},
        Interpretation:"Research only. Expected slots begin at the first recorded sample per market window in the last 14 days. Missing slots may include Railway downtime or missing quotes; this is not verified provider uptime. Gamma is not independent Chainlink verification. No paper or live trading authorization."}
    if archiveError{out.Status="RESEARCH_ARCHIVE_ERROR";return out}
    if now.IsZero(){out.Status="INVALID_AS_OF_TIME";return out}
    cutoff:=now.Add(-14*24*time.Hour)
    first:=[2]time.Time{}
    for _,s:=range samples{
        i:=-1
        if s.WindowSeconds==300{i=0}else if s.WindowSeconds==900{i=1}
        if i<0||s.CapturedAt.Before(cutoff)||s.CapturedAt.After(now){continue}
        w:=&out.Windows[i]
        if s.Policy!=SamplingPolicy||!validSnapshot(s)||(s.Winner!=""&&s.ResolvedAt.After(now)){w.Invalid++;continue}
        w.Captured++
        if w.FirstCaptureAt==nil||s.CapturedAt.Before(*w.FirstCaptureAt){x:=s.CapturedAt;w.FirstCaptureAt=&x}
        if w.LastCaptureAt==nil||s.CapturedAt.After(*w.LastCaptureAt){x:=s.CapturedAt;w.LastCaptureAt=&x}
        slot:=s.MarketEnd.Add(-time.Duration(s.WindowSeconds)*time.Second)
        if first[i].IsZero()||slot.Before(first[i]){first[i]=slot}
        if s.Winner==""{w.Pending++}else{w.Resolved++}
        if now.After(s.MarketEnd.Add(30*time.Minute)){
            w.Matured++
            if s.Winner==""{w.Overdue++}else{w.ResolvedMatured++}
        }
    }
    any,backlog,gaps,bad:=false,false,false,false
    for i:=range out.Windows{
        w:=&out.Windows[i]
        w.SettlementRatePct=coveragePercent(w.ResolvedMatured,w.Matured)
        if w.Invalid>0{bad=true}
        if w.Overdue>0{backlog=true}
        if w.Captured==0{
            w.MissingReason="No first fixed-time capture; coverage denominator unknown."
            continue
        }
        any=true
        duration:=time.Duration(300)*time.Second
        if i==1{duration=900*time.Second}
        lastEligible:=now.Add(-120*time.Second)
        expected:=1
        if !lastEligible.Before(first[i]){expected=int(lastEligible.Sub(first[i])/duration)+1}
        if expected<w.Captured{expected=w.Captured}
        missing:=expected-w.Captured
        w.ExpectedSlots=&expected
        w.MissingSlots=&missing
        w.CaptureRatePct=coveragePercent(w.Captured,expected)
        if missing>0{gaps=true}
        switch{
        case w.Matured==0:w.MissingReason="No matured samples for settlement-rate calculation."
        case w.Overdue>0:w.MissingReason="Some observed markets lack a Gamma label more than 30 minutes after expiration."
        case missing>0:w.MissingReason="Theoretical market slots missing; root cause cannot be established from archive alone."
        default:w.MissingReason="Collecting prospective market results; no trading authorization."
        }
    }
    switch{
    case bad:out.Status="INVALID_COMPARISON_RECORDS"
    case !any:out.Status="NO_FIXED_TIME_SAMPLES"
    case backlog:out.Status="SETTLEMENT_BACKLOG_REVIEW"
    case gaps:out.Status="CAPTURE_GAPS_REVIEW"
    default:out.Status="INSUFFICIENT_FORWARD_EVIDENCE"
    }
    return out
}
func(r *Recorder) Quality(now time.Time)CoverageQuality{
    if r==nil{return AssessCoverage(nil,now,true)}
    items:=r.Snapshot()
    q:=AssessCoverage(items,now,r.Error())
    c:=Compare(items,now)
    for i:=range q.Windows{q.Windows[i].ForwardResolved=c.AllWindows[i].ForwardResolved}
    return q
}
