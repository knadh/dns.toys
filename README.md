<img width="150" src="https://user-images.githubusercontent.com/547147/171995179-b9d2faae-d659-4260-99df-04c62c171f6f.png" />

dns.toys is a DNS server that takes creative liberties with the DNS protocol to offer handy utilities and services that are easily accessible via the command line.

For docs, visit [**www.dns.toys**](https://www.dns.toys)

## Sample commands
```shell
dig help @dns.toys

dig mumbai.time @dns.toys

dig newyork.weather @dns.toys

dig 42km-mi.unit @dns.toys

dig 100USD-INR.fx @dns.toys

dig myip @dns.toys

dig 987654321.words @dns.toys
```

## Running locally
- Clone the repo
- Copy `config.sample.toml` to `config.toml` and edit the config
- Run `make build` to build the binary. Then run `./dnstoys.bin`
