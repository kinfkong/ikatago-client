package platform

// Oss represents the oss bucket of this platform
type Oss struct {
	BucketEndpoint string `json:"bucketEndpoint"`
	Bucket         string `json:"Bucket"`
}

type Http struct {
	GetUrl     *string `json:"getUrl"`
	GetPostUrl *string `json:"getPostUrl"`
}

// Platform represents the platform
type Platform struct {
	Name string `json:"name"`
	Oss  Oss    `json:"oss"`
	Http *Http  `json:"http"`
}
