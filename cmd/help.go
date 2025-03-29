package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// 명령행 인자 처리 및 도움말 출력
func handleCommandLineArgs() (bool, error) {
	// 도움말 표시 처리
	helpCmd := flag.NewFlagSet("help", flag.ExitOnError)

	// 명령행 인자가 없으면 정상 실행
	if len(os.Args) < 2 {
		return false, nil
	}

	// 명령 처리
	switch os.Args[1] {
	case "help", "--help", "-h":
		helpCmd.Parse(os.Args[2:])
		printHelp()
		return true, nil

	case "version", "--version", "-v":
		fmt.Println("Dynamic Board 버전 1.0.0")
		return true, nil

	default:
		// 알 수 없는 명령 처리
		fmt.Fprintf(os.Stderr, "알 수 없는 명령: %s\n", os.Args[1])
		printHelp()
		return true, fmt.Errorf("알 수 없는 명령: %s", os.Args[1])
	}
}

// 도움말 출력
func printHelp() {
	progName := filepath.Base(os.Args[0])
	fmt.Printf("사용법: %s [명령] [옵션]\n\n", progName)
	fmt.Println("명령:")
	fmt.Println("  help\t\t이 도움말을 표시합니다")
	fmt.Println("  version\t버전 정보를 표시합니다")
	fmt.Println()
}