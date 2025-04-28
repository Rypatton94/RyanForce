package handlers

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"testing"
)

// TestMain sets up in-memory DB before any tests run
func TestMain(m *testing.M) {
	// Initialize temporary in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to in-memory test database")
	}
	config.DB = db

	// Migrate schema
	err = config.DB.AutoMigrate(&models.User{}, &models.Ticket{}, &models.Comment{}, &models.Account{})
	if err != nil {
		panic("failed to migrate test database schema")
	}

	// Run tests
	code := m.Run()

	// Exit with test result
	os.Exit(code)
}

// mockToken generates a JWT token for testing with given role
func mockToken(role string) (string, error) {
	return utils.GenerateJWT(1, "test@example.com", role)
}

// setupMockSession creates a fake session token for a given role
func setupMockSession(t *testing.T, role string) {
	token, err := mockToken(role)
	if err != nil {
		t.Fatalf("Failed to generate mock token: %v", err)
	}
	err = utils.SaveSession(token)
	if err != nil {
		t.Fatalf("Failed to save mock session: %v", err)
	}
}

// cleanupSession clears session after tests
func cleanupSession() {
	utils.ClearSession()
}

// ======= Admin Role Tests =======
func TestCLIHandlersAsAdmin(t *testing.T) {
	setupMockSession(t, "admin")
	defer cleanupSession()

	t.Run("AdminWhoami", func(t *testing.T) {
		handleWhoami()
	})

	t.Run("AdminHelp", func(t *testing.T) {
		handleHelp()
	})

	t.Run("AdminListTickets", func(t *testing.T) {
		handleListTickets()
	})

	t.Run("AdminViewLogs", func(t *testing.T) {
		handleViewLogs()
	})
}

// ======= Tech Role Tests =======
func TestCLIHandlersAsTech(t *testing.T) {
	setupMockSession(t, "tech")
	defer cleanupSession()

	t.Run("TechWhoami", func(t *testing.T) {
		handleWhoami()
	})

	t.Run("TechHelp", func(t *testing.T) {
		handleHelp()
	})

	t.Run("TechListTickets", func(t *testing.T) {
		handleListTickets()
	})
}

// ======= Client Role Tests =======
func TestCLIHandlersAsClient(t *testing.T) {
	setupMockSession(t, "client")
	defer cleanupSession()

	t.Run("ClientWhoami", func(t *testing.T) {
		handleWhoami()
	})

	t.Run("ClientHelp", func(t *testing.T) {
		handleHelp()
	})

	t.Run("ClientListTickets", func(t *testing.T) {
		handleListTickets()
	})
}

// ======= Expired Session Behavior =======
func TestSessionExpiredBehavior(t *testing.T) {
	utils.ClearSession()

	t.Run("NoSessionWhoami", func(t *testing.T) {
		handleWhoami()
	})

	t.Run("NoSessionHelp", func(t *testing.T) {
		handleHelp()
	})
}

// ======= Specific Functional Tests =======

func TestHandleViewTicket(t *testing.T) {
	setupMockSession(t, "admin")
	defer cleanupSession()

	// Create fake ticket
	ticket := models.Ticket{
		Title:       "Test Ticket",
		Description: "Test ticket description.",
		Priority:    "medium",
		Status:      "open",
		ClientID:    1,
	}
	if err := config.DB.Create(&ticket).Error; err != nil {
		t.Fatalf("Failed to create test ticket: %v", err)
	}

	// Create a temporary file containing the ticket ID
	tempFile, err := os.CreateTemp("", "input")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(fmt.Sprintf("%d\n", ticket.ID))
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if _, err := tempFile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek temp file: %v", err)
	}

	// Swap os.Stdin temporarily
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = tempFile

	t.Run("ViewTicket", func(t *testing.T) {
		handleViewTicket()
	})
}
