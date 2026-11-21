package research100
import("testing")
func TestRound091ArchiveCapacity(t *testing.T){
 d:=Dataset{Items:make([]Observation,2)}
 if x:=AssessArchiveCapacity(d,1);x.WithinLimit||x.Status!="ARCHIVE_ITEM_CAP_EXCEEDED"{t.Fatalf("unbounded report %+v",x)}
 if x:=AssessArchiveCapacity(d,50000);!x.WithinLimit{t.Fatalf("incorrect cap %+v",x)}
}
