package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"football-cli/internal/api"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "경고: .env 파일을 불러오지 못했습니다")
		}
	}

	// API 클라이언트 초기화 (인자: timeoutMs만, 키는 내부에서 환경변수로 읽음)
	client, err := api.NewClient(5000)
	if err != nil {
		fmt.Fprintf(os.Stderr, "클라이언트 초기화 실패: %v\n", err)
		os.Exit(1)
	}

	// MCP 서버 생성
	s := server.NewMCPServer(
		"football-cli",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// matches 툴 등록
	s.AddTool(
		mcp.NewTool("get_matches",
			mcp.WithDescription("리그별 경기 일정 및 결과 조회. 날짜·팀·상태 필터 가능"),
			mcp.WithString("league",
				mcp.Required(),
				mcp.Description("리그 약어: EPL, LaLiga, Bundesliga, SerieA, Ligue1"),
			),
			mcp.WithString("date",
				mcp.Description("조회 날짜 (YYYY-MM-DD). 기본값: 오늘"),
			),
			mcp.WithString("from",
				mcp.Description("시작 날짜 (YYYY-MM-DD)"),
			),
			mcp.WithString("to",
				mcp.Description("종료 날짜 (YYYY-MM-DD)"),
			),
			mcp.WithString("team",
				mcp.Description("팀 이름 (부분 검색 가능, 예: Arsenal)"),
			),
			mcp.WithString("status",
				mcp.Description("경기 상태: live, upcoming, finished"),
			),
		),
		handleGetMatches(client),
	)

	// standings 툴 등록
	s.AddTool(
    	mcp.NewTool("get_standings",
        	mcp.WithDescription("리그 순위 조회"),
        	mcp.WithString("league",
            	mcp.Required(),
            	mcp.Description("리그 약어: EPL, LaLiga, Bundesliga, SerieA, Ligue1"),
        	),
    	),
    	handleGetStandings(client),
	)

	// team-info 툴 등록
	s.AddTool(
		mcp.NewTool("get_team_info",
			mcp.WithDescription("팀 정보 조회. 창단연도, 경기장, 소속 리그 포함"),
			mcp.WithString("team",
				mcp.Required(),
				mcp.Description("팀 이름 (부분 검색 가능, 예: Arsenal)"),
			),
		),
		handleGetTeamInfo(client),
	)

	// player-stats 툴 등록
	s.AddTool(
		mcp.NewTool("get_player_stats",
			mcp.WithDescription("선수 정보 조회. 포지션, 나이, 국적, 소속팀 포함"),
			mcp.WithString("player",
				mcp.Required(),
				mcp.Description("선수 이름 (부분 검색 가능, 예: Salah)"),
			),
			mcp.WithString("league",
				mcp.Description("리그 약어: EPL, LaLiga, Bundesliga, SerieA, Ligue1. 생략 시 전체 검색"),
			),
		),
		handleGetPlayerStats(client),
	)

	// predict 툴 등록
	s.AddTool(
		mcp.NewTool("predict_match",
			mcp.WithDescription("두 팀의 경기 결과 예측. 순위·맞대결·홈 어드밴티지 기반"),
			mcp.WithString("home",
				mcp.Required(),
				mcp.Description("홈팀 이름 (부분 검색 가능, 예: Arsenal)"),
			),
			mcp.WithString("away",
				mcp.Required(),
				mcp.Description("어웨이팀 이름 (부분 검색 가능, 예: Chelsea)"),
			),
			mcp.WithBoolean("explain",
				mcp.Description("예측 근거 상세 출력 여부"),
			),
		),
		handlePredictMatch(client),
	)

	// stdio 모드로 실행
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "서버 오류: %v\n", err)
		os.Exit(1)
	}
}

