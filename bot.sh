#!/bin/bash
ps -aux | grep /home/work/manga-bot/manga-bot | grep -v grep | awk '{ print $2 }' | xargs kill
ps -aux | grep phantomjs/phantomjs | grep -v grep | awk '{ print $2 }' | xargs kill
/home/work/manga-bot/manga-bot 2>&1 | tee /home/work/manga-bot/logs.txt &
