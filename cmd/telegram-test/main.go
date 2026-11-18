// telegram-test only sends when the operator explicitly supplies --send.
package main
import(
 "context"
 "flag"
 "fmt"
 "os"
 "time"
 "github.com/twappleon/SkewLab/internal/notify"
)
func main(){
 send:=flag.Bool("send",false,"explicitly send one TEST notification to the configured Telegram chat")
 flag.Parse()
 if !*send{fmt.Fprintln(os.Stderr,"No message sent. Use --send to perform one explicit Telegram test.");return}
 c:=notify.FromEnv(".")
 if c==nil{fmt.Fprintln(os.Stderr,"Telegram token and numeric chat ID are required");os.Exit(1)}
 ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel()
 if err:=c.SendTest(ctx);err!=nil{fmt.Fprintf(os.Stderr,"Telegram test failed: %s\n",err);os.Exit(1)}
 fmt.Println("Telegram test accepted by API; check the intended recipient chat.")
}
