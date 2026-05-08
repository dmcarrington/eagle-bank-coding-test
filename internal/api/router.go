package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/davidcarrington/eagle-bank/internal/api/middleware"
	"github.com/davidcarrington/eagle-bank/internal/config"
	"github.com/davidcarrington/eagle-bank/internal/service"
	"github.com/davidcarrington/eagle-bank/internal/store"
)

type Deps struct {
	Store  *store.Store
	Config config.Config
}

func NewRouter(deps Deps) *gin.Engine {
	userStore := store.NewUserStore(deps.Store.DB)
	accountStore := store.NewAccountStore(deps.Store.DB)

	txnStore := store.NewTransactionStore(deps.Store.DB)

	users := service.NewUserService(userStore, accountStore)
	accounts := service.NewAccountService(accountStore)
	txns := service.NewTransactionService(accountStore, txnStore)

	uh := &userHandler{users: users, cfg: deps.Config}
	ah := &accountHandler{accounts: accounts}
	th := &transactionHandler{txns: txns}

	r := gin.New()
	// Trust the Docker network (host: 172.17.0.1) and localhost for real client IPs. In production, this would be the load balancer's subnet.
	err := r.SetTrustedProxies([]string{"172.17.0.0/16", "127.0.0.1"})
	if err != nil {
		panic(err)
	}
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Open routes
	r.POST("/v1/users", uh.createUser)
	r.POST("/v1/auth/login", uh.login)

	// Authenticated routes
	authed := r.Group("/v1", middleware.RequireAuth(deps.Config.JWTSecret))
	authed.GET("/users/:userId", uh.getUser)
	authed.PATCH("/users/:userId", uh.updateUser)
	authed.DELETE("/users/:userId", uh.deleteUser)

	authed.POST("/accounts", ah.createAccount)
	authed.GET("/accounts", ah.listAccounts)
	authed.GET("/accounts/:accountNumber", ah.getAccount)
	authed.PATCH("/accounts/:accountNumber", ah.updateAccount)
	authed.DELETE("/accounts/:accountNumber", ah.deleteAccount)

	authed.POST("/accounts/:accountNumber/transactions", th.createTransaction)
	authed.GET("/accounts/:accountNumber/transactions", th.listTransactions)
	authed.GET("/accounts/:accountNumber/transactions/:transactionId", th.getTransaction)

	return r
}
