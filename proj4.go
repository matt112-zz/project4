package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func init() {
	initInstructionTable()
	initRegisterTable()
	flush()

	// f, err := os.OpenFile("testfmtfile.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	fmt.Fatalf("error opening file: %v", err)
	// }
	// defer f.Close()

	// fmt.SetOutput(f)
	// fmt.SetLevel(fmt.PanicLevel)
}

type Instruction struct {
	opCode, opd1, opd2 int32
}

const (
	INSTRUCTION_SIZE = 12
	INT_SIZE         = 4
	BYT_SIZE         = 1
	BYTE_SIZE        = 4
	BYT_DIR          = ".BYT"
	INT_DIR          = ".INT"
	INS              = "INS"
	MEMORY_SIZE      = 10000
	REGISTER_SIZE    = 14
)

var symbol_table = make(map[string]int32)
var instruction_table = make(map[string]int32)
var register_table = make(map[string]int32)
var memory = make([]byte, MEMORY_SIZE)
var registers [REGISTER_SIZE]int32
var PC_to_ASM_Line = make(map[int32]int)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func flush() {
	for i := 0; i < REGISTER_SIZE; i++ {
		registers[i] = 0
	}
}

func initRegisterTable() {
	register_table["R1"] = 1
	register_table["R2"] = 2
	register_table["R3"] = 3
	register_table["R4"] = 4
	register_table["R5"] = 5
	register_table["R6"] = 6
	register_table["R7"] = 7
	register_table["R8"] = 8
	register_table["PC"] = 9
	register_table["SP"] = 10 // stack pointer
	register_table["SL"] = 11 // stack limit
	register_table["SB"] = 12 // bottom of stack
	register_table["FP"] = 13 // frame pointer. bottom of current frame
}

func initInstructionTable() {
	// JUMP INSTRUCTIONS
	instruction_table["JMP"] = 1
	instruction_table["JMR"] = 2
	instruction_table["BNZ"] = 3
	instruction_table["BGT"] = 4
	instruction_table["BLT"] = 5
	instruction_table["BRZ"] = 6

	// MOVE INSTRUCTIONS
	instruction_table["MOV"] = 7
	instruction_table["LDA"] = 8
	instruction_table["STR"] = 9
	instruction_table["LDR"] = 10
	instruction_table["STB"] = 11
	instruction_table["LDB"] = 12

	// ARITHMETIC INSTRUCTIONS
	instruction_table["ADD"] = 13
	instruction_table["ADI"] = 14
	instruction_table["SUB"] = 15
	instruction_table["MUL"] = 16
	instruction_table["DIV"] = 17

	// fmtICAL INSTRUCTIONS
	instruction_table["AND"] = 18
	instruction_table["OR"] = 19

	// COMPARE INSTRUCTIONS
	instruction_table["CMP"] = 20

	// TRAPS
	instruction_table["TRP"] = 21

	// REGISTER INDIRECT ADDRSSING
	instruction_table["STRI"] = 22
	instruction_table["LDRI"] = 23
	instruction_table["STBI"] = 24
	instruction_table["LDBI"] = 25
}

func writeInt(mem []byte, val string) error {
	log.Trace("writeInt() with ", val)

	if i, err := strconv.Atoi(val); err == nil {
		ui := uint32(i)
		binary.LittleEndian.PutUint32(mem, ui)
		return nil
	} else {
		return err
	}
}

func writeBytecode(mem []byte, bytecode string) {
	log.Trace("writeBytecode() with ", bytecode)

	tokens := strings.Fields(bytecode)
	writeInt(mem, tokens[0])
	writeInt(mem[4:], tokens[1])
	if len(tokens) == 3 {
		writeInt(mem[8:], tokens[2])
	}
}

func fetch(mem []byte) Instruction {
	// log.Trace("fetchInstruction()")

	var ins int32
	var op1 int32
	var op2 int32

	ins = int32(binary.LittleEndian.Uint32(mem[0:]))
	op1 = int32(binary.LittleEndian.Uint32(mem[4:]))
	op2 = int32(binary.LittleEndian.Uint32(mem[8:]))

	IR := Instruction{
		opCode: ins,
		opd1:   op1,
		opd2:   op2,
	}

	// log.Trace("With ins(", ins, ") op1(", op1, ") op2(", op2, ")")
	return IR
}

func fetchInt(mem []byte) int32 {
	// log.Trace("fetchInt()")
	var v int32
	v = int32(binary.LittleEndian.Uint32(mem[0:]))
	return v
}

func printMemory() {
	for i, str := range memory {
		println(i, ": ", str)
	}
	println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
}

