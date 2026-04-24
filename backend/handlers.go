package backend

import (
	"github.com/danielgtaylor/huma/v2"
)

type EmptyOutput struct {
	Body struct{}
}

type PlaintextOutput struct {
	ContentType string `header:"Content-Type"`
	Body        string
}

type BytesOutput struct {
	ContentType  string `header:"Content-Type"`
	CacheControl string `header:"Cache-Control"`
	ETag         string `header:"ETag"`
	Body         []byte
}

func RegisterHandlers(api huma.API) {

	huma.Register(api, ExportOperation, Export)
	huma.Register(api, ImportOperation, Import)
	huma.Register(api, GetStoreOperation, GetStore)
	huma.Register(api, GetStoreVersionOperation, GetStoreVersion)
	huma.Register(api, ResetStoreOperation, ResetStore)
	huma.Register(api, GetFirstFreeIDOperation, GetFirstFreeID)
	huma.Register(api, CreateEntityOperation, CreateEntity)
	huma.Register(api, PatchEntityOperation, PatchEntity)
	huma.Register(api, ReplaceEntityOperation, ReplaceEntity)
	huma.Register(api, GetQRCodeOperation, GetQRCode)
	huma.Register(api, GetEntityQRCodeOperation, GetEntityQRCode)
	huma.Register(api, GetNextIDOperation, GetNextID)
	huma.Register(api, GetFirstLabeledIDOperation, GetFirstLabeledID)
	huma.Register(api, FindLocationsOperation, FindLocations)
	huma.Register(api, FindLocationsFullOperation, FindLocationsFull)
	huma.Register(api, GetChildrenFullOperation, GetChildrenFull)
	huma.Register(api, GetChildrenFullRecursiveOperation, GetChildrenFullRecursive)
	huma.Register(api, GetEntityContainsOperation, GetEntityContains)
	huma.Register(api, ListEntityIDsOperation, ListEntityIDs)
	huma.Register(api, GetAllEntitiesOperation, GetAllEntities)
	huma.Register(api, GetEntityOperation, GetEntity)
	huma.Register(api, DeleteEntityOperation, DeleteEntity)

	huma.Register(api, GetArtifactQRCodeOperation, GetArtifactQRCode)
	huma.Register(api, DeleteArtifactOperation, DeleteArtifact)
	huma.Register(api, ListArtifactsOperation, ListArtifacts)
	huma.Register(api, CreateArtifactStoreOperation, CreateArtifactStore)
	huma.Register(api, GetArtifactStoreOperation, GetArtifact)

	huma.Register(api, ListRecordsOperation, ListRecords)
	huma.Register(api, GetRecordOperation, GetRecord)
	huma.Register(api, CreateRecordOperation, CreateRecord)
	huma.Register(api, DeleteRecordOperation, DeleteRecord)

	huma.Register(api, VisualizeGraphRecordsOperation, VisualizeGraphRecords)

	huma.Register(api, CreateArtifactOperation, CreateArtifact)
	huma.Register(api, GetArtifactOperation, GetArtifact)

	huma.Register(api, ListTagsOperation, ListTags)
	huma.Register(api, GetTagOperation, GetTag)
	huma.Register(api, CreateTagOperation, CreateTag)
	huma.Register(api, DeleteTagOperation, DeleteTag)
}
