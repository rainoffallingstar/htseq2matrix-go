package htseq

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ReadHTSeqFiles reads all HTSeq files matching the pattern in the directory
func ReadHTSeqFiles(dir, postfix string) ([]HTSeqSample, error) {
	pattern := filepath.Join(dir, "*"+postfix)

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find HTSeq files with pattern %s: %w", pattern, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no HTSeq files found with pattern %s in %s", postfix, dir)
	}

	samples := make([]HTSeqSample, 0, len(files))

	for _, filePath := range files {
		sample, err := readHTSeqFile(filePath, postfix, dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}
		samples = append(samples, sample)
		fmt.Printf("processing %s\n", sample.SampleID)
	}

	return samples, nil
}

// readHTSeqFile reads a single HTSeq file
func readHTSeqFile(filePath, postfix, htseqDir string) (HTSeqSample, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return HTSeqSample{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Extract sample ID by removing postfix and directory path
	// First get the base filename
	baseName := filepath.Base(filePath)
	// Then remove the postfix to get sample ID
	sampleID := strings.TrimSuffix(baseName, postfix)

	scanner := bufio.NewScanner(file)
	records := make([]HTSeqRecord, 0)

	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Skip comment lines (e.g., enva banner starting with #)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Skip HTSeq summary lines
		if strings.HasPrefix(line, "__") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return HTSeqSample{}, fmt.Errorf("invalid line format (expected 2 columns): %s", line)
		}

		geneID := strings.TrimSpace(parts[0])
		countStr := strings.TrimSpace(parts[1])

		count, err := strconv.ParseFloat(countStr, 64)
		if err != nil {
			return HTSeqSample{}, fmt.Errorf("invalid count value '%s': %w", countStr, err)
		}

		records = append(records, HTSeqRecord{
			GeneID: geneID,
			Count:  count,
		})
	}

	if err := scanner.Err(); err != nil {
		return HTSeqSample{}, fmt.Errorf("error reading file: %w", err)
	}

	return HTSeqSample{
		SampleID: sampleID,
		Records:  records,
		Path:     filePath,
	}, nil
}

// ToCountMap converts HTSeqSample records to a map for efficient merging
func (s *HTSeqSample) ToCountMap() GeneCountMap {
	countMap := make(GeneCountMap, len(s.Records))
	for _, record := range s.Records {
		countMap[record.GeneID] = record.Count
	}
	return countMap
}
