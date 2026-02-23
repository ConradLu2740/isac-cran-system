package handler

import (
	"isac-cran-system/internal/model"
	"isac-cran-system/internal/service"
	"isac-cran-system/pkg/errors"
	"isac-cran-system/pkg/response"

	"github.com/gin-gonic/gin"
)

type IRSHandler struct {
	service *service.IRSService
}

func NewIRSHandler(service *service.IRSService) *IRSHandler {
	return &IRSHandler{service: service}
}

func (h *IRSHandler) Configure(c *gin.Context) {
	var req model.IRSConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		response.ErrorWithCode(c, errors.CodeInvalidIRSConfig, err.Error())
		return
	}

	config, err := h.service.Configure(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, config)
}

func (h *IRSHandler) GetStatus(c *gin.Context) {
	status, err := h.service.GetStatus(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, status)
}

func (h *IRSHandler) GetCurrentConfig(c *gin.Context) {
	config := h.service.GetCurrentConfig()
	if config == nil {
		response.NotFound(c, "no active IRS configuration")
		return
	}

	response.Success(c, config)
}

func (h *IRSHandler) ApplyOptimal(c *gin.Context) {
	var req struct {
		TargetAngle float64 `json:"target_angle" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	config, err := h.service.ApplyOptimalPhaseShifts(c.Request.Context(), req.TargetAngle)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, config)
}

type ChannelHandler struct {
	service *service.ChannelService
}

func NewChannelHandler(service *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{service: service}
}

func (h *ChannelHandler) Collect(c *gin.Context) {
	var req model.ChannelCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	data, err := h.service.CollectData(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, data)
}

func (h *ChannelHandler) Query(c *gin.Context) {
	var query model.ChannelDataQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query parameters: "+err.Error())
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	data, total, err := h.service.QueryData(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessPage(c, data, total, query.Page, query.PageSize)
}

func (h *ChannelHandler) GetRealtime(c *gin.Context) {
	data, err := h.service.GetRealtimeData(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, data)
}

type AlgorithmHandler struct {
	service *service.AlgorithmService
}

func NewAlgorithmHandler(service *service.AlgorithmService) *AlgorithmHandler {
	return &AlgorithmHandler{service: service}
}

func (h *AlgorithmHandler) RunBeamforming(c *gin.Context) {
	var req struct {
		ExperimentID string                  `json:"experiment_id" binding:"required"`
		Params       model.BeamformingParams `json:"params" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.RunBeamforming(c.Request.Context(), req.ExperimentID, &req.Params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AlgorithmHandler) RunDOA(c *gin.Context) {
	var req struct {
		ExperimentID string          `json:"experiment_id" binding:"required"`
		Params       model.DOAParams `json:"params" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.RunDOA(c.Request.Context(), req.ExperimentID, &req.Params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AlgorithmHandler) GetResult(c *gin.Context) {
	experimentID := c.Param("id")
	if experimentID == "" {
		response.BadRequest(c, "experiment id is required")
		return
	}

	result, err := h.service.GetResult(c.Request.Context(), experimentID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AlgorithmHandler) ListResults(c *gin.Context) {
	algorithmType := c.Query("algorithm_type")
	page := 1
	pageSize := 20

	results, total, err := h.service.ListResults(c.Request.Context(), model.AlgorithmType(algorithmType), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessPage(c, results, total, page, pageSize)
}

type SensorHandler struct {
	service *service.SensorService
}

func NewSensorHandler(service *service.SensorService) *SensorHandler {
	return &SensorHandler{service: service}
}

func (h *SensorHandler) List(c *gin.Context) {
	sensorType := c.Query("sensor_type")

	sensors, err := h.service.ListSensors(c.Request.Context(), model.SensorType(sensorType))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, sensors)
}

func (h *SensorHandler) GetData(c *gin.Context) {
	var query model.SensorDataQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query parameters: "+err.Error())
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	data, err := h.service.GetSensorData(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, data)
}

func (h *SensorHandler) ReadSensor(c *gin.Context) {
	sensorID := c.Param("id")
	if sensorID == "" {
		response.BadRequest(c, "sensor id is required")
		return
	}

	data, err := h.service.ReadSensor(c.Request.Context(), sensorID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, data)
}

func (h *SensorHandler) StartCollection(c *gin.Context) {
	err := h.service.StartCollection(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "sensor data collection started", nil)
}

func (h *SensorHandler) StopCollection(c *gin.Context) {
	h.service.StopCollection()
	response.SuccessWithMessage(c, "sensor data collection stopped", nil)
}

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (h *SystemHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status":    "healthy",
		"timestamp": "now",
	})
}

func (h *SystemHandler) Info(c *gin.Context) {
	response.Success(c, gin.H{
		"name":        "ISAC-CRAN System",
		"version":     "1.0.0",
		"description": "Intelligent Reflecting Surface assisted C-RAN ISAC Experimental Prototype System",
	})
}
