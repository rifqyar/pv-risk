package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
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
	tmpl := template.New("").Funcs(template.FuncMap{
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
	})

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
	r.GET("/", controller.ShowDashboard)
	r.GET("/dashboard", controller.ShowDashboard)
	r.GET("/assessment/form", controller.ShowForm)
	r.POST("/submit", controller.SubmitAssessment)
	r.GET("/assessment/list", controller.ShowListAssessment)
	r.GET("/assessment/view/:id", controller.ViewAssessmentDetail)
	r.GET("/assessment/edit/:id", controller.EditAssessment)
	r.DELETE("/assessment/delete/:id", controller.DeleteAssessment)
	r.POST("/assessment/update-validate/:id", controller.UpdateValidate)

	r.GET("/api/equipment-autofill/:id", controller.GetEquipmentAutofill)
	r.GET("/api/assessment-detail/:id", controller.GetAssessmentByID)

	// === DEV ===
	r.Run(":8080")

	// ================= SERVER =================
	// srv := &http.Server{
	// 	Addr:    ":" + port,
	// 	Handler: r,
	// }

	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("server error: %v", err)
	// 	}
	// }()

	// // ================= WAIT SERVER READY =================
	// waitForServer(baseURL)

	// // ================= RUN DESKTOP =================
	// runWebview(baseURL)

	// // ================= SHUTDOWN =================
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// if err := srv.Shutdown(ctx); err != nil {
	// 	log.Println("shutdown error:", err)
	// }

	// log.Println("app closed cleanly")
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
