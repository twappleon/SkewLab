package polymarketv2
import("fmt";"testing";"time";"math")
func fixture(now time.Time)Book{return Book{AssetID:"test",Timestamp:fmt.Sprint(now.UnixMilli()),Bids:[]Level{{Price:"0.48",Size:"100"}},Asks:[]Level{{Price:"0.52",Size:"100"}}}}
func TestDepthAndStale(t *testing.T){
 now:=time.Now().UTC();if !Evaluate(fixture(now),now,5).Executable{t.Fatal("healthy depth rejected")}
 if Evaluate(fixture(now.Add(-30*time.Second)),now,5).Executable{t.Fatal("stale book accepted")}
 b:=fixture(now);b.Asks=nil;if Evaluate(b,now,5).Executable{t.Fatal("empty book accepted")}
 b=fixture(now);b.Bids[0].Size="1";if Evaluate(b,now,5).Executable{t.Fatal("insufficient exit depth accepted")}
}
func TestNoTradeUntilValidated(t *testing.T){
 now:=time.Now().UTC();q:=Evaluate(fixture(now),now,5)
 p:=ProbabilityEvidence{Probability:.9,Lower95:.85,SampleSize:99,ResolvedBefore:now.Add(-time.Hour),IndependentSource:"historical labeled samples",OutOfSample:true}
 if Assess(q,p,now,.07,5).Decision!="NO_TRADE"{t.Fatal("insufficient sample accepted")}
 p.SampleSize=200;p.OutOfSample=false
 if Assess(q,p,now,.07,5).Decision!="NO_TRADE"{t.Fatal("in-sample estimate accepted")}
 p.OutOfSample=true
 a:=Assess(q,p,now,.07,5)
 if a.Decision!="PAPER_CANDIDATE_ONLY"||a.NetEVUSD<=0||a.Shares*q.BuyVWAP+a.ExpectedCostsUSD>5.0001{t.Fatalf("incorrect budget or gate: %+v",a)}
 if math.Abs(Fee(100,.5,.07)-1.75)>1e-9{t.Fatal("incorrect taker fee formula")}
}