// stringArg : Arguments 맵에서 문자열 파라미터를 안전하게 추출
// Arguments가 any 타입이라 타입 단언(type assertion) 필요
func stringArg(args map[string]any, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// matches 툴 핸들러
func handleGetMatches(client *api.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Arguments를 map으로 타입 단언 후 파라미터 추출
		args, _ := req.Params.Arguments.(map[string]any)
		league := stringArg(args, "league")
		date   := stringArg(args, "date")
		from   := stringArg(args, "from")
		to     := stringArg(args, "to")
		team   := stringArg(args, "team")
		status := stringArg(args, "status")

		// 리그 약어 → ID 변환 (LeagueIDs 맵 직접 사용)
		leagueID, ok := api.LeagueIDs[league]
		if !ok {
			return mcp.NewToolResultError(
				fmt.Sprintf("지원하지 않는 리그입니다: %s (EPL, LaLiga, Bundesliga, SerieA, Ligue1)", league),
			), nil
		}

		// API 호출
		result, err := client.GetMatches(leagueID, date, from, to, team, status)
		if err != nil {
			if err.Error() == "NO_DATA" {
				return mcp.NewToolResultError("해당 조건의 경기 데이터가 없습니다"), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("API 오류: %v", err)), nil
		}

		// JSON 직렬화 후 반환
		b, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("응답 직렬화 실패"), nil
		}

		return mcp.NewToolResultText(string(b)), nil
	}
}

// standings 툴 핸들러
func handleGetStandings(client *api.Client) server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, _ := req.Params.Arguments.(map[string]any)
        league := stringArg(args, "league")

        leagueID, ok := api.LeagueIDs[league]
        if !ok {
            return mcp.NewToolResultError(
                fmt.Sprintf("지원하지 않는 리그입니다: %s", league),
            ), nil
        }

        result, err := client.GetStandings(leagueID)
        if err != nil {
            if err.Error() == "NO_DATA" {
                return mcp.NewToolResultError("순위 데이터가 없습니다"), nil
            }
            return mcp.NewToolResultError(fmt.Sprintf("API 오류: %v", err)), nil
        }

        b, err := json.Marshal(result)
        if err != nil {
            return mcp.NewToolResultError("응답 직렬화 실패"), nil
        }

        return mcp.NewToolResultText(string(b)), nil
    }
}

// team-info 툴 핸들러
func handleGetTeamInfo(client *api.Client) server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, _ := req.Params.Arguments.(map[string]any)
        team := stringArg(args, "team")

        result, err := client.GetTeamInfo(team)
        if err != nil {
            if err.Error() == "NO_DATA" {
                return mcp.NewToolResultError(fmt.Sprintf("팀을 찾을 수 없습니다: %s", team)), nil
            }
            return mcp.NewToolResultError(fmt.Sprintf("API 오류: %v", err)), nil
        }

        b, err := json.Marshal(result)
        if err != nil {
            return mcp.NewToolResultError("응답 직렬화 실패"), nil
        }

        return mcp.NewToolResultText(string(b)), nil
    }
}

// player-stats 툴 핸들러
func handleGetPlayerStats(client *api.Client) server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, _ := req.Params.Arguments.(map[string]any)
        player := stringArg(args, "player")
        league := stringArg(args, "league")

        result, err := client.GetPlayerStats(player, league)
        if err != nil {
            if err.Error() == "NO_DATA" {
                return mcp.NewToolResultError(fmt.Sprintf("선수를 찾을 수 없습니다: %s", player)), nil
            }
            return mcp.NewToolResultError(fmt.Sprintf("API 오류: %v", err)), nil
        }

        b, err := json.Marshal(result)
        if err != nil {
            return mcp.NewToolResultError("응답 직렬화 실패"), nil
        }

        return mcp.NewToolResultText(string(b)), nil
    }
}

// predict 툴 핸들러
func handlePredictMatch(client *api.Client) server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args, _ := req.Params.Arguments.(map[string]any)
        home := stringArg(args, "home")
        away := stringArg(args, "away")

        // explain은 bool 타입이라 별도 추출
        explain := false
        if v, ok := args["explain"]; ok {
            explain, _ = v.(bool)
        }

        result, err := client.Predict(home, away, explain)
        if err != nil {
            if err.Error() == "NO_DATA" {
                return mcp.NewToolResultError(
                    fmt.Sprintf("팀을 찾을 수 없습니다: %s 또는 %s", home, away),
                ), nil
            }
            return mcp.NewToolResultError(fmt.Sprintf("API 오류: %v", err)), nil
        }

        b, err := json.Marshal(result)
        if err != nil {
            return mcp.NewToolResultError("응답 직렬화 실패"), nil
        }

        return mcp.NewToolResultText(string(b)), nil
    }
}