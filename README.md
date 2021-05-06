# manga-bot
### This is Telegram Bot Mangakaka
*It's can grab random manga chapter from Readmanga.live and than send to telegram channel like instant view telegra.ph page.*

### What do you need for run (Ubuntu example):
- ```$ sudo apt update```
- install git ```$ sudo apt install git```
- install golang ```sudo apt install golang```
- set golang project path `$ sudo nano ~/.profile` add ```export GOPATH=$HOME/work``` to the end of the file: 
- need to download <https://phantomjs.org/download.html>, put in *phantomgo* dir and install libs:
  - ```$ sudo apt-get install libfreetype6 libfreetype6-dev```
  - ```$ sudo apt-get install libfontconfig1 libfontconfig1-dev```
> - Look at `phantomjsPath` variable in `phantom.go`
>   - Make sure that your Phantomgo PATH is `GOPATH + /src/github.com/k4s/phantomgo/phantomjs/phantomjs`
> - About ```bot.sh```
>   - This is needed to remove the application from processes because `phantomjs` does not close correctly
- build app
- edit path in ```bot.sh``` and add to crontab
- set GOPATH env to crontab

### TODO:
- [X] Checker *if manga haven't chapters*
- [X] Checker *if need to buy manga than select other manga*
- [X] Will add manga description in message (added URL to manga page)
- [] Make all xpath strings like vars
