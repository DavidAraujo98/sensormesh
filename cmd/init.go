/*
Copyright © 2023 David Araújo <davidaraujo98@github.io>
*/
package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sensormesh/cmd/shared"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/spf13/cobra"
)

var (
	IpfsApi          *shell.Shell
	IpfsPath         string
	swarmkey         string
	nodename         string
	swarmKeyFilePath = filepath.Join(os.Getenv("IPFS_PATH"), "swarm.key")
)

func createSwarmKeyFile() {
	// Clear old swarm key file
	os.Remove(swarmKeyFilePath)

	// If swarm key not provided, creates new one
	if swarmkey == "" {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			panic(fmt.Errorf("error while trying to read random source for swarm key: %v", err))
		}
		swarmkey = hex.EncodeToString(key)
	}

	var (
		file *os.File
		err  error
	)
	exists, _ := shared.Exists(swarmKeyFilePath)
	if !exists {
		err = os.MkdirAll(filepath.Dir(swarmKeyFilePath), 0700)
		if err != nil {
			panic(fmt.Errorf("error creating directories: %v", err))
		}
		file, err = os.Create(swarmKeyFilePath)
		if err != nil {
			panic(fmt.Errorf("error generating swarm file: %v", err))
		}
		defer file.Close()
		fmt.Fprintln(file, "/key/swarm/psk/1.0.0/")
		fmt.Fprintln(file, "/base16/")
		fmt.Fprintln(file, swarmkey)
	}
	fmt.Println("[+] Swarm key: ", swarmkey)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local FireMesh configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Checking if IPFS configs exist
		// TODO FEATURE - Allow to user to specify the IPFS repo path
		location := filepath.Join(os.Getenv("IPFS_PATH"), "config")
		_, err := shared.Exists(filepath.Join(os.Getenv("IPFS_PATH"), "config"))
		if err != nil {
			panic(fmt.Errorf("configuration file not set at "+location+". Try running 'ipfs init' first: %s", err))
		}

		// Creates new swarm key file
		createSwarmKeyFile()

		// Load sensormesh configurations to Viper
		shared.LoadConfigurationFromFile()

		// Set the node's initial configurati
		shared.ViperConfs.Set("name", nodename)
		shared.ViperConfs.Set("logfile", shared.LogFilePath)
		shared.ViperConfs.Set("swarmkey", swarmkey)
		shared.ViperConfs.WriteConfig()

		fmt.Println("[+] New sensormesh node " + shared.ViperConfs.GetString("name") + " created !")
	},
}

func init() {
	initCmd.Flags().StringVar(&swarmkey, "swarmkey", "", "IPFS private network swarm key, if none provided, creates a new one")
	initCmd.Flags().StringVar(&nodename, "nodename", "SensorMeshNode", "IPFS private network swarm key, if none provided, creates a new one")
	initCmd.Flags().StringVar(&shared.LogFilePath, "logfile", shared.LogFilePath, "Path destination for logfile, Defaults to '~/.sensormesh/sensormesh.log'")
	rootCmd.AddCommand(initCmd)
}