func firstpass(filename string) {
	log.Trace("~~~~~~~~~~~~~Start of first pass~~~~~~~~~~~~~~~~")

	var count int32
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	var label string
	var directive string
	tokens := make([]string, 0)

	for scanner.Scan() {
		line = scanner.Text()

		if i := strings.Index(line, "#"); i > -1 {
			line = line[:i]
		}
		tokens = strings.Fields(line)

		// ignore empty lines
		if len(tokens) == 0 {
			continue
		}

		log.Trace("First pass with line: ", line, " at memory location : ", count)

		// traps and such
		if len(tokens) == 2 {
			log.Trace("There are two tokens")
			if _, ok := instruction_table[tokens[0]]; ok {
				// log.Trace("found ", tokens[0], " in instruction table!")
				count += INSTRUCTION_SIZE
			} else {
				directive = tokens[0]
				if directive == INT_DIR {
					count += INT_SIZE
				} else if directive == BYT_DIR {
					count += BYT_SIZE
				}
			}
		} else if len(tokens) == 3 {
			log.Trace("There are 3 tokens")
			if _, ok := instruction_table[tokens[0]]; ok {
				count += INSTRUCTION_SIZE
			} else if _, ok := instruction_table[tokens[1]]; ok {
				// log.Trace("found ", tokens[0], " in instruction table!")
				if _, ok := symbol_table[tokens[0]]; !ok {
					// log.Trace("Label not in symbol table. Label for an instruction.")
					symbol_table[tokens[0]] = count
					count += INSTRUCTION_SIZE
				} else {
					log.Trace(tokens[0])
					log.Fatal("Duplicate symbol in table")
				}
			} else { // directive
				label = tokens[0]
				directive = tokens[1]
				if _, ok := symbol_table[label]; !ok {
					// log.Trace("Label not in symbol table")
					symbol_table[label] = count
					if directive == INT_DIR {
						count += INT_SIZE
					} else if directive == BYT_DIR {
						count += BYT_SIZE
					}
				} else {
					log.Trace(label)
					log.Fatal("Duplicate symbol in table")
				}
			}
		} else if len(tokens) == 4 {
			log.Trace("There are 4 tokens")
			label = tokens[0]
			if _, ok := symbol_table[label]; !ok {
				symbol_table[label] = count
				count += INSTRUCTION_SIZE
			} else {
				// log.Trace(label)
				// fmt.Fatal("Duplicate symbol in table")
			}
		}
	}

	log.Trace("~~~~~~~~~~~~~~~~~End of first pass~~~~~~~~~~~~~~")

}

func secondpass(filename string) {
	log.Trace("~~~~~~~~~~~~Start of second pass~~~~~~~~~~~~~~~")

	var count int32
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	var instruction string
	var label string
	var data string
	var directive string
	var bytecode string
	var newByte int
	var lineNum = 0
	pArr := false
	tokens := make([]string, 0)
	rest := make([]string, 0)
	operand := make([]string, 0)

	for scanner.Scan() {
		lineNum++
		line = scanner.Text()
		if i := strings.Index(line, "#"); i > -1 {
			line = line[:i]
		}

		tokens = strings.Fields(line)

		// ignore empty lines
		if len(tokens) == 0 {
			continue
		}

		log.Trace("Second pass with line: ", line, " at memory location : ", count)

		// traps and such
		if len(tokens) == 2 {
			log.Trace("There are 2 tokens")

			if _, ok := instruction_table[tokens[0]]; ok {
				// log.Trace("found ", tokens[0], " in instruction table!")
				instruction = tokens[0]
				operand = tokens[1:]
				bytecode = instruction_to_bytecode(instruction, operand)
				writeBytecode(memory[count:], bytecode)
				count += INSTRUCTION_SIZE
			} else { // data for the arr
				directive = tokens[0]
				data = tokens[1]
				if directive == INT_DIR {
					writeInt(memory[count:], data)
					count += INT_SIZE
				} else if directive == BYT_DIR {
					newByte, err = strconv.Atoi(data)
					memory[count] = byte(newByte)
					count += BYT_SIZE
				}
			}
		} else if len(tokens) == 3 {
			log.Trace("There are 3 tokens")

			if _, ok := instruction_table[tokens[0]]; ok {
				instruction = tokens[0]
				rest = tokens[1:]
				bytecode = instruction_to_bytecode(instruction, rest)
				writeBytecode(memory[count:], bytecode)
				count += INSTRUCTION_SIZE
			} else if _, ok := instruction_table[tokens[1]]; ok {
				// log.Trace("found ", tokens[0], " in instruction table!")
				if _, ok := symbol_table[tokens[0]]; ok {
					// log.Trace("Label not in symbol table. Label for an instruction.")
					instruction = tokens[1]
					rest = tokens[2:]
					bytecode = instruction_to_bytecode(instruction, rest)
					writeBytecode(memory[count:], bytecode)
					count += INSTRUCTION_SIZE
				} else {
					log.Trace(tokens[0])
					log.Fatal("Symbol not found on first pass")
				}
			} else { // directive
				label = tokens[0]
				if pArr == false {
					pArr = true
				}
				directive = tokens[1]
				data = tokens[2]
				if _, ok := symbol_table[label]; ok {
					if directive == INT_DIR {
						writeInt(memory[count:], data)
						count += INT_SIZE
					} else if directive == BYT_DIR {
						newByte, err = strconv.Atoi(data)
						memory[count] = byte(newByte)
						count += BYT_SIZE
					}
				} else {
					// log.Trace(label)
					// fmt.Fatal("Duplicate symbol in table")
				}
			}
		} else if len(tokens) == 4 {
			log.Trace("There are 4 tokens")

			label = tokens[0]
			instruction = tokens[1]
			rest = tokens[2:]
			bytecode = instruction_to_bytecode(instruction, rest)
			writeBytecode(memory[count:], bytecode)
			if _, ok := symbol_table[label]; ok {
				count += INSTRUCTION_SIZE
			} else {
				// log.Trace(label)
				// fmt.Fatal("Duplicate symbol in table")
			}
		}
		PC_to_ASM_Line[count] = lineNum
	}

	log.Trace("~~~~~~~~~~~~~End of second pass~~~~~~~~~~~~~~~~~~")
}

