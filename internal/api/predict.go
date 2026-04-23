package api

import (
	"fmt"
	"strings"
	"time"
)

// H2H API 응답 구조체
type H2HAPIResponse struct {
	Matches []Match `json:"matches"`
}

// CLI가 최종 출력하는 예측 응답 구조
type PredictResponse struct {
	Match      MatchInfo      `json:"match"`
	Prediction PredictionInfo `json:"prediction"`
	Basis      BasisInfo      `json:"basis"`
	Reasoning  []string       `json:"reasoning,omitempty"` // --explain 시에만 포함
	DataFreshness string      `json:"data_freshness"`
}

// 경기 기본 정보
type MatchInfo struct {
	HomeTeam string `json:"home_team"`
	AwayTeam string `json:"away_team"`
}

// 예측 결과
type PredictionInfo struct {
	Winner     string      `json:"winner"`     // "home" | "away" | "draw"
	Confidence float64     `json:"confidence"` // 0.0 ~ 1.0
	Scores     ScoreOutput `json:"scores"`
}

// 예측 근거 데이터
type BasisInfo struct {
	HomeForm   string `json:"home_form"`
	AwayForm   string `json:"away_form"`
	H2HSummary string `json:"h2h_summary"`
}

// Predict : 두 팀의 경기 결과 예측
func (c *Client) Predict(homeTeam, awayTeam string, explain bool) (*PredictResponse, error) {
	// 1. 두 팀이 속한 리그 순위 데이터 조회
	homeStanding, awayStanding, leagueID, err := c.findTeamStandings(homeTeam, awayTeam)
	if err != nil {
		return nil, err
	}

	// 2. H2H 데이터 조회
	h2hWins, h2hDraws, h2hLosses, err := c.getH2H(homeStanding.Team.ID, awayStanding.Team.ID)
	if err != nil {
		// H2H 실패해도 예측은 진행
		h2hWins, h2hDraws, h2hLosses = 0, 0, 0
	}
	_ = leagueID

	// 3. 승리 확률 계산
	homeProb, drawProb, awayProb := calcProbability(
		homeStanding, awayStanding,
		h2hWins, h2hDraws, h2hLosses,
	)

	// 4. 예측 승자 결정
	winner, confidence := determineWinner(homeProb, drawProb, awayProb)

	// 5. 예상 스코어 계산
	homeScore, awayScore := calcExpectedScore(homeStanding, awayStanding)

	// 6. H2H 요약 문자열
	h2hTotal := h2hWins + h2hDraws + h2hLosses
	h2hSummary := fmt.Sprintf("최근 맞대결 %d경기: %d승 %d무 %d패",
		h2hTotal, h2hWins, h2hDraws, h2hLosses)

	resp := &PredictResponse{
		Match: MatchInfo{
			HomeTeam: homeStanding.Team.Name,
			AwayTeam: awayStanding.Team.Name,
		},
		Prediction: PredictionInfo{
			Winner:     winner,
			Confidence: confidence,
			Scores: ScoreOutput{
				Home: &homeScore,
				Away: &awayScore,
			},
		},
		Basis: BasisInfo{
			HomeForm:   fmt.Sprintf("%d위 (승점 %d)", homeStanding.Position, homeStanding.Points),
			AwayForm:   fmt.Sprintf("%d위 (승점 %d)", awayStanding.Position, awayStanding.Points),
			H2HSummary: h2hSummary,
		},
		DataFreshness: time.Now().UTC().Format(time.RFC3339),
	}

	// --explain 시 reasoning 추가
	if explain {
		resp.Reasoning = buildReasoning(
			homeStanding, awayStanding,
			h2hWins, h2hDraws, h2hLosses,
			homeProb, drawProb, awayProb,
		)
	}

	return resp, nil
}

// findTeamStandings : 두 팀을 5대 리그 순위에서 검색
func (c *Client) findTeamStandings(homeTeam, awayTeam string) (*TeamStanding, *TeamStanding, int, error) {
	for _, leagueID := range LeagueIDs {
		standings, err := c.getRawStandings(leagueID)
		if err != nil {
			continue
		}

		var home, away *TeamStanding
		for i, s := range standings {
			if strings.Contains(strings.ToLower(s.Team.Name), strings.ToLower(homeTeam)) {
				home = &standings[i]
			}
			if strings.Contains(strings.ToLower(s.Team.Name), strings.ToLower(awayTeam)) {
				away = &standings[i]
			}
		}

		if home != nil && away != nil {
			return home, away, leagueID, nil
		}
	}
	return nil, nil, 0, fmt.Errorf("NO_DATA")
}

