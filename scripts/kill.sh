#!/bin/sh

ps aux | grep -v grep | grep enen

killall enen
