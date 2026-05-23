package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type NewPipelineController struct {
}

func PipelineController() *NewPipelineController {
	return &NewPipelineController{}
}

func (ctrl *NewPipelineController) ShowForm(c *gin.Context) {
	c.HTML(http.StatusOK, "pipeline_assessment_form.html", gin.H{
		"ActiveMenu": "assessment-pipeline",
	})
}

func (ctrl *NewPipelineController) SubmitAssessment(c *gin.Context) {}

func (ctrl *NewPipelineController) ShowListAssessment(c *gin.Context) {}

func (ctrl *NewPipelineController) ViewAssessmentDetail(c *gin.Context) {}

func (ctrl *NewPipelineController) EditAssessment(c *gin.Context) {}

func (ctrl *NewPipelineController) DeleteAssessment(c *gin.Context) {}
