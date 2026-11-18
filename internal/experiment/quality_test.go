package experiment

import (
    "math"
    "testing"
    "time"
)
func qualityFixture(tm time.Time,duration int,name string)Sample{
    length:=time.Duration(duration)*time.Second
    start:=tm.Truncate(length)
    at:=start.Add(100*time.Second)
    return Sample{Policy:SamplingPolicy,Slug:name,ConditionID:"cid-"+name,
        WindowSeconds:duration,CapturedAt:at,MarketEnd:start.Add(length),
        UpQuoteAt:at,DownQuoteAt:at,UpBid:.49,UpAsk:.51,UpAskDepth:90,
        DownBid:.48,DownAsk:.50,DownAskDepth:90,MomentumPct:.03,Skew:.2,
        BTCSpotSource:"test"}
}
func TestCoverageUnseenAndCorrupt(t *testing.T){
    now:=time.Date(2026,11,18,10,0,0,0,time.UTC)
    q:=AssessCoverage(nil,now,false)
    if q.Status!="NO_FIXED_TIME_SAMPLES"||q.TradingAuthorized||q.Windows[0].ExpectedSlots!=nil||q.Windows[1].CaptureRatePct!=nil{t.Fatalf("unseen fabricated: %+v",q)}
    q=AssessCoverage(nil,now,true)
    if !q.ArchiveError||q.Status!="RESEARCH_ARCHIVE_ERROR"||q.TradingAuthorized{t.Fatalf("corrupt archive: %+v",q)}
}
func TestCoverageMissingSlotsNeverInventOtherWindow(t *testing.T){
    start:=time.Date(2026,11,18,8,0,0,0,time.UTC)
    a:=qualityFixture(start,300,"first")
    b:=qualityFixture(start.Add(10*time.Minute),300,"third")
    q:=AssessCoverage([]Sample{a,b},start.Add(16*time.Minute),false)
    w:=q.Windows[0]
    if w.ExpectedSlots==nil||*w.ExpectedSlots!=3||w.MissingSlots==nil||*w.MissingSlots!=1||w.Captured!=2||w.CaptureRatePct==nil||math.Abs(*w.CaptureRatePct-66.67)>.01{t.Fatalf("wrong coverage: %+v",w)}
    if q.Windows[1].ExpectedSlots!=nil||q.Windows[1].CaptureRatePct!=nil||q.Status!="CAPTURE_GAPS_REVIEW"||q.TradingAuthorized{t.Fatalf("improper inference: %+v",q)}
}
func TestCoverageSettlementMaturityAndFutureRejection(t *testing.T){
    start:=time.Date(2026,11,18,8,0,0,0,time.UTC)
    a:=qualityFixture(start,300,"first")
    b:=qualityFixture(start.Add(5*time.Minute),300,"second")
    a.Winner="UP";a.ResolvedAt=a.MarketEnd.Add(time.Minute)
    now:=start.Add(45*time.Minute)
    q:=AssessCoverage([]Sample{a,b},now,false)
    w:=q.Windows[0]
    if w.Matured!=2||w.ResolvedMatured!=1||w.Overdue!=1||w.SettlementRatePct==nil||*w.SettlementRatePct!=50||q.Status!="SETTLEMENT_BACKLOG_REVIEW"{t.Fatalf("maturity: %+v",q)}
    b.Winner="DOWN";b.ResolvedAt=now.Add(time.Hour)
    q=AssessCoverage([]Sample{a,b},now,false)
    if q.Windows[0].Invalid!=1||q.Status!="INVALID_COMPARISON_RECORDS"{t.Fatalf("future label allowed: %+v",q)}
}
func TestCoverageNoRateForTooNewOrOld(t *testing.T){
    now:=time.Date(2026,11,18,10,0,0,0,time.UTC)
    old:=qualityFixture(now.Add(-15*24*time.Hour),300,"old")
    q:=AssessCoverage([]Sample{old},now,false)
    if q.Windows[0].Captured!=0||q.Windows[0].ExpectedSlots!=nil{t.Fatal("old sample included")}
    recent:=qualityFixture(now.Add(-8*time.Minute),300,"recent")
    recent.Winner="UP";recent.ResolvedAt=now.Add(-time.Minute)
    q=AssessCoverage([]Sample{recent},now,false)
    if q.Windows[0].Captured!=1||q.Windows[0].Matured!=0||q.Windows[0].SettlementRatePct!=nil||q.TradingAuthorized{t.Fatalf("unmatured counted: %+v",q)}
}
