package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/heroku/color"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/buildpacks/pack/internal/commands"
	"github.com/buildpacks/pack/internal/config"
	ilogging "github.com/buildpacks/pack/internal/logging"
	h "github.com/buildpacks/pack/testhelpers"
)

func TestAddBuildpackRegistry(t *testing.T) {
	color.Disable(true)
	defer color.Disable(false)

	spec.Run(t, "Commands", testAddBuildpackRegistryCommand, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testAddBuildpackRegistryCommand(t *testing.T, when spec.G, it spec.S) {
	when("AddBuildpackRegistry", func() {
		var (
			tmpDir     string
			configFile string
			outBuf     bytes.Buffer
			command    *cobra.Command
			logger     = ilogging.NewLogWithWriters(&outBuf, &outBuf)
			assert     = h.NewAssertionManager(t)
		)

		it.Before(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "pack-home-*")
			assert.Nil(err)

			configFile = filepath.Join(tmpDir, "config.toml")
			command = commands.AddBuildpackRegistry(logger, config.Config{}, configFile)
		})

		it.After(func() {
			assert.Nil(os.RemoveAll(tmpDir))
		})

		when("add buildpack registry", func() {
			it("adds to registry list", func() {
				command.SetArgs([]string{"bp", "https://github.com/buildpacks/registry-index/"})
				assert.Succeeds(command.Execute())

				cfg, err := config.Read(configFile)
				assert.Nil(err)
				assert.Equal(len(cfg.Registries), 1)
				assert.Equal(cfg.Registries[0].Name, "bp")
				assert.Equal(cfg.Registries[0].Type, "github")
				assert.Equal(cfg.Registries[0].URL, "https://github.com/buildpacks/registry-index/")
				assert.Equal(cfg.DefaultRegistryName, "")
			})
		})

		when("default is true", func() {
			it("sets newly added registry as the default", func() {
				command.SetArgs([]string{"bp", "https://github.com/buildpacks/registry-index/", "--default"})
				assert.Succeeds(command.Execute())

				cfg, err := config.Read(configFile)
				assert.Nil(err)
				assert.Equal(len(cfg.Registries), 1)
				assert.Equal(cfg.Registries[0].Name, "bp")
				assert.Equal(cfg.DefaultRegistryName, "bp")
			})
		})

		when("validation", func() {
			it("fails with missing args", func() {
				command.SetOut(ioutil.Discard)
				command.SetArgs([]string{})
				err := command.Execute()
				assert.ErrorContains(err, "accepts 2 arg")
			})

			it("should validate type", func() {
				command.SetArgs([]string{"bp", "https://github.com/buildpacks/registry-index/", "--type=bogus"})
				assert.Error(command.Execute())

				output := outBuf.String()
				h.AssertContains(t, output, "'bogus' is not a valid type. Supported types are: 'git', 'github'.")
			})

			it("should throw error when registry already exists", func() {
				command := commands.AddBuildpackRegistry(logger, config.Config{
					Registries: []config.Registry{
						{
							Name: "bp",
							Type: "github",
							URL:  "https://github.com/buildpacks/registry-index/",
						},
					},
				}, configFile)
				command.SetArgs([]string{"bp", "https://github.com/buildpacks/registry-index/"})
				assert.Error(command.Execute())

				output := outBuf.String()
				h.AssertContains(t, output, "Buildpack registry 'bp' already exists.")
			})
		})
	})
}