func instruction_to_bytecode(instruction string, operands []string) string {
	log.Trace("instruction_to_bytecode() with instruction(", instruction, ") and operands(", operands, ")")
	var bytecode string

	if instruction == "STR" || instruction == "LDR" || instruction == "STB" || instruction == "LDB" {
		if _, ok := register_table[operands[1]]; ok { // if second operand is a register
			instruction += "I"
		}
	}

	switch instruction {
	case "JMP":
		return fmt.Sprintf("%d %d", instruction_table[instruction], symbol_table[operands[0]])
	case "JMR":
		return fmt.Sprintf("%d %d", instruction_table[instruction], register_table[operands[0]])
	case "BNZ":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "BGT":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "BLT":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "BRZ":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "MOV":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "LDA":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "STR":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "LDR":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "STB":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "LDB":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], symbol_table[operands[1]])
	case "ADD":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "ADI":
		return fmt.Sprintf("%d %d %s", instruction_table[instruction], register_table[operands[0]], operands[1])
	case "SUB":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "MUL":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "DIV":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "AND":
		log.Trace("AND not implemented")
	case "OR":
		log.Trace("OR not implemented")
	case "CMP":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "TRP":
		return fmt.Sprintf("%d %s", instruction_table[instruction], operands[0])
	case "STRI":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "LDRI":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "STBI":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	case "LDBI":
		return fmt.Sprintf("%d %d %d", instruction_table[instruction], register_table[operands[0]], register_table[operands[1]])
	}
	return bytecode
}

