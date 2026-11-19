package research100
import("testing";"time")
func TestRound024RawTimingDetectsAllMalformedCases(t *testing.T){
 now:=time.Date(2026,11,20,10,0,0,0,time.UTC)
 a:=qualityFixture(now,"early");a.Winner="UP";a.ResolvedAt=a.CapturedAt
 b:=qualityFixture(now,"noat");b.Winner="DOWN"
 c:=qualityFixture(now,"orphan");c.ResolvedAt=now.Add(time.Hour)
 d:=qualityFixture(now,"expiry");d.ExpiresAt=d.CapturedAt
 x:=AuditRawSettlementTiming([]Observation{a,b,c,d})
 if x.RawRecords!=4||x.EarlyPublishedBeforeExpiry!=1||x.LabelWithoutTimestamp!=1||x.TimestampWithoutLabel!=1||x.InvalidCaptureExpiry!=1||x.Status!="REJECT_UNTRUSTWORTHY_RAW_SETTLEMENT_TIMING"{t.Fatalf("missed malformed input %+v",x)}
}
