package project

import (
	"time"

	"smap-project/internal/model"
	"smap-project/pkg/paginator"
)

// CreateInput represents input for creating a new project
type CreateInput struct {
	Name                  string
	Description           *string
	Status                string
	FromDate              time.Time
	ToDate                time.Time
	BrandName             string
	CompetitorNames       []string
	BrandKeywords         []string
	CompetitorKeywordsMap map[string][]string // Map competitor name to their keywords
}

// UpdateInput represents input for updating a project
type UpdateInput struct {
	ID                    string
	Name                  *string
	Description           *string
	Status                *string
	FromDate              *time.Time
	ToDate                *time.Time
	BrandName             *string
	CompetitorNames       []string
	BrandKeywords         []string
	CompetitorKeywordsMap map[string][]string
}

// ProjectOutput represents output for a single project
type ProjectOutput struct {
	Project model.Project
}

// GetProjectOutput represents output for multiple projects with pagination
type GetProjectOutput struct {
	Projects  []model.Project
	Paginator paginator.Paginator
}

// GetOneInput represents input for getting a single project
type GetOneInput struct {
	ID string
}

// ListInput represents input for listing projects
type ListInput struct {
	Filter Filter
}

// GetInput represents input for getting projects with pagination
type GetInput struct {
	Filter        Filter
	PaginateQuery paginator.PaginateQuery
}

// Filter represents filtering options for projects
type Filter struct {
	IDs        []string
	Statuses   []string
	CreatedBy  *string // User ID who created the projects
	SearchName *string // Search by project name
}
