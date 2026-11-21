package research100
import("testing";"time")
func TestRound079MissingExitQuotesMustStayMissing(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now.Add(-time.Minute),"a")
 d,e:=BuildDataset([]Observation{a});if e!=nil{t.Fatal(e)}
 x:=AssessWitnessCoverage(d,nil,now)
 if x.Entries!=1||x.Missing!=1||x.Matched!=0||x.Fraction==nil||*x.Fraction!=0{t.Fatalf("invented matched evidence %+v",x)}
 e0:=AssessWitnessCoverage(Dataset{},nil,now)
 if e0.Fraction!=nil||e0.Status!="UNKNOWN_NO_ENTRIES"{t.Fatalf("zero denominator %+v",e0)}
}
