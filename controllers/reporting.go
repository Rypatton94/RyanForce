package controllers

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// ReportStatus prints the number of tickets in each status category.
func ReportStatus() {
	type Result struct {
		Status string
		Count  int64
	}

	var results []Result
	config.DB.Model(&models.Ticket{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results)

	fmt.Println("\nTicket Count by Status")
	fmt.Println("--------------------------")
	for _, r := range results {
		fmt.Printf("%-25s : %d\n", r.Status, r.Count)
	}
	utils.LogInfo("[Report] Status report generated")
}

// ReportPriority prints the number of tickets by priority level.
func ReportPriority() {
	type Result struct {
		Priority string
		Count    int64
	}

	var results []Result
	config.DB.Model(&models.Ticket{}).
		Select("priority, COUNT(*) as count").
		Group("priority").
		Scan(&results)

	fmt.Println("\nTicket Count by Priority")
	fmt.Println("----------------------------")
	for _, r := range results {
		fmt.Printf("%-10s : %d\n", r.Priority, r.Count)
	}
	utils.LogInfo("[Report] Priority report generated")
}

// ReportUnassigned lists all tickets that are not assigned to a technician.
func ReportUnassigned() {
	var tickets []models.Ticket
	config.DB.Where("tech_id IS NULL").Find(&tickets)

	fmt.Println("\nUnassigned Tickets")
	fmt.Println("----------------------")
	if len(tickets) == 0 {
		fmt.Println("All tickets are assigned.")
		utils.LogInfo("[Report] No unassigned tickets found")
		return
	}

	for _, t := range tickets {
		fmt.Printf("ID: %d | Title: %s | Status: %s\n", t.ID, t.Title, t.Status)
	}
	utils.LogInfo(fmt.Sprintf("[Report] Found %d unassigned tickets", len(tickets)))
}

// ReportResolutionMetrics calculates SLA compliance and average resolution time.
func ReportResolutionMetrics() {
	var tickets []models.Ticket
	config.DB.Where("closed_at IS NOT NULL").Find(&tickets)

	if len(tickets) == 0 {
		fmt.Println("No closed tickets available for resolution metrics.")
		utils.LogInfo("[Report] No closed tickets available for SLA metrics")
		return
	}

	slaTargets := map[string]time.Duration{
		"low":      72 * time.Hour,
		"medium":   48 * time.Hour,
		"high":     24 * time.Hour,
		"critical": 4 * time.Hour,
	}

	var totalTime time.Duration
	slaCompliance := map[string]struct {
		total  int
		onTime int
	}{}

	for _, t := range tickets {
		resolutionTime := t.ClosedAt.Sub(t.CreatedAt)
		totalTime += resolutionTime
		priority := t.Priority
		sla := slaTargets[priority]
		entry := slaCompliance[priority]
		entry.total++
		if resolutionTime <= sla {
			entry.onTime++
		}
		slaCompliance[priority] = entry
	}

	avgResolution := totalTime / time.Duration(len(tickets))

	fmt.Println("\nAverage Resolution Time")
	fmt.Println("--------------------------")
	fmt.Printf("Overall Average: %v\n", avgResolution)

	fmt.Println("\nSLA Compliance by Priority")
	fmt.Println("-----------------------------")
	for priority, data := range slaCompliance {
		rate := float64(data.onTime) / float64(data.total) * 100
		fmt.Printf("%-10s | On-time: %2d/%2d (%.1f%%)\n", priority, data.onTime, data.total, rate)
	}
	utils.LogInfo("[Report] SLA resolution metrics generated")
}

// ReportOverdueTickets identifies open tickets that have exceeded their SLA deadline.
func ReportOverdueTickets() {
	var tickets []models.Ticket
	config.DB.Where("status != ?", "closed").Find(&tickets)

	if len(tickets) == 0 {
		fmt.Println("No open tickets found.")
		utils.LogInfo("[Report] No open tickets found for overdue scan")
		return
	}

	slaTargets := map[string]time.Duration{
		"low":      72 * time.Hour,
		"medium":   48 * time.Hour,
		"high":     24 * time.Hour,
		"critical": 4 * time.Hour,
	}

	now := time.Now()
	var overdue []models.Ticket

	for _, t := range tickets {
		sla, ok := slaTargets[t.Priority]
		if !ok {
			continue
		}
		if now.Sub(t.CreatedAt) > sla {
			overdue = append(overdue, t)
		}
	}

	fmt.Println("\nOverdue Tickets (open tickets past SLA deadline)")
	fmt.Println("-----------------------------------------------------")

	if len(overdue) == 0 {
		fmt.Println("All open tickets are within SLA.")
		utils.LogInfo("[Report] All open tickets are within SLA")
		return
	}

	for _, t := range overdue {
		elapsed := now.Sub(t.CreatedAt)
		fmt.Printf("ID: %d | Priority: %-8s | Status: %-20s | Open for: %s\n",
			t.ID, t.Priority, t.Status, elapsed.Round(time.Minute))
	}
	utils.LogInfo(fmt.Sprintf("[Report] Found %d overdue tickets", len(overdue)))
}

// ReportAll prints status, priority, SLA, and overdue reports.
func ReportAll() {
	fmt.Println("\n========== RYANFORCE REPORT SUMMARY ==========\n")
	ReportStatus()
	ReportPriority()
	ReportResolutionMetrics()
	ReportOverdueTickets()
	fmt.Println("\n================================================\n")
	utils.LogInfo("[Report] Full summary report generated")
}

// ExportTicketsCSV writes all tickets to a CSV in the program's logs directory.
func ExportTicketsCSV() {
	var tickets []models.Ticket
	config.DB.Find(&tickets)

	if len(tickets) == 0 {
		fmt.Println("No tickets found to export.")
		utils.LogWarning("[Export] No tickets found to export")
		return
	}

	file, err := os.Create("tickets_export.csv")
	if err != nil {
		fmt.Println("[Error] Unable to create file:", err)
		utils.LogError("[Export] Failed to create export file", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"ID", "Title", "Description", "Priority", "Status",
		"ClientID", "TechID", "CreatedAt", "ClosedAt",
	})

	for _, t := range tickets {
		techID := ""
		if t.TechID != nil {
			techID = strconv.Itoa(int(*t.TechID))
		}
		closedAt := ""
		if t.ClosedAt != nil {
			closedAt = t.ClosedAt.Format("2006-01-02 15:04:05")
		}

		record := []string{
			strconv.Itoa(int(t.ID)),
			t.Title,
			t.Description,
			t.Priority,
			t.Status,
			strconv.Itoa(int(t.ClientID)),
			techID,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			closedAt,
		}
		writer.Write(record)
	}

	fmt.Println("Exported tickets to tickets_export.csv")
	utils.LogInfo("[Export] Ticket export completed")
}
