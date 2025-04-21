package rest

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/NhutHuyDev/sgbank/internal/infra/db"
	"github.com/NhutHuyDev/sgbank/internal/token"
	"github.com/gin-gonic/gin"
)

type transferDTO struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) transferHandler(ctx *gin.Context) {
	var req transferDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	fromAccount, valid := server.isValidAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(AuthorizationPayloadKey).(*token.Payload)
	if authPayload.Username != fromAccount.Owner {
		err := errors.New("from account doestn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	_, valid = server.isValidAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	toAccount, err := server.Store.GetAccount(ctx, req.ToAccountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if toAccount.Currency != req.Currency {
		err := fmt.Errorf("currency of from_account and to_account mismatch: %s vs %s", req.Currency, toAccount.Currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.Store.TransferTx(ctx, arg)
	if err != nil {
		fmt.Println(err)
		if err.Error() == "the balance of the from account is insufficient" {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) isValidAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.Store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}
