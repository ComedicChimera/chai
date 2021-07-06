; NOTE TO READER: This is temporary assembly and probably does not work yet.  
; I am leaving this here as a "template" for now.
 
global _start

extern __init
extern __main
extern __exit

section .text

_start:
    ; set up the end of the stack frame linked list
    mov $0, %rbp
    push %rbp # rip = 0
    push %rbp # rbp = 0
    mov %rsp, %rbp

    ; TODO: handle cmdl args

    ; TODO: initialize runtime and standard library
    
    ; initialize global variables; call global `init`
    call __init

    ; call the user's main function
    call __main

    ; gracefully exit our program
    mov %eax, $0
    call __exit
    ret 



