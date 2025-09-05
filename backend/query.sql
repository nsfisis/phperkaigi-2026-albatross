-- name: GetUserByID :one
SELECT * FROM users
WHERE users.user_id = $1
LIMIT 1;

-- name: GetUserIDByUsername :one
SELECT user_id FROM users
WHERE users.username = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, display_name, is_admin)
VALUES ($1, $1, false)
RETURNING user_id;

-- name: UpdateUserIconPath :exec
UPDATE users
SET icon_path = $2
WHERE user_id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY users.user_id;

-- name: GetUserAuthByUsername :one
SELECT * FROM users
JOIN user_auths ON users.user_id = user_auths.user_id
WHERE users.username = $1
LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
SET
    display_name = $2,
    icon_path = $3,
    is_admin = $4,
    label = $5
WHERE user_id = $1;

-- name: CreateUserAuth :exec
INSERT INTO user_auths (user_id, auth_type)
VALUES ($1, $2);

-- name: ListPublicGames :many
SELECT * FROM games
JOIN problems ON games.problem_id = problems.problem_id
WHERE is_public = true
ORDER BY games.game_id;

-- name: ListAllGames :many
SELECT * FROM games
ORDER BY games.game_id;

-- name: UpdateGameStartedAt :exec
UPDATE games
SET started_at = $2
WHERE game_id = $1;

-- name: GetGameByID :one
SELECT * FROM games
JOIN problems ON games.problem_id = problems.problem_id
WHERE games.game_id = $1
LIMIT 1;

-- name: UpdateGame :exec
UPDATE games
SET
    game_type = $2,
    is_public = $3,
    display_name = $4,
    duration_seconds = $5,
    started_at = $6,
    problem_id = $7
WHERE game_id = $1;

-- name: ListMainPlayers :many
SELECT * FROM game_main_players
JOIN users ON game_main_players.user_id = users.user_id
WHERE game_main_players.game_id = ANY($1::INT[])
ORDER BY game_main_players.user_id;

-- name: AddMainPlayer :exec
INSERT INTO game_main_players (game_id, user_id)
VALUES ($1, $2);

-- name: RemoveAllMainPlayers :exec
DELETE FROM game_main_players
WHERE game_id = $1;

-- name: ListTestcasesByGameID :many
SELECT * FROM testcases
WHERE testcases.problem_id = (SELECT problem_id FROM games WHERE game_id = $1)
ORDER BY testcases.testcase_id;

-- name: CreateTestcaseResult :exec
INSERT INTO testcase_results (submission_id, testcase_id, status, stdout, stderr)
VALUES ($1, $2, $3, $4, $5);

-- name: AggregateTestcaseResults :one
SELECT
    CASE
        WHEN COUNT(*) < (SELECT COUNT(*) FROM testcases WHERE problem_id =
                         (SELECT problem_id FROM games WHERE game_id =
                          (SELECT game_id FROM submissions AS s WHERE s.submission_id = $1)))
        THEN 'running'
        WHEN COUNT(CASE WHEN r.status = 'internal_error' THEN 1 END) > 0 THEN 'internal_error'
        WHEN COUNT(CASE WHEN r.status = 'timeout'        THEN 1 END) > 0 THEN 'timeout'
        WHEN COUNT(CASE WHEN r.status = 'runtime_error'  THEN 1 END) > 0 THEN 'runtime_error'
        WHEN COUNT(CASE WHEN r.status = 'wrong_answer'   THEN 1 END) > 0 THEN 'wrong_answer'
        ELSE 'success'
    END AS status
FROM testcase_results AS r
WHERE r.submission_id = $1;

-- name: GetLatestState :one
SELECT * FROM game_states
LEFT JOIN submissions ON game_states.best_score_submission_id = submissions.submission_id
WHERE game_states.game_id = $1 AND game_states.user_id = $2
LIMIT 1;

-- name: GetLatestStatesOfMainPlayers :many
SELECT * FROM game_main_players
LEFT JOIN game_states ON game_main_players.game_id = game_states.game_id AND game_main_players.user_id = game_states.user_id
LEFT JOIN submissions ON game_states.best_score_submission_id = submissions.submission_id
WHERE game_main_players.game_id = $1;

-- name: GetRanking :many
SELECT
    sqlc.embed(submissions),
    sqlc.embed(users)
FROM game_states
JOIN users ON game_states.user_id = users.user_id
JOIN submissions ON game_states.best_score_submission_id = submissions.submission_id
WHERE game_states.game_id = $1
ORDER BY submissions.code_size ASC, submissions.created_at ASC
LIMIT 30;

-- name: GetQualifyingRanking :many
SELECT
    u.username AS username,
    u.label AS user_label,
    s1.code_size AS code_size_1,
    s2.code_size AS code_size_2,
    (s1.code_size + s2.code_size) AS total_code_size,
    s1.created_at AS submitted_at_1,
    s2.created_at AS submitted_at_2
FROM game_states gs1
JOIN submissions s1 ON gs1.best_score_submission_id = s1.submission_id
JOIN game_states gs2 ON gs1.user_id = gs2.user_id
JOIN submissions s2 ON gs2.best_score_submission_id = s2.submission_id
JOIN users u ON gs1.user_id = u.user_id
WHERE gs1.game_id = $1 AND gs2.game_id = $2
ORDER BY total_code_size ASC;

-- name: UpdateCode :exec
INSERT INTO game_states (game_id, user_id, code, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (game_id, user_id)
DO UPDATE SET code = EXCLUDED.code;

-- name: UpdateCodeAndStatus :exec
INSERT INTO game_states (game_id, user_id, code, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (game_id, user_id)
DO UPDATE SET code = EXCLUDED.code, status = EXCLUDED.status;

-- name: CreateSubmission :one
INSERT INTO submissions (game_id, user_id, code, code_size, status)
VALUES ($1, $2, $3, $4, 'running')
RETURNING submission_id;

-- name: UpdateSubmissionStatus :exec
UPDATE submissions
SET status = $2
WHERE submission_id = $1;

-- name: UpdateGameStateStatus :exec
UPDATE game_states
SET status = $3
WHERE game_id = $1 AND user_id = $2;

-- name: SyncGameStateBestScoreSubmission :exec
UPDATE game_states
SET best_score_submission_id = (
    SELECT submission_id FROM submissions AS s
    WHERE s.game_id = $1 AND s.user_id = $2 AND s.status = 'success'
    ORDER BY s.code_size ASC, s.created_at ASC
    LIMIT 1
)
WHERE game_id = $1 AND user_id = $2;

-- name: ListSubmissionIDs :many
SELECT submission_id FROM submissions;

-- name: ListGameStateIDs :many
SELECT game_id, user_id FROM game_states;

-- name: ListProblems :many
SELECT * FROM problems
ORDER BY problem_id;

-- name: GetProblemByID :one
SELECT * FROM problems
WHERE problem_id = $1
LIMIT 1;

-- name: CreateProblem :one
INSERT INTO problems (title, description, language, sample_code)
VALUES ($1, $2, $3, $4)
RETURNING problem_id;

-- name: UpdateProblem :exec
UPDATE problems
SET
    title = $2,
    description = $3,
    language = $4,
    sample_code = $5
WHERE problem_id = $1;
