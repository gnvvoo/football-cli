package cmd

import (
	"football-cli/internal/api"
	"football-cli/internal/output"
	"os"

	"github.com/spf13/cobra"
)

// standings 커맨드 flag 변수
var (
	standingsLeague string
	standingsSeason int
)

var standingsCmd = &cobra.Command{
	Use:   "standings",
	Short: "리그 순위 조회",
	Long:  "리그별 현재 순위 조회. --league 필수.",
	RunE:  runStandings,
}

func init() {
	// standings 커맨드를 루트 커맨드에 등록
	rootCmd.AddCommand(standingsCmd)

	// standings 전용 flag 등록
	standingsCmd.Flags().StringVar(&standingsLeague, "league", "", "리그 (EPL|LaLiga|Bundesliga|SerieA|Ligue1) [필수]")
	standingsCmd.Flags().IntVar(&standingsSeason, "season", 0, "시즌 연도 (예: 2024, 기본값: 현재 시즌)")

	// --league 필수 flag로 지정
	standingsCmd.MarkFlagRequired("league")
}

func runStandings(cmd *cobra.Command, args []string) error {
	// 리그 약어 → API ID 변환
	leagueID, ok := api.LeagueIDs[standingsLeague]
	if !ok {
		PrintError(
			"INVALID_LEAGUE",
			"유효하지 않은 리그입니다: "+standingsLeague,
			[]string{"EPL", "LaLiga", "Bundesliga", "SerieA", "Ligue1"},
		)
		os.Exit(ExitInvalidArgs)
	}

	// API 클라이언트 생성
	client, err := api.NewClient(Timeout)
	if err != nil {
		PrintError("AUTH_FAILURE", err.Error(), nil)
		os.Exit(ExitAuthFailure)
	}

	// 순위 데이터 조회
	resp, err := client.GetStandings(leagueID)
	if err != nil {
		switch err.Error() {
		case "NO_DATA":
			PrintError("NO_DATA", "순위 데이터가 없습니다.", nil)
			os.Exit(ExitNoData)
		case "AUTH_FAILURE":
			PrintError("AUTH_FAILURE", "API 인증에 실패했습니다. API 키를 확인해주세요.", nil)
			os.Exit(ExitAuthFailure)
		default:
			PrintError("API_FAILURE", err.Error(), nil)
			os.Exit(ExitAPIFailure)
		}
	}

	// --json 플래그 시 JSON 출력
	if JSONOutput {
		return output.PrintJSON(resp)
	}

	// 텍스트 테이블 출력
	var rows []output.StandingRow
	for _, s := range resp.Standings {
		rows = append(rows, output.StandingRow{
			Rank:   s.Rank,
			Team:   s.Team,
			Played: s.Played,
			Won:    s.Won,
			Drawn:  s.Drawn,
			Lost:   s.Lost,
			GF:     s.GF,
			GA:     s.GA,
			GD:     s.GD,
			Points: s.Points,
			Form:   s.Form,
		})
	}
	output.PrintStandingsTable(resp.League, resp.Season, rows)

	return nil
}