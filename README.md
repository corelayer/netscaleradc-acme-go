# Let's Encrypt for NetScaler ADC
## Table of Contents
[Introduction](#introduction)</br>
[Changelog](#changelog)</br>
&nbsp;&nbsp;&nbsp;&nbsp;[v0.3.1](#v031)</br>
[System Requirements](#system-requirements)</br>
&nbsp;&nbsp;&nbsp;&nbsp;[Operating System](#operating-system)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Native NetScaler ADC binary](#native-netscaler-adc-binary)</br>
&nbsp;&nbsp;&nbsp;&nbsp;[Certificate Authorities](#certificate-authorities)</br>
&nbsp;&nbsp;&nbsp;&nbsp;[NetScaler ADC](#netscaler-adc)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[User Permission](#user-permissions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[CLI Commands](#cli-commands)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[NetScaler ADM](#netscaler-adm)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Terraform](#terraform)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Running on NetScaler natively](#running-on-netscaler-adc-natively)</br>
&nbsp;&nbsp;&nbsp;&nbsp;[Running Lens](#running-lens)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Request mode](#request-mode)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Environment variables](#environment-variables)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Defining environment variables](#defining-environment-variables)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[CLI](#cli)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Environment variables file](#environment-variables-file)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Referencing environment variables](#referencing-environment-variables)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Integrations](#integrations)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[1Password](#1password)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Configuration mode](#configuration-mode)</br>
&nbsp;&nbsp;&nbsp;&nbsp;[Configuration](#configuration)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Global configuration](#global-configuration)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Config path](#config-path)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Organizations](#organizations)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Users](#users)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Provider parameters](#provider-parameters)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Examples](#examples)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Certificate configuration](#certificate-configuration)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Request](#request)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Challenge](#challenge)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Service](#service)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Type](#type)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Provider](#provider)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Provider parameters](#provider-parameters-1)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Disable DNS propagation check](#disablednspropagationcheck)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Installation](#installation)</br>

---
## Introduction

Let's Encrypt for NetScaler ADC (aka LENS) is a tool which allows you to generate certificates based on the well-known ACME protocol. It is based on the fantastic library from the people at [https://github.com/go-acme/lego](https://github.com/go-acme/lego) to provide the functionality to talk to different DNS providers, but now also NetScaler ADC.

[Back to top](#lets-encrypt-for-netscaler-adc)

---
## Changelog
### v0.3.1
- Changed global application flags to accommodate a global configuration file and environment variables file flag
  - changed -f / --config to -c / --configFile
  - added -e / --envFile
  - This also frees up the -f parameter to be changed later to a --force parameter

- Added provider parameters to global configuration file for use with DNS providers which require environment variables to be set when being used.
  - For more information on available providers, see [https://go-acme.github.io/lego/dns/](https://go-acme.github.io/lego/dns/)

[Back to top](#lets-encrypt-for-netscaler-adc)

---
## System requirements
### Operating system
We provide binaries for different operating systems and architectures:
- Linux (amd64/arm64)
- MacOS (Intel/Apple Silicon)
- Windows (amd64/arm64)
- FreeBSD (amd64/arm64), versions > FreeBSD 11

[Back to top](#lets-encrypt-for-netscaler-adc)

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

[Back to top](#lets-encrypt-for-netscaler-adc)

### Certificate Authorities

By default, we support both staging and production environments for Let's Encrypt.
Lens is designed to work with other certificate authorities who provide access through the ACME protocol.

*If you are a user of other ACME-protocol based services, such as Sectigo, please reach out, so we can ensure maximum compatibility!*

[Back to top](#lets-encrypt-for-netscaler-adc)

### NetScaler ADC
#### User permissions
- You will need a user account with the following permissions:
    - High-Availability Node: show
    - DNS TXT Record: show, add, remove
    - Responder Action: add, remove
    - Responder Policy: add, remove
    - Responder Policy Global Binding: show, bind, unbind
    - System File: show, add, remove
    - SSL CertKey: show, add, update, remove
    - SSL Virtual Server / Service: bind
    - Save config

##### High-availability Node Permissions
```regexp
(^show\s+ha\s+node\s+0)
```
##### DNS TXT Record Permissions
```regexp
(^show\s+dns\s+txtRec\s+_acme-challenge\..*)
(^add\s+dns\s+txtRec\s+_acme-challenge\..*\.\s+[a-zA-Z0-9._-]+\s+-TTL\s+\d+)
(^rm\s+dns\s+txtRec\s+_acme-challenge\..*\.\s+-recordId\s+\d+)
```
##### Responder Permissions
```regexp
(^add\s+responder\s+action\s+RSA_LENS_.*\d{14}\s+respondwith\s+q{"HTTP/1.1\s+200\s+OK\\r\\n\\r\\n[a-zA-Z0-9._-]+"})
(^rm\s+responder\s+action\s+RSA_LENS_.*\d{14})

(^add\s+responder\s+policy\s+RSP_LENS_.*_\d{14}\s+"HTTP.REQ.HOSTNAME.EQ\(\\".*\\"\)\s+&&\s+HTTP.REQ.URL.EQ\(\\"/.well-known/acme-challenge/[a-zA-Z0-9_-]*\\"\)" RSA_LENS_.*_\d{14})
(^rm\s+responder\s+policy\s+RSP_LENS_.*\d{14})

(^show\s+responder\s+global)|(^show\s+responder\s+global\s+-type\s+REQ_OVERRIDE)
(^bind\s+responder\s+global\s+RSP_LENS_.*\d{14}.*\s+\d+\s+END\s+-type\s+REQ_OVERRIDE)
(^unbind\s+responder\s+global\s+RSP_LENS_.*\d{14}\s+-type\s+REQ_OVERRIDE)
```
##### System File Permissions
```regexp
(^show\s+system\s+file\s+.*_\d{14}\.(cer|key)\s+-fileLocation\s+"/nsconfig/ssl/LENS/.*")
(^add\s+system\s+file\s+.*_\d{14}\.(cer|key)\s+-fileLocation\s+"/nsconfig/ssl/LENS/")
(^rm\s+system\s+file\s+.*_\d{14}\.(cer|key)\s+-fileLocation\s+"/nsconfig/ssl/LENS/.*")
```
##### SSL CertKey Permissions
```regexp
(^show\s+ssl\s+certKey\s+LENS_.*)|(^show\s+ssl\s+certKey)
(^add\s+ssl\s+certKey\s+LENS_.*\s+-cert\s+"/nsconfig/ssl/LENS/.*_\d{14}.cer"\s+-key\s+"/nsconfig/ssl/LENS/.*_\d{14}.key".*)
(^update\s+ssl\s+certKey\s+LENS_.*\s+-cert\s+"/nsconfig/ssl/LENS/.*_\d{14}.cer"\s+-key\s+"/nsconfig/ssl/LENS/.*_\d{14}.key".*)
(^rm\s+ssl\s+certKey\s+LENS_.*)
```
##### SSL Virtual Server / Service Permissions
```regexp
(^bind\s+ssl\s+vserver\s+.*\s+-priority\s+\d+\s+-certkeyName\s+LENS_.*)
(^bind\s+ssl\s+service\s+.*\s+-priority\s+\d+\s+-certkeyName\s+LENS_.*)
```
##### Save configuration Permissions
```regexp
(^save ns config)
```

For easy configuration, we provide the necessary commands to create the command policy on NetScaler ADC in the section [below](#cli-commands).

[Back to top](#lets-encrypt-for-netscaler-adc)

##### CLI Commands
###### Command Policies
You can copy-paste the entire block below into the command-line of a NetScaler ADC instance to add the required command policies to the system.
```text
# High-Availability Node Permissions
add system cmdPolicy CMD_LENS_HA_SHOW ALLOW "(^show\\s+ha\\s+node\\s+0)"

# DNS TXT Record Permissions
add system cmdPolicy CMD_LENS_DNS_TXT_SHOW ALLOW "(^show\\s+dns\\s+txtRec\\s+_acme-challenge\\..*)"
add system cmdPolicy CMD_LENS_DNS_TXT_ADD ALLOW "(^add\\s+dns\\s+txtRec\\s+_acme-challenge\\..*\\.\\s+[a-zA-Z0-9._-]+\\s+-TTL\\s+\\d+)"
add system cmdPolicy CMD_LENS_DNS_TXT_RM ALLOW "(^rm\\s+dns\\s+txtRec\\s+_acme-challenge\\..*\\.\\s+-recordId\\s+\\d+)"

# Responder Permissions
add system cmdPolicy CMD_LENS_RSA_ADD ALLOW q<(^add\s+responder\s+action\s+RSA_LENS_.*\d{14}\s+respondwith\s+q{"HTTP/1.1\s+200\s+OK\\r\\n\\r\\n[a-zA-Z0-9._-]+"})>
add system cmdPolicy CMD_LENS_RSA_RM ALLOW "(^rm\\s+responder\\s+action\\s+RSA_LENS_.*\\d{14})"
add system cmdPolicy CMD_LENS_RSP_ADD ALLOW q<(^add\s+responder\s+policy\s+RSP_LENS_.*_\d{14}\s+"HTTP.REQ.HOSTNAME.EQ\(\\".*\\"\)\s+&&\s+HTTP.REQ.URL.EQ\(\\"/.well-known/acme-challenge/[a-zA-Z0-9_-]*\\"\)" RSA_LENS_.*_\d{14})>
add system cmdPolicy CMD_LENS_RSP_RM ALLOW "(^rm\\s+responder\\s+policy\\s+RSP_LENS_.*\\d{14})"
add system cmdPolicy CMD_LENS_RSPL_GLOBAL_SHOW ALLOW "(^show\\s+responder\\s+global)|(^show\\s+responder\\s+global\\s+-type\\s+REQ_OVERRIDE)"
add system cmdPolicy CMD_LENS_RSPL_GLOBAL_BIND ALLOW "(^bind\\s+responder\\s+global\\s+RSP_LENS_.*\\d{14}.*\\s+\\d+\\s+END\\s+-type\\s+REQ_OVERRIDE)"
add system cmdPolicy CMD_LENS_RSPL_GLOBAL_UNBIND ALLOW "(^unbind\\s+responder\\s+global\\s+RSP_LENS_.*\\d{14}\\s+-type\\s+REQ_OVERRIDE)"

# System File Permissions
add system cmdPolicy CMD_LENS_SYSTEMFILE_SHOW ALLOW "(^show\\s+system\\s+file\\s+.*_\\d{14}\\.(cer|key)\\s+-fileLocation\\s+\"/nsconfig/ssl/LENS/.*\")"
add system cmdPolicy CMD_LENS_SYSTEMFILE_ADD ALLOW "(^add\\s+system\\s+file\\s+.*_\\d{14}\\.(cer|key)\\s+-fileLocation\\s+\"/nsconfig/ssl/LENS/\")"
add system cmdPolicy CMD_LENS_SYSTEMFILE_RM ALLOW "(^rm\\s+system\\s+file\\s+.*_\\d{14}\\.(cer|key)\\s+-fileLocation\\s+\"/nsconfig/ssl/LENS/.*\")"

# SSL CertKey Permissions
add system cmdPolicy CMD_LENS_SSL_CERTKEY_SHOW ALLOW "(^show\\s+ssl\\s+certKey\\s+LENS_.*)|(^show\\s+ssl\\s+certKey)"
add system cmdPolicy CMD_LENS_SSL_CERTKEY_ADD ALLOW "(^add\\s+ssl\\s+certKey\\s+LENS_.*\\s+-cert\\s+\"/nsconfig/ssl/LENS/.*_\\d{14}.cer\"\\s+-key\\s+\"/nsconfig/ssl/LENS/.*_\\d{14}.key\".*)"
add system cmdPolicy CMD_LENS_SSL_CERTKEY_UPDATE ALLOW "(^update\\s+ssl\\s+certKey\\s+LENS_.*\\s+-cert\\s+\"/nsconfig/ssl/LENS/.*_\\d{14}.cer\"\\s+-key\\s+\"/nsconfig/ssl/LENS/.*_\\d{14}.key\".*)"
add system cmdPolicy CMD_LENS_SSL_CERTKEY_RM ALLOW "(^rm\\s+ssl\\s+certKey\\s+LENS_.*)"

# SSL Virtual Server / Service Permissions
add system cmdPolicy CMD_LENS_SSL_VSERVER_BIND ALLOW "(^bind\\s+ssl\\s+vserver\\s+.*\\s+-priority\\s+\\d+\\s+-certkeyName\\s+LENS_.*)"
add system cmdPolicy CMD_LENS_SSL_SERVICE_BIND ALLOW "(^bind\\s+ssl\\s+service\\s+.*\\s+-priority\\s+\\d+\\s+-certkeyName\\s+LENS_.*)"

# Save Config Permissions
add system cmdPolicy CMD_LENS_SAVE_CONFIG ALLOW "(^save ns config)"
```

###### Create user Account
Create a user account on NetScaler ADC for use with ```lens```.</br>

**Notes**:
- if you your environment has autoLogin set to ```false``` in the connection settings, a timeout of 60 seconds should be more than enough.
- Disable external authentication for the user
- Only allow management access to NetScaler ADC using Nitro API

```text
add system user <username> <password> -externalAuth DISABLED -timeout 60 -maxsession 20 -allowedManagementInterface API
```

###### Assign permissions
Bind the required command policies to any user that needs to operate on NetScaler ADC:
```text
# High-Availability Node Permissions
bind system user <username> CMD_LENS_HA_SHOW 1001

# DNS TXT Record Permissions
bind system user <username> CMD_LENS_DNS_TXT_SHOW 1101
bind system user <username> CMD_LENS_DNS_TXT_ADD 1102
bind system user <username> CMD_LENS_DNS_TXT_RM 1103

# Responder Permissions
bind system user <username> CMD_LENS_RSA_ADD 1201
bind system user <username> CMD_LENS_RSA_RM 1202
bind system user <username> CMD_LENS_RSP_ADD 1211
bind system user <username> CMD_LENS_RSP_RM 1212
bind system user <username> CMD_LENS_RSPL_GLOBAL_SHOW 1221
bind system user <username> CMD_LENS_RSPL_GLOBAL_BIND 1222
bind system user <username> CMD_LENS_RSPL_GLOBAL_UNBIND 1223

# System File Permissions
bind system user <username> CMD_LENS_SYSTEMFILE_SHOW 1301
bind system user <username> CMD_LENS_SYSTEMFILE_ADD 1302
bind system user <username> CMD_LENS_SYSTEMFILE_RM 1303

# SSL CertKey Permissions
bind system user <username> CMD_LENS_SSL_CERTKEY_SHOW 1401
bind system user <username> CMD_LENS_SSL_CERTKEY_ADD 1402
bind system user <username> CMD_LENS_SSL_CERTKEY_UPDATE 1403
bind system user <username> CMD_LENS_SSL_CERTKEY_RM 1404

# SSL Virtual Server / Service Permissions
bind system user <username> CMD_LENS_SSL_VSERVER_BIND 1501
bind system user <username> CMD_LENS_SSL_SERVICE_BIND 1502

# Save Config Permissions
bind system user <username> CMD_LENS_SAVE_CONFIG 1999
```

[Back to top](#lets-encrypt-for-netscaler-adc)

##### NetScaler ADM
###### Configuration Job

A configuration job template for ADM is available in the assets folder.

[Back to top](#lets-encrypt-for-netscaler-adc)

##### Terraform
```text
</TBD>
```

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Running on NetScaler ADC natively
If you run the binary natively on NetScaler ADC:
- You will need internet access to connect to your ACME service of choice if you want to run natively on NetScaler ADC
- You wil need connectivity with either the NSIP or SNIP address for the environments to which you will connect.

[Back to top](#lets-encrypt-for-netscaler-adc)

---
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

[Back to top](#lets-encrypt-for-netscaler-adc)

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

[Back to top](#lets-encrypt-for-netscaler-adc)

### Environment variables

Environment variables can be set in two ways:
- Directly on the command-line
- Using an .env file

**NOTE: both can be used at the same time**

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Defining environment variables
##### CLI

Always prefix the environment variable with ```LENS_```.
Other environment variables will not be used to replace variable placeholders in the config files.

Example:<br/>
```LENS_NAME=corelayer_acme lens request -a```

[Back to top](#lets-encrypt-for-netscaler-adc)

##### Environment Variables File

You do not need to prefix the environment variable in the file.
However, when referencing the variable in a config file, you **must** prefix

Example: ```variables.env```
```text
NAME=corelayer_acme
```

CLI: ```lens request -e variables.env```

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Referencing environment variables

You can reference the environment variables in the global configuration file.</br>
If we take the preceding sections as an example, we have LENS_NAME or NAME as an environment variable.</br>
We can now use that variable as a reference using ${LENS_NAME} as the value of a parameter.

See [Multiple environments - with environment variable file](#multiple-environments---with-environment-variable-file) for more information.

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Integrations
##### 1Password
Using 1password-CLI (```op```), you can integrate your password vault for use with environment variables.</br>
For more information, check out [1Password Command-Line tool](https://1password.com/downloads/command-line/) and the [documentation](https://developer.1password.com/docs/cli/get-started/).

When you have set up 1password-CLI, you can start referencing items in your vault:
- [Load secrets into the environment](https://developer.1password.com/docs/cli/secrets-environment-variables)
- [Load secrets into config files](https://developer.1password.com/docs/cli/secrets-config-files)

[Back to top](#lets-encrypt-for-netscaler-adc)

### Configuration mode

**Not implemented**

The goal is to be able to configure lens from the command line.

[Back to top](#lets-encrypt-for-netscaler-adc)

---
## Configuration

Configuration for lens is done using YAML files and is split up in 2 parts:
- global configuration
- certificate configuration

As the global configuration needs to have all account details for the different environments to which you need to connect, this file is separated and can be stored in a secured location with appropriate permissions.
- On Linux, this would typically be /etc/corelayer/lens for example, which can have the necessary permissions to only allow root access or access from the user account intended to run lens.

The individual certificate configuration files can be stored elsewhere on the system with more permissive access, as it will only contain the certificate configuration. References to the environments are made using the name of the configured environment.

[Back to top](#lets-encrypt-for-netscaler-adc)

### Global configuration
```yaml
configPath: <path to the individual certificate configuration files>
organizations:
  - name: <organization name>
    environments:
      - name: <environment name>
        type: <standalone | hapair | cluster>
        management:
          name: <name for the management address>
          address: <management ip address / fqdn>
        nodes:
          - name: <name for the individual node>
            address: <NSIP address / fqdn>
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
  - name: <user name for reference in certificate configuration files>
    email: <user e-mail address>
    eab:
      kid: <kid for external account binding>
      hmacEncoded: <hmac key for external account binding>
  - name: <user name for reference in certificate configuration files>
    email: <user e-mail address>
    eab:
      kid: <kid for external account binding>
      hmacEncoded: <hmac key for external account binding>
providerParameters:
  - name: <name for the set of parameters>
    variables:
      - name: <environment variable name>
        value: <environment variable value>
      - name: <environment variable name>
        value: <environment variable value>
```

As you can see, the global configuration has several sections, which we will discuss in more detail below:
- [config path](#config-path)
- [organizations](#organizations)
- [users](#users)
- [provider parameters](#provider-parameters)

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Config path

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Organizations

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Users

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Provider parameters

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Examples
- [Standalone - using SNIP](#standalone---using-snip)
- [Standalone - using NSIP](#standalone---using-nsip)
- [High-Availability pair - using SNIP](#high-availability-pair---using-snip)
- [High-Availability pair - using NSIP](#high-availability-pair---using-nsip)
- [Multiple environments](#multiple-environments)
- [Multiple environments - with environment variable file](#multiple-environments---with-environment-variable-file)

[Back to top](#lets-encrypt-for-netscaler-adc)

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
users:
  - name: corelayer_acme
    email: fake@email.com
```

[Back to top](#lets-encrypt-for-netscaler-adc)

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
users:
  - name: corelayer_acme
    email: fake@email.com
```

[Back to top](#lets-encrypt-for-netscaler-adc)

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
users:
  - name: corelayer_acme
    email: fake@email.com
```

[Back to top](#lets-encrypt-for-netscaler-adc)

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
users:
  - name: corelayer_acme
    email: fake@email.com
```

[Back to top](#lets-encrypt-for-netscaler-adc)

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
users:
  - name: corelayer_acme
    email: fake@email.com
```

[Back to top](#lets-encrypt-for-netscaler-adc)

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
users:
  - name: ${LENS_NAME1}
    email: fake@email.com
```

[Back to top](#lets-encrypt-for-netscaler-adc)

### Certificate configuration
```yaml
name: <name>
request:
  target:
    organization: <organization name>
    environment: <environment name>
  user: <user name referenced from users section in global config file>
  challenge:
    service: LE_STAGING | LE_PRODUCTION | <custom url>
    type: <http-01 | dns-01>
    provider: <netscaler-http-global | netscaler-adns | <name of dns provider>
    providerParameters: <providerParameters name from global config file>
    disableDnsPropagationCheck: <true | false>
  keyType: <RSA20248 | RSA4096 | RSA8192 | EC256 | EC384>
  content:
    commonName: <common name>
    subjectAlternativeNames:
      - <subjectAlternativeName>
      - <subjectAlternativeName>
    subjectAlternativeNamesFile: <filename | filepath>
installation:
  - target:
      organization: <organization name>
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

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Request
This section holds all the details to be able to request a certificate from your ACME service of choice.
We need to specify the organization and environment name to select which NetScaler to talk to.

[Back to top](#lets-encrypt-for-netscaler-adc)

##### Challenge
###### Service
You can either choose one of the pre-defined services, or specify your own ACME Service URL.
- ```LE_STAGING```: Let's Encrypt STAGING Environment
- ```LE_PRODUCTION```: Let's Encrypt PRODUCTION Environment

[Back to top](#lets-encrypt-for-netscaler-adc)

###### Type
We currently either support ```http-01``` or ```dns-01``` as the challenge type.

[Back to top](#lets-encrypt-for-netscaler-adc)

###### Provider
This tool is primarily meant for use with NetScaler ADC, both for the certificate request as for the installation of the certificate.
However, we do support external DNS providers.

- ```netscaler-http-global```
- ```netscaler-adns```

**Other DNS providers are to be enabled in a future releases.**

[Back to top](#lets-encrypt-for-netscaler-adc)

###### Provider parameters
This tool is primarily meant for use with NetScaler ADC, both for the certificate request as for the installation of the certificate.
However, we do support external DNS providers.

- ```netscaler-http-global```
- ```netscaler-adns```

**Other DNS providers are to be enabled in a future releases.**

[Back to top](#lets-encrypt-for-netscaler-adc)

###### DisableDnsPropagationCheck
In case you are executing a challenge from within a network that has split-DNS (different DNS responses on the internet compared to the local network), you might need to set ```DisableDnsPropagationCheck``` to ```true```.</br>When enabled, lens will not wait for any propagation to happen, nor will it check if propagation has succeeded in order for it to complete the challenge.

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Installation
Once the certificate request is done, we can install the certificate onto multiple ssl vservers in multiple environments.
This is especially useful when having SAN-certificates or wildcard certificates, so they can be bound appropriately on different NetScaler environments.

**Note that you cannot have the option ```replaceDefaultCertificate``` set to ```true``` while having endpoints defined under "sslVserver" and/or "sslServices"**

[Back to top](#lets-encrypt-for-netscaler-adc)

#### Examples
- [Simple certificate](#simple-certificate)
- [SAN certificate - using manual entries](#san-certificate---using-manual-entries)
- [SAN certificate - replace default NetScaler certificate](#san-certificate---replace-default-netscaler-certificate)
- [SAN certificate - using external file](#san-certificate---using-external-file)

[Back to top](#lets-encrypt-for-netscaler-adc)

##### Simple certificate
Certificate configuration:
```yaml
name: corelogic_dev
request:
  target:
    organization: corelayer
    environment: development
  user: corelayer_acme
  challenge:
    service: LE_STAGING
    type: http-01
    provider: netscaler-http-global
  keyType: RSA4096
  content:
    commonName: corelogic.dev.corelayer.eu
installation:
  - target:
      organization: corelayer
      environment: development
    sslVirtualServers:
      - name: CSV_DEV_SSL
        sniEnabled: true
```

[Back to top](#lets-encrypt-for-netscaler-adc)

##### SAN certificate - using manual entries
Certificate configuration:
```yaml
name: corelogic_dev
request:
  target:
    organization: corelayer
    environment: development
  user: corelayer_acme
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
  - target:
      organization: corelayer
      environment: development
    sslVirtualServers:
      - name: CSV_DEV_SSL
        sniEnabled: true
```

[Back to top](#lets-encrypt-for-netscaler-adc)

##### SAN certificate - Replace default NetScaler certificate
Certificate configuration:
```yaml
name: corelogic_dev
request:
  target:
    organization: corelayer
    environment: development
  user: corelayer_acme
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
  - target:
      organization: corelayer
      environment: development
    replaceDefaultCertificate: true
```

[Back to top](#lets-encrypt-for-netscaler-adc)

##### SAN certificate - using external file
Certificate configuration:
```yaml
name: corelogic_dev
request:
  target:
    organization: corelayer
    environment: development
  user: corelayer_acme
  challenge:
    service: LE_STAGING
    type: http-01
    provider: netscaler-http-global
  keyType: RSA4096
  content:
    commonName: corelogic.dev.corelayer.eu
    subjectAlternativeNamesFile: corelogic_dev_san.txt
installation:
  - target:
      organization: corelayer
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

[Back to top](#lets-encrypt-for-netscaler-adc)

##### Simple certificate - multiple installations
Certificate configuration:
```yaml
name: corelogic_dev
request:
  target:
    organization: corelayer
    environment: development
  user: corelayer_acme
  challenge:
    service: LE_STAGING
    type: http-01
    provider: netscaler-http-global
  keyType: RSA4096
  content:
    commonName: corelogic.dev.corelayer.eu
installation:
  - target:
      organization: corelayer
      environment: development
    sslVirtualServers:
      - name: CSV_DEV_SSL
        sniEnabled: true
      - name: CSV_PUBLICDEV_SSL
        sniEnabled: false
  - target:
      organization: corelayer
      environment: test
    sslVirtualServers:
      - name: CSV_TST_SSL
        sniEnabled: true
```

[Back to top](#lets-encrypt-for-netscaler-adc)

---