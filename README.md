# How to use

## Prequistes

You have to have file xlsx with format like this

```
Timestamp	B3 Q1 Current : l1	B3 Q1 Current : l2	B3 Q1 Current : l3	B3 Q2 Current : l1	B3 Q2 Current : l2	B3 Q2 Current : l3	B3 Q3 Current : l1	B3 Q3 Current : l2	B3 Q3 Current : l3	B3 Q4 Current : l1	B3 Q4 Current : l2	B3 Q4 Current : l3	B3 Q5 Current : l1	B3 Q5 Current : l2	B3 Q5 Current : l3	B3 Q6 Current : l1	B3 Q6 Current : l2	B3 Q6 Current : l3	B3 Q7 Current : l1	B3 Q7 Current : l2	B3 Q7 Current : l3	B3 Q8 Current : l1	B3 Q8 Current : l2	B3 Q8 Current : l3	B3 Q9 Current : l1	B3 Q9 Current : l2	B3 Q9 Current : l3	B3 Q10 Current : l1	B3 Q10 Current : l2	B3 Q10 Current : l3	B3 Q11 Current : l1	B3 Q11 Current : l2	B3 Q11 Current : l3	B3 Q12 Current : l1	B3 Q12 Current : l2	B3 Q12 Current : l3	B3 Q13 Current : l1	B3 Q13 Current : l2	B3 Q13 Current : l3	B3 Q14 Current : l1	B3 Q14 Current : l2	B3 Q14 Current : l3	B3 Q15 Current : l1	B3 Q15 Current : l2	B3 Q15 Current : l3	B3 Q16 Current : l1	B3 Q16 Current : l2	B3 Q16 Current : l3	B3 Q17 Current : l1	B3 Q17 Current : l2	B3 Q17 Current : l3	B3 Q18 Current : l1	B3 Q18 Current : l2	B3 Q18 Current : l3
31/07/2025 23:50:00	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	0	2.707	4.785	2.28	1.794	0	1.847	1.642	3.871	2.33	5.192	2.225	2.676	1.262	2.752	3.927	3.389	2.365	1.198	2.284	2.48	4.38	2.167	3.66	1.315	2.382	1.147	3.578
```

## Generate Summary

```
reski@Reskis-M4-Pro bdx-parser % ./bin/pdu-parser A4.xlsx 
```

### Example

```
reski@Reskis-M4-Pro bdx-parser % ./bin/pdu-parser A4.xlsx 
Loaded data from A4.xlsx for PDU A4: 54 columns processed
Output written to total_a4.csv
Processing completed successfully!
Input: A4.xlsx
Output: total_a4.csv
PDU processed: A4
```

## Fill into Template

The key is you already have output from Generate Summary per each PDU
then execute

```
./bin/monthly-filler total_a4.csv 
```

### Example

```
reski@Reskis-M4-Pro bdx-parser % ./bin/monthly-filler total_a4.csv 
Loaded PDU A4 data from total_a4.csv
Loading existing filled template: filled_monthly_report.csv (preserving previous data)
Found PDU A1: header at row 1, data rows 2-19
Found PDU A2: header at row 20, data rows 21-38
Found PDU A3: header at row 39, data rows 40-57
Found PDU A4: header at row 58, data rows 59-76
Found PDU A5: header at row 77, data rows 78-95
Found PDU B1: header at row 96, data rows 97-114
Found PDU B2: header at row 115, data rows 116-133
Found PDU B3: header at row 134, data rows 135-152
Found PDU B4: header at row 153, data rows 154-171
Found PDU B5: header at row 172, data rows 173-190
Found PDU C1: header at row 191, data rows 192-209
Found PDU C2: header at row 210, data rows 211-228
Found PDU C3: header at row 229, data rows 230-247
Found PDU C4: header at row 248, data rows 249-266
Found PDU C5: header at row 267, data rows 268-285
Loaded template from filled_monthly_report.csv (286 rows, 15 PDU sections)
Filling data into PDU A4 section (rows 59-76)
Successfully filled A4 data into PDU A4 section (including summary calculations)
Exported filled data to filled_monthly_report.csv

=== Processing completed successfully! ===
PDU Data Source: total_a4.csv
Monthly Template: monthlyjune2025.xlsx
Output: filled_monthly_report.csv
Filled PDU section: A4
Summary calculations: ✅ Current Min/AVG/Max per rack calculated
Mode: Preserve existing data ✅

📝 Next steps:
   Run: ./bin/monthly-filler total_<next_pdu>.csv
   This will add more PDU data while keeping existing sections intact.

📊 Summary per Rack columns now filled with:
   - Current Min: Minimum across L1/L2/L3 min values
   - Current AVG: Average across L1/L2/L3 avg values
   - Current Max: Maximum across L1/L2/L3 max values
```
