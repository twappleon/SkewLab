package research100
import("math";"testing";"time")
func TestRound082PairedMidpointSumsToOne(t *testing.T){
 now:=time.Date(2026,11,20,20,0,0,0,time.UTC)
 s:=pairedFixture(now,"m");s.UpBid=.69;s.UpAsk=.71;s.DownBid=.29;s.DownAsk=.31
 x:=AssessMarketProxy(s,now)
 if !x.Valid||math.Abs(x.UpFraction-.7)>1e-12||math.Abs(x.UpFraction+x.DownFraction-1)>1e-12{t.Fatalf("bad proxy %+v",x)}
}
