package role

import (
	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

type CreateOptions struct {
	Name string
}

type UpdateOptions struct {
	Model models.Role
	Name  string
}

type Filter struct {
	IDs   []string
	ID    string
	Alias []string
	Code  []string
}

type ListOptions struct {
	Filter
}

type GetOneOptions struct {
	Filter
}

type GetOptions struct {
	Filter
	PagQuery paginator.PaginateQuery
}
