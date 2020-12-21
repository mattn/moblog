# moblog

Mobile Blogging System

## Installation

Setup your postfix mail server to use procmail.

```
mailbox_command = /usr/bin/procmail
```

Configure your .procmailrc to handle e-mail to the command.

```
PATH=/bin:/usr/bin
MAILDIR=$HOME/Maildir
DEFAULT=$MAILDIR/
SPAM=$MAILDIR/.spam/
LOCKFILE=$HOME/.lockmail
LOGFILE=$HOME/.procmail.log
#VERBOSE=ON # 詳細ログ出力

:0
* ^X-Original-To: moblog@example.com
| /path/to/the/moblog/moblog -a blog@example.com -s me@example.com -d /path/to/jekyll/blog
```

*WARN* You have to use private e-mail address for `-a` flag. 

## License

MIT

## Author

Yasuhiro Matsumoto
