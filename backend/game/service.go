package game

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"albatross-2026-backend/db"
)

type HubInterface interface {
	CalcCodeSize(code string, language string) int
	EnqueueTestTasks(ctx context.Context, submissionID, gameID, userID int, language, code string) error
}

type Service struct {
	q   db.Querier
	txm db.TxManager
	hub HubInterface
}

func NewService(q db.Querier, txm db.TxManager, hub HubInterface) *Service {
	return &Service{q: q, txm: txm, hub: hub}
}

// Domain types

type Player struct {
	UserID      int
	Username    string
	DisplayName string
	IconPath    *string
	IsAdmin     bool
	Label       *string
}

type ProblemDetail struct {
	ProblemID   int
	Title       string
	Description string
	Language    string
	SampleCode  string
}

type Detail struct {
	GameID          int
	GameType        string
	IsPublic        bool
	DisplayName     string
	DurationSeconds int
	StartedAt       *time.Time
	Problem         ProblemDetail
	MainPlayers     []Player
}

type LatestState struct {
	Code                 string
	Score                *int
	BestScoreSubmittedAt *int64
	Status               string
}

type RankingEntry struct {
	Player      Player
	Score       int
	SubmittedAt int64
	Code        *string
}

type SubmissionDetail struct {
	SubmissionID int
	GameID       int
	Code         string
	CodeSize     int
	Status       string
	CreatedAt    int64
}

// Helper functions

func IsGameRunning(startedAt pgtype.Timestamp, durationSeconds int32) bool {
	if !startedAt.Valid {
		return false
	}
	endTime := startedAt.Time.Add(time.Duration(durationSeconds) * time.Second)
	return time.Now().Before(endTime)
}

func IsGameFinished(startedAt pgtype.Timestamp, durationSeconds int32) bool {
	if !startedAt.Valid {
		return false
	}
	endTime := startedAt.Time.Add(time.Duration(durationSeconds) * time.Second)
	return !time.Now().Before(endTime)
}

func playerFromMainPlayerRow(row db.ListMainPlayersRow) Player {
	return Player{
		UserID:      int(row.UserID),
		Username:    row.Username,
		DisplayName: row.DisplayName,
		IconPath:    row.IconPath,
		IsAdmin:     row.IsAdmin,
		Label:       row.Label,
	}
}

func gameDetailFromPublicRow(row db.ListPublicGamesRow) Detail {
	var startedAt *time.Time
	if row.StartedAt.Valid {
		t := row.StartedAt.Time
		startedAt = &t
	}
	return Detail{
		GameID:          int(row.GameID),
		GameType:        row.GameType,
		IsPublic:        row.IsPublic,
		DisplayName:     row.DisplayName,
		DurationSeconds: int(row.DurationSeconds),
		StartedAt:       startedAt,
		Problem: ProblemDetail{
			ProblemID:   int(row.ProblemID),
			Title:       row.Title,
			Description: row.Description,
			Language:    row.Language,
			SampleCode:  row.SampleCode,
		},
	}
}

func gameDetailFromGetRow(row db.GetGameByIDRow) Detail {
	var startedAt *time.Time
	if row.StartedAt.Valid {
		t := row.StartedAt.Time
		startedAt = &t
	}
	return Detail{
		GameID:          int(row.GameID),
		GameType:        row.GameType,
		IsPublic:        row.IsPublic,
		DisplayName:     row.DisplayName,
		DurationSeconds: int(row.DurationSeconds),
		StartedAt:       startedAt,
		Problem: ProblemDetail{
			ProblemID:   int(row.ProblemID),
			Title:       row.Title,
			Description: row.Description,
			Language:    row.Language,
			SampleCode:  row.SampleCode,
		},
	}
}

// Service methods

func (s *Service) ListPublicGames(ctx context.Context) ([]Detail, error) {
	gameRows, err := s.q.ListPublicGames(ctx)
	if err != nil {
		return nil, err
	}
	games := make([]Detail, len(gameRows))
	gameIDs := make([]int32, len(gameRows))
	gameID2Index := make(map[int32]int, len(gameRows))
	for i, row := range gameRows {
		games[i] = gameDetailFromPublicRow(row)
		gameIDs[i] = row.GameID
		gameID2Index[row.GameID] = i
	}
	mainPlayerRows, err := s.q.ListMainPlayers(ctx, gameIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range mainPlayerRows {
		idx := gameID2Index[row.GameID]
		games[idx].MainPlayers = append(games[idx].MainPlayers, playerFromMainPlayerRow(row))
	}
	return games, nil
}

func (s *Service) GetGameByID(ctx context.Context, gameID int, isAdmin bool) (Detail, error) {
	row, err := s.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Detail{}, ErrNotFound
		}
		return Detail{}, err
	}
	if !row.IsPublic && !isAdmin {
		return Detail{}, ErrNotFound
	}
	game := gameDetailFromGetRow(row)
	mainPlayerRows, err := s.q.ListMainPlayers(ctx, []int32{int32(gameID)})
	if err != nil {
		return Detail{}, err
	}
	for _, playerRow := range mainPlayerRows {
		game.MainPlayers = append(game.MainPlayers, playerFromMainPlayerRow(playerRow))
	}
	return game, nil
}

