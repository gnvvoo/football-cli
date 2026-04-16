package main

import (
	"football-cli/cmd"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// godotenv.Load() : .env 파일을 읽고 환경변수로 등록

	// err := godotenv.Load()
	// if err != nil { }
	// 위 코드와 동일한 문장. 한 문장으로 표현.
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Println("경고: .env 파일을 불러오지 못했습니다")
		}
	}

	// CLI 실행
	cmd.Execute()
}
