package test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/NhutHuyDev/sgbank/internal/infra/db"
	"github.com/NhutHuyDev/sgbank/pkg/utils"
	_ "github.com/lib/pq"
)

var testQueries *db.Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := utils.LoadConfig("../../../../", "app")
	if err != nil {
		log.Fatal("cannot load config:", err)

	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	testQueries = db.New(testDB)

	os.Exit(m.Run())
}
