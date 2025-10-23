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
	ContentType string `header:"Content-Type"`
	Body        []byte
}

func RegisterHandlers(api huma.API) {

	huma.Register(api, GetStoreOperation, GetStore)
	huma.Register(api, GetStoreVersionOperation, GetStoreVersion)
	huma.Register(api, GetFirstFreeIDOperation, GetFirstFreeID)
	huma.Register(api, CreateEntityOperation, CreateEntity)
	huma.Register(api, PatchEntityOperation, PatchEntity)
	huma.Register(api, GetEntityOperation, GetEntity)
	huma.Register(api, DeleteEntityOperation, DeleteRecord)

	huma.Register(api, ListRecordsOperation, ListRecords)
	huma.Register(api, GetRecordOperation, GetRecord)
	huma.Register(api, CreateRecordOperation, CreateRecord)
	huma.Register(api, DeleteRecordOperation, DeleteRecord)

	huma.Register(api, VisualizeGraphRecordsOperation, VisualizeGraphRecords)

	huma.Register(api, ListTagsOperation, ListTags)
	huma.Register(api, GetTagOperation, GetTag)
	huma.Register(api, CreateTagOperation, CreateTag)
	huma.Register(api, DeleteTagOperation, DeleteTag)
}
