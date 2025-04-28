package web

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"encoding/csv"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AdminReports handles GET /admin/reports
// Displays system metrics and recent logs for the admin.
// AdminReports handles GET /admin/reports
// Displays ticket metrics and filtered/paginated audit logs.
func AdminReports(c *gin.Context) {
	// Query params
	afterParam := c.Query("after")
	beforeParam := c.Query("before")
	search := c.Query("search")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	// Parse time filters
	var afterTime, beforeTime time.Time
	var err error

	if afterParam != "" {
		afterTime, err = time.Parse("2006-01-02T15:04", afterParam)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid 'after' datetime format.")
			return
		}
	}
	if beforeParam != "" {
		beforeTime, err = time.Parse("2006-01-02T15:04", beforeParam)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid 'before' datetime format.")
			return
		}
	}

	// Parse pagination values
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}

	// Ticket summary queries
	type StatusResult struct {
		Status string
		Count  int64
	}
	type PriorityResult struct {
		Priority string
		Count    int64
	}

	var statuses []StatusResult
	var priorities []PriorityResult

	db := config.DB.Model(&models.Ticket{})
	if !afterTime.IsZero() {
		db = db.Where("created_at >= ?", afterTime)
	}
	if !beforeTime.IsZero() {
		db = db.Where("created_at <= ?", beforeTime)
	}

	db.Select("status, COUNT(*) as count").Group("status").Scan(&statuses)
	db.Select("priority, COUNT(*) as count").Group("priority").Scan(&priorities)

	// Load all logs (for filtering + pagination)
	allLogs := utils.LoadRecentLogs("logs/audit.log", 0)
	if search != "" {
		filtered := []string{}
		for _, line := range allLogs {
			if strings.Contains(strings.ToLower(line), strings.ToLower(search)) {
				filtered = append(filtered, line)
			}
		}
		allLogs = filtered
	}

	// Paginate logs
	start := (page - 1) * limit
	end := start + limit
	if start > len(allLogs) {
		start = len(allLogs)
	}
	if end > len(allLogs) {
		end = len(allLogs)
	}
	logs := allLogs[start:end]

	// Render template
	c.HTML(http.StatusOK, "admin_reports.html", gin.H{
		"statuses":   statuses,
		"priorities": priorities,
		"logs":       logs,
		"after":      afterParam,
		"before":     beforeParam,
		"search":     search,
		"page":       page,
		"limit":      limit,
		"logCount":   len(allLogs),
	})
}

// ExportReportCSV handles GET /admin/reports/export
// Returns ticket summary data as a downloadable CSV file.
func ExportReportCSV(c *gin.Context) {
	afterParam := c.Query("after")
	beforeParam := c.Query("before")

	var afterTime, beforeTime time.Time
	var err error

	if afterParam != "" {
		afterTime, err = time.Parse("2006-01-02T15:04", afterParam)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid 'after' datetime format.")
			return
		}
	}

	if beforeParam != "" {
		beforeTime, err = time.Parse("2006-01-02T15:04", beforeParam)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid 'before' datetime format.")
			return
		}
	}

	var statuses []struct {
		Status string
		Count  int64
	}
	var priorities []struct {
		Priority string
		Count    int64
	}

	db := config.DB.Model(&models.Ticket{})
	if !afterTime.IsZero() {
		db = db.Where("created_at >= ?", afterTime)
	}
	if !beforeTime.IsZero() {
		db = db.Where("created_at <= ?", beforeTime)
	}

	db.Select("status, COUNT(*) as count").Group("status").Scan(&statuses)
	db.Select("priority, COUNT(*) as count").Group("priority").Scan(&priorities)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=report.csv")
	writer := csv.NewWriter(c.Writer)

	writer.Write([]string{"Section", "Value", "Count"})
	for _, s := range statuses {
		writer.Write([]string{"Status", s.Status, strconv.FormatInt(s.Count, 10)})
	}
	for _, p := range priorities {
		writer.Write([]string{"Priority", p.Priority, strconv.FormatInt(p.Count, 10)})
	}

	writer.Flush()
}

func ExportAuditCSV(c *gin.Context) {
	search := c.Query("search")
	logs := utils.LoadRecentLogs("logs/audit.log", 0)

	if search != "" {
		filtered := []string{}
		for _, line := range logs {
			if strings.Contains(strings.ToLower(line), strings.ToLower(search)) {
				filtered = append(filtered, line)
			}
		}
		logs = filtered
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=audit_logs.csv")
	writer := csv.NewWriter(c.Writer)

	writer.Write([]string{"Log Entry"})
	for _, entry := range logs {
		writer.Write([]string{entry})
	}

	writer.Flush()
}
