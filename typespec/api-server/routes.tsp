using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace AlbatrossApi;

// ---------- Auth ----------

@route("/login")
@post
@operationId("postLogin")
op postLogin(@body body: {
  username: string;
  password: string;
}): {
  @body body: {
    user: User;
  };
} | UnauthorizedError;

@route("/logout")
@post
@operationId("postLogout")
op postLogout(): {
  @statusCode statusCode: 200;
} | UnauthorizedError;

@route("/me")
@get
@operationId("getMe")
op getMe(): {
  @body body: {
    user: User;
  };
} | UnauthorizedError;

// ---------- Games ----------

@route("/games")
@get
@operationId("getGames")
op getGames(): {
  @body body: {
    games: Game[];
  };
} | UnauthorizedError | ForbiddenError;

@route("/games/{game_id}")
@get
@operationId("getGame")
op getGame(@path game_id: integer): {
  @body body: {
    game: Game;
  };
} | UnauthorizedError | ForbiddenError | NotFoundError;

// ---------- Play ----------

@route("/games/{game_id}/play/latest_state")
@get
@operationId("getGamePlayLatestState")
op getGamePlayLatestState(@path game_id: integer): {
  @body body: {
    state: LatestGameState;
  };
} | UnauthorizedError | ForbiddenError | NotFoundError;

@route("/games/{game_id}/play/code")
@post
@operationId("postGamePlayCode")
op postGamePlayCode(
  @path game_id: integer,
  @body body: {
    code: string;
  },
): {
  @statusCode statusCode: 200;
} | UnauthorizedError | ForbiddenError | NotFoundError;

@route("/games/{game_id}/play/submit")
@post
@operationId("postGamePlaySubmit")
op postGamePlaySubmit(
  @path game_id: integer,
  @body body: {
    code: string;
  },
): {
  @statusCode statusCode: 200;
} | UnauthorizedError | ForbiddenError | NotFoundError;

// ---------- Watch ----------

@route("/games/{game_id}/watch/ranking")
@get
@operationId("getGameWatchRanking")
op getGameWatchRanking(@path game_id: integer): {
  @body body: {
    ranking: RankingEntry[];
  };
} | UnauthorizedError | ForbiddenError | NotFoundError;

@route("/games/{game_id}/watch/latest_states")
@get
@operationId("getGameWatchLatestStates")
op getGameWatchLatestStates(@path game_id: integer): {
  @body body: {
    states: Record<LatestGameState>;
  };
} | UnauthorizedError | ForbiddenError | NotFoundError;

// ---------- Tournament ----------

@route("/tournaments/{tournament_id}")
@get
@operationId("getTournament")
op getTournament(@path tournament_id: integer): {
  @body body: {
    tournament: Tournament;
  };
} | UnauthorizedError | ForbiddenError | NotFoundError;
