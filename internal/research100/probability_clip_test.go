package research100
import("math";"testing")
func TestRound043ClippedExtremesAndRejectNaN(t *testing.T){
 x:=ClipProbability(0,.000001)
 if !x.Accepted||x.Clipped!=.000001{t.Fatalf("zero clip %+v",x)}
 y:=ClipProbability(1,.000001)
 if !y.Accepted||y.Clipped!=.999999{t.Fatalf("one clip %+v",y)}
 if z:=ClipProbability(math.NaN(),.01);z.Accepted{t.Fatal("NaN accepted")}
 if z:=ClipProbability(.5,0);z.Accepted{t.Fatal("unbounded log loss")}
}
