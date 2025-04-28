package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

// InitLogger initializes three separate log files for info, warning, and error logs.
func InitLogger(toConsole bool) {
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		fmt.Println("[Logger] Failed to create logs directory:", err)
		return
	}

	infoFile, err := os.OpenFile("logs/info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("[Logger] Failed to open info.log:", err)
		return
	}

	warnFile, err := os.OpenFile("logs/warn.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("[Logger] Failed to open warn.log:", err)
		return
	}

	errorFile, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("[Logger] Failed to open error.log:", err)
		return
	}

	infoLogger = log.New(infoFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(warnFile, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Optional: send logs to stdout for debugging purposes
	if toConsole {
		infoLogger.SetOutput(os.Stdout)
		warnLogger.SetOutput(os.Stdout)
		errorLogger.SetOutput(os.Stdout)
	}
}

func LogInfo(message string) {
	if infoLogger != nil {
		_ = infoLogger.Output(2, message)
	}
}

func LogWarning(message string) {
	if warnLogger != nil {
		_ = warnLogger.Output(2, message)
	}
}

func LogError(message string, err error) {
	if errorLogger != nil {
		_ = errorLogger.Output(2, message+": "+err.Error())
	}
}

// New — Log Info with IP Address
func LogInfoIP(message string, ip string) {
	formatted := formatIPMessage(message, ip)
	LogInfo(formatted)
}

// New — Log Warning with IP Address
func LogWarningIP(message string, ip string) {
	formatted := formatIPMessage(message, ip)
	LogWarning(formatted)
}

// New — Log Error with IP Address
func LogErrorIP(message string, err error, ip string) {
	formatted := formatIPMessage(message, ip)
	LogError(formatted, err)
}

// New — helper to format IPs into message
func formatIPMessage(message, ip string) string {
	if ip == "" {
		ip = "CLI-Local"
	}
	return fmt.Sprintf("(IP: %s) %s", ip, message)
}

func LoadRecentLogs(path string, max int) []string {
	lines := []string{}
	file, err := os.Open(path)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > max {
		return lines[len(lines)-max:]
	}
	return lines
}
