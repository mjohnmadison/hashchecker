# HashChecker
Keeps an eye on my ETH miner hash rate using ethermine.org and Claymore Miner; resets if reported hash rate falls below threshold. Drivers crash, crap happens and sometimes you need some maintenance on your rig.

Download pre-compiled binary for Windows or Linux in Releases.

### Usage
```
hashrate wallet_address expected_hashrate
```
I use a batch script or shell script to start my miner. It's a hard coded filename (start-miner.bat or start-miner.sh) and should be in the same directory as this Go binary.

Both the start-miner script and this script start on boot of the OS with offset sleep intervals (start-miner first obviously). I have the power options in the BIOS to "Always On" as well in case of power failure.

This allows your miners to be completely self healing and hands off.

### Develop/Dependencies
```
go get github.com/fatih/color
```
I've added a Linux path although I don't mine on Linux anymore. Not sure it works 100% on *nix (PRs welcome!)

Basic logic:
  - Check the reported hashrate to the pool against what you should be getting
  - Kill/restart the miner if it falls below
  - If you get 3 failures in a row, reboot machine

This keeps my miner running itself, start your miner on boot along with this script and never worry about a driver crashing and losing hash power!

# TODO:
- Add a check for internet connection. I've had it just reboot every 30 minutes until my internet came back. Probably should just wait and check network again.
- Log errors to a file.
