EOT         .BYT        3
NL          .BYT        10
Space       .BYT        32

ZERO        .INT    0
I           .INT    1
II          .INT    2
III         .INT    3
IV          .INT    4
V           .INT    5
VI          .INT    6
VII         .INT    7
VIII        .INT    8
IX          .INT    9
X           .INT    10

cZERO       .BYT    48
cI          .BYT    49
cII         .BYT    50
cIII        .BYT    51
cIV         .BYT    52
cV          .BYT    53
cVI         .BYT    54
cVII        .BYT    55
cVIII       .BYT    56
cIX         .BYT    57


###### int fib(int n) ######
fib     MOV     R8      SP          # test for SO with Ret Addr + PFP + passed parameter + locals and temps
        ADI     R8      -16          # space for 5 int temps
        CMP     R8      SL          # test for SO
        BLT     R8      OVERFLOW
        ADI     SP      -20          # adjust SP to word on top of AR

        MOV     R6      FP          # 
        ADI     R6      -8         # location for param n
        LDR     R5      R6
        MOV     R7      R5
        ADI     R7      -1
        ADI     R6      -4
        STR     R7      R6          # store (n-1) in t1
        ADI     R7      -1
        ADI     R6      -4
        STR     R7      R6          # store (n-2) in t2
        
        MOV     R6      FP          # 
        ADI     R6      -8         # location for param n
        LDR     R8      R6
        LDR     R5      I           # if (n <= 1)
        CMP     R5      R8
        BLT     R5      fib_else
    # return n
        MOV     SP      FP
        MOV     R4      SP          
        ADI     R4      -4      # point at PFP
        LDR     FP      R4
        MOV     R6      SP
        CMP     R6      SB
        BGT     R6      UNDERFLOW

    # store return value
    # function complete. return to caller
        MOV     R6      SP      # SP points to return adddress
        LDR     R6      R6
        STR     R8      SP      # set return value 
        JMR     R6

    # prepare for else condition call
fib_else TRP 99
        MOV     R8      SP          # compute space needed for activation record
        ADI     R8      -4          # Adjust for space needed (Rtn Address & PFP)
        ADI     R8      -4          # Adjust for space for passed paramters
        CMP     R8      SL          # test for SO
        BLT     R8      OVERFLOW

        MOV     R8      FP          # Save FP in R8, this will be the PFP
        ADI     FP      -12         # t1
        LDR     R1      FP
        MOV     FP      SP          # Point at Current Activation Record (FP = SP)
        ADI     SP      -0          # Adjust SP for Rtn Address. To be pushed on later
        MOV     R6      SP          # R6 holding address for Rtn Address
        ADI     SP      -4          # Space for PFP
        STR     R8      SP          # PFP to Top of Stack

    # Pass parameters on the stack
        ADI     SP      -4
        STR     R1      SP          # t1 = n-1
        
    # Update the return address in the Stack Frame
        ADI     SP      -4          # SP points to word before activation rec
        MOV     R7      PC
        ADI     R7      48          # Compute Return Address (12 * # of instructions until instruction after JMP)
        STR     R7      R6          # push return address on the stack. R7 is holding ret address -> R6 holding stack location for ret address
        
    # Jump to function    
        JMP     fib

    # get fib(n) return value
        LDR     R3      SP          # should be store in a temp
        MOV     R6      FP          # 
        ADI     R6      -20         # location for t3
        STR     R3      R6          # store return value in t3

    # prepare for fib(n-2)
        MOV     R8      SP          # compute space needed for activation record
        ADI     R8      -4          # Adjust for space needed (Rtn Address & PFP)
        ADI     R8      -4          # Adjust for space for passed paramters
        CMP     R8      SL          # test for SO
        BLT     R8      OVERFLOW

        TRP 99
        MOV     R8      FP          # Save FP in R8, this will be the PFP
        ADI     FP      -16         # t2
        LDR     R1      FP
        MOV     FP      SP          # Point at Current Activation Record (FP = SP)
        ADI     SP      -0          # Adjust SP for Rtn Address. To be pushed on later
        MOV     R6      SP          # R6 holding address for Rtn Address
        ADI     SP      -4          # Space for PFP
        STR     R8      SP          # PFP to Top of Stack

    # Pass parameters on the stack
        TRP 99
        ADI     SP      -4
        STR     R1      SP          # t2 = n-2
        
    # Update the return address in the Stack Frame
        ADI     SP      -4          # SP points to word before activation rec
        MOV     R7      PC
        ADI     R7      48          # Compute Return Address (12 * # of instructions until instruction after JMP)
        STR     R7      R6          # push return address on the stack. R7 is holding ret address -> R6 holding stack location for ret address
        
    # Jump to function    
        JMP     fib

    # get fib(n) return value from TOS -4 (prev act record)
        LDR     R3      SP          
        MOV     R6      FP          
        ADI     R6      -24         
        STR     R3      R6          # store return value in t4

    # set t5 (t3+t4)
        MOV     R6      FP
        ADI     R6      -20
        LDR     R7      R6      # t3
        ADI     R6      -4
        LDR     R8      R6      # t4
        ADD     R8      R7
        ADI     R6      -4
        STR     R8      R6
        # MOV     R8      R6      # R8 holding addr or t5

    # begin return call
        MOV     SP      FP
        MOV     R4      SP          
        ADI     R4      -4      # point at PFP
        LDR     FP      R4
        MOV     R6      SP
        CMP     R6      SB
        BGT     R6      UNDERFLOW

    # store return value
    # function complete. return to caller
        MOV     R6      SP      # 
        LDR     R6      R6      # R6 has return address
        STR     R8      SP      # set return value 
        JMR     R6




###### START OF PROGRAM ######
START   MOV     R8      SP          # compute space needed for activation record
        ADI     R8      -8          # Adjust for space needed (Rtn Address & PFP)
        ADI     R8      -16         # Adjust for space for passed paramters reset(1, 0, 0, 0) <- 4 ints = 16 bytes
        CMP     R8      SL          # test for SO
        BLT     R8      OVERFLOW

        TRP     2                   # get int from user
        LDB     R1      ZERO        # check stop condition
        CMP     R1      R3
        BRZ     R1      fib_stop

                                    # prepare for fib function
                                    # store ret and PFP
        MOV     R8      FP          # Save FP in R8, this will be the PFP
        MOV     FP      SP          # Point at Current Activation Record (FP = SP)
        ADI     SP      -0          # Adjust SP for Rtn Address. To be pushed on later
        MOV     R6      SP          # R6 holding address for Rtn Address
        ADI     SP      -4          # Space for PFP
        STR     R8      SP          # PFP to Top of Stack

    # Pass parameters on the stack
        ADI     SP      -4
        STR     R3      SP          # n value for fib(n)
        
    # Update the return address in the Stack Frame
        ADI     SP      -4          # SP points to word before activation rec
        MOV     R7      PC
        ADI     R7      48          # Compute Return Address (12 * # of instructions until instruction after JMP)
        STR     R7      R6          # push return address on the stack. R7 is holding ret address -> R6 holding stack location for ret address
        
    # Jump to function    
        JMP     fib

    # get fib(n) return value from TOS -4 (prev act record)
        LDR     R3      SP

        TRP     1


END     LDR SS  R3


UNDERFLOW   TRP     0
OVERFLOW    TRP     0