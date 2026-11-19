package research100

import (
 "errors"
)

type FingerprintIntegrity struct {
 Observations int `json:"observations"`
 DigestSHA256 string `json:"digest_sha256"`
 Stable bool `json:"stable"`
 Status string `json:"status"`
}

// VerifyEncodedDatasetFingerprint accepts only strict JSON and deterministic
// canonical data; a hash is not an authenticated signature or oracle proof.
func VerifyEncodedDatasetFingerprint(raw []byte,expectedDigest string) (FingerprintIntegrity,error) {
 out:=FingerprintIntegrity{Status:"UNVERIFIED_ARCHIVE"}
 dataset,e:=DecodeStrictObservations(raw)
 if e!=nil{return out,e}
 fp:=FingerprintDataset(dataset)
 out.Observations=fp.Observations
 out.DigestSHA256=fp.DigestSHA256
 out.Stable=expectedDigest!=""&&expectedDigest==fp.DigestSHA256
 out.Status="DIGEST_MISMATCH_OR_NOT_PROVIDED"
 if out.Stable{out.Status="CANONICAL_DIGEST_MATCH_NOT_SOURCE_AUTHENTICATION"}
 if expectedDigest!=""&&!out.Stable{return out,errors.New("archive fingerprint mismatch")}
 return out,nil
}
