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
// @Security CookieAuth
// @Param id path string true "Project ID"
// @Success 200 {object} ProjectResp
// @Failure 400 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/{id} [get]
func (h handler) Detail(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processDetailReq(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Detail.processDetailReq: %v", err)
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
// @Security CookieAuth
// @Param ids query []string false "Filter by project IDs"
// @Param statuses query []string false "Filter by statuses"
// @Param search_name query string false "Search by project name"
// @Success 200 {array} ProjectResp
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects [get]
func (h handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	input, sc, err := h.processListReq(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.List.processListReq: %v", err)
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
// @Security CookieAuth
// @Param ids query []string false "Filter by project IDs"
// @Param statuses query []string false "Filter by statuses"
// @Param search_name query string false "Search by project name"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} ProjectListResp
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/page [get]
func (h handler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	input, sc, err := h.processGetReq(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Get.processGetReq: %v", err)
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
// @Security CookieAuth
// @Param request body CreateReq true "Project creation request"
// @Success 201 {object} ProjectResp
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects [post]
func (h handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	input, sc, err := h.processCreateReq(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Create.processCreateReq: %v", err)
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
// @Security CookieAuth
// @Param id path string true "Project ID"
// @Param request body UpdateReq true "Project update request"
// @Success 200 {object} ProjectResp
// @Failure 400 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/{id} [put]
func (h handler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	input, _, sc, err := h.processUpdateReq(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Update.processUpdateReq: %v", err)
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
// @Security CookieAuth
// @Param id path string true "Project ID"
// @Success 204 "No Content"
// @Failure 400 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 403 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/{id} [delete]
func (h handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processDeleteReq(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.Delete.processDeleteReq: %v", err)
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

// @Summary Suggest keywords
// @Description Suggest niche and negative keywords based on brand name
// @Tags Projects
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body SuggestKeywordsReq true "Suggestion request"
// @Success 200 {object} SuggestKeywordsResp
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/keywords/suggest [post]
func (h handler) SuggestKeywords(c *gin.Context) {
	ctx := c.Request.Context()

	brandName, sc, err := h.processSuggestKeywordsReq(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.SuggestKeywords.processSuggestKeywordsReq: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	niche, negative, err := h.uc.SuggestKeywords(ctx, sc, brandName)
	if err != nil {
		h.l.Errorf(ctx, "project.http.SuggestKeywords.SuggestKeywords: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newSuggestKeywordsResp(niche, negative))
}

// @Summary Dry run keywords
// @Description Fetch sample data for keywords
// @Tags Projects
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body DryRunKeywordsReq true "Dry run request"
// @Success 200 {object} DryRunKeywordsResp
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /projects/keywords/dry-run [post]
func (h handler) DryRunKeywords(c *gin.Context) {
	ctx := c.Request.Context()

	keywords, sc, err := h.processDryRunKeywordsReq(c)
	if err != nil {
		h.l.Errorf(ctx, "project.http.DryRunKeywords.processDryRunKeywordsReq: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	posts, err := h.uc.DryRunKeywords(ctx, sc, keywords)
	if err != nil {
		h.l.Errorf(ctx, "project.http.DryRunKeywords.DryRunKeywords: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	response.OK(c, h.newDryRunKeywordsResp(posts))
}
