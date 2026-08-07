package runner

import (
	"fmt"

	"github.com/yurivict/ollama/x/mlxrunner"
)

func Execute(args []string) error {
	if args[0] == "runner" {
		args = args[1:]
	}

	if len(args) > 0 {
		switch args[0] {
		case "--mlx-engine":
			return mlxrunner.Execute(args[1:])
		}
	}
	return fmt.Errorf("unknown runner engine, expected --mlx-engine")
}
