package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/pkg/response"
)

// @Summary Upload file
// @Description Upload a file to MinIO storage
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param file formData file true "File to upload"
// @Param bucket_name formData string true "Bucket name"
// @Success 201 {object} uploadItem "Created"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/uploads [POST]
func (h handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processCreateRequest(c)
	if err != nil {
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	o, err := h.uc.Create(ctx, sc, req.toInput())
	if err != nil {
		h.l.Errorf(ctx, "internal.upload.http.Create.uc.Create: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	response.OK(c, h.newItem(o))
}

// @Summary Get upload detail
// @Description Get upload by ID
// @Tags Upload
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param id path string true "Upload ID"
// @Success 200 {object} uploadItem "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/uploads/{id} [GET]
func (h handler) Detail(c *gin.Context) {
	ctx := c.Request.Context()

	id, sc, err := h.processDetailRequest(c)
	if err != nil {
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	o, err := h.uc.Detail(ctx, sc, id)
	if err != nil {
		h.l.Errorf(ctx, "internal.upload.http.Detail.uc.Detail: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	response.OK(c, h.newItem(o))
}

// @Summary Get all uploads
// @Description Get all uploads with pagination
// @Tags Upload
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param bucket_name query string false "Filter by bucket name"
// @Param original_name query string false "Filter by original name"
// @Param source query string false "Filter by source"
// @Param created_user_id query string false "Filter by created user ID"
// @Success 200 {object} getUploadResp "Success"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /api/v1/uploads [GET]
func (h handler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	req, sc, err := h.processGetRequest(c)
	if err != nil {
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	o, err := h.uc.Get(ctx, sc, req.toInput())
	if err != nil {
		h.l.Errorf(ctx, "internal.upload.http.Get.uc.Get: %v", err)
		response.Error(c, h.mapErrorCode(err), h.d)
		return
	}

	response.OK(c, h.newGetResp(o))
}
