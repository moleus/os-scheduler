# OS scheduling algorithms implementations


# Theory
*selection function* - chooses a process among ready for execution.
Three quantities:
- `w` = time spend in system so far, waiting ??
- `e` = time spent in execution so far
- `s` = total service time required by the process, including `e`; generally, this quan- tity must be estimated or supplied by the user

*HRRN* - $$ max(\frac{w+s}{s}) $$

*decision mode* - when to execute selection function
1. *Nonpreemptive* - process executes until terminates or blocks to wait for I/O
2. *Preemptive* - running process can be interrupted at _any_ time

# Time measure
- service time (Ts) - total sum of CPU and IO cycles
- turnaround time (Tr) - total time in system. Ts + waiting
- normalized turnaround (Tr/Ts) - relative delay experienced by a process
