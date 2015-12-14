# Transhift

A tiny and simple peer-to-peer file mover.

### Usage

##### Prerequisites

* Both you and your friend (or other machine) must have the binary installed
* You and your friend should agree on a password via other messaging means (SMS, IRC, IRL, etc.); this doesn't have to be super secure, but you definitely shouldn't use one that you've used for other online accounts
* If you are *sending* a file to your friend, ask them for their external IP address (tell them to see [whatsmyip.org](http://www.whatsmyip.org/)); if you are *receiving* a file from your friend, be sure to tell them *your* external IP address
* UPnP should be enabled on NAT protected machines that are to receive files; this is usually enabled by default

##### Sending a file to a friend

1. Tell your friend to run `transhift download <password>` where `<password>` is the password you and your friend agreed on.
2. On your own machine, run `transhift upload <peer> <password> <file>` where `<peer>` is the external IP address of your friend's machine, `<password>` is the password you and your friend agreed on, and `<file>` is the relative or absolute path of the file you would like to send them

##### Receiving a file from your friend

1. On your own machine, run `transhift download <password>` where `<password>` is the password you and your friend agreed on.
2. Tell your friend to run `transhift upload <peer> <password> <file>` where `<peer>` is the external IP address of your machine, `<password>` is the password you and your friend agreed on, and `<file>` is the relative or absolute path of the file they would like to send you

### Quick Start

To download and install, run:

```bash
curl -s https://raw.githubusercontent.com/transhift/transhift/master/install.sh | bash
```

Or run each command individually from the [install script](https://github.com/transhift/transhift/blob/master/install.sh).

To test it out, run `transhift --version` and you should see `Transhift version 0.1.0`.
