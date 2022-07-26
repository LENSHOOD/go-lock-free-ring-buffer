#!/bin/sh
go test -bench MPMC$ -benchtime=10s -count=10 | benchstat /dev/stdin
