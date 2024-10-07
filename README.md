# Caduceus
The Caduceus is a symbol of Hermes or Mercury in Greek and Roman mythology. Caduceus symbol is identified with thieves, merchants, and messengers, and Mercury is said to be a patron of thieves and outlaws

Caduceus is a tool to scan IPs or CIDRs for certificates. This allows finding hidden domains, new organizations, etc.

Input Support: Either IPs & CIDRs separated by commas, or a file with IPs/CIDRs on each line, or file contains ip:port format list. Or stdin.

Inspired by [CloudRecon](https://github.com/g0ldencybersec/CloudRecon)

# Install
** You must have CGO enabled, and may have to install gcc to run CloudRecon**
```sh
sudo apt install gcc
```

```sh
go install github.com/g0ldencybersec/Caduceus/cmd/caduceus@latest
```

Note:
Don't forget to [set your `GOPATH`](https://github.com/golang/go/wiki/SettingGOPATH) before installing.

# Options
```sh
-i <IPs/CIDRs or File> 
  -c int
        How many goroutines running concurrently (default 100)
  -debug
        Add this flag if you want to see failures/timeouts
  -h    Show the program usage message
  -i string
        Either IPs & CIDRs separated by commas, or a file with IPs/CIDRs on each line (default "NONE")
        TO USE STDIN, DONT USE THIS FLAG
  -j    print cert data as jsonl
  -p string
        TLS ports to check for certificates (default "443")
  -t int
        Timeout for TLS handshake (default 4)
  -wc
        print wildcards to stdout
```