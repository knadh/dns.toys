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

dig ip @dns.toys

dig 987654321.words @dns.toys

dig pi @dns.toys

dig 100dec-hex.base @dns.toys

dig fun.dict @dns.toys

dig A12.9352,77.6245/12.9698,77.7500.aerial @dns.toys
```

## Running locally

- Clone the repo
- Copy `config.sample.toml` to `config.toml` and edit the config
- Make sure you have a copy of the `cities15000.txt` file at the root of this directory (instructions are in the `config.sample.toml` file)
- Run `make build` to build the binary and then run `./dnstoys.bin`

## Others

- [DnsToys.NET](https://github.com/fatihdgn/DnsToys.NET) - A .net client library for the service.
