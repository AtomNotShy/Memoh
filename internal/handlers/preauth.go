package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/memohai/memoh/internal/accounts"
	"github.com/memohai/memoh/internal/auth"
	"github.com/memohai/memoh/internal/bots"
	"github.com/memohai/memoh/internal/identity"
	"github.com/memohai/memoh/internal/preauth"
)

type PreauthHandler struct {
	service        *preauth.Service
	botService     *bots.Service
	accountService *accounts.Service
}

func NewPreauthHandler(service *preauth.Service, botService *bots.Service, accountService *accounts.Service) *PreauthHandler {
	return &PreauthHandler{
		service:        service,
		botService:     botService,
		accountService: accountService,
	}
}

func (h *PreauthHandler) Register(e *echo.Echo) {
	group := e.Group("/bots/:bot_id/preauth_keys")
	group.POST("", h.Issue)
}

type preauthIssueRequest struct {
	TTLSeconds int `json:"ttl_seconds"`
}

func (h *PreauthHandler) Issue(c echo.Context) error {
	channelIdentityID, err := h.requireChannelIdentityID(c)
	if err != nil {
		return err
	}
	botID := strings.TrimSpace(c.Param("bot_id"))
	if botID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bot id is required")
	}
	if _, err := h.authorizeBotAccess(c.Request().Context(), channelIdentityID, botID); err != nil {
		return err
	}
	var req preauthIssueRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ttl := 24 * time.Hour
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}
	key, err := h.service.Issue(c.Request().Context(), botID, channelIdentityID, ttl)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, key)
}

func (h *PreauthHandler) requireChannelIdentityID(c echo.Context) (string, error) {
	channelIdentityID, err := auth.ChannelIdentityIDFromContext(c)
	if err != nil {
		return "", err
	}
	if err := identity.ValidateChannelIdentityID(channelIdentityID); err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return channelIdentityID, nil
}

func (h *PreauthHandler) authorizeBotAccess(ctx context.Context, channelIdentityID, botID string) (bots.Bot, error) {
	if h.botService == nil || h.accountService == nil {
		return bots.Bot{}, echo.NewHTTPError(http.StatusInternalServerError, "bot services not configured")
	}
	isAdmin, err := h.accountService.IsAdmin(ctx, channelIdentityID)
	if err != nil {
		return bots.Bot{}, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	bot, err := h.botService.AuthorizeAccess(ctx, channelIdentityID, botID, isAdmin, bots.AccessPolicy{AllowPublicMember: false})
	if err != nil {
		if errors.Is(err, bots.ErrBotNotFound) {
			return bots.Bot{}, echo.NewHTTPError(http.StatusNotFound, "bot not found")
		}
		if errors.Is(err, bots.ErrBotAccessDenied) {
			return bots.Bot{}, echo.NewHTTPError(http.StatusForbidden, "bot access denied")
		}
		return bots.Bot{}, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return bot, nil
}
