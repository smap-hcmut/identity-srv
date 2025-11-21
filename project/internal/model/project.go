package model

import (
	"encoding/json"
	"time"

	"smap-project/internal/sqlboiler"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
)

// Project represents a project entity in the domain model
type Project struct {
	ID                    string
	Name                  string
	Description           *string
	Status                string
	FromDate              time.Time
	ToDate                time.Time
	BrandName             string
	CompetitorNames       []string
	BrandKeywords         []string
	CompetitorKeywordsMap map[string][]string // JSON map of competitor -> keywords
	CreatedBy             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

// NewProjectFromDB converts SQLBoiler Project to domain Project
func NewProjectFromDB(p *sqlboiler.Project) *Project {
	if p == nil {
		return nil
	}

	project := &Project{
		ID:              p.ID,
		Name:            p.Name,
		Status:          p.Status,
		FromDate:        p.FromDate,
		ToDate:          p.ToDate,
		BrandName:       p.BrandName,
		CompetitorNames: []string(p.CompetitorNames),
		BrandKeywords:   []string(p.BrandKeywords),
		CreatedBy:       p.CreatedBy,
	}

	if p.Description.Valid {
		desc := p.Description.String
		project.Description = &desc
	}

	if p.CompetitorKeywordsMap.Valid {
		var kwMap map[string][]string
		if err := json.Unmarshal(p.CompetitorKeywordsMap.JSON, &kwMap); err == nil {
			project.CompetitorKeywordsMap = kwMap
		}
	}

	if p.CreatedAt.Valid {
		project.CreatedAt = p.CreatedAt.Time
	}

	if p.UpdatedAt.Valid {
		project.UpdatedAt = p.UpdatedAt.Time
	}

	if p.DeletedAt.Valid {
		deletedAt := p.DeletedAt.Time
		project.DeletedAt = &deletedAt
	}

	return project
}

// ToDBProject converts domain Project to SQLBoiler Project
func (p *Project) ToDBProject() *sqlboiler.Project {
	dbProject := &sqlboiler.Project{
		ID:              p.ID,
		Name:            p.Name,
		Status:          p.Status,
		FromDate:        p.FromDate,
		ToDate:          p.ToDate,
		BrandName:       p.BrandName,
		CompetitorNames: types.StringArray(p.CompetitorNames),
		BrandKeywords:   types.StringArray(p.BrandKeywords),
		CreatedBy:       p.CreatedBy,
	}

	if p.Description != nil {
		dbProject.Description = null.StringFrom(*p.Description)
	}

	if p.CompetitorKeywordsMap != nil {
		if kwMapJSON, err := json.Marshal(p.CompetitorKeywordsMap); err == nil {
			dbProject.CompetitorKeywordsMap = null.JSONFrom(kwMapJSON)
		}
	}

	if !p.CreatedAt.IsZero() {
		dbProject.CreatedAt = null.TimeFrom(p.CreatedAt)
	}

	if !p.UpdatedAt.IsZero() {
		dbProject.UpdatedAt = null.TimeFrom(p.UpdatedAt)
	}

	if p.DeletedAt != nil {
		dbProject.DeletedAt = null.TimeFrom(*p.DeletedAt)
	}

	return dbProject
}

// ProjectStatus constants
const (
	ProjectStatusDraft     = "draft"
	ProjectStatusActive    = "active"
	ProjectStatusCompleted = "completed"
	ProjectStatusArchived  = "archived"
	ProjectStatusCancelled = "cancelled"
)

// IsValidStatus checks if the given status is valid
func IsValidProjectStatus(status string) bool {
	validStatuses := []string{
		ProjectStatusDraft,
		ProjectStatusActive,
		ProjectStatusCompleted,
		ProjectStatusArchived,
		ProjectStatusCancelled,
	}

	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}
