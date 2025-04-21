package test

import (
	"os"
	"testing"
	"time"

	"github.com/NhutHuyDev/sgbank/internal/infra/db"
	"github.com/NhutHuyDev/sgbank/internal/rest"
	"github.com/NhutHuyDev/sgbank/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store) *rest.Server {
	config := utils.Config{
		TokenSymmetricKey:   utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := rest.NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
