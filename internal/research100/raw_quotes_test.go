package research100
import("math";"testing";"time")
func TestRound035RawQuoteAnomaliesBeforeAdmission(t *testing.T){
 now:=time.Date(2026,11,19,12,0,0,0,time.UTC)
 crossed:=qualityFixture(now,"cross");crossed.Bid=.60;crossed.Ask=.50
 missing:=qualityFixture(now,"missing");missing.BidDepth=0
 broken:=qualityFixture(now,"nan");broken.Ask=math.NaN()
 clean:=qualityFixture(now,"good")
 got:=AuditRawQuotes([]Observation{crossed,missing,broken,clean})
 if got.CrossedOrLocked!=1||got.MissingTopOfBook!=1||got.InvalidNumerics!=1||got.Records!=4||got.Status!="RAW_BOOK_ANOMALIES_REQUIRE_REJECTION"{t.Fatalf("raw book audit %+v",got)}
}
