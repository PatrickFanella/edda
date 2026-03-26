package main

import (
	"fmt"
	"os"

	"github.com/PatrickFanella/game-master/internal/config"
	"github.com/PatrickFanella/game-master/internal/llm"
)

func main() {
	cfg, err := config.Load(os.Getenv("GM_CONFIG"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	if _, err := llm.NewLLMProvider(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "initialize llm provider: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "game-master server scaffold ready (provider=%s)\n", cfg.LLM.Provider)
}
