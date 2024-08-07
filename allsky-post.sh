#!/bin/bash

curl -q --max-time 5 --connect-timeout 1 http://localhost:18888/notify \
        -d AS_25544ALT="$AS_25544ALT" \
        -d AS_25544VISIBLE="$AS_25544VISIBLE" \
        -d AS_DATE="$AS_DATE" \
        -d AS_EXPOSURE_US="$AS_EXPOSURE_US" \
        -d AS_GAIN="$AS_GAIN" \
        -d AS_METEORCOUNT="$AS_METEORCOUNT" \
        -d AS_STARCOUNT="$AS_STARCOUNT" \
        -d AS_SUN_SUNRISE="$AS_SUN_SUNRISE" \
        -d AS_SUN_SUNSET="$AS_SUN_SUNSET" \
        -d AS_TEMPERATURE_C="$AS_TEMPERATURE_C" \
        -d AS_TIME="$AS_TIME" \
        -d CURRENT_IMAGE="$CURRENT_IMAGE" \
        -d DATE_NAME="$DATE_NAME"


