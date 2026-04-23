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

// CLI가 최종 출력하는 선수 스탯 구조 (복수 결과)
type PlayerStatsResponse struct {
	Players       []PlayerOutput `json:"players"`
	DataFreshness string         `json:"data_freshness"`
}

// 선수 출력 구조
type PlayerOutput struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	DateOfBirth string `json:"date_of_birth"`
	Age         int    `json:"age"`
	Nationality string `json:"nationality"`
	Team        string `json:"team"`
}

// calcAge : 생년월일 문자열("1998-03-21")로 현재 나이 계산
func calcAge(dob string) int {
	t, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - t.Year()
	// 아직 생일이 지나지 않은 경우 1살 빼기
	if now.Month() < t.Month() || (now.Month() == t.Month() && now.Day() < t.Day()) {
		age--
	}
	return age
}

// GetPlayerStats : 선수 이름으로 선수 정보 조회
// league가 빈 문자열이면 5대 리그 전체에서 검색
// 동명이인을 고려해 매칭되는 선수를 모두 반환
func (c *Client) GetPlayerStats(playerName, league string) (*PlayerStatsResponse, error) {
	// 검색할 리그 ID 목록 결정
	leagueIDs := make(map[string]int)
	if league != "" {
		id, ok := LeagueIDs[league]
		if !ok {
			return nil, fmt.Errorf("INVALID_LEAGUE")
		}
		leagueIDs[league] = id
	} else {
		leagueIDs = LeagueIDs
	}

	var found []PlayerOutput

	for _, leagueID := range leagueIDs {
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
				Name  string   `json:"name"`
				Squad []Player `json:"squad"`
			}

			if err := c.Get(squadEndpoint, &squadResp); err != nil {
				continue
			}

			for _, p := range squadResp.Squad {
				if strings.Contains(strings.ToLower(p.Name), strings.ToLower(playerName)) {
					found = append(found, PlayerOutput{
						ID:          p.ID,
						Name:        p.Name,
						Position:    p.Position,
						DateOfBirth: p.DateOfBirth,
						Age:         calcAge(p.DateOfBirth),
						Nationality: p.Nationality,
						Team:        squadResp.Name,
					})
				}
			}
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("NO_DATA")
	}

	return &PlayerStatsResponse{
		Players:       found,
		DataFreshness: time.Now().UTC().Format(time.RFC3339),
	}, nil
}