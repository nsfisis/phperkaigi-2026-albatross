package admin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"albatross-2026-backend/config"
	"albatross-2026-backend/db"
	"albatross-2026-backend/game"
	"albatross-2026-backend/session"
	"albatross-2026-backend/tournament"
)

// mockQuerier implements db.Querier for admin handler testing.
type mockQuerier struct {
	db.Querier
	getUserByIDFunc                         func(ctx context.Context, userID int32) (db.User, error)
	listUsersFunc                           func(ctx context.Context) ([]db.User, error)
	updateUserFunc                          func(ctx context.Context, arg db.UpdateUserParams) error
	listAllGamesFunc                        func(ctx context.Context) ([]db.Game, error)
	getGameByIDFunc                         func(ctx context.Context, gameID int32) (db.GetGameByIDRow, error)
	listProblemsFunc                        func(ctx context.Context) ([]db.Problem, error)
	getProblemByIDFunc                      func(ctx context.Context, problemID int32) (db.Problem, error)
	createGameFunc                          func(ctx context.Context, arg db.CreateGameParams) (int32, error)
	createProblemFunc                       func(ctx context.Context, arg db.CreateProblemParams) (int32, error)
	updateProblemFunc                       func(ctx context.Context, arg db.UpdateProblemParams) error
	listTestcasesByProblemIDFunc            func(ctx context.Context, problemID int32) ([]db.Testcase, error)
	getTestcaseByIDFunc                     func(ctx context.Context, testcaseID int32) (db.Testcase, error)
	createTestcaseFunc                      func(ctx context.Context, arg db.CreateTestcaseParams) (int32, error)
	updateTestcaseFunc                      func(ctx context.Context, arg db.UpdateTestcaseParams) error
	deleteTestcaseFunc                      func(ctx context.Context, testcaseID int32) error
	deleteTestcaseResultsBySubmissionIDFunc func(ctx context.Context, submissionID int32) error
	listMainPlayersFunc                     func(ctx context.Context, gameIDs []int32) ([]db.ListMainPlayersRow, error)
	listSubmissionIDsFunc                   func(ctx context.Context) ([]int32, error)
	getSubmissionsByGameIDFunc              func(ctx context.Context, gameID int32) ([]db.Submission, error)
	getLatestSubmissionsByGameIDFunc        func(ctx context.Context, gameID int32) ([]db.Submission, error)
	getSubmissionByIDFunc                   func(ctx context.Context, submissionID int32) (db.Submission, error)
	getTestcaseResultsBySubmIDFunc          func(ctx context.Context, submissionID int32) ([]db.TestcaseResult, error)
	updateSubmissionStatusFunc              func(ctx context.Context, arg db.UpdateSubmissionStatusParams) error
	listTestcasesByGameIDFunc               func(ctx context.Context, gameID int32) ([]db.Testcase, error)
	updateGameStartedAtFunc                 func(ctx context.Context, arg db.UpdateGameStartedAtParams) error
	listTournamentsFunc                     func(ctx context.Context) ([]db.Tournament, error)
	getTournamentByIDFunc                   func(ctx context.Context, tournamentID int32) (db.Tournament, error)
	createTournamentFunc                    func(ctx context.Context, arg db.CreateTournamentParams) (int32, error)
	updateTournamentFunc                    func(ctx context.Context, arg db.UpdateTournamentParams) error
	listTournamentEntriesFunc               func(ctx context.Context, tournamentID int32) ([]db.ListTournamentEntriesRow, error)
	deleteTournamentEntriesFunc             func(ctx context.Context, tournamentID int32) error
	createTournamentEntryFunc               func(ctx context.Context, arg db.CreateTournamentEntryParams) error
	listTournamentMatchesFunc               func(ctx context.Context, tournamentID int32) ([]db.TournamentMatch, error)
	createTournamentMatchFunc               func(ctx context.Context, arg db.CreateTournamentMatchParams) error
	updateTournamentMatchGameFunc           func(ctx context.Context, arg db.UpdateTournamentMatchGameParams) error
	updateGameFunc                          func(ctx context.Context, arg db.UpdateGameParams) error
	removeAllMainPlayersFunc                func(ctx context.Context, gameID int32) error
	addMainPlayerFunc                       func(ctx context.Context, arg db.AddMainPlayerParams) error
	aggregateTestcaseResultsFunc            func(ctx context.Context, submissionID int32) (string, error)
	listGameStateIDsFunc                    func(ctx context.Context) ([]db.ListGameStateIDsRow, error)
	syncGameStateBestScoreSubmissionFunc    func(ctx context.Context, arg db.SyncGameStateBestScoreSubmissionParams) error
}

func (m *mockQuerier) GetUserByID(ctx context.Context, userID int32) (db.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return db.User{}, pgx.ErrNoRows
}

