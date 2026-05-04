package controller // Sesuaikan dengan nama package lu

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"pv-risk/config" // Sesuaikan dengan struktur folder lu

	"github.com/gin-gonic/gin"
)

// 1. Siapkan Struct khusus untuk Baris Tabel di Dashboard
type DashboardActionItem struct {
	TagNumber          string
	GoverningComponent string
	RiskLevel          string
	NextInspectionYear int
}

// 2. Siapkan Struct Utama untuk menampung semua data KPI & Tabel
type DashboardData struct {
	TotalAssets       int
	HighRisk          int
	MediumRisk        int
	InspectionsDue    int
	RecentAssessments []DashboardActionItem // Array of struct untuk looping di HTML
}

func ShowDashboard(c *gin.Context) {
	var data DashboardData
	db := config.DB
	currentYear := time.Now().Year()

	// ==========================================
	// A. QUERY UNTUK 4 KPI CARDS (TOP ROW)
	// ==========================================

	// 1. Total Assets (Menghitung jumlah equipment unik yang sudah di-ases)
	db.QueryRow("SELECT COUNT(DISTINCT equipment_id) FROM assessments").Scan(&data.TotalAssets)

	// 2. High Risk Count
	db.QueryRow("SELECT COUNT(id) FROM assessment_results WHERE risk_level = 'HIGH RISK'").Scan(&data.HighRisk)

	// 3. High Risk Count
	db.QueryRow("SELECT COUNT(id) FROM assessment_results WHERE risk_level = 'MEDIUM RISK'").Scan(&data.MediumRisk)

	// 4. Inspections Due (Jatuh tempo tahun ini atau sudah lewat)
	db.QueryRow("SELECT COUNT(id) FROM assessment_results WHERE next_inspection_year <= ?", currentYear).Scan(&data.InspectionsDue)

	// ==========================================
	// B. QUERY JOIN UNTUK TABEL ACTION ITEMS (BOTTOM ROW)
	// ==========================================
	// Kita ambil 5 aset dengan skor risiko paling tinggi (risk_index DESC)
	// Pakai COALESCE biar kalau ada data NULL di database, Golang nggak panik (panic error).
	query := `
		SELECT 
			COALESCE(t.tag_number, 'Unknown') as tag_number, 
			COALESCE(r.governing_component, '-') as governing_component, 
			COALESCE(r.risk_level, 'Pending') as risk_level, 
			COALESCE(r.next_inspection_year, 0) as next_inspection_year
		FROM assessment_results r
		JOIN assessments a ON r.assessment_id = a.id
		JOIN trx_equipments t ON a.equipment_id = t.id
		ORDER BY r.risk_index DESC, r.next_inspection_year ASC
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error fetching dashboard table data:", err)
		// Tetap render halamannya walaupun tabelnya error (graceful degradation)
		c.HTML(http.StatusOK, "dashboard.html", data)
		return
	}
	defer rows.Close()

	// ==========================================
	// C. LOOPING ROWS.SCAN() (Ini yang kemaren kurang bro!)
	// ==========================================
	for rows.Next() {
		var item DashboardActionItem

		// Scan data dari setiap baris SQL ke dalam variabel Struct item
		err := rows.Scan(
			&item.TagNumber,
			&item.GoverningComponent,
			&item.RiskLevel,
			&item.NextInspectionYear,
		)

		if err != nil {
			log.Println("Error scanning row:", err)
			continue // Kalau 1 baris error, lanjut ke baris berikutnya, jangan matiin server
		}

		// Masukkan baris yang sukses di-scan ke dalam Array/Slice utama
		data.RecentAssessments = append(data.RecentAssessments, item)
	}

	// Chart Query
	var highRisk, mediumRisk, lowRisk int
	db.QueryRow("SELECT COUNT(id) FROM assessment_results WHERE risk_level = 'HIGH RISK'").Scan(&mediumRisk)
	db.QueryRow("SELECT COUNT(id) FROM assessment_results WHERE risk_level = 'MEDIUM RISK'").Scan(&mediumRisk)
	db.QueryRow("SELECT COUNT(id) FROM assessment_results WHERE risk_level = 'LOW RISK'").Scan(&lowRisk)

	// 2. Tambahan: Rekap Tahun Inspeksi buat Bar Chart (5 Tahun ke depan)
	rowsYears, _ := db.Query(`
		SELECT next_inspection_year, COUNT(id) 
		FROM assessment_results 
		WHERE next_inspection_year >= ? 
		GROUP BY next_inspection_year 
		ORDER BY next_inspection_year ASC 
		LIMIT 5
	`, currentYear)

	var years []int
	var yearCounts []int
	for rowsYears.Next() {
		var y, c int
		rowsYears.Scan(&y, &c)
		years = append(years, y)
		yearCounts = append(yearCounts, c)
	}
	rowsYears.Close()

	yearsJSON, _ := json.Marshal(years)
	countsJSON, _ := json.Marshal(yearCounts)

	// ==========================================
	// D. RENDER KE HTML
	// ==========================================
	// Lempar struct "data" ke file template "dashboard.html"
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Dashboard":   data,
		"CurrentYear": currentYear,
		"ActiveMenu":  "dashboard", // <--- INI KUNCINYA
		"HighRisk":    highRisk,
		"MediumRisk":  mediumRisk,
		"LowRisk":     lowRisk,
		"YearsJSON":   string(yearsJSON),
		"CountsJSON":  string(countsJSON),
	})
}
