package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/operatorai/kettle-cli/command"
	"github.com/operatorai/kettle-cli/config"
	"github.com/operatorai/kettle-cli/templates"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project from a template.",
	Long: `🆕 The kettle CLI tool automatically creates a directory
 with all of the boiler plate that you need from a template.
	
The create command will create a directory with all the code to get you started.`,
	Args: validateCreateArgs,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func validateCreateArgs(cmd *cobra.Command, args []string) error {
	// Validate that a template was given
	if len(args) == 0 {
		return errors.New("please specify a template")
	}
	return nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Get the directory where the template is (or has been cloned to)
	templatePath, isTempDir, err := templates.GetTemplate(args[0])
	if err != nil {
		return formatError(err)
	}
	if isTempDir {
		defer os.RemoveAll(templatePath)
	}

	// Read the template config
	templateConfig, err := templates.ReadConfig(templatePath)
	if err != nil {
		return formatError(err)
	}

	// Create the directory where the template will be populated
	projectName, directoryPath, err := createProjectDirectory()
	if err != nil {
		return formatError(err)
	}

	// Ask the user for any input that is required
	templateConfig.ProjectName = projectName
	templateValues := map[string]string{
		"ProjectName": projectName,
	}
	for i, templateEntry := range templateConfig.Template {
		userInput, err := command.PromptForString(templateEntry.Prompt)
		if err != nil {
			return cleanUp(directoryPath, err)
		}
		templateConfig.Template[i].Value = userInput
		templateValues[templateEntry.Key] = userInput
	}

	// The template files are in a subdirectory of templatePath
	templateDirectory := path.Join(templatePath, "template")
	err = filepath.Walk(templateDirectory, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			if config.DebugMode {
				fmt.Printf("error accessing a path %q: %v\n", filePath, err)
				return err
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Create the target path
		targetPath := strings.Replace(filePath, templateDirectory, "", 1)
		targetPath = path.Join(directoryPath, targetPath)

		// Create the target file
		if err := createFile(targetPath, filePath, templateValues); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return cleanUp(directoryPath, err)
	}

	err = templates.WriteConfig(directoryPath, templateConfig)
	if err != nil {
		return cleanUp(directoryPath, err)
	}
	fmt.Println("\n✅  Created: ", directoryPath)
	return nil
}

func createProjectDirectory() (string, string, error) {
	directoryName, err := command.PromptForString("Directory name")
	if err != nil {
		return "", "", err
	}

	directoryPath, err := templates.GetRelativeDirectory(directoryName)
	if err != nil {
		return "", "", err
	}

	// Validate that the function path does *not* already exist
	exists, err := templates.PathExists(directoryPath)
	if err != nil {
		return "", "", err
	}
	if exists {
		return "", "", fmt.Errorf("directory already exists: %s", directoryPath)
	}

	// Create a directory with the function name
	if err := os.Mkdir(directoryPath, os.ModePerm); err != nil {
		return "", "", err
	}
	return directoryName, directoryPath, nil
}

func createFile(targetPath, filePath string, templateValues interface{}) error {
	// Read the source file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Create the parent directory
	parentDir, _ := path.Split(targetPath)
	err = os.MkdirAll(parentDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Create the target file
	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Populate the target file by executing the template
	_, fileName := path.Split(filePath)
	tmpl, err := template.New(fileName).Parse(string(data))
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, templateValues)
	if err != nil {
		return err
	}
	return nil
}

func cleanUp(directoryPath string, err error) error {
	cleanupErr := os.RemoveAll(directoryPath)
	if cleanupErr != nil {
		fmt.Println("\n Failed to clean up: ", directoryPath, cleanupErr)
	}
	return err
}
