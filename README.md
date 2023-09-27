# netscaleradc-acme-go
## Introduction

Let's Encrypt for NetScaler ADC (aka LENS) is a tool which allows you to generate certificates based on the well-known ACME protocol. It is based on the fantastic library from the people at [https://github.com/go-acme/lego](https://github.com/go-acme/lego) to provide the functionality to talk to different DNS providers, but now also NetScaler ADC.

## Changelog
### v0.3.1
- Changed global application flags to accomodate a global configuration file and environment variables file flag
  - changed -f / --config to -c / --configFile
  - added -e / --envFile
  - This also frees up the -f parameter to be changed later to a --force parameter

- Added provider parameters to global configuration file for use with DNS providers which require environment variables to be set when being used.
  - For more information on available providers, see [https://go-acme.github.io/lego/dns/](https://go-acme.github.io/lego/dns/)

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

By default, we support both staging and production environments for Let's Encrypt.
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
```text
</TBD>
```

#### Running on NetScaler ADC natively
If you run the binary natively on NetScaler ADC:
- You will need internet access to connect to your ACME service of choice if you want to run natively on NetScaler ADC
- You wil need connectivity with either the NSIP or SNIP address for the environments to which you will connect.

## Running Lens
```
    __    _______   _______
   / /   / ____/ | / / ___/
  / /   / __/ /  |/ /\__ \
 / /___/ /___/ /|  /___/ /
/_____/_____/_/ |_//____/

Let's Encrypt for NetScaler ADC

Usage:
  lens [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  request     Request mode

Flags:
  -c, --configFile string   config file name (default "config.yaml")
  -e, --envFile string      environment file name (default "variables.env")
  -h, --help                help for lens
  -l, --loglevel string     log level
  -p, --path string         config file path, do not use with -s
  -s, --search strings      config file search paths, do not use with -p (default [/etc/corelayer/lens,/nsconfig/ssl/LENS,$HOME/.lens,$PWD,%APPDATA%/corelayer/lens,%LOCALAPPDATA%/corelayer/lens,%PROGRAMDATA%/corelayer/lens])

Use "lens [command] --help" for more information about a command.
```

**NOTE: do not use the same file name for both the config file and environment variables file, e.g. ```config.yaml``` and ```config.env```, unless you specify the path of the config files using the ```-p``` flag**<br/>
Due to a bug in [spf13/viper](https://github.com/spf13/viper/issues/1163), the wrong file might get loaded in memory, causing an unexpected application exit.

By default, lens will be looking for a global configuration file in the following paths:
- /etc/corelayer/lens
- /nsconfig/ssl/LENS
- $HOME/.lens
- $PWD (the current working directory)
- %APPDATA%
- %LOCALAPPDATA%
- %PROGRAMDATA%

Global Flags:
- -c / --configFile: allows you to specify a custom global configuration file
- -e / --envFile: allows you to specify an environment variables file
- -p / --path: allows you to specify to path for the global configuration file
- -s / --search: allows you to specify multiple search paths

- -l / --loglevel [debug | info | warning | error]: allows you to specify a loglevel, the default is "info"

### Request mode
```
    __    _______   _______
   / /   / ____/ | / / ___/
  / /   / __/ /  |/ /\__ \
 / /___/ /___/ /|  /___/ /
/_____/_____/_/ |_//____/

Let's Encrypt for NetScaler ADC - Request Mode

Usage:
  lens request [flags]

Flags:
  -a, --all           request all
  -h, --help          help for request
  -n, --name string   request name

Global Flags:
  -c, --configFile string   config file name (default "config.yaml")
  -e, --envFile string      environment file name (default "variables.env")
  -l, --loglevel string     log level
  -p, --path string         config file path, do not use with -s
  -s, --search strings      config file search paths, do not use with -p (default [/etc/corelayer/lens,/nsconfig/ssl/LENS,$HOME/.lens,$PWD,%APPDATA%/corelayer/lens,%LOCALAPPDATA%/corelayer/lens,%PROGRAMDATA%/corelayer/lens])
```

You can either request one single certificate or all configured certificates at once.
For a single certificate, you need to specify the name of the configuration using the -n/--name flag.

Flags:
- -a / --all: make a request for all configured certificates
- -n / --name: specify the certificate to be requested

*Both flags are mutually exclusive!*

The global flags are still applicable and can be used accordingly.

### Environment variables

Environment variables can be set in two ways:
- Directly on the command-line
- Using an .env file

**NOTE: both can be used at the same time**

#### CLI

Always prefix the environment variable with ```LENS_```.
Other environment variables will not be used to replace variable placeholders in the config files.

Example:<br/>
```LENS_NAME=corelayer_acme lens request -a```

#### Environment Variables File

You do not need to prefix the environment variable in the file.
However, when referencing the variable in a config file, you **must** prefix

Example: ```variables.env```
```text
NAME=corelayer_acme
```

CLI: ```lens request -e variables.env```

#### Referencing environment variables

You can reference the environment variables in the global configuration file.</br>
If we take the preceding sections as an example, we have LENS_NAME or NAME as an environment variable.</br>
We can now use that variable as a reference using ${LENS_NAME} as the value of a parameter.

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
organizations:
  - name: <organization name>
    environments:
      - name: <environment name>
        type: <standalone | hapair | cluster>
        management:
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
acmeUsers:
  - name: <acme username>
    email: <acme e-mail address>
providerParameters:
  - name: <name for the set of parameters>
    variables:
      - name: <environment variable name>
        value: <environment variable value>
      - name: <environment variable name>
        value: <environment variable value>
```

#### Examples
- [Standalone - using SNIP](#standalone---using-snip)
- [Standalone - using NSIP](#standalone---using-nsip)
- [High-Availability pair - using SNIP](#high-availability-pair---using-snip)
- [High-Availability pair - using NSIP](#high-availability-pair---using-nsip)
- [Multiple environments](#multiple-environments)

##### Standalone - using SNIP
Global configuration:
```yaml
configPath: conf.d
organizations:
  - name: corelayer
    environments:
      - name: development
        type: standalone
        management:
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
acmeUsers:
  - name: corelayer_acme
    email: fake@email.com
```

##### Standalone - using NSIP
Global configuration:
```yaml
configPath: conf.d
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
acmeUsers:
  - name: corelayer_acme
    email: fake@email.com
```

##### High-availability Pair - using SNIP
Global configuration:
```yaml
configPath: conf.d
organizations:
  - name: corelayer
    environments:
      - name: development
        type: hapair
        management:
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
acmeUsers:
  - name: corelayer_acme
    email: fake@email.com
```

##### High-availability Pair - using NSIP
Global configuration:
```yaml
configPath: conf.d
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
acmeUsers:
  - name: corelayer_acme
    email: fake@email.com
```

##### Multiple environments
Global configuration:
```yaml
configPath: conf.d
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
      - name: test
        type: hapair
        management:
          name: vpx-tst
          address: vpx-tst.test.local
        nodes:
          - name: vpx-tst-001
            address: 192.168.2.11
          - name: vpx-tst-002
            address: 192.168.2.12
        credentials:
          username: nsroot
          password: nsroot
        connectionSettings:
          useSsl: true
          timeout: 3000
          validateServerCertificate: false
          logTlsSecrets: false
          autoLogin: false
acmeUsers:
  - name: corelayer_acme
    email: fake@email.com
```

##### Multiple environments - with environment variable file
Environment variables file:
```text
NAME1=corelayer_acme1
DEV_PASS=secretPassword
TST_PASS=anotherPassword
```

Global configuration:
```yaml
configPath: conf.d
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
          password: ${LENS_DEV_PASS}
        connectionSettings:
          useSsl: true
          timeout: 3000
          validateServerCertificate: false
          logTlsSecrets: false
          autoLogin: false
      - name: test
        type: hapair
        management:
          name: vpx-tst
          address: vpx-tst.test.local
        nodes:
          - name: vpx-tst-001
            address: 192.168.2.11
          - name: vpx-tst-002
            address: 192.168.2.12
        credentials:
          username: nsroot
          password: {LENS_TST_PASS}
        connectionSettings:
          useSsl: true
          timeout: 3000
          validateServerCertificate: false
          logTlsSecrets: false
          autoLogin: false
acmeUsers:
  - name: ${LENS_NAME1}
    email: fake@email.com
```

### Certificate configuration
```yaml
name: <name>
request:
  organization: <organization name>
  environment: <environment name>
  acmeUser: <acme username>
  challenge:
    service: LE_STAGING | LE_PRODUCTION | <custom url>
    type: <http-01 | dns-01>
    provider: <netscaler-http-global | netscaler-adns | <name of dns provider>
    providerParameters: <providerParameters name>
    disableDnsPropagationCheck: <true | false>
  keyType: <RSA20248 | RSA4096 | RSA8192 | EC256 | EC384>
  content:
    commonName: <common name>
    subjectAlternativeNames:
      - <subjectAlternativeName>
      - <subjectAlternativeName>
    subjectAlternativeNamesFile: <filename | filepath>
installation:
  - organization: <organization name>
    environment: <environment name>
    replaceDefaultCertificate: <true | false>
    sslVirtualServers:
      - name: <ssl vserver name>
        sniEnabled: <true | false>
    sslServices:
      - name: <ssl service name>
        sniEnabled: <true | false>
```
As you can see, the configuration is split up in two parts:
- request
- installation

#### Request
This section holds all the details to be able to request a certificate from your ACME service of choice.
We need to specify the organization and environment name to select which NetScaler to talk to.

##### Challenge
###### Service
You can either choose one of the pre-defined services, or specify your own ACME Service URL.
- ```LE_STAGING```: Let's Encrypt STAGING Environment
- ```LE_PRODUCTION```: Let's Encrypt PRODUCTION Environment

###### Type
We currently either support ```http-01``` or ```dns-01``` as the challenge type.

###### Provider
This tool is primarily meant for use with NetScaler ADC, both for the certificate request as for the installation of the certificate.
However, we do support external DNS providers.

- ```netscaler-http-global```
- ```netscaler-adns```

**Other DNS providers are to be enabled in a future releases.**

###### DisableDnsPropagationCheck
In case you are executing a challenge from within a network that has split-DNS (different DNS responses on the internet compared to the local network), you might need to set ```DisableDnsPropagationCheck``` to ```true```.</br>When enabled, lens will not wait for any propagation to happen, nor will it check if propagation has succeeded in orde for it to complete the challenge.


#### Installation
Once the certificate request is done, we can install the certificate onto multiple ssl vservers in multiple environments.
This is especially useful when having SAN-certificates or wildard certificates, so they can be bound appropriately on different NetScaler environments.

**Note that you cannot have the option ```replaceDefaultCertificate``` set to ```true``` while having endpoints defined under "sslVserver" and/or "sslServices"**

#### Examples
- [Simple certificate](#simple-certificate)
- [SAN certificate - using manual entries](#san-certificate---using-manual-entries)
- [SAN certificate - replace default NetScaler certificate](#san-certificate---replace-default-netscaler-certificate)
- [SAN certificate - using external file](#san-certificate---using-external-file)

##### Simple certificate
Certificate configuration:
```yaml
name: corelogic_dev
request:
  organization: corelayer
  environment: development
  acmeUser: corelayer_acme
  challenge:
    service: LE_STAGING
    type: http-01
    provider: netscaler-http-global
  keyType: RSA4096
  content:
    commonName: corelogic.dev.corelayer.eu
installation:
  - organization: corelayer
    environment: development
    sslVirtualServers:
      - name: CSV_DEV_SSL
        sniEnabled: true
```

##### SAN certificate - using manual entries
Certificate configuration:
```yaml
name: corelogic_dev
request:
  organization: corelayer
  environment: development
  acmeUser: corelayer_acme
  challenge:
    service: LE_STAGING
    type: http-01
    provider: netscaler-http-global
  keyType: RSA4096
  content:
    commonName: corelogic.dev.corelayer.eu
    subjectAlternativeNames:
      - demo.dev.corelayer.eu
      - my.dev.corelayer.eu
installation:
  - organization: corelayer
    environment: development
    sslVirtualServers:
      - name: CSV_DEV_SSL
        sniEnabled: true
```

##### SAN certificate - Replace default NetScaler certificate
Certificate configuration:
```yaml
name: corelogic_dev
request:
  organization: corelayer
  environment: development
  acmeUser: corelayer_acme
  challenge:
    service: LE_STAGING
    type: http-01
    provider: netscaler-http-global
  keyType: RSA4096
  content:
    commonName: vpx.dev.corelayer.eu
    subjectAlternativeNames:
      - vpx-001.dev.corelayer.eu
      - vpx-002.dev.corelayer.eu
installation:
  - organization: corelayer
    environment: development
    replaceDefaultCertificate: true
```

##### SAN certificate - using external file
Certificate configuration:
```yaml
name: corelogic_dev
request:
  organization: corelayer
  environment: development
  acmeUser: corelayer_acme
  challenge:
    service: LE_STAGING
    type: http-01
    provider: netscaler-http-global
  keyType: RSA4096
  content:
    commonName: corelogic.dev.corelayer.eu
    subjectAlternativeNamesFile: corelogic_dev_san.txt
installation:
  - organization: corelayer
    environment: development
    sslVirtualServers:
      - name: CSV_DEV_SSL
        sniEnabled: true
```

Subject Alternative Names File (stored next to the certificate configuration file):
```text
demo.dev.corelayer.eu
my.dev.corelayer.eu
```

##### Simple certificate - multiple installations
Certificate configuration:
```yaml
name: corelogic_dev
request:
  organization: corelayer
  environment: development
  acmeUser: corelayer_acme
  challenge:
    service: LE_STAGING
    type: http-01
    provider: netscaler-http-global
  keyType: RSA4096
  content:
    commonName: corelogic.dev.corelayer.eu
installation:
  - organization: corelayer
    environment: development
    sslVirtualServers:
      - name: CSV_DEV_SSL
        sniEnabled: true
      - name: CSV_PUBLICDEV_SSL
        sniEnabled: false
  - organization: corelayer
    environment: test
    sslVirtualServers:
      - name: CSV_TST_SSL
        sniEnabled: true
```