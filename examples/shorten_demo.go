package main

import (
	"fmt"
	"strings"
)

// shortenLine truncates lines exceeding 512 characters and appends "..."
func shortenLine(line string) string {
	const maxLength = 512
	if len(line) > maxLength {
		return line[:maxLength] + "..."
	}
	return line
}

func main() {
	fmt.Println("Saw Line Shortening Demo")
	fmt.Println("========================")
	fmt.Println()

	// Create some example log lines of varying lengths
	shortLine := "2024-04-15 14:30:00 INFO This is a normal log message"
	
	mediumLine := "2024-04-15 14:30:01 INFO Processing request with parameters: " +
		strings.Repeat("data", 50) + " and additional context information"
	
	longLine := "2024-04-15 14:30:02 ERROR Database query failed with extremely long error message: " +
		strings.Repeat("Very long error details with stack trace and context information. ", 20) +
		"End of error message"
	
	veryLongJSON := `{"timestamp":"2024-04-15T14:30:03Z","level":"INFO","message":"Processing large payload","payload":"` +
		strings.Repeat("x", 600) +
		`","user_id":"12345","request_id":"abc-def-ghi"}`

	// Display examples
	examples := []struct {
		name string
		line string
	}{
		{"Short Line (< 512 chars)", shortLine},
		{"Medium Line (~250 chars)", mediumLine},
		{"Long Line (~1400 chars)", longLine},
		{"Very Long JSON (~650 chars)", veryLongJSON},
	}

	for i, example := range examples {
		fmt.Printf("Example %d: %s\n", i+1, example.name)
		fmt.Printf("Original length: %d characters\n", len(example.line))
		fmt.Println()
		
		fmt.Println("WITHOUT --shorten:")
		fmt.Println(example.line)
		fmt.Println()
		
		shortened := shortenLine(example.line)
		fmt.Println("WITH --shorten:")
		fmt.Println(shortened)
		fmt.Printf("Shortened length: %d characters\n", len(shortened))
		
		if len(example.line) > 512 {
			fmt.Printf("✂️  Truncated %d characters\n", len(example.line)-512)
		} else {
			fmt.Println("✓ No truncation needed")
		}
		
		fmt.Println()
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println()
	}

	// Usage examples
	fmt.Println("Usage Examples:")
	fmt.Println()
	fmt.Println("  # Get logs with line shortening")
	fmt.Println("  saw get production --shorten")
	fmt.Println("  saw get production -s")
	fmt.Println()
	fmt.Println("  # Watch logs with line shortening")
	fmt.Println("  saw watch production --shorten")
	fmt.Println("  saw watch production -s")
	fmt.Println()
	fmt.Println("  # Combine with other flags")
	fmt.Println("  saw get production --shorten --pretty --filter ERROR")
	fmt.Println("  saw watch production -s --prefix api")
	fmt.Println()
	
	fmt.Println("Benefits:")
	fmt.Println("  • Prevents extremely long lines from wrapping and cluttering terminal")
	fmt.Println("  • Makes logs more readable when dealing with large payloads")
	fmt.Println("  • Useful for quick scanning of log streams")
	fmt.Println("  • Lines longer than 512 chars are cut and marked with '...'")
}