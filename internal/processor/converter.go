package processor

import (
	"fmt"

	"github.com/gerui/htseq2matrix-go/internal/database"
	"github.com/gerui/htseq2matrix-go/pkg/dataframe"
)

const HighUnmappedRateThreshold = 0.20

// GeneIDConversionStats describes how many input IDs were converted or retained.
type GeneIDConversionStats struct {
	TotalCount       int
	ConvertedCount   int
	UnmappedCount    int
	UnmappedRate     float64
	HighUnmappedRate bool
}

// ConvertGeneIDs converts known gene IDs to symbols and preserves unknown IDs.
func ConvertGeneIDs(
	dataFrame *dataframe.DataFrame,
	geneDatabase database.GeneDatabase,
	species string,
) (*dataframe.DataFrame, GeneIDConversionStats, error) {
	if dataFrame == nil {
		return nil, GeneIDConversionStats{}, fmt.Errorf("data frame is nil")
	}
	if geneDatabase == nil {
		return nil, GeneIDConversionStats{}, fmt.Errorf("gene database is nil")
	}

	result := dataframe.NewDataFrame(dataFrame.Columns)
	statistics := GeneIDConversionStats{TotalCount: dataFrame.NumRows}

	for rowIndex := 0; rowIndex < dataFrame.NumRows; rowIndex++ {
		geneID := dataFrame.RowLabels[rowIndex]
		outputGeneID := geneID

		if symbol, found := geneDatabase.GetSymbolBySpecies(geneID, species); found && symbol != "" {
			outputGeneID = symbol
			statistics.ConvertedCount++
		} else {
			statistics.UnmappedCount++
		}

		result.AddRow(outputGeneID, dataFrame.Data[rowIndex])
	}

	if statistics.TotalCount > 0 {
		statistics.UnmappedRate = float64(statistics.UnmappedCount) / float64(statistics.TotalCount)
	}
	statistics.HighUnmappedRate = statistics.UnmappedRate > HighUnmappedRateThreshold

	fmt.Printf("transforming Entrezeid to Symbol\n")
	fmt.Printf(
		"Converted %d gene IDs, %d not found in database (retained original IDs)\n",
		statistics.ConvertedCount,
		statistics.UnmappedCount,
	)
	if statistics.HighUnmappedRate {
		fmt.Printf(
			"WARNING: %.1f%% of gene IDs were not mapped; verify that the %s mapping database matches the HTSeq gene_id namespace\n",
			statistics.UnmappedRate*100,
			species,
		)
	}

	return result, statistics, nil
}
