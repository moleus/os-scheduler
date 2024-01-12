#!/bin/sh

output="./results"
stats="./results/stats"
variant="data/91582.txt"

mkdir -p $output
mkdir -p $stats

./main -cpus 2 -algo fcfs -input $variant -log warn -output $output/fcfs.txt -procStats $stats/fcfs.stats
./main -cpus 2 -algo rr1  -input $variant -log warn -output $output/rr1.txt  -procStats $stats/rr1.stats
./main -cpus 2 -algo rr4  -input $variant -log warn -output $output/rr4.txt  -procStats $stats/rr4.stats
./main -cpus 2 -algo spn  -input $variant -log warn -output $output/spn.txt  -procStats $stats/spn.stats
./main -cpus 2 -algo srt  -input $variant -log warn -output $output/srt.txt  -procStats $stats/srt.stats
./main -cpus 2 -algo hrrn -input $variant -log warn -output $output/hrrn.txt -procStats $stats/hrrn.stats
