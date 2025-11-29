package http

import (
	"time"

	"smap-project/internal/model"
	"smap-project/pkg/paginator"
)

// ProjectResp represents the HTTP response for a single project
type ProjectResp struct {
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

// ProjectListResp represents the HTTP response for multiple projects with pagination
type ProjectListResp struct {
	Projects  []ProjectResp       `json:"projects"`
	Paginator paginator.Paginator `json:"paginator"`
}

func (h handler) newProjectResp(p model.Project) ProjectResp {
	return ProjectResp{
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

func (h handler) newProjectListResp(projects []model.Project) []ProjectResp {
	resp := make([]ProjectResp, len(projects))
	for i, p := range projects {
		resp[i] = h.newProjectResp(p)
	}
	return resp
}

func (h handler) newProjectPageResp(projects []model.Project, pag paginator.Paginator) ProjectListResp {
	return ProjectListResp{
		Projects:  h.newProjectListResp(projects),
		Paginator: pag,
	}
}

type SuggestKeywordsResp struct {
	NicheKeywords    []string `json:"niche_keywords"`
	NegativeKeywords []string `json:"negative_keywords"`
}

func (h handler) newSuggestKeywordsResp(niche []string, negative []string) SuggestKeywordsResp {
	return SuggestKeywordsResp{
		NicheKeywords:    niche,
		NegativeKeywords: negative,
	}
}

type DryRunKeywordsResp struct {
	Posts []interface{} `json:"posts"`
}

func (h handler) newDryRunKeywordsResp(p []interface{}) DryRunKeywordsResp {
	return DryRunKeywordsResp{
		Posts: p,
	}
}
