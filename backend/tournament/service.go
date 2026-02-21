package tournament

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"albatross-2026-backend/db"
	"albatross-2026-backend/game"
)

type Service struct {
	q   db.Querier
	txm db.TxManager
}

func NewService(q db.Querier, txm db.TxManager) *Service {
	return &Service{q: q, txm: txm}
}

func nextPowerOf2(n int) int {
	p := 1
	for p < n {
		p *= 2
	}
	return p
}

func log2Int(n int) int {
	r := 0
	for n > 1 {
		n /= 2
		r++
	}
	return r
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

type Entry struct {
	User Player
	Seed int
}

type Match struct {
	TournamentMatchID int
	Round             int
	Position          int
	GameID            *int
	Player1           *Player
	Player2           *Player
	Player1Score      *int
	Player2Score      *int
	WinnerUserID      *int
	IsBye             bool
}

type Bracket struct {
	TournamentID int
	DisplayName  string
	BracketSize  int
	NumRounds    int
	Entries      []Entry
	Matches      []Match
}

// StandardBracketSeeds returns the seed assignments for each slot in a standard
// single-elimination bracket.
func StandardBracketSeeds(bracketSize int) []int {
	seeds := make([]int, bracketSize)
	seeds[0] = 1
	for size := 2; size <= bracketSize; size *= 2 {
		temp := make([]int, size)
		for i := 0; i < size/2; i++ {
			temp[i*2] = seeds[i]
			temp[i*2+1] = size + 1 - seeds[i]
		}
		copy(seeds, temp)
	}
	return seeds
}

func findSeedByUserID(entries []Entry, userID int) int {
	for _, e := range entries {
		if e.User.UserID == userID {
			return e.Seed
		}
	}
	return 0
}

func (s *Service) GetTournament(ctx context.Context, tournamentID int) (Bracket, error) {
	t, err := s.q.GetTournamentByID(ctx, int32(tournamentID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Bracket{}, game.ErrNotFound
		}
		return Bracket{}, err
	}

	entryRows, err := s.q.ListTournamentEntries(ctx, int32(tournamentID))
	if err != nil {
		return Bracket{}, err
	}

	seedToUser := make(map[int]Player)
	entries := make([]Entry, len(entryRows))
	for i, e := range entryRows {
		u := Player{
			UserID:      int(e.UserID),
			Username:    e.Username,
			DisplayName: e.DisplayName,
			IconPath:    e.IconPath,
			IsAdmin:     e.IsAdmin,
			Label:       e.Label,
		}
		seedToUser[int(e.Seed)] = u
		entries[i] = Entry{
			User: u,
			Seed: int(e.Seed),
		}
	}

	matchRows, err := s.q.ListTournamentMatches(ctx, int32(tournamentID))
	if err != nil {
		return Bracket{}, err
	}

	bracketSize := int(t.BracketSize)
	numRounds := int(t.NumRounds)
	bracketSeeds := StandardBracketSeeds(bracketSize)

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
		gameRow, err := s.q.GetGameByID(ctx, gid)
		if err != nil {
			continue
		}
		if !gameRow.StartedAt.Valid {
			continue
		}
		rankingRows, err := s.q.GetRanking(ctx, gid)
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
		player1   *Player
		player2   *Player
		p1Score   *int
		p2Score   *int
		winnerUID *int
		isBye     bool
	}
	resultByKey := make(map[matchKey]*matchResult)

	for round := range numRounds {
		numPositions := bracketSize / (1 << (round + 1))
		for pos := range numPositions {
			m, exists := matchByKey[matchKey{round, pos}]
			mr := &matchResult{}

			if round == 0 {
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
					if mr.player1 != nil && mr.player2 != nil {
						if rr.winnerID == mr.player1.UserID || rr.winnerID == mr.player2.UserID {
							w := rr.winnerID
							mr.winnerUID = &w
						} else {
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

	// Build response matches
	apiMatches := make([]Match, 0, len(matchRows))
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

			apiMatches = append(apiMatches, Match{
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

	return Bracket{
		TournamentID: int(t.TournamentID),
		DisplayName:  t.DisplayName,
		BracketSize:  bracketSize,
		NumRounds:    numRounds,
		Entries:      entries,
		Matches:      apiMatches,
	}, nil
}

// CreateTournament creates a new tournament with the given number of participants.
func (s *Service) CreateTournament(ctx context.Context, displayName string, numParticipants int) (int, error) {
	if numParticipants < 2 {
		return 0, errors.New("num_participants must be >= 2")
	}

	bracketSize := nextPowerOf2(numParticipants)
	numRounds := log2Int(bracketSize)

	var tournamentID int32
	err := s.txm.RunInTx(ctx, func(qtx db.Querier) error {
		var err error
		tournamentID, err = qtx.CreateTournament(ctx, db.CreateTournamentParams{
			DisplayName: displayName,
			BracketSize: int32(bracketSize),
			NumRounds:   int32(numRounds),
		})
		if err != nil {
			return err
		}
		for round := range numRounds {
			numPositions := bracketSize / (1 << (round + 1))
			for pos := range numPositions {
				if err := qtx.CreateTournamentMatch(ctx, db.CreateTournamentMatchParams{
					TournamentID: tournamentID,
					Round:        int32(round),
					Position:     int32(pos),
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(tournamentID), nil
}

// SeedEntry represents a seed-to-user mapping.
type SeedEntry struct {
	Seed   int
	UserID int
}

// MatchGame represents a match-to-game mapping.
type MatchGame struct {
	MatchID int
	GameID  *int32
}

// UpdateTournamentParams holds parameters for updating a tournament.
type UpdateTournamentParams struct {
	TournamentID int
	DisplayName  string
	SeedEntries  []SeedEntry
	MatchGames   []MatchGame
}

// UpdateTournament updates a tournament's display name, entries, and match games.
func (s *Service) UpdateTournament(ctx context.Context, params UpdateTournamentParams) error {
	t, err := s.q.GetTournamentByID(ctx, int32(params.TournamentID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return game.ErrNotFound
		}
		return err
	}

	return s.txm.RunInTx(ctx, func(qtx db.Querier) error {
		if err := qtx.UpdateTournament(ctx, db.UpdateTournamentParams{
			TournamentID: int32(params.TournamentID),
			DisplayName:  params.DisplayName,
			BracketSize:  t.BracketSize,
			NumRounds:    t.NumRounds,
		}); err != nil {
			return err
		}

		if err := qtx.DeleteTournamentEntries(ctx, int32(params.TournamentID)); err != nil {
			return err
		}
		for _, se := range params.SeedEntries {
			if err := qtx.CreateTournamentEntry(ctx, db.CreateTournamentEntryParams{
				TournamentID: int32(params.TournamentID),
				UserID:       int32(se.UserID),
				Seed:         int32(se.Seed),
			}); err != nil {
				return err
			}
		}

		for _, mg := range params.MatchGames {
			if err := qtx.UpdateTournamentMatchGame(ctx, db.UpdateTournamentMatchGameParams{
				TournamentMatchID: int32(mg.MatchID),
				GameID:            mg.GameID,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

// EditData holds data needed for the tournament edit page.
type EditData struct {
	Tournament  db.Tournament
	SeedUserMap map[int]int
	Matches     []db.TournamentMatch
}

// GetTournamentEditData retrieves the data needed for editing a tournament.
func (s *Service) GetTournamentEditData(ctx context.Context, tournamentID int) (EditData, error) {
	t, err := s.q.GetTournamentByID(ctx, int32(tournamentID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EditData{}, game.ErrNotFound
		}
		return EditData{}, err
	}

	entryRows, err := s.q.ListTournamentEntries(ctx, int32(tournamentID))
	if err != nil {
		return EditData{}, err
	}
	seedUserMap := make(map[int]int)
	for _, e := range entryRows {
		seedUserMap[int(e.Seed)] = int(e.UserID)
	}

	matchRows, err := s.q.ListTournamentMatches(ctx, int32(tournamentID))
	if err != nil {
		return EditData{}, err
	}

	return EditData{
		Tournament:  t,
		SeedUserMap: seedUserMap,
		Matches:     matchRows,
	}, nil
}
