package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/NhutHuyDev/sgbank/internal/gapi"
	"github.com/NhutHuyDev/sgbank/internal/infra/db"
	"github.com/NhutHuyDev/sgbank/internal/rest"
	"github.com/NhutHuyDev/sgbank/pb"
	"github.com/NhutHuyDev/sgbank/pkg/utils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config, err := utils.LoadConfig(".", "app")
	if err != nil {
		log.Fatal().Msgf("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db: ", err)
	}

	RunDbMigration(config.MigrationUrl, config.DBSource)

	store := db.NewStore(conn)

	// go RunGateWayServer(config, store)
	go RunHttpServer(config, store)

	RunGrpcServer(config, store)
}

func RunDbMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msgf("cannot create new migrate instance: ", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msgf("faild to run migrate up: ", err)
	}

	log.Info().Msg("db migrated successfully")
}

func RunHttpServer(config utils.Config, store db.Store) {
	server, err := rest.NewServer(config, store)
	if err != nil {
		log.Fatal().Msgf(err.Error())
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msgf("cannot start HTTP server: ", err)
	}
}

func RunGrpcServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	err = server.Start(config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msgf("cannot start gRPC server: ", err)
	}
}

func RunGateWayServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	grpcMux := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pb.RegisterSgbankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msgf("cannot register handler server", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msgf("cannot create listener", err)
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Msgf("cannot start HTTP gateway server", err)
	}
}
