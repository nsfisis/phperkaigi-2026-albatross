package admin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"albatross-2026-backend/account"
	"albatross-2026-backend/config"
	"albatross-2026-backend/db"
	"albatross-2026-backend/game"
	"albatross-2026-backend/session"
	"albatross-2026-backend/tournament"
)

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

type Handler struct {
	gameSvc       *game.Service
	tournamentSvc *tournament.Service
	q             db.Querier
	conf          *config.Config
}

func NewHandler(gameSvc *game.Service, tournamentSvc *tournament.Service, q db.Querier, conf *config.Config) *Handler {
	return &Handler{gameSvc: gameSvc, tournamentSvc: tournamentSvc, q: q, conf: conf}
}

func (h *Handler) newAdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := session.GetUserFromContext(c.Request().Context())
			if !ok {
				return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"login")
			}
			if !user.IsAdmin {
				return echo.NewHTTPError(http.StatusForbidden)
			}
			return next(c)
		}
	}
}

func (h *Handler) RegisterHandlers(g *echo.Group) {
	g.Use(newAssetsMiddleware())
	g.Use(h.newAdminMiddleware())

	g.GET("/dashboard", h.getDashboard)

	g.GET("/online-qualifying-ranking", h.getOnlineQualifyingRanking)
	g.POST("/fix", h.postFix)

	g.GET("/users", h.getUsers)
	g.GET("/users/:userID", h.getUserEdit)
	g.POST("/users/:userID", h.postUserEdit)
	g.POST("/users/:userID/fetch-icon", h.postUserFetchIcon)

	g.GET("/games", h.getGames)
	g.GET("/games/new", h.getGameNew)
	g.POST("/games/new", h.postGameNew)
	g.GET("/games/:gameID", h.getGameEdit)
	g.POST("/games/:gameID", h.postGameEdit)
	g.POST("/games/:gameID/start", h.postGameStart)
	g.GET("/games/:gameID/submissions", h.getSubmissions)
	g.POST("/games/:gameID/submissions/rejudge-latest", h.postSubmissionsRejudgeLatest)
	g.POST("/games/:gameID/submissions/rejudge-all", h.postSubmissionsRejudgeAll)
	g.GET("/games/:gameID/submissions/:submissionID", h.getSubmissionDetail)
	g.POST("/games/:gameID/submissions/:submissionID/rejudge", h.postSubmissionRejudge)

	g.GET("/problems", h.getProblems)
	g.GET("/problems/new", h.getProblemNew)
	g.POST("/problems/new", h.postProblemNew)
	g.GET("/problems/:problemID", h.getProblemEdit)
	g.POST("/problems/:problemID", h.postProblemEdit)
	g.GET("/problems/:problemID/testcases", h.getTestcases)
	g.GET("/problems/:problemID/testcases/new", h.getTestcaseNew)
	g.POST("/problems/:problemID/testcases/new", h.postTestcaseNew)
	g.GET("/problems/:problemID/testcases/:testcaseID", h.getTestcaseEdit)
	g.POST("/problems/:problemID/testcases/:testcaseID", h.postTestcaseEdit)
	g.POST("/problems/:problemID/testcases/:testcaseID/delete", h.postTestcaseDelete)

	g.GET("/tournaments", h.getTournaments)
	g.GET("/tournaments/new", h.getTournamentNew)
	g.POST("/tournaments/new", h.postTournamentNew)
	g.GET("/tournaments/:tournamentID", h.getTournamentEdit)
	g.POST("/tournaments/:tournamentID", h.postTournamentEdit)
}

func (h *Handler) getDashboard(c echo.Context) error {
	return c.Render(http.StatusOK, "dashboard", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Dashboard",
	})
}

