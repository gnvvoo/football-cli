package api

import (
	"fmt"
	"strings"
	"time"
)

// football-data.org 스쿼드 응답 구조체
type SquadAPIResponse struct {
	Squad []Player `json:"squad"`
}

// 선수 기본 정보
type Player struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	DateOfBirth string `json:"dateOfBirth"` // "1998-03-21"
	Nationality string `json:"nationality"`
}

// CLI가 최종 출력하는 선수 스탯 구조
type PlayerStatsResponse struct {
	Player        PlayerOutput `json:"player"`
	DataFreshness string       `json:"data_freshness"`
}

// 선수 출력 구조
type PlayerOutput struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	DateOfBirth string `json:"date_of_birth"`
	Nationality string `json:"nationality"`
	Team        string `json:"team"`
}

// GetPlayerStats : 선수 이름으로 선수 정보 조회
// football-data.org 무료 플랜은 선수 스탯을 제공하지 않아
// 스쿼드 목록에서 선수를 찾아 기본 정보만 반환
func (c *Client) GetPlayerStats(playerName string) (*PlayerStatsResponse, error) {
	// 5대 리그 팀 목록에서 선수 검색
	for _, leagueID := range LeagueIDs {
		teamsEndpoint := fmt.Sprintf("/competitions/%d/teams", leagueID)

		var teamsResp struct {
			Teams []TeamInfoAPIResponse `json:"teams"`
		}

		if err := c.Get(teamsEndpoint, &teamsResp); err != nil {
			continue
		}

		// 각 팀 스쿼드에서 선수 검색
		for _, team := range teamsResp.Teams {
			squadEndpoint := fmt.Sprintf("/teams/%d", team.ID)

			var squadResp struct {
				Name  string `json:"name"`
				Squad []Player `json:"squad"`
			}

			if err := c.Get(squadEndpoint, &squadResp); err != nil {
				continue
			}

			for _, p := range squadResp.Squad {
				if strings.Contains(strings.ToLower(p.Name), strings.ToLower(playerName)) {
					return &PlayerStatsResponse{
						Player: PlayerOutput{
							ID:          p.ID,
							Name:        p.Name,
							Position:    p.Position,
							DateOfBirth: p.DateOfBirth,
							Nationality: p.Nationality,
							Team:        squadResp.Name,
						},
						DataFreshness: time.Now().UTC().Format(time.RFC3339),
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("NO_DATA")
}