package research100

import (
 "encoding/json"
 "testing"
 "time"
)

func TestRound098MismatchedFingerprintFailsClosed(t *testing.T) {
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 o:=qualityFixture(now,"sample")
 b,e:=json.Marshal([]Observation{o});if e!=nil{t.Fatal(e)}
 d,e:=DecodeStrictObservations(b);if e!=nil{t.Fatal(e)}
 correct:=FingerprintDataset(d)
 good,e:=VerifyEncodedDatasetFingerprint(b,correct.DigestSHA256)
 if e!=nil||!good.Stable||good.Observations!=1{t.Fatalf("fingerprint %+v %v",good,e)}
 if _,e:=VerifyEncodedDatasetFingerprint(b,"deadbeef");e==nil{t.Fatal("modified archive accepted")}
}
