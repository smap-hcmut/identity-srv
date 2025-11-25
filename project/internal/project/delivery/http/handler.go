package http

import (
	"net/http"
	"slices"

	"smap-project/pkg/response"

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

	id, sc, err := h.processDetailRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Detail.processDetailRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Detail(ctx, sc, id)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "project.http.Detail.Detail: %v", err)
		} else {
			h.l.Warnf(ctx, "project.http.Detail.Detail: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newProjectResp(output.Project))
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

	input, sc, err := h.processListRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.List.processListRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	projects, err := h.uc.List(ctx, sc, input)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "project.http.List.List: %v", err)
		} else {
			h.l.Warnf(ctx, "project.http.List.List: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newProjectListResp(projects))
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

	input, sc, err := h.processGetRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Get.processGetRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Get(ctx, sc, input)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "project.http.Get.Get: %v", err)
		} else {
			h.l.Warnf(ctx, "project.http.Get.Get: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newProjectPageResp(output.Projects, output.Paginator))
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

	input, sc, err := h.processCreateRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Create.processCreateRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Create(ctx, sc, input)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "project.http.Create.Create: %v", err)
		} else {
			h.l.Warnf(ctx, "project.http.Create.Create: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newProjectResp(output.Project))
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

	input, _, sc, err := h.processUpdateRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Update.processUpdateRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	output, err := h.uc.Update(ctx, sc, input)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "project.http.Update.Update: %v", err)
		} else {
			h.l.Warnf(ctx, "project.http.Update.Update: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newProjectResp(output.Project))
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

	id, sc, err := h.processDeleteRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Delete.processDeleteRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	err = h.uc.Delete(ctx, sc, id)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "project.http.Delete.Delete: %v", err)
		} else {
			h.l.Warnf(ctx, "project.http.Delete.Delete: %v", err)
		}
		response.Error(c, err, h.discord)
		return
	}

	c.Status(http.StatusNoContent)
}
