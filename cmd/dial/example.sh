#!/bin/sh

printf "GET /programming/cshell.hrm HTTP/1.1\r\n\
Host: www.textfiles.com\r\n\
Accept: text/html\r\n\
Accept-Language: en-US,en;q=0.9\r\n\
DNT: 0\r\n\
Referer: http://www.textfiles.com/programming/\r\n\
Upgrade-Insecure-Requests: 0\r\n\
User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36\r\n\r\n" | dial textfiles.com:80
