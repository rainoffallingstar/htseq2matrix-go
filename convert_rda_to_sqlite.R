#!/usr/bin/env Rscript

# Convert R .rda gene database files to SQLite format
# This script is run once to prepare the database for the Go application

library(RSQLite)

# Get the directory where this script is located
# This allows the script to find the data files regardless of where it's run from
args <- commandArgs(trailingOnly = FALSE)
script_path <- sub("--file=", "", args[grep("--file=", args)])
script_dir <- dirname(script_path)

# Paths to the R data files - relative to script directory
rda_human <- file.path(script_dir, "data", "Entrez_Gene_Id_db.rda")
rda_mouse <- file.path(script_dir, "data", "Entrez_Gene_Id_db_mmu.rda")

# Output SQLite database - also relative to script directory
db_path <- file.path(script_dir, "data", "gene_mappings.sqlite")

cat("Script directory:", script_dir, "\n")
cat("Human database:", rda_human, "\n")
cat("Mouse database:", rda_mouse, "\n")
cat("Output database:", db_path, "\n\n")

# Check if R files exist
if (!file.exists(rda_human)) {
  stop(paste("Human database file not found:", rda_human))
}
if (!file.exists(rda_mouse)) {
  stop(paste("Mouse database file not found:", rda_mouse))
}

# Create SQLite database
con <- dbConnect(RSQLite::SQLite(), db_path)

# Convert human database
cat("Converting human gene database...\n")
load(rda_human)
human_mapping <- data.frame(
  gene_id = Entrez_Gene_Id_db$ENSEMBL,
  symbol = Entrez_Gene_Id_db$SYMBOL,
  stringsAsFactors = FALSE
)
dbWriteTable(con, "human_ensembl_to_symbol", human_mapping, row.names = FALSE)
cat(sprintf("  Converted %d human gene mappings\n", nrow(human_mapping)))

# Convert mouse database
cat("Converting mouse gene database...\n")
load(rda_mouse)
mouse_mapping <- data.frame(
  gene_id = Entrez_Gene_Id_db_mmu$UNIPROT,
  symbol = Entrez_Gene_Id_db_mmu$SYMBOL,
  stringsAsFactors = FALSE
)
dbWriteTable(con, "mouse_uniprot_to_symbol", mouse_mapping, row.names = FALSE)
cat(sprintf("  Converted %d mouse gene mappings\n", nrow(mouse_mapping)))

# Clean up
dbDisconnect(con)

cat("\nDatabase conversion complete!\n")
cat(sprintf("SQLite database created at: %s\n", db_path))
