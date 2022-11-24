package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/kerti/idcra-api/model"
	"github.com/kerti/idcra-api/service"
)

type studentsConnectionResolver struct {
	students   []*model.Student
	totalCount int
	from       *string
	to         *string
}

func (r *studentsConnectionResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

func (r *studentsConnectionResolver) Edges() *[]*studentsEdgeResolver {
	l := make([]*studentsEdgeResolver, len(r.students))
	for i := range l {
		l[i] = &studentsEdgeResolver{
			cursor: service.EncodeCursor(&(r.students[i].ID)),
			model:  r.students[i],
		}
	}
	return &l
}

func (r *studentsConnectionResolver) PageInfo() *pageInfoResolver {
	var startCursor graphql.ID
	var endCursor graphql.ID

	hasNext := false

	if r.from != nil {
		startCursor = service.EncodeCursor(r.from)
	}

	if r.to != nil {
		endCursor = service.EncodeCursor(r.to)
		hasNext = true
	} else {
		hasNext = false
	}

	return &pageInfoResolver{
		startCursor: &startCursor,
		endCursor:   &endCursor,
		hasNextPage: hasNext,
	}
}
