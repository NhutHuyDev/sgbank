package rest

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RenewAccessTokenDTO struct {
	RefreshToken string `json:"access_token" binding:"required"`
}

type RenewAccessTokenRes struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewTokenHandler(ctx *gin.Context) {
	var req RenewAccessTokenDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	refreshPayload, err := server.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	session, err := server.Store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if session.IsBlocked {
		err = fmt.Errorf("blocked session")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	if session.Username != refreshPayload.Username {
		err = fmt.Errorf("incorrect session user")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	if session.RefreshToken != req.RefreshToken {
		err = fmt.Errorf("mismatched session token")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	refreshToken, refreshPayload, err := server.TokenMaker.CreateToken(session.Username, server.Config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	ctx.JSON(http.StatusOK, RenewAccessTokenRes{
		AccessToken:          refreshToken,
		AccessTokenExpiresAt: refreshPayload.ExpiredAt,
	})
}