func (m *mockQuerier) ListUsers(ctx context.Context) ([]db.User, error) {
	if m.listUsersFunc != nil {
		return m.listUsersFunc(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) UpdateUser(ctx context.Context, arg db.UpdateUserParams) error {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) ListAllGames(ctx context.Context) ([]db.Game, error) {
	if m.listAllGamesFunc != nil {
		return m.listAllGamesFunc(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) GetGameByID(ctx context.Context, gameID int32) (db.GetGameByIDRow, error) {
	if m.getGameByIDFunc != nil {
		return m.getGameByIDFunc(ctx, gameID)
	}
	return db.GetGameByIDRow{}, pgx.ErrNoRows
}

func (m *mockQuerier) ListProblems(ctx context.Context) ([]db.Problem, error) {
	if m.listProblemsFunc != nil {
		return m.listProblemsFunc(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) GetProblemByID(ctx context.Context, problemID int32) (db.Problem, error) {
	if m.getProblemByIDFunc != nil {
		return m.getProblemByIDFunc(ctx, problemID)
	}
	return db.Problem{}, pgx.ErrNoRows
}

func (m *mockQuerier) CreateGame(ctx context.Context, arg db.CreateGameParams) (int32, error) {
	if m.createGameFunc != nil {
		return m.createGameFunc(ctx, arg)
	}
	return 1, nil
}

func (m *mockQuerier) CreateProblem(ctx context.Context, arg db.CreateProblemParams) (int32, error) {
	if m.createProblemFunc != nil {
		return m.createProblemFunc(ctx, arg)
	}
	return 1, nil
}

func (m *mockQuerier) UpdateProblem(ctx context.Context, arg db.UpdateProblemParams) error {
	if m.updateProblemFunc != nil {
		return m.updateProblemFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) ListTestcasesByProblemID(ctx context.Context, problemID int32) ([]db.Testcase, error) {
	if m.listTestcasesByProblemIDFunc != nil {
		return m.listTestcasesByProblemIDFunc(ctx, problemID)
	}
	return nil, nil
}

func (m *mockQuerier) GetTestcaseByID(ctx context.Context, testcaseID int32) (db.Testcase, error) {
	if m.getTestcaseByIDFunc != nil {
		return m.getTestcaseByIDFunc(ctx, testcaseID)
	}
	return db.Testcase{}, pgx.ErrNoRows
}

func (m *mockQuerier) CreateTestcase(ctx context.Context, arg db.CreateTestcaseParams) (int32, error) {
	if m.createTestcaseFunc != nil {
		return m.createTestcaseFunc(ctx, arg)
	}
	return 1, nil
}

func (m *mockQuerier) UpdateTestcase(ctx context.Context, arg db.UpdateTestcaseParams) error {
	if m.updateTestcaseFunc != nil {
		return m.updateTestcaseFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) DeleteTestcase(ctx context.Context, testcaseID int32) error {
	if m.deleteTestcaseFunc != nil {
		return m.deleteTestcaseFunc(ctx, testcaseID)
	}
	return nil
}

func (m *mockQuerier) ListMainPlayers(ctx context.Context, gameIDs []int32) ([]db.ListMainPlayersRow, error) {
	if m.listMainPlayersFunc != nil {
		return m.listMainPlayersFunc(ctx, gameIDs)
	}
	return nil, nil
}

func (m *mockQuerier) ListSubmissionIDs(ctx context.Context) ([]int32, error) {
	if m.listSubmissionIDsFunc != nil {
		return m.listSubmissionIDsFunc(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) GetSubmissionsByGameID(ctx context.Context, gameID int32) ([]db.Submission, error) {
	if m.getSubmissionsByGameIDFunc != nil {
		return m.getSubmissionsByGameIDFunc(ctx, gameID)
	}
	return nil, nil
}

func (m *mockQuerier) GetLatestSubmissionsByGameID(ctx context.Context, gameID int32) ([]db.Submission, error) {
	if m.getLatestSubmissionsByGameIDFunc != nil {
		return m.getLatestSubmissionsByGameIDFunc(ctx, gameID)
	}
	return nil, nil
}

func (m *mockQuerier) GetSubmissionByID(ctx context.Context, submissionID int32) (db.Submission, error) {
	if m.getSubmissionByIDFunc != nil {
		return m.getSubmissionByIDFunc(ctx, submissionID)
	}
	return db.Submission{}, pgx.ErrNoRows
}

func (m *mockQuerier) GetTestcaseResultsBySubmissionID(ctx context.Context, submissionID int32) ([]db.TestcaseResult, error) {
	if m.getTestcaseResultsBySubmIDFunc != nil {
		return m.getTestcaseResultsBySubmIDFunc(ctx, submissionID)
	}
	return nil, nil
}

func (m *mockQuerier) ListTestcasesByGameID(ctx context.Context, gameID int32) ([]db.Testcase, error) {
	if m.listTestcasesByGameIDFunc != nil {
		return m.listTestcasesByGameIDFunc(ctx, gameID)
	}
	return nil, nil
}

func (m *mockQuerier) UpdateGameStartedAt(ctx context.Context, arg db.UpdateGameStartedAtParams) error {
	if m.updateGameStartedAtFunc != nil {
		return m.updateGameStartedAtFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) ListTestcasesByProblemIDForUpdate(ctx context.Context, problemID int32) ([]db.Testcase, error) {
	return m.ListTestcasesByProblemID(ctx, problemID)
}

func (m *mockQuerier) ListTournaments(ctx context.Context) ([]db.Tournament, error) {
	if m.listTournamentsFunc != nil {
		return m.listTournamentsFunc(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) GetTournamentByID(ctx context.Context, tournamentID int32) (db.Tournament, error) {
	if m.getTournamentByIDFunc != nil {
		return m.getTournamentByIDFunc(ctx, tournamentID)
	}
	return db.Tournament{}, pgx.ErrNoRows
}

func (m *mockQuerier) CreateTournament(ctx context.Context, arg db.CreateTournamentParams) (int32, error) {
	if m.createTournamentFunc != nil {
		return m.createTournamentFunc(ctx, arg)
	}
	return 1, nil
}

func (m *mockQuerier) UpdateTournament(ctx context.Context, arg db.UpdateTournamentParams) error {
	if m.updateTournamentFunc != nil {
		return m.updateTournamentFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) ListTournamentEntries(ctx context.Context, tournamentID int32) ([]db.ListTournamentEntriesRow, error) {
	if m.listTournamentEntriesFunc != nil {
		return m.listTournamentEntriesFunc(ctx, tournamentID)
	}
	return nil, nil
}

func (m *mockQuerier) DeleteTournamentEntries(ctx context.Context, tournamentID int32) error {
	if m.deleteTournamentEntriesFunc != nil {
		return m.deleteTournamentEntriesFunc(ctx, tournamentID)
	}
	return nil
}

func (m *mockQuerier) CreateTournamentEntry(ctx context.Context, arg db.CreateTournamentEntryParams) error {
	if m.createTournamentEntryFunc != nil {
		return m.createTournamentEntryFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) ListTournamentMatches(ctx context.Context, tournamentID int32) ([]db.TournamentMatch, error) {
	if m.listTournamentMatchesFunc != nil {
		return m.listTournamentMatchesFunc(ctx, tournamentID)
	}
	return nil, nil
}

func (m *mockQuerier) CreateTournamentMatch(ctx context.Context, arg db.CreateTournamentMatchParams) error {
	if m.createTournamentMatchFunc != nil {
		return m.createTournamentMatchFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) DeleteTestcaseResultsBySubmissionID(ctx context.Context, submissionID int32) error {
	if m.deleteTestcaseResultsBySubmissionIDFunc != nil {
		return m.deleteTestcaseResultsBySubmissionIDFunc(ctx, submissionID)
	}
	return nil
}

func (m *mockQuerier) UpdateSubmissionStatus(ctx context.Context, arg db.UpdateSubmissionStatusParams) error {
	if m.updateSubmissionStatusFunc != nil {
		return m.updateSubmissionStatusFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) UpdateTournamentMatchGame(ctx context.Context, arg db.UpdateTournamentMatchGameParams) error {
	if m.updateTournamentMatchGameFunc != nil {
		return m.updateTournamentMatchGameFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) UpdateGame(ctx context.Context, arg db.UpdateGameParams) error {
	if m.updateGameFunc != nil {
		return m.updateGameFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) RemoveAllMainPlayers(ctx context.Context, gameID int32) error {
	if m.removeAllMainPlayersFunc != nil {
		return m.removeAllMainPlayersFunc(ctx, gameID)
	}
	return nil
}

func (m *mockQuerier) AddMainPlayer(ctx context.Context, arg db.AddMainPlayerParams) error {
	if m.addMainPlayerFunc != nil {
		return m.addMainPlayerFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) AggregateTestcaseResults(ctx context.Context, submissionID int32) (string, error) {
	if m.aggregateTestcaseResultsFunc != nil {
		return m.aggregateTestcaseResultsFunc(ctx, submissionID)
	}
	return "pass", nil
}

func (m *mockQuerier) ListGameStateIDs(ctx context.Context) ([]db.ListGameStateIDsRow, error) {
	if m.listGameStateIDsFunc != nil {
		return m.listGameStateIDsFunc(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) SyncGameStateBestScoreSubmission(ctx context.Context, arg db.SyncGameStateBestScoreSubmissionParams) error {
	if m.syncGameStateBestScoreSubmissionFunc != nil {
		return m.syncGameStateBestScoreSubmissionFunc(ctx, arg)
	}
	return nil
}

// mockGameHub implements game.HubInterface for testing.
type mockGameHub struct {
	enqueueTestTasksFunc func(ctx context.Context, submissionID, gameID, userID int, language, code string) error
}

func (m *mockGameHub) CalcCodeSize(_ string, _ string) int {
	return 0
}

func (m *mockGameHub) EnqueueTestTasks(ctx context.Context, submissionID, gameID, userID int, language, code string) error {
	if m.enqueueTestTasksFunc != nil {
		return m.enqueueTestTasksFunc(ctx, submissionID, gameID, userID, language, code)
	}
	return nil
}

// mockTxManager implements db.TxManager for testing.
// By default it passes the provided querier to the function.
type mockTxManager struct {
	q           db.Querier
	runInTxFunc func(ctx context.Context, fn func(q db.Querier) error) error
}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(q db.Querier) error) error {
	if m.runInTxFunc != nil {
		return m.runInTxFunc(ctx, fn)
	}
	if m.q != nil {
		return fn(m.q)
	}
	return fn(&mockQuerier{})
}

// mockRenderer implements echo.Renderer for testing.
type mockRenderer struct {
	lastTemplateName string
	lastData         any
}

func (r *mockRenderer) Render(_ io.Writer, name string, data any, _ echo.Context) error {
	r.lastTemplateName = name
	r.lastData = data
	return nil
}

func newTestHandler(q *mockQuerier) *Handler {
	hub := &mockGameHub{}
	txm := &mockTxManager{q: q}
	gameSvc := game.NewService(q, txm, hub)
	tournamentSvc := tournament.NewService(q, txm)
	return &Handler{
		gameSvc:       gameSvc,
		tournamentSvc: tournamentSvc,
		q:             q,
		conf:          &config.Config{BasePath: "/test/"},
	}
}

func newTestHandlerWithHub(q *mockQuerier, hub *mockGameHub) *Handler {
	txm := &mockTxManager{q: q}
	gameSvc := game.NewService(q, txm, hub)
	tournamentSvc := tournament.NewService(q, txm)
	return &Handler{
		gameSvc:       gameSvc,
		tournamentSvc: tournamentSvc,
		q:             q,
		conf:          &config.Config{BasePath: "/test/"},
	}
}

func newEchoContext(method, path string, params map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Renderer = &mockRenderer{}
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if params != nil {
		names := make([]string, 0, len(params))
		values := make([]string, 0, len(params))
		for k, v := range params {
			names = append(names, k)
			values = append(values, v)
		}
		c.SetParamNames(names...)
		c.SetParamValues(values...)
	}
	return c, rec
}

func newEchoContextWithForm(path string, params map[string]string, form url.Values) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Renderer = &mockRenderer{}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if params != nil {
		names := make([]string, 0, len(params))
		values := make([]string, 0, len(params))
		for k, v := range params {
			names = append(names, k)
			values = append(values, v)
		}
		c.SetParamNames(names...)
		c.SetParamValues(values...)
	}
	return c, rec
}

// --- Admin middleware tests ---

func setUserInContext(c echo.Context, user *db.User) {
	ctx := session.SetUserInContext(c.Request().Context(), user)
	c.SetRequest(c.Request().WithContext(ctx))
}

func TestAdminMiddleware_NoUser(t *testing.T) {
	h := newTestHandler(&mockQuerier{})
	middleware := h.newAdminMiddleware()

	c, rec := newEchoContext(http.MethodGet, "/admin/dashboard", nil)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
}

func TestAdminMiddleware_NonAdminUser(t *testing.T) {
	h := newTestHandler(&mockQuerier{})
	middleware := h.newAdminMiddleware()

	c, _ := newEchoContext(http.MethodGet, "/admin/dashboard", nil)
	setUserInContext(c, &db.User{UserID: 1, IsAdmin: false})

	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for non-admin user")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusForbidden)
	}
}

func TestAdminMiddleware_AdminUser(t *testing.T) {
	h := newTestHandler(&mockQuerier{})
	middleware := h.newAdminMiddleware()

	c, rec := newEchoContext(http.MethodGet, "/admin/dashboard", nil)
	setUserInContext(c, &db.User{UserID: 1, IsAdmin: true})

	called := false
	handler := middleware(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// --- Handler tests ---

func TestGetUsers_Success(t *testing.T) {
	q := &mockQuerier{
		listUsersFunc: func(_ context.Context) ([]db.User, error) {
			return []db.User{
				{UserID: 1, Username: "alice", DisplayName: "Alice", IsAdmin: false},
				{UserID: 2, Username: "bob", DisplayName: "Bob", IsAdmin: true},
			}, nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContext(http.MethodGet, "/admin/users", nil)
	err := h.getUsers(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetUsers_DBError(t *testing.T) {
	q := &mockQuerier{
		listUsersFunc: func(_ context.Context) ([]db.User, error) {
			return nil, errors.New("db error")
		},
	}
	h := newTestHandler(q)

	c, _ := newEchoContext(http.MethodGet, "/admin/users", nil)
	err := h.getUsers(c)
	if err == nil {
		t.Fatal("expected error")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusInternalServerError)
	}
}

func TestGetUserEdit_Success(t *testing.T) {
	q := &mockQuerier{
		getUserByIDFunc: func(_ context.Context, userID int32) (db.User, error) {
			return db.User{UserID: userID, Username: "alice", DisplayName: "Alice"}, nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContext(http.MethodGet, "/admin/users/1", map[string]string{"userID": "1"})
	err := h.getUserEdit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetUserEdit_InvalidID(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	c, _ := newEchoContext(http.MethodGet, "/admin/users/abc", map[string]string{"userID": "abc"})
	err := h.getUserEdit(c)
	if err == nil {
		t.Fatal("expected error for invalid userID")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusBadRequest)
	}
}

func TestGetUserEdit_NotFound(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	c, _ := newEchoContext(http.MethodGet, "/admin/users/999", map[string]string{"userID": "999"})
	err := h.getUserEdit(c)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

func TestPostUserEdit_Success(t *testing.T) {
	var updatedParams db.UpdateUserParams
	q := &mockQuerier{
		updateUserFunc: func(_ context.Context, arg db.UpdateUserParams) error {
			updatedParams = arg
			return nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"display_name": {"Alice Updated"},
		"icon_path":    {"files/img/alice/icon.png"},
		"is_admin":     {"on"},
		"label":        {"sponsor"},
	}
	c, rec := newEchoContextWithForm("/admin/users/1", map[string]string{"userID": "1"}, form)

	err := h.postUserEdit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if updatedParams.UserID != 1 {
		t.Errorf("UserID = %d, want 1", updatedParams.UserID)
	}
	if updatedParams.DisplayName != "Alice Updated" {
		t.Errorf("DisplayName = %q, want %q", updatedParams.DisplayName, "Alice Updated")
	}
	if !updatedParams.IsAdmin {
		t.Error("IsAdmin should be true")
	}
}

func TestPostUserEdit_EmptyOptionalFields(t *testing.T) {
	var updatedParams db.UpdateUserParams
	q := &mockQuerier{
		updateUserFunc: func(_ context.Context, arg db.UpdateUserParams) error {
			updatedParams = arg
			return nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"display_name": {"Bob"},
	}
	c, _ := newEchoContextWithForm("/admin/users/2", map[string]string{"userID": "2"}, form)

	err := h.postUserEdit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedParams.IconPath != nil {
		t.Errorf("IconPath should be nil for empty value, got %v", updatedParams.IconPath)
	}
	if updatedParams.IsAdmin {
		t.Error("IsAdmin should be false when not in form")
	}
	if updatedParams.Label != nil {
		t.Errorf("Label should be nil for empty value, got %v", updatedParams.Label)
	}
}

func TestGetGames_Success(t *testing.T) {
	q := &mockQuerier{
		listAllGamesFunc: func(_ context.Context) ([]db.Game, error) {
			return []db.Game{
				{GameID: 1, GameType: "golf", DisplayName: "Game 1", DurationSeconds: 300, ProblemID: 1},
			}, nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContext(http.MethodGet, "/admin/games", nil)
	err := h.getGames(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestPostGameNew_Success(t *testing.T) {
	var createdParams db.CreateGameParams
	q := &mockQuerier{
		createGameFunc: func(_ context.Context, arg db.CreateGameParams) (int32, error) {
			createdParams = arg
			return 1, nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"game_type":        {"golf"},
		"is_public":        {"on"},
		"display_name":     {"Test Game"},
		"duration_seconds": {"300"},
		"problem_id":       {"1"},
	}
	c, rec := newEchoContextWithForm("/admin/games/new", nil, form)

	err := h.postGameNew(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if createdParams.GameType != "golf" {
		t.Errorf("GameType = %q, want %q", createdParams.GameType, "golf")
	}
	if !createdParams.IsPublic {
		t.Error("IsPublic should be true")
	}
	if createdParams.DurationSeconds != 300 {
		t.Errorf("DurationSeconds = %d, want 300", createdParams.DurationSeconds)
	}
	if createdParams.ProblemID != 1 {
		t.Errorf("ProblemID = %d, want 1", createdParams.ProblemID)
	}
}

func TestPostGameNew_InvalidDuration(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	form := url.Values{
		"game_type":        {"golf"},
		"display_name":     {"Test Game"},
		"duration_seconds": {"invalid"},
		"problem_id":       {"1"},
	}
	c, _ := newEchoContextWithForm("/admin/games/new", nil, form)

	err := h.postGameNew(c)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusBadRequest)
	}
}

func TestPostGameNew_InvalidProblemID(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	form := url.Values{
		"game_type":        {"golf"},
		"display_name":     {"Test Game"},
		"duration_seconds": {"300"},
		"problem_id":       {"invalid"},
	}
	c, _ := newEchoContextWithForm("/admin/games/new", nil, form)

	err := h.postGameNew(c)
	if err == nil {
		t.Fatal("expected error for invalid problem_id")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusBadRequest)
	}
}

func TestGetProblems_Success(t *testing.T) {
	q := &mockQuerier{
		listProblemsFunc: func(_ context.Context) ([]db.Problem, error) {
			return []db.Problem{
				{ProblemID: 1, Title: "Hello World", Description: "Print hello", Language: "php"},
			}, nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContext(http.MethodGet, "/admin/problems", nil)
	err := h.getProblems(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestPostProblemNew_Success(t *testing.T) {
	var createdParams db.CreateProblemParams
	q := &mockQuerier{
		createProblemFunc: func(_ context.Context, arg db.CreateProblemParams) (int32, error) {
			createdParams = arg
			return 1, nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"title":       {"FizzBuzz"},
		"description": {"Write FizzBuzz"},
		"language":    {"php"},
		"sample_code": {"<?php echo 1;"},
	}
	c, rec := newEchoContextWithForm("/admin/problems/new", nil, form)

	err := h.postProblemNew(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if createdParams.Title != "FizzBuzz" {
		t.Errorf("Title = %q, want %q", createdParams.Title, "FizzBuzz")
	}
	if createdParams.Language != "php" {
		t.Errorf("Language = %q, want %q", createdParams.Language, "php")
	}
}

func TestGetProblemEdit_Success(t *testing.T) {
	q := &mockQuerier{
		getProblemByIDFunc: func(_ context.Context, problemID int32) (db.Problem, error) {
			return db.Problem{ProblemID: problemID, Title: "Test", Language: "php"}, nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContext(http.MethodGet, "/admin/problems/1", map[string]string{"problemID": "1"})
	err := h.getProblemEdit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetProblemEdit_NotFound(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	c, _ := newEchoContext(http.MethodGet, "/admin/problems/999", map[string]string{"problemID": "999"})
	err := h.getProblemEdit(c)
	if err == nil {
		t.Fatal("expected error for non-existent problem")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

func TestPostProblemEdit_Success(t *testing.T) {
	var updatedParams db.UpdateProblemParams
	q := &mockQuerier{
		updateProblemFunc: func(_ context.Context, arg db.UpdateProblemParams) error {
			updatedParams = arg
			return nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"title":       {"Updated Title"},
		"description": {"Updated Desc"},
		"language":    {"php"},
		"sample_code": {"<?php echo 2;"},
	}
	c, rec := newEchoContextWithForm("/admin/problems/1", map[string]string{"problemID": "1"}, form)

	err := h.postProblemEdit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if updatedParams.ProblemID != 1 {
		t.Errorf("ProblemID = %d, want 1", updatedParams.ProblemID)
	}
	if updatedParams.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", updatedParams.Title, "Updated Title")
	}
}

func TestGetTestcases_Success(t *testing.T) {
	q := &mockQuerier{
		getProblemByIDFunc: func(_ context.Context, problemID int32) (db.Problem, error) {
			return db.Problem{ProblemID: problemID, Title: "Test Problem"}, nil
		},
		listTestcasesByProblemIDFunc: func(_ context.Context, _ int32) ([]db.Testcase, error) {
			return []db.Testcase{
				{TestcaseID: 1, ProblemID: 1, Stdin: "in", Stdout: "out"},
			}, nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContext(http.MethodGet, "/admin/problems/1/testcases", map[string]string{"problemID": "1"})
	err := h.getTestcases(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestPostTestcaseNew_Success(t *testing.T) {
	var createdParams db.CreateTestcaseParams
	q := &mockQuerier{
		createTestcaseFunc: func(_ context.Context, arg db.CreateTestcaseParams) (int32, error) {
			createdParams = arg
			return 1, nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"stdin":  {"hello"},
		"stdout": {"world"},
	}
	c, rec := newEchoContextWithForm("/admin/problems/1/testcases/new", map[string]string{"problemID": "1"}, form)

	err := h.postTestcaseNew(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if createdParams.ProblemID != 1 {
		t.Errorf("ProblemID = %d, want 1", createdParams.ProblemID)
	}
	if createdParams.Stdin != "hello" {
		t.Errorf("Stdin = %q, want %q", createdParams.Stdin, "hello")
	}
}

func TestPostTestcaseEdit_Success(t *testing.T) {
	q := &mockQuerier{
		getTestcaseByIDFunc: func(_ context.Context, testcaseID int32) (db.Testcase, error) {
			return db.Testcase{TestcaseID: testcaseID, ProblemID: 1}, nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"stdin":  {"updated_in"},
		"stdout": {"updated_out"},
	}
	c, rec := newEchoContextWithForm("/admin/problems/1/testcases/1", map[string]string{
		"problemID":  "1",
		"testcaseID": "1",
	}, form)

	err := h.postTestcaseEdit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
}

func TestPostTestcaseEdit_WrongProblem(t *testing.T) {
	q := &mockQuerier{
		getTestcaseByIDFunc: func(_ context.Context, testcaseID int32) (db.Testcase, error) {
			return db.Testcase{TestcaseID: testcaseID, ProblemID: 2}, nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"stdin":  {"in"},
		"stdout": {"out"},
	}
	c, _ := newEchoContextWithForm("/admin/problems/1/testcases/1", map[string]string{
		"problemID":  "1",
		"testcaseID": "1",
	}, form)

	err := h.postTestcaseEdit(c)
	if err == nil {
		t.Fatal("expected error when testcase belongs to different problem")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

func TestPostTestcaseDelete_Success(t *testing.T) {
	deletedID := int32(0)
	q := &mockQuerier{
		getTestcaseByIDFunc: func(_ context.Context, testcaseID int32) (db.Testcase, error) {
			return db.Testcase{TestcaseID: testcaseID, ProblemID: 1}, nil
		},
		deleteTestcaseFunc: func(_ context.Context, testcaseID int32) error {
			deletedID = testcaseID
			return nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContextWithForm("/admin/problems/1/testcases/5/delete", map[string]string{
		"problemID":  "1",
		"testcaseID": "5",
	}, url.Values{})

	err := h.postTestcaseDelete(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if deletedID != 5 {
		t.Errorf("deleted testcase ID = %d, want 5", deletedID)
	}
}

func TestPostTestcaseDelete_WrongProblem(t *testing.T) {
	q := &mockQuerier{
		getTestcaseByIDFunc: func(_ context.Context, testcaseID int32) (db.Testcase, error) {
			return db.Testcase{TestcaseID: testcaseID, ProblemID: 99}, nil
		},
	}
	h := newTestHandler(q)

	c, _ := newEchoContextWithForm("/admin/problems/1/testcases/5/delete", map[string]string{
		"problemID":  "1",
		"testcaseID": "5",
	}, url.Values{})

	err := h.postTestcaseDelete(c)
	if err == nil {
		t.Fatal("expected error when testcase belongs to different problem")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

func TestPostGameStart_Success(t *testing.T) {
	q := &mockQuerier{
		getGameByIDFunc: func(_ context.Context, gameID int32) (db.GetGameByIDRow, error) {
			return db.GetGameByIDRow{GameID: gameID, ProblemID: 1}, nil
		},
		listTestcasesByProblemIDFunc: func(_ context.Context, _ int32) ([]db.Testcase, error) {
			return []db.Testcase{{TestcaseID: 1, ProblemID: 1}}, nil
		},
		updateGameStartedAtFunc: func(_ context.Context, _ db.UpdateGameStartedAtParams) error {
			return nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContextWithForm("/admin/games/1/start", map[string]string{"gameID": "1"}, url.Values{})

	err := h.postGameStart(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
}

func TestPostGameStart_GameNotFound(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	c, _ := newEchoContextWithForm("/admin/games/999/start", map[string]string{"gameID": "999"}, url.Values{})

	err := h.postGameStart(c)
	if err == nil {
		t.Fatal("expected error for non-existent game")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

func TestPostGameStart_NoTestcases(t *testing.T) {
	q := &mockQuerier{
		getGameByIDFunc: func(_ context.Context, gameID int32) (db.GetGameByIDRow, error) {
			return db.GetGameByIDRow{GameID: gameID, ProblemID: 1}, nil
		},
		listTestcasesByProblemIDFunc: func(_ context.Context, _ int32) ([]db.Testcase, error) {
			return []db.Testcase{}, nil
		},
	}
	h := newTestHandler(q)

	c, _ := newEchoContextWithForm("/admin/games/1/start", map[string]string{"gameID": "1"}, url.Values{})

	err := h.postGameStart(c)
	if err == nil {
		t.Fatal("expected error when no testcases")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusBadRequest)
	}
}

func TestGetSubmissions_Success(t *testing.T) {
	q := &mockQuerier{
		getSubmissionsByGameIDFunc: func(_ context.Context, _ int32) ([]db.Submission, error) {
			return []db.Submission{
				{
					SubmissionID: 1,
					GameID:       1,
					UserID:       1,
					Status:       "pass",
					CodeSize:     42,
					CreatedAt:    pgtype.Timestamp{Valid: true},
				},
			}, nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContext(http.MethodGet, "/admin/games/1/submissions", map[string]string{"gameID": "1"})
	err := h.getSubmissions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetSubmissionDetail_Success(t *testing.T) {
	q := &mockQuerier{
		getSubmissionByIDFunc: func(_ context.Context, submissionID int32) (db.Submission, error) {
			return db.Submission{
				SubmissionID: submissionID,
				GameID:       1,
				UserID:       1,
				Code:         "<?php echo 1;",
				CodeSize:     14,
				Status:       "pass",
				CreatedAt:    pgtype.Timestamp{Valid: true},
			}, nil
		},
		getTestcaseResultsBySubmIDFunc: func(_ context.Context, _ int32) ([]db.TestcaseResult, error) {
			return []db.TestcaseResult{
				{
					TestcaseResultID: 1,
					SubmissionID:     1,
					TestcaseID:       1,
					Status:           "pass",
					CreatedAt:        pgtype.Timestamp{Valid: true},
				},
			}, nil
		},
	}
	h := newTestHandler(q)

	c, rec := newEchoContext(http.MethodGet, "/admin/games/1/submissions/1", map[string]string{
		"gameID":       "1",
		"submissionID": "1",
	})
	err := h.getSubmissionDetail(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetSubmissionDetail_NotFound(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	c, _ := newEchoContext(http.MethodGet, "/admin/games/1/submissions/999", map[string]string{
		"gameID":       "1",
		"submissionID": "999",
	})
	err := h.getSubmissionDetail(c)
	if err == nil {
		t.Fatal("expected error for non-existent submission")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

// --- Tournament admin tests ---

func TestGetTournaments_Empty(t *testing.T) {
	h := newTestHandler(&mockQuerier{})
	c, rec := newEchoContext(http.MethodGet, "/admin/tournaments", nil)
	err := h.getTournaments(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetTournaments_WithData(t *testing.T) {
	h := newTestHandler(&mockQuerier{
		listTournamentsFunc: func(_ context.Context) ([]db.Tournament, error) {
			return []db.Tournament{
				{TournamentID: 1, DisplayName: "Tournament A", BracketSize: 4},
				{TournamentID: 2, DisplayName: "Tournament B", BracketSize: 8},
			}, nil
		},
	})
	c, _ := newEchoContext(http.MethodGet, "/admin/tournaments", nil)
	err := h.getTournaments(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTournamentNew(t *testing.T) {
	h := newTestHandler(&mockQuerier{})
	c, rec := newEchoContext(http.MethodGet, "/admin/tournaments/new", nil)
	err := h.getTournamentNew(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestPostTournamentNew_Success(t *testing.T) {
	var createdParams db.CreateTournamentParams
	var matchCount int
	q := &mockQuerier{
		createTournamentFunc: func(_ context.Context, arg db.CreateTournamentParams) (int32, error) {
			createdParams = arg
			return 1, nil
		},
		createTournamentMatchFunc: func(_ context.Context, _ db.CreateTournamentMatchParams) error {
			matchCount++
			return nil
		},
	}
	h := newTestHandler(q)

	form := url.Values{
		"display_name":     {"Test Tournament"},
		"num_participants": {"3"},
	}
	c, rec := newEchoContextWithForm("/admin/tournaments/new", nil, form)
	err := h.postTournamentNew(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if createdParams.DisplayName != "Test Tournament" {
		t.Errorf("display_name = %q, want %q", createdParams.DisplayName, "Test Tournament")
	}
	if createdParams.BracketSize != 4 {
		t.Errorf("bracket_size = %d, want 4", createdParams.BracketSize)
	}
	if createdParams.NumRounds != 2 {
		t.Errorf("num_rounds = %d, want 2", createdParams.NumRounds)
	}
	// bracket_size=4, num_rounds=2 → round 0: 2 matches + round 1: 1 match = 3 matches
	if matchCount != 3 {
		t.Errorf("match count = %d, want 3", matchCount)
	}
}

func TestPostTournamentNew_InvalidParticipants(t *testing.T) {
	h := newTestHandler(&mockQuerier{})
	form := url.Values{
		"display_name":     {"Test"},
		"num_participants": {"abc"},
	}
	c, _ := newEchoContextWithForm("/admin/tournaments/new", nil, form)
	err := h.postTournamentNew(c)
	if err == nil {
		t.Fatal("expected error for invalid num_participants")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusBadRequest)
	}
}

func TestGetTournamentEdit_NotFound(t *testing.T) {
	h := newTestHandler(&mockQuerier{})
	c, _ := newEchoContext(http.MethodGet, "/admin/tournaments/999", map[string]string{
		"tournamentID": "999",
	})
	err := h.getTournamentEdit(c)
	if err == nil {
		t.Fatal("expected error for non-existent tournament")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

func TestGetTournamentEdit_Success(t *testing.T) {
	h := newTestHandler(&mockQuerier{
		getTournamentByIDFunc: func(_ context.Context, _ int32) (db.Tournament, error) {
			return db.Tournament{
				TournamentID: 1,
				DisplayName:  "Test",
				BracketSize:  4,
				NumRounds:    2,
			}, nil
		},
		listTournamentMatchesFunc: func(_ context.Context, _ int32) ([]db.TournamentMatch, error) {
			return []db.TournamentMatch{
				{TournamentMatchID: 1, Round: 0, Position: 0},
			}, nil
		},
	})
	c, rec := newEchoContext(http.MethodGet, "/admin/tournaments/1", map[string]string{
		"tournamentID": "1",
	})
	err := h.getTournamentEdit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestPostTournamentEdit_NotFound(t *testing.T) {
	h := newTestHandler(&mockQuerier{})
	form := url.Values{
		"display_name": {"Updated"},
	}
	c, _ := newEchoContextWithForm("/admin/tournaments/999", map[string]string{
		"tournamentID": "999",
	}, form)
	err := h.postTournamentEdit(c)
	if err == nil {
		t.Fatal("expected error for non-existent tournament")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

// --- Rejudge tests ---

func TestPostSubmissionRejudge_Success(t *testing.T) {
	var enqueuedSubmissionID, enqueuedGameID, enqueuedUserID int
	var enqueuedLanguage, enqueuedCode string

	q := &mockQuerier{
		getSubmissionByIDFunc: func(_ context.Context, submissionID int32) (db.Submission, error) {
			return db.Submission{
				SubmissionID: submissionID,
				GameID:       1,
				UserID:       10,
				Code:         "<?php echo 1;",
				CodeSize:     14,
				Status:       "wrong_answer",
				CreatedAt:    pgtype.Timestamp{Valid: true},
			}, nil
		},
		getGameByIDFunc: func(_ context.Context, gameID int32) (db.GetGameByIDRow, error) {
			return db.GetGameByIDRow{GameID: gameID, ProblemID: 1, Language: "php"}, nil
		},
	}

	hub := &mockGameHub{
		enqueueTestTasksFunc: func(_ context.Context, submissionID, gameID, userID int, language, code string) error {
			enqueuedSubmissionID = submissionID
			enqueuedGameID = gameID
			enqueuedUserID = userID
			enqueuedLanguage = language
			enqueuedCode = code
			return nil
		},
	}

	h := newTestHandlerWithHub(q, hub)

	c, rec := newEchoContextWithForm("/admin/games/1/submissions/5/rejudge", map[string]string{
		"gameID":       "1",
		"submissionID": "5",
	}, url.Values{})

	err := h.postSubmissionRejudge(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if enqueuedSubmissionID != 5 {
		t.Errorf("enqueued submission ID = %d, want 5", enqueuedSubmissionID)
	}
	if enqueuedGameID != 1 {
		t.Errorf("enqueued game ID = %d, want 1", enqueuedGameID)
	}
	if enqueuedUserID != 10 {
		t.Errorf("enqueued user ID = %d, want 10", enqueuedUserID)
	}
	if enqueuedLanguage != "php" {
		t.Errorf("enqueued language = %q, want %q", enqueuedLanguage, "php")
	}
	if enqueuedCode != "<?php echo 1;" {
		t.Errorf("enqueued code = %q, want %q", enqueuedCode, "<?php echo 1;")
	}
}

func TestPostSubmissionRejudge_SubmissionNotFound(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	c, _ := newEchoContextWithForm("/admin/games/1/submissions/999/rejudge", map[string]string{
		"gameID":       "1",
		"submissionID": "999",
	}, url.Values{})

	err := h.postSubmissionRejudge(c)
	if err == nil {
		t.Fatal("expected error for non-existent submission")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

func TestPostSubmissionsRejudgeLatest_Success(t *testing.T) {
	var enqueuedIDs []int

	q := &mockQuerier{
		getGameByIDFunc: func(_ context.Context, gameID int32) (db.GetGameByIDRow, error) {
			return db.GetGameByIDRow{GameID: gameID, ProblemID: 1, Language: "php"}, nil
		},
		getLatestSubmissionsByGameIDFunc: func(_ context.Context, _ int32) ([]db.Submission, error) {
			return []db.Submission{
				{SubmissionID: 10, GameID: 1, UserID: 1, Code: "<?php echo 1;", CreatedAt: pgtype.Timestamp{Valid: true}},
				{SubmissionID: 20, GameID: 1, UserID: 2, Code: "<?php echo 2;", CreatedAt: pgtype.Timestamp{Valid: true}},
			}, nil
		},
	}

	hub := &mockGameHub{
		enqueueTestTasksFunc: func(_ context.Context, submissionID, _, _ int, _, _ string) error {
			enqueuedIDs = append(enqueuedIDs, submissionID)
			return nil
		},
	}

	h := newTestHandlerWithHub(q, hub)

	c, rec := newEchoContextWithForm("/admin/games/1/submissions/rejudge-latest", map[string]string{
		"gameID": "1",
	}, url.Values{})

	err := h.postSubmissionsRejudgeLatest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if len(enqueuedIDs) != 2 {
		t.Fatalf("enqueued count = %d, want 2", len(enqueuedIDs))
	}
	if enqueuedIDs[0] != 10 || enqueuedIDs[1] != 20 {
		t.Errorf("enqueued IDs = %v, want [10, 20]", enqueuedIDs)
	}
}

func TestPostSubmissionsRejudgeAll_Success(t *testing.T) {
	var enqueuedIDs []int

	q := &mockQuerier{
		getGameByIDFunc: func(_ context.Context, gameID int32) (db.GetGameByIDRow, error) {
			return db.GetGameByIDRow{GameID: gameID, ProblemID: 1, Language: "php"}, nil
		},
		getSubmissionsByGameIDFunc: func(_ context.Context, _ int32) ([]db.Submission, error) {
			return []db.Submission{
				{SubmissionID: 10, GameID: 1, UserID: 1, Code: "<?php echo 1;", CreatedAt: pgtype.Timestamp{Valid: true}},
				{SubmissionID: 11, GameID: 1, UserID: 1, Code: "<?php echo 11;", CreatedAt: pgtype.Timestamp{Valid: true}},
				{SubmissionID: 20, GameID: 1, UserID: 2, Code: "<?php echo 2;", CreatedAt: pgtype.Timestamp{Valid: true}},
			}, nil
		},
	}

	hub := &mockGameHub{
		enqueueTestTasksFunc: func(_ context.Context, submissionID, _, _ int, _, _ string) error {
			enqueuedIDs = append(enqueuedIDs, submissionID)
			return nil
		},
	}

	h := newTestHandlerWithHub(q, hub)

	c, rec := newEchoContextWithForm("/admin/games/1/submissions/rejudge-all", map[string]string{
		"gameID": "1",
	}, url.Values{})

	err := h.postSubmissionsRejudgeAll(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if len(enqueuedIDs) != 3 {
		t.Fatalf("enqueued count = %d, want 3", len(enqueuedIDs))
	}
	if enqueuedIDs[0] != 10 || enqueuedIDs[1] != 11 || enqueuedIDs[2] != 20 {
		t.Errorf("enqueued IDs = %v, want [10, 11, 20]", enqueuedIDs)
	}
}

func TestPostSubmissionsRejudgeAll_GameNotFound(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	c, _ := newEchoContextWithForm("/admin/games/999/submissions/rejudge-all", map[string]string{
		"gameID": "999",
	}, url.Values{})

	err := h.postSubmissionsRejudgeAll(c)
	if err == nil {
		t.Fatal("expected error for non-existent game")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}
