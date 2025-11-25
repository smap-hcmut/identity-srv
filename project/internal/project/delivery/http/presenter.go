package http

import (
	"time"

	"smap-project/internal/model"
	"smap-project/pkg/paginator"
)

// ProjectResponse represents the HTTP response for a single project
type ProjectResponse struct {
	ID                    string              `json:"id"`
	Name                  string              `json:"name"`
	Description           *string             `json:"description,omitempty"`
	Status                string              `json:"status"`
	FromDate              time.Time           `json:"from_date"`
	ToDate                time.Time           `json:"to_date"`
	BrandName             string              `json:"brand_name"`
	CompetitorNames       []string            `json:"competitor_names,omitempty"`
	BrandKeywords         []string            `json:"brand_keywords"`
	CompetitorKeywordsMap map[string][]string `json:"competitor_keywords_map,omitempty"`
	CreatedBy             string              `json:"created_by"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}

// ProjectListResponse represents the HTTP response for multiple projects with pagination
type ProjectListResponse struct {
	Projects  []ProjectResponse   `json:"projects"`
	Paginator paginator.Paginator `json:"paginator"`
}

func (h handler) newProjectResp(p model.Project) ProjectResponse {
	return ProjectResponse{
		ID:                    p.ID,
		Name:                  p.Name,
		Description:           p.Description,
		Status:                p.Status,
		FromDate:              p.FromDate,
		ToDate:                p.ToDate,
		BrandName:             p.BrandName,
		CompetitorNames:       p.CompetitorNames,
		BrandKeywords:         p.BrandKeywords,
		CompetitorKeywordsMap: p.CompetitorKeywordsMap,
		CreatedBy:             p.CreatedBy,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}

func (h handler) newProjectListResp(projects []model.Project) []ProjectResponse {
	resp := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		resp[i] = h.newProjectResp(p)
	}
	return resp
}

func (h handler) newProjectPageResp(projects []model.Project, pag paginator.Paginator) ProjectListResponse {
	return ProjectListResponse{
		Projects:  h.newProjectListResp(projects),
		Paginator: pag,
	}
}
