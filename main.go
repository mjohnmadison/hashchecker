package main

// Version: 1.3
// Script to check hashrate of miner and reset if it falls below expected hashrate
// Script will kill and restart the miner a few times and recheck hashrate
// If hashrate continues to be low, it will restart the computer
// Requries 2 arguments. First is wallet address/api key, second is expected hashrate
// TODO: Log errors or warnings to logfile

import (
	"encoding/json"
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

type stats struct {
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
	retry := 0       // Set initial retry
	wt := os.Args[1] // Your wallet address.
	er := os.Args[2] // Expected reported hashrate of the miner.
	for {            // Start the loop
		if !minerRunning() { // Check miner is running each time
			startMiner()
			retry++
		}
		url := "https://api.ethermine.org/miner/" + wt + "/currentStats"
		cur, err := currentHashRate(url)
		if err != nil {
			log.Println(err)
		}
		hr, _ := strconv.Atoi(er) // Convert Expected hashrate into int
		// Compare actual hashrate from the pool and what your miner should be getting
		if !checkHashRate(cur, hr) {
			fixMiner(retry)
			retry++
		} else {
			retry = 1 // Reset retry counter if we get success
		}
		time.Sleep(60 * time.Second) // Check stats every 60 seconds
	}
}

func currentHashRate(url string) (int, error) {
	webClient := http.Client{
		Timeout: time.Second * 5, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "Mozilla")

	res, getErr := webClient.Do(req)
	if getErr != nil {
		return 0, getErr
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return 0, readErr
	}

	hashrate := stats{}
	jsonErr := json.Unmarshal(body, &hashrate)
	if jsonErr != nil {
		return 0, jsonErr
	}

	return hashrate.Data.ReportedHashrate / 1000000, nil
}

func checkHashRate(rate int, expected int) bool {
	if rate < expected {
		color.Red("Hashrate Bad: %d\n", rate)
		return false
	}
	color.Green("Hashrate Good: %d\n", rate)
	return true
}

func fixMiner(retry int) {
	if retry > 2 {
		reboot()
	}
	color.Yellow("Attemping to reset miner...")
	if minerRunning() {
		// Restart it and add to retry var
		killMiner()
		startMiner()
		retry++
	}
}

func minerRunning() bool {
	if runtime.GOOS == "windows" {
		if minerRunningWin() {
			return true
		}
	} else if runtime.GOOS == "linux" {
		if minerRunningLin() {
			return true
		}
	}
	return false
}

func minerRunningWin() bool {
	c := color.New(color.FgGreen, color.Bold)
	r := color.New(color.FgRed, color.Bold)
	// Windows sucks, need to find a better way to run a pipe command or get this working:
	// tasklist | find /I /C "EthDcrMiner" -- this would spit out number of those prcesses as int
	// For now, we make it janky
	p, err := exec.Command("cmd.exe", "/C", "tasklist").Output()
	if err != nil {
		log.Println("EXEC ERROR:", err)
	}
	pid := strings.TrimSpace(string(p))
	if strings.Contains(pid, "EthDcrMiner") {
		c.Printf("Miner running: ")
		return true
	}
	r.Printf("Miner NOT running: ")
	return false
}

func minerRunningLin() bool {
	c := color.New(color.FgGreen, color.Bold)
	r := color.New(color.FgRed, color.Bold)
	cmd := "ps aux | grep [E]thDcrMiner | wc -l" // Regex to eliminate the grep returning the grep command.
	p, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Println("EXEC ERROR:", err)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(p)))
	if pid > 0 {
		c.Printf("Miner running: ")
		return true
	}
	r.Printf("Miner NOT running: ")
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
	c := color.New(color.FgCyan, color.Bold)
	c.Printf("Starting Miner... ")
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "start", "cmd", "/C", "start-miner.lnk") // FTW Winbloze
		cmd.Start()                                                               // Fork miner
	} else { // Linux
		cmd := exec.Command("bash", "-c", "./start-miner.sh")
		cmd.Start() // Fork miner
	}
	time.Sleep(600 * time.Second) // Wait 10 minutes for stats to populate on pool
}

func reboot() {
	color.Red("Rebooting Machine: too many failures in a row")
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/C", "shutdown /r /t 00").CombinedOutput()
	} else {
		exec.Command("bash", "-c", "sudo reboot").CombinedOutput()
	}
	os.Exit(1)
}
