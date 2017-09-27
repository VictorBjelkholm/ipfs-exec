package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/mholt/archiver"
	homedir "github.com/mitchellh/go-homedir"
)

func main() {
	// TODO
	// Implement special commands such as:
	// randomize-ports: assigns 0 to all the ports in a configuration
	// reset-repository: removes and inits it again
	// remove-repository: removes repository
	if len(os.Args) < 2 {
		fmt.Println("Missing ipfs version to use")
		return
	}
	version := os.Args[1]
	if len(os.Args) < 3 {
		if version == "ls" {
			fmt.Println("Checking versions")
			resp, err := http.Get("https://dist.ipfs.io/go-ipfs/versions")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode == 200 { // OK
				bodyBytes, _ := ioutil.ReadAll(resp.Body)
				bodyString := string(bodyBytes)
				fmt.Println(bodyString)
			}
			return
		}
		fmt.Println("Missing ipfs subcommand")
		return
	}
	command := os.Args[2:]
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	repoPath := path.Join(home, ".ipfs-exec", version)
	binPath := path.Join(home, ".ipfs-exec", "bin")
	thisVersionPath := path.Join(binPath, "ipfs-"+version)
	// Debug output
	// fmt.Println("Version", version)
	// fmt.Println("Command", command)
	// fmt.Println("Repo path", repoPath)
	// fmt.Println("Binaries path", binPath)
	// fmt.Println("This version's path", thisVersionPath)

	// https://dist.ipfs.io/go-ipfs/v0.4.11-rc2/go-ipfs_v0.4.11-rc2_darwin-amd64.tar.gz
	if _, err := os.Stat(thisVersionPath); os.IsNotExist(err) {
		fmt.Println("Downloading version...")
		out, err := ioutil.TempFile("", version)
		if err != nil {
			panic(err)
		}
		defer out.Close()
		url := "https://dist.ipfs.io/go-ipfs/" + version + "/go-ipfs_" + version + "_darwin-amd64.tar.gz"
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			panic("Did not manage to download version " + version + " from url " + url)
		}
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			panic(err)
		}
		dir, err := ioutil.TempDir("", version+"-extracted")
		if err != nil {
			panic(err)
		}
		archiver.TarGz.Open(out.Name(), dir)
		finalBinary := path.Join(dir, "go-ipfs", "ipfs")

		os.MkdirAll(binPath, 0777)
		err = os.Rename(finalBinary, thisVersionPath)
		if err != nil {
			panic(err)
		}
		RunIPFSCmd(repoPath, thisVersionPath, []string{"init"})
		// Now replace ports!

		// run ipfs config Addresses.Swarm
		// take output and replace 4001 with 0
		// run ipfs config --json Addreses.Swarm $output
		ReplacePortInAddress(repoPath, thisVersionPath, "Addresses.Swarm", "4001", true)
		ReplacePortInAddress(repoPath, thisVersionPath, "Addresses.API", "5001", false)
		ReplacePortInAddress(repoPath, thisVersionPath, "Addresses.Gateway", "8080", false)
		// ipfs config --json Addresses.Swarm "$()"
		// replacePortCmd := thisVersionPath + " config Addresses.Swarm | sed 's|4001|0|g'"
		// cmd := []string{"config", "--json", "Addresses.Swarm", "\"$(" + replacePortCmd + ")\""}
		// RunIPFSCmd(repoPath, thisVersionPath, cmd)
	}
	fmt.Println("Running " + thisVersionPath)
	// Now, execute the command
	cmd := exec.Command(thisVersionPath, command...)
	cmd.Env = append(os.Environ(),
		"IPFS_PATH="+repoPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

func RunIPFSCmd(repo string, version string, command []string) {
	cmd := exec.Command(version, command...)
	cmd.Env = append(os.Environ(),
		"IPFS_PATH="+repo,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func RunIPFSCmdWithOutput(repo string, version string, command []string) string {
	cmd := exec.Command(version, command...)
	cmd.Env = append(os.Environ(),
		"IPFS_PATH="+repo,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return string(out)
}

func ReplacePortInAddress(repo string, version string, address string, oldPort string, json bool) {
	swarmOutput := RunIPFSCmdWithOutput(repo, version, []string{"config", address})
	modifiedSwarmOutput := strings.Replace(swarmOutput, oldPort, "0", -1)
	modifiedSwarmOutput = strings.Replace(modifiedSwarmOutput, "\n", "", -1)
	args := []string{"config", address, modifiedSwarmOutput}
	if json {
		args = append(args, "--json")
	}
	RunIPFSCmd(repo, version, args)
}
