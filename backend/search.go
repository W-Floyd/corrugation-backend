package backend

import (
	"errors"
)

func dotProduct(v1 []float64, v2 []float64) (dotProduct float64, err error) {
	if len(v1) != len(v2) {
		err = errors.New("vectors should have same lenghth")
		return
	}
	for i := 0; i < len(v1); i++ {
		dotProduct += v1[i] * v2[i]
	}
	return
}

func SearchByArtifact(search string, threshold float64) (recordResults []struct {
	id    uint
	score float64
}, err error) {
	es, err := GetArtifactEmbeddings()
	if err != nil {
		return
	}

	searchEmbeddings, err := GenerateEmbeddings(search)
	if err != nil {
		return
	}

	for _, e := range es {
		var p float64
		p, err = dotProduct(e.embedding, searchEmbeddings)
		if err != nil {
			return
		}
		if p > threshold && e.artifact.RecordID != nil {
			recordResults = append(recordResults, struct {
				id    uint
				score float64
			}{
				id:    *e.artifact.RecordID,
				score: p,
			})
		}
	}

	return

}

func SearchByRecord(search string, threshold float64) (recordResults []struct {
	id    uint
	score float64
}, err error) {
	es, err := GetRecordEmbeddings()
	if err != nil {
		return
	}

	searchEmbeddings, err := GenerateEmbeddings(search)
	if err != nil {
		return
	}

	var similarityScore = map[uint]float64{}

	for id, e := range es {
		var p float64
		p, err = dotProduct(e, searchEmbeddings)
		if err != nil {
			return
		}
		similarityScore[id] = p
	}

	for id, score := range similarityScore {
		if score > threshold {
			recordResults = append(recordResults, struct {
				id    uint
				score float64
			}{
				id:    id,
				score: score,
			})
		}
	}

	return
}
