package database

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// GeneDatabase provides gene ID to symbol conversion
type GeneDatabase interface {
	GetSymbol(geneID string) (string, bool)
	GetSymbolBySpecies(geneID, species string) (string, bool)
	Close() error
}

// CSVDatabase implements GeneDatabase using CSV files (no CGO required)
type CSVDatabase struct {
	humanMap map[string]string
	mouseMap map[string]string
	mu       sync.RWMutex
}

// NewCSVDatabase creates a new CSV-based database instance
func NewCSVDatabase() *CSVDatabase {
	return &CSVDatabase{
		humanMap: make(map[string]string),
		mouseMap: make(map[string]string),
	}
}

// LoadSpecies loads the gene database for a specific species from CSV files
// dbPath should be the path to the directory containing gene_mapping_*.csv files
// species should be "human" or "mouse"
func (c *CSVDatabase) LoadSpecies(dbPath, species string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If already loaded, return success
	if species == "human" && len(c.humanMap) > 0 {
		return nil
	}
	if species == "mouse" && len(c.mouseMap) > 0 {
		return nil
	}

	// Determine which file to load
	var filename string
	var targetMap map[string]string
	if species == "human" {
		filename = "gene_mapping_human.csv"
		targetMap = c.humanMap
	} else {
		filename = "gene_mapping_mouse.csv"
		targetMap = c.mouseMap
	}

	filePath := filepath.Join(dbPath, filename)

	// Open CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open gene mapping file %s: %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'         // Set tab as delimiter
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Skip header if present
	firstLine := true
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	for _, record := range records {
		if len(record) < 2 {
			continue
		}

		// Check if first line is header
		if firstLine {
			firstLine = false
			if record[0] == "gene_id" || record[0] == "GeneID" || record[0] == "input_id" {
				continue // Skip header
			}
		}

		geneID := record[0]
		symbol := record[1]
		targetMap[geneID] = symbol
	}

	return nil
}

// GetSymbol returns the gene symbol for a gene ID
// Returns (symbol, true) if found, ("", false) if not found
func (c *CSVDatabase) GetSymbol(geneID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Try human map first
	if symbol, ok := c.humanMap[geneID]; ok {
		return symbol, true
	}
	// Try mouse map
	if symbol, ok := c.mouseMap[geneID]; ok {
		return symbol, true
	}
	return "", false
}

// GetSymbolBySpecies returns the gene symbol for a gene ID in a specific species
func (c *CSVDatabase) GetSymbolBySpecies(geneID, species string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if species == "human" {
		symbol, ok := c.humanMap[geneID]
		return symbol, ok
	}
	symbol, ok := c.mouseMap[geneID]
	return symbol, ok
}

// Close closes the database connection (no-op for CSV)
func (c *CSVDatabase) Close() error {
	return nil
}

// LoadBothSpecies loads both human and mouse databases
func (c *CSVDatabase) LoadBothSpecies(dbPath string) error {
	if err := c.LoadSpecies(dbPath, "human"); err != nil {
		return fmt.Errorf("failed to load human database: %w", err)
	}
	if err := c.LoadSpecies(dbPath, "mouse"); err != nil {
		return fmt.Errorf("failed to load mouse database: %w", err)
	}
	return nil
}

// ReadCSVMapping reads a CSV gene mapping file into a map
func ReadCSVMapping(csvPath string) (map[string]string, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	mapping := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// Skip header line
	if scanner.Scan() {
		line := scanner.Text()
		if line == "gene_id\tsymbol" || line == "GeneID\tSymbol" {
			// Skip header
		} else {
			// First line is data, parse it
			parts := splitLine(line, '\t')
			if len(parts) >= 2 {
				mapping[parts[0]] = parts[1]
			}
		}
	}

	for scanner.Scan() {
		parts := splitLine(scanner.Text(), '\t')
		if len(parts) >= 2 {
			mapping[parts[0]] = parts[1]
		}
	}

	return mapping, scanner.Err()
}

func splitLine(s string, sep rune) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for _, r := range s {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case sep:
			if !inQuotes {
				parts = append(parts, current.String())
				current.Reset()
				continue
			}
			fallthrough
		default:
			current.WriteRune(r)
		}
	}
	parts = append(parts, current.String())
	return parts
}

// DetectSpecies determines the species based on the postfix.
// Recognized values: "human" when postfix contains "human",
// "mouse" when postfix contains "mouse", "unknown" otherwise.
func DetectSpecies(postfix string) string {
	lower := strings.ToLower(postfix)
	if strings.Contains(lower, "human") {
		return "human"
	}
	if strings.Contains(lower, "mouse") {
		return "mouse"
	}
	return "unknown"
}
