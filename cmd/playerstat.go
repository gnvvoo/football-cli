package cmd

import (
	"football-cli/internal/api"
	"football-cli/internal/output"
	"os"

	"github.com/spf13/cobra"
)

// player-stats 커맨드 flag 변수
var playerStatsPlayer string

var playerStatsCmd = &cobra.Command{
	Use:   "player-stats",
	Short: "선수 정보 조회",
	Long:  "선수 이름으로 선수 정보 조회. --player 필수.",
	RunE:  runPlayerStats,
}

func init() {
	rootCmd.AddCommand(playerStatsCmd)

	playerStatsCmd.Flags().StringVar(&playerStatsPlayer, "player", "", "선수 이름 (부분 검색 가능) [필수]")
	playerStatsCmd.MarkFlagRequired("player")
}

func runPlayerStats(cmd *cobra.Command, args []string) error {
	client, err := api.NewClient(Timeout)
	if err != nil {
		PrintError("AUTH_FAILURE", err.Error(), nil)
		os.Exit(ExitAuthFailure)
	}

	resp, err := client.GetPlayerStats(playerStatsPlayer)
	if err != nil {
		switch err.Error() {
		case "NO_DATA":
			PrintError("NO_DATA", "선수를 찾을 수 없습니다: "+playerStatsPlayer, nil)
			os.Exit(ExitNoData)
		case "AUTH_FAILURE":
			PrintError("AUTH_FAILURE", "API 인증에 실패했습니다. API 키를 확인해주세요.", nil)
			os.Exit(ExitAuthFailure)
		default:
			PrintError("API_FAILURE", err.Error(), nil)
			os.Exit(ExitAPIFailure)
		}
	}

	if JSONOutput {
		return output.PrintJSON(resp)
	}

	output.PrintPlayerStats(
		resp.Player.Name,
		resp.Player.Position,
		resp.Player.DateOfBirth,
		resp.Player.Nationality,
		resp.Player.Team,
	)

	return nil
}