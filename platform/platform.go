package platform

// Oss represents the oss bucket of this platform
type Oss struct {
	BucketEndpoint string `json:"bucketEndpoint"`
	Bucket         string `json:"Bucket"`
}

// Platform represents the platform
type Platform struct {
	Name string `json:"name"`
	Oss  Oss    `json:"oss"`
}
