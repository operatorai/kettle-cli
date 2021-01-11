package command

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
	"github.com/operatorai/operator/config"
)

func getSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
	s.Suffix = "  Working..."
	s.Start()
	return s
}

func Execute(command string, args []string) error {
	osCmd := exec.Command(command, args...)
	if config.DebugMode {
		fmt.Println("\n", command, strings.Join(args, " "))
		osCmd.Stderr = os.Stderr
		osCmd.Stdout = os.Stdout
	}

	s := getSpinner()
	defer s.Stop()
	if err := osCmd.Run(); err != nil {
		return err
	}
	return nil
}

func ExecuteWithResult(command string, args []string) ([]byte, error) {
	osCmd := exec.Command(command, args...)
	if config.DebugMode {
		fmt.Println("\n", command, strings.Join(args, " "))
		osCmd.Stderr = os.Stderr
	}

	s := getSpinner()
	defer s.Stop()
	output, err := osCmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

// PromptForValue shows a prompt (using a map's keys) to the user and returns
// the value that is indexed at that key
func PromptForValue(label string, values map[string]string, addNoneOfThese bool) (string, error) {
	valueLabels := []string{}
	for valueLabel, _ := range values {
		valueLabels = append(valueLabels, valueLabel)
	}
	sort.Strings(valueLabels)
	if addNoneOfThese {
		valueLabels = append(valueLabels, config.PromptNoneOfTheseOption)
	}

	prompt := promptui.Select{
		Label: label,
		Items: valueLabels,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	if addNoneOfThese && result == config.PromptNoneOfTheseOption {
		return "", nil
	}
	return values[result], nil
}

func PromptToConfirm(label string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return false, err
	}

	if strings.ToLower(result) == "y" {
		return true, nil
	}
	return false, nil
}
