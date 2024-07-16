#!/bin/bash

FILE_NAME="multipleFrameSizePlotter.py"
RESULT_PATH="/home/ron/Desktop/RemoteStudioLive/Results"
SETUP="Wifi connection - LAN"

open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 1"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 2"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 3"/"$FILE_NAME" &

open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 1"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 2"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 3"/"$FILE_NAME" &

SETUP="VPN Connection"

open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 1"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 2"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 3"/"$FILE_NAME" &

open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 1"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 2"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 3"/"$FILE_NAME" &

SETUP="Wired Connection"

open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 1"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 2"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"UDP"/"Set 3"/"$FILE_NAME" &

open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 1"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 2"/"$FILE_NAME" &
open "$RESULT_PATH"/"$SETUP"/"TCP"/"Set 3"/"$FILE_NAME" &



