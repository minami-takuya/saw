package internal

type S3EventDetail struct {
	Version   string   `json:"version"`
	Bucket    S3Bucket `json:"bucket"`
	Object    S3Object `json:"object"`
	RequestID string   `json:"request-id"`
	Requester string   `json:"requester"`
	SourceIP  string   `json:"source-ip-address"`
	Reason    string   `json:"reason"`
}

type S3Bucket struct {
	Name string `json:"name"`
}

type S3Object struct {
	Key       string `json:"key"`
	Size      int64  `json:"size"`
	ETag      string `json:"etag"`
	VersionID string `json:"version-id,omitempty"`
	Sequencer string `json:"sequencer"`
}
