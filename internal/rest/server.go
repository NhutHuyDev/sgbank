package rest

import (
	"fmt"
	"net/http"

	"github.com/NhutHuyDev/sgbank/internal/infra/db"
	"github.com/NhutHuyDev/sgbank/internal/token"
	"github.com/NhutHuyDev/sgbank/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	Config     utils.Config
	Store      db.Store
	TokenMaker token.Maker
	Router     *gin.Engine
}

func NewServer(config utils.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		Config:     config,
		Store:      store,
		TokenMaker: tokenMaker,
	}

	router := gin.Default()

	_ = router.SetTrustedProxies([]string{"192.168.1.1"})

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", currencyValidator)
	}

	router.GET("/v1/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "OKE",
		})
	})
	router.POST("/v1/users", server.createUserHandler)
	router.POST("/v1/users/sign-in", server.signInHandler)

	authRoutes := router.Group("/").Use(AuthMiddleware(server.TokenMaker))

	authRoutes.GET("/v1/accounts", server.listAccountsHandler)
	authRoutes.GET("/v1/accounts/:id", server.getAccountHandler)
	authRoutes.POST("/v1/accounts", server.createAccountHandler)

	authRoutes.POST("/v1/transfers", server.transferHandler)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
	})

	server.Router = router

	return server, nil
}

func (server *Server) StartServer(address string) error {
	return server.Router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
