package research100
import("os";"path/filepath";"testing")
func TestRound093PrivateArchiveRejectsWorldReadableAndLinks(t *testing.T){
 dir:=t.TempDir();p:=filepath.Join(dir,"archive.json")
 if e:=os.WriteFile(p,[]byte("[]"),0600);e!=nil{t.Fatal(e)}
 if _,e:=ReadPrivateAuditFile(p,1024);e!=nil{t.Fatal(e)}
 if e:=os.Chmod(p,0644);e!=nil{t.Fatal(e)}
 if _,e:=ReadPrivateAuditFile(p,1024);e==nil{t.Fatal("world-readable archive accepted")}
 if e:=os.Chmod(p,0600);e!=nil{t.Fatal(e)}
 link:=filepath.Join(dir,"link.json");if e:=os.Symlink(p,link);e!=nil{t.Fatal(e)}
 if _,e:=ReadPrivateAuditFile(link,1024);e==nil{t.Fatal("symlink accepted")}
}
