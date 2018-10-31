package main

// Script to check hashrate of miner and reset if it falls below expected hashrate
// Script will kill and restart the miner a few times and recheck hashrate
// If hashrate continues to be low, it will restart the computer
// Requries 2 arguments. First is wallet address/api key, second is expected hashrate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type hash struct {
	Status string `json:"status"`
	Data   struct {
		Time             int         `json:"time"`
		LastSeen         int         `json:"lastSeen"`
		ReportedHashrate int         `json:"reportedHashrate"`
		CurrentHashrate  float64     `json:"currentHashrate"`
		ValidShares      int         `json:"validShares"`
		InvalidShares    int         `json:"invalidShares"`
		StaleShares      int         `json:"staleShares"`
		AverageHashrate  float64     `json:"averageHashrate"`
		ActiveWorkers    int         `json:"activeWorkers"`
		Unpaid           int64       `json:"unpaid"`
		Unconfirmed      interface{} `json:"unconfirmed"`
		CoinsPerMin      float64     `json:"coinsPerMin"`
		UsdPerMin        float64     `json:"usdPerMin"`
		BtcPerMin        float64     `json:"btcPerMin"`
	} `json:"data"`
}

func main() {
	// Variables
	retry := 1       // Set initial retry
	wt := os.Args[1] // Your wallet address.
	er := os.Args[2] // Expected reported hashrate of the miner.
	for {            // Start the loop
		url := "https://api.ethermine.org/miner/" + wt + "/currentStats"
		cur := currentHashRate(url)
		hr, _ := strconv.Atoi(er) // Convert Expected hashrate into int
		// Compare actual hashrate from the pool and what your miner should be getting
		if !checkHashRate(cur, hr) {
			fixMiner(retry)
			retry++
		} else {
			retry = 1 // Reset retry counter if we get success
		}
		time.Sleep(30 * time.Second)
	}
}

func currentHashRate(url string) int {
	webClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla")

	res, getErr := webClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	hashrate := hash{}
	jsonErr := json.Unmarshal(body, &hashrate)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return hashrate.Data.ReportedHashrate / 1000000
}

func checkHashRate(rate int, expected int) bool {
	if rate < expected {
		color.Red("Hashrate Bad: %d", rate)
		return false
	}
	color.Green("Hashrate Good: %d", rate)
	return true
}

func fixMiner(retry int) {
	if retry > 3 {
		reboot()
	}
	fmt.Println("Fixing Miner")
	if runtime.GOOS == "windows" {
		fmt.Println("Fixing for Windows")
		if !minerRunningWin() {
			startMiner()
			retry++
		} else {
			// Restart it and add to retry var
			fmt.Println("Restarting Miner")
			killMiner()
			startMiner()
			retry++
		}
	} else if runtime.GOOS == "linux" {
		fmt.Println("Fixing for Linux")
		if !minerRunningLin() {
			startMiner()
			retry++
		} else {
			// Restart it and add to retry var
			fmt.Println("Restarting Miner")
			killMiner()
			startMiner()
			retry++
		}
	}
}

func minerRunningWin() bool {
	// Windows sucks, need to find a better way to run a pipe command or get this working:
	// tasklist | find /I /C "EthDcrMiner" -- this would spit out number of those prcesses as int
	// For now, we make it janky
	p, err := exec.Command("cmd.exe", "/C", "tasklist").Output()
	if err != nil {
		log.Fatal("EXEC ERROR:", err)
	}
	pid := strings.TrimSpace(string(p))
	if strings.Contains(pid, "EthDcrMiner") {
		fmt.Println("Miner is running")
		return true
	}
	return false
}

func minerRunningLin() bool {
	cmd := "ps aux | grep EthDcrMiner | wc -l"
	p, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal("EXEC ERROR:", err)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(p)))
	if pid > 0 {
		fmt.Println("Miner is running")
		return true
	}
	return false
}

func killMiner() {
	if runtime.GOOS == "windows" {
		cmd := "taskkill.exe /IM EthDcrMiner* /F"
		exec.Command("cmd", "/C", cmd).Output()
	} else {
		cmd := "pkill -f EthDcrMiner"
		exec.Command("bash", "-c", cmd).Output()
	}
}

func startMiner() {
	fmt.Println("Starting Miner")
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "start", "cmd", "/C", "start-miner.lnk") // FTW Winbloze
		cmd.Start()                                                               // Fork miner
	} else {
		cmd := exec.Command("bash", "-c", "./start-miner.sh")
		cmd.Start() // Fork miner
	}
	time.Sleep(600 * time.Second) // Wait 10 minutes for status to populate on pool
}

func reboot() {
	fmt.Println("Rebooting Machine; too many failures in a row")
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/C", "shutdown /r /t 00").CombinedOutput() // Bye Winders
	} else {
		exec.Command("bash", "-c", "sudo reboot").CombinedOutput()
	}
	time.Sleep(5 * time.Second)
}
