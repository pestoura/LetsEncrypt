package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/letsencrypt/boulder/cmd"
	"github.com/letsencrypt/boulder/test"
)

// TestConfigValidation checks that each of the components which register a
// validation tagged Config struct at init time can be used to successfully
// validate their corresponding test configuration files.
func TestConfigValidation(t *testing.T) {
	configPath := "../../test"
	if os.Getenv("BOULDER_CONFIG_DIR") == "test/config-next" {
		configPath = "../../test/config-next"
	}

	// Each component is a set of `cmd` package name and a list of paths to
	// configuration files to validate.
	components := make(map[string][]string)

	// For each component, add the paths to the configuration files to validate.
	// By default we assume that the configuration file is named after the
	// component. However, there are some exceptions to this rule. We've added
	// special cases for these components.
	for _, cmdName := range cmd.AvailableConfigs() {
		var fileNames []string
		switch cmdName {
		case "boulder-publisher":
			fileNames = []string{"publisher.json"}
		case "boulder-wfe2":
			fileNames = []string{"wfe2.json"}
		case "boulder-va":
			fileNames = []string{"va.json"}
		case "boulder-remoteva":
			fileNames = []string{
				"va-remote-a.json",
				"va-remote-b.json",
			}
		case "boulder-ra":
			fileNames = []string{"ra.json"}
		case "nonce-service":
			fileNames = []string{
				"nonce-a.json",
				"nonce-b.json",
			}
		case "boulder-ca":
			fileNames = []string{
				"ca-a.json",
				"ca-b.json",
			}
		case "boulder-sa":
			fileNames = []string{"sa.json"}
		case "boulder-observer":
			fileNames = []string{"observer.yml"}

		default:
			fileNames = []string{cmdName + ".json"}
		}
		components[cmdName] = append(components[cmdName], fileNames...)
	}
	t.Parallel()
	for cmdName, paths := range components {
		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				err := cmd.ReadAndValidateConfigFile(cmdName, fmt.Sprintf("%s/%s", configPath, path))
				test.AssertNotError(t, err, fmt.Sprintf("Failed to validate config file %q", path))
			})
		}
	}
}