func (s *Service) SaveCode(ctx context.Context, gameID int, userID int32, code string) error {
	gameRow, err := s.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if !IsGameRunning(gameRow.StartedAt, gameRow.DurationSeconds) {
		return ErrGameNotRunning
	}
	return s.q.UpdateCode(ctx, db.UpdateCodeParams{
		GameID: int32(gameID),
		UserID: userID,
		Code:   code,
		Status: "none",
	})
}

func (s *Service) SubmitCode(ctx context.Context, gameID int, userID int32, code string) error {
	gameRow, err := s.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	language := gameRow.Language
	codeSize := s.hub.CalcCodeSize(code, language)

	if !IsGameRunning(gameRow.StartedAt, gameRow.DurationSeconds) {
		return ErrGameNotRunning
	}

	var submissionID int32
	err = s.txm.RunInTx(ctx, func(qtx db.Querier) error {
		if err := qtx.UpdateCodeAndStatus(ctx, db.UpdateCodeAndStatusParams{
			GameID: int32(gameID),
			UserID: userID,
			Code:   code,
			Status: "running",
		}); err != nil {
			return err
		}
		var err error
		submissionID, err = qtx.CreateSubmission(ctx, db.CreateSubmissionParams{
			GameID:   int32(gameID),
			UserID:   userID,
			Code:     code,
			CodeSize: int32(codeSize),
		})
		return err
	})
	if err != nil {
		return err
	}

	return s.hub.EnqueueTestTasks(ctx, int(submissionID), gameID, int(userID), language, code)
}

func (s *Service) GetLatestState(ctx context.Context, gameID int, userID int32) (LatestState, error) {
	row, err := s.q.GetLatestState(ctx, db.GetLatestStateParams{
		GameID: int32(gameID),
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LatestState{Status: "none"}, nil
		}
		return LatestState{}, err
	}
	var score *int
	if row.CodeSize != nil {
		s := int(*row.CodeSize)
		score = &s
	}
	var submittedAt *int64
	if row.CreatedAt.Valid {
		ts := row.CreatedAt.Time.Unix()
		submittedAt = &ts
	}
	return LatestState{
		Code:                 row.Code,
		Score:                score,
		BestScoreSubmittedAt: submittedAt,
		Status:               row.Status,
	}, nil
}

func (s *Service) GetWatchLatestStates(ctx context.Context, gameID int, userID *int32, isAdmin bool) (map[int]LatestState, error) {
	rows, err := s.q.GetLatestStatesOfMainPlayers(ctx, int32(gameID))
	if err != nil {
		return nil, err
	}
	states := make(map[int]LatestState, len(rows))
	for _, row := range rows {
		var code string
		if row.Code != nil {
			code = *row.Code
		}
		var status string
		if row.Status != nil {
			status = *row.Status
		} else {
			status = "none"
		}
		var score *int
		if row.CodeSize != nil {
			s := int(*row.CodeSize)
			score = &s
		}
		var submittedAt *int64
		if row.CreatedAt.Valid {
			ts := row.CreatedAt.Time.Unix()
			submittedAt = &ts
		}

		if userID != nil && row.UserID == *userID && !isAdmin {
			return nil, ErrForbidden
		}

		states[int(row.UserID)] = LatestState{
			Code:                 code,
			Score:                score,
			BestScoreSubmittedAt: submittedAt,
			Status:               status,
		}
	}
	return states, nil
}

func (s *Service) GetRanking(ctx context.Context, gameID int) ([]RankingEntry, bool, error) {
	gameRow, err := s.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, ErrNotFound
		}
		return nil, false, err
	}
	finished := IsGameFinished(gameRow.StartedAt, gameRow.DurationSeconds)

	rows, err := s.q.GetRanking(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, finished, nil
		}
		return nil, false, err
	}
	ranking := make([]RankingEntry, len(rows))
	for i, row := range rows {
		var code *string
		if finished {
			code = &row.Submission.Code
		}
		ranking[i] = RankingEntry{
			Player: Player{
				UserID:      int(row.User.UserID),
				Username:    row.User.Username,
				DisplayName: row.User.DisplayName,
				IconPath:    row.User.IconPath,
				IsAdmin:     row.User.IsAdmin,
				Label:       row.User.Label,
			},
			Score:       int(row.Submission.CodeSize),
			SubmittedAt: row.Submission.CreatedAt.Time.Unix(),
			Code:        code,
		}
	}
	return ranking, finished, nil
}

