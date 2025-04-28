package main

import (
	"RyanForce/config"
	"RyanForce/controllers"
	"RyanForce/handlers"
	"RyanForce/routes"
	"RyanForce/utils"
	"html/template"
	"strconv"

	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// RyanForce CRM - Main Entry Point
// Initializes database, seeds demo data, and launches CLI or WebUI
func main() {
	utils.InitLogger(false) // Log setup

	config.Connect()

	// Check command-line arguments
	args := os.Args[1:]
	mode := "cli"
	shouldSeed := false

	for _, arg := range args {
		switch strings.ToLower(arg) {
		case "web":
			mode = "web"
		case "cli":
			mode = "cli"
		case "seed":
			shouldSeed = true
		default:
			fmt.Printf("[Startup] Unknown argument '%s' (ignored)\n", arg)
		}
	}

	if shouldSeed {
		fmt.Println("[Startup] Forcing database reseed...")
		utils.LogInfo("[Startup] Admin manually requested database reseed.")
		controllers.ClearDatabase(false) // will clear and reseed
	}

	// If we didn't seed-only, start the selected mode
	if mode == "web" {
		startWeb()
	} else {
		startCLIWithSession()
	}
}

// startWeb initializes and runs the Gin-based WebUI server
func startWeb() {
	fmt.Println("[Startup] Launching WebUI on http://localhost:8080")

	// Register template helpers
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"join":     strings.Join,
		"inc":      func(i int) int { return i + 1 },
		"dec":      func(i int) int { return i - 1 },
		"multiply": func(a, b int) int { return a * b },
		"itoa":     strconv.Itoa,
	})

	r.LoadHTMLGlob("web/templates/*.html")
	routes.SetupRouterWithEngine(r)

	if err := r.Run(":8080"); err != nil {
		utils.LogError("[WebUI] Failed to start server", err)
	}
}

// startCLIWithSession handles session checks and displays dashboard in CLI mode
func startCLIWithSession() {
	token, err := utils.LoadSession()
	if err != nil || token == "" {
		fmt.Println("[Session expired] Please log in again.")
		utils.ClearSession()

		token, claims, err := handlers.PromptLogin()
		if err != nil {
			utils.LogError("[Startup] Login failed", err)
			return
		}
		if err := utils.SaveSession(token); err != nil {
			utils.LogError("[Startup] Failed to save session", err)
			return
		}
		utils.LogInfo("[Startup] New session saved after login")
		handlers.DisplayDashboard(claims)
	} else {
		claims, _ := utils.ParseJWT(token) // Should not error here if LoadSession succeeded
		utils.LogInfo("[Startup] Resuming session")
		handlers.DisplayDashboard(claims)
	}

	startCLI()
}

// startCLI begins the interactive CLI input loop
func startCLI() {
	reader := bufio.NewReader(os.Stdin)

	for {
		utils.LogInfo("[CLI] Waiting for user input")
		fmt.Print("RyanForce > ")
		input, err := reader.ReadString('\n')
		if err != nil {
			utils.LogError("[CLI] Failed to read input", err)
			continue
		}

		input = strings.TrimSpace(strings.ToLower(input))
		switch input {
		case "logout", "exit", "quit":
			utils.LogInfo("[CLI] Session manually closed by user")
			utils.ClearSession()
			return
		default:
			handlers.HandleCommand(input)
		}
	}
}
