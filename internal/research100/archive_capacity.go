package research100
type ArchiveCapacity struct{
 Observed int `json:"observed"`
 Maximum int `json:"maximum"`
 WithinLimit bool `json:"within_limit"`
 Status string `json:"status"`
}
func AssessArchiveCapacity(d Dataset,max int)ArchiveCapacity{
 out:=ArchiveCapacity{Observed:len(d.Items),Maximum:max,Status:"INVALID_CAPACITY_LIMIT"}
 if max<=0||max>50000{return out}
 out.WithinLimit=len(d.Items)<=max
 if out.WithinLimit{out.Status="OFFLINE_ARCHIVE_WITHIN_LIMIT"}else{out.Status="ARCHIVE_ITEM_CAP_EXCEEDED"}
 return out
}
