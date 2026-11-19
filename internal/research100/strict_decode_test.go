package research100
import("encoding/json";"testing";"time")
func TestRound092RejectUnknownArchiveFields(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now,"a")
 b,e:=json.Marshal([]Observation{a});if e!=nil{t.Fatal(e)}
 if _,e=DecodeStrictObservations(b);e!=nil{t.Fatal(e)}
 if _,e=DecodeStrictObservations(append(b,[]byte(" {}")...));e==nil{t.Fatal("trailing JSON accepted")}
 if _,e=DecodeStrictObservations([]byte(`[{"id":"m","secret_field":"unexpected"}]`));e==nil{t.Fatal("unknown schema accepted")}
}
