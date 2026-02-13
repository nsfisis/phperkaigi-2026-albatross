package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/nullable"

	"albatross-2026-backend/auth"
	"albatross-2026-backend/db"
)

type Handler struct {
	q   *db.Queries
	hub GameHubInterface
}

type GameHubInterface interface {
	CalcCodeSize(code string, language string) int
	EnqueueTestTasks(ctx context.Context, submissionID, gameID, userID int, language, code string) error
}

func (h *Handler) PostLogin(ctx context.Context, request PostLoginRequestObject) (PostLoginResponseObject, error) {
	username := request.Body.Username
	password := request.Body.Password
	userID, err := auth.Login(ctx, h.q, username, password)
	if err != nil {
		log.Printf("login failed: %v", err)
		var msg string
		if errors.Is(err, auth.ErrForteeLoginTimeout) {
			msg = "ログインに失敗しました"
		} else {
			msg = "ユーザー名またはパスワードが誤っています"
		}
		return PostLogin401JSONResponse{
			UnauthorizedJSONResponse: UnauthorizedJSONResponse{
				Message: msg,
			},
		}, nil
	}

	user, err := h.q.GetUserByID(ctx, int32(userID))
	if err != nil {
		return PostLogin401JSONResponse{
			UnauthorizedJSONResponse: UnauthorizedJSONResponse{
				Message: "ログインに失敗しました",
			},
		}, nil
	}

	jwt, err := auth.NewJWT(&user)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return PostLogin200JSONResponse{
		Token: jwt,
	}, nil
}

