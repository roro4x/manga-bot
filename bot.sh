#!/bin/bash
ps -aux | grep rand-grab-manga-bot/rand-grab-manga-bot | grep -v grep | awk '{ print $2 }' | xargs kill
ps -aux | grep phantomjs/phantomjs | grep -v grep | awk '{ print $2 }' | xargs kill
/home/rand-grab-manga-bot/rand-grab-manga-bot 2>&1 | tee /home/rand-grab-manga-bot/logs.txt &
