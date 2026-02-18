package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/nullable"

	"albatross-2026-backend/auth"
	"albatross-2026-backend/config"
	"albatross-2026-backend/db"
)

type AuthenticatorInterface interface {
	Login(ctx context.Context, username, password string) (int, error)
}

type Handler struct {
	q    db.Querier
	txm  db.TxManager
	hub  GameHubInterface
	auth AuthenticatorInterface
	conf *config.Config
}

type GameHubInterface interface {
	CalcCodeSize(code string, language string) int
	EnqueueTestTasks(ctx context.Context, submissionID, gameID, userID int, language, code string) error
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
	userID, err := h.auth.Login(ctx, username, password)
	if err != nil {
		slog.Error("login failed", "error", err)
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
	if sessionID, ok := GetSessionIDFromContext(ctx); ok {
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
	gameRows, err := h.q.ListPublicGames(ctx)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	games := make([]Game, len(gameRows))
	gameIDs := make([]int32, len(gameRows))
	gameID2Index := make(map[int32]int, len(gameRows))
	for i, row := range gameRows {
		var startedAt *int64
		if row.StartedAt.Valid {
			startedAtTimestamp := row.StartedAt.Time.Unix()
			startedAt = &startedAtTimestamp
		}
		games[i] = Game{
			GameID:          int(row.GameID),
			GameType:        GameType(row.GameType),
			IsPublic:        row.IsPublic,
			DisplayName:     row.DisplayName,
			DurationSeconds: int(row.DurationSeconds),
			StartedAt:       startedAt,
			Problem: Problem{
				ProblemID:   int(row.ProblemID),
				Title:       row.Title,
				Description: row.Description,
				Language:    ProblemLanguage(row.Language),
				SampleCode:  row.SampleCode,
			},
		}
		gameIDs[i] = row.GameID
		gameID2Index[row.GameID] = i
	}
	mainPlayerRows, err := h.q.ListMainPlayers(ctx, gameIDs)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	for _, row := range mainPlayerRows {
		idx := gameID2Index[row.GameID]
		game := &games[idx]
		game.MainPlayers = append(game.MainPlayers, User{
			UserID:      int(row.UserID),
			Username:    row.Username,
			DisplayName: row.DisplayName,
			IconPath:    row.IconPath,
			IsAdmin:     row.IsAdmin,
			Label:       toNullable(row.Label),
		})
	}
	return GetGames200JSONResponse{
		Games: games,
	}, nil
}

func (h *Handler) GetGame(ctx context.Context, request GetGameRequestObject, user *db.User) (GetGameResponseObject, error) {
	gameID := request.GameID
	row, err := h.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetGame404JSONResponse{
				Message: "Game not found",
			}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !row.IsPublic && !user.IsAdmin {
		return GetGame404JSONResponse{
			Message: "Game not found",
		}, nil
	}
	var startedAt *int64
	if row.StartedAt.Valid {
		startedAtTimestamp := row.StartedAt.Time.Unix()
		startedAt = &startedAtTimestamp
	}
	mainPlayerRows, err := h.q.ListMainPlayers(ctx, []int32{int32(gameID)})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	mainPlayers := make([]User, len(mainPlayerRows))
	for i, playerRow := range mainPlayerRows {
		mainPlayers[i] = User{
			UserID:      int(playerRow.UserID),
			Username:    playerRow.Username,
			DisplayName: playerRow.DisplayName,
			IconPath:    playerRow.IconPath,
			IsAdmin:     playerRow.IsAdmin,
			Label:       toNullable(playerRow.Label),
		}
	}
	game := Game{
		GameID:          int(row.GameID),
		GameType:        GameType(row.GameType),
		IsPublic:        row.IsPublic,
		DisplayName:     row.DisplayName,
		DurationSeconds: int(row.DurationSeconds),
		StartedAt:       startedAt,
		Problem: Problem{
			ProblemID:   int(row.ProblemID),
			Title:       row.Title,
			Description: row.Description,
			Language:    ProblemLanguage(row.Language),
			SampleCode:  row.SampleCode,
		},
		MainPlayers: mainPlayers,
	}
	return GetGame200JSONResponse{
		Game: game,
	}, nil
}

func (h *Handler) GetGamePlayLatestState(ctx context.Context, request GetGamePlayLatestStateRequestObject, user *db.User) (GetGamePlayLatestStateResponseObject, error) {
	gameID := request.GameID
	row, err := h.q.GetLatestState(ctx, db.GetLatestStateParams{
		GameID: int32(gameID),
		UserID: user.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetGamePlayLatestState200JSONResponse{
				State: LatestGameState{
					Code:                 "",
					Score:                nullable.NewNullNullable[int](),
					BestScoreSubmittedAt: nullable.NewNullNullable[int64](),
					Status:               None,
				},
			}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return GetGamePlayLatestState200JSONResponse{
		State: LatestGameState{
			Code:                 row.Code,
			Score:                toNullableWith(row.CodeSize, func(x int32) int { return int(x) }),
			BestScoreSubmittedAt: nullable.NewNullableWithValue(row.CreatedAt.Time.Unix()),
			Status:               ExecutionStatus(row.Status),
		},
	}, nil
}

func (h *Handler) GetGameWatchLatestStates(ctx context.Context, request GetGameWatchLatestStatesRequestObject, user *db.User) (GetGameWatchLatestStatesResponseObject, error) {
	gameID := request.GameID
	rows, err := h.q.GetLatestStatesOfMainPlayers(ctx, int32(gameID))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	states := make(map[string]LatestGameState, len(rows))
	for _, row := range rows {
		var code string
		if row.Code != nil {
			code = *row.Code
		}
		var status ExecutionStatus
		if row.Status != nil {
			status = ExecutionStatus(*row.Status)
		} else {
			status = None
		}
		states[strconv.Itoa(int(row.UserID))] = LatestGameState{
			Code:                 code,
			Score:                toNullableWith(row.CodeSize, func(x int32) int { return int(x) }),
			BestScoreSubmittedAt: nullable.NewNullableWithValue(row.CreatedAt.Time.Unix()),
			Status:               status,
		}

		if row.UserID == user.UserID && !user.IsAdmin {
			return GetGameWatchLatestStates403JSONResponse{
				Message: "You are one of the main players of this game",
			}, nil
		}
	}
	return GetGameWatchLatestStates200JSONResponse{
		States: states,
	}, nil
}

func (h *Handler) GetGameWatchRanking(ctx context.Context, request GetGameWatchRankingRequestObject, _ *db.User) (GetGameWatchRankingResponseObject, error) {
	gameID := request.GameID

	gameRow, err := h.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetGameWatchRanking404JSONResponse{
				Message: "Game not found",
			}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	gameFinished := isGameFinished(gameRow)

	rows, err := h.q.GetRanking(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetGameWatchRanking200JSONResponse{}, nil
		}
	}
	ranking := make([]RankingEntry, len(rows))
	for i, row := range rows {
		var code nullable.Nullable[string]
		if gameFinished {
			code = nullable.NewNullableWithValue(row.Submission.Code)
		} else {
			code = nullable.NewNullNullable[string]()
		}
		ranking[i] = RankingEntry{
			Player: User{
				UserID:      int(row.User.UserID),
				Username:    row.User.Username,
				DisplayName: row.User.DisplayName,
				IconPath:    row.User.IconPath,
				IsAdmin:     row.User.IsAdmin,
				Label:       toNullable(row.User.Label),
			},
			Score:       int(row.Submission.CodeSize),
			SubmittedAt: row.Submission.CreatedAt.Time.Unix(),
			Code:        code,
		}
	}
	return GetGameWatchRanking200JSONResponse{
		Ranking: ranking,
	}, nil
}

func (h *Handler) PostGamePlayCode(ctx context.Context, request PostGamePlayCodeRequestObject, user *db.User) (PostGamePlayCodeResponseObject, error) {
	gameID := request.GameID

	gameRow, err := h.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PostGamePlayCode404JSONResponse{
				Message: "Game not found",
			}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !isGameRunning(gameRow) {
		return PostGamePlayCode403JSONResponse{
			Message: "Game is not running",
		}, nil
	}

	err = h.q.UpdateCode(ctx, db.UpdateCodeParams{
		GameID: int32(gameID),
		UserID: user.UserID,
		Code:   request.Body.Code,
		Status: "none",
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return PostGamePlayCode200Response{}, nil
}

func (h *Handler) PostGamePlaySubmit(ctx context.Context, request PostGamePlaySubmitRequestObject, user *db.User) (PostGamePlaySubmitResponseObject, error) {
	gameID := request.GameID
	code := request.Body.Code

	gameRow, err := h.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PostGamePlaySubmit404JSONResponse{}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	language := gameRow.Language
	codeSize := h.hub.CalcCodeSize(code, language)

	if !isGameRunning(gameRow) {
		return PostGamePlaySubmit403JSONResponse{
			Message: "Game is not running",
		}, nil
	}

	var submissionID int32
	err = h.txm.RunInTx(ctx, func(qtx db.Querier) error {
		if err := qtx.UpdateCodeAndStatus(ctx, db.UpdateCodeAndStatusParams{
			GameID: int32(gameID),
			UserID: user.UserID,
			Code:   code,
			Status: "running",
		}); err != nil {
			return err
		}
		var err error
		submissionID, err = qtx.CreateSubmission(ctx, db.CreateSubmissionParams{
			GameID:   int32(gameID),
			UserID:   user.UserID,
			Code:     code,
			CodeSize: int32(codeSize),
		})
		return err
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = h.hub.EnqueueTestTasks(ctx, int(submissionID), gameID, int(user.UserID), language, code)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return PostGamePlaySubmit200Response{}, nil
}

func (h *Handler) GetTournament(ctx context.Context, request GetTournamentRequestObject, _ *db.User) (GetTournamentResponseObject, error) {
	tournamentID := int32(request.TournamentID)

	tournament, err := h.q.GetTournamentByID(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetTournament404JSONResponse{Message: "Tournament not found"}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	entryRows, err := h.q.ListTournamentEntries(ctx, tournamentID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	seedToUser := make(map[int]User)
	entries := make([]TournamentEntry, len(entryRows))
	for i, e := range entryRows {
		u := User{
			UserID:      int(e.UserID),
			Username:    e.Username,
			DisplayName: e.DisplayName,
			IconPath:    e.IconPath,
			IsAdmin:     e.IsAdmin,
			Label:       toNullable(e.Label),
		}
		seedToUser[int(e.Seed)] = u
		entries[i] = TournamentEntry{
			User: u,
			Seed: int(e.Seed),
		}
	}

	matchRows, err := h.q.ListTournamentMatches(ctx, tournamentID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	bracketSize := int(tournament.BracketSize)
	numRounds := int(tournament.NumRounds)
	bracketSeeds := standardBracketSeeds(bracketSize)

	// Index matches by (round, position)
	type matchKey struct{ round, position int }
	matchByKey := make(map[matchKey]db.TournamentMatch)
	for _, m := range matchRows {
		matchByKey[matchKey{int(m.Round), int(m.Position)}] = m
	}

	// Collect game IDs for batch fetching
	gameIDs := make(map[int32]bool)
	for _, m := range matchRows {
		if m.GameID != nil {
			gameIDs[*m.GameID] = true
		}
	}

	// Fetch rankings for all games that have started
	type rankingResult struct {
		scores   map[int]int // userID -> score
		winnerID int
	}
	gameRankings := make(map[int32]*rankingResult)
	for gid := range gameIDs {
		gameRow, err := h.q.GetGameByID(ctx, gid)
		if err != nil {
			continue
		}
		if !gameRow.StartedAt.Valid {
			continue
		}
		rankingRows, err := h.q.GetRanking(ctx, gid)
		if err != nil || len(rankingRows) == 0 {
			continue
		}
		rr := &rankingResult{scores: make(map[int]int)}
		for i, r := range rankingRows {
			rr.scores[int(r.User.UserID)] = int(r.Submission.CodeSize)
			if i == 0 {
				rr.winnerID = int(r.User.UserID)
			}
		}
		gameRankings[gid] = rr
	}

	// Build match results bottom-up
	type matchResult struct {
		player1   *User
		player2   *User
		p1Score   *int
		p2Score   *int
		winnerUID *int
		isBye     bool
	}
	resultByKey := make(map[matchKey]*matchResult)

	for round := 0; round < numRounds; round++ {
		numPositions := bracketSize / (1 << (round + 1))
		for pos := 0; pos < numPositions; pos++ {
			m, exists := matchByKey[matchKey{round, pos}]
			mr := &matchResult{}

			if round == 0 {
				// First round: resolve players from bracket seeds
				slot1 := pos * 2
				slot2 := pos*2 + 1
				seed1 := bracketSeeds[slot1]
				seed2 := bracketSeeds[slot2]

				if u, ok := seedToUser[seed1]; ok {
					mr.player1 = &u
				}
				if u, ok := seedToUser[seed2]; ok {
					mr.player2 = &u
				}
			} else {
				// Later rounds: resolve from child match winners
				child1 := resultByKey[matchKey{round - 1, pos * 2}]
				child2 := resultByKey[matchKey{round - 1, pos*2 + 1}]

				if child1 != nil && child1.winnerUID != nil {
					if u, ok := seedToUser[findSeedByUserID(entries, *child1.winnerUID)]; ok {
						mr.player1 = &u
					}
				}
				if child2 != nil && child2.winnerUID != nil {
					if u, ok := seedToUser[findSeedByUserID(entries, *child2.winnerUID)]; ok {
						mr.player2 = &u
					}
				}
			}

			// Check for bye
			if mr.player1 == nil && mr.player2 != nil {
				mr.isBye = true
				uid := mr.player2.UserID
				mr.winnerUID = &uid
			} else if mr.player1 != nil && mr.player2 == nil {
				mr.isBye = true
				uid := mr.player1.UserID
				mr.winnerUID = &uid
			}

			// Resolve scores from game
			if exists && m.GameID != nil && !mr.isBye {
				if rr, ok := gameRankings[*m.GameID]; ok {
					if mr.player1 != nil {
						if s, ok := rr.scores[mr.player1.UserID]; ok {
							score := s
							mr.p1Score = &score
						}
					}
					if mr.player2 != nil {
						if s, ok := rr.scores[mr.player2.UserID]; ok {
							score := s
							mr.p2Score = &score
						}
					}
					// Winner is the one with the best (lowest) score in the ranking
					if mr.player1 != nil && mr.player2 != nil {
						if rr.winnerID == mr.player1.UserID || rr.winnerID == mr.player2.UserID {
							w := rr.winnerID
							mr.winnerUID = &w
						} else {
							// Both players have scores; pick the one with lower score
							if mr.p1Score != nil && mr.p2Score != nil {
								if *mr.p1Score <= *mr.p2Score {
									w := mr.player1.UserID
									mr.winnerUID = &w
								} else {
									w := mr.player2.UserID
									mr.winnerUID = &w
								}
							}
						}
					}
				}
			}

			resultByKey[matchKey{round, pos}] = mr
		}
	}

	// Build API response matches
	apiMatches := make([]TournamentMatch, 0, len(matchRows))
	for round := 0; round < numRounds; round++ {
		numPositions := bracketSize / (1 << (round + 1))
		for pos := 0; pos < numPositions; pos++ {
			m, exists := matchByKey[matchKey{round, pos}]
			mr := resultByKey[matchKey{round, pos}]

			matchID := 0
			var gameID *int
			if exists {
				matchID = int(m.TournamentMatchID)
				if m.GameID != nil {
					gid := int(*m.GameID)
					gameID = &gid
				}
			}

			apiMatches = append(apiMatches, TournamentMatch{
				TournamentMatchID: matchID,
				Round:             round,
				Position:          pos,
				GameID:            gameID,
				Player1:           mr.player1,
				Player2:           mr.player2,
				Player1Score:      mr.p1Score,
				Player2Score:      mr.p2Score,
				WinnerUserID:      mr.winnerUID,
				IsBye:             mr.isBye,
			})
		}
	}

	return GetTournament200JSONResponse{
		Tournament: Tournament{
			TournamentID: int(tournament.TournamentID),
			DisplayName:  tournament.DisplayName,
			BracketSize:  bracketSize,
			NumRounds:    numRounds,
			Entries:      entries,
			Matches:      apiMatches,
		},
	}, nil
}

func findSeedByUserID(entries []TournamentEntry, userID int) int {
	for _, e := range entries {
		if e.User.UserID == userID {
			return e.Seed
		}
	}
	return 0
}

// standardBracketSeeds returns the seed assignments for each slot in a standard
// single-elimination bracket. For bracket_size=8:
// Position: [0]=1, [1]=8, [2]=5, [3]=4, [4]=3, [5]=6, [6]=7, [7]=2
// This ensures Seed 1 vs Seed 2 are on opposite sides, and higher seeds face lower seeds.
func standardBracketSeeds(bracketSize int) []int {
	seeds := make([]int, bracketSize)
	seeds[0] = 1
	// Build the bracket by repeatedly splitting
	for size := 2; size <= bracketSize; size *= 2 {
		// For each pair in the current level, the new opponent for seed[i]
		// is (size + 1 - seed[i])
		temp := make([]int, size)
		for i := 0; i < size/2; i++ {
			temp[i*2] = seeds[i]
			temp[i*2+1] = size + 1 - seeds[i]
		}
		copy(seeds, temp)
	}
	return seeds
}

func isGameRunning(game db.GetGameByIDRow) bool {
	if !game.StartedAt.Valid {
		return false
	}
	endTime := game.StartedAt.Time.Add(time.Duration(game.DurationSeconds) * time.Second)
	return time.Now().Before(endTime)
}

func isGameFinished(game db.GetGameByIDRow) bool {
	if !game.StartedAt.Valid {
		return false
	}
	endTime := game.StartedAt.Time.Add(time.Duration(game.DurationSeconds) * time.Second)
	return !time.Now().Before(endTime)
}

func toNullable[T any](p *T) nullable.Nullable[T] {
	if p == nil {
		return nullable.NewNullNullable[T]()
	}
	return nullable.NewNullableWithValue(*p)
}

func toNullableWith[T, U any](p *T, f func(T) U) nullable.Nullable[U] {
	if p == nil {
		return nullable.NewNullNullable[U]()
	}
	return nullable.NewNullableWithValue(f(*p))
}
