# Go Web Restart
Automatically restart go web application when you change your source code. Allow to save you sometime from restarting that web process manually.
Inspired from Beego's Bee command that allow for automatically restart go web service, i decided to give my own solution on this problem.
**Bear in mind that i built it in a rush and for my own use**, feel free to fork and improve it :]. 
The original name is actually Gin Web Restart since i'm trying to use it for Gin Web Framework.
### Usage
1. Download and install it
```sh
go get -u github.com/firmanmm/go-webrestart
```
2. Navigate to your current go source code
3. Run go-webrestart
```sh
go-webrestart
```
4. Try to change your main.go file and see it rebuild and restart
### Parameter

| Parameter | Description | Example  |
| ------------- |:-------------:| -----:|
| -v      | show more information to the console | go-webrestart -v |
| -p      | pass variable to go build method      |   go-webrestart -p "-tags=jsoniter" |
| -e      | add extra extension to watch for (default .go)      | go-webrestart -e .tpl .html .env |

### Test
+ Tested on Beego Web Framework and working under Windows 10 using Ubuntu 16.04 (WSL).
+ Tested on Gin Web Framework and working under Windows 10 using Ubuntu 16.04 (WSL).

