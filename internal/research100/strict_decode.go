package research100
import("bytes";"encoding/json";"errors";"io")
const MaxAuditArchiveBytes=32<<20
func DecodeStrictObservations(data []byte)(Dataset,error){
 if len(data)==0||len(data)>MaxAuditArchiveBytes{return Dataset{},errors.New("missing or oversized audit archive")}
 dec:=json.NewDecoder(bytes.NewReader(data))
 dec.DisallowUnknownFields()
 var records []Observation
 if e:=dec.Decode(&records);e!=nil{return Dataset{},e}
 var extra any
 if e:=dec.Decode(&extra);e!=io.EOF{
  if e==nil{return Dataset{},errors.New("trailing JSON value")}
  return Dataset{},e
 }
 return BuildDataset(records)
}
