package api

import (
	"fmt"
	"time"
)

// football-data.org 순위 응답 구조체
type StandingsAPIResponse struct {
	Competition Competition      `json:"competition"`
	Season      Season           `json:"season"`
	Standings   []StandingTable  `json:"standings"`
}

// 시즌 정보
type Season struct {
	StartDate string `json:"startDate"` // "2025-08-08"
	EndDate   string `json:"endDate"`   // "2026-05-24"
}

// 순위 테이블 (홈/어웨이/전체 세 가지로 나뉘어서 옴)
type StandingTable struct {
	Type  string         `json:"type"`  // "TOTAL", "HOME", "AWAY"
	Table []TeamStanding `json:"table"`
}

// 팀 순위 정보
type TeamStanding struct {
	Position       int    `json:"position"`
	Team           Team   `json:"team"`
	PlayedGames    int    `json:"playedGames"`
	Won            int    `json:"won"`
	Draw           int    `json:"draw"`
	Lost           int    `json:"lost"`
	GoalsFor       int    `json:"goalsFor"`
	GoalsAgainst   int    `json:"goalsAgainst"`
	GoalDifference int    `json:"goalDifference"`
	Points         int    `json:"points"`
	Form           string `json:"form"` // "W,W,D,L,W"
}

// CLI가 최종 출력하는 순위 응답 구조
type StandingsResponse struct {
	League        string              `json:"league"`
	Season        string              `json:"season"`
	Standings     []TeamStandingOutput `json:"standings"`
	DataFreshness string              `json:"data_freshness"`
}

// CLI가 최종 출력하는 팀 순위 데이터
type TeamStandingOutput struct {
	Rank   int    `json:"rank"`
	Team   string `json:"team"`
	Played int    `json:"played"`
	Won    int    `json:"won"`
	Drawn  int    `json:"drawn"`
	Lost   int    `json:"lost"`
	GF     int    `json:"gf"`
	GA     int    `json:"ga"`
	GD     int    `json:"gd"`
	Points int    `json:"points"`
	Form   string `json:"form"` // "WWDLW"
}

// GetStandings : 리그 순위 조회
// leagueID: 리그 ID (예: 2021)
func (c *Client) GetStandings(leagueID int) (*StandingsResponse, error) {
	endpoint := fmt.Sprintf("/competitions/%d/standings", leagueID)

	var apiResp StandingsAPIResponse
	if err := c.Get(endpoint, &apiResp); err != nil {
		return nil, err
	}

	// TOTAL 테이블만 사용 (HOME, AWAY 제외)
	var totalTable []TeamStanding
	for _, s := range apiResp.Standings {
		if s.Type == "TOTAL" {
			totalTable = s.Table
			break
		}
	}

	if len(totalTable) == 0 {
		return nil, fmt.Errorf("NO_DATA")
	}

	output := &StandingsResponse{
		League:        apiResp.Competition.Name,
		Season:        apiResp.Season.StartDate[:4], // "2025-08-08" → "2025"
		DataFreshness: time.Now().UTC().Format(time.RFC3339),
	}

	for _, t := range totalTable {
		// Form 변환 "W,W,D,L,W" → "WWDLW"
		form := formatForm(t.Form)

		output.Standings = append(output.Standings, TeamStandingOutput{
			Rank:   t.Position,
			Team:   t.Team.Name,
			Played: t.PlayedGames,
			Won:    t.Won,
			Drawn:  t.Draw,
			Lost:   t.Lost,
			GF:     t.GoalsFor,
			GA:     t.GoalsAgainst,
			GD:     t.GoalDifference,
			Points: t.Points,
			Form:   form,
		})
	}

	return output, nil
}

// formatForm : "W,W,D,L,W" → "WWDLW" 변환
func formatForm(form string) string {
	result := ""
	for _, c := range form {
		if c != ',' {
			result += string(c)
		}
	}
	return result
}