package output

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gerui/htseq2matrix-go/pkg/dataframe"
)

// WriteMatrix writes a DataFrame to TSV file
func WriteMatrix(df *dataframe.DataFrame, outputDir, filename string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filepath := filepath.Join(outputDir, filename)

	if err := df.WriteTSV(filepath); err != nil {
		return fmt.Errorf("failed to write matrix: %w", err)
	}

	return nil
}

// WriteCountMatrix writes the raw count matrix
func WriteCountMatrix(df *dataframe.DataFrame, outputDir string) error {
	return WriteMatrix(df, outputDir, "matrix_count.txt")
}

// WriteNormalizedMatrix writes the log2-normalized matrix
func WriteNormalizedMatrix(df *dataframe.DataFrame, outputDir string) error {
	return WriteMatrix(df, outputDir, "matrix_norm.txt")
}