func virtualmachine() {
	log.Trace("~~~~~~~~~~~~~~~~~Running Virtual Machine~~~~~~~~~~~~~~")

	nums := []int32{10, 3, 5, 7, 2, 9, 12, 14, 0}
	var (
		val, d_reg, s_reg, m_addr                  int32
		s_addr, mode                               int32
		s_reg_name, d_reg_name, g_reg_name, l_addr int32
	)
	var PC int32
	var err error
	var IR Instruction
	kb_buff := make([]byte, 0)
	kb_buff = []byte("9\n")
	reader := bufio.NewReader(os.Stdin)
	branch := false
	done := false
	stack := int32(MEMORY_SIZE)
	registers[register_table["FP"]] = stack - 4
	registers[register_table["SL"]] = stack - 1000 // todo: where is heap?
	registers[register_table["SP"]] = stack - 4
	registers[register_table["SB"]] = stack
	log.Trace("Stack at location(", stack, ") in memory")
	PC = symbol_table["START"]
	registers[register_table["PC"]] = PC
	for !done {
		branch = false
		PC = registers[register_table["PC"]]
		IR = fetch(memory[PC:])

		switch IR.opCode {
		// TRAPS
		case instruction_table["TRP"]:
			mode = IR.opd1
			if mode == 0 {
				log.Info("TRP 0: End of program")
				done = true
				break
			} else if mode == 1 {
				log.Info("TRP 1")
				fmt.Print(registers[3])
			} else if mode == 2 {
				log.Info("TRP 2")
				// fmt.Print("Enter a num for fib: ")
				registers[3] = nums[0]
				nums = nums[1:]
				// _, _ = fmt.Scanf("%d", &registers[3])
			} else if mode == 3 {
				log.Info("TRP 3")
				val = registers[3]
				if val == 12 { // new line
					fmt.Print("\n")
				} else if val == 32 { // space
					fmt.Print(" ")
				} else {
					fmt.Print(string(val))
				}
			} else if mode == 4 {
				log.Info("TRP 4")
				if len(kb_buff) == 0 {
					fmt.Print("Enter chars: ")
					kb_buff, _ = reader.ReadBytes('\n')
				}
				registers[3] = int32(kb_buff[0])
				log.Info("~~~~~~~~~~~~~~~Register 3: ", string(registers[3]))
				kb_buff = kb_buff[1:]
			} else if mode == 99 {
				log.Info("TRP 99: ***********DEBUGGING LINE ", PC_to_ASM_Line[PC+12], " OF ASSEMBLY CODE****************")
			} else if mode == 98 {
				log.Info("~~~~~~~~~~~~~~Stack Pointer: ", registers[register_table["SP"]])
			}
		// JUMP INSTRUCTIONS
		case instruction_table["JMP"]:
			branch = true
			l_addr = IR.opd1
			registers[register_table["PC"]] = l_addr
			log.Infof("JMP instruction: Branching to label at location(%d)", l_addr)
			// printMemory()
		case instruction_table["JMR"]:
			branch = true
			l_addr = registers[IR.opd1]
			registers[register_table["PC"]] = l_addr
			log.Infof("JMR instruction: Branching to address in s_reg(%d)", l_addr)
		case instruction_table["BNZ"]:
			s_reg_name = IR.opd1
			s_reg = registers[s_reg_name]
			if s_reg != 0 {
				branch = true
				l_addr = IR.opd2
				registers[register_table["PC"]] = l_addr
			}
			log.Infof("BNZ instruction: Branching to label at location(%d) if s_reg(%d) != 0", l_addr, s_reg_name)
		case instruction_table["BGT"]:
			s_reg_name = IR.opd1
			s_reg = registers[s_reg_name]
			if s_reg > 0 {
				branch = true
				l_addr = IR.opd2
				registers[register_table["PC"]] = l_addr
			}
			log.Infof("BGT instruction: Branching to label at location(%d) if s_reg(%d) > 0", l_addr, s_reg_name)
		case instruction_table["BLT"]:
			s_reg_name = IR.opd1
			s_reg = registers[s_reg_name]
			if s_reg < 0 {
				branch = true
				l_addr = IR.opd2
				registers[register_table["PC"]] = l_addr
			}
			log.Infof("BLT instruction: Branching to label at location(%d) if s_reg(%d) < 0", l_addr, s_reg_name)
		case instruction_table["BRZ"]:
			s_reg_name = IR.opd1
			s_reg = registers[s_reg_name]
			if s_reg == 0 {
				branch = true
				l_addr = IR.opd2
				registers[register_table["PC"]] = l_addr
			}
			log.Infof("BRZ instruction: Branching to label at location(%d) if s_reg(%d) = 0", l_addr, s_reg_name)

		// MOVE INSTRUCTIONS
		case instruction_table["MOV"]:
			d_reg_name = IR.opd1
			s_reg_name = IR.opd2
			registers[d_reg_name] = registers[s_reg_name]
			log.Infof("MOV instruction: Moving s_reg(%d) into d_reg(%d)", s_reg_name, d_reg_name)
		case instruction_table["LDA"]:
			d_reg = IR.opd1
			l_addr = IR.opd2
			registers[d_reg] = l_addr
			log.Infof("LDA instruction: Loading address of label at location(%d) in d_reg(%d)", l_addr, d_reg)
		case instruction_table["STR"]:
			s_reg_name = IR.opd1
			l_addr = IR.opd2
			writeInt(memory[l_addr:], strconv.Itoa(int(registers[s_reg_name])))
			log.Infof("STR instruction: Storing data into Mem/Label at location(%d) in s_reg(%d)", l_addr, s_reg_name)
			// printMemory()
		case instruction_table["LDR"]:
			d_reg = IR.opd1
			s_addr = IR.opd2
			registers[d_reg] = fetchInt(memory[s_addr:])
			log.Infof("LDR instruction: Loading d_reg(%d) with data from Mem/Label at location(%d)", d_reg, s_addr)
		case instruction_table["STB"]:
			s_reg = IR.opd1
			s_addr = IR.opd2
			memory[s_addr] = byte(registers[s_reg])
			log.Infof("STB instruction: Storing byte into Mem/Label at location(%d) from s_reg(%d)", s_addr, s_reg)
		case instruction_table["LDB"]:
			d_reg = IR.opd1
			s_addr = IR.opd2
			registers[d_reg] = int32(memory[s_addr])
			log.Infof("LDB instruction: Loading byte into d_reg(%d) from Mem/Label at location (%d)", d_reg, s_addr)

		// ARITHMETIC INSTRUCTIONS
		case instruction_table["ADD"]:
			d_reg_name = IR.opd1
			d_reg = registers[d_reg_name]
			s_reg_name = IR.opd2
			s_reg = registers[s_reg_name]
			registers[d_reg_name] = d_reg + s_reg
			log.Info("ADD instructions")
		case instruction_table["ADI"]:
			d_reg_name = IR.opd1
			registers[d_reg_name] += IR.opd2
			log.Infof("ADI instruction: Adding immediatte(%d) to d_reg(%d)", IR.opd2, d_reg_name)
		case instruction_table["SUB"]:
			d_reg_name = IR.opd1
			d_reg = registers[d_reg_name]
			s_reg_name = IR.opd2
			s_reg = registers[s_reg_name]
			registers[d_reg_name] = d_reg - s_reg
			log.Info("SUB instructions")
		case instruction_table["MUL"]:
			d_reg_name = IR.opd1
			d_reg = registers[d_reg_name]
			s_reg_name = IR.opd2
			s_reg = registers[s_reg_name]
			registers[d_reg_name] = d_reg * s_reg
			log.Info("MUL instructions")
		case instruction_table["DIV"]:
			d_reg_name = IR.opd1
			d_reg = registers[d_reg_name]
			s_reg_name = IR.opd2
			s_reg = registers[s_reg_name]
			registers[d_reg_name] = d_reg / s_reg
			log.Info("DIV instructions")

		// LOGICAL INSTRUCTIONS
		case instruction_table["AND"]:
			log.Info("AND not implemented in VM")
		case instruction_table["OR"]:
			log.Info("OR not implemented in VM")

		// COMPARE INSRUCTIONS
		case instruction_table["CMP"]:

			d_reg_name = IR.opd1
			d_reg = registers[d_reg_name]
			s_reg_name = IR.opd2
			s_reg = registers[s_reg_name]
			if d_reg == s_reg {
				registers[d_reg_name] = 0
			} else if d_reg < s_reg {
				registers[d_reg_name] = -1
			} else if d_reg > s_reg {
				registers[d_reg_name] = 1
			}
			log.Info("CMP instructions")

		// REGISTER INDIRECT ADDRESSING INSTRUCTIONS
		case instruction_table["STRI"]:
			s_reg_name = IR.opd1
			g_reg_name = IR.opd2
			m_addr = registers[g_reg_name]
			writeInt(memory[m_addr:], strconv.Itoa(int(registers[s_reg_name])))
			log.Infof("STRI instruction: Storing data from s_reg(%d) to memory at RG(%d)", s_reg_name, g_reg_name)
		case instruction_table["LDRI"]:
			d_reg_name = IR.opd1
			g_reg_name = IR.opd2
			m_addr = registers[g_reg_name]
			registers[d_reg_name] = fetchInt(memory[m_addr:])
			log.Infof("LDRI instruction: Load d_reg(%d) with data from memory at d_reg(%d)", d_reg_name, g_reg_name)
		case instruction_table["STBI"]:
			s_reg_name = IR.opd1
			g_reg_name = IR.opd2
			m_addr = registers[g_reg_name]
			memory[m_addr] = byte(registers[s_reg_name])
			// printMemory()
			log.Info("STBI instructions")
		case instruction_table["LDBI"]:
			d_reg_name = IR.opd1
			g_reg_name = IR.opd2
			m_addr = registers[g_reg_name]
			registers[d_reg_name] = int32(memory[m_addr])
			log.Info("LDBI instructions")
		}

		check(err)

		if !branch {
			registers[register_table["PC"]] += INSTRUCTION_SIZE
		}
	}
	log.Trace("~~~~~~~~~~~~~~End of virtualmachine()~~~~~~~~~~~~~~~~~~")
}

func main() {
	customFormatter := new(log.TextFormatter)
	log.SetFormatter(customFormatter)
	log.SetLevel(log.InfoLevel)
	customFormatter.DisableTimestamp = true
	_ = os.Remove("log.txt")
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	// log.SetOutput(ioutil.Discard)

	filename := "proj4.asm"
	firstpass(filename)
	secondpass(filename)
	virtualmachine()
}
