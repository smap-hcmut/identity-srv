package http

import (
	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type createReq struct {
	Name string `json:"name" binding:"required"`
}

func (r createReq) toInput() role.CreateInput {
	return role.CreateInput{
		Name: r.Name,
	}
}

type updateReq struct {
	ID   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (r updateReq) toInput() role.UpdateInput {
	return role.UpdateInput{
		ID:   r.ID,
		Name: r.Name,
	}
}

type deleteReq struct {
	IDs []string `json:"ids" binding:"required"`
}

func (r deleteReq) validate() error {
	if len(r.IDs) == 0 {
		return errWrongBody
	}

	for _, id := range r.IDs {
		if _, err := primitive.ObjectIDFromHex(id); err != nil {
			return errWrongBody
		}
	}

	return nil
}

type detailResp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Alias       string `json:"alias"`
	Description string `json:"description,omitempty"`
}

func (h handler) newGetOneResp(output role.GetOneOutput) detailResp {
	return detailResp{
		ID:          output.Role.ID.Hex(),
		Name:        output.Role.Name,
		Code:        output.Role.Code,
		Alias:       output.Role.Alias,
		Description: output.Role.Description,
	}
}

func (h handler) newDetailResp(output role.CreateOutput) detailResp {
	return detailResp{
		ID:          output.Role.ID.Hex(),
		Name:        output.Role.Name,
		Code:        output.Role.Code,
		Alias:       output.Role.Alias,
		Description: output.Role.Description,
	}
}

func (h handler) newUpdateResp(output role.UpdateOutput) detailResp {
	return detailResp{
		ID:          output.Role.ID.Hex(),
		Name:        output.Role.Name,
		Code:        output.Role.Code,
		Alias:       output.Role.Alias,
		Description: output.Role.Description,
	}
}

type listRespItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Alias       string `json:"alias"`
	Description string `json:"description,omitempty"`
}

type listReq struct {
	IDs   []string `form:"ids[]"`
	Alias []string `form:"alias[]"`
	Code  []string `form:"code[]"`
}

func (r listReq) toInput() role.ListInput {
	return role.ListInput{
		Filter: role.Filter{
			IDs:   r.IDs,
			Alias: r.Alias,
			Code:  r.Code,
		},
	}
}

func (h handler) newListResp(output role.ListOutput) []listRespItem {
	items := make([]listRespItem, 0, len(output.Roles))
	for _, roleItem := range output.Roles {
		items = append(items, listRespItem{
			ID:          roleItem.ID.Hex(),
			Name:        roleItem.Name,
			Code:        roleItem.Code,
			Alias:       roleItem.Alias,
			Description: roleItem.Description,
		})
	}
	return items
}

type getOneReq struct {
	ID string
}

func (r getOneReq) toInput() role.GetOneInput {
	return role.GetOneInput{
		Filter: role.Filter{ID: r.ID},
	}
}

type getReq struct {
	IDs   []string `form:"ids[]"`
	Alias []string `form:"alias[]"`
	Code  []string `form:"code[]"`
}

func (r getReq) toInput() role.GetInput {
	return role.GetInput{
		Filter: role.Filter{
			IDs:   r.IDs,
			Alias: r.Alias,
			Code:  r.Code,
		},
	}
}

type getMetaResponse struct {
	paginator.PaginatorResponse
}

type getResp struct {
	Items []listRespItem  `json:"items"`
	Meta  getMetaResponse `json:"meta"`
}

func (h handler) newGetResp(output role.GetOutput) getResp {
	listOutput := role.ListOutput{Roles: output.Roles}
	return getResp{
		Items: h.newListResp(listOutput),
		Meta:  getMetaResponse{PaginatorResponse: output.Paginator.ToResponse()},
	}
}