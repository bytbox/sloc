#!/bin/sh
pod2man -c "" -r "`git describe`" reposloc > reposloc.1