func (h *Handler) getOnlineQualifyingRanking(c echo.Context) error {
	game1, err := strconv.Atoi(c.QueryParam("game_1"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game_1")
	}
	game2, err := strconv.Atoi(c.QueryParam("game_2"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game_2")
	}
	game3, err := strconv.Atoi(c.QueryParam("game_3"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game_3")
	}

	rows, err := h.q.GetQualifyingRanking(c.Request().Context(), db.GetQualifyingRankingParams{
		GameID:   int32(game1),
		GameID_2: int32(game2),
		GameID_3: int32(game3),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	entries := make([]echo.Map, len(rows))
	for i, r := range rows {
		entries[i] = echo.Map{
			"Rank":         i + 1,
			"Username":     r.Username,
			"UserLabel":    r.UserLabel,
			"Score1":       r.CodeSize1,
			"Score2":       r.CodeSize2,
			"Score3":       r.CodeSize3,
			"TotalScore":   r.TotalCodeSize,
			"SubmittedAt1": r.SubmittedAt1.Time.In(jst).Format("2006-01-02T15:04"),
			"SubmittedAt2": r.SubmittedAt2.Time.In(jst).Format("2006-01-02T15:04"),
			"SubmittedAt3": r.SubmittedAt3.Time.In(jst).Format("2006-01-02T15:04"),
		}
	}
	return c.Render(http.StatusOK, "online_qualifying_ranking", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Online Qualifying Ranking",
		"Entries":  entries,
	})
}

func (h *Handler) postFix(c echo.Context) error {
	if err := h.gameSvc.FixSubmissionStatuses(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/dashboard")
}

func (h *Handler) getUsers(c echo.Context) error {
	rows, err := h.q.ListUsers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	users := make([]echo.Map, len(rows))
	for i, u := range rows {
		users[i] = echo.Map{
			"UserID":      u.UserID,
			"Username":    u.Username,
			"DisplayName": u.DisplayName,
			"IconPath":    u.IconPath,
			"IsAdmin":     u.IsAdmin,
		}
	}

	return c.Render(http.StatusOK, "users", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Users",
		"Users":    users,
	})
}

func (h *Handler) getUserEdit(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user id")
	}
	row, err := h.q.GetUserByID(c.Request().Context(), int32(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Render(http.StatusOK, "user_edit", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "User Edit",
		"User": echo.Map{
			"UserID":      row.UserID,
			"Username":    row.Username,
			"DisplayName": row.DisplayName,
			"IconPath":    row.IconPath,
			"IsAdmin":     row.IsAdmin,
			"Label":       row.Label,
		},
	})
}

func (h *Handler) postUserEdit(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user_id")
	}

	displayName := c.FormValue("display_name")
	iconPathRaw := c.FormValue("icon_path")
	isAdmin := (c.FormValue("is_admin") != "")
	labelRaw := c.FormValue("label")

	var iconPath *string
	if iconPathRaw != "" {
		iconPath = &iconPathRaw
	}
	var label *string
	if labelRaw != "" {
		label = &labelRaw
	}

	err = h.q.UpdateUser(c.Request().Context(), db.UpdateUserParams{
		UserID:      int32(userID),
		DisplayName: displayName,
		IconPath:    iconPath,
		IsAdmin:     isAdmin,
		Label:       label,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/users")
}

func (h *Handler) postUserFetchIcon(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user id")
	}
	row, err := h.q.GetUserByID(c.Request().Context(), int32(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	go func() {
		err := account.FetchIcon(context.Background(), h.q, int(row.UserID))
		if err != nil {
			slog.Error("failed to fetch icon", "error", err)
			// The failure is intentionally ignored. Retry manually if needed.
		}
	}()
	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/users")
}

func (h *Handler) getGames(c echo.Context) error {
	rows, err := h.q.ListAllGames(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	games := make([]echo.Map, len(rows))
	for i, g := range rows {
		var startedAt string
		if g.StartedAt.Valid {
			startedAt = g.StartedAt.Time.In(jst).Format("2006-01-02T15:04")
		}
		games[i] = echo.Map{
			"GameID":          g.GameID,
			"GameType":        g.GameType,
			"IsPublic":        g.IsPublic,
			"DisplayName":     g.DisplayName,
			"DurationSeconds": g.DurationSeconds,
			"StartedAt":       startedAt,
			"ProblemID":       g.ProblemID,
		}
	}

	return c.Render(http.StatusOK, "games", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Games",
		"Games":    games,
	})
}

func (h *Handler) getGameNew(c echo.Context) error {
	problemRows, err := h.q.ListProblems(c.Request().Context())
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	var problems []echo.Map
	for _, p := range problemRows {
		problems = append(problems, echo.Map{
			"ProblemID": int(p.ProblemID),
			"Title":     p.Title,
		})
	}

	return c.Render(http.StatusOK, "game_new", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "New Game",
		"Problems": problems,
	})
}

func (h *Handler) postGameNew(c echo.Context) error {
	gameType := c.FormValue("game_type")
	isPublic := (c.FormValue("is_public") != "")
	displayName := c.FormValue("display_name")
	durationSeconds, err := strconv.Atoi(c.FormValue("duration_seconds"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid duration_seconds")
	}
	var problemID int
	{
		problemIDRaw := c.FormValue("problem_id")
		problemIDInt, err := strconv.Atoi(problemIDRaw)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem_id")
		}
		problemID = problemIDInt
	}

	_, err = h.q.CreateGame(c.Request().Context(), db.CreateGameParams{
		GameType:        gameType,
		IsPublic:        isPublic,
		DisplayName:     displayName,
		DurationSeconds: int32(durationSeconds),
		ProblemID:       int32(problemID),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/games")
}

func (h *Handler) getGameEdit(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game id")
	}
	row, err := h.q.GetGameByID(c.Request().Context(), int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var startedAt string
	if row.StartedAt.Valid {
		startedAt = row.StartedAt.Time.In(jst).Format("2006-01-02T15:04")
	}

	mainPlayerRows, err := h.q.ListMainPlayers(c.Request().Context(), []int32{int32(gameID)})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	mainPlayer1 := 0
	if len(mainPlayerRows) > 0 {
		mainPlayer1 = int(mainPlayerRows[0].UserID)
	}
	mainPlayer2 := 0
	if len(mainPlayerRows) > 1 {
		mainPlayer2 = int(mainPlayerRows[1].UserID)
	}

	problemRows, err := h.q.ListProblems(c.Request().Context())
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	var problems []echo.Map
	for _, p := range problemRows {
		problems = append(problems, echo.Map{
			"ProblemID": int(p.ProblemID),
			"Title":     p.Title,
		})
	}

	userRows, err := h.q.ListUsers(c.Request().Context())
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	var users []echo.Map
	for _, r := range userRows {
		users = append(users, echo.Map{
			"UserID":   int(r.UserID),
			"Username": r.Username,
		})
	}

	return c.Render(http.StatusOK, "game_edit", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Game Edit",
		"Game": echo.Map{
			"GameID":          row.GameID,
			"GameType":        row.GameType,
			"IsPublic":        row.IsPublic,
			"DisplayName":     row.DisplayName,
			"DurationSeconds": row.DurationSeconds,
			"StartedAt":       startedAt,
			"ProblemID":       row.ProblemID,
			"MainPlayer1":     mainPlayer1,
			"MainPlayer2":     mainPlayer2,
		},
		"Problems": problems,
		"Users":    users,
	})
}

func (h *Handler) postGameEdit(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game id")
	}

	gameType := c.FormValue("game_type")
	isPublic := (c.FormValue("is_public") != "")
	displayName := c.FormValue("display_name")
	durationSeconds, err := strconv.Atoi(c.FormValue("duration_seconds"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid duration_seconds")
	}
	var problemID int
	{
		problemIDRaw := c.FormValue("problem_id")
		problemIDInt, err := strconv.Atoi(problemIDRaw)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem_id")
		}
		problemID = problemIDInt
	}
	var startedAt *time.Time
	{
		startedAtRaw := c.FormValue("started_at")
		if startedAtRaw != "" {
			startedAtJST, err := time.ParseInLocation("2006-01-02T15:04", startedAtRaw, jst)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid started_at")
			}
			startedAtUTC := startedAtJST.UTC()
			startedAt = &startedAtUTC
		}
	}

	var changedStartedAt pgtype.Timestamp
	if startedAt == nil {
		changedStartedAt = pgtype.Timestamp{
			Valid: false,
		}
	} else {
		changedStartedAt = pgtype.Timestamp{
			Time:  *startedAt,
			Valid: true,
		}
	}

	mainPlayers := []int{}
	mainPlayer1Raw := c.FormValue("main_player_1")
	if mainPlayer1Raw != "" && mainPlayer1Raw != "0" {
		mainPlayer1, err := strconv.Atoi(mainPlayer1Raw)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid main_player_1")
		}
		mainPlayers = append(mainPlayers, mainPlayer1)
	}
	mainPlayer2Raw := c.FormValue("main_player_2")
	if mainPlayer2Raw != "" && mainPlayer2Raw != "0" {
		mainPlayer2, err := strconv.Atoi(mainPlayer2Raw)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid main_player_2")
		}
		mainPlayers = append(mainPlayers, mainPlayer2)
	}

	err = h.gameSvc.UpdateGameWithPlayers(c.Request().Context(), game.UpdateGameParams{
		GameID:          gameID,
		GameType:        gameType,
		IsPublic:        isPublic,
		DisplayName:     displayName,
		DurationSeconds: durationSeconds,
		StartedAt:       changedStartedAt,
		ProblemID:       problemID,
		MainPlayerIDs:   mainPlayers,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/games")
}

func (h *Handler) postGameStart(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game id")
	}

	err = h.gameSvc.StartGame(c.Request().Context(), gameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Game not found")
		}
		if errors.Is(err, game.ErrNoTestcases) {
			return echo.NewHTTPError(http.StatusBadRequest, "No testcases")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/games")
}

func (h *Handler) getSubmissions(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game_id")
	}

	submissions, err := h.q.GetSubmissionsByGameID(c.Request().Context(), int32(gameID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	entries := make([]echo.Map, len(submissions))
	for i, r := range submissions {
		entries[i] = echo.Map{
			"SubmissionID": r.SubmissionID,
			"UserID":       r.UserID,
			"Status":       r.Status,
			"CodeSize":     r.CodeSize,
			"CreatedAt":    r.CreatedAt.Time.In(jst).Format("2006-01-02T15:04"),
		}
	}

	return c.Render(http.StatusOK, "submissions", echo.Map{
		"BasePath":    h.conf.BasePath,
		"Title":       "Submissions",
		"GameID":      gameID,
		"Submissions": entries,
	})
}

func (h *Handler) getSubmissionDetail(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game_id")
	}

	submissionID, err := strconv.Atoi(c.Param("submissionID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid submission_id")
	}

	submission, err := h.q.GetSubmissionByID(c.Request().Context(), int32(submissionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	testcaseResultRows, err := h.q.GetTestcaseResultsBySubmissionID(c.Request().Context(), int32(submissionID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	testcaseResults := make([]echo.Map, len(testcaseResultRows))
	for i, r := range testcaseResultRows {
		testcaseResults[i] = echo.Map{
			"TestcaseResultID": r.TestcaseResultID,
			"TestcaseID":       r.TestcaseID,
			"Status":           r.Status,
			"CreatedAt":        r.CreatedAt.Time.In(jst).Format("2006-01-02T15:04"),
			"Stdout":           r.Stdout,
			"Stderr":           r.Stderr,
		}
	}

	return c.Render(http.StatusOK, "submission_detail", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Submission Detail",
		"GameID":   gameID,
		"Submission": echo.Map{
			"SubmissionID": submission.SubmissionID,
			"UserID":       submission.UserID,
			"Status":       submission.Status,
			"CodeSize":     submission.CodeSize,
			"CreatedAt":    submission.CreatedAt.Time.In(jst).Format("2006-01-02T15:04"),
			"Code":         submission.Code,
		},
		"TestcaseResults": testcaseResults,
	})
}

func (h *Handler) postSubmissionRejudge(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game_id")
	}

	submissionID, err := strconv.Atoi(c.Param("submissionID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid submission_id")
	}

	ctx := c.Request().Context()

	submission, err := h.q.GetSubmissionByID(ctx, int32(submissionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	g, err := h.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := h.gameSvc.RejudgeSubmission(ctx, submission.SubmissionID, int(submission.GameID), int(submission.UserID), g.Language, submission.Code); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%sadmin/games/%d/submissions/%d", h.conf.BasePath, gameID, submissionID))
}

func (h *Handler) postSubmissionsRejudgeLatest(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game_id")
	}

	err = h.gameSvc.RejudgeLatestSubmissionsByGame(c.Request().Context(), gameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%sadmin/games/%d/submissions", h.conf.BasePath, gameID))
}

func (h *Handler) postSubmissionsRejudgeAll(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game_id")
	}

	err = h.gameSvc.RejudgeAllSubmissionsByGame(c.Request().Context(), gameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%sadmin/games/%d/submissions", h.conf.BasePath, gameID))
}

func (h *Handler) getProblems(c echo.Context) error {
	rows, err := h.q.ListProblems(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	problems := make([]echo.Map, len(rows))
	for i, p := range rows {
		problems[i] = echo.Map{
			"ProblemID":   p.ProblemID,
			"Title":       p.Title,
			"Description": p.Description,
			"Language":    p.Language,
		}
	}

	return c.Render(http.StatusOK, "problems", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Problems",
		"Problems": problems,
	})
}

func (h *Handler) getProblemNew(c echo.Context) error {
	return c.Render(http.StatusOK, "problem_new", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "New Problem",
	})
}

func (h *Handler) postProblemNew(c echo.Context) error {
	title := c.FormValue("title")
	description := c.FormValue("description")
	language := c.FormValue("language")
	sampleCode := c.FormValue("sample_code")

	_, err := h.q.CreateProblem(c.Request().Context(), db.CreateProblemParams{
		Title:       title,
		Description: description,
		Language:    language,
		SampleCode:  sampleCode,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/problems")
}

func (h *Handler) getProblemEdit(c echo.Context) error {
	problemID, err := strconv.Atoi(c.Param("problemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem id")
	}
	row, err := h.q.GetProblemByID(c.Request().Context(), int32(problemID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Render(http.StatusOK, "problem_edit", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Problem Edit",
		"Problem": echo.Map{
			"ProblemID":   row.ProblemID,
			"Title":       row.Title,
			"Description": row.Description,
			"Language":    row.Language,
			"SampleCode":  row.SampleCode,
		},
	})
}

func (h *Handler) postProblemEdit(c echo.Context) error {
	problemID, err := strconv.Atoi(c.Param("problemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem_id")
	}

	title := c.FormValue("title")
	description := c.FormValue("description")
	language := c.FormValue("language")
	sampleCode := c.FormValue("sample_code")

	err = h.q.UpdateProblem(c.Request().Context(), db.UpdateProblemParams{
		ProblemID:   int32(problemID),
		Title:       title,
		Description: description,
		Language:    language,
		SampleCode:  sampleCode,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/problems")
}

func (h *Handler) getTestcases(c echo.Context) error {
	problemID, err := strconv.Atoi(c.Param("problemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem id")
	}

	// Get problem info
	problem, err := h.q.GetProblemByID(c.Request().Context(), int32(problemID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Get testcases for this problem
	rows, err := h.q.ListTestcasesByProblemID(c.Request().Context(), int32(problemID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	testcases := make([]echo.Map, len(rows))
	for i, tc := range rows {
		testcases[i] = echo.Map{
			"TestcaseID": tc.TestcaseID,
			"ProblemID":  tc.ProblemID,
			"Stdin":      tc.Stdin,
			"Stdout":     tc.Stdout,
		}
	}

	return c.Render(http.StatusOK, "testcases", echo.Map{
		"BasePath":  h.conf.BasePath,
		"Title":     "Testcases for " + problem.Title,
		"Problem":   echo.Map{"ProblemID": problem.ProblemID, "Title": problem.Title},
		"Testcases": testcases,
	})
}

func (h *Handler) getTestcaseNew(c echo.Context) error {
	problemID, err := strconv.Atoi(c.Param("problemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem id")
	}

	// Get problem info
	problem, err := h.q.GetProblemByID(c.Request().Context(), int32(problemID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Render(http.StatusOK, "testcase_new", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "New Testcase for " + problem.Title,
		"Problem":  echo.Map{"ProblemID": problem.ProblemID, "Title": problem.Title},
	})
}

func (h *Handler) postTestcaseNew(c echo.Context) error {
	problemID, err := strconv.Atoi(c.Param("problemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem_id")
	}
	stdin := c.FormValue("stdin")
	stdout := c.FormValue("stdout")

	_, err = h.q.CreateTestcase(c.Request().Context(), db.CreateTestcaseParams{
		ProblemID: int32(problemID),
		Stdin:     stdin,
		Stdout:    stdout,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/problems/"+strconv.Itoa(problemID)+"/testcases")
}

func (h *Handler) getTestcaseEdit(c echo.Context) error {
	problemID, err := strconv.Atoi(c.Param("problemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem id")
	}
	testcaseID, err := strconv.Atoi(c.Param("testcaseID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid testcase id")
	}

	// Get problem info
	problem, err := h.q.GetProblemByID(c.Request().Context(), int32(problemID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Get testcase info and verify it belongs to this problem
	testcase, err := h.q.GetTestcaseByID(c.Request().Context(), int32(testcaseID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Verify the testcase belongs to the specified problem
	if testcase.ProblemID != int32(problemID) {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return c.Render(http.StatusOK, "testcase_edit", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Edit Testcase for " + problem.Title,
		"Problem":  echo.Map{"ProblemID": problem.ProblemID, "Title": problem.Title},
		"Testcase": echo.Map{
			"TestcaseID": testcase.TestcaseID,
			"ProblemID":  testcase.ProblemID,
			"Stdin":      testcase.Stdin,
			"Stdout":     testcase.Stdout,
		},
	})
}

func (h *Handler) postTestcaseEdit(c echo.Context) error {
	problemID, err := strconv.Atoi(c.Param("problemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem_id")
	}
	testcaseID, err := strconv.Atoi(c.Param("testcaseID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid testcase_id")
	}

	// Verify the testcase belongs to this problem before updating
	testcase, err := h.q.GetTestcaseByID(c.Request().Context(), int32(testcaseID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if testcase.ProblemID != int32(problemID) {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	stdin := c.FormValue("stdin")
	stdout := c.FormValue("stdout")

	err = h.q.UpdateTestcase(c.Request().Context(), db.UpdateTestcaseParams{
		TestcaseID: int32(testcaseID),
		ProblemID:  int32(problemID),
		Stdin:      stdin,
		Stdout:     stdout,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/problems/"+strconv.Itoa(problemID)+"/testcases")
}

func (h *Handler) postTestcaseDelete(c echo.Context) error {
	problemID, err := strconv.Atoi(c.Param("problemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid problem_id")
	}
	testcaseID, err := strconv.Atoi(c.Param("testcaseID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid testcase_id")
	}

	// Verify the testcase belongs to this problem before deleting
	testcase, err := h.q.GetTestcaseByID(c.Request().Context(), int32(testcaseID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if testcase.ProblemID != int32(problemID) {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	err = h.q.DeleteTestcase(c.Request().Context(), int32(testcaseID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/problems/"+strconv.Itoa(problemID)+"/testcases")
}

func (h *Handler) getTournaments(c echo.Context) error {
	rows, err := h.q.ListTournaments(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	tournaments := make([]echo.Map, len(rows))
	for i, t := range rows {
		tournaments[i] = echo.Map{
			"TournamentID": t.TournamentID,
			"DisplayName":  t.DisplayName,
			"BracketSize":  t.BracketSize,
		}
	}

	return c.Render(http.StatusOK, "tournaments", echo.Map{
		"BasePath":    h.conf.BasePath,
		"Title":       "Tournaments",
		"Tournaments": tournaments,
	})
}

func (h *Handler) getTournamentNew(c echo.Context) error {
	return c.Render(http.StatusOK, "tournament_new", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "New Tournament",
	})
}

func (h *Handler) postTournamentNew(c echo.Context) error {
	displayName := c.FormValue("display_name")
	numParticipants, err := strconv.Atoi(c.FormValue("num_participants"))
	if err != nil || numParticipants < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid num_participants")
	}

	_, err = h.tournamentSvc.CreateTournament(c.Request().Context(), displayName, numParticipants)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/tournaments")
}

func (h *Handler) getTournamentEdit(c echo.Context) error {
	tournamentID, err := strconv.Atoi(c.Param("tournamentID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tournament id")
	}
	ctx := c.Request().Context()

	editData, err := h.tournamentSvc.GetTournamentEditData(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	t := editData.Tournament

	matchGameMap := make(map[int]int)
	var matches []echo.Map
	for _, m := range editData.Matches {
		gameID := 0
		if m.GameID != nil {
			gameID = int(*m.GameID)
		}
		matchGameMap[int(m.TournamentMatchID)] = gameID
		matches = append(matches, echo.Map{
			"MatchID":     int(m.TournamentMatchID),
			"Round":       int(m.Round),
			"Position":    int(m.Position),
			"Description": "R" + strconv.Itoa(int(m.Round)) + "P" + strconv.Itoa(int(m.Position)),
			"IsBye":       false,
		})
	}

	userRows, err := h.q.ListUsers(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	var users []echo.Map
	for _, r := range userRows {
		users = append(users, echo.Map{
			"UserID":   int(r.UserID),
			"Username": r.Username,
		})
	}

	gameRows, err := h.q.ListAllGames(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	var games []echo.Map
	for _, g := range gameRows {
		games = append(games, echo.Map{
			"GameID":      int(g.GameID),
			"DisplayName": g.DisplayName,
		})
	}

	seeds := make([]int, t.BracketSize)
	for i := range seeds {
		seeds[i] = i + 1
	}

	return c.Render(http.StatusOK, "tournament_edit", echo.Map{
		"BasePath": h.conf.BasePath,
		"Title":    "Tournament Edit",
		"Tournament": echo.Map{
			"TournamentID": t.TournamentID,
			"DisplayName":  t.DisplayName,
			"BracketSize":  t.BracketSize,
			"NumRounds":    t.NumRounds,
		},
		"Seeds":        seeds,
		"SeedUserMap":  editData.SeedUserMap,
		"Matches":      matches,
		"MatchGameMap": matchGameMap,
		"Users":        users,
		"Games":        games,
	})
}

func (h *Handler) postTournamentEdit(c echo.Context) error {
	tournamentID, err := strconv.Atoi(c.Param("tournamentID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid tournament id")
	}
	ctx := c.Request().Context()

	// We need the tournament to know bracket size for seed parsing,
	// and matches for match game parsing.
	editData, err := h.tournamentSvc.GetTournamentEditData(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	displayName := c.FormValue("display_name")

	// Parse seed → user assignments
	var seedEntries []tournament.SeedEntry
	for seed := 1; seed <= int(editData.Tournament.BracketSize); seed++ {
		raw := c.FormValue("seed_" + strconv.Itoa(seed))
		if raw == "" || raw == "0" {
			continue
		}
		userID, err := strconv.Atoi(raw)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid seed_"+strconv.Itoa(seed))
		}
		seedEntries = append(seedEntries, tournament.SeedEntry{Seed: seed, UserID: userID})
	}

	// Parse match → game assignments
	var matchGames []tournament.MatchGame
	for _, m := range editData.Matches {
		raw := c.FormValue("match_" + strconv.Itoa(int(m.TournamentMatchID)))
		var gameID *int32
		if raw != "" && raw != "0" {
			gid, err := strconv.Atoi(raw)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid match game")
			}
			gid32 := int32(gid)
			gameID = &gid32
		}
		matchGames = append(matchGames, tournament.MatchGame{MatchID: int(m.TournamentMatchID), GameID: gameID})
	}

	err = h.tournamentSvc.UpdateTournament(ctx, tournament.UpdateTournamentParams{
		TournamentID: tournamentID,
		DisplayName:  displayName,
		SeedEntries:  seedEntries,
		MatchGames:   matchGames,
	})
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, h.conf.BasePath+"admin/tournaments")
}
