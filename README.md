# moblog

Mobile Blogging System

## Usege

```
Usage of ./moblog:
  -a string
    	accept e-mail from (default "*")
  -d string
    	repository of jekyll (default "/path/to/jekyll/blog")
  -m string
    	mail server (default "localhost:25")
  -s string
    	e-mail sender (default "moblog@example.com")
  -t	use html
```

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
#VERBOSE=ON

:0
* ^X-Original-To: moblog@example.com
| /path/to/the/moblog/moblog -a blog@example.com -s me@example.com -d /path/to/jekyll/blog
```

*WARN* You have to use private e-mail address for `-a` flag. 

## License

MIT

## Author

Yasuhiro Matsumoto
