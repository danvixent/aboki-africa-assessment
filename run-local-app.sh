#!/bin/bash
set -e

cd cmd
go run main.go -config_path="../config/config.yml"
