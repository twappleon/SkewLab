// polymarket-scan performs read-only public research; it cannot trade.
package main
import (
 "context"
 "encoding/json"
 "flag"
 "fmt"
 "os"
 "time"
 "github.com/twappleon/SkewLab/internal/polymarketscan"
)
func main(){
 limit:=flag.Int("limit",50,"public markets to inspect (1..100)");flag.Parse()
 p:=polymarketscan.DefaultPolicy()
 ctx,cancel:=context.WithTimeout(context.Background(),15*time.Second);defer cancel()
 markets,err:=polymarketscan.Fetch(ctx,nil,*limit)
 if err!=nil{fmt.Fprintln(os.Stderr,"scan unavailable (no fabricated data):",err);os.Exit(1)}
 result,err:=polymarketscan.Scan(time.Now(),markets,p)
 if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
 encoder:=json.NewEncoder(os.Stdout);encoder.SetIndent("","  ")
 if err:=encoder.Encode(result);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
}
