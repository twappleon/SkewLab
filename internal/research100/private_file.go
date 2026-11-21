package research100
import("errors";"io";"os")
func ReadPrivateAuditFile(path string,maxBytes int64)([]byte,error){
 if path==""||maxBytes<=0||maxBytes>MaxAuditArchiveBytes{return nil,errors.New("invalid private archive access")}
 info,e:=os.Lstat(path);if e!=nil{return nil,e}
 if !info.Mode().IsRegular(){return nil,errors.New("private archive must be a regular file, not a link")}
 if info.Mode().Perm()&0077!=0{return nil,errors.New("private archive permissions expose group/other access")}
 if info.Size()<=0||info.Size()>maxBytes{return nil,errors.New("private archive absent or too large")}
 f,e:=os.Open(path);if e!=nil{return nil,e};defer f.Close()
 data,e:=io.ReadAll(io.LimitReader(f,maxBytes+1));if e!=nil{return nil,e}
 if len(data)==0||int64(len(data))>maxBytes{return nil,errors.New("private archive exceeds byte cap")}
 return data,nil
}
