package upload

import (
	"github.com/nguyentantai21042004/smap-api/internal/models"
	pag "github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

type CreateOptions struct {
	Upload models.Upload
}

type GetOptions struct {
	Filter   Filter
	PagQuery pag.PaginateQuery
}
