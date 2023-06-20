#!/bin/sh
pod2man -c "" -r "`git describe --tags`" reposloc > reposloc.1
