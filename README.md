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


# Lab variant 91582
program input:
```
CPU(6);IO2(16);CPU(6);IO2(20);CPU(10);IO1(16);CPU(8);IO2(20);CPU(2);IO1(12);CPU(4);IO1(18);
CPU(4);IO2(18);CPU(10);IO2(16);CPU(4);IO1(14);CPU(10);IO1(12);CPU(6);IO1(20);
CPU(2);IO2(12);CPU(6);IO1(16);CPU(4);IO2(12);CPU(2);IO1(14);CPU(6);IO1(12);
CPU(10);IO1(20);CPU(4);IO2(12);CPU(10);IO1(10);CPU(8);IO2(14);CPU(10);IO1(18);
CPU(8);IO2(14);CPU(8);IO2(20);CPU(8);IO1(10);CPU(6);IO2(20);CPU(8);IO2(18);CPU(2);IO2(20);
CPU(12);IO2(10);CPU(48);IO2(18);CPU(12);IO1(18);CPU(24);IO1(14);CPU(48);IO1(20);CPU(24);IO2(14);CPU(36);IO2(10);
```
