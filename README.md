# hashchecker
Keeps an eye on my ETH miner hash rate using ethermine.org; resets if hash rate falls below threshold.

Download binary for Windows or Linux in Releases.

### Usage
```
hashrate <wallet_address> <expected hashrate>
```
### Develop/Dependencies
```
go get github.com/fatih/color
```
I've added a Linux path although I don't mine on Linux anymore. Not sure it works 100% on *nix

Basic logic:
  - Check the reported hashrate to the pool against what you should be getting
  - Kill/restart the miner if it falls below
  - If you get 3 failures in a row, reboot machine

This keeps my miner running itself, start your miner on boot along with this script and never worry about a driver crashing and losing hash power!

# TODO:
- Add a check for internet connection. I've had it just reboot every 30 minutes until my internet came back. Probably should just wait and check network again.
