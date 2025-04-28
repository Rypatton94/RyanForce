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

func TestMain(m *testing.M) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to in-memory test database")
	}
	config.DB = db

	err = config.DB.AutoMigrate(&models.User{}, &models.Ticket{}, &models.Comment{}, &models.Account{})
	if err != nil {
		panic("failed to migrate test database schema")
	}

	code := m.Run()
	os.Exit(code)
}

func mockToken(role string) (string, error) {
	return utils.GenerateJWT(1, "test@example.com", role)
}

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

func TestCLIHandlersAsAdmin(t *testing.T) {
	setupMockSession(t, "admin")
	defer utils.ClearSession()

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

	t.Run("AdminAssignTicket", func(t *testing.T) {
		t.Skip("Skipping because it needs manual stdin input")
	})
}

func TestCLIHandlersAsTech(t *testing.T) {
	setupMockSession(t, "tech")
	defer utils.ClearSession()

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

func TestCLIHandlersAsClient(t *testing.T) {
	setupMockSession(t, "client")
	defer utils.ClearSession()

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

func TestSessionExpiredBehavior(t *testing.T) {
	utils.ClearSession()

	t.Run("NoSessionWhoami", func(t *testing.T) {
		handleWhoami()
	})

	t.Run("NoSessionHelp", func(t *testing.T) {
		handleHelp()
	})
}

func TestHandleViewTicket(t *testing.T) {
	setupMockSession(t, "admin")
	defer utils.ClearSession()

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

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = tempFile

	t.Run("ViewTicket", func(t *testing.T) {
		handleViewTicket()
	})
}
