package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"pv-risk/config"
	"pv-risk/controller"
	"pv-risk/migrations"
	"pv-risk/seeder"
	"runtime"

	"github.com/gin-gonic/gin"
)

//go:embed templates/* templates/layouts/* templates/partial/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}

	if err != nil {
		log.Println(err)
	}
}

func main() {

	// Initialize database connection
	config.InitDB()
	migrations.Migrate(config.DB)
	seeder.SeedAll(config.DB)

	r := gin.Default()
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

	// tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))
	// tmpl = template.Must(tmpl.ParseGlob("templates/layouts/*.html"))
	tmpl, err := tmpl.ParseFS(templateFS,
		"templates/*.html",         // Untuk backup_assessment_form.html dll di root templates
		"templates/layouts/*.html", // Untuk master.html
		"templates/partial/*.html", // Untuk partials kamu
	)

	if err != nil {
		log.Fatal("Gagal parsing template: ", err)
	}

	r.SetHTMLTemplate(tmpl)
	// r.Static("/static", "./static")
	staticContent, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(staticContent))

	r.GET("/", controller.ShowForm)
	r.POST("/submit", controller.SubmitAssessment)

	r.Run(":8080")
	// // Run server in goroutine
	// go func() {
	// 	r.Run(":8080")
	// }()

	// // Wait until server ready
	// for i := 0; i < 20; i++ {
	// 	resp, err := http.Get("http://localhost:8080")
	// 	if err == nil {
	// 		resp.Body.Close()
	// 		break
	// 	}
	// 	time.Sleep(300 * time.Millisecond)
	// }

	// openBrowser("http://localhost:8080")

	// select {}
}
