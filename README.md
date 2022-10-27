# Silent Deny Discord Bot

Open source code for a Discord bot that can silently delete messages that match a deny list of regular expressions.
Helpful to relieve crypto server spam! It does not notify the user of message deletion.

A bot invite to any public server can be found [here](https://discord.com/api/oauth2/authorize?client_id=1034999125093142669&permissions=8&scope=bot). The bot needs manage messages permissions.

## Install

Go 1.18 or Docker

**With Go**

```
git clone http://github.com/rauljordan/silent-deny && cd silent-deny
go build .
./silent-deny -token=<BOT_TOKEN> -denylist=/path/to/denylist.txt
```

**With Docker**

```
docker run rauljordan/denylist -token=<BOT_TOKEN> -denylist=/path/to/denylist.txt
```

## Usage

```
Usage of ./silent-deny:
  -denylist string
        Filepath to denylist of regular expressions, separated by new line delimiters
  -token string
        Discord bot token
```

A denylist file is a simple text file of line-separated regular expressions to match messages. Example:

```
uni-airdrop\.org
.*airdrop
uni-airdrop
```
