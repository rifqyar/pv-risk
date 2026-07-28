package main

import (
	// "context"

	"context"
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pv-risk/config"
	"pv-risk/controller"
	"pv-risk/migrations"
	"pv-risk/seeder"

	"github.com/gin-gonic/gin"
	webview "github.com/webview/webview_go"
)

// ================= EMBED =================

//go:embed templates/* templates/layouts/* templates/partials/*
var templateFS embed.FS

//go:embed static/* assets/*
var staticFS embed.FS

// ================= MAIN =================

func main() {

	port := "8080"
	baseURL := "http://localhost:" + port

	// ================= SINGLE INSTANCE =================
	if isServerRunning(baseURL) {
		runWebview(baseURL)
		return
	}

	// ================= INIT =================
	config.InitDB()
	migrations.Migrate(config.DB)
	seeder.SeedAll(config.DB)

	r := gin.Default()

	// ================= TEMPLATE =================
	tmpl := template.New("").Funcs(appTemplateFuncs())
	var err error
	tmpl, err = tmpl.ParseFS(templateFS,
		"templates/*.html",
		"templates/layouts/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		log.Fatal("template error:", err)
	}

	r.SetHTMLTemplate(tmpl)

	// ================= STATIC =================
	staticContent, _ := fs.Sub(staticFS, "static")
	assetsContent, _ := fs.Sub(staticFS, "assets")

	r.StaticFS("/static", http.FS(staticContent))
	r.StaticFS("/assets", http.FS(assetsContent))

	// ================= ROUTES =================
	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "splash.html", nil)
	})
	r.GET("/dashboard", controller.ShowDashboard)

	// Master Data
	masterDataRoute := r.Group("/master-data")
	{
		masterDataRoute.GET("", controller.ShowMasterData)
		masterDataRoute.POST("/equipment/save", controller.SaveMasterEquipment)
		masterDataRoute.POST("/equipment/delete/:id", controller.DeleteMasterEquipment)
		masterDataRoute.POST("/shell-material/save", controller.SaveShellMaterial)
		masterDataRoute.POST("/shell-material/delete/:id", controller.DeleteShellMaterial)
		masterDataRoute.POST("/simple-material/save", controller.SaveSimpleMaterial)
		masterDataRoute.POST("/simple-material/delete/:kind/:id", controller.DeleteSimpleMaterial)
	}

	// Assessment PV
	pressureVesselRoute := r.Group("/assessment")
	{
		pressureVesselRoute.GET("/form", controller.ShowForm)
		pressureVesselRoute.POST("/submit", controller.SubmitAssessment)
		pressureVesselRoute.GET("/list", controller.ShowListAssessment)
		pressureVesselRoute.GET("/view/:id", controller.ViewAssessmentDetail)
		pressureVesselRoute.GET("/edit/:id", controller.EditAssessment)
		pressureVesselRoute.DELETE("/delete/:id", controller.DeleteAssessment)
		pressureVesselRoute.POST("/update-validate/:id", controller.UpdateValidate)
	}

	// Assessment Pipeline
	pipelineRoute := r.Group("/assessment-pipeline")
	{
		pipelineController := controller.PipelineController()
		pipelineRoute.GET("/form", pipelineController.ShowForm)
		pipelineRoute.POST("/submit", pipelineController.SubmitAssessment)
		pipelineRoute.GET("/list", pipelineController.ShowListAssessment)
		pipelineRoute.GET("/view/:id", pipelineController.ViewAssessmentDetail)
		pipelineRoute.GET("/compare/:id", pipelineController.CompareAssessment)
		pipelineRoute.GET("/edit/:id", pipelineController.EditAssessment)
		pipelineRoute.POST("/update/:id", pipelineController.UpdateAssessment)
		pipelineRoute.POST("/preview", pipelineController.PreviewAssessment)
		pipelineRoute.POST("/calculate/:id", pipelineController.CalculateAssessment)
		pipelineRoute.POST("/approve/:id", pipelineController.ApproveAssessment)
		pipelineRoute.POST("/export/pdf/:id/audit", pipelineController.AuditPDFExport)
		pipelineRoute.GET("/export/excel/:id", pipelineController.ExportExcel)
		pipelineRoute.DELETE("/delete/:id", pipelineController.DeleteAssessment)
		pipelineRoute.GET("/gas", pipelineController.ShowGasComingSoon)
		pipelineRoute.GET("/master-data", pipelineController.ShowPipelineMasterData)
		pipelineRoute.POST("/master-data/material/save", pipelineController.SavePipelineMaterial)
		pipelineRoute.POST("/master-data/material/activate/:id", pipelineController.ActivatePipelineMaterial)
		pipelineRoute.POST("/master-data/material/deactivate/:id", pipelineController.DeactivatePipelineMaterial)
		pipelineRoute.POST("/master-data/inspection-method/save", pipelineController.SavePipelineInspectionMethod)
		pipelineRoute.POST("/master-data/inspection-method/activate/:id", pipelineController.ActivatePipelineInspectionMethod)
		pipelineRoute.POST("/master-data/inspection-method/deactivate/:id", pipelineController.DeactivatePipelineInspectionMethod)
	}

	r.GET("/api/equipment-autofill/:id", controller.GetEquipmentAutofill)
	r.GET("/api/assessment-detail/:id", controller.GetAssessmentByID)

	// === DEV ===
	// r.Run(":8080")

	// ================= SERVER =================
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// ================= WAIT SERVER READY =================
	waitForServer(baseURL)

	// ================= RUN DESKTOP =================
	runWebview(baseURL)

	// ================= SHUTDOWN =================
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("shutdown error:", err)
	}

	log.Println("app closed cleanly")
}

func appTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"toJSON": func(v interface{}) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("null")
			}
			return template.JS(b)
		},
		"fmtNum":               formatTemplateNumber,
		"isPipelineGasService": isPipelineGasService,
	}
}

func isPipelineGasService(service string) bool {
	switch strings.ToLower(strings.TrimSpace(service)) {
	case "gas", "natural gas", "dwr gas", "wet gas":
		return true
	default:
		return false
	}
}

func formatTemplateNumber(value interface{}) string {
	var numeric float64
	switch v := value.(type) {
	case float64:
		numeric = v
	case float32:
		numeric = float64(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	default:
		return ""
	}
	formatted := strconv.FormatFloat(numeric, 'f', 4, 64)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	if formatted == "-0" {
		return "0"
	}
	return formatted
}

// ================= HELPER =================

func isServerRunning(url string) bool {
	resp, err := http.Get(url)
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		return true
	}
	return false
}

func waitForServer(url string) {
	for i := 0; i < 20; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
	log.Println("warning: server may not be fully ready")
}

func runWebview(url string) {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("Fire")
	w.SetSize(1200, 800, webview.HintNone)
	w.Navigate(url)
	w.Run()
}
