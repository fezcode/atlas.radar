//go:build gobake
package bake_recipe

import (
	"fmt"
	"github.com/fezcode/gobake"
)

func Run(bake *gobake.Engine) error {
	// Load metadata
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		return fmt.Errorf("error loading recipe.piml: %v", err)
	}

	// --- Tasks ---

	bake.Task("setup", "Installs required tools", func(ctx *gobake.Context) error {
		return ctx.InstallTools()
	})

	bake.Task("build", "Builds the binary for multiple platforms", func(ctx *gobake.Context) error {
		ctx.Log("Building %s v%s...", bake.Info.Name, bake.Info.Version)

		targets := []struct {
			os   string
			arch string
		}{
			{"linux", "amd64"},
			{"linux", "arm64"},
			{"windows", "amd64"},
			{"windows", "arm64"},
			{"darwin", "amd64"},
			{"darwin", "arm64"},
		}

		err := ctx.Mkdir("build")
		if err != nil {
			return err
		}

		for _, t := range targets {
			output := "build/" + bake.Info.Name + "-" + t.os + "-" + t.arch
			if t.os == "windows" {
				output += ".exe"
			}

			ldflags := "-s -w -X main.Version=" + bake.Info.Version
			ctx.Env = []string{"CGO_ENABLED=0"}
			err := ctx.Run("go", "build", "-ldflags", ldflags, "-o", output, "main.go")
			if err != nil {
				return err
			}
		}
		return nil
	})

	bake.Task("test", "Runs project tests", func(ctx *gobake.Context) error {
		ctx.Log("Running tests...")
		return ctx.Run("go", "test", "./...")
	})

	bake.Task("clean", "Removes build artifacts", func(ctx *gobake.Context) error {
		ctx.Log("Cleaning up...")
		return ctx.Remove("build")
	})
	
	return nil
}
