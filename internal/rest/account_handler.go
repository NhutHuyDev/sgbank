package rest

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/NhutHuyDev/sgbank/internal/infra/db"
	"github.com/NhutHuyDev/sgbank/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type CreateAccountDTO struct {
	Currency string `json:"currency" binding:"required,oneof=USD EUR CAD"`
}

func (server *Server) createAccountHandler(ctx *gin.Context) {
	var req CreateAccountDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(AuthorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.Store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"account": account,
	})
}

type GetAccountDTO struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type GetAccountRes struct {
	Account db.Account `json:"account"`
}

func (server *Server) getAccountHandler(ctx *gin.Context) {
	var req GetAccountDTO
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.Store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(AuthorizationPayloadKey).(*token.Payload)
	if authPayload.Username != account.Owner {
		err := errors.New("account doestn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, GetAccountRes{
		Account: account,
	})
}

type listAccountsDTO struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccountsHandler(ctx *gin.Context) {
	var req listAccountsDTO
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(AuthorizationPayloadKey).(*token.Payload)
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.Store.ListAccounts(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}
