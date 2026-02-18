using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace AlbatrossApi;

// ---------- Error ----------

model Error {
  message: string;
}

// ---------- Error Responses ----------

@error
model UnauthorizedError {
  @statusCode statusCode: 401;
  @body body: Error;
}

@error
model ForbiddenError {
  @statusCode statusCode: 403;
  @body body: Error;
}

@error
model NotFoundError {
  @statusCode statusCode: 404;
  @body body: Error;
}

// ---------- Enums ----------

enum GameType {
  `1v1`,
  multiplayer,
}

enum ProblemLanguage {
  php,
  swift,
}

enum ExecutionStatus {
  none,
  running,
  success,
  wrong_answer,
  timeout,
  compile_error,
  runtime_error,
  internal_error,
}

// ---------- Models ----------

model User {
  user_id: integer;
  username: string;
  display_name: string;
  icon_path?: string;
  is_admin: boolean;
  label: string | null;
}

model Problem {
  problem_id: integer;
  title: string;
  description: string;
  language: ProblemLanguage;
  sample_code: string;
}

model Game {
  game_id: integer;
  game_type: GameType;
  is_public: boolean;
  display_name: string;
  duration_seconds: integer;

  @extension("x-go-type", "int64")
  started_at?: integer;

  problem: Problem;
  main_players: User[];
}

model LatestGameState {
  code: string;
  score: integer | null;

  @extension("x-go-type", "int64")
  best_score_submitted_at: integer | null;

  status: ExecutionStatus;
}

model RankingEntry {
  player: User;
  score: integer;

  @extension("x-go-type", "int64")
  submitted_at: integer;

  code: string | null;
}

model Tournament {
  tournament_id: integer;
  display_name: string;
  bracket_size: integer;
  num_rounds: integer;
  entries: TournamentEntry[];
  matches: TournamentMatch[];
}

model TournamentEntry {
  user: User;
  seed: integer;
}

model TournamentMatch {
  tournament_match_id: integer;
  round: integer;
  position: integer;
  game_id?: integer;
  player1?: User;
  player2?: User;
  player1_score?: integer;
  player2_score?: integer;
  winner_user_id?: integer;
  is_bye: boolean;
}
