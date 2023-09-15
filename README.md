# netscaleradc-acme-go
## Introduction

Let's Encrypt for NetScaler ADC (aka LENS) is a tool which allows you to generate certificates based on the well-known ACME protocol. It is based on the fantastic library from the people at [https://github.com/go-acme/lego](https://github.com/go-acme/lego) to provide the functionality to talk to different DNS providers, but now also NetScaler ADC.

## System requirements
### Operating system
We provide binaries for different operating systems and architectures:
- Linux (amd64/arm64)
- MacOS (Intel/Apple Silicon)
- Windows (amd64/arm64)
- FreeBSD (amd64/arm64), versions > FreeBSD 11

#### Native NetScaler ADC binary
Lens was initially designed to be run on NetScaler appliances directly as well.</br>
Since NetScaler is based on a **heavily modified** FreeBSD, this shouldn't pose any problems.

This tool is based on Go 1.21, because we wanted to make use of the standard library structured logging (slog) to provide a uniform logging infrastructure across all our tools/libraries.
However, since Go 1.19, support for FreeBSD 11 and older was dropped, given that the last version of FreeBSD 11 (11.4) was end-of-life since June 23, 2020.
For more information, visit [https://tip.golang.org/doc/go1.18#freebsd](https://tip.golang.org/doc/go1.18#freebsd)


The latest release of NetScaler, 14.1 is still based on FreeBSD 11.4
```
> uname -p -r -s -m
FreeBSD 11.4-NETSCALER-14.1 amd64 amd64
```
Again, NetScaler is based on a **heavily modified** FreeBSD operating system, but base OS is still FreeBSD 11.4.
As such, it is currently not possible to run the tool on a NetScaler directly.

### Certificate Authorities

By default we support both staging and production environments for Let's Encrypt.
Lens is designed to work with other certificate authorities who provide access through the ACME protocol.

*If you are a user of other ACME-protocol based services, such as Sectigo, please reach out so we can ensure maximum compatibility!*

### NetScaler ADC
#### User permissions
- You will need a user account with the following permissions:
    - \<TBD>

For easy configuration, we provide the necessary commands to create the command policy on NetScaler ADC in the section below.
##### CLI Commands
```text
</TBD>
```
##### Terraform
```terraform
</TBD>
```

#### Running on NetScaler ADC natively
If you run the binary natively on NetScaler ADC:
- You will need internet access to connect to your ACME service of choice if you want to run natively on NetScaler ADC
- You wil need connectivity with either the NSIP or SNIP address for the environments to which you will connect.

## Running Lens
```
Let's Encrypt for NetScaler ADC

Usage:
  lens [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  daemon      Daemon mode
  help        Help about any command
  request     Request mode

Flags:
  -f, --file string       config file name (default "config.yaml")
  -h, --help              help for lens
  -l, --loglevel string   log level
  -p, --path string       config file path, do not use with -s
  -s, --search strings    config file search paths, do not use with -p (default [/etc/corelayer/lens,/nsconfig/ssl/LENS,$HOME/.lens,$PWD])

Use "lens [command] --help" for more information about a command.
```

By default, lens will be looking for a global configuration file in the following paths:
- /etc/corelayer/lens
- /nsconfig/ssl/LENS
- $HOME/.lens
- $PWD (the current working directory)

Global Flags:
- -f / --file: allows you to specify a custom global configuration file
- -p / --path: allows you to specify to path for the global configuration file
- -s / --search: allows you to specify multiple search paths

- -l / --loglevel [debug | info | warning | error]: allows you to specify a loglevel, the default is "info"

### Request mode
```
Let's Encrypt for NetScaler ADC - Request Mode

Usage:
  lens request [flags]

Flags:
  -a, --all           request all (default true)
  -h, --help          help for request
  -n, --name string   request name

Global Flags:
  -f, --file string       config file name (default "config.yaml")
  -l, --loglevel string   log level
  -p, --path string       config file path, do not use with -s
  -s, --search strings    config file search paths, do not use with -p (default [/etc/corelayer/lens,/nsconfig/ssl/LENS,$HOME/.lens,$PWD])
```

You can either request one single certificate or all configured certificates at once.
For a single certificate, you need to specify the name of the configuration using the -n/--name flag.

Flags:
- -a / --all: make a request for all configured certificates
- -n / --name: specify the certificate to be requested

*Both flags are mutually exclusive!*

The global flags are still applicable and can be used accordingly.


### Daemon mode
```
Let's Encrypt for NetScaler ADC - Daemon Mode

Usage:
  lens daemon [flags]

Flags:
  -h, --help   help for daemon

Global Flags:
  -f, --file string       config file name (default "config.yaml")
  -l, --loglevel string   log level
  -p, --path string       config file path, do not use with -s
  -s, --search strings    config file search paths, do not use with -p (default [/etc/corelayer/lens,/nsconfig/ssl/LENS,$HOME/.lens,$PWD])
```

**Not implemented**

The goal is to run lens as a daemon which verifies the actual state of the current certificates and request new ones accordingly.

### Configuration mode

**Not implemented**

The goal is to be able to configure lens from the command line.

## Configuration

Configuration for lens is done using YAML files and is split up in 2 parts:
- global configuration
- certificate configuration

As the global configuration needs to have all account details for the different environments to which you need to connect, this file is separated and can be stored in a secured location with appropriate permissions.
- On Linux, this would typically be /etc/corelayer/lens for example, which can have the necessary permissions to only allow root access or access from the user account intended to run lens.

The individual certificate configuration files can be stored elsewhere on the system with more permissive access, as it will only contain the certificate configuration. References to the environments are made using the name of the configured environment.

### Global configuration
```
configPath: <path to the individua certificate configuration files>
daemon:
  address: <ip address>
  port: <port>>
organizations:
  - name: <organization name>
    environments:
      - name: <environment name>
        type: <standalone | hapair | cluster>
        snip:
          name: <name for the SNIP address>
          address: <SNIP address>
        nodes:
          - name: <hostname>
            address: <NSIP address>
        credentials:
          username: <username>
          password: <password>
        connectionSettings:
          useSsl: <true | false>
          timeout: 3000
          validateServerCertificate: <true | false>
          logTlsSecrets: <true | false>
          autoLogin: <true | false>
users:
  - name: <acme username>
    email: <acme e-mail address>
```

#### Examples
- [Standalone - using SNIP](#standalone---using-snip)
- [Standalone - using NSIP](#standalone---using-nsip)
- [High-Availability pair - using SNIP](#high-availability-pair---using-snip)
- [High-Availability pair - using NSIP](#high-availability-pair---using-nsip)

##### Standalone - using SNIP
Global configuration:
```yaml
configPath: conf.d
daemon:
  address: 127.0.0.1
  port: 12345
organizations:
  - name: corelayer
    environments:
      - name: development
        type: standalone
        snip:
          name: vpx-dev-snip
          address: 192.168.1.10
        nodes:
          - name: vpx-dev-001
            address: 192.168.1.11
        credentials:
          username: nsroot
          password: nsroot
        connectionSettings:
          useSsl: true
          timeout: 3000
          validateServerCertificate: false
          logTlsSecrets: false
          autoLogin: false
users:
  - name: corelayer_acme
    email: fake@email.com
```

##### Standalone - using NSIP
Global configuration:
```yaml
configPath: conf.d
daemon:
  address: 127.0.0.1
  port: 12345
organizations:
  - name: corelayer
    environments:
      - name: development
        type: standalone
        nodes:
          - name: vpx-dev-001
            address: 192.168.1.11
        credentials:
          username: nsroot
          password: nsroot
        connectionSettings:
          useSsl: true
          timeout: 3000
          validateServerCertificate: false
          logTlsSecrets: false
          autoLogin: false
users:
  - name: corelayer_acme
    email: fake@email.com
```

##### High-availability Pair - using SNIP
Global configuration:
```yaml
configPath: conf.d
daemon:
  address: 127.0.0.1
  port: 12345
organizations:
  - name: corelayer
    environments:
      - name: development
        type: hapair
        snip:
          name: vpx-dev-snip
          address: 192.168.1.10
        nodes:
          - name: vpx-dev-001
            address: 192.168.1.11
          - name: vpx-dev-002
            address: 192.168.1.12
        credentials:
          username: nsroot
          password: nsroot
        connectionSettings:
          useSsl: true
          timeout: 3000
          validateServerCertificate: false
          logTlsSecrets: false
          autoLogin: false
users:
  - name: corelayer_acme
    email: fake@email.com
```

##### High-availability Pair - using NSIP
Global configuration:
```yaml
configPath: conf.d
daemon:
  address: 127.0.0.1
  port: 12345
organizations:
  - name: corelayer
    environments:
      - name: development
        type: hapair
        nodes:
          - name: vpx-dev-001
            address: 192.168.1.11
          - name: vpx-dev-002
            address: 192.168.1.12
        credentials:
          username: nsroot
          password: nsroot
        connectionSettings:
          useSsl: true
          timeout: 3000
          validateServerCertificate: false
          logTlsSecrets: false
          autoLogin: false
users:
  - name: corelayer_acme
    email: fake@email.com
```

### Certificate configuration
```yaml
name: <name>
acmeRequest:
  organization: <organization name>
  environment: <environment name>
  username: <acme username>
  service: LE_STAGING | LE_PRODUCTION | <custom address>
  type: <netscaler-http-global | netscaler-adns | http | <name of dns provider>
  keytype: <RSA20248 | RSA4096 | RSA8192 | EC256 | EC384>
  commonName: <common name>
  subjectAlternativeNames:
    - <subjectAlternativeName>
    - <subjectAlternativeName>
bindpoints:
  - organization: <organization name>
    environment: <environment name>
    sslVservers:
      - name: <ssl vserver name>
        sniEnabled: <true | false>
```
As you can see, the configuration is split up in two parts:
- acme request
- bindpoints

#### ACME Request
This section holds all the details to be able to request a certificate from your ACME service of choice.
We need to specify the organization and environment name to select which NetScaler to talk to.

#### Bindpoints
Once the certificate request is done, we can install the certificate onto multiple ssl vservers in multiple environments.
This is especially useful when having SAN-certificates or wildard certificates, so they can be bound appropriately on different NetScaler environments.

#### Examples
- [Simple certificate](#simple-certificate)
- [SAN certificate - using manual entries](#san-certificate---using-manual-entries)
- [SAN certificate - using external file](#san-certificate---using-external-file)

##### Simple certificate
Certificate configuration:
```yaml
name: corelogic_dev
acmeRequest:
  organization: corelayer
  environment: development
  username: corelayer_acme
  service: LE_STAGING
  type: netscaler-http-global
  keytype: RSA4096
  commonName: corelogic.dev.corelayer.eu
bindpoints:
  - organization: corelayer
    environment: development
    sslVservers:
      - name: CSV_DEV_SSL
        sniEnabled: true
```

##### SAN certificate - using manual entries
Certificate configuration:
```yaml
name: corelogic_dev
acmeRequest:
  organization: corelayer
  environment: development
  username: corelayer_acme
  service: LE_STAGING
  type: netscaler-http-global
  keytype: RSA4096
  commonName: corelogic.dev.corelayer.eu
  subjectAlternativeNames:
    - demo.dev.corelayer.eu
    - my.dev.corelayer.eu
bindpoints:
  - organization: corelayer
    environment: development
    sslVservers:
      - name: CSV_DEV_SSL
        sniEnabled: true
```

##### SAN certificate - using external file
Certificate configuration:
```yaml
name: corelogic_dev
acmeRequest:
  organization: corelayer
  environment: development
  username: corelayer_acme
  service: LE_STAGING
  type: netscaler-http-global
  keytype: RSA4096
  commonName: corelogic.dev.corelayer.eu
  subjectAlternativeNamesFile: corelogic_dev_san.txt
bindpoints:
  - organization: corelayer
    environment: development
    sslVservers:
      - name: CSV_DEV_SSL
        sniEnabled: true
```

Subject Alternative Names File (stored next to the certificate configuration file):
```text
demo.dev.corelayer.eu
my.dev.corelayer.eu
```