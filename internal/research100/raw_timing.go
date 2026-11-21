package research100
import("time")
type RawSettlementTiming struct{
 RawRecords int `json:"raw_records"`
 InvalidCaptureExpiry int `json:"invalid_capture_expiry"`
 EarlyPublishedBeforeExpiry int `json:"early_published_before_expiry"`
 LabelWithoutTimestamp int `json:"label_without_timestamp"`
 TimestampWithoutLabel int `json:"timestamp_without_label"`
 Status string `json:"status"`
}
func AuditRawSettlementTiming(raw []Observation)RawSettlementTiming{
 out:=RawSettlementTiming{RawRecords:len(raw),Status:"RAW_CHRONOLOGY_CHECKED_ONLY"}
 for _,o:=range raw{
  if o.CapturedAt.IsZero()||o.ExpiresAt.IsZero()||!o.CapturedAt.Before(o.ExpiresAt){out.InvalidCaptureExpiry++}
  if o.Winner!=""&&o.ResolvedAt.IsZero(){out.LabelWithoutTimestamp++}
  if o.Winner==""&&!o.ResolvedAt.IsZero(){out.TimestampWithoutLabel++}
  if o.Winner!=""&&!o.ResolvedAt.IsZero()&&!o.ExpiresAt.IsZero()&&o.ResolvedAt.Before(o.ExpiresAt){out.EarlyPublishedBeforeExpiry++}
 }
 if out.InvalidCaptureExpiry+out.EarlyPublishedBeforeExpiry+out.LabelWithoutTimestamp+out.TimestampWithoutLabel>0{out.Status="REJECT_UNTRUSTWORTHY_RAW_SETTLEMENT_TIMING"}
 return out
}
var _ time.Duration // keeps timestamp units explicit for independent validation
