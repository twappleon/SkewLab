package research100

import "time"

type ReleaseAssurance struct {
 Status string `json:"status"`
 Observations int `json:"observations"`
 PublishedLabels int `json:"published_labels"`
 CanonicalSHA256 string `json:"canonical_sha256"`
 Blockers []string `json:"blockers"`
 StrictInputRequired bool `json:"strict_input_required"`
 CredentialsRequired bool `json:"credentials_required"`
 TradingAuthorized bool `json:"trading_authorized"`
 Note string `json:"note"`
}

// ReviewReleaseAssurance produces auditable, read-only data provenance and
// gates; it cannot verify historical authenticity or grant transaction rights.
func ReviewReleaseAssurance(d Dataset,asOf time.Time) ReleaseAssurance {
 gate:=ReviewReleaseGate(d,asOf)
 fp:=FingerprintDataset(d)
 out:=ReleaseAssurance{
  Status:gate.Status,Observations:gate.Observations,
  PublishedLabels:gate.PublishedLabels,CanonicalSHA256:fp.DigestSHA256,
  Blockers:gate.Blockers,StrictInputRequired:true,
  CredentialsRequired:false,TradingAuthorized:false,
  Note:"SHA-256 is NOT a signature; this report only checks recorded research. No paper or real trading approved.",
 }
 return out
}
