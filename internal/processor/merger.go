package processor

import (
	"fmt"
	"sort"

	"github.com/gerui/htseq2matrix-go/internal/htseq"
	"github.com/gerui/htseq2matrix-go/pkg/dataframe"
)

// MergeSamples merges multiple HTSeq samples into a single DataFrame
// Uses left-join semantics like dplyr::left_join in the R script
func MergeSamples(samples []htseq.HTSeqSample) (*dataframe.DataFrame, error) {
	if len(samples) == 0 {
		return nil, fmt.Errorf("no samples to merge")
	}

	// Collect all gene IDs across all samples
	allGeneIDs := make(map[string]bool)
	sampleCountMaps := make([]htseq.GeneCountMap, len(samples))
	sampleIDs := make([]string, len(samples))

	for i, sample := range samples {
		sampleIDs[i] = sample.SampleID
		countMap := sample.ToCountMap()
		sampleCountMaps[i] = countMap

		for geneID := range countMap {
			allGeneIDs[geneID] = true
		}
	}

	// Sort gene IDs for consistent ordering
	sortedGeneIDs := make([]string, 0, len(allGeneIDs))
	for geneID := range allGeneIDs {
		sortedGeneIDs = append(sortedGeneIDs, geneID)
	}
	sort.Strings(sortedGeneIDs)

	// Build columns: Gene, sample1, sample2, ...
	columns := append([]string{"Gene"}, sampleIDs...)
	df := dataframe.NewDataFrame(columns)

	// Add rows for each gene ID
	for _, geneID := range sortedGeneIDs {
		values := make([]float64, len(samples))
		for i, countMap := range sampleCountMaps {
			if count, ok := countMap[geneID]; ok {
				values[i] = count
			} else {
				// Gene not present in this sample = NA (missing value)
				values[i] = htseq.NA
			}
		}
		df.AddRow(geneID, values)
	}

	return df, nil
}
