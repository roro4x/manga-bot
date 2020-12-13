# rand-grab-manga-bot
### This is Telegram Bot Mangakaka v0.1.1.0000
*It's can grab random manga chapter from Readmanga.live and than send to telegram channel like instant view telegra.ph page.*

### What do you need for run (Ubuntu example):
- ```$ sudo apt update```
- install git ```$ sudo apt install git```
- install golang:
  - ```$ wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz```
  - ```$ sudo tar -C /usr/local -xzf go1.13.linux-amd64.tar.gz```
  - `$ sudo nano ~/.profile` add the following to the end of the file: 
    - ```export GOPATH=$HOME/work```
    - ```export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin```
- need to download <https://phantomjs.org/download.html>, put in *phantomgo* dir and install libs:
  - ```$ sudo apt-get install libfreetype6 libfreetype6-dev```
  - ```$ sudo apt-get install libfontconfig1 libfontconfig1-dev```
- build app
- ```$ ./prog &```
- ```$ kill *pid*```
> - Look at `phantomjsPath` variable in `phantom.go`
>   - Make sure that your Phantomgo PATH is `/src/github.com/k4s/phantomgo/phantomjs/phantomjs`


### TODO:
- [X] Checker *if manga haven't chapters*
- [X] Checker *if need to buy manga than select other manga*
- [ ] Will add exception processing (mb need to use https://github.com/sethgrid/pester)
- [ ] Will add manga description in message
- [ ] Dockerfile
