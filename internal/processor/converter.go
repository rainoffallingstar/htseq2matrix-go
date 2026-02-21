package processor

import (
	"fmt"

	"github.com/gerui/htseq2matrix-go/internal/database"
	"github.com/gerui/htseq2matrix-go/pkg/dataframe"
)

// ConvertGeneIDs converts gene IDs to gene symbols using the database
// Uses right-join semantics: keeps ALL database entries, HTSeq IDs not in DB are dropped
func ConvertGeneIDs(df *dataframe.DataFrame, geneDB database.GeneDatabase, species string) (*dataframe.DataFrame, error) {
	result := dataframe.NewDataFrame(df.Columns)

	conversionCount := 0
	missingCount := 0

	for i := 0; i < df.NumRows; i++ {
		geneID := df.RowLabels[i]

		if symbol, ok := geneDB.GetSymbolBySpecies(geneID, species); ok {
			// Successfully converted
			result.AddRow(symbol, df.Data[i])
			conversionCount++
		} else {
			// Gene ID not found in database
			// In R's right_join, these are dropped (not in output)
			missingCount++
		}
	}

	fmt.Printf("transforming Entrezeid to Symbol\n")
	fmt.Printf("Converted %d gene IDs, %d not found in database\n", conversionCount, missingCount)

	return result, nil
}
