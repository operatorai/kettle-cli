package config

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func getConfigPath() string {
	// Find the home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return path.Join(home, ".operator.yaml")
}

func setConfigPath() {
	// Find the home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Search config in home directory with name ".operator" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigName(".operator")
	viper.SetConfigType("yaml")
}

func Read() error {
	setConfigPath()
	viper.SetDefault(DeploymentType, GoogleCloudRun)
	viper.SetDefault(Runtime, Python)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

func Write() {
	configPath := getConfigPath()
	if err := viper.SafeWriteConfigAs(configPath); err != nil {
		if os.IsNotExist(err) {
			log.Println("Creating new config file")
			err = viper.WriteConfigAs(configPath)
		}
	}
	viper.WriteConfigAs(configPath)
}