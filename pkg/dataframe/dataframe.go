package dataframe

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
)

// DataFrame represents a gene expression matrix
type DataFrame struct {
	Columns   []string    // Column names (including "Gene" as first column)
	RowLabels []string    // Gene symbols
	Data      [][]float64 // 2D slice [row][col-1] (excluding Gene column)
	NumRows   int
	NumCols   int
}

// NewDataFrame creates an empty DataFrame with specified columns
func NewDataFrame(columns []string) *DataFrame {
	return &DataFrame{
		Columns: columns,
		Data:    make([][]float64, 0),
		NumCols: len(columns),
	}
}

// AddRow adds a row to the DataFrame
func (df *DataFrame) AddRow(gene string, values []float64) {
	df.RowLabels = append(df.RowLabels, gene)
	df.Data = append(df.Data, values)
	df.NumRows++
}

// GetRowSum calculates sum of a row (excluding Gene column)
// Handles R-specific edge cases: NaN values are skipped, -Inf propagates
func (df *DataFrame) GetRowSum(rowIndex int) float64 {
	sum := 0.0
	hasValidValue := false

	for _, val := range df.Data[rowIndex] {
		if math.IsNaN(val) {
			continue
		}
		if math.IsInf(val, -1) {
			// Single -Inf makes entire sum -Inf (R behavior)
			return math.Inf(-1)
		}
		sum += val
		hasValidValue = true
	}

	if !hasValidValue {
		// All values were NaN
		return math.NaN()
	}
	return sum
}

// IsRowValid checks if a row passes filters (sum != 0, != -Inf, != NA)
func (df *DataFrame) IsRowValid(rowIndex int) bool {
	sum := df.GetRowSum(rowIndex)
	return sum != 0 && !math.IsInf(sum, -1) && !math.IsNaN(sum)
}

// AggregateByGene aggregates duplicate genes by taking max value for each column
// This replicates R's data.table behavior: temp[, lapply(.SD, max, na.rm=TRUE), by=Gene]
func (df *DataFrame) AggregateByGene() *DataFrame {
	aggregated := make(map[string][]float64)
	geneOrder := make([]string, 0)

	for i := 0; i < df.NumRows; i++ {
		gene := df.RowLabels[i]
		values := df.Data[i]

		if existing, ok := aggregated[gene]; ok {
			// Take max for each column, handling NaN like R's na.rm=TRUE
			for j := 0; j < len(values); j++ {
				if !math.IsNaN(values[j]) && (math.IsNaN(existing[j]) || values[j] > existing[j]) {
					existing[j] = values[j]
				}
			}
		} else {
			// Copy values to avoid aliasing issues
			newValues := make([]float64, len(values))
			copy(newValues, values)
			aggregated[gene] = newValues
			geneOrder = append(geneOrder, gene)
		}
	}

	// Build result preserving gene order
	result := NewDataFrame(df.Columns)
	for _, gene := range geneOrder {
		result.AddRow(gene, aggregated[gene])
	}
	return result
}

// FilterValidRows removes rows with zero/NA/-Inf sums
func (df *DataFrame) FilterValidRows() *DataFrame {
	result := NewDataFrame(df.Columns)
	for i := 0; i < df.NumRows; i++ {
		if df.IsRowValid(i) {
			result.AddRow(df.RowLabels[i], df.Data[i])
		}
	}
	return result
}

// Normalize applies log2(x+1) transformation
func (df *DataFrame) Normalize() *DataFrame {
	result := NewDataFrame(df.Columns)
	for i := 0; i < df.NumRows; i++ {
		normalized := make([]float64, len(df.Data[i]))
		for j := 0; j < len(df.Data[i]); j++ {
			// log2(x + 1) - handles zero counts correctly
			// NaN values propagate through
			if !math.IsNaN(df.Data[i][j]) {
				normalized[j] = math.Log2(df.Data[i][j] + 1)
			} else {
				normalized[j] = math.NaN()
			}
		}
		result.AddRow(df.RowLabels[i], normalized)
	}
	return result
}

// WriteTSV writes DataFrame to TSV file
func (df *DataFrame) WriteTSV(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	defer writer.Flush()

	// Write header
	if err := writer.Write(df.Columns); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for i := 0; i < df.NumRows; i++ {
		row := make([]string, df.NumCols)
		row[0] = df.RowLabels[i]
		for j := 1; j < df.NumCols; j++ {
			val := df.Data[i][j-1]
			row[j] = formatFloat(val)
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	return nil
}

// formatFloat converts float64 to string, handling special values
func formatFloat(f float64) string {
	if math.IsNaN(f) {
		return "NA"
	}
	if math.IsInf(f, -1) {
		return "-Inf"
	}
	if math.IsInf(f, 1) {
		return "Inf"
	}
	// Use default formatting - Go will format appropriately
	return fmt.Sprintf("%g", f)
}

// SortRows sorts rows by gene symbol
func (df *DataFrame) SortRows() {
	// Create indices for sorting
	indices := make([]int, df.NumRows)
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return df.RowLabels[indices[i]] < df.RowLabels[indices[j]]
	})

	// Reorder data
	newRowLabels := make([]string, df.NumRows)
	newData := make([][]float64, df.NumRows)

	for i, idx := range indices {
		newRowLabels[i] = df.RowLabels[idx]
		newData[i] = df.Data[idx]
	}

	df.RowLabels = newRowLabels
	df.Data = newData
}
