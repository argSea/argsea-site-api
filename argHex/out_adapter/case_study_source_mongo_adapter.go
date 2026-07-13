package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// legacyProjectDoc reads only the fields the caselog migration needs off a raw
// project doc: the dormant caseStudy string plus the header-seed inputs. The
// live Project struct no longer decodes caseStudy, so this narrow read is the
// one place the legacy field is still pulled off the wire.
type legacyProjectDoc struct {
	Id        primitive.ObjectID   `bson:"_id"`
	Title     string               `bson:"title"`
	ShortDesc string               `bson:"shortDesc"`
	FirstLit  string               `bson:"firstLit"`
	Tags      []string             `bson:"tags"`
	Facts     []domain.ProjectFact `bson:"facts"`
	CaseStudy string               `bson:"caseStudy"`
}

type caseStudySourceMongoAdapter struct {
	store *stores.Mordor
}

func NewCaseStudySourceMongoAdapter(store *stores.Mordor) out_port.CaseStudySource {
	return caseStudySourceMongoAdapter{
		store: store,
	}
}

// LegacyCaseStudies returns every project doc still carrying a non-empty
// caseStudy string, ordered by _id so the migration's log line is stable across
// runs.
func (c caseStudySourceMongoAdapter) LegacyCaseStudies() (domain.LegacyCaseStudies, error) {
	var docs []legacyProjectDoc
	filter := bson.M{"caseStudy": bson.M{"$exists": true, "$ne": ""}}

	if _, err := c.store.Find(filter, 0, 0, bson.D{{Key: "_id", Value: 1}}, &docs); nil != err {
		return nil, err
	}

	out := make(domain.LegacyCaseStudies, 0, len(docs))

	for _, doc := range docs {
		out = append(out, domain.LegacyCaseStudy{
			ProjectId: doc.Id.Hex(),
			Title:     doc.Title,
			ShortDesc: doc.ShortDesc,
			FirstLit:  doc.FirstLit,
			Tags:      doc.Tags,
			Facts:     doc.Facts,
			CaseStudy: doc.CaseStudy,
		})
	}

	return out, nil
}