// getH2H : 두 팀의 최근 맞대결 결과 조회
func (c *Client) getH2H(homeID, awayID int) (wins, draws, losses int, err error) {
	endpoint := fmt.Sprintf("/teams/%d/matches?competitions=PL,PD,BL1,SA,FL1&limit=10", homeID)

	var apiResp struct {
		Matches []Match `json:"matches"`
	}

	if err := c.Get(endpoint, &apiResp); err != nil {
		return 0, 0, 0, err
	}

	for _, m := range apiResp.Matches {
		// 두 팀 간의 경기만 필터링
		if (m.HomeTeam.ID == homeID && m.AwayTeam.ID == awayID) ||
			(m.HomeTeam.ID == awayID && m.AwayTeam.ID == homeID) {
			if m.Score.FullTime.Home == nil || m.Score.FullTime.Away == nil {
				continue
			}
			homeGoals := *m.Score.FullTime.Home
			awayGoals := *m.Score.FullTime.Away

			if m.HomeTeam.ID == homeID {
				if homeGoals > awayGoals {
					wins++
				} else if homeGoals == awayGoals {
					draws++
				} else {
					losses++
				}
			} else {
				if awayGoals > homeGoals {
					wins++
				} else if homeGoals == awayGoals {
					draws++
				} else {
					losses++
				}
			}
		}
	}
	return wins, draws, losses, nil
}

// calcProbability : 승/무/패 확률 계산
func calcProbability(home, away *TeamStanding, h2hW, h2hD, h2hL int) (homeProb, drawProb, awayProb float64) {
	// PlayedGames로 수정
	homeAvg := float64(home.Points) / float64(max(home.PlayedGames, 1))
	awayAvg := float64(away.Points) / float64(max(away.PlayedGames, 1))

	h2hTotal := h2hW + h2hD + h2hL
	homeH2H := 0.5
	if h2hTotal > 0 {
		homeH2H = float64(h2hW*3+h2hD) / float64(h2hTotal*3)
	}

	homeBonus := 0.1
	homeScore := homeAvg*0.6 + homeH2H*0.4 + homeBonus
	awayScore := awayAvg*0.6 + (1-homeH2H)*0.4

	diff := homeScore - awayScore
	if diff < 0 {
		diff = -diff
	}
	drawProb = 0.3 - diff*0.1
	if drawProb < 0.1 {
		drawProb = 0.1
	}

	total := homeScore + awayScore + drawProb
	homeProb = homeScore / total
	awayProb = awayScore / total
	drawProb = drawProb / total

	return homeProb, drawProb, awayProb
}
// determineWinner : 가장 높은 확률의 결과 반환
func determineWinner(homeProb, drawProb, awayProb float64) (string, float64) {
	if homeProb >= drawProb && homeProb >= awayProb {
		return "home", homeProb
	}
	if awayProb >= homeProb && awayProb >= drawProb {
		return "away", awayProb
	}
	return "draw", drawProb
}

// calcExpectedScore : 예상 스코어 계산
func calcExpectedScore(home, away *TeamStanding) (int, int) {
	// GoalsFor, GoalsAgainst, PlayedGames로 수정
	homeAvgGoals := float64(home.GoalsFor) / float64(max(home.PlayedGames, 1))
	awayAvgGoals := float64(away.GoalsFor) / float64(max(away.PlayedGames, 1))
	homeAvgConcede := float64(home.GoalsAgainst) / float64(max(home.PlayedGames, 1))
	awayAvgConcede := float64(away.GoalsAgainst) / float64(max(away.PlayedGames, 1))

	homeExpected := (homeAvgGoals + awayAvgConcede) / 2
	awayExpected := (awayAvgGoals + homeAvgConcede) / 2

	return int(homeExpected + 0.5), int(awayExpected + 0.5)
}

// buildReasoning : --explain 시 근거 문자열 배열 생성
func buildReasoning(home, away *TeamStanding, h2hW, h2hD, h2hL int, homeProb, drawProb, awayProb float64) []string {
	return []string{
		fmt.Sprintf("%s 현재 %d위 (승점 %d, 경기당 평균 %.2f점)",
			home.Team.Name, home.Position, home.Points,
			float64(home.Points)/float64(max(home.PlayedGames, 1))),
		fmt.Sprintf("%s 현재 %d위 (승점 %d, 경기당 평균 %.2f점)",
			away.Team.Name, away.Position, away.Points,
			float64(away.Points)/float64(max(away.PlayedGames, 1))),
		fmt.Sprintf("최근 맞대결: %d승 %d무 %d패", h2hW, h2hD, h2hL),
		fmt.Sprintf("홈 어드밴티지 +10%% 적용"),
		fmt.Sprintf("홈팀 승 확률 %.0f%% / 무승부 %.0f%% / 원정팀 승 확률 %.0f%%",
			homeProb*100, drawProb*100, awayProb*100),
	}
}

// max : 두 정수 중 큰 값 반환
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}