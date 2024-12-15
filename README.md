<img width="150" src="https://user-images.githubusercontent.com/547147/171995179-b9d2faae-d659-4260-99df-04c62c171f6f.png" />

dns.toys is a DNS server that takes creative liberties with the DNS protocol to offer handy utilities and services that are easily accessible via the command line.

For docs, visit [**www.dns.toys**](https://www.dns.toys)

## Sample commands for Linux

```shell
dig help @dns.toys

dig mumbai.time @dns.toys

dig 2023-05-28T14:00-bengaluru-berlin/de.time @dns.toys

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

## Using @dns-toys on Windows

### Using PowerShell (Recomended)

You can use Resolve-DnsName cmdlet in Windows PowerShell to interact with @dns-toys:

* **Open PowerShell**
   - Press `Win + X` and select **Windows PowerShell**.

* **Use `Resolve-DnsName`**
     - **Get the current time in a city**:
       ```powershell
       Resolve-DnsName -Name mumbai.time -Server dns.toys
       ```
     - **Check the weather**:
       ```powershell
       Resolve-DnsName -Name newyork.weather -Server dns.toys
       ```
     - **Convert units**:
       ```powershell
       Resolve-DnsName -Name 42km-mi.unit -Server dns.toys
       ```
     - **Currency conversion**:
       ```powershell
       Resolve-DnsName -Name 100USD-INR.fx -Server dns.toys
       ```

* **Format the Output**:
   - To make the output more readable, you can format it using `Format-List`:
     ```powershell
     Resolve-DnsName -Name mumbai.time -Server dns.toys | Format-List
     ```


### Using dig

* **Install `dig` Utility**:
   - You can install the `dig` utility, which is part of the BIND tools, using a package manager like Chocolatey:
     ```powershell
     choco install bind-toolsonly
     ```

* **Using `dig` with @dns-toys**

    Once `dig` is installed, you can use it to query @dns-toys. Here are some examples:

- **Get the current time in a city**:
  ```powershell
  dig mumbai.time @dns.toys
  ```
- **Currency conversion**:
  ```powershell
  dig 100USD-INR.fx @dns.toys
  ```

## Running locally

- Clone the repo
- Copy `config.sample.toml` to `config.toml` and edit the config
- Make sure you have a copy of the `cities15000.txt` file at the root of this directory (instructions are in the `config.sample.toml` file)
- Make sure to download the `wordnet` from [Wordnet website](https://wordnetcode.princeton.edu/3.0/WNdb-3.0.tar.gz).(more instructions are in the `config.sample.toml` file)
- Extract the tarball and rename extracted the directory to `wordnet`
- Run `make build` to build the binary and then run `./dnstoys.bin`
- Query against the locally running server
```shell
    dig <query> +short @127.0.0.1 -p 5354
```

## Others

- [DnsToys.NET](https://github.com/fatihdgn/DnsToys.NET) - A .net client library for the service.
