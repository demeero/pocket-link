package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func New(linkClient linkpb.LinkServiceClient) http.Handler {
	e := echo.New()
	middlewares(e)
	e.Any("/*", redirect(linkClient))
	return e
}

func middlewares(e *echo.Echo) {
	e.Pre(middleware.AddTrailingSlash())
}

func redirect(linkClient linkpb.LinkServiceClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		shortened := strings.Trim(c.Request().URL.Path, "/")
		resp, err := linkClient.GetLink(c.Request().Context(), &linkpb.GetLinkRequest{Shortened: shortened})
		if err != nil {
			zap.L().Error("error get link for path", zap.Error(err), zap.String("path", shortened))
			return err
		}
		u, err := url.Parse(resp.GetLink().GetOriginal())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("invalid original URL: %w", err))
		}
		if u.Scheme == "" {
			u.Scheme = "http"
		}
		return c.Redirect(http.StatusFound, u.String())
	}
}
