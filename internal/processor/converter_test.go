package processor

import (
	"testing"

	"github.com/gerui/htseq2matrix-go/internal/database"
	"github.com/gerui/htseq2matrix-go/pkg/dataframe"
)

type testGeneDatabase struct {
	mapping map[string]string
}

func (database *testGeneDatabase) GetSymbol(geneID string) (string, bool) {
	symbol, found := database.mapping[geneID]
	return symbol, found
}

func (database *testGeneDatabase) GetSymbolBySpecies(geneID, species string) (string, bool) {
	symbol, found := database.mapping[geneID]
	return symbol, found
}

func (database *testGeneDatabase) Close() error {
	return nil
}

func TestConvertGeneIDsPreservesUnmappedIDsAndReturnsStatistics(t *testing.T) {
	dataFrame := dataframe.NewDataFrame([]string{"Gene", "sample"})
	dataFrame.AddRow("ENSG000001", []float64{10})
	dataFrame.AddRow("ENSG_UNMAPPED", []float64{20})

	geneDatabase := &testGeneDatabase{
		mapping: map[string]string{"ENSG000001": "KNOWN"},
	}

	converted, statistics, err := ConvertGeneIDs(dataFrame, geneDatabase, "human")
	if err != nil {
		t.Fatalf("ConvertGeneIDs returned an unexpected error: %v", err)
	}
	if converted.NumRows != 2 {
		t.Fatalf("converted row count = %d, want 2", converted.NumRows)
	}
	if converted.RowLabels[0] != "KNOWN" {
		t.Errorf("mapped row label = %q, want KNOWN", converted.RowLabels[0])
	}
	if converted.RowLabels[1] != "ENSG_UNMAPPED" {
		t.Errorf("unmapped row label = %q, want original ID", converted.RowLabels[1])
	}
	if converted.Data[1][0] != 20 {
		t.Errorf("unmapped row count = %v, want 20", converted.Data[1][0])
	}
	if statistics.TotalCount != 2 || statistics.ConvertedCount != 1 || statistics.UnmappedCount != 1 {
		t.Errorf("conversion statistics = %+v, want total=2 converted=1 unmapped=1", statistics)
	}
	if statistics.UnmappedRate != 0.5 {
		t.Errorf("unmapped rate = %v, want 0.5", statistics.UnmappedRate)
	}
	if !statistics.HighUnmappedRate {
		t.Error("50%% unmapped should be reported as a high unmapped rate")
	}
}

func TestConvertGeneIDsPreservesMouseEnsemblIDsWhenEmbeddedMappingUsesDifferentNamespace(t *testing.T) {
	geneDatabase := database.NewEmbeddedDatabase()
	if err := geneDatabase.LoadSpecies("", "mouse"); err != nil {
		t.Fatalf("load embedded mouse mapping: %v", err)
	}

	dataFrame := dataframe.NewDataFrame([]string{"Gene", "sample"})
	dataFrame.AddRow("ENSMUSG00000000001", []float64{17})

	converted, statistics, err := ConvertGeneIDs(dataFrame, geneDatabase, "mouse")
	if err != nil {
		t.Fatalf("ConvertGeneIDs returned an unexpected error: %v", err)
	}
	if converted.RowLabels[0] != "ENSMUSG00000000001" || converted.Data[0][0] != 17 {
		t.Fatalf("mouse Ensembl ID/count was not preserved: %q %v", converted.RowLabels[0], converted.Data[0])
	}
	if statistics.UnmappedCount != 1 || statistics.UnmappedRate != 1 || !statistics.HighUnmappedRate {
		t.Fatalf("unexpected conversion statistics: %+v", statistics)
	}
}

func TestConvertGeneIDsDoesNotFlagAcceptableUnmappedRate(t *testing.T) {
	dataFrame := dataframe.NewDataFrame([]string{"Gene", "sample"})
	mapping := make(map[string]string)
	for geneIndex := 0; geneIndex < 5; geneIndex++ {
		geneID := string(rune('A' + geneIndex))
		dataFrame.AddRow(geneID, []float64{float64(geneIndex + 1)})
		if geneIndex < 4 {
			mapping[geneID] = "symbol_" + geneID
		}
	}

	_, statistics, err := ConvertGeneIDs(dataFrame, &testGeneDatabase{mapping: mapping}, "human")
	if err != nil {
		t.Fatalf("ConvertGeneIDs returned an unexpected error: %v", err)
	}
	if statistics.HighUnmappedRate {
		t.Errorf("an unmapped rate equal to the %.0f%% threshold should not be high", HighUnmappedRateThreshold*100)
	}
}
