package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/davidcarrington/eagle-bank/internal/api/middleware"
	"github.com/davidcarrington/eagle-bank/internal/auth"
	"github.com/davidcarrington/eagle-bank/internal/config"
	"github.com/davidcarrington/eagle-bank/internal/domain"
	"github.com/davidcarrington/eagle-bank/internal/service"
)

type userHandler struct {
	users *service.UserService
	cfg   config.Config
}

// Request types

type createUserRequest struct {
	Name        string         `json:"name" binding:"required"`
	Email       string         `json:"email" binding:"required,email"`
	Password    string         `json:"password" binding:"required,min=8"`
	PhoneNumber string         `json:"phoneNumber" binding:"required"`
	Address     addressRequest `json:"address" binding:"required"`
}

type addressRequest struct {
	Line1    string `json:"line1" binding:"required"`
	Line2    string `json:"line2"`
	Line3    string `json:"line3"`
	Town     string `json:"town" binding:"required"`
	County   string `json:"county" binding:"required"`
	Postcode string `json:"postcode" binding:"required"`
}

type updateUserRequest struct {
	Name        *string         `json:"name"`
	Email       *string         `json:"email" binding:"omitempty,email"`
	PhoneNumber *string         `json:"phoneNumber"`
	Address     *addressRequest `json:"address" binding:"omitempty"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Response types

type addressResponse struct {
	Line1    string `json:"line1"`
	Line2    string `json:"line2,omitempty"`
	Line3    string `json:"line3,omitempty"`
	Town     string `json:"town"`
	County   string `json:"county"`
	Postcode string `json:"postcode"`
}

type userResponse struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	PhoneNumber      string          `json:"phoneNumber"`
	Address          addressResponse `json:"address"`
	CreatedTimestamp string          `json:"createdTimestamp"`
	UpdatedTimestamp string          `json:"updatedTimestamp"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		Address: addressResponse{
			Line1:    u.Address.Line1,
			Line2:    u.Address.Line2,
			Line3:    u.Address.Line3,
			Town:     u.Address.Town,
			County:   u.Address.County,
			Postcode: u.Address.Postcode,
		},
		CreatedTimestamp: u.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedTimestamp: u.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// Handlers

func (h *userHandler) createUser(c *gin.Context) {
	var req createUserRequest
	if !bindJSON(c, &req) {
		return
	}

	u, err := h.users.Create(c.Request.Context(), service.CreateUserInput{
		Name:        req.Name,
		Email:       req.Email,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		Address: domain.Address{
			Line1:    req.Address.Line1,
			Line2:    req.Address.Line2,
			Line3:    req.Address.Line3,
			Town:     req.Address.Town,
			County:   req.Address.County,
			Postcode: req.Address.Postcode,
		},
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toUserResponse(u))
}

func (h *userHandler) login(c *gin.Context) {
	var req loginRequest
	if !bindJSON(c, &req) {
		return
	}

	u, err := h.users.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}

	token, expiresAt, err := auth.SignToken(h.cfg.JWTSecret, u.ID, h.cfg.JWTTTL)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		Token:     token,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	})
}

func (h *userHandler) getUser(c *gin.Context) {
	u, err := h.users.Get(c.Request.Context(), middleware.CallerID(c), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(u))
}

func (h *userHandler) updateUser(c *gin.Context) {
	var req updateUserRequest
	if !bindJSON(c, &req) {
		return
	}

	input := service.UpdateUserInput{
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}
	if req.Address != nil {
		input.Address = &domain.Address{
			Line1:    req.Address.Line1,
			Line2:    req.Address.Line2,
			Line3:    req.Address.Line3,
			Town:     req.Address.Town,
			County:   req.Address.County,
			Postcode: req.Address.Postcode,
		}
	}

	u, err := h.users.Update(c.Request.Context(), middleware.CallerID(c), c.Param("userId"), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(u))
}

func (h *userHandler) deleteUser(c *gin.Context) {
	err := h.users.Delete(c.Request.Context(), middleware.CallerID(c), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
