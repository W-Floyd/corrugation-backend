package backend

import (
	"errors"
)

const (
	minimumImageSearchConfidence float64 = 0.2
	minimumTextSearchConfidence  float64 = 0.9
)

func dotProduct(v1 []float64, v2 []float64) (dotProduct float64, err error) {
	if len(v1) != len(v2) {
		err = errors.New("vectors should have same length")
		return
	}
	for i := range len(v1) {
		dotProduct += v1[i] * v2[i]
	}
	return
}

func SearchByArtifact(search string, recordIDs []uint) (recordResults []struct {
	id    uint
	score float64
}, err error) {
	es, err := GetArtifactEmbeddings(recordIDs)
	if err != nil {
		return
	}

	searchEmbeddings, err := GenerateImageQueryEmbeddings(search)
	if err != nil {
		return
	}

	for _, e := range es {
		if e.recordID == nil {
			continue
		}
		var p float64
		p, err = dotProduct(e.embedding, searchEmbeddings)
		if err != nil {
			return
		}
		recordResults = append(recordResults, struct {
			id    uint
			score float64
		}{id: *e.recordID, score: p})
	}

	return
}

func SearchByRecord(search string) (recordResults []struct {
	id    uint
	score float64
}, err error) {
	es, err := GetRecordEmbeddings()
	if err != nil {
		return
	}

	searchEmbeddings, err := GenerateTextQueryEmbeddings(search)
	if err != nil {
		return
	}

	for id, e := range es {
		var p float64
		p, err = dotProduct(e, searchEmbeddings)
		if err != nil {
			return
		}
		recordResults = append(recordResults, struct {
			id    uint
			score float64
		}{id: id, score: p})
	}

	return
}
