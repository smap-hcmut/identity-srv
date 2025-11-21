package http

import (
	"net/http"

	"smap-project/internal/project"
	"smap-project/pkg/response"
	"smap-project/pkg/scope"

	"github.com/gin-gonic/gin"
)

// @Summary Get project detail
// @Description Get a single project by ID
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {object} ProjectResponse
// @Failure 400 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/{id} [get]
func (h handler) Detail(c *gin.Context) {
	ctx := c.Request.Context()
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		c.Abort()
		return
	}

	id := c.Param("id")
	if id == "" {
		response.Error(c, ErrInvalidID, nil)
		return
	}

	output, err := h.uc.Detail(ctx, sc, id)
	if err != nil {
		if err == project.ErrProjectNotFound {
			response.Error(c, err, nil)
			return
		}
		if err == project.ErrUnauthorized {
			response.Error(c, err, nil)
			return
		}
		h.l.Errorf(ctx, "project.delivery.http.Detail: %v", err)
		response.Error(c, err, nil)
		return
	}

	resp := ToProjectResponse(output.Project)
	response.OK(c, resp)
}

// @Summary List all projects
// @Description Get all projects for the authenticated user without pagination
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids query []string false "Filter by project IDs"
// @Param statuses query []string false "Filter by statuses"
// @Param search_name query string false "Search by project name"
// @Success 200 {array} ProjectResponse
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects [get]
func (h handler) List(c *gin.Context) {
	ctx := c.Request.Context()
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		c.Abort()
		return
	}

	var req GetProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err, nil)
		return
	}

	input := req.ToListInput()
	projects, err := h.uc.List(ctx, sc, input)
	if err != nil {
		h.l.Errorf(ctx, "project.delivery.http.List: %v", err)
		response.Error(c, err, nil)
		return
	}

	resp := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		resp[i] = ToProjectResponse(p)
	}

	response.OK(c, resp)
}

// @Summary Get projects with pagination
// @Description Get projects for the authenticated user with pagination
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids query []string false "Filter by project IDs"
// @Param statuses query []string false "Filter by statuses"
// @Param search_name query string false "Search by project name"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} ProjectListResponse
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/page [get]
func (h handler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		c.Abort()
		return
	}

	var req GetProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, err, nil)
		return
	}

	input := req.ToGetInput()
	output, err := h.uc.Get(ctx, sc, input)
	if err != nil {
		h.l.Errorf(ctx, "project.delivery.http.Get: %v", err)
		response.Error(c, err, nil)
		return
	}

	resp := ToProjectListResponse(output.Projects, output.Paginator)
	response.OK(c, resp)
}

// @Summary Create a new project
// @Description Create a new project for the authenticated user
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateProjectRequest true "Project creation request"
// @Success 201 {object} ProjectResponse
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects [post]
func (h handler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		c.Abort()
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err, nil)
		return
	}

	input := req.ToCreateInput()
	output, err := h.uc.Create(ctx, sc, input)
	if err != nil {
		if err == project.ErrInvalidStatus {
			response.Error(c, err, nil)
			return
		}
		if err == project.ErrInvalidDateRange {
			response.Error(c, err, nil)
			return
		}
		h.l.Errorf(ctx, "project.delivery.http.Create: %v", err)
		response.Error(c, err, nil)
		return
	}

	resp := ToProjectResponse(output.Project)
	response.OK(c, resp)
}

// @Summary Update a project
// @Description Update an existing project
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param request body UpdateProjectRequest true "Project update request"
// @Success 200 {object} ProjectResponse
// @Failure 400 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/{id} [put]
func (h handler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		c.Abort()
		return
	}

	id := c.Param("id")
	if id == "" {
		response.Error(c, ErrInvalidID, nil)
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err, nil)
		return
	}

	input := req.ToUpdateInput(id)
	output, err := h.uc.Update(ctx, sc, input)
	if err != nil {
		if err == project.ErrProjectNotFound {
			response.Error(c, err, nil)
			return
		}
		if err == project.ErrUnauthorized {
			response.Error(c, err, nil)
			return
		}
		if err == project.ErrInvalidStatus {
			response.Error(c, err, nil)
			return
		}
		if err == project.ErrInvalidDateRange {
			response.Error(c, err, nil)
			return
		}
		h.l.Errorf(ctx, "project.delivery.http.Update: %v", err)
		response.Error(c, err, nil)
		return
	}

	resp := ToProjectResponse(output.Project)
	response.OK(c, resp)
}

// @Summary Delete a project
// @Description Soft delete a project (sets deleted_at timestamp)
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 204 "No Content"
// @Failure 400 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/{id} [delete]
func (h handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		response.Unauthorized(c)
		c.Abort()
		return
	}

	id := c.Param("id")
	if id == "" {
		response.Error(c, ErrInvalidID, nil)
		return
	}

	err := h.uc.Delete(ctx, sc, id)
	if err != nil {
		if err == project.ErrProjectNotFound {
			response.Error(c, err, nil)
			return
		}
		if err == project.ErrUnauthorized {
			response.Error(c, err, nil)
			return
		}
		h.l.Errorf(ctx, "project.delivery.http.Delete: %v", err)
		response.Error(c, err, nil)
		return
	}

	c.Status(http.StatusNoContent)
}
