package api

import (
	"fmt"
	"time"
)

// football-data.org 팀 정보 응답 구조체
type TeamInfoAPIResponse struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	ShortName   string       `json:"shortName"`
	Founded     int          `json:"founded"`
	Venue       string       `json:"venue"`
	RunningCompetitions []Competition `json:"runningCompetitions"`
}

// CLI가 최종 출력하는 팀 정보 구조
type TeamInfoResponse struct {
	Team          TeamDetail     `json:"team"`
	DataFreshness string         `json:"data_freshness"`
}

// 팀 상세 정보
type TeamDetail struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Founded  int    `json:"founded"`
	Venue    string `json:"venue"`
	Leagues  []string `json:"leagues"`
}

// GetTeamInfo : 팀 이름으로 팀 정보 조회
// football-data.org는 팀 이름 검색을 지원하지 않아서
// 리그 팀 목록에서 검색하는 방식으로 구현
func (c *Client) GetTeamInfo(teamName string) (*TeamInfoResponse, error) {
	// 5대 리그 팀 목록에서 검색
	for _, leagueID := range LeagueIDs {
		endpoint := fmt.Sprintf("/competitions/%d/teams", leagueID)

		var apiResp struct {
			Teams []TeamInfoAPIResponse `json:"teams"`
		}

		if err := c.Get(endpoint, &apiResp); err != nil {
			continue
		}

		// 팀 이름 부분 매칭
		for _, t := range apiResp.Teams {
			if containsTeam(t.Name, t.ShortName, teamName) {
				leagues := make([]string, 0)
				for _, comp := range t.RunningCompetitions {
					leagues = append(leagues, comp.Name)
				}

				return &TeamInfoResponse{
					Team: TeamDetail{
						ID:      t.ID,
						Name:    t.Name,
						Founded: t.Founded,
						Venue:   t.Venue,
						Leagues: leagues,
					},
					DataFreshness: time.Now().UTC().Format(time.RFC3339),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("NO_DATA")
}