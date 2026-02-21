package htseq

import "math"

// NA represents missing values (R's NA)
// Not a constant because math.NaN() returns a float64
var NA = math.NaN()

// HTSeqRecord represents a single gene record from HTSeq output
type HTSeqRecord struct {
	GeneID string
	Count  float64
}

// HTSeqSample represents all records from a single HTSeq file
type HTSeqSample struct {
	SampleID string
	Records  []HTSeqRecord
	Path     string
}

// GeneCountMap maps GeneID to Count for efficient lookups
type GeneCountMap map[string]float64
