package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "경고: .env 파일을 불러오지 못했습니다")
		}
	}

	// MCP 서버 생성
	s := server.NewMCPServer(
		"football-cli",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// 툴은 다음 커밋에서 추가

	// stdio 모드로 실행 (Claude Desktop 연결 방식)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "서버 오류: %v\n", err)
		os.Exit(1)
	}
}