// UpdateGameParams holds parameters for updating a game with its players.
type UpdateGameParams struct {
	GameID          int
	GameType        string
	IsPublic        bool
	DisplayName     string
	DurationSeconds int
	StartedAt       pgtype.Timestamp
	ProblemID       int
	MainPlayerIDs   []int
}

func (s *Service) StartGame(ctx context.Context, gameID int) error {
	gameRow, err := s.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	testcases, err := s.q.ListTestcasesByProblemID(ctx, gameRow.ProblemID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if len(testcases) == 0 {
		return ErrNoTestcases
	}

	startedAt := time.Now().Add(10 * time.Second)
	return s.q.UpdateGameStartedAt(ctx, db.UpdateGameStartedAtParams{
		GameID: int32(gameID),
		StartedAt: pgtype.Timestamp{
			Time:  startedAt,
			Valid: true,
		},
	})
}

func (s *Service) UpdateGameWithPlayers(ctx context.Context, params UpdateGameParams) error {
	return s.txm.RunInTx(ctx, func(qtx db.Querier) error {
		if err := qtx.UpdateGame(ctx, db.UpdateGameParams{
			GameID:          int32(params.GameID),
			GameType:        params.GameType,
			IsPublic:        params.IsPublic,
			DisplayName:     params.DisplayName,
			DurationSeconds: int32(params.DurationSeconds),
			StartedAt:       params.StartedAt,
			ProblemID:       int32(params.ProblemID),
		}); err != nil {
			return err
		}
		if err := qtx.RemoveAllMainPlayers(ctx, int32(params.GameID)); err != nil {
			return err
		}
		for _, userID := range params.MainPlayerIDs {
			if err := qtx.AddMainPlayer(ctx, db.AddMainPlayerParams{
				GameID: int32(params.GameID),
				UserID: int32(userID),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) RejudgeSubmission(ctx context.Context, submissionID int32, gameID int, userID int, language, code string) error {
	err := s.txm.RunInTx(ctx, func(qtx db.Querier) error {
		if err := qtx.DeleteTestcaseResultsBySubmissionID(ctx, submissionID); err != nil {
			return err
		}
		return qtx.UpdateSubmissionStatus(ctx, db.UpdateSubmissionStatusParams{
			SubmissionID: submissionID,
			Status:       "running",
		})
	})
	if err != nil {
		return err
	}
	return s.hub.EnqueueTestTasks(ctx, int(submissionID), gameID, userID, language, code)
}

func (s *Service) RejudgeLatestSubmissionsByGame(ctx context.Context, gameID int) error {
	gameRow, err := s.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	submissions, err := s.q.GetLatestSubmissionsByGameID(ctx, int32(gameID))
	if err != nil {
		return err
	}

	for _, sub := range submissions {
		if err := s.RejudgeSubmission(ctx, sub.SubmissionID, int(sub.GameID), int(sub.UserID), gameRow.Language, sub.Code); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RejudgeAllSubmissionsByGame(ctx context.Context, gameID int) error {
	gameRow, err := s.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	submissions, err := s.q.GetSubmissionsByGameID(ctx, int32(gameID))
	if err != nil {
		return err
	}

	for _, sub := range submissions {
		if err := s.RejudgeSubmission(ctx, sub.SubmissionID, int(sub.GameID), int(sub.UserID), gameRow.Language, sub.Code); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) FixSubmissionStatuses(ctx context.Context) error {
	submissionIDs, err := s.q.ListSubmissionIDs(ctx)
	if err != nil {
		return err
	}
	for _, submissionID := range submissionIDs {
		as, err := s.q.AggregateTestcaseResults(ctx, submissionID)
		if err != nil {
			return err
		}
		if err := s.q.UpdateSubmissionStatus(ctx, db.UpdateSubmissionStatusParams{
			SubmissionID: submissionID,
			Status:       as,
		}); err != nil {
			return err
		}
	}

	gameStates, err := s.q.ListGameStateIDs(ctx)
	if err != nil {
		return err
	}
	for _, r := range gameStates {
		if err := s.q.SyncGameStateBestScoreSubmission(ctx, db.SyncGameStateBestScoreSubmissionParams(r)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetSubmissions(ctx context.Context, gameID int, userID int32) ([]SubmissionDetail, error) {
	_, err := s.q.GetGameByID(ctx, int32(gameID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	rows, err := s.q.GetSubmissionsByGameIDAndUserID(ctx, db.GetSubmissionsByGameIDAndUserIDParams{
		GameID: int32(gameID),
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	submissions := make([]SubmissionDetail, len(rows))
	for i, row := range rows {
		submissions[i] = SubmissionDetail{
			SubmissionID: int(row.SubmissionID),
			GameID:       int(row.GameID),
			Code:         row.Code,
			CodeSize:     int(row.CodeSize),
			Status:       row.Status,
			CreatedAt:    row.CreatedAt.Time.Unix(),
		}
	}
	return submissions, nil
}