func (h *Handler) GetGames(ctx context.Context, _ GetGamesRequestObject, _ *auth.JWTClaims) (GetGamesResponseObject, error) {
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
			GameType:        GameGameType(row.GameType),
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

func (h *Handler) GetGame(ctx context.Context, request GetGameRequestObject, user *auth.JWTClaims) (GetGameResponseObject, error) {
	gameID := request.GameID
	row, err := h.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetGame404JSONResponse{
				NotFoundJSONResponse: NotFoundJSONResponse{
					Message: "Game not found",
				},
			}, nil
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !row.IsPublic && !user.IsAdmin {
		return GetGame404JSONResponse{
			NotFoundJSONResponse: NotFoundJSONResponse{
				Message: "Game not found",
			},
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
		GameType:        GameGameType(row.GameType),
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

func (h *Handler) GetGamePlayLatestState(ctx context.Context, request GetGamePlayLatestStateRequestObject, user *auth.JWTClaims) (GetGamePlayLatestStateResponseObject, error) {
	gameID := request.GameID
	userID := user.UserID
	row, err := h.q.GetLatestState(ctx, db.GetLatestStateParams{
		GameID: int32(gameID),
		UserID: int32(userID),
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

func (h *Handler) GetGameWatchLatestStates(ctx context.Context, request GetGameWatchLatestStatesRequestObject, user *auth.JWTClaims) (GetGameWatchLatestStatesResponseObject, error) {
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

		if int(row.UserID) == user.UserID && !user.IsAdmin {
			return GetGameWatchLatestStates403JSONResponse{
				ForbiddenJSONResponse: ForbiddenJSONResponse{
					Message: "You are one of the main players of this game",
				},
			}, nil
		}
	}
	return GetGameWatchLatestStates200JSONResponse{
		States: states,
	}, nil
}

func (h *Handler) GetGameWatchRanking(ctx context.Context, request GetGameWatchRankingRequestObject, _ *auth.JWTClaims) (GetGameWatchRankingResponseObject, error) {
	gameID := request.GameID
	rows, err := h.q.GetRanking(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetGameWatchRanking200JSONResponse{}, nil
		}
	}
	ranking := make([]RankingEntry, len(rows))
	for i, row := range rows {
		// TODO: check if game is finished.
		code := &row.Submission.Code
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
			Code:        toNullable(code),
		}
	}
	return GetGameWatchRanking200JSONResponse{
		Ranking: ranking,
	}, nil
}

func (h *Handler) PostGamePlayCode(ctx context.Context, request PostGamePlayCodeRequestObject, user *auth.JWTClaims) (PostGamePlayCodeResponseObject, error) {
	gameID := request.GameID
	userID := user.UserID
	// TODO: check if the game is running
	err := h.q.UpdateCode(ctx, db.UpdateCodeParams{
		GameID: int32(gameID),
		UserID: int32(userID),
		Code:   request.Body.Code,
		Status: "none",
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return PostGamePlayCode200Response{}, nil
}

func (h *Handler) PostGamePlaySubmit(ctx context.Context, request PostGamePlaySubmitRequestObject, user *auth.JWTClaims) (PostGamePlaySubmitResponseObject, error) {
	gameID := request.GameID
	userID := user.UserID
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
	// TODO: check if the game is running
	// TODO: transaction
	err = h.q.UpdateCodeAndStatus(ctx, db.UpdateCodeAndStatusParams{
		GameID: int32(gameID),
		UserID: int32(userID),
		Code:   code,
		Status: "running",
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	submissionID, err := h.q.CreateSubmission(ctx, db.CreateSubmissionParams{
		GameID:   int32(gameID),
		UserID:   int32(userID),
		Code:     code,
		CodeSize: int32(codeSize),
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	err = h.hub.EnqueueTestTasks(ctx, int(submissionID), gameID, userID, language, code)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return PostGamePlaySubmit200Response{}, nil
}

func (h *Handler) GetTournament(ctx context.Context, request GetTournamentRequestObject, _ *auth.JWTClaims) (GetTournamentResponseObject, error) {
	gameIDs := []int32{
		int32(request.Params.Game1),
		int32(request.Params.Game2),
		int32(request.Params.Game3),
		int32(request.Params.Game4),
		int32(request.Params.Game5),
	}

	matches := make([]TournamentMatch, 0, 5)

	for _, gameID := range gameIDs {
		gameRow, err := h.q.GetGameByID(ctx, gameID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		playerRows, err := h.q.ListMainPlayers(ctx, []int32{gameID})
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		var player1, player2 *User
		if len(playerRows) > 0 {
			p1 := User{
				UserID:      int(playerRows[0].UserID),
				Username:    playerRows[0].Username,
				DisplayName: playerRows[0].DisplayName,
				IconPath:    playerRows[0].IconPath,
				IsAdmin:     playerRows[0].IsAdmin,
				Label:       toNullable(playerRows[0].Label),
			}
			player1 = &p1
		}
		if len(playerRows) > 1 {
			p2 := User{
				UserID:      int(playerRows[1].UserID),
				Username:    playerRows[1].Username,
				DisplayName: playerRows[1].DisplayName,
				IconPath:    playerRows[1].IconPath,
				IsAdmin:     playerRows[1].IsAdmin,
				Label:       toNullable(playerRows[1].Label),
			}
			player2 = &p2
		}

		var winnerID *int
		var player1Score, player2Score *int

		if gameRow.StartedAt.Valid {
			rankingRows, err := h.q.GetRanking(ctx, gameID)
			if err == nil && len(rankingRows) > 0 {
				// Find scores for each player
				for _, ranking := range rankingRows {
					userID := int(ranking.User.UserID)
					score := int(ranking.Submission.CodeSize)

					if player1 != nil && player1.UserID == userID {
						player1Score = &score
						if winnerID == nil {
							winnerID = &userID
						}
					}
					if player2 != nil && player2.UserID == userID {
						player2Score = &score
						if winnerID == nil {
							winnerID = &userID
						}
					}
				}
			}
		}

		match := TournamentMatch{
			GameID:       int(gameID),
			Player1:      player1,
			Player2:      player2,
			Player1Score: player1Score,
			Player2Score: player2Score,
			Winner:       winnerID,
		}
		matches = append(matches, match)
	}

	return GetTournament200JSONResponse{
		Tournament: Tournament{
			Matches: matches,
		},
	}, nil
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
