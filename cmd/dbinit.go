package cmd

import (
	"database/sql"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "github.com/mattn/go-sqlite3"

	"pinjol/internal"
)

var dbInitCmd = &cobra.Command{
	Use:   "db-init",
	Short: "Initialize the database",
	Long:  `Initialize the database schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		runDBInit()
	},
}

func init() {
	dbInitCmd.Flags().StringP("db-path", "d", "./pinjol.db", "Database path")

	viper.BindPFlags(dbInitCmd.Flags())
}

func runDBInit() {
	dbPath := viper.GetString("db-path")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := internal.InitDatabase(db); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	log.Printf("database initialized successfully at %s", dbPath)
}
