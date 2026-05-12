package cmd

import (
	"football-cli/internal/api"
	"football-cli/internal/output"
	"football-cli/internal/schema"
	"os"

	"github.com/spf13/cobra"
)

// player-stats 커맨드 flag 변수
var (
	playerStatsPlayer string
	playerStatsLeague string
)

var playerStatsCmd = &cobra.Command{
	Use:   "player-stats",
	Short: "선수 정보 조회",
	Long:  "선수 이름으로 선수 정보 조회. --player 필수. --league로 검색 범위 축소 가능.",
	RunE:  runPlayerStats,
}

func init() {
	rootCmd.AddCommand(playerStatsCmd)

	playerStatsCmd.Flags().StringVar(&playerStatsPlayer, "player", "", "선수 이름 (부분 검색 가능) [필수]")
	playerStatsCmd.Flags().StringVar(&playerStatsLeague, "league", "", "리그 (EPL|LaLiga|Bundesliga|SerieA|Ligue1) [선택]")

	// playerStatsCmd.MarkFlagRequired("player")
}

func runPlayerStats(cmd *cobra.Command, args []string) error {
	// --schema 플래그 시 스키마 출력 후 종료
	if SchemaFlag {
		s, _ := schema.GetCommandSchema("player-stats")
		return output.PrintJSON(s)
	}

	// --schema 없을 때만 필수 플래그 검증
	if playerStatsPlayer == "" {
		PrintError("INVALID_ARGS", "--player 플래그가 필요합니다.", nil)
		os.Exit(ExitInvalidArgs)
	}

	client, err := api.NewClient(Timeout)
	if err != nil {
		PrintError("AUTH_FAILURE", err.Error(), nil)
		os.Exit(ExitAuthFailure)
	}

	resp, err := client.GetPlayerStats(playerStatsPlayer, playerStatsLeague)
	if err != nil {
		if err.Error() == "INVALID_LEAGUE" {
			PrintError("INVALID_LEAGUE", "유효하지 않은 리그입니다: "+playerStatsLeague,
				[]string{"EPL", "LaLiga", "Bundesliga", "SerieA", "Ligue1"})
			os.Exit(ExitInvalidArgs)
		}
		handleAPIError(err, "선수를 찾을 수 없습니다: "+playerStatsPlayer)
	}

	if JSONOutput {
		return output.PrintJSON(resp)
	}

	var rows []output.PlayerStatRow
	for _, p := range resp.Players {
		rows = append(rows, output.PlayerStatRow{
			Name:        p.Name,
			Position:    p.Position,
			DateOfBirth: p.DateOfBirth,
			Age:         p.Age,
			Nationality: p.Nationality,
			Team:        p.Team,
		})
	}
	output.PrintPlayerStatsTable(rows)

	return nil
}
