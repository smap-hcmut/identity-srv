package role

import (
	"github.com/nguyentantai21042004/smap-api/internal/models"
	pag "github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

type CreateInput struct {
	Name string
}

type UpdateInput struct {
	ID   string
	Name string
}

type GetOneInput struct {
	Filter Filter
}

type GetInput struct {
	Filter   Filter
	PagQuery pag.PaginateQuery
}

type ListInput struct {
	Filter Filter
}

type CreateOutput struct {
	Role models.Role
}

type UpdateOutput struct {
	Role models.Role
}

type DetailOutput struct {
	Role models.Role
}

type GetOneOutput struct {
	Role models.Role
}

type GetOutput struct {
	Roles     []models.Role
	Paginator pag.Paginator
}

type ListOutput struct {
	Roles []models.Role
}
