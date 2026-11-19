package research100
import("crypto/sha256";"encoding/hex";"encoding/json";"sort")
type ArchiveFingerprint struct{
 DigestSHA256 string `json:"digest_sha256"`
 Observations int `json:"observations"`
 Status string `json:"status"`
}
func FingerprintDataset(d Dataset)ArchiveFingerprint{
 clone:=append([]Observation{},d.Items...)
 sort.Slice(clone,func(i,j int)bool{return clone[i].ID<clone[j].ID})
 b,e:=json.Marshal(clone)
 if e!=nil{return ArchiveFingerprint{Observations:len(clone),Status:"SERIALIZATION_FAILURE"}}
 sum:=sha256.Sum256(b)
 return ArchiveFingerprint{DigestSHA256:hex.EncodeToString(sum[:]),Observations:len(clone),Status:"DIGEST_NOT_A_SIGNATURE_OR_PROOF_OF_AUTHENTICITY"}
}
