package backend

const (
	errorRecordNotFound            = "record not found"
	errorMoreRecordsThanExpected   = "more records than expected"
	errorTagNotFound               = "tag not found"
	errorMoreTagsThanExpected      = "more tags than expected"
	errorArtifactNotFound          = "artifact not found"
	errorMoreArtifactsThanExpected = "more artifacts than expected"
	topLevelName                   = "World"
)

var (
	infinityAddress    = "http://localhost:8002"
	infinityImageModel = "wkcn/TinyCLIP-ViT-8M-16-Text-3M-YFCC15M"
	infinityTextModel  = "wkcn/TinyCLIP-ViT-8M-16-Text-3M-YFCC15M"
)

func SetInfinityConfig(address, textModel string, imageModel string) {
	infinityAddress = address
	infinityImageModel = imageModel
	infinityTextModel = textModel
}
