package api

import (
	"fmt"
	"strings"
	"time"
)

// football-data.org의 경기 목록 응답 구조체
type MatchesAPIResponse struct {
	Matches []Match `json:"matches"`
}

// API-Football의 경기 데이터 구조체
type Match struct {
	ID          int        `json:"id"`
	UtcDate     string     `json:"utcDate"`   // ISO8601 형식
	Status      string     `json:"status"`    // "SCHEDULED", "IN_PLAY", "FINISHED" 등
	HomeTeam    MatchTeam  `json:"homeTeam"`
	AwayTeam    MatchTeam  `json:"awayTeam"`
	Score       MatchScore `json:"score"`
	Competition Competition `json:"competition"`
	Venue       string     `json:"venue"`
}

// 팀 기본 정보
type MatchTeam struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 득점 정보
// 포인터(*int)를 쓰는 이유 : 경기 전엔 점수가 null이라서
// int면 null을 0으로 오해 가능성 → *int면 null = nil로 구분 가능
type MatchScore struct {
	FullTime ScoreDetail `json:"fullTime"`
	HalfTime ScoreDetail `json:"halfTime"`
}

// 점수 상세
type ScoreDetail struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

// 대회 정보
type Competition struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// --manifest 스키마의 matches output 구조
type MatchesResponse struct {
	Matches       []MatchOutput `json:"matches"`
	DataFreshness string        `json:"data_freshness"`
}

// CLI가 최종 출력하는 경기 데이터 구조
// API 응답을 -> 지정한 스키마에 맞게 변환 후 출력
type MatchOutput struct {
	ID       int         `json:"id"`
	Date     string      `json:"date"`
	Status   string      `json:"status"`
	HomeTeam string      `json:"home_team"`
	AwayTeam string      `json:"away_team"`
	Score    ScoreOutput `json:"score"`
	League   string      `json:"league"`
	Venue    string      `json:"venue"`
}

// 점수 출력 구조
type ScoreOutput struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

// API-Football에서 경기 목록을 가져오는 함수
// leagueID: 리그 ID (예: 2021)
// date: "2026-04-13" 형식, 빈 문자열이면 오늘 날짜 사용
// team: 팀 이름 필터 (빈 문자열이면 전체)
// status: "live", "upcoming", "finished" (빈 문자열이면 전체)
func (c *Client) GetMatches(leagueID int, date, team, status string) (*MatchesResponse, error) {
	// 날짜가 없으면 오늘 날짜 사용
	if date == "" {
		// Go의 날짜 포맷은 2006-01-02 기준
		date = time.Now().Format("2006-01-02")
	}

	// 엔드포인트 조합
	endpoint := fmt.Sprintf("/competitions/%d/matches?dateFrom=%s&dateTo=%s",
		leagueID,
		date,
		date,
	)

	// API 호출
	var apiResp MatchesAPIResponse
	if err := c.Get(endpoint, &apiResp); err != nil {
		return nil, err
	}

	// 결과가 없으면 NO_DATA
	if len(apiResp.Matches) == 0 {
		return nil, fmt.Errorf("NO_DATA")
	}

	// API 응답을 출력 구조체로 변환
	output := &MatchesResponse{
		DataFreshness: time.Now().UTC().Format(time.RFC3339),
	}

	for _, m := range apiResp.Matches {
		// 상태 필터링
		if status != "" && !matchStatus(m.Status, status) {
			continue
		}

		// 팀 필터링 (대소문자 구분 없는 부분 문자열 매칭)
		if team != "" && !containsTeam(m.HomeTeam.Name, m.AwayTeam.Name, team) {
			continue
		}

		output.Matches = append(output.Matches, MatchOutput{
			ID:       m.ID,
			Date:     m.UtcDate,
			Status:   m.Status,
			HomeTeam: m.HomeTeam.Name,
			AwayTeam: m.AwayTeam.Name,
			Score: ScoreOutput{
				Home: m.Score.FullTime.Home,
				Away: m.Score.FullTime.Away,
			},
			League: m.Competition.Name,
			Venue:  m.Venue,
		})
	}

	return output, nil
}

// API 상태 코드를 status 필터에 맞게 변환
func matchStatus(apiStatus, filter string) bool {
	switch filter {
	case "live":
		return apiStatus == "IN_PLAY" || apiStatus == "PAUSED"
	case "upcoming":
		return apiStatus == "SCHEDULED" || apiStatus == "TIMED"
	case "finished":
		return apiStatus == "FINISHED"
	}
	return true // 필터 없으면 전체 반환
}

// 홈 또는 어웨이 팀 이름에 검색어가 포함되는지 확인
// 대소문자 구분 없이 부분 매칭
func containsTeam(home, away, query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(home), q) ||
		strings.Contains(strings.ToLower(away), q)
}