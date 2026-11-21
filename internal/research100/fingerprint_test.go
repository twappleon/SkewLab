package research100
import("testing";"time")
func TestRound094OrderIndependentArchiveFingerprint(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 a:=qualityFixture(now,"a");b:=qualityFixture(now.Add(time.Second),"b")
 d,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 f,e:=BuildDataset([]Observation{b,a});if e!=nil{t.Fatal(e)}
 x:=FingerprintDataset(d);y:=FingerprintDataset(f)
 if x.DigestSHA256!=y.DigestSHA256||len(x.DigestSHA256)!=64{t.Fatalf("unstable digest %+v %+v",x,y)}
 b.Momentum=.4;z,e:=BuildDataset([]Observation{a,b});if e!=nil{t.Fatal(e)}
 if FingerprintDataset(z).DigestSHA256==x.DigestSHA256{t.Fatal("mutation undetected")}
}
