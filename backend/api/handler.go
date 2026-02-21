package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"albatross-2026-backend/auth"
	"albatross-2026-backend/config"
	"albatross-2026-backend/db"
	"albatross-2026-backend/game"
	"albatross-2026-backend/session"
	"albatross-2026-backend/tournament"
)

type AuthenticatorInterface interface {
	Login(ctx context.Context, username, password string) (int, error)
}

type Handler struct {
	gameSvc       *game.Service
	tournamentSvc *tournament.Service
	auth          AuthenticatorInterface
	conf          *config.Config
	q             db.Querier // for session management (login/logout)
}

type postLoginCookieResponse struct {
	cookie http.Cookie
	body   PostLogin200JSONResponse
}

func (r postLoginCookieResponse) VisitPostLoginResponse(w http.ResponseWriter) error {
	http.SetCookie(w, &r.cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	return json.NewEncoder(w).Encode(r.body)
}

func (h *Handler) PostLogin(ctx context.Context, request PostLoginRequestObject) (PostLoginResponseObject, error) {
	username := request.Body.Username
	password := request.Body.Password
	ip := session.GetClientIPFromContext(ctx)

	userID, err := h.auth.Login(ctx, username, password)
	if err != nil {
		slog.Warn("login failed", "username", username, "ip", ip, "reason", err.Error())
		var msg string
		if errors.Is(err, auth.ErrForteeLoginTimeout) {
			msg = "ログインに失敗しました"
		} else {
			msg = "ユーザー名またはパスワードが誤っています"
		}
		return PostLogin401JSONResponse{
			Message: msg,
		}, nil
	}

	dbUser, err := h.q.GetUserByID(ctx, int32(userID))
	if err != nil {
		return PostLogin401JSONResponse{
			Message: "ログインに失敗しました",
		}, nil
	}

	sessionID, err := auth.GenerateSessionID()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	hashedID := auth.HashSessionID(sessionID)
	expiresAt := pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true}
	if err := h.q.CreateSession(ctx, db.CreateSessionParams{
		SessionID: hashedID,
		UserID:    dbUser.UserID,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	slog.Info("login succeeded", "username", username, "user_id", dbUser.UserID, "ip", ip)

	return postLoginCookieResponse{
		cookie: http.Cookie{
			Name:     "albatross_session",
			Value:    sessionID,
			Path:     h.conf.BasePath,
			MaxAge:   86400,
			HttpOnly: true,
			Secure:   !h.conf.IsLocal,
			SameSite: http.SameSiteLaxMode,
		},
		body: PostLogin200JSONResponse{
			User: User{
				UserID:      int(dbUser.UserID),
				Username:    dbUser.Username,
				DisplayName: dbUser.DisplayName,
				IconPath:    dbUser.IconPath,
				IsAdmin:     dbUser.IsAdmin,
				Label:       toNullable(dbUser.Label),
			},
		},
	}, nil
}

func (h *Handler) GetMe(_ context.Context, _ GetMeRequestObject, user *db.User) (GetMeResponseObject, error) {
	return GetMe200JSONResponse{
		User: User{
			UserID:      int(user.UserID),
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IconPath:    user.IconPath,
			IsAdmin:     user.IsAdmin,
			Label:       toNullable(user.Label),
		},
	}, nil
}

type postLogoutCookieResponse struct {
	cookie http.Cookie
}

func (r postLogoutCookieResponse) VisitPostLogoutResponse(w http.ResponseWriter) error {
	http.SetCookie(w, &r.cookie)
	w.WriteHeader(200)
	return nil
}

func (h *Handler) PostLogout(ctx context.Context, _ PostLogoutRequestObject, _ *db.User) (PostLogoutResponseObject, error) {
	if sessionID, ok := session.GetSessionIDFromContext(ctx); ok {
		_ = h.q.DeleteSession(ctx, sessionID)
	}
	return postLogoutCookieResponse{
		cookie: http.Cookie{
			Name:     "albatross_session",
			Value:    "",
			Path:     h.conf.BasePath,
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   !h.conf.IsLocal,
			SameSite: http.SameSiteLaxMode,
		},
	}, nil
}

func (h *Handler) GetGames(ctx context.Context, _ GetGamesRequestObject, _ *db.User) (GetGamesResponseObject, error) {
	games, err := h.gameSvc.ListPublicGames(ctx)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	apiGames := make([]Game, len(games))
	for i, g := range games {
		apiGames[i] = toAPIGame(g)
	}
	return GetGames200JSONResponse{Games: apiGames}, nil
}

func (h *Handler) GetGame(ctx context.Context, request GetGameRequestObject, user *db.User) (GetGameResponseObject, error) {
	isAdmin := user != nil && user.IsAdmin
	g, err := h.gameSvc.GetGameByID(ctx, request.GameID, isAdmin)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return GetGame404JSONResponse{Message: "Game not found"}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return GetGame200JSONResponse{Game: toAPIGame(g)}, nil
}

func (h *Handler) GetGamePlayLatestState(ctx context.Context, request GetGamePlayLatestStateRequestObject, user *db.User) (GetGamePlayLatestStateResponseObject, error) {
	state, err := h.gameSvc.GetLatestState(ctx, request.GameID, user.UserID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return GetGamePlayLatestState200JSONResponse{State: toAPILatestState(state)}, nil
}

func (h *Handler) GetGameWatchLatestStates(ctx context.Context, request GetGameWatchLatestStatesRequestObject, user *db.User) (GetGameWatchLatestStatesResponseObject, error) {
	var userID *int32
	var isAdmin bool
	if user != nil {
		userID = &user.UserID
		isAdmin = user.IsAdmin
	}
	stateMap, err := h.gameSvc.GetWatchLatestStates(ctx, request.GameID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, game.ErrForbidden) {
			return GetGameWatchLatestStates403JSONResponse{
				Message: "You are one of the main players of this game",
			}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	states := make(map[string]LatestGameState, len(stateMap))
	for uid, s := range stateMap {
		states[strconv.Itoa(uid)] = toAPILatestState(s)
	}
	return GetGameWatchLatestStates200JSONResponse{States: states}, nil
}

func (h *Handler) GetGameWatchRanking(ctx context.Context, request GetGameWatchRankingRequestObject, _ *db.User) (GetGameWatchRankingResponseObject, error) {
	ranking, _, err := h.gameSvc.GetRanking(ctx, request.GameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return GetGameWatchRanking404JSONResponse{Message: "Game not found"}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	apiRanking := make([]RankingEntry, len(ranking))
	for i, r := range ranking {
		apiRanking[i] = toAPIRankingEntry(r)
	}
	return GetGameWatchRanking200JSONResponse{Ranking: apiRanking}, nil
}

func (h *Handler) GetGamePlaySubmissions(ctx context.Context, request GetGamePlaySubmissionsRequestObject, user *db.User) (GetGamePlaySubmissionsResponseObject, error) {
	submissions, err := h.gameSvc.GetSubmissions(ctx, request.GameID, user.UserID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return GetGamePlaySubmissions404JSONResponse{Message: "Game not found"}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	apiSubmissions := make([]Submission, len(submissions))
	for i, s := range submissions {
		apiSubmissions[i] = toAPISubmission(s)
	}
	return GetGamePlaySubmissions200JSONResponse{Submissions: apiSubmissions}, nil
}

func (h *Handler) PostGamePlayCode(ctx context.Context, request PostGamePlayCodeRequestObject, user *db.User) (PostGamePlayCodeResponseObject, error) {
	err := h.gameSvc.SaveCode(ctx, request.GameID, user.UserID, request.Body.Code)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return PostGamePlayCode404JSONResponse{Message: "Game not found"}, nil
		}
		if errors.Is(err, game.ErrGameNotRunning) {
			return PostGamePlayCode403JSONResponse{Message: "Game is not running"}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return PostGamePlayCode200Response{}, nil
}

func (h *Handler) PostGamePlaySubmit(ctx context.Context, request PostGamePlaySubmitRequestObject, user *db.User) (PostGamePlaySubmitResponseObject, error) {
	err := h.gameSvc.SubmitCode(ctx, request.GameID, user.UserID, request.Body.Code)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return PostGamePlaySubmit404JSONResponse{}, nil
		}
		if errors.Is(err, game.ErrGameNotRunning) {
			return PostGamePlaySubmit403JSONResponse{Message: "Game is not running"}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return PostGamePlaySubmit200Response{}, nil
}

func (h *Handler) GetTournament(ctx context.Context, request GetTournamentRequestObject, _ *db.User) (GetTournamentResponseObject, error) {
	t, err := h.tournamentSvc.GetTournament(ctx, request.TournamentID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return GetTournament404JSONResponse{Message: "Tournament not found"}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return GetTournament200JSONResponse{Tournament: toAPITournament(t)}, nil
}
