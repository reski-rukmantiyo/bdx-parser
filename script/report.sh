#!/bin/bash

# Run the report script for each PDU
for pdu in a1 a2 a3 a4 a5 b1 b2 b3 b4 b5 c1 c2 c3 c4 c5; do
    go run pkg/report/report.go total_${pdu}.csv monthly-june-2025.xlsx result.csv
